# Project Template

Go + SvelteKit + Postgres template with Auth0 identity mapping, opaque
server-side sessions, and a single-process Dokku deployment. The frontend is
static-first: Go serves the compiled SvelteKit site and `/api` from one origin
in production.

## Prerequisites

- [mise](https://mise.jdx.dev/) (manages Go 1.27, Node 24 and env vars)
- [Docker](https://www.docker.com/) (local dev environment)
- [just](https://just.systems/) (task runner)
- [yarn](https://classic.yarnpkg.com/) (Node package manager)

Run `just install-tools` to install the pinned versions of sqlc and goose.

## Quick Start

1. Clone and rename:
   ```bash
   gh repo create myapp --private --template thrgamon/project-template
   cd myapp
   ```

2. Update the Go module path:
   ```bash
   fd -t f -e go -x sed -i '' 's|github.com/thrgamon/project-template|github.com/thrgamon/myapp|g' {}
   go mod edit -module github.com/thrgamon/myapp
   ```

3. Start development:
   ```bash
   just dev
   ```

   The first command creates a gitignored `.worktree.env` with a unique
   Compose project name, PostgreSQL database, volume set, and local port block
   for this worktree. It prints the chosen ports; use `FRONTEND_PORT` and
   `BACKEND_PORT` from that file to open the app. This lets multiple agents run
   `just dev` in separate worktrees without sharing containers, databases, or
   fixture state.

## Project Structure

```
cmd/server/          # Go entrypoint
internal/
  api/               # HTTP handlers (HandlerConfig struct)
  auth/              # Gin adapters for shared Auth0 middleware
  config/            # Environment-based config
  db/                # sqlc generated (DO NOT EDIT)
  domain/            # API request/response types
  middleware/         # Request ID, logging
  server/            # HTTP server setup, routing, CORS
migrations/          # goose SQL migrations
queries/             # sqlc SQL query files
frontend/            # SvelteKit static frontend
  src/routes/        # Pages (login, dashboard)
  src/lib/api.ts     # Handwritten browser DTOs and API client
  vite.config.ts     # SvelteKit options (adapter-static) and dev proxy
e2e/                 # Playwright end-to-end tests
monitoring/          # Grafana, Prometheus, Loki, Tempo configs
deploy/              # Dokku entrypoint script
```

## API types and code generation

After changing `migrations/` or `queries/`:

```bash
just sync
```

One generator is deliberately kept:

| Generator | Input | Output |
|-----------|-------|--------|
| sqlc | `migrations/` + `queries/` | `internal/db/` |

Keep browser-facing request and response DTOs hand-written in
`frontend/src/lib/api.ts`. They are a small, explicit boundary rather than a
generated mirror of Go structs. Test owned HTTP endpoints whenever that
contract changes.

## Rendering model

The default is a static SvelteKit build (`adapter-static`) served by Go. It is
one process, uses same-origin cookies, and needs no production CORS setup.

SSR is an explicit escape hatch: only adopt it for a documented requirement
that static HTML and browser API calls cannot meet. Then use adapter-node, add
a defined production process, and cover the server-rendered route end-to-end.

## Auth Flow

Auth0 verifies identity; this app decides who may access it. There is no
password or public registration endpoint.

1. An application owner runs `go run ./cmd/provision-user` with the exact
   Auth0 issuer and subject, creating the local user and explicit membership.
2. `GET /api/auth/login` begins Auth0 Authorization Code + PKCE authentication.
3. `GET /api/auth/callback` verifies the signed token, maps `(issuer, subject)`
   to an active membership, and stores only a hash of the opaque session token.
4. `GET /api/auth/me` returns the local user and CSRF token; state-changing
   browser requests send that value as `X-CSRF-Token`.
5. `POST /api/auth/logout` requires that CSRF token and revokes the current
   server-side session immediately.

The shared dependency is vendored from
`github.com/thrgamon/infra/go/auth` at
`v0.0.0-20260924045113-b1d1f320b79f`; Docker builds use that snapshot and do
not need credentials for the private infra repository.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Backend server port (the local `just backend` command assigns a worktree-specific port) |
| `DATABASE_URL` | `postgres://...localhost.../myapp` | PostgreSQL connection string (the local `just backend` command assigns an isolated database) |
| `ENVIRONMENT` | `development` | `development` or `production` |
| `SESSION_MAX_AGE` | `604800` | Session duration in seconds (7 days) |
| `COOKIE_SECURE` | `false` | Set `true` in production (HTTPS only) |
| `AUTH_ALLOW_INSECURE_COOKIES` | `false` | Set `true` only for local HTTP development; production must leave it unset. |
| `AUTH0_ISSUER_URL` | required | Auth0 issuer URL, including `https://` |
| `AUTH0_CLIENT_ID` | required | Confidential web application client ID |
| `AUTH0_CLIENT_SECRET` | required | Confidential web application client secret |
| `AUTH0_REDIRECT_URL` | required | Exact registered callback URL ending in `/api/auth/callback` |
| `AUTH_STATE_SECRET` | required | At least 32 random bytes used to sign short-lived login transactions |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | (empty) | Set to enable OpenTelemetry (no-op if unset) |
| `OTEL_EXPORTER_OTLP_HEADERS` | (empty) | OTLP auth, e.g. `Authorization=Basic%20<base64>` for Grafana Cloud (see Production monitoring) |
| `STATIC_DIR` | (empty) | Production SvelteKit build directory served by Go |

## Production monitoring

Deployed apps run on the shared Dokku host, whose monitoring is defined in
`thrgamon/infra` [`docs/monitoring.md`](https://github.com/thrgamon/infra/blob/main/docs/monitoring.md).
That file is the source of truth; in short:

- Keep `GET /api/health` unauthenticated and returning 200 when the app is
  healthy. Dokku uses it as the startup check (`app.json`) and the Grafana
  synthetic check uses it for uptime and TLS expiry.
- Register the app in infra `apps.yaml` with `health_path: /api/health` and
  run the synthetics reconciler there.
- Container metrics, logs, Postgres metrics and the shared "Dokku App"
  dashboard need no per-app setup. Do not create a per-app dashboard.
- To export traces, metrics and logs, issue a Grafana Cloud token named after
  the app on the `dokku` access policy (region `prod-au-southeast-1`; tokens
  from other regions are rejected) and set
  `OTEL_EXPORTER_OTLP_ENDPOINT=https://otlp-gateway-prod-au-southeast-1.grafana.net/otlp`
  and `OTEL_EXPORTER_OTLP_HEADERS=Authorization=Basic%20<base64 of 1540301:token>`.

### Concurrent worktrees

Use the `just` development commands (or `./scripts/compose`), rather than
plain `docker compose`. The wrapper creates `.worktree.env` once per worktree
and passes it to Compose. The file is local and gitignored; delete it only if
you deliberately want a new isolated local environment. `just db-reset` only
destroys the current worktree's database volume.

## Commands

```bash
just dev              # Start all services
just test             # Run Go tests
just check            # Lint + test + type-check
just fmt              # Format Go code
just sync             # Regenerate sqlc output + check the SvelteKit app
just migrate          # Run migrations
just e2e              # Run Playwright tests
just dev-monitoring   # Start with Grafana/Prometheus/Loki/Tempo
just dokku-deploy     # Deploy to Dokku
just install-hooks    # Install pre-push hook
```

## Deployment (Dokku)

1. Create app: `dokku apps:create myapp` and build from the production Dockerfile: `dokku builder-dockerfile:set myapp dockerfile-path Dockerfile.dokku`
2. Create DB: `dokku postgres:create myapp-db && dokku postgres:link myapp-db myapp`.
   Backups are automatic: `thrgamon/infra`'s pg-backup job dumps every Dokku Postgres service nightly, so no per-database schedule is needed.
   If the app holds bulk derived or re-importable data, exclude it in `backups/exclude-table-data.conf` in that repo (see its `docs/backups.md`).
3. Set production Auth0 configuration through Dokku secrets: `ENVIRONMENT=production`, `COOKIE_SECURE=true`, `AUTH0_ISSUER_URL`, `AUTH0_CLIENT_ID`, `AUTH0_CLIENT_SECRET`, `AUTH0_REDIRECT_URL`, and `AUTH_STATE_SECRET`. Do not set `COOKIE_DOMAIN`; session cookies are host-only.
4. Add the deploy job to `.github/workflows/ci.yml`. The template ships without one, because this repository is public and the shared delivery workflow in the private `thrgamon/infra` repository can only be called from private repositories owned by the same account:

   ```yaml
     deploy:
       if: github.event_name == 'push' && github.ref == 'refs/heads/main'
       needs: [migrations, lint, test, check, sync, image]
       uses: thrgamon/infra/.github/workflows/dokku-deploy.yml@main
       with:
         app_name: myapp
         healthcheck_url: https://myapp.tomgamon.xyz/api/health
       secrets:
         DOKKU_SSH_KEY: ${{ secrets.DOKKU_SSH_KEY }}
   ```

   Add the `DOKKU_SSH_KEY` repository secret, a key authorised on the Dokku server (see `thrgamon/infra` `docs/ci-deploy.md`). Every push to `main` that passes CI then deploys.

`just dokku-deploy` (with a `dokku` git remote) remains available for a manual first deploy or recovery.

Migrations run automatically on deploy via `app.json` predeploy hook. The
Docker build compiles SvelteKit output; Go serves it from `STATIC_DIR`.
