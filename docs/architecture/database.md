# spcase v1.0.0 — PostgreSQL Schema

Каноническая схема находится в `migrations/00001_init_schema.sql`, индексы — в `00002_add_indexes.sql`, runtime grants — в `00004_grant_runtime_privileges.sql`, изоляция Goose metadata — в `00005_isolate_goose_metadata.sql`. `00003_seed_dev_data.sql` содержит только development fixtures и отсутствует в production allowlist.

## 1. Types and tables

### `user_role`

PostgreSQL enum: `USER`, `JURY`, `ADMIN`.

### `users`

Единая таблица identities для участников, жюри и администраторов:

| Column | Type | Rules |
|---|---|---|
| `id` | UUID | PK, `gen_random_uuid()` |
| `full_name` | VARCHAR(255) | NOT NULL |
| `university` | VARCHAR(255) | required by CHECK only for USER |
| `email` | VARCHAR(255) | NOT NULL, case-insensitive unique index |
| `telegram` | VARCHAR(100) | required by CHECK only for USER |
| `password_hash` | VARCHAR(255) | NOT NULL, bcrypt value |
| `role` | `user_role` | NOT NULL |
| `auth_version` | INTEGER | NOT NULL, default 1, greater than zero |
| `disabled_at` | TIMESTAMPTZ | NULL means enabled |
| `created_at` | TIMESTAMPTZ | NOT NULL |

`auth_version` изменяется при password/account/session invalidation operations. Protected requests сравнивают его и role с JWT. `disabled_at` немедленно запрещает login и использование существующих tokens.

### `teams`

| Column | Type | Rules |
|---|---|---|
| `id` | UUID | PK |
| `name` | VARCHAR(100) | NOT NULL, case-insensitive unique index |
| `invite_code` | VARCHAR(8) | NOT NULL, unique, `^[A-Z0-9]{8}$` |
| `captain_id` | UUID | NOT NULL, references `users`, delete restricted |
| `created_at`, `updated_at` | TIMESTAMPTZ | NOT NULL |

Captain обязан одновременно существовать в `team_members`. Это обеспечивается deferred composite FK `(team.id, captain_id) → team_members(team_id, user_id)`.

### `team_members`

| Column | Type | Rules |
|---|---|---|
| `team_id` | UUID | FK to teams, ON DELETE CASCADE |
| `user_id` | UUID | FK to users, delete restricted |
| `joined_at` | TIMESTAMPTZ | NOT NULL |

PK — `(team_id, user_id)`; unique `user_id` запрещает участие в нескольких командах. Constraint trigger допускает только accounts с ролью USER.

Максимум четыре участника обеспечивается транзакционным repository check под team row lock. Минимум два участника применяется к submission, а не к существованию команды.

### `submissions`

Одна актуальная работа на команду:

| Column | Type | Rules |
|---|---|---|
| `id` | UUID | PK |
| `team_id` | UUID | UNIQUE FK, ON DELETE CASCADE |
| `solution_url` | TEXT | NOT NULL, HTTP(S)-like CHECK |
| `updated_at` | TIMESTAMPTZ | NOT NULL |

Application дополнительно валидирует URL и ограничивает его 2048 bytes.

### `evaluations`

Одна строка на критерий:

| Column | Type | Rules |
|---|---|---|
| `id` | UUID | PK |
| `jury_id` | UUID | FK to users, delete restricted |
| `team_id` | UUID | FK to teams, ON DELETE CASCADE |
| `criterion_id` | SMALLINT | 1..6 |
| `score` | SMALLINT | 0..10 |
| `updated_at` | TIMESTAMPTZ | NOT NULL |

Unique `(jury_id, team_id, criterion_id)`. Constraint trigger требует роль JURY у автора. FK на `teams`, а не на `submissions`, отражает историческую семантику: потеря текущего submission не удаляет оценки, а удаление team удаляет их каскадно.

### `evaluation_state`

Singleton row с `singleton_id=1`:

- `is_closed`;
- `closed_at` и `closed_by`, обязательные только в закрытом состоянии;
- `updated_at`.

`closed_by` ссылается на ADMIN. Удаление и truncate singleton запрещены triggers.

### `evaluation_state_events`

