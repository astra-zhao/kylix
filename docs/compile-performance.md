# Compile-Time Performance (v6.0.0 #7)

> **Status**: v6.0.0 — first compile-time benchmark for the Kylix compiler.
> Target: the bootstrap sources (`src/*.klx`, **7 files / 7451 lines**), the
> largest real-world Kylix codebase, compiled on both backends with the
> incremental cache both cold and warm.

## Results

**Environment**: Apple Silicon (arm64, M-series), Go 1.25, LLVM 22 (llc + clang + opt).
**Method**: `benchmarks/compile_time.sh` — 3 wall-clock rounds per scenario, median reported. All artifacts produced in a throwaway temp dir.

| 场景 | 第1次 | 第2次 | 第3次 | 中位数 |
|---|---|---|---|---|
| Go 冷编译（无缓存） | 31ms | 29ms | 28ms | **29ms** |
| Go 热编译（增量缓存） | 23ms | 23ms | 22ms | **23ms** |
| LLVM -O0 | 11452ms | 11527ms | 11857ms | **11527ms** |
| LLVM -O2 | 580ms | 575ms | 570ms | **575ms** |

## Key Observations

- **Go 后端编译 7451 行只需 ~29ms** — 它只生成 Go 代码，不做 `go build`。增量缓存命中时 23ms（缓存提升约 1.3×）。
- **LLVM -O0 需要 11.5s** — 生成完整 LLVM IR + `llc` 汇编。alloca/load/store 风格的 IR（每个变量都 alloca）未经优化直接交给 llc，代码量大。
- **LLVM -O2 反而只要 575ms（-O0 的 1/20）** — `opt --O2` 先把 IR 做 mem2reg/inlining/DCE 优化，IR 体积大幅缩小，`llc` 处理优化后的小 IR 反而快得多。这与运行时相反（runtime 里 -O2 快），编译期 -O2 也快。
- **LLVM 编译时间 vs Go 后端**：-O0 约 400×、-O2 约 20×。LLVM 后端的目标是"产出自包含原生二进制、无 Go 依赖"，编译时间换运行时间 + 部署形态，属预期权衡。

## v6.1.0 顺带修复（-O2 才可用的前置条件）

`--llvm-opt=2` 走 `opt` 优化，其 IR 校验比 `llc -O0` 严格。此前 bootstrap 编译在 -O0 下勉强通过（且从未验证 -O2），本轮修复两个类型错误后 **-O0 / -O2 均稳定产出二进制**：

1. **`arr := nil`（动态数组赋 nil）**：RHS 是 null ptr，但 slice 槽需 store `{ptr null, i64 0, i64 0}` 零值 — 修复见 `stmt.go` emitAssign。
2. **`FloatToStr(x)` builtin 缺失**：bootstrap 里当作普通函数 → 返回 i64（截断），接口方法实参类型错配 — 修复见 `expr.go` emitFloatToStr（snprintf `%.17g`）。

## v6.0.0 #8: -O2 正确性验证

51 教程以 `--llvm-opt=2` 全部编译 + 运行通过（51/51），且每个教程的 stdout 与 `-O0` 产物**逐字节一致**——`-O2` 不破坏正确性。前置条件是 v6.1.0 修复的两个 IR 类型错误（`arr := nil` / `FloatToStr`），否则 `opt` 校验直接拒绝 IR。

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
- `CHANGELOG.md` v6.0.0 — benchmark results + the two LLVM type fixes.

---

**Last Updated**: 2026-08-10 (v6.0.0)
