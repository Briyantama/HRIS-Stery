#!/bin/bash

# ============================================================
# HRIS-Stery Database Health Check
# ============================================================
# Verifies database connectivity, schema state, and migration status

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Database configuration (from Makefile defaults)
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-hris_db}"
DB_USER="${DB_USER:-hris_admin}"
DB_PASSWORD="${DB_PASSWORD:-hris_admin_secret}"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ──────────────────────────────────────────────────────────────
# Helper functions
# ──────────────────────────────────────────────────────────────

check_connectivity() {
    echo -e "\n${BLUE}Database Connectivity${NC}"

    # Check if port is open
    if timeout 3 bash -c "</dev/tcp/$DB_HOST/$DB_PORT" 2>/dev/null; then
        echo -e "  ${GREEN}✓${NC} Database host reachable ($DB_HOST:$DB_PORT)"
    else
        echo -e "  ${RED}✗${NC} Cannot reach database ($DB_HOST:$DB_PORT)"
        return 1
    fi

    # Try to connect with psql
    if command -v psql &>/dev/null; then
        if PGPASSWORD="$DB_PASSWORD" timeout 5 psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1" &>/dev/null 2>&1; then
            echo -e "  ${GREEN}✓${NC} PostgreSQL connection successful"
        else
            echo -e "  ${YELLOW}⊘${NC} Cannot authenticate to PostgreSQL (psql available but failed)"
        fi
    else
        echo -e "  ${YELLOW}⊘${NC} psql not found (cannot fully test connection)"
    fi

    return 0
}

check_schema() {
    echo -e "\n${BLUE}Schema Status${NC}"

    if ! command -v psql &>/dev/null; then
        echo -e "  ${YELLOW}⊘${NC} psql not available (skipping schema check)"
        return 0
    fi

    # List all schemas
    local schemas=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT schema_name FROM information_schema.schemata
        WHERE schema_name NOT IN ('pg_catalog', 'information_schema')
        ORDER BY schema_name
    " 2>/dev/null || echo "")

    if [ -z "$schemas" ]; then
        echo -e "  ${YELLOW}⊘${NC} No application schemas found"
        return 0
    fi

    echo "  Schemas:"
    echo "$schemas" | while read -r schema; do
        [ -n "$schema" ] && echo -e "    ${GREEN}✓${NC} $schema"
    done

    return 0
}

check_migration_status() {
    echo -e "\n${BLUE}Migration Status${NC}"

    if ! command -v migrate &>/dev/null; then
        echo -e "  ${YELLOW}⊘${NC} golang-migrate not found (cannot check migration status)"
        return 0
    fi

    # Check each service
    for svc in auth attendance employee leave audit notification document; do
        local migration_dir="$PROJECT_ROOT/services/$svc-service/migrations"

        if [ ! -d "$migration_dir" ]; then
            continue
        fi

        if [ "$(find "$migration_dir" -name "*.up.sql" 2>/dev/null | wc -l)" -eq 0 ]; then
            continue
        fi

        local version=$(PGPASSWORD="$DB_PASSWORD" migrate -path "$migration_dir" -database "postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable" version 2>/dev/null || echo "0")

        if [ -n "$version" ] && [ "$version" != "0" ]; then
            echo -e "  ${GREEN}✓${NC} $svc: version $version"
        else
            echo -e "  ${YELLOW}⊘${NC} $svc: no migrations applied"
        fi
    done

    return 0
}

check_performance() {
    echo -e "\n${BLUE}Performance Metrics${NC}"

    if ! command -v psql &>/dev/null; then
        echo -e "  ${YELLOW}⊘${NC} psql not available (cannot measure performance)"
        return 0
    fi

    # Measure query latency
    local start=$(date +%s%N)
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1" &>/dev/null 2>&1
    local end=$(date +%s%N)
    local latency=$(( (end - start) / 1000000 ))  # Convert to ms

    echo -e "  ${GREEN}✓${NC} Query latency: ${latency}ms"

    # Check connection pool (if PgBouncer is running)
    if timeout 2 bash -c "</dev/tcp/localhost/6432" 2>/dev/null; then
        echo -e "  ${GREEN}✓${NC} PgBouncer (connection pool) is available on port 6432"
    else
        echo -e "  ${YELLOW}⊘${NC} PgBouncer not found on port 6432"
    fi

    return 0
}

wait_for_database() {
    echo -e "\n${BLUE}Waiting for database...${NC}"

    local max_attempts=30
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if timeout 3 bash -c "</dev/tcp/$DB_HOST/$DB_PORT" 2>/dev/null; then
            if command -v psql &>/dev/null; then
                if PGPASSWORD="$DB_PASSWORD" timeout 5 psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1" &>/dev/null 2>&1; then
                    echo -e "${GREEN}✓${NC} Database is ready"
                    return 0
                fi
            else
                echo -e "${GREEN}✓${NC} Database port is open"
                return 0
            fi
        fi

        echo -ne "."
        sleep 1
        ((attempt++))
    done

    echo -e "\n${RED}✗${NC} Database did not become ready (timeout after $((max_attempts))s)"
    return 1
}

# ──────────────────────────────────────────────────────────────
# Main
# ──────────────────────────────────────────────────────────────

main() {
    local command=${1:-check}

    case "$command" in
        check)
            echo -e "${BLUE}════════════════════════════════════════════════════════${NC}"
            echo -e "${BLUE}  Database Health Check${NC}"
            echo -e "${BLUE}════════════════════════════════════════════════════════${NC}"

            check_connectivity || true
            check_schema || true
            check_migration_status || true
            check_performance || true

            echo -e "\n${BLUE}════════════════════════════════════════════════════════${NC}\n"
            ;;
        wait)
            wait_for_database
            ;;
        *)
            echo "Database Health Check Utilities"
            echo ""
            echo "Usage: $0 <command>"
            echo ""
            echo "Commands:"
            echo "  check   Check database health (connectivity, schema, migrations)"
            echo "  wait    Wait for database to become ready (for startup scripts)"
            ;;
    esac
}

main "$@"
