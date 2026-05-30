# Kylix - Modern Pascal Language

Kylix is a modern reimagining of Pascal, designed to compile to Go. It combines the clarity and simplicity of Pascal with modern language features.

## Features

### Core Pascal Features
- Strong typing with type inference
- Procedures and functions
- Control structures (if, while, for, case, repeat)
- Records and arrays
- Exception handling

### Modern Additions
- **Type Inference**: `var x := 42;`
- **Lambda Expressions**: `var square = (x: Integer) -> x * x;`
- **Generics**: `TList<T>`
- **Async/Await**: `async function FetchData(): String;`
- **Pattern Matching**: `match value { 0 => 'zero', _ => 'other' }`
- **Classes & Interfaces**: Object-oriented programming support
- **Properties**: With getters and setters
- **ForEach Loops**: `for item in collection do`
- **String Interpolation**: `'Hello, ${name}!'`
- **Modern Exception Handling**: try/except/finally

## Installation

```bash
# Clone the repository
cd kylix

# Build the compiler
go build -o kylix

# Add to PATH (optional)
export PATH=$PATH:$(pwd)
```

## Usage

```bash
# Compile Kylix source to Go
./kylix program.klx

# Compile and run
./kylix -run program.klx

# Show tokens (debugging)
./kylix -tokens program.klx

# Show AST (debugging)
./kylix -ast program.klx

# Specify output file
./kylix -o output.go program.klx
```

## Examples

### Hello World
```pascal
program Hello;
begin
  WriteLn('Hello, Kylix World!');
end.
```

### Type Inference
```pascal
var count := 42;
var message := 'Inferred as string';
var ratio := 3.14;
```

### Functions
```pascal
function Add(a: Integer; b: Integer): Integer;
begin
  result := a + b;
end;

// Lambda
var square := (x: Integer) -> x * x;
```

### Classes
```pascal
class Animal
private
  var Name: String;
public
  constructor Create(name: String);
  begin
    Name := name;
  end;
  
  procedure Speak; virtual;
  begin
    WriteLn(Name, ' makes a sound');
  end;
end;

class Dog inherits Animal
public
  procedure Speak; override;
  begin
    WriteLn(Name, ' barks!');
  end;
end;
```

### Pattern Matching
```pascal
match value {
  0 => 'zero',
  1, 2, 3 => 'small',
  when value > 100 => 'large',
  _ => 'other'
};
```

### Async/Await
```pascal
async function FetchData(url: String): String;
begin
  // Async operations here
  result := 'Data from ' + url;
end;

var data := await FetchData('http://example.com');
```

### Exception Handling
```pascal
try
begin
  var result := Divide(10, 0);
end
except
begin
  WriteLn('An error occurred');
end
finally
begin
  WriteLn('Cleanup code');
end
end;
```

## Language Reference

### Types
- `Integer` - 64-bit integer (maps to Go's `int64`)
- `Real` - 64-bit float (maps to Go's `float64`)
- `Boolean` - Boolean value
- `String` - String value
- `Char` - Single character (maps to Go's `byte`)

### Operators
- Arithmetic: `+`, `-`, `*`, `/`, `div`, `mod`
- Comparison: `=`, `<>`, `<`, `>`, `<=`, `>=`
- Logical: `and`, `or`, `not`, `xor`
- Assignment: `:=`, `=`

### Control Structures
- `if/then/else`
- `while/do`
- `for/to/downto`
- `for/in` (foreach)
- `repeat/until`
- `case/of`
- `match` (pattern matching)
- `try/except/finally`

### Declarations
- `var` - Variable declaration
- `const` - Constant declaration
- `type` - Type declaration
- `function` - Function with return value
- `procedure` - Procedure (no return value)
- `class` - Class declaration
- `interface` - Interface declaration

## Project Structure

```
kylix/
├── token/       # Token definitions
├── lexer/       # Lexical analyzer
├── ast/         # Abstract Syntax Tree
├── parser/      # Parser
├── generator/   # Go code generator
├── examples/    # Example programs
└── main.go      # Compiler driver
```

## Roadmap

### Phase 1: Transpiler (Current)
- ✅ Lexer and parser
- ✅ AST generation
- ✅ Go code generation
- ✅ Basic language features
- ✅ Modern features (lambdas, async, pattern matching)

### Phase 2: IDE Tool
- Self-hosted compiler written in Kylix
- Syntax highlighting
- Code completion
- Error reporting
- Build system integration

### Phase 3: Framework
- Spring Boot-like framework
- Dependency injection
- Web server
- ORM
- Auto-configuration

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

MIT License
