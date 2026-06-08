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
        test test-go test-integration test-laravel test-svelte test-e2e \
        lint lint-go lint-laravel lint-svelte \
        build build-go build-laravel build-svelte \
        install-laravel reinstall-laravel install-svelte install-apps \
        dev-laravel dev-svelte dev-apps \
        format-laravel format-svelte \
        all-go all-laravel all-svelte all-apps \
        require-laravel-vendor require-svelte-modules \
        tidy vet clean clean-go clean-laravel clean-svelte setup setup-apps \
        go-services

# ── Discovered paths (only *-service dirs; excludes _shared) ───────────────
GO_SERVICES := $(sort $(patsubst services/%/,%,$(dir $(wildcard services/*-service/go.mod))))
MIGRATE_SERVICES := $(sort $(patsubst services/%/,%,$(dir $(wildcard services/*-service/migrations))))

API_GATEWAY_DIR := apps/api-gateway
WEB_DIR         := apps/web

HAS_API_GATEWAY := $(wildcard $(API_GATEWAY_DIR)/composer.json)
HAS_WEB         := $(wildcard $(WEB_DIR)/package.json)

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
	@echo "Laravel API Gateway ($(API_GATEWAY_DIR), skipped if composer.json missing):"
	@echo "  make install-laravel  composer install (Windows: disables 7zip extractor)"
	@echo "  make reinstall-laravel  remove vendor/ and reinstall from composer.lock"
	@echo "  make dev-laravel      php artisan serve (port 8000)"
	@echo "  make test-laravel     Pest unit/feature tests"
	@echo "  make lint-laravel     Laravel Pint (--test)"
	@echo "  make format-laravel   Laravel Pint (auto-fix)"
	@echo "  make build-laravel    composer install + optimize autoloader"
	@echo "  make all-laravel      lint + format + test"
	@echo "  make clean-laravel    clear Laravel caches"
	@echo ""
	@echo "SvelteKit Frontend ($(WEB_DIR), skipped if package.json missing):"
	@echo "  make install-svelte   npm ci (or npm install)"
	@echo "  make dev-svelte       Vite dev server"
	@echo "  make preview-svelte   production preview (port 4173)"
	@echo "  make test-svelte      Vitest unit tests"
	@echo "  make test-e2e         Playwright E2E tests"
	@echo "  make lint-svelte      ESLint + svelte-check"
	@echo "  make format-svelte    Prettier (if configured)"
	@echo "  make build-svelte     SvelteKit production build"
	@echo "  make all-svelte       lint + test + build"
	@echo "  make clean-svelte     remove .svelte-kit and build/"
	@echo ""
	@echo "Apps (aggregates):"
	@echo "  make install-apps     install-laravel + install-svelte"
	@echo "  make dev-apps         print dev server instructions"
	@echo "  make all-apps         all-laravel + all-svelte"
	@echo ""
	@echo "Aggregates:"
	@echo "  make test             test-go + laravel + svelte"
	@echo "  make lint             lint-go + laravel + svelte"
	@echo "  make build            build-go + laravel + svelte"
	@echo "  make clean            go clean + laravel + svelte artifacts"
	@echo "  make setup            dev + migrate-up"
	@echo "  make setup-apps       install-apps (after scaffold is ready)"
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
	@command -v php >/dev/null 2>&1 || echo "WARN: php not found (needed for Laravel)"
	@command -v composer >/dev/null 2>&1 || echo "WARN: composer not found (needed for Laravel)"
	@command -v node >/dev/null 2>&1 || echo "WARN: node not found (needed for SvelteKit)"
	@command -v npm >/dev/null 2>&1 || echo "WARN: npm not found (needed for SvelteKit)"
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
# Laravel API Gateway (optional — skipped when composer.json missing)
# ──────────────────────────────────────────────────────────────
install-laravel:
ifneq ($(HAS_API_GATEWAY),)
	cd $(API_GATEWAY_DIR) && COMPOSER_DISABLE_7ZIP=1 composer install --no-progress --prefer-dist --optimize-autoloader
else
	@echo "SKIP install-laravel: $(API_GATEWAY_DIR)/composer.json not present"
endif

reinstall-laravel:
ifneq ($(HAS_API_GATEWAY),)
	rm -rf $(API_GATEWAY_DIR)/vendor
	cd $(API_GATEWAY_DIR) && COMPOSER_DISABLE_7ZIP=1 composer install --no-progress --prefer-dist --optimize-autoloader
else
	@echo "SKIP reinstall-laravel: $(API_GATEWAY_DIR)/composer.json not present"
endif

dev-laravel:
ifneq ($(HAS_API_GATEWAY),)
	cd $(API_GATEWAY_DIR) && php artisan serve --host=0.0.0.0 --port=8000
else
	@echo "SKIP dev-laravel: $(API_GATEWAY_DIR)/composer.json not present"
endif

test-laravel:
ifneq ($(HAS_API_GATEWAY),)
	cd $(API_GATEWAY_DIR) && ./vendor/bin/pest --parallel
else
	@echo "SKIP test-laravel: $(API_GATEWAY_DIR)/composer.json not present"
endif

lint-laravel:
ifneq ($(HAS_API_GATEWAY),)
	cd $(API_GATEWAY_DIR) && ./vendor/bin/pint --test
else
	@echo "SKIP lint-laravel: $(API_GATEWAY_DIR)/composer.json not present"
endif

format-laravel:
ifneq ($(HAS_API_GATEWAY),)
	cd $(API_GATEWAY_DIR) && ./vendor/bin/pint
else
	@echo "SKIP format-laravel: $(API_GATEWAY_DIR)/composer.json not present"
endif

build-laravel:
ifneq ($(HAS_API_GATEWAY),)
	cd $(API_GATEWAY_DIR) && composer install --no-progress --prefer-dist --optimize-autoloader
else
	@echo "SKIP build-laravel: $(API_GATEWAY_DIR)/composer.json not present"
endif

clean-laravel:
ifneq ($(HAS_API_GATEWAY),)
	cd $(API_GATEWAY_DIR) && php artisan cache:clear 2>/dev/null || true
	cd $(API_GATEWAY_DIR) && php artisan config:clear 2>/dev/null || true
	cd $(API_GATEWAY_DIR) && php artisan route:clear 2>/dev/null || true
	cd $(API_GATEWAY_DIR) && php artisan view:clear 2>/dev/null || true
	@echo "Laravel caches cleared (vendor/ kept)"
else
	@echo "SKIP clean-laravel: $(API_GATEWAY_DIR)/composer.json not present"
endif

require-laravel-vendor:
ifneq ($(HAS_API_GATEWAY),)
	@if [ ! -d "$(API_GATEWAY_DIR)/vendor" ]; then \
		echo "ERROR: $(API_GATEWAY_DIR)/vendor missing. Run: make install-laravel"; \
		exit 1; \
	fi
endif

all-laravel: require-laravel-vendor lint-laravel format-laravel test-laravel

# ──────────────────────────────────────────────────────────────
# SvelteKit Frontend (optional — skipped when package.json missing)
# ──────────────────────────────────────────────────────────────
install-svelte:
ifneq ($(HAS_WEB),)
	@for attempt in 1 2 3; do \
		echo "→ npm install (attempt $$attempt/3)..."; \
		(cd $(WEB_DIR) && npm install) && exit 0; \
		[ $$attempt -lt 3 ] && echo "npm failed (ECONNRESET/network?), retrying in 5s..." && sleep 5; \
	done; \
	echo "ERROR: npm install failed after 3 attempts. Check network/proxy or run: cd $(WEB_DIR) && npm install"; \
	exit 1
else
	@echo "SKIP install-svelte: $(WEB_DIR)/package.json not present"
endif

dev-svelte:
ifneq ($(HAS_WEB),)
	cd $(WEB_DIR) && npm run dev
else
	@echo "SKIP dev-svelte: $(WEB_DIR)/package.json not present"
endif

preview-svelte:
ifneq ($(HAS_WEB),)
	cd $(WEB_DIR) && npm run preview -- --port 4173 --host
else
	@echo "SKIP preview-svelte: $(WEB_DIR)/package.json not present"
endif

test-svelte:
ifneq ($(HAS_WEB),)
	cd $(WEB_DIR) && npm run test:unit
else
	@echo "SKIP test-svelte: $(WEB_DIR)/package.json not present"
endif

test-e2e:
ifneq ($(HAS_WEB),)
	cd $(WEB_DIR) && npx playwright install --with-deps chromium
	cd $(WEB_DIR) && npm run build
	cd $(WEB_DIR) && PLAYWRIGHT_BASE_URL=http://localhost:4173 npm run test:e2e
else
	@echo "SKIP test-e2e: $(WEB_DIR)/package.json not present"
endif

lint-svelte:
ifneq ($(HAS_WEB),)
	cd $(WEB_DIR) && npm run lint && npm run check
else
	@echo "SKIP lint-svelte: $(WEB_DIR)/package.json not present"
endif

format-svelte:
ifneq ($(HAS_WEB),)
	@if cd $(WEB_DIR) && npm run | grep -q 'format'; then \
		cd $(WEB_DIR) && npm run format; \
	else \
		echo "SKIP format-svelte: no npm run format script in $(WEB_DIR)/package.json"; \
	fi
else
	@echo "SKIP format-svelte: $(WEB_DIR)/package.json not present"
endif

build-svelte:
ifneq ($(HAS_WEB),)
	cd $(WEB_DIR) && npm run build
else
	@echo "SKIP build-svelte: $(WEB_DIR)/package.json not present"
endif

clean-svelte:
ifneq ($(HAS_WEB),)
	cd $(WEB_DIR) && rm -rf .svelte-kit build
	@echo "Removed SvelteKit build artifacts (node_modules kept)"
else
	@echo "SKIP clean-svelte: $(WEB_DIR)/package.json not present"
endif

require-svelte-modules:
ifneq ($(HAS_WEB),)
	@if [ ! -d "$(WEB_DIR)/node_modules" ]; then \
		echo "ERROR: $(WEB_DIR)/node_modules missing. Run: make install-svelte"; \
		exit 1; \
	fi
endif

all-svelte: require-svelte-modules lint-svelte test-svelte build-svelte

# ──────────────────────────────────────────────────────────────
# Apps — combined helpers
# ──────────────────────────────────────────────────────────────
install-apps: install-laravel install-svelte

dev-apps:
	@echo "Start app dev servers in separate terminals:"
	@echo "  make dev-laravel   → http://localhost:8000"
	@echo "  make dev-svelte    → http://localhost:5173 (default Vite port)"

all-apps: all-laravel all-svelte all-go
	@echo "All app checks passed."

setup-apps: install-apps
	@echo "App dependencies installed. Run 'make dev-apps' for next steps."

# ──────────────────────────────────────────────────────────────
# Aggregates
# ──────────────────────────────────────────────────────────────
test: test-go test-laravel test-svelte

lint: lint-go lint-laravel lint-svelte

build: build-go build-laravel build-svelte

# ──────────────────────────────────────────────────────────────
# Cleanup
# ──────────────────────────────────────────────────────────────
clean-go:
	$(call go_foreach,go clean ./...)

clean: clean-go clean-laravel clean-svelte

setup: dev migrate-up
	@echo "Dev environment ready. Run 'make dev-logs' to watch."
