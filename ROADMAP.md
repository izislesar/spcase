# spcase — Roadmap

Текущий статус: production candidate; выполняется frontend-миграция.
Backend, database и production-инфраструктура реализованы и проверены.
Принято решение заменить server-rendered frontend на независимое современное
frontend-приложение до продолжения staging/production cutover
(см. `docs/decisions/0001-frontend-v2.md`).

Frontend-миграция находится на этапе публичной визуальной системы (фаза 4):
технический foundation и поведенческий контракт завершены, публичная
визуальная реализация прошла итерации Stage 4 / 4A / 4B. Stage 4B технически
принят, но визуально НЕ принят — статус `ITERATE` (текущее состояние human
review: `docs/frontend/visual-acceptance.md`). Следующий шаг — Stage 4C.

Staging acceptance и production deployment НЕ являются ближайшей фазой: они
возобновляются только после завершения frontend-миграции.

## Реализованный baseline

- [x] Go HTTP-приложение со слоями handler → service → repository → PostgreSQL
- [x] Регистрация и аутентификация USER/JURY/ADMIN, RBAC и ревокация JWT через состояние аккаунта
- [x] Транзакционное управление командами и Hard Lock
- [x] Сдача решений, оценивание по шести критериям и управление lifecycle оценивания
- [x] Admin statistics, XLSX export и одноразовый bootstrap первого ADMIN
- [x] PostgreSQL migrations, production allowlist и development seed
- [x] Server-rendered frontend со встроенными шаблонами и assets (текущая реализация, legacy — запланирован к замене)
- [x] Multi-stage Docker images, Compose, Nginx и health checks
- [x] Unit, race, security и PostgreSQL integration test suites
- [x] Минимальный observability baseline и operator runbook production cutover

## Фаза 1. Context engineering

- [x] 1.1 Реструктурировать контекст репозитория: `AGENTS.md`, `docs/`, `ROADMAP.md`
- [ ] 1.2 Поддерживать документацию актуальной при изменении кода и архитектуры

## Фаза 2. Frontend behavioral contract

- [x] 2.1 Провести аудит существующего server-rendered frontend (`web/`)
- [x] 2.2 Зафиксировать поведенческий контракт: маршруты, состояния, формы, ошибки, role-specific flows USER/JURY/ADMIN — `docs/frontend/legacy-contract.md`
- [x] 2.3 Определить критерии parity для будущего cutover — `docs/frontend/legacy-contract.md`, acceptance gates в `docs/frontend/cutover-plan.md`

## Фаза 3. Новый frontend foundation

- [x] 3.1 Создать независимое frontend-приложение по утверждённому стеку (`docs/frontend/architecture.md`) — реализовано в `frontend/` (Vite 8, React 19, TypeScript strict, React Router data mode, TanStack Query, CSS Modules); production cutover НЕ выполнен
- [x] 3.2 Настроить toolchain, проверки и интеграцию с существующим `/api/v1` — pnpm, Vite 8, TypeScript strict, Biome, dev-proxy к `localhost:8000`
- [x] 3.3 Настроить Playwright-окружение для end-to-end проверок — desktop/mobile projects и routing smoke tests; прогон требует установленных браузеров

## Фаза 4. Публичная визуальная система

- [~] 4.1 Реализовать публичные страницы по утверждённой design direction (`docs/frontend/design-direction.md`) — реализация существует и прошла итерации Stage 4 / 4A / 4B; Stage 4B технически принят, human visual acceptance: **ITERATE** (не ACCEPT), см. `docs/frontend/visual-acceptance.md`
- [~] 4.2 Реализовать адаптивную композицию от 320 px, touch-only и reduced motion — механизмы реализованы; финальная композиция 320/375 px пересматривается в Stage 4C
- [ ] 4.3 Stage 4C — визуальная итерация по результатам human review (`docs/frontend/visual-acceptance.md`): wide/full-bleed композиция, heterogeneous editorial scenes, сокращение card chrome, mobile 320/375 recomposition

**Gate:** фаза 5 НЕ начинается, пока human review явно не зафиксирует visual
ACCEPT в `docs/frontend/visual-acceptance.md`. Технически валидный commit не
равен визуально принятому.

## Фаза 5. Миграция USER/JURY/ADMIN

- [ ] 5.1 Перенести USER flows (регистрация, команда, submission)
- [ ] 5.2 Перенести JURY flows (workspace, оценивание)
- [ ] 5.3 Перенести ADMIN flows (статистика, export, lifecycle оценивания)

## Фаза 6. Parity, mobile и accessibility

- [ ] 6.1 Проверить behavioral parity против зафиксированного контракта фазы 2
- [ ] 6.2 Проверить mobile composition, touch, hover-free и reduced-motion сценарии
- [ ] 6.3 Проверить accessibility baseline

## Фаза 7. Frontend cutover

