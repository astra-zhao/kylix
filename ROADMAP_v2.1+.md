# Kylix v2.1+ Roadmap

> **Current Status**: v2.0.0 (2026-06-17) — Production-ready with complete toolchain  
> **Next Major**: v2.1.0 — Enhanced Type System & Standard Library

---

## 🎯 Short-term Goals (v2.1.0 — 2-3 weeks)

### 1. Multi-Parameter Generic Constraints
**Priority: High** | **Effort: 3 days**

Current limitation:
```pascal
type TMap<K: IComparable, V> = class  // ❌ Only validates single param
```

Goal:
- Support multi-parameter constraint validation
- Track parameter ordering from `TypeDecl.TypeParams`
- Validate each type argument against its constraint

**Files**: `pkg/compiler/typecheck.go`

---

### 2. Class → Interface Implementation Mapping
**Priority: High** | **Effort: 4 days**

Current limitation:
```pascal
type
  TMyType = class implements IComparable
    function CompareTo(): Integer;
  end;
var box: TBox<TMyType>;  // ✅ Passes (assumes custom types satisfy)
```

Goal:
- Build class → interface mapping during `collectDeclarations`
- Verify method signatures match interface requirements
- Reject classes that claim to implement but don't have matching methods

**Files**: `pkg/compiler/typecheck.go`

---

### 3. Enhanced Type Inference
**Priority: Medium** | **Effort: 3 days**

Current support:
- Literals (Integer, String, Boolean, Real)
- Function calls
- Simple infix expressions

Missing:
- Array literals: `var arr := [1, 2, 3]` → infer `array of Integer`
- Map literals: `var m := {'a': 1}` → infer `map[String]Integer`
- Lambda expressions: `var f := (x: Integer) -> x * 2` → infer function type
- Ternary: `var x := if cond then 1 else 0` → infer Integer

**Files**: `pkg/compiler/typecheck.go` (`inferExprType`)

---

### 4. Standard Library Kylix-ification
**Priority: High** | **Effort: 3+ weeks**

Current state:
- 4 stdlib modules have `.klx` declarations (`sysutil`, `datetime`, `regex`, `jsonutil`)
- Core runtime still in Go (`stdlib/*.go`)

Goal — Phase 1 (v2.1):
- Rewrite `strutil` in Kylix (string manipulation, no external deps)
- Rewrite `mathutil` in Kylix (common math functions)
- Keep performance-critical parts in Go (JSON parsing, regex, file I/O)

Goal — Phase 2 (v2.2+):
- Self-host more stdlib modules
- Performance comparison: Kylix vs Go implementation

**Files**: `stdlib/klx/`, new `stdlib/src/` for Kylix implementations

---

## 🚀 Medium-term Goals (v2.2.0 — 1-2 months)

### 5. Incremental Compilation
**Priority: Medium** | **Effort: 1 week**

Current behavior:
- Full recompilation on every `kylix build`
- Slow for large projects

Goal:
- Track file modification times
- Cache `.go` output and ASTs
- Only recompile changed files + dependents
- Store build cache in `.kylix/cache/`

**Files**: `pkg/compiler/cache.go` (exists, needs expansion)

---

### 6. Package-level Type Checking
**Priority: Medium** | **Effort: 5 days**

Current limitation:
- Cross-file type checking only via LSP
- CLI `kylix check` only validates single files

Goal:
- `kylix check .` checks all `.klx` in project
- Resolve cross-file dependencies (uses clauses)
- Report undeclared types from other units

**Files**: `pkg/compiler/`, `cmd/kylix/cmd_check.go`

---

### 7. Debugger Integration
**Priority: Low** | **Effort: 2 weeks**

Goal:
- Generate Go code with `//line` directives (already done)
- Integrate with Delve (Go debugger)
- `kylix debug program.klx` launches debugger
- Breakpoints map back to `.klx` source lines

**Files**: New `pkg/debugger/`, `cmd/kylix/cmd_debug.go`

---

### 8. REPL Enhancements
**Priority: Low** | **Effort: 3 days**

Current REPL:
- Basic expression evaluation
- No multi-line input
- No persistent state

Goal:
- Multi-line mode (detect incomplete statements)
- Save/load REPL sessions
- Autocomplete via LSP
- Syntax highlighting

**Files**: `cmd/kylix/cmd_repl.go`

---

## 🔮 Long-term Vision (v3.0+ — 3+ months)

### 9. WebAssembly Target
**Priority: Low** | **Effort: 1 month**

Goal:
- Compile Kylix → WASM instead of Go
- Run Kylix in browser
- Enable interactive playground on kylix.top

**Approach**:
- Option A: Kylix → Go → TinyGo → WASM
- Option B: New WASM backend in `generator/`

---

### 10. Native Code Generation
**Priority: Low** | **Effort: 3+ months**

Current:
- Kylix → Go → native (via Go compiler)
- Indirect, but leverages Go's excellent codegen

