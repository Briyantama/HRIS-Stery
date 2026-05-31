# ============================================================
# HRIS-Stery — Developer Makefile
# ============================================================
# All commands assume you are at the monorepo root.
# Prerequisites: Docker, Go 1.24+, Node 20+, PHP 8.3+, buf, golang-migrate

.PHONY: help dev dev-down dev-reset dev-logs \
        proto-deps proto-gen proto-lint proto-breaking \
        migrate-up migrate-down migrate-status \
        test test-go test-laravel test-svelte \
        lint lint-go lint-laravel lint-svelte \
        build build-go build-laravel build-svelte \
        clean

# ──────────────────────────────────────────────────────────────
# Help
# ──────────────────────────────────────────────────────────────
help:
	@echo ""
	@echo "HRIS-Stery Developer Commands"
	@echo "────────────────────────────────────"
	@echo "Dev environment:"
	@echo "  make dev          Start all infrastructure (Postgres, Redis, NATS, MinIO, Jaeger)"
	@echo "  make dev-down     Stop infrastructure"
	@echo "  make dev-reset    Stop + destroy all volumes (clean slate)"
	@echo "  make dev-logs     Tail all container logs"
	@echo ""
	@echo "Proto:"
	@echo "  make proto-deps     Fetch/update buf module dependencies (googleapis, etc.)"
	@echo "  make proto-gen      Run buf generate (rebuild all Go stubs)"
	@echo "  make proto-lint     Run buf lint"
	@echo "  make proto-breaking Check for breaking changes against main"
	@echo ""
	@echo "Migrations:"
	@echo "  make migrate-up     Run all pending migrations for all services"
	@echo "  make migrate-down   Roll back last migration for all services"
	@echo ""
	@echo "Testing:"
	@echo "  make test           Run all tests (Go unit + integration, Laravel, Svelte)"
	@echo "  make test-go        Go unit tests only (-short)"
	@echo "  make test-laravel   Laravel Pest tests"
	@echo "  make test-svelte    SvelteKit Vitest + Playwright"
	@echo ""
	@echo "Lint:"
	@echo "  make lint           Lint all (Go, PHP, TypeScript)"
	@echo "  make lint-go        GolangCI-Lint for all Go services"
	@echo "  make lint-laravel   Laravel Pint for API Gateway"
	@echo "  make lint-svelte    ESLint + Stylelint for SvelteKit"
	@echo "  make tidy           Run go mod tidy for all Go services"
	@echo "  make vet            Run go vet for all Go services"
	@echo ""

# ──────────────────────────────────────────────────────────────
# Dev environment
# ──────────────────────────────────────────────────────────────
COMPOSE = docker compose -f deploy/docker/docker-compose.yml

dev:
	$(COMPOSE) up -d
	@echo ""
	@echo "Infrastructure running:"
	@echo "  PostgreSQL  → localhost:5432"
	@echo "  PgBouncer   → localhost:6432 (use this for app connections)"
	@echo "  Redis       → localhost:6379"
	@echo "  NATS        → localhost:4222  | Monitor: http://localhost:8222"
	@echo "  MinIO       → http://localhost:9001 (minioadmin / minioadmin123)"
	@echo "  Jaeger      → http://localhost:16686"
	@echo "  Prometheus  → http://localhost:9090"
	@echo ""

dev-down:
	$(COMPOSE) down

dev-reset:
	$(COMPOSE) down -v
	@echo "All volumes destroyed."

dev-logs:
	$(COMPOSE) logs -f

dev-ps:
	$(COMPOSE) ps

# ──────────────────────────────────────────────────────────────
# Proto
# ──────────────────────────────────────────────────────────────
proto-deps:
	cd proto && buf dep update

proto-gen: proto-deps
	cd proto && buf generate
	@echo "Proto stubs regenerated in gen/"

proto-lint: proto-deps
	cd proto && buf lint

proto-breaking:
	cd proto && buf breaking --against ".git#branch=main,subdir=proto"

# ──────────────────────────────────────────────────────────────
# Migrations
# Uses PgBouncer on 6432 for consistency with app connections.
# hris_admin role has BYPASSRLS for migrations.
# ──────────────────────────────────────────────────────────────
DB_URL = postgres://hris_admin:hris_admin_secret@localhost:5432/hris_db?sslmode=disable

SERVICES = auth-service employee-service attendance-service leave-service \
           notification-service audit-service payroll-service ai-service

migrate-up:
	@for svc in $(SERVICES); do \
		echo "→ Migrating $$svc..."; \
		migrate -path services/$$svc/migrations -database "$(DB_URL)" up; \
	done
	@echo "All migrations applied."

migrate-down:
	@for svc in $(SERVICES); do \
		echo "← Rolling back $$svc (1 step)..."; \
		migrate -path services/$$svc/migrations -database "$(DB_URL)" down 1; \
	done

migrate-status:
	@for svc in $(SERVICES); do \
		echo "Status: $$svc"; \
		migrate -path services/$$svc/migrations -database "$(DB_URL)" version; \
	done

# ──────────────────────────────────────────────────────────────
# Tests
# ──────────────────────────────────────────────────────────────
test: test-go test-laravel test-svelte

test-go:
	@for svc in $(SERVICES) ai-service; do \
		echo "Testing $$svc..."; \
		cd services/$$svc && go test -short -race ./... && cd ../..; \
	done

test-laravel:
	cd apps/api-gateway && ./vendor/bin/pest --parallel

test-svelte:
	cd apps/web && npm run test:unit

# ──────────────────────────────────────────────────────────────
# Lint
# ──────────────────────────────────────────────────────────────
lint: lint-go lint-laravel lint-svelte

lint-go:
	@for svc in $(SERVICES) ai-service; do \
		echo "Linting $$svc..."; \
		cd services/$$svc && golangci-lint run ./... && cd ../..; \
	done

lint-laravel:
	cd apps/api-gateway && ./vendor/bin/pint --test

lint-svelte:
	cd apps/web && npm run lint && npm run check

# ──────────────────────────────────────────────────────────────
# Build (local, not Docker)
# ──────────────────────────────────────────────────────────────
build-go:
	@for svc in $(SERVICES) ai-service; do \
		echo "Building $$svc..."; \
		cd services/$$svc && go build ./cmd/... && cd ../..; \
	done

build-svelte:
	cd apps/web && npm run build

tidy:
	@echo "Tidying all Go services..."
	@for svc in $(SERVICES) ai-service; do \
		echo "Tidying $$svc..."; \
		cd services/$$svc && go mod tidy && cd ../..; \
	done

vet:
	@echo "Running go vet for all Go services..."
	@for svc in $(SERVICES) ai-service; do \
		echo "Running go vet for $$svc..."; \
		cd services/$$svc && go vet ./... && cd ../..; \
	done

# ──────────────────────────────────────────────────────────────
# Cleanup
# ──────────────────────────────────────────────────────────────
clean:
	@echo "Cleaning all Go services..."
	@for svc in $(SERVICES) ai-service; do \
		echo "Cleaning $$svc..."; \
		cd services/$$svc && go clean ./... && cd ../..; \
	done
	@echo "Cleaning SvelteKit..."
	cd apps/web && rm -rf .svelte-kit build node_modules
	@echo "Cleaned."

# ──────────────────────────────────────────────────────────────
# Convenience
# ──────────────────────────────────────────────────────────────
setup: dev migrate-up
	@echo "Dev environment ready. Run 'make dev-logs' to watch."
