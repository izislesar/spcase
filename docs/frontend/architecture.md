# Frontend target architecture

> **Status: approved target direction; migration in progress.**
> The current production frontend is the server-rendered implementation in
> `web/`, embedded in the Go binary. It remains the running system and the
> behavioral reference until parity is captured and cutover is accepted.
> See `../decisions/0001-frontend-v2.md` for the decision record.
>
> The independent `frontend/` project exists and is the migration target:
> a Vite 8 + React 19 + TypeScript (strict) application with React Router
> (data mode), TanStack Query, the `/api/v1` fetch transport, CSS
> token/base primitives, Biome and Playwright configuration. Stage 4E was
> technically implemented but human review on 2026-08-11 rejected its visual
> direction. Stage 4F (dark de-stylization) is the next approved public design
> implementation stage (`visual-acceptance.md`). Product UX semantics live in
> `experience-model.md`. Production cutover is defined (`cutover-plan.md`)
> but NOT executed. Details: `../../frontend/AGENTS.md`.

## Direction

The frontend will become an **independent application** that consumes the
existing Go backend through `/api/v1`. The current server-rendered `web/`
implementation is retained until behavioral parity is demonstrated and the
cutover is explicitly accepted.

Target end state:

- static frontend delivery separated from Go API responsibilities;
- cookie/auth contract and `/api/v1` behavior preserved unchanged;
- Go backend keeps API, business logic, persistence and (until cutover) the
  legacy server-rendered pages.

## Approved target stack

Core:

- React 19
- TypeScript (strict)
- Vite 8
- React Router 8
- React Compiler where compatible
- TanStack Query
- React Hook Form
- Zod

Styling and visual system:

- CSS Modules plus modern native CSS
- CSS custom properties
- Grid / Subgrid
- Container Queries
- OKLCH / `color-mix()`
- SVG only where a real interface/brand need exists; decorative illustration is
  not a default architecture requirement

Motion and browser interaction:

- `motion` is installed and is the default component-level animation library
- View Transitions are a progressive enhancement
- CSS scroll-driven animations are a progressive enhancement when they solve
  a real information/interaction problem
- GSAP/ScrollTrigger, Lenis and Rive are **not default project dependencies**;
  adding any of them requires an explicit task-level product need and approval
- product workspaces should prefer platform/CSS/Motion primitives and sparse
  state-oriented animation over narrative choreography

Tooling:

- Playwright (end-to-end)
- Biome (lint/format)
- pnpm (package manager)

## Progressive enhancement rule

Experimental browser features (View Transitions, scroll-driven animations,
Container Queries, `color-mix()`, Subgrid) must **progressively enhance a
usable baseline**. Core functionality must remain fully usable without them.

## Current implementation state

- The independent `frontend/` project exists (pnpm, Vite 8, React 19,
  TypeScript strict, Biome, Playwright).
- Vite builds `dist/index.html` plus fingerprinted assets under
  `dist/assets/*` — the output structure expected by the cutover plan.
- Routing: React Router Data Mode (`createBrowserRouter`); framework mode
  is not used.
- API requests go to the relative path `/api/v1` through a single fetch
  client; in development a Vite proxy forwards `/api/v1` to the Go backend
  on `localhost:8000`, keeping calls same-origin.
- Browser credential model: `credentials: "same-origin"`; the
  `access_token` HttpOnly cookie is never read or stored by the frontend.
- Behavioral requirements are fixed in `legacy-contract.md`; production
  delivery topology is defined in `cutover-plan.md`. The cutover is defined
  but NOT yet executed.

## Related frontend authorities

- `design-direction.md` — stable visual constitution;
- `experience-model.md` — stable product UX model;
- `visual-acceptance.md` — current human visual verdict;
- `legacy-contract.md` — behavioral parity requirements;
- `cutover-plan.md` — defined production delivery/cutover mechanics.

The target cutover topology is already defined in `cutover-plan.md`; it is not
implemented yet.
