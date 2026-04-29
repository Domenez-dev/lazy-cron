#!/usr/bin/env bash
# lazy-cron installer
#
# Usage (one-liner):
#   bash -c "$(curl -fsSL https://raw.githubusercontent.com/domenez-dev/lazy-cron/main/install.sh)"
#
# Options:
#   PREFIX=/usr/local bash -c "$(curl ...)"   install to /usr/local/bin (default)
#   PREFIX=/usr       bash -c "$(curl ...)"   install to /usr/bin

set -euo pipefail

REPO="domenez-dev/lazy-cron"
BINARY="lazy-cron"
PREFIX="${PREFIX:-/usr/local}"
BIN_DIR="${PREFIX}/bin"

# Colours
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
BOLD='\033[1m'
RESET='\033[0m'

info()    { echo -e "${CYAN}==>${RESET} ${BOLD}$*${RESET}"; }
success() { echo -e "${GREEN}==>${RESET} $*"; }
error()   { echo -e "${RED}error:${RESET} $*" >&2; exit 1; }

# Detect OS
if [[ "$(uname -s)" != "Linux" ]]; then
  error "lazy-cron only supports Linux."
fi

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  *)       error "Unsupported architecture: $ARCH" ;;
esac

# Detect package manager for .deb / .rpm installs (optional, falls back to binary)
detect_pkg_manager() {
  if command -v dpkg &>/dev/null; then echo "deb"
  elif command -v rpm &>/dev/null; then echo "rpm"
  else echo "binary"
  fi
}

PKG_TYPE="$(detect_pkg_manager)"

# Get latest version from GitHub API
info "Fetching latest release..."
LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\(.*\)".*/\1/')

if [[ -z "$LATEST" ]]; then
  error "Could not fetch latest release. Check your internet connection."
fi

info "Latest version: ${LATEST}"

BASE_URL="https://github.com/${REPO}/releases/download/${LATEST}"
VERSION="${LATEST#v}"

install_deb() {
  local FILE="${BINARY}_${VERSION}_${ARCH}.deb"
  local URL="${BASE_URL}/${FILE}"
  local TMP=$(mktemp /tmp/lazy-cron-XXXXXX.deb)
  info "Downloading ${FILE}..."
  curl -fsSL "$URL" -o "$TMP"
  info "Installing .deb package (may ask for sudo)..."
  sudo dpkg -i "$TMP"
  rm -f "$TMP"
}

install_rpm() {
  local FILE="${BINARY}-${VERSION}-1.x86_64.rpm"
  local URL="${BASE_URL}/${FILE}"
  local TMP=$(mktemp /tmp/lazy-cron-XXXXXX.rpm)
  info "Downloading ${FILE}..."
  curl -fsSL "$URL" -o "$TMP"
  info "Installing .rpm package (may ask for sudo)..."
  sudo rpm -i "$TMP" 2>/dev/null || sudo rpm -U "$TMP"
  rm -f "$TMP"
}

install_binary() {
  local ARCHIVE="${BINARY}-${LATEST}-linux-${ARCH}.tar.gz"
  local URL="${BASE_URL}/${ARCHIVE}"
  local TMP=$(mktemp -d /tmp/lazy-cron-XXXXXX)
  info "Downloading ${ARCHIVE}..."
  curl -fsSL "$URL" -o "${TMP}/${ARCHIVE}"
  tar -xzf "${TMP}/${ARCHIVE}" -C "$TMP"
  info "Installing to ${BIN_DIR} (may ask for sudo)..."
  sudo mkdir -p "$BIN_DIR"
  sudo install -m755 "${TMP}/${BINARY}-${LATEST}-linux-${ARCH}" "${BIN_DIR}/${BINARY}"
  rm -rf "$TMP"
}

case "$PKG_TYPE" in
  deb)    install_deb    ;;
  rpm)    install_rpm    ;;
  binary) install_binary ;;
esac

# Verify
if command -v "$BINARY" &>/dev/null; then
  success "lazy-cron ${LATEST} installed successfully!"
  echo -e "  Run: ${BOLD}lazy-cron${RESET}"
else
  # Binary install but not in PATH
  success "lazy-cron installed to ${BIN_DIR}/${BINARY}"
  echo -e "  Make sure ${BIN_DIR} is in your PATH, then run: ${BOLD}lazy-cron${RESET}"
fi
