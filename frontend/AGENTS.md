# AGENTS.md — SPCase frontend context router

## Current state

`frontend/` is the independent React migration target for SPCase. It is NOT
wired into production yet; legacy `web/` remains the running frontend and the
behavioral reference until explicit cutover.

Stage 4E (`73185c5`, `feat(frontend): consolidate editorial art direction`) is
technically complete but was **visually rejected by human review on
2026-08-11**. Do not treat its composition, illustration, palette usage or
"controlled imperfection" devices as approved visual precedent.

The next approved implementation stage is **Stage 4F — dark de-stylization**.
Its authority is `docs/frontend/design-direction.md` plus the current verdict
in `docs/frontend/visual-acceptance.md`.

USER/JURY/ADMIN routes are still mostly structural shells. Do not start Phase
5 unless the task explicitly says Stage 4F has received human ACCEPT.

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
- `motion` is installed and available for justified interaction/state motion

No Tailwind and no component-library visual system in the React target.

Do not add GSAP, ScrollTrigger, Lenis, Rive or another runtime visual framework
without explicit task-level approval.

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

**Dark, content-first competition interface.**

The frontend should feel like a real championship with a straightforward
interface, not like a designer's concept website.

Governing tests:

> Nothing exists solely to make the page look designed.

> If an element cannot justify itself without the phrase "visual interest",
> remove it.

### Dark palette

The default direction is a continuous near-black/deep-navy canvas with warm
light text, subtle rules and one deep-red primary accent. Secondary color is
rare and muted.

Do not reproduce the Stage 4D/4E mustard + cyan + coral + navy field system.
Do not solve the dark direction with nested lighter cards.

Dark must not drift into cyberpunk, terminal, developer-tool, neon, glass,
glow, luxury-black or blue/purple-gradient aesthetics.

### Content before decoration

Prefer real competition information over illustration or ornamental geometry.
Dates, deadlines, team size, schedule, scores, files and lifecycle state are
valid visual material.

The current gear/flag/machine/podium/sheet illustration language is not an
approved identity requirement. Stage 4F should remove it from dominant public
composition and must not replace it with another decorative illustration
system.

### Natural variation, not designed imperfection

Do not implement a quota of grid escapes, broken rules, crops, rotations or
asymmetric tricks. The old `90% discipline / 10% disobedience` formula is no
longer an implementation rule.

Allow different content to produce different density and spacing, but do not
manufacture irregularity to look human.

### Public vs product surfaces

PUBLIC is calm and content-first. Large type is allowed when it reflects real
hierarchy, but public pages should not depend on illustration, giant numerals,
color blocks or slogan copy.

USER/JURY workspaces are information-dense operational tools. ADMIN is
primarily utilitarian.

Do not scale landing-page grammar into operational workspaces.

## Anti-slop / anti-AI contract

Reject as dominant grammar:

- Swiss/editorial agency-site imitation;
- giant stage numerals;
- decorative `01 · SECTION` labels;
- Bento/equal-card layouts;
- KPI-card dashboards;
- universal rounded dark surfaces;
- art-panel auth layouts;
- gear/flag/document/podium illustration motifs;
- large multicolor section fields;
- deliberate broken-grid tricks;
- slogans replacing factual labels;
- glassmorphism, glow, gradients, neon;
- terminal/HUD cosplay;
- decorative scroll choreography.

A primitive is allowed when the information or interaction genuinely requires
it. These are anti-default rules, not syntactic bans.

## Motion policy

Motion is deliberately sparse. Preferred reasons:

1. navigation continuity;
2. state transition;
3. direct interaction feedback;
4. truthful temporal progression where it improves comprehension.

Public scroll reveal, decorative parallax, scroll drift and floating artwork
are not approved defaults.

Respect `prefers-reduced-motion`. Static state must be complete and clear.

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
