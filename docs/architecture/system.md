# spcase v1.0.0 — System Architecture

## 1. Назначение и границы

spcase — монолитная платформа кейс-чемпионата. Один Go-процесс обслуживает REST API и server-rendered HTML, PostgreSQL хранит состояние, Nginx является единственной точкой входа Compose. Redis, message broker и отдельные application-сервисы отсутствуют.

## 2. Runtime topology

```text
Browser
  │ HTTPS
  ▼
External TLS ingress / tunnel
  │ HTTP to 127.0.0.1:8080
  ▼
Nginx container
  ├── /static/* ──► compiled static assets
  └── pages + /api/v1/* ──► app:8000
                                  │
                                  ▼
                           PostgreSQL:5432
                           ├── postgres (bootstrap/admin)
                           ├── spcase_migrator (schema owner)
                           └── spcase_app (runtime DML)
```

Compose не реализует публичный TLS. Nginx публикуется только на `127.0.0.1:${NGINX_PORT:-8080}`; сертификаты, DNS и внешний HTTPS ingress находятся вне репозитория. Порты приложения и PostgreSQL на host не публикуются.

## 3. Application layers

```text
net/http handler
      ↓
service
      ↓
repository
      ↓
pgxpool / PostgreSQL
```

- `internal/domain` — роли, membership state, команды, submissions, evaluations, отчётные модели и стабильные domain errors.
- `internal/delivery/http/v1` — DTO, JSON parsing, HTTP status mapping и handlers.
- `internal/delivery/http/middleware` — authentication, RBAC, Hard Lock, CORS, recovery, API errors, security headers и cache policy.
- `internal/service` — validation и бизнес-правила.
- `internal/repository` — SQL, транзакции, row locking и преобразование PostgreSQL errors.
- `internal/pkg/postgres` — конфигурация `pgxpool`, connect/ping и session timeouts.

Handlers не выполняют SQL. Services не зависят от HTTP. Repositories не формируют HTTP responses.

### Stack

- Go `net/http` and `http.ServeMux`;
- PostgreSQL through `pgx/v5` and `pgxpool`;
- bcrypt password hashes;
- HS256 JWT through `golang-jwt/jwt/v5`;
- UUID identities;
- XLSX export through `excelize`;
- embedded HTML/templates/static assets;
- Docker Compose and Nginx.

### Repository structure

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

### Configuration

Application reads optional local `.env`; existing environment variables win.

Required runtime values:

- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`;
- `JWT_SECRET`, `JURY_REGISTRATION_KEY`;
- `CORS_ALLOWED_ORIGINS` containing HTTPS origins;
- `REGISTRATION_DEADLINE`, `SUBMISSION_DEADLINE` in RFC3339;
- `NO_TEAM_TELEGRAM_URL` as HTTPS URL.

Optional/defaulted values: `PORT=8000`, `APP_DOMAIN=spcase.ru`, `DB_STATEMENT_TIMEOUT=15s`, `DB_LOCK_TIMEOUT=5s`.

Secrets must be at least 32 characters, have sufficient diversity and must not match known placeholder/weak patterns. Registration deadline must precede submission deadline.

Compose database provisioning additionally requires `POSTGRES_ADMIN_PASSWORD`, fixed role names `DB_MIGRATOR_USER=spcase_migrator` and `DB_APP_USER=spcase_app`, and separate `DB_MIGRATOR_PASSWORD`/`DB_APP_PASSWORD`. Эти credentials не читаются Go configuration напрямую. Compose подставляет migrator credentials в `DB_USER`/`DB_PASSWORD` только migration container, а application credentials — только runtime service и запускаемому через него `admin-bootstrap`. Runtime service получает явно перечисленную конфигурацию вместо полного env-файла, поэтому administrative и migrator passwords в его environment отсутствуют. Legacy `DB_USER`/`DB_PASSWORD` сохранены только для rollback/direct-process compatibility и normal Compose не используются.

### Startup, pool and HTTP limits

`cmd/app` loads configuration, creates and pings a PostgreSQL pool, constructs all dependencies, parses embedded templates and starts HTTP. Startup is bounded by 10 seconds.

Pool defaults: max 10, min 2 connections, one-hour lifetime, 30-minute idle time and one-minute health checks. PostgreSQL session statement/lock timeouts are configured on every connection.

HTTP limits: 5s header read, 15s read, 30s write, 60s idle and 1 MiB headers. SIGINT/SIGTERM initiates graceful shutdown with a 15-second timeout.

## 4. Repositories and services

Repositories:

- `UserPostgres` — accounts, account projection, auth version, disabled state, first ADMIN.
- `TeamPostgres` — teams, membership и транзакционные mutations.
- `SubmissionPostgres` — получение и atomic upsert решения.
- `ScorePostgres` — atomic batch upsert, team-scoped lifecycle locking и чтение evaluations/totals.
- `QueryPostgres` — jury workspace, evaluation state, admin stats и exports.

Services:

- `AuthService`, `UserService`, `TeamService`, `SubmissionService`, `ScoreService`, `JuryService`, `AdminService`, `ExportService`, `PublicService`.
- `AdminBootstrapService` используется отдельным CLI, а не HTTP API.

## 5. HTTP composition

`cmd/app` создаёт repositories, services и handlers вручную и регистрирует маршруты в `http.ServeMux`. Глобальная цепочка:

```text
SecurityHeaders
  → NoStoreSensitiveResponses
  → RequestID
  → RequestLogging
  → CORS
  → Recovery
  → APIErrorResponses
  → ServeMux
