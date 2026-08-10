# Frontend visual acceptance

> **Purpose:** current human visual evaluation of the new frontend, kept
> separately from the stable design direction (`design-direction.md`).
> This document is transient but authoritative: it records what human review
> accepted, what it rejected and what the next visual stage must improve.
> Update it after every human visual review.
>
> **Acceptance principle:** a technically valid, committed implementation is
> NOT automatically a visually accepted one. Do not proceed to Stage 5 until
> a human review explicitly records visual **ACCEPT** here.

## Current status

**Stage 4D implemented — awaiting human visual review**

The Stage 4D implementation is committed and technically validated
(typecheck, Biome, production build, Playwright test discovery). It adds
the interaction/motion/UX-polish layer on top of the accepted Stage 4C
static baseline WITHOUT redesigning it: a shared editorial motion layer
(Motion, reduced-motion aware), route view transitions with named surface
morphs (schedule title, coral register field, turquoise login field),
the bottom-navigation traveling marker and partial scroll-direction
response, the authored hero entrance with layered scene assembly, pointer
and scroll depth, individualized format-stage choreography, per-surface
mosaic interactions, /schedule scroll storytelling and the orphan-letter
title fix, the PublicStatus editorial loading/error/empty component,
FAQ layout animation, footer and auth-shell motion, and CTA/link
microinteractions. Human visual review has NOT yet accepted the result;
the verdict remains **ITERATE** until a human records **ACCEPT** below.

## KEEP

- React/frontend technical foundation;
- API/query behavior;
- Manrope as the display face;
- current white/navy/mustard/turquoise/coral palette;
- desktop bottom-navigation concept;
- accessibility and mobile mechanisms;
- sparse native motion;
- current `/schedule` behavioral architecture.

## KEEP BUT AMPLIFY

- schedule storytelling concept;
- large typography;
- flat SVG approach;
- saturated color fields.

## REWORK HEAVILY

- hero scale and composition;
- excessive centered/safe `max-width` feeling;
- format/stages composition;
- footer artwork;
- schedule visual scale.

## REPLACE

- icon-pack-like illustration grammar;
- repeated circle + object + spark/plus formula;
- universal mosaic tile anatomy;
- three near-identical stage-card anatomy.

## ADD / TARGET FOR STAGE 4C

- wide/full-bleed public composition primitive;
- larger asymmetric editorial hero scene;
- heterogeneous illustrations/scenes;
- true editorial collage rather than Bento-like tiles;
- reduced card chrome/radius/shadows;
- visual shells for public auth routes where appropriate, without
  implementing Stage 5 business functionality;
- intentional 320/375 mobile recomposition.

## DO NOT TOUCH DURING VISUAL ITERATION

- backend;
- SQL;
- API contracts;
- nginx/Docker/cutover;
- legacy `web/`;
- query semantics;
- authentication semantics;
- accessibility invariants.

## Human acceptance surfaces

A human visual review covers at minimum:

- homepage full scroll at desktop, approximately 1440×900 viewport;
- `/schedule` desktop;
- homepage at 375 px;
- homepage at 320 px;
- navigation/menu behavior;
- public/auth-route visual continuity;
- reduced-motion state.

## Gate

Stage 5 is blocked until this document records an explicit human visual
**ACCEPT** for the surfaces above.
