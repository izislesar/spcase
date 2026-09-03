# spcase

Платформа кейс-чемпионата (spcase.ru). Монолитное Go-приложение с REST API и
server-rendered frontend, PostgreSQL как хранилище состояния, Docker Compose
production-окружение с Nginx как единственной точкой входа.

Текущий статус: production candidate в переходном периоде. Server-rendered
frontend (`web/`) — действующая реализация и поведенческий эталон; выполняется
его замена на независимое React-приложение (`frontend/`) — migration target,
production cutover ещё НЕ выполнен
(`docs/decisions/0001-frontend-v2.md`, `ROADMAP.md`).

## Компоненты

- `cmd/app` — HTTP-приложение (handler → service → repository → PostgreSQL);
- `cmd/admin-bootstrap` — одноразовое создание первого ADMIN;
- `cmd/healthcheck` — readiness-проба для контейнера;
- `web/` — server-rendered frontend (templates, Tailwind/esbuild assets,
  встраиваются в бинарник) — текущий legacy frontend;
- `frontend/` — независимое React-приложение (Vite, TypeScript), цель
  frontend-миграции; в production-delivery пока не подключено;
- `migrations/` — Goose-миграции и production allowlist;
- `scripts/` — production migration, PostgreSQL role cutover и rehearsals;
- `Dockerfile`, `docker-compose*.yml`, `nginx.conf` — deployment.

## Разработка

Требования: Go, Node.js, Docker, доступный PostgreSQL.

Frontend assets собираются в Go-бинарник. Для воспроизводимой сборки:

```bash
make frontend-build
```

Во время разработки CSS и JavaScript пересобираются раздельно:
`npm run watch:css` и `npm run watch:js`.

Новый React-frontend (`frontend/`) разрабатывается независимо (pnpm, Vite;
dev-server проксирует `/api/v1` на Go backend `localhost:8000`):

```bash
cd frontend
pnpm install
pnpm dev          # Vite dev server на :5173
pnpm typecheck
pnpm check        # Biome lint + format
pnpm build
```

Миграции используют `DATABASE_URL` либо DB-переменные из локального `.env`:

```bash
make migrate-up        # локальная разработка
make migrate-status
make migrate-reset     # полный откат и повторное применение (только локально)
```

`migrations/00003_seed_dev_data.sql` — только локальные fixtures: 1 Admin,
3 Jury, 8 USER-аккаунтов, 3 команды и 2 submissions. Пароль всех
seed-аккаунтов — `password`. Production использует отдельный workflow
`make migrate-production` с allowlist `migrations/production.txt`; dev seed
туда не входит. Не используйте `migrate-up`, `migrate-down` или
`migrate-reset` в production.

## Трекинг задач (beads)

