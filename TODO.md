# spcase v1.0.0 — Production Readiness Roadmap

Текущий статус: production candidate. Функциональные backend, database и frontend-фазы завершены; дальнейшая работа направлена только на проверку и безопасный ввод в эксплуатацию.

## Реализованный baseline

- [x] Go HTTP-приложение со слоями handler → service → repository → PostgreSQL
- [x] Регистрация и аутентификация USER/JURY/ADMIN, RBAC и ревокация JWT через состояние аккаунта
- [x] Транзакционное управление командами и Hard Lock
- [x] Сдача решений, оценивание по шести критериям и управление lifecycle оценивания
- [x] Admin statistics, XLSX export и одноразовый bootstrap первого ADMIN
- [x] PostgreSQL migrations, production allowlist и development seed
- [x] Server-rendered frontend со встроенными шаблонами и assets
- [x] Multi-stage Docker images, Compose, Nginx и health checks
- [x] Unit, race, security и PostgreSQL integration test suites

## 1. Staging environment

- [ ] 1.1 Подготовить отдельные staging DNS, PostgreSQL volume и `.env.production` с уникальными секретами
- [ ] 1.2 Собрать все Compose targets и поднять полный stack через production migration flow
- [ ] 1.3 Выполнить одноразовый ADMIN bootstrap и проверить повторный отказ
- [ ] 1.4 Провести smoke-тест browser pages, static assets и всех API-контуров
- [ ] 1.5 Проверить restart, graceful shutdown, readiness gating и сохранность данных

## 2. Integration and acceptance testing

- [ ] 2.1 Запустить `go test -tags=integration ./internal/...` на изолированной staging PostgreSQL
- [ ] 2.2 Подтвердить concurrent join, детерминированный порядок блокировок и повторную проверку Hard Lock в БД
- [ ] 2.3 Проверить production migration allowlist на чистой БД и повторный idempotent запуск
- [ ] 2.4 Выполнить end-to-end сценарии USER, JURY и ADMIN через Nginx
- [ ] 2.5 Проверить frontend в поддерживаемых браузерах и на мобильных viewport

## 3. Deployment preparation

- [ ] 3.1 Зафиксировать immutable image tags/digests для staging и production
- [ ] 3.2 Настроить внешний TLS ingress для loopback-порта Nginx и проверить DNS
- [ ] 3.3 Проверить передачу client IP, trusted proxy headers и rate limits за реальным ingress
- [ ] 3.4 Описать deploy, rollback приложения и действия при неуспешной миграции
- [ ] 3.5 Проверить значения deadlines, CORS origins, APP_DOMAIN и Telegram URL перед запуском

## 4. Security review

- [ ] 4.1 Повторить `make security-check` на release commit
- [ ] 4.2 Проверить генерацию, хранение, ротацию и разграничение доступа к секретам
- [ ] 4.3 Проверить cookie, CORS, TLS, security headers и auth revocation в staging
- [x] 4.4 Разделить PostgreSQL privileges для app, migrator и administrative operations
- [ ] 4.5 Провести abuse-тест регистрации, login rate limits, payload limits и error disclosure

## 5. Backup and restore

- [ ] 5.1 Определить backup schedule, retention, encryption, RPO и RTO
- [ ] 5.2 Создать backup staging database
- [ ] 5.3 Восстановить backup в отдельную БД и проверить схему, данные и запуск приложения
- [ ] 5.4 Зафиксировать пошаговую restore-процедуру и ответственных

## 6. Observability

- [x] 6.1 JSON application logs в stdout/stderr
- [x] 6.2 Liveness/readiness endpoints и container health checks
- [x] 6.3 Nginx access/error logs и request ID
- [ ] 6.4 Настроить сбор и срок хранения application, Nginx и PostgreSQL logs
- [ ] 6.5 Добавить внешние uptime/readiness checks и alerting
- [ ] 6.6 Определить минимальные operational dashboards и thresholds до production deploy
