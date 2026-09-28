#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd -- "$SCRIPT_DIR/../.." && pwd)"
ENV_FILE="${FASTIMG_ENV_FILE:-$SCRIPT_DIR/fastimg.env}"
COMPOSE_FILE="$SCRIPT_DIR/docker-compose.yml"

log() { printf '[fastimg] %s\n' "$*"; }
die() { printf '[fastimg] ERROR: %s\n' "$*" >&2; exit 1; }

usage() {
    cat <<'EOF'
Usage: deploy.sh [--check] [--build-only]

Environment:
  FASTIMG_ENV_FILE       API env file (default: deploy/docker/fastimg.env)
  FASTIMG_WEB_PORT       Host port for the web container, set in env file
  FASTIMG_PULL_IMAGES    Set to 1 to pull newer base images before building

The compose file intentionally contains only FastImg API and web services.
PostgreSQL and Redis are external services configured in fastimg.env.
EOF
}

env_value() {
    local key="$1"
    sed -n "s/^${key}=//p" "$ENV_FILE" | tail -n 1
}

require_value() {
    local key="$1" value
    value="$(env_value "$key")"
    [[ -n "$value" ]] || die "$key is missing in $ENV_FILE"
}

validate_env() {
    [[ -f "$ENV_FILE" ]] || die "missing $ENV_FILE; copy fastimg.env.example and fill production values"
    for key in APP_ENV APP_DEBUG APP_URL APP_KEY JWT_SECRET DB_HOST DB_PORT DB_DATABASE DB_USERNAME DB_PASSWORD DB_SSLMODE REDIS_HOST REDIS_PORT CORS_ALLOWED_ORIGINS; do
        require_value "$key"
    done

    [[ "$(env_value APP_ENV)" == "production" ]] || die "APP_ENV must be production"
    [[ "$(env_value APP_DEBUG)" == "false" ]] || die "APP_DEBUG must be false"
    if [[ "$(env_value APP_URL)" != https://* ]]; then
        [[ "${FASTIMG_ALLOW_HTTP:-0}" == "1" && "$(env_value APP_URL)" == http://* ]] || die "APP_URL must use https://"
        log "WARNING: allowing temporary HTTP APP_URL for local/Baota bootstrap; replace it with the final HTTPS domain before public launch"
    fi
    if [[ "$(env_value DB_SSLMODE)" == "disable" ]]; then
        [[ "${FASTIMG_ALLOW_LOCAL_DB_SSL_DISABLE:-0}" == "1" && "$(env_value DB_HOST)" == 127.* || "${FASTIMG_ALLOW_LOCAL_DB_SSL_DISABLE:-0}" == "1" && "$(env_value DB_HOST)" == localhost || "${FASTIMG_ALLOW_LOCAL_DB_SSL_DISABLE:-0}" == "1" && "$(env_value DB_HOST)" == ::1 ]] || die "DB_SSLMODE=disable is forbidden for production"
        log "WARNING: allowing local PostgreSQL without SSL for initial bootstrap; configure TLS before public launch"
    fi
    local app_key jwt_secret
    app_key="$(env_value APP_KEY)"
    jwt_secret="$(env_value JWT_SECRET)"
    [[ ${#app_key} -ge 32 ]] || die "APP_KEY must be a real secret of at least 32 characters"
    [[ ${#jwt_secret} -ge 32 ]] || die "JWT_SECRET must be a real secret of at least 32 characters"
    [[ "$(env_value CORS_ALLOWED_ORIGINS)" != *'*'* ]] || die "CORS_ALLOWED_ORIGINS must not contain *"
}

compose() {
    docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" "$@"
}

if [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
    usage
    exit 0
fi
validate_env
command -v docker >/dev/null 2>&1 || die "docker is required"
docker compose version >/dev/null 2>&1 || die "Docker Compose v2 is required"

services="$(compose config --services)"
if printf '%s\n' "$services" | grep -Eq '^(postgres|redis|database)$'; then
    die "the deployment must not start a database or Redis container"
fi

if [[ "${1:-}" == "--check" ]]; then
    compose config --quiet
    log "configuration is valid; external PostgreSQL and Redis are required"
    exit 0
fi

compose config --quiet
if [[ "${FASTIMG_PULL_IMAGES:-0}" == "1" ]]; then
    log "pulling base images"
    compose build --pull
else
    compose build
fi

if [[ "${1:-}" == "--build-only" ]]; then
    log "images built successfully"
    exit 0
fi

log "running database migrations against the configured external PostgreSQL"
compose run --rm api artisan migrate --no-ansi

log "starting FastImg API and web containers"
compose up -d

web_port="$(env_value FASTIMG_WEB_PORT)"
web_port="${web_port:-8080}"
for attempt in $(seq 1 30); do
    if curl --fail --silent --show-error "http://127.0.0.1:${web_port}/api/v1/discovery/status" >/dev/null; then
        log "FastImg is healthy at http://127.0.0.1:${web_port}/"
        compose ps
        exit 0
    fi
    sleep 2
done

compose ps
die "FastImg did not become healthy; inspect: docker compose --env-file $ENV_FILE -f $COMPOSE_FILE logs --tail=200"
