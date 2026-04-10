#!/usr/bin/env sh
# ajolote installer
# Usage: curl -fsSL https://raw.githubusercontent.com/ajolote-ai/ajolote/main/install.sh | sh

set -e

REPO="ajolote-ai/ajolote"
BINARY="ajolote"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# ── detect OS and arch ────────────────────────────────────────────────────────

OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Linux)   os="linux" ;;
  Darwin)  os="darwin" ;;
  *)       echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

case "$ARCH" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

# ── resolve latest version ────────────────────────────────────────────────────

if [ -z "$VERSION" ]; then
  VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' \
    | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')"
fi

if [ -z "$VERSION" ]; then
  echo "Could not determine latest release. Set VERSION manually:" >&2
  echo "  VERSION=v0.1.0 sh install.sh" >&2
  exit 1
fi

# ── download and install ──────────────────────────────────────────────────────

ARCHIVE="${BINARY}_${os}_${arch}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "Downloading ajolote ${VERSION} (${os}/${arch})..."
curl -fsSL "$URL" -o "${TMP}/${ARCHIVE}"

tar -xzf "${TMP}/${ARCHIVE}" -C "$TMP"

# Check if install dir is writable; if not, use sudo
if [ -w "$INSTALL_DIR" ]; then
  mv "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
else
  echo "Installing to ${INSTALL_DIR} (may require sudo)..."
  sudo mv "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
fi

chmod +x "${INSTALL_DIR}/${BINARY}"

echo ""
echo "ajolote ${VERSION} installed to ${INSTALL_DIR}/${BINARY}"
echo ""
echo "Get started:"
echo "  cd your-project"
echo "  ajolote init"
