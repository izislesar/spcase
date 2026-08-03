#!/bin/sh

# check-backup-freshness.sh — validates a backup evidence manifest against an
# approved RPO threshold. Exits 0 when the newest backup is fresh and verified,
# 1 when the manifest is missing, invalid, unverified, or stale, and 2 on usage
# errors. It never reads database credentials and never prints manifest data
# beyond field names.
#
# Manifest contract (written atomically, chmod 600, by the backup job):
#   {
#     "timestamp":  "2026-01-02T03:04:05Z",   RFC3339 backup completion time
#     "database":   "spcase",                 logical database name
#     "size_bytes": 123456,                   positive integer archive size
#     "sha256":     "<64 lowercase hex>",     archive checksum
#     "verified":   true,                     archive verification succeeded
#     "tool":       "pg_dump 16.6"            producing tool and version
#   }
#
# Usage:
#   check-backup-freshness.sh <manifest-path> <max-age-seconds>

set -eu

if [ "$#" -ne 2 ]; then
    echo "usage: $0 <manifest-path> <max-age-seconds>" >&2
    exit 2
fi

manifest=$1
max_age=$2

fail() {
    echo "backup_freshness_failed: $*" >&2
    exit 1
}

case "$max_age" in
    ""|*[!0-9]*)
        echo "usage: max-age-seconds must be a positive integer" >&2
        exit 2
        ;;
esac
if [ "$max_age" -lt 60 ]; then
    echo "usage: max-age-seconds must be at least 60" >&2
    exit 2
fi

if ! command -v jq >/dev/null 2>&1; then
    echo "jq is required for backup freshness validation" >&2
    exit 2
fi

[ -f "$manifest" ] || fail "manifest not found: $manifest"
[ ! -L "$manifest" ] || fail "manifest must not be a symlink: $manifest"

permissions=$(stat -c '%a' "$manifest")
case "$permissions" in
    *00) ;;
    *)
        fail "manifest permissions $permissions allow group/other access; expected 600 or stricter"
        ;;
esac

if ! jq empty "$manifest" >/dev/null 2>&1; then
    fail "manifest is not valid JSON"
fi

field_check=$(jq -r '
    def is_sha256: type == "string" and test("^[0-9a-f]{64}$");
    if (.timestamp | type) != "string" then "timestamp"
    elif (.database | type) != "string" or (.database | gsub("\\s"; "")) == "" then "database"
    elif (.size_bytes | type) != "number" or .size_bytes < 1 or (.size_bytes | floor) != .size_bytes then "size_bytes"
    elif (.sha256 | is_sha256) | not then "sha256"
    elif .verified != true then "verified"
    elif (.tool | type) != "string" or (.tool | gsub("\\s"; "")) == "" then "tool"
    else "" end
' "$manifest")
if [ -n "$field_check" ]; then
    fail "manifest field is missing or invalid: $field_check"
fi

timestamp=$(jq -r '.timestamp' "$manifest")
if ! backup_epoch=$(date --date="$timestamp" +%s 2>/dev/null); then
    fail "manifest timestamp is not a valid RFC3339 datetime"
fi
now_epoch=$(date +%s)
age=$((now_epoch - backup_epoch))
if [ "$age" -lt -300 ]; then
    fail "manifest timestamp is more than 5 minutes in the future"
fi
if [ "$age" -gt "$max_age" ]; then
    fail "latest verified backup is ${age}s old, exceeding the ${max_age}s threshold"
fi

echo "backup freshness OK: verified backup age ${age}s within ${max_age}s threshold"
