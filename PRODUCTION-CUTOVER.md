# PostgreSQL production cutover

## Purpose

This runbook converts one existing spcase PostgreSQL database to the v1.0.0
three-role model and deploys the tested application revision. It is an operator
procedure for a separately approved maintenance window; it is not authorization
to access or change production.

Execution order is: approve and record the release, quiesce writes, create and
verify a fresh backup, run the ownership/ACL cutover, migrate to version 5,
start the application, verify it, and enter the approved observation period.

## Scope

Object ownership and ACL mutations are limited to the explicitly selected
database and its `public` schema. PostgreSQL roles are cluster-wide: the
procedure creates or normalizes only the fixed roles `spcase_migrator` and
`spcase_app`, then transfers non-extension application objects to
`spcase_migrator` and applies documented runtime ACLs. Confirm that those fixed
roles are not shared with unrelated databases in the cluster. The procedure
does not drop, disable, or rotate the legacy role and does not remove
transitional secrets.

Fresh-volume initialization and disposable rehearsals are separate workflows.
Do not run this procedure against a fresh empty volume or a rehearsal database.

## Roles and responsibilities

| Role | Responsibility | Required access | Failure action | Required approval |
| --- | --- | --- | --- | --- |
| Cutover operator | Executes this runbook and retains command evidence | Release host, Docker Compose, approved secrets | Stops at the first failed gate and preserves output | Maintenance-window and cutover approval |
| Database administrator | Verifies ownership, backup, restore capacity, cutover output and database health | PostgreSQL superuser and backup destination | Diagnoses database failures; prepares clean rollback target | Backup/restore and database-change approval |
| Application verifier | Records baseline and runs smoke/business checks | Public endpoint and test accounts for each required role | Reports the first failing workflow without changing data outside the approved probes | Smoke-test plan approval |
| Rollback decision owner | Makes the explicit fix-forward or rollback decision | All evidence and incident status | Orders write freeze continuation, rollback, or abort | Named decision authority before maintenance |
| Incident communicator | Maintains the deployment/incident timeline | Status channel and retained evidence | Announces impact, decision and recovery state | Communication plan approval |

One person may fill more than one label only when that assignment is explicitly
approved. Record the approved assignees outside this repository.

## Required access

- a clean checkout of the tested release revision;
- Docker Engine and Docker Compose access on the deployment host;
- PostgreSQL network access as the existing administrative superuser;
- access to the current Compose environment and the approved secret manager;
- write access to an encrypted backup destination with sufficient free space;
- a clean rollback database/cluster or restore environment;
- application test accounts and production-safe smoke-test data.

## Required secrets

Load secrets from the approved secret manager without placing values in shell
history, logs, this document, or the Git tree. The procedure requires:

- `POSTGRES_ADMIN_PASSWORD` for bootstrap/restore administration;
- `DB_MIGRATOR_PASSWORD` for `spcase_migrator`;
- `DB_APP_PASSWORD` for `spcase_app`;
- the retained legacy database credential for rollback;
- application secrets already required by the production environment.

The three PostgreSQL passwords must differ. Before execution, require the
variables without printing them:

```bash
: "${POSTGRES_ADMIN_PASSWORD:?missing admin secret}"
: "${DB_MIGRATOR_PASSWORD:?missing migrator secret}"
: "${DB_APP_PASSWORD:?missing application secret}"
```

## Pre-maintenance checks

1. Record `<TESTED_COMMIT_SHA>`, image digests, current application health,
   migration version, database/schema/object owners, current database size and
   the approved maintenance ticket.
2. Confirm all five operator labels, the rollback target, backup retention,
   observation-period placeholder and decision channel.
3. In the release checkout, verify the exact revision and a clean artifact:

   ```bash
   test "$(git rev-parse HEAD)" = '<TESTED_COMMIT_SHA>'
   git diff --quiet
   git diff --cached --quiet
   test -z "$(git ls-files --others --exclude-standard)"
   docker compose --env-file '<SECURE_PRODUCTION_ENV_FILE>' config --quiet
   ```

