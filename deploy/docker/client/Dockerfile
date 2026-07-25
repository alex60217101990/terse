# syntax=docker/dockerfile:1
# --- build stage: zig cross-compiles the POSIX client to a static musl binary ---
FROM alpine:3.20 AS build
ARG ZIG_VERSION=0.16.0
ARG TARGETARCH
RUN apk add --no-cache curl xz && \
    case "$TARGETARCH" in \
      amd64) ZARCH=x86_64; MUSL=x86_64-linux-musl ;; \
      arm64) ZARCH=aarch64; MUSL=aarch64-linux-musl ;; \
    esac && \
    curl -fsSL "https://ziglang.org/download/${ZIG_VERSION}/zig-${ZARCH}-linux-${ZIG_VERSION}.tar.xz" \
      | tar -xJ -C /opt && mv /opt/zig-* /opt/zig && echo "$MUSL" > /musl-target
COPY client/qc.c /src/qc.c
RUN mkdir -p /out && \
    /opt/zig/zig cc -O2 -target "$(cat /musl-target)" -o /out/qdf-hookc /src/qc.c && \
    /opt/zig/zig cc --version >/dev/null

# --- final stage ---
FROM scratch
LABEL org.opencontainers.image.title="qdf-hookc" \
      org.opencontainers.image.description="Native qdf-hook socket client" \
      org.opencontainers.image.source="https://github.com/alex60217101990/qdf-hook"
COPY --from=build /out/qdf-hookc /qdf-hookc
ENTRYPOINT ["/qdf-hookc"]
