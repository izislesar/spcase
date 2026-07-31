# syntax=docker/dockerfile:1.7

FROM node:24-bookworm-slim@sha256:6f7b03f7c2c8e2e784dcf9295400527b9b1270fd37b7e9a7285cf83b6951452d AS frontend-build

WORKDIR /workspace

COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci --no-audit --no-fund

COPY tailwind.config.js ./
COPY web/src ./web/src
COPY web/template ./web/template
RUN mkdir -p web/static/css web/static/js && npm run build


FROM golang:1.26.5-bookworm@sha256:1ecb7edf62a0408027bd5729dfd6b1b8766e578e8df93995b225dfd0944eb651 AS go-build

WORKDIR /workspace

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
COPY --from=frontend-build /workspace/web/static ./web/static

ARG TARGETOS=linux
ARG TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS="${TARGETOS}" GOARCH="${TARGETARCH}" \
    go build -buildvcs=false -trimpath -ldflags="-s -w" -o /out/spcase ./cmd/app && \
    CGO_ENABLED=0 GOOS="${TARGETOS}" GOARCH="${TARGETARCH}" \
    go build -buildvcs=false -trimpath -ldflags="-s -w" -o /out/healthcheck ./cmd/healthcheck && \
    CGO_ENABLED=0 GOOS="${TARGETOS}" GOARCH="${TARGETARCH}" \
    go build -buildvcs=false -trimpath -ldflags="-s -w" -o /out/admin-bootstrap ./cmd/admin-bootstrap


FROM golang:1.26.5-bookworm@sha256:1ecb7edf62a0408027bd5729dfd6b1b8766e578e8df93995b225dfd0944eb651 AS migrator-build

ARG GOOSE_VERSION=v3.27.2
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOBIN=/out \
    go install \
    -tags="no_clickhouse no_libsql no_mssql no_mysql no_sqlite3 no_vertica no_ydb" \
    github.com/pressly/goose/v3/cmd/goose@${GOOSE_VERSION}


FROM debian:bookworm-slim@sha256:7b140f374b289a7c2befc338f42ebe6441b7ea838a042bbd5acbfca6ec875818 AS migrator

RUN apt-get update && \
    apt-get install --no-install-recommends -y ca-certificates make && \
    rm -rf /var/lib/apt/lists/* && \
    groupadd --gid 65532 nonroot && \
    useradd --uid 65532 --gid 65532 --no-create-home --shell /usr/sbin/nologin nonroot

WORKDIR /workspace

COPY --from=migrator-build /out/goose /usr/local/bin/goose
COPY --chown=nonroot:nonroot Makefile ./
COPY --chown=nonroot:nonroot scripts/migrate-production.sh ./scripts/migrate-production.sh
COPY --chown=nonroot:nonroot migrations ./migrations

USER nonroot:nonroot

ENTRYPOINT ["make", "migrate-production"]


FROM nginx:1.30.0-alpine-slim@sha256:2fb5d772cea6ef1a8dab525df1b9485289eee167d26af9613fce27a12c060caa AS nginx

COPY nginx.conf /etc/nginx/nginx.conf
COPY --from=frontend-build --chown=101:101 /workspace/web/static /usr/share/nginx/html/static

USER 101:101
EXPOSE 8080
STOPSIGNAL SIGQUIT
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["wget", "-q", "--spider", "http://127.0.0.1:8080/api/v1/health/ready"]

ENTRYPOINT ["nginx"]
CMD ["-g", "daemon off;"]


FROM gcr.io/distroless/static-debian12:nonroot@sha256:f5b485ea962d9bd1186b2f6b3a061191539b905b82ec395de78cbfae51f20e35 AS runtime

WORKDIR /app

COPY --from=go-build --chown=nonroot:nonroot /out/spcase /app/spcase
COPY --from=go-build --chown=nonroot:nonroot /out/healthcheck /app/healthcheck
COPY --from=go-build --chown=nonroot:nonroot /out/admin-bootstrap /app/admin-bootstrap

ENV PORT=8000

USER nonroot:nonroot
EXPOSE 8000
STOPSIGNAL SIGTERM
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/app/healthcheck"]

ENTRYPOINT ["/app/spcase"]
