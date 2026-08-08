# Legacy frontend behavioral contract

> **Purpose:** migration parity contract for replacing the server-rendered
> frontend (`web/`). Every requirement below describes **observable behavior**
> of the current implementation that the future frontend must preserve, unless
> explicitly listed in "Not automatically preserved". This is not an
> implementation description: Alpine.js, GSAP, Lenis and the template structure
> are legacy mechanisms, not requirements.
>
> **Sources:** `web/handler.go`, `web/handler_test.go`, `web/template/*.html`,
> `web/src/app.js`, `web/src/interactions/focus.js`, `web/src/animations/*.js`,
> `web/src/input.css`, cross-checked against `../contracts/http-api.md`,
> `../domain/business-rules.md`, `internal/domain/errors.go`,
> `internal/delivery/http/middleware/hardlock.go` and the service test suites.
> Where documentation and code disagreed, code and tests won.

## 1. Legacy route map

Browser routes served by the Go web handler (`web/handler.go`). All page
routes return `200 text/html` with `Cache-Control: no-store` to **any**
visitor — there is no server-side page authorization; access control happens
through the API and client-side redirects (see AUTH requirements).

| Route | Page | Auth required to use |
|---|---|---|
| `/` | Home: hero, format, schedule preview, FAQ, team entry modal | No (team modal needs USER session) |
| `/schedule` | Full schedule | No |
| `/no-team` | Team-finding instructions + Telegram chat link | No |
| `/login` | Participant/admin login | No |
| `/register` | Participant registration | No |
| `/dashboard` | Participant workspace (team, submission) | USER |
| `/jury/login` | Jury login | No |
| `/jury/register` | Jury registration by secret key | No |
| `/jury/teams` | Jury workspace (scoring) | JURY |
| `/admin` | Admin panel (stats, export, evaluation state) | ADMIN |
| `/jury` | 307 Temporary Redirect → `/jury/teams` | — |
| any other path | 404 (plain Go `http.NotFound`, no styled page) | — |
| `/static/*` | Embedded assets, `Cache-Control: public, max-age=31536000, immutable` | No |

Asset URLs carry a content-derived `?v=<12-char hash>` cache-buster
(`web/handler.go:104`, verified by `web/handler_test.go`).

## 2. Frontend → API dependency map

All API calls use `credentials: "include"`, `Accept: application/json`, and
`Content-Type: application/json` when a body is present. Base path `/api/v1`.

| Page component | Endpoints | On 401 |
|---|---|---|
| Home (`homePage`) | `GET /info`, `GET /faq`, `GET /schedule`, `GET /user/me` (probe for team-entry section) | ignored silently (no redirect) |
| Home team modal | `POST /team/create`, `POST /team/join` | redirect `/login` |
| Schedule page | `GET /schedule` | n/a (public) |
| No-team page | `GET /no-team` | n/a (public) |
| Login page | `POST /auth/login` | shown as form error |
| Register page | `POST /auth/register` | shown as form error |
| Jury login/register | `POST /jury/login`, `POST /jury/register` | shown as toast |
| Dashboard | `GET /user/me`, `GET /info`, `GET /team/my`, `POST /team/create`, `POST /team/join`, `POST /team/leave`, `POST /team/kick`, `POST /team/transfer-ownership`, `DELETE /team/disband`, `POST /team/submit`, `POST /auth/logout` | redirect `/login` |
| Jury workspace | `GET /jury/teams`, `GET /jury/evaluations`, `POST /jury/evaluations`, `POST /auth/logout` | redirect `/jury/login` |
| Admin panel | `GET /admin/stats`, `GET /admin/export/excel`, `POST /admin/evaluations/open`, `POST /admin/evaluations/close`, `POST /auth/logout` | redirect `/login` |

The 401 redirect target is contextual: paths starting with `/jury` go to
`/jury/login`, everything else to `/login` (`app.js:21-27, 89-91`).

## 3. Role capability matrix

Browser pages are served to everyone; capabilities below are enforced by the
API plus client-side behavior.

