#!/usr/bin/env python3
"""
Merge i18n locale JSON files by copying missing keys.

This script is intended for keeping `bin/locale/<lang>.json` consistent with
`internal/locale/en.json` (or any other "source of truth" locale file):

- For every target JSON, it adds keys that are present in `source` but missing
  in the target.
- Existing keys in the targets are NOT overwritten.
- Output is written back to the same target file.

All JSON is read/written as UTF-8 with `ensure_ascii=False` to preserve
non-ASCII characters (Cyrillic, CJK, etc.).

Example:
  python scripts/merge_locales.py internal/locale/en.json bin/locale/fr.json

Usage:
  merge_locales.py <source.json> <target1.json> [<target2.json> ...]
"""
import json
import sys
from pathlib import Path

def load_json(p: Path) -> dict[str, str]:
    """
    Load JSON as a dict[str, str].

    If the file doesn't exist, returns an empty dict. (In normal usage you
    should pass existing target files.)
    """
    if not p.exists():
        return {}
    return json.loads(p.read_text(encoding="utf-8"))

def write_json(p: Path, data: dict[str, str]) -> None:
    """Write JSON with stable formatting and UTF-8."""
    p.write_text(
        json.dumps(data, ensure_ascii=False, indent=2) + "\n",
        encoding="utf-8",
    )

def merge(source: Path, targets: list[str]) -> dict[str, list[str]]:
    """
    Copy missing keys from `source` into each target JSON.

    Returns a mapping:
      target_path -> list of keys that were added
    """
    src = load_json(source)
    added: dict[str, list[str]] = {}

    for t in targets:
        tp = Path(t)
        dst = load_json(tp)
        added_keys: list[str] = []

        for k, v in src.items():
            if k not in dst:
                dst[k] = v
                added_keys.append(k)

        write_json(tp, dst)
        added[tp.as_posix()] = added_keys

    return added

def main():
    if len(sys.argv) < 3:
        print(
            "Usage: merge_locales.py <source.json> <target1.json> [<target2.json> ...]"
        )
        sys.exit(2)

    source = Path(sys.argv[1])
    targets = sys.argv[2:]

    added = merge(source, targets)
    for f, keys in added.items():
        print(f"{f}: added {len(keys)} keys")

if __name__ == '__main__':
    main()
