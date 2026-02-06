#!/usr/bin/env bash
# Regenerate .cursor/hooks.json and .claude/settings.json from hooks/config.yaml.
# Run from repo root. Use in pre-commit or after editing config.yaml.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$REPO_ROOT"
make -C hooks all
make -C hooks config