| Capability | Anonymous | USER | JURY | ADMIN |
|---|---|---|---|---|
| Public pages/data (`/`, `/schedule`, `/no-team`, FAQ, info) | ✓ | ✓ | ✓ | ✓ |
| Home team-entry section (create/join modal) | hidden | only when `team_status = NO_TEAM` | hidden | hidden |
| Register participant | ✓ | page accessible, API decides | ✓ (page) | ✓ (page) |
| Login via `/login` | ✓ | ✓ → `/dashboard` | rejected by API (wrong role) | ✓ → `/admin` |
| Login via `/jury/login` | ✓ | rejected by API | ✓ → `/jury/teams` | rejected by API |
| Jury registration by secret key | ✓ | n/a | ✓ | n/a |
| `/dashboard` page usage | redirected to `/login` (on API 401) | full | client-redirected to `/jury/teams` | client-redirected to `/admin` |
| `/jury/teams` page usage | redirected to `/jury/teams`→`/jury/login` (on API 401) | API 403 → error toast, empty page | full | API 403 → error toast |
| `/admin` page usage | redirected to `/login` (on API 401) | API 403 → error toast | API 403 → error toast | full |
| Team create/join/leave/kick/transfer/disband | — | role/membership-dependent (see TEAM) | — | — |
| Submit/update solution | — | captain only, 2–4 members, before deadline | — | — |
| View/join via invite code | — | any USER without a team | — | — |
| Score teams | — | — | ✓ while evaluations open | — |
| View stats, export XLSX, open/close evaluations | — | — | — | ✓ |

Notes:

- Only the dashboard performs a client-side role redirect
  (`app.js:412-415`: non-USER roles are forwarded to their own workspace).
  The jury and admin workspaces rely purely on API 401/403 outcomes.
- `GET /user/me` accepts USER and ADMIN; the dashboard still redirects ADMIN
  away because `profile.role !== "USER"`.

## 4. Deadline / lock behavior matrix

Server enforcement is authoritative (`docs/domain/business-rules.md`,
`internal/domain/errors.go`); the frontend mirrors lock states in UI.

| Boundary | Server behavior | Frontend behavior |
|---|---|---|
| `REGISTRATION_DEADLINE` | `POST /auth/register` rejected (403) | home countdown switches to "Регистрация завершена" and stops; registration form stays submittable and shows the server error message |
| `SUBMISSION_DEADLINE − 1h` (Hard Lock) | `leave`, `kick`, `transfer-ownership`, `disband` → 403 `MUTATIONS_LOCKED` (middleware + PostgreSQL-time recheck) | `GET /team/my.mutations_locked` disables those buttons and shows a lock notice; a live 403 `MUTATIONS_LOCKED` sets the same state without reload; create/join/submit remain available |
| `SUBMISSION_DEADLINE` | `POST /team/submit` → 403 `DEADLINE_PASSED` | submission form disabled + "Приём решений завершён" notice, computed from `GET /info.submission_deadline` against the client clock (30 s polling) and set immediately on a live 403 `DEADLINE_PASSED` |
| Evaluation state closed (ADMIN toggle) | `POST /jury/evaluations` → 403 `EVALUATIONS_LOCKED` | jury inputs/save buttons disabled, read-only notice; set from `GET /jury/teams.evaluations_locked` at load and live on a 403 `EVALUATIONS_LOCKED` |
| Team loses 2nd member (leave/kick) | existing submission deleted in the same transaction; evaluations retained as history | destructive confirmation dialog warns that the solution link will be deleted when the team has a submission and exactly 2 members |

Lock codes are delivered to the UI through a dedicated channel: a 403 whose
error code is `MUTATIONS_LOCKED`, `DEADLINE_PASSED` or `EVALUATIONS_LOCKED`
dispatches a `spcase:lock` event that the open page applies to its state
(`app.js:6-10, 92-94`). The parity requirement is the outcome — lock state
propagates to an open page without reload — not the event mechanism.

## 5. Requirements

### PUBLIC

- PUBLIC-001 — `/` renders without authentication and loads `GET /info`,
  `GET /faq` and `GET /schedule` in parallel; failures show an error
  notification and the page still finishes its loading state.
