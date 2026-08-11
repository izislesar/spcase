# Frontend visual acceptance

> **Purpose:** current human evaluation of the React frontend, separate from
> the stable design constitution (`design-direction.md`) and product UX model
> (`experience-model.md`). This file is intentionally transient and is written
> only after human review.
>
> A technically valid implementation is not automatically visually accepted.

## Current status

**Stage 4F human review completed 2026-08-11 — verdict: FOUNDATION ACCEPTED / VISUAL INCOMPLETE**

Stage 4F (`4f6cbb2`) successfully removed the Stage 4E AI/editorial cliches and
produced a technically clean dark content-first baseline. Human review agreed
with the deletion direction but found the rendered public frontend too flat,
sparse and visually close to a polished wireframe.

Stage 4F is therefore **not rejected**. Its cleanup is the foundation for the
next iteration. The next approved implementation step is **Stage 4G — spatial
identity without decorative illustration**.

## What Stage 4F got right

Keep these decisions:

- dark near-black/deep-navy canvas;
- warm light text and restrained deep-red accent;
- removal of mustard/cyan/coral section fields;
- removal of HeroScene/GearScene/PennantScene/PodiumScene/SheetStack/BubbleScene
  and other public illustration scenes;
- deletion of the Mosaic/Bento section;
- removal of giant `01/02/03` identity numerals;
- factual copy over slogans;
- much lower motion density;
- quiet auth forms;
- information-first schedule;
- no fabricated LIVE/NOW/count/state data;
- no generic dashboard/card system.

Do not undo this cleanup merely to make the site look fuller.

## Why Stage 4F is not visually complete

Human review of the rendered Stage 4F frontend found that the de-stylization
removed too much visual hierarchy together with the cliches.

Observed issues:

- large desktop areas feel unused rather than intentionally spacious;
- hero lacks a memorable visual/product identity moment;
- the public flow can read like a very clean wireframe;
- Format is informative but visually under-articulated;
- the dark material system has little depth;
- the interface needs stronger UX feedback and spatial hierarchy without
  returning to illustration or poster tricks.

The correct response is **not** to restore Stage 4E. The response is to add a
small, disciplined spatial system on top of the Stage 4F skeleton.

## Stage 4G approved direction

North star:

**Dark Spatial Competition Interface**

Stage 4G adds pseudo-3D and material depth using the existing frontend stack.
The governing rule is:

> Visual complexity is allowed when it creates spatial hierarchy, interaction,
> orientation, or a memorable product object.

The homepage may contain one signature spatial object/composition. Shallow
structural depth may also appear in Format and, where useful, Schedule.

FAQ and forms remain mostly flat and functional.

## Stage 4G depth limits

Use the depth model from `design-direction.md`:

- `Z0` — normal 2D interface;
- `Z1` — shallow structural depth;
- `Z2` — rare signature spatial object, primarily hero.

Do not add a site-wide 3D scene system.

Stage 4G should use CSS perspective/transforms, layered DOM/SVG where justified
and the already-installed Motion library. Do not add Three.js, React Three
Fiber, WebGL frameworks, shaders or another visual runtime dependency.

## Visual material rules

Allowed when purposeful:

- subtle material plane gradients;
- soft cast shadows between meaningful spatial layers;
- restrained edge highlights;
- small pointer tilt on a signature object;
- shallow foreground/background response on meaningful stage/day objects.

Still rejected:

- decorative illustration;
- floating abstract blobs/cubes;
- glow/neon;
- glassmorphism;
- cyberpunk/HUD/terminal language;
- rotating logos;
- decorative mechanical/industrial details;
- ubiquitous floating cards;
- constant parallax;
- large scroll-driven spectacle;
- giant numbers and Stage 4E poster grammar.

## Stage 4G human acceptance surfaces

Review at minimum:

- `/` full scroll at desktop (~1440×900);
- hero pointer interaction with a fine pointer;
- `/` at 375 px;
- `/` at 320 px;
- `/schedule` desktop + mobile;
- `/login` and `/register` desktop + mobile;
- navigation/menu behavior;
- reduced-motion behavior;
- keyboard focus and contrast.

## Stage 4G review questions

1. Does the hero now have a memorable SPCase-specific visual presence without
   looking like decorative 3D artwork?
2. Does pseudo-3D carry information/hierarchy or merely fill empty space?
3. Is the dark interface richer without becoming cyberpunk, glassy or
   developer-tool-like?
4. Is Format more engaging while still readable as the actual competition
   process?
5. Is Schedule still primarily information, not a 3D interaction demo?
6. Are FAQ and auth restrained enough to create contrast with the signature
   spatial moments?
7. Does motion feel like direct physical response rather than choreography?
8. Does the interface remain complete and convincing under reduced motion?
9. Does mobile flatten gracefully rather than preserving desktop spectacle at
   the expense of content?
10. Has any old illustration/poster cliche returned under a different name?

## Documentation ownership

Coding agents do not declare visual acceptance and do not edit this file.
Human review + ChatGPT Web own this document unless the user explicitly changes
that policy.

## Gate

Phase 5 is blocked until Stage 4G receives an explicit human **ACCEPT** in this
file. A technically valid Stage 4G commit is not enough.
