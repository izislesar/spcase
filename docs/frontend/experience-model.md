# Frontend experience model

> **Status: approved product UX constitution.**
> This document defines how the competition should be experienced through the
> frontend. It does not override backend/business/API contracts and does not
> prescribe exact mockups or component structure.

## Core proposition

SPCase is not a SaaS application that happens to host a competition.

The interface represents a **live competition with stages, deadlines,
artifacts, roles and consequential moments**. A user should quickly understand:

1. where they are in the competition;
2. what is available now;
3. what is locked or unavailable;
4. what happens next;
5. what action matters most.

The product should feel like a real event with a direct digital interface, not
like an editorial concept laid on top of a CRUD application.

Visual rules live in `design-direction.md`; behavioral truth remains in
`legacy-contract.md`, `docs/domain/business-rules.md` and
`docs/contracts/http-api.md`.

## Competition lifecycle

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

Not every state must exist as a new backend enum. Presentation is derived from
authoritative data and contracts; never invent server state to make the UI
look more "live".

## Information hierarchy before visual metaphor

Every important surface should begin with the information needed to act.
Useful categories are:

- identity: team, case, participant, event;
- state: current lifecycle condition;
- time: date, deadline, remaining window where authoritative;
- material: case files, submission files, criteria, results;
- people: team members, jury context, roles;
- action: the current next step;
- consequence: what a high-salience action changes or locks.

Do not invent a dossier, ticket, stamp, scoreboard or technical metaphor when
plain information hierarchy communicates the state better.

## Homepage

The homepage is the public entry point to the current championship.

Its first job is to answer basic questions with truthful content:

- what is SPCase;
- where/when it happens;
- whether/when registration is available;
- who can participate / team format where known;
- what the primary next action is.

When backend data supports a current lifecycle state, surface it. When it does
not, use stable factual event information rather than fake LIVE/NOW/countdown
concepts.

The homepage does not need an illustration or metaphor to feel complete.

## Format

Explain how the championship works in direct language.

Each stage should communicate:

- stage name;
- date/window when authoritative;
- what participants do;
- what outcome moves them forward.

Small sequence numbers are acceptable as metadata. Stage numbers should not
become the primary visual content.

## Schedule

Schedule is a primary information surface, not a decorative timeline story.

Prioritize:

- date grouping;
- times;
- event names;
- locations/details;
- current/next event when authoritative;
- fast scanning on desktop and mobile.

The schedule itself should provide the visual structure. Illustration and
narrative motion are unnecessary.

## Auth

Login and registration should feel like entering the championship, but remain
plain functional forms.

Prioritize:

- clear context/title;
- concise registration/deadline information where truthful;
- stable fields;
- validation and errors;
- one obvious primary action;
- obvious link between login and registration.

Do not add a decorative art panel merely to make auth visually distinctive.

## Participant workspace

The main participant surface is a work area, not a generic dashboard and not a
stylized dossier prop.

Preferred hierarchy:

- team identity;
- current competition state;
- assigned/available case;
- deadline/lock state;
- team membership;
- work/submission artifacts;
- next action.

Avoid `Welcome back`, KPI cards or marketing-style hero content. Use lists,
rules, metadata and bounded surfaces only when they improve comprehension.

## Team formation

Team formation is a roster-building workflow.

Emphasize:

- current members;
- captain/role where authoritative;
- available capacity;
- invite/join actions;
- restrictions that prevent further changes;
- transition into a locked/competition state.

Hard Lock and lifecycle behavior remain defined by backend/domain contracts.
Visual treatment must not imply reversibility where none exists.

## Case

A case is the central work artifact.

After authoritative release, show:

- case title/context;
- instructions;
- files/materials;
- relevant dates/deadlines;
- current participant work state.

Before release, show only states actually supported by authoritative data.

Do not needlessly style the case as a sealed envelope, issued dossier or other
physical metaphor. The actual case content is enough.

## Submission

Submission is a high-consequence workflow and deserves more interaction care
than routine navigation.

Make clear:

- required/optional artifacts;
- selected/uploaded files;
- validation/readiness;
- deadline;
- consequence of submitting/locking according to the real contract;
- final action;
- confirmed post-submit state.

Confirmation may be visually strong, but it should be based on real status,
time and files rather than a decorative stamp metaphor.

Do not imply irreversibility unless the contract actually makes the action
irreversible.

## Jury

JURY is a scoring workspace, not a dashboard and not a themed judging prop.

Prioritize:

- team/case identity;
- submission material;
- scoring criteria;
- current values and totals where authoritative;
- validation/lifecycle state;
- score submission/finalization.

Repeated judging work requires low motion and high scanability.

## Results

Results may carry stronger hierarchy because they are a meaningful event
moment, but data stays primary.

Prioritize:

- ranking/outcome;
- score;
- team identity;
- scoring provenance where available.

Do not add celebratory visual noise that obscures comparison.

## Persistent lifecycle orientation

Where useful, a compact lifecycle indicator may orient users across routes.
It must correspond to real state and remain understandable without animation.

Do not create fake continuous percentage progress for a stage-based lifecycle.

## Interaction hierarchy

At a given state, prefer one obvious primary action over several equally loud
CTAs. Secondary actions should remain secondary.

High-consequence actions (join/finalize/submit/evaluate where applicable)
require clearer consequence communication and confirmation than routine
navigation.

## Public / product / admin density

- PUBLIC: calm, direct, content-first, sparse when content is sparse.
- USER/JURY: denser operational information, minimal decoration.
- ADMIN: utilitarian, highly scannable, operationally safe.

Shared identity comes from type, palette, spacing, interaction and product
language; it does not require shared decorative motifs.

## Accessibility and truthfulness

- Do not encode state only in color or motion.
- Reduced-motion users receive equivalent information.
- Never invent deadlines, lock states, IDs, scores, eligibility or lifecycle
  transitions from visual design needs.
- Do not trade form/control predictability for visual character.
- When visual intent conflicts with an authoritative behavioral contract,
  preserve the contract and report the design constraint/gap.
