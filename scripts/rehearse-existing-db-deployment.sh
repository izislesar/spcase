#!/bin/sh

set -eu

umask 077

: "${SPCASE_CONFIRM_EXISTING_DB_DEPLOYMENT_REHEARSAL:?Set SPCASE_CONFIRM_EXISTING_DB_DEPLOYMENT_REHEARSAL=YES}"
if [ "$SPCASE_CONFIRM_EXISTING_DB_DEPLOYMENT_REHEARSAL" != "YES" ]; then
    echo "refusing deployment rehearsal without SPCASE_CONFIRM_EXISTING_DB_DEPLOYMENT_REHEARSAL=YES" >&2
    exit 1
fi

for command in docker go curl jq sha256sum psql; do
    if ! command -v "$command" >/dev/null 2>&1; then
        echo "$command is required for the existing-database deployment rehearsal" >&2
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
work_dir=$(mktemp -d "${TMPDIR:-/tmp}/spcase-existing-db-deployment.XXXXXX")
run_id="$(date +%s)-$$-${work_dir##*.}"
safe_run_id=$(printf '%s' "$run_id" | tr '[:upper:]' '[:lower:]' | tr -cd '[:alnum:]')

source_network="spcase-deployment-source-$run_id"
source_volume="spcase-deployment-source-$run_id"
source_db_container="spcase-deployment-source-db-$run_id"
source_app_container="spcase-deployment-source-app-$run_id"
source_database="spcase_legacy"

target_project="spcase-deployment-target-$safe_run_id"
target_volume="${target_project}_postgres_data"
target_restore_network="spcase-deployment-restore-$run_id"
target_restore_container="spcase-deployment-restore-db-$run_id"
target_database="spcase_restored"

rollback_network="spcase-deployment-rollback-$run_id"
rollback_volume="spcase-deployment-rollback-$run_id"
rollback_container="spcase-deployment-rollback-db-$run_id"
rollback_database="spcase_rollback"

app_image="spcase-deployment-app:$run_id"
migrator_image="spcase-deployment-migrator:$run_id"
nginx_image="spcase-deployment-nginx:$run_id"
postgres_image="postgres:16-bookworm@sha256:92620daddcd947f8d5ab5ba66e848702fe443d87fed30c4cea8e389fd78dfc55"
backup_file="$work_dir/spcase-legacy.dump"
compose_env="$work_dir/target.env"

legacy_role="spcase_legacy"
legacy_password="Legacy-${safe_run_id}-Db9!"
admin_password="Admin-${safe_run_id}-Pg8!"
migrator_password="Migrator-${safe_run_id}-Db7!"
app_password="Runtime-${safe_run_id}-Db6!"
account_password="Account-${safe_run_id}-User5!"
jwt_secret="jwt-${safe_run_id}-9Aq!2Ws#3Ed\$4Rf%5Tg&6Yh*7Uj(8Ik)"
jury_key="jury-${safe_run_id}-8Zx!7Cv#6Bn\$5Mm%4As&3Df*2Gh(1Jk)"

compose() {
    docker compose --project-name "$target_project" \
        --file "$repository_dir/docker-compose.yml" \
        --file "$repository_dir/docker-compose.staging.yml" \
        --env-file "$compose_env" "$@"
}

