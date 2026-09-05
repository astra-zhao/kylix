# Kylix Complete Tutorial

Welcome to the complete Kylix tutorial! This tutorial covers all working features in Kylix **v0.6.9** with tested, runnable examples — **50 numbered examples across 20 chapters**, plus the `math_helper.klx` unit companion and a `test.klx` smoke file.

Test status (v0.6.9):

- ✅ Go backend sweep: **51/51** (`examples/complete-tutorial/test_all.sh`)
- ✅ LLVM backend sweep: **51/51** (`examples/complete-tutorial/test_all_llvm.sh`, native binaries, no Go at runtime)
- ✅ Bootstrap (self-hosted, no-Go) sweep: **50 PASS + 1 SKIP** (`scripts/test_bootstrap_all.sh`; example33 is multi-file and verified host-side)
- ✅ The compiler itself is written in Kylix and reaches an IR fixed point (gen1 ≡ gen2 ≡ gen3)

## What is Kylix?

Kylix is a modern Pascal compiler: by default it transpiles to readable Go code (compiled with `go build`), and with `--backend=llvm` it emits LLVM IR directly to produce a native binary — no Go toolchain needed at runtime.

## Prerequisites

- Kylix compiler (v0.6.9 or later)
- Either **Go 1.18+** (Go backend) or **LLVM** (`llc`/`clang`, native backend). Run `kylix doctor` to preflight the LLVM environment.
- Native backend links against system libraries for some stdlib modules (openssl, sqlite3, curl) as needed.

## Tutorial Structure

### 1. Basics (6 examples) - `01_basics/`
- `example01_hello.klx` - Hello World
- `example02_variables.klx` - Variable declarations and types
- `example03_constants.klx` - Constants
- `example04_type_inference.klx` - Type inference with `:=`
- `example05_operators.klx` - Arithmetic, comparison, logical operators
- `example06_comments.klx` - Single-line comments

### 2. Control Flow (5 examples) - `02_control_flow/`
- `example07_if_else.klx` - If-then-else statements
- `example08_while.klx` - While loops
- `example09_for_to.klx` - For loops (to/downto)
- `example10_repeat.klx` - Repeat-until loops
- `example11_case.klx` - Case statements

### 3. Functions (4 examples) - `03_functions/`
- `example13_functions.klx` - Functions and procedures
- `example14_recursion.klx` - Recursive functions
- `example15_lambda.klx` - Lambdas and closures
- `example16_multireturn.klx` - Multiple return values

### 4. OOP (4 examples) - `04_oop/`
- `example17_class_fields.klx` - Class fields
- `example18_class_methods.klx` - Class methods (`self.Field`)
- `example19_inheritance.klx` - Class inheritance
- `example40_declarative_oop.klx` - `var p := TPerson.Create` pattern with inheritance

### 5. Generics (1 example) - `05_generics/`
- `example21_generic_class.klx` - Generic stack class

### 6. Advanced Types (5 examples) - `06_advanced_types/`
- `example20_enum.klx` - Enum types
- `example22_records.klx` - Record types
- `example23_arrays.klx` - Fixed and dynamic arrays
- `example24_map.klx` - Map type (hash tables)
- `example25_string_ops.klx` - String operations

### 7. Core Functions (1 example) - `07_stdlib_core/`
- `example29_basic_funcs.klx` - Max, Min, Abs functions

### 8. stdlib Utils (4 examples) - `08_stdlib_utils/`
- `example36_sysutil.klx` - File system / environment utilities
- `example37_jsonutil.klx` - JSON encode/decode
- `example38_datetime.klx` - Date and time
- `example39_regex.klx` - Regular expressions

### 9. Exceptions (2 examples) - `10_exceptions/`
- `example27_try_except.klx` - Try-except blocks
- `example28_finally.klx` - Try-finally and try-except-finally

### 10. Modules (1 example + unit) - `11_modules/`
- `math_helper.klx` - Unit definition (`unit`/`interface`/`implementation`)
- `example33_use_module.klx` - Using units with `uses`

### 11. Special Features (7 examples) - `12_special_features/`
- `example41_attributes.klx` - `[Attribute]` annotation syntax (`[Controller]`, `[Get]`, `[Inject]`, `[Entity]`)
- `example42_kylixboot_autowire.klx` - KylixBoot `[Controller]` + `[Get]` auto route registration
- `example43_kylixboot_di.klx` - KylixBoot `[Service]` + `[Inject]` DI auto-wiring
- `example44_kylixboot_proc_handler.klx` - Procedure-style route handler
- `example45_validation_annotations.klx` - `[Required]`/`[Email]`/`[Min]`/`[MinLen]` field validators
- `example46_security_annotations.klx` - `[Authenticated]`/`[Role]` per-route security guards
- `example47_orm_annotations.klx` - ORM annotations (`[Entity]`/`[Repository]`/`[Query]`)

