#!/bin/sh

set -eu

goose_bin=${1:?goose binary is required}
goose_driver=${2:?goose driver is required}
database_url=${3:?database URL is required}
migrations_dir=${4:?migrations directory is required}
manifest=${5:?production migration manifest is required}

if [ ! -f "$manifest" ]; then
    echo "production migration manifest not found: $manifest" >&2
    exit 1
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
            echo "invalid production migration entry: $migration" >&2
            exit 1
            ;;
        00003.sql|00003_*.sql)
            echo "refusing to run development seed migration in production" >&2
            exit 1
            ;;
        *.sql)
            ;;
        *)
            echo "invalid production migration entry: $migration" >&2
            exit 1
            ;;
    esac

    source_file="$migrations_dir/$migration"
    if [ ! -f "$source_file" ]; then
        echo "production migration not found: $source_file" >&2
        exit 1
    fi

    cp "$source_file" "$temporary_dir/$migration"
    migration_count=$((migration_count + 1))

    migration_version=$(printf '%s\n' "$migration" | sed -n 's/^\([0-9][0-9]*\)_.*/\1/p')
    migration_version=$(printf '%s\n' "$migration_version" | sed 's/^0*//; s/^$/0/')
    case "$migration_version" in
        ""|*[!0-9]*)
            echo "invalid versioned production migration: $migration" >&2
            exit 1
            ;;
    esac
    if [ "$migration_version" -gt "$expected_version" ]; then
        expected_version=$migration_version
    fi
done < "$manifest"

if [ "$migration_count" -eq 0 ]; then
    echo "production migration manifest is empty" >&2
    exit 1
fi

"$goose_bin" -dir "$temporary_dir" "$goose_driver" "$database_url" up

version_output=$("$goose_bin" -dir "$temporary_dir" "$goose_driver" "$database_url" version 2>&1)
current_version=${version_output##* }
case "$current_version" in
    ""|*[!0-9]*)
        echo "unable to verify production migration version: $version_output" >&2
        exit 1
        ;;
esac

if [ "$current_version" -ne "$expected_version" ]; then
    echo "production database version $current_version does not match allowed version $expected_version" >&2
    exit 1
fi
