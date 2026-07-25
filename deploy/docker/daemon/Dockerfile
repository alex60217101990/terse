# syntax=docker/dockerfile:1
# --- build stage ---
# Pin the builder to the NATIVE build platform and cross-compile with Go's
# GOOS/GOARCH (CGO is off), so an arm64 image is built on an amd64 runner at
# native speed instead of under QEMU emulation. BUILDPLATFORM/TARGETOS/
# TARGETARCH are provided automatically by buildx. The cache mounts persist
# the module and build caches across runs (paired with GHA cache in CI).
FROM --platform=$BUILDPLATFORM golang:1.26 AS build
ARG VERSION=dev
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath \
      -ldflags="-s -w -X main.appVersion=${VERSION}" \
      -o /out/qdf-hook ./cmd/qdf-hook

# --- final stage ---
FROM scratch
LABEL org.opencontainers.image.title="qdf-hookd" \
      org.opencontainers.image.description="Resident qdf-hook daemon" \
      org.opencontainers.image.source="https://github.com/alex60217101990/terse"
COPY --from=build /out/qdf-hook /qdf-hook
USER 65532:65532
ENV HOME=/data
VOLUME ["/data"]
HEALTHCHECK --interval=30s --timeout=3s CMD ["/qdf-hook", "daemon", "--ping"]
ENTRYPOINT ["/qdf-hook"]
CMD ["daemon", "--serve"]
