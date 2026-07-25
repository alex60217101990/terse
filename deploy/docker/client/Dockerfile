# syntax=docker/dockerfile:1
# --- build stage: zig cross-compiles the POSIX client to a static musl binary ---
# Pin to the NATIVE build platform: zig itself is a cross-compiler, so we fetch
# the zig toolchain for the BUILD host arch (BUILDARCH) and cross-compile to the
# requested target's musl triple via `-target`. That keeps the whole build off
# QEMU — an arm64 client is produced on an amd64 runner without emulation. The
# zig download is cached across runs. BUILDARCH/TARGETARCH come from buildx.
# Base pinned by digest (Scorecard Pinned-Dependencies), pulled from
# mirror.gcr.io (Google's pull-through cache of Docker Hub official images) to
# avoid Docker Hub rate limits/outages in CI.
FROM --platform=$BUILDPLATFORM mirror.gcr.io/library/alpine:3.20@sha256:d9e853e87e55526f6b2917df91a2115c36dd7c696a35be12163d44e6e2a4b6bc AS build
ARG ZIG_VERSION=0.16.0
ARG BUILDARCH
ARG TARGETARCH
RUN --mount=type=cache,target=/zig-dl \
    apk add --no-cache curl xz && \
    case "$BUILDARCH" in \
      amd64) HOSTARCH=x86_64 ;; \
      arm64) HOSTARCH=aarch64 ;; \
      *) echo "unsupported BUILDARCH=$BUILDARCH" >&2; exit 1 ;; \
    esac && \
    case "$TARGETARCH" in \
      amd64) echo "x86_64-linux-musl"  > /musl-target ;; \
      arm64) echo "aarch64-linux-musl" > /musl-target ;; \
      *) echo "unsupported TARGETARCH=$TARGETARCH" >&2; exit 1 ;; \
    esac && \
    tarball="zig-${HOSTARCH}-linux-${ZIG_VERSION}.tar.xz" && \
    if [ ! -f "/zig-dl/${tarball}" ]; then \
      curl -fsSL "https://ziglang.org/download/${ZIG_VERSION}/${tarball}" -o "/zig-dl/${tarball}"; \
    fi && \
    tar -xJ -C /opt -f "/zig-dl/${tarball}" && mv /opt/zig-* /opt/zig
COPY client/qc.c /src/qc.c
RUN mkdir -p /out && \
    /opt/zig/zig cc -O2 -target "$(cat /musl-target)" -o /out/qdf-hookc /src/qc.c

# --- final stage ---
FROM scratch
LABEL org.opencontainers.image.title="qdf-hookc" \
      org.opencontainers.image.description="Native qdf-hook socket client" \
      org.opencontainers.image.source="https://github.com/alex60217101990/terse"
COPY --from=build /out/qdf-hookc /qdf-hookc
ENTRYPOINT ["/qdf-hookc"]
