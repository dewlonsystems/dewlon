#!/bin/sh
# Installs the dewlon CLI from its GitHub releases.
#
#   curl -fsSL https://raw.githubusercontent.com/dewlonsystems/dewlon/main/scripts/install.sh | sh
#
# Environment:
#   DEWLON_VERSION      version to install, e.g. v0.2.0 (default: latest)
#   DEWLON_INSTALL_DIR  where to put the binary (default: /usr/local/bin if
#                       writable, otherwise ~/.local/bin)

set -eu

REPO="dewlonsystems/dewlon"
BIN="dewlon"

say() { printf '%s\n' "$*"; }
die() { printf 'error: %s\n' "$*" >&2; exit 1; }
need() { command -v "$1" >/dev/null 2>&1 || die "$1 is required but not installed"; }

need curl
need tar
need uname

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
  linux | darwin) ;;
  *) die "unsupported OS: $os (on Windows, use scripts/install.ps1)" ;;
esac

arch=$(uname -m)
case "$arch" in
  x86_64 | amd64) arch=amd64 ;;
  aarch64 | arm64) arch=arm64 ;;
  *) die "unsupported architecture: $arch" ;;
esac

version="${DEWLON_VERSION:-latest}"
if [ "$version" = "latest" ]; then
  # Following the redirect avoids the API and its rate limits.
  final=$(curl -fsSL -o /dev/null -w '%{url_effective}' "https://github.com/$REPO/releases/latest") ||
    die "could not look up the latest release"
  version="${final##*/}"
fi
case "$version" in
  v*) ;;
  *) version="v$version" ;;
esac
number="${version#v}"

archive="${BIN}_${number}_${os}_${arch}.tar.gz"
base="https://github.com/$REPO/releases/download/$version"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

say "Downloading $BIN $version for $os/$arch..."
curl -fsSL -o "$tmp/$archive" "$base/$archive" || die "download failed: $base/$archive"
curl -fsSL -o "$tmp/checksums.txt" "$base/checksums.txt" || die "could not download checksums.txt"

expected=$(awk -v f="$archive" '$2 == f { print $1 }' "$tmp/checksums.txt")
[ -n "$expected" ] || die "no checksum listed for $archive"

if command -v sha256sum >/dev/null 2>&1; then
  actual=$(sha256sum "$tmp/$archive" | awk '{ print $1 }')
elif command -v shasum >/dev/null 2>&1; then
  actual=$(shasum -a 256 "$tmp/$archive" | awk '{ print $1 }')
else
  die "sha256sum or shasum is needed to verify the download"
fi
[ "$expected" = "$actual" ] || die "checksum mismatch for $archive (expected $expected, got $actual)"

tar -xzf "$tmp/$archive" -C "$tmp" "$BIN"

dir="${DEWLON_INSTALL_DIR:-}"
if [ -z "$dir" ]; then
  if [ -w /usr/local/bin ]; then
    dir=/usr/local/bin
  else
    dir="$HOME/.local/bin"
  fi
fi
mkdir -p "$dir"
install -m 0755 "$tmp/$BIN" "$dir/$BIN"

say "Installed $BIN $version to $dir/$BIN"

case ":$PATH:" in
  *":$dir:"*) ;;
  *) say "Note: $dir is not on your PATH. Add it to your shell profile to run '$BIN' directly." ;;
esac