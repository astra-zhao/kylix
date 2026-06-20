# Kylix v2.1+ Roadmap

> **Current Status**: v2.4.0 (2026-06-20) — Polish & Ecosystem complete  
> **Next**: v2.5.0 — Toolchain Deepening  
> **Last Updated**: 2026-06-20

---

## ✅ Completed Versions (v2.1.0 – v2.4.0)

All original v2.1-v2.3 goals were completed ahead of schedule (6 days vs. 5 months planned).

### v2.1.0 ✅ (2026-06-19) — Enhanced Type System & stdlib Phase 1

| # | Task | Status |
|---|------|--------|
| 1 | Multi-parameter generic constraints | ✅ |
| 2 | Class → Interface implementation mapping | ✅ |
| 3 | Enhanced type inference (Boolean, array, nil, not) | ✅ |
| 4 | stdlib Phase 1: `strutil` (8 fn) + `mathutil` (12 fn) | ✅ |

### v2.2.0 ✅ (2026-06-19) — Engineering Quality

| # | Task | Status |
|---|------|--------|
| 1 | GitHub Actions CI/CD (ci.yml + release.yml) | ✅ |
| 2 | Generic constraint method signature verification | ✅ |
| 3 | Project-level type checking (`CheckProject`) | ✅ |
| 4 | Incremental compilation activated (`BuildCache`) | ✅ |
| 5 | stdlib Phase 2: `arrayutil` (8 fn) + `collections` (6 fn) | ✅ |

### v2.3.0 ✅ (2026-06-19) — Developer Experience

| # | Task | Status |
|---|------|--------|
| 1 | LSP incremental synchronization (textDocumentSync 2) | ✅ |
| 2 | REPL Tab completion + `:load` + `:type` | ✅ |
| 3 | kylix test: Setup/Teardown + `--filter` | ✅ |
| 4 | i18n framework (21 codes × 2 languages) | ✅ |
| 5 | Delve debugger integration (`kylix debug`) | ✅ |
| 6 | WebAssembly backend (`--wasm` + `--tinygo`) | ✅ |

### v2.4.0 ✅ (2026-06-20) — Polish & Ecosystem

| # | Task | Status |
|---|------|--------|
| 1 | i18n fully integrated into typecheck | ✅ |
| 2 | REPL `:type` real inference (`compiler.InferType`) | ✅ |
| 3 | SetLength fixed (Go generic `__kylixSetLength[T any]`) | ✅ |
| 5 | Package manager: nested deps + `kylix.lock` | ✅ |
| 6 | stdlib Phase 3: `stringbuilder` (5 fn) + `resulttype` (6 fn) | ✅ |

---

## 📊 Cumulative Metrics (as of v2.4.0)

| Metric | Value |
|--------|-------|
| Go test packages | 13 (all passing) |
| Go-level tests | ~200+ |
| Kylix-level stdlib tests | 39 (6 modules) |
| Pure Kylix stdlib functions | 45 |
| CLI commands | 17 |
| Error codes (i18n) | 21 (Chinese + English) |
| Native build targets | 5 (linux/darwin/windows × amd64/arm64) |
| WASM targets | 2 (Go standard + TinyGo) |

---

## 🎯 v2.5.0 — Toolchain Deepening (2026-07)

**Theme**: Finish the "infrastructure ready but not fully wired" items from v2.3-v2.4.

### 1. LSP Refactoring Actions
**Priority: High** | **Effort: 1 week**

- `textDocument/rename` — rename symbol across all files
- `textDocument/codeAction` — extract function, inline variable
- Leverage existing `ReferenceWalker` for cross-file scope

**Files**: `pkg/lsp/handler_navigation.go`, new `pkg/lsp/handler_refactor.go`

### 2. `kylix doc` Code Example Extraction
**Priority: Medium** | **Effort: 3 days**

- Extract fenced code blocks (` ```pascal ... ``` `) from `//` comments
- Include in generated Markdown as runnable examples
- Auto-test extracted examples (optional)

**Files**: `pkg/docgen/docgen.go`

### 3. `kylix bench` Memory Allocation Report
**Priority: Medium** | **Effort: 3 days**

- Add `--mem` flag to report B/op + allocs/op
- Use Go runtime `ReadMemStats` in harness
- Output: `BenchmarkFib  1000000  1234 ns/op  240 B/op  3 allocs/op`

**Files**: `pkg/testrunner/runner.go`, `cmd/kylix/cmd_bench.go`

### 4. `iter` Iterator Module
**Priority: Medium** | **Effort: 5 days**

```pascal
uses iter;

// Map/Filter/Reduce on arrays
var nums := [1, 2, 3, 4, 5];
var doubled := iter.Map(nums, function(x: Integer): Integer begin result := x * 2; end);
var evens := iter.Filter(doubled, function(x: Integer): Boolean begin result := (x mod 2) = 0; end);
var sum := iter.Reduce(evens, 0, function(acc: Integer; x: Integer): Integer begin result := acc + x; end);
```

**Files**: `stdlib/src/iter.klx`, `stdlib/src/iter_test.klx`

### 5. Class Method External Definition Fix
**Priority: High** | **Effort: 3 days**

Current: methods declared in class body + defined outside generate duplicate Go methods.
Goal: support Pascal-style `function TClass.Method()` outside class body.

**Files**: `generator/generator_types.go`

---

## 🚀 v2.6.0 — Performance & Optimization (2026-08)

### 1. Parallel Compilation
**Priority: High** | **Effort: 1 week**

- Parse + generate independent units in parallel (goroutine pool)
- `sync.WaitGroup` + worker pool in `CompileProject`
- Race detector tested
- Target: 10-file project > 30% speedup

