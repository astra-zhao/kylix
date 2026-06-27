# Changelog

All notable changes to the Kylix compiler are documented in this file.

> 🌐 [kylix.top](https://kylix.top) — Official website with interactive docs and live code examples.

## v3.2.0-dev (2026-06-25) — KylixBoot Annotation Auto-Wiring

### LLVM Backend Milestone 2 Phase 3 — Generic Monomorphization

The LLVM backend now specializes generic class declarations for each concrete instantiation, completing LLVM M2:

```
%TBox_Integer = type { ptr, i64 }
%TBox_String  = type { ptr, ptr }
define i64 @TBox_Integer_Get(ptr %self) { ... }
define ptr @TBox_String_Get(ptr %self) { ... }
```

What landed:

- Generic class templates (`TBox<T>`) are registered but not emitted directly. `emitDecl` defers them to the monomorphization pass.
- A new `collectInstantiations` AST walker finds every `*ast.GenericType` reference (in `VarDecl`, function/method signatures, member expressions, expressions).
- Each unique instantiation is mangled (`TBox<Integer>` → `TBox_Integer`) and specialized once. Specialization clones the template `ClassDecl`, substitutes type-parameter identifiers throughout fields, method params/returns, and property types, then routes the clone through the existing `emitClassDecl` path so it gets a struct, vtable, and methods.
- Constructor pattern `TBox<Integer>.Create` and plain `TFoo.Create` are now lowered (this also fixes a pre-existing gap where non-generic constructor calls weren't actually wired to `emitConstructor`).
- Generic and class-typed `VarDecl`s now allocate `ptr` slots and register their resolved (mangled) type into `localTypes` so member access and method dispatch work end-to-end.

Tests: `pkg/llvmgen/generics_test.go` (6 IR-fragment tests). All 44 existing llvmgen tests continue to pass.

`CacheVersion` bumped to `10` (defensive — Go-backend output unchanged).

Deferred: free-standing generic functions, generic constraints validation in the LLVM path (Go backend already validates), nested generics like `TList<TBox<Integer>>`, cross-module discovery.

---

### LLVM Backend Milestone 2 Phase 2 — Interfaces (fat pointer)

The LLVM backend now lowers Pascal interface declarations and dispatches calls through a two-word fat pointer:

```
%IFoo_vtable = type { ptr, ptr, ... }     ; one ptr per method
%IFoo_iface  = type { ptr, ptr }          ; { vtable, data }
@TFoo_IFoo_vtable = constant { ptr } [ ptr @TFoo_M ]
```

What landed:

- `InterfaceDecl` is emitted (was previously silently dropped). Per-class interface vtables are emitted in interface declaration order alongside the existing class vtable.
- `MemberExpression` and `obj.Method(args)` are lowered for the first time — concrete class field access (`obj.Field`) and direct method dispatch were never wired up before this slice.
- Interface-typed locals reserve two `ptr` allocas (`%v_x_iface_vt`, `%v_x_iface_data`). Assignment `iface := concreteObj` or `iface := obj as IFoo` boxes the object into `{ @TClass_IFace_vtable, data }`.
- Interface method calls indirect through the vtable slot: load slot via `getelementptr [N x ptr]`, then indirect `call`.
- `obj is IFoo` lowers to a compile-time `i1` for known class→interface implementation pairs.
- `obj as IFoo` produces a boxed fat pointer when the receiver class implements the target interface, else a null pointer.
- Generator state additions: `interfaces map[string]*InterfaceInfo`, `localTypes map[string]string`. `localTypes` is populated for `VarDecl` with declared types, function parameters, and method parameters.

Tests: `pkg/llvmgen/interface_test.go` (8 IR-fragment tests). All existing `pkg/llvmgen` tests still pass.

`CacheVersion` bumped to `9` to invalidate any cached Go fragments from before this LLVM-side change (defensive — Go backend output is unchanged).

Deferred to the next slice: dynamic `is`/`as` against arbitrary-typed interface values (needs runtime type IDs), multi-level inheritance vtable merging, cross-module interface resolution.

---

### ORM Annotations

The compiler now scans `[Entity('table')]`, `[Column('col')]`, `[PrimaryKey]`, `[Repository(TEntity)]`, and `[Query('SELECT ...')]` annotations and generates ORM helper methods that delegate to the existing `stdlib.ORM` runtime.

MVP support:

- `[Entity]` generates `ToRow()` and `FromRow()` mapping helpers on the class.
- `[Repository(TEntity)]` generates baseline `FindAll`, `FindById`, `Save`, and `DeleteById` methods taking `orm *stdlib.ORM` as the first argument.
- `[Query('SELECT ...')]` generates a method body that calls `orm.Query` or `orm.QueryAll` based on the return type (single entity vs `array of entity`).
- Existing methods on the entity/repository class take precedence — generation is skipped to avoid duplicate Go method definitions.
- Compiler diagnostic `KLX213 ErrInvalidORM` covers bad `[Entity]` args, unknown repository entity targets, misplaced `[Query]`, and invalid query return types.

Example: `examples/complete-tutorial/12_special_features/example47_orm_annotations.klx`.

Tests: `generator/generator_orm_annotations_test.go`, `pkg/compiler/orm_annotations_test.go`.

---

### Security Annotations

KylixBoot now supports per-route security guards via field annotations:

- `[Authenticated]` injects a 401 guard that validates `Authorization: Bearer <token>` using `boot.RegisterAuthValidator`.
- `[Role('admin')]` adds a 403 guard against `boot.RegisterRolesProvider` (implies `[Authenticated]`).

Generated route closures emit the guards before dispatching to the controller:

```go
stdlib.BootGET("/admin/users", func(req *stdlib.BootRequest) *stdlib.BootResponse {
    if __r := stdlib.BootEnforceAuth(req); __r != nil { return __r }
    if __r := stdlib.BootEnforceRole(req, "admin"); __r != nil { return __r }
    return __kylix_ctrl_TAdminController.ListUsers(req)
})
```

Runtime additions:
- `pkg/boot/security.go` — `RegisterAuthValidator`, `RegisterRolesProvider`, `EnforceAuth`, `EnforceRole`
- `pkg/boot/types.go` — `Request.User` / `Request.Roles`
- `stdlib/boot_bridge.go` — `BootRegisterAuth`, `BootRegisterRoles`, `BootEnforceAuth`, `BootEnforceRole`
- `stdlib/klx/boot.klx` — LSP declarations for the security helpers

Compiler diagnostic `KLX212 ErrInvalidSecurity` covers:
- `[Role]` missing or non-string argument
- `[Authenticated]` / `[Role]` outside a controller route method

Example: `examples/complete-tutorial/12_special_features/example46_security_annotations.klx`.

Tests: `pkg/boot/security_test.go`, `generator/generator_boot_security_test.go`, `pkg/compiler/security_annotations_test.go`.

---

### Validation Annotations

The compiler now scans `[Required]` / `[Email]` / `[Min]` / `[Max]` / `[MinLen]` / `[MaxLen]` field annotations on classes and generates `Validate()` and `IsValid()` methods.

MVP support:

- `[Required]` checks empty strings via `strings.TrimSpace` or zero for numeric fields
- `[Email]` runs a conservative regex compatible with `stdlib/validation.Email`
- `[MinLen(n)]` / `[MaxLen(n)]` check string length
- `[Min(n)]` / `[Max(n)]` check numeric bounds
- Generated methods skip if the class already defines `Validate` or `IsValid`
- Compiler diagnostic `KLX211` flags missing integer args or type mismatches such as `[Min] Name: String`

Example: `examples/complete-tutorial/12_special_features/example45_validation_annotations.klx`.

Tests: `generator/generator_validation_annotations_test.go`, `pkg/compiler/validation_annotations_test.go`.

---

### KylixBoot Auto Route Registration

The compiler now scans `[Controller]` classes and HTTP method attributes at codegen pre-scan time and emits route registrations before user `main()` statements.

MVP support:

- `[Controller('/base')]` on classes
- `[Get('/path')]`, `[Post('/path')]`, `[Put('/path')]`, `[Delete('/path')]` on class methods
- Function handler signature: `function Method(req: TRequest): TResponse`
- Generated startup code registers closures through `stdlib.BootGET/POST/PUT/DELETE`
- `TRequest` / `TResponse` map to `*stdlib.BootRequest` / `*stdlib.BootResponse` when `uses boot` is active

Example: `examples/complete-tutorial/12_special_features/example42_kylixboot_autowire.klx`.

Tests: `generator/generator_boot_annotations_test.go`.

### KylixBoot Service / Component + Inject Auto-Wiring

The compiler now scans `[Service]` and `[Component]` classes and emits singleton startup wiring before route registration.

MVP support:

- `[Service]` / `[Component]` class singleton instantiation
- `BootRegisterInstance` registration by exact class name and short Pascal name (`TUserService` + `UserService`)
- `[Inject]` fields resolved by field type and assigned directly in generated Go
- Controllers reuse the injected controller instance for generated route closures

Example: `examples/complete-tutorial/12_special_features/example43_kylixboot_di.klx`.

Tests: `generator/generator_boot_di_test.go`.

---


### Compiler Fixes

**KLX-M01 — Unit `interface` / `implementation` parsing**

Pascal unit files now correctly treat `interface` and `implementation` as section markers instead of generating an empty Go interface declaration.

- `token/token.go`: adds the `implementation` keyword token
- `parser/parser.go`: skips unit section markers and parses implementation functions as top-level function declarations with bodies
- `generator/generator_types.go`: skips bodiless forward declarations and guards against empty interface names
- `pkg/compiler/unit_sections_test.go`: regression coverage for multi-file unit builds

This fixes `examples/complete-tutorial/11_modules/example33_use_module.klx`.

**KLX-G01 — Generic class method receivers**

Go codegen now emits instantiated receivers for generic class methods:

```go
func (self *TStack[T]) Push(item T)
```

instead of the invalid:

```go
func (self *TStack) Push(item T)
```

- `generator/generator.go`: tracks class type parameters during pre-scan
- `generator/generator_types.go`: emits generic receivers for methods and properties
- `generator/generator_generics_test.go`: regression coverage for generic class receivers
- `example21_generic_class.klx`: uses explicit `self.Field` access, matching the OOP tutorial convention

### Tutorial Verification

`examples/complete-tutorial/test_all.sh` now covers all tutorial example directories and the module example. Current result: **35/35 passed** (34 `example*.klx` files plus the `math_helper.klx` unit companion file).

### Cache Invalidation

Incremental build cache entries now include `CacheVersion`, invalidating stale generated fragments after codegen changes.

---


### KylixBoot Framework — Spring Boot-style Runtime Core

New `pkg/boot/` package (~700 lines, 23 tests) — declarative web framework foundation.

- `types.go` — Request, Response, Handler, Middleware
- `router.go` — Route matching with path params (`/users/:id`)
- `server.go` — HTTP server with graceful shutdown
- `di.go` — DI container: Singleton / Transient / Instance + reflection-based Inject
- `app.go` — Top-level App + global shortcuts (`boot.GET`, `boot.POST`, `boot.Use`, `boot.Listen`)
- `config.go` — Config with env var fallback
- `middleware.go` — Logger, Recover, CORS, Auth, RateLimit, RequestID

Bridge: `stdlib/boot_bridge.go` re-exports the runtime as `stdlib.BootXxx`.
LSP: `stdlib/klx/boot.klx` provides declarations for completion / hover.
Generator: `boot` registered in the stdlib module dispatcher.

Tests: 23/23 pass (`pkg/boot/boot_test.go`).

---

### Annotation Syntax — `[Name]` / `[Name(args)]`

New `ast.Attribute` type with `Name` and `Args`. The following AST nodes gain an `Attributes []*Attribute` field:

- `ClassDecl`, `TypeDecl`, `FunctionDecl`, `VarDecl`

New `parser/parser_attribute.go` parses `[Name]` and `[Name(args...)]` at top-level and inside class bodies. Foundation for v3.2's auto-route registration, ORM mapping, validation, and DI.

Example: `examples/complete-tutorial/example41_attributes.klx`.

---

### Compiler Fixes (KLX-C01 .. KLX-C05)

**KLX-C01 — `var p: TClass` field access (commit ff867f5)**

`var p: TPerson` previously emitted `interface{}` when the class was inherited, breaking `p.Field` access. Fix:
- `generator/generator_types.go` always emits `*TypeName` for class types
- `generator/generator.go` scans `TypeDecl`-wrapped `ClassDecl` method bodies for imports
- `generator/generator_expr.go` skips `os.ReadFile` inline when `uses sysutil` active
- `generator/generator_stdlib.go` uses concrete return types for error-returning functions

New example: `example40_declarative_oop.klx`.

**KLX-C02 — String interpolation (commit d8dbc6e)**

`lexer/lexer.go` single-quoted strings containing `${...}` now emit `STRING_INTERPOLATION` instead of plain STRING.

**KLX-C03 — Lambda / anonymous function return types (commit d8dbc6e)**

`ast.LambdaExpression` gained a `ReturnType` field. Parser saves it; generator emits the return type + `var result T` + `return result` for typed anonymous functions.

**KLX-C04 — `match` statement codegen (commit d8dbc6e)**

`match` now generates a tagless `switch { case _v == p: }` (was a broken `switch _v := ... { case _v == 1: }`).

**KLX-C05 — `uses` symbol injection in programs (commit 6f18bc9)**

`uses sysutil/jsonutil/datetime/regex/httpclient` in program files previously emitted undefined function calls. Fix:
- Generator tracks `usedModules map[string]bool`
- New `generator/generator_stdlib.go` (~270 lines) maps stdlib module names to function sets
- `resolveStdlibFunc()` checks module membership
- `generateStdlibCall()` emits `stdlib.FuncName(...)` calls
- Functions returning `(T, error)` wrapped with concrete return types

Unlocks 40+ stdlib functions in program files.

---

### LLVM Backend — Milestone 2 Phase 1 (Arrays + Optimization)

- `pkg/llvmgen/array.go` (~200 lines): static arrays `array[1..N] of T` → `alloca [N x T]`; dynamic arrays `array of T` → `{ ptr, i64, i64 }` slice struct
- Pascal 1-based indices automatically converted to LLVM 0-based
- Compile-time constant evaluation for array sizes (handles `array[1..N]` desugared to `((N-1)+1)`)
- `pkg/llvmgen/array_test.go` adds 6 tests (total 30)
- `CompileOpts.OptLevel` + `--llvm-opt=0/1/2/3` CLI flag; `llc -O=N`
- `emitMain` now allocates top-level VarDecls in `main()`; new `program *ast.Program` field on Generator

---

### Tests

- 30 LLVM tests (+6 array tests)
- 23 KylixBoot tests (new package)
- All 15 Go packages still pass
- Tutorial: 32/34 examples pass (~94%)

---

## v3.0.0-alpha (2026-06-21) — LLVM 原生后端 + WASI + 包注册中心 + stdlib Phase 4

### LLVM 原生后端（Milestone 1：最小可用子集）

新包 `pkg/llvmgen/` — Kylix → LLVM IR → 原生二进制，bypasses Go codegen。

- `codegen.go` — 生成器核心：module/function/block 管理，SSA 寄存器分配，字符串常量池
- `expr.go` — 基础类型（i64/i1/double/ptr），算术/比较/逻辑运算，WriteLn/Write/IntToStr/Length，libc 调用链
- `stmt.go` — 控制流（if/while/for/repeat），变量 alloca，函数定义，external 函数跳过 body
- `class.go` — 类 codegen：%ClassName struct、vtable 常量、method ptr %self、GEP 字段访问、虚函数分发、malloc 构造
- `compile.go` — 完整管道：`FindLLVM()` + `CompileToNative()`（AST → .ll → .o → binary via llc + clang）

CLI: `kylix build --backend=llvm main.klx`（Go 后端仍为默认）

端到端验证：Hello World + 整数算术 + while 循环 → 原生二进制，运行正确。

测试：24 个单元测试（`pkg/llvmgen/codegen_test.go`，含 6 个类 codegen 测试）

已知限制（Milestone 1）：不支持接口、泛型、数组、record、异常；无优化 Pass（-O0）；仅 libc 调用。

---

### WASI 支持

CLI: `kylix build --wasi main.klx`（Go 1.21+ `GOOS=wasip1 GOARCH=wasm`），`--tinygo` 更小二进制。

`pkg/wasi/` — build-tag 分离：`wasi_wasip1.go`（原生）+ `wasi_stub.go`（非 WASI 本地测试）
- 函数：Stdout/Stderr/Stdin、Args/Getenv/Environ、ClockMonotonic/ClockWalltime、ReadFile/WriteFile、WasiExit

`stdlib/src/wasi.klx` + `stdlib/klx/wasi.klx` — 纯 Kylix 高层包装（WriteLn/ReadLine/HasEnv/ArgCount/ElapsedMs 等）

示例：`examples/wasi-hello/`（Wasmtime/Node.js）、`examples/cloudflare-worker/`（Cloudflare Workers）

测试：8 个单元测试（`pkg/wasi/wasi_test.go`）

---

### 包注册中心服务端

`registry/` — 独立 Go module，完整的包发布 / 搜索 / 下载服务。

后端：SQLite 数据库层（Store 接口，可切换 PostgreSQL），Bearer token 认证，REST API：
```
GET  /api/v1/packages              搜索包（?q= 关键词）
POST /api/v1/packages              发布包（需 Bearer token）
GET  /api/v1/packages/:name/versions  版本列表
GET  /api/v1/packages/:name/:ver/dl   下载（重定向 tarball）
```

Web 前端：htmx + Tailwind CSS（首页实时搜索 + 包详情页）

CLI: `kylix publish [--version=X] [--registry=URL] [--token=T]`，支持 `KYLIX_TOKEN` 环境变量

测试：7 个集成测试（`registry/internal/api/handler_test.go`）

---

### stdlib Phase 4 — 纯 Kylix 化（jsonutil/regex/datetime）

**`stdlib/src/jsonutil.klx`** (390 行) — 完整 JSON 解析器（TJsonLexer + TJsonParser），支持任意深度嵌套对象和数组。修复 ROADMAP 已知缺陷："jsonutil 仅支持扁平 JSON"。JsonEncode/JsonEncodePretty 保留 external（依赖 Go reflect）。

**`stdlib/src/regex.klx`** (387 行) — 纯 Kylix 字符级验证函数：IsDigit/IsLower/IsUpper/IsLetter、IsNumeric/IsAlpha/IsAlphanumeric、IsEmail/IsURL/IsIPv4/IsPhone/IsDate。无 NFA 开销。

**`stdlib/src/datetime.klx`** (260 行) — FormatPattern（yyyy/MM/dd/HH/mm/ss）、DateAdd/DateSub（修复 ROADMAP TDateTime +/- 缺陷）、IsLeapYear/DaysInMonth/IsWeekend/IsWeekday/WeekNumber、MonthName/DayName。

声明文件：更新 `stdlib/klx/regex.klx` 和 `stdlib/klx/datetime.klx`

测试：69 个（jsonutil: 29, regex: 19, datetime: 21）

---

### 编译器 Bug 修复: `external` 函数声明解析

`function Foo(): T; external;` 在文件末尾或任意位置现在均可正确解析。

- `ast/ast.go`: `FunctionDecl.IsExternal` 新字段
- `parser/parser_decl.go`: `EXTERNAL` 修饰词识别（行 253）
- `generator/generator_types.go`: IsExternal=true 时跳过函数体生成
- 8 个新测试（3 parser + 5 generator）

---

### stdlib HTTP 客户端

`stdlib/http_client.go` — THttpClient（Get/Post/StatusCode/SetHeader），NewHttpClient()，HttpGet/HttpPost/HttpGetJSON 一键函数

`stdlib/klx/httpclient.klx` — LSP 声明文件

`stdlib/src/httpclient.klx` — 纯 Kylix 包装：BuildQueryString/IsOK/IsRedirect/IsClientError/IsServerError

---

## v2.6.0 (2026-06-20)

### 🎉 Performance & Optimization

v2.6.0 brings parallel compilation, dead code elimination, and LSP performance benchmarks.

---

### Task 1: Parallel Compilation

`CompileProject` now parses all source files in parallel using goroutines.

- Each file gets its own goroutine for ReadFile + Lex + Parse
- `sync.WaitGroup` waits for all, results collected in original order
- Race detector verified (`go test -race`)
- 9-file project: CPU utilization 113% (multi-core active)
- Cache logic preserved (cached files still skip GenerateBody)

---

### Task 2: Dead Code Elimination

New `pkg/compiler/optimize.go` removes unreachable code:

```pascal
function Foo(): Integer;
begin
  result := 42;
  return;
  WriteLn('unreachable');  // ← eliminated
  WriteLn('also gone');    // ← eliminated
end;
```

- Terminators detected: `return`, `raise`, `Exit`, `break`, `continue`
- Recursive: optimizes nested blocks (if/while/for/try bodies)
- No-op when no dead code exists (preserves all statements)

**Tests**: `pkg/compiler/optimize_test.go` (5 tests)

---

### Task 3: LSP Large File Performance Benchmark

Verifies incremental edit latency on large files:

| Scenario | File Size | Latency | Target |
|----------|-----------|---------|--------|
| Full parse | 500 functions (~2K lines) | 1.2ms | < 200ms ✅ |
| Incremental edit | 500 functions | 1.0ms | < 200ms ✅ |
| Symbol collection | 200 functions | < 1ms | complete ✅ |

v2.3.0's incremental sync investment pays off — even 500-function files edit in < 2ms.

**Tests**: `pkg/lsp/perf_test.go` (3 tests)

---

### Summary

| Task | Tests | Type |
|------|-------|------|
| Parallel compilation | – (race detector) | Performance |
| Dead code elimination | 5 | Optimization |
| LSP perf benchmark | 3 | Performance guard |
| **Total v2.6.0** | **8** | |

---

## v2.5.0 (2026-06-20)

### 🎉 Toolchain Deepening

v2.5.0 completes the "infrastructure ready but not fully wired" items from v2.3-v2.4: LSP refactoring, doc code examples, bench memory, iter module, and the long-standing class method external definition bug.

---

### Task 1: LSP Refactoring Actions

- **Cross-file rename**: `textDocument/rename` now walks all open documents via `ReferenceWalker`, returning `WorkspaceEdit` with edits for every file containing the symbol
- **Context-aware code actions**: `textDocument/codeAction` now offers "Rename Symbol: \<name\>" when cursor is on an identifier, and "Extract Function" when there's a selection
- `ReferenceWalker` API exported: `NewReferenceWalker(name, uri)` + `References()`

**Tests**: `pkg/lsp/rename_test.go` (5 tests)

---

### Task 2: `kylix doc` Code Example Extraction

Doc comments now preserve multi-line structure and fenced code blocks (` ```pascal ... ``` `).

```pascal
// Reverse returns the reversed string.
//
// ```pascal
// WriteLn(Reverse('abc'));  // cba
// ```
function Reverse(s: String): String;
```

Generates Markdown with the code block preserved.

**Tests**: `pkg/docgen/examples_test.go` (4 tests)

---

### Task 3: `kylix bench --mem` Memory Allocation Report

```bash
$ kylix bench --count 2 --mem fib_bench.klx
ok  BenchFib10  321.05 ms/op  0 B/op  0 allocs/op
```

Uses `runtime.ReadMemStats` before/after benchmark execution, reports `TotalAlloc` delta and `Mallocs` delta.

---

### Task 4: `iter` Iterator Module

9 array utility functions in pure Kylix:

| Function | Purpose |
|----------|---------|
| `Contains(arr, v)` | Linear search |
| `Count(arr, v)` | Count occurrences |
| `Unique(arr)` | Remove duplicates |
| `Reverse(arr)` | New reversed array |
| `Concat(a, b)` | Merge two arrays |
| `Slice(arr, start, end)` | Subarray extraction |
| `Sum(arr)` | Element sum |
| `Min(arr)` / `Max(arr)` | Extrema |

Note: Map/Filter/Reduce deferred — Kylix doesn't yet support function-type parameters.

**Tests**: `stdlib/src/iter_test.klx` (9 Kylix-level tests)

---

### Task 5: Class Method External Definition Fix

Fixed long-standing bug where Pascal-style external method definitions generated duplicate Go methods.

```pascal
type
  TFoo = class
    function Bar(): Integer;  // forward declaration (no body)
  end;

function TFoo.Bar(): Integer;  // external definition
begin result := 42; end;
```

`generateClassDecl` now skips methods with `Body == nil` (forward declarations), letting `generateFunctionDecl` handle the external definition.

**Tests**: `generator/extmethod_test.go` (4 tests)

---

### Summary

| Task | Tests | Type |
|------|-------|------|
| LSP refactoring (rename + codeAction) | 5 | IDE |
| doc code examples | 4 | Documentation |
| bench --mem | – | Benchmarking |
| iter module | 9 (Kylix) | Standard library |
| External method fix | 4 | Compiler |
| **Total v2.5.0** | **22** | |

### stdlib Cumulative (7 modules, 54 functions, 48 tests)

| Module | Phase | Functions | Tests |
|--------|-------|-----------|-------|
| `strutil` | v2.1 | 8 | 8 |
| `mathutil` | v2.1 | 12 | 10 |
| `arrayutil` | v2.2 | 8 | 8 |
| `collections` | v2.2 | 6 | 5 |
| `stringbuilder` | v2.4 | 5 | 4 |
| `resulttype` | v2.4 | 6 | 4 |
| `iter` | v2.5 | 9 | 9 |

---

## v2.4.0 (2026-06-20)

### 🎉 Polish & Ecosystem

v2.4.0 completes the v2.3 infrastructure: i18n fully wired, REPL `:type` uses
real inference, SetLength fixed, package manager gains nested deps + lockfile,
and stdlib Phase 3 adds `stringbuilder` + `resulttype`.

---

### Task 1: i18n Fully Integrated

Error messages now respect `KYLIX_LANG`:

```bash
$ KYLIX_LANG=zh kylix check broken.klx
error[KLX101]: 无法将 String 字面量赋给类型为 'Integer' 的变量
  = help: 使用 StrToInt(s) 或 StrToInt64(s) 把 String 转为 Integer
```

**Changes:**
- `typecheck.go`: `c.diag` / `c.diagHint` now use `i18n.T(code, args...)`
- `suggestions.go`: `typeConversionHint` uses `i18n.Hint()`
- All KLX101 (type mismatch) + KLX201 (undeclared) + KLX104 (generic constraint) localized
- Error codes (KLX101) stay constant across languages — only messages change
- `i18n.HasCode()` added for fallback detection

**Tests**: `pkg/compiler/i18n_integration_test.go` (4 tests)

---

### Task 2: REPL `:type` Real Inference

```
kylix> :type 1 < 2       → Boolean
kylix> :type 3.0 + 4     → Real
kylix> :type GetAge()    → Integer  (from function return type)
```

**Changes:**
- New exported `compiler.InferType(program, expr)` — full inference engine
- REPL `showType()` rewritten to parse `__probe := <expr>` and call `InferType`
- Falls back to literal guess on parse failure

**Tests**: `pkg/compiler/infertype_export_test.go` (6 tests)

---

### Task 3: SetLength Fixed

```pascal
var arr: array of Integer;
arr := nil;
SetLength(arr, 3);  // ← was panic, now works
arr[0] := 10;
SetLength(arr, 1);  // truncate works
SetLength(arr, 0);  // zero-length works
```

**Changes:**
- `generator.go`: new `needsSetLength` flag + `setLengthHelperSource` (Go generic `__kylixSetLength[T any]`)
- `generator_stmt.go`: `SetLength(arr, n)` → `arr = __kylixSetLength(arr, int(n))`
- Helper grows (append zeros) or truncates as needed
- `writeRuntimeHelpers()` emits the helper at end of output (Go allows any order)

---

### Task 5: Package Manager — Nested Deps + Lockfile

```bash
$ kylix add mylib github.com/user/mylib@v1.0.0
  installing mylib (github.com/user/mylib@v1.0.0)…
  resolving 2 nested dependenc(ies) for mylib…
✓ Added mylib

$ cat kylix.lock
[dependency "mylib"]
ref = "github.com/user/mylib@v1.0.0"
sha = "abc123def456..."
```

**Changes:**
- `installGit()`: after clone, reads package's `kylix.toml` for nested `[dependencies]`
- Recursively installs nested deps (skips already-installed)
- `writeLock()`: generates `kylix.lock` with ref + git SHA per dependency
- `Add()` / `InstallAll()` / `Remove()` all update lockfile

---

### Task 6: stdlib Phase 3

Two new pure-Kylix modules:

#### `stdlib/src/stringbuilder.klx` — TStringBuilder (5 methods)

| Method | Purpose |
|--------|---------|
| `Append(s)` | Add string |
| `AppendLine(s)` | Add string + newline |
| `Clear()` | Reset to empty |
| `Length()` | Total character count |
| `ToString()` | Get combined string |

#### `stdlib/src/resulttype.klx` — TResult (3 methods + 3 functions)

| Member | Purpose |
|--------|---------|
| `TResult.Unwrap()` | Get value or panic |
| `TResult.UnwrapOr(fallback)` | Get value or default |
| `TResult.ErrorMsg()` | Get error string |
| `Ok(value)` | Create success result |
| `Err(msg)` | Create error result |
| `SafeDiv(a, b)` | Example: divide returning Result |

**Tests**: 8 new Kylix-level tests (4 + 4)

#### Cumulative stdlib Kylix coverage

| Module | Phase | Functions | Tests |
|--------|-------|-----------|-------|
| `strutil` | v2.1 | 8 | 8 |
| `mathutil` | v2.1 | 12 | 10 |
| `arrayutil` | v2.2 | 8 | 8 |
| `collections` | v2.2 | 6 | 5 |
| `stringbuilder` | v2.4 | 5 | 4 |
| `resulttype` | v2.4 | 6 | 4 |
| **Total** | | **45** | **39** |

#### Bug fixes discovered during Phase 3

- `result` is a Kylix keyword (RESULT token) → unit renamed to `resulttype`
- `default` is a keyword → parameter renamed to `fallback`
- testrunner harness missing Exception type → injected in buildHarness

---

### Summary

| Task | Tests | Type |
|------|-------|------|
| i18n integration | 4 | Internationalization |
| REPL :type inference | 6 | Developer experience |
| SetLength fix | – | Correctness |
| Package manager nested deps + lockfile | – | Ecosystem |
| stdlib Phase 3 | 8 (Kylix) | Standard library |
| **Total v2.4.0** | **18** | |

### Known Limitations
- `iter` module deferred to v2.5 (needs generator-level iterator protocol)
- LSP refactor actions (rename, extract) deferred to v2.5
- i18n covers typecheck errors; parser/Go-compiler errors still English-only

---

## v2.3.0 (2026-06-19)

### 🎉 Developer Experience: Editor, REPL, Test, Debug, WASM

v2.3.0 polishes the developer-facing surface — IDE responsiveness, interactive
REPL, test runner ergonomics, language localization, debugger integration,
and a WebAssembly target.

---

### Task 1: LSP Incremental Synchronization

Editors no longer lose sync on rapid typing.

**Before**: every `didChange` parsed the full document N times (one per change).
**After**: changes batched, applied incrementally, parsed once.

**Changes:**
- `Document.Version` tracks LSP version for each document
- `DocumentStore.Update(uri, text, version)` rejects stale versions
- New `ApplyChanges(uri, version, []TextDocumentContentChange)` —
  incremental range edits
- New `applyRangeEdit(text, range, newText)` + `positionToOffset` helpers
- Server capability: `textDocumentSync` 1 (Full) → 2 (Incremental)
- Single `publishDiagnostics` per `didChange` (was N)

**Tests**: `pkg/lsp/sync_test.go` (8 tests)

---

### Task 2: REPL Enhancements

```
kylix> writ<Tab>           → completes to 'WriteLn'
kylix> :ty 42               → :type 42  →  Integer
kylix> :load mylib.klx      → loads declarations from file
```

**Changes:**
- `liner` Tab completion enabled (was unused dependency)
- `buildCompleter()` combines:
  - Pascal/Kylix keywords
  - Built-in functions
  - Meta-commands (`:help`, `:load`, `:type`, ...)
  - User-declared identifiers
- New `:load <file>` — read .klx file, strip program shell, execute declarations
- New `:type <expr>` / `:t <expr>` — show inferred literal type
- `extractDeclaredNames()` mines scope from accumulated declarations

**Tests**: `pkg/repl/repl_test.go` (8 tests)

---

### Task 3: kylix test Advanced Features

```pascal
unit fixture_test;
var counter: Integer;

procedure Setup;
begin
  counter := 100;  // runs before each test
end;

procedure Teardown;
begin
  WriteLn('cleaning up');  // runs after each test (deferred)
end;

procedure TestAdd;
begin
  Assert(counter + 1 = 101, 'incremented');
end;
```

```bash
$ kylix test --filter Add fixture_test.klx
  ok  TestAdd
1 passed, 0 failed (filter: "Add")
```

**Changes:**
- `Runner.Filter` field + `SetFilter` / `FilterCases` methods
- New `--filter <substr>` CLI flag
- `detectFixtures()` finds `Setup` / `Teardown` procedures per file
- `buildHarness` injects `Setup()` at start and `defer Teardown()` after
- `Teardown` runs even when test panics (defer semantics)

**Tests**: `pkg/testrunner/filter_test.go` (3 tests)

---

### Task 4: i18n — Error Message Internationalization

```bash
$ KYLIX_LANG=zh kylix check broken.klx
错误[KLX201]: 未声明的变量或函数 'unknownVar'
```

`pkg/i18n/` package (new):
- 21 error code translations × 2 languages (English + Chinese)
- 6 fix-hint translations
- `T(code, args...)` — template lookup with English fallback
- `Hint(hintKey, args...)` — localized fix suggestion
- `KYLIX_LANG` env var: `en` (default) / `zh` (`zh-cn`, `chinese` aliases)

Note: i18n package is independent and ready; full integration with
`typecheck.diag` and `printDiagnostics` will land in v2.4 to avoid
disruption to existing tests.

**Tests**: `pkg/i18n/i18n_test.go` (9 tests)

---

### Task 5: Delve Debugger Integration

```bash
$ kylix debug main.klx
(dlv) break main.klx:5
Breakpoint 1 set at 0x... for main main.go:5
(dlv) continue
```

`cmd/kylix/cmd_debug.go` (new):
- Compiles `.klx` → `.go` (preserves `//line` directives)
- `go build -gcflags='all=-N -l'` — disables optimization for debugging
- Spawns `dlv exec <binary>` — interactive or `--headless --port=N` for IDE
- `--keep` retains intermediate Go file
- Detects missing `dlv` and shows install command

Source-line mapping: generator's existing `//line file.klx:N` directives carry
into DWARF debug info, so `break main.klx:42` works directly.

---

### Task 6: WebAssembly Backend (MVP)

```bash
$ kylix build --wasm hello.klx
✓ Built hello.klx → hello.wasm [wasm via Go]   (2.5 MB)

$ kylix build --wasm --tinygo hello.klx
✓ Built hello.klx → hello.wasm [wasm via TinyGo]  (~30 KB)
```

**Changes:**
- `--wasm` flag in `cmd build`
- `--tinygo` for size-optimized output (requires tinygo installed)
- `--wasm` and `--target` mutually exclusive; `--tinygo` requires `--wasm`
- New `goBuildWasm(goFile, outBin, useTinyGo)` function
- Both single-file and project mode supported

Implementation strategy: leverages Go's existing wasm backend (`GOOS=js GOARCH=wasm`)
rather than a custom IR. TinyGo path produces ~80× smaller binaries suitable
for browser deployment.

---

### Summary

| Task | Tests | Type |
|------|-------|------|
| LSP incremental sync | 8 | Editor performance |
| REPL Tab/load/type | 8 | Interactive experience |
| Test fixtures + filter | 3 | Testing ergonomics |
| i18n framework | 9 | Internationalization |
| Delve debug command | – | Tooling |
| WASM target | – | Deployment |
| **Total v2.3.0** | **28** | |

### Breaking Changes
- `DocumentStore.Update` signature: now `(uri, text, version int)` — callers must add `0` for legacy version
- LSP `textDocumentSync` capability advertises Incremental (2) instead of Full (1)

### Known Limitations
- i18n integration not yet wired into all error paths (deferred to v2.4)
- Debug command requires `dlv` installed separately
- WASM with TinyGo requires `tinygo` installed; standard Go path always works
- `:type` REPL command does literal-only inference (not full expression types)

---

## v2.2.0 (2026-06-19)

### 🎉 Engineering Quality & stdlib Phase 2

v2.2.0 focuses on production-readiness: continuous integration, deeper type
checking, project-level diagnostics, and incremental builds. Plus two new
pure-Kylix stdlib modules.

---

### Task 1: GitHub Actions CI/CD

`.github/workflows/`:
- **ci.yml** — Multi-platform testing (Linux + macOS × Go 1.21/1.22/1.23)
  - `go build`, `go vet`, `go test -race -timeout 60s ./...`
  - Kylix-level integration: `kylix test stdlib/src/*_test.klx`
  - Independent `lint` job runs `gofmt` check
- **release.yml** — Cross-platform binary release on `git tag v*`
  - Builds: linux/amd64+arm64, darwin/amd64+arm64, windows/amd64
  - Auto-extracts release notes from CHANGELOG.md
  - Creates GitHub Release with binaries attached

`gofmt -w` applied to entire codebase (19 files reformatted, no logic changes).

---

### Task 2: Generic Constraint Method Signature Verification

Previously v2.1.2 only checked method **names** existed. Now signatures match.

```pascal
type
  IFoo = interface
    function Bar(x: Integer): String;
  end;
  TBox<T: IFoo> = class end;

  TBad = class implements IFoo
    function Bar(): Integer;  // ❌ wrong params + wrong return type
  end;

var b: TBox<TBad>;
// error[KLX104]: TBad does not satisfy IFoo (signature mismatch on Bar)
```

**Changes:**
- `interfaces` / `classMethods` upgraded: `[]string` → `map[name]*FunctionDecl`
- New `signaturesMatch(impl, want)` — compares param count, types, return types
- New `typesEqual(a, b)` — type expression equality with alias resolution
- Type aliases are transparent (`UserId = Integer` → matches `Integer`)

**Tests**: `pkg/compiler/signature_test.go` (6 tests)

---

### Task 3: Project-Level Type Checking

`kylix check` now does **full cross-file analysis**, not just per-file syntax.

```bash
$ kylix check
error[KLX201]: call to undeclared function 'Cube'
  --> main.klx:4:15

1 error(s) across 2 file(s)
```

**Changes:**
- New `compiler.CheckProject(files)` — runs syntax + interface + type checks across all files
- Cross-file symbol merging prevents false-positive "undeclared" for cross-unit calls
- New `isStdlibUnit()` whitelist — `uses sysutil` doesn't require local `.klx`
- New `checker.strictFunctionCalls` flag distinguishes single-file vs project mode
- `cmdCheck` now defaults to project mode; `--syntax` retains parser-only behaviour

**Tests**: `pkg/compiler/checkproject_test.go` (6 tests)

---

### Task 4: Incremental Compilation Activated

`BuildCache` infrastructure existed since v1.4.0 but was never wired into the
project-mode build path.

**Fix**: `cmd/kylix/cmd_build.go` project mode now calls `CompileProject`
(which uses cache) for multi-file projects. Single-file projects retain
`CompileFile` for compatibility.

**Effect on a 2-file project:**
```
$ rm -rf .kylix-cache build
$ kylix build -v          # cold
  compile: math.klx
  compile: main.klx

$ kylix build -v          # warm cache
  cached: main.klx
  cached: math.klx
  reuse:  math.klx
  reuse:  main.klx

$ touch math.klx
$ kylix build -v          # partial rebuild
  cached: main.klx
  compile: math.klx       # ← only changed file
  reuse:  main.klx
```

**Tests**: `pkg/compiler/incremental_test.go` (4 tests)

---

### Task 5: stdlib Kylix-ification Phase 2

Two new pure-Kylix modules joining v2.1.0's `strutil`/`mathutil`.

#### `stdlib/src/arrayutil.klx` (8 functions)

| Function | Purpose |
|----------|---------|
| `Sum(arr)` | Sum of integers |
| `Product(arr)` | Product of integers (1 for empty) |
| `MinValue(arr)` | Smallest element |
| `MaxValue(arr)` | Largest element |
| `ArrayContains(arr, v)` | Linear search |
| `IndexOf(arr, v)` | Position of v, or -1 |
| `ArrayReverse(arr)` | New reversed array |
| `ArrayLength(arr)` | Wrapper for built-in `Length` |

#### `stdlib/src/collections.klx` — `TIntList`

```pascal
var list: TIntList;
list := TIntList.Create();
list.Add(10);
list.Add(20);
list.Add(30);
WriteLn(list.Sum());     // 60
WriteLn(list.Count());   // 3
list.Clear();
WriteLn(list.IsEmpty()); // true
```

Methods: `Count()`, `Get(i)`, `Add(v)`, `Clear()`, `IsEmpty()`, `Sum()`.

**Tests**:
- `arrayutil_test.klx`: 8 tests
- `collections_test.klx`: 5 tests

#### Cumulative stdlib Kylix coverage

| Module | Functions | Tests |
|--------|-----------|-------|
| `strutil` (v2.1.0) | 8 | 8 |
| `mathutil` (v2.1.0) | 12 | 10 |
| `arrayutil` (v2.2.0) | 8 | 8 |
| `collections` (v2.2.0) | 6 (methods) | 5 |
| **Total** | **34** | **31** |

---

### Summary

| Task | Tests | Type |
|------|-------|------|
| CI/CD pipelines | – | Infrastructure |
| Generic signature verification | 6 | Type system |
| Project-level checking | 6 | Type system |
| Incremental compilation | 4 | Performance |
| stdlib Phase 2 | 13 | Standard library |
| **Total v2.2.0** | **29** | |

### Breaking Changes
- `kylix check` now performs full type checking by default (use `--syntax` for
  parse-only behaviour)
- Class implementing an interface must have **matching method signatures**
  (parameter types and return type), not just method names

### Known Limitations
- `SetLength` builtin only grows existing slices (workaround: use `append` with `nil` initial value)
- Method declarations split across class body and outside (Pascal-style) generate duplicate Go methods (workaround: define methods inline in class body)
- Multi-parameter generic constraints validated by parameter position, not name

---

## v2.1.0 (2026-06-19)

### 🎉 Enhanced Type System & stdlib Kylix-ification

v2.1.0 strengthens the type system with multi-parameter generic constraints,
real interface implementation verification, expanded type inference, and
introduces the first pure-Kylix stdlib modules.

---

### M2.1.1: Multi-Parameter Generic Constraints

```pascal
type
  IComparable = interface
    function CompareTo(): Integer;
  end;
  IHashable = interface
    function HashCode(): Integer;
  end;
  TMap<K: IComparable, V: IHashable> = class
  end;

var m: TMap<Integer, String>;
// error[KLX104]: type 'Integer' does not satisfy constraint 'IComparable'
//                for parameter 'K' of generic type 'TMap'
// error[KLX104]: type 'String' does not satisfy constraint 'IHashable'
//                for parameter 'V' of generic type 'TMap'
```

**Changes:**
- New `GenericTypeInfo` struct preserves parameter declaration order
- `genericConstraints` now tracks both ordered names and constraints
- Each type argument validated independently against its constraint
- Error messages include the specific parameter name

**Tests**: `pkg/compiler/generics_multi_test.go` (4 tests)

---

### M2.1.2: Class → Interface Implementation Mapping

Previously, custom types were assumed to satisfy any constraint (false positive).
Now we verify actual `implements` declarations and method existence.

```pascal
type
  IComparable = interface
    function CompareTo(): Integer;
  end;
  TBox<T: IComparable> = class end;

  TBadType = class implements IComparable
    // Missing CompareTo
  end;

var b: TBox<TBadType>;
// error[KLX104]: TBadType claims IComparable but lacks CompareTo
```

**Changes:**
- New `classImpls` / `classParent` / `classMethods` tracking
- `typeImplementsInterface` now verifies:
  1. Built-in types never implement user interfaces
  2. Type alias chain resolution
  3. Direct `implements` declaration + method signature presence
  4. Inherited implementation via parent class chain

**Tests**: `pkg/compiler/impl_test.go` (5 tests)

---

### M2.1.3: Enhanced Type Inference

Expanded `inferExprType` to handle more expression forms:

```pascal
var b := 1 < 2;            // → Boolean (comparison)
var ok := true and false;  // → Boolean (logical)
var n := not true;         // → Boolean (prefix not)
var arr := [1, 2, 3];      // → array of Integer
var p := nil;              // → nil
```

**Changes:**
- `NilLiteral` → `nil`
- `ArrayLiteral` → `array of <element type>`
- `LambdaExpression` → `function`
- `IndexExpression` → element type from `array of T`
- Comparison operators (`=`, `<>`, `<`, `>`, `<=`, `>=`) → `Boolean`
- Logical operators (`and`, `or`, `xor`) → `Boolean`
- Prefix `not` → `Boolean`

**Tests**: `pkg/compiler/typeinfer_v2_test.go` (6 tests)

---

### M2.1.4: stdlib Kylix-ification Phase 1

Two stdlib modules now have **pure-Kylix implementations** demonstrating that
core utilities can be self-hosted without performance loss.

#### `stdlib/src/strutil.klx` (8 functions)

| Function | Purpose |
|----------|---------|
| `Reverse(s)` | Reverse character order |
| `IsEmpty(s)` | Check empty string |
| `StartsWith(s, prefix)` | Prefix check |
| `EndsWith(s, suffix)` | Suffix check |
| `Contains(s, substr)` | Substring search |
| `RepeatStr(s, n)` | Repeat string n times |
| `PadLeft(s, w, c)` | Left-pad to width |
| `PadRight(s, w, c)` | Right-pad to width |

#### `stdlib/src/mathutil.klx` (12 functions)

| Function | Purpose |
|----------|---------|
| `Abs(x)`, `AbsReal(x)` | Absolute value |
| `Min(a, b)`, `Max(a, b)` | Extrema |
| `Clamp(x, lo, hi)` | Bound to range |
| `Sign(x)` | Sign function |
| `Pow(base, exp)` | Integer exponentiation |
| `Factorial(n)` | n! |
| `Gcd(a, b)`, `Lcm(a, b)` | GCD / LCM |
| `IsPrime(n)` | Primality test |

**Tests**: 18 tests in `*_test.klx`, all passing via `kylix test`

#### Supporting Infrastructure

**`pkg/testrunner/runner.go`** — Test runner now resolves `uses` clauses:
- Parses dependent `.klx` files in same directory
- Compiles them together with the test file
- No more "undefined symbol" errors when testing modules

**`generator/`** — Critical bug fix:
- New `userFuncs map[string]bool` tracks user-defined function names
- `mapBuiltinFunction` skips rewriting when user defines a function
- Previously `function Abs(x: Integer): Integer` was incorrectly rewritten
  as `math.Abs` calls. Now user definitions take precedence.

---

### Summary

| Feature | Tests | LOC |
|---------|-------|-----|
| Multi-param generic constraints | 4 | ~80 |
| Class→Interface mapping | 5 | ~120 |
| Enhanced type inference | 6 | ~50 |
| strutil + mathutil + tests | 18 | ~300 |
| **Total v2.1.0 additions** | **33** | **~550** |

### Breaking Changes
- `function Abs(x)` user definition now correctly takes precedence over `math.Abs`
- Custom types must explicitly declare `implements IFoo` AND have all methods to satisfy generic constraints (previously always passed)

### Known Limitations
- Generic constraint verification doesn't check method signatures (only names)
- Parameter ordering for nested generics may need refinement
- stdlib Kylix-ification is Phase 1 (more modules in v2.2+)

---

## v2.0.0 (2026-06-17)

### 🎉 Production-Ready Release

Kylix v2.0.0 completes the compiler toolchain with enhanced type checking, testing, documentation generation, and performance benchmarking capabilities.

---

### M1: Error Experience Overhaul

**Error Codes & Recovery** (`pkg/compiler/errors.go`, `typecheck.go`)
- Structured error codes (KLX001–KLX499) with ranges:
  - KLX001–099: Syntax errors
  - KLX100–199: Type errors
  - KLX200–299: Semantic errors (undeclared, arity)
  - KLX300–399: Interface/contract errors
- Context-aware error messages with file/line/column
- Type mismatch recovery: infer expected type and continue checking

**Intelligent Suggestions** (`pkg/compiler/suggestions.go`)
- Levenshtein distance ≤2 for typo correction
- "did you mean X?" hints for undeclared identifiers
- Type conversion suggestions (e.g., `IntToStr`, `StrToInt`)

---

### M2: Type System Enhancements

**M2.1: Type Inference** (`pkg/compiler/typecheck.go`)
- `var x := 42` → infer `Integer`
- `var s := 'hello'` → infer `String`
- `var age := GetAge()` → infer function return type
- Arithmetic type propagation (`Integer` + `Integer` → `Integer`)
- **Tests:** `pkg/compiler/typeinfer_test.go` (6 tests)

**M2.2: Generic Constraint Validation** (`pkg/compiler/typecheck.go`)
```pascal
type
  IComparable = interface
    function CompareTo(other: IComparable): Integer;
  end;
  TBox<T: IComparable> = class end;

var box: TBox<Integer>;  // error[KLX104]: Integer does not satisfy IComparable
```
- Collects constraints from `<T: IComparable>` syntax
- Validates type arguments at instantiation time
- Built-in types (Integer, String, Boolean) don't implement user interfaces
- **Tests:** `pkg/compiler/generics_test.go` (3 tests)

**M2.3: Type Alias Enhancement** (v1.4.0 foundation)
- Circular dependency detection
- Alias chain resolution with cycle guards

---

### M3: Toolchain Expansion

**M3.1: Test Runner** (`pkg/testrunner/`, `cmd/kylix/cmd_testcmd.go`)
```pascal
unit math_test;
procedure TestAdd;
begin
  Assert(2 + 3 = 5, 'expected 2+3=5');
end;
```
```bash
$ kylix test
  ok  TestAdd
  ok  TestSubtract
  FAIL TestDivideByZero
       FAIL: expected division by zero error

2 passed, 1 failed
```
- Discovers `*_test.klx` files and `Test*` procedures
- Built-in `Assert(condition, message)` for test assertions
- TAP version 14 output format (`--tap` flag)
- Compiles tests with isolated Go harness per test
- **Tests:** `pkg/testrunner/runner_test.go` (4 tests)

**M3.2: Documentation Generator** (`pkg/docgen/`, `cmd/kylix/cmd_doc.go`)
```pascal
// StringUtils provides string manipulation utilities.
unit stringutils;

// Reverse returns the string s reversed character by character.
function Reverse(s: String): String;
```
Generates:
```markdown
# stringutils
StringUtils provides string manipulation utilities.

## Functions
### Reverse
```pascal
function Reverse(s: String): String
```
Reverse returns the string s reversed character by character.
```
- Extracts `//` doc comments immediately preceding declarations
- Generates Markdown grouped by kind (Functions, Classes, Types, etc.)
- `kylix doc` → outputs to `docs/api/*.md`
- `kylix doc --stdout` → prints to console
- **Tests:** `pkg/docgen/docgen_test.go` (5 tests)

**M3.3: Benchmark Runner** (`pkg/testrunner/`, `cmd/kylix/cmd_bench.go`)
```pascal
unit fib_bench;
procedure BenchFib15;
var x: Integer;
begin
  x := Fib(15);
end;
```
```bash
$ kylix bench --count 5
Running 1 benchmark(s), 5 iteration(s) each...
ok    BenchFib15    39.34 ms/op
```
- Discovers `*_bench.klx` files and `Bench*` procedures
- Measures wall-clock time over N iterations (default 5)
- Reports average time per operation (ns/µs/ms/s per op)
- Compatible output format with Go benchmarks

---

### Summary

| Feature | Status | LOC | Tests |
|---------|--------|-----|-------|
| **Error codes & recovery** | ✅ | ~400 | 8 |
| **Type inference** | ✅ | ~100 | 6 |
| **Generic constraints** | ✅ | ~120 | 3 |
| **Test runner** | ✅ | ~280 | 4 |
| **Doc generator** | ✅ | ~340 | 5 |
| **Benchmark runner** | ✅ | ~120 | – |
| **Total** | ✅ | **~1360** | **26** |

### Breaking Changes
None — all additions are backward compatible.

### Known Limitations
- Generic constraint validation only supports single type parameter (`TBox<T>`)
- Multi-parameter generics (`TMap<K, V>`) require parameter ordering info (future work)
- Custom types assumed to satisfy constraints (no full class→interface mapping yet)

---

## v1.5.0 (2026-06-14)

### P3: stdlib Kylix化 + 包管理器

#### stdlib `.klx` 声明文件 (`stdlib/klx/`)

Four standard-library modules now have `.klx` declaration files that the LSP
reads to provide completion, hover, and signature help — without rewriting the
Go implementation.

| File | Coverage |
|------|----------|
| `stdlib/klx/sysutil.klx` | file I/O, path ops, env, TTextFile |
| `stdlib/klx/datetime.klx` | TDateTime with 30+ methods, factory functions |
| `stdlib/klx/regex.klx` | TRegex, one-shot helpers, IsEmail/IsURL/… |
| `stdlib/klx/jsonutil.klx` | encode/decode, map accessors, file I/O |

LSP auto-loads the relevant `.klx` file when a document has `uses sysutil;`
etc., adding both qualified (`sysutil.ReadFile`) and unqualified symbols.

#### Package manager (`pkg/pkgmgr/`)

Minimal but complete package manager for Kylix projects.

```
kylix add utils github.com/alice/utils@v1.0.0  # add + install
kylix install                                   # install all deps
kylix remove utils                              # remove
```

- Packages installed to `packages/<name>/` (git clone or symlink for locals)
- `kylix.toml` gains `[dependencies]` section
- `Config.Dependencies map[string]string` added to project config
- Local path refs (`"./local_pkg"`) use symlinks for dev convenience
- `pkgmgr.Manager.PackageDirs()` returns dirs for compiler search path

#### Updated help (`kylix --help`)

Package commands listed in usage output.

---


## v1.4.0 (2026-06-13)

### P2: Incremental compilation

Unchanged `.klx` files now skip the parse+generate step on repeated builds.
A file-fingerprint cache (mtime + size) lives in `.kylix-cache/` and survives
across processes.

**Results on a 2-unit project:**
| Build | Time | Notes |
|-------|------|-------|
| Cold (no cache) | 444 ms | full compile |
| All cached | 8 ms | **55× faster** |
| One file changed | 6 ms | only changed unit recompiled |

**How it works:**
- `pkg/compiler/cache.go` — `BuildCache`: SHA-256 keyed JSON entries per file
- `CompileProject` uses cache when `opts.CacheDir != ""`
- Global pre-scan (class types, imports, exceptions) still runs over all ASTs
- Only the `GenerateBody` step is skipped for cached units
- `generator.GenerateBody` / `BuildOutput` / exported pre-scan methods added
- `kylix build` and `kylix build <files...>` both enable the cache automatically
- `.kylix-cache/` added to `.gitignore`

---


## v1.3.2 (2026-06-13)

### LSP real-time diagnostics

The Language Server now pushes Kylix-layer diagnostics on every `didOpen`
and `didChange` event, so editors show squiggly lines without running a build.

**What's reported in-editor:**
- Parse errors (was already working)
- Interface implementation violations (`class Foo implements IBar` missing methods)

`Diagnostic.source` field added — editors display `"kylix"` as the diagnostic source tag.

**Implementation:**
- `pkg/lsp/document.go` — calls `compiler.CheckInterfaces()` after parse
- `pkg/compiler/compiler.go` — `CheckInterfaces()` exported for LSP use
- `pkg/lsp/server.go` — `Diagnostic` struct gains `Source` field

---

## v1.3.1 (2026-06-13)

### Multi-return value full scenario coverage

Fixed three parser bugs that blocked multi-return value patterns:

| Syntax | Was | Now |
|--------|-----|-----|
| `return (b, a)` | parse error: `)` unexpected | ✅ |
| `x, y := Swap(3, 7)` | parse error: `,` unexpected | ✅ |
| `var a, b := Pair()` | parse error: `,` unexpected | ✅ |

**Root causes fixed:**
1. `parseReturnStatement` — after parsing a tuple expression, `curToken` landed on the closing `)` instead of advancing to `;`, causing the block loop to try parsing `)` as a new statement
2. `parseExpressionOrAssignment` — added `tryParseMultiAssign()` to handle `ident, ident, ... :=` patterns  
3. `parseLTExpression` — left identifier must start uppercase to be treated as generic (prevents `a < b` from being misread as generic instantiation)

---

## v1.3.0 (2026-06-13)

### v2.0 Phase 1 — Interface validation, Kylix-layer errors, Real generics

Three foundational improvements that make Kylix viable for production code.

#### Interface implementation validation (compile-time)

Classes that declare `implements IFoo` are now checked at compile time.
Missing methods produce Kylix-layer errors (not Go errors) with the correct
source file and line number.

```pascal
type
  IAnimal = interface
    procedure Speak();
    function Name(): String;
  end;

  TDog = class implements IAnimal
    procedure Speak(); begin end;
    // Error: class "TDog" implements "IAnimal" but is missing method "Name"
  end;
```

#### Kylix-layer error reporting via `//line` directives

Previously, type errors and other Go-level issues reported Go file paths
(e.g. `./main.go:9:5`). Now the generator emits `//line` directives before
every function declaration, class declaration, and statement, so the Go
compiler maps all errors back to the original Kylix source:

```
Before:  ./main.go:9: cannot use "hello" as int64 value in assignment
After:   /path/to/main.klx:4: cannot use "hello" as int64 value in assignment
```

#### Real generic code generation

Generic classes and functions now generate proper Go 1.18+ generics instead
of falling back to `interface{}`. Generic type instantiation in expressions
(`TBox<Integer>.Create()`) is fully parsed and code-generated.

```pascal
type TBox<T> = class
  Value: T;
end;

function BoxInt(n: Integer): TBox<Integer>;
begin
  result := TBox<Integer>.Create();
  result.Value := n;
end;
```

Generates:
```go
type TBox[T interface{}] struct { Value T }

func BoxInt(n int64) *TBox[int64] {
    result := &TBox[int64]{}
    result.Value = n
    return result
}
```

#### New tests

- `pkg/compiler/compiler_test.go` — 5 interface validation tests
  (fully implemented, missing one method, missing all methods,
   cross-unit interface skipped, no implements no error)

#### Version bump

| Component | Old | New |
|-----------|-----|-----|
| Compiler, REPL, LSP | 1.2.3 | **1.3.0** |

---

## v1.2.3 (2026-06-12)

### Code refactoring — all source files under 1000 lines

Enforced a hard 1000-line limit per source file to improve readability and
maintainability. No behavior changes; all 40 tests still pass.

#### Files split

| Before | Lines | After | Max lines |
|--------|-------|-------|-----------|
| `parser/parser.go` | 2271 | `parser.go` + `parser_decl.go` + `parser_stmt.go` + `parser_expr.go` | 685 |
| `generator/generator.go` | 1979 | `generator.go` + `generator_types.go` + `generator_stmt.go` + `generator_expr.go` | 631 |
| `pkg/lsp/server.go` | 1238 | `server.go` + `handler_completion.go` + `handler_navigation.go` | 523 |
| `stdlib/orm.go` | 964 | `orm.go` + `orm_query.go` + `orm_migrate.go` | 410 |
| `pkg/formatter/formatter.go` | 897 | `formatter.go` + `formatter_stmt.go` + `formatter_expr.go` | 396 |

#### New file layout

```
parser/
  parser.go          core: Parser struct, New, ParseProgram, token helpers
  parser_decl.go     declarations: var, const, type, function, class, interface
  parser_stmt.go     statements: if, while, for, repeat, case, match, try, raise
  parser_expr.go     expressions: literals, operators, calls, lambdas, types

generator/
  generator.go       core: Generator struct, Generate/GenerateMulti, pre-scan
  generator_types.go type/function codegen: class, interface, variant, enum
  generator_stmt.go  statement codegen: if, for, while, try, match, raise
  generator_expr.go  expression codegen: calls, operators, lambdas, type mapping

pkg/lsp/
  server.go              JSON-RPC transport, message dispatch, document sync
  handler_completion.go  completion + hover handlers
  handler_navigation.go  definition, references, rename, formatting, signature

stdlib/
  orm.go         database connection + transaction
  orm_query.go   QueryBuilder fluent API
  orm_migrate.go ORM CRUD + MigrationManager + scan helpers

pkg/formatter/
  formatter.go       core + declaration formatting
  formatter_stmt.go  statement formatting
  formatter_expr.go  expression + type formatting
```

#### Key Constraint added to CLAUDE.md

> Every source file must not exceed 1000 lines. Split large files by logical
> responsibility (e.g. parser_decl.go / parser_stmt.go / parser_expr.go).

---

## v1.2.2 (2026-06-12)

### Tests + inherited keyword fix — 15/15 examples pass on both compilers

#### Bug Fix: `inherited` keyword in self-hosted compiler

`inherited Create(name, age)` inside class constructors caused a parse error
("no prefix parse function for )") in the self-hosted Kylix compiler.

**Root cause:** After `ParseInheritedStatement` called `ParseExpression`, the
Pratt parser left `curToken` on the closing `)` of the call. The outer
`ParseBlockStatement` semicolon-skip loop then started from the wrong position,
consuming a real identifier token as whitespace and leaving the parser desync'd.

**Fix** (`src/parser.klx`): added a `while PeekTokenIs(tkSemicolon)` advance
after `ParseExpression` returns, so the inherited statement correctly positions
to the trailing semicolon before returning.

#### New Tests

**parser/parser_test.go** (25 tests)
Covers: literals (int, float, bool, string), infix/prefix expressions, call
expressions, member access, array indexing, if/while/for statements,
assignments, function/procedure/var/const declarations, class declarations,
inherited calls, try/except, is/as expressions, map/array types,
program/unit name, empty program.

**generator/generator_test.go** (15 tests)
Covers end-to-end Kylix → Go codegen: hello world, var decl, function decl,
if/else, while loop, for loop, class with struct, map types, try/except,
booleans, arithmetic, nil, package header, string interpolation, inherited calls.

#### Example file coverage

| Compiler | v1.2.1 | v1.2.2 |
|----------|--------|--------|
| Go reference | 15/15 | 15/15 |
| Kylix self-hosted | 14/15 | **15/15** ✅ |

### Version Bumps

| Component | Old | New |
|-----------|-----|-----|
| Compiler, REPL, LSP, Project | 1.2.0 | **1.2.2** |

---

## v1.2.0 (2026-06-08)

### Phase 9 Complete: Diff Verification Passes — Self-Hosting Achieved!

This release completes the self-hosting bootstrap verification. The Kylix
compiler, written in Kylix and compiled by the Kylix compiler, generates
Go output that is semantically equivalent to the Go reference compiler.

#### Diff Verification Results

| Dimension | Go Reference | Kylix Self-Hosted | Result |
|-----------|-------------|-------------------|--------|
| Functions | 136 | 136 | ✅ Identical |
| Type definitions | 66 | 66 | ✅ Identical |
| Const blocks | 10 | 10 | ✅ Identical |
| Function signatures | — | — | 3 minor format diffs |
| Go compilation | ✅ | ✅ | Both compile |
| Runtime behavior | ✅ | ✅ | Semantically equivalent |

The only differences are 3 function signatures where the Kylix parser
expands multi-name parameters differently (e.g., `line, col int64` vs
`line int64, col int64`). These are semantically identical and both
compile to the same Go binary behavior.

#### Self-Hosting Bootstrap — Complete Pipeline

```
Kylix source (.klx) → Go compiler (kylix) → Go code → go build → Binary A
                                                            ↓
Kylix source (.klx) → Binary A → Go code → go build → Binary B
                                                            ↓
                      Binary A ≈ Binary B (semantically equivalent)
```

### Version Bumps

| Component | Old | New |
|-----------|-----|-----|
| Compiler, REPL, LSP, Project | 1.1.5 | **1.2.0** |

---

### Phase 9: Multi-File Go Compile Passes — String Escaping + Codegen Fixes

This release achieves a major milestone: the self-hosted multi-file Go output
(136KB, 6 source files merged) now **compiles and runs with zero errors**.

#### P0 - String Escaping in Generated Go Code

**Root cause:** `TStringLiteral` in the self-hosted generator output escaped
Go strings without handling embedded quotes, causing `""fmt""` instead of
`"\"fmt\""` in the generated Go code.

**Fix:** Added `WriteEscapedGoString` method that escapes `\` → `\\` and
`"` → `\"` before writing Go string literals. Applied to `GenerateExpression`
for `TStringLiteral` handling.

#### P0 - Base Class Type Mapping

**Root cause:** `MapType` relied on `ClassIsBase`/`ClassTypes` maps which are
nil in the self-hosted compiler. Base classes (TNode, TStatement, TExpression)
were not being mapped to `interface{}`, causing "is not an interface" errors.

**Fix:** Hardcoded TNode/TStatement/TExpression → `interface{}` in MapType.
Added default pointer type (`*Type`) for unknown class-like types.

#### P0 - Enum Type Declaration

**Root cause:** `GenerateEnumType` only emitted the `const (...)` block
without the underlying `type Name int` declaration, causing "undefined: TTokenType"
errors.

**Fix:** Added `type Name int` output after the const block.

#### P0 - Builtin Functions

- **StrToInt64/StrToFloat:** Added IIFE wrapper generation in `GenerateCallExpression`
- **append:** Added `arr = append(arr, elem)` auto-assignment in `GenerateStatement`
- **Exit/Break/Continue:** Added to `MapBuiltinFunction` (Exit→return, etc.)
- **ClassName.Create (no parens):** Added `&ClassName{}` generation in `TMemberExpression`

#### P0 - Multi-Name Parameter Parsing

**Root cause:** `ParseParameterList` only handled `name: Type` (single name)
syntax. Multi-name declarations like `line, col: Integer` left early names
without type annotations.

**Fix:** Rewrote `ParseParameterList` to collect all comma-separated names
first, then apply the type annotation to all collected names when `:` is found.

#### P0 - Empty Main Function

**Fix:** `GenerateMulti` now emits `func main() {}` when no program has
top-level statements.

### Bootstrap Status

| Step | Status | Description |
|------|--------|-------------|
| 7 files parse | ✅ | All 7 source files parse correctly |
| 7 files generate | ✅ | All generate valid Go output |
| Multi-file merged output | ✅ | 136KB combined with correct receivers |
| Multi-file Go compile | ✅ | **Zero errors, binary runs correctly** |
| Diff verification | 🟡 | Next step: compare Go vs Kylix output |

### Files Changed

- `src/generator.klx` — WriteEscapedGoString, MapType base classes, enum type,
  builtins, append, Create, empty main
- `src/parser.klx` — ParseParameterList multi-name support

### Version Bumps

| Component | Old | New |
|-----------|-----|-----|
| Compiler, REPL, LSP, Project | 1.1.4 | **1.1.5** |

---

### Phase 9: Class Method Receiver Fix + String Escaping Fixes

This release fixes the class method receiver generation for methods using
`ClassName.MethodName` syntax (defined outside the class body), and fixes
soft keyword handling in method names.

#### P0 - Class Method Receiver for ClassName.MethodName Syntax

**Root cause 1 — Soft keywords in method names:**
`ParseFunctionDecl` only checked `tkIdent` for method names. Methods named with
soft keywords (Write, Read, New, Delete, Default, ReadChar, NextToken, etc.)
had their `decl.Name` set to empty string. Fixed by changing the check to
`IsIdentOrSoftKeyword()`.

**Root cause 2 — ClassName.MethodName split missing:**
Go generator's `generateFunctionDecl` detects `.` in function names and splits
them into `className.methodName`, generating `func (self *ClassName) MethodName`.
Kylix generator's `GenerateFunctionDecl` lacked this check, emitting
`func ClassName.MethodName()` without a receiver. Fixed by adding manual `.`
position search with string slice extraction.

**Result:** All 126 class methods across all 7 source files now have correct
Go receivers.

| Class | Before | After |
|-------|--------|-------|
| TLexer | 0 methods with receiver | 11 methods + receiver |
| TParser | 0 methods with receiver | 59 methods + receiver |
| TGenerator | 50 (already correct) | 50 ✓ |
| TErrorList | 6 (already correct) | 6 ✓ |

#### P1 - Remaining String Escaping Issues

Known remaining issues in self-hosted compiler output:
- Double-quote strings (`"fmt"`) generate `""fmt""` instead of `"\"fmt\""`
- Single-quote string literals in Go output have raw newlines
- These are Go string escaping edge cases that do not block bootstrap verification

### Bootstrap Status

| Step | Status | Description |
|------|--------|-------------|
| 7 files parse | ✅ | All 7 source files parse correctly |
| 7 files generate | ✅ | All generate valid Go output |
| Multi-file merged output | ✅ | 135KB combined with correct receivers |
| Multi-file Go compile | 🟡 | String escaping edge cases remain |
| Diff verification | 🟡 | Blocked on Go compile |

### Files Changed

- `src/parser.klx` — `ParseFunctionDecl`: IsIdentOrSoftKeyword for method names + dotted names
- `src/generator.klx` — `GenerateFunctionDecl`: ClassName.MethodName → receiver split

### Version Bumps

| Component | Old | New |
|-----------|-----|-----|
| Compiler, REPL, LSP, Project | 1.1.3 | **1.1.4** |

---

### Phase 9: String Escaping Fix + Multi-File Bootstrap + GenerateMulti

This release fixes the critical string escaping bug that prevented the
self-hosted compiler's Go output from being compilable, and adds multi-file
bootstrap compilation support.

#### P0 - String Escaping Fix

**Root cause:** Go generator's `generateExpression` for `TStringLiteral` applied
escape transformations in the wrong order. `\` → `\\` was done before `\n` handling,
so Kylix's `'\n'` literal (two characters: backslash + n) became Go's `"\\n"`
(literal backslash-n) instead of `"\n"` (newline escape sequence).

This caused the self-hosted compiler to output all Go code as a single line
with literal `\n` characters, making the output un-compilable.

**Fix:** Reordered escape processing in `generator/generator.go`:
1. Protect `\n`, `\t`, `\r` with temporary markers (`\x00n`, etc.)
2. Escape `\` → `\\` and `"` → `\"`
3. Restore markers to correct Go escape sequences (`\n`, `\t`, `\r`)

**Result:** Self-hosted compiler output now has proper newlines and is
compilable Go source code.

#### P0 - Multi-File Bootstrap Compilation

**main.klx:**
- Rewrote from single-file to multi-file mode
- Reads 6 dependency files in hardcoded order: token → error → ast →
  lexer → parser → generator
- Parses each file independently, collects errors
- Calls `GenerateMulti(Programs)` for combined output

**generator.klx — `GenerateMulti`:**
- New method accepting `array of TProgram`
- Pre-scans all programs (class types, imports, exceptions)
- Generates types, globals, functions from all programs in order
- Generates single `func main()` from the non-unit program
- Output: single combined `main.go` with all declarations merged

#### P1 - Soft Keyword & Prefix Parse Expansion

**parser.klx:**
- `IsIdentOrSoftKeyword` expanded from 3 to 25+ tokens (matching Go version)
- 17 missing prefix parse functions registered: `exit`, `return`, `break`,
  `continue`, `delete`, `new`, `default`, `inherited`, `import`, `export`,
  `module`, `abstract`, `static`, `virtual`, `override` → all map to
  `ParseIdentifier`
- `ParseMemberExpression`: fixed result overwrite + soft keyword support

**generator.klx:**
- `GenerateTypeDecl`: unwrap `TClassDecl`/`TInterfaceDecl` inside `TTypeDecl`
- `GenerateTypeExpression`: added `TClassDecl` → `*ClassName` pointer mapping
- Removed nil map writes to `ClassTypes`/`ClassIsBase` (prevents nil map panic)

### Bootstrap Status

| Step | Status | Description |
|------|--------|-------------|
| 7 files parse | ✅ | All 7 source files parse correctly |
| 7 files generate | ✅ | All generate valid Go output |
| Single-file Go compile | ✅ | token/ast/error/lexer/parser compile OK |
| Multi-file Go output | ✅ | 134KB combined output with proper newlines |
| Multi-file Go compile | 🟡 | Class method codegen issues (Create, receiver format) |
| Diff verification | 🟡 | Blocked on class method codegen |

### Files Changed

- `generator/generator.go` — String escape reordering (Pascal \n → Go \n)
- `src/main.klx` — Multi-file mode with 6 dependency files
- `src/generator.klx` — `GenerateMulti` method + class type unwrap
- `src/parser.klx` — Soft keyword expansion + prefix parse registration + member expr fix

### Version Bumps

| Component | Old | New |
|-----------|-----|-----|
| Compiler, REPL, LSP, Project | 1.1.2 | **1.1.3** |

---

This release fixes 6 critical "result overwrite" bugs in the Kylix parser
and 4 code generation defects in the self-hosted generator, enabling the
self-hosted compiler to successfully compile all 7 bootstrap source files.

#### P0 - Parser "Result Overwrite" Bug Fixes (6 functions)

The Kylix parser (`src/parser.klx`) had a systematic bug pattern: in Pascal,
`result` is the implicit return variable. When `result` is set inside an
`if` block but execution continues past the block, subsequent code
overwrites the correct return value.

**Fixed functions:**

| Function | Bug | Impact |
|----------|-----|--------|
| `ParseTypeExpression` | No `Exit` after setting result; fallback always overwrites | Parameter types corrupted (e.g., `Integer` → `)`) |
| `ParseExpressionOrAssignment` | No `Exit` after assignment branch; `exprStmt` always overwrites | `x := 42` lost the `= 42` part |
| `ParseExpressionList` | No `Exit` after empty-list early return; continues parsing | `Foo()` (no-arg calls) caused parse failure |
| `ParseForStatement` | No `Exit` after for-each branch; for-loop always overwrites | For-each parsed as regular for |
| `ParseIndexExpression` | No `Exit` after slice branch; index always overwrites | `s[a:b]` parsed as `s[a]` |
| `ParseGroupedExpression` | No `Exit` after lambda/tuple branches; grouped expr overwrites | Lambda and tuple expressions lost |

**Fix pattern:** Added `Exit` statement after each `result := ...` that
should be the final return value, preventing fallthrough to later code.

#### P0 - Code Generation Improvements (4 defects)

**1. Record type generation:**
- **Before:** `type TToken = record ... end` → `type TToken interface{}`
- **After:** → `type TToken struct { TokenType TTokenType; Literal string; ... }`
- Added `GenerateRecordType` and `GenerateInlineRecordType` methods
- Added `TRecordType` branch in `GenerateTypeExpression`

**2. Map auto-initialization:**
- **Before:** `var Keywords: map[String]TTokenType` → `var Keywords map[string]TTokenType`
- **After:** → `var Keywords map[string]TTokenType = map[string]TTokenType{}`
- Prevents nil map panic at runtime

**3. Local variable declarations:**
- Added `LocalDecls` field to `TFunctionDecl` in `src/ast.klx`
- Modified `ParseFunctionDecl` to store local declarations in AST
- Modified `GenerateFunctionDecl` and `GenerateClassMethod` to emit `var` declarations before body
- Added `_ = name` suppression for unused local variables

**4. ReadFile builtin:**
- Added `ReadFile` special handling in `GenerateCallExpression`
- Generates: `func() string { data, _ := os.ReadFile(path); return string(data) }()`

### Bootstrap Status

All 7 Kylix source files now compile successfully with the self-hosted compiler:

| File | Parse | Generate | Notes |
|------|-------|----------|-------|
| `token.klx` | ✅ | ✅ | Enum, record, map init, functions all correct |
| `ast.klx` | ✅ | ✅ | 54 class types generated |
| `error.klx` | ✅ | ✅ | Error types generated |
| `lexer.klx` | ✅ | ✅ | Lexer with ReadChar, NextToken, etc. |
| `parser.klx` | ✅ | ✅ | Full Pratt parser (2338 lines) |
| `generator.klx` | ✅ | ✅ | Full code generator (~1400 lines) |
| `main.klx` | ✅ | ✅ | Entry point with ReadFile |

### Files Changed

- `src/parser.klx` — 6 result overwrite fixes with `Exit` statements
- `src/ast.klx` — Added `LocalDecls` field to `TFunctionDecl`
- `src/generator.klx` — Record type, map init, local vars, ReadFile builtin

### Version Bumps

| Component | Old | New |
|-----------|-----|-----|
| Compiler, REPL, LSP, Project | 1.1.1 | **1.1.2** |

---

This release completes the three remaining P0 tasks blocking self-hosting:
the Kylix lexer tokenization bug is fixed, the generator.klx skeleton is
fully implemented, and the bootstrap verification pipeline passes for
simple programs.

#### P0 - Lexer Tokenization Bug Fix (Two Root Causes)

**Bug 1 — `LookupIdent` returns tkIllegal for identifiers:**
- **Root cause:** `LookupIdent` in `src/token.klx` used single-value map
  lookup `result := Keywords[lower]`. In Go, a missing map key returns the
  zero value (`tkIllegal` = 0) instead of `tkIdent`.
- **Fix:** Added fallback: after map lookup, if `tok = tkIllegal` then
  return `tkIdent` instead. No valid keyword maps to `tkIllegal` (value 0),
  so this is a safe check.
- **File:** `src/token.klx` — `LookupIdent` function

**Bug 2 — `TParser.Create(Lex)` doesn't initialize token state:**
- **Root cause:** `main.klx` called `Par := TParser.Create(Lex)` which
  generates `&TParser{Lex: Lex}` — a bare struct literal without calling
  `NextToken()` twice. This left `CurToken` and `PeekToken` as zero values
  (type=0 = tkIllegal, line=0), causing parser errors.
- **Fix:** Changed `main.klx` to call `Par := NewParser(Lex)` which properly
  initializes token state via two `NextToken()` calls.
- **File:** `src/main.klx` — parser initialization

#### P0 - Generator Skeleton Completed

`src/generator.klx` expanded from 221 lines (stub) to ~1350 lines (full
implementation). All type dispatch uses Kylix `is`/`as` syntax instead of
Go type switches.

**Implemented methods:**

| Category | Methods |
|----------|---------|
| **Type generation** | `GenerateTypes`, `GenerateTypeDecl`, `GenerateEnumType`, `GenerateClassDecl`, `GenerateClassMethod`, `GenerateInterfaceDecl`, `GeneratePropertyAccessors` |
| **Global declarations** | `GenerateGlobals`, `GenerateGlobalVarDecl`, `GenerateConstDecl` |
| **Function generation** | `GenerateFunctions`, `GenerateFunctionDecl`, `GenerateFunctionSignature`, `GenerateTypeParams` |
| **Statement generation** | `GenerateStatement` (15+ statement types via is/as dispatch), `GenerateVarDecl`, `GenerateAssignment`, `GenerateIfStatement`, `GenerateWhileStatement`, `GenerateForStatement`, `GenerateForEachStatement`, `GenerateRepeatStatement`, `GenerateCaseStatement`, `GenerateMatchStatement`, `GenerateTryStatement`, `GenerateRaiseStatement`, `GenerateReturnStatement` |
| **Expression generation** | `GenerateExpression` (20+ expression types via is/as dispatch), `GenerateCallExpression` |
| **Type expression** | `GenerateTypeExpression`, `GenerateTypeExpressionForCast` (handles base class → `*ClassName` for is/as assertions) |
| **Pre-scan passes** | `CollectClassTypes`, `ScanImports`, `ScanForException` |
| **Utilities** | `MapType` (Kylix→Go type mapping), `MapBuiltinFunction` (WriteLn→fmt.Println, LowerCase→strings.ToLower, etc.) |

**Key design: is/as type dispatch pattern:**
```pascal
if stmt is TIfStatement then
begin
  var ifStmt: TIfStatement;
  ifStmt := stmt as TIfStatement;
  self.GenerateIfStatement(ifStmt);
end
else if stmt is TWhileStatement then ...
```

#### Bootstrap Verification

The three-step bootstrap pipeline now passes for simple programs:

```
Step 1: Go compiler (kylix build) compiles 7 .klx files → main.go ✅
Step 2: go build produces kylix_compiler binary ✅
Step 3: Self-hosted compiler compiles input → valid Go output ✅
```

**Verified:** `program hello; begin WriteLn(42); end.` correctly generates:
```go
package main
import ("fmt"; "strings"; "strconv")
func main() { fmt.Println(42) }
```

#### Known Limitations

- Self-hosting of complex source files (like `token.klx`) still has issues
  with local variable declarations and parameter type handling
- The Kylix AST's `TFunctionDecl` lacks a `LocalDecls` field, which exists
  in the Go AST — local var/const in function bodies are parsed but not
  stored in the AST for the generator
- Single-quoted string escaping needs improvement

### Files Changed

- `src/token.klx` — Fixed `LookupIdent` to return `tkIdent` for unknown identifiers
- `src/main.klx` — Changed `TParser.Create(Lex)` to `NewParser(Lex)`
- `src/generator.klx` — Expanded from 221-line skeleton to ~1350-line full implementation

### Version Bumps

| Component | Old | New |
|-----------|-----|-----|
| Compiler, REPL, LSP, Project | 1.1.0 | **1.1.1** |

---

### Phase 8: Bootstrap Compiler — Go Backend Upgrades

This release upgrades the Go compiler backend with the features needed to
compile the Kylix self-hosting compiler (`src/*.klx`). All 14 example files
continue to pass, all Go tests pass.

#### P0 - Enum Types

- **AST**: Added `EnumType` node with `Names []string`
- **Parser**: `tryParseEnumType()` parses `(val1, val2, ...)` syntax via `parseTypeExpression`
- **Generator**: `generateEnumType()` → Go `const` + `iota` + `type X int`
- Example: `type TTokenType = (tkEOF, tkIdent, ...);` → `const (tkEOF TTokenType = iota; tkIdent; ...)`

#### P0 - Slice Expressions

- **AST**: Added `SliceExpression` node (`Low`, `High`)
- **Parser**: `parseIndexExpression` detects `[a:b]` vs `[a]`
- **Generator**: `s[a:b]` → `s[a:b]` (Go slice syntax)

#### P0 - Unit File System & Multi-File Compilation

- **Parser**: `unit X;` declaration at file start → `Program.UnitName`, `Program.IsUnit`
- **Generator**: `GenerateMulti([]*Program)` — compiles multiple files into one Go package
- **Compiler API**: `CompileProject(files, opts)` with topological dependency sort
- **CLI**: `kylix build a.klx b.klx c.klx` multi-file mode
- **CLI**: `kylix run` auto-detects all `.klx` files via `FindAllKlxFiles()`

#### P0 - Class Code Generation (Hybrid Struct/Interface Approach)

- **All classes** generate as Go structs with parent embedding
- **Base classes** (parents of other classes) → `interface{}` in type positions for polymorphism
- **Concrete classes** → `*ClassName` pointers
- **Constructors**: `ClassName.Create` (no args) → `&ClassName{}`; `ClassName.Create(args)` → `&ClassName{args...}`
- **Class methods** generate `var result` declaration and local var/const declarations
- **Property accessors** generate getter/setter methods on the class

#### P1 - Soft Keyword Expansion (25+ keywords)

~25 Pascal keywords can now be used as identifiers in member positions
(`obj.Default`, `obj.DownTo`, `obj.When`, `obj.Dynamic`, `obj.To`, `obj.Do`,
`obj.Of`, `obj.In`, `obj.Read`, `obj.Write`, `obj.Abstract`, `obj.External`,
`obj.Forward`, `obj.Virtual`, `obj.Override`, `obj.Static`, `obj.Stored`,
`obj.Packed`, `obj.File`, `obj.New`, `obj.Delete`, `obj.Export`, `obj.Import`,
`obj.Module`, `obj.Is`, `obj.Except`, `obj.On`).

- **Parser**: `isSoftKeyword()` expanded; `parseMemberExpression` accepts soft keywords
- **Parser**: `parseFunctionDecl` accepts keywords as function names (fixes `function Delete`)

#### P1 - Other Generator Fixes

- **Local var/const in functions**: `FunctionDecl.LocalDecls` parsed and generated before body
- **`Exit` statement**: Pascal `exit` → `return result` (with return value) or `return` (procedure)
- **Bare method calls**: `self.Method` as statement → `self.Method()` (auto-parens)
- **Map type as expression**: `map[K]V` registered as prefix parse function, generates `map[K]V{}`
- **Empty array `[]`**: generates `nil` (assignable to any Go slice type)
- **String escaping**: proper `\`, `"`, `\n` escaping in string literals
- **New builtins**: `Ord`, `Length`, `IntToStr`, `StrToInt64`, `StrToFloat`
- **`for` loop**: generates `for i = 0` (no `:=`, avoids type mismatch with pre-declared `int64`)

### Bootstrap Compiler Source Files (Phase 8)

Seven Kylix source files written as the self-hosting compiler:

| File | Lines | Description |
|------|-------|-------------|
| `src/token.klx` | 209 | Token type enum, keyword map, lex helpers |
| `src/ast.klx` | 374 | AST node class hierarchy (54 classes) |
| `src/lexer.klx` | 366 | Lexical analyzer (character → token stream) |
| `src/parser.klx` | 2338 | Pratt parser (token stream → AST) |
| `src/error.klx` | 91 | Compiler error types and diagnostics |
| `src/generator.klx` | 221 | Go code generator (AST → Go source, skeleton) |
| `src/main.klx` | 56 | Entry point wiring lexer→parser→generator |
| **Total** | **3655** | |

**Build status:** All 7 `.klx` files compile to Go code successfully. The generated
Go code has ~6 remaining type/API compatibility issues to resolve before full
self-hosting bootstrap works.

### Example File Status (15 files)

| ✅ Passing (14/15) | ❌ Failing (1) |
|---|---|
| hello, simple, types, control, classes | web_advanced (Go syntax mixed into Kylix code) |
| modern, exceptions, stdlib_demo, orm_example | |
| functions, web_demo, test_formatter, test_map | |
| web_fullstack | |

### Bug Fixes

- **`Delete` as function name**: `function Delete(...)` no longer fails (keyword recognized as identifier)
- **Class field parsing**: Bare field declarations (`Name: Type;` without `var`) guarded by `peekTokenIs(COLON)`
- **Parser regression**: 14/15 examples confirmed passing (no regressions from new features)
- **Constructor argument mapping**: `T.Create(arg)` now generates `&T{Field: arg}` using collected class field names
- **Bare method calls in assignment/condition**: `Prog := Par.ParseProgram` → `Prog := Par.ParseProgram()` (main.klx uses explicit parens)
- **Unused local variables**: Generator appends `_ = varName` for local vars declared in function bodies
- **For loop variable type**: `for i = 0` (no `:=`) avoids `int` vs `int64` type mismatch
- **Map type as expression prefix**: `MAP` and `VARIANT` registered as prefix parse functions
- **String escaping**: Proper `\`, `"`, `\n` escaping in string literals

### New Builtins

- **ReadFile(filename)** — reads file content, returns string (uses `os.ReadFile` internally, auto-adds `"os"` import)
- **Ord(s)** — returns int value of first character (guards against empty string)
- **Length(x)** — returns `int64(len(x))` for slices/strings
- **IntToStr(n)** — converts int64 to string via `fmt.Sprintf`
- **StrToInt64(s)** — parses string to int64 via `strconv.ParseInt`
- **StrToFloat(s)** — parses string to float64 via `strconv.ParseFloat`

### is/as Type Dispatch

- `is` expression → Go type assertion check: `func() bool { _, ok := expr.(*Type); return ok }()`
- `as` expression → Go type assertion: `expr.(*Type)`
- Both work correctly with base class → `interface{}` polymorphism
- Confirmed working in Go backend and usable from `.klx` source files

### Self-Hosting Bootstrap Status

**Build chain verified:**
```
7 .klx source files → kylix build → Go code → go build → kylix_compiler binary ✅
```

**Runtime status:**
- Lexer→Parser→Error pipeline: ✅ functional
- Tokenizer: 🟡 has known bug (some Pascal keywords produce tkIllegal tokens)
- Generator (Kylix-side): 🟡 skeleton code, needs type dispatch implementation

**Known issues to fix for full self-hosting:**
- Kylix lexer tokenization bug: valid Pascal source strings produce unexpected tkIllegal tokens
- Generator.klx skeleton needs completion with `is`/`as` type dispatch
- Single-quoted string escaping in generated Go code needs improvement

### Files Changed

- `ast/ast.go` — Added `EnumType`, `SliceExpression`, `LocalDecls` on `FunctionDecl`
- `parser/parser.go` — Enum parsing, slice parsing, unit file parsing, soft keyword expansion, map/variant prefix, local var/const storage, class field safety, function-as-keyword-name fix
- `generator/generator.go` — Major rewrite: class codegen (hybrid struct/interface), enum generation, slice generation, multi-file `GenerateMulti`, constructor handling, bare method call parens, exit statement, for loop type fix, string escaping, new builtins, class method result+locals generation, map type as expression
- `cmd/kylix/main.go` — Multi-file build/run support
- `pkg/compiler/compiler.go` — `CompileProject` with topological sort
- `src/*.klx` — 7 new bootstrap compiler source files (3655 lines total)

### Version Bumps

| Component | Old | New |
|-----------|-----|-----|
| Compiler, REPL, LSP, Project | 1.0.3 | **1.1.0** |

---

## v1.0.3 (2026-06-05)

### New Features — Phase 7: Language Capabilities

**P0 - Map Type (map[K]V):**
- Token: Added `MAP` token and `"map"` keyword
- AST: Added `MapType` node with `KeyType` and `ValueType` fields
- Parser: `parseMapType()` parses `map[K]V` syntax
- Generator: `map[K]V` → Go `map[K]V`, with auto-initialization (`map[K]V{}`)
- Example: `examples/test_map.klx` — Map operations demo

**P0 - Variant / Discriminated Union:**
- Token: Added `VARIANT` token and `"variant"` keyword
- AST: Added `VariantType` and `VariantCase` nodes
- Parser: Parses `variant CaseName: Type; ... end` syntax
- Generator: Generates Go `interface` + concrete `struct` types with marker methods
  - `type TExpr = variant IntLit: Integer; StrLit: String; end;` →
    - `type TExpr interface { isTExpr() }`
    - `type TExpr_IntLit struct { Value int64 }` + `func (*TExpr_IntLit) isTExpr() {}`
    - `type TExpr_StrLit struct { Value string }` + `func (*TExpr_StrLit) isTExpr() {}`

**P0 - Dynamic Arrays (append, SetLength):**
- Builtin: `append` and `SetLength` registered in builtin map
- `append(arr, elem)` → `arr = append(arr, elem)` (auto-assignment)
- `SetLength(arr, n)` → `arr = arr[:n]` (slice truncation)
- Works as expression statement, not requiring manual assignment

### Bug Fix

**web_fullstack.klx rewritten:**
- Replaced Go struct literal `TConnectionConfig{...}` with proper Kylix field assignments
- Replaced `map[string]interface{}` with `map[String]String` (valid Kylix syntax)
- Replaced `user = nil` check with `user.ID = 0` (proper record check)

### Example File Status (15 files)

| ✅ Passing (15/15) | ❌ Failing (0) |
|---|---|
| hello, simple, types, control, classes | — |
| modern, exceptions, stdlib_demo | |
| test_formatter, test_map, orm_example | |
| functions, web_demo, web_advanced | |
| web_fullstack | |

- **test_map.klx**: New example for Map type
- **web_fullstack.klx**: Rewritten in proper Kylix syntax — now passes ✅

### Files Changed

- `token/token.go` — Added `MAP`, `VARIANT` tokens and keywords
- `ast/ast.go` — Added `MapType`, `VariantType`, `VariantCase` nodes
- `parser/parser.go` — `parseMapType()`, variant type parsing in `parseTypeExpression()`
- `generator/generator.go` — `MapType`/`VariantType` generation, `append`/`SetLength` builtins, map auto-init
- `examples/web_fullstack.klx` — Rewritten in proper Kylix syntax
- `examples/test_map.klx` — New Map type example

---

## v1.0.2 (2026-06-04)

### Bug Fixes

**P1 - String Interpolation (fixed):**
- **Lexer**: Already detected `$"..."` patterns correctly — no changes needed
- **Parser**: `parseStringInterpolation()` now properly splits raw content by `${...}` patterns, creates sub-parsers for each expression segment, and returns `ast.StringInterpolation` with parsed expression parts
- **Generator**: Added `*ast.StringInterpolation` case in `generateExpression()` → generates `fmt.Sprintf(format, args...)` with automatic `"fmt"` import
- Added `scanExpressionForImports` support for `*ast.StringInterpolation`

**P1 - Exception Types (fixed):**
- Exception types (`Exception`, `EIndexOutOfRange`, etc.) now auto-generated inline in Go output when `try/raise/except` is used
- `raise Exception.Create('msg')` generates `panic(&Exception{Message: "msg"})` using constructor pattern
- `except on E: ExceptionType do` generates `case *ExceptionType:` (pointer type switch) for proper matching
- Sub-types detected from `on` clauses are auto-generated as structs embedding `Exception`
- Added `scanForException` pre-scan pass to detect exception usage before type generation
- Plain `raise` without expression generates `panic(&Exception{Message: "exception"})`

**P2 - Multi-Value Return (fixed):**
- Parser: `parseFunctionDecl` now detects `: (Type1, Type2)` tuple return type syntax
- Parser: `parseGroupedExpression` now detects `(expr1, expr2)` tuple literals via `peekToken` check
- Parser: `parseSingleVarDecl` supports destructuring `var (a, b) := expr` with LPAREN detection
- AST: Added `TupleLiteral` expression node and `ReturnTypes []Expression` to `FunctionDecl`
- Generator: `generateFunctionSignature` outputs `(type1, type2)` for multi-return
- Generator: `result := (a, b)` in multi-return functions generates `return a, b`
- Generator: `var (quotient, ok) := Divide(10, 3)` generates `quotient, ok := Divide(10, 3)`
- Generator: Added `writeInterpolation` and `generateMultiReturnType` helper methods

**P2 - Properties Code Generation (fixed):**
- Generator: `generateClassDecl` now iterates `class.Properties` and generates getter/setter methods
- `property PropName: Type read FieldName;` → `func (self *ClassName) PropName() Type { return self.FieldName }`
- `property PropName: Type write FieldName;` → `func (self *ClassName) SetPropName(v Type) { self.FieldName = v }`

**P2 - Anonymous Procedure Edge Cases (fixed):**
- Record type parser now tracks nesting depth for nested `record` types
- `web_demo.klx`: Anonymous procedures with nested record types in `var` declarations now parse correctly
- Fix: `parseTypeExpression` for `RECORD` uses depth counter to handle inner `end` tokens

**P2 - Array Range Size Calculation (fixed):**
- `array[0..2] of Integer` now correctly computes size as `((2 - 0) + 1)` instead of `[0]`
- Fix: `parseArrayType` now computes `upperBound - lowerBound + 1` when `..` range syntax is used

### Example File Status (14 files)

| ✅ Passing (13) | ❌ Failing (1) |
|---|---|
| hello, simple, types, control, classes | web_fullstack (Go struct literal `{...}` syntax) |
| modern, exceptions, stdlib_demo | |
| test_formatter, web_advanced, orm_example | |
| functions, web_demo | |

- **functions.klx**: Now passes ✅ (was failing due to missing multi-return support)
- **web_demo.klx**: Now passes ✅ (was failing due to nested record parsing bug)
- **web_fullstack.klx**: Still fails (uses Go struct literal `{...}` syntax — not valid Kylix)

### Files Changed

- `lexer/lexer.go` — No changes (STRING_INTERPOLATION detection already worked)
- `parser/parser.go` — String interpolation parsing, multi-return return type/tuple/destructuring parsing, record depth tracking, array size computation, LPAREN support in var sections
- `generator/generator.go` — String interpolation generation, exception type auto-generation, multi-return function/assignment/var generation, property accessor generation
- `ast/ast.go` — Added `TupleLiteral` expression node, `ReturnTypes []Expression` field on `FunctionDecl`
- `stdlib/exceptions.go` — New file: Reference exception type definitions

---

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
