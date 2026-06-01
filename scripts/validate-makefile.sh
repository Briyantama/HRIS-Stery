#!/usr/bin/env bash
# Validates every Makefile target and records pass/fail.
# Usage: ./scripts/validate-makefile.sh
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

LOG_DIR="${ROOT}/.make-validate"
mkdir -p "$LOG_DIR"
RESULTS="${LOG_DIR}/results.tsv"
echo -e "target\texit\tstatus" > "$RESULTS"

run_target() {
  local target="$1"
  local log="${LOG_DIR}/${target}.log"
  local allow_fail="${2:-0}"
  echo "======== make ${target} ========"
  set +e
  make "$target" >"$log" 2>&1
  local code=$?
  set -e
  local status="PASS"
  if [ "$code" -ne 0 ]; then
    if [ "$allow_fail" -eq 1 ]; then
      status="EXPECTED_FAIL"
    else
      status="FAIL"
    fi
  fi
  echo -e "${target}\t${code}\t${status}" >> "$RESULTS"
  tail -n 5 "$log" | sed 's/^/  /'
  echo ""
}

# Non-destructive / local targets
run_target help
run_target check-tools
run_target go-services
run_target proto-deps
run_target proto-lint
run_target proto-gen
run_target proto-breaking 1
run_target tidy
run_target vet
run_target test-go
run_target build-go
run_target lint-go 1
run_target test-laravel 1
run_target test-svelte 1
run_target lint-laravel 1
run_target lint-svelte 1
run_target build-svelte 1
run_target clean
run_target test 1
run_target lint 1
run_target build 1

# Infra (may fail without Docker)
run_target dev-ps 1
run_target migrate-status 1

echo "Results written to ${RESULTS}"
column -t "$RESULTS" 2>/dev/null || cat "$RESULTS"
