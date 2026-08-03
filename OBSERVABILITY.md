# spcase — Observability Baseline

This document is the operator reference for the minimum production
observability baseline. It describes what signals exist, where they come from,
what an operator should conclude from each one, and what is still pending
external decisions.

Nothing in this document claims centralized monitoring, external alert
delivery, or long-term log retention. Those remain open production approvals
listed at the end.

## 1. Signal inventory

Classification legend:

- **implemented** — a concrete mechanism exists in this repository now;
- **logs** — derivable from existing structured application or Nginx logs;
- **docker** — available from the Docker engine or Compose without changes;
- **external** — requires a future external monitoring/alerting service;
- **ingress-blocked** — cannot exist before the production TLS ingress is selected.

| Signal | Status | Source |
|---|---|---|
| Application process alive | implemented | `GET /api/v1/health/live` |
| Application readiness | implemented | `GET /api/v1/health/ready` + container health checks |
| PostgreSQL connectivity | implemented | readiness ping (2s timeout), `pg_isready` health check |
| PostgreSQL pool acquisition failures | logs | `database_readiness_failed` events + `http_request_completed` with `status >= 500` |
| Migration success or failure | implemented | migrator exit code, `migration_failed` events, Compose `depends_on: service_completed_successfully` gating |
| HTTP request count | logs | `http_request_completed` events, Nginx access log |
| HTTP response status classes | logs | `status` field in `http_request_completed`, Nginx `$status` |
| HTTP request latency | logs | `duration_ms` in `http_request_completed`, Nginx `$request_time` / `$upstream_response_time` |
| Panic recovery count | logs | `panic_recovered` events |
| Authentication failure count | logs | `http_request_completed` with `status = 401` on auth routes; Nginx 401/429 counts |
| Container restart count | docker | `docker inspect -f '{{.RestartCount}}'`, `docker compose ps` |
| Disk usage | docker / external | host `df`, `docker system df`; alerting requires external monitoring |
| Backup age | implemented (checker) | `scripts/check-backup-freshness.sh` against a backup evidence manifest; the production manifest producer is pending the backup destination decision |
| TLS certificate expiry | ingress-blocked | must be monitored on the external TLS ingress once selected |

## 2. Health endpoint semantics

- `GET /api/v1/health/live` answers only whether the Go process can respond.
  It never touches PostgreSQL or any external service and always returns
  `200 {"status":"ok","timestamp":...}` while the process runs.
- `GET /api/v1/health/ready` answers whether the application can safely serve
  traffic. It pings PostgreSQL with a bounded 2-second timeout and returns
  `200 {"status":"ready",...}` on success or `503` with the stable
  `{"error":{"code":"NOT_READY",...}}` contract on failure. Database or network
  error details are never included in the response body.
- Both endpoints are unauthenticated, read-only, and cheap: liveness performs
  no I/O; readiness is at most one bounded ping per call.
- The container health check (`cmd/healthcheck`) probes readiness, so an
  unhealthy container means "cannot serve traffic", not "process dead".
- During a PostgreSQL outage, readiness returns 503 and liveness keeps
  returning 200. This distinction is verified by
  `scripts/rehearse-observability.sh`.

## 3. Structured application logs

Application logs are JSON (`slog`) on stdout. Request-scoped entries contain:

```text
time, level, msg, event, request_id, method, route, status, duration_ms
```

- `route` is the registered Go route template (for example
  `GET /api/v1/team/create`) or the fixed value `unmatched`. Raw URL paths are
  deliberately not logged by the request logger to keep cardinality bounded;
  the Nginx access log remains the source for raw paths.
- `request_id` correlates every request log with the Nginx access log (Nginx
  forwards its `$request_id` as `X-Request-ID`).

Stable event names:

