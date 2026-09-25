#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
backend_root="$repo_root/backend"
cache_root="${FASTIMG_GOCACHE_ROOT:-${XDG_CACHE_HOME:-$HOME/.cache}/fastimg/go-build}"

mkdir -p "$cache_root"
export GOCACHE="$cache_root"
printf 'FastImg Go cache: %s\n' "$GOCACHE"
cd "$backend_root"
exec go "$@"
