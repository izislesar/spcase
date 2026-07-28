# spcase

Backend кейс-чемпионата на Go и PostgreSQL.

## Frontend

HTML-шаблоны, скомпилированные Tailwind CSS и Alpine.js bundle встраиваются
в Go-бинарник. Для воспроизводимой сборки assets:

```bash
make frontend-build
```

Во время разработки CSS и JavaScript можно пересобирать раздельно командами
`npm run watch:css` и `npm run watch:js`.

## Миграции

Команды используют `DATABASE_URL` либо DB-переменные из локального `.env`:

```bash
make migrate-up
make migrate-down
make migrate-status
make migrate-reset
```

`migrate-reset` полностью откатывает схему и заново применяет все миграции.

`migrations/00003_seed_dev_data.sql` содержит только локальные fixtures:
1 Admin, 3 Jury, 8 USER-аккаунтов, 3 команды и 2 submissions. Пароль всех
seed-аккаунтов — `password`. В production следует применять только миграции
`00001` и `00002`.

Интеграционные тесты создают и удаляют изолированную схему в указанной БД:

```bash
SPCASE_TEST_DATABASE_URL='postgres://postgres:postgres@127.0.0.1:5432/spcase_test?sslmode=disable' \
  go test -tags=integration ./internal/repository
```
