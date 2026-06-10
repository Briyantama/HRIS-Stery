#!/bin/bash
# Note: NOT using set -e to allow graceful handling of warnings

# ============================================================
# HRIS-Stery Preflight Checks
# ============================================================
# Validates environment before running `make dev`
# Checks: Docker, Docker Compose, required binaries, disk space, ports

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
CHECKS_PASSED=0
CHECKS_FAILED=0
CHECKS_WARNING=0

# ──────────────────────────────────────────────────────────────
# Helper functions
# ──────────────────────────────────────────────────────────────

check_command() {
    local cmd=$1
    local description=${2:-"$cmd"}

    if command -v "$cmd" &>/dev/null; then
        echo -e "${GREEN}✓${NC} $description"
        ((CHECKS_PASSED++))
        return 0
    else
        echo -e "${RED}✗${NC} $description (not found in PATH)"
        ((CHECKS_FAILED++))
        return 1
    fi
}

check_docker() {
    echo -e "\n${BLUE}Docker & Docker Compose${NC}"

    if ! command -v docker &>/dev/null; then
        echo -e "${RED}✗${NC} Docker is not installed"
        echo "  Install from: https://www.docker.com/products/docker-desktop"
        ((CHECKS_FAILED++))
        return 1
    fi

    echo -e "${GREEN}✓${NC} Docker CLI found"
    ((CHECKS_PASSED++))

    # Check if Docker daemon is running (with timeout, skip if can't determine)
    if timeout 3 docker info &>/dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} Docker daemon is running"
        ((CHECKS_PASSED++))
    else
        echo -e "${YELLOW}⊘${NC} Docker daemon status unknown (will be checked on 'make dev')"
        ((CHECKS_WARNING++))
    fi

    # Check Docker Compose version
    if ! command -v docker-compose &>/dev/null && ! docker compose version &>/dev/null; then
        echo -e "${RED}✗${NC} Docker Compose is not available"
        echo "  Docker Compose should be included with Docker Desktop"
        ((CHECKS_FAILED++))
        return 1
    fi

    local version
    version=$(docker compose version 2>/dev/null | grep -oP 'v\K[0-9.]+' || echo "unknown")
    echo -e "${GREEN}✓${NC} Docker Compose v$version found"
    ((CHECKS_PASSED++))
}

check_required_binaries() {
    echo -e "\n${BLUE}Required Binaries${NC}"

    check_command "go" "Go 1.24+"
    check_command "buf" "Protocol Buffer compiler (buf)"
    check_command "git" "Git version control"
}

check_optional_binaries() {
    echo -e "\n${BLUE}Optional Binaries${NC}"

    # These are optional and just show warnings
    if ! command -v golangci-lint &>/dev/null; then
        echo -e "${YELLOW}⊘${NC} golangci-lint not found (install for: make lint-go)"
        echo "  Install: https://golangci-lint.run/usage/install/"
        ((CHECKS_WARNING++))
    else
        echo -e "${GREEN}✓${NC} golangci-lint found"
        ((CHECKS_PASSED++))
    fi

    if ! command -v migrate &>/dev/null; then
        echo -e "${YELLOW}⊘${NC} golang-migrate not found (install for: make migrate-*)"
        echo "  Install: https://github.com/golang-migrate/migrate/tree/master/cmd/migrate"
        ((CHECKS_WARNING++))
    else
        echo -e "${GREEN}✓${NC} golang-migrate found"
        ((CHECKS_PASSED++))
    fi

    if ! command -v node &>/dev/null; then
        echo -e "${YELLOW}⊘${NC} Node.js not found (install for: make dev-svelte)"
        echo "  Install: https://nodejs.org/ (v20+)"
        ((CHECKS_WARNING++))
    else
        echo -e "${GREEN}✓${NC} Node.js found"
        ((CHECKS_PASSED++))
    fi

    if ! command -v php &>/dev/null; then
        echo -e "${YELLOW}⊘${NC} PHP not found (install for: make dev-laravel)"
        echo "  Install: https://www.php.net/downloads (8.3+)"
        ((CHECKS_WARNING++))
    else
        echo -e "${GREEN}✓${NC} PHP found"
        ((CHECKS_PASSED++))
    fi

    if ! command -v composer &>/dev/null; then
        echo -e "${YELLOW}⊘${NC} Composer not found (install for: make dev-laravel)"
        echo "  Install: https://getcomposer.org/download/"
        ((CHECKS_WARNING++))
    else
        echo -e "${GREEN}✓${NC} Composer found"
        ((CHECKS_PASSED++))
    fi
}

