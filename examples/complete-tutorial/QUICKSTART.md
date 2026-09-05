# Kylix Tutorial - Quick Start Guide

## Installation Check

```bash
kylix version
# Should show: Kylix 0.6.9 (or later)
```

Optional — preflight the native LLVM backend (no Go needed at runtime):

```bash
kylix doctor
```

## 5-Minute Quick Start

### 1. Hello World (30 seconds)

```bash
cat > hello.klx << 'ENDKLX'
program Hello;
begin
  WriteLn('Hello, Kylix!');
end.
ENDKLX

kylix run hello.klx
# or: kylix build hello.klx && go run hello.go
```

### 2. Variables and Functions (2 minutes)

```bash
cat > calc.klx << 'ENDKLX'
program Calculator;

function Add(a: Integer; b: Integer): Integer;
begin
  result := a + b;
end;

function Multiply(x: Integer; y: Integer): Integer;
begin
  result := x * y;
end;

var
  sum: Integer;
  product: Integer;

begin
  sum := Add(10, 20);
  product := Multiply(5, 6);

  WriteLn('10 + 20 = ', sum);
  WriteLn('5 * 6 = ', product);
end.
ENDKLX

kylix run calc.klx
```

### 3. Control Flow (2 minutes)

```bash
cat > loops.klx << 'ENDKLX'
program Loops;

var
  i: Integer;
  sum: Integer;

begin
  // For loop
  WriteLn('Counting:');
  for i := 1 to 5 do
    WriteLn(i);

  // While loop
  sum := 0;
  i := 1;
  while i <= 10 do
  begin
    sum := sum + i;
    i := i + 1;
  end;
  WriteLn('Sum of 1-10: ', sum);

  // Case statement
  var day := 3;
  case day of
    1: WriteLn('Monday');
    2: WriteLn('Tuesday');
    3: WriteLn('Wednesday');
  end;
end.
ENDKLX

kylix run loops.klx
```

## Run All Tutorial Examples

```bash
# Go backend (51/51)
KYLIX=/path/to/kylix bash examples/complete-tutorial/test_all.sh

# LLVM backend (51/51, native binaries)
bash examples/complete-tutorial/test_all_llvm.sh
```

## Learn by Category

### Beginners (Start Here)
```bash
cd examples/complete-tutorial/01_basics
# Run each example:
kylix run example01_hello.klx
kylix run example02_variables.klx
# ... continue through example06
```

### Intermediate
```bash
cd examples/complete-tutorial/02_control_flow
# All 5 control flow examples

cd examples/complete-tutorial/03_functions
# Functions, recursion, lambda, multi-return
```

### Advanced
```bash
cd examples/complete-tutorial/06_advanced_types
# Enum, arrays, maps, records, string ops

cd examples/complete-tutorial/10_exceptions
# Exception handling

cd examples/complete-tutorial/12_special_features
# Annotations, KylixBoot, validation, security, ORM
```

## Common Patterns

### Function with Return Value
```pascal
function Square(x: Integer): Integer;
begin
  result := x * x;
end;
```

### Procedure (No Return)
```pascal
procedure PrintInfo(name: String; age: Integer);
begin
  WriteLn('Name: ', name);
  WriteLn('Age: ', age);
end;
```

### Type Inference
```pascal
var count := 42;           // Integer
var message := 'Hello';    // String
var ratio := 3.14;         // Real
var active := true;        // Boolean
```

### Arrays and Maps
```pascal
var numbers: array[0..4] of Integer;
numbers[0] := 10;
numbers[1] := 20;

var scores: map[String]Integer;
scores['Alice'] := 95;
WriteLn(scores['Alice']);
```

### Multiple Return Values
```pascal
function DivMod(a: Integer; b: Integer): (Integer, Integer);
begin
  result := (a div b, a mod b);
end;

var q, r: Integer;
(q, r) := DivMod(17, 5);
```

### Exception Handling
```pascal
try
begin
  // risky code
end
except
begin
  WriteLn('Error!');
end
finally
begin
  WriteLn('Cleanup');
end
end;
```

## Next Steps

1. ✅ Run the 5-minute examples above
2. ✅ Read through `examples/complete-tutorial/README.md`
3. ✅ Try examples in `01_basics/` and `02_control_flow/`
4. ✅ Learn functions, recursion and lambdas
5. ✅ Master advanced types (arrays, maps, records, enum)
6. ✅ Explore annotations and the KylixBoot framework

## Get Help

- Tutorial: `examples/complete-tutorial/README.md` (中文版: `README_CN.md`)
- Beginner walkthrough: `docs/TUTORIAL_FOR_BEGINNERS_CN.md`
- Examples: `examples/complete-tutorial/*/example*.klx`
- Full sweep: `examples/complete-tutorial/test_all.sh`
- Index: `examples/complete-tutorial/INDEX.md`

Happy coding! 🚀
