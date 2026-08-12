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

The product should feel like a real event with a strong digital interface, not
like an editorial concept laid on top of CRUD and not like a generic dark
application shell.

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
authoritative data and contracts; never invent server state to make the UI look
more live.

## Information hierarchy first, spatial hierarchy second

Every important surface begins with the information needed to act:

- identity: team, case, participant, event;
- state: current lifecycle condition;
- time: date, deadline, remaining window where authoritative;
- material: case files, submission files, criteria, results;
- people: team members, jury context, roles;
- action: the current next step;
- consequence: what a high-salience action changes or locks.

Spatial depth may reinforce these relationships, but it must not replace them.
A user should understand the page even if all transforms are flattened.

Do not invent a dossier, ticket, stamp, scoreboard, machine or technical
metaphor when plain information hierarchy communicates the state better.

## Composition carries experience

The user should feel the hierarchy of the championship before interacting with
anything. Visual footprint should track product importance.

The accepted public composition uses this rule:

> Fill space with scale, relationships and depth — not with more components.

Practical consequences:

- the public hero may be physically/typographically large because it establishes
  event identity;
- Format should communicate one progression rather than three unrelated boxes;
- Schedule and FAQ may occupy wide desktop surfaces because their real data/text
  can carry the composition;
- auth and ordinary controls stay compact when their task is compact;
- empty space is allowed when it frames hierarchy, but not when it merely reveals
  an underscaled or disconnected composition.

Do not manufacture content, labels or product states to increase density. Use
real information more confidently first.

## Spatial interaction model

Depth is a progressive enhancement with three roles:

- **identity:** a rare memorable public spatial object may make SPCase
  recognizable;
- **hierarchy:** a meaningful object may advance/recede to distinguish active,
  selected or supporting information;
- **transition:** a real state/context change may be expressed as a restrained
  physical relationship.

Depth must not become a parallel navigation system or hide information.

Pointer tilt, hover depth and similar responses are optional enhancements for
fine-pointer users. Keyboard, touch and reduced-motion users receive the same
information and actions without needing the effect.

## Homepage

The homepage is the public entry point to the current championship.

Its first job is still to answer basic questions truthfully:

- what is SPCase;
- where/when it happens;
- whether/when registration is available;
- who can participate / team format where known;
- what the primary next action is.

The homepage may use one signature spatial composition to make the championship
feel tangible and memorable. The accepted public hero makes that composition large and
integrated enough to participate in the whole hero, not read as a small card
stack beside the content. It should carry or organize real event information
rather than act as an unrelated illustration.

When backend data supports a current lifecycle state, surface it. When it does
not, use stable factual information rather than fake LIVE/NOW/count concepts.

## Format

Explain how the championship works in direct language.

Each stage communicates:

- stage name;
- date/window when authoritative;
- what participants do;
- what outcome moves them forward.

The stage sequence may use shallow spatial progression when that makes the
relationship clearer or more engaging. On desktop, prefer one connected
progression over several independent stage panels. It must still read as a
simple vertical sequence on touch/mobile/reduced contexts.

Do not turn stage numbers into primary visual content and do not make the user
manipulate a 3D object to read the process.

## Schedule

Schedule is a primary information surface.

Prioritize:

- date grouping;
- times;
- event names;
- locations/details;
- current/next event when authoritative;
- fast scanning on desktop and mobile.

Schedule data itself provides most of the visual structure. Use more of the
available desktop width for times, date groups and event hierarchy before
adding any new visual object. Shallow depth may clarify day context or
selection, but the schedule must never become a 3D carousel or require
interaction to reveal ordinary events.

## Auth

Login and registration should feel connected to SPCase while remaining plain
functional forms.

Prioritize:

- clear context/title;
- concise registration/deadline information where truthful;
- stable fields;
- validation and errors;
- one obvious primary action;
- obvious link between login and registration where behavior permits.

A restrained background/spatial identity element is optional. It must never
compete with the form or become necessary for understanding the page.

## Participant workspace

