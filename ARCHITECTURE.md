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

## 4. Repositories and services

Repositories:

- `UserPostgres` — accounts, account projection, auth version, disabled state, first ADMIN.
- `TeamPostgres` — teams, membership и транзакционные mutations.
- `SubmissionPostgres` — получение и atomic upsert решения.
- `ScorePostgres` — atomic batch upsert и чтение evaluations/totals.
- `QueryPostgres` — jury workspace, evaluation state, admin stats и exports.

Services:

- `AuthService`, `UserService`, `TeamService`, `SubmissionService`, `ScoreService`, `JuryService`, `AdminService`, `ExportService`, `PublicService`.
- `AdminBootstrapService` используется отдельным CLI, а не HTTP API.

## 5. HTTP composition

`cmd/app` создаёт repositories, services и handlers вручную и регистрирует маршруты в `http.ServeMux`. Глобальная цепочка:

```text
SecurityHeaders
  → NoStoreSensitiveResponses
  → CORS
  → Recovery
  → APIErrorResponses
  → ServeMux
```

На защищённых маршрутах дополнительно применяются `AuthMiddleware` и `RequireRoles`. Для leave/kick/transfer/disband используется `HardLockMiddleware`; repository повторно проверяет lock по времени PostgreSQL после получения row locks.

HTTP server имеет read/write/header/idle timeouts, лимит заголовков и graceful shutdown до 15 секунд. Логи приложения — JSON через `slog` в stdout.

## 6. Authentication and authorization

JWT подписывается HS256 и живёт 24 часа. Claims: subject user UUID, role, `auth_version`, `iat`, `nbf`, `exp`. Team identity в token не хранится.

`access_token` устанавливается с `HttpOnly`, `Secure`, `SameSite=Lax`, `Path=/` и `Domain=APP_DOMAIN`. На каждом защищённом запросе middleware:

1. проверяет единственную cookie и JWT;
2. загружает актуальные role, `auth_version` и `disabled_at`;
3. отклоняет удалённый/disabled account, смену role и устаревшую auth version;
4. помещает immutable principal в request context.

RBAC-роли: `USER`, `JURY`, `ADMIN`. Team membership — отдельное вычисляемое состояние `NO_TEAM`, `IN_TEAM`, `CAPTAIN`.

## 7. Frontend

`web/handler.go` встраивает `web/template` и `web/static` через `go:embed`. Templates парсятся при старте; ошибка останавливает приложение. Известные страницы обслуживаются точными маршрутами, неизвестные получают 404.

Tailwind CSS и JavaScript bundle собираются Node/esbuild stage. Asset hash добавляется к URL для cache busting. Nginx раздаёт `/static/*` с immutable cache; Go handler содержит те же assets для прямого запуска без Nginx. HTML и sensitive API responses имеют `no-store`.

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

При создании чистого volume PostgreSQL init script оставляет `postgres` единственной administrative superuser-role, создаёт non-superuser `spcase_migrator` и `spcase_app`, передаёт migrator ownership database/schema и отзывает PUBLIC privileges. Production migrator применяет только `migrations/production.txt`; migration `00004` выдаёт runtime role точные grants, development seed исключён.

Разделение вводится в два шага для безопасной работы с существующими volumes. Migrator уже подключается через `DB_MIGRATOR_*`; Go application и `admin-bootstrap` пока используют совместимые legacy `DB_USER`/`DB_PASSWORD`. После administrative ownership cutover они переключаются на `DB_APP_*`. Поведение Go-приложения при этом не меняется.

Первый ADMIN создаётся отдельно через `cmd/admin-bootstrap`, пароль читается только из stdin. Bootstrap требует только runtime DML и после переключения использует `spcase_app`, а не administrative database credentials.

## 9. Health and edge behavior

- `GET /api/v1/health/live` проверяет только живой Go-процесс.
- `GET /api/v1/health/ready` выполняет PostgreSQL ping с timeout 2 секунды.
- Application healthcheck вызывает readiness endpoint.
- Nginx формирует request ID, ограничивает login/registration, задаёт proxy timeouts, gzip и единый JSON для edge errors.

Внешний ingress обязан завершать TLS и безопасно доставлять трафик на loopback Nginx. Модель trusted client IP должна проверяться в deployment environment.
