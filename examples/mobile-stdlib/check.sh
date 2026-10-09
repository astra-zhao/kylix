#!/bin/bash
# Link the portable stdlib probe for android or ios.
# android: NDK clang, both ABIs, plus sqlite amalgamation when sqlite3.c exists.
# ios: macOS + Xcode. simulator libc probe, simulator sqlite (-lsqlite3), device libc probe.
# Does not boot an emulator or a simulator.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
KYLIX="${KYLIX:-$ROOT/kylix}"
MODE="${1:-}"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

if [[ ! -x "$KYLIX" ]]; then
  echo "kylix binary not found at $KYLIX" >&2
  exit 1
fi

defined_in() {
  local bin="$1" sym="$2" out
  # Read nm to completion. grep -q in a pipefail script returns 141 when it
  # closes the pipe early, which fails the check even though the symbol is there.
  if [[ "$bin" == *.so ]]; then
    out="$(nm -D --defined-only "$bin")"
    grep -Eq " ${sym}$" <<< "$out"
  else
    # Mach-O C symbols carry a leading underscore.
    out="$(nm -gU "$bin")"
    grep -Eq " _?${sym}$" <<< "$out"
  fi
}

file_has() {
  local info
  info="$(file "$1")"
  grep -q "$2" <<< "$info"
}

build() {
  local target="$1" src="$2" out="$3"
  echo "build $target $src -> $out"
  "$KYLIX" build --backend=llvm --target "$target" --shared -o "$out" "$src"
}

case "$MODE" in
android)
  build android/arm64 "$ROOT/examples/mobile-stdlib/probe.klx" "$WORK/probe-arm64.so"
  file_has "$WORK/probe-arm64.so" "ELF"
  file_has "$WORK/probe-arm64.so" "ARM aarch64"
  defined_in "$WORK/probe-arm64.so" kylix_sha256
  defined_in "$WORK/probe-arm64.so" kylix_md5
  build android/amd64 "$ROOT/examples/mobile-stdlib/probe.klx" "$WORK/probe-amd64.so"
  file_has "$WORK/probe-amd64.so" "ELF"
  file_has "$WORK/probe-amd64.so" "x86-64"
  defined_in "$WORK/probe-amd64.so" kylix_sha256
  if [[ -f "$ROOT/third_party/sqlite/sqlite3.c" || -n "${KYLIX_SQLITE_SRC:-}" ]]; then
    build android/arm64 "$ROOT/examples/mobile-stdlib/dbprobe.klx" "$WORK/db-arm64.so"
    defined_in "$WORK/db-arm64.so" sqlite3_open
  else
    echo "skip android db: sqlite3.c not present (run scripts/fetch_sqlite_amalgamation.sh)"
  fi
  ;;
ios)
  build ios/simulator-arm64 "$ROOT/examples/mobile-stdlib/probe.klx" "$WORK/probe-sim.dylib"
  file_has "$WORK/probe-sim.dylib" "Mach-O"
  defined_in "$WORK/probe-sim.dylib" kylix_sha256
  defined_in "$WORK/probe-sim.dylib" kylix_md5
  build ios/simulator-arm64 "$ROOT/examples/mobile-stdlib/dbprobe.klx" "$WORK/db-sim.dylib"
  file_has "$WORK/db-sim.dylib" "Mach-O"
  build ios/arm64 "$ROOT/examples/mobile-stdlib/probe.klx" "$WORK/probe-dev.dylib"
  file_has "$WORK/probe-dev.dylib" "Mach-O"
  defined_in "$WORK/probe-dev.dylib" kylix_sha256
  ;;
*)
  echo "usage: $0 android|ios" >&2
  exit 1
  ;;
esac

echo "mobile stdlib link ok ($MODE)"
