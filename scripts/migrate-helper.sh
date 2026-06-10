#!/bin/bash

# ============================================================
# HRIS-Stery Migration Helper Utilities
# ============================================================
# Provides utilities for migration management:
# - create: Create a new migration pair
# - list: Show all migrations
# - verify: Verify migration integrity
# - status: Show migration status per service
# - rollback-check: Verify rollback safety

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ──────────────────────────────────────────────────────────────
# Helper functions
# ──────────────────────────────────────────────────────────────

create_migration() {
    local service=$1
    local description=$2

    if [ -z "$service" ] || [ -z "$description" ]; then
        echo -e "${RED}Usage: create <service> <description>${NC}"
        echo "Example: create auth add_user_index"
        return 1
    fi

    local service_dir="$PROJECT_ROOT/services/$service-service"
    local migration_dir="$service_dir/migrations"

    if [ ! -d "$service_dir" ]; then
        echo -e "${RED}Service not found: $service-service${NC}"
        return 1
    fi

    if [ ! -d "$migration_dir" ]; then
        mkdir -p "$migration_dir"
    fi

    # Find next version number
    local next_version=1
    if [ "$(ls -1 "$migration_dir" 2>/dev/null | wc -l)" -gt 0 ]; then
        next_version=$(ls -1 "$migration_dir" | grep -oE '^[0-9]+' | sort -n | tail -1)
        ((next_version++))
    fi

    # Format version with leading zeros
    local version=$(printf "%03d" "$next_version")

    # Create migration files
    local up_file="$migration_dir/${version}_${description}.up.sql"
    local down_file="$migration_dir/${version}_${description}.down.sql"

    # Create up migration template
    cat > "$up_file" <<'EOF'
-- Migration: <DESCRIPTION>
-- Direction: up
-- Created: <TIMESTAMP>
--
-- This migration makes the following changes:
-- TODO: Add description
--

BEGIN;

-- TODO: Add migration SQL

COMMIT;
EOF

    # Create down migration template
    cat > "$down_file" <<'EOF'
-- Migration: <DESCRIPTION>
-- Direction: down
-- Created: <TIMESTAMP>
--
-- This migration reverses the changes from the up migration.
-- IMPORTANT: Ensure this migration can safely rollback the corresponding up migration.
--

BEGIN;

-- TODO: Add rollback SQL

COMMIT;
EOF

    # Update placeholders
    local timestamp=$(date -u +'%Y-%m-%d %H:%M:%S UTC')
    sed -i "s/<DESCRIPTION>/$description/g" "$up_file" "$down_file"
    sed -i "s/<TIMESTAMP>/$timestamp/g" "$up_file" "$down_file"

    echo -e "${GREEN}✓${NC} Migration created: $service-service"
    echo "  Up:   $up_file"
    echo "  Down: $down_file"
    echo ""
    echo "Next steps:"
    echo "  1. Edit the migration files to add your SQL"
    echo "  2. Run: make migrate-verify"
    echo "  3. Run: make migrate-up"
    echo "  4. Test: make migrate-down && make migrate-up"
}

list_migrations() {
    echo -e "\n${BLUE}════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  All Migrations${NC}"
    echo -e "${BLUE}════════════════════════════════════════════════════════${NC}\n"

    for svc in auth attendance employee leave audit notification document; do
        local migration_dir="$PROJECT_ROOT/services/$svc-service/migrations"

        if [ ! -d "$migration_dir" ]; then
            continue
        fi

        local count=$(find "$migration_dir" -name "*.up.sql" 2>/dev/null | wc -l)
        if [ "$count" -eq 0 ]; then
            echo -e "${YELLOW}$svc-service:${NC} no migrations"
            continue
        fi

        echo -e "${BLUE}$svc-service:${NC}"
        find "$migration_dir" -name "*.up.sql" | sort | while read -r file; do
            local name=$(basename "$file" .up.sql)
            local version=$(echo "$name" | grep -oE '^[0-9]+')
            local desc=$(echo "$name" | sed "s/^${version}_//")
            echo "  $version  $desc"
        done
        echo ""
    done

    echo -e "${BLUE}════════════════════════════════════════════════════════${NC}\n"
}

