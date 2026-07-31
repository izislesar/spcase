# spcase v1.0.0 — HTTP API

Base path: `/api/v1`. JSON requests должны иметь `Content-Type: application/json`, содержать ровно один object и не могут иметь неизвестные fields. Максимальный body — 1 MiB.

## 1. Authentication contract

Успешная регистрация/login устанавливает cookie:

```text
access_token=<JWT>; Domain=<APP_DOMAIN>; Path=/; HttpOnly; Secure; SameSite=Lax
```

JWT действует 24 часа. Protected endpoints проверяют token и актуальные `role`, `auth_version`, `disabled_at` в PostgreSQL. General login принимает USER и ADMIN; jury login принимает только JURY.

Обозначения access: Public, USER, JURY, ADMIN, Authenticated.

## 2. Error contract

```json
{
  "error": {
    "code": "STABLE_ERROR_CODE",
    "message": "Human-readable message"
  }
}
```

Основное отображение:

- 400 — validation, invalid request, no team, invalid evaluation/URL;
- 401 — invalid credentials, missing/invalid/revoked token, disabled account;
- 403 — wrong role, deadline/registration/Hard Lock, invalid jury key, evaluations closed;
- 404 — unknown route/resource/invite/member/submission;
- 405 — method not allowed;
- 409 — duplicate email/team name, existing membership, full team;
- 415 — non-JSON request;
- 500 — opaque `INTERNAL_ERROR`;
- 503 — readiness failure.

Nginx возвращает тот же envelope для 413 `INVALID_REQUEST`, 429 `RATE_LIMITED` и upstream 503 `SERVICE_UNAVAILABLE`.

## 3. Public endpoints

| Method and path | Success response |
|---|---|
| `GET /health/live` | 200 `{status:"ok",timestamp}`; database не проверяется |
| `GET /health/ready` | 200 `{status:"ready",timestamp}`; 503 `NOT_READY` при failed DB ping |
| `GET /info` | deadlines и текущие `is_registration_open`, `is_submission_open` |
| `GET /schedule` | `{events:[{id,title,start_time,description}]}` |
| `GET /faq` | `{faq:[{id,question,answer}]}` |
| `GET /no-team` | `{message,telegram_chat_url}` |

## 4. Authentication and profile

### `POST /auth/register` — Public

Request: `{full_name, university, email, telegram, password}`.

Создаёт USER до registration deadline, устанавливает cookie. Response 201: `{id,email,role:"USER"}`.

### `POST /auth/login` — Public

Request: `{email,password}`. Допустимые роли: USER, ADMIN.

Response 200: `{status:"success",role}` и access cookie.

### `POST /jury/register` — Public

Request: `{secret_key,full_name,email,password}`.

Создаёт JURY при корректном `JURY_REGISTRATION_KEY`, устанавливает cookie. Response 201: `{message:"Jury registered successfully"}`.

### `POST /jury/login` — Public

Request: `{email,password}`. Допустима только роль JURY.

Response 200: `{status:"success",role:"JURY"}` и access cookie.

### `POST /auth/logout` — Authenticated

Истекает browser cookie. Response 200: `{status:"logged_out"}`. Endpoint не изменяет `auth_version`.

### `GET /user/me` — USER, ADMIN

Response 200:

```json
{
  "id": "uuid",
  "full_name": "...",
  "university": "...",
  "email": "...",
  "telegram": "...",
  "role": "USER",
  "team_status": "NO_TEAM|IN_TEAM|CAPTAIN",
  "team_id": "uuid-or-null"
}
```

## 5. Team endpoints — USER

| Method and path | Request | Success |
|---|---|---|
| `POST /team/create` | `{name}` | 201 `{id,name,invite_code,captain_id}` |
| `POST /team/join` | `{invite_code}` | 200 `{message,team_id,team_name}` |
| `GET /team/my` | — | 200 team details |
| `POST /team/leave` | empty body | 200 message |
| `POST /team/kick` | `{user_id}` | 200 message |
| `POST /team/transfer-ownership` | `{new_captain_id}` | 200 message |
| `DELETE /team/disband` | empty body | 200 message |
| `POST /team/submit` | `{solution_url}` | 200 `{status:"submitted",solution_url,updated_at}` |

`GET /team/my` возвращает `id`, `name`, `invite_code`, `captain_id`, `status_badge`, `mutations_locked`, `members[]` и optional `submission`.

Kick, transfer и disband требуют captain; leave запрещён captain. Leave/kick/transfer/disband блокируются за час до submission deadline. Submit требует captain, 2–4 участников, HTTP(S) URL и время строго до deadline.

## 6. Jury endpoints — JURY

### `GET /jury/teams`

Возвращает только команды с submission:

```json
{
  "teams": [{
    "team_id": "uuid",
    "team_name": "...",
    "solution_url": "https://...",
    "is_evaluated_by_me": false,
    "members_count": 4
  }],
  "evaluations_locked": false
}
```

### `GET /jury/evaluations`

Response: `{evaluations:[{team_id,criterion_id,score}]}` для текущего JURY.

### `POST /jury/evaluations`

Request:

```json
{
  "team_id": "uuid",
  "scores": [
    {"criterion_id": 1, "score": 0}
  ]
}
```

`scores` обязан содержать ровно шесть уникальных criteria 1..6; score каждого — 0..10. Team должен иметь submission, evaluation lifecycle должен быть open. Response 200: `{message:"Scores saved successfully"}`.

## 7. Admin endpoints — ADMIN

| Method and path | Success |
|---|---|
| `GET /admin/stats` | `{total_users,total_teams,submitted_solutions,total_juries,evaluations_closed}` |
| `GET /admin/export/excel` | XLSX `hackathon_results.xlsx` |
| `POST /admin/evaluations/close` | 200 `{message:"Evaluations closed"}` |
| `POST /admin/evaluations/open` | 200 `{message:"Evaluations opened"}` |

Open/close operations идемпотентны по текущему state; audit event создаётся только при реальном переходе.

## 8. Browser routes

Go web handler обслуживает `/`, `/schedule`, `/no-team`, `/login`, `/register`, `/dashboard`, `/jury/login`, `/jury/register`, `/jury/teams`, `/admin`; `/jury` перенаправляется на `/jury/teams`. Assets доступны под `/static/`.
