# Participant workspace UX contract

> **Status: approved Phase 5A UX constitution.**
> This document defines the participant authenticated shell and `/dashboard`
> experience. Backend/domain/API contracts remain authoritative for behavior and
> state. It does not authorize invented lifecycle data.

## Purpose

The participant area is a **competition workspace**, not a generic dashboard.
Its job is to orient the participant and surface the next meaningful action with
minimum friction.

The entry surface should answer:

1. Where am I in the product/competition context?
2. Do I have a team, and who is in it?
3. What can or must I do next?
4. Which deadline/lock condition matters now?
5. Where do I continue working?

Do not optimize `/dashboard` for the number of visible modules.

## Authenticated layout contract

All USER authenticated pages must render inside one reusable workspace shell (or
an equivalent shared layout boundary). The shell owns:

- main content bounds and top spacing;
- desktop/mobile gutters;
- product navigation;
- participant/team context when available;
- account/logout affordance where the existing product contract permits it;
- `main` landmark and stable page hierarchy;
- responsive transition between desktop and mobile navigation.

**Invariant:** no authenticated route may place page-level content directly at
viewport `(0,0)` because that route forgot local spacing. Fix shell/layout
architecture, never the individual page with arbitrary `margin-left`, `top` or
absolute positioning.

The shell should be small enough to support future `/team`, case and submission
surfaces without becoming a framework inside the application.

## Public navigation vs product navigation

Do not reuse public navigation mechanically.

Public navigation answers questions about the championship. Participant
navigation answers questions about the participant's work. They may share brand
tokens and motion primitives but have different information architecture.

Avoid a generic app sidebar with many icons. The current product has too few
participant destinations to justify that default. Prefer compact textual/task
navigation.

## Dashboard grammar

The dominant grammar is:

```text
one workspace
  ├─ current context/state
  ├─ team region
  ├─ deadline/lock region when authoritative
  └─ next action / continuation
```

Not:

```text
[team card] [deadline card] [status card]
[case card] [members card]  [stats card]
```

Do not create generic abstractions such as `DashboardCard`, `StatCard`, `Widget`
or `OverviewTile` for Phase 5A.

## State priority

At any given state, prefer one obvious primary next action. Secondary actions
should remain visibly secondary.

The UI may present, when authoritative:

- participant identity;
- team membership/name;
- captain/member role;
- team members/capacity;
- relevant deadline;
- mutation/lock state;
- availability of the next existing workflow route.

Do not invent current case, completion percentage, file counts, result state,
team metrics or lifecycle completion for visual density.

## No-team state

No-team is a first-class product state, not an error and not a decorative empty
state. It should:

- state plainly that the participant is not currently in a team;
- expose only valid create/join/invite-related actions supported by the existing
  routes/contracts;
- surface relevant team-size/lifecycle restrictions if authoritative and useful;
- avoid giant icons, illustrations or empty-state cards.

If creation/join is implemented on another route, the dashboard should route to
that workflow rather than duplicate it.

## Has-team state

A participant with a team should see the team as a working entity, not as one
widget among many. Use a coherent team/workspace region with:

- team identity;
- member roster/roles where authoritative;
- current relevant state;
- one primary continuation action;
- deadline/lock information where it affects action.

Prefer typography, alignment, rules and one shared material surface over nested
boxes.

## Lifecycle/product navigation

A compact participant rail/navigation may use concepts such as team, case, work
and submission, but exact route/state labels must match authoritative product
behavior.

Structural navigation is allowed even when the backend cannot determine stage
completion. In that case do **not** show fake completed/current markers.

Never use a fake continuous progress percentage for a stage-based workflow.

On mobile, simplify the rail into compact navigation; do not preserve a wide
desktop construct at the cost of clarity.

## Spatial depth

Participant UI is predominantly **Z0**.

A single **Z1** coherent work surface may be used when it clarifies:

- current work hierarchy;
- relationship between primary and supporting state;
- consequential state later in the workflow.

Do not use Z2 public-hero spectacle in routine authenticated work. Do not tilt
inputs, forms, member rows or navigation. No idle/pointer spectacle.

Depth must correspond to hierarchy/state; if flattening the panel changes
nothing except “coolness”, flatten it.

## Typography and density

Product typography is denser than the public hero. Use a clear workspace title,
section labels, readable body, metadata and tabular time/data where appropriate.

Do not use giant slogans or motivational copy.

Desktop should use a wide structural canvas while keeping prose/data measures
local. Mobile should reorder regions by task importance, not simply stack every
desktop column in source order if that produces a poor task flow.

## Color/material

Reuse the accepted dark material system:

- near-black/deep-navy canvas;
- warm foreground;
- muted neutral metadata;
- graphite material planes only when structurally meaningful;
- restrained deep red for primary/important state.

Do not add a new participant accent palette, glass, glow or card shadows.

## Motion

Workspace motion is quieter than the public hero. Valid reasons:

1. navigation state;
2. menu state;
3. direct control feedback;
4. meaningful product-state transition.

No page reveal choreography, card staggers, parallax, idle floating or dashboard
entrance animation.

## Loading/error/locked states

Participant state surfaces must distinguish pending, success/domain-empty, error
and locked/forbidden outcomes according to the API/domain contract.

Errors should explain what happened and what the participant can do next. Server
lock/deadline rejections remain authoritative; do not hide or reinterpret them
for a smoother visual flow.

## Responsive requirements

Validate at minimum:

- wide desktop / 1440×900;
- 375 px;
- 320 px.

Requirements:

- stable shell gutters;
- no viewport-corner leakage;
- no horizontal overflow;
- primary action remains obvious;
- team/member information stays scannable;
- deadline/state is not hidden;
- touch targets >= 44 px;
- depth flattens when useful;
- no hover-only functionality.

## Accessibility

Preserve:

- `main` and navigation landmarks;
- logical heading hierarchy;
- keyboard navigation/focus;
- WCAG contrast;
- reduced-motion equivalence;
- textual state in addition to color/depth;
- DOM reading order independent of visual transforms.

## Phase 5A scope

Implement now:

- authenticated participant shell/layout contract;
- participant product navigation foundation;
- `/dashboard` no-team and has-team states from real data;
- next-action hierarchy;
- responsive workspace behavior;
- only immediately required shared primitives.

Do not fully implement yet:

- complete team create/join/invite/edit workflow (Phase 5B);
- case/material workspace (Phase 5C);
- submission workflow (Phase 5D);
- JURY;
- ADMIN;
- results;
- speculative profile/settings/notification systems.

Existing future routes may remain incomplete but must participate in the shared
authenticated layout when architecturally appropriate.

## Acceptance test

Phase 5A is successful when a participant can enter the authenticated product
and immediately understand their truthful team/context and next available action
without the interface resembling a generic SaaS dashboard. The shell must make
the old viewport-corner defect structurally impossible.
