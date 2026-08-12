# AGENTS.md — SPCase frontend context router

## Current state

`frontend/` is the independent React migration target for SPCase. It is NOT
wired into production yet; legacy `web/` remains the running frontend and the
behavioral reference until explicit cutover.

Stage 4G (`db9753a`, `feat(frontend): add restrained spatial identity`) is
technically complete. Human review on 2026-08-11 accepted the **direction** but
found the composition incomplete: the dark spatial language is right, while
its scale and relationships are too tentative on wide desktop surfaces.

The next approved implementation stage is **Stage 4H — spatial composition &
density**. It must keep the Stage 4F cleanup and Stage 4G spatial vocabulary,
then strengthen composition through larger visual footprint, connected spatial
relationships and better use of real content. Its authority is
`docs/frontend/design-direction.md` plus the current verdict in
`docs/frontend/visual-acceptance.md`.

Do not add a new visual language to solve density. Do not restore Stage 4E's
illustration/poster grammar. Do not start Phase 5 unless the task explicitly
says Stage 4H has received human ACCEPT.

## Documentation ownership

Project documentation is owned by the human + ChatGPT Web workflow.

Coding agents must **not edit `.md` project/context documents** as part of code
implementation unless the user explicitly overrides this rule for that task.
This includes:

- `AGENTS.md`;
- `README.md`;
- `ROADMAP.md`;
- `frontend/AGENTS.md`;
- `docs/**/*.md`.

When code work reveals a documentation gap, report it in the final report. Do
not silently update the docs yourself.

## Read context by task

### Public visual/composition task

Read:

1. `docs/frontend/design-direction.md`
2. `docs/frontend/visual-acceptance.md`

### New USER/JURY/ADMIN workflow

Read:

1. `docs/frontend/experience-model.md`
2. the relevant sections of `docs/frontend/legacy-contract.md`
3. the relevant endpoints in `docs/contracts/http-api.md`
4. `docs/frontend/design-direction.md` for visual constraints

### API/query/auth integration

Read the relevant `docs/contracts/http-api.md` and behavioral-contract
sections. Do not invent backend behavior from UI needs.

### Cutover/deployment task

Read `docs/frontend/cutover-plan.md`, `docs/architecture/system.md` and ADR
`docs/decisions/0001-frontend-v2.md`.

Do not load the full `legacy-contract.md` for a purely visual public-page edit.

## Approved stack

- React 19 + TypeScript strict + Vite 8
- React Router 8 in data mode (`createBrowserRouter`), never framework mode
- pnpm only (`packageManager` is pinned)
- TanStack Query for server state
- React Hook Form + Zod for forms/validation
- Biome for lint/format
- Playwright for desktop/mobile e2e
- CSS Modules + modern native CSS
- CSS custom properties, Grid/Flex, container queries, OKLCH/`color-mix()`
- CSS perspective/3D transforms for progressive spatial UI
- `motion` is installed and available for justified interaction/state/spatial
  response

No Tailwind and no component-library visual system in the React target.

Do not add GSAP, ScrollTrigger, Lenis, Rive, Three.js, React Three Fiber, a
WebGL framework or another runtime visual framework without explicit task-level
approval. Stage 4H must continue using the existing CSS/Motion stack; it is a
composition pass, not a new rendering-stack decision.

## API contract

- All API calls go through `src/lib/api/client.ts` (`apiGet`/`apiPost`/…).
- Base path is the relative string `/api/v1`; never add a production API base
  URL or token transport.
- Use `credentials: "same-origin"`.
- Add `Content-Type: application/json` only when a body is present.
- Parse `{ "error": { "code": string, "message": string } }` through the
  existing `ApiError` path; never expose raw payloads or stack traces.
- Auth is the `access_token` HttpOnly cookie. Frontend code never reads,
  persists or forwards tokens through localStorage/URLs.

## Product and design model

North star:

**Dark Spatial Competition Interface.**

Stage 4F's dark/content-first cleanup remains the baseline. Stage 4G established
the approved spatial/material language. Stage 4H does not add another style; it
makes that language compositionally confident.

Governing tests:

> Fill space with scale, relationships and depth — not with more components.

> Visual complexity is allowed when it creates spatial hierarchy, interaction,
> orientation, or a memorable product object.

> Negative space is intentional when it gives hierarchy room. Dead space is an
> accidental absence of compositional relationship; do not fix it with filler.

### Composition and density

- Prefer one large coherent visual system over several small decorative objects.
- Wide desktop layout may use substantially more horizontal space while local
  reading columns remain narrow enough for comfortable text.
