#!/bin/bash
# bundle_llvm.sh — 捆绑 LLVM 工具链到 kylix 二进制旁（分发形态 B，v0.6.3）。
#
# 产出: <dest>/llvm/bin/{llc,clang,opt} + <dest>/llvm/lib/*.dylib|*.so
# 之后 kylix 的 FindLLVM 会优先使用可执行文件旁的 llvm/bin（自带编译器，
# 不依赖目标机装 LLVM）。
#
# Usage:
#   bash scripts/bundle_llvm.sh [dest_dir]     # 默认当前目录
#
# 说明:
#   - macOS: 复制 llc/opt/clang + libLLVM.dylib/libclang-cpp.dylib，并给工具
#     add_rpath @executable_path/../lib（找不到时提示 DYLD_LIBRARY_PATH）。
#   - Linux: 复制 llc/opt/clang + libLLVM-*.so/libclang-cpp*.so，运行时需
#     LD_LIBRARY_PATH=<dest>/llvm/lib（或 ldconfig 配置）。
set -euo pipefail

DEST=${1:-.}
BIN="$DEST/llvm/bin"
LIB="$DEST/llvm/lib"
mkdir -p "$BIN" "$LIB"

# 定位 LLVM 安装目录（优先 PATH，其次 Homebrew / 常见路径）。
LLC=$(command -v llc 2>/dev/null || true)
if [ -z "$LLC" ] && [ -x /opt/homebrew/opt/llvm/bin/llc ]; then
  LLVM_BIN=/opt/homebrew/opt/llvm/bin
elif [ -z "$LLC" ] && [ -x /usr/local/opt/llvm/bin/llc ]; then
  LLVM_BIN=/usr/local/opt/llvm/bin
else
  LLVM_BIN=$(dirname "$LLC")
fi
echo "> LLVM source bin: $LLVM_BIN"

# 复制工具：只捆绑编译核心 llc/opt（+ clang 仅当 PATH 里就是 Homebrew LLVM）。
# 不捆绑 clang 的依赖 dylib 链（libclang-cpp → libLLVM → … 的 @rpath 处理在
# 各平台差异大），链接步骤用系统 clang（macOS 有 /usr/bin/clang，Linux apt clang）。
cp "$LLVM_BIN/llc" "$BIN/"
[ -x "$LLVM_BIN/opt" ] && cp "$LLVM_BIN/opt" "$BIN/"

# 复制依赖库。
if [ "$(uname)" = "Darwin" ]; then
  for lib in libLLVM.dylib libclang-cpp.dylib libc++.1.dylib libc++abi.dylib; do
    find /opt/homebrew/opt/llvm/lib /usr/local/opt/llvm/lib -maxdepth 1 -name "$lib" \
      -exec cp {} "$LIB/" \; 2>/dev/null || true
  done
  # 让复制的工具在 ./llvm/lib 找 dylib（@executable_path/../lib 即 llvm/lib）。
  for t in "$BIN"/llc "$BIN"/opt "$BIN"/clang*; do
    [ -x "$t" ] || continue
    install_name_tool -add_rpath "@executable_path/../lib" "$t" 2>/dev/null || true
  done
else
  find /usr/lib /usr/local/lib -maxdepth 2 \( -name 'libLLVM-*.so*' -o -name 'libclang-cpp*.so*' \) 2>/dev/null \
    | head -6 | xargs -r -I{} cp {} "$LIB/" 2>/dev/null || true
  echo "> Linux: 运行 kylix 前请 export LD_LIBRARY_PATH=$LIB"
fi

echo "✓ bundled LLVM to $BIN (+ libs in $LIB)"
echo "  现在 <dest>/kylix 可自包含运行（FindLLVM 优先用 llvm/bin）。"
