# Frontend visual acceptance

> **Purpose:** current human evaluation of the React frontend, separate from
> the stable design constitution (`design-direction.md`) and product UX model
> (`experience-model.md`). This file is intentionally transient and is written
> only after human review.
>
> A technically valid implementation is not automatically visually accepted.

## Current status

**Stage 4G human review completed 2026-08-11 — verdict: DIRECTION ACCEPTED / COMPOSITION INCOMPLETE**

Stage 4G (`db9753a`) successfully added a restrained pseudo-3D/material identity
on top of the Stage 4F dark foundation without restoring the rejected
illustration/poster grammar. Human review likes the concept and wants to keep
it.

The remaining problem is not the visual language. It is **scale, spatial
integration and density**. The interface still feels raw/tentative on wide
screens and leaves too much dead space.

The next approved implementation step is **Stage 4H — spatial composition &
density**.

## What Stage 4G got right

Keep these decisions:

- dark near-black/deep-navy canvas;
- warm light text and restrained deep-red accent;
- no decorative illustration scene system;
- one Z2 hero spatial moment rather than site-wide 3D;
- graphite material planes with restrained lighting/shadows;
- fine-pointer tilt only as progressive enhancement;
- Format as the only current Z1 public progression surface;
- Schedule, FAQ, auth and navigation largely Z0;
- no Three.js/WebGL/new visual dependency;
- no global reveal choreography or idle animation;
- no fabricated LIVE/NOW/count/state data.

Do not undo these decisions to make the site feel fuller.

## Why Stage 4G is not yet visually accepted

Human review of the rendered Stage 4G frontend found:

- the hero artifact is visually too small/tentative relative to the desktop
  viewport;
- its layered construction can read as a stack of dark cards rather than one
  SPCase-specific physical/system object;
- the conventional text-left/object-right relationship still leaves large areas
  of dead space rather than meaningful negative space;
- Format technically uses Z-depth, but visually still reads as three dark
  bounded panels and the progression relationship is weak;
- overall desktop content footprint is too cautious in places;
- Schedule preview/full schedule can use typography, aligned time/data and width
  more confidently;
- FAQ is too visually small for the available canvas and can become a stronger
  full-width typographic surface without adding decoration.

The phrase "fill the empty space" in this review does **not** authorize filler.
It means the existing important content and spatial systems should occupy and
relate to the canvas more confidently.

## Stage 4H approved direction

Stage 4H is a **composition pass, not a new art direction**.

Governing rule:

> **Fill space with scale, relationships and depth — not with more components.**

Keep the Dark Spatial Competition Interface north star.

The intended changes are:

- make the hero composition materially larger and more confident on wide
  desktop;
- evolve the hero from a visible layered-card stack into one integrated spatial
  chassis/assembly with shared silhouette/geometry;
- allow hero typography more scale than ordinary sections;
- introduce a meaningful large backplane/supporting depth relationship if it
  belongs to the same hero object;
- make Format read as one connected shallow spatial progression, not three
  separate cards/panels;
- let Schedule gain visual weight through real data, wider alignment, stronger
  time/date hierarchy and horizontal footprint;
- let FAQ use more width as a typographic surface;
- keep auth/forms quiet and mostly 2D;
- preserve the restrained motion budget.

## What Stage 4H must NOT do

Do not solve density by adding:

- more cards;
- a second hero illustration;
- decorative SVG/geometry;
- abstract cubes/blobs;
- fake industrial/HUD labels;
- more colors;
- more metadata than the product needs;
- invented facts/state;
- new section-reveal choreography;
- WebGL/Three.js/R3F;
- a new component/rendering library.

Do not restore Stage 4E giant numerals, flags, gears, colored bands, poster
asymmetry or Bento grammar.

## Composition acceptance questions

1. Does the hero command the desktop viewport without feeling like a generic
   3D landing-page object?
2. Does the spatial object read as one integrated SPCase object rather than
   several floating cards?
3. Is empty space now intentional negative space with clear relationships rather
   than unused canvas?
4. Is hero typography strong enough without making every section a poster?
5. Does Format visually read as one progression before the user notices the
   individual stages?
6. Did Schedule become visually richer through information composition rather
   than cards/decoration?
7. Does FAQ use the canvas confidently while remaining a simple accordion?
8. Are flat Z0 surfaces still a useful contrast to the hero/Format spatial
   moments?
9. Does mobile flatten/recompose cleanly without inheriting desktop-scale gaps?
10. Did any new element appear solely because a region looked empty?

## Stage 4H human acceptance surfaces

Review at minimum:

- `/` full scroll at wide desktop (~1440×900 and preferably ~1920 wide);
- hero pointer interaction with a fine pointer;
- `/` at 375 px;
- `/` at 320 px;
- `/schedule` desktop + mobile;
- `/login` and `/register` desktop + mobile;
- navigation/menu behavior;
- reduced-motion behavior;
- keyboard focus and contrast.

## Documentation ownership

Coding agents do not declare visual acceptance and do not edit this file.
Human review + ChatGPT Web own this document unless the user explicitly changes
that policy.

## Gate

Phase 5 is blocked until Stage 4H receives an explicit human **ACCEPT** in this
file. A technically valid Stage 4H commit is not enough.
