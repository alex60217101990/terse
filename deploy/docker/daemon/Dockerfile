# syntax=docker/dockerfile:1
# --- build stage ---
FROM golang:1.26 AS build
ARG VERSION=dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath \
      -ldflags="-s -w -X main.appVersion=${VERSION}" \
      -o /out/qdf-hook ./cmd/qdf-hook

# --- final stage ---
FROM scratch
LABEL org.opencontainers.image.title="qdf-hookd" \
      org.opencontainers.image.description="Resident qdf-hook daemon" \
      org.opencontainers.image.source="https://github.com/alex60217101990/qdf-hook"
COPY --from=build /out/qdf-hook /qdf-hook
USER 65532:65532
ENV HOME=/data
VOLUME ["/data"]
HEALTHCHECK --interval=30s --timeout=3s CMD ["/qdf-hook", "daemon", "--ping"]
ENTRYPOINT ["/qdf-hook"]
CMD ["daemon", "--serve"]
