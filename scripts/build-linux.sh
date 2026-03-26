#!/usr/bin/env bash
set -euo pipefail

if ! command -v go >/dev/null 2>&1; then
  echo "error: Go is not installed. Install Go first: https://go.dev/doc/install" >&2
  exit 1
fi

if ! command -v gcc >/dev/null 2>&1; then
  echo "error: gcc is required for CGO (sqlite). Install build tools first." >&2
  echo "ubuntu/debian: sudo apt update && sudo apt install -y build-essential" >&2
  exit 1
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

echo "==> Building lost for linux/amd64..."
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=gcc go build -ldflags="-s -w" -o lost-linux-amd64 ./cmd

echo "==> Build complete: $ROOT_DIR/lost-linux-amd64"

if [[ "${1:-}" == "--install" ]]; then
  echo "==> Installing to /usr/local/bin/lost (may require sudo)..."
  sudo install -m 0755 lost-linux-amd64 /usr/local/bin/lost
  echo "==> Installed. Try: lost --help"
else
  echo "==> To install globally: ./scripts/build-linux.sh --install"
  echo "==> To run directly: ./lost-linux-amd64 --help"
fi
