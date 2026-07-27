# api.md — Спецификация REST API (OpenAPI / Swagger Format)

---

## 1. Общие Стандарты API

### 1.1. Базовые параметры

* **Base URL:** `/api/v1`
* **Формат данных:** `application/json` (все тела запросов и ответов).
* **Кодировка:** `UTF-8`

### 1.2. Аутентификация и Сессия

Все защищенные эндпоинты требуют наличие валидного JWT-токена в `HttpOnly` Cookie с именем `access_token`.

* При отсутствии куки или истечении срока действия токена бэкенд возвращает `401 Unauthorized`.
* При попытке доступа к ресурсу, не соответствующему роли пользователя (`USER`, `JURY`, `ADMIN`), бэкенд возвращает `403 Forbidden`.

---

## 2. Спецификация Эндпоинтов

### 2.1. Публичный контур (Public Flow)

#### `GET /api/v1/health`

Проверка работоспособности сервиса (Liveness/Readiness probe).

* **Auth:** Не требуется
* **Response 200 OK:**

```json
{
  "status": "ok",
  "timestamp": "2026-07-27T12:00:00Z"
}

```

#### `GET /api/v1/info`

Конфигурация чемпионата и таймер обратного отсчета.

* **Auth:** Не требуется
* **Response 200 OK:**

```json
{
  "registration_deadline": "2026-10-15T18:00:00Z",
  "submission_deadline": "2026-10-17T21:00:00Z",
  "is_registration_open": true,
  "is_submission_open": false
}

```

#### `GET /api/v1/schedule`

Расписание и таймлайн проведения мероприятия.

* **Auth:** Не требуется
* **Response 200 OK:**

```json
{
  "events": [
    {
      "id": 1,
      "title": "Старт регистрации",
      "start_time": "2026-09-01T10:00:00Z",
      "description": "Открытие формы регистрации команд"
    },
    {
      "id": 2,
      "title": "Очное открытие и выдача ТЗ",
      "start_time": "2026-10-16T10:00:00Z",
      "description": "Презентация кейса организаторами на площадке"
    }
  ]
}

```

#### `GET /api/v1/faq`

Список вопросов и ответов для аккордеона.

* **Auth:** Не требуется
* **Response 200 OK:**

```json
{
  "faq": [
    {
      "id": 1,
      "question": "Сколько человек может быть в команде?",
      "answer": "В составе команды должно быть от 2 до 4 человек."
    }
  ]
}

```

#### `GET /api/v1/no-team`

Информация и инструкция для участников без команды.

* **Auth:** Не требуется
* **Response 200 OK:**

```json
{
  "message": "Если у вас нет команды, перейдите в закрытый Telegram-чат для поиска сокомандников.",
  "telegram_chat_url": "https://t.me/joinchat/example_hash"
}

```

---

### 2.2. Контур Авторизации и Профиля (`/auth`, `/user`)

#### `POST /api/v1/auth/register`

Регистрация нового участника.

* **Auth:** Не требуется
* **Request Body:**

```json
{
  "full_name": "Иванов Иван Иванович",
  "university": "РУДН",
  "email": "ivanov@example.com",
  "telegram": "@ivanov",
  "password": "strongpassword123"
}

```

* **Response 201 Created:** (Автоматически выставляет `Set-Cookie: access_token=...`)

```json
{
  "id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
  "email": "ivanov@example.com",
  "role": "USER"
}

```

* **Response 409 Conflict:** `{"error": "EMAIL_ALREADY_EXISTS"}`

#### `POST /api/v1/auth/login`

Вход участника или администратора.

* **Auth:** Не требуется
* **Request Body:**

```json
{
  "email": "ivanov@example.com",
  "password": "strongpassword123"
}

```

* **Response 200 OK:** (Записывает `access_token` в `HttpOnly` Cookie)

```json
{
  "status": "success",
  "role": "USER"
}

```

* **Response 401 Unauthorized:** `{"error": "INVALID_CREDENTIALS"}`

#### `POST /api/v1/auth/logout`

Выход из системы (очистка Cookie).

* **Auth:** Требуется (`USER`, `JURY`, `ADMIN`)
* **Response 200 OK:** (Выставляет просроченную куку `access_token=; Max-Age=0`)

```json
{
  "status": "logged_out"
}

```

