#!/usr/bin/env bash
# Install hooks into another repo. Builds bins in this repo, copies only .hooks/bin and
# .hooks/config.yaml into TARGET, then runs gen-config. Usage: ./install.sh /path/to/target/repo
set -euo pipefail

TARGET="${1:?usage: $0 /path/to/target/repo}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Recommend installing into a git repo
if [[ "$(git -C "$TARGET" rev-parse --is-inside-work-tree 2>/dev/null)" != "true" ]]; then
  echo "warning: $TARGET is not a git repository; using a git repo is recommended" >&2
  read -r -p "Continue anyway? [y/N] " reply
  if [[ ! "$reply" =~ ^[yY]$ ]]; then
    echo "aborted" >&2
    exit 1
  fi
fi

# Source = directory containing install.sh (hooks repo root)
SOURCE_DIR="$SCRIPT_DIR"
if [[ ! -f "$SOURCE_DIR/config.yaml" || ! -d "$SOURCE_DIR/cmd" ]]; then
  echo "error: config.yaml and cmd/ not found in $SOURCE_DIR" >&2
  exit 1
fi

# Build binaries in source repo
make -C "$SOURCE_DIR" all

# Install only bin and config into target
mkdir -p "$TARGET/.hooks/bin"
cp "$SOURCE_DIR"/bin/* "$TARGET/.hooks/bin/"
cp "$SOURCE_DIR/config.yaml" "$TARGET/.hooks/config.yaml"

cd "$TARGET"
.hooks/bin/gen-config

echo "installed to $TARGET/.hooks; wrote $TARGET/.cursor/ and $TARGET/.claude/"
if [[ -f "$TARGET/.cursor/hooks.env" ]]; then
  echo "optional: source $TARGET/.cursor/hooks.env"
fi
