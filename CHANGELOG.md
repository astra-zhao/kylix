# Changelog

All notable changes to the Kylix compiler are documented in this file.

## v1.0.1 (2026-06-03)

### Bug Fixes

**P0 - Critical (Infinite Loops & Crashes):**
- **`inherits` keyword silently ignored**: `class Dog inherits Animal` now correctly sets the parent class and generates Go struct embedding
- **Anonymous procedure/function parsing**: `procedure()` and `function()` are now parsed as expressions. Support for local declarations (var, const, type) in anonymous functions
- **Match wildcard `_` generates invalid Go**: `_ => body` now correctly generates `default:`
- **Match multi-pattern and `when` guard**: `2, 3 =>` and `when condition =>` now correctly parsed
- **`{ }` block comment conflict**: Removed `{...}` as Pascal comment syntax (conflicted with match block braces). Only `//` and `(* *)` are recognized
- **Case statement infinite loop**: Fixed missing `nextToken()` after `parseExpression` in case values
- **While/for loop parsing**: Fixed missing `nextToken()` after condition and From/To expressions
- **Function type as parameter**: `function Apply(fn: function(Integer): Integer)` no longer infinite loops
- **Array range syntax**: `array[0..2] of Integer` now correctly parses
- **Consecutive `//` comments**: Multiple line comments no longer cause parse errors
- **Parameter parsing**: Added iteration guard to prevent infinite loop

**P1 - High Priority:**
- **Constructor code generation**: `Dog.Create(args)` now generates `&Dog{args}`
- **Match statement import scanning**: Built-ins inside match branches now trigger Go imports
- **`match` keyword as identifier**: `match` can now be used as variable/field name
- **`result` keyword as identifier**: `result` can now be used as variable/field name
- **`try/except` with `begin...end` blocks**: `except begin...end` now correctly parsed
- **`finally` without `begin`**: Bare statements in finally block now supported

### Files Changed

- `lexer/lexer.go` — Removed `{}` comment syntax, fixed consecutive comment lines
- `parser/parser.go` — 12 parser fixes across match, case, while, for, try, array, function type parsing
- `generator/generator.go` — Match multi-pattern, wildcard, constructor generation
- `ast/ast.go` — Added `AdditionalPatterns` to MatchBranch

### Example File Status (14 files)

| ✅ Passing (11) | ❌ Failing (3) |
|---|---|
| hello, simple, types, control, classes | functions (multi-return — feature gap) |
| modern, exceptions, stdlib_demo | web_demo (3 errors — anon proc edge cases) |
| test_formatter, web_advanced, orm_example | web_fullstack (12 errors — Go syntax in examples) |

### Version Bumps

| Component | Old | New |
|-----------|-----|-----|
| Compiler, REPL, LSP, Project, VSCode | 1.0.0 | **1.0.1** |

### Known Issues (v1.0.2+)

| Priority | Issue |
|----------|-------|
| P1 | String interpolation broken (lexer→parser→generator) |
| P1 | Exception types not defined in Go runtime |
| P2 | Multi-value return `(Real, Boolean)` not supported |
| P2 | Properties silently dropped in code generation |
| P2 | No multi-file compilation |
| P2 | Map/dictionary type not supported |
| P2 | No lexer/parser/generator unit tests |
| P3 | 18 tokens defined but unhandled |
| P3 | LSP code actions are stubs |
| P3 | REPL no uses/class declaration detection |

---

## v1.0.0 (2026-06-01)

**🎉 First stable release**

This release marks the completion of all 5 planned phases. Kylix is now a full-featured modern Pascal compiler targeting Go.

### New Standard Library Modules

| Module | `uses` | Description |
|--------|--------|-------------|
| `sysutil` | `uses sysutil` | File I/O, directory operations, path utilities, environment variables |
| `jsonutil` | `uses jsonutil` | JSON encode/decode, type-safe accessors, file I/O |
| `datetime` | `uses datetime` | Date/time creation, arithmetic, formatting, parsing, comparisons |
| `regex` | `uses regex` | Pattern matching, find/replace, split, email/URL/numeric validators |

### Language Features (Phase 4)

- **Generic type parameters** — declare generics on classes and functions:
  ```pascal
  type TPair<T1, T2> = class ... end;
  function CreatePair<T>(x: T; y: T): TPair<T, T>;
  ```
- **Exception handling ON clause** — typed exception catching:
  ```pascal
  try
    raise Exception.Create('error');
  except
    on E: Exception do WriteLn(E.Message);
  end;
  ```
- **Constructor / Destructor / Inherited** keywords
- **Lambda expression parameter parsing** — `(x: Integer) -> x * x`
- **Async/Await** code generation (goroutine + channel pattern)

### Standard Library (Phase 3)

