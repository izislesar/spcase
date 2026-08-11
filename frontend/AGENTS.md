# AGENTS.md — SPCase frontend context router

## Current state

`frontend/` is the independent React migration target for SPCase. It is NOT
wired into production yet; legacy `web/` remains the running frontend and the
behavioral reference until explicit cutover.

The public React surfaces have a technically validated Stage 4D implementation,
but the human review on 2026-08-11 returned **ITERATE**, not ACCEPT. Preserve
the established visual DNA, but do not treat current DOM/CSS/composition or
motion density as a design source of truth. The next stage is Stage 4E: art-
direction consolidation around the approved product/visual model.

USER/JURY/ADMIN routes are still mostly structural shells. Do not extrapolate
the public landing-page grammar into those workspaces by default.

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
- SVG/DOM/CSS-first graphics
- `motion` is installed and is the default animation library

No Tailwind and no component-library visual system in the React target.

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

**Editorial Competition OS with controlled imperfection.**

The product should feel like a live competition system — part editorial
publication, scoreboard, dossier and judging desk — not like generic SaaS.

Use roughly **90% discipline / 10% disobedience**:

- discipline: hierarchy, grid, readability, predictable controls,
  accessibility, stable semantics;
- disobedience: rare asymmetric composition, uneven whitespace, intentional
  crop/grid escape, oversized numbering, physical marks/stamps and deliberate
  density changes.

Imperfection must create tension, not fake handmade texture. Do not add random
rotation, paper grain, scratches, noisy backgrounds or arbitrary misalignment
merely to look "human".

### Public vs product surfaces

Public routes may be expressive: large type, illustration, asymmetric fields,
strong whitespace and rare spectacle motion.

USER/JURY workspaces are information-dense operational tools: document-like
layouts, rules, labels, tables, status, identifiers and restrained motion.

ADMIN is primarily utilitarian.

Never scale the landing-page visual grammar directly into operational
workspaces.

## Anti-slop contract

Avoid as default visual grammar:

- Bento-grid or equal-card composition;
- KPI-card dashboards;
- universal rounded containers;
- pill/badge overload;
- glassmorphism, gradients, glow or generic 3D;
- stock "feature icon + title + paragraph" sections;
- abstract decorative blobs/geometries without product semantics;
- repetitive identical illustration formulas;
- "Welcome back" dashboard hero patterns;
- continuous reveal/parallax choreography;
- decorative motion whose only purpose is to demonstrate animation.

A card is allowed when the underlying object is semantically a self-contained
card. A rounded corner or animation is allowed when it earns its role. These
are constraints against defaults, not blanket bans on primitives.

## Motion policy

Motion must have a product reason. Preferred classes:

1. navigation continuity / shared surface transitions;
2. state transitions (opened, selected, submitted, error, locked);
3. information progression (current stage, schedule/time progression);
4. microresponse (hover/press/focus);
5. rare public-page spectacle, primarily the hero.

Respect `prefers-reduced-motion` at every motion call site or through a shared
primitive that guarantees it. Static state must remain fully understandable.

Pointer parallax and scroll drift are not default primitives. Do not add GSAP,
Lenis or Rive without an explicit task-level product need and approval.

## Code rules

- Keep `pnpm typecheck`, `pnpm check` and `pnpm build` green.
- Russian UI copy; document language remains `ru`.
- Mobile is first-class: deliberate composition from 320 px upward, touch
  targets at least 44 px, no hover-dependent functionality.
- Accessibility invariants in `legacy-contract.md` are mandatory.
- Experimental CSS/browser features must progressively enhance a usable
  baseline.
- No unjustified runtime dependencies.
- Current implementation is evidence, not visual authority, when it conflicts
  with the approved design direction or current human verdict.

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
