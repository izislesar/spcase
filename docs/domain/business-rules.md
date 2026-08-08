# spcase v1.0.0 — Business Rules

## 1. Identity and roles

All identities are accounts in `users` with one global role:

- `USER` — participant profile and team operations;
- `JURY` — submitted-team registry and personal evaluations;
- `ADMIN` — statistics, XLSX export and evaluation lifecycle control.

Roles are mutually exclusive. Team membership state (`NO_TEAM`, `IN_TEAM`, `CAPTAIN`) is derived from current database relations and is never stored in JWT.

USER requires full name, university, email, Telegram and password. JURY requires secret key, full name, email and password. The first ADMIN is created only by the bootstrap CLI.

Participant registration is allowed strictly before `REGISTRATION_DEADLINE`. Jury registration is controlled by `JURY_REGISTRATION_KEY`, not that deadline; the key is verified with a SHA-256-backed constant-time comparison. The bootstrap CLI serializes creation of the first ADMIN through a PostgreSQL advisory transaction lock and accepts the password only from stdin.

## 2. Authentication and revocation

- Password length: 8..72 bytes; hashes use bcrypt.
- Email is trimmed, lowercased and case-insensitively unique.
- Access JWT lifetime: 24 hours; algorithm: HS256.
- Claims include account UUID, role and `auth_version`; team information is excluded.
- Cookie is `HttpOnly`, `Secure`, `SameSite=Lax`, scoped to `APP_DOMAIN` and `/`.

Every protected request validates current account state from PostgreSQL. Token is rejected when account is missing/disabled or role/auth version differs from claims. Password changes, disabled-state changes and auth-version increments advance `auth_version`.

Logout deletes the browser cookie only. A copied token remains usable until expiry or an account-state revocation occurs.

## 3. Team lifecycle

- One USER can belong to at most one team.
- Team size is 1..4; at least two members are required only to retain/submit a solution.
- Creating a team makes its creator captain and member atomically.
- Team name is case-insensitively unique and at most 100 characters.
- Invite code is cryptographically generated from `A-Z0-9`, exactly 8 characters; generation retries collisions up to eight times.
- Join normalizes invite code to uppercase and is serialized under team/user locks.
- Captain cannot leave or kick themselves.
- Kick, transfer ownership and disband require the current captain.
- New captain must already be a member of the same team.
- Disband deletes membership, submission and evaluations through cascades but keeps user accounts.

Status badge:

- `SUBMITTED` when a submission exists;
- otherwise `READY` for 2–4 members;
- otherwise `SEARCHING`.

## 4. Team Hard Lock

At `SUBMISSION_DEADLINE - 1 hour` the following endpoints close:

- leave;
- kick;
- transfer ownership;
- disband.

The middleware performs an early check. Repository repeats it using PostgreSQL `clock_timestamp()` after row locks, preventing a request queued before the boundary from committing after it.

Current implementation does not apply this lock to create or join.

If leave/kick reduces a team below two members, its existing submission is deleted in the same transaction. Existing evaluations are retained as historical team records; they become writable again only after the team regains eligibility, creates a current submission and evaluations are open. Disband is different: deleting the team cascades to both submission and evaluations.

## 5. Submissions

- Only the current captain may submit.
- Team must contain 2–4 members.
- URL must be absolute HTTP or HTTPS, have a hostname and be at most 2048 bytes.
- Submission is accepted strictly before `SUBMISSION_DEADLINE`.
- One current submission exists per team; later valid submissions update URL and timestamp.
- Repository rechecks captain, size and database time under a team row lock.
- Jury workspace exposes only teams with a submission.

## 6. Jury scoring

Each JURY submits exactly six scores for one submitted team:

- criterion IDs: every integer 1..6 exactly once;
- score: integer 0..10;
- maximum per-jury team total: 60.

The batch is validated before persistence, sorted by criterion and upserted atomically. A JURY can update their own six rows while evaluation lifecycle is open. Evaluation identity is unique by `(jury_id, team_id, criterion_id)`.

Scoring, submission writes and destructive membership mutations serialize on the affected team row. The scoring lock order is team → global evaluation state → current submission. If scoring obtains the team lock first, its complete batch commits before a later eligibility loss; the later mutation removes the submission but retains the committed historical evaluations. If the membership mutation commits first, scoring fails with `SUBMISSION_NOT_FOUND`.

`is_evaluated_by_me` becomes true only when the current JURY has all six criterion rows for the team.

Team total is the sum of all persisted criterion scores from all juries. Missing jury evaluations do not block aggregation. `evaluated_by_count` counts distinct juries with rows for that team.

## 7. Evaluation lifecycle

`evaluation_state` is a global singleton:

- open: JURY batch writes are allowed;
- closed: writes fail with `EVALUATIONS_LOCKED`;
- reads, admin statistics and export remain available.

Only ADMIN may open or close evaluations. The state row is locked during transition. Repeating the current state is idempotent; an immutable `OPEN`/`CLOSE` audit event is appended only when state changes.

Closing and reopening evaluations does not delete or reset scores. A close transition serializes with in-flight scoring through the state row: an already locked score batch completes first, otherwise the batch observes the closed state and fails with `EVALUATIONS_LOCKED`.

## 8. Administration and export

Admin statistics report USER count, team count, submitted solutions, JURY count and current evaluation state.

XLSX export contains:

- summary rows with team, captain, members, solution, total score and jury coverage;
- detail rows with team, jury, criterion and score.

Teams without submissions/evaluations remain in the summary with explicit default values. Export aggregation avoids multiplying scores by team-member joins.

## 9. Public state

Public info derives registration/submission openness from current UTC time and configured deadlines. Schedule and FAQ are immutable JSON content embedded in the binary. The no-team endpoint returns the configured HTTPS Telegram URL.
