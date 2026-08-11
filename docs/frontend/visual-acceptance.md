# Frontend visual acceptance

> **Purpose:** current human evaluation of the React frontend, separate from
> the stable design constitution (`design-direction.md`) and product UX model
> (`experience-model.md`). This file is intentionally transient and is written
> only after human review.
>
> A technically valid implementation is not automatically visually accepted.

## Current status

**Stage 4E human review completed 2026-08-11 — verdict: REJECT DIRECTION**

Stage 4E was technically coherent and faithfully implemented the previous
brief, but the rendered result exposed a deeper design problem. It moved away
from generic SaaS and into another recognizable AI/template grammar:
"Swiss/editorial creative-agency site".

The implementation should not be incrementally polished from that direction.
The next approved step is **Stage 4F — dark de-stylization**.

## Why Stage 4E was rejected

The reviewed homepage/format composition over-relied on a set of visible
"authored" devices:

- oversized `01 / 02 / 03` numerals;
- `01 · ФОРМАТ`-style editorial labels;
- deliberate asymmetric placement;
- thin rules used as graphic signatures;
- flat gear/flag/machine illustration;
- deep navy + coral + mustard/cyan identity fields;
- a large coral final-stage band;
- deliberate grid breaks/crops;
- sparse, highly polished agency-site whitespace;
- slogan-like copy such as `Три этапа. Одна сильная работа.`.

Individually these techniques are legitimate. Together they read as a
recognizable generated editorial concept rather than an inevitable interface
for a real championship.

The central lesson is:

> Do not design "human imperfection" as a visible feature.

If the irregularities can be enumerated as a list of signature tricks, they
have become styling rather than character.

## What remains valid

Keep the engineering foundation:

- React/Vite/TypeScript architecture;
- API/query/auth behavior;
- accessibility and responsive mechanisms;
- `PublicStatus` behavior;
- React Router/View Transition foundation where it remains useful;
- Motion as an available dependency, but with much lower use;
- schedule as an information-first temporal surface;
- the principle that USER/JURY/ADMIN must not inherit generic SaaS dashboard
  grammar.

Visual ingredients are no longer protected merely because Stage 4D/4E used
them.

## Retired visual assumptions

The following are no longer part of the approved visual DNA by default:

- cream/off-white canvas as the main site background;
- four-color mustard/cyan/coral/navy identity;
- flat cartoon/editorial illustration system;
- gears, flags, podiums, sheet stacks and machines as recurring motifs;
- oversized stage numbers as identity;
- designed grid escapes and broken rules;
- "controlled imperfection" as a quota or checklist;
- large colored fields as section rhythm;
- public-page spectacle as a requirement.

These may not be retained merely for continuity with earlier stages.

## Stage 4F approved direction

The new north star is:

**Dark, content-first competition interface.**

Governing rules:

- nothing exists solely to make the page look designed;
- delete before designing;
- content, hierarchy, state and interaction are the primary visual material;
- default canvas is near-black/deep navy;
- warm light text, subtle rules and one deep-red accent dominate the palette;
- secondary colors are rare and muted;
- no replacement illustration system;
- no manufactured imperfection;
- no slogan copy where direct factual language is better;
- most public surfaces live on one continuous dark canvas;
- auth becomes especially quiet and functional;
- schedule derives visual interest from the schedule itself;
- product surfaces remain operational and information-first;
- motion is largely limited to navigation, state and direct interaction.

Detailed rules live in `design-direction.md` and `experience-model.md`.

## Stage 4F anti-AI acceptance test

During human review, reject a surface if it primarily looks like any of the
following:

- Swiss/editorial portfolio concept;
- dark SaaS with nested cards;
- cyberpunk/devtool/terminal UI;
- luxury-black marketing site;
- giant-number editorial layout;
- red/blue/neon glass interface;
- illustration-led landing page;
- deliberate "look how asymmetric this is" composition;
- typography slogans replacing factual information;
- generic motion showcase.

For any questionable element ask:

> Can its existence be justified without saying "visual interest"?

If not, it should normally be removed.

## Stage 4F human acceptance surfaces

Review at minimum:

- `/` full scroll at desktop (~1440×900);
- `/` at 375 px;
- `/` at 320 px;
- `/schedule` desktop + mobile;
- `/login` and `/register` desktop + mobile;
- navigation/menu behavior;
- reduced-motion behavior;
- high-contrast/focus states;
- evidence that old illustration assets/color-field grammar no longer dominate
  public composition.

Specific review questions:

1. Does the site feel like a real championship interface rather than a design
   concept?
2. Does the dark palette feel institutional/calm rather than cyber/luxury?
3. Is the page strong when all decorative artwork is ignored?
4. Is empty space allowed to remain empty?
5. Are typography and real information doing most of the visual work?
6. Does any section contain an element whose only clear purpose is aesthetic
   signaling?
7. Are auth and schedule more functional than theatrical?
8. Is motion sparse enough that a static screenshot still carries the design?

## Documentation ownership

Coding agents do not declare visual acceptance and do not edit this file.
Human review + ChatGPT Web own this document unless the user explicitly changes
that policy.

## Gate

Phase 5 is blocked until Stage 4F receives an explicit human **ACCEPT** in this
file. A technically valid commit is not enough.
