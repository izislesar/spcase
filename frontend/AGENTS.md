# AGENTS.md — SPCase frontend agent context

Independent React application for SPCase (spcase.ru). The foundation was
scaffolded in roadmap phase 3; since then a working public visual
implementation (home, schedule and the other public routes) has been built
and iterated through Stage 4/4A/4B. NOT wired into production; the legacy
server-rendered `web/` remains the running frontend and the behavioral
authority until cutover.

## Approved stack

- React 19 + TypeScript (strict) + Vite 8
- React Router 8 in **data mode** (`createBrowserRouter` in `src/app/router.tsx`) — never framework mode
- pnpm (only package manager; `packageManager` field is pinned)
- TanStack Query (server state), React Hook Form + Zod (forms/validation)
- Biome (lint + format; `pnpm check` must stay clean)
- Playwright (e2e, desktop + mobile projects; `pnpm test:e2e`)
- CSS Modules + modern native CSS: custom properties in `@layer tokens`,
  reset/primitives in `@layer base`, OKLCH, container queries, Grid/Flex

## API contract

- All API calls go through `src/lib/api/client.ts` (`apiGet`/`apiPost`/…).
- Base path is the relative string `/api/v1` — never configurable, no env
  var, no production base URL. Dev reaches the Go backend via the Vite
  proxy to `http://localhost:8000` (`vite.config.ts`).
- Always `credentials: "same-origin"`; `Content-Type: application/json` only
  when a body is present.
- Server error envelope is `{ "error": { "code": string, "message": string } }`;
  parse it via `src/lib/api/errors.ts` (`ApiError`). Never expose raw
  payloads or stack traces to users.
- Auth is the `access_token` HttpOnly cookie; the frontend never reads or
  stores tokens. No localStorage tokens, no token-in-URL.

## Code rules

- TypeScript strict: `tsc --noEmit` covers `src`, `vite.config.ts`,
  `playwright.config.ts`, `e2e`. Keep it green (`pnpm typecheck`).
- Russian UI copy; `lang="ru"`.
- Mobile is first-class: mobile-first functional parity, compose from
  320 px, touch targets ≥ 44 px, no hover-dependent functionality.
- Accessibility is mandatory, not optional: keep the ACCESSIBILITY
  requirements of `docs/frontend/legacy-contract.md` satisfied in every
  change.
- Respect `prefers-reduced-motion`; `scroll-behavior: smooth` is the only
  default motion. Experimental features (container queries, `color-mix()`,
  view transitions) must progressively enhance a usable baseline.
- SVG/DOM/CSS-first graphics.
- No unjustified runtime dependencies (see Dependencies below); no backend
  contract invention — the API contract in `docs/contracts/http-api.md` and
  the error envelope are fixed.

## Dependencies

Restrained policy: prefer existing deps and platform features. Motion, GSAP,
Lenis and Rive are approved in the architecture doc but may only be added
when a concrete feature justifies them. No Tailwind, no component libraries.

## Design direction

Graphic-editorial / illustrative-modernist. Explicitly rejected: generic
B2B SaaS look, glassmorphism, arbitrary purple/blue gradients, decorative
WebGL, excessive motion. See `docs/frontend/design-direction.md`.

Before any visual work, read BOTH:

- `docs/frontend/design-direction.md` — the stable visual north star;
- `docs/frontend/visual-acceptance.md` — the current human visual review
  state (what is accepted, what is rejected, what the next stage must
  improve, and the gate before Stage 5).

Important: the current DOM/CSS/component structure is NOT automatically a
visual source of truth. When human visual acceptance rejects an
implementation pattern, agents may substantially recompose the visual layer
(markup structure, CSS, illustration components) while preserving behavior
and accessibility. A technically valid, committed implementation is not
necessarily a visually accepted one.

## Behavioral parity

Defined by `docs/frontend/legacy-contract.md`. The legacy `web/` frontend is
authoritative for behavior until cutover; do not invent flows here that
contradict it. Route map and cutover mechanics:
`docs/frontend/cutover-plan.md`.

## Commands

```bash
pnpm install
pnpm dev          # Vite dev server on :5173, /api/v1 proxied to :8000
pnpm build        # typecheck + production build into dist/
pnpm typecheck
pnpm check        # Biome lint + format check
pnpm test:e2e     # Playwright (needs browsers installed)
```
