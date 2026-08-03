#!/bin/sh

set -eu

goose_bin=${1:?goose binary is required}
goose_driver=${2:?goose driver is required}
database_url=${3:?database URL is required}
migrations_dir=${4:?migrations directory is required}
manifest=${5:?production migration manifest is required}

# fail emits a stable, machine-grep-able migration_failed event for operators
# and log pipelines, then exits nonzero.
fail() {
    echo "migration_failed: $*" >&2
    exit 1
}

if [ ! -f "$manifest" ]; then
    fail "production migration manifest not found: $manifest"
fi

temporary_dir=$(mktemp -d "${TMPDIR:-/tmp}/spcase-production-migrations.XXXXXX")
cleanup() {
    rm -f "$temporary_dir"/*.sql
    rmdir "$temporary_dir"
}
trap cleanup EXIT
trap 'exit 1' HUP INT TERM

migration_count=0
expected_version=0
while IFS= read -r migration || [ -n "$migration" ]; do
    case "$migration" in
        ""|\#*)
            continue
            ;;
        */*|*".."*)
            fail "invalid production migration entry: $migration"
            ;;
        00003.sql|00003_*.sql)
            fail "refusing to run development seed migration in production"
            ;;
        *.sql)
            ;;
        *)
            fail "invalid production migration entry: $migration"
            ;;
    esac

    source_file="$migrations_dir/$migration"
    if [ ! -f "$source_file" ]; then
        fail "production migration not found: $source_file"
    fi

    cp "$source_file" "$temporary_dir/$migration"
    migration_count=$((migration_count + 1))

    migration_version=$(printf '%s\n' "$migration" | sed -n 's/^\([0-9][0-9]*\)_.*/\1/p')
    migration_version=$(printf '%s\n' "$migration_version" | sed 's/^0*//; s/^$/0/')
    case "$migration_version" in
        ""|*[!0-9]*)
            fail "invalid versioned production migration: $migration"
            ;;
    esac
    if [ "$migration_version" -gt "$expected_version" ]; then
        expected_version=$migration_version
    fi
done < "$manifest"

if [ "$migration_count" -eq 0 ]; then
    fail "production migration manifest is empty"
fi

if ! "$goose_bin" -dir "$temporary_dir" "$goose_driver" "$database_url" up; then
    fail "goose up failed"
fi

if ! version_output=$("$goose_bin" -dir "$temporary_dir" "$goose_driver" "$database_url" version 2>&1); then
    fail "unable to verify production migration version: $version_output"
fi
current_version=${version_output##* }
case "$current_version" in
    ""|*[!0-9]*)
        fail "unable to verify production migration version: $version_output"
        ;;
esac

if [ "$current_version" -ne "$expected_version" ]; then
    fail "production database version $current_version does not match allowed version $expected_version"
fi