### 12. stdlib Phase 6 (1 example) - `13_stdlib_phase6/`
- `example48_phase6_net_crypto_encoding.klx` - SHA-256, Base64, BCrypt, CSV, HMAC, MD5

### 13. Request Body Binding (1 example) - `14_body_binding/`
- `example49_body_binding.klx` - `[Body(TEntity)]` JSON request body binding with `Validate()`/`IsValid()` checks

### 14. JWT Authentication (1 example) - `15_jwt/`
- `example50_jwt_auth.klx` - `JwtSign`/`JwtVerify` + `BootRegisterJwtAuth` for `[Authenticated]` route guards

### 15. OpenAPI / Swagger (1 example) - `16_openapi/`
- `example51_openapi.klx` - `[Controller]`/`[Get]`/`[Post]`/`[Body]`/`[Authenticated]`/`[Role]` → OpenAPI 3.1 YAML via `kylix doc --openapi`

### 16. Database (1 example) - `17_database/`
- `example52_database.klx` - SQLite in-memory DB with `DbOpenSQLite`/`DbExec`/`DbQueryScalar`/`DbQueryRows`, parameterized queries

### 17. Cache (1 example) - `18_cache/`
- `example53_cache.klx` - Thread-safe LRU cache with `NewCache`/`Put`/`GetString`/`Has`/`Delete`/`Size`/`Clear`

### 18. HTTP Client (1 example) - `19_http/`
- `example54_http.klx` - `THttpClient` with GET/POST/PUT/DELETE, one-shot helpers, `THttpResponse` (status+body)

### 19. WebSocket (1 example) - `20_websocket/`
- `example55_websocket.klx` - RFC 6455 WebSocket client/server (`WsDial`/`WsAccept`/`WsSend`/`WsRecv`/`WsClose`), pure stdlib

### 20. Variant (2 examples) - `21_variant/`
- `example56_variant.klx` - Variant scalars and arrays (type-tagged runtime values)
- `example57_variant_map.klx` - `map[String]Variant` with type-tagged values (`row['col']` style access)

## How to Run Examples

### Single File

```bash
cd examples/complete-tutorial/01_basics
kylix build example01_hello.klx
go run example01_hello.go
```

Or with the native LLVM backend (no Go at runtime):

```bash
kylix run example01_hello.klx            # auto-detects backend
kylix build --backend=llvm example01_hello.klx   # force native binary
```

### Multi-File (Modules)

```bash
cd examples/complete-tutorial/11_modules
kylix build math_helper.klx example33_use_module.klx
go run main.go
```

### All Examples in a Category

```bash
cd examples/complete-tutorial/02_control_flow
for f in example*.klx; do
  echo "=== $f ==="
  kylix build "$f"
  go run "${f%.klx}.go"
  echo ""
done
```

### Full Sweep

```bash
# Go backend (51/51)
KYLIX=/path/to/kylix bash examples/complete-tutorial/test_all.sh

# LLVM backend (51/51, native binaries)
bash examples/complete-tutorial/test_all_llvm.sh
```

## Language Features Reference

### Variables and Types

```pascal
var x: Integer;           // Integer type
var name: String;         // String type
var pi: Real;             // Float type
var active: Boolean;      // Boolean type

var count := 42;          // Type inference
```

### Control Flow

```pascal
// If statement
if x > 5 then
  WriteLn('Greater')
else
  WriteLn('Not greater');

// While loop
while i < 10 do
begin
  i := i + 1;
end;

// For loop
for i := 1 to 10 do
  WriteLn(i);

// Repeat-until
repeat
  WriteLn(i);
  i := i - 1;
until i <= 0;

// Case statement
case day of
  1: WriteLn('Monday');
  2: WriteLn('Tuesday');
  6, 7: WriteLn('Weekend');
end;
```

### Functions

```pascal
function Add(a: Integer; b: Integer): Integer;
begin
  result := a + b;
end;

procedure Greet(name: String);
begin
  WriteLn('Hello, ', name);
end;

// Multiple return values
function DivMod(a: Integer; b: Integer): (Integer, Integer);
begin
  result := (a div b, a mod b);
end;

var q, r: Integer;
(q, r) := DivMod(17, 5);
```

### Arrays and Collections

