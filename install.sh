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

resolve_latest_semver() {
  # Pick the highest vX.Y.Z tag among GitHub Releases, ignoring non-semver tags like "stable".
  # Uses GNU sort -V.
  local api="https://api.github.com/repos/${REPO}/releases?per_page=100"
  local json tags latest
  if ! json="$(curl -fsSL "$api")"; then
    return 1
  fi
  tags="$(printf '%s' "$json" | grep -oE '"tag_name"\s*:\s*"v[0-9]+\.[0-9]+\.[0-9]+"' | sed -E 's/.*"(v[0-9]+\.[0-9]+\.[0-9]+)".*/\1/' | sort -u)"
  if [[ -z "$tags" ]]; then
    return 1
  fi
  latest="$(printf '%s\n' "$tags" | sort -V | tail -n 1)"
  [[ -n "$latest" ]] || return 1
  printf '%s' "$latest"
}

if [[ "$VERSION" == "latest" ]]; then
  if VERSION="$(resolve_latest_semver)"; then
    :
  else
    # Fallback: GitHub's idea of "latest" release.
    VERSION="latest"
  fi
fi

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

"$INSTALL_DIR/$BIN_NAME" version || true
