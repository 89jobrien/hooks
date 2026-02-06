#!/usr/bin/env bash
# Install hooks into another repo.
# Usage: from this repo root, run: ./hooks/install.sh /path/to/target/repo
# Or: from hooks dir, run: ./install.sh /path/to/target/repo (source = parent dir)
set -euo pipefail

TARGET="${1:?usage: $0 /path/to/target/repo}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SOURCE_REPO="$(cd "$SCRIPT_DIR/.." && pwd)"

if [[ ! -d "$SOURCE_REPO/hooks" ]]; then
  echo "error: expected hooks dir at $SOURCE_REPO/hooks" >&2
  exit 1
fi

mkdir -p "$TARGET"
cp -R "$SOURCE_REPO/hooks" "$TARGET/hooks"
cd "$TARGET"
make -C hooks all
make -C hooks config
echo "installed hooks to $TARGET (.cursor/hooks.json, .claude/settings.json, hooks/bin/)"
echo "optional: source $TARGET/.cursor/hooks.env before starting Cursor to set per-hook env"