```pascal
// Fixed array
var numbers: array[0..9] of Integer;
numbers[0] := 42;

// Map
var scores: map[String]Integer;
scores['Alice'] := 95;
WriteLn(scores['Alice']);
```

### Records

```pascal
type
  TPoint = record
    X: Real;
    Y: Real;
  end;

var point: TPoint;
point.X := 10.5;
point.Y := 20.3;
```

### Annotations

```pascal
[Controller('/api/users')]
type
  TUserController = class
    [Inject]
    UserRepo: TUserRepository;

    [Get('/')]
    function ListUsers(req: TRequest): TResponse;
    begin
      result := req.JSON(UserRepo.FindAll());
    end;
  end;
```

### Exception Handling

```pascal
try
begin
  // Code that might raise exception
  result := SafeDivide(10, 0);
end
except
begin
  WriteLn('Error occurred');
end
finally
begin
  WriteLn('Cleanup');
end
end;
```

### Modules (Units)

```pascal
// math_helper.klx
unit MathHelper;

interface
function Square(x: Integer): Integer;

implementation
function Square(x: Integer): Integer;
begin
  result := x * x;
end;
end.

// main.klx
program Main;
uses MathHelper;
begin
  WriteLn(Square(5));
end.
```

## Known Limitations (v0.6.9)

- **Windows**: `net` (Winsock) and `regex` (pcre2) are stubs — real implementations are planned with Windows machine verification.
- **Variant arithmetic** (`v + 1`) is LLVM-backend only; Go's `interface{}` cannot do operator arithmetic. Comparisons and printing work on both backends.
- **example33** (multi-file unit build) is exercised host-side; the no-Go bootstrap sweep reports it as SKIP.

## Tips and Best Practices

1. **Always use `begin`/`end` blocks** for multi-statement bodies
2. **Declare variables before use** - either with `var` or with type inference `:=`
3. **Use `result :=`** in functions to set return value
4. **Multi-return requires pre-declared variables** - can't use `:=` with tuple assignment
5. **Arrays are 0-indexed** in Kylix (with optional Pascal-style 1-based ranges)
6. **Maps auto-initialize** - no need for explicit initialization
7. **Use `self.Field` inside class methods** to access instance fields
8. **For class instance vars, both `var p: TPerson` and `var p := TPerson.Create` work**

## Quick Start Example

Create `hello.klx`:

```pascal
program Hello;

function Greet(name: String): String;
begin
  result := 'Hello, ' + name + '!';
end;

begin
  WriteLn(Greet('Kylix'));
  WriteLn('Welcome to modern Pascal!');
end.
```

Compile and run:

```bash
kylix run hello.klx
```

## Further Learning

- Official website: [kylix.top](https://kylix.top)
- GitHub: [Kylix repository](https://github.com/astra-zhao/kylix)
- Beginner-friendly walkthrough: [`docs/TUTORIAL_FOR_BEGINNERS_CN.md`](../../docs/TUTORIAL_FOR_BEGINNERS_CN.md)
- Check `CHANGELOG.md` for version-specific features
- Read `ROADMAP.md` for upcoming features

## Example Categories Summary

| Category | Examples | Status |
|----------|----------|--------|
| Basics | 6 | ✅ All work |
| Control Flow | 5 | ✅ All work |
| Functions (incl. lambda) | 4 | ✅ All work |
| OOP (incl. declarative) | 4 | ✅ All work |
| Generics | 1 | ✅ Works |
| Advanced Types | 5 | ✅ All work |
| Core Functions | 1 | ✅ Works |
| stdlib Utils (sysutil/jsonutil/datetime/regex) | 4 | ✅ All work |
| Exceptions | 2 | ✅ All work |
| Modules (unit/uses) | 1 + unit | ✅ Works |
| Annotations / KylixBoot / Validation / Security / ORM | 7 | ✅ All work |
| stdlib Phase 6 (crypto / encoding) | 1 | ✅ Works |
| Request Body Binding | 1 | ✅ Works |
| JWT Authentication | 1 | ✅ Works |
| OpenAPI / Swagger | 1 | ✅ Works |
| Database (SQLite) | 1 | ✅ Works |
| Cache (LRU) | 1 | ✅ Works |
| HTTP Client | 1 | ✅ Works |
| WebSocket | 1 | ✅ Works |
| Variant (scalars/arrays + map) | 2 | ✅ All work |
| **Total** | **50 + unit + smoke** | **Go 51/51 · LLVM 51/51 · bootstrap 50+1 SKIP** |

Happy coding with Kylix! 🚀
