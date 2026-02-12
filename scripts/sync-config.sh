#!/usr/bin/env bash
# Regenerate .cursor/hooks.json and .claude/settings.json from config.
# Run from repo root (or with script at repo/hooks/scripts/sync-config.sh).
# Supports: .hooks/config.yaml (install.sh layout) or hooks/config.yaml (subdir layout).
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$REPO_ROOT"

if [[ -f .hooks/config.yaml ]]; then
	./.hooks/bin/gen-config
elif [[ -f hooks/config.yaml ]]; then
	make -C hooks all
	make -C hooks config
else
	echo "no hooks/config.yaml or .hooks/config.yaml found; run from repo root" >&2
	exit 1
fi
