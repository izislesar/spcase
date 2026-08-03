#!/bin/sh

set -eu

umask 077

: "${SPCASE_CONFIRM_POSTGRES_ROLE_CUTOVER_REHEARSAL:?Set SPCASE_CONFIRM_POSTGRES_ROLE_CUTOVER_REHEARSAL=YES}"
if [ "$SPCASE_CONFIRM_POSTGRES_ROLE_CUTOVER_REHEARSAL" != "YES" ]; then
    echo "refusing PostgreSQL role cutover rehearsal without SPCASE_CONFIRM_POSTGRES_ROLE_CUTOVER_REHEARSAL=YES" >&2
    exit 1
fi

if ! command -v docker >/dev/null 2>&1; then
    echo "docker is required for the PostgreSQL cutover rehearsal" >&2
    exit 1
fi
if ! docker info >/dev/null 2>&1; then
    echo "Docker daemon is unavailable" >&2
    exit 1
fi
if ! command -v go >/dev/null 2>&1; then
    echo "go is required for the PostgreSQL integration rehearsal" >&2
    exit 1
fi

script_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
repository_dir=$(CDPATH='' cd -- "$script_dir/.." && pwd)
work_dir=$(mktemp -d "${TMPDIR:-/tmp}/spcase-cutover-rehearsal.XXXXXX")
run_id="$(date +%s)-$$-${work_dir##*.}"
network="spcase-cutover-$run_id"
source_container="spcase-cutover-source-$run_id"
restored_container="spcase-cutover-restored-$run_id"
source_volume="spcase-cutover-source-$run_id"
restored_volume="spcase-cutover-restored-$run_id"
migrator_image="spcase-cutover-migrator:$run_id"
postgres_image="postgres:16-bookworm@sha256:92620daddcd947f8d5ab5ba66e848702fe443d87fed30c4cea8e389fd78dfc55"
source_database="spcase_legacy"
restored_database="spcase_restored"
legacy_role="spcase_legacy"
legacy_password="legacy-$run_id"
migrator_password="migrator-$run_id"
app_password="application-$run_id"
backup_file="$work_dir/spcase-legacy.dump"

cleanup() {
    exit_status=$?
    cleanup_failed=0
    trap - EXIT HUP INT TERM

    docker rm --force "$source_container" "$restored_container" >/dev/null 2>&1 || true
    docker volume rm "$source_volume" "$restored_volume" >/dev/null 2>&1 || true
    docker network rm "$network" >/dev/null 2>&1 || true
    docker image rm "$migrator_image" >/dev/null 2>&1 || true
    if ! rm -rf "$work_dir"; then
        echo "cleanup failed: could not remove work directory: $work_dir" >&2
        cleanup_failed=1
    fi

    if ! docker info >/dev/null 2>&1; then
        echo "cleanup failed: Docker is unavailable; resource removal cannot be verified" >&2
        cleanup_failed=1
    else
        for container in "$source_container" "$restored_container"; do
            if docker container inspect "$container" >/dev/null 2>&1; then
                echo "cleanup failed: container remains: $container" >&2
                cleanup_failed=1
            fi
        done
        for volume in "$source_volume" "$restored_volume"; do
            if docker volume inspect "$volume" >/dev/null 2>&1; then
                echo "cleanup failed: volume remains: $volume" >&2
                cleanup_failed=1
            fi
        done
        if docker network inspect "$network" >/dev/null 2>&1; then
            echo "cleanup failed: network remains: $network" >&2
            cleanup_failed=1
        fi
        if docker image inspect "$migrator_image" >/dev/null 2>&1; then
            echo "cleanup failed: image remains: $migrator_image" >&2
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
        if [ "$attempt" -ge 60 ]; then
            echo "PostgreSQL did not become ready in $container" >&2
            return 1
        fi
        if docker exec "$container" psql \
            --no-psqlrc --tuples-only --no-align --username "$database_user" --dbname "$database" \
            --command 'SELECT 1' >/dev/null 2>&1; then
            consecutive_successes=$((consecutive_successes + 1))
        else
            consecutive_successes=0
        fi
        sleep 1
    done
}