- PUBLIC-002 — The home hero shows a live countdown to
  `info.registration_deadline` (`Nd HH:MM:SS`, ticking every second); at/after
  the deadline it switches to a terminal "registration closed" state and the
  timer stops and is cleaned up.
- PUBLIC-003 — Dates from the API are formatted in the visitor's local
  timezone using the `ru-RU` long-date/short-time format; missing/invalid
  values render as "—". The schedule pages state that times are shown in the
  participant's timezone.
- PUBLIC-004 — The home schedule preview and the `/schedule` page render the
  server-provided event list (title, start time, description) as an ordered
  timeline; while loading, a polite `role="status"` loading indicator is
  exposed.
- PUBLIC-005 — FAQ renders as an accordion: one open item at a time, toggle
  button with `aria-expanded`/`aria-controls`, answer region
  `aria-labelledby`-linked and `aria-hidden` when collapsed.
- PUBLIC-006 — `/no-team` loads instructions and `telegram_chat_url` from
  `GET /no-team`; until the URL arrives, the Telegram button is inert
  (`aria-disabled`, removed from tab order, click prevented); the message
  block is `aria-live="polite"`.
- PUBLIC-007 — The team-entry section on `/` appears only for an
  authenticated USER with `team_status = NO_TEAM` (probed via `GET /user/me`;
  401/403 are ignored silently). It offers "I have a team" (modal) and "I
  need a team" (`/no-team`).
- PUBLIC-008 — Navigation (desktop links and mobile menu) covers: home
  sections (`/#about`, `/#faq`), `/schedule`, `/dashboard`, `/login`; the
  footer links `/register`, `/schedule`, `/jury/login`.
- PUBLIC-009 — All pages are served with `Cache-Control: no-store`; static
  assets are immutable with a content-derived version query. HTML must never
  be served from cache after mutations.
- PUBLIC-010 — Unknown browser paths produce 404. `/jury` redirects (307) to
  `/jury/teams`.
- PUBLIC-011 — The home page heading is the restrained "СПК кейс-чемпионат"
  without promotional slogans (pinned by `web/handler_test.go`).

### AUTH

- AUTH-001 — Login/registration success sets the `access_token` cookie
  server-side (`HttpOnly; Secure; SameSite=Lax; Path=/; Domain=APP_DOMAIN`);
  the frontend never reads the token and always sends credentials with every
  API call.
- AUTH-002 — Participant login (`/login`) validates client-side (email must
  contain "@", password non-empty, field-level error messages with
  `role="alert"`), posts to `POST /auth/login`, then redirects by the role in
  the response: USER → `/dashboard`, ADMIN → `/admin`, JURY → `/jury/teams`.
- AUTH-003 — General login accepts only USER and ADMIN accounts; jury
  credentials are rejected by the API and surface as an error notification.
- AUTH-004 — Participant registration (`/register`) requires all five fields
  (ФИО, ВУЗ, Telegram, email, password ≥ 8 chars; client checks non-empty,
  email shape, password length, single form-level error), posts to
  `POST /auth/register`, then redirects to `/dashboard`. Server errors
  (duplicate email, closed registration) are shown as the form error.
- AUTH-005 — Jury registration (`/jury/register`) requires secret key, ФИО,
  email, password ≥ 8; the key field is masked (`type="password"`,
  `autocomplete="off"`); success redirects to `/jury/teams`. Jury login
  (`/jury/login`) accepts only JURY accounts.
- AUTH-006 — Logout (available on dashboard, jury workspace, admin panel)
  posts `POST /auth/logout` and always navigates to `/`, even if the request
  fails. Logout does not revoke the token server-side (documented contract).
- AUTH-007 — Any API 401 outside auth pages triggers a full-page redirect to
  the contextual login page (`/jury/*` → `/jury/login`, else `/login`). On
  auth pages, 401 is shown as an error instead of redirecting.
- AUTH-008 — Auth pages remain reachable while authenticated; no client-side
  "already logged in" redirect exists. Server-side session state decides.
- AUTH-009 — Auth forms carry correct `autocomplete` tokens (`email`,
  `current-password`, `new-password`, `name`) to support password managers.