The main participant surface is a **competition work area**, not a generic
dashboard. `/dashboard` is the participant entry point and should answer, in
order:

1. what state/context the participant is in now;
2. what team they belong to (or that no team exists);
3. what the most meaningful next action is;
4. which deadline/lock condition matters now when authoritative;
5. where to continue into team/case/submission work.

The authenticated shell owns page bounds, product navigation and responsive
workspace structure. No authenticated page-level content may render directly
against the viewport edge because a route forgot its own margin/padding.

Preferred hierarchy:

- team/participant identity;
- current competition state;
- next action;
- deadline/lock state;
- team membership;
- assigned/available case;
- work/submission artifacts as later phases expose them.

A participant with no team receives a deliberate no-team product state with
truthful valid actions, not a broken/empty dashboard and not an illustrated
empty-state card.

Avoid `Welcome back`, KPI cards, analytics, charts, marketing hero content or a
generic sidebar. Most workspace content is Z0. Shallow Z1 depth is permitted
only when one coherent work surface benefits from it.

Detailed Phase 5A rules live in `participant-workspace.md`.

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
Visual depth must not imply reversibility where none exists.

## Case

A case is the central work artifact.

After authoritative release, show:

- case title/context;
- instructions;
- files/materials;
- relevant dates/deadlines;
- current participant work state.

Before release, show only states actually supported by authoritative data.

A case may eventually be treated as a spatially foregrounded work artifact, but
not as a decorative sealed envelope or fake physical prop.

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

A real transition from editable to submitted/locked may use restrained depth or
material-state change, but the information and consequence must remain explicit
in text/state. Do not imply irreversibility unless the contract actually makes
the action irreversible.

## Jury

JURY is a scoring workspace, not a dashboard and not a themed judging prop.

Prioritize:

- team/case identity;
- submission material;
- scoring criteria;
- current values and totals where authoritative;
- validation/lifecycle state;
- score submission/finalization.

Repeated judging work requires low motion and high scanability. Spatial effects
should be rare and state-oriented.

## Results

Results may carry stronger hierarchy because they are a meaningful event
moment, but data stays primary.

Prioritize:

- ranking/outcome;
- score;
- team identity;
- scoring provenance where available.

A result reveal may eventually use stronger depth/motion than routine product
work, but comparison must remain clear and accessible.

## Persistent lifecycle orientation

Where useful, a compact lifecycle/product navigation rail may orient participants
across routes. Structure/navigation may exist before the backend exposes enough
truth to mark stages complete, but completion/current-state semantics must not
be fabricated.

Do not create fake continuous percentage progress for a stage-based lifecycle.
On mobile, simplify the navigation rather than preserving a desktop rail at all
costs.

## Interaction hierarchy

At a given state, prefer one obvious primary action over several equally loud
CTAs. Secondary actions should remain secondary.

High-consequence actions require clearer consequence communication and
confirmation than routine navigation.

Spatial response must follow interaction importance rather than compete with
it. A decorative object should never move more prominently than the action the
user is trying to complete.

## Public / product / admin density

- PUBLIC: stronger visual identity is allowed; one or a few spatial moments may
  be memorable, while factual content remains primary. Wide-screen composition
  should use scale and relationships confidently rather than adding filler.
- USER/JURY: denser operational information; depth is mainly state/hierarchy.
- ADMIN: utilitarian, highly scannable, predominantly 2D.

Shared identity comes from dark material language, typography, restrained red,
interaction and product semantics. It does not require every surface to use
pseudo-3D.

## Accessibility and truthfulness

- Do not encode state only in color, depth or motion.
- Reduced-motion users receive equivalent information.
- Never invent deadlines, lock states, IDs, scores, eligibility or lifecycle
  transitions from visual design needs.
- Do not trade form/control predictability for visual character.
- Pointer-driven spatial response is progressive enhancement only.
- Important text must remain readable; avoid perspective distortion on body
  copy and controls.
- When visual intent conflicts with an authoritative behavioral contract,
  preserve the contract and report the design constraint/gap.
