# spcase

SPCase (spcase.ru) is a case-championship platform implemented as a Go HTTP
service with PostgreSQL storage and Docker/Nginx deployment.

## Repository layout

- `cmd/app` — HTTP API service;
- `cmd/admin-bootstrap` — one-shot ADMIN bootstrap CLI;
- `cmd/healthcheck` — container readiness probe;
- `internal/domain` — entities, business rules, stable errors;
- `internal/service` — application and business logic;
- `internal/repository` — PostgreSQL persistence;
- `internal/delivery/http` — HTTP handlers and middleware;
- `migrations` — Goose migrations and production allowlist;
- `scripts` — database operations and production rehearsals;
- `Dockerfile`, `docker-compose*.yml`, `nginx.conf` — deployment.

## Development

Requirements: Go, Docker, and PostgreSQL for integration tests.

```bash
go test -race ./...
go vet ./...
go build ./...
```

Database migrations:

```bash
make migrate-up
make migrate-status
make migrate-reset   # local development only
```

Production migrations are restricted to `migrations/production.txt` and run
through `make migrate-production`.

## Beads

Beads (`bd`) is the durable project task system. `.beads/issues.jsonl` is the
canonical task store for this repository; there is no second planning system.

```bash
bd prime
bd ready
bd show <id>
bd update <id> --claim
bd close <id> --reason="Completed; validation: <commands>"
```

Do not create roadmap files, planning documents, Markdown TODO lists, or ad-hoc
memory files. Put durable work, blockers, dependencies, and project memory in
Beads.

## Local staging

Copy `.env.staging.example` to `.env.staging`, fill the required secrets, then
validate and start the stack:

```bash
docker compose -p spcase-staging \
  -f docker-compose.yml -f docker-compose.staging.yml \
  --env-file .env.staging config --quiet

docker compose -p spcase-staging \
  -f docker-compose.yml -f docker-compose.staging.yml \
  --env-file .env.staging up --detach --build --wait
```

Health endpoints:

```bash
curl --fail http://127.0.0.1:18080/api/v1/health/live
curl --fail http://127.0.0.1:18080/api/v1/health/ready
```

## Production image

```bash
docker build --pull -t spcase:local .
docker compose --env-file .env.production config --quiet
docker compose --env-file .env.production build --pull
docker compose --env-file .env.production up --detach --wait
```

The application image contains only the Go binaries and runtime dependencies.
PostgreSQL migrations run separately through the migrator image.
