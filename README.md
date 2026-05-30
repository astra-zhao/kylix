# Kylix - Modern Pascal Language

[![中文文档](https://img.shields.io/badge/lang-中文-red.svg)](SUMMARY.md)

Kylix is a modern reimagining of Pascal, designed to compile to Go. It combines the clarity and simplicity of Pascal with modern language features, and ships with a full IDE toolchain and editor integrations.

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
git clone https://github.com/astra-zhao/kylix.git
cd kylix

# Build the compiler
go build -o kylix cmd/kylix/main.go

# Add to PATH (optional)
export PATH=$PATH:$(pwd)
```

## Quick Start

```bash
# Create a new project
./kylix new myapp
cd myapp

# Compile and run
./kylix run

# Check syntax
./kylix check

# Format code
./kylix fmt
```

## CLI Commands

```bash
kylix new <name>       # Create a new project
kylix build            # Compile project or file
kylix run              # Compile and run
kylix check            # Syntax check (no code generation)
kylix fmt              # Format source files
kylix repl             # Interactive REPL
kylix lsp              # Start LSP server (for editors)
kylix version          # Show version
kylix help             # Show help
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
├── cmd/kylix/          # CLI entry point
├── pkg/
│   ├── compiler/       # Compilation API
│   ├── project/        # Project management (kylix.toml)
│   ├── lsp/            # Language Server Protocol server
│   └── repl/           # Interactive REPL
├── token/              # Token definitions
├── lexer/              # Lexical analyzer
├── ast/                # Abstract Syntax Tree
├── parser/             # Parser (Pratt parsing)
├── generator/          # Go code generator
├── examples/           # Example programs
├── vscode-ext/         # VS Code extension
└── docs/               # Documentation
    ├── KYLIX_IDE_USER_MANUAL.md
    ├── KYLIX_DEV_GUIDE.md
    └── KYLIX_TOOLS_EXPLAINED.md
```

## Editor Integration

### VS Code
The `vscode-ext/` directory contains a full VS Code extension with:
- Syntax highlighting
- Language configuration (brackets, comments, folding)
- LSP client integration

```bash
cd vscode-ext
npm install
# Press F5 in VS Code to launch extension
```

### Other Editors
Kylix LSP supports any editor with LSP client:
```json
{
  "command": ["kylix", "lsp"],
  "filetypes": ["kylix"]
}
```

## Documentation

- [IDE User Manual](docs/KYLIX_IDE_USER_MANUAL.md) - Complete CLI and editor guide
- [Developer Guide](docs/KYLIX_DEV_GUIDE.md) - Architecture, internals, and contributing
- [Tools Explained](docs/KYLIX_TOOLS_EXPLAINED.md) - Beginner-friendly tool descriptions

## Roadmap

### Phase 1: Transpiler ✅
- ✅ Lexer and parser
- ✅ AST generation
- ✅ Go code generation
- ✅ Basic language features
- ✅ Modern features (lambdas, async, pattern matching)

### Phase 2: IDE Tool ✅
- ✅ CLI toolchain (new, build, run, check, fmt, repl, lsp)
- ✅ Project management (kylix.toml)
- ✅ LSP server with completion and hover
- ✅ VS Code extension with syntax highlighting
- ✅ Interactive REPL
- ✅ Comprehensive documentation

### Phase 3: Framework (In Progress)
- Spring Boot-like framework
- Dependency injection
- Web server
- ORM
- Auto-configuration

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

MIT License