- AUTH-010 — Submitting auth forms disables the submit button and exposes
  busy state (`aria-busy`, "Проверяем…/Создаём…") until the request settles.

### TEAM

- TEAM-001 — A USER without a team sees the onboarding state: create-team
  form (name ≤ 100 chars, required) and join-by-invite form side by side,
  plus a link to `/no-team`.
- TEAM-002 — Invite-code inputs normalize on every keystroke: uppercase,
  strip everything except `A-Z0-9`, truncate to 8 characters; the join button
  stays disabled until exactly 8 characters are present.
- TEAM-003 — Successful create/join from the home modal navigates to
  `/dashboard`; from the dashboard it reloads the page with a success
  notification. API errors (duplicate name, unknown/full invite, already in
  a team) surface as error notifications.
- TEAM-004 — The team view shows: status badge (`SEARCHING` "В поиске" /
  `READY` "Команда готова" / `SUBMITTED` "Решение сдано"), team name,
  captain marker, invite code with copy button, member count `n / 4` with an
  eligibility hint ("can submit" vs "needs at least one more member").
- TEAM-005 — Copy invite writes the code to the clipboard, switches the
  button label to a confirmation for ~2 s, and shows an error notification
  when the clipboard is unavailable (no non-clipboard fallback exists).
- TEAM-006 — Each roster row shows full name, university, a Telegram contact
  link (`https://t.me/<handle>` with leading `@` stripped, new tab,
  `rel="noopener noreferrer"`), a captain badge, and — for the captain only,
  on non-captain rows — a kick button.
- TEAM-007 — Leave (non-captain) and kick (captain) require a native
  confirmation; when the team has a submission and exactly 2 members, the
  confirmation explicitly warns that the solution link will be deleted.
- TEAM-008 — Kick refreshes team data in place after success; leave,
  transfer and disband reload the page after success. (Asymmetry is current
  behavior; see ambiguities.)
- TEAM-009 — Transfer ownership opens a modal listing non-captain members as
  a radio group (`fieldset`/`legend`); confirm is disabled until a member is
  chosen. It is unavailable for single-member teams (button disabled when
  fewer than 2 members).
- TEAM-010 — Disband opens a modal with an explicit irreversibility warning
  (team, roster, solution link and related evaluations are deleted) and
  requires an explicit destructive confirm. Focus starts on "Отмена".
- TEAM-011 — Leave/kick/transfer/disband buttons are disabled while
  `mutations_locked` is true, and a lock notice explains that composition is
  frozen while submission remains possible until the deadline.
- TEAM-012 — Captain-only controls (kick, transfer, disband, submission
  section) are hidden for regular members; the leave button is hidden for
  the captain.
- TEAM-013 — While any team mutation request is in flight, the relevant
  action buttons are disabled (busy state).

### SUBMISSION

- SUBMISSION-001 — The submission section is visible to the captain only.
- SUBMISSION-002 — With an existing submission, the saved URL is shown as an
  external link (new tab, `rel="noopener noreferrer"`) with a "Заменить"
  button that opens the edit form; cancel restores the saved URL.
- SUBMISSION-003 — The URL field is client-validated (`^https?://` prefix,
  field-level error with `aria-invalid`/`aria-describedby`, `type="url"`,
  `inputmode="url"`); a help text states the 2-member minimum and the
  http(s) requirement.
- SUBMISSION-004 — The submit button is disabled while busy, while the
  deadline has passed, or while the team has fewer than 2 members.
- SUBMISSION-005 — Successful submit updates the view in place: saved-URL
  state, badge switches to `SUBMITTED`, success notification; the API upsert
  semantics (later valid submissions replace URL and timestamp) apply.
- SUBMISSION-006 — After the submission deadline the form is replaced by a
  "DEADLINE_PASSED · Приём решений завершён" notice and all submission
  controls are disabled; this state derives from `/info.submission_deadline`
  (client clock, re-checked every 30 s) and from live `DEADLINE_PASSED` API
  rejections.
- SUBMISSION-007 — During Hard Lock (last hour), submission stays available
  while team-composition mutations are blocked; the lock notice states this
  explicitly.

### JURY

