#!/bin/sh

# rehearse-observability.sh — disposable failure-mode rehearsal for the
# observability baseline. It builds unique images, starts an isolated copy of
# the tracked Compose stack under a unique project name, and verifies that the
# documented operational signals actually fire:
#
#   1.  healthy startup, liveness, readiness
#   2.  end-to-end request-ID correlation (nginx -> application logs)
#   3.  PostgreSQL pause -> readiness 503, liveness stays 200
#   4.  controlled dependency 5xx visible in structured logs
#   5.  PostgreSQL restore -> readiness recovers, schema intact
#   6.  container restart visible via RestartCount
#   7.  migration failure observable with a stable event name
#   8.  SIGTERM graceful shutdown events and clean exit
#   9.  no secrets present in application logs
#
# All containers, networks, volumes, and images are unique to this run and are
# removed by the cleanup trap. Production and staging resources are never
# touched: the rehearsal uses its own Compose project, images, and volume.

set -eu

umask 077

: "${SPCASE_CONFIRM_OBSERVABILITY_REHEARSAL:?Set SPCASE_CONFIRM_OBSERVABILITY_REHEARSAL=YES}"
if [ "$SPCASE_CONFIRM_OBSERVABILITY_REHEARSAL" != "YES" ]; then
    echo "refusing observability rehearsal without SPCASE_CONFIRM_OBSERVABILITY_REHEARSAL=YES" >&2
    exit 1
fi

for command in docker curl jq; do
    if ! command -v "$command" >/dev/null 2>&1; then
        echo "$command is required for the observability rehearsal" >&2
        exit 1
    fi
done
if ! docker info >/dev/null 2>&1; then
    echo "Docker daemon is unavailable" >&2
    exit 1
fi
if ! docker compose version >/dev/null 2>&1; then
    echo "Docker Compose is unavailable" >&2
    exit 1
fi

script_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
repository_dir=$(CDPATH='' cd -- "$script_dir/.." && pwd)
work_dir=$(mktemp -d "${TMPDIR:-/tmp}/spcase-observability.XXXXXX")
run_id="$(date +%s)-$$-${work_dir##*.}"
safe_run_id=$(printf '%s' "$run_id" | tr '[:upper:]' '[:lower:]' | tr -cd '[:alnum:]')

project="spcase-observability-$safe_run_id"
app_image="spcase-observability-app:$safe_run_id"
migrator_image="spcase-observability-migrator:$safe_run_id"
nginx_image="spcase-observability-nginx:$safe_run_id"
compose_env="$work_dir/rehearsal.env"
database_name="spcase_observability"

random_hex() {
    od -An -N32 -tx1 /dev/urandom | tr -d ' \n'
}

admin_password=$(random_hex)
migrator_password=$(random_hex)
app_password=$(random_hex)
jwt_secret=$(random_hex)
jury_key=$(random_hex)

compose() {
    docker compose --project-name "$project" \
        --file "$repository_dir/docker-compose.yml" \
        --file "$repository_dir/docker-compose.staging.yml" \
        --env-file "$compose_env" "$@"
}

cleanup() {
    exit_status=$?
    cleanup_failed=0
    trap - EXIT HUP INT TERM

    if [ "$exit_status" -ne 0 ] && [ -f "$compose_env" ]; then
        echo "diagnostics: rehearsal failed, capturing disposable stack state" >&2
        compose ps --all >&2 2>/dev/null || true
        for service in app nginx db; do
            echo "diagnostics: last logs for $service" >&2
            compose logs --tail 30 "$service" >&2 2>/dev/null || true
        done
    fi

    echo "cleanup: removing resources created by observability rehearsal $run_id"
    if [ -f "$compose_env" ]; then
        compose down --volumes --remove-orphans >/dev/null 2>&1 || true
    fi
    docker image rm "$app_image" "$migrator_image" "$nginx_image" >/dev/null 2>&1 || true
    if ! rm -rf "$work_dir"; then
        echo "cleanup failed: could not remove work directory: $work_dir" >&2
        cleanup_failed=1
    fi

    if ! docker info >/dev/null 2>&1; then
        echo "cleanup failed: Docker is unavailable; resource removal cannot be verified" >&2
        cleanup_failed=1
    else
        if [ -n "$(docker ps --all --quiet --filter "label=com.docker.compose.project=$project")" ] || \
           [ -n "$(docker network ls --quiet --filter "label=com.docker.compose.project=$project")" ] || \
           [ -n "$(docker volume ls --quiet --filter "label=com.docker.compose.project=$project")" ]; then
            echo "cleanup failed: Compose resources remain for project $project" >&2
            cleanup_failed=1
        fi
        for image in "$app_image" "$migrator_image" "$nginx_image"; do
            if docker image inspect "$image" >/dev/null 2>&1; then
                echo "cleanup failed: image remains: $image" >&2
                cleanup_failed=1
            fi
        done
    fi
    if [ -e "$work_dir" ]; then
        echo "cleanup failed: work directory remains: $work_dir" >&2
        cleanup_failed=1
    fi
    if [ "$cleanup_failed" -ne 0 ] && [ "$exit_status" -eq 0 ]; then
        exit_status=1
    fi
    exit "$exit_status"
}
trap cleanup EXIT
trap 'exit 1' HUP INT TERM

