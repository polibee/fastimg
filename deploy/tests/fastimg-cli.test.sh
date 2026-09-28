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
grep -q -- 'member URL' <<<"$help"
grep -q -- 'admin URL' <<<"$help"
grep -q -- 'API URL' <<<"$help"

echo 'fastimg CLI tests: PASS'
