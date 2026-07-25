VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
GO      ?= go
LDFLAGS := -s -w -X main.appVersion=$(VERSION)
GOFLAGS := -trimpath
BIN     := bin
PREFIX  ?= $(HOME)/.local
DESTDIR := $(PREFIX)/bin
SOCK    := $(HOME)/.qdf-hook/d.sock

.PHONY: build client client-all install modernize align fix fmt tidy vet lint test bench cover check \
        docker-daemon docker-client docker-all docker-push clean

build:
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BIN)/qdf-hook ./cmd/qdf-hook

client:
	zig cc -O2 -o $(BIN)/qdf-hookc client/qc.c

# Cross-compile the native client for all Unix targets via zig cc.
client-all:
	@mkdir -p $(BIN)
	zig cc -O2 -target aarch64-macos      -o $(BIN)/qdf-hookc-darwin-arm64 client/qc.c
	zig cc -O2 -target x86_64-macos       -o $(BIN)/qdf-hookc-darwin-amd64 client/qc.c
	zig cc -O2 -target aarch64-linux-musl -o $(BIN)/qdf-hookc-linux-arm64  client/qc.c && strip $(BIN)/qdf-hookc-linux-arm64 || true
	zig cc -O2 -target x86_64-linux-musl  -o $(BIN)/qdf-hookc-linux-amd64  client/qc.c && strip $(BIN)/qdf-hookc-linux-amd64 || true

# Install qdf-hook + qdf-hookc into PREFIX/bin (default ~/.local/bin).
#
# NEVER cp over a running daemon's binary in place: the daemon executes that
# exact inode, and rewriting its pages invalidates the code signature so macOS
# SIGKILLs (Killed: 9) every subsequent exec. So this target (1) QUITs any live
# daemon, then (2) installs each binary to a temp name and `mv`s it into place —
# an atomic rename gives a fresh inode, leaving the old one untouched. On macOS
# it re-signs ad-hoc so the freshly written file is allowed to run.
install: build client
	@mkdir -p "$(DESTDIR)"
	@printf 'QUIT\n' | "$(BIN)/qdf-hookc" "$(SOCK)" 2>/dev/null || true
	@for b in qdf-hook qdf-hookc; do \
		cp "$(BIN)/$$b" "$(DESTDIR)/.$$b.new.$$$$" && chmod +x "$(DESTDIR)/.$$b.new.$$$$" && \
		mv -f "$(DESTDIR)/.$$b.new.$$$$" "$(DESTDIR)/$$b" && \
		if [ "$$(uname)" = Darwin ]; then codesign --force -s - "$(DESTDIR)/$$b" >/dev/null 2>&1 || true; fi && \
		echo "installed $(DESTDIR)/$$b"; \
	done
	@echo "Run 'qdf-hook init' (upgrades hook wiring), then restart Claude Code. The daemon self-restarts via SessionStart."

modernize:
	$(GO) run golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize@latest --fix ./...

align:
	$(GO) run github.com/dkorunic/betteralign/cmd/betteralign@latest -apply ./...

fix:
	$(GO) fix ./...

fmt:
	$(GO) run mvdan.cc/gofumpt@latest -w .
	$(GO) run golang.org/x/tools/cmd/goimports@latest -w .

tidy:
	$(GO) mod tidy

vet:
	$(GO) vet ./...

lint:
	$(GO) run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run --timeout=5m

test:
	CGO_ENABLED=0 $(GO) test -race -count=1 -timeout=10m -coverprofile=coverage.out -covermode=atomic ./...

bench:
	$(GO) test -bench=. -benchmem -count=12 ./...

cover: test
	$(GO) tool cover -html=coverage.out -o coverage.html

# Non-mutating CI gate: modernize/align run in report mode and fail on a diff.
check: fmt vet lint
	$(GO) run golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize@latest ./... || (echo "modernize found changes"; exit 1)
	$(GO) run github.com/dkorunic/betteralign/cmd/betteralign@latest ./... || (echo "betteralign found changes"; exit 1)
	$(MAKE) test

docker-daemon:
	docker build -f deploy/docker/daemon/Dockerfile -t qdf-hookd:$(VERSION) --build-arg VERSION=$(VERSION) .

docker-client:
	docker build -f deploy/docker/client/Dockerfile -t qdf-hookc:$(VERSION) --build-arg TARGETARCH=$$(go env GOARCH) .

docker-all: docker-daemon docker-client

docker-push: docker-all
	docker push qdf-hookd:$(VERSION)
	docker push qdf-hookc:$(VERSION)

clean:
	rm -rf $(BIN) dist coverage.out coverage.html