#### `GET /api/v1/user/me`

Получение профиля текущего пользователя и статуса его команды.

* **Auth:** Требуется (`USER`, `ADMIN`)
* **Response 200 OK:**

```json
{
  "id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
  "full_name": "Иванов Иван Иванович",
  "university": "РУДН",
  "email": "ivanov@example.com",
  "telegram": "@ivanov",
  "role": "USER",
  "team_status": "CAPTAIN", // Варианты: "NO_TEAM", "IN_TEAM", "CAPTAIN"
  "team_id": "c39a2828-82ef-4171-a75d-354a93edb8f2"
}

```

---

### 2.3. Управление командами (`/team`) — *Role: USER*

#### `POST /api/v1/team/create`

Создание новой команды (текущий пользователь становится капитаном).

* **Auth:** Требуется (`USER`)
* **Request Body:**

```json
{
  "name": "CyberPunks"
}

```

* **Response 201 Created:**

```json
{
  "id": "c39a2828-82ef-4171-a75d-354a93edb8f2",
  "name": "CyberPunks",
  "invite_code": "X8K2M9N7",
  "captain_id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
}

```

* **Response 409 Conflict:** `{"error": "ALREADY_IN_TEAM"}`

#### `POST /api/v1/team/join`

Вступление в команду по 8-значному инвайт-коду.

* **Auth:** Требуется (`USER`)
* **Request Body:**

```json
{
  "invite_code": "X8K2M9N7"
}

```

* **Response 200 OK:**

```json
{
  "message": "Successfully joined team",
  "team_id": "c39a2828-82ef-4171-a75d-354a93edb8f2",
  "team_name": "CyberPunks"
}

```

* **Response 400 Bad Request:** `{"error": "TEAM_FULL"}` (Если в команде уже 4 человека)
* **Response 404 Not Found:** `{"error": "INVALID_INVITE_CODE"}`

#### `GET /api/v1/team/my`

Получение полной информации о своей команде, её составе и сданном решении.

* **Auth:** Требуется (`USER`)
* **Response 200 OK:**

```json
{
  "id": "c39a2828-82ef-4171-a75d-354a93edb8f2",
  "name": "CyberPunks",
  "invite_code": "X8K2M9N7",
  "captain_id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
  "status_badge": "READY", // Варианты: "SEARCHING", "READY", "SUBMITTED"
  "members": [
    {
      "id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
      "full_name": "Иванов Иван Иванович",
      "telegram": "@ivanov",
      "is_captain": true
    },
    {
      "id": "b11fbc99-1c0b-4ef8-bb6d-7bb9bd380a22",
      "full_name": "Петров Петр Петрович",
      "telegram": "@petrov",
      "is_captain": false
    }
  ],
  "submission": {
    "solution_url": "https://disk.yandex.ru/d/example",
    "updated_at": "2026-10-17T19:30:00Z"
  }
}

```

#### `POST /api/v1/team/leave`

Выход из текущей команды (доступно только рядовому участнику, но не капитану).

* **Auth:** Требуется (`USER`)
* **Response 200 OK:** `{"message": "Successfully left team"}`
* **Response 400 Bad Request:** `{"error": "CAPTAIN_CANNOT_LEAVE"}` (Капитан должен передать права или расформировать команду)
* **Response 403 Forbidden:** `{"error": "MUTATIONS_LOCKED"}` (За 1 час до дедлайна)

#### `POST /api/v1/team/kick`

Исключение участника из команды (только для Капитана).

* **Auth:** Требуется (`USER` с правами капитана)
* **Request Body:**

```json
{
  "user_id": "b11fbc99-1c0b-4ef8-bb6d-7bb9bd380a22"
}

```

* **Response 200 OK:** `{"message": "User kicked successfully"}`
* **Response 403 Forbidden:** `{"error": "MUTATIONS_LOCKED"}`

#### `POST /api/v1/team/transfer-ownership`

Передача прав капитана другому участнику команды.

* **Auth:** Требуется (`USER` с правами капитана)
* **Request Body:**

```json
{
  "new_captain_id": "b11fbc99-1c0b-4ef8-bb6d-7bb9bd380a22"
}

```

* **Response 200 OK:** `{"message": "Ownership transferred successfully"}`

#### `DELETE /api/v1/team/disband`

