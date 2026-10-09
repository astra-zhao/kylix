#!/bin/bash
# Download the pinned SQLite amalgamation into third_party/sqlite/.
# Android db links this file. It is not committed (sqlite3.c is about 9MB).
# iOS uses the system libsqlite3.tbd and does not need this script.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DEST="$ROOT/third_party/sqlite"
VER="3460100"
ZIP="sqlite-amalgamation-${VER}.zip"
URL="https://www.sqlite.org/2024/${ZIP}"
SHA="77823cb110929c2bcb0f5d48e4833b5c59a8a6e40cdea3936b99e199dbbe5784"

mkdir -p "$DEST"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

curl -fL --retry 3 -o "$TMP/$ZIP" "$URL"
echo "$SHA  $TMP/$ZIP" | sha256sum -c -
unzip -q -o "$TMP/$ZIP" -d "$TMP"
cp "$TMP/sqlite-amalgamation-${VER}/sqlite3.c" "$DEST/sqlite3.c"
cp "$TMP/sqlite-amalgamation-${VER}/sqlite3.h" "$DEST/sqlite3.h"
echo "wrote $DEST/sqlite3.c"
