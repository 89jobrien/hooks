#!/usr/bin/env bash
# Print audit and cost summary. Uses HOOK_AUDIT_DIR and HOOK_COST_DIR or ~/.cursor/audit and ~/.cursor/cost.
# Audit: line count from log files modified in last 24h (approximate). Cost: sum of tokens from cost.log.
set -euo pipefail
AUDIT_DIR="${HOOK_AUDIT_DIR:-$HOME/.cursor/audit}"
COST_DIR="${HOOK_COST_DIR:-$HOME/.cursor/cost}"
calls=0
tokens=0

if [[ -d "$AUDIT_DIR" ]]; then
  while IFS= read -r f; do
    calls=$((calls + $(wc -l < "$f" 2>/dev/null || echo 0)))
  done < <(find "$AUDIT_DIR" -maxdepth 1 -name "audit-*.log" -mtime -1 2>/dev/null)
fi

if [[ -f "$COST_DIR/cost.log" ]]; then
  while IFS= read -r line; do
    if [[ "$line" =~ tokens=([0-9]+) ]]; then
      tokens=$((tokens + BASH_REMATCH[1]))
    fi
  done < "$COST_DIR/cost.log"
fi

echo "Recent: $calls tool calls, ~$tokens tokens (estimate)"