Полное расформирование команды (только для Капитана).

* **Auth:** Требуется (`USER` с правами капитана)
* **Response 200 OK:** `{"message": "Team disbanded successfully"}`

---

### 2.4. Сдача решений (`/team/submit`) — *Role: CAPTAIN*

#### `POST /api/v1/team/submit`

Сохранение или изменение URL-ссылки на решение (Яндекс.Диск / Google Drive).

* **Auth:** Требуется (`USER` с правами капитана, состав команды $\ge 2$)
* **Request Body:**

```json
{
  "solution_url": "https://disk.yandex.ru/d/example_folder"
}

```

* **Response 200 OK:**

```json
{
  "status": "submitted",
  "solution_url": "https://disk.yandex.ru/d/example_folder",
  "updated_at": "2026-10-17T20:15:00Z"
}

```

* **Response 400 Bad Request:** `{"error": "INVALID_URL_FORMAT"}` / `{"error": "MINIMUM_2_MEMBERS_REQUIRED"}`
* **Response 403 Forbidden:** `{"error": "DEADLINE_PASSED"}` (Hard Lock после 00:00:00)

---

### 2.5. Контур Жюри (`/jury`)

#### `POST /api/v1/jury/register`

Регистрация эксперта по секретному ключу из `.env`.

* **Auth:** Не требуется
* **Request Body:**

```json
{
  "secret_key": "SUPER_SECRET_JURY_KEY_FROM_ENV",
  "full_name": "Смирнов Алексей Владимирович",
  "email": "jury_smirnov@example.com",
  "password": "jurypassword123"
}

```

* **Response 201 Created:** `{"message": "Jury registered successfully"}`
* **Response 403 Forbidden:** `{"error": "INVALID_SECRET_KEY"}`

#### `POST /api/v1/jury/login`

Авторизация эксперта жюри.

* **Auth:** Не требуется
* **Request Body:**

```json
{
  "email": "jury_smirnov@example.com",
  "password": "jurypassword123"
}

```

* **Response 200 OK:** (Записывает `access_token` с ролью `JURY` в `HttpOnly` Cookie)

#### `GET /api/v1/jury/teams`

Получение списка всех сформированных команд, составов и ссылок на работы.

* **Auth:** Требуется (`JURY`)
* **Response 200 OK:**

```json
{
  "teams": [
    {
      "team_id": "c39a2828-82ef-4171-a75d-354a93edb8f2",
      "team_name": "CyberPunks",
      "solution_url": "https://disk.yandex.ru/d/example_folder",
      "is_evaluated_by_me": true,
      "members_count": 3
    }
  ]
}

```

#### `GET /api/v1/jury/evaluations`

Получение всех оценок, ранее выставленных текущим жюри.

* **Auth:** Требуется (`JURY`)
* **Response 200 OK:**

```json
{
  "evaluations": [
    {
      "team_id": "c39a2828-82ef-4171-a75d-354a93edb8f2",
      "criterion_id": 1,
      "score": 9
    },
    {
      "team_id": "c39a2828-82ef-4171-a75d-354a93edb8f2",
      "criterion_id": 2,
      "score": 8
    }
  ]
}

```

#### `POST /api/v1/jury/evaluations`

Сохранение/обновление персональных оценок для конкретной команды.

* **Auth:** Требуется (`JURY`)
* **Request Body:**

```json
{
  "team_id": "c39a2828-82ef-4171-a75d-354a93edb8f2",
  "scores": [
    {"criterion_id": 1, "score": 9},
    {"criterion_id": 2, "score": 8},
    {"criterion_id": 3, "score": 10}
  ]
}

```

* **Response 200 OK:** `{"message": "Scores saved successfully"}`

---

### 2.6. Административный контур (`/admin`)

#### `GET /api/v1/admin/stats`

Сводная статистика платформы.

* **Auth:** Требуется (`ADMIN`)
* **Response 200 OK:**

```json
{
  "total_users": 240,
  "total_teams": 65,
  "submitted_solutions": 58,
  "total_juries": 8
}

```

#### `GET /api/v1/admin/export/excel`

Генерация и скачивание `.xlsx` файла с итоговыми результатами.

* **Auth:** Требуется (`ADMIN`)
* **Response 200 OK:** Бинарный файл Excel (`Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`).