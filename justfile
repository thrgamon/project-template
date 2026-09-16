# Pinned versions of the external tools the codegen and migration steps need.
# tygo is not listed: it is pinned in go.mod and run via `go tool tygo`.
SQLC_VERSION := "v1.31.1"
GOOSE_VERSION := "v3.27.3"

default:
    @just --list

# Install the pinned external tools into $GOBIN
install-tools:
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@{{SQLC_VERSION}}
    go install github.com/pressly/goose/v3/cmd/goose@{{GOOSE_VERSION}}

# --- Development ---

# Start all services (postgres, backend, frontend)
dev:
    ./scripts/compose up --build

# Stop all services
dev-down:
    ./scripts/compose down

# View backend logs
logs:
    ./scripts/compose logs -f backend

# Start with monitoring stack
dev-monitoring:
    ./scripts/compose --profile monitoring up --build

# Run Go backend locally (outside docker)
backend:
    ./scripts/worktree-env && . ./.worktree.env && DATABASE_URL="postgres://postgres:postgres@localhost:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" PORT="${BACKEND_PORT}" go run ./cmd/server

# Run frontend locally (outside docker)
frontend:
    ./scripts/worktree-env && . ./.worktree.env && yarn install && API_URL="http://localhost:${BACKEND_PORT}" yarn run dev -- -p "${FRONTEND_PORT}"

# Install git hooks
install-hooks:
    cp scripts/pre-push .git/hooks/pre-push
    chmod +x .git/hooks/pre-push

# --- Code Generation ---

# Regenerate sqlc Go types from SQL queries
sqlc:
    sqlc generate

# Regenerate TypeScript API types from internal/domain
tygo:
    go tool tygo generate

# Full sync: sqlc + tygo + type check
sync: sqlc tygo
    yarn run check
    go vet ./...

# --- Quality ---

# Run Go linters
lint:
    golangci-lint run ./...
    go mod tidy

# Run all tests
test:
    go test -race ./...

# Format Go code (gofmt + goimports, configured in .golangci.yml)
fmt:
    golangci-lint fmt ./...

# Frontend lint
fe-lint:
    yarn run lint

# Frontend lint with auto-fix
fe-lint-fix:
    yarn run lint:fix

# Frontend format
fe-fmt:
    yarn run format

# Run all checks (lint + test + type-check)
check: lint test
    yarn run check
    yarn run lint

# --- E2E Tests ---

# Run Playwright e2e tests against this worktree's frontend service
e2e:
    ./scripts/worktree-env && . ./.worktree.env && BASE_URL="http://localhost:${FRONTEND_PORT}" yarn playwright test

# Run Playwright with UI mode
e2e-ui:
    ./scripts/worktree-env && . ./.worktree.env && BASE_URL="http://localhost:${FRONTEND_PORT}" yarn playwright test --ui

# --- Database ---

# Run migrations up
migrate:
    ./scripts/worktree-env && . ./.worktree.env && goose -dir migrations postgres "postgres://postgres:postgres@localhost:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" up

# Roll back one migration
migrate-down:
    ./scripts/worktree-env && . ./.worktree.env && goose -dir migrations postgres "postgres://postgres:postgres@localhost:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" down

# Reset all migrations
migrate-reset:
    ./scripts/worktree-env && . ./.worktree.env && goose -dir migrations postgres "postgres://postgres:postgres@localhost:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" reset

# Show migration status
migrate-status:
    ./scripts/worktree-env && . ./.worktree.env && goose -dir migrations postgres "postgres://postgres:postgres@localhost:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" status

# Create a new migration
migrate-create NAME:
    goose -dir migrations create {{NAME}} sql

# Reset database (destroy volume and recreate)
db-reset:
    ./scripts/compose down -v
    ./scripts/compose up -d postgres
    @echo "Waiting for postgres..."
    @sleep 3
    just migrate

# --- Build ---

# Build production Docker image
build:
    docker build -t myapp -f Dockerfile .

# Build Dokku Docker image
dokku-build:
    docker build -t myapp-dokku -f Dockerfile.dokku .

# Clean build artifacts
clean:
    rm -rf bin/ tmp/ coverage.out docs/
    rm -rf .next build node_modules/.cache test-results playwright-report

# --- Dokku Deployment ---

_dokku-host:
    @echo "${DOKKU_HOST}"

# Deploy to Dokku
dokku-deploy:
    git push dokku main

# View Dokku app logs
dokku-logs:
    ssh dokku@$(just _dokku-host) logs myapp -t

# View Dokku app config
dokku-config:
    ssh dokku@$(just _dokku-host) config:show myapp

# View Dokku app process status
dokku-ps:
    ssh dokku@$(just _dokku-host) ps:report myapp

# Connect to Dokku database
dokku-db-connect:
    ssh dokku@$(just _dokku-host) postgres:connect myapp-db

# Backup Dokku database locally
dokku-db-backup:
    ssh dokku@$(just _dokku-host) postgres:export myapp-db > myapp-db-backup.sql

# --- Monitoring ---

# Start monitoring stack
monitoring-up:
    ./scripts/compose -f monitoring/docker-compose.yml up -d

# Stop monitoring stack
monitoring-down:
    ./scripts/compose -f monitoring/docker-compose.yml down

# View monitoring logs
monitoring-logs:
    ./scripts/compose -f monitoring/docker-compose.yml logs -f

# Restart monitoring stack
monitoring-restart:
    ./scripts/compose -f monitoring/docker-compose.yml restart
