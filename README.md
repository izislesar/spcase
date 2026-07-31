# spcase

Backend кейс-чемпионата на Go и PostgreSQL.

## Frontend

HTML-шаблоны, скомпилированные Tailwind CSS, Alpine.js и motion-слой на
Lenis/GSAP встраиваются в Go-бинарник. Для воспроизводимой сборки assets:

```bash
make frontend-build
```

Во время разработки CSS и JavaScript можно пересобирать раздельно командами
`npm run watch:css` и `npm run watch:js`.

## Проверки безопасности

Полный набор обязательных Go- и npm-проверок запускается одной командой:

```bash
make security-check
```

Команда завершается с ошибкой при падении тестов, `go vet`, `govulncheck` или
`npm audit`. Закреплённая версия `govulncheck` загружается Go toolchain
автоматически и не требует глобальной установки.

## Миграции

Команды используют `DATABASE_URL` либо DB-переменные из локального `.env`:

```bash
make migrate-up
make migrate-production
make migrate-down
make migrate-status
make migrate-reset
```

`migrate-up`, `migrate-down` и `migrate-reset` предназначены для локальной
разработки. `migrate-reset` полностью откатывает схему и заново применяет все
миграции.

`migrations/00003_seed_dev_data.sql` содержит только локальные fixtures:
1 Admin, 3 Jury, 8 USER-аккаунтов, 3 команды и 2 submissions. Пароль всех
seed-аккаунтов — `password`.

В production разрешено запускать только отдельный workflow:

```bash
DATABASE_URL='postgres://...' make migrate-production
```

Он применяет allowlist из `migrations/production.txt` и не видит dev seed
`00003`. Новую production schema-миграцию необходимо явно добавить в этот
allowlist после review. Не используйте `migrate-up`, `migrate-down` или
`migrate-reset` в production.

После production-миграций первый ADMIN создаётся отдельной одноразовой
CLI-командой. Пароль принимается только через stdin и не передаётся аргументом
или переменной окружения:

```bash
systemd-ask-password "Initial ADMIN password:" |
  go run ./cmd/admin-bootstrap \
    -full-name "Production Administrator" \
    -email "admin@example.com"
```

Для собранного бинарника используется тот же интерфейс:

```bash
systemd-ask-password "Initial ADMIN password:" |
  ./admin-bootstrap -full-name "Production Administrator" -email "admin@example.com"
```

Команда использует конфигурацию приложения из ENV/`.env`, создаёт аккаунт только
при отсутствии роли `ADMIN` и завершает повторный запуск ошибкой. В stdout не
выводятся имя, email или пароль.

Интеграционные тесты создают и удаляют изолированную схему в указанной БД:

```bash
SPCASE_TEST_DATABASE_URL='postgres://postgres:postgres@127.0.0.1:5432/spcase_test?sslmode=disable' \
  go test -tags=integration ./internal/...
```

## Local staging

Local staging использует production Compose с отдельным override, project name,
volume, image tags и loopback-портами. PostgreSQL публикуется только на
`127.0.0.1` для host-side integration tests.

Подготовка:

```bash
cp .env.staging.example .env.staging
chmod 600 .env.staging
openssl rand -hex 32
openssl rand -base64 48
openssl rand -base64 48
```

Заполните `DB_PASSWORD`, `JWT_SECRET` и `JURY_REGISTRATION_KEY`, затем:

```bash
docker compose \
  -p spcase-staging \
  -f docker-compose.yml \
  -f docker-compose.staging.yml \
  --env-file .env.staging \
  config --quiet

docker compose \
  -p spcase-staging \
  -f docker-compose.yml \
  -f docker-compose.staging.yml \
  --env-file .env.staging \
  up --detach --build --wait
```

Проверки:

```bash
curl --fail http://127.0.0.1:18080/api/v1/health/live
curl --fail http://127.0.0.1:18080/api/v1/health/ready

SPCASE_TEST_DATABASE_URL='postgres://spcase_staging:<password>@127.0.0.1:15432/spcase_staging?sslmode=disable' \
  go test -tags=integration ./internal/...
```

Первый ADMIN создаётся существующим CLI внутри application image:

```bash
systemd-ask-password "Initial staging ADMIN password:" |
  docker compose \
    -p spcase-staging \
    -f docker-compose.yml \
    -f docker-compose.staging.yml \
    --env-file .env.staging \
    run --rm --no-deps -T \
    --entrypoint /app/admin-bootstrap app \
    -full-name "Staging Administrator" \
    -email "admin@staging.spcase.ru"
```