**Files**: `pkg/compiler/compiler.go`

### 2. Constant Propagation + Dead Code Elimination
**Priority: Medium** | **Effort: 5 days**

- Basic const folding: `const MAX = 5; var arr: array[0..MAX-1]` → `array[0..4]`
- Remove unreachable code after `return`/`raise`
- Skip unused local variables (suppress Go `_ = varName` hacks)

**Files**: new `pkg/compiler/optimize.go`

### 3. LSP Large File Performance Benchmark
**Priority: Low** | **Effort: 2 days**

- Synthetic 10K-line .klx file
- Measure didChange → diagnostics latency
- Add to CI as performance regression guard
- Target: < 50ms per incremental edit

**Files**: `pkg/lsp/sync_test.go`

---

## 🔮 v3.0.0 — Architecture Breakthrough (2026-Q4)

### 1. LLVM Native Backend
**Priority: High** | **Effort: 2-3 months**

Goal: Kylix → LLVM IR → native binary, bypassing Go entirely.

Benefits:
- Smaller binaries (no Go runtime)
- Better optimizations (LLVM pass pipeline)
- Direct control over code generation

Approach:
- New `pkg/llvmgen/` — AST → LLVM IR via `llvm-go` bindings
- Keep Go backend as fallback (`--backend=go`)
- Benchmark: LLVM vs Go output size + speed

### 2. Package Registry Server
**Priority: Medium** | **Effort: 1 month**

- `kylix.top/packages` — browseable package index
- `kylix publish` — upload package to registry
- Semantic versioning + dependency resolution
- Mirror caching for GitHub packages

### 3. stdlib Complete Kylix-ification (Phase 4+)
**Priority: Medium** | **Effort: 2 weeks**

Rewrite remaining Go-implemented modules in Kylix:
- `jsonutil` — JSON encode/decode
- `regex` — pattern matching (via Go FFI bridge)
- `datetime` — date/time arithmetic

Keep performance-critical parts in Go via `external` declarations.

### 4. WASI Support
**Priority: Low** | **Effort: 2 weeks**

- WASI syscall layer for WASM target
- File I/O, environment variables, clock in browser/WASI runtime
- Enables server-side WASM (Cloudflare Workers, Fastly Compute)

---

## 🐛 Remaining Known Issues

### Type System
- [ ] No variance checking for generic types (covariance/contravariance) — v2.7+
- [ ] Multi-return functions not fully integrated with type inference — v2.5

### Compiler
- [ ] Error recovery sometimes continues with invalid AST state — v2.5
- [ ] No const propagation or dead code elimination — v2.6
- [ ] Class method external definitions generate duplicate Go methods — v2.5

### Standard Library
- [ ] `TDateTime` operators (+, -) not implemented — v2.5
- [ ] `TRegex` doesn't support named capture groups — v2.6
- [ ] `jsonutil` only supports flat JSON — v3.0 (Phase 4)

### Tooling
- [ ] `kylix doc` doesn't extract inline code examples — v2.5
- [ ] `kylix bench` doesn't report memory allocations — v2.5
- [ ] LSP rename refactoring not implemented — v2.5
- [ ] LSP code actions (extract/inline) not implemented — v2.5

### Infrastructure
- [x] ~~No CI/CD pipeline~~ → ✅ v2.2
- [ ] No automated regression testing across platforms — v2.5
- [ ] Website (kylix.top) needs examples update for v2.x features — v2.5

---

## 📝 Documentation Gaps

- [ ] "Getting Started" tutorial for beginners — v2.5
- [ ] Generic constraints usage guide — v2.5
- [ ] Testing best practices guide — v2.5
- [ ] Performance optimization guide — v2.6
- [ ] Migration guide from Delphi/FreePascal — v2.6
- [ ] LSP setup for VS Code / Neovim / other editors — v2.5

---

## 🎓 Community & Ecosystem

### Short-term (v2.5)
- [ ] Publish v2.4.0 announcement
- [ ] Create Discord/Slack community
- [ ] Set up GitHub Discussions for Q&A

### Medium-term (v3.0)
- [ ] Package registry (kylix.top/packages)
- [ ] Example project gallery
- [ ] Video tutorials / screencasts

### Long-term (post-v3.0)
- [ ] Conference talks / workshops
- [ ] Corporate sponsors / foundation

---

## 📅 Updated Timeline

| Version | Target Date | Theme | Status |
|---------|-------------|-------|--------|
| v2.1.0 | ~~2026-07-15~~ | Enhanced Types & stdlib Phase 1 | ✅ 2026-06-19 |
| v2.2.0 | ~~2026-09-01~~ | Incremental Build & Tooling | ✅ 2026-06-19 |
| v2.3.0 | ~~2026-11-01~~ | LSP + REPL + Debug + WASM | ✅ 2026-06-19 |
| v2.4.0 | — | Polish & Ecosystem | ✅ 2026-06-20 |
| **v2.5.0** | **2026-07** | **Toolchain Deepening** | 📋 Plan |
| **v2.6.0** | **2026-08** | **Performance & Optimization** | 📋 Plan |
| **v3.0.0** | **2026-Q4** | **LLVM Backend + Registry** | 📋 Long-term |

---

## 🤝 Contributing

Priority areas for external contributors:
1. **Standard library modules** — pure Kylix implementations (`iter`, `sortutil`, `tree`)
2. **LSP features** — refactorings, code actions, semantic highlighting
3. **Documentation** — tutorials, guides, migration from Delphi/FPC
4. **Platform testing** — Windows, Linux, macOS, WASM

---

**License**: MIT