- JURY-001 — The workspace loads `GET /jury/teams` and
  `GET /jury/evaluations` in parallel and merges them: every listed team
  gets six criterion inputs pre-filled with the jury's saved scores, default
  0 when no saved row exists.
- JURY-002 — Only teams with a submission are listed; each team card shows
  name, member count, an evaluated marker ("Оценено" / "Требует оценки")
  driven by `is_evaluated_by_me`, and an external "Открыть решение ↗" link
  (new tab, `rel="noopener noreferrer"`).
- JURY-003 — Exactly six numeric score inputs per team (ids 1–6), `min=0
  max=10 step=1`, required; the per-team running total `x / 60` updates
  live.
- JURY-004 — Client-side validation before save: all six values must be
  integers in 0–10; otherwise an error notification and no request. The
  server re-validates and remains authoritative.
- JURY-005 — Saving posts all six scores in one batch; during the save the
  team's inputs and button are disabled ("Сохраняем…"); on success the team
  flips to "Оценено" (`is_evaluated_by_me = true`) with a success
  notification.
- JURY-006 — A "Скрыть проверенные" filter hides evaluated teams; a counter
  shows "Показано: X из Y". The empty state distinguishes "all evaluated"
  from "no submissions yet".
- JURY-007 — When evaluations are closed (`evaluations_locked` at load or a
  live `EVALUATIONS_LOCKED` rejection), all score inputs and save buttons
  turn read-only/disabled and a lock notice explains saved scores remain
  viewable.
- JURY-008 — The workspace finishes its loading state even when initial
  fetches fail (error notification + empty list), keeping the page usable
  for retry via reload.
- JURY-009 — While the page is open, a rejected save with
  `EVALUATIONS_LOCKED` both switches the workspace to locked mode and shows
  the error message.

### ADMIN

- ADMIN-001 — The panel loads `GET /admin/stats` and renders four metrics
  (participants, teams, submitted solutions, jury), each falling back to "—"
  when absent.
- ADMIN-002 — The Excel export button fetches
  `GET /admin/export/excel` with credentials, downloads the response as
  `hackathon_results.xlsx` via an object-URL anchor, and shows a busy state
  ("Готовим файл…") plus a success notification. 401 redirects to `/login`;
  error envelopes surface as notifications.
- ADMIN-003 — The evaluation-state control reflects
  `stats.evaluations_closed`: one button toggles to the opposite state
  (`POST /admin/evaluations/open|close`), disabled while in flight; the row
  visually distinguishes open vs closed state and the copy notes the change
  is global and server-audited.
- ADMIN-004 — After a successful toggle the local state and button label
  update immediately (idempotent server semantics: no audit event when state
  is unchanged — server-side concern, no frontend dependency).

### ERROR

- ERROR-001 — API errors are parsed from the stable envelope
  `{error:{code,message}}`; the server `message` is shown to the user.
  Unparseable error bodies produce a generic "server returned an invalid
  response" error; network failures produce a dedicated connectivity
  message. Users never see raw payloads or stack details.
- ERROR-002 — Global failures (load errors, mutation errors) surface as
  toast notifications; toasts auto-dismiss after 5 s, are manually closable,
  and use `role="alert"` for errors and `role="status"` otherwise
  (`aria-atomic`).
- ERROR-003 — Field-level errors (login credentials, registration form,
  submission URL) render inline with `role="alert"`, `aria-invalid` on the
  input and `aria-describedby` linkage — not only as toasts.
- ERROR-004 — A 401 triggers the contextual login redirect (AUTH-007); a 403
  with a lock code updates lock state (Section 4); any other 4xx/5xx shows
  the server message as a toast without navigation.
- ERROR-005 — Pages must reach a terminal loading state on fetch failure
  (loading indicators clear, empty/error states appear); an open page must
  not spin forever.
- ERROR-006 — Edge-level errors from Nginx (413, 429, 503) arrive in the
  same error envelope (`docs/contracts/http-api.md` §2) and are handled by
  the same ERROR-001 path.

### ACCESSIBILITY

- ACCESSIBILITY-001 — Every page has a skip link ("К содержанию") targeting
  `#main-content`, visible on focus; `<main id="main-content">` wraps page
  content; documents declare `lang="ru"`.
