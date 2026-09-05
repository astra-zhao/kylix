# Compile-Time Performance (v0.6.0 #7, updated for v0.6.5)

> **Status**: v0.6.0 — first compile-time benchmark for the Kylix compiler;
> updated for the v0.6.5 LLVM performance work.
> Target: the bootstrap sources (`src/*.klx`, now **9 files** incl. `llvmgen.klx` + `stdlib_ir.klx`; 7 files / 7451 lines at v0.6.0), the
> largest real-world Kylix codebase, compiled on both backends with the
> incremental cache both cold and warm.

## Results

**Environment**: Apple Silicon (arm64, M-series), Go 1.25, LLVM 22 (llc + clang + opt).
**Method**: `benchmarks/compile_time.sh` — 3 wall-clock rounds per scenario, median reported (v0.6.0). v0.6.5 numbers are single-run measurements on the same machine.

| 场景 | v0.6.0 中位数 | v0.6.5 实测 |
|---|---|---|
| Go 冷编译（无缓存） | **29ms** | ~30ms |
| Go 热编译（增量缓存） | **23ms** | ~23ms |
| LLVM -O0 | **11527ms** | **~380ms** |
| LLVM -O0（缓存命中） | — | **~250ms** |
| LLVM -O2 | **575ms** | ~300ms（冷） / ~40ms（缓存命中） |

## Key Observations

- **Go 后端编译 7451 行只需 ~29ms** — 它只生成 Go 代码，不做 `go build`。增量缓存命中时 23ms（缓存提升约 1.3×）。
- **LLVM -O0 从 11.5s 降到 ~0.38s（v0.6.5，约 30×）** — 三个改动叠加：
  1. 进程内 DCE 从"每个死 def 全模块扫描"改成**单遍 token 计数**（O(defs × len) → O(len)）；
  2. `-O0` 默认先跑 **`opt --passes=mem2reg`**（alloca→SSA 提升）——这正是 v0.6.0 里 `-O2` 比 `-O0` 快 20× 的机制（IR 缩小 → llc 快），现在 `-O0` 也享受；
  3. `.o` 缓存查询**提前到 opt/llc 之前**，热构建直接跳过两者。
- **LLVM -O2 热构建从 575ms 降到 ~40ms** — 缓存命中时只做 parse + IR 生成 + 链接，跳过 `opt` 和 `llc`。
- **LLVM 编译时间 vs Go 后端**：现在 `-O0` 约 13×、`-O2` 热构建约 1.7×（v0.6.5）。LLVM 后端的目标是"产出自包含原生二进制、无 Go 依赖"，编译时间换运行时间 + 部署形态，属预期权衡。

## v0.6.1 顺带修复（-O2 才可用的前置条件）

`--llvm-opt=2` 走 `opt` 优化，其 IR 校验比 `llc -O0` 严格。此前 bootstrap 编译在 -O0 下勉强通过（且从未验证 -O2），本轮修复两个类型错误后 **-O0 / -O2 均稳定产出二进制**：

1. **`arr := nil`（动态数组赋 nil）**：RHS 是 null ptr，但 slice 槽需 store `{ptr null, i64 0, i64 0}` 零值 — 修复见 `stmt.go` emitAssign。
2. **`FloatToStr(x)` builtin 缺失**：bootstrap 里当作普通函数 → 返回 i64（截断），接口方法实参类型错配 — 修复见 `expr.go` emitFloatToStr（snprintf `%.17g`）。

## v0.6.0 #8: -O2 正确性验证

51 教程以 `--llvm-opt=2` 全部编译 + 运行通过（51/51），且每个教程的 stdout 与 `-O0` 产物**逐字节一致**——`-O2` 不破坏正确性。前置条件是 v0.6.1 修复的两个 IR 类型错误（`arr := nil` / `FloatToStr`），否则 `opt` 校验直接拒绝 IR。

```bash
# Verify -O2 doesn't change any tutorial's output:
OUTDIR=/tmp/o0 bash examples/complete-tutorial/test_all_llvm.sh
OUTDIR=/tmp/o2 LLVM_OPT=2 bash examples/complete-tutorial/test_all_llvm.sh
diff -r /tmp/o0 /tmp/o2    # → identical
```

## Reproduce

```bash
# CLI-level timing (single build, includes cache hit report)
kylix build --time src/*.klx

# Full 3-round benchmark table (Go cold/warm + LLVM -O0/-O2)
bash benchmarks/compile_time.sh
```

## Related

- [LLVM Backend Performance Guide](llvm-performance.md) — runtime performance (`benchmarks/llvm/`), optimization levels reference.
- `CHANGELOG.md` v0.6.0 — benchmark results + the two LLVM type fixes.

---

**Last Updated**: 2026-09-04 (v0.6.9)
