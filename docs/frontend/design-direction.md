# Frontend design direction

> **Status: approved stable visual north star.**
> These are design principles, not mockups, and not a snapshot of any single
> implemented stage. They apply to the independent frontend described in
> `architecture.md`; the legacy server-rendered `web/` implementation is
> unchanged. Current human review state of the visual implementation lives
> separately in `visual-acceptance.md` — this document does not record
> per-stage verdicts.

## Art direction

The approved direction is **graphic-editorial, illustrative modernism**:
visually distinctive but restrained. Visual richness comes from scale,
illustration, color and composition — never from UI complexity.

Core grammar: **graphic maximalism + simple UI**. Large graphic surfaces and
bespoke artwork carry the identity; interface chrome stays minimal. Never the
reverse.

### Canvas and color

- White / very-light editorial canvas.
- Deep navy typography and ink.
- Saturated flat color fields: mustard, cyan/turquoise, coral/pink/red
  accents.
- Large color fields create deliberate editorial rhythm; not every section
  is equally loud, and there is no mechanical one-color-per-section
  sequence.
- **No gradients.** No glassmorphism, no soft glow layers.

### Typography

- One friendly heavy grotesk display face — **Manrope** (or a similar
  friendly heavy grotesk with full Cyrillic), used at heavy weights with
  tight tracking for display type and the wordmark.
- Oversized typography is a primary design tool: character comes from
  scale, spacing, line breaks and positioning, not decorative effects.
- The system sans stack carries body, UI and tabular numerals. No other
  webfonts.

### Illustration

- Original flat vector illustration with visual weight **comparable to
  typography** — artwork is a protagonist, not decoration.
- Thick simple silhouettes, flat fills, a limited per-scene palette. Each
  scene illustrates a real piece of the product.
- Illustration sets must be **heterogeneous**: scenes differ compositionally
  from each other and must NOT read as an icon pack.
- Do not repeat one formula across scenes — in particular avoid
  "object inside a colored circle" as a default and avoid recurring
  sparks/pluses/flags as decorative filler.
- No stock-illustration look, no abstract identity geometry, no 3D.

### Composition

- Prefer **full-bleed / wide editorial compositions** over a universally
  centered safe container; `max-width` is a tool, not the identity.
- Asymmetry is intentional; scenes should differ compositionally.
- Editorial collage, not Bento Grid: avoid converting editorial composition
  into uniform tiles, and avoid a universal card/tile anatomy.
- Some graphic surfaces should have **no radius, no shadow and no card
  chrome**; rounded poster cards are an accent, not a default.
- Composition may intentionally crop or extend artwork inside controlled
  bounds.
- One strong visual gesture per section rather than many simultaneous
  effects; adjacent sections are not equally loud.

### UI and motion

- Simple UI: minimal chrome, no pill-control dashboards, no repetitive card
  anatomy as the visual identity.
- Animation is deliberate and **sparse**; everything has a static, fully
  understandable baseline and reduced motion collapses all of it.

## Explicitly rejected

- generic B2B SaaS appearance and language;
- repetitive card dashboards or Bento grids as the visual identity;
- icon-pack illustration grammar (same formula repeated across scenes);
- generic purple/blue gradients;
- glassmorphism;
- decorative WebGL;
- arbitrary 3D;
- particle backgrounds;
- excessive motion;
- visual complexity without a product reason.

## Mobile as first-class composition

Mobile is a first-class composition target, not a collapsed desktop layout.
Mobile remains **art-directed**: each viewport gets a deliberate
recomposition, not a stack of shrunken desktop cards.

Core functionality must work:

- from 320 px upward;
- on touch-only devices;
- without hover;
- without a precise pointer;
- with reduced motion.

## Operational workspaces

USER/JURY/ADMIN workflows are operational tools: they prioritize **clarity
and efficiency over theatrical animation**. The editorial visual identity
applies, but motion and decorative gestures are minimized where users perform
repeated task-focused work.
