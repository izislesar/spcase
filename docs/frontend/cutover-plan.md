# Frontend cutover architecture plan

> **Status: defined, not implemented.** This document fixes the exact
> current → target topology and the migration mechanics for replacing the
> embedded server-rendered frontend (`web/`) with the independent React
> application decided in `../decisions/0001-frontend-v2.md` and specified in
> `architecture.md`. Behavioral parity requirements live in
> `legacy-contract.md`; this plan covers ownership, delivery, build and
> deployment mechanics only. Nothing here changes running code.

## 1. Current URL ownership

Facts from `nginx.conf`, `cmd/app/main.go:258-294`, `web/handler.go` and the
`Dockerfile`.

**nginx (edge, the only published port `127.0.0.1:8080`):**

- `^~ /static/` — served **directly by nginx** from
  `/usr/share/nginx/html/static` (copied from the legacy frontend build in the
  Dockerfile `nginx` stage), with `Cache-Control: public, max-age=31536000,
  immutable` via the `$static_cache_control` map; `location = /static`
  308-redirects to `/static/`.
- `~ ^/api/v1(?:/|$)` — proxied to `app:8000` with edge error envelopes for
  413/429/502-504 and per-IP rate limits on the four login/registration URIs.
- `location /` — **everything else proxied to `app:8000`**, including all
  browser pages. nginx owns no HTML today.

**Go application (`app:8000`, not published):**

- `GET/POST /api/v1/*` — the complete API including health
  (`/api/v1/health/live|ready`), public data, auth, team, jury, admin and the
  XLSX export (`cmd/app/main.go:258-292`).
- `GET /` → `web.Handler` — the ten browser pages (`/`, `/schedule`,
  `/no-team`, `/login`, `/register`, `/dashboard`, `/jury/login`,
  `/jury/register`, `/jury/teams`, `/admin`), the 307 `/jury` → `/jury/teams`
  redirect, plain-text 404 for unknown paths, and `/static/*` **a second
  time** from `go:embed` (for direct runs without nginx).
- Both layers are wrapped by the global middleware chain: security headers,
  no-store for sensitive API prefixes, request ID, logging, CORS, recovery,
  API error normalization (`cmd/app/main.go:296-306`).

## 2. Target topology (confirmed)

```text
Browser
  │
  ▼
nginx (only published port, loopback)
  ├── ^~ /assets/            → built React static files (fingerprinted, immutable cache)
  ├── ~  ^/api/v1(?:/|$)     → Go app (proxy; rate limits + JSON edge envelopes unchanged)
  ├── =  /index.html, /      → index.html (Cache-Control: no-store)
  └── location /             → try_files $uri /index.html   (React Router fallback, no-store)
                                  │
                                  ▼ (proxy only)
                            Go app:8000 — /api/v1/* ONLY
```

- The React application owns every browser route: `/`, `/schedule`,
  `/no-team`, `/login`, `/register`, `/dashboard`, `/jury`, `/jury/login`,
  `/jury/register`, `/jury/teams`, `/admin`, and the 404 experience.
- Go owns exactly `/api/v1/*`. After final cutover Go neither renders nor
  embeds any frontend file; `GET /` web wiring is removed.
- nginx becomes the sole static/HTML delivery layer. The legacy Go
  "direct-run without nginx" frontend capability disappears — an accepted
  consequence; API-only direct runs stay possible.
- No new backend service is introduced; the React app is static files inside
  the existing nginx image.

This matches the preferred target in the stage assignment; nothing found in
the repository invalidates it.

## 3. Routes that must remain Go/API routes

All of `/api/v1/*`, verified against `cmd/app/main.go` and
`../contracts/http-api.md`:

- health: `GET /api/v1/health/live`, `GET /api/v1/health/ready` (also used by
  the nginx container healthcheck — must never hit the SPA fallback);
- public data: `/api/v1/info`, `/api/v1/schedule`, `/api/v1/faq`,
  `/api/v1/no-team`;
- auth: register/login/logout endpoints (rate-limited at the edge);
- team, submission, jury, admin endpoints;
- download: `GET /api/v1/admin/export/excel` (binary XLSX; must stay a
  proxied API route so `credentials: include` fetch + blob download keeps
  working exactly as in `legacy-contract.md` ADMIN-002).

