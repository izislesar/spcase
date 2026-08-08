# Frontend target architecture

> **Status: approved target direction; foundation scaffolded (roadmap phase 3).**
> The current frontend is the server-rendered implementation in `web/`,
> embedded in the Go binary. It remains the running system and the behavioral
> reference until parity is captured and cutover is accepted.
> See `../decisions/0001-frontend-v2.md` for the decision record.
>
> Phase 3 delivered `frontend/`: a Vite 8 + React 19 + TypeScript (strict)
> application with React Router 8 (data mode) covering the legacy route map
> as placeholders, the `/api/v1` fetch transport, CSS token/base primitives,
> Biome and Playwright configuration. This is foundation only — no parity,
> no visual design, no production wiring. Details: `../../frontend/AGENTS.md`.

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
- SVG-first graphics

Motion:

- Motion for component-level animation
- View Transitions as progressive enhancement
- CSS scroll-driven animations as progressive enhancement
- GSAP + ScrollTrigger only for justified narrative sequences
- Lenis only where justified on public experiences; not required for
  application workspaces
- optional Rive for rare signature interactions

Tooling:

- Playwright (end-to-end)
- Biome (lint/format)
- pnpm (package manager)

## Progressive enhancement rule

Experimental browser features (View Transitions, scroll-driven animations,
Container Queries, `color-mix()`, Subgrid) must **progressively enhance a
usable baseline**. Core functionality must remain fully usable without them.

## What is not decided yet

- exact project layout and build/deploy wiring for the new frontend;
- how static frontend delivery is separated from the Go service in Compose;
- the behavioral-parity checklist (produced by the contract-audit stage).

Do not treat this document as permission to start the implementation: the
migration begins only after the existing frontend's behavioral contract is
captured (see `../../ROADMAP.md`).
