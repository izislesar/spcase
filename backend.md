# spcase v1.0.0 — Backend Implementation

## 1. Stack

- Go `net/http` and `http.ServeMux`;
- PostgreSQL through `pgx/v5` and `pgxpool`;
- bcrypt password hashes;
- HS256 JWT through `golang-jwt/jwt/v5`;
- UUID identities;
- XLSX export through `excelize`;
- embedded HTML/templates/static assets;
- Docker Compose and Nginx.

## 2. Repository structure

```text
cmd/
  app/                 application and dependency wiring
  healthcheck/         readiness probe binary
  admin-bootstrap/     one-time first ADMIN creation
internal/
  config/              .env/ENV loading and validation
  domain/              entities, rules and stable errors
  pkg/postgres/        pgxpool setup
  repository/          PostgreSQL implementations
  service/             use cases and validation
  delivery/http/
    middleware/        cross-cutting HTTP controls
    v1/                DTO, handlers and response mapping
migrations/            Goose schema, indexes and dev seed
scripts/               production migration wrapper
web/
  template/            server-rendered pages
  src/                 Tailwind/JavaScript sources
  static/              compiled embedded assets
```

## 3. Configuration

Application reads optional local `.env`; existing environment variables win.

Required runtime values:

- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`;
- `JWT_SECRET`, `JURY_REGISTRATION_KEY`;
- `CORS_ALLOWED_ORIGINS` containing HTTPS origins;
- `REGISTRATION_DEADLINE`, `SUBMISSION_DEADLINE` in RFC3339;
- `NO_TEAM_TELEGRAM_URL` as HTTPS URL.

Optional/defaulted values:

- `PORT=8000`;
- `APP_DOMAIN=spcase.ru`;
- `DB_STATEMENT_TIMEOUT=15s`;
- `DB_LOCK_TIMEOUT=5s`.

Secrets must be at least 32 characters, have sufficient diversity and must not match known placeholder/weak patterns. Registration deadline must precede submission deadline.

Compose database provisioning additionally requires `POSTGRES_ADMIN_PASSWORD`, fixed role names `DB_MIGRATOR_USER=spcase_migrator` and `DB_APP_USER=spcase_app`, and separate `DB_MIGRATOR_PASSWORD`/`DB_APP_PASSWORD`. Эти credentials не читаются Go configuration напрямую. Compose подставляет migrator credentials в `DB_USER`/`DB_PASSWORD` только migration container, а application credentials — только runtime service и запускаемому через него `admin-bootstrap`. Runtime service получает явно перечисленную конфигурацию вместо полного env-файла, поэтому administrative и migrator passwords в его environment отсутствуют. Legacy `DB_USER`/`DB_PASSWORD` сохранены только для rollback/direct-process compatibility и normal Compose не используются.

## 4. Startup and shutdown

`cmd/app` loads configuration, creates and pings a PostgreSQL pool, constructs all dependencies, parses embedded templates and starts HTTP. Startup is bounded by 10 seconds.

Pool defaults: max 10, min 2 connections, one-hour lifetime, 30-minute idle time and one-minute health checks. PostgreSQL session statement/lock timeouts are configured on every connection.

HTTP limits: 5s header read, 15s read, 30s write, 60s idle and 1 MiB headers. SIGINT/SIGTERM initiates graceful shutdown with a 15-second timeout.

## 5. Authentication

Participant passwords are 8..72 bytes and stored using bcrypt default cost. Email is normalized to lowercase and validated.

JWT claims contain user ID, role and `auth_version`; lifetime is 24 hours and algorithm is restricted to HS256. Authentication middleware always reloads the current account projection. Password change, account enable/disable, role change or explicit auth-version increment invalidate prior tokens through projection mismatch.

Logout only expires the cookie in the browser. It does not perform server-side token revocation.

Participant registration stops at `REGISTRATION_DEADLINE`. Jury registration uses a SHA-256-backed constant-time comparison with `JURY_REGISTRATION_KEY`. General login accepts USER/ADMIN; jury login accepts JURY.

The first ADMIN is not created over HTTP. `cmd/admin-bootstrap` reads the password from stdin and uses a PostgreSQL advisory lock to allow exactly one administrator.

## 6. Services and repositories

Services own validation and orchestration:

- auth/user identity;
- team lifecycle and displayed state;
- submission deadline and URL rules;
- jury workspace and score validation;
- evaluation lifecycle/admin statistics;
- XLSX export;
- embedded public schedule/FAQ content.

Repositories execute direct SQL:

- user/account state;
- teams and membership transactions;
- submissions;
- evaluation batches/totals;
- read models, lifecycle state and exports.

Team and submission writes use database transactions and row locks. Business deadlines are checked in service for fast rejection and again against PostgreSQL time inside critical transactions where applicable.

## 7. HTTP middleware

- `AuthMiddleware` validates token plus current account state.
- `RequireRoles` enforces RBAC.
- `HardLockMiddleware` blocks selected team mutation routes.
- `CORSMiddleware` permits credentials only from configured HTTPS origins.
- `RecoveryMiddleware` converts panics to opaque errors.
- `APIErrorResponses` normalizes ServeMux 404/405.
- `SecurityHeaders` sets nosniff, DENY framing, referrer and permissions policies.
- `NoStoreSensitiveResponses` disables caching for auth/user/team/jury/admin API.

JSON decoder requires one object, rejects unknown fields and limits request bodies to 1 MiB. Internal errors are logged but returned as stable opaque JSON.

## 8. Frontend delivery

Templates and built assets are embedded in the Go binary. Templates are parsed during startup. HTML uses `no-store`; static assets use immutable one-year caching and a content-derived version query.

The Docker frontend stage runs `npm ci`, Tailwind and esbuild. Nginx serves the same compiled static assets directly; all other routes are proxied to Go.

## 9. Deployment components

- PostgreSQL 16 with SCRAM host auth, checksums, healthcheck and named volume.
- Fresh-volume role initialization with `postgres` for cluster administration, `spcase_migrator` as database/schema owner and `spcase_app` for runtime DML.
- One-shot migrator gated by database health.
- Distroless nonroot app with compiled healthcheck.
- Nonroot Nginx with login/registration rate limits, request IDs, gzip and proxy timeouts.

App/PostgreSQL are isolated on the internal backend network. Nginx alone binds to host loopback. External TLS termination is deliberately outside Compose.

Production migrations use an explicit allowlist and never apply development seed data. Application startup does not run migrations. `spcase_migrator` owns schema objects and is the Goose identity; `spcase_app` has no DDL or ownership privileges, receives explicit per-table grants from `00004_grant_runtime_privileges.sql` and is denied access to Goose metadata by `00005_isolate_goose_metadata.sql`. Application and first ADMIN bootstrap both use `spcase_app` in fresh Compose deployments.

Existing databases are converted only by the explicit administrative
`scripts/cutover-postgres-roles.sh`; it is not part of application startup or
fresh-volume init. The tool requires a confirmed target database and legacy
owner, validates all pre-cutover owners, transfers non-extension `public`
relations/routines/types, applies exact runtime/default ACLs and performs
post-validation. `scripts/rehearse-postgres-role-cutover.sh` proves the procedure
against a separate custom-format backup restore, including the full integration
suite and an idempotent second run. The opt-in
`scripts/rehearse-existing-db-deployment.sh` adds tracked Compose application,
admin-bootstrap, Nginx ingress, runtime smoke/restart and an independent rollback
restore to the same disposable legacy-to-version-5 path. Existing installations
must complete that manual cutover before deploying the tracked `DB_APP_*` runtime wiring. Legacy
credentials and role remain available only for an explicitly reverted Compose
configuration until production smoke tests pass.

## 10. Testing

Default verification:

```bash
go test -race -count=1 ./...
go vet ./...
go build ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.6.0 ./...
npm audit
```

`make security-check` runs tests, vet, vulnerability scan and npm audit.

PostgreSQL integration suite:

```bash
SPCASE_TEST_MIGRATOR_DATABASE_URL='postgres://spcase_migrator:...' \
SPCASE_TEST_APP_DATABASE_URL='postgres://spcase_app:...' \
go test -race -count=1 -tags=integration ./internal/...
```

The source database must already be at production migration version 5. The harness verifies both connection roles, creates/migrates/cleans an isolated schema as `spcase_migrator`, and constructs every repository with the `spcase_app` pool. Runtime grants for the isolated objects are copied from the already migrated `public` schema, so tests cannot silently broaden application privileges. It covers schema integrity, migrations/seed, concurrent joins, lock ordering, database deadline checks, submission invalidation, evaluation atomicity, query plans, aggregation, timeouts and concurrent ADMIN bootstrap.

Fresh-database role ACL check:

```bash
SPCASE_TEST_MIGRATOR_DATABASE_URL='postgres://spcase_migrator:...' \
SPCASE_TEST_APP_DATABASE_URL='postgres://spcase_app:...' \
go test -tags=integration ./internal/repository -run '^TestPostgresRolePrivileges$'
```

`SPCASE_TEST_DATABASE_URL` is rejected when supplied alone because it would collapse the production privilege boundary.

Infrastructure validation uses `docker compose --env-file .env.production config --quiet`, followed by image build and runtime smoke tests described in README.
