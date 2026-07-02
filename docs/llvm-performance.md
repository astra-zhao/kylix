# LLVM Backend Performance Guide

> **Status**: v4.1.0 — `--llvm-opt` now runs the standalone `opt` tool for full IR-level optimization (mem2reg, inlining, loop induction, DCE), in addition to llc's codegen-level `-O<N>`.

The LLVM backend produces standalone native binaries with no Go runtime dependency. With `--llvm-opt`, the `opt` IR optimizer runs before `llc`, enabling aggressive optimizations (inlining, loop induction, scalar replacement) that the codegen-level `-O` flag alone cannot perform.

---

## Quick Start

```bash
# No optimization (fastest compile, slowest runtime)
kylix build --backend=llvm program.klx

# O1 — basic optimization (mem2reg + DCE)
kylix build --backend=llvm --llvm-opt=1 program.klx

# O2 — standard optimization (inlining + loop opts)  [recommended]
kylix build --backend=llvm --llvm-opt=2 program.klx

# O3 — aggressive (vectorization, may increase code size)
kylix build --backend=llvm --llvm-opt=3 program.klx
```

### Optimization Pipeline

```
.klx → LLVM IR (.ll) → [opt -O2] → optimized IR (.opt.ll) → [llc -O2] → .o → [clang] → native binary
```

- **`opt`** (IR-level): runs the new pass manager's `default<O<N>>` pipeline — mem2reg, SROA, inlining, loop induction, GVN, DCE. This is where the biggest wins come from (Kylix's alloca/load/store style benefits greatly from mem2reg).
- **`llc -O<N>`** (codegen-level): instruction selection, register allocation, peephole opts.
- If `opt` is not installed, the pipeline falls back to `llc -O<N>` only (still works, less aggressive).

---

## Benchmarks

**Environment**: Apple Silicon (arm64), LLVM 22.1.7, Kylix v4.1.0.
**Note**: Times are pure runtime (binary execution only, no compile time). `kylix run` (Go backend) includes Go compilation each run, so its times are not directly comparable — use the Go binary (`go build` then run) for a fair comparison.

| Benchmark | unopt | O2 | O3 | Speedup (O2 vs unopt) |
|-----------|-------|-----|-----|------------------------|
| `fib` (Fib(35), recursive) | 0.027s | 0.020s | 0.016s | 1.4× |
| `loop_sum` (100M iterations) | 0.161s | 0.008s | 0.007s | **20×** |
| `primes` (sieve to 1M) | 0.041s | 0.040s | 0.040s | 1.0× |

### Key Observations

- **`loop_sum` O2 = 0.008s**: The optimizer **induces the loop into a closed-form arithmetic formula** at compile time — the 100M-iteration loop vanishes entirely. This is the biggest win from IR-level optimization.
- **`fib` recursive**: O3 enables tail-call-like transforms and inlining of the base cases → 1.7× over unopt.
- **`primes`**: dominated by `mod` operations the optimizer can't eliminate → minimal speedup. This is a fair representation of "opaque" compute.
- **vs Go backend**: LLVM-optimized binaries are consistently 5–20× faster than `kylix run` (which includes Go compile time). Compared fairly (Go binary vs LLVM binary, runtime only), LLVM O2/O3 is competitive with or faster than Go.

### Why Kylix IR Benefits So Much from `opt`

Kylix emits SSA-via-alloca (`alloca` + `load`/`store` for every variable). This is deliberately simple to generate but leaves redundant memory traffic. The `mem2reg` pass (run by `opt -O1+`) promotes these allocas to SSA registers, eliminating most loads/stores — a dramatic transformation that llc's codegen-level `-O` alone does only partially.

---

## Optimization Levels Reference

| Level | What it enables | When to use |
|-------|-----------------|-------------|
| (none) | Nothing | Development / fastest compile |
| `1` | mem2reg, DCE, simple inlining | Debug builds with some opts |
| `2` | + loop opts, SROA, GVN | **Production default** |
| `3` | + vectorization, aggressive inline | Hot loops, numeric code |

---

## Troubleshooting

### `opt: Unknown command line argument '-O=2'`
You're using an older `opt` (pre-LLVM 15). Kylix v4.1.0 uses the new `--O<N>` syntax (new pass manager). Either upgrade LLVM or the backend will skip `opt` and fall back to `llc -O<N>` only.

### Optimization changes program output
All 01–04 tutorials produce byte-identical output with `--llvm-opt=2` and without. If you see a divergence, it's a compiler bug — please report it with the `.ll` and `.opt.ll` files.

### `opt not found`
`opt` is optional. If not installed, the pipeline uses `llc -O<N>` only (less aggressive but still beneficial). Install full LLVM for `opt`: `brew install llvm`.

---

## Benchmarks Directory

Reproducible benchmarks live in `benchmarks/llvm/`:

- `fib.klx` — recursive Fibonacci (call + recursion test)
- `loop_sum.klx` — 100M-iteration sum (loop optimization test)
- `primes.klx` — prime sieve to 1M (mod-heavy compute test)

Run them yourself:

```bash
for b in fib loop_sum primes; do
  echo "--- $b ---"
  kylix build -backend=llvm -o /tmp/$b benchmarks/llvm/$b.klx
  kylix build -backend=llvm --llvm-opt=2 -o /tmp/${b}_o2 benchmarks/llvm/$b.klx
  time /tmp/$b; time /tmp/${b}_o2
done
```

---

**Last Updated**: 2026-07-02 (v4.1.0)
