# Frontend design direction

> **Status: approved stable visual constitution.**
> This document describes enduring art-direction rules, not a snapshot of a
> particular implementation stage. Current human verdicts live in
> `visual-acceptance.md`; product/interaction semantics live in
> `experience-model.md`.

## North star

The SPCase frontend is an **Editorial Competition OS with controlled
imperfection**.

It should feel like a live competition system expressed through editorial
design: part publication, scoreboard, dossier, event graphics and judging
desk. It must not read as generic B2B SaaS, a component-library showcase or an
"award-site" motion demo.

The visual system is intentionally disciplined but not polished to sterility.
Use roughly:

**90% discipline / 10% disobedience.**

The discipline proves competence. The disobedience creates character.

## Controlled imperfection

Controlled imperfection is not sloppiness and not simulated handmade texture.
It is deliberate tension inside a strong system.

### Discipline

- clear information hierarchy;
- coherent grid and spacing logic;
- predictable interaction and control placement;
- accessible contrast, focus and target sizes;
- stable typography and product semantics;
- consistent state language;
- intentional responsive composition.

### Disobedience

Use sparingly and mainly at high-salience moments:

- asymmetric composition;
- uneven but purposeful whitespace;
- a display element escaping or crossing the main grid;
- deliberate crop at a viewport or section boundary;
- oversized identifiers or numerals;
- rare physical marks such as a stamp, registration mark or broken rule;
- deliberate changes in visual density between neighboring regions;
- a strong color field that is not balanced symmetrically by another one.

A violation must be authored. Random misalignment, arbitrary rotation and
inconsistent spacing are not character.

### Do not fake "human"

Do not add paper grain, scratched textures, distressed type, random rotations,
noise, fake print errors or hand-drawn wobble merely to make the interface
look less digital. Imperfection should come from composition and hierarchy,
not from decorative aging.

## Existing visual DNA to preserve

The Stage 4D review confirmed that the following direction remains valuable:

- very light / off-white editorial canvas;
- deep navy as primary ink;
- saturated flat mustard, cyan/turquoise and coral/red fields;
- Manrope as the primary display face;
- oversized typography and strong line breaks;
- flat SVG/DOM/CSS-first illustration;
- simple silhouettes and limited palettes;
- strong contrast between large graphic surfaces and minimal UI chrome;
- desktop bottom-navigation concept where it remains usable;
- mobile and reduced-motion mechanisms.

These are ingredients, not a frozen layout. Existing components may be
substantially recomposed when the human verdict requires it.

## Canvas and color

- Prefer white / very-light editorial canvas with deep navy ink.
- Use mustard, cyan/turquoise and coral/red as large flat fields or semantic
  accents, not as a mechanical "one color per section" sequence.
- Let color create rhythm and state emphasis.
- No gradients, glassmorphism or soft glow layers.
- Do not introduce a rainbow status palette by default. State must remain
  understandable through text, shape and hierarchy, not color alone.

## Typography

- Manrope remains the display face with full Cyrillic support.
- Oversized type is a primary graphic tool.
- Character comes from scale, line breaks, crop, spacing and placement rather
  than decorative effects.
- Body/UI copy uses the system sans stack.
- Tabular data and identifiers may use a restrained monospace/system-mono role
  where it strengthens the dossier/system language; do not add another
  webfont solely for this.
- Define distinct roles for display, headline, body, meta and data rather than
  solving hierarchy with more cards.

## Composition

- Prefer wide/full-bleed editorial composition where it serves the surface;
  `max-width` is a tool, not the identity.
- Asymmetry is intentional and different sections may use different internal
  compositions.
- Grid is an information framework, not a card generator.
- Do not default to `repeat(3, 1fr)` feature sections, universal Bento grids or
  equal-height tiles.
- Some surfaces should contain plain type, rules and whitespace with no
  container at all.
- Rounded cards are accents, not universal anatomy.
- Adjacent sections need not share equal visual loudness or whitespace.
- One strong visual gesture usually beats several simultaneous gestures.

## Illustration and semantic graphics

Illustration must carry product meaning or identity. It is not filler for
empty layout space.

- Original flat vector language; no stock illustration look.
- Illustration sets should be heterogeneous rather than an icon pack.
- Repeated formulae such as "object inside circle + spark" are rejected as a
  default.
- No arbitrary 3D or abstract identity blobs.
- Reusable graphic motifs should acquire stable product semantics where
  possible: stage marker, case/document, submission, locked state, jury mark,
  current position, result, etc.
- A graphic motif should be able to recur across public pages, product
  workspaces and event collateral without becoming decorative noise.

## Public surfaces

Public routes are allowed to be expressive.

They may use:

- very large type;
- art-directed illustration;
- asymmetric color fields;
- unusual whitespace and crop;
- heterogeneous section composition;
- rare high-salience motion.

The homepage should increasingly behave like the live cover/status board of a
competition rather than a static marketing funnel. Its content and emphasis
should reflect the current competition state when backend data supports it.

Public expressiveness does not justify repetitive agency-landing-page
patterns such as hero → three cards → color block → FAQ with uniform reveal
animation.

## Product surfaces

USER and JURY surfaces are operational tools. Their visual grammar should be
more document-like and information-dense:

- dossier/work-sheet composition;
- rules and separators;
- identifiers;
- explicit state and deadlines;
- tables/lists where the information is tabular;
- restrained color fields;
- minimal decorative illustration;
- motion primarily for state and continuity.

Do not scale the landing-page grammar directly into product workspaces.

## Admin surfaces

ADMIN is primarily utilitarian. Prioritize scanability, correctness,
operational safety and dense information. Brand identity may appear through
type, rules and color accents, but theatrical composition is inappropriate.

## Motion

Motion is a supporting system, not a design theme.

Preferred reasons for motion:

1. navigation continuity;
2. state transition;
3. information progression;
4. interaction feedback;
5. rare public-page spectacle.

Avoid continuous reveal choreography, default scroll drift, pointer parallax
on routine surfaces, and animation whose main purpose is to demonstrate
animation.

Every motion surface must have a static, understandable baseline and a strong
reduced-motion path.

## Anti-slop contract

The following are rejected as default grammar:

- generic B2B SaaS appearance or copy;
- KPI-card dashboard intros;
- universal rounded cards;
- Bento grids used as identity;
- pill/badge overload;
- feature-icon + heading + paragraph triptychs;
- abstract gradient spheres/blobs;
- glassmorphism and glow;
- arbitrary purple/blue gradients;
- decorative WebGL/3D;
- stock icon-pack illustration;
- repetitive reveal/fade-up choreography;
- decorative motion without product meaning;
- "Welcome back" dashboard hero patterns;
- equal visual polish and equal density everywhere.

These are anti-default rules, not primitive bans. Use a card, pill, radius or
animation when the information or interaction genuinely calls for it.

## Mobile as first-class composition

Mobile is not a collapsed desktop layout. Art-direct the 320/375 px
composition deliberately.

Core functionality must work:

- from 320 px upward;
- on touch-only devices;
- without hover or precise pointer;
- with reduced motion;
- with touch targets at least 44 px where interactive.

Controlled imperfection must not become unpredictable mobile layout. On small
screens, clarity wins when tension and usability conflict.
