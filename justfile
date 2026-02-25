default:
    @just --list

# --- Development ---

# Start all services (postgres, backend, frontend)
dev:
    docker compose up --build

# Start with file watching (auto-rebuild on changes)
dev-watch:
    docker compose watch

# Stop all services
dev-down:
    docker compose down

# View backend logs
logs:
    docker compose logs -f backend

# --- Code Generation ---

# Regenerate sqlc Go types from SQL queries
sqlc:
    sqlc generate

# Regenerate swagger.json from Go annotations
api-docs:
    swag init -g cmd/server/main.go -o docs --parseInternal

# Regenerate frontend TypeScript client from swagger
api-types:
    npx orval

# Full sync: sqlc + swagger + orval + type check
sync: sqlc api-docs api-types
    npm run type-check
    go vet ./...

# --- Quality ---

lint:
    golangci-lint run --fix
    go mod tidy

test:
    go test -race -coverprofile=coverage.out ./...
    npm test

fmt:
    gofmt -s -w .

check: lint test
    npm run type-check

# --- Database ---

migrate-up:
    goose -dir migrations postgres "${DATABASE_URL}" up

migrate-down:
    goose -dir migrations postgres "${DATABASE_URL}" down

migrate-create NAME:
    goose -dir migrations create {{NAME}} sql

migrate-status:
    goose -dir migrations postgres "${DATABASE_URL}" status

# --- Build ---

build:
    docker build -t myapp-backend -f Dockerfile .
    docker build -t myapp-frontend -f Dockerfile.frontend .

clean:
    rm -rf bin/ tmp/ coverage.out docs/
    rm -rf .svelte-kit build node_modules/.cache