Append-only audit:

- UUID `id`;
- `action IN ('OPEN','CLOSE')`;
- `admin_id`, обязанный иметь роль ADMIN;
- `created_at`.

Update, delete и truncate запрещены triggers. Событие создаётся только при фактическом изменении state.

## 2. Role integrity

Deferred constraint triggers проверяют роли при membership, evaluation и evaluation-state writes. Отдельный trigger запрещает менять роль пользователя, пока с ним связаны:

- team membership для USER;
- evaluations для JURY;
- evaluation state/events для ADMIN.

## 3. Indexes

Кроме PK/UNIQUE indexes миграция создаёт:

- `uq_users_email_ci ON users (LOWER(email))`;
- `uq_teams_name_ci ON teams (LOWER(name))`;
- `idx_evaluations_team_id ON evaluations(team_id)`;
- `idx_teams_captain_id ON teams(captain_id)`.

Integration tests выполняют `EXPLAIN` критических queries.

## 4. Transactional operations

- Team create блокирует captain account, проверяет role/membership и атомарно создаёт team + captain membership.
- Join блокирует team по invite code и user; capacity проверяется внутри той же transaction.
- Leave/kick/transfer/disband блокируют team `FOR UPDATE`, затем user rows в стабильном UUID-порядке. Hard Lock повторно вычисляется через `clock_timestamp()` после locks.
- Leave/kick атомарно удаляют существующий submission, если состав стал меньше двух; исторические evaluations сохраняются.
- Submission блокирует team, повторно проверяет captain, member count и database time, затем выполняет upsert.
- Evaluation batch блокирует team `FOR SHARE`, затем singleton state `FOR SHARE` и текущий submission `FOR SHARE`; после проверок атомарно upsert-ит ровно шесть criteria в порядке criterion ID.
- Open/close evaluation берёт singleton `FOR UPDATE`, обновляет state и append-only event в одной transaction.
- First ADMIN bootstrap сериализован PostgreSQL advisory transaction lock.

Все repository transactions используют `READ COMMITTED`. Pool задаёт PostgreSQL `statement_timeout` и `lock_timeout` из configuration.

Lifecycle lock order начинается с team row. Membership mutations продолжают к user rows в стабильном UUID-порядке; scoring продолжает к evaluation state и submission. Поэтому scoring сериализован только с mutation той же команды, open/close сериализован со всеми score writes через singleton state, а операции разных команд не блокируют друг друга на team locks.

## 5. Database roles

- `postgres` — cluster bootstrap, создание ролей и аварийные administrative operations. Это единственная superuser-role; приложение и migrator не должны использовать её после завершения перехода.
- `spcase_migrator` — LOGIN без `SUPERUSER`, `CREATEDB`, `CREATEROLE`, `REPLICATION` и `BYPASSRLS`; владеет application database и schema `public`, устанавливает extension и выполняет DDL через Goose.
- `spcase_app` — LOGIN без administrative/DDL privileges; имеет `CONNECT`, schema `USAGE` и только требуемый runtime DML.

Текущие runtime grants:

- `users`: `SELECT`, `INSERT`, `UPDATE`;
- `teams`: `SELECT`, `INSERT`, `UPDATE`, `DELETE`;
- `team_members`: `SELECT`, `INSERT`, `DELETE`;
- `submissions`: `SELECT`, `INSERT`, `UPDATE`, `DELETE`;
- `evaluations`: `SELECT`, `INSERT`, `UPDATE`;
- `evaluation_state`: `SELECT`, `UPDATE`;
- `evaluation_state_events`: `INSERT`;
- application sequences: отсутствуют; UUID defaults используют `gen_random_uuid()` и не требуют sequence privileges;
- `user_role` type: `USAGE`.

`spcase_app` не получает schema `CREATE`, ownership или доступ к таблице/sequence Goose metadata. PUBLIC database/schema privileges отозваны. Default privileges объектов, создаваемых `spcase_migrator`, обеспечивают runtime DML и sequence access; каждая schema migration обязана уточнять grants для объектов с более узкой моделью доступа.

## 6. Migrations

Production workflow копирует только файлы из `migrations/production.txt` во временный каталог, запрещает development seed, запускает Goose `up` и проверяет итоговую schema version.

