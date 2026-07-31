# spcase v1.0.0 — PostgreSQL Schema

Каноническая схема находится в `migrations/00001_init_schema.sql`, индексы — в `00002_add_indexes.sql`, runtime grants — в `00004_grant_runtime_privileges.sql`. `00003_seed_dev_data.sql` содержит только development fixtures и отсутствует в production allowlist.

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

Unique `(jury_id, team_id, criterion_id)`. Constraint trigger требует роль JURY у автора.

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
- Leave/kick атомарно удаляют существующий submission, если состав стал меньше двух.
- Submission блокирует team, повторно проверяет captain, member count и database time, затем выполняет upsert.
- Evaluation batch берёт `FOR SHARE` на singleton state, требует существующий submission и атомарно upsert-ит ровно шесть criteria.
- Open/close evaluation берёт singleton `FOR UPDATE`, обновляет state и append-only event в одной transaction.
- First ADMIN bootstrap сериализован PostgreSQL advisory transaction lock.

Все repository transactions используют `READ COMMITTED`. Pool задаёт PostgreSQL `statement_timeout` и `lock_timeout` из configuration.

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
- sequences: `USAGE`, `SELECT`, `UPDATE`;
- `user_role` type: `USAGE`.

`spcase_app` не получает schema `CREATE`, ownership или доступ к migration metadata. PUBLIC database/schema privileges отозваны. Default privileges объектов, создаваемых `spcase_migrator`, обеспечивают runtime DML и sequence access; каждая schema migration обязана уточнять grants для объектов с более узкой моделью доступа.

## 6. Migrations

Production workflow копирует только файлы из `migrations/production.txt` во временный каталог, запрещает development seed, запускает Goose `up` и проверяет итоговую schema version.

Текущий production allowlist:

- `00001_init_schema.sql`;
- `00002_add_indexes.sql`;
- `00004_grant_runtime_privileges.sql`.

`00003_seed_dev_data.sql` применяется только локальными development commands и обратим собственной Down-секцией.

На чистом PostgreSQL entrypoint запускает `scripts/init-postgres-roles.sh`: создаёт non-superuser roles, передаёт database/schema ownership migrator и задаёт default privileges. Затем одноразовый migrator применяет production allowlist. На существующем volume роли и ownership должны быть подготовлены отдельной administrative cutover-процедурой до применения `00004`; изменение `POSTGRES_*` variables само по себе не меняет уже созданный cluster.

Пока сохраняется переходный режим: migrator уже подключается как `spcase_migrator`, но application и `admin-bootstrap` продолжают читать legacy `DB_USER`/`DB_PASSWORD`. Их переключение на `spcase_app` выполняется после подготовки существующих databases и не является частью этой migration.
