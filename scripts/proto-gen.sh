#!/usr/bin/env bash
# Regenerates all Go gRPC stubs and grpc-gateway proxies from proto definitions.
# Requires: buf (https://buf.build/docs/installation)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

echo "→ Running buf generate..."
cd "${REPO_ROOT}/proto"
buf generate

echo "✓ Proto stubs regenerated in ${REPO_ROOT}/gen/"