```

На защищённых маршрутах дополнительно применяются `AuthMiddleware` и `RequireRoles`. Для leave/kick/transfer/disband используется `HardLockMiddleware`; repository повторно проверяет lock по времени PostgreSQL после получения row locks.

Submission, scoring и destructive team mutations разделяют team row как lifecycle lock. Scoring использует порядок team → evaluation state → submission; membership paths начинают с той же team row, затем блокируют users в UUID-порядке. Это не сериализует независимые команды. Evaluations принадлежат team как исторические записи: invalidation текущего submission их сохраняет, disband удаляет каскадно.

HTTP server имеет read/write/header/idle timeouts, лимит заголовков и graceful shutdown до 15 секунд. Логи приложения — JSON через `slog` в stdout. `RequestID` принимает строго валидный входящий `X-Request-ID` (8–64 символа из `[A-Za-z0-9._-]`, в том числе `$request_id` от Nginx) либо генерирует UUIDv4, возвращает его в response header и помещает в request context. `RequestLogging` пишет одно событие `http_request_completed` на запрос с полями `request_id`, `method`, `route` (шаблон маршрута, не сырой path), `status` и `duration_ms`; `status >= 500` логируется на уровне ERROR. Operational runbook со стабильными event names и alert definitions — `../runbooks/observability.md`.

## 6. Authentication and authorization

JWT подписывается HS256 и живёт 24 часа. Claims: subject user UUID, role, `auth_version`, `iat`, `nbf`, `exp`. Team identity в token не хранится.

`access_token` устанавливается с `HttpOnly`, `Secure`, `SameSite=Lax`, `Path=/` и `Domain=APP_DOMAIN`. На каждом защищённом запросе middleware:

1. проверяет единственную cookie и JWT;
2. загружает актуальные role, `auth_version` и `disabled_at`;
3. отклоняет удалённый/disabled account, смену role и устаревшую auth version;
4. помещает immutable principal в request context.

RBAC-роли: `USER`, `JURY`, `ADMIN`. Team membership — отдельное вычисляемое состояние `NO_TEAM`, `IN_TEAM`, `CAPTAIN`.

## 7. Frontend

Текущий frontend — server-rendered и встроен в Go-бинарник: `web/handler.go` встраивает `web/template` и `web/static` через `go:embed`. Templates парсятся при старте; ошибка останавливает приложение. Известные страницы обслуживаются точными маршрутами, неизвестные получают 404.

Tailwind CSS и JavaScript bundle собираются Node/esbuild stage. Asset hash добавляется к URL для cache busting. Nginx раздаёт `/static/*` с immutable cache; Go handler содержит те же assets для прямого запуска без Nginx. HTML и sensitive API responses имеют `no-store`.

Этот server-rendered frontend является legacy и запланирован к замене на независимое frontend-приложение (см. `../frontend/architecture.md` и решение `../decisions/0001-frontend-v2.md`). До завершения parity и cutover он остаётся текущей реализацией и поведенческим эталоном; его поведение не меняется в рамках подготовки миграции.

## 8. Docker deployment flow

Dockerfile содержит stages:

1. `frontend-build` — `npm ci`, Tailwind и esbuild;
2. `go-build` — static `spcase` и `healthcheck`;
3. `migrator-build`/`migrator` — pinned Goose и production migrations;
4. `nginx` — reverse proxy и static assets;
5. `runtime` — distroless nonroot application.

Compose startup:

```text
fresh db role initialization → db healthy → migrator completed → app ready → nginx
```

Containers используют nonroot users, dropped capabilities, `no-new-privileges` и read-only filesystems; writable temporary storage предоставляется только через tmpfs. PostgreSQL data хранится в named volume.

При создании чистого volume PostgreSQL init script оставляет `postgres` единственной administrative superuser-role, создаёт non-superuser `spcase_migrator` и `spcase_app`, передаёт migrator ownership database/schema и отзывает PUBLIC privileges. Production migrator применяет только `migrations/production.txt`; migration `00004` выдаёт runtime role точные grants, `00005` изолирует таблицу и sequence Goose metadata, development seed исключён.

Разделение вводится в два шага для безопасной работы с существующими volumes. Fresh volume использует initdb hook; существующая БД требует отдельного backup, restore rehearsal и ручного `scripts/cutover-postgres-roles.sh`. Cutover переносит ownership application objects и Goose metadata на `spcase_migrator`, задаёт runtime/default ACL и сохраняет legacy runtime role до smoke tests. Tracked Compose подключает migrator через `DB_MIGRATOR_*`, а Go application и `admin-bootstrap` через `DB_APP_*`; Compose отображает их в стандартные `DB_USER`/`DB_PASSWORD`, не передавая runtime container administrative или migrator secrets. Existing installation переходит на это wiring только после ручного database cutover.

Полная disposable-проверка перехода выполняется отдельным
`scripts/rehearse-existing-db-deployment.sh`: legacy source, verified backup,
независимый restored volume, idempotent cutover, migration 5, tracked
app/Nginx runtime, restart и отдельный rollback restore. Сценарий не является
частью Compose startup и не заменяет production backup approval.

Первый ADMIN создаётся отдельно через `cmd/admin-bootstrap`, пароль читается только из stdin. Bootstrap использует `spcase_app` и требует только runtime DML. Healthcheck обращается к readiness HTTP endpoint и не загружает database credentials самостоятельно.

## 9. Health and edge behavior

- `GET /api/v1/health/live` проверяет только живой Go-процесс.
- `GET /api/v1/health/ready` выполняет PostgreSQL ping с timeout 2 секунды; при ошибке возвращает стабильный `503 NOT_READY` без деталей БД и пишет событие `database_readiness_failed`.
- Application healthcheck вызывает readiness endpoint.
- Nginx формирует request ID, ограничивает login/registration, задаёт proxy timeouts, gzip и единый JSON для edge errors. Приложение принимает этот ID через middleware `RequestID` и включает его во все request-scoped логи, что даёт end-to-end корреляцию Nginx ↔ application logs.
- Все Compose services используют ограниченную локальную ротацию логов (`json-file`, 10 MiB × 5 файлов на service); это не заменяет централизованное хранение.
- Failure-mode rehearsal `scripts/rehearse-observability.sh` воспроизводимо проверяет startup, health semantics при потере PostgreSQL, видимость 5xx и restart, observability migration failure и graceful shutdown на disposable resources.

Внешний ingress обязан завершать TLS и безопасно доставлять трафик на loopback Nginx. Модель trusted client IP должна проверяться в deployment environment.