4. Set only reviewed non-secret targets, then load secrets through the approved
   mechanism:

   ```bash
   export PGHOST='<POSTGRES_HOST>'
   export PGPORT='<POSTGRES_PORT>'
   export PGSSLMODE='<APPROVED_SSL_MODE>'
   export POSTGRES_ADMIN_USER='<EXISTING_SUPERUSER>'
   export SPCASE_CUTOVER_DATABASE='<APPLICATION_DATABASE>'
   export SPCASE_LEGACY_DB_ROLE='<CURRENT_DATABASE_OWNER_ROLE>'
   export SPCASE_ENV_FILE='<SECURE_PRODUCTION_ENV_FILE>'
   export BACKUP_DIR='<APPROVED_ENCRYPTED_BACKUP_DIRECTORY>'
   ```

5. Verify destination capacity and administrative connectivity:

   ```bash
   df -Pk "$BACKUP_DIR"
   PGPASSWORD="$POSTGRES_ADMIN_PASSWORD" psql --no-psqlrc --no-password \
     --host "$PGHOST" --port "$PGPORT" --username "$POSTGRES_ADMIN_USER" \
     --dbname "$SPCASE_CUTOVER_DATABASE" --set=ON_ERROR_STOP=1 \
     --command 'SELECT current_database(), current_user'
   ```

6. Have the database administrator compare owners and migration version with
   the recorded baseline. Supported pre-cutover production versions are 2, 4
   and 5; any applied development seed migration 3 is rejected. Unexpected
   schemas, owners or versions are `NO-GO` and require review.

## Write freeze or service quiescence

1. Announce the start of maintenance.
2. Stop every application instance, worker and external writer. For the tracked
   single-host Compose deployment:

   ```bash
   docker compose --env-file "$SPCASE_ENV_FILE" stop nginx app
   ```

3. As the database administrator, verify that no legacy/application writer
   sessions remain. Do not terminate unknown sessions without separate approval.
4. Keep writes frozen through backup, cutover, migration and smoke verification.
   Inability to quiesce writes is `NO-GO`.

## Backup procedure

Create a new custom-format backup after writes are frozen. Use a run-specific
path on the approved destination:

```bash
umask 077
BACKUP_PATH="$BACKUP_DIR/spcase-pre-cutover-<UTC_TIMESTAMP>.dump"
PGPASSWORD="$POSTGRES_ADMIN_PASSWORD" pg_dump --no-password \
  --host "$PGHOST" --port "$PGPORT" --username "$POSTGRES_ADMIN_USER" \
  --dbname "$SPCASE_CUTOVER_DATABASE" --format=custom \
  --file "$BACKUP_PATH"
```

Do not reuse or overwrite a previous backup path.

## Backup verification

```bash
test -s "$BACKUP_PATH"
pg_restore --list "$BACKUP_PATH" >/dev/null
sha256sum "$BACKUP_PATH" > "$BACKUP_PATH.sha256"
sha256sum --check "$BACKUP_PATH.sha256"
wc -c "$BACKUP_PATH"
```

Record the path, byte size, SHA-256, `pg_restore --list` result, PostgreSQL
version and retention policy. The database administrator must confirm that the
archive can be restored into the available clean rollback target. Any backup,
checksum, listing or restore failure is `NO-GO`.

## Cutover command

Run exactly once as the cutover operator while writes remain frozen:

