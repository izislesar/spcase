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

**Stage 4B — TECHNICALLY ACCEPTED / VISUALLY ITERATE**

The Stage 4B implementation is committed and technically sound (behavior,
accessibility mechanisms, responsiveness machinery). Human visual review did
not accept the result; the next design step is Stage 4C.

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
