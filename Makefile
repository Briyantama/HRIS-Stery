# ============================================================
# HRIS-Stery — Developer Makefile
# ============================================================
# Run from monorepo root.
# Prerequisites: Docker, Go 1.24+, buf, golang-migrate
# Optional: golangci-lint, Node 20+ (apps/web), PHP 8.3+ (apps/api-gateway)

SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c

.PHONY: help check-tools require-proto-gen \
        dev dev-down dev-reset dev-logs dev-ps \
        proto-deps proto-gen proto-lint proto-breaking \
        migrate-up migrate-down migrate-status \
        test test-go test-integration test-laravel test-svelte \
        lint lint-go lint-laravel lint-svelte \
        build build-go build-laravel build-svelte \
        tidy vet clean clean-svelte setup \
        go-services

# ── Discovered paths (only *-service dirs; excludes _shared) ───────────────
GO_SERVICES := $(sort $(patsubst services/%/,%,$(dir $(wildcard services/*-service/go.mod))))
MIGRATE_SERVICES := $(sort $(patsubst services/%/,%,$(dir $(wildcard services/*-service/migrations))))

HAS_API_GATEWAY := $(wildcard apps/api-gateway/composer.json)
HAS_WEB := $(wildcard apps/web/package.json)

COMPOSE := docker compose -f deploy/docker/docker-compose.yml
DB_URL ?= postgres://hris_admin:hris_admin_secret@localhost:5432/hris_db?sslmode=disable

# ──────────────────────────────────────────────────────────────
# Help
# ──────────────────────────────────────────────────────────────
help:
	@echo ""
	@echo "HRIS-Stery Developer Commands"
	@echo "────────────────────────────────────"
	@echo "Go services (auto-detected): $(GO_SERVICES)"
	@echo "Migrate services:            $(MIGRATE_SERVICES)"
	@echo ""
	@echo "Dev environment:"
	@echo "  make dev              Start infrastructure (Docker Compose)"
	@echo "  make dev-down         Stop infrastructure"
	@echo "  make dev-reset        Stop and destroy volumes"
	@echo "  make dev-logs         Tail container logs"
	@echo "  make dev-ps           Show container status"
	@echo ""
	@echo "Proto:"
	@echo "  make proto-deps       Fetch buf module dependencies"
	@echo "  make proto-gen        buf generate → gen/"
	@echo "  make proto-lint       buf lint"
	@echo "  make proto-breaking   buf breaking vs main (requires git)"
	@echo ""
	@echo "Migrations (requires Postgres + migrate CLI):"
	@echo "  make migrate-up       Apply migrations for: $(MIGRATE_SERVICES)"
	@echo "  make migrate-down     Roll back one step per service"
	@echo "  make migrate-status   Show migration version per service"
	@echo ""
	@echo "Go:"
	@echo "  make tidy             go mod tidy for all Go services + gen/go"
	@echo "  make vet              go vet for all Go services"
	@echo "  make test-go          Unit tests (-short) for all Go services"
	@echo "  make test-integration Integration tests (INTEGRATION=true)"
	@echo "  make build-go         go build ./cmd/... for all Go services"
	@echo "  make lint-go          golangci-lint (skip if not installed)"
	@echo "  make format-go        go fmt for all Go services"
	@echo "  make update-go-mods   update go.mod for all Go services"
	@echo "  make all-go           run all go vet, fmt, and lint commands for all Go services"
	@echo ""
	@echo "Apps (skipped if directory missing):"
	@echo "  make test-laravel     Pest (apps/api-gateway)"
	@echo "  make test-svelte      Vitest (apps/web)"
	@echo "  make lint-laravel     Laravel Pint"
	@echo "  make lint-svelte      ESLint + svelte-check"
	@echo "  make build-svelte     SvelteKit production build"
	@echo ""
	@echo "Aggregates:"
	@echo "  make test             test-go + optional app tests"
	@echo "  make lint             lint-go + optional app lints"
	@echo "  make build            build-go + optional build-svelte"
	@echo "  make clean            go clean + SvelteKit artifacts"
	@echo "  make setup            dev + migrate-up"
	@echo "  make check-tools      Verify required CLI tools"
	@echo "  make go-services      Print discovered Go service names"
	@echo ""

go-services:
	@echo "$(GO_SERVICES)"

check-tools:
	@echo "Checking tools..."
	@command -v go >/dev/null 2>&1 || { echo "MISSING: go"; exit 1; }
	@command -v buf >/dev/null 2>&1 || { echo "MISSING: buf (https://buf.build/docs/installation)"; exit 1; }
	@command -v migrate >/dev/null 2>&1 || echo "WARN: migrate not found (needed for migrate-*)"
	@command -v docker >/dev/null 2>&1 || echo "WARN: docker not found (needed for dev)"
	@command -v golangci-lint >/dev/null 2>&1 || echo "WARN: golangci-lint not found (lint-go will skip)"
	@echo "OK: core tools present"

# ──────────────────────────────────────────────────────────────
# Dev environment
# ──────────────────────────────────────────────────────────────
dev:
	$(COMPOSE) up -d
	@echo ""
	@echo "Infrastructure running:"
	@echo "  PostgreSQL  → localhost:5432"
	@echo "  PgBouncer   → localhost:6432 (app connections)"
	@echo "  Redis       → localhost:6379 (password: hris_redis_secret)"
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

proto-breaking: proto-deps
	@if git rev-parse --git-dir >/dev/null 2>&1; then \
		cd proto && buf breaking --against ".git#branch=main,subdir=proto"; \
	else \
		echo "SKIP proto-breaking: not a git repository (clone with git to enable)"; \
	fi

# ──────────────────────────────────────────────────────────────
# Migrations — direct Postgres (5432); migrations use BYPASSRLS admin
# ──────────────────────────────────────────────────────────────
migrate-up:
	@if [ -z "$(MIGRATE_SERVICES)" ]; then \
		echo "No services with migrations/ found"; exit 1; \
	fi
	@command -v migrate >/dev/null 2>&1 || { echo "ERROR: install golang-migrate"; exit 1; }
	@for svc in $(MIGRATE_SERVICES); do \
		echo "→ Migrating $$svc..."; \
		migrate -path "services/$$svc/migrations" -database "$(DB_URL)" up || exit 1; \
	done
	@echo "All migrations applied."

migrate-down:
	@if [ -z "$(MIGRATE_SERVICES)" ]; then \
		echo "No services with migrations/ found"; exit 1; \
	fi
	@command -v migrate >/dev/null 2>&1 || { echo "ERROR: install golang-migrate"; exit 1; }
	@for svc in $(MIGRATE_SERVICES); do \
		echo "← Rolling back $$svc (1 step)..."; \
		migrate -path "services/$$svc/migrations" -database "$(DB_URL)" down 1 || exit 1; \
	done

migrate-status:
	@if [ -z "$(MIGRATE_SERVICES)" ]; then \
		echo "No services with migrations/ found"; exit 1; \
	fi
	@command -v migrate >/dev/null 2>&1 || { echo "ERROR: install golang-migrate"; exit 1; }
	@for svc in $(MIGRATE_SERVICES); do \
		echo "Status: $$svc"; \
		migrate -path "services/$$svc/migrations" -database "$(DB_URL)" version || true; \
	done

# ──────────────────────────────────────────────────────────────
# Go — loop helpers
# ──────────────────────────────────────────────────────────────
define go_foreach
	@if [ -z "$(GO_SERVICES)" ]; then \
		echo "No Go services found under services/*-service/"; exit 1; \
	fi
	@for svc in $(GO_SERVICES); do \
		echo "→ $$svc..."; \
		(cd "services/$$svc" && $(1)) || exit 1; \
	done
endef

tidy:
	@echo "Tidying gen/go..."
	@if [ -f gen/go/go.mod ]; then (cd gen/go && go mod tidy); fi
	$(call go_foreach,go mod tidy)
	@echo "Done."

vet:
	$(call go_foreach,go vet ./...)

update-go-mods:
	$(call go_foreach,go get -u ./...)
	$(call go_foreach,go mod tidy)
	$(call go_foreach,go mod verify)
	$(call go_foreach,go mod download)

format-go:
	$(call go_foreach,go fmt ./...)

# ── Generated code guard ─────────────────────────────────────
require-proto-gen:
	@if [ ! -f gen/go/go.mod ]; then \
		echo "ERROR: gen/go missing. Run: make proto-gen"; \
		exit 1; \
	fi

test-go: require-proto-gen
	$(call go_foreach,go test -short ./...)

build-go: require-proto-gen
	$(call go_foreach,go build ./cmd/...)

test-integration: require-proto-gen
	@for svc in $(GO_SERVICES); do \
		if [ -d "services/$$svc/internal/integration" ]; then \
			echo "→ Integration tests: $$svc..."; \
			(cd "services/$$svc" && INTEGRATION=true go test -run Integration -race -timeout 120s ./internal/integration/...) || exit 1; \
		else \
			echo "⊘ No integration tests: $$svc"; \
		fi; \
	done

lint-go:
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "SKIP lint-go: golangci-lint not installed (https://golangci-lint.run/welcome/install/)"; \
		exit 0; \
	fi
	$(call go_foreach,golangci-lint run --timeout=5m ./...)

all-go:
	$(call go_foreach,go vet ./...)
	$(call go_foreach,go fmt ./...)
	$(call go_foreach,golangci-lint run --timeout=5m ./...)

# ──────────────────────────────────────────────────────────────
# Apps (optional — skipped when directory missing)
# ──────────────────────────────────────────────────────────────
test-laravel:
ifneq ($(HAS_API_GATEWAY),)
	cd apps/api-gateway && ./vendor/bin/pest --parallel
else
	@echo "SKIP test-laravel: apps/api-gateway not present"
endif

test-svelte:
ifneq ($(HAS_WEB),)
	cd apps/web && npm run test:unit
else
	@echo "SKIP test-svelte: apps/web not present"
endif

lint-laravel:
ifneq ($(HAS_API_GATEWAY),)
	cd apps/api-gateway && ./vendor/bin/pint --test
else
	@echo "SKIP lint-laravel: apps/api-gateway not present"
endif

lint-svelte:
ifneq ($(HAS_WEB),)
	cd apps/web && npm run lint && npm run check
else
	@echo "SKIP lint-svelte: apps/web not present"
endif

build-laravel:
ifneq ($(HAS_API_GATEWAY),)
	@echo "SKIP build-laravel: no standard build target defined for API gateway"
else
	@echo "SKIP build-laravel: apps/api-gateway not present"
endif

build-svelte:
ifneq ($(HAS_WEB),)
	cd apps/web && npm run build
else
	@echo "SKIP build-svelte: apps/web not present"
endif

# ──────────────────────────────────────────────────────────────
# Aggregates
# ──────────────────────────────────────────────────────────────
test: test-go test-laravel test-svelte

lint: lint-go lint-laravel lint-svelte

build: build-go build-svelte

# ──────────────────────────────────────────────────────────────
# Cleanup
# ──────────────────────────────────────────────────────────────
clean:
	$(call go_foreach,go clean ./...)
	@$(MAKE) clean-svelte

clean-svelte:
ifneq ($(HAS_WEB),)
	cd apps/web && rm -rf .svelte-kit build
	@echo "Removed SvelteKit build artifacts (node_modules kept)"
else
	@echo "SKIP clean-svelte: apps/web not present"
endif

setup: dev migrate-up
	@echo "Dev environment ready. Run 'make dev-logs' to watch."