```bash
PGHOST="$PGHOST" \
PGPORT="$PGPORT" \
PGSSLMODE="$PGSSLMODE" \
POSTGRES_ADMIN_USER="$POSTGRES_ADMIN_USER" \
POSTGRES_ADMIN_PASSWORD="$POSTGRES_ADMIN_PASSWORD" \
SPCASE_CUTOVER_DATABASE="$SPCASE_CUTOVER_DATABASE" \
SPCASE_LEGACY_DB_ROLE="$SPCASE_LEGACY_DB_ROLE" \
DB_MIGRATOR_PASSWORD="$DB_MIGRATOR_PASSWORD" \
DB_APP_PASSWORD="$DB_APP_PASSWORD" \
SPCASE_CONFIRM_EXISTING_DB_CUTOVER=YES \
./scripts/cutover-postgres-roles.sh
```

Retain complete stdout/stderr. The object/ACL stage is transactional. Database
ownership is transferred after that transaction; a failure at or after this
boundary may leave transferred objects or database ownership in place even
though the script reports failure. Do not infer success from partial output.
Use the final validation result and operator gates. Re-running does not rotate
existing role passwords and requires the current target-role credentials.

## Migration command

Apply only `migrations/production.txt` as `spcase_migrator`:

```bash
docker compose --env-file "$SPCASE_ENV_FILE" run --rm migrator
docker compose --env-file "$SPCASE_ENV_FILE" run --rm migrator
```

The first run must reach Goose version 5; the second must report no pending
migrations. Do not run `goose down`, `make migrate-down`, or migration `00003`.
Migration 00005 Down restores the version-4 Goose sequence ACL and is not the
production rollback mechanism.

## Application startup

```bash
docker compose --env-file "$SPCASE_ENV_FILE" up --detach --wait
```

Application and `admin-bootstrap` must receive runtime credentials mapped from
`DB_APP_*`; migrator must receive `DB_MIGRATOR_*`. Do not add a legacy fallback.
Do not run first-admin bootstrap on an existing database unless the approved
deployment plan explicitly establishes that no administrator exists.

## Smoke tests

Run through the production ingress, not the container-local application port:

```bash
curl --fail --show-error 'https://<PRODUCTION_HOST>/api/v1/health/live'
curl --fail --show-error 'https://<PRODUCTION_HOST>/api/v1/health/ready'
curl --fail --show-error 'https://<PRODUCTION_HOST>/api/v1/info'
```

The application verifier must also check, with approved accounts and safe data:

- participant and jury authentication;
- one protected admin route;
- team read and an approved reversible team operation;
- submission read/update under the approved smoke record;
- evaluation read/write and lifecycle control if approved for the window;
- logout and subsequent token rejection.

As the database administrator, verify `pg_stat_activity` shows application
sessions as `spcase_app`, never `postgres`, the legacy owner or
`spcase_migrator`. Verify version 5, migrator ownership, exact runtime ACLs and
no Goose metadata access from `spcase_app`.

## Go/no-go criteria

`GO` requires every item below:

- approved maintenance window and confirmed operator roles;
- verified database-admin access and known expected owners;
- sufficient approved backup destination space;
- a successful fresh backup, recorded checksum and successful
  `pg_restore --list`;
- an available rollback target/restore environment and demonstrated ability to
  restore the archive;
- recorded tested repository revision and image digests;
- required secrets available through the approved channel;
- recorded production health and data baseline;
- no unexplained Git diff or untracked file in the release artifact;
- successful cutover post-validation, migration version 5, readiness and smoke
  tests.

`NO-GO` applies to backup failure, unreadable backup, missing checksum, failed
admin connection, unknown database owner, unexpected object owners or schema,
unexpected migration version, missing or shared secrets, insufficient disk
space, inability to quiesce writes, unresolved failed rehearsal, inability to
restore the backup, or any unexplained release diff.

## Rollback triggers

Immediate rollback is required when safe runtime cannot be established:

- ownership/ACL post-validation failure after mutation;
- migration failure that cannot be safely and immediately re-run;
- application readiness or authentication failure;
- failure of a critical protected route;
- unexpected pre-existing-data fingerprint change;
- unrecoverable elevated HTTP 5xx;
- inability to connect as `spcase_app`;
- evidence that runtime uses an administrative, migrator or legacy role.