verify_migrations() {
    echo -e "\n${BLUE}════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  Migration Verification${NC}"
    echo -e "${BLUE}════════════════════════════════════════════════════════${NC}\n"

    local total_pass=0
    local total_fail=0

    for svc in auth attendance employee leave audit notification document; do
        local migration_dir="$PROJECT_ROOT/services/$svc-service/migrations"

        if [ ! -d "$migration_dir" ]; then
            continue
        fi

        echo -e "${BLUE}$svc-service:${NC}"

        local has_migrations=0
        for up_file in "$migration_dir"/*.up.sql; do
            [ -f "$up_file" ] || continue
            has_migrations=1

            local name=$(basename "$up_file" .up.sql)
            local down_file="${up_file%.up.sql}.down.sql"

            if [ -f "$down_file" ]; then
                echo -e "  ${GREEN}✓${NC} $name (up ↔ down)"
                ((total_pass++))
            else
                echo -e "  ${RED}✗${NC} $name (missing down)"
                ((total_fail++))
            fi
        done

        if [ "$has_migrations" -eq 0 ]; then
            echo "  No migrations found"
        fi

        echo ""
    done

    echo -e "${BLUE}════════════════════════════════════════════════════════${NC}"
    echo "Summary:"
    echo -e "  ${GREEN}Valid:${NC}   $total_pass"
    echo -e "  ${RED}Failed:${NC}  $total_fail"
    echo -e "${BLUE}════════════════════════════════════════════════════════${NC}\n"

    if [ "$total_fail" -gt 0 ]; then
        return 1
    fi
    return 0
}

rollback_check() {
    echo -e "\n${BLUE}════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  Rollback Safety Check${NC}"
    echo -e "${BLUE}════════════════════════════════════════════════════════${NC}\n"

    for svc in auth attendance employee leave audit notification document; do
        local migration_dir="$PROJECT_ROOT/services/$svc-service/migrations"

        if [ ! -d "$migration_dir" ]; then
            continue
        fi

        local down_files=$(find "$migration_dir" -name "*.down.sql" | sort)
        if [ -z "$down_files" ]; then
            continue
        fi

        echo -e "${BLUE}$svc-service:${NC}"

        while IFS= read -r down_file; do
            local name=$(basename "$down_file" .down.sql)

            # Check if down migration has content (more than just comments)
            local content=$(grep -v "^--" "$down_file" | grep -v "^$" | wc -l)

            if [ "$content" -lt 3 ]; then
                echo -e "  ${YELLOW}⊘${NC} $name (has minimal content - check if reversible)"
            else
                echo -e "  ${GREEN}✓${NC} $name (appears reversible)"
            fi
        done < <(echo "$down_files")

        echo ""
    done

    echo -e "${BLUE}════════════════════════════════════════════════════════${NC}\n"
}

# ──────────────────────────────────────────────────────────────
# Main
# ──────────────────────────────────────────────────────────────

main() {
    local command=${1:-help}

    case "$command" in
        create)
            create_migration "$2" "$3"
            ;;
        list)
            list_migrations
            ;;
        verify)
            verify_migrations
            ;;
        rollback-check)
            rollback_check
            ;;
        help|*)
            echo "Migration Helper Utilities"
            echo ""
            echo "Usage: $0 <command> [args]"
            echo ""
            echo "Commands:"
            echo "  create <service> <desc>   Create a new migration pair"
            echo "                            Example: create auth add_user_index"
            echo ""
            echo "  list                      List all migrations"
            echo ""
            echo "  verify                    Verify migration integrity"
            echo ""
            echo "  rollback-check            Check rollback safety"
            echo ""
            echo "  help                      Show this help message"
            ;;
    esac
}

main "$@"
