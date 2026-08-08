# ADR 0001: Replace the server-rendered frontend with an independent frontend application

- Status: **Approved — migration not yet executed**
- Date: 2026-08

## Context

SPCase v1.0.0 ships a server-rendered frontend: Go templates, compiled
Tailwind CSS and a small Alpine.js/motion layer embedded into the Go binary
via `go:embed` (`web/`). This implementation is complete and stable, but it
limits the visual and interactive ambitions for the public product surface and
couples frontend delivery to the Go release cycle.

The project is entering a new phase: rebuild the frontend as an independent
modern application while preserving the existing backend.

## Decision

Replace the embedded/server-rendered frontend with an independent frontend
application that consumes the existing `/api/v1` backend. The approved target
stack and design direction are recorded in `../frontend/architecture.md` and
`../frontend/design-direction.md`.

## Constraints

- The existing `web/` implementation **remains the authoritative frontend**
  until behavioral parity is demonstrated and cutover is accepted.
- Go application behavior must not be changed as part of context restructuring
  or frontend preparation; the API contract (`../contracts/http-api.md`) and
  cookie/auth behavior are preserved.
- The frontend behavioral contract must be **captured from the existing
  implementation before any part of it is deleted**.
- Cookie/auth/API compatibility must be preserved end to end: same
  `access_token` cookie contract, same endpoints, same error envelope.
- The eventual deployment architecture should separate static frontend
  delivery from Go API responsibilities; the exact wiring is decided during
  the frontend-foundation stage, not here.
- Final cutover requires parity testing, mobile/accessibility verification and
  **explicit acceptance**; legacy `web/` is retired only after that acceptance.

## Consequences

- Staging acceptance and production deployment are deferred until after the
  frontend migration gate (see `../../ROADMAP.md`).
- Backend work continues against a stable `/api/v1` contract; API changes
  during the migration require explicit justification because two frontends
  may consume them.
- Until cutover, every user-facing behavioral change must keep the legacy
  frontend working or be explicitly sequenced in the roadmap.
- This ADR intentionally does not fix implementation details (project layout,
  build wiring, hosting of static assets) that have not yet been validated.