Future:
- Direct LLVM backend
- Kylix → LLVM IR → native
- Better control over optimizations

---

### 11. Language Server Protocol Extensions
**Priority: Medium** | **Effort: 2 weeks**

Current LSP features:
- Completion, hover, diagnostics, signature help
- Go to definition, find references

Missing:
- Rename refactoring (use LSP rename, not just CLI)
- Code actions (extract function, inline variable)
- Semantic highlighting
- Inlay hints (show inferred types)

**Files**: `pkg/lsp/`

---

## 📊 Priority Matrix

| Feature | Priority | Effort | Impact | v2.1? |
|---------|----------|--------|--------|-------|
| Multi-param generics | High | 3d | High | ✅ |
| Class→Interface mapping | High | 4d | High | ✅ |
| Enhanced type inference | Medium | 3d | Medium | ✅ |
| stdlib Kylix-ification | High | 3w | High | 🔄 Phase 1 |
| Incremental compilation | Medium | 1w | High | ❌ v2.2 |
| Package-level checking | Medium | 5d | Medium | ❌ v2.2 |
| Debugger | Low | 2w | Medium | ❌ v2.2 |
| REPL enhancements | Low | 3d | Low | ❌ v2.2 |
| WASM target | Low | 1m | Low | ❌ v3.0+ |
| Native codegen | Low | 3m+ | Medium | ❌ v3.0+ |
| LSP extensions | Medium | 2w | Medium | ❌ v2.2 |

---

## 🐛 Known Issues & Technical Debt

### Type System
- [ ] Circular type alias detection doesn't report all cycles in complex graphs
- [ ] Generic constraint violation messages could be more specific
- [ ] No variance checking for generic types (covariance/contravariance)

### Compiler
- [ ] Error recovery sometimes continues with invalid AST state
- [ ] Multi-return functions not fully integrated with type inference
- [ ] No const propagation or dead code elimination

### Standard Library
- [ ] `TDateTime` operators (+, -) not implemented
- [ ] `TRegex` doesn't support named capture groups
- [ ] `jsonutil` only supports flat JSON (no nested arrays/objects well)

### Tooling
- [ ] `kylix test` doesn't support setup/teardown hooks
- [ ] `kylix bench` doesn't report memory allocations
- [ ] `kylix doc` doesn't extract inline code examples from comments
- [ ] LSP occasionally loses sync on rapid edits

### Infrastructure
- [ ] No CI/CD pipeline (GitHub Actions)
- [ ] No automated regression testing across platforms
- [ ] Website (kylix.top) needs examples update for v2.0 features

---

## 📝 Documentation Gaps

- [ ] No "Getting Started" tutorial for beginners
- [ ] Generic constraints usage guide missing
- [ ] Testing best practices guide
- [ ] Performance optimization guide
- [ ] Migration guide from Delphi/FreePascal
- [ ] LSP setup for VS Code / Neovim / other editors

---

## 🎓 Community & Ecosystem

### Short-term
- [ ] Publish v2.0.0 announcement on Reddit/HN/Lobsters
- [ ] Create Discord/Slack community
- [ ] Set up GitHub Discussions for Q&A

### Medium-term
- [ ] Package registry (kylix.top/packages)
- [ ] Example project gallery
- [ ] Video tutorials / screencasts

### Long-term
- [ ] Conference talks / workshops
- [ ] Corporate sponsors / foundation
- [ ] Annual Kylix conference

---

## 🚦 Decision Points

### Should we prioritize...
1. **Compatibility** vs **Innovation**?
   - Current: Leaning toward innovation (modern features)
   - Trade-off: Some Delphi/FPC code won't compile
   
2. **Performance** vs **Simplicity**?
   - Current: Simplicity (Kylix→Go is indirect but simple)
   - Alternative: Direct LLVM backend for performance
   
3. **Self-hosting** vs **Go runtime**?
   - Current: Hybrid (compiler in Go, some stdlib in Kylix)
   - Goal: More self-hosting, but keep perf-critical parts in Go

---

## 📅 Tentative Timeline

| Version | Target Date | Theme |
|---------|-------------|-------|
| v2.1.0 | 2026-07-15 | Enhanced Types & stdlib Phase 1 |
| v2.2.0 | 2026-09-01 | Incremental Build & Tooling |
| v2.3.0 | 2026-11-01 | LSP Extensions & REPL |
| v3.0.0 | 2027-Q1 | WebAssembly / Native Codegen |

---

## 🤝 Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Code style guide
- PR submission process
- Issue triage workflow
- Testing requirements

Priority areas for external contributors:
1. Standard library modules (pure Kylix implementations)
2. LSP features (refactorings, code actions)
3. Documentation (tutorials, guides, examples)
4. Platform testing (Windows, Linux, macOS)

---

**Last Updated**: 2026-06-17  
**Maintainer**: @astra-zhao  
**License**: MIT