| Event | Level | Meaning |
|---|---|---|
| `http_server_started` | INFO | HTTP listener is up |
| `http_request_completed` | INFO (`status < 500`), ERROR (`status >= 500`) | one per HTTP request |
| `panic_recovered` | ERROR | handler panic converted to opaque 500; logs `panic_type` only, never the panic value |
| `database_startup_failed` | ERROR | pool creation/ping failed at startup; process exits |
| `database_readiness_failed` | WARN | readiness ping failed |
| `database_pool_closed` | INFO | PostgreSQL pool closed during shutdown |
| `graceful_shutdown_started` | INFO | SIGTERM/SIGINT received, draining begins |
| `graceful_shutdown_completed` | INFO | HTTP server drained and stopped within the 15s timeout |
| `graceful_shutdown_failed` | ERROR | drain exceeded the timeout; connections were force-closed |
| `migration_failed` | stderr (migrator) | production migration script failure; migrator exits nonzero and the app never starts |
| `backup_freshness_failed` | stderr (checker) | backup manifest missing, invalid, unverified, or stale |

Sensitive-data rules enforced by design: no passwords, jury keys, raw JWTs,
cookies, database URLs, request bodies, or panic values are ever logged. The
rehearsal script asserts that no deployment secret appears in application or
Nginx logs.

## 4. Request correlation

`internal/delivery/http/middleware/request_id.go` implements correlation:

- a valid inbound `X-Request-ID` (8–64 characters from `[A-Za-z0-9._-]`) is
  accepted; anything else — missing, malformed, oversized, or containing
  control characters — is replaced with a fresh UUIDv4;
- the ID is returned in the `X-Request-ID` response header and stored in the
  request context;
- request, error, readiness, and panic logs include `request_id`;
- the ID is a correlation token only and is never used for authentication.

Behind Nginx, the Nginx-generated `$request_id` is forwarded, so one ID joins
the Nginx access log line and all application log lines for a request.

## 5. Docker log retention

All four Compose services (`db`, `migrator`, `app`, `nginx`) use the shared
`x-log-retention` anchor in `docker-compose.yml`:

```yaml
logging:
  driver: json-file
  options:
    max-size: "10m"
    max-file: "5"
```

Expected maximum local retention is 50 MiB per service (5 files × 10 MiB),
about 200 MiB for the whole stack. These values are a baseline, not an approved
production policy; adjust only with a documented reason.

This is local rotation only. It is not centralized log storage, not long-term
retention, and not tamper-evident archival. PostgreSQL statement-level logging
(`log_statement`, `log_min_duration_statement`) is a separate database
configuration concern and is intentionally not enabled by this baseline.

## 6. Log-based operational queries

Until external monitoring exists, these queries answer the baseline questions
against the local JSON logs:

```bash
# Is the application up and ready right now?
docker compose ps                                  # health status per service
curl -s -o /dev/null -w '%{http_code}\n' http://127.0.0.1:8080/api/v1/health/ready

# Did migrations succeed?
docker compose ps --all migrator                   # must be "Exited (0)"
docker compose logs migrator | grep migration_failed

# Are containers restarting?
docker inspect -f '{{.Name}} restarts={{.RestartCount}}' $(docker compose ps -q)

# Are 5xx errors increasing?
docker compose logs app --since 1h | grep '"event":"http_request_completed"' \
  | jq -r 'select(.status >= 500) | .route' | sort | uniq -c | sort -rn

# Database readiness / pool failures?
docker compose logs app --since 1h | grep -c database_readiness_failed

# Panics?
docker compose logs app --since 24h | grep -c panic_recovered

# Authentication failure volume (via Nginx)?
docker compose logs nginx --since 1h | grep -E ' (401|429) ' | awk '{print $9}' | sort | uniq -c

# Latency distribution from application logs?
docker compose logs app --since 1h | grep '"event":"http_request_completed"' \
  | jq -r '.duration_ms' | sort -n | awk '{a[NR]=$1} END {print "p50="a[int(NR*0.5)], "p95="a[int(NR*0.95)], "max="a[NR]}'
```

## 7. Alert definitions

Thresholds marked `<...>` are placeholders: they require measurement on staging
and explicit approval before production use. No threshold below is an approved
production value.

### Availability

