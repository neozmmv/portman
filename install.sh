#!/usr/bin/env bash
set -euo pipefail

REPO="neozmmv/portman"
BIN_NAME="portman"
INSTALL_DIR="/usr/local/bin"

ARCH="$(uname -m)"

case "$ARCH" in
  x86_64)
    ASSET="portman-linux-amd64"
    ;;
  aarch64|arm64)
    ASSET="portman-linux-arm64"
    ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

VERSION="${PORTMAN_VERSION:-latest}"
if [[ "$VERSION" == "latest" ]]; then
  URL="https://github.com/$REPO/releases/latest/download/$ASSET"
else
  URL="https://github.com/$REPO/releases/download/$VERSION/$ASSET"
fi

TMP_BIN="$(mktemp -t portman.XXXXXX)"

echo "Installing $BIN_NAME ($ASSET) from $VERSION"
curl -fsSL "$URL" -o "$TMP_BIN"
chmod +x "$TMP_BIN"

sudo mv "$TMP_BIN" "$INSTALL_DIR/$BIN_NAME"
sudo chmod +x "$INSTALL_DIR/$BIN_NAME"

echo "Installed at $INSTALL_DIR/$BIN_NAME"
echo "Run: sudo portman"

if command -v "$BIN_NAME" >/dev/null 2>&1; then
  "$BIN_NAME" version || true
fi
