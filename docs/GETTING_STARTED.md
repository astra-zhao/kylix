# Kylix Getting Started — Step by Step

> This tutorial walks through 5 tested Kylix examples, from Hello World to classes and objects, introducing the core features of the language one step at a time.
>
> **Requirements**: Kylix compiler installed (v0.6.9 or later)
>
> **How to run**: `kylix run example.klx` (auto-detects the Go/LLVM backend; falls back to LLVM when no Go toolchain is present)

> 🇨🇳 中文版: [GETTING_STARTED_CN.md](GETTING_STARTED_CN.md)

---

## Example 1: Hello World — your first program

The simplest Kylix program, printing "Hello, Kylix!".

```pascal
program HelloWorld;
begin
  WriteLn('Hello, Kylix!');
end.
```

**Run**:
```bash
kylix run example1_hello.klx
```

**Output**:
```
Hello, Kylix!
```

**Notes**:
- `program` declares the program name
- `begin...end.` wraps the main program block
- `WriteLn()` prints a line of text

---

## Example 2: Variables and Types — storing data

Demonstrates the basic data types: `String`, `Integer`, `Real`, `Boolean`.

```pascal
program Variables;
var
  name: String;
  age: Integer;
  score: Real;
  passed: Boolean;
begin
  name := 'Alice';
  age := 25;
  score := 89.5;
  passed := score >= 60.0;

  WriteLn('Name: ' + name);
  WriteLn('Age: ' + IntToStr(age));
  if passed then
    WriteLn('Status: Passed')
  else
    WriteLn('Status: Failed');
end.
```

**Output**:
```
Name: Alice
Age: 25
Status: Passed
```

**Notes**:
- `var` declares variables
- `:=` is the assignment operator
- `IntToStr()` converts an integer to a string (for concatenation)
- `if...then...else` conditional branching

---

## Example 3: Functions and Procedures — reusing code

Functions return values, procedures don't. Demonstrates recursion (Factorial) and parameter passing.

```pascal
program FunctionDemo;

function Add(a: Integer; b: Integer): Integer;
begin
  result := a + b;
end;

function Factorial(n: Integer): Integer;
begin
  if n <= 1 then
    result := 1
  else
    result := n * Factorial(n - 1);
end;

procedure Greet(name: String);
begin
  WriteLn('Hello, ', name, '!');
end;

var
  x, y, sum: Integer;
  fact: Integer;
begin
  x := 10;
  y := 20;
  sum := Add(x, y);
  WriteLn(IntToStr(x), ' + ', IntToStr(y), ' = ', IntToStr(sum));

  fact := Factorial(5);
  WriteLn('5! = ', IntToStr(fact));

  Greet('Kylix User');
end.
```

**Output**:
```
10 + 20 = 30
5! = 120
Hello, Kylix User!
```

**Notes**:
- `function` returns a value (assigned via `result`)
- `procedure` returns nothing
- Functions can call themselves recursively (`Factorial`)

---

## Example 4: Loops — repeating work

Demonstrates `while` loops and nested loops.

```pascal
program LoopsDemo;
var
  i, sum: Integer;
  j: Integer;
begin
  // While loop - sum 1 to 5
  WriteLn('=== For Loop ===');
  sum := 0;
  i := 1;
  while i <= 5 do
  begin
    sum := sum + i;
    i := i + 1;
  end;
  WriteLn('Sum 1-5: ' + IntToStr(sum));

  // Nested loops - multiplication table
  WriteLn('=== Multiplication Table (3x3) ===');
  i := 1;
  while i <= 3 do
  begin
    j := 1;
    while j <= 3 do
    begin
      WriteLn(IntToStr(i) + ' x ' + IntToStr(j) + ' = ' + IntToStr(i * j));
      j := j + 1;
    end;
    i := i + 1;
  end;
end.
```

**Output**:
```
=== For Loop ===
Sum 1-5: 15
=== Multiplication Table (3x3) ===
1 x 1 = 1
1 x 2 = 2
1 x 3 = 3
2 x 1 = 2
2 x 2 = 4
2 x 3 = 6
3 x 1 = 3
3 x 2 = 6
3 x 3 = 9
```

**Notes**:
- `while...do` loops (runs while the condition is true)
- `begin...end` wraps multiple statements
- Nested loops: the outer loop controls rows, the inner loop columns

---

## Example 5: Classes and Objects — OOP

Demonstrates class definitions, object creation, and field access.

```pascal
program ClassDemo;

type
  TPerson = class
  public
    Name: String;
    Age: Integer;
  end;

var
  person: TPerson;
begin
  person := TPerson.Create;
  person.Name := 'Bob';
  person.Age := 30;
  WriteLn('Person: ' + person.Name + ', Age: ' + IntToStr(person.Age));
end.
```

**Output**:
```
Person: Bob, Age: 30
```

**Notes**:
- `type...class` defines a class
- `public` declares public fields
- `TPerson.Create` creates an object instance
- `.` accesses fields

---

## Where to go next

**Intermediate topics**:
- Arrays and records: `array of Integer`, `record...end`
- Exception handling: `try...except...finally`
- Generics: `TList<T>`
- Interfaces: `interface...end`

**Advanced topics**:
- Web server: `uses web; app := createServer(8080);`
- JSON processing: `uses jsonutil; obj := ParseJSON(str);`
- File I/O: `uses sysutil; content := ReadFile('data.txt');`
- WASI builds: `kylix build --wasi main.klx`
- LLVM native backend: `kylix build --backend=llvm main.klx` (no Go dependency; run `kylix doctor` to preflight the environment)
- The self-hosting compiler: the compiler itself is written in Kylix (`src/*.klx`), and the bootstrap has closed the no-Go loop (v0.6.9)

**Full docs**: [README.md](../README.md) | [Website](https://kylix.top) | [CHANGELOG](../CHANGELOG.md)

---

## FAQ

**Q: How do I check the version?**
```bash
kylix version
```

**Q: How do I format code?**
```bash
kylix fmt myfile.klx
```

**Q: How do I run tests?**
```bash
kylix test myfile_test.klx
```

**Q: The build fails — what now?**
- Check syntax: every statement needs a semicolon `;`
- Function return values: use `result :=`, not `return`
- Declare variables first: `var x: Integer;` before `begin`
- String concatenation: use `+` to join strings

---

**Version**: applies to Kylix v0.6.9
**Last updated**: 2026-09-04