Rule of thumb: the SPA fallback must never answer a URL beginning with
`/api/`, a health endpoint, a download, or an asset path. The nginx
`location ~ ^/api/v1(?:/|$)` regex already outranks prefix locations, and
fingerprinted assets live under their own `^~` prefix, so this is
structurally guaranteed (§4).

## 4. React Router history fallback mechanics

Required nginx shape (final form; exact syntax settled at implementation):

```nginx
location ^~ /assets/ {
    try_files $uri =404;          # fingerprinted React build output
    # Cache-Control: public, max-age=31536000, immutable (via extended map)
}

location ~ ^/api/v1(?:/|$) {      # unchanged from today
    proxy_pass http://spcase_app;
    # rate limits + 413/429/503 JSON envelopes unchanged
}

location / {
    try_files $uri /index.html;   # SPA fallback
    # Cache-Control: no-store on index.html / fallback responses
}
```

Consequences, verified against current config:

- `/api/v1/*` (regex location) always wins over `location /`; unknown API
  paths still reach Go and return the JSON 404 envelope — they never receive
  HTML.
- Requests for real files (`/favicon.ico`, `/manifest.webmanifest`,
  `/robots.txt`, anything actually present in the web root) are served
  directly; only misses fall back to `index.html`.
- Browser refresh / deep link on any React route (`/dashboard`,
  `/jury/teams`, …) returns `index.html` with 200; React Router then renders
  the route and the app's auth probe (`GET /user/me`) decides any redirect —
  the outcome-level parity requirements AUTH-007 and the role matrix in
  `legacy-contract.md` are preserved by client code, not by the server.
- The legacy `/jury` → `/jury/teams` 307 (Go today) becomes a React Router
  redirect. An exact nginx `location = /jury { return 307 /jury/teams; }`
  is an acceptable alternative; decide at implementation, default to the
  router.
- nginx `add_header` caveat: `add_header` inside a `location` replaces the
  inherited server-level list. The four security headers
  (`X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`,
  `Permissions-Policy`) and `X-Request-ID` must be repeated (or moved into
  an included snippet) in every location that sets its own `Cache-Control`.

## 5. Same-origin cookie authentication

Current mechanism (verified in `internal/delivery/http/v1/auth.go` and
`internal/delivery/http/middleware/auth.go`):

- `access_token` cookie: `HttpOnly; Secure; SameSite=Lax; Path=/;
  Domain=<APP_DOMAIN>`, set by the API on register/login, expired on logout;
  every protected request revalidates role/`auth_version`/`disabled_at`
  against PostgreSQL.
- The frontend never touches the token; all API calls use
  `credentials: "include"` against the same origin.

Cutover compatibility: the React app is served from **the same origin** as
the API (nginx terminates both). Cookies, `SameSite=Lax` CSRF posture, the
CORS origin allowlist and `Domain=<APP_DOMAIN>` keep working with **zero
backend changes**. Serving the SPA from a different origin is explicitly
rejected for this migration: it would require CORS credential changes and
re-examine CSRF, for no demonstrated benefit.

## 6. Development topology

```text
Browser → Vite dev server (localhost:5173)
            ├── React app + HMR
            └── proxy /api/v1 → http://localhost:8000  (go run ./cmd/app)
```

- The Vite dev-server proxy makes API calls same-origin, so cookie
  semantics match production exactly; `appType: "spa"` gives the history
  fallback for free in development.
- Local backend configuration: `APP_DOMAIN=localhost` so the
  `Domain=localhost` cookie is stored for the `localhost` origin (ports are
  not part of cookie scope). `CORS_ALLOWED_ORIGINS` does not need the Vite
  origin while the proxy is used; add `http://localhost:5173` only if a
  developer bypasses the proxy (direct cross-origin calls are rejected 403
  by CORS middleware otherwise — existing behavior, keep it).
- The Go API runs unchanged (`go run ./cmd/app` + local PostgreSQL, per
  `README.md`); the legacy `make frontend-build` workflow remains until
  final cutover for the legacy UI.

## 7. Production topology

As §2: nginx serves the React build and proxies `/api/v1/*`; the Go
container is API+business logic only. External TLS ingress stays out of
scope of the Compose stack exactly as today (loopback bind, unchanged).
Container healthchecks are unaffected: both the app and nginx healthchecks
hit `/api/v1/health/ready`, which remains a Go route.

## 8. Assets, hashing and caching

