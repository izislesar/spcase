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

## Established visual rules (Stage 4B)

The public implementation fixed these reusable rules; they are the defaults
for later stages (token values live in `frontend/src/styles/tokens.css`,
shared artwork in `frontend/src/components/graphics/illustrations.tsx`).
They supersede the earlier Stage 4A identity (open-ring metaphor, the
disc/block/half-disc semantic grammar and the Unbounded display face),
which is retired: graphic elements are art direction, not encoded product
symbolism.

- Grammar: **graphic maximalism + simple UI**. Visual richness comes from
  typography, composition, saturated flat color fields and bespoke flat
  vector illustration; interface chrome stays minimal. Never the reverse.
- Canvas and ink: near-white editorial canvas, deep navy ink
  (`--color-ink`), one coral accent (`--color-accent` /
  `--color-accent-strong` for text and buttons).
- Saturated flat fields: mustard (`--color-mustard`), turquoise
  (`--color-turquoise`), navy (`--color-navy`). Large fields create
  deliberate editorial rhythm (homepage: white → mustard panel → white
  mosaic → turquoise → white → navy footer); no gradients, no mechanical
  one-token-per-section sequence.
- Typography: one self-hosted display face — **Manrope** (variable 200–800,
  WOFF2 cyrillic + latin subsets, SIL OFL 1.1, `font-display: swap`), a
  friendly heavy grotesk with full Cyrillic, used at weight 800 with tight
  tracking for display type and the wordmark. Character comes from scale,
  spacing, line breaks and positioning, not decorative effects. The system
  sans stack carries body, UI and tabular numerals. No other webfonts.
- Illustration: original flat vector scenes — thick simple silhouettes,
  flat fills, limited per-scene palette, small spark accents, occasional
  dashed travel paths. Each scene illustrates a real piece of the product
  (case-solving machine, stages, calendar, cup, questions, pennant).
  No stock-illustration look, no abstract identity geometry, no 3D.
- Hero rule: heavy headline with deliberate line breaks on the left, ONE
  large bespoke scene on the right; the artwork interacts with the DOM
  typography spatially, never replaces it.
- Section rule: one dominant graphic move per section — an oversized
  colored panel (format), an irregular editorial mosaic of heterogeneous
  tiles (visual navigation, not feature cards), typographic rows on a flat
  field (schedule preview), quiet hairline rows (FAQ), a closing scene
  (footer). Adjacent sections are not equally loud.
- Poster surfaces: rounded rectangles with soft physical depth
  (`--shadow-poster`) and slight tilts are the tile/poster language;
  dashboard-style cards, bento grids and pill controls are rejected.
- Navigation: on desktop a thin fixed bottom bar (brand + the four
  destinations, reserved-space active marker); on mobile the compact top
  header with the focus-managed overlay menu.
- Motion stays sparse: a one-shot hero entrance assembly, a scroll-driven
  timeline progress line (guarded native CSS scroll timeline), small
  hover-only tile lifts. Everything has a static, fully understandable
  baseline; reduced motion collapses all of it.
- Accent is never used for body text on dark fields; on saturated fields
  text is navy ink or on-field pairs with checked contrast.