# wait_for_http_status <url> <expected-status> <max-attempts>
wait_for_http_status() {
    url=$1
    expected=$2
    max_attempts=$3
    attempt=0
    while :; do
        attempt=$((attempt + 1))
        status=$(curl --noproxy '*' --silent --output /dev/null --max-time 10 \
            --write-out '%{http_code}' "$url" || true)
        if [ "$status" = "$expected" ]; then
            return 0
        fi
        if [ "$attempt" -ge "$max_attempts" ]; then
            echo "HTTP endpoint did not reach status $expected: $url (last status $status)" >&2
            return 1
        fi
        sleep 1
    done
}

cat > "$compose_env" <<EOF
POSTGRES_ADMIN_PASSWORD=$admin_password
DB_MIGRATOR_USER=spcase_migrator
DB_MIGRATOR_PASSWORD=$migrator_password
DB_APP_USER=spcase_app
DB_APP_PASSWORD=$app_password
DB_NAME=$database_name
APP_DOMAIN=127.0.0.1
CORS_ALLOWED_ORIGINS=https://127.0.0.1
REGISTRATION_DEADLINE=2099-01-01T00:00:00Z
SUBMISSION_DEADLINE=2099-01-02T00:00:00Z
NO_TEAM_TELEGRAM_URL=https://t.me/spcase_rehearsal
JWT_SECRET=$jwt_secret
JURY_REGISTRATION_KEY=$jury_key
SPCASE_APP_IMAGE=$app_image
SPCASE_MIGRATOR_IMAGE=$migrator_image
SPCASE_NGINX_IMAGE=$nginx_image
NGINX_PORT=0
STAGING_DB_PORT=0
EOF
chmod 600 "$compose_env"

echo "phase 1/9: build isolated rehearsal images"
docker build --target runtime --tag "$app_image" "$repository_dir" >/dev/null
docker build --target migrator --tag "$migrator_image" "$repository_dir" >/dev/null
docker build --target nginx --tag "$nginx_image" "$repository_dir" >/dev/null

echo "phase 2/9: start the isolated stack and verify healthy startup"
compose up --detach --wait
nginx_container=$(compose ps --quiet nginx)
app_container=$(compose ps --quiet app)
db_container=$(compose ps --quiet db)
nginx_port=$(docker port "$nginx_container" 8080/tcp | sed -n 's/.*://p')
runtime_url="http://127.0.0.1:$nginx_port"
wait_for_http_status "$runtime_url/api/v1/health/live" 200 30
wait_for_http_status "$runtime_url/api/v1/health/ready" 200 30
if ! docker logs "$app_container" 2>&1 | grep -q '"event":"http_server_started"'; then
    echo "http_server_started event missing from application logs" >&2
    exit 1
fi
log_driver=$(docker inspect --format '{{.HostConfig.LogConfig.Type}}' "$app_container")
log_max_size=$(docker inspect --format '{{index .HostConfig.LogConfig.Config "max-size"}}' "$app_container")
log_max_file=$(docker inspect --format '{{index .HostConfig.LogConfig.Config "max-file"}}' "$app_container")
if [ "$log_driver" != "json-file" ] || [ "$log_max_size" != "10m" ] || [ "$log_max_file" != "5" ]; then
    echo "application container is missing bounded log rotation: driver=$log_driver max-size=$log_max_size max-file=$log_max_file" >&2
    exit 1