План работ ведётся в [beads](https://github.com/steveyegge/beads) (`bd`) —
distributed issue tracker для coding-агентов; конфиг и экспорт живут в
`.beads/`, источник истины — локальная Dolt-база (`.beads/embeddeddolt/`,
в git не коммитится). `.beads/issues.jsonl` — экспорт для просмотра и обмена,
не backup.

```bash
brew install beads        # или: npm install -g @beads/bd
bd init                   # первый запуск в клоне: создаст локальную базу
bd import                 # разовый импорт начального плана из .beads/issues.jsonl
bd ready                  # задачи без блокеров; bd prime — контекст для агента
bd dolt push              # публикует базу в refs/dolt/data этого же remote;
                          # после первого push свежие клоны поднимаются через bd bootstrap
```

## Проверки

```bash
gofmt -w .
go test -race ./...
go vet ./...
go build ./...
```

Полный набор обязательных Go- и npm-проверок (tests, `go vet`, `govulncheck`,
`npm audit`):

```bash
make security-check
```

PostgreSQL integration suite требует мигрированную disposable-БД версии 5 и
две роли (legacy `SPCASE_TEST_DATABASE_URL` намеренно не принимается):

```bash
SPCASE_TEST_MIGRATOR_DATABASE_URL='postgres://spcase_migrator:<password>@127.0.0.1:5432/spcase_test?sslmode=disable' \
SPCASE_TEST_APP_DATABASE_URL='postgres://spcase_app:<password>@127.0.0.1:5432/spcase_test?sslmode=disable' \
  go test -race -count=1 -tags=integration ./internal/...
```

Инфраструктурные изменения проверяются через
`docker compose --env-file .env.production config --quiet`.

## Local staging

Local staging использует production Compose с отдельным override, project
name, volume, image tags и loopback-портами. PostgreSQL публикуется только на
`127.0.0.1` для host-side integration tests.

```bash
cp .env.staging.example .env.staging
chmod 600 .env.staging
# заполнить administrative, migrator, application, JWT и jury secrets
# (openssl rand -hex 32 / openssl rand -base64 48); legacy DB_USER/DB_PASSWORD
# для fresh staging оставить пустыми

docker compose -p spcase-staging \
  -f docker-compose.yml -f docker-compose.staging.yml \
  --env-file .env.staging config --quiet

docker compose -p spcase-staging \
  -f docker-compose.yml -f docker-compose.staging.yml \
  --env-file .env.staging up --detach --build --wait
```

Проверки:

```bash
curl --fail http://127.0.0.1:18080/api/v1/health/live
curl --fail http://127.0.0.1:18080/api/v1/health/ready
```

Первый ADMIN создаётся существующим CLI внутри application image:

```bash
systemd-ask-password "Initial staging ADMIN password:" |
  docker compose -p spcase-staging \
    -f docker-compose.yml -f docker-compose.staging.yml \
    --env-file .env.staging run --rm --no-deps -T \
    --entrypoint /app/admin-bootstrap app \
    -full-name "Staging Administrator" \
    -email "admin@staging.spcase.ru"
```

Остановка без удаления данных: тот же Compose с `down` (без `--volumes`,
если staging data должны сохраниться).

## Production-контейнер и Compose

Образ собирает frontend assets отдельным Node.js stage, затем статические
Go-бинарники; runtime — `distroless/static` от `nonroot`, без shell и
toolchain. PostgreSQL в образ не входит.

```bash
docker build --pull -t spcase:local .
```

Контейнер ожидает уже мигрированную базу; миграции выполняются отдельно через
`make migrate-production`. Приложение принимает `SIGTERM` и завершает
HTTP-сервер с grace period до 15 секунд.

Production Compose поднимает PostgreSQL 16, одноразовый migrator, приложение
и Nginx в закрытой backend-сети. Единственная точка входа — Nginx на
`127.0.0.1:8080`; порты PostgreSQL и приложения на host не публикуются.
Порядок старта фиксирован: db healthy → migrator completed → app ready →
nginx. Migrator получает только `DB_MIGRATOR_*`, приложение и admin-bootstrap —
только `DB_APP_*`; administrative/migrator secrets в runtime-контейнере
отсутствуют.

```bash
cp .env.production.example .env.production
chmod 600 .env.production
# заполнить все секреты уникальными значениями

docker compose --env-file .env.production config --quiet
docker compose --env-file .env.production build --pull
docker compose --env-file .env.production up --detach --wait
```

Named volume `spcase_postgres_data` при `down` сохраняется; `--volumes`
удаляет базу без возможности восстановления.

Первый ADMIN в production создаётся отдельной одноразовой CLI-командой,
пароль принимается только через stdin:

```bash
systemd-ask-password "Initial ADMIN password:" |
  go run ./cmd/admin-bootstrap \
    -full-name "Production Administrator" \
    -email "admin@example.com"
```

Команда создаёт аккаунт только при отсутствии роли `ADMIN` и завершает
повторный запуск ошибкой; в stdout не выводятся имя, email или пароль.

## Документация

Авторитетные источники по областям:

- `docs/architecture/system.md` — архитектура системы, конфигурация,
  deployment topology;
- `docs/architecture/database.md` — схема, миграции, транзакции, роли и ACL
  PostgreSQL;
- `docs/domain/business-rules.md` — бизнес-правила и lifecycle;
- `docs/contracts/http-api.md` — HTTP API и browser routes;
- `docs/frontend/` — архитектура нового frontend, visual constitution
  (`design-direction.md`), product UX model (`experience-model.md`), current
  human visual verdict, legacy behavioral contract и cutover plan;
- `docs/decisions/0001-frontend-v2.md` — решение о замене frontend;
- `docs/runbooks/observability.md` — observability baseline и alert
  definitions;
- `docs/runbooks/postgres-cutover.md` — пошаговый operator runbook
  ownership/ACL cutover существующей БД (go/no-go, backup, rollback);
- `ROADMAP.md` — фазы и текущий этап;
- `AGENTS.md` — контекст и инварианты для coding-агентов.

Existing-volume PostgreSQL cutover, observability rehearsals и production
go/no-go выполняются строго по runbook-ам в `docs/runbooks/` и требуют
отдельных approvals.
