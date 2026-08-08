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
