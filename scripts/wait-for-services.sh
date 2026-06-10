#!/bin/bash
set -e

# ============================================================
# HRIS-Stery Service Health Check
# ============================================================
# Polls Docker Compose services until they report healthy
# or times out after MAX_WAIT seconds

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Configuration
MAX_WAIT=120  # seconds
CHECK_INTERVAL=5  # seconds
COMPOSE_FILE="${PROJECT_ROOT}/deploy/docker/docker-compose.yml"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ──────────────────────────────────────────────────────────────
# Helper functions
# ──────────────────────────────────────────────────────────────

wait_for_service() {
    local service=$1
    local port=$2
    local timeout=${3:-$MAX_WAIT}
    local start_time=$(date +%s)

    echo -ne "${BLUE}Waiting for $service (port $port)...${NC}"

    while true; do
        local current_time=$(date +%s)
        local elapsed=$((current_time - start_time))

        if [ $elapsed -gt $timeout ]; then
            echo -e " ${RED}TIMEOUT (${elapsed}s)${NC}"
            return 1
        fi

        if nc -z localhost "$port" 2>/dev/null; then
            echo -e " ${GREEN}ready (${elapsed}s)${NC}"
            return 0
        fi

        echo -ne "."
        sleep $CHECK_INTERVAL
    done
}

check_container_health() {
    local container=$1
    local max_attempts=30
    local attempt=1

    echo -ne "${BLUE}Checking $container health...${NC}"

    while [ $attempt -le $max_attempts ]; do
        local health_status=$(docker inspect --format='{{.State.Health.Status}}' "$container" 2>/dev/null || echo "")

        if [ "$health_status" = "healthy" ]; then
            echo -e " ${GREEN}healthy${NC}"
            return 0
        elif [ "$health_status" = "unhealthy" ]; then
            echo -e " ${RED}unhealthy${NC}"
            docker logs "$container" 2>&1 | tail -5
            return 1
        elif [ -z "$health_status" ]; then
            # No health check defined, consider it ready
            echo -e " ${GREEN}running${NC}"
            return 0
        fi

        echo -ne "."
        sleep $CHECK_INTERVAL
        ((attempt++))
    done

    echo -e " ${YELLOW}timeout waiting for health${NC}"
    return 0  # Don't fail if health check takes too long
}

# ──────────────────────────────────────────────────────────────
# Main
# ──────────────────────────────────────────────────────────────

main() {
    echo -e "\n${BLUE}═══════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  Waiting for services to be ready...${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}\n"

    # List of services and their ports
    local services=(
        "postgres:5432"
        "redis:6379"
        "nats:4222"
        "minio:9000"
        "jaeger:16686"
        "prometheus:9090"
    )

    local failed_services=0

    # Wait for each service
    for service_port in "${services[@]}"; do
        local service="${service_port%%:*}"
        local port="${service_port##*:}"

        if ! wait_for_service "$service" "$port"; then
            echo -e "${RED}✗ Failed to reach $service on port $port${NC}"
            ((failed_services++))
        fi
    done

    # Check container health
    echo ""
    local containers=(
        "postgres"
        "redis"
        "nats"
        "minio"
    )

    for container in "${containers[@]}"; do
        local full_container="hris-stery-$container-1"
        if docker inspect "$full_container" &>/dev/null 2>&1; then
            if ! check_container_health "$full_container"; then
                echo -e "${YELLOW}Warning: $full_container may not be healthy${NC}"
            fi
        fi
    done

    # Print summary
    echo -e "\n${BLUE}═══════════════════════════════════════════════════════${NC}"

    if [ $failed_services -eq 0 ]; then
        echo -e "${GREEN}✅ All services are ready${NC}"
        echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}\n"
        return 0
    else
        echo -e "${RED}❌ $failed_services service(s) failed to start${NC}"
        echo "Run 'make dev-logs' to see container logs"
        echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}\n"
        return 1
    fi
}

main "$@"
