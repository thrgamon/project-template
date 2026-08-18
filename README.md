# Project Template

Go + Next.js + Postgres template with session-based authentication, type-safe code generation, and Dokku deployment.

## Prerequisites

- [mise](https://mise.jdx.dev/) (manages Go, Node versions and env vars)
- [Docker](https://www.docker.com/) (local dev environment)
- [just](https://just.systems/) (task runner)
- [yarn](https://classic.yarnpkg.com/) (Node package manager)

Run `just install-tools` to install the pinned versions of sqlc and goose.
tygo needs no install: it is pinned in `go.mod` and run via `go tool tygo`.

## Quick Start

1. Clone and rename:
   ```bash
   gh repo create myapp --template thrgamon/project-template
   cd myapp
   ```

2. Update the Go module path:
   ```bash
   fd -t f -e go -x sed -i '' 's|github.com/thrgamon/project-template|github.com/thrgamon/myapp|g' {}
   sed -i '' 's|github.com/thrgamon/project-template|github.com/thrgamon/myapp|g' tygo.yaml
   go mod edit -module github.com/thrgamon/myapp
   ```

3. Update `mise.toml` with your app name for `POSTGRES_DB` and `DATABASE_URL`.

4. Start development:
   ```bash
   just dev
   ```

   Backend: http://localhost:8080, Frontend: http://localhost:3000

## Project Structure

```
cmd/server/          # Go entrypoint
internal/
  api/               # HTTP handlers (HandlerConfig struct)
  auth/              # Auth service + middleware
  config/            # Environment-based config
  db/                # sqlc generated (DO NOT EDIT)
  domain/            # API request/response types (source of truth, feeds tygo)
  middleware/         # Request ID, logging
  server/            # HTTP server setup, routing, CORS
migrations/          # goose SQL migrations
queries/             # sqlc SQL query files
src/                 # Next.js App Router frontend
  app/               # Pages (login, register, dashboard)
  lib/               # Auth context, query provider
  lib/api/types.ts   # tygo generated from internal/domain (DO NOT EDIT)
  lib/api/client.ts  # Typed fetch wrapper over the Go API
  lib/api/hooks.ts   # React Query hooks built on the client
  components/        # Shared components (ErrorBanner, shadcn/ui)
e2e/                 # Playwright end-to-end tests
monitoring/          # Grafana, Prometheus, Loki, Tempo configs
deploy/              # Dokku entrypoint script
```

## Code Generation

After changing `migrations/`, `queries/`, or `internal/domain/`:

```bash
just sync
```

Two generators, each with one input and one output:

| Generator | Input | Output |
|-----------|-------|--------|
| sqlc | `migrations/` + `queries/` | `internal/db/` |
| tygo | `internal/domain/` | `src/lib/api/types.ts` |

`internal/domain` is the single source of truth for the API shape. The
TypeScript types are generated from it; `src/lib/api/client.ts` and
`src/lib/api/hooks.ts` are hand-written and consume those types, so a
mismatch between the Go response and the frontend is a type error.

tygo is pinned in `go.mod` as a tool dependency, so it needs no separate
install: `go tool tygo generate` works on a fresh clone.

## Auth Flow

Session-based authentication using HTTP-only cookies:

1. **Register** -- `POST /api/auth/register` -- creates user + session, sets cookie
2. **Login** -- `POST /api/auth/login` -- validates credentials, sets cookie
3. **Me** -- `GET /api/auth/me` -- returns current user (requires auth)
4. **Logout** -- `POST /api/auth/logout` -- deletes session, clears cookie

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Backend server port |
| `DATABASE_URL` | `postgres://...localhost.../myapp` | PostgreSQL connection string |
| `ENVIRONMENT` | `development` | `development` or `production` |
| `SESSION_MAX_AGE` | `604800` | Session duration in seconds (7 days) |
| `COOKIE_SECURE` | `false` | Set `true` in production (HTTPS only) |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | (empty) | Set to enable OpenTelemetry (no-op if unset) |
| `API_URL` | `http://localhost:8080` | Backend URL for Next.js rewrites |

## Commands

```bash
just dev              # Start all services
just test             # Run Go tests
just check            # Lint + test + type-check
just fmt              # Format Go code
just sync             # Regenerate sqlc + tygo output
just migrate          # Run migrations
just e2e              # Run Playwright tests
just dev-monitoring   # Start with Grafana/Prometheus/Loki/Tempo
just dokku-deploy     # Deploy to Dokku
just install-hooks    # Install pre-push hook
```

## Deployment (Dokku)

1. Create app: `dokku apps:create myapp`
2. Create DB: `dokku postgres:create myapp-db && dokku postgres:link myapp-db myapp`
3. Set config: `dokku config:set myapp ENVIRONMENT=production COOKIE_SECURE=true`
4. Add remote: `git remote add dokku dokku@your-server:myapp`
5. Deploy: `just dokku-deploy`

Migrations run automatically on deploy via `app.json` predeploy hook.