fi

echo "phase 3/9: verify end-to-end request correlation"
edge_request_id=$(curl --noproxy '*' --silent --output /dev/null --dump-header - \
    "$runtime_url/api/v1/health/live" | sed -n 's/^[Xx]-[Rr]equest-[Ii][Dd]: //p' | tr -d '\r')
if [ -z "$edge_request_id" ]; then
    echo "nginx response is missing the X-Request-ID header" >&2
    exit 1
fi
if ! docker logs "$app_container" 2>&1 | grep -F -q "\"request_id\":\"$edge_request_id\""; then
    echo "application logs do not contain the nginx-propagated request ID" >&2
    exit 1
fi
custom_id="rehearsal-$safe_run_id-0001"
direct_response=$(docker exec "$nginx_container" \
    wget -q -S -O /dev/null --header "X-Request-ID: $custom_id" \
    http://app:8000/api/v1/health/live 2>&1 || true)
if ! printf '%s' "$direct_response" | grep -i -q "x-request-id: $custom_id"; then
    echo "application did not accept a valid inbound request ID" >&2
    exit 1
fi
oversized_id="rehearsal-$safe_run_id-$(printf 'a%.0s' $(seq 1 80))"
direct_response=$(docker exec "$nginx_container" \
    wget -q -S -O /dev/null --header "X-Request-ID: $oversized_id" \
    http://app:8000/api/v1/health/live 2>&1 || true)
returned_id=$(printf '%s' "$direct_response" | sed -n 's/^[[:space:]]*[Xx]-[Rr]equest-[Ii][Dd]: //p' | tr -d '\r ')
if [ -z "$returned_id" ] || [ "$returned_id" = "$oversized_id" ]; then
    echo "application did not replace an oversized inbound request ID" >&2
    exit 1
fi
if ! printf '%s' "$returned_id" | grep -qE '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'; then
    echo "replacement request ID is not a UUID: $returned_id" >&2
    exit 1
fi

echo "phase 4/9: pause PostgreSQL and verify readiness/liveness semantics"
docker pause "$db_container" >/dev/null
wait_for_http_status "$runtime_url/api/v1/health/ready" 503 60
wait_for_http_status "$runtime_url/api/v1/health/live" 200 10
if ! docker logs "$app_container" 2>&1 | grep -q '"event":"database_readiness_failed"'; then
    echo "database_readiness_failed event missing from application logs" >&2
    exit 1
fi
# Through nginx the app's 503 is intentionally replaced by the edge JSON
# contract (nginx.conf @api_unavailable); the app-level NOT_READY body contract
# is covered by unit tests in internal/delivery/http/v1/public_test.go.
readiness_body=$(curl --noproxy '*' --silent --max-time 10 "$runtime_url/api/v1/health/ready")
case "$readiness_body" in
    *'"SERVICE_UNAVAILABLE"'*) ;;
    *)
        echo "edge readiness failure body does not use the stable SERVICE_UNAVAILABLE contract: $readiness_body" >&2
        exit 1
        ;;
esac

echo "phase 5/9: verify a controlled dependency failure is visible as a 5xx"
login_status=""
for attempt in 1 2 3; do
    login_status=$(curl --noproxy '*' --silent --output /dev/null --max-time 25 \
        --write-out '%{http_code}' --request POST --header 'Content-Type: application/json' \
        --data '{"email":"rehearsal@example.test","password":"rehearsal-password"}' \
        "$runtime_url/api/v1/auth/login" || true)
    if [ "$login_status" = "500" ]; then
        break
    fi
    sleep 2
done
if [ "$login_status" != "500" ]; then
    echo "login during PostgreSQL outage returned HTTP ${login_status:-timeout}, expected 500" >&2
    exit 1
fi
if ! docker logs "$app_container" 2>&1 | grep -q '"event":"http_request_completed"[^}]*"status":500'; then
    echo "no http_request_completed event with status 500 found in application logs" >&2
    exit 1
fi

echo "phase 6/9: restore PostgreSQL and verify readiness recovery"
docker unpause "$db_container" >/dev/null
wait_for_http_status "$runtime_url/api/v1/health/ready" 200 60
migration_version=$(docker exec "$db_container" psql --no-psqlrc --tuples-only --no-align \
    --username postgres --dbname "$database_name" \
    --command 'SELECT MAX(version_id) FILTER (WHERE is_applied) FROM goose_db_version' | tr -d '[:space:]')
