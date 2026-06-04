# Kylix Project Context

Kylix is a modern Pascal-to-Go transpiler. The compiler is written in Go and targets Go output.

## Current State: v1.0.2 (2026-06-04)

- Phase 6 complete: string interpolation, exception types, multi-return, properties, nested record fix
- 13/14 example files pass (93%)
- 1 failing: web_fullstack.klx (uses Go struct literal `{...}` syntax — not valid Kylix)

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
- `stdlib/` — Go standard library (web, orm, template, etc.)

## Phase 6 Completed → v1.0.2

- String interpolation: `$"${expr}"` → `fmt.Sprintf(...)`
- Exception types: `raise/except/on` generate proper Go exception types
- Multi-value return: `: (Type1, Type2)` + tuple literals + destructuring
- Properties: `property Name: Type read Field;` → getter/setter methods
- Nested record fix: depth tracking for `end` in nested record types
- Array range fix: `[0..2]` now computes correct size `(2-0+1)`

## Next: Phase 7 → v1.1.0

Priority: Map type → Variant/Union types → Dynamic arrays → Enums → Multi-file modules

## Key Constraints

- Go backend stays the same (Kylix → Go → binary)
- AST nodes use classes (not variant records)
- Never commit/push without explicit user permission