- Vite emits fingerprinted filenames under `assets/` (e.g.
  `assets/index-D3adb33f.js`); `index.html` references them by hash.
- Fingerprinted assets: `Cache-Control: public, max-age=31536000,
  immutable` — extend the existing `$static_cache_control` map (today keyed
  on `~^/static/`) with `~^/assets/`.
- `index.html` and every fallback response: `Cache-Control: no-store`
  (parity with PUBLIC-009: HTML is never cached). This is stricter than the
  current nginx behavior for HTML (today Go sets `no-store`; nginx passes it
  through) and must be set explicitly because nginx will now originate HTML.
- Non-fingerprinted public files (favicon, manifest, robots) get a short or
  no cache — decide at implementation; do not make them immutable.
- The legacy `/static/` location and its immutable map entry survive until
  legacy assets are removed at final cutover.

## 9. Removing go:embed / template / static responsibilities

Current responsibilities of the `web` package and `web/handler.go`:
page routing + titles, template parsing at startup, `go:embed` of
`template/` and `static/`, asset content-hash (`?v=`) cache busting,
no-store HTML, immutable `/static/*`, `/jury` redirect, plain 404.

Removal sequence (final cutover stage only, after the acceptance gates in
§17):

1. nginx serves all browser routes and static assets (§4) — deployed and
   verified **while Go still serves pages underneath** (proxy target for
   `location /` is simply no longer hit for page routes).
2. Delete `mux.Handle("GET /", webHandler)` and web handler construction in
   `cmd/app/main.go`; Go becomes API-only.
3. Delete the `web/` directory (`handler.go`, `handler_test.go`,
   `template/`, `src/`, `static/`) and the root legacy JS toolchain
   (`package.json`, `package-lock.json`, `tailwind.config.js`).
4. Update Dockerfile (§10), Makefile (§12), README/AGENTS pointers.

Until step 2, the embedded frontend remains in the binary as a zero-cost
fallback; rollback stays a config/image-level operation (§16).

## 10. Docker build implications

- `frontend-build` stage: today `npm ci` + Tailwind + esbuild over `web/`.
  At the foundation stage this becomes (or is joined by) a pnpm + Vite build
  of `frontend/` producing `dist/`. Node base image can be reused; pnpm is
  fetched via corepack — pin it like the rest of the toolchain.
- `nginx` stage: copy `frontend/dist` (not `web/static`) into
  `/usr/share/nginx/html`; the web root then contains `index.html`,
  `assets/`, and public files. During transition it may hold **both**
  `static/` (legacy) and the React build.
- `go-build` stage: `COPY --from=frontend-build … web/static` and the
  dependency on the frontend stage are removed at final cutover; the Go
  binaries no longer embed assets, shrinking the image.
- `migrator`, `runtime` stages: unchanged except the `runtime` image no
  longer carries embedded assets.
- Reproducibility rules from `AGENTS.md` stand: pinned base-image digests,
  lockfile-driven dependency install (`pnpm install --frozen-lockfile`).

## 11. nginx changes at final cutover

Single file: `nginx.conf`.

1. Add `^~ /assets/` static location with immutable cache (map extension).
2. Change `location /` from `proxy_pass http://spcase_app;` to
   `try_files $uri /index.html;` with explicit `Cache-Control: no-store`
   and repeated security headers (§4 caveat).
3. Keep `~ ^/api/v1(?:/|$)` byte-identical (rate limits, envelopes,
   proxy headers).
4. Remove `^~ /static/` and `location = /static` together with legacy
   assets (or keep them harmlessly until asset deletion; removal is part of
   cleanup).
5. Keep `root /usr/share/nginx/html;` — the same path, new content.

## 12. Compose implications

- `docker-compose.yml`: no structural change required. The `nginx` service
  keeps its build target, networks, loopback port and healthcheck; the `app`
  service keeps all environment variables. No new service is added (static
  content rides in the nginx image).
- `docker-compose.staging.yml`: unchanged (db port override only).
- Image tags (`SPCASE_APP_IMAGE`, `SPCASE_NGINX_IMAGE`) are the rollback
  handles — tag the cutover pair explicitly in staging/production deploys.
- Optional (later, not required): a `frontend` dev-only Compose service for
  the Vite server. Not part of the production topology.

## 13. Environment / configuration boundary

