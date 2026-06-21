# Kylix v3.0.0-alpha Complete Tutorial

Welcome to the complete Kylix tutorial! This tutorial covers all working features in Kylix v3.0.0-alpha with tested, runnable examples.

## What is Kylix?

Kylix is a modern Pascal-to-Go transpiler that brings modern language features to the Pascal syntax. Write Pascal, get Go performance.

## Prerequisites

- Kylix compiler (v3.0.0-alpha or later)
- Go 1.18+ (for running generated code)

## Tutorial Structure

This tutorial contains **23 working examples** organized into 10 categories:

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

### 3. Functions (3 examples) - `03_functions/`
- `example13_functions.klx` - Functions and procedures
- `example14_recursion.klx` - Recursive functions
- `example16_multireturn.klx` - Multiple return values

### 4. Generics (1 example) - `05_generics/`
- `example21_generic_class.klx` - Generic stack class

### 5. Advanced Types (3 examples) - `06_advanced_types/`
- `example22_records.klx` - Record types
- `example23_arrays.klx` - Fixed arrays
- `example24_map.klx` - Map type (hash tables)

### 6. Core Functions (1 example) - `07_stdlib_core/`
- `example29_basic_funcs.klx` - Max, Min, Abs functions

### 7. Exceptions (2 examples) - `10_exceptions/`
- `example27_try_except.klx` - Try-except blocks
- `example28_finally.klx` - Try-finally and try-except-finally

### 8. Modules (2 examples) - `11_modules/`
- `math_helper.klx` - Unit definition
- `example33_use_module.klx` - Using units with `uses`

## How to Run Examples

### Single File

```bash
cd /tmp/kylix_complete/01_basics
kylix build example01_hello.klx
go run example01_hello.go
```

### Multi-File (Modules)

```bash
cd /tmp/kylix_complete/11_modules
kylix build math_helper.klx example33_use_module.klx
go run main.go
```

### All Examples in a Category

```bash
cd /tmp/kylix_complete/02_control_flow
for f in example*.klx; do
  echo "=== $f ==="
  kylix build "$f"
  go run "${f%.klx}.go"
  echo ""
done
```

## Language Features Reference

### Variables and Types

```pascal
var x: Integer;           // Integer type
var name: String;         // String type
var pi: Real;            // Float type
var active: Boolean;     // Boolean type

var count := 42;         // Type inference
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

## Known Limitations (v3.0.0-alpha)

### Not Working / Buggy
- **OOP (Classes)**: Field access in methods generates incorrect Go code (missing `self.` prefix)
- **Lambda expressions**: Type declarations fail to compile
- **Match expressions**: Syntax not fully implemented
- **Enum types**: Parse errors with enum syntax
- **String interpolation**: Not processed (prints literal `${var}`)
- **Multi-line comments**: `{ }` and `(* *)` syntax not supported
- **Write() function**: Only `WriteLn()` is available
- **For..in loops**: Limited support for iterating collections

### Working Features
✅ Basic types (Integer, String, Real, Boolean)
✅ Type inference with `:=`
✅ All control flow (if, while, for, repeat, case)
✅ Functions and procedures
✅ Recursion
✅ Multi-return values
✅ Arrays (fixed size)
✅ Map types
✅ Record types
✅ Generic classes
✅ Exception handling (try/except/finally)
✅ Modules (unit/uses)
✅ Operators (arithmetic, comparison, logical)

## Tips and Best Practices

1. **Always use `begin`/`end` blocks** for multi-statement bodies
2. **Declare variables before use** - either with `var` or with type inference `:=`
3. **Use `result :=`** in functions to set return value
4. **Multi-return requires pre-declared variables** - can't use `:=` with tuple assignment
5. **Arrays are 0-indexed** in Kylix
6. **Maps auto-initialize** - no need for explicit initialization
7. **Only single-line comments work** - use `//` not `{ }` or `(* *)`

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
kylix build hello.klx
go run hello.go
```

## Further Learning

- Official website: [kylix.top](https://kylix.top)
- GitHub: [Kylix repository](https://github.com/your-repo/kylix)
- Check `CHANGELOG.md` for version-specific features
- Read `ROADMAP.md` for upcoming features

## Example Categories Summary

| Category | Examples | Status |
|----------|----------|--------|
| Basics | 6 | ✅ All work |
| Control Flow | 5 | ✅ All work |
| Functions | 3 | ✅ All work |
| OOP | 0 | ❌ Compiler bugs |
| Generics | 1 | ✅ Works |
| Advanced Types | 3 | ✅ All work |
| Core Functions | 1 | ✅ Works |
| Exceptions | 2 | ✅ All work |
| Modules | 2 | ✅ All work |
| **Total** | **23** | **21 fully working** |

Happy coding with Kylix! 🚀