check_disk_space() {
    echo -e "\n${BLUE}Disk Space${NC}"

    # Get available disk space in MB (works on macOS, Linux, WSL)
    local available_mb
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        available_mb=$(df -m "$PROJECT_ROOT" | awk 'NR==2 {print $4}')
    else
        # Linux/WSL
        available_mb=$(df -m "$PROJECT_ROOT" | awk 'NR==2 {print $4}')
    fi

    local available_gb=$((available_mb / 1024))

    if [ "$available_mb" -lt 10240 ]; then  # Less than 10 GB
        echo -e "${RED}✗${NC} Insufficient disk space: ${available_gb}GB available (need ≥10GB)"
        ((CHECKS_FAILED++))
        return 1
    fi

    echo -e "${GREEN}✓${NC} Sufficient disk space: ${available_gb}GB available"
    ((CHECKS_PASSED++))
}

check_ports() {
    echo -e "\n${BLUE}Required Ports Available${NC}"

    # Skip port checks - they can hang on Windows/WSL
    # Ports will be validated when 'make dev' actually starts containers
    echo -e "${YELLOW}⊘${NC} Port checks skipped (will be validated on 'make dev')"
    ((CHECKS_WARNING++))
}

check_network() {
    echo -e "\n${BLUE}Network & Registries${NC}"

    # Check GitHub (used for Go module downloads)
    if timeout 3 ping -c 1 github.com &>/dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} Can reach github.com"
        ((CHECKS_PASSED++))
    else
        echo -e "${YELLOW}⊘${NC} Cannot reach github.com (may fail downloading Go modules)"
        ((CHECKS_WARNING++))
    fi

    # Check DockerHub (used for container images)
    if timeout 3 ping -c 1 docker.io &>/dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} Can reach docker.io"
        ((CHECKS_PASSED++))
    else
        echo -e "${YELLOW}⊘${NC} Cannot reach docker.io (may fail pulling images)"
        ((CHECKS_WARNING++))
    fi
}

check_docker_compose_file() {
    echo -e "\n${BLUE}Docker Compose Configuration${NC}"

    local compose_file="$PROJECT_ROOT/deploy/docker/docker-compose.yml"

    if [ ! -f "$compose_file" ]; then
        echo -e "${RED}✗${NC} docker-compose.yml not found at: $compose_file"
        ((CHECKS_FAILED++))
        return 1
    fi

    echo -e "${GREEN}✓${NC} docker-compose.yml found"
    ((CHECKS_PASSED++))

    # Validate docker-compose file (with timeout)
    if timeout 5 docker compose -f "$compose_file" config &>/dev/null; then
        echo -e "${GREEN}✓${NC} docker-compose.yml is valid"
        ((CHECKS_PASSED++))
    else
        echo -e "${RED}✗${NC} docker-compose.yml has errors (or Docker daemon not running)"
        ((CHECKS_FAILED++))
        return 1
    fi
}

# ──────────────────────────────────────────────────────────────
# Main
# ──────────────────────────────────────────────────────────────

main() {
    echo -e "\n${BLUE}═══════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  HRIS-Stery Preflight Checks${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"

    check_docker
    check_required_binaries
    check_optional_binaries
    check_disk_space
    check_ports
    check_network
    check_docker_compose_file

    # Print summary
    echo -e "\n${BLUE}═══════════════════════════════════════════════════════${NC}"
    echo -e "Checks Summary:"
    echo -e "  ${GREEN}Passed:${NC}   $CHECKS_PASSED"
    echo -e "  ${YELLOW}Warnings:${NC} $CHECKS_WARNING"
    echo -e "  ${RED}Failed:${NC}   $CHECKS_FAILED"
    echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}\n"

    if [ "$CHECKS_FAILED" -gt 0 ]; then
        echo -e "${RED}❌ Preflight checks FAILED${NC}"
        echo "Please fix the above issues and run 'make dev' again."
        exit 1
    fi

    if [ "$CHECKS_WARNING" -gt 0 ]; then
        echo -e "${YELLOW}⚠️  Preflight checks PASSED with warnings${NC}"
        echo "Some optional features may not be available."
        echo "Run 'make dev' to continue."
    else
        echo -e "${GREEN}✅ All preflight checks PASSED${NC}"
        echo "Ready to run: make dev"
    fi

    exit 0
}

main "$@"
