# Project Template

A reusable GitHub template for Go + SvelteKit + Postgres projects with session-based authentication and structured request logging.

## Prerequisites

- [mise](https://mise.jdx.dev/) (manages Go, Node versions and env vars)
- [Docker](https://www.docker.com/) (local dev environment)
- [just](https://just.systems/) (task runner)
- [sqlc](https://sqlc.dev/) (Go code generation from SQL)
- [swag](https://github.com/swaggo/swag) (Swagger doc generation)

## Quick Start

1. Clone this template and rename:
   ```bash
   gh repo create myapp --template thrgamon/project-template
   cd myapp
   ```

2. Update the Go module path:
   ```bash
   # Replace all occurrences of github.com/thrgamon/project-template
   find . -type f -name '*.go' -exec sed -i '' 's|github.com/thrgamon/project-template|github.com/thrgamon/myapp|g' {} +
   go mod edit -module github.com/thrgamon/myapp
   ```

3. Update `mise.toml` — change `POSTGRES_DB` and `DATABASE_URL` to your app name.

4. Start development:
   ```bash
   just dev
   ```

This starts Postgres, the Go backend (with hot reload via air), and the SvelteKit frontend.

## Project Structure

```
cmd/server/          # Go entrypoint
internal/
  api/               # HTTP handlers with swag annotations
  auth/              # Auth service + middleware
  config/            # Environment-based config
  db/                # sqlc generated (DO NOT EDIT)
  domain/            # Request/response types
  middleware/         # Request ID, logging
  server/            # HTTP server setup, routing
migrations/          # goose SQL migrations
queries/             # sqlc SQL query files
src/                 # SvelteKit frontend
  routes/            # Pages (login, register, dashboard)
  lib/stores/        # Auth state (Svelte 5 runes)
  lib/api/generated/ # Orval generated client (DO NOT EDIT)
```

## Auth Flow

Session-based authentication using HTTP-only cookies:

1. **Register** — `POST /api/auth/register` — creates user + session, sets cookie
2. **Login** — `POST /api/auth/login` — validates credentials, creates session, sets cookie
3. **Me** — `GET /api/auth/me` — returns current user (requires auth)
4. **Logout** — `POST /api/auth/logout` — deletes session, clears cookie

Protected routes use the `auth.RequireAuth` middleware which reads the `session_token` cookie and validates against the database.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Backend server port |
| `DATABASE_URL` | `postgres://...localhost.../myapp` | PostgreSQL connection string |
| `ENVIRONMENT` | `development` | `development` or `production` |
| `SESSION_MAX_AGE` | `604800` | Session duration in seconds (7 days) |
| `COOKIE_SECURE` | `false` | Set `true` in production (HTTPS only) |
| `COOKIE_DOMAIN` | (empty) | Cookie domain restriction |

## Code Generation

After changing migrations, queries, or handler annotations:

```bash
just sync
```

This runs: `sqlc generate` → `swag init` → `npx orval` → type checks.

## Commands

```bash
just dev          # Start all services
just test         # Run all tests
just sync         # Regenerate all code
just migrate-up   # Run migrations
just lint         # Run linters
just check        # Lint + test + type-check
```