- [ ] 7.1 Разделить доставку статического frontend и Go API в deployment
- [ ] 7.2 Выполнить тестируемый cutover с явным acceptance
- [ ] 7.3 Вывести legacy `web/` frontend из эксплуатации после принятия parity

## Фаза 8. Возобновление staging acceptance

Работы ниже сохранены из предыдущего production-readiness плана и становятся
актуальными после завершения frontend cutover.

### Staging environment

- [ ] 8.1 Подготовить отдельные staging DNS, PostgreSQL volume и `.env.production` с уникальными секретами
- [ ] 8.2 Собрать все Compose targets и поднять полный stack через production migration flow
- [ ] 8.3 Выполнить одноразовый ADMIN bootstrap и проверить повторный отказ
- [ ] 8.4 Провести smoke-тест browser pages, static assets и всех API-контуров
- [ ] 8.5 Проверить restart, graceful shutdown, readiness gating и сохранность данных

### Integration and acceptance testing

- [ ] 8.6 Запустить `go test -tags=integration ./internal/...` на изолированной staging PostgreSQL
- [ ] 8.7 Подтвердить concurrent join, детерминированный порядок блокировок и повторную проверку Hard Lock в БД
- [ ] 8.8 Проверить production migration allowlist на чистой БД и повторный idempotent запуск
- [ ] 8.9 Выполнить end-to-end сценарии USER, JURY и ADMIN через Nginx
- [x] 8.10 Устранить submission/evaluation lifecycle race: team-scoped lock order, historical evaluation contract, deterministic concurrency tests и structural integrity probes

### Deployment preparation

- [ ] 8.11 Зафиксировать immutable image tags/digests для staging и production
- [ ] 8.12 Настроить внешний TLS ingress для loopback-порта Nginx и проверить DNS
- [ ] 8.13 Проверить передачу client IP, trusted proxy headers и rate limits за реальным ingress
- [ ] 8.14 Описать deploy, rollback приложения и действия при неуспешной миграции
- [ ] 8.15 Проверить значения deadlines, CORS origins, APP_DOMAIN и Telegram URL перед запуском

### Security review

- [ ] 8.16 Повторить `make security-check` на release commit
- [ ] 8.17 Проверить генерацию, хранение, ротацию и разграничение доступа к секретам
- [ ] 8.18 Проверить cookie, CORS, TLS, security headers и auth revocation в staging
- [ ] 8.19 Завершить разделение PostgreSQL privileges для app, migrator и administrative operations
  - [x] Fresh-volume roles, migrator ownership, runtime grants и изоляция Goose metadata
  - [x] Разделить setup/runtime credentials в integration tests
  - [x] Реализовать и отрепетировать ownership/ACL cutover на disposable `pg_dump`/`pg_restore` копии
  - [x] Переключить fresh Compose application и admin-bootstrap на `DB_APP_*`
  - [x] Выполнить restored-existing-database deployment rehearsal с application runtime
  - [x] Подготовить reviewed operator runbook с go/no-go, observation и rollback gates (`docs/runbooks/postgres-cutover.md`)
  - [x] Провести final technical review cutover procedure
  - [ ] Утвердить production backup/rollback plan
  - [ ] Выполнить реальный ownership/ACL cutover после backup approval
  - [ ] Завершить утверждённый production observation period
  - [ ] Удалить transitional credentials после закрытия observation period
  - [ ] Вывести legacy database role из эксплуатации после отдельного approval
- [ ] 8.20 Провести abuse-тест регистрации, login rate limits, payload limits и error disclosure

### Backup and restore

- [ ] 8.21 Определить backup schedule, retention, encryption, RPO и RTO
- [ ] 8.22 Создать backup staging database
- [ ] 8.23 Восстановить backup в отдельную БД и проверить схему, данные и запуск приложения
- [ ] 8.24 Зафиксировать пошаговую restore-процедуру и ответственных

### Observability

- [x] 8.25 JSON application logs в stdout/stderr
- [x] 8.26 Liveness/readiness endpoints и container health checks
- [x] 8.27 Nginx access/error logs и request ID
- [ ] 8.28 Настроить сбор и срок хранения application, Nginx и PostgreSQL logs
  - [x] Ограниченная локальная ротация Docker logs для всех services (`json-file`, 10 MiB × 5)
  - [ ] Централизованный log destination и долгосрочное хранение
- [ ] 8.29 Добавить внешние uptime/readiness checks и alerting
  - [x] Alert definitions с placeholder thresholds задокументированы (`docs/runbooks/observability.md`)
  - [ ] Доставка alert-ов и утверждённые production thresholds
- [ ] 8.30 Определить минимальные operational dashboards и thresholds до production deploy
- [x] 8.31 Минимальный observability baseline: signal inventory, request correlation, стабильные event names, backup freshness checker (`scripts/check-backup-freshness.sh`), failure-mode rehearsal (`scripts/rehearse-observability.sh`) и runbook `docs/runbooks/observability.md`

## Фаза 9. Production deployment

- [ ] 9.1 Production deployment — только после явного approval и завершения фаз 7–8
