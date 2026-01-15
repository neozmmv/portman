#!/usr/bin/env bash
set -e

REPO="neozmmv/portman"
BIN_NAME="portman"
INSTALL_DIR="/usr/local/bin"

LATEST="stable"
URL="https://github.com/$REPO/releases/download/$LATEST/$BIN_NAME"

echo "Installing $BIN_NAME from $LATEST"
curl -fL "$URL" -o "/tmp/$BIN_NAME"
chmod +x "/tmp/$BIN_NAME"

sudo mv "/tmp/$BIN_NAME" "$INSTALL_DIR/$BIN_NAME"
sudo chmod +x "$INSTALL_DIR/$BIN_NAME"

echo "Installed at $INSTALL_DIR/$BIN_NAME"
echo "Run: sudo portman"
