# backend.md — Спецификация Backend-Архитектуры (Golang)

---

## 1. Общие Принципы и Серверная Архитектура

Бэкенд представляет собой высокопроизводительный **чистый монолит на Golang**. Системная архитектура ориентирована на максимальную производительность, минимальный memory footprint и полное отсутствие избыточных внешних зависимостей (оверинжиниринга).

### Ключевые требования:

* **Асинхронность и конкурентность:** Использование встроенной модели goroutines/channels и контекстов (`context.Context`) для отмены таймаутов.
* **Изоляция:** Развертывание системы под управлением Arch Linux (host OS) с проксированием через **Nginx** (Reverse Proxy, SSL-termination, Rate Limiting).
* **Stateless-архитектура:** Сервер не хранит состояние сессий в памяти или БД. Перезапуск Go-бинарника (например, через `systemd`) происходит без разрыва сессий авторизованных пользователей.
* **Zero-Redis Policy:** Отказ от Redis в пользу эффективного connection pool к PostgreSQL (`pgxpool`) и при необходимости In-Memory кэширования в памяти Go-процесса (`sync.RWMutex` / `atomic.Value`).

---

## 2. Структура Репозитория (Standard Go Layout)

Применяется слоистая Clean Architecture (Domain-Driven Design layout):

```text
.
├── cmd/
│   └── app/
│       └── main.go                 # Инициализация конфигов, БД, роутера и graceful shutdown
├── internal/
│   ├── config/
│   │   └── config.go               # Загрузка переменной окружения (.env) через cleanenv/godotenv
│   ├── domain/                     # Бизнес-сущности и ошибки
│   │   ├── user.go
│   │   ├── team.go
│   │   ├── submission.go
│   │   ├── evaluation.go
│   │   └── errors.go
│   ├── delivery/
│   │   └── http/                   # HTTP-адаптеры (Fiber / Chi / Net/HTTP)
│   │       ├── middleware/
│   │       │   ├── auth.go         # Проверка JWT из HttpOnly Cookie
│   │       │   ├── role.go         # RBAC (USER, JURY, ADMIN)
│   │       │   └── hardlock.go     # Блокировка операций по таймеру
│   │       └── v1/                 # Хэндлеры эндпоинтов V1
│   │           ├── auth.go
│   │           ├── user.go
│   │           ├── team.go
│   │           ├── submission.go
│   │           ├── jury.go
│   │           ├── admin.go
│   │           └── public.go
│   ├── service/                    # Бизнес-логика (Use Cases)
│   │   ├── auth_service.go
│   │   ├── team_service.go
│   │   ├── submission_service.go
│   │   ├── jury_service.go
│   │   └── export_service.go
│   └── repository/                 # Data Access Layer (PostgreSQL)
│       └── postgres/
│           ├── user_repo.go
│           ├── team_repo.go
│           ├── submission_repo.go
│           └── evaluation_repo.go
├── migrations/                     # SQL миграции (goose)
│   ├── 000001_init_schema.up.sql
│   └── 000001_init_schema.down.sql
├── go.mod
└── go.sum

```
---

## 3. Environment Variables & Domain Configuration

Target Domain: `spcase.ru`

### Required ENV Variables:
- `APP_DOMAIN`: `spcase.ru` (используется для параметра Domain в Set-Cookie)
- `CORS_ALLOWED_ORIGINS`: `https://spcase.ru,https://www.spcase.ru`

### Security Policy Rules for Code Generation:
1. **HttpOnly Cookies:** Все Auth-токены выставляются с параметрами:
   - `Domain`: `spcase.ru` (или берутся из `cfg.AppDomain`)
   - `HttpOnly`: `true`
   - `Secure`: `true`
   - `SameSite`: `Lax`
2. **CORS Headers:** Мидлварь CORS должна разрешать только origin `https://spcase.ru` и обязательно включать `AllowCredentials: true`.

---

## 4. Аутентификация и Безопасность (JWT & Security)

### 4.1. Механизм JWT

