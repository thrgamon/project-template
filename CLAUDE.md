# Project Name

## Build & Test
- `just test` to run Go tests, `just check` for all checks
- `just lint` for Go linters, `just fe-lint` for Biome
- `just sync` to regenerate sqlc + tygo output after schema/API changes
- `just e2e` to run Playwright tests (requires dev server)
- Use `go vet` before pushing

## Architecture
- Go backend: cmd/server/ + internal/ (Gin, pgx, sqlc)
- Next.js frontend: src/ (App Router, React Query, shadcn/ui, Tailwind v4)
- PostgreSQL with sqlc for type-safe queries
- `internal/domain` is the single source of truth for the API shape; tygo generates `src/lib/api/types.ts` from it
- Handler uses HandlerConfig struct for dependency injection
- OpenTelemetry opt-in: set OTEL_EXPORTER_OTLP_ENDPOINT to enable (traces, metrics, logs, Go runtime)
- `internal/telemetry/telemetry.go` initializes OTEL SDK; `otelgin` middleware on Gin router for automatic HTTP tracing

## Auth
- Session-based with HTTP-only cookies (no JWT)
- Auth service in internal/auth/, middleware reads session_token cookie
- Sessions stored in DB, cleaned up hourly

## Conventions
- Error wrapping: fmt.Errorf("context: %w", err)
- Add or change an API type in internal/domain, then run `just sync`. Never hand-edit src/lib/api/types.ts
- Handlers return domain types, never gin.H. Errors return domain.ErrorResponse
- Read the authenticated user with auth.GetUser(c), never c.Get("user_id")
- Generated code: internal/db/ (sqlc), src/lib/api/types.ts (tygo) - DO NOT EDIT, and both are committed
- Run `just sync` after changing migrations/, queries/, or internal/domain/
- Always create migrations with `just migrate-create <name>` (generates unique version). Never hand-create migration files.
- Use shadcn/ui components, not raw HTML for interactive elements
- shadcn skills are installed in .claude/skills/shadcn for AI-assisted component work
- Add components via `npx shadcn@latest add <component>`, search with `npx shadcn@latest search`
- Use semantic selectors in e2e tests (getByRole, getByText)
- Frontend calls the API through src/lib/api/hooks.ts (React Query) or src/lib/api/client.ts. Do not call fetch directly
- Use yarn for package management, never npm (`npx` is fine for one-off tools like shadcn)

## Deployment
- Dokku via Dockerfile.dokku (unified Go + Next.js container)
- Migrations run on deploy via app.json predeploy hook
- `just dokku-deploy` to push, `just dokku-logs` to tail
