#!/bin/bash
set -e

# ============================================================
# HRIS-Stery Migration Validation
# ============================================================
# Validates migration state for all services
# Checks: migration files exist, down migrations exist, migrations are valid

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
SERVICES_OK=0
SERVICES_FAILED=0

# ──────────────────────────────────────────────────────────────
# Helper functions
# ──────────────────────────────────────────────────────────────

validate_service_migrations() {
    local service=$1
    local migration_dir="$PROJECT_ROOT/services/$service/migrations"

    echo -ne "${BLUE}Validating $service...${NC} "

    if [ ! -d "$migration_dir" ]; then
        echo -e "${YELLOW}no migrations directory${NC}"
        return 0
    fi

    local up_migrations=()
    local down_migrations=()

    # Find all .up.sql files
    while IFS= read -r file; do
        local filename=$(basename "$file")
        up_migrations+=("$filename")
    done < <(find "$migration_dir" -name "*.up.sql" -type f | sort)

    # Find all .down.sql files
    while IFS= read -r file; do
        local filename=$(basename "$file")
        down_migrations+=("$filename")
    done < <(find "$migration_dir" -name "*.down.sql" -type f | sort)

    if [ ${#up_migrations[@]} -eq 0 ]; then
        echo -e "${YELLOW}no migrations found${NC}"
        return 0
    fi

    # Check that every .up.sql has a corresponding .down.sql
    local missing_down=0
    for up_file in "${up_migrations[@]}"; do
        local version="${up_file%%_*}"
        local down_file="${version}_$(echo "$up_file" | sed 's/.*_//' | sed 's/\.up\.sql$/.down.sql/')"

        # Find if down file exists
        local found=0
        for down in "${down_migrations[@]}"; do
            if [[ "$down" == *"$version"*"down.sql" ]]; then
                found=1
                break
            fi
        done

        if [ $found -eq 0 ]; then
            echo -e "${RED}missing down migration for: $up_file${NC}"
            ((missing_down++))
        fi
    done

    if [ $missing_down -gt 0 ]; then
        echo -e "${RED}✗ Missing $missing_down down migration(s)${NC}"
        ((SERVICES_FAILED++))
        return 1
    fi

    # Check for valid migration names
    local invalid_count=0
    for file in "$migration_dir"/*.sql; do
        filename=$(basename "$file")
        # Valid format: NNN_description.{up,down}.sql
        if ! [[ "$filename" =~ ^[0-9]{3}_.+\.(up|down)\.sql$ ]]; then
            echo -e "${RED}invalid migration name: $filename (expected: NNN_description.{up,down}.sql)${NC}"
            ((invalid_count++))
        fi
    done

    if [ $invalid_count -gt 0 ]; then
        echo -e "${RED}✗ Found $invalid_count invalid migration name(s)${NC}"
        ((SERVICES_FAILED++))
        return 1
    fi

    echo -e "${GREEN}✓ ${#up_migrations[@]} migrations (all have down)${NC}"
    ((SERVICES_OK++))
    return 0
}

list_services() {
    # Find all services with migrations directories
    local services=()
    for dir in "$PROJECT_ROOT"/services/*-service; do
        if [ -d "$dir/migrations" ]; then
            local service=$(basename "$dir" | sed 's/-service$//')
            services+=("$service")
        fi
    done

    echo "${services[@]}"
}

# ──────────────────────────────────────────────────────────────
# Main
# ──────────────────────────────────────────────────────────────

main() {
    echo -e "\n${BLUE}═══════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  Migration Validation${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}\n"

    local services=$(list_services)

    if [ -z "$services" ]; then
        echo -e "${RED}No migration services found in services/ directory${NC}"
        exit 1
    fi

    # Validate each service
    for service in $services; do
        validate_service_migrations "$service"
    done

    # Print summary
    echo -e "\n${BLUE}═══════════════════════════════════════════════════════${NC}"
    echo "Migration Validation Summary:"
    echo -e "  ${GREEN}Valid:${NC}   $SERVICES_OK"
    echo -e "  ${RED}Failed:${NC}  $SERVICES_FAILED"
    echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}\n"

    if [ $SERVICES_FAILED -gt 0 ]; then
        echo -e "${RED}❌ Migration validation FAILED${NC}"
        echo "Please fix the above issues:"
        echo "  1. All .up.sql files must have corresponding .down.sql files"
        echo "  2. Migration files must be named: NNN_description.{up,down}.sql"
        echo "  3. The version number (NNN) must match between up and down files"
        exit 1
    else
        echo -e "${GREEN}✅ All migrations are valid${NC}"
        exit 0
    fi
}

main "$@"