database_fingerprint() {
    container=$1
    database=$2
    database_user=$3
    docker exec "$container" psql \
        --no-psqlrc --tuples-only --no-align --set=ON_ERROR_STOP=1 \
        --username "$database_user" --dbname "$database" --command "
            SELECT md5(concat_ws('|',
                (SELECT string_agg(id::text, ',' ORDER BY id) FROM public.users),
                (SELECT string_agg(id::text, ',' ORDER BY id) FROM public.teams),
                (SELECT string_agg(team_id::text || ':' || user_id::text, ',' ORDER BY team_id, user_id) FROM public.team_members),
                (SELECT string_agg(id::text, ',' ORDER BY id) FROM public.submissions),
                (SELECT string_agg(id::text, ',' ORDER BY id) FROM public.evaluations)
            ));
        "
}

acl_fingerprint() {
    container=$1
    database=$2
    docker exec "$container" psql \
        --no-psqlrc --tuples-only --no-align --set=ON_ERROR_STOP=1 \
        --username postgres --dbname "$database" --command "
            SELECT md5(concat_ws('|',
                (SELECT string_agg(object.relkind::text || ':' || object.relname || ':' || pg_get_userbyid(object.relowner), ','
                                   ORDER BY object.relkind, object.relname)
                 FROM pg_class object
                 JOIN pg_namespace namespace ON namespace.oid = object.relnamespace
                 WHERE namespace.nspname = 'public'),
                (SELECT string_agg(object.relname || ':' || privileges.grantee::text || ':' || privileges.privilege_type || ':' ||
                                   privileges.grantor::text || ':' || privileges.is_grantable::text, ','
                                   ORDER BY object.relname, privileges.grantee, privileges.privilege_type,
                                            privileges.grantor, privileges.is_grantable)
                 FROM pg_class object
                 JOIN pg_namespace namespace ON namespace.oid = object.relnamespace
                 CROSS JOIN LATERAL aclexplode(object.relacl) privileges
                 WHERE namespace.nspname = 'public'),
                (SELECT namespace.nspowner::text || ':' || COALESCE(string_agg(
                            privileges.grantee::text || ':' || privileges.privilege_type || ':' ||
                            privileges.grantor::text || ':' || privileges.is_grantable::text, ','
                            ORDER BY privileges.grantee, privileges.privilege_type,
                                     privileges.grantor, privileges.is_grantable), '')
                 FROM pg_namespace namespace
                 LEFT JOIN LATERAL aclexplode(namespace.nspacl) privileges ON true
                 WHERE namespace.nspname = 'public'
                 GROUP BY namespace.nspowner),
                (SELECT string_agg(defaults.defaclobjtype::text || ':' || privileges.grantee::text || ':' ||
                                   privileges.privilege_type || ':' || privileges.grantor::text || ':' ||
                                   privileges.is_grantable::text, ','
                                   ORDER BY defaults.defaclobjtype, privileges.grantee, privileges.privilege_type,
                                            privileges.grantor, privileges.is_grantable)
                 FROM pg_default_acl defaults
                 JOIN pg_roles owner ON owner.oid = defaults.defaclrole
                 CROSS JOIN LATERAL aclexplode(defaults.defaclacl) privileges
                 WHERE owner.rolname = 'spcase_migrator')
            ));
        "
}

echo "rehearsal: creating isolated source and restored PostgreSQL resources ($run_id)"
docker network create "$network" >/dev/null
docker volume create "$source_volume" >/dev/null
docker volume create "$restored_volume" >/dev/null
docker build --target migrator --tag "$migrator_image" "$repository_dir" >/dev/null