- ACCESSIBILITY-002 — Keyboard focus is always visible (`:focus-visible`
  outline with offset and halo) and never removed.
- ACCESSIBILITY-003 — Modals are true modal dialogs: `role="dialog"`,
  `aria-modal="true"`, labelled by their heading; triggers use
  `aria-haspopup="dialog"`, `aria-controls` and `aria-expanded`.
- ACCESSIBILITY-004 — While a dialog or the mobile menu is open, focus is
  trapped inside it: Tab/Shift+Tab cycle within the scope, background
  siblings are made `inert` (except live regions/alerts/toast stack), focus
  moves to a designated initial element (`data-focus-initial`) or the first
  focusable element, and focus returns to the original trigger on close.
- ACCESSIBILITY-005 — Escape closes any open dialog and the mobile menu.
  Click-outside closes the team-entry and transfer dialogs; the destructive
  disband dialog does **not** close on outside click (explicit choice
  required) — see ambiguities.
- ACCESSIBILITY-006 — Dynamic state changes are announced: loading
  indicators use `role="status"` + `aria-live="polite"` with accessible
  names; pages expose `aria-busy` while loading; lock notices use
  `role="status"`; the no-team instruction is `aria-live="polite"`.
- ACCESSIBILITY-007 — Content hidden before the interactive layer initializes
  (`x-cloak`) is not rendered visible prematurely; loading skeletons are
  announced instead of raw empty markup.
- ACCESSIBILITY-008 — Icon-only/decorative elements are `aria-hidden`;
  interactive icons have accessible names (close buttons, menu toggle with
  dynamic open/close label, scroll cue).
- ACCESSIBILITY-009 — Reduced motion is fully honored (RESPONSIVE-004) — it
  is an accessibility requirement, not a visual preference.
- ACCESSIBILITY-010 — Form fields have persistent visible `<label for>` or
  fieldset/legend structure (transfer dialog radio group); help and error
  texts are programmatically associated via `aria-describedby`.

### RESPONSIVE

- RESPONSIVE-001 — At viewports ≤ 64rem the desktop nav links and header
  login button are replaced by the mobile menu (`<details>`/`<summary>`
  disclosure); the menu closes on link activation and Escape, and while open
  the page scroll is locked (scroll lock applies at this breakpoint).
- RESPONSIVE-002 — Opening any modal locks page scroll (document-level
  overflow lock plus smooth-scroll suspension); closing restores it. Nested
  scroll areas (modal panels, the mobile menu panel) keep native scrolling
  and are excluded from the smooth-scroll hijack.
- RESPONSIVE-003 — Hover-dependent effects (magnetic buttons, hero tilt)
  initialize only on devices with `(hover: hover) and (pointer: fine)`;
  touch/coarse-pointer devices get the same functionality without them.
  Parallax intensity is reduced on compact viewports (< 900px).
- RESPONSIVE-004 — Under `prefers-reduced-motion: reduce`: smooth scrolling
  is disabled, scroll-driven and intro animations do not run, CSS
  animations/transitions are near-instant; the preference is observed live
  (changing it at runtime toggles behavior without reload).
- RESPONSIVE-005 — Layout has dedicated treatments at ≤ 64rem and ≤ 48rem
  breakpoints (including wordmark simplification); pages remain functional
  and readable down to small mobile widths.
- RESPONSIVE-006 — Header gains a solid/blurred background state after
  ~24 px of scroll; purely visual but must not break pointer interaction
  with nav controls.

### LIFECYCLE

- LIFECYCLE-001 — Page components clean up on teardown: countdown/clock
  intervals, lock-event listeners and animation/focus observers are removed
  (`destroy` paths); re-initialization must not double-register (pinned by
  `web/handler_test.go` for template-level init).
- LIFECYCLE-002 — The page is bfcache-safe: on `pagehide` with
  `persisted=true` (bfcache entry) live state is kept so the page resumes
  intact; on a real unload all global listeners, observers and tickers are
  torn down.
