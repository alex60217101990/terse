# qdf-hookc

`qdf-hookc` is a tiny (~30 line) POSIX AF_UNIX client for the qdf-hook daemon: it
connects to the daemon's unix socket, streams stdin to it, half-closes the
write side (`shutdown(SHUT_WR)`) so the daemon can read the request to EOF,
then copies the daemon's reply to stdout — the same protocol `internal/daemon`'s
`DialAndProxy` speaks, byte-for-byte. It has no runtime dependencies and is
built with `zig cc` (e.g. `zig cc -O2 -o qdf-hookc client/qc.c`), so a single
Zig toolchain can cross-compile it for every POSIX target the hook needs
without a platform-specific C toolchain. On platforms without an AF_UNIX
socket (Windows) or wherever a native binary isn't available, the Go
`qdf-hook post` path is the universal fallback — it speaks the identical
protocol in pure Go and is always correct, just marginally slower to start
than the native client.
