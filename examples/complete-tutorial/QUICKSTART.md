# Kylix Tutorial - Quick Start Guide

## Installation Check

```bash
kylix version
# Should show: Kylix 3.0.0-alpha
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

kylix build hello.klx
go run hello.go
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

kylix build calc.klx
go run calc.go
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

kylix build loops.klx
go run loops.go
```

## Run All Tutorial Examples

```bash
cd /tmp/kylix_complete
./test_all.sh
```

## Learn by Category

### Beginners (Start Here)
```bash
cd /tmp/kylix_complete/01_basics
# Run each example:
kylix build example01_hello.klx && go run example01_hello.go
kylix build example02_variables.klx && go run example02_variables.go
# ... continue through example06
```

### Intermediate
```bash
cd /tmp/kylix_complete/02_control_flow
# All 5 control flow examples

cd /tmp/kylix_complete/03_functions
# Functions, recursion, multi-return
```

### Advanced
```bash
cd /tmp/kylix_complete/06_advanced_types
# Arrays, maps, records

cd /tmp/kylix_complete/10_exceptions
# Exception handling
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

### Arrays
```pascal
var numbers: array[0..4] of Integer;
numbers[0] := 10;
numbers[1] := 20;
```

### Maps
```pascal
var scores: map[String]Integer;
scores['Alice'] := 95;
WriteLn(scores['Alice']);
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

## What to Avoid (Bugs in v3.0.0-alpha)

❌ Don't use classes (field access broken)
❌ Don't use lambda expressions
❌ Don't use match expressions
❌ Don't use enum types
❌ Don't use `{ }` comments (use `//` instead)
❌ Don't use `Write()` (use `WriteLn()` instead)

## Next Steps

1. ✅ Run the 5-minute examples above
2. ✅ Read through `/tmp/kylix_complete/README.md`
3. ✅ Try examples in `/tmp/kylix_complete/01_basics/`
4. ✅ Explore control flow examples
5. ✅ Learn functions and recursion
6. ✅ Master advanced types (arrays, maps, records)

## Get Help

- Tutorial: `/tmp/kylix_complete/README.md`
- Examples: `/tmp/kylix_complete/*/example*.klx`
- Test: `/tmp/kylix_complete/test_all.sh`
- Summary: `/tmp/kylix_complete/SUMMARY.md`

Happy coding! 🚀
