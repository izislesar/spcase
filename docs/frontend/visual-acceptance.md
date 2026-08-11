# Frontend visual acceptance

> **Purpose:** current human evaluation of the React frontend, separate from
> the stable design constitution (`design-direction.md`) and product UX model
> (`experience-model.md`). This file is intentionally transient and should be
> updated after each meaningful human visual review.
>
> A technically valid implementation is not automatically a visually accepted
> one.

## Current status

**Stage 4D human review completed 2026-08-11 — verdict: ITERATE**

Stage 4D is technically valid and established a useful motion/interaction
foundation. Human review concluded that the frontend has a recognizable and
promising visual DNA, but its composition still reads too often as a polished
premium/agency landing page rather than a product-specific competition
system.

The problem is not insufficient polish. Further decorative polish would likely
make the result worse. The next iteration must shift effort from effects to
composition, semantics, lifecycle UX and controlled imperfection.

## Keep

- React/Vite/TypeScript frontend foundation;
- API/query/auth architecture;
- Manrope display typography;
- off-white / navy / mustard / cyan-turquoise / coral-red palette;
- large editorial typography;
- flat SVG/DOM/CSS-first visual language;
- existing accessibility, mobile and reduced-motion mechanisms;
- desktop bottom-navigation concept where usability remains sound;
- route/view-transition technical foundation;
- `PublicStatus` concept;
- schedule as a timeline/information surface.

## Keep, but change how it is used

### Illustration

Keep the flat editorial illustration language, but stop using artwork as a
solution to empty space. Motifs should increasingly carry product/event
semantics and illustration sets must remain heterogeneous.

### Motion

Keep Motion and the working reduced-motion/view-transition infrastructure, but
reduce decorative motion density. Motion must support navigation continuity,
state, information progression or interaction feedback. Public hero spectacle
is the exception, not the default.

### Large color fields

Keep them, but use them to express hierarchy/state/rhythm rather than a
mechanical sequence of visually balanced sections.

## Rework

- homepage composition as a conventional hero/feature-section sequence;
- feature/stage areas that still resolve into equal or near-equal cards;
- excessively uniform polish and spacing;
- illustration placement that is decorative rather than semantic;
- public-page motion that competes with content;
- any assumption that public landing-page grammar should become the USER/JURY
  workspace grammar;
- schedule animation if it exists primarily as scroll spectacle rather than
  temporal orientation.

## New approved direction for Stage 4E

The stable north star is now:

**Editorial Competition OS with controlled imperfection.**

Stage 4E must consolidate the system around:

- roughly 90% discipline / 10% deliberate disobedience;
- authored asymmetry, uneven whitespace, crop/grid escape and oversized
  identifiers in selected high-salience areas;
- no fake handmade texture or random imperfection;
- homepage as a live competition cover/status board where current APIs support
  it;
- a semantic graphic language rather than decorative filler;
- explicit separation of expressive PUBLIC surfaces from document-like
  USER/JURY surfaces and utilitarian ADMIN surfaces;
- competition lifecycle orientation;
- reduced dependence on cards/tiles as composition;
- reduced decorative motion.

The detailed rules live in `design-direction.md` and `experience-model.md`.

## Anti-slop acceptance test

During review, explicitly reject a surface if its primary identity could be
summarized as one of these generic patterns:

- hero + equal feature cards + colored CTA block;
- Bento grid as the main composition language;
- KPI-card dashboard;
- universal rounded-container UI;
- icon + title + paragraph repeated as a section system;
- generic "Welcome back" product shell;
- gradients/glass/glow/abstract 3D as identity;
- continuous fade/reveal/parallax choreography;
- decorative irregularity added merely to look handmade.

A primitive is not rejected in isolation. The test concerns the dominant
visual grammar.

## Do not touch during Stage 4E art-direction consolidation

Unless the task explicitly requires otherwise, do not change:

- backend behavior;
- SQL/migrations;
- API contracts;
- nginx/Docker/cutover topology;
- legacy `web/` behavior;
- auth/query semantics;
- accessibility invariants.

## Stage 4E human acceptance surfaces

Review at minimum:

- homepage full scroll at desktop (~1440×900);
- `/schedule` desktop;
- homepage at 375 px;
- homepage at 320 px;
- navigation/menu behavior;
- public/auth-route visual continuity;
- reduced-motion state;
- evidence that the public composition no longer defaults to equal-card/
  agency-landing-page grammar.

Product workflow implementation remains Phase 5. Before coding those flows,
`experience-model.md` and `design-direction.md` are the authoritative UX/design
context.

## Gate

Phase 5 is blocked until a human review records Stage 4E **ACCEPT** here.
Technical validation or a commit alone does not satisfy this gate.