1. **Токен:** Stateless JWT, содержащий `user_id`, `role` (`USER`, `JURY`, `ADMIN`), `team_id` (опционально) и `exp` (время жизни — 24 часа).
2. **Передача:** Токен передается **исключительно** через куку `access_token`.
3. **Параметры Cookie:**
* `HttpOnly: true` (защита от XSS, недоступно из JavaScript).
* `Secure: true` (передача только по HTTPS).
* `SameSite: http.SameSiteLaxMode` (защита от CSRF).
* `Path: "/"`



### 4.2. Регистрация Жюри

* Эндпоинт `POST /api/v1/jury/register` проверяет заголовок или поле `secret_key`.
* Бэкенд сравнивает значение с переменной окружения `JURY_REGISTRATION_KEY` из `.env` за постоянное время (`subtle.ConstantTimeCompare`), предотвращая атаки по времени (timing attacks).

---

## 5. Бэкенд-Логика и Бизнес-Правила

### 5.1. Генерация Инвайт-Кодов

При создании команды бэкенд генерирует случайный 8-значный криптографически стойкий код (`[A-Z0-9]`) с помощью `crypto/rand`.

```go
// Пример: "X8K2M9N7"
func GenerateInviteCode() string

```

Код проверяется на коллизии уникальным индексом `UNIQUE` в PostgreSQL.

### 5.2. Транзакции и Атомарность (ACID)

Все операции мутации состава команды выполняются строго в рамках изолированных PostgreSQL-транзакций (`pgx.Tx`) с уровнем изоляции `Read Committed` или `Repeatable Read` во избежание race condition при одновременном входе участников:

* **Join Team:** Бэкенд блокирует запись команды (`SELECT ... FOR UPDATE`), проверяет условие `COUNT(members) < 4`, добавляет пользователя и коммитит транзакцию.

### 5.3. Механизм Hard Lock

Бэкенд высчитывает дедлайны на основе системного времени `time.Now().UTC()` и значений из конфигуратора:

1. **Team Hard Lock (`deadline - 1 hour`):** Middleware `HardLockMiddleware` перехватывает запросы к `POST /team/leave`, `POST /team/kick`, `POST /team/transfer-ownership`, `DELETE /team/disband` и возвращает `403 Forbidden` (`"Team mutations are locked 1 hour before submission deadline"`).
2. **Submission Hard Lock (`deadline`):** Роут `POST /team/submit` моментально отклоняет запросы после наступления дедлайна с HTTP-кодом `403 Forbidden` (`"Submission deadline has passed"`).

---

## 6. Генерация Excel-Отчетов (Admin Export)

Для экспорта итоговых оценок используется высокопроизводительная библиотека `[github.com/xuri/excelize/v2](https://github.com/xuri/excelize/v2)`.

### Алгоритм сборки отчета:

1. Выполняется один оптимизированный SQL-запрос с `LEFT JOIN` таблиц `teams`, `users`, `submissions` и `evaluations`.
2. В памяти Go формируется документ Excel:
* **Лист 1 (Сводка):** Таблица команд, ФИО капитана, состав, ссылка на решение, итоговая сумма баллов от всех жюри.
* **Лист 2 (Детализация по жюри):** Матрица оценок (Команда $\times$ Жюри $\times$ Критерии).


3. Бинарный поток данных стримится напрямую в ответ клиенту без сохранения промежуточных файлов на диск сервера:
* `Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`
* `Content-Disposition: attachment; filename="hackathon_results.xlsx"`



---

## 7. Обработка Ошибок и Статус-Коды

Бэкенд возвращает унифицированный JSON-ответ при ошибках:

```json
{
  "error": {
    "code": "TEAM_FULL",
    "message": "Maximum team capacity (4 members) reached"
  }
}

```

### Маппинг Статусов:

* `400 Bad Request` — Ошибка валидации входящих данных (неверный формат URL, короткий пароль).
* `401 Unauthorized` — Отсутствует или просрочен JWT в Cookie.
* `403 Forbidden` — Нарушение прав доступа или сработавший Hard Lock.
* `404 Not Found` — Команда по инвайт-коду или ресурс не найден.
* `409 Conflict` — Пользователь уже состоит в другой команде.
* `500 Internal Server Error` — Необработанная ошибка БД или сервера (детали логируются на сервере, пользователю отдается обезличенное сообщение).