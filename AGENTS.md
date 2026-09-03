# AGENTS.md — SPCase repository context router

## Project identity

SPCase (spcase.ru) is a case-championship platform with a Go backend,
PostgreSQL, Docker Compose/Nginx production infrastructure, a legacy
server-rendered frontend in `web/`, and an independent React migration target
in `frontend/`.

The backend/database/production baseline is mature. Frontend replacement is
in progress and production cutover has NOT happened. Current project status
and the next approved stage live in `ROADMAP.md`.

For any task under `frontend/`, read `frontend/AGENTS.md` first.

## Architectural boundaries

```text
HTTP Handler → Service → Repository → PostgreSQL
```

- Handlers own HTTP concerns, not SQL.
- Services own business logic, not HTTP transport.
- Repositories own persistence, not HTTP responses.
- Domain owns entities, rules and stable errors.
- `web/` is the legacy behavioral reference until frontend cutover.
- `frontend/` is the independent React migration target consuming `/api/v1`.

Do not move logic between layers, add abstractions, or perform style-only
refactors without a concrete task reason.

## Authority model

There is no universal "code always wins" rule. Use the source that is
authoritative for the concern being changed.

| Concern | Authority order |
|---|---|
| Business behavior / lifecycle | `docs/domain/business-rules.md` → `docs/contracts/http-api.md` → tests/current code as implementation evidence |
| Database ownership / schema / transactions | `docs/architecture/database.md` → migrations/tests/current code |
| System architecture / deployment | accepted ADRs + `docs/architecture/system.md` → runbooks/cutover plans → current implementation |
| Frontend behavior / parity | `docs/frontend/legacy-contract.md` + `docs/contracts/http-api.md` → current legacy behavior/tests |
| Frontend UX model | `docs/frontend/experience-model.md` → behavioral contracts |
| Frontend visual intent | `docs/frontend/design-direction.md` → latest `docs/frontend/visual-acceptance.md` → current implementation as material to evaluate |
| Current phase / next work | `ROADMAP.md` |
| Historical reasoning | Git history |

A disagreement between implementation and an authoritative contract is a
discrepancy to resolve, not an automatic reason to declare either side
correct. Do not silently rewrite behavior or documentation to hide a conflict.

## Context loading

Read only the documents needed for the task. Do not load the entire repository
context by default.

| Task | Read first |
|---|---|
| Backend/service/repository work | relevant code + `docs/domain/business-rules.md`; add `docs/contracts/http-api.md` when HTTP behavior is involved |
| Database/migration/ACL work | `docs/architecture/database.md` + relevant migration/runbook |
| Public frontend visual work | `frontend/AGENTS.md` + `docs/frontend/design-direction.md` + `docs/frontend/visual-acceptance.md` |
| New USER/JURY/ADMIN workflow | `frontend/AGENTS.md` + `docs/frontend/experience-model.md` + relevant `legacy-contract.md` and HTTP contract sections |
| Frontend API integration | `frontend/AGENTS.md` + `docs/contracts/http-api.md` + relevant behavioral contract |
| Frontend cutover | `docs/frontend/cutover-plan.md` + system architecture + ADR 0001 |
| Observability / PostgreSQL production operations | matching runbook + architecture document |

Use short cross-references instead of duplicating stable facts across files.

## Documentation ownership

Project/context documentation is authored and maintained by the human +
ChatGPT Web workflow. Coding agents are implementation-only by default.

Unless the user explicitly overrides this rule for a task, coding agents must
not edit project Markdown/context files, including `AGENTS.md`, `README.md`,
`ROADMAP.md`, `frontend/AGENTS.md` and `docs/**/*.md`. If implementation work
reveals that documentation is stale or incomplete, report the discrepancy in
the final report so the documentation owner can update it separately.

Documentation-only commits should remain separate from implementation commits
when practical.

## Global invariants

- Preserve established behavior unless the task explicitly changes an
  authoritative contract or fixes a proven bug.
- All database changes go through Goose migrations. Production migrations use
  only `migrations/production.txt`; never run dev seed `00003` or destructive
  migration helpers in production.
- PostgreSQL privilege separation is security-critical: `spcase_migrator`
  owns schema/DDL, `spcase_app` has runtime DML only, and `postgres` is not a
  runtime application identity. Do not broaden runtime privileges.
- Team lifecycle, Hard Lock, submission and evaluation concurrency rules
  (including row-lock ordering and PostgreSQL-time rechecks) must not be
  weakened.
- Legacy `web/` remains behaviorally intact until the explicit frontend
  cutover stage.
- Experimental browser features must progressively enhance a usable baseline.

## Security constraints

- Never hardcode, print, commit, summarize or copy secrets into agent context.
- Never log passwords, jury keys, JWTs, cookies, database URLs or request
  bodies containing credentials.
- Context/archive bundles must exclude real `.env`, `.env.local`,
  `.env.staging`, `.env.production` and other secret-bearing environment
  files. Keep only public `*.example` templates.
