#!/usr/bin/env bash
# Unified test script for MemoNexus monorepo
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "Running MemoNexus tests..."

# Go tests
echo "Running Go tests..."
cd "$PROJECT_ROOT/packages/backend"
go test -v -cover ./...

# Flutter tests
echo "Running Flutter tests..."
cd "$PROJECT_ROOT/apps/frontend"
flutter test --coverage

echo "All tests complete!"