- **Web Framework** — HTTP server with routing (GET/POST/PUT/DELETE), path parameters, middleware chain, JSON/HTML responses
- **DI Container** — Singleton, transient, and scoped lifetimes
- **Configuration** — Auto-config from JSON files + environment variables with priority layering
- **Middleware Suite** — CORS, authentication, rate limiting, request ID, logging
- **Request Validation** — Required fields, min/max length, email, pattern, range checks
- **ORM** — MySQL, PostgreSQL, SQLite support with query builder and migrations
- **Template Engine** — Layouts, partials, custom functions, caching
- **Auto-Configuration** — Multi-source config loading with environment detection

### Tooling Improvements (Phase 5)

- **REPL**:
  - Added `github.com/peterh/liner` for readline support
  - Persistent command history (saved to `~/.kylix_repl_history`)
  - ↑/↓ arrow keys for history navigation
  - Lexer-based `isCompleteStatement` detection (replaced fragile string heuristics)
  - Separate `errOut` writer — stderr goes to `os.Stderr`, not merged with stdout
  - Ctrl-C cancels multiline input cleanly
- **Formatter**:
  - `formatClassDecl` now outputs visibility modifiers (`public`, `private`, `protected`)
  - `formatClassDecl` now iterates and outputs `Properties`
  - `formatConstDecl` outputs type annotation when present
  - Added `token` package import for visibility constants
- **Generator**: Added stdlib import mappings for `sysutil`, `jsonutil`, `datetime`, `regex`

### Version Bumps

| Component | Old | New |
|-----------|-----|-----|
| Compiler (`cmd/kylix/main.go`) | 0.2.0 | **1.0.0** |
| REPL (`pkg/repl/repl.go`) | 0.3.0 | **1.0.0** |
| LSP Server (`pkg/lsp/server.go`) | 0.3.0 | **1.0.0** |
| Project Config (`pkg/project/project.go`) | 0.1.0 | **1.0.0** |
| VS Code Extension (`vscode-ext/package.json`) | 0.2.0 | **1.0.0** |

### Files Added

- `stdlib/sysutil.go` — File I/O and system utilities (~220 lines)
- `stdlib/jsonutil.go` — JSON encoding/decoding (~155 lines)
- `stdlib/datetime.go` — Date and time operations (~230 lines)
- `stdlib/regex.go` — Regular expression utilities (~180 lines)
- `stdlib/stdlib_new_test.go` — 32 new tests for all four modules
- `examples/stdlib_demo.klx` — Stdlib demo program
- `CHANGELOG.md` — This file

### Tests

- 32 new stdlib tests — all passing
- Full test suite: `go test ./...` — all packages pass

---

## v0.3.0 (2026-05-31)

### Phase 4: Language Enhancements

- Generic type parameter declarations (classes and functions)
- Exception handling with ON clause (`on E: ExceptionType do`)
- Constructor/destructor/inherited keywords
- Lambda expression parameter parsing
- Async/await code generation improvements (goroutine + channel)
- Updated formatter for new syntax (generics, ON clause)
- Updated generator for type parameters and exception type-switch

### Phase 3: Web Framework

- HTTP server based on Go `net/http`
- Routing: GET, POST, PUT, DELETE with path parameters (`/users/:id`)
- Middleware support (logger, CORS, auth, rate limit, request ID)
- JSON request/response handling
- Static file serving
- Anonymous procedures and functions
- DI container, config system, validation, ORM, template engine, auto-config
- VS Code extension with syntax highlighting, snippets, completions

---

## v0.2.0 (2026-05-30)

### Phase 2: IDE Toolchain

- CLI toolchain: `new`, `build`, `run`, `check`, `fmt`, `repl`, `lsp`, `version`
- Project management with `kylix.toml`
- LSP server with code completion and hover documentation
- VS Code extension with syntax highlighting
- Interactive REPL with multiline support and session persistence
- Comprehensive documentation (user manual, developer guide, tools explained)

---

## v0.1.0 (2026-05-29)

### Phase 1: Compiler Core

- Lexer with full Pascal token support (comments, strings, operators)
- Pratt parser with correct operator precedence
- Complete AST node definitions
- Go code generator with builtin function mapping (WriteLn → fmt.Println, etc.)
- Type mapping (Integer → int64, Real → float64, String → string)
- Language features:
  - Variables, constants, type declarations
  - Functions and procedures
  - Control structures (if, while, for, case, repeat)
  - Records and arrays
  - Classes and interfaces
  - Properties with getters/setters
  - Type inference (`var x := 42;`)
  - Lambda expressions
  - Pattern matching (`match value { ... }`)
  - Async/Await
  - ForEach loops
  - String interpolation
  - Exception handling (try/except/finally)
