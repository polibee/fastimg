#!/usr/bin/env bash
set -Eeuo pipefail

ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)"
CLI="$ROOT/deploy/fastimg-cli.sh"

[[ -f "$CLI" ]] || { echo "missing deploy CLI" >&2; exit 1; }
bash -n "$CLI"
help="$(bash "$CLI" --help)"
grep -q -- '--mode source|docker' <<<"$help"
grep -q -- 'PostgreSQL' <<<"$help"
grep -q -- 'Redis' <<<"$help"
grep -q -- 'loopback upstream' <<<"$help"
grep -q -- 'Baota' <<<"$help"
grep -q -- 'automatic' <<<"$help"
grep -q -- '127.0.0.1' <<<"$help"

echo 'fastimg CLI tests: PASS'
