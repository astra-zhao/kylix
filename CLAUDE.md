# Kylix Project Context

Kylix is a modern Pascal-to-Go transpiler. The compiler is written in Go and targets Go output.

## Current State: v1.5.0 (2026-06-14)

- Phase 6–10 complete: v2.0 核心特性全部就绪
- 增量编译：55× 加速（v1.4.0）
- LSP 实时诊断：parse 错误 + 接口验证（v1.3.2）
- stdlib Kylix 化：4 个 `.klx` 声明文件，LSP 自动加载（v1.5.0）
- 包管理器：`kylix add/install/remove`，支持 git + 本地路径（v1.5.0）
- All Go tests pass (152 tests across 8 packages)
- All source files ≤ 1000 lines (refactored in v1.2.3)
- Interface implementation validated at compile time (v1.3.0)
- Kylix-layer error reporting via //line directives (v1.3.0)
- Real Go 1.18+ generics generated for generic classes/functions (v1.3.0)

## Key Documents

- [ROADMAP.md](ROADMAP.md) — Development roadmap through v2.0.0
- [TECHNICAL_DEBT.md](TECHNICAL_DEBT.md) — Known issues and improvement backlog
- [TASKS.md](TASKS.md) — Detailed task breakdown
- [CHANGELOG.md](CHANGELOG.md) — Version history

## Architecture

- `token/token.go` — Token type definitions and keyword map
- `lexer/lexer.go` — Lexical analyzer (character → token stream)
- `ast/ast.go` — AST node definitions (interfaces + concrete types)
- `parser/parser.go` — Pratt parser core; `parser_decl.go` declarations; `parser_stmt.go` statements; `parser_expr.go` expressions
- `generator/generator.go` — Generator core + pre-scan; `generator_types.go` type/func codegen; `generator_stmt.go` statement codegen; `generator_expr.go` expression codegen
- `cmd/kylix/main.go` — CLI entry point
- `pkg/compiler/` — Compilation API + incremental cache
- `pkg/pkgmgr/` — Package manager (add/install/remove)
- `pkg/repl/` — Interactive REPL
- `pkg/lsp/` — Language Server Protocol
- `stdlib/` — Go standard library (web, orm, template, exceptions, etc.)
- `stdlib/klx/` — Kylix declaration files for LSP completion

## Completed Phases

### Phase 6 → v1.0.2
- String interpolation, exception types, multi-return, properties
- Nested record fix, array range fix, memory leak fix

### Phase 7 → v1.0.3
- Map type: `map[K]V` → Go `map[K]V`, auto-init
- Variant type: `variant ... end` → Go interface + struct
- Dynamic arrays: `append(arr, elem)`, `SetLength(arr, n)`
- web_fullstack.klx rewritten in proper Kylix syntax

### Phase 8 → v1.1.2
- Enum type, slice expressions, unit file system, multi-file compilation
- Self-hosted compiler passes 15/15 examples

### Phase 9 → v1.2.0
- Bootstrap verification complete

### Phase 10 → v1.3.0–v1.5.0 (v2.0 core features)
- v1.3.0: Interface validation, Kylix-layer errors, real generics
- v1.3.1: Multi-return full coverage
- v1.3.2: LSP real-time diagnostics
- v1.4.0: Incremental compilation (55× speedup)
- v1.5.0: stdlib `.klx` declarations + package manager

## Next: Phase 11 — v2.0 engineering quality

**Priority 1: Correctness fixes (see TECHNICAL_DEBT.md)**
- `CompileFile` incremental cache integration
- `topoSortWithFiles` file path alignment fix
- `GenerateBody` exception types output stability

**Priority 2: Feature gaps**
- Type checking layer (MVP: assignment type mismatch, undeclared vars, arity check)
- Package manager integration into compiler search path
- `kylix add` git install logic fix

**Priority 3: Test coverage**
- `pkg/pkgmgr` — 0 → 5+ tests
- `pkg/compiler/cache.go` — 0 → 3+ tests
- `pkg/lsp` stdlib loading — 0 → 2+ tests
- `stdlib/klx/*.klx` parseability — 0 → 1 test
- parser generics/multi-return — 0 → 3+ tests

**Target:** 60%+ test coverage on critical paths before v2.0.0 release

## Key Constraints

- Go backend stays the same (Kylix → Go → binary)
- AST nodes use classes (not variant records)
- Never commit/push without explicit user permission
- **每个源文件不超过 1000 行**：大文件按功能拆分（例如 parser_decl.go / parser_stmt.go / parser_expr.go）
- build=go build -o /tmp/kylix_bin ./cmd/kylix/ && /tmp/kylix_bin build src/token.klx src/ast.klx src/error.klx src/lexer.klx src/parser.klx src/generator.klx src/main.klx && go build -o /tmp/kylix_self . && echo "Self-hosted compiler rebuilt OK"

## Known Issues (v1.5.0)

See [TECHNICAL_DEBT.md](TECHNICAL_DEBT.md) for the complete list. Top 3 to fix first:
1. Package manager not integrated into compiler search path (2.4)
2. `topoSortWithFiles` file path alignment bug (1.2)
3. `pkg/pkgmgr` + `pkg/compiler/cache` have zero tests (3.1)