A transient external dependency, ingress configuration error, or application
container issue may be fixed forward only when the database gates, data
fingerprint and role identity are valid, the rollback decision owner approves,
and the maintenance window remains open. A cutover-script failure before any
mutation may be corrected and retried after database-administrator review. A
failure at or after mutation is not automatically retryable. Do not automate
the decision.

## Rollback procedure

1. The rollback decision owner explicitly orders rollback; record the trigger.
2. Keep writes frozen and stop the new runtime:

   ```bash
   docker compose --env-file "$SPCASE_ENV_FILE" stop nginx app
   ```

3. Preserve failed-deployment logs and database evidence. Do not run migration
   Down and do not try to reverse ownership ad hoc.
4. The database administrator provisions a clean replacement database or clean
   volume, with the retained legacy role and credential. Never restore over the
   failed database.
5. Verify the recorded checksum again, then restore:

   ```bash
   sha256sum --check "$BACKUP_PATH.sha256"
   PGPASSWORD="$POSTGRES_ADMIN_PASSWORD" pg_restore --no-password \
     --host '<ROLLBACK_POSTGRES_HOST>' --port '<ROLLBACK_POSTGRES_PORT>' \
     --username "$POSTGRES_ADMIN_USER" --dbname '<CLEAN_ROLLBACK_DATABASE>' \
     --exit-on-error --no-owner --role="$SPCASE_LEGACY_DB_ROLE" "$BACKUP_PATH"
   ```

6. Verify the restored migration version, data fingerprint/counts and legacy
   ownership against the pre-cutover evidence.
7. Deploy the recorded previous Compose revision explicitly with the retained
   legacy `DB_USER`/`DB_PASSWORD`, pointing only to the clean restored target.
8. Repeat health, authentication and critical-route smoke tests before reopening
   writes. Preserve the failed database for investigation according to policy.

## Observation period

After all GO gates, reopen writes and observe for
`<APPROVED_OBSERVATION_PERIOD>`. The duration must be approved outside this
document. During the period, verify:

- availability and readiness;
- authentication, admin route and key application workflows;
- PostgreSQL errors and unexpected permission-denied errors;
- container health and restart counts;
- HTTP 5xx rate and latency signals;
- data-integrity signals and pre-existing-record checks;
- backup artifact integrity, access and retention.

Legacy credentials and the legacy role remain available until the rollback
decision owner explicitly closes observation.

## Post-cutover verification

At observation close, retain evidence that:

- database, `public` schema, application objects and Goose metadata are owned by
  `spcase_migrator`;
- `spcase_app` has only the documented DML/type/schema privileges, no DDL,
  ownership or Goose access;
- application sessions use `spcase_app` and migrations use
  `spcase_migrator`;
- Goose version is 5 with no pending production migrations;
- application data, health, authentication and protected workflows are valid;
- restart counts, PostgreSQL errors and HTTP 5xx remained within approved gates;
- the verified backup remains retained and readable.

## Legacy credential retirement criteria

Retirement is a later, separately authorized change. It requires closed and
accepted observation, an approved post-cutover backup, no runtime or deployment
reference to legacy credentials, a tested recovery path that no longer depends
on the legacy role, and explicit database-administrator and rollback-owner
approval. Only then may transitional `DB_USER`/`DB_PASSWORD` be removed and the
legacy role disabled or dropped through a reviewed administrative procedure.

## Evidence to retain

- approved window, role assignments and decision timeline;
- tested commit SHA, image digests and clean-tree checks;
- redacted environment/config validation;
- pre/post owners, ACLs, migration version and data fingerprints;
- backup path, size, checksum, listing, restore test and retention metadata;
- full cutover/migrator output without secrets;
- container role-isolation evidence and `pg_stat_activity` identity;
- health, smoke, observation and restart evidence;
- go/no-go and any rollback decision, commands and verification results.
