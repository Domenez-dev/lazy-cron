#!/usr/bin/env bash
# release.sh — build and package lazy-cron for a new release
#
# Usage:
#   ./release.sh v0.1.0
#
# What it produces in dist/:
#   lazy-cron-v0.1.0-linux-amd64.tar.gz
#   lazy-cron-v0.1.0-linux-arm64.tar.gz
#   lazy-cron_v0.1.0_amd64.deb
#   lazy-cron_v0.1.0_arm64.deb
#   lazy-cron-v0.1.0-1.x86_64.rpm
#   checksums.sha256
#
# Requirements:
#   go, nfpm (yay -S nfpm)
#
# Upload the dist/ contents to a GitHub release manually or with gh:
#   gh release create v0.1.0 dist/* --title "v0.1.0" --notes "Release notes here"

set -euo pipefail

VERSION="${1:-}"
if [[ -z "$VERSION" ]]; then
  echo "Usage: $0 <version>   (e.g. $0 v0.1.0)"
  exit 1
fi

# Strip leading 'v' for package version fields
PKG_VERSION="${VERSION#v}"

MODULE="github.com/domenez-dev/lazy-cron"
BINARY="lazy-cron"
CMD_PATH="./cmd/lazy-cron"
DIST="dist"
LDFLAGS="-s -w -X ${MODULE}/internal/styles.AppVersion=${PKG_VERSION}"

echo "==> Building lazy-cron ${VERSION}"
rm -rf "$DIST"
mkdir -p "$DIST"

build() {
  local GOARCH=$1
  local OUT="${DIST}/${BINARY}-${VERSION}-linux-${GOARCH}"
  echo "    linux/${GOARCH}"
  GOOS=linux GOARCH="$GOARCH" CGO_ENABLED=0 go build \
    -ldflags "$LDFLAGS" \
    -o "$OUT" \
    "$CMD_PATH"
}

build amd64
build arm64

echo "==> Creating archives"
for ARCH in amd64 arm64; do
  BIN="${DIST}/${BINARY}-${VERSION}-linux-${ARCH}"
  ARCHIVE="${DIST}/${BINARY}-${VERSION}-linux-${ARCH}.tar.gz"
  tar -czf "$ARCHIVE" -C "$DIST" "$(basename "$BIN")" \
    -C "$(pwd)" README.md LICENSE
  echo "    $ARCHIVE"
done

echo "==> Building .deb and .rpm packages (requires nfpm)"
if command -v nfpm &>/dev/null; then
  for ARCH in amd64 arm64; do
    # Point nfpm at the right binary
    NFPM_ARCH="$ARCH"
    BIN_PATH="${DIST}/${BINARY}-${VERSION}-linux-${ARCH}"

    # Write a temp nfpm config with correct version and binary path
    TMPCONF=$(mktemp /tmp/nfpm-XXXXXX.yaml)
    sed \
      -e "s|__VERSION__|${PKG_VERSION}|g" \
      -e "s|__BIN__|${BIN_PATH}|g" \
      -e "s|__ARCH__|${NFPM_ARCH}|g" \
      packaging/nfpm.yaml > "$TMPCONF"

    nfpm package --packager deb --config "$TMPCONF" --target "$DIST/"
    if [[ "$ARCH" == "amd64" ]]; then
      nfpm package --packager rpm --config "$TMPCONF" --target "$DIST/"
    fi
    rm "$TMPCONF"
  done
else
  echo "    nfpm not found — skipping .deb/.rpm (yay -S nfpm)"
fi

echo "==> Checksums"
cd "$DIST"
sha256sum ./* > checksums.sha256
cat checksums.sha256
cd ..

echo ""
echo "Done. Files in ${DIST}/:"
ls -lh "$DIST/"
echo ""
echo "To publish:"
echo "  git tag ${VERSION} && git push origin ${VERSION}"
echo "  gh release create ${VERSION} ${DIST}/* --title \"${VERSION}\" --generate-notes"
echo "  (or upload dist/ files manually on github.com)"
