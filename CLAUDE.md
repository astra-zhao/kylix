# Kylix Project Context

Kylix is a modern Pascal-to-Go transpiler. The compiler is written in Go and targets Go output.

## Current State: v1.0.3 (2026-06-05)

- Phase 6 & 7 complete: bug fixes + language capabilities (map, variant, dynamic arrays)
- 15/15 example files pass (100%)
- All Go tests pass

## Key Documents

- [ROADMAP.md](ROADMAP.md) — Full development roadmap through self-hosting
- [TASKS.md](TASKS.md) — Detailed task breakdown
- [CHANGELOG.md](CHANGELOG.md) — Version history

## Architecture

- `token/token.go` — Token type definitions and keyword map
- `lexer/lexer.go` — Lexical analyzer (character → token stream)
- `ast/ast.go` — AST node definitions (interfaces + concrete types)
- `parser/parser.go` — Pratt parser (token stream → AST)
- `generator/generator.go` — Go code generator (AST → Go source)
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

## Next: Phase 8 → v2.0.0

编写 compiler.klx — 用 Kylix 写 Kylix 编译器

## Key Constraints

- Go backend stays the same (Kylix → Go → binary)
- AST nodes use classes (not variant records)
- Never commit/push without explicit user permission
