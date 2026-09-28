#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd -- "$SCRIPT_DIR/.." && pwd)"
MODE="${FASTIMG_DEPLOY_MODE:-}"
NON_INTERACTIVE="${FASTIMG_NON_INTERACTIVE:-0}"
CHECK_ONLY=0
APP_URL_VALUE="${FASTIMG_APP_URL:-}"
CORS_VALUE="${FASTIMG_CORS_ALLOWED_ORIGINS:-}"
DB_HOST_VALUE="${FASTIMG_DB_HOST:-127.0.0.1}"
DB_PORT_VALUE="${FASTIMG_DB_PORT:-5432}"
DB_DATABASE_VALUE="${FASTIMG_DB_DATABASE:-fastimg}"
DB_USERNAME_VALUE="${FASTIMG_DB_USERNAME:-}"
DB_PASSWORD_VALUE="${FASTIMG_DB_PASSWORD:-}"
DB_SSLMODE_VALUE="${FASTIMG_DB_SSLMODE:-verify-full}"
APP_USER_VALUE="${FASTIMG_APP_USER:-fastimg}"
REDIS_HOST_VALUE="${FASTIMG_REDIS_HOST:-}"
REDIS_PORT_VALUE="${FASTIMG_REDIS_PORT:-6379}"
REDIS_PASSWORD_VALUE="${FASTIMG_REDIS_PASSWORD:-}"
API_PORT_VALUE="${FASTIMG_API_PORT:-8080}"
WEB_PORT_VALUE="${FASTIMG_WEB_PORT:-8080}"
INSTALL_ROOT_VALUE="${FASTIMG_INSTALL_ROOT:-/opt/fastimg}"
ENV_FILE_VALUE="${FASTIMG_ENV_FILE:-}"
ADMIN_EMAIL_VALUE="${FASTIMG_ADMIN_EMAIL:-}"
ADMIN_NAME_VALUE="${FASTIMG_ADMIN_NAME:-Administrator}"
ADMIN_PASSWORD_VALUE="${FASTIMG_ADMIN_PASSWORD:-}"
BOOTSTRAP_ADMIN="${FASTIMG_BOOTSTRAP_ADMIN:-1}"

log() { printf '[fastimg] %s\n' "$*"; }
die() { printf '[fastimg] ERROR: %s\n' "$*" >&2; exit 1; }

usage() {
    cat <<'EOF'
FastImg deployment CLI

Usage: deploy/fastimg-cli.sh [options]

Options:
  --mode source|docker   Choose Linux source deployment or Docker deployment.
  --env-file PATH        Write the generated API environment to PATH.
  --check                Validate PostgreSQL/Redis and print endpoints only.
  --non-interactive      Read all values from FASTIMG_* environment variables.
  -h, --help             Show this help.

The CLI asks for PostgreSQL host, database, username and password, then probes
Redis automatically. APP_KEY and JWT_SECRET are generated locally. It never
starts a PostgreSQL or Redis container and never prints passwords.

After deployment it prints the member URL, admin URL and API URL. Put the
printed upstream values into your Baota reverse-proxy/site configuration.
EOF
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --mode) [[ $# -ge 2 ]] || die '--mode requires source or docker'; MODE="$2"; shift 2 ;;
        --env-file) [[ $# -ge 2 ]] || die '--env-file requires a path'; ENV_FILE_VALUE="$2"; shift 2 ;;
        --check) CHECK_ONLY=1; shift ;;
        --non-interactive) NON_INTERACTIVE=1; shift ;;
        -h|--help) usage; exit 0 ;;
        *) die "unknown option: $1" ;;
    esac
done

is_interactive() { [[ "$NON_INTERACTIVE" != "1" ]]; }

ask() {
    local label="$1" current="$2" answer
    if ! is_interactive; then printf '%s' "$current"; return; fi
    if [[ -n "$current" ]]; then
        read -r -p "$label [$current]: " answer
        printf '%s' "${answer:-$current}"
    else
        read -r -p "$label: " answer
        printf '%s' "$answer"
    fi
}

ask_required() {
    local value
    value="$(ask "$1" "$2")"
    [[ -n "$value" ]] || die "$1 is required"
    printf '%s' "$value"
}

ask_secret() {
    local label="$1" current="$2" answer
    [[ -n "$current" ]] && { printf '%s' "$current"; return; }
    is_interactive || die "$label is required in non-interactive mode"
    read -r -s -p "$label: " answer
    printf '\n' >&2
    printf '%s' "$answer"
}

