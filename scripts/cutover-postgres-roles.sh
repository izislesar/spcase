#!/bin/sh

set -eu

umask 077

: "${PGHOST:?PGHOST must identify the PostgreSQL server}"
: "${POSTGRES_ADMIN_USER:?POSTGRES_ADMIN_USER must be set}"
: "${POSTGRES_ADMIN_PASSWORD:?POSTGRES_ADMIN_PASSWORD must be set}"
: "${SPCASE_CUTOVER_DATABASE:?SPCASE_CUTOVER_DATABASE must be set explicitly}"
: "${DB_MIGRATOR_PASSWORD:?DB_MIGRATOR_PASSWORD must be set}"
: "${DB_APP_PASSWORD:?DB_APP_PASSWORD must be set}"
: "${SPCASE_LEGACY_DB_ROLE:?SPCASE_LEGACY_DB_ROLE must be set}"
: "${SPCASE_CONFIRM_EXISTING_DB_CUTOVER:?SPCASE_CONFIRM_EXISTING_DB_CUTOVER must be set}"

PGPORT=${PGPORT:-5432}
export PGPORT

if [ "$SPCASE_CONFIRM_EXISTING_DB_CUTOVER" != "YES" ]; then
    echo "refusing existing-database cutover without SPCASE_CONFIRM_EXISTING_DB_CUTOVER=YES" >&2
    exit 1
fi
if [ "$SPCASE_CUTOVER_DATABASE" = "template0" ] || [ "$SPCASE_CUTOVER_DATABASE" = "template1" ]; then
    echo "refusing to cut over a PostgreSQL template database" >&2
    exit 1
fi
if [ "$SPCASE_LEGACY_DB_ROLE" = "spcase_migrator" ] || [ "$SPCASE_LEGACY_DB_ROLE" = "spcase_app" ]; then
    echo "SPCASE_LEGACY_DB_ROLE must identify the pre-cutover owner" >&2
    exit 1
fi
if [ "$POSTGRES_ADMIN_PASSWORD" = "$DB_MIGRATOR_PASSWORD" ] || \
   [ "$POSTGRES_ADMIN_PASSWORD" = "$DB_APP_PASSWORD" ] || \
   [ "$DB_MIGRATOR_PASSWORD" = "$DB_APP_PASSWORD" ]; then
    echo "administrator, migrator and application passwords must be different" >&2
    exit 1
fi
if ! command -v psql >/dev/null 2>&1; then
    echo "psql is required for PostgreSQL role cutover" >&2
    exit 1
fi

case "$0" in
    */*) script_directory=${0%/*} ;;
    *) script_directory=. ;;
esac
sql_file=$script_directory/sql/cutover-postgres-roles.sql
if [ ! -r "$sql_file" ]; then
    echo "cutover SQL file is not readable: $sql_file" >&2
    exit 1
fi

export POSTGRES_ADMIN_USER DB_MIGRATOR_PASSWORD DB_APP_PASSWORD SPCASE_LEGACY_DB_ROLE
export SPCASE_CONFIRM_EXISTING_DB_CUTOVER SPCASE_CUTOVER_DATABASE

echo "PostgreSQL existing-volume cutover: preflight, mutation and administrative validation"
PGPASSWORD=$POSTGRES_ADMIN_PASSWORD
export PGPASSWORD
psql \
    --no-psqlrc \
    --host "$PGHOST" \
    --port "$PGPORT" \
    --username "$POSTGRES_ADMIN_USER" \
    --dbname "$SPCASE_CUTOVER_DATABASE" \
    --no-password \
    --file "$sql_file"

echo "PostgreSQL existing-volume cutover: validating migrator credentials"
PGPASSWORD=$DB_MIGRATOR_PASSWORD
export PGPASSWORD
migrator_identity=$(psql \
    --no-psqlrc \
    --host "$PGHOST" \
    --port "$PGPORT" \
    --username spcase_migrator \
    --dbname "$SPCASE_CUTOVER_DATABASE" \
    --no-password \
    --tuples-only \
    --no-align \
    --command 'SELECT current_user')
if [ "$migrator_identity" != "spcase_migrator" ]; then
    echo "migrator credential validation returned an unexpected role" >&2
    exit 1
fi

echo "PostgreSQL existing-volume cutover: validating runtime credentials"
PGPASSWORD=$DB_APP_PASSWORD
export PGPASSWORD
app_identity=$(psql \
    --no-psqlrc \
    --host "$PGHOST" \
    --port "$PGPORT" \
    --username spcase_app \
    --dbname "$SPCASE_CUTOVER_DATABASE" \
    --no-password \
    --tuples-only \
    --no-align \
    --command 'SELECT current_user')
if [ "$app_identity" != "spcase_app" ]; then
    echo "application credential validation returned an unexpected role" >&2
    exit 1
fi

unset PGPASSWORD
echo "PostgreSQL existing-volume role cutover completed and validated"
