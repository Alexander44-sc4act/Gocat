#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GIF_DIR="${SCRIPT_DIR}/gifs"

if ! command -v vhs &>/dev/null; then
  echo "VHS not found. Install: go install github.com/charmbracelet/vhs@latest (needs ttyd + ffmpeg)"
  exit 1
fi

if [ ! -x /tmp/gcdemo/gocat ]; then
  echo "Building gocat demo binary..."
  mkdir -p /tmp/gcdemo
  (cd "${SCRIPT_DIR}/.." && go build -o /tmp/gcdemo/gocat .)
fi

mkdir -p "${GIF_DIR}"

TAPES=(
  version
  console
  scan
  payload
  listen-connect
  session
  transfer
  stabilize
)

FAILED=0
for tape in "${TAPES[@]}"; do
  echo -n "Recording ${tape}... "
  if (cd "${SCRIPT_DIR}/.." && vhs "demos/${tape}.tape" 2>/dev/null); then
    echo OK
  else
    echo FAIL
    FAILED=$((FAILED + 1))
  fi
done

echo "Done: $(( ${#TAPES[@]} - FAILED ))/${#TAPES[@]} recorded into ${GIF_DIR}/"
[ "$FAILED" -eq 0 ]
