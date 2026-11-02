#!/usr/bin/env bash
# Download the baton CLI from the baton-sdk release that matches go.mod.
# The CLI must read the c1z format that the connector writes, so its version
# has to match the baton-sdk version. Output: ./baton in the current directory.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
version="$(awk '$1 == "github.com/conductorone/baton-sdk" {print $2}' "$root/go.mod" | head -n 1)"
[ -n "$version" ] || { echo "baton-sdk version not found in go.mod" >&2; exit 1; }

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$arch" in
  x86_64) arch=amd64 ;;
  aarch64) arch=arm64 ;;
esac

base="https://github.com/conductorone/baton-sdk/releases/download/${version}"
archive="baton-${version}-${os}-${arch}.tar.gz"
checksums="baton_${version#v}_checksums.txt"

curl -fsSLO "${base}/${archive}"
curl -fsSLO "${base}/${checksums}"
if command -v sha256sum >/dev/null; then sum=(sha256sum -c -); else sum=(shasum -a 256 -c -); fi
grep " ${archive}\$" "${checksums}" | "${sum[@]}"
tar xzf "${archive}" baton
rm -f "${archive}" "${checksums}"
./baton --version
