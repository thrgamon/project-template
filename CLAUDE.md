# Project Name

## Build & Test
- `just test` to run Go tests, `just check` for all checks
- `just lint` for Go linters, `just fe-lint` for Biome
- `just sync` to regenerate sqlc output and check the SvelteKit app
- `just e2e` to run Playwright tests (requires dev server)
- Use `go vet` before pushing

## Concurrent Worktrees
- Always start local services with `just dev`, `just dev-monitoring`, or `./scripts/compose`; do not use plain `docker compose`.
- The first Compose command creates a gitignored `.worktree.env` containing a stable, unique Compose project name, ports, PostgreSQL database, and isolated volumes for this worktree.
- Read `.worktree.env` to find the frontend/backend URLs. `just db-reset` destroys only this worktree's database and fixture state.

## Architecture
- Go backend: cmd/server/ + internal/ (Gin, pgx, sqlc)
- SvelteKit frontend: frontend/ (adapter-static, Vite)
- PostgreSQL with sqlc for type-safe queries
- Production is static-first: Go serves frontend/build and `/api` from one origin
- SSR is an explicit exception: document the need, use adapter-node, and add end-to-end coverage before introducing a Node runtime
- Handler uses HandlerConfig struct for dependency injection
- OpenTelemetry opt-in: set OTEL_EXPORTER_OTLP_ENDPOINT to enable (traces, metrics, logs, Go runtime)
- `internal/telemetry/telemetry.go` initializes OTEL SDK; `otelgin` middleware on Gin router for automatic HTTP tracing

## Auth
- Auth0 identity with opaque, hashed server-side sessions (no JWT in the app)
- Shared module lives at `github.com/thrgamon/infra/go/auth`; it is vendored
  and framework adapters live in `internal/auth/`
- Membership is provisioned explicitly by issuer and subject; never add public
  password signup, email linking, or first-user ownership

## Conventions
- Error wrapping: fmt.Errorf("context: %w", err)
- Keep browser DTOs in frontend/src/lib/api.ts small and hand-written. Validate owned API changes with endpoint tests.
- Handlers return domain types, never gin.H. Errors return domain.ErrorResponse
- Read the authenticated user with auth.GetUser(c), never c.Get("user_id")
- Generated code: internal/db/ (sqlc) - DO NOT EDIT; it is committed
- Run `just sync` after changing migrations/ or queries/
- Always create migrations with `just migrate-create <name>` (generates unique version). Never hand-create migration files.
- Use semantic selectors in e2e tests (getByRole, getByText)
- Frontend calls the API through frontend/src/lib/api.ts; do not scatter fetch calls through routes
- Use yarn for package management, never npm (`npx` is fine for one-off tools like shadcn)

## Deployment
- Dokku via Dockerfile.dokku (unified Go + static SvelteKit container)
- Migrations run on deploy via app.json predeploy hook
- `just dokku-deploy` to push, `just dokku-logs` to tail