| Alert | Signal source | Condition | Severity | Operator action | Escalation |
|---|---|---|---|---|---|
| Liveness failure | external probe of `/api/v1/health/live` | probe fails 3 consecutive times | critical | check `docker compose ps`, `docker compose logs app`; restart the app container if the process is wedged | if restarts repeat, roll back to the previous approved image |
| Readiness failure | external probe of `/api/v1/health/ready`, container health | unhealthy longer than `<APPROVED_READINESS_GRACE>` | critical | check `database_readiness_failed` events and the `db` container health; restore PostgreSQL first, do not restart the app blindly | sustained failure after DB recovery → rollback procedure in PRODUCTION-CUTOVER.md |
| Repeated container restart | Docker `RestartCount` | increase above `<APPROVED_RESTART_THRESHOLD>` within one hour | high | inspect logs for `panic_recovered`/`database_startup_failed`; check host memory | recurring crash loop → roll back image, escalate to engineering |
| Nginx upstream failure | Nginx error log, edge 503 responses | `upstream` errors or 503 from `@api_unavailable` | high | check app container health and backend network | if app healthy but upstream fails → investigate DNS resolver/network |

### HTTP

| Alert | Signal source | Condition | Severity | Operator action | Escalation |
|---|---|---|---|---|---|
| Elevated 5xx rate | `http_request_completed` status counts | 5xx share above `<APPROVED_5XX_THRESHOLD>` over 5 minutes | high | group 5xx by `route`; check `database_readiness_failed` and `panic_recovered` correlation | sustained rise after dependency recovery → roll back |
| Sustained latency increase | `duration_ms` distribution, Nginx `$request_time` | p95 above `<APPROVED_LATENCY_THRESHOLD>` over 15 minutes | medium | check DB health, disk, slow queries, host load | growing latency with rising 5xx → treat as availability incident |
| Abnormal authentication-failure volume | 401 counts on auth routes | above `<APPROVED_AUTH_FAILURE_THRESHOLD>` | medium | review Nginx access log source IPs; confirm rate limits engage (429s) | credential-stuffing pattern → tighten edge controls, escalate security review |
| Rate-limit rejection anomaly | Nginx 429 counts | above `<APPROVED_RATE_LIMIT_THRESHOLD>` or zero 429s under obvious attack | medium | verify `limit_req` zones work; check for misconfigured legit clients | attack beyond Nginx capacity → escalate to ingress/network controls |

### PostgreSQL

| Alert | Signal source | Condition | Severity | Operator action | Escalation |
|---|---|---|---|---|---|
| Connection failure | `database_readiness_failed`, `pg_isready` health check | continuous failures | critical | check `db` container, volume, host disk | DB not recoverable in place → restore from latest verified backup |
| Pool acquisition timeout | `http_request_completed` 5xx correlated with DB errors | above `<APPROVED_POOL_FAILURE_THRESHOLD>` | high | check pool saturation (`pg_stat_activity`), slow queries, `statement_timeout` hits | persistent saturation → capacity review before any config change |
| Migration failure | migrator exit code, `migration_failed` events | any nonzero migrator exit | critical | the app does not start by design; read migrator logs; do not force-start the app | follow the failed-migration path in PRODUCTION-CUTOVER.md |
| Disk usage threshold | host `df`, `docker system df` | above `<APPROVED_DISK_THRESHOLD>` | high | free space, verify log rotation is active, check volume growth | below headroom for a backup + restore → emergency capacity action |
| Database container unhealthy | Docker health check | unhealthy longer than `<APPROVED_DB_GRACE>` | critical | `docker compose logs db`; check data volume integrity | corrupted volume → restore from latest verified backup |

### Backup

| Alert | Signal source | Condition | Severity | Operator action | Escalation |
|---|---|---|---|---|---|
| Backup older than RPO | `check-backup-freshness.sh` exit 1 (stale) | age above `<APPROVED_RPO>` | high | run the backup job manually; investigate the scheduled job | repeated missed backups → incident: data at risk |
| Backup command failure | backup job exit code / stderr | any nonzero exit | high | read job logs; verify destination access and disk | consecutive failures → treat as missed backup |
| Checksum/list verification failure | `pg_restore --list`, manifest `verified` flag | verification fails | critical | discard the corrupt archive, re-run backup immediately, check storage integrity | repeated corruption → replace backup destination |
| Restore rehearsal overdue | restore rehearsal records | older than `<APPROVED_RESTORE_REHEARSAL_INTERVAL>` | medium | schedule and run the documented restore rehearsal | skipped rehearsals block backup-plan approval |

### TLS

