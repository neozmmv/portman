#!/usr/bin/env bash
set -e

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

LATEST="$(curl -fsSL https://api.github.com/repos/$REPO/releases/latest | grep tag_name | cut -d '"' -f 4)"

URL="https://github.com/$REPO/releases/download/$LATEST/$ASSET"

echo "Installing $BIN_NAME ($ASSET) from $LATEST"

curl -fL "$URL" -o "/tmp/$BIN_NAME"
chmod +x "/tmp/$BIN_NAME"

sudo mv "/tmp/$BIN_NAME" "$INSTALL_DIR/$BIN_NAME"
sudo chmod +x "$INSTALL_DIR/$BIN_NAME"

echo "Installed at $INSTALL_DIR/$BIN_NAME"
echo "Run: sudo portman"