- Backend keeps every current variable (`DB_*`, `JWT_SECRET`,
  `JURY_REGISTRATION_KEY`, deadlines, `NO_TEAM_TELEGRAM_URL`, `APP_DOMAIN`,
  `CORS_ALLOWED_ORIGINS`, …) with unchanged semantics (`docs/architecture/
  system.md` → Configuration).
- The frontend needs **no runtime configuration and no secrets**: the API
  base is the same-origin relative path `/api/v1`; deadlines, registration
  state, Telegram URL and public content arrive at runtime from
  `GET /api/v1/info`, `/api/v1/schedule`, `/api/v1/faq`, `/api/v1/no-team`
  exactly as the legacy frontend consumes them today.
- Build-time `VITE_*` variables, if ever introduced, are public by design
  (embedded in the bundle); never place secrets there. Development proxy
  target is the only expected one.
- The cookie domain stays `APP_DOMAIN`; no frontend coupling to its value.

## 14. Deep links, refresh and mobile behavior

- Any React route must survive a cold load: refresh, shared link, mobile
  browser restart. Guaranteed by the `try_files $uri /index.html` fallback
  (§4) plus client-side auth probing; requirement carried from
  `legacy-contract.md` (route map + AUTH-007).
- Deep links to legacy-only server behaviors (`/jury` redirect, 404) are
  handled by the router per §4.
- Mobile/touch/reduced-motion parity is a behavioral concern covered by
  RESPONSIVE/ACCESSIBILITY requirements in `legacy-contract.md`, not by
  deployment topology; no server mechanism is involved.
- bfcache safety remains client-side (LIFECYCLE-002); `no-store` HTML means
  navigations after logout/mutations always re-fetch fresh HTML — unchanged
  from today.

## 15. Security implications

Verified current posture:

- **CSRF:** there is no token-based CSRF protection anywhere in the
  repository. Defense-in-depth today = `SameSite=Lax` cookie + strict
  credentialed CORS allowlist (disallowed `Origin` → 403
  `internal/delivery/http/middleware/cors.go`) + JSON-only bodies with
  `Content-Type: application/json` (forces preflight). Same-origin React
  delivery preserves all three exactly; no new CSRF surface is introduced.
- **CSP:** no Content-Security-Policy exists today (neither nginx nor Go
  sets one). Introducing a CSP is a foundation-stage hardening option —
  Vite output is CSP-friendly (external scripts/styles) — but it is **not**
  a cutover requirement and must not be smuggled into the cutover change.
- **Security headers:** the four baseline headers are set by both nginx and
  Go today (nginx hides the Go duplicates for proxied responses). When nginx
  starts originating HTML, it must keep emitting them on static/fallback
  responses (§4 `add_header` inheritance caveat).
- **Cookies:** `HttpOnly` means the SPA cannot and must not read the token;
  session state is always probed through `/api/v1/user/me` — identical to
  legacy behavior.
- **Rate limits** on login/registration URIs live at the edge and are
  untouched.
- **no-store on sensitive API prefixes** (`cache_control.go`) is
  path-based and unaffected by the cutover.
- The new frontend must not introduce token-in-URL, localStorage tokens, or
  any client-side secret handling — the cookie contract in
  `legacy-contract.md` AUTH-001 stands.

## 16. Rollback strategy

Cutover is a **delivery-layer change only**: no database, API, or business
logic changes. Rollback is therefore image/config-level:

1. Pre-cutover: record the running `SPCASE_APP_IMAGE`/`SPCASE_NGINX_IMAGE`
   tags (the last legacy pair).
2. Cutover deploy ships the new nginx image (SPA + fallback config) paired
   with the app image that still contains the embedded legacy UI.
3. Failure before legacy deletion: redeploy the recorded previous nginx
   image tag — page routes immediately proxy to Go again, which still
   serves the full legacy UI. No data migration, no credential or role
   changes involved.
4. Failure after `web/` deletion: rebuild the recorded pre-cutover git
   revision (tagged) and redeploy that pair; the deletion commit is the
   point of no cheap return, which is why §17 gates it.
5. The rehearsal path already exists: local staging Compose
   (`README.md` → Local staging) with production images validates the exact
   cutover pair before any production change.

## 17. Acceptance gates before deleting `web/`

All must pass, in order, with evidence recorded:

