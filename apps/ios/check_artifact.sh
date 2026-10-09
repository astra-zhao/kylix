#!/usr/bin/env bash
# check_artifact.sh — Mach-O archive symbols, then an SDK link smoke.
#
#   bash apps/ios/check_artifact.sh path/to/libkylixcore.a simulator
#   bash apps/ios/check_artifact.sh path/to/libkylixcore.a device
#
# simulator: nm must list every apps/shared/mobile_exports.list name, then
# xcrun --sdk iphonesimulator clang links a tiny C main into an arm64
# Mach-O whose LC_BUILD_VERSION platform is IOSSIMULATOR.
# device: the same, against the iphoneos SDK (platform IOS). That binary
# is not signed and is not installed on a phone. Real-device login is the
# manual checklist in docs/MOBILE_APPS.md.
#
# Refuses to run off Darwin. CI calls this on a macos runner.
set -euo pipefail

ARCHIVE="${1:-}"
KIND="${2:-}"
if [ -z "$ARCHIVE" ] || [ -z "$KIND" ]; then
  echo "usage: $0 <libkylixcore.a> simulator|device" >&2
  exit 2
fi
if [ "$(uname -s)" != "Darwin" ]; then
  echo "iOS artifact checks must run on macOS with Xcode command line tools." >&2
  exit 1
fi
if [ ! -f "$ARCHIVE" ]; then
  echo "missing $ARCHIVE" >&2
  exit 1
fi

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
LIST="$ROOT/apps/shared/mobile_exports.list"

case "$KIND" in
  simulator|sim)
    SDK=iphonesimulator
    TARGET=arm64-apple-ios16.0.0-simulator
    PLATFORM=IOSSIMULATOR
    ;;
  device)
    SDK=iphoneos
    TARGET=arm64-apple-ios16.0.0
    PLATFORM=IOS
    ;;
  *) echo "usage: $0 <archive> simulator|device" >&2; exit 2 ;;
esac

desc=$(file "$ARCHIVE")
echo "$desc"
if ! printf '%s\n' "$desc" | grep -q "ar archive"; then
  echo "expected an ar archive" >&2
  exit 1
fi

if command -v llvm-nm >/dev/null 2>&1; then
  NM=llvm-nm
else
  NM=nm
fi
blob=$("$NM" "$ARCHIVE")
echo "$blob" | grep -E 'mc_|kylix_free' || true

while IFS= read -r sym; do
  case "$sym" in
    ''|\#*) continue ;;
  esac
  if ! printf '%s\n' "$blob" | grep -E -q "(^|[^A-Za-z0-9])_?${sym}(\$|[^A-Za-z0-9_])"; then
    echo "missing archive symbol ${sym}" >&2
    exit 1
  fi
done < "$LIST"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
cat > "$tmp/smoke.c" <<'EOF'
extern const char *mc_login_path(void);
extern const char *mc_notes_path(void);
extern const char *mc_logout_path(void);
extern void kylix_free(void *);
int main(void) {
    const char *login = mc_login_path();
    const char *notes = mc_notes_path();
    const char *logout = mc_logout_path();
    if (login == 0 || notes == 0 || logout == 0) {
        return 2;
    }
    kylix_free((void *)login);
    kylix_free((void *)notes);
    kylix_free((void *)logout);
    return 0;
}
EOF

sdkroot=$(xcrun --sdk "$SDK" --show-sdk-path)
xcrun --sdk "$SDK" clang -arch arm64 --target="$TARGET" -isysroot "$sdkroot" \
  -o "$tmp/smoke" "$tmp/smoke.c" "$ARCHIVE"
linked=$(file "$tmp/smoke")
echo "$linked"
if ! printf '%s\n' "$linked" | grep -q "Mach-O 64-bit executable arm64"; then
  echo "link smoke did not produce an arm64 Mach-O executable" >&2
  exit 1
fi
build=$(vtool -show-build "$tmp/smoke")
echo "$build"
if [ "$PLATFORM" = "IOSSIMULATOR" ]; then
  if ! printf '%s\n' "$build" | grep -E -q "platform IOSSIMULATOR([[:space:]]|$)"; then
    echo "expected platform IOSSIMULATOR" >&2
    exit 1
  fi
else
  if printf '%s\n' "$build" | grep -q "IOSSIMULATOR"; then
    echo "device link produced a simulator binary" >&2
    exit 1
  fi
  if ! printf '%s\n' "$build" | grep -E -q "platform IOS([[:space:]]|$)"; then
    echo "expected platform IOS" >&2
    exit 1
  fi
fi

echo "ios artifact OK ($KIND)"