| Alert | Signal source | Condition | Severity | Operator action | Escalation |
|---|---|---|---|---|---|
| Certificate expiry | external ingress certificate monitoring | fewer than `<APPROVED_TLS_EXPIRY_DAYS>` days remaining | high | renew certificate via the ingress provider | failed renewal → incident before expiry |

TLS monitoring is **pending until the production ingress is selected**; the
Compose stack terminates no TLS itself.

## 8. Backup freshness interface

The production backup destination is not yet selected, so no provider
integration exists. What exists is a provider-agnostic freshness contract:

- the backup job (whatever it becomes) writes a small JSON evidence manifest
  next to each backup, atomically (write to a temporary file in the same
  directory, `chmod 600`, then `mv`):

```json
{
  "timestamp": "2026-01-02T03:04:05Z",
  "database": "spcase",
  "size_bytes": 123456,
  "sha256": "<64 lowercase hex of the archive>",
  "verified": true,
  "tool": "pg_dump 16.6"
}
```

- `verified` must only be `true` after the archive passed an integrity check
  (for example `pg_restore --list` for custom-format dumps);
- the manifest must never contain passwords, connection strings, or signed
  URLs;
- `scripts/check-backup-freshness.sh <manifest> <max-age-seconds>` validates
  format, permissions, checksum shape, verification status, and age against
  the approved RPO, exiting nonzero on any failure. Wire it into cron or the
  future alerting stack.

Production backups themselves are **not** marked implemented: destination,
schedule, encryption, retention, and RPO remain TODO section 5 approvals.

## 9. Failure-mode rehearsal

The observability failure modes are reproducibly validated by:

```bash
SPCASE_CONFIRM_OBSERVABILITY_REHEARSAL=YES ./scripts/rehearse-observability.sh
```

The rehearsal builds unique images, starts an isolated copy of the Compose
stack under a unique project, and verifies: healthy startup, liveness,
readiness, end-to-end request-ID correlation, PostgreSQL pause (readiness 503
with `database_readiness_failed`, liveness stays 200), a controlled dependency
5xx visible in structured logs, readiness recovery after restore with schema
intact, container restart visibility (fresh `StartedAt` plus repeated
`http_server_started` events; note that Docker `RestartCount` increments only
for daemon restarts after genuine process crashes, not for API-initiated
stop/kill/restart), migration failure observability with nonzero exit, SIGTERM
graceful-shutdown events with exit code 0, and absence of secrets in logs. A
cleanup trap removes only resources created by the run and verifies their
removal.

It uses only disposable Docker resources and never touches staging or
production volumes.

## 10. Graceful shutdown

On SIGTERM (also `docker stop`) the application:

1. logs `graceful_shutdown_started`;
2. stops accepting new connections;
3. lets bounded in-flight requests finish within a 15-second drain timeout
   (otherwise force-closes and logs `graceful_shutdown_failed`);
4. logs `graceful_shutdown_completed`;
5. closes the PostgreSQL pool (`database_pool_closed`) and exits with code 0.

Compose grants the app a 20-second `stop_grace_period`, which covers the
15-second drain. The rehearsal asserts the events and the clean exit code.

## 11. Metrics decision

No `/metrics` endpoint is exposed. Rationale:

- every baseline question above is answerable from the structured logs and
  Docker state that already exist;
- the application sits behind Nginx on an internal network; a metrics endpoint
  on the public mux would be proxied to the internet by the current
  `location /` rule, and restricting it safely requires the ingress decision
  that has not been made yet;
- adding a separate listener or a metrics dependency now would expand the
  architecture without a demonstrated need.

If Prometheus-style metrics are approved later, they must be served on a
separate internal listener, use only bounded labels (method, route template,
status class), and never expose user IDs, team IDs, raw paths, or secrets.

## 12. Remaining external requirements

Not implemented here, required before or during production acceptance:

- centralized log destination and long-term retention (local rotation only today);
- external alert delivery (probe runner, notification channels);
- approved thresholds for every `<...>` placeholder above;
- backup destination, schedule, encryption, retention, and approved RPO;
- TLS ingress with certificate-expiry and real-client-IP monitoring;
- staging acceptance of this baseline, then production approval.
