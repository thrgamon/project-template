# Comprehensive Playwright E2E Test Suite

## Summary

End-to-end test suite using Playwright that tests every functional surface of the Dianne web app against a near-production stack. Tests use real OpenAI LLM calls, real Claude Code with OAuth authentication for channel bridge delegation, and a test double email client injected via dependency injection. The test environment is fully managed via Docker Compose and runnable with a single command.

## Decisions

- **LLM calls:** Real OpenAI API (gpt-4.1-mini for cost efficiency). Tests validate the full agent loop including tool dispatch.
- **Email client:** Go interface + in-process test double. Swapped in when `ENVIRONMENT=test`. Returns canned emails covering newsletters, threads, unsubscribe headers.
- **Claude channels:** Real Claude Code with production OAuth authentication. Credentials mounted from the host. Separated into its own Playwright project (`e2e-channels`) due to cost and latency.
- **Test environment:** Fully managed via `just e2e`. Docker Compose builds the app from `Dockerfile.dokku` (near-production), starts Postgres, waits for health checks, runs Playwright, tears down.
- **DB isolation:** Truncate tables per test file via a test-only `POST /api/test/reset` endpoint (only available when `ENVIRONMENT=test`). Each file seeds its own data.

## Backend Changes

### EmailClient Interface

Define an `EmailClient` interface in `internal/email/types.go` extracted from what the email plugin and sync service consume:

```go
type EmailClient interface {
    ListEmails(ctx context.Context, opts ListOptions) ([]Email, error)
    GetEmail(ctx context.Context, id string) (*Email, error)
    SendEmail(ctx context.Context, msg OutgoingEmail) error
}
```

The exact method set will be determined by auditing the callers in `emailplugin` and `emailsync.Service`.

### TestEmailClient

`internal/email/testclient.go` implements `EmailClient` with a fixed dataset of ~10 emails:
- 3 regular messages (1 unread)
- 4 newsletters with `List-Unsubscribe` headers (2 with RFC 8058 `List-Unsubscribe-Post`, 2 with plain URL)
- 2 threaded conversation messages
- 1 email with attachment flag

Wired in `cmd/dianne-kernel/main.go` when `cfg.Environment == "test"` instead of the real Fastmail JMAP client.

### SyncClient Interface

The email sync service (`emailsync.Service`) uses a separate sync client (`email.SyncClient` from `internal/email/sync.go`) for JMAP delta sync. Define an interface and test double for this as well:

```go
type SyncClient interface {
    GetState(ctx context.Context) (string, error)
    GetChanges(ctx context.Context, sinceState string) (*Changes, error)
    GetEmailsFull(ctx context.Context, ids []string) ([]SyncEmail, error)
    QueryAllIDs(ctx context.Context) ([]string, string, error)
}
```

The test implementation returns a stable state token and the canned email set, simulating a completed initial sync with no pending changes.

### Test Reset Endpoint

`internal/api/handler_test_reset.go`:

```go
func (h *Handler) TestReset(c *gin.Context) {
    // Only available when ENVIRONMENT=test
    // Truncates all data tables (not schema/migrations)
    // Seeds test user: test@dianne.test / testpassword123
}
```

Registered in `server_kernel.go` only when `cfg.Environment == "test"`:

```go
if strings.EqualFold(sc.Cfg.Environment, "test") {
    api.POST("/test/reset", sc.Handler.TestReset)
}
```

## Test Infrastructure

### Docker Compose

`docker-compose.test.yml` (override file used with `docker-compose.yml`):

```yaml
services:
  app:
    build:
      context: .
      dockerfile: Dockerfile.dokku
    environment:
      - ENVIRONMENT=test
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - DATABASE_URL=postgres://postgres:postgres@postgres:5432/dianne_test?sslmode=disable
      - TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN}
    volumes:
      - ~/.claude:/home/claude/.claude:ro  # Claude Code OAuth credentials
    depends_on:
      postgres:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:5000/health"]
      interval: 5s
      timeout: 3s
      retries: 30

  postgres:
    image: pgvector/pgvector:pg16
    environment:
      POSTGRES_DB: dianne_test
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 2s
      timeout: 3s
      retries: 10

  playwright:
    image: mcr.microsoft.com/playwright:v1.45.0-jammy
    working_dir: /app
    command: npx playwright test
    environment:
      - BASE_URL=http://app:5000
    volumes:
      - ./e2e:/app/e2e
      - ./playwright.config.ts:/app/playwright.config.ts
      - ./package.json:/app/package.json
    depends_on:
      app:
        condition: service_healthy
```

