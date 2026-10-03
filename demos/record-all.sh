#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GIF_DIR="${SCRIPT_DIR}/gifs"

# Check for VHS
if ! command -v vhs &>/dev/null; then
  echo "❌ VHS not found. Install: brew install vhs"
  exit 1
fi

# Check for gocat binary
if ! command -v gocat &>/dev/null; then
  echo "⚠️  gocat not in PATH. Building..."
  (cd "${SCRIPT_DIR}/.." && go build -o gocat . && export PATH="${SCRIPT_DIR}/..:${PATH}")
fi

mkdir -p "${GIF_DIR}"

TAPES=(
  version
  doctor
  scan
  console
  payload
  listen-connect
  connect
  transfer
  stabilize
  session
  session-cmd
  portforward
  proxy
  websocket
  dns-tunnel
  tunnel
  sniffer
  interfaces
  convert
  mcp
  multi-listen
  script
  unix
)

TOTAL=${#TAPES[@]}
CURRENT=0
FAILED=0

echo "🎬 Recording ${TOTAL} demos..."
echo ""

for tape in "${TAPES[@]}"; do
  CURRENT=$((CURRENT + 1))
  TAPE_FILE="${SCRIPT_DIR}/${tape}.tape"

  if [[ ! -f "${TAPE_FILE}" ]]; then
    echo "⏭️  [${CURRENT}/${TOTAL}] Skipping ${tape} (tape file not found)"
    continue
  fi

  echo -n "🎥 [${CURRENT}/${TOTAL}] Recording ${tape}... "
  if vhs "${TAPE_FILE}" 2>/dev/null; then
    echo "✅"
  else
    echo "❌"
    FAILED=$((FAILED + 1))
  fi
done

echo ""
echo "═══════════════════════════════════"
echo "📊 Results: $((TOTAL - FAILED))/${TOTAL} demos recorded"
echo "📁 GIFs saved to: ${GIF_DIR}/"

if [[ ${FAILED} -gt 0 ]]; then
  echo "⚠️  ${FAILED} recordings failed"
  exit 1
fi

echo ""
echo "🎉 All demos recorded successfully!"
ls -lh "${GIF_DIR}"/*.gif 2>/dev/null || true
