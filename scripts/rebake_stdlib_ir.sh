#!/usr/bin/env bash
# rebake_stdlib_ir.sh — v0.9.0: 重烘 src/stdlib_ir.klx 全链路（修复 v0.6.9 重烘断裂债）。
#
# 此前 cover.klx 从未入库、/tmp/stdir_cover/cover.ll 清理后无法重烘（TECHNICAL_DEBT.md
# v0.8.0 第二条）。现在 scripts/cover.klx 入库 + 本脚本一键重烘，CI/本地均不依赖临时残留。
#
# 用法：bash scripts/rebake_stdlib_ir.sh
# 产出：src/stdlib_ir.klx（stderr 打印各段行数与签名数，预期 signatures: 139+）
#
# 步骤：
#   1. go build host 编译器 → /tmp/kylix_bin（可用 KYLIX_BIN 覆盖跳过）
#   2. host `--backend=llvm` 生成 cover + 10 个教程的 .ll 到 /tmp/stdir_cover/
#      （只需编译通过，不运行——提取器解析 .ll 文本）
#   3. extract_stdlib_ir.py 烘焙 → src/stdlib_ir.klx
#   4. 抽查签名数（预期与上一版一致；新增 stdlib 函数时该数应单调不减）
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT=/tmp/stdir_cover
TUT=$ROOT/examples/complete-tutorial
KYLIX_BIN=${KYLIX_BIN:-/tmp/kylix_bin}

if [ -z "${KYLIX_BIN:-}" ] || [ ! -x "$KYLIX_BIN" ] || [ "${REBUILD_COMPILER:-0}" = 1 ]; then
  (cd "$ROOT" && go build -o /tmp/kylix_bin ./cmd/kylix/)
  KYLIX_BIN=/tmp/kylix_bin
fi

mkdir -p "$OUT" "$OUT/src"
emit() { # emit <源.klx> <输出名> — IR 名 = 源路径 .klx→.ll，故先把源拷进 $OUT
  cp "$1" "$OUT/$2.klx"
  "$KYLIX_BIN" build --backend=llvm -o "$OUT/$2_bin" "$OUT/$2.klx" > /dev/null
  # verify 门：host llc 走 -disable-verify（v0.6.5 性能），坏 IR 会被静默放过；
  # 烘焙数据必须 verify-clean（bootstrap sweep 的 llc 不带该 flag）。
  OPT=${OPT:-/opt/homebrew/opt/llvm/bin/opt}
  "$OPT" -passes=verify "$OUT/$2.ll" -o /dev/null || {
    echo "VERIFY FAILED: $2.ll — host emitted invalid IR, aborting rebake" >&2; exit 1; }
  echo "  emitted $2.ll (verified)"
}

echo "[1/3] emitting IR..."
emit "$ROOT/scripts/cover.klx" cover
emit "$TUT/08_stdlib_utils/example36_sysutil.klx" example36_sysutil
emit "$TUT/08_stdlib_utils/example37_jsonutil.klx" example37_jsonutil
emit "$TUT/08_stdlib_utils/example38_datetime.klx" example38_datetime
emit "$TUT/08_stdlib_utils/example39_regex.klx" example39_regex
emit "$TUT/13_stdlib_phase6/example48_phase6_net_crypto_encoding.klx" example48_phase6_net_crypto_encoding
emit "$TUT/15_jwt/example50_jwt_auth.klx" example50_jwt_auth
emit "$TUT/17_database/example52_database.klx" example52_database
emit "$TUT/18_cache/example53_cache.klx" example53_cache
emit "$TUT/19_http/example54_http.klx" example54_http
emit "$TUT/20_websocket/example55_websocket.klx" example55_websocket

echo "[2/3] baking..."
python3 "$ROOT/scripts/extract_stdlib_ir.py" > "$ROOT/src/stdlib_ir.klx"

echo "[3/3] verify signatures"
grep -c 'append(StdFnNames' "$ROOT/src/stdlib_ir.klx"
echo "done: src/stdlib_ir.klx regenerated"