ask_yes_no() {
    local label="$1" default="${2:-y}" answer
    is_interactive || [[ "$default" == "y" ]]
    read -r -p "$label [${default}/$( [[ "$default" == "y" ]] && printf 'N' || printf 'Y' )]: " answer
    [[ "${answer:-$default}" =~ ^[Yy]$ ]]
}

require_command() { command -v "$1" >/dev/null 2>&1 || die "$1 is required"; }

random_secret() {
    require_command openssl
    openssl rand -hex 32
}

validate_url() {
    [[ "$1" == https://* ]] || die 'APP_URL must start with https://'
    [[ "$1" != *$'\n'* && "$1" != *' '* ]] || die 'APP_URL must not contain spaces or newlines'
}

probe_redis() {
    require_command redis-cli
    local candidate result candidate_password
    local -a candidates=()
    [[ -n "$REDIS_HOST_VALUE" ]] && candidates+=("$REDIS_HOST_VALUE")
    candidates+=(127.0.0.1 localhost)
    for candidate in "${candidates[@]}"; do
        result="$(redis-cli -h "$candidate" -p "$REDIS_PORT_VALUE" ping 2>&1 || true)"
        if [[ "$result" == PONG ]]; then
            REDIS_HOST_VALUE="$candidate"
            REDIS_PASSWORD_VALUE=''
            log "Redis detected at ${REDIS_HOST_VALUE}:${REDIS_PORT_VALUE}"
            return 0
        fi
        if [[ "$result" == *NOAUTH* ]]; then
            candidate_password="$REDIS_PASSWORD_VALUE"
            if [[ -z "$candidate_password" && "$NON_INTERACTIVE" != 1 ]]; then
                candidate_password="$(ask_secret "Redis password for ${candidate}:${REDIS_PORT_VALUE}" '')"
            fi
            if [[ -n "$candidate_password" ]] && redis-cli -h "$candidate" -p "$REDIS_PORT_VALUE" -a "$candidate_password" --no-auth-warning ping 2>/dev/null | grep -qx PONG; then
                REDIS_HOST_VALUE="$candidate"
                REDIS_PASSWORD_VALUE="$candidate_password"
                log "Redis detected at ${REDIS_HOST_VALUE}:${REDIS_PORT_VALUE} with authentication"
                return 0
            fi
        fi
    done
    return 1
}

configure_redis() {
    if probe_redis; then return; fi
    log 'Redis was not found on local candidates; enter the external Redis connection.'
    REDIS_HOST_VALUE="$(ask_required 'Redis host' "$REDIS_HOST_VALUE")"
    REDIS_PORT_VALUE="$(ask 'Redis port' "$REDIS_PORT_VALUE")"
    if [[ -z "$REDIS_PASSWORD_VALUE" && "$NON_INTERACTIVE" != 1 ]]; then
        REDIS_PASSWORD_VALUE="$(ask_secret 'Redis password (leave empty when disabled)' '')"
    fi
    if [[ -n "$REDIS_PASSWORD_VALUE" ]]; then
        redis-cli -h "$REDIS_HOST_VALUE" -p "$REDIS_PORT_VALUE" -a "$REDIS_PASSWORD_VALUE" --no-auth-warning ping 2>/dev/null | grep -qx PONG || die 'Redis authentication check failed'
    else
        redis-cli -h "$REDIS_HOST_VALUE" -p "$REDIS_PORT_VALUE" ping | grep -qx PONG || die 'Redis connectivity check failed'
    fi
    log "Redis connected at ${REDIS_HOST_VALUE}:${REDIS_PORT_VALUE}"
}

configure_database() {
    require_command psql
    DB_HOST_VALUE="$(ask_required 'PostgreSQL host' "$DB_HOST_VALUE")"
    DB_PORT_VALUE="$(ask 'PostgreSQL port' "$DB_PORT_VALUE")"
    DB_DATABASE_VALUE="$(ask_required 'PostgreSQL database' "$DB_DATABASE_VALUE")"
    DB_USERNAME_VALUE="$(ask_required 'PostgreSQL username' "$DB_USERNAME_VALUE")"
    DB_PASSWORD_VALUE="$(ask_secret 'PostgreSQL password' "$DB_PASSWORD_VALUE")"
    DB_SSLMODE_VALUE="$(ask 'PostgreSQL SSL mode (verify-full/require/disable)' "$DB_SSLMODE_VALUE")"
    [[ "$DB_SSLMODE_VALUE" != disable ]] || die 'PostgreSQL SSL mode disable is not allowed by the production deployment'
    PGPASSWORD="$DB_PASSWORD_VALUE" PGSSLMODE="$DB_SSLMODE_VALUE" psql -h "$DB_HOST_VALUE" -p "$DB_PORT_VALUE" -U "$DB_USERNAME_VALUE" -d "$DB_DATABASE_VALUE" -Atqc 'select 1' >/dev/null || die 'PostgreSQL connectivity check failed'
    log "PostgreSQL connected at ${DB_HOST_VALUE}:${DB_PORT_VALUE}/${DB_DATABASE_VALUE}"
}

configure_common() {
    APP_URL_VALUE="$(ask_required 'Public HTTPS URL, for example https://img.example.com' "$APP_URL_VALUE")"
    validate_url "$APP_URL_VALUE"
    CORS_VALUE="$(ask 'CORS allowed origins (comma-separated)' "${CORS_VALUE:-$APP_URL_VALUE}")"
    [[ "$CORS_VALUE" != *'*'* ]] || die 'CORS_ALLOWED_ORIGINS must not contain *'
    configure_database
    configure_redis
}

write_env() {
    local app_key jwt_secret env_dir timestamp backup_path temp_env
    app_key="$(random_secret)"
    jwt_secret="$(random_secret)"
    env_dir="$(dirname -- "$ENV_FILE_VALUE")"
    mkdir -p "$env_dir"
    if [[ -e "$ENV_FILE_VALUE" ]]; then
        ask_yes_no "Overwrite existing environment file $ENV_FILE_VALUE?" n || die "refusing to overwrite $ENV_FILE_VALUE"
        timestamp="$(date -u +%Y%m%d%H%M%S)"
        backup_path="${ENV_FILE_VALUE}.bak.${timestamp}"
        cp -p "$ENV_FILE_VALUE" "$backup_path"
        log "Existing environment backed up to $backup_path"
    fi
    umask 077
    temp_env="$(mktemp "${ENV_FILE_VALUE}.tmp.XXXXXX")"
    {
        printf 'APP_NAME=FastImg\nAPP_ENV=production\nAPP_DEBUG=false\n'
        printf 'APP_URL=%s\nAPP_HOST=0.0.0.0\nAPP_PORT=%s\n' "$APP_URL_VALUE" "$API_PORT_VALUE"
        printf 'APP_KEY=%s\nJWT_SECRET=%s\n' "$app_key" "$jwt_secret"
        printf 'CORS_ALLOWED_ORIGINS=%s\n' "$CORS_VALUE"
        printf 'DB_CONNECTION=postgres\nDB_HOST=%s\nDB_PORT=%s\nDB_DATABASE=%s\nDB_USERNAME=%s\nDB_PASSWORD=%s\nDB_SSLMODE=%s\n' "$DB_HOST_VALUE" "$DB_PORT_VALUE" "$DB_DATABASE_VALUE" "$DB_USERNAME_VALUE" "$DB_PASSWORD_VALUE" "$DB_SSLMODE_VALUE"
        printf 'REDIS_HOST=%s\nREDIS_PORT=%s\nREDIS_PASSWORD=%s\n' "$REDIS_HOST_VALUE" "$REDIS_PORT_VALUE" "$REDIS_PASSWORD_VALUE"
    } > "$temp_env"
    chmod 0640 "$temp_env"
    mv -f "$temp_env" "$ENV_FILE_VALUE"
    log "Generated protected environment file: $ENV_FILE_VALUE"
}

bootstrap_admin() {
    [[ "$BOOTSTRAP_ADMIN" == 1 ]] || return 0
    [[ -n "$ADMIN_EMAIL_VALUE" || "$NON_INTERACTIVE" != 1 ]] || { log 'FASTIMG_ADMIN_EMAIL is empty; skipping administrator bootstrap'; return 0; }
    ADMIN_EMAIL_VALUE="$(ask_required 'Initial administrator email' "$ADMIN_EMAIL_VALUE")"
    ADMIN_NAME_VALUE="$(ask 'Initial administrator name' "$ADMIN_NAME_VALUE")"
    ADMIN_PASSWORD_VALUE="$(ask_secret 'Initial administrator password (12+ characters)' "$ADMIN_PASSWORD_VALUE")"
    [[ ${#ADMIN_PASSWORD_VALUE} -ge 12 ]] || die 'initial administrator password must contain at least 12 characters'
    if [[ "$MODE" == docker ]]; then
        printf '%s\n' "$ADMIN_PASSWORD_VALUE" | docker compose --env-file "$ENV_FILE_VALUE" -f "$SCRIPT_DIR/docker/docker-compose.yml" run --rm -T api artisan admin:bootstrap --email "$ADMIN_EMAIL_VALUE" --name "$ADMIN_NAME_VALUE"
    else
        local app_binary="$INSTALL_ROOT_VALUE/current/backend/fastimg-api"
        [[ -x "$app_binary" ]] || die "deployed API binary not found: $app_binary"
        if [[ "$(id -u)" -eq 0 ]]; then
            printf '%s\n' "$ADMIN_PASSWORD_VALUE" | runuser -u "$APP_USER_VALUE" -- "$app_binary" artisan admin:bootstrap --email "$ADMIN_EMAIL_VALUE" --name "$ADMIN_NAME_VALUE"
        else
            printf '%s\n' "$ADMIN_PASSWORD_VALUE" | "$app_binary" artisan admin:bootstrap --email "$ADMIN_EMAIL_VALUE" --name "$ADMIN_NAME_VALUE"
        fi
    fi
}

deploy_application() {
    if [[ "$MODE" == docker ]]; then
        FASTIMG_ENV_FILE="$ENV_FILE_VALUE" FASTIMG_WEB_PORT="$WEB_PORT_VALUE" bash "$SCRIPT_DIR/docker/deploy.sh"
    else
        [[ "$(id -u)" -eq 0 ]] || die 'source deployment must run as root (use sudo)'
        FASTIMG_SOURCE_ROOT="$REPO_ROOT" FASTIMG_ENV_FILE="$ENV_FILE_VALUE" FASTIMG_INSTALL_ROOT="$INSTALL_ROOT_VALUE" FASTIMG_APP_USER="$APP_USER_VALUE" FASTIMG_API_PORT="$API_PORT_VALUE" bash "$SCRIPT_DIR/linux/deploy.sh"
    fi
}

print_urls() {
    local origin="${APP_URL_VALUE%/}"
    printf '\n'
    log 'Deployment endpoints'
    printf '  member URL: %s/\n' "$origin"
    printf '  admin URL:  %s/admin/\n' "$origin"
    printf '  API URL:    %s/api/\n' "$origin"
    printf '  media URL:  %s/i/{id}?signature=...\n' "$origin"
    printf '\n'
    log 'Baota reverse-proxy targets'
    if [[ "$MODE" == docker ]]; then
        printf '  web/app upstream: http://127.0.0.1:%s\n' "$WEB_PORT_VALUE"
        printf '  API is proxied by the web container under /api/\n'
    else
        printf '  frontend static root: %s/current/web\n' "$INSTALL_ROOT_VALUE"
        printf '  API upstream: http://127.0.0.1:%s\n' "$API_PORT_VALUE"
        printf '  proxy /api/, /i/, /s/, /sitemap.xml and /robots.txt to the API upstream\n'
    fi
    printf '\n'
    log 'Manual production configuration after login'
    printf '  1. TLS certificate, DNS and Baota reverse proxy.\n'
    printf '  2. Email provider and sender address for registration verification.\n'
    printf '  3. Cloudflare Turnstile keys, if registration protection is enabled.\n'
    printf '  4. Payment channels (PayPal, XCash, NOWPayments) and public callbacks.\n'
    printf '  5. R2/OSS/COS credentials, CDN domain, watermark, ads, SEO and sitemap.\n'
    printf '  6. PostgreSQL backup schedule, restore drill and external monitoring.\n'
}

main() {
    require_command curl
    if [[ -z "$MODE" ]]; then MODE="$(ask 'Deployment mode (source/docker)' source)"; fi
    [[ "$MODE" == source || "$MODE" == docker ]] || die 'deployment mode must be source or docker'
    if [[ -z "$ENV_FILE_VALUE" ]]; then
        if [[ "$MODE" == docker ]]; then ENV_FILE_VALUE="$SCRIPT_DIR/docker/fastimg.env"; else ENV_FILE_VALUE="${INSTALL_ROOT_VALUE}/shared/.env"; fi
    fi
    if [[ "$MODE" == docker ]]; then
        WEB_PORT_VALUE="$(ask 'Docker web host port' "$WEB_PORT_VALUE")"
    else
        API_PORT_VALUE="$(ask 'API local port' "$API_PORT_VALUE")"
        INSTALL_ROOT_VALUE="$(ask 'Linux install root' "$INSTALL_ROOT_VALUE")"
    fi
    configure_common
    if [[ "$CHECK_ONLY" == 1 ]]; then
        log 'Configuration checks passed; deployment was not started.'
        print_urls
        return 0
    fi
    write_env
    deploy_application
    bootstrap_admin
    print_urls
}

main "$@"