Текущий production allowlist:

- `00001_init_schema.sql`;
- `00002_add_indexes.sql`;
- `00004_grant_runtime_privileges.sql`;
- `00005_isolate_goose_metadata.sql`.

`00003_seed_dev_data.sql` применяется только локальными development commands и обратим собственной Down-секцией.

PostgreSQL integration harness требует отдельные
`SPCASE_TEST_MIGRATOR_DATABASE_URL` и `SPCASE_TEST_APP_DATABASE_URL`. Migrator
создаёт, мигрирует и удаляет случайно именованную isolated schema; runtime pool
подключается к той же schema как `spcase_app` и используется всеми repositories,
transactions и concurrency tests. Legacy single-connection variable не используется.

Исходная база должна уже быть на production migration version 5. Harness
проверяет обе connection roles; runtime grants для isolated objects копируются из
уже мигрированной `public` schema, поэтому тесты не могут незаметно расширить
privileges приложения. Suite покрывает schema integrity, migrations/seed,
concurrent joins, database deadline checks, historical evaluation retention,
deterministic scoring против leave/kick/disband и submission races,
evaluation-state serialization, per-team lock independence, query plans,
aggregation, timeouts и concurrent ADMIN bootstrap.

Fresh-database role ACL check:

```bash
SPCASE_TEST_MIGRATOR_DATABASE_URL='postgres://spcase_migrator:...' \
SPCASE_TEST_APP_DATABASE_URL='postgres://spcase_app:...' \
go test -tags=integration ./internal/repository -run '^TestPostgresRolePrivileges$'
```

`SPCASE_TEST_DATABASE_URL` отклоняется при одиночной передаче, потому что один
DDL-capable connection схлопывает production privilege boundary.

На чистом PostgreSQL entrypoint запускает `scripts/init-postgres-roles.sh`: создаёт non-superuser roles, передаёт database/schema ownership migrator и задаёт default privileges. Затем одноразовый migrator применяет production allowlist.

Существующий volume переводится только вручную через
`scripts/cutover-postgres-roles.sh`, после verified backup. Требуются explicit
database, superuser connection, фактический legacy owner и confirmation guard.
Preflight допускает ownership только у legacy role или `spcase_migrator`
(`pg_database_owner` также допустим для legacy `public` schema), отклоняет
неожиданные user schemas/owners, участие target roles в role memberships и
принимает только production Goose versions 2, 4 или 5; applied development seed
version 3 отклоняется. Transactional stage переносит non-extension tables, sequences, views, materialized views,
foreign tables, routines, enums/domains/types и связанные indexes; database
ownership меняется отдельной командой. Constraints, triggers, defaults, data и
Goose history не переписываются.

После cutover database, `public`, application objects и Goose metadata принадлежат
`spcase_migrator`; `spcase_app` получает только документированную DML matrix,
`user_role` usage и никаких DDL/Goose privileges. Default privileges совпадают с
fresh-volume model. Для custom legacy owner сохраняются transitional runtime DML
grants, но ownership/DDL отзываются; role не отключается и не удаляется.
Повторный запуск сначала проверяет already-converted state и остаётся
семантически idempotent. `scripts/rehearse-postgres-role-cutover.sh` проверяет это
на отдельной `pg_dump`/`pg_restore` копии и очищает созданные disposable resources.
`scripts/rehearse-existing-db-deployment.sh` дополнительно объединяет verified
custom-format backup, независимые target/rollback restores, cutover, migration 5,
tracked Compose runtime под `spcase_app`, API smoke и restart/persistence в одной
явно подтверждаемой disposable-процедуре. Исходный legacy fingerprint должен
совпасть после restore и rollback; converted fingerprint может отличаться только
версией миграции и намеренно созданными smoke-test records/events.

Tracked Compose подключает migrator как `spcase_migrator`, а application и
`admin-bootstrap` как `spcase_app`; role-specific secrets отображаются в
стандартные `DB_USER`/`DB_PASSWORD` только внутри соответствующего container.
Legacy variables и role пока сохраняются для явного rollback existing
installations. Реальное existing-volume deployment требует предварительного
backup approval и ручного cutover; автоматического credential fallback нет.