if [ "$migration_version" != "5" ]; then
    echo "migration version is $migration_version after database restore, expected 5" >&2
    exit 1
fi

echo "phase 7/9: restart the application container and verify restart visibility"
# Note: Docker RestartCount only increments for daemon restarts after a
# genuine process crash, not for API-initiated stop/kill/restart. The
# deterministic operator-visible signals of any restart are a fresh StartedAt
# and a repeated http_server_started event in the application logs.
started_before=$(docker inspect --format '{{.State.StartedAt}}' "$app_container")
starts_logged_before=$(docker logs "$app_container" 2>&1 | grep -c '"event":"http_server_started"' || true)
docker restart "$app_container" >/dev/null
wait_for_http_status "$runtime_url/api/v1/health/ready" 200 90
started_after=$(docker inspect --format '{{.State.StartedAt}}' "$app_container")
starts_logged_after=$(docker logs "$app_container" 2>&1 | grep -c '"event":"http_server_started"' || true)
if [ "$started_after" = "$started_before" ] || [ "$starts_logged_after" -le "$starts_logged_before" ]; then
    echo "application restart is not visible: StartedAt $started_before -> $started_after, startup events $starts_logged_before -> $starts_logged_after" >&2
    exit 1
fi

echo "phase 8/9: verify migration failure observability in an isolated run"
backend_network=$(docker network ls --quiet --filter "label=com.docker.compose.project=$project" \
    --filter "name=backend" | head -n 1)
if [ -z "$backend_network" ]; then
    echo "rehearsal backend network not found" >&2
    exit 1
fi
set +e
docker run --rm --network "$backend_network" \
    --env DB_HOST=db --env DB_PORT=5432 --env DB_NAME="$database_name" \
    --env DB_USER=spcase_migrator --env DB_PASSWORD="definitely-wrong-$safe_run_id" \
    "$migrator_image" > "$work_dir/migration-failure.log" 2>&1
migration_status=$?
set -e
if [ "$migration_status" -eq 0 ]; then
    echo "migrator unexpectedly succeeded with invalid credentials" >&2
    exit 1
fi
if ! grep -q 'migration_failed' "$work_dir/migration-failure.log"; then
    echo "migration_failed event missing from migrator output" >&2
    exit 1
fi

echo "phase 9/9: verify graceful shutdown and absence of secrets in logs"
docker stop "$app_container" >/dev/null
app_exit_code=$(docker inspect --format '{{.State.ExitCode}}' "$app_container")
if [ "$app_exit_code" != "0" ]; then
    echo "application exited with code $app_exit_code after SIGTERM, expected 0" >&2
    exit 1
fi
for event in graceful_shutdown_started graceful_shutdown_completed database_pool_closed; do
    if ! docker logs "$app_container" 2>&1 | grep -q "\"event\":\"$event\""; then
        echo "$event event missing from application logs" >&2
        exit 1
    fi
done
for secret in "$app_password" "$migrator_password" "$admin_password" "$jwt_secret" "$jury_key"; do
    if docker logs "$app_container" 2>&1 | grep -F -q "$secret"; then
        echo "a secret value was found in application logs" >&2
        exit 1
    fi
    if docker logs "$nginx_container" 2>&1 | grep -F -q "$secret"; then
        echo "a secret value was found in nginx logs" >&2
        exit 1
    fi
done

printf '%s\n' \
    "observability rehearsal result: PASS" \
    "healthy startup, liveness and readiness verified" \
    "request correlation verified end-to-end (nginx request ID present in application logs)" \
    "PostgreSQL pause: readiness 503 with database_readiness_failed events; liveness stayed 200" \
    "controlled dependency failure produced a logged http_request_completed status 500" \
    "PostgreSQL restore: readiness recovered, migration version 5 intact" \
    "application restart visible: fresh StartedAt and startup events $starts_logged_before -> $starts_logged_after" \
    "migration failure observable with migration_failed event and nonzero exit" \
    "SIGTERM: graceful_shutdown_started/completed and database_pool_closed logged, exit code 0" \
    "no rehearsal secrets found in application or nginx logs" \
    "cleanup will remove only resources created by run $run_id"
