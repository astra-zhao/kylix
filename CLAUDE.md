# Kylix Project Context

Kylix is a modern Pascal-to-Go transpiler. The compiler is written in Go and targets Go output.

## Current State: v1.3.0 (2026-06-13)

- Phase 6–9 complete: bootstrap verified, self-hosted compiler passes 15/15 examples
- All Go tests pass (parser: 25, generator: 15, compiler: 5)
- All source files ≤ 1000 lines (refactored in v1.2.3)
- Interface implementation validated at compile time (v1.3.0)
- Kylix-layer error reporting via //line directives (v1.3.0)
- Real Go 1.18+ generics generated for generic classes/functions (v1.3.0)

## Key Documents

- [ROADMAP.md](ROADMAP.md) — Full development roadmap through self-hosting
- [TASKS.md](TASKS.md) — Detailed task breakdown
- [CHANGELOG.md](CHANGELOG.md) — Version history

## Architecture

- `token/token.go` — Token type definitions and keyword map
- `lexer/lexer.go` — Lexical analyzer (character → token stream)
- `ast/ast.go` — AST node definitions (interfaces + concrete types)
- `parser/parser.go` — Pratt parser core; `parser_decl.go` declarations; `parser_stmt.go` statements; `parser_expr.go` expressions
- `generator/generator.go` — Generator core + pre-scan; `generator_types.go` type/func codegen; `generator_stmt.go` statement codegen; `generator_expr.go` expression codegen
- `cmd/kylix/main.go` — CLI entry point
- `pkg/compiler/` — Compilation API
- `pkg/repl/` — Interactive REPL
- `pkg/lsp/` — Language Server Protocol
- `stdlib/` — Go standard library (web, orm, template, exceptions, etc.)

## Completed Phases

### Phase 6 → v1.0.2
- String interpolation, exception types, multi-return, properties
- Nested record fix, array range fix, memory leak fix

### Phase 7 → v1.0.3
- Map type: `map[K]V` → Go `map[K]V`, auto-init
- Variant type: `variant ... end` → Go interface + struct
- Dynamic arrays: `append(arr, elem)`, `SetLength(arr, n)`
- web_fullstack.klx rewritten in proper Kylix syntax

## Next: v2.0.0 — Production-grade self-hosted compiler

- `kylix build --target=<os>/<arch>` 跨平台交叉编译
- 基础类型错误诊断（Kylix 层报错，不依赖 Go 编译器错误信息）
- 泛型代码生成真实实现（当前为空壳）
- 接口实现编译时验证（`implements` 子句）
- stdlib 逐步迁移为 Kylix 源码

## Key Constraints

- Go backend stays the same (Kylix → Go → binary)
- AST nodes use classes (not variant records)
- Never commit/push without explicit user permission
- **每个源文件不超过 1000 行**：大文件按功能拆分（例如 parser_decl.go / parser_stmt.go / parser_expr.go）
- build=go build -o /tmp/kylix_bin ./cmd/kylix/ && /tmp/kylix_bin build src/token.klx src/ast.klx src/error.klx src/lexer.klx src/parser.klx src/generator.klx src/main.klx && go build -o /tmp/kylix_self . && echo "Self-hosted compiler rebuilt OK"
