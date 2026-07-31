#!/bin/sh

set -eu

: "${POSTGRES_DB:?POSTGRES_DB must be set}"
: "${POSTGRES_USER:?POSTGRES_USER must be set}"
: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD must be set}"
: "${DB_MIGRATOR_USER:?DB_MIGRATOR_USER must be set}"
: "${DB_MIGRATOR_PASSWORD:?DB_MIGRATOR_PASSWORD must be set}"
: "${DB_APP_USER:?DB_APP_USER must be set}"
: "${DB_APP_PASSWORD:?DB_APP_PASSWORD must be set}"

if [ "$POSTGRES_USER" != "postgres" ]; then
    echo "POSTGRES_USER must be postgres" >&2
    exit 1
fi
if [ "$DB_MIGRATOR_USER" != "spcase_migrator" ]; then
    echo "DB_MIGRATOR_USER must be spcase_migrator" >&2
    exit 1
fi
if [ "$DB_APP_USER" != "spcase_app" ]; then
    echo "DB_APP_USER must be spcase_app" >&2
    exit 1
fi
if [ "$DB_MIGRATOR_PASSWORD" = "$DB_APP_PASSWORD" ]; then
    echo "migrator and application passwords must be different" >&2
    exit 1
fi
if [ "$POSTGRES_PASSWORD" = "$DB_MIGRATOR_PASSWORD" ] || [ "$POSTGRES_PASSWORD" = "$DB_APP_PASSWORD" ]; then
    echo "administrator, migrator and application passwords must be different" >&2
    exit 1
fi

psql --set=ON_ERROR_STOP=1 \
    --username "$POSTGRES_USER" \
    --dbname "$POSTGRES_DB" \
    --set=migrator_password="$DB_MIGRATOR_PASSWORD" \
    --set=app_password="$DB_APP_PASSWORD" <<'SQL'
SELECT format(
    'CREATE ROLE spcase_migrator LOGIN PASSWORD %L NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS',
    :'migrator_password'
)
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'spcase_migrator') \gexec

SELECT format(
    'CREATE ROLE spcase_app LOGIN PASSWORD %L NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS',
    :'app_password'
)
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'spcase_app') \gexec

ALTER ROLE spcase_migrator
    LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS
    PASSWORD :'migrator_password';
ALTER ROLE spcase_app
    LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS
    PASSWORD :'app_password';

SELECT format('ALTER DATABASE %I OWNER TO spcase_migrator', current_database()) \gexec
ALTER SCHEMA public OWNER TO spcase_migrator;

SELECT format('REVOKE ALL PRIVILEGES ON DATABASE %I FROM PUBLIC', current_database()) \gexec
SELECT format('GRANT CONNECT ON DATABASE %I TO spcase_migrator', current_database()) \gexec
SELECT format('GRANT CONNECT ON DATABASE %I TO spcase_app', current_database()) \gexec

REVOKE ALL PRIVILEGES ON SCHEMA public FROM PUBLIC;
GRANT USAGE, CREATE ON SCHEMA public TO spcase_migrator;
GRANT USAGE ON SCHEMA public TO spcase_app;

ALTER DEFAULT PRIVILEGES FOR ROLE spcase_migrator IN SCHEMA public
    REVOKE ALL PRIVILEGES ON TABLES FROM PUBLIC;
ALTER DEFAULT PRIVILEGES FOR ROLE spcase_migrator IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO spcase_app;
ALTER DEFAULT PRIVILEGES FOR ROLE spcase_migrator IN SCHEMA public
    REVOKE ALL PRIVILEGES ON SEQUENCES FROM PUBLIC;
ALTER DEFAULT PRIVILEGES FOR ROLE spcase_migrator IN SCHEMA public
    GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO spcase_app;
ALTER DEFAULT PRIVILEGES FOR ROLE spcase_migrator IN SCHEMA public
    REVOKE ALL PRIVILEGES ON TYPES FROM PUBLIC;
ALTER DEFAULT PRIVILEGES FOR ROLE spcase_migrator IN SCHEMA public
    GRANT USAGE ON TYPES TO spcase_app;
SQL
