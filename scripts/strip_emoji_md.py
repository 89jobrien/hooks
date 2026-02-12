#!/usr/bin/env python3
"""
Strip emoji and fix whitespace in project markdown.
- Removes emoji (Unicode symbols, dingbats, pictographs)
- Collapses double spaces to single
- Fixes table cell spacing (|  | -> | |)
- Trims leading space before ** on a line
Excludes: .venv, venv, node_modules, .git, .cursor
"""
from __future__ import annotations

import re
import sys
from pathlib import Path

# Unicode ranges: symbols, dingbats, misc symbols, emoticons, symbols & pictographs, etc.
EMOJI_PATTERN = re.compile(
    "["
    "\u2600-\u26FF"   # Misc symbols
    "\u2700-\u27BF"   # Dingbats
    "\u2B50-\u2B55"   # Stars etc
    "\uFE00-\uFE0F"   # Variation selectors (often with emoji)
    "\U0001F300-\U0001F9FF"   # Misc symbols and pictographs, supplemental
    "\U0001F600-\U0001F64F"   # Emoticons
    "\U0001F1E0-\U0001F1FF"   # Flags
    "\U00002702-\U000027B0"
    "\U000024C2-\U0001F251"
    "]+",
    re.UNICODE,
)

EXCLUDES = {".venv", "venv", "node_modules", ".git", ".cursor"}


def clean_line(line: str) -> str:
    s = EMOJI_PATTERN.sub("", line)
    s = re.sub(r"  +", " ", s)
    s = re.sub(r"^\s+(\*\*)", r"\1", s)
    return s


def clean_table_cells(content: str) -> str:
    return re.sub(r"\|\s{2,}\|", "| |", content)


def clean_file(path: Path) -> bool:
    try:
        text = path.read_text(encoding="utf-8", errors="replace")
    except OSError:
        return False
    lines = [clean_line(line) for line in text.splitlines()]
    new_content = clean_table_cells("\n".join(lines))
    if new_content != text:
        path.write_text(new_content, encoding="utf-8")
        return True
    return False


def main() -> None:
    root = Path(__file__).resolve().parent.parent
    modified = 0
    for path in root.rglob("*.md"):
        parts = path.relative_to(root).parts
        if any(p in EXCLUDES for p in parts):
            continue
        if clean_file(path):
            print(path)
            modified += 1
    if modified:
        print(f"Cleaned {modified} file(s).", file=sys.stderr)
    sys.exit(0)


if __name__ == "__main__":
    main()
