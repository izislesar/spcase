# Frontend experience model

> **Status: approved product UX constitution.**
> This document defines how the competition should be experienced through the
> frontend. It does not override backend/business/API contracts and does not
> prescribe exact mockups or component structure.

## Core proposition

SPCase is not a SaaS application that happens to host a competition.

The interface represents a **live competition with stages, deadlines,
artifacts, roles and irreversible moments**. Product UX should make a user
understand where they are in that competition, what is available now, what is
locked, what happens next and what action matters most.

The conceptual blend is:

```text
editorial publication
× competition scoreboard
× operational dossier
× event system
× judging sheet
```

The visual constitution is `design-direction.md`; behavioral truth remains in
`legacy-contract.md`, `docs/domain/business-rules.md` and
`docs/contracts/http-api.md`.

## Competition lifecycle

The UX should consistently express the relevant lifecycle rather than treating
routes as unrelated pages.

Canonical participant progression:

```text
REGISTRATION
    ↓
TEAM FORMATION
    ↓
CASE RELEASE
    ↓
WORK
    ↓
SUBMISSION OPEN
    ↓
SUBMITTED / LOCKED
    ↓
JURY
    ↓
RESULTS
```

Not every state must exist as a new backend enum. The frontend derives its
presentation from existing authoritative data and contracts; do not invent
server state merely to match this diagram.

At any important participant surface, answer these questions quickly:

1. What stage is the competition in?
2. What is my/team state?
3. What can I do now?
4. What is the next irreversible or deadline-bound event?

## Stable product language

Use compact identifiers and metadata to make the system feel concrete and
operational rather than generic.

Examples of the language, subject to real available data:

```text
SPK/26
TEAM/042
CASE/03
DOC/02
REV/04
JURY/07
```

Do not fabricate IDs the backend does not expose. The principle is stable
naming and metadata hierarchy, not fake data.

Useful visual roles:

- display: event/stage/title;
- meta: IDs, timestamps, role, version, status;
- data: countdowns, scores, counts, deadlines;
- action: the single current next step;
- rule/marker: progression, grouping, current position.

## Homepage as live cover

The homepage should evolve from static marketing toward a **live cover/status
board** of the championship.

Its strongest message should depend on the real current stage when data is
available. Examples:

- registration open → deadline and team formation are primary;
- case active → active work window and case status are primary;
- submission window → deadline/state are primary;
- completed competition → results become primary.

This is a presentation rule, not authorization to invent missing API state.
If current APIs cannot support a desired state safely, document the gap rather
than guessing.

## Participant workspace: dossier, not dashboard

The main participant workspace should read like the team's active competition
dossier rather than a generic dashboard.

Preferred hierarchy:

- team identity;
- current stage;
- assigned/available case;
- deadline or lock state;
- team membership;
- work artifacts/submission state;
- next action.

Avoid starting the surface with generic KPI cards or "Welcome back" copy. Use
plain hierarchy, rules, lists, metadata and strong state communication before
introducing containers.

## Team formation

Team formation should feel like assembling a roster for an event, not filling
an account-settings page.

Emphasize:

- current membership;
- captain/role where authoritative;
- available capacity/invite or join actions;
- conditions that prevent further changes;
- transition into locked/competition state.

Hard Lock and related lifecycle behavior remain defined by backend/domain
contracts; visual treatment must not imply reversibility where none exists.

## Case release: issued dossier

A case is an issued competition artifact, not a generic downloadable card.

Before release, if authoritative data exposes the condition, the UI may use a
sealed/locked treatment. After release, present the case as an issued dossier
with title, metadata, files/instructions and current work state.

A release moment is allowed to have more visual ceremony than routine product
navigation because it is a meaningful lifecycle event.

## Submission: a deliberate ritual

Submission is one of the highest-salience interactions in the product. Treat
it like fixing/filing a competition result, not dropping a file into a generic
upload widget.

The surface should make clear:

- required and optional artifacts;
- validation/readiness state;
- deadline;
- consequences of submission/lock according to the actual contract;
- final action;
- confirmed post-submit state.

After successful submission, use a strong stable confirmation treatment
(e.g. issued timestamp/status/stamp language) while remaining accessible and
truthful to backend state.

Do not imply an irreversible action unless the contract really makes it
irreversible.

## Jury: judging desk

JURY surfaces are a judging workspace, not a dashboard.

The ideal desktop mental model is a judging sheet beside the team/submission
material. Prioritize:

- team/case identity;
- submission material;
- scoring criteria;
- current values and totals where authoritative;
- validation and lifecycle state;
- deliberate score submission/finalization.

Repeated judging work requires efficiency and low motion. Visual character
comes from hierarchy, typography, rules and state — not decorative scenes.

## Results

Results are an event moment and may receive stronger art direction than
routine product screens. The hierarchy should emphasize ranking/outcome and
make scoring provenance understandable where the product exposes it.

Do not turn results into celebratory visual noise that obscures data.

## Schedule and temporal state

Schedule is an information visualization of time, not a decorative scrolling
story.

If current-time data is available, make the user's temporal position legible:
what happened, what is current, what is next. Animation may reinforce current
progress but must not be required to understand the schedule.

## Persistent lifecycle orientation

Where useful, surfaces may share a compact lifecycle/stage indicator so users
can orient themselves across routes. It should describe real product state and
collapse gracefully on mobile.

Do not create a fake percentage progress bar when the underlying lifecycle is
stage-based rather than continuous.

## Semantic graphic language

Reusable motifs should gain stable meaning. Candidate roles include:

- current-stage marker;
- milestone/flag;
- issued case/document;
- submission/fixed state;
- locked state;
- jury mark/evaluation;
- result/rank;
- dotted or ruled progression path.

Do not force every motif onto every screen. A semantic motif is useful because
it can recur when the same concept recurs.

## Interaction hierarchy

At a given state, prefer one obvious primary action over several equally loud
CTAs. Secondary actions should visually remain secondary.

High-salience lifecycle actions (join/finalize/submit/evaluate where actually
applicable) deserve stronger confirmation and consequence communication than
routine navigation.

## Public / product / admin density

- Public: expressive, spacious, illustrative.
- USER/JURY product: information-dense, document-like, task-focused.
- ADMIN: utilitarian, highly scannable, operationally safe.

Shared identity does not require shared layout anatomy.

## Accessibility and truthfulness

- Never trade readability or control predictability for controlled
  imperfection.
- Do not encode state only in color or motion.
- Reduced-motion users must receive equivalent information.
- Never invent deadlines, lock states, IDs, scores, eligibility or lifecycle
  transitions from visual design needs.
- When design intent conflicts with an authoritative behavioral contract,
  preserve the contract and document the design constraint/gap.
