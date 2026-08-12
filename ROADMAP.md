# spcase — Roadmap

Текущий статус: production candidate; выполняется frontend-миграция.
Backend, database и production-инфраструктура реализованы и проверены.
Принято решение заменить server-rendered frontend на независимое современное
frontend-приложение до продолжения staging/production cutover
(см. `docs/decisions/0001-frontend-v2.md`).

Frontend-миграция находится на этапе публичной визуальной системы (фаза 4).
Технический foundation и поведенческий контракт завершены. Stage 4F (`4f6cbb2`)
зафиксировал правильный dark/content-first baseline. Stage 4G (`db9753a`)
добавил ограниченную spatial/pseudo-3D identity; human review 2026-08-11 признал
направление **DIRECTION ACCEPTED / COMPOSITION INCOMPLETE**: язык глубины,
материалов и restrained motion верный, но desktop-композиция слишком робкая —
hero-object мал и близок к layered-card stack, Format читается как три тёмные
панели, а часть пространства остаётся dead space вместо осмысленной negative
space. Следующий утверждённый шаг — **Stage 4H: spatial composition & density**:
не добавлять новый visual vocabulary, а усилить масштаб, отношения между
поверхностями и информационный footprint существующей системы. Governing rule:
**fill space with scale, relationships and depth — not with more components**.
Никаких новых runtime visual dependencies. См. `docs/frontend/design-direction.md`,
`docs/frontend/experience-model.md` и `docs/frontend/visual-acceptance.md`.

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
- [~] 1.2 Поддерживать документацию актуальной через human + ChatGPT Web documentation pass; coding agents по умолчанию меняют только реализацию

## Фаза 2. Frontend behavioral contract

- [x] 2.1 Провести аудит существующего server-rendered frontend (`web/`)
- [x] 2.2 Зафиксировать поведенческий контракт: маршруты, состояния, формы, ошибки, role-specific flows USER/JURY/ADMIN — `docs/frontend/legacy-contract.md`
- [x] 2.3 Определить критерии parity для будущего cutover — `docs/frontend/legacy-contract.md`, acceptance gates в `docs/frontend/cutover-plan.md`

## Фаза 3. Новый frontend foundation

- [x] 3.1 Создать независимое frontend-приложение по утверждённому стеку (`docs/frontend/architecture.md`) — реализовано в `frontend/` (Vite 8, React 19, TypeScript strict, React Router data mode, TanStack Query, CSS Modules); production cutover НЕ выполнен
- [x] 3.2 Настроить toolchain, проверки и интеграцию с существующим `/api/v1` — pnpm, Vite 8, TypeScript strict, Biome, dev-proxy к `localhost:8000`
- [x] 3.3 Настроить Playwright-окружение для end-to-end проверок — desktop/mobile projects и routing smoke tests; прогон требует установленных браузеров

## Фаза 4. Публичная визуальная система

- [~] 4.1 Реализовать публичные страницы по утверждённой design direction (`docs/frontend/design-direction.md`) — implementation продолжается; human visual acceptance ещё не получен
- [~] 4.2 Реализовать адаптивную композицию от 320 px, touch-only и reduced motion — технические механизмы существуют; финальная acceptance переносится в Stage 4H
- [x] 4.3 Stage 4C — wide/full-bleed composition, heterogeneous editorial scenes, уменьшение card chrome, mobile 320/375 recomposition — технически реализовано
- [x] 4.4 Stage 4D — Motion/view-transition interaction layer, bottom-nav marker, hero/schedule/auth choreography, PublicStatus и reduced-motion пути — технически реализовано; human review выявил избыточную agency-landing-page/motion-polish модель
- [x] 4.5 Stage 4E — art-direction consolidation (`73185c5`) — технически реализовано; human review 2026-08-11: **REJECT DIRECTION** из-за giant-number/editorial-rule/illustration/color-field/forced-asymmetry AI-template grammar
- [x] 4.6 Stage 4F — dark de-stylization (`4f6cbb2`): тёмный content-first canvas, один основной red accent, удалены dominant illustration/color-field/giant-number grammar, Mosaic и большая часть decorative motion; human review: **FOUNDATION ACCEPTED / VISUAL INCOMPLETE**
- [x] 4.7 Stage 4G — spatial identity without decorative illustration (`db9753a`): один Z2 hero artifact, Z1 Format depth, material planes и restrained pointer tilt без illustration system/new runtime dependency; human review: **DIRECTION ACCEPTED / COMPOSITION INCOMPLETE**
- [ ] 4.8 Stage 4H — spatial composition & density: сохранить Stage 4G language, увеличить visual footprint и wide-screen confidence; hero переосмыслить как один интегрированный spatial chassis/backplane вместо card-stack impression, Format связать в единую spatial progression, Schedule/FAQ использовать ширину и typography/data scale; не добавлять новые decorative components, illustration system, WebGL/Three.js или новую цветовую грамматику; обязательный desktop/mobile/reduced-motion human visual review

**Gate:** фаза 5 НЕ начинается, пока Stage 4H human review явно не зафиксирует
visual **ACCEPT** в `docs/frontend/visual-acceptance.md`. Технически валидный
commit не равен визуально принятому.

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
