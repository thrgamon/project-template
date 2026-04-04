# Project Name

## Build & Test
- `just test` to run Go tests, `just check` for all checks
- `just lint` for Go linters, `just fe-lint` for Biome
- `just sync` to regenerate all code after schema/API changes
- `just e2e` to run Playwright tests (requires dev server)
- Use `go vet` before pushing

## Architecture
- Go backend: cmd/server/ + internal/ (Gin, pgx, sqlc)
- Next.js frontend: src/ (App Router, React Query, shadcn/ui, Tailwind v4)
- PostgreSQL with sqlc for type-safe queries
- Swagger from Go annotations (swag), TypeScript client from Orval (React Query + Zod + MSW)
- Handler uses HandlerConfig struct for dependency injection
- OpenTelemetry opt-in: set OTEL_EXPORTER_OTLP_ENDPOINT to enable (traces, metrics, logs, Go runtime)
- `internal/telemetry/telemetry.go` initializes OTEL SDK; `otelgin` middleware on Gin router for automatic HTTP tracing

## Auth
- Session-based with HTTP-only cookies (no JWT)
- Auth service in internal/auth/, middleware reads session_token cookie
- Sessions stored in DB, cleaned up hourly

## Conventions
- Error wrapping: fmt.Errorf("context: %w", err)
- Domain types use validate:"required" tags for swag
- Generated code: internal/db/, docs/, src/lib/api/generated/ - DO NOT EDIT
- Run `just sync` after changing migrations/, queries/, or handler annotations
- Always create migrations with `just migrate-create <name>` (generates unique version). Never hand-create migration files.
- Use shadcn/ui components, not raw HTML for interactive elements
- shadcn skills are installed in .claude/skills/shadcn for AI-assisted component work
- Add components via `npx shadcn@latest add <component>`, search with `npx shadcn@latest search`
- Use semantic selectors in e2e tests (getByRole, getByText)
- Frontend types in src/lib/types.ts, schemas in src/lib/schemas.ts

## Deployment
- Dokku via Dockerfile.dokku (unified Go + Next.js container)
- Migrations run on deploy via app.json predeploy hook
- `just dokku-deploy` to push, `just dokku-logs` to tail