- LIFECYCLE-003 — The deadline countdown and the submission-deadline clock
  tick without page reload (1 s and 30 s respectively) and stop cleanly at
  their terminal states.
- LIFECYCLE-004 — Lock-state changes arriving while a page is open apply
  without reload (Section 4); data mutations either refresh in place or
  reload — the user never sees stale team/score state after a completed
  action.

## 6. Not automatically preserved

Implementation-specific behaviors that exist today but are **not** parity
requirements (the new frontend may replace them with equivalents):

- Alpine.js (`x-data`/`x-show`/`x-cloak`), GSAP/ScrollTrigger and Lenis as
  technologies; the target stack (`architecture.md`) defines replacements.
- The `spcase:lock` / `spcase:scroll-lock` CustomEvent mechanism — the
  requirement is the propagated outcome, not the bus.
- `window.location.reload()` after leave/transfer/disband and full
  `window.location.assign` navigations — a SPA should update state instead.
- Native `window.confirm()` dialogs for leave/kick — the destructive-warning
  semantics (TEAM-007) must stay, the widget may change.
- Exact timings: toast 5 s auto-dismiss, "copied" 2 s reset, 30 s deadline
  polling, 1 s countdown tick.
- Duplicate client-side validation rules (email "@", URL regex, score
  range) — keep client validation, but the server is authoritative and the
  exact regexes are not contractual.
- Generic criterion labels "Критерий 1..6" — a content gap, not mandated
  copy (see ambiguities).
- Visual/motion design itself: grain overlay, hero orbit/parallax, magnetic
  buttons, scroll-cue animation, reveal animations, badge color classes.
- Lenis smooth scrolling as a feature — the target architecture makes Lenis
  optional; scroll-lock outcomes (RESPONSIVE-002) remain required.
- The `?v=<hash>` cache-busting scheme and immutable static caching as an
  asset-delivery mechanism (the new frontend gets its own build hashing);
  the *no-store HTML* requirement (PUBLIC-009) stands.
- Plain Go `404 page not found` body for unknown routes — the new frontend
  should provide a proper 404 experience.

## 7. Open ambiguities

Behavior the code exhibits but no test or document establishes as intended:

1. **Disband dialog has no click-outside close** (team-entry and transfer
   dialogs have it; `dashboard.html:206` lacks `@click.outside`). Possibly
   deliberate for a destructive action; unverified.
2. **No client-side role guard on `/jury/teams` and `/admin`.** A USER or
   JURY opening them gets an API 403 toast and an empty/skeleton page rather
   than a redirect to their own workspace; only `/dashboard` redirects by
   role. Asymmetry may be accidental.
3. **`is_registration_open` / `is_submission_open` from `GET /info` are
   unused by the frontend.** Registration/submission forms stay submittable
   after the respective deadlines and rely on server rejection; the UI
   disables submission by computing the deadline locally instead. Intended
   contract for these flags is unclear.
4. **Post-mutation refresh is inconsistent**: kick → in-place reload of
   `/team/my`; leave/transfer/disband → full page reload; submit → in-place
   patch. No test pins any of these.
5. **Client clock drives deadline UI** (countdown, submission lock) while
   the server enforces by PostgreSQL time; skew shows stale enabled/disabled
   states until the next API rejection. No mitigation exists or is
   documented.
6. **Criteria have no names.** All six are labelled "Критерий N" in the jury
   UI; neither API nor docs carry criterion titles. Whether real criterion
   names are a product requirement is undecided.
7. **Anonymous visitors can open `/dashboard`, `/jury/teams`, `/admin`**
   (200 HTML) and are bounced only after the first API 401 — a brief
   unauthenticated render of workspace shells is accepted today.
8. **`/register` and `/login` are not hidden from authenticated users**;
   re-registering while logged in overwrites the session cookie. No guard or
   document defines intent.
9. **Clipboard copy has no fallback** for insecure contexts/older browsers —
   only an error toast. Whether a manual-copy fallback is required is
   undecided.
10. **Score inputs default to 0**, and an all-zeros batch is a valid
    evaluation (0 is a legal score). A jury can mark a team "evaluated" by
    saving untouched zeros; no explicit "unset" state exists.