1. **Parity:** every requirement in `legacy-contract.md` is either verified
   against the React app (Playwright + manual checklist) or explicitly
   accepted as changed per its "Not automatically preserved" list.
2. **Automated:** Playwright suite green on desktop + mobile viewports,
   touch emulation, and `prefers-reduced-motion`; accessibility checks per
   the ACCESSIBILITY requirements.
3. **API contract:** backend test suites unchanged and green
   (`make security-check`, integration suite); zero backend behavioral diff
   except removal of page routes (step 2 of §9 happens only after gates
   1–4).
4. **Staging:** full cutover pair deployed to local staging; smoke of every
   route in the route map, deep-link/refresh checks, download flow, login
   rate-limit behavior, cache headers (immutable assets vs no-store HTML).
5. **Acceptance:** explicit human sign-off on parity and staging evidence.
6. Only then: §9 steps 2–4 (Go wiring removal, `web/` and legacy toolchain
   deletion) in a separate, revertible commit.

## 18. File-by-file migration surface

**Created by future stages (new frontend foundation):**

- `frontend/` — new application: `package.json`, `pnpm-lock.yaml`,
  `vite.config.ts`, `tsconfig.json`, `index.html`, `src/`, Playwright
  config, Biome config (exact layout settled at foundation stage).
- `frontend/AGENTS.md` — frontend-specific agent context (planned in
  roadmap phase 3).
- Updates to `.gitignore` for `frontend/dist/` and pnpm artifacts.

**Modified at cutover (delivery layer):**

- `nginx.conf` — §11 changes.
- `Dockerfile` — `frontend-build` stage retargeted to `frontend/`;
  `nginx` stage copies `dist/`; `go-build` drops the `web/static` COPY.
- `docker-compose.yml` — no structural change; possibly comments only.
  `docker-compose.staging.yml` — unchanged.
- `Makefile` — `frontend-build` target points at the new toolchain;
  `security-check` swaps `npm audit` for the pnpm equivalent.
- `README.md` — development instructions (Vite dev topology).
- `AGENTS.md` — doc-map/pointers once `frontend/` exists.
- `docs/frontend/architecture.md`, `ROADMAP.md` — status updates.

**Deleted at final cutover (after §17 gates):**

- `web/` entirely: `handler.go`, `handler_test.go`, `template/` (11 files),
  `src/` (`app.js`, `animations/*`, `interactions/*`, `input.css`),
  `static/` (compiled `css/app.css`, `js/app.js`).
- Root `package.json`, `package-lock.json`, `tailwind.config.js`
  (legacy toolchain; the new one lives in `frontend/`).
- `cmd/app/main.go` — only the web handler construction and
  `mux.Handle("GET /", webHandler)` lines are removed (edit, not deletion).

**Must not change through the whole migration:**

- `internal/` (API handlers, middleware, services, repositories, config),
  `migrations/`, `scripts/`, `cmd/healthcheck`, `cmd/admin-bootstrap`,
  Compose service/env structure, `go.mod`/`go.sum` (no new Go dependencies;
  the `embed` stdlib import disappears with the `web` package).

## 19. Open architecture risks / decisions deferred to implementation

1. **Asset prefix**: **confirmed at foundation stage.** The Vite build emits
   fingerprinted assets under `dist/assets/` with the default `assetsDir`
   (`assets`), so the `/assets/` prefix stands as planned.
2. **`/jury` redirect**: React Router redirect vs nginx 307 — defaulted to
   router, low risk either way.
3. **CSP introduction**: valuable hardening, deliberately out of cutover
   scope; if adopted, it lands in the foundation stage with its own
   validation (nonce/hash strategy for any inline payload Vite/React may
   emit).
4. **Partial (per-route) cutover**: technically possible via exact nginx
   locations, but rejected as the default — the React bundle covers all
   routes, and mixing server-rendered and SPA pages produces double
   navigation semantics. Atomic all-routes cutover with image-level
   rollback is the recommended path; product may still request staged
   public-route-first rollout, which would need its own plan addendum.
5. **Nginx `add_header` repetition**: the inheritance caveat (§4) is the
   most likely source of a silent security-header regression; the staging
   gate must assert headers on HTML, asset and API responses alike.
6. **Vite dev proxy on non-localhost setups** (remote dev, containerized
   dev): only `localhost` topology is specified here; anything else needs
   `APP_DOMAIN`/CORS review at that time.