Остановка без удаления данных:

```bash
docker compose \
  -p spcase-staging \
  -f docker-compose.yml \
  -f docker-compose.staging.yml \
  --env-file .env.staging \
  down
```

Не используйте `--volumes`, если staging data должны сохраниться.

## Production-контейнер

Образ собирает frontend assets отдельным Node.js stage, затем статические
Go-бинарники приложения и readiness-пробы. Runtime основан на
`distroless/static`, запускается от пользователя `nonroot` и не содержит shell,
исходный код или build toolchain. PostgreSQL в образ не входит и остается
внешней зависимостью.

Сборка:

```bash
docker build --pull -t spcase:local .
```

Для локальной проверки создайте отдельный env-файл на основе `.env.example`.
Задайте независимые случайные `JWT_SECRET` и `JURY_REGISTRATION_KEY`, а
`DB_HOST` укажите как адрес PostgreSQL, доступный из контейнера. При PostgreSQL
на хосте Docker это `host.docker.internal`:

```bash
cp .env.example .env.container
docker run --rm --name spcase \
  --detach \
  --env-file .env.container \
  --add-host=host.docker.internal:host-gateway \
  --read-only \
  --cap-drop=ALL \
  --security-opt=no-new-privileges:true \
  -p 127.0.0.1:8000:8000 \
  spcase:local
```

Контейнер ожидает уже мигрированную production-базу. Миграции выполняются
отдельно через `make migrate-production`; образ приложения не применяет их при
старте.

Проверка readiness и frontend:

```bash
docker inspect --format '{{.State.Health.Status}}' spcase
curl --fail http://127.0.0.1:8000/api/v1/health/ready
curl --fail http://127.0.0.1:8000/
curl --fail http://127.0.0.1:8000/static/css/app.css
```

Основной процесс принимает `SIGTERM` и завершает HTTP-сервер с grace period до
15 секунд. Для ручной остановки дайте контейнеру больший timeout:

```bash
docker stop --timeout=20 spcase
```

## Production Compose

Compose запускает PostgreSQL 16, одноразовый production migrator и приложение в
закрытой backend-сети. Nginx подключён к backend и отдельной edge-сети. Порты
PostgreSQL и приложения на host не публикуются. Единственная точка входа — Nginx
на `127.0.0.1:8080`; такой bind не открывает сервис во внешнюю сеть и
предназначен для локального reverse proxy/tunnel.

Подготовьте production-конфигурацию и заполните все пустые секреты:

```bash
cp .env.production.example .env.production
chmod 600 .env.production
openssl rand -hex 32
openssl rand -base64 48
openssl rand -base64 48
```

Каждое сгенерированное значение используйте один раз. Проверить конфигурацию,
собрать образы и поднять stack:

```bash
docker compose --env-file .env.production config --quiet
docker compose --env-file .env.production build --pull
docker compose --env-file .env.production up --detach --wait
```

Порядок старта фиксирован: PostgreSQL должен пройти `pg_isready`, migrator
должен успешно завершить `make migrate-production`, затем запускается приложение
и только после его readiness — Nginx. Migrator использует
`migrations/production.txt`, поэтому dev seed `00003` не применяется.

Nginx раздаёт собранные `/static/*` напрямую с immutable cache, а
server-rendered страницы и `/api/v1/*` проксирует в `app:8000`. Для login и
registration endpoints включены отдельные per-IP rate limits. Входные
`X-Forwarded-*` и `X-Request-ID` не передаются как есть: Nginx формирует
доверенные значения для приложения и добавляет request ID в ответ.

Проверить состояние:

```bash
docker compose --env-file .env.production ps
docker compose --env-file .env.production logs migrator
docker compose --env-file .env.production exec app /app/healthcheck
curl --fail http://127.0.0.1:8080/
curl --fail http://127.0.0.1:8080/static/css/app.css
curl --fail http://127.0.0.1:8080/api/v1/health/ready
```

Прямой доступ к `app:8000` и PostgreSQL с host отсутствует. Менять bind Nginx с
`127.0.0.1` на `0.0.0.0` не следует: внешний TLS ingress подключается к
локальному порту на следующем deployment-этапе.

Повторный запуск production migrations безопасен. Обычный перезапуск с
сохранением данных:

```bash
docker compose --env-file .env.production down
docker compose --env-file .env.production up --detach --wait
```

Named volume `spcase_postgres_data` при `down` сохраняется. Флаг `--volumes`
удаляет базу без возможности восстановления и для обычного deploy/restart
использоваться не должен.
