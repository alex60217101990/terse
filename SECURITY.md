# Security Policy

## Supported versions

The latest tagged release and the `main` branch receive security fixes.

## Reporting a vulnerability

Please report suspected vulnerabilities **privately** through GitHub Security
Advisories — the repository's **Security → Report a vulnerability** tab, or:

<https://github.com/alex60217101990/terse/security/advisories/new>

Do not open a public issue for a security report. We aim to acknowledge within a
few business days and will coordinate a fix and disclosure timeline with you.

## Scope

`qdf-hook` is a local Claude Code hook. It reads tool output on stdin and the
user's own files, and writes a **rebuildable** cache under `~/.qdf-hook`. It
performs no outbound network I/O and stores no credentials. The security-
relevant surfaces are therefore:

- **Decoding attacker/model-controlled input** — the hook parses arbitrary JSON
  from Claude Code and on-disk cache blobs; decode paths must bound allocation
  by the input and never panic on malformed data (see the native fuzz target in
  `internal/protocol`).
- **Zero-copy lifetime** — pooled buffers and `unsafe` string/byte views must
  never outlive or alias a reused/freed backing store.
- **The PreToolUse deny path** — the one place the hook answers "unchanged,
  don't re-read" without content in hand; it must never cause stale content to
  be served as authoritative.
