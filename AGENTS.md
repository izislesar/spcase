# AGENTS.md — SPCase agent context

## Mission

SPCase is a backend/API platform built as a Go service with PostgreSQL and
Docker/Nginx deployment. Keep the repository backend-only until product scope
is explicitly defined in Beads.

This file is the only repository-level agent instruction file. Beads is the
only durable work-state system. Source code and tests are the implementation
truth.

## Context contract

At the start of every session:

```bash
bd prime
bd ready
```

Before editing, inspect the selected bead and its blockers:

```bash
bd show <id>
bd update <id> --claim
```

Do not create planning Markdown files, TODO lists, roadmap files, ad-hoc
memory files, or duplicate architecture notes. If information must survive a
session reset, put it in Beads (`bd remember`, issue descriptions, comments,
dependencies) rather than another document.

Load only the code needed for the selected bead. Do not reconstruct project
state from historical documents or git history unless the bead explicitly
requires historical investigation.

## Architecture

```text
HTTP Handler → Service → Repository → PostgreSQL
```

- Handlers own HTTP transport concerns.
- Services own business rules and orchestration.
- Repositories own persistence and SQL.
- Domain owns entities, rules, and stable errors.
- Migrations own database schema changes.
- Nginx is the external HTTP gateway; the Go service exposes the API.

Do not move responsibilities between layers or add abstractions without a
concrete bead requiring them.

## Non-negotiable invariants

- Preserve established API behavior unless the active bead explicitly changes
  the contract.
- All schema changes use Goose migrations.
- Production migrations use only `migrations/production.txt`.
- Never run development seed migration `00003` in production.
- `spcase_migrator` owns schema/DDL; `spcase_app` gets runtime DML only.
- Never broaden runtime PostgreSQL privileges.
- Authentication, authorization, cookie security, CORS, rate limits, input
  validation, error opacity, and auth-version revocation checks are security
  boundaries.
- Team lifecycle, Hard Lock, submission, evaluation concurrency, row-lock
  ordering, and PostgreSQL-time rechecks must not be weakened.
- Never hardcode, print, commit, or copy secrets into code, issues, logs, or
  agent context.
- Never rewrite git history unless explicitly requested.
- Never perform production deployment or migration unless explicitly requested.

## Work protocol

1. `bd prime` and `bd ready`.
2. Select one unblocked bead.
3. `bd show <id>` and inspect relevant code/tests.
4. Claim it with `bd update <id> --claim`.
5. Implement the smallest correct change.
6. Run focused tests first, then the repository validation gates relevant to
   the change.
7. Record newly discovered durable work as Beads dependencies/tasks.
8. Close only completed beads, with evidence:

```bash
bd close <id> --reason="Completed; validation: <commands>"
```

If a product or architecture decision is required, do not invent it. Create
or update a bead describing the decision/blocker and stop at that boundary.

## Validation

For Go changes:

```bash
gofmt -w .
go test -race ./...
go vet ./...
go build ./...
```

For security-sensitive or release-facing backend changes:

```bash
make security-check
```

For infrastructure changes:

```bash
docker compose config --quiet
```

Integration tests require the disposable PostgreSQL environment described by
the repository's environment examples and test configuration.

## Git policy

- Work on the active branch selected by the user/orchestrator.
- Review `git diff` before committing.
- Keep commits atomic and scoped to the active bead when commits are requested.
- Do not push, tag, merge, or deploy unless explicitly authorized.
