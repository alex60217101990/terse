#!/usr/bin/env bash
set -euo pipefail

BINARY="qdf-hook"
INSTALL_DIR="${GOPATH:-$HOME/go}/bin"

echo "Building $BINARY..."
go build -ldflags="-s -w" -o "$INSTALL_DIR/$BINARY" ./cmd/qdf-hook/

echo "Installed: $INSTALL_DIR/$BINARY"
"$INSTALL_DIR/$BINARY" version

echo ""
echo "Add to ~/.claude/settings.json:"
cat <<'EOF'
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Read|Glob|Grep",
        "hooks": [{"type": "command", "command": "qdf-hook read"}]
      },
      {
        "matcher": "Bash|PowerShell",
        "hooks": [{"type": "command", "command": "qdf-hook bash"}]
      }
    ],
    "PreCompact": [
      {"matcher": ".*", "hooks": [{"type": "command", "command": "qdf-hook precompact"}]}
    ],
    "PostCompact": [
      {"matcher": ".*", "hooks": [{"type": "command", "command": "qdf-hook postcompact"}]}
    ]
  }
}
EOF