docker run --detach \
    --name "$source_container" \
    --network "$network" \
    --publish 127.0.0.1::5432 \
    --env POSTGRES_DB="$source_database" \
    --env POSTGRES_USER="$legacy_role" \
    --env POSTGRES_PASSWORD="$legacy_password" \
    --volume "$source_volume:/var/lib/postgresql/data" \
    "$postgres_image" >/dev/null
wait_for_postgres "$source_container" "$source_database" "$legacy_role"

mkdir "$work_dir/pre-cutover-migrations"
cp "$repository_dir/migrations/00001_init_schema.sql" "$work_dir/pre-cutover-migrations/"
cp "$repository_dir/migrations/00002_add_indexes.sql" "$work_dir/pre-cutover-migrations/"
cp "$repository_dir/migrations/00003_seed_dev_data.sql" "$work_dir/pre-cutover-migrations/"
chmod 711 "$work_dir"
chmod 755 "$work_dir/pre-cutover-migrations"
chmod 644 "$work_dir/pre-cutover-migrations"/*.sql
docker run --rm \
    --network "$network" \
    --env PGPASSWORD="$legacy_password" \
    --volume "$work_dir/pre-cutover-migrations:/migrations:ro" \
    --entrypoint /usr/local/bin/goose \
    "$migrator_image" \
    -dir /migrations postgres \
    "postgres://$legacy_role@$source_container:5432/$source_database?sslmode=disable" up

source_port=$(docker port "$source_container" 5432/tcp | sed -n 's/.*://p')
if [ -z "$source_port" ]; then
    echo "unable to resolve source PostgreSQL host port" >&2
    exit 1
fi
seed_rejection_log="$work_dir/seed-rejection.log"
set +e
PGHOST=127.0.0.1 \
PGPORT="$source_port" \
POSTGRES_ADMIN_USER="$legacy_role" \
POSTGRES_ADMIN_PASSWORD="$legacy_password" \
DB_MIGRATOR_PASSWORD="$migrator_password" \
DB_APP_PASSWORD="$app_password" \
SPCASE_CUTOVER_DATABASE="$source_database" \
SPCASE_LEGACY_DB_ROLE="$legacy_role" \
SPCASE_CONFIRM_EXISTING_DB_CUTOVER=YES \
    "$repository_dir/scripts/cutover-postgres-roles.sh" >"$seed_rejection_log" 2>&1
seed_rejection_status=$?
set -e
if [ "$seed_rejection_status" -eq 0 ] || \
   ! grep -q 'development seed migration 00003 must not be applied in production' "$seed_rejection_log"; then
    echo "cutover preflight did not reject development seed migration 00003" >&2
    exit 1
fi
seed_rejection_state=$(docker exec "$source_container" psql \
    --no-psqlrc --tuples-only --no-align --username "$legacy_role" --dbname "$source_database" \
    --command "SELECT pg_get_userbyid(datdba) || '|' ||
        (SELECT count(*) FROM pg_roles WHERE rolname IN ('spcase_migrator', 'spcase_app'))
        FROM pg_database WHERE datname = current_database()")
if [ "$seed_rejection_state" != "$legacy_role|0" ]; then
    echo "development-seed rejection changed source ownership or target roles" >&2
    exit 1
fi
docker run --rm \
    --network "$network" \
    --env PGPASSWORD="$legacy_password" \
    --volume "$work_dir/pre-cutover-migrations:/migrations:ro" \
    --entrypoint /usr/local/bin/goose \
    "$migrator_image" \
    -dir /migrations postgres \
    "postgres://$legacy_role@$source_container:5432/$source_database?sslmode=disable" down
source_version=$(docker exec "$source_container" psql \
    --no-psqlrc --tuples-only --no-align --username "$legacy_role" --dbname "$source_database" \
    --command 'SELECT MAX(version_id) FILTER (WHERE is_applied) FROM public.goose_db_version')
source_users=$(docker exec "$source_container" psql \
    --no-psqlrc --tuples-only --no-align --username "$legacy_role" --dbname "$source_database" \
    --command 'SELECT COUNT(*) FROM public.users')
if [ "$source_version" != "2" ] || [ "$source_users" != "0" ]; then
    echo "development-seed rollback did not restore the version-2 source baseline" >&2
    exit 1
fi
echo "rehearsal: development seed preflight rejection passed without mutation"

docker exec "$source_container" psql \
    --no-psqlrc --set=ON_ERROR_STOP=1 --username "$legacy_role" --dbname "$source_database" \
    --command "
        BEGIN;
        INSERT INTO users (id, full_name, university, email, telegram, password_hash, role) VALUES
            ('10000000-0000-4000-8000-000000000001', 'Rehearsal Admin', NULL, 'admin@rehearsal.test', NULL, 'hash', 'ADMIN'),
            ('20000000-0000-4000-8000-000000000001', 'Rehearsal Jury', NULL, 'jury@rehearsal.test', NULL, 'hash', 'JURY'),
            ('30000000-0000-4000-8000-000000000001', 'Rehearsal Captain', 'University', 'captain@rehearsal.test', '@captain', 'hash', 'USER'),
            ('30000000-0000-4000-8000-000000000002', 'Rehearsal Member', 'University', 'member@rehearsal.test', '@member', 'hash', 'USER');
        INSERT INTO teams (id, name, invite_code, captain_id)
        VALUES ('40000000-0000-4000-8000-000000000001', 'Rehearsal Team', 'REHEAR01', '30000000-0000-4000-8000-000000000001');
        INSERT INTO team_members (team_id, user_id) VALUES
            ('40000000-0000-4000-8000-000000000001', '30000000-0000-4000-8000-000000000001'),
            ('40000000-0000-4000-8000-000000000001', '30000000-0000-4000-8000-000000000002');
        INSERT INTO submissions (id, team_id, solution_url)
        VALUES ('50000000-0000-4000-8000-000000000001', '40000000-0000-4000-8000-000000000001', 'https://example.test/rehearsal');
        INSERT INTO evaluations (id, jury_id, team_id, criterion_id, score)
        VALUES ('60000000-0000-4000-8000-000000000001', '20000000-0000-4000-8000-000000000001', '40000000-0000-4000-8000-000000000001', 1, 7);
        COMMIT;
    " >/dev/null

source_fingerprint=$(database_fingerprint "$source_container" "$source_database" "$legacy_role")
echo "rehearsal: creating pg_dump custom-format backup"
echo "backup command: pg_dump --format=custom --username $legacy_role --dbname $source_database"
docker exec "$source_container" pg_dump \
    --format=custom --username "$legacy_role" --dbname "$source_database" \
    --file=/tmp/spcase-legacy.dump
docker cp "$source_container:/tmp/spcase-legacy.dump" "$backup_file" >/dev/null
test -s "$backup_file"

docker run --detach \
    --name "$restored_container" \
    --network "$network" \
    --publish 127.0.0.1::5432 \
    --env POSTGRES_DB="$restored_database" \
    --env POSTGRES_USER=postgres \
    --env POSTGRES_PASSWORD="$legacy_password" \
    --volume "$restored_volume:/var/lib/postgresql/data" \
    "$postgres_image" >/dev/null
wait_for_postgres "$restored_container" "$restored_database" postgres
docker exec --interactive --env REHEARSAL_LEGACY_PASSWORD="$legacy_password" \
    "$restored_container" psql \
    --no-psqlrc --set=ON_ERROR_STOP=1 --username postgres --dbname "$restored_database" \
    >/dev/null <<'SQL'
\getenv legacy_password REHEARSAL_LEGACY_PASSWORD
SELECT format(
    'CREATE ROLE spcase_legacy LOGIN PASSWORD %L NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS',
    :'legacy_password'
) \gexec
ALTER DATABASE spcase_restored OWNER TO spcase_legacy;
SQL
docker cp "$backup_file" "$restored_container:/tmp/spcase-legacy.dump" >/dev/null
docker exec "$restored_container" pg_restore \
    --exit-on-error --no-owner --role="$legacy_role" --username postgres --dbname "$restored_database" \
    /tmp/spcase-legacy.dump

restored_fingerprint=$(database_fingerprint "$restored_container" "$restored_database" postgres)
if [ "$restored_fingerprint" != "$source_fingerprint" ]; then
    echo "restored application data does not match the source backup" >&2
    exit 1
fi
docker exec "$restored_container" psql \
    --no-psqlrc --set=ON_ERROR_STOP=1 --username postgres --dbname "$restored_database" \
    --command "
        SELECT current_database(),
               pg_get_userbyid((SELECT datdba FROM pg_database WHERE datname = current_database())) AS database_owner,
               pg_get_userbyid((SELECT nspowner FROM pg_namespace WHERE nspname = 'public')) AS schema_owner,
               (SELECT MAX(version_id) FROM public.goose_db_version WHERE is_applied) AS migration_version,
               (SELECT COUNT(*) FROM public.users) AS users,
               (SELECT COUNT(*) FROM public.teams) AS teams,
               (SELECT COUNT(*) FROM public.submissions) AS submissions,
               (SELECT COUNT(*) FROM public.evaluations) AS evaluations;
    "

target_port=$(docker port "$restored_container" 5432/tcp | sed -n 's/.*://p')
if [ -z "$target_port" ]; then
    echo "unable to resolve restored PostgreSQL host port" >&2
    exit 1
fi

run_cutover() {
    PGHOST=127.0.0.1 \
    PGPORT="$target_port" \
    POSTGRES_ADMIN_USER=postgres \
    POSTGRES_ADMIN_PASSWORD="$legacy_password" \
    DB_MIGRATOR_PASSWORD="$migrator_password" \
    DB_APP_PASSWORD="$app_password" \
    SPCASE_CUTOVER_DATABASE="$restored_database" \
    SPCASE_LEGACY_DB_ROLE="$legacy_role" \
    SPCASE_CONFIRM_EXISTING_DB_CUTOVER=YES \
    "$repository_dir/scripts/cutover-postgres-roles.sh"
}

echo "rehearsal: running first existing-volume cutover"
run_cutover

docker run --rm \
    --network "$network" \
    --env DB_HOST="$restored_container" \
    --env DB_PORT=5432 \
    --env DB_NAME="$restored_database" \
    --env DB_USER=spcase_migrator \
    --env DB_PASSWORD="$migrator_password" \
    "$migrator_image"

migrator_url="postgres://spcase_migrator:$migrator_password@127.0.0.1:$target_port/$restored_database?sslmode=disable"
app_url="postgres://spcase_app:$app_password@127.0.0.1:$target_port/$restored_database?sslmode=disable"
SPCASE_TEST_MIGRATOR_DATABASE_URL="$migrator_url" \
SPCASE_TEST_APP_DATABASE_URL="$app_url" \
go test -race -count=1 -tags=integration "$repository_dir/internal/..."

after_first_fingerprint=$(database_fingerprint "$restored_container" "$restored_database" postgres)
if [ "$after_first_fingerprint" != "$source_fingerprint" ]; then
    echo "application data changed after cutover and migrations" >&2
    exit 1
fi
before_second_acl=$(acl_fingerprint "$restored_container" "$restored_database")

docker exec "$restored_container" psql \
    --no-psqlrc --set=ON_ERROR_STOP=1 --username postgres --dbname "$restored_database" \
    --command 'CREATE ROLE spcase_acl_inheritor; GRANT spcase_app TO spcase_acl_inheritor' >/dev/null
membership_rejection_log="$work_dir/membership-rejection.log"
set +e
run_cutover >"$membership_rejection_log" 2>&1
membership_rejection_status=$?
set -e
if [ "$membership_rejection_status" -eq 0 ] || \
   ! grep -q 'existing target roles must not participate in role memberships' "$membership_rejection_log"; then
    echo "cutover preflight did not reject target-role membership drift" >&2
    exit 1
fi
docker exec "$restored_container" psql \
    --no-psqlrc --set=ON_ERROR_STOP=1 --username postgres --dbname "$restored_database" \
    --command 'REVOKE spcase_app FROM spcase_acl_inheritor; DROP ROLE spcase_acl_inheritor' >/dev/null
if [ "$(acl_fingerprint "$restored_container" "$restored_database")" != "$before_second_acl" ]; then
    echo "membership-drift rejection changed ownership or ACL state" >&2
    exit 1
fi
echo "rehearsal: target-role membership preflight rejection passed without mutation"

echo "rehearsal: running idempotent second cutover"
run_cutover
after_second_acl=$(acl_fingerprint "$restored_container" "$restored_database")
if [ "$after_second_acl" != "$before_second_acl" ]; then
    echo "ownership or ACL fingerprint changed on the second cutover" >&2
    exit 1
fi
if [ "$(database_fingerprint "$restored_container" "$restored_database" postgres)" != "$source_fingerprint" ]; then
    echo "application data changed on the second cutover" >&2
    exit 1
fi

docker run --rm \
    --network "$network" \
    --env DB_HOST="$restored_container" \
    --env DB_PORT=5432 \
    --env DB_NAME="$restored_database" \
    --env DB_USER=spcase_migrator \
    --env DB_PASSWORD="$migrator_password" \
    "$migrator_image"

docker exec --env PGPASSWORD="$app_password" "$restored_container" psql \
    --host=127.0.0.1 --username=spcase_app --dbname="$restored_database" \
    --no-psqlrc --set=ON_ERROR_STOP=1 --command "
        BEGIN;
        INSERT INTO users (full_name, university, email, telegram, password_hash, role)
        VALUES ('Runtime Probe', 'University', 'runtime-probe@rehearsal.test', '@runtime_probe', 'hash', 'USER');
        UPDATE users SET full_name = full_name WHERE email = 'runtime-probe@rehearsal.test';
        SELECT id FROM users WHERE email = 'runtime-probe@rehearsal.test';
        ROLLBACK;
    " >/dev/null

if docker exec --env PGPASSWORD="$app_password" "$restored_container" psql \
    --host=127.0.0.1 --username=spcase_app --dbname="$restored_database" \
    --no-psqlrc --set=ON_ERROR_STOP=1 \
    --command "CREATE TABLE public.cutover_ddl_must_fail (id integer)" >/dev/null 2>&1; then
    echo "spcase_app unexpectedly created a table" >&2
    exit 1
fi
if docker exec --env PGPASSWORD="$app_password" "$restored_container" psql \
    --host=127.0.0.1 --username=spcase_app --dbname="$restored_database" \
    --no-psqlrc --set=ON_ERROR_STOP=1 \
    --command "SELECT * FROM public.goose_db_version" >/dev/null 2>&1; then
    echo "spcase_app unexpectedly read Goose metadata" >&2
    exit 1
fi

docker exec --env PGPASSWORD="$migrator_password" "$restored_container" psql \
    --host=127.0.0.1 --username=spcase_migrator --dbname="$restored_database" \
    --no-psqlrc --set=ON_ERROR_STOP=1 --command "
        SELECT current_database(),
               pg_get_userbyid((SELECT datdba FROM pg_database WHERE datname = current_database())) AS database_owner,
               pg_get_userbyid((SELECT nspowner FROM pg_namespace WHERE nspname = 'public')) AS schema_owner,
               pg_get_userbyid((SELECT relowner FROM pg_class WHERE oid = 'public.goose_db_version'::regclass)) AS goose_table_owner,
               pg_get_userbyid((SELECT relowner FROM pg_class WHERE oid = 'public.goose_db_version_id_seq'::regclass)) AS goose_sequence_owner,
               (SELECT MAX(version_id) FROM public.goose_db_version WHERE is_applied) AS migration_version;
    "

echo "rehearsal result: seed rejection, backup, restore, cutover, integration, ACL and second-run checks passed"
