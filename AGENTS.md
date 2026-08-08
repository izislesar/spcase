# AGENTS.md — SPCase agent context

## Project identity

SPCase (spcase.ru) is a case-championship platform: Go backend, PostgreSQL,
server-rendered frontend, Docker Compose production environment with Nginx.

Current state: **production candidate in transition**. Backend, database and
production infrastructure are complete. An approved decision
(`docs/decisions/0001-frontend-v2.md`) will replace the server-rendered
frontend with an independent frontend application. That migration has NOT
started: the existing `web/` frontend is the current implementation and the
behavioral reference until parity is captured and cutover is accepted.

## Architectural boundaries

```text
HTTP Handler → Service → Repository → PostgreSQL
```

- Handlers handle HTTP concerns only (no SQL).
- Services contain business logic (no HTTP dependencies).
- Repositories contain persistence logic (no HTTP responses).
- Domain contains entities, rules and stable errors.
- Frontend (current): server-rendered templates/assets embedded via `web/`.
- Frontend (future): independent application consuming `/api/v1` — see
  `docs/frontend/architecture.md`. Do not implement it outside the roadmap
  stages.

Do not move logic between layers or introduce new abstractions without a
strong, stated reason. Existing working code is preferred over theoretical
improvement; no style-only refactors.

## Source of truth

Priority order:

1. Current working code
2. Documentation map below
3. `ROADMAP.md`
4. Git history

When documentation disagrees with code, code wins; fix the documentation,
not the behavior, unless there is a proven bug.

Read the document matching your task before editing:

| Task area | Authoritative document |
|---|---|
| System architecture, layers, config, deployment topology | `docs/architecture/system.md` |
| Schema, migrations, transactions, PostgreSQL roles/ACL | `docs/architecture/database.md` |
| Business rules, lifecycles, scoring, deadlines | `docs/domain/business-rules.md` |
| HTTP endpoints, auth cookie, error contract | `docs/contracts/http-api.md` |
| Frontend target direction and stack | `docs/frontend/architecture.md`, `docs/frontend/design-direction.md` |
| Legacy frontend behavioral parity contract | `docs/frontend/legacy-contract.md` |
| Frontend cutover topology and mechanics | `docs/frontend/cutover-plan.md` |
| Frontend replacement decision | `docs/decisions/0001-frontend-v2.md` |
| Observability operations | `docs/runbooks/observability.md` |
| PostgreSQL production cutover | `docs/runbooks/postgres-cutover.md` |
| Phased plan and current stage | `ROADMAP.md` |
| Human onboarding, dev commands | `README.md` |

Keep these documents current when you change what they describe. Do not
duplicate facts across documents; use short cross-references.

## Global invariants

- Code behavior is preserved unless the task explicitly targets a proven bug.
- All database changes go through Goose migrations; production migrations use
  only the `migrations/production.txt` allowlist; never run dev seed
  (`00003`) or `migrate-down`/`migrate-reset` in production.
- PostgreSQL privilege separation is security-critical: `spcase_migrator`
  owns schema/DDL, `spcase_app` has runtime DML only, `postgres` is not used
  at runtime. Do not broaden runtime privileges.
- Team lifecycle, Hard Lock, submission and evaluation concurrency rules
  (row-lock ordering, PostgreSQL-time rechecks) must not be weakened.
- The legacy `web/` frontend stays behaviorally intact until the frontend
  cutover stage.
- Experimental browser features (when the new frontend exists) must
  progressively enhance a usable baseline.

## Security constraints

- Never hardcode secrets; never log passwords, jury keys, JWTs, cookies,
  database URLs or request bodies.
- Do not weaken authentication, authorization checks, cookie flags, CORS,
  rate limits or input validation.
- Do not expose sensitive details in error responses; internal errors stay
  opaque.
- Auth revocation relies on `auth_version`/`disabled_at` rechecks against
  PostgreSQL on every protected request — do not bypass.

## Validation

After Go code changes, run:

```bash
gofmt -w .
go test -race ./...
go vet ./...
go build ./...
```

For infrastructure changes: `docker compose config --quiet`.

Full gate: `make security-check`. Do not remove tests to make builds pass.
Never run production cutover, migration or deployment procedures unless the
task explicitly requires and authorizes them.

## Git policy

- Work directly on local `main`.
- Atomic commits are allowed when the task requests them.
- Never push, tag, or deploy without explicit user instruction.
- Never rewrite history unless explicitly requested.
- Before committing: review `git diff`, include only intended changes.

## Scope discipline

- Work on one logical task at a time; do not combine unrelated improvements.
- Before changing code: understand the problem, inspect related files, check
  existing patterns, find the smallest correct solution.
- Keep edits scoped to what the request implies; leave unrelated refactors,
  renames and metadata churn alone.
- Prefer existing dependencies and the Go standard library; justify any new
  dependency.
- Do not create placeholder files or duplicate systems.