- Visual footprint should track importance: hero may be large; FAQ/forms do not
  need spectacle.
- Vary section density naturally. Do not normalize every section to the same
  height, padding or component anatomy.
- Do not fill empty regions with cards, labels, SVG filler, abstract geometry or
  extra copy merely to occupy space.

### Depth model

- `Z0`: normal 2D UI — forms, FAQ, navigation, body content, most product work.
- `Z1`: shallow structural depth — connected stage/day/state relationships.
- `Z2`: rare signature spatial object — primarily the public hero.

Do not make every section spatial. A flat surface next to a spatial surface is
intentional contrast.

### Dark material language

Use near-black/deep-navy canvas, warm text, restrained neutral material planes
and one deep-red accent.

Subtle gradients, shadows and edge highlights are allowed only to make
meaningful spatial layers readable. They are not general page decoration.

Do not drift into cyberpunk, HUD, terminal, neon, glass, glow, luxury-black or
purple/blue-gradient aesthetics.

### Hero / Format expectations

Hero should use one integrated spatial assembly/chassis rather than a visible
stack of independent floating cards. The object may be substantially larger on
wide screens and should make surrounding negative space feel related to its
geometry.

Format should read as one connected progression, not three separate dark cards
with barely perceptible `translateZ` differences.

### No return to illustration

Do not recreate or replace Stage 4E's gear/flag/machine/podium/sheet scene
system. Spatial richness should be built from meaningful surfaces/content, not
3D illustration or abstract floating objects.

### Public vs product surfaces

PUBLIC may have a few high-salience spatial moments. Hero is the signature
surface; Format may use connected Z1 depth. Schedule and FAQ should gain visual
weight primarily from data/typography/width, not extra 3D objects.

FAQ and auth remain predominantly Z0. Forms must stay conventional.

USER/JURY workspaces will use depth mainly for state/hierarchy once Phase 5 is
approved. ADMIN remains primarily utilitarian.

Do not scale a hero scene into operational workspaces.

## Anti-slop / anti-AI contract

Reject as dominant grammar:

- Stage 4E Swiss/editorial poster devices;
- decorative gear/flag/document/podium scenes;
- giant stage numerals;
- Bento/equal-card layouts;
- KPI-card dashboards;
- generic dark floating cards;
- multiple small spatial plates used where one coherent object would work;
- a tiny signature object floating inside a mostly unrelated empty desktop hero;
- three separate dark stage panels presented as spatial progression;
- floating abstract cubes/blobs;
- rotating 3D logo/object demos;
- glassmorphism, glow, neon;
- cyberpunk/terminal/HUD styling;
- meaningless industrial screws/grids/coordinates;
- pseudo-3D on every section;
- large cursor-following/parallax scenes;
- constant idle animation;
- slogan copy replacing factual labels.

A primitive is allowed when the information, interaction or spatial hierarchy
genuinely requires it. These are anti-default rules, not syntactic bans.

## Motion policy

Preferred reasons:

1. navigation continuity;
2. state transition;
3. direct interaction feedback;
4. truthful temporal progression;
5. restrained spatial response on a meaningful object.

Signature pointer response should normally be a few degrees of tilt, not large
travel/rotation. No idle floating or continuous object animation.

Respect `prefers-reduced-motion`. Static state must be complete and clear.
Touch/keyboard users must not depend on hover/pointer depth.

## Code rules

- Keep `pnpm typecheck`, `pnpm check` and `pnpm build` green.
- Russian UI copy; document language remains `ru`.
- Mobile is first-class: deliberate composition from 320 px upward, touch
  targets at least 44 px, no hover-dependent functionality.
- Accessibility invariants in `legacy-contract.md` are mandatory.
- Experimental CSS/browser features must progressively enhance a usable
  baseline.
- Important text/controls should not be distorted by strong perspective.
- No unjustified runtime dependencies.
- Current implementation is evidence, not visual authority, when it conflicts
  with the approved design direction or current human verdict.
- Do not fabricate API state, dates, counts or lifecycle information for visual
  effect.

## Behavioral parity

Behavioral parity is defined by `docs/frontend/legacy-contract.md`. Preserve
legacy outcomes until cutover unless an authoritative behavioral contract is
explicitly changed. Route/cutover mechanics live in
`docs/frontend/cutover-plan.md`.

## Commands

```bash
pnpm install
pnpm dev
pnpm typecheck
pnpm check
pnpm build
pnpm test:e2e
```
