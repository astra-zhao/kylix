#!/bin/bash
# test_bootstrap_all.sh — v0.6.9 P3/#44: run every tutorial through the
# bootstrap compiler's no-Go pipeline (--emit-llvm → llc → clang → run) and
# diff the output against the host-compiled binary.
#
# Usage:
#   KYLIX=/tmp/kylix_bin BOOT=/tmp/main_self_p2_dbg \
#     bash scripts/test_bootstrap_all.sh
#
# Link flags are chosen per program by scanning the emitted IR (mirrors host
# compile.go): -lcrypto for OpenSSL symbols, -lsqlite3 for db, -lcurl for
# httpclient. Server-ish tutorials get a timeout so a blocking accept can't
# hang the sweep; their output is compared on what was printed before exit.

KYLIX=${KYLIX:-/tmp/kylix_bin}
BOOT=${BOOT:-/tmp/main_self_p2_dbg}
LLC=${LLC:-/opt/homebrew/opt/llvm/bin/llc}
TUT=${TUT:-examples/complete-tutorial}
OUT=${OUT:-/tmp/boot_tut}
TIMEOUT=${TIMEOUT:-15}

# Known bootstrap-emitter limitations (v0.6.9 P3), counted separately:
#  - example15_lambda:  anonymous procedure/function literals (lambda values)
#                       are not lowerable yet (lambda var calls go undefined)
#  - example50_jwt_auth: one alloca's use escapes its block (domination error
#                        in the jwt claims path) — IR verify fails
KNOWN_FAILS=" example15_lambda example50_jwt_auth "

mkdir -p "$OUT"
PASS=0; FAIL=0; SKIP=0; KNOWN=0

for f in "$TUT"/*/*.klx; do
  name=$(basename "$f" .klx)
  # unit files (no main statements) are not runnable programs — they are
  # exercised via the multi-file test below (math_helper + example33)
  if [ "$name" == "math_helper" ]; then
    continue
  fi
  # ---- host reference ----
  if ! "$KYLIX" build --backend=llvm -o "$OUT/${name}_host" "$f" >/dev/null 2>&1; then
    echo "SKIP $name (host build failed)"
    SKIP=$((SKIP+1)); continue
  fi
  host_out=$(timeout "$TIMEOUT" "$OUT/${name}_host" 2>/dev/null </dev/null)
  host_rc=$?

  # ---- bootstrap pipeline ----
  if ! "$BOOT" --emit-llvm "$f" > "$OUT/$name.ll" 2>"$OUT/$name.emit.err"; then
    echo "FAIL $name (bootstrap emit)"; head -3 "$OUT/$name.emit.err" | sed 's/^/    /'
    FAIL=$((FAIL+1)); continue
  fi
  if ! "$LLC" -filetype=obj "$OUT/$name.ll" -o "$OUT/$name.o" 2>"$OUT/$name.llc.err"; then
    echo "FAIL $name (llc)"; head -4 "$OUT/$name.llc.err" | sed 's/^/    /'
    FAIL=$((FAIL+1)); continue
  fi
  extra=""
  grep -q "@__kylix_crypto_\|@EVP_\|@SHA256\|@MD5\|@PKCS5" "$OUT/$name.ll" && \
    extra="$extra -L/opt/homebrew/opt/openssl@3/lib -lcrypto -Wl,-rpath,/opt/homebrew/opt/openssl@3/lib"
  grep -q "@sqlite3_" "$OUT/$name.ll" && extra="$extra -lsqlite3"
  grep -q "@curl_" "$OUT/$name.ll" && \
    extra="$extra -L/opt/homebrew/opt/curl/lib -lcurl -Wl,-rpath,/opt/homebrew/opt/curl/lib"
  if ! clang "$OUT/$name.o" -o "$OUT/${name}_boot" $extra 2>"$OUT/$name.clang.err"; then
    echo "FAIL $name (link)"; head -4 "$OUT/$name.clang.err" | sed 's/^/    /'
    FAIL=$((FAIL+1)); continue
  fi
  boot_out=$(timeout "$TIMEOUT" "$OUT/${name}_boot" 2>/dev/null </dev/null)
  boot_rc=$?

  if [ "$boot_out" == "$host_out" ]; then
    echo "PASS $name"
    PASS=$((PASS+1))
  elif [[ "$KNOWN_FAILS" == *" $name "* ]]; then
    echo "KNOWN $name (limitation, see script header)"
    KNOWN=$((KNOWN+1))
  else
    echo "DIFF $name (host_rc=$host_rc boot_rc=$boot_rc)"
    diff <(echo "$host_out") <(echo "$boot_out") | head -8 | sed 's/^/    /'
    FAIL=$((FAIL+1))
  fi
done

# ---- multi-file module test (math_helper + example33) ----
name=example33_use_module
if [ -f "$TUT/11_modules/example33_use_module.klx" ]; then
  if "$KYLIX" build --backend=llvm -o "$OUT/${name}_host" \
      "$TUT/11_modules/math_helper.klx" "$TUT/11_modules/$name.klx" >/dev/null 2>&1; then
    host_out=$("$OUT/${name}_host" 2>/dev/null)
    if "$BOOT" --emit-llvm "$TUT/11_modules/math_helper.klx" "$TUT/11_modules/$name.klx" \
        > "$OUT/$name.ll" 2>/dev/null \
        && "$LLC" -filetype=obj "$OUT/$name.ll" -o "$OUT/$name.o" \
        && clang "$OUT/$name.o" -o "$OUT/${name}_boot"; then
      boot_out=$("$OUT/${name}_boot" 2>/dev/null)
      if [ "$boot_out" == "$host_out" ]; then
        echo "PASS $name (multi-file)"; PASS=$((PASS+1))
      else
        echo "DIFF $name (multi-file)"; FAIL=$((FAIL+1))
      fi
    else
      echo "FAIL $name (multi-file pipeline)"; FAIL=$((FAIL+1))
    fi
  fi
fi

echo ""
echo "=== bootstrap pipeline: PASS=$PASS FAIL=$FAIL KNOWN=$KNOWN SKIP=$SKIP"
[ "$FAIL" -eq 0 ]