### Justfile

```
e2e:
    docker compose -f docker-compose.yml -f docker-compose.test.yml up --build --abort-on-container-exit --exit-code-from playwright

e2e-channels:
    PLAYWRIGHT_PROJECT=e2e-channels docker compose -f docker-compose.yml -f docker-compose.test.yml up --build --abort-on-container-exit --exit-code-from playwright
```

### Playwright Config

Two projects:
- `e2e` (default): all test files except `channels.spec.ts`. 45s timeout.
- `e2e-channels`: only `channels.spec.ts`. 180s timeout.

```typescript
export default defineConfig({
  testDir: './e2e',
  timeout: 45_000,
  use: {
    baseURL: process.env.BASE_URL || 'http://localhost:5000',
    trace: 'on-first-retry',
  },
  projects: [
    {
      name: 'e2e',
      testIgnore: ['**/channels.spec.ts'],
    },
    {
      name: 'e2e-channels',
      testMatch: ['**/channels.spec.ts'],
      timeout: 180_000,
    },
  ],
});
```

## Test Helpers

### `e2e/helpers/auth.ts`

```typescript
// loginViaUI(page) -- fills login form, submits, waits for dashboard
// loginViaAPI(context) -- POST /api/auth/login, stores cookie on context
// TEST_USER = { email: 'test@dianne.test', password: 'testpassword123' }
```

### `e2e/helpers/reset.ts`

```typescript
// resetDB(request) -- POST /api/test/reset, seeds test user
// Called in beforeAll() of each test file
```

### `e2e/helpers/wait.ts`

```typescript
// waitForStreamComplete(page) -- waits for SSE streaming to finish (no more tokens arriving)
// pollUntil(fn, timeout) -- generic polling helper for async operations
```

## Test Files

### `e2e/auth.spec.ts`

- Register a new user, verify redirect to chat
- Login with valid credentials, verify session
- Login with invalid credentials, verify error
- Logout, verify redirect
- Access protected page without auth, verify redirect to login
- Session persists across page reload

### `e2e/todos.spec.ts`

- Add a todo, verify it appears in the list
- Complete a todo, verify it's marked done
- Move a todo between sections (today/this week/someday)

### `e2e/goals.spec.ts`

- Create a goal with title and description
- Verify it appears in the list

### `e2e/themes.spec.ts`

- Create a theme
- Perform a check-in with notes
- Verify updated check-in state

### `e2e/projects.spec.ts`

- Create a project
- Verify it appears in the list

### `e2e/commitments.spec.ts`

- Create a commitment with due date
- Fulfill a commitment
- Break a commitment

### `e2e/reminders.spec.ts`

- Create a reminder
- Cancel a reminder
- Verify list reflects changes

### `e2e/bookmarks.spec.ts`

- Create a bookmark with URL and title
- Mark as read
- Archive
- Delete

### `e2e/books.spec.ts`

- Add a book
- Update status (reading, completed)
- Update notes

### `e2e/emails.spec.ts`

- Stats cards show correct counts (from test email client data)
- Subscription list shows newsletters with unsubscribe headers
- Filter subscriptions by sender name
- Select senders and unsubscribe
- Verify success/manual results display correctly

### `e2e/chat.spec.ts`

- Send a simple message, verify streamed response appears
- Send a tool-triggering message (e.g. "add a todo called 'buy milk'"), verify the todo was created by navigating to the todos page
- Send a memory-related message (e.g. "remember that I like Go"), verify via the memory page or a follow-up question
- Chat history: start a conversation, navigate away, come back, verify messages load from sidebar

### `e2e/settings.spec.ts`

- Toggle a plugin off, verify it appears disabled
- Toggle it back on

### `e2e/channels.spec.ts` (separate project)

- Send a delegation message via chat (e.g. "write a python hello world script")
- Verify the coding session starts (check for session status indicators in the UI or poll via API)
- Wait for completion (up to 120s)
- Verify the session reports success

## Out of Scope

- Voice/WebRTC testing (requires audio I/O)
- Telegram bot testing (requires Telegram API)
- Phone call/Telnyx testing (requires SIP)
- Push notification testing (requires service worker)
- Mobile/iOS app testing
- Performance/load testing