cleanup() {
    exit_status=$?
    cleanup_failed=0
    trap - EXIT HUP INT TERM

    echo "cleanup: removing resources created by rehearsal $run_id"
    if [ -f "$compose_env" ]; then
        compose down --volumes --remove-orphans >/dev/null 2>&1 || true
    fi
    docker rm --force \
        "$source_app_container" "$source_db_container" \
        "$target_restore_container" "$rollback_container" >/dev/null 2>&1 || true
    docker volume rm "$source_volume" "$target_volume" "$rollback_volume" >/dev/null 2>&1 || true
    docker network rm "$source_network" "$target_restore_network" "$rollback_network" >/dev/null 2>&1 || true
    docker image rm "$app_image" "$migrator_image" "$nginx_image" >/dev/null 2>&1 || true
    if ! rm -rf "$work_dir"; then
        echo "cleanup failed: could not remove work directory: $work_dir" >&2
        cleanup_failed=1
    fi

    if ! docker info >/dev/null 2>&1; then
        echo "cleanup failed: Docker is unavailable; resource removal cannot be verified" >&2
        cleanup_failed=1
    else
        for container in "$source_app_container" "$source_db_container" \
            "$target_restore_container" "$rollback_container"; do
            if docker container inspect "$container" >/dev/null 2>&1; then
                echo "cleanup failed: container remains: $container" >&2
                cleanup_failed=1
            fi
        done
        for volume in "$source_volume" "$target_volume" "$rollback_volume"; do
            if docker volume inspect "$volume" >/dev/null 2>&1; then
                echo "cleanup failed: volume remains: $volume" >&2
                cleanup_failed=1
            fi
        done
        for network in "$source_network" "$target_restore_network" "$rollback_network"; do
            if docker network inspect "$network" >/dev/null 2>&1; then
                echo "cleanup failed: network remains: $network" >&2
                cleanup_failed=1
            fi
        done
        for image in "$app_image" "$migrator_image" "$nginx_image"; do
            if docker image inspect "$image" >/dev/null 2>&1; then
                echo "cleanup failed: image remains: $image" >&2
                cleanup_failed=1
            fi
        done
        if [ -n "$(docker ps --all --quiet --filter "label=com.docker.compose.project=$target_project")" ] || \
           [ -n "$(docker network ls --quiet --filter "label=com.docker.compose.project=$target_project")" ] || \
           [ -n "$(docker volume ls --quiet --filter "label=com.docker.compose.project=$target_project")" ]; then
            echo "cleanup failed: Compose resources remain for project $target_project" >&2
            cleanup_failed=1
        fi
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

wait_for_postgres() {
    container=$1
    database=$2
    database_user=$3
    attempt=0
    consecutive_successes=0
    while [ "$consecutive_successes" -lt 2 ]; do
        attempt=$((attempt + 1))
        if [ "$attempt" -gt 90 ]; then
            echo "PostgreSQL did not become ready in $container" >&2
            return 1
        fi
        if docker exec "$container" pg_isready --username "$database_user" --dbname "$database" >/dev/null 2>&1 && \
           docker exec "$container" psql --no-psqlrc --tuples-only --no-align \
               --username "$database_user" --dbname "$database" --command 'SELECT 1' >/dev/null 2>&1; then
            consecutive_successes=$((consecutive_successes + 1))
        else
            consecutive_successes=0
        fi
        sleep 1
    done
}

wait_for_http() {
    url=$1
    attempt=0
    while :; do
        attempt=$((attempt + 1))
        status=$(curl --noproxy '*' --silent --output /dev/null --write-out '%{http_code}' "$url" || true)
        if [ "$status" = "200" ]; then
            return 0
        fi
        if [ "$attempt" -gt 90 ]; then
            echo "HTTP endpoint did not become ready: $url (last status $status)" >&2
            return 1
        fi
        sleep 1
    done
}

database_fingerprint() {
    container=$1
    database=$2
    database_user=$3
    docker exec "$container" psql --no-psqlrc --tuples-only --no-align --set=ON_ERROR_STOP=1 \
        --username "$database_user" --dbname "$database" --command "
            SELECT concat_ws('|',
                COALESCE((SELECT MAX(version_id) FILTER (WHERE is_applied) FROM public.goose_db_version), 0),
                (SELECT COUNT(*) FROM public.users),
                (SELECT COUNT(*) FROM public.teams),
                (SELECT COUNT(*) FROM public.team_members),
                (SELECT COUNT(*) FROM public.submissions),
                (SELECT COUNT(*) FROM public.evaluations),
                (SELECT COUNT(*) FROM public.evaluation_state_events),
                md5(COALESCE((SELECT string_agg(id::text || ':' || email || ':' || role::text || ':' || auth_version::text,
                    ',' ORDER BY id) FROM public.users), '')),
                md5(COALESCE((SELECT string_agg(id::text || ':' || name || ':' || invite_code || ':' || captain_id::text,
                    ',' ORDER BY id) FROM public.teams), '')),
                md5(COALESCE((SELECT string_agg(team_id::text || ':' || user_id::text,
                    ',' ORDER BY team_id, user_id) FROM public.team_members), '')),
                md5(COALESCE((SELECT string_agg(id::text || ':' || team_id::text || ':' || solution_url,
                    ',' ORDER BY id) FROM public.submissions), '')),
                md5(COALESCE((SELECT string_agg(id::text || ':' || jury_id::text || ':' || team_id::text || ':' ||
                    criterion_id::text || ':' || score::text, ',' ORDER BY id) FROM public.evaluations), ''))
            );
        "
}

preserved_fingerprint() {
    container=$1
    database=$2
    database_user=${3:-postgres}
    docker exec "$container" psql --no-psqlrc --tuples-only --no-align --set=ON_ERROR_STOP=1 \
        --username "$database_user" --dbname "$database" --command "
            SELECT md5(concat_ws('|',
                (SELECT string_agg(id::text || ':' || email || ':' || role::text,
                    ',' ORDER BY email) FROM users WHERE email IN (
                        'admin@rehearsal.test', 'captain@rehearsal.test',
                        'member@rehearsal.test', 'jury@rehearsal.test'
                    )),
                (SELECT string_agg(id::text || ':' || name || ':' || invite_code || ':' || captain_id::text,
                    ',' ORDER BY name) FROM teams WHERE name = 'Existing Rehearsal Team'),
                (SELECT string_agg(team_id::text || ':' || user_id::text,
                    ',' ORDER BY user_id) FROM team_members WHERE team_id =
                        (SELECT id FROM teams WHERE name = 'Existing Rehearsal Team')),
                (SELECT string_agg(id::text || ':' || team_id::text,
                    ',' ORDER BY id) FROM submissions WHERE team_id =
                        (SELECT id FROM teams WHERE name = 'Existing Rehearsal Team')),
                (SELECT string_agg(id::text || ':' || jury_id::text || ':' || team_id::text || ':' || criterion_id::text,
                    ',' ORDER BY id) FROM evaluations WHERE team_id =
                        (SELECT id FROM teams WHERE name = 'Existing Rehearsal Team'))
            ));
        "
}

ownership_summary() {
    container=$1
    database=$2
    docker exec "$container" psql --no-psqlrc --tuples-only --no-align --set=ON_ERROR_STOP=1 \
        --username postgres --dbname "$database" --command "
            SELECT concat_ws('|',
                pg_get_userbyid((SELECT datdba FROM pg_database WHERE datname = current_database())),
                pg_get_userbyid((SELECT nspowner FROM pg_namespace WHERE nspname = 'public')),
                (SELECT string_agg(DISTINCT pg_get_userbyid(relowner), ',' ORDER BY pg_get_userbyid(relowner))
                 FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
                 WHERE n.nspname = 'public' AND c.relkind IN ('r','S'))
            );
        "
}

api_request() {
    expected=$1
    method=$2
    url=$3
    cookie=${4:-}
    data=${5:-}
    response_file=$6
    set -- --noproxy '*' --silent --show-error --output "$response_file" --write-out '%{http_code}' \
        --request "$method" --header 'Content-Type: application/json'
    if [ -n "$cookie" ]; then
        set -- "$@" --header "Cookie: access_token=$cookie"
    fi
    if [ -n "$data" ]; then
        set -- "$@" --data "$data"
    fi
    status=$(curl "$@" "$url")
    if [ "$status" != "$expected" ]; then
        echo "unexpected HTTP status for $method $url: got $status, want $expected" >&2
        jq -c '{error: .error.code // "unexpected_response"}' "$response_file" >&2 2>/dev/null || true
        return 1
    fi
}

login_token() {
    endpoint=$1
    email=$2
    password=$3
    response_file=$4
    header_file=$5
    status=$(curl --noproxy '*' --silent --show-error --output "$response_file" --dump-header "$header_file" \
        --write-out '%{http_code}' --request POST --header 'Content-Type: application/json' \
        --data "{\"email\":\"$email\",\"password\":\"$password\"}" "$endpoint")
    if [ "$status" != "200" ]; then
        echo "login failed with HTTP $status" >&2
        return 1
    fi
    token=$(sed -n 's/^[Ss]et-[Cc]ookie: access_token=\([^;]*\).*/\1/p' "$header_file" | head -n 1)
    if [ -z "$token" ]; then
        echo "login did not return an access token cookie" >&2
        return 1
    fi
    printf '%s' "$token"
}

echo "phase 1/12: build isolated rehearsal images"
docker build --target runtime --tag "$app_image" "$repository_dir" >/dev/null
docker build --target migrator --tag "$migrator_image" "$repository_dir" >/dev/null
docker build --target nginx --tag "$nginx_image" "$repository_dir" >/dev/null

echo "phase 2/12: create a legacy source database at production migration version 2"
docker network create "$source_network" >/dev/null
docker volume create "$source_volume" >/dev/null
docker run --detach --name "$source_db_container" --network "$source_network" \
    --env POSTGRES_DB="$source_database" --env POSTGRES_USER="$legacy_role" \
    --env POSTGRES_PASSWORD="$legacy_password" --volume "$source_volume:/var/lib/postgresql/data" \
    "$postgres_image" >/dev/null
wait_for_postgres "$source_db_container" "$source_database" "$legacy_role"

mkdir "$work_dir/legacy-migrations"
cp "$repository_dir/migrations/00001_init_schema.sql" "$work_dir/legacy-migrations/"
cp "$repository_dir/migrations/00002_add_indexes.sql" "$work_dir/legacy-migrations/"
chmod 711 "$work_dir"
chmod 755 "$work_dir/legacy-migrations"
chmod 644 "$work_dir/legacy-migrations"/*.sql
docker run --rm --network "$source_network" --env PGPASSWORD="$legacy_password" \
    --volume "$work_dir/legacy-migrations:/migrations:ro" --entrypoint /usr/local/bin/goose \
    "$migrator_image" -dir /migrations postgres \
    "postgres://$legacy_role@$source_db_container:5432/$source_database?sslmode=disable" up >/dev/null

printf '%s\n' "$account_password" | docker run --rm --interactive --network "$source_network" \
    --env DB_HOST="$source_db_container" --env DB_PORT=5432 --env DB_NAME="$source_database" \
    --env DB_USER="$legacy_role" --env DB_PASSWORD="$legacy_password" \
    --env DB_STATEMENT_TIMEOUT=15s --env DB_LOCK_TIMEOUT=5s --env PORT=8000 \
    --env APP_DOMAIN=127.0.0.1 --env REGISTRATION_DEADLINE=2099-01-01T00:00:00Z \
    --env CORS_ALLOWED_ORIGINS=https://127.0.0.1 \
    --env SUBMISSION_DEADLINE=2099-01-02T00:00:00Z \
    --env NO_TEAM_TELEGRAM_URL=https://t.me/spcase_rehearsal \
    --env JWT_SECRET="$jwt_secret" --env JURY_REGISTRATION_KEY="$jury_key" \
    --entrypoint /app/admin-bootstrap "$app_image" \
    -full-name "Existing Administrator" -email "admin@rehearsal.test" >/dev/null
docker run --detach --name "$source_app_container" --network "$source_network" --publish 127.0.0.1::8000 \
    --env DB_HOST="$source_db_container" --env DB_PORT=5432 --env DB_NAME="$source_database" \
    --env DB_USER="$legacy_role" --env DB_PASSWORD="$legacy_password" \
    --env DB_STATEMENT_TIMEOUT=15s --env DB_LOCK_TIMEOUT=5s --env PORT=8000 \
    --env APP_DOMAIN=127.0.0.1 --env REGISTRATION_DEADLINE=2099-01-01T00:00:00Z \
    --env CORS_ALLOWED_ORIGINS=https://127.0.0.1 \
    --env SUBMISSION_DEADLINE=2099-01-02T00:00:00Z \
    --env NO_TEAM_TELEGRAM_URL=https://t.me/spcase_rehearsal \
    --env JWT_SECRET="$jwt_secret" --env JURY_REGISTRATION_KEY="$jury_key" "$app_image" >/dev/null
source_port=$(docker port "$source_app_container" 8000/tcp | sed -n 's/.*://p')
source_url="http://127.0.0.1:$source_port/api/v1"
wait_for_http "$source_url/health/ready"

echo "phase 3/12: create representative source state through application APIs"
captain_cookie="$work_dir/captain.cookie"
member_cookie="$work_dir/member.cookie"
jury_cookie="$work_dir/jury.cookie"
response="$work_dir/response.json"
api_request 201 POST "$source_url/auth/register" "" \
    "{\"full_name\":\"Existing Captain\",\"university\":\"University\",\"email\":\"captain@rehearsal.test\",\"telegram\":\"@captain\",\"password\":\"$account_password\"}" "$response"
captain_token=$(login_token "$source_url/auth/login" "captain@rehearsal.test" "$account_password" "$response" "$captain_cookie")
api_request 201 POST "$source_url/auth/register" "" \
    "{\"full_name\":\"Existing Member\",\"university\":\"University\",\"email\":\"member@rehearsal.test\",\"telegram\":\"@member\",\"password\":\"$account_password\"}" "$response"
member_token=$(login_token "$source_url/auth/login" "member@rehearsal.test" "$account_password" "$response" "$member_cookie")
api_request 201 POST "$source_url/team/create" "$captain_token" '{"name":"Existing Rehearsal Team"}' "$response"
team_id=$(jq -er '.id' "$response")
invite_code=$(jq -er '.invite_code' "$response")
api_request 200 POST "$source_url/team/join" "$member_token" "{\"invite_code\":\"$invite_code\"}" "$response"
api_request 200 POST "$source_url/team/submit" "$captain_token" \
    '{"solution_url":"https://example.test/existing-solution"}' "$response"
api_request 201 POST "$source_url/jury/register" "" \
    "{\"secret_key\":\"$jury_key\",\"full_name\":\"Existing Jury\",\"email\":\"jury@rehearsal.test\",\"password\":\"$account_password\"}" "$response"
jury_token=$(login_token "$source_url/jury/login" "jury@rehearsal.test" "$account_password" "$response" "$jury_cookie")
api_request 200 POST "$source_url/jury/evaluations" "$jury_token" \
    "{\"team_id\":\"$team_id\",\"scores\":[{\"criterion_id\":1,\"score\":7},{\"criterion_id\":2,\"score\":8},{\"criterion_id\":3,\"score\":9},{\"criterion_id\":4,\"score\":6},{\"criterion_id\":5,\"score\":8},{\"criterion_id\":6,\"score\":7}]}" "$response"

source_fingerprint=$(database_fingerprint "$source_db_container" "$source_database" "$legacy_role")
source_preserved=$(preserved_fingerprint "$source_db_container" "$source_database" "$legacy_role")
source_ownership=$(docker exec "$source_db_container" psql --no-psqlrc --tuples-only --no-align \
    --username "$legacy_role" --dbname "$source_database" --command "
        SELECT concat_ws('|',
            pg_get_userbyid((SELECT datdba FROM pg_database WHERE datname = current_database())),
            pg_get_userbyid((SELECT nspowner FROM pg_namespace WHERE nspname = 'public')),
            (SELECT string_agg(DISTINCT pg_get_userbyid(relowner), ',' ORDER BY pg_get_userbyid(relowner))
             FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
             WHERE n.nspname = 'public' AND c.relkind IN ('r','S'))
        );")
printf 'source fingerprint: %s\nsource ownership: %s\n' "$source_fingerprint" "$source_ownership"

echo "phase 4/12: create and verify a new custom-format backup"
docker exec "$source_db_container" pg_dump --format=custom --username "$legacy_role" \
    --dbname "$source_database" --file=/tmp/spcase-legacy.dump
docker cp "$source_db_container:/tmp/spcase-legacy.dump" "$backup_file" >/dev/null
test -s "$backup_file"
docker run --rm --volume "$backup_file:/backup/spcase.dump:ro" "$postgres_image" \
    pg_restore --list /backup/spcase.dump >/dev/null
backup_size=$(wc -c < "$backup_file" | tr -d ' ')
backup_checksum=$(sha256sum "$backup_file" | awk '{print $1}')
if [ "$backup_size" -lt 1024 ] || [ -z "$backup_checksum" ]; then
    echo "backup artifact failed size or checksum validation" >&2
    exit 1
fi
printf 'backup verified: custom format, %s bytes, sha256 %s\n' "$backup_size" "$backup_checksum"

echo "phase 5/12: restore backup into an independent target volume"
docker network create "$target_restore_network" >/dev/null
docker volume create --label "com.docker.compose.project=$target_project" \
    --label com.docker.compose.volume=postgres_data "$target_volume" >/dev/null
docker run --detach --name "$target_restore_container" --network "$target_restore_network" \
    --publish 127.0.0.1::5432 --env POSTGRES_DB="$target_database" --env POSTGRES_USER=postgres \
    --env POSTGRES_PASSWORD="$admin_password" --volume "$target_volume:/var/lib/postgresql/data" \
    "$postgres_image" >/dev/null
wait_for_postgres "$target_restore_container" "$target_database" postgres
docker exec --interactive --env LEGACY_PASSWORD="$legacy_password" "$target_restore_container" psql \
    --no-psqlrc --set=ON_ERROR_STOP=1 --username postgres --dbname "$target_database" >/dev/null <<'SQL'
\getenv legacy_password LEGACY_PASSWORD
SELECT format(
    'CREATE ROLE spcase_legacy LOGIN PASSWORD %L NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS',
    :'legacy_password'
) \gexec
ALTER DATABASE spcase_restored OWNER TO spcase_legacy;
SQL
docker cp "$backup_file" "$target_restore_container:/tmp/spcase-legacy.dump" >/dev/null
docker exec "$target_restore_container" pg_restore --exit-on-error --no-owner --role="$legacy_role" \
    --username postgres --dbname "$target_database" /tmp/spcase-legacy.dump

restored_fingerprint=$(database_fingerprint "$target_restore_container" "$target_database" postgres)
restored_ownership=$(ownership_summary "$target_restore_container" "$target_database")
if [ "$restored_fingerprint" != "$source_fingerprint" ] || [ "$restored_ownership" != "$source_ownership" ]; then
    echo "restored target does not match source data or legacy ownership" >&2
    exit 1
fi
printf 'restore verified: fingerprint %s, ownership %s\n' "$restored_fingerprint" "$restored_ownership"

target_port=$(docker port "$target_restore_container" 5432/tcp | sed -n 's/.*://p')
run_cutover() {
    PGHOST=127.0.0.1 PGPORT="$target_port" POSTGRES_ADMIN_USER=postgres \
    POSTGRES_ADMIN_PASSWORD="$admin_password" DB_MIGRATOR_PASSWORD="$migrator_password" \
    DB_APP_PASSWORD="$app_password" SPCASE_CUTOVER_DATABASE="$target_database" \
    SPCASE_LEGACY_DB_ROLE="$legacy_role" SPCASE_CONFIRM_EXISTING_DB_CUTOVER=YES \
        "$repository_dir/scripts/cutover-postgres-roles.sh"
}

echo "phase 6/12: run ownership and ACL cutover twice"
run_cutover
first_cutover_fingerprint=$(database_fingerprint "$target_restore_container" "$target_database" postgres)
first_cutover_ownership=$(ownership_summary "$target_restore_container" "$target_database")
run_cutover
second_cutover_fingerprint=$(database_fingerprint "$target_restore_container" "$target_database" postgres)
second_cutover_ownership=$(ownership_summary "$target_restore_container" "$target_database")
if [ "$first_cutover_fingerprint" != "$source_fingerprint" ] || \
   [ "$second_cutover_fingerprint" != "$source_fingerprint" ] || \
   [ "$first_cutover_ownership" != "$second_cutover_ownership" ] || \
   [ "$second_cutover_ownership" != "spcase_migrator|spcase_migrator|spcase_migrator" ]; then
    echo "cutover changed data or failed ownership/idempotence validation" >&2
    exit 1
fi

echo "phase 7/12: apply production migrations as spcase_migrator and verify no pending work"
docker run --rm --network "$target_restore_network" --env DB_HOST="$target_restore_container" \
    --env DB_PORT=5432 --env DB_NAME="$target_database" --env DB_USER=spcase_migrator \
    --env DB_PASSWORD="$migrator_password" "$migrator_image" >/dev/null
docker run --rm --network "$target_restore_network" --env DB_HOST="$target_restore_container" \
    --env DB_PORT=5432 --env DB_NAME="$target_database" --env DB_USER=spcase_migrator \
    --env DB_PASSWORD="$migrator_password" "$migrator_image" >/dev/null
version=$(docker exec --env PGPASSWORD="$migrator_password" "$target_restore_container" psql \
    --host=127.0.0.1 --username=spcase_migrator --dbname="$target_database" --no-psqlrc \
    --tuples-only --no-align --command 'SELECT MAX(version_id) FILTER (WHERE is_applied) FROM goose_db_version')
if [ "$version" != "5" ]; then
    echo "target migration version is $version, expected 5" >&2
    exit 1
fi
if [ "$(preserved_fingerprint "$target_restore_container" "$target_database")" != "$source_preserved" ]; then
    echo "preserved application records changed during cutover or migration" >&2
    exit 1
fi

echo "phase 8/12: run the separate-role integration suite and explicit ACL probes"
(
    cd "$repository_dir"
    SPCASE_TEST_MIGRATOR_DATABASE_URL="postgres://spcase_migrator:$migrator_password@127.0.0.1:$target_port/$target_database?sslmode=disable" \
    SPCASE_TEST_APP_DATABASE_URL="postgres://spcase_app:$app_password@127.0.0.1:$target_port/$target_database?sslmode=disable" \
        go test -race -count=1 -tags=integration ./internal/...
)
docker exec --env PGPASSWORD="$app_password" "$target_restore_container" psql --host=127.0.0.1 \
    --username=spcase_app --dbname="$target_database" --no-psqlrc --set=ON_ERROR_STOP=1 \
    --command "BEGIN; SELECT COUNT(*) FROM users; INSERT INTO users
        (full_name, university, email, telegram, password_hash, role)
        VALUES ('ACL Probe','University','acl-probe@rehearsal.invalid','@acl_probe','hash','USER');
        UPDATE users SET full_name=full_name WHERE email='acl-probe@rehearsal.invalid';
        ROLLBACK;" >/dev/null
for denied_sql in \
    'CREATE TABLE public.rehearsal_forbidden(id integer)' \
    'ALTER TABLE public.users ADD COLUMN rehearsal_forbidden integer' \
    'DROP TABLE public.users' \
    'TRUNCATE public.users' \
    'ALTER SCHEMA public OWNER TO spcase_app' \
    'SELECT * FROM public.goose_db_version' \
    "SELECT nextval('public.goose_db_version_id_seq')"; do
    if docker exec --env PGPASSWORD="$app_password" "$target_restore_container" psql --host=127.0.0.1 \
        --username=spcase_app --dbname="$target_database" --no-psqlrc --set=ON_ERROR_STOP=1 \
        --command "$denied_sql" >/dev/null 2>&1; then
        echo "spcase_app unexpectedly executed a forbidden database operation" >&2
        exit 1
    fi
done

cat > "$compose_env" <<EOF
POSTGRES_ADMIN_PASSWORD=$admin_password
DB_MIGRATOR_USER=spcase_migrator
DB_MIGRATOR_PASSWORD=$migrator_password
DB_APP_USER=spcase_app
DB_APP_PASSWORD=$app_password
DB_USER=$legacy_role
DB_PASSWORD=$legacy_password
DB_NAME=$target_database
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

echo "phase 9/12: start the tracked Compose runtime against the converted restored volume"
docker rm --force "$target_restore_container" >/dev/null
compose up --detach --wait
nginx_container=$(compose ps --quiet nginx)
app_container=$(compose ps --quiet app)
migrator_container=$(compose ps --all --quiet migrator)
nginx_port=$(docker port "$nginx_container" 8080/tcp | sed -n 's/.*://p')
runtime_url="http://127.0.0.1:$nginx_port"
wait_for_http "$runtime_url/api/v1/health/ready"

app_secret_check=$(docker inspect "$app_container" | jq -r --arg app "$app_password" \
    '.[0].Config.Env as $e | (($e|index("DB_USER=spcase_app")) != null and ($e|index("DB_PASSWORD="+$app)) != null and
     ($e|map(startswith("POSTGRES_ADMIN_PASSWORD=") or startswith("DB_MIGRATOR_PASSWORD=") or startswith("DB_APP_PASSWORD="))|any|not))')
migrator_secret_check=$(docker inspect "$migrator_container" | jq -r --arg migrator "$migrator_password" \
    '.[0].Config.Env as $e | (($e|index("DB_USER=spcase_migrator")) != null and ($e|index("DB_PASSWORD="+$migrator)) != null and
     ($e|map(startswith("POSTGRES_ADMIN_PASSWORD=") or startswith("DB_APP_PASSWORD="))|any|not))')
if [ "$app_secret_check" != "true" ] || [ "$migrator_secret_check" != "true" ]; then
    echo "Compose service credential isolation failed" >&2
    exit 1
fi

for path in /api/v1/health/live /api/v1/health/ready / /api/v1/info; do
    status=$(curl --noproxy '*' --silent --output /dev/null --write-out '%{http_code}' "$runtime_url$path")
    if [ "$status" != "200" ]; then
        echo "$path returned HTTP $status" >&2
        exit 1
    fi
done

echo "phase 10/12: run production-ingress application smoke tests"
admin_token=$(login_token "$runtime_url/api/v1/auth/login" "admin@rehearsal.test" "$account_password" \
    "$response" "$work_dir/admin-runtime.cookie")
api_request 200 GET "$runtime_url/api/v1/admin/stats" "$admin_token" "" "$response"
captain_token=$(login_token "$runtime_url/api/v1/auth/login" "captain@rehearsal.test" "$account_password" \
    "$response" "$work_dir/captain-runtime.cookie")
jury_token=$(login_token "$runtime_url/api/v1/jury/login" "jury@rehearsal.test" "$account_password" \
    "$response" "$work_dir/jury-runtime.cookie")
api_request 200 GET "$runtime_url/api/v1/team/my" "$captain_token" "" "$response"
test "$(jq -r '.id' "$response")" = "$team_id"
test "$(jq -r '.submission.solution_url' "$response")" = "https://example.test/existing-solution"
api_request 200 POST "$runtime_url/api/v1/team/submit" "$captain_token" \
    '{"solution_url":"https://example.test/existing-solution"}' "$response"
api_request 200 POST "$runtime_url/api/v1/jury/evaluations" "$jury_token" \
    "{\"team_id\":\"$team_id\",\"scores\":[{\"criterion_id\":1,\"score\":7},{\"criterion_id\":2,\"score\":8},{\"criterion_id\":3,\"score\":9},{\"criterion_id\":4,\"score\":6},{\"criterion_id\":5,\"score\":8},{\"criterion_id\":6,\"score\":7}]}" "$response"
api_request 200 GET "$runtime_url/api/v1/jury/evaluations" "$jury_token" "" "$response"
test "$(jq '.evaluations | length' "$response")" = "6"
api_request 200 POST "$runtime_url/api/v1/admin/evaluations/close" "$admin_token" '{}' "$response"
api_request 200 POST "$runtime_url/api/v1/admin/evaluations/open" "$admin_token" '{}' "$response"
api_request 201 POST "$runtime_url/api/v1/auth/register" "" \
    "{\"full_name\":\"Runtime Registration\",\"university\":\"University\",\"email\":\"runtime-$safe_run_id@rehearsal.test\",\"telegram\":\"@runtime\",\"password\":\"$account_password\"}" "$response"
runtime_token=$(login_token "$runtime_url/api/v1/auth/login" "runtime-$safe_run_id@rehearsal.test" \
    "$account_password" "$response" "$work_dir/runtime.cookie")
api_request 201 POST "$runtime_url/api/v1/team/create" "$runtime_token" \
    "{\"name\":\"Runtime Team $safe_run_id\"}" "$response"
api_request 200 DELETE "$runtime_url/api/v1/team/disband" "$runtime_token" "" "$response"
api_request 200 POST "$runtime_url/api/v1/auth/logout" "$captain_token" '{}' "$response"

bootstrap_container="spcase-deployment-bootstrap-$run_id"
set +e
printf '%s\n' "$account_password" | compose run --no-deps --name "$bootstrap_container" \
    --entrypoint /app/admin-bootstrap app -full-name "Duplicate Administrator" \
    -email "duplicate-admin@rehearsal.test" >/dev/null 2>&1
bootstrap_status=$?
set -e
if [ "$bootstrap_status" -eq 0 ]; then
    echo "admin bootstrap unexpectedly created a second administrator" >&2
    exit 1
fi
bootstrap_secret_check=$(docker inspect "$bootstrap_container" | jq -r --arg app "$app_password" \
    '.[0].Config.Env as $e | (($e|index("DB_USER=spcase_app")) != null and ($e|index("DB_PASSWORD="+$app)) != null and
     ($e|map(startswith("POSTGRES_ADMIN_PASSWORD=") or startswith("DB_MIGRATOR_PASSWORD=") or startswith("DB_APP_PASSWORD="))|any|not))')
docker rm "$bootstrap_container" >/dev/null
if [ "$bootstrap_secret_check" != "true" ]; then
    echo "admin-bootstrap credential isolation failed" >&2
    exit 1
fi
active_runtime=$(compose exec -T db psql --no-psqlrc --tuples-only --no-align --username postgres \
    --dbname "$target_database" --command "SELECT COUNT(*) FROM pg_stat_activity WHERE usename='spcase_app'")
if [ "$active_runtime" -lt 1 ]; then
    echo "no active spcase_app runtime connection was observed" >&2
    exit 1
fi
if [ "$(preserved_fingerprint "$(compose ps --quiet db)" "$target_database")" != "$source_preserved" ]; then
    echo "pre-existing restored records changed during runtime smoke tests" >&2
    exit 1
fi

echo "phase 11/12: restart the tracked stack and verify persistence"
post_smoke_fingerprint=$(database_fingerprint "$(compose ps --quiet db)" "$target_database" postgres)
compose down --remove-orphans >/dev/null
compose up --detach --wait
nginx_container=$(compose ps --quiet nginx)
nginx_port=$(docker port "$nginx_container" 8080/tcp | sed -n 's/.*://p')
runtime_url="http://127.0.0.1:$nginx_port"
wait_for_http "$runtime_url/api/v1/health/ready"
post_restart_fingerprint=$(database_fingerprint "$(compose ps --quiet db)" "$target_database" postgres)
if [ "$post_restart_fingerprint" != "$post_smoke_fingerprint" ]; then
    echo "database fingerprint changed across normal stack restart" >&2
    exit 1
fi
admin_token=$(login_token "$runtime_url/api/v1/auth/login" "admin@rehearsal.test" "$account_password" \
    "$response" "$work_dir/admin-restart.cookie")
api_request 200 GET "$runtime_url/api/v1/admin/stats" "$admin_token" "" "$response"
if [ "$(preserved_fingerprint "$(compose ps --quiet db)" "$target_database")" != "$source_preserved" ]; then
    echo "pre-existing restored data was not preserved after restart" >&2
    exit 1
fi

echo "phase 12/12: restore the original backup into an independent rollback target"
docker network create "$rollback_network" >/dev/null
docker volume create "$rollback_volume" >/dev/null
docker run --detach --name "$rollback_container" --network "$rollback_network" \
    --env POSTGRES_DB="$rollback_database" --env POSTGRES_USER=postgres \
    --env POSTGRES_PASSWORD="$admin_password" --volume "$rollback_volume:/var/lib/postgresql/data" \
    "$postgres_image" >/dev/null
wait_for_postgres "$rollback_container" "$rollback_database" postgres
docker exec --interactive --env LEGACY_PASSWORD="$legacy_password" "$rollback_container" psql \
    --no-psqlrc --set=ON_ERROR_STOP=1 --username postgres --dbname "$rollback_database" >/dev/null <<'SQL'
\getenv legacy_password LEGACY_PASSWORD
SELECT format(
    'CREATE ROLE spcase_legacy LOGIN PASSWORD %L NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS',
    :'legacy_password'
) \gexec
ALTER DATABASE spcase_rollback OWNER TO spcase_legacy;
SQL
docker cp "$backup_file" "$rollback_container:/tmp/spcase-legacy.dump" >/dev/null
docker exec "$rollback_container" pg_restore --exit-on-error --no-owner --role="$legacy_role" \
    --username postgres --dbname="$rollback_database" /tmp/spcase-legacy.dump
rollback_fingerprint=$(database_fingerprint "$rollback_container" "$rollback_database" postgres)
rollback_ownership=$(ownership_summary "$rollback_container" "$rollback_database")
rollback_operational=$(docker exec --env PGPASSWORD="$legacy_password" "$rollback_container" psql \
    --host=127.0.0.1 --username="$legacy_role" --dbname="$rollback_database" --no-psqlrc \
    --tuples-only --no-align --command 'SELECT COUNT(*) FROM users')
if [ "$rollback_fingerprint" != "$source_fingerprint" ] || \
   [ "$rollback_ownership" != "$source_ownership" ] || [ "$rollback_operational" -lt 4 ]; then
    echo "rollback restore failed fingerprint, ownership or operational validation" >&2
    exit 1
fi

printf '%s\n' \
    "rehearsal result: PASS" \
    "source/restore fingerprint: $source_fingerprint" \
    "converted/restarted fingerprint: $post_restart_fingerprint" \
    "rollback fingerprint: $rollback_fingerprint" \
    "backup: custom format, $backup_size bytes, sha256 $backup_checksum" \
    "rollback configuration boundary: deploy the prior Compose revision explicitly with transitional DB_USER/DB_PASSWORD; no automatic fallback was used" \
    "cleanup will remove only resources and backup artifacts created by run $run_id"