- Do not weaken authentication, authorization, cookie flags, CORS, rate
  limits, input validation or error opacity.
- Auth revocation relies on `auth_version` / `disabled_at` rechecks against
  PostgreSQL on every protected request; do not bypass them.

## Validation

For Go changes:

```bash
gofmt -w .
go test -race ./...
go vet ./...
go build ./...
```

For the independent React frontend:

```bash
cd frontend
pnpm typecheck
pnpm check
pnpm build
# pnpm test:e2e when the task affects browser behavior and browsers are installed
```

For infrastructure changes: `docker compose config --quiet` with the relevant
Compose files/env template.

`make security-check` remains part of the backend/legacy release gate; do not
assume it replaces the independent `frontend/` pnpm checks.

Never run production cutover, migration or deployment procedures unless the
task explicitly requires and authorizes them.

## Git policy

- Work directly on local `main`.
- Atomic commits are allowed when the task requests them.
- Never push, tag or deploy without explicit user instruction.
- Never rewrite history unless explicitly requested.
- Before committing, review `git diff` and include only intended changes.

## Scope discipline

- Work on one logical task at a time.
- Inspect related code and the authoritative documents before editing.
- Prefer the smallest correct change; leave unrelated cleanup alone.
- Prefer existing dependencies and platform features; justify any new runtime
  dependency.
- Do not create duplicate systems, placeholder architecture, or speculative
  abstractions.

<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:970c3bf2 -->
## Beads Issue Tracker

This project uses **bd (beads)** for issue tracking. Run `bd prime` to see full workflow context and commands.

### Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --claim  # Claim work
bd close <id>         # Complete work
```

### Rules

- Use `bd` for ALL task tracking — do NOT use TodoWrite, TaskCreate, or markdown TODO lists
- Run `bd prime` for detailed command reference and session close protocol
- Use `bd remember` for persistent knowledge — do NOT use MEMORY.md files

**Architecture in one line:** issues live in a local Dolt DB; sync uses `refs/dolt/data` on your git remote; `.beads/issues.jsonl` is a passive export. See https://github.com/gastownhall/beads/blob/main/docs/SYNC_CONCEPTS.md for details and anti-patterns.

## Agent Context Profiles

The managed Beads block is task-tracking guidance, not permission to override repository, user, or orchestrator instructions.

- **Conservative (default)**: Use `bd` for task tracking. Do not run git commits, git pushes, or Dolt remote sync unless explicitly asked. At handoff, report changed files, validation, and suggested next commands.
- **Minimal**: Keep tool instruction files as pointers to `bd prime`; use the same conservative git policy unless active instructions say otherwise.
- **Team-maintainer**: Only when the repository explicitly opts in, agents may close beads, run quality gates, commit, and push as part of session close. A current "do not commit" or "do not push" instruction still wins.

## Session Completion

This protocol applies when ending a Beads implementation workflow. It is subordinate to explicit user, repository, and orchestrator instructions.

1. **File issues for remaining work** - Create beads for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **Handle git/sync by active profile**:
   ```bash
   # Conservative/minimal/default: report status and proposed commands; wait for approval.
   git status

   # Team-maintainer opt-in only, unless current instructions forbid it:
   git pull --rebase
   bd dolt push
   git push
   git status
   ```
5. **Hand off** - Summarize changes, validation, issue status, and any blocked sync/commit/push step

**Critical rules:**
- Explicit user or orchestrator instructions override this Beads block.
- Do not commit or push without clear authority from the active profile or the current user request.
- If a required sync or push is blocked, stop and report the exact command and error.
<!-- END BEADS INTEGRATION -->

<!-- BEGIN BEADS CODEX SETUP: generated by bd setup codex -->
## Beads Issue Tracker

Use Beads (`bd`) for durable task tracking in repositories that include it. Use the `beads` skill at `.agents/skills/beads/SKILL.md` (project install) or `~/.agents/skills/beads/SKILL.md` (global install) for Beads workflow guidance, then use the `bd` CLI for issue operations.

### Quick Reference

```bash
bd ready                # Find available work
bd show <id>            # View issue details
bd update <id> --claim  # Claim work
bd close <id>           # Complete work
bd prime                # Refresh Beads context
```

### Rules

- Use `bd` for all task tracking; do not create markdown TODO lists.
- Run `bd prime` when Beads context is missing or stale. Codex 0.129.0+ can load Beads context automatically through native hooks; use `/hooks` to inspect or toggle them.
- Keep persistent project memory in Beads via `bd remember`; do not create ad hoc memory files.

**Architecture in one line:** issues live in a local Dolt DB; sync uses `refs/dolt/data` on your git remote; `.beads/issues.jsonl` is a passive export. See https://github.com/gastownhall/beads/blob/main/docs/SYNC_CONCEPTS.md for details and anti-patterns.
<!-- END BEADS CODEX SETUP -->
