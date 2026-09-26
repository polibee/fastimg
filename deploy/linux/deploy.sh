#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd -- "$SCRIPT_DIR/../.." && pwd)"
INSTALL_ROOT="${FASTIMG_INSTALL_ROOT:-/opt/fastimg}"
SOURCE_ROOT="${FASTIMG_SOURCE_ROOT:-$REPO_ROOT}"
ENV_FILE="${FASTIMG_ENV_FILE:-$INSTALL_ROOT/shared/.env}"
APP_USER="${FASTIMG_APP_USER:-fastimg}"
APP_GROUP="${FASTIMG_APP_GROUP:-$APP_USER}"
API_PORT="${FASTIMG_API_PORT:-8080}"
CHECK_DEPENDENCIES="${FASTIMG_CHECK_DEPENDENCIES:-1}"
INSTALL_NGINX="${FASTIMG_INSTALL_NGINX:-0}"
RELEASE_ID="$(date -u +%Y%m%d%H%M%S)"
RELEASE_ROOT="$INSTALL_ROOT/releases/$RELEASE_ID"

log() { printf '[fastimg] %s\n' "$*"; }
die() { printf '[fastimg] ERROR: %s\n' "$*" >&2; exit 1; }

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
    [[ -f "$ENV_FILE" ]] || die "missing $ENV_FILE; copy deploy/docker/fastimg.env.example and fill production values"
    for key in APP_ENV APP_DEBUG APP_URL APP_KEY JWT_SECRET APP_PORT DB_HOST DB_PORT DB_DATABASE DB_USERNAME DB_PASSWORD DB_SSLMODE REDIS_HOST REDIS_PORT CORS_ALLOWED_ORIGINS; do
        require_value "$key"
    done
    [[ "$(env_value APP_ENV)" == "production" ]] || die "APP_ENV must be production"
    [[ "$(env_value APP_DEBUG)" == "false" ]] || die "APP_DEBUG must be false"
    [[ "$(env_value APP_URL)" == https://* ]] || die "APP_URL must use https://"
    [[ "$(env_value DB_SSLMODE)" != "disable" ]] || die "DB_SSLMODE=disable is forbidden for production"
    local app_key jwt_secret
    app_key="$(env_value APP_KEY)"
    jwt_secret="$(env_value JWT_SECRET)"
    [[ ${#app_key} -ge 32 ]] || die "APP_KEY must be at least 32 characters"
    [[ ${#jwt_secret} -ge 32 ]] || die "JWT_SECRET must be at least 32 characters"
    [[ "$(env_value CORS_ALLOWED_ORIGINS)" != *'*'* ]] || die "CORS_ALLOWED_ORIGINS must not contain *"
}

require_command() {
    command -v "$1" >/dev/null 2>&1 || die "$1 is required"
}

run_as_app() {
    if [[ "$(id -u)" -eq 0 ]]; then
        runuser -u "$APP_USER" -- "$@"
    else
        "$@"
    fi
}

check_external_dependencies() {
    [[ "$CHECK_DEPENDENCIES" == "1" ]] || return 0
    require_command psql
    require_command redis-cli
    log "checking external PostgreSQL"
    PGPASSWORD="$(env_value DB_PASSWORD)" PGSSLMODE="$(env_value DB_SSLMODE)" \
        psql -h "$(env_value DB_HOST)" -p "$(env_value DB_PORT)" \
        -U "$(env_value DB_USERNAME)" -d "$(env_value DB_DATABASE)" \
        -Atqc 'select 1' >/dev/null
    log "checking external Redis"
    if [[ -n "$(env_value REDIS_PASSWORD)" ]]; then
        redis-cli -h "$(env_value REDIS_HOST)" -p "$(env_value REDIS_PORT)" \
            -a "$(env_value REDIS_PASSWORD)" --no-auth-warning ping | grep -qx PONG
    else
        redis-cli -h "$(env_value REDIS_HOST)" -p "$(env_value REDIS_PORT)" ping | grep -qx PONG
    fi
}

ensure_user_and_directories() {
    [[ "$(id -u)" -eq 0 ]] || die "run this installer as root so it can create a systemd service"
    if ! getent group "$APP_GROUP" >/dev/null; then
        groupadd --system "$APP_GROUP"
    fi
    if ! id -u "$APP_USER" >/dev/null 2>&1; then
        useradd --system --gid "$APP_GROUP" --home-dir "$INSTALL_ROOT" --shell /usr/sbin/nologin "$APP_USER"
    fi
    install -d -o "$APP_USER" -g "$APP_GROUP" "$INSTALL_ROOT/shared/storage/fastimg" "$INSTALL_ROOT/shared/storage/logs"
    install -d -o "$APP_USER" -g "$APP_GROUP" "$INSTALL_ROOT/releases"
    install -d -o root -g "$APP_GROUP" -m 0750 "$INSTALL_ROOT/shared"
    chown "$APP_USER:$APP_GROUP" "$ENV_FILE"
    chmod 0640 "$ENV_FILE"
}

build_release() {
    require_command go
    require_command node
    if command -v pnpm >/dev/null 2>&1; then
        PNPM=(pnpm)
    else
        require_command corepack
        PNPM=(corepack pnpm)
    fi
    [[ -f "$SOURCE_ROOT/admin/pnpm-lock.yaml" ]] || die "admin/pnpm-lock.yaml is missing"
    [[ -f "$SOURCE_ROOT/backend/go.mod" ]] || die "backend/go.mod is missing"

    [[ ! -e "$RELEASE_ROOT" ]] || die "release directory already exists: $RELEASE_ROOT"
    install -d "$RELEASE_ROOT/backend" "$RELEASE_ROOT/web"

    log "building the SSG frontend"
    (
        cd "$SOURCE_ROOT/admin"
        "${PNPM[@]}" install --frozen-lockfile
        SSG_PUBLIC_ORIGIN="$(env_value APP_URL)" VITE_API_BASE_URL="" "${PNPM[@]}" run build:ssg
    )
    cp -a "$SOURCE_ROOT/admin/dist/." "$RELEASE_ROOT/web/"

    log "building the Go API"
    (
        cd "$SOURCE_ROOT/backend"
        go build -trimpath -ldflags='-s -w' -o "$RELEASE_ROOT/backend/fastimg-api" .
    )
    cp -a "$SOURCE_ROOT/backend/public" "$RELEASE_ROOT/backend/public"
    cp -a "$SOURCE_ROOT/backend/resources" "$RELEASE_ROOT/backend/resources"
    ln -s "$INSTALL_ROOT/shared/storage" "$RELEASE_ROOT/backend/storage"
    ln -s "$ENV_FILE" "$RELEASE_ROOT/backend/.env"
    chown -R "$APP_USER:$APP_GROUP" "$RELEASE_ROOT"
}

install_systemd_service() {
    local rendered
    rendered="$(mktemp)"
    sed \
        -e "s#__FASTIMG_ROOT__#${INSTALL_ROOT}#g" \
        -e "s#__FASTIMG_USER__#${APP_USER}#g" \
        -e "s#__FASTIMG_GROUP__#${APP_GROUP}#g" \
        "$SCRIPT_DIR/fastimg-api.service" > "$rendered"
    install -o root -g root -m 0644 "$rendered" /etc/systemd/system/fastimg-api.service
    rm -f "$rendered"
    ln -sfn "$RELEASE_ROOT" "$INSTALL_ROOT/current"
    systemctl daemon-reload
    systemctl enable fastimg-api.service >/dev/null
    systemctl restart fastimg-api.service
}

install_nginx() {
    [[ "$INSTALL_NGINX" == "1" ]] || return 0
    require_command nginx
    require_command systemctl
    local domain="${FASTIMG_DOMAIN:-}" cert="${FASTIMG_TLS_CERT:-}" key="${FASTIMG_TLS_KEY:-}" rendered
    [[ -n "$domain" && -n "$cert" && -n "$key" ]] || die "FASTIMG_DOMAIN, FASTIMG_TLS_CERT and FASTIMG_TLS_KEY are required when FASTIMG_INSTALL_NGINX=1"
    [[ -f "$cert" && -f "$key" ]] || die "TLS certificate files do not exist"
    rendered="$(mktemp)"
    sed \
        -e "s#__FASTIMG_DOMAIN__#${domain}#g" \
        -e "s#__FASTIMG_WEB_ROOT__#${INSTALL_ROOT}/current/web#g" \
        -e "s#__FASTIMG_API_PORT__#${API_PORT}#g" \
        -e "s#__FASTIMG_TLS_CERT__#${cert}#g" \
        -e "s#__FASTIMG_TLS_KEY__#${key}#g" \
        "$REPO_ROOT/deploy/nginx/fastimg-source.conf.example" > "$rendered"
    install -o root -g root -m 0644 "$rendered" /etc/nginx/sites-available/fastimg.conf
    ln -sfn /etc/nginx/sites-available/fastimg.conf /etc/nginx/sites-enabled/fastimg.conf
    rm -f "$rendered"
    nginx -t
    systemctl reload nginx
}

health_check() {
    require_command curl
    for attempt in $(seq 1 30); do
        if curl --fail --silent --show-error "http://127.0.0.1:${API_PORT}/api/v1/discovery/status" >/dev/null; then
            return 0
        fi
        sleep 2
    done
    systemctl status fastimg-api.service --no-pager || true
    journalctl -u fastimg-api.service -n 100 --no-pager || true
    die "FastImg API did not become healthy"
}

validate_env
[[ "${FASTIMG_API_PORT:-}" =~ ^[0-9]+$ ]] || die "FASTIMG_API_PORT must be numeric when provided"
API_PORT="${FASTIMG_API_PORT:-$(env_value APP_PORT)}"
ensure_user_and_directories
check_external_dependencies
build_release

log "running database migrations against the configured external PostgreSQL"
run_as_app "$RELEASE_ROOT/backend/fastimg-api" artisan migrate --no-ansi

install_systemd_service
health_check
install_nginx

log "FastImg source deployment completed"
log "API: http://127.0.0.1:${API_PORT}/api/v1/discovery/status"
log "Web files: $INSTALL_ROOT/current/web"
log "Service: systemctl status fastimg-api.service"
log "Back up PostgreSQL and validate real payment/TLS/object-storage/pressure gates before public launch"
