# Frontend design direction

> **Status: approved art direction for the future frontend.**
> These are design principles, not mockups. They apply to the independent
> frontend described in `architecture.md`; the current server-rendered `web/`
> implementation is unchanged.

## Art direction

The approved direction is **graphic-editorial, illustrative modernism**:
visually distinctive but restrained.

Principles:

- large typography;
- strong flat color fields;
- custom SVG/vector illustration;
- generous whitespace;
- unusual but simple composition;
- one strong visual gesture per section rather than many simultaneous effects;
- animation is deliberate and sparse.

## Explicitly rejected

- generic B2B SaaS appearance;
- repetitive card dashboards as the visual identity;
- generic purple/blue gradients;
- glassmorphism;
- decorative WebGL;
- arbitrary 3D;
- particle backgrounds;
- excessive motion;
- visual complexity without a product reason.

## Mobile as first-class composition

Mobile is a first-class composition target, not a collapsed desktop layout.

Core functionality must work:

- from 320 px upward;
- on touch-only devices;
- without hover;
- without a precise pointer;
- with reduced motion.

Desktop compositions must not simply collapse into stacked cards on mobile;
each viewport gets a deliberate composition.

## Operational workspaces

USER/JURY/ADMIN workflows are operational tools: they prioritize **clarity
and efficiency over theatrical animation**. The editorial visual identity
applies, but motion and decorative gestures are minimized where users perform
repeated task-focused work.

## Established visual rules (Stage 4A)

The public homepage implementation fixed these reusable rules; they are the
defaults for later stages (token values live in `frontend/src/styles/tokens.css`,
shared artwork in `frontend/src/components/graphics/grammar.tsx`):

- Graphic grammar: one small family of flat vector forms with product
  meaning. The case is an **open ring** (unresolved problem; its gap narrows
  through the stages and closes at the final); the team is a set of
  independent elements — **disc, block, half-disc**. The public page tells
  this as one story: hero (scattered elements around the open ring) →
  format vignettes (scatter → converge → assemble) → the resolved mark
  (closed ring around one assembled composition) in the final vignette and
  the footer. Schedule markers and the brand mark echo the same forms.
  Reuse these primitives instead of inventing new decoration.
- Palette: warm paper base, near-black ink, one vermilion accent and two
  deliberate supporting flat fields — warm sun yellow (`--color-sun`) and
  mint (`--color-mint`). Section rhythm on the homepage: paper → sun → dark
  ink → mint → accent footer. One dominant field per section; no gradients;
  colors are not all used in every section.
- Typography: one self-hosted display face — **Unbounded** (variable
  200–900, WOFF2 cyrillic + latin subsets, SIL OFL 1.1, `font-display:
  swap`) for headings and the wordmark; the system sans stack carries body,
  UI and tabular numerals. No other webfonts.
- Hero rule: typography is part of the composition (oversized first line
  with an inline team disc as its full stop, offset accent second line),
  with the open case ring anchored behind it — adjacency and tension, never
  text painted over the ink stroke.
- Signature mechanism: the trajectory — a bent SVG path with three hops
  between the format stages, with one accent disc riding it via
  `offset-path` + a CSS scroll-driven animation (no JS, no render loop).
  The fully drawn path and the disc resting at the finish are the baseline
  for reduced motion and unsupported browsers. One such gesture per page.
- Flat fields, not cards: sections are large flat color fields; content
  uses hairline rules, indexes and grammar markers instead of boxed cards.
- Accent is never used for body text on the dark field; on the accent
  footer field the focus ring switches to on-accent.
