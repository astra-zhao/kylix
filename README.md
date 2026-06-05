# Kylix - Modern Pascal Language

[![Official Site](https://img.shields.io/badge/official-kylix.top-4f6ef7.svg)](https://kylix.top)
[![中文文档](https://img.shields.io/badge/lang-中文-red.svg)](SUMMARY.md)
[![Version](https://img.shields.io/badge/version-1.0.3-blue.svg)](CHANGELOG.md)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

Kylix is a modern reimagining of Pascal, designed to compile to Go. It combines the clarity and simplicity of Pascal with modern language features, and ships with a full IDE toolchain and editor integrations.

> 🌐 **Official Website**: [https://kylix.top](https://kylix.top) — interactive docs, live examples, and the full feature showcase.

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
- **Generics**: Declare type parameters: `TList<T>`, `function Foo<T>(x: T): T`
- **Generic Type References**: `TList<Integer>`, `TPair<String, Integer>`
- **Async/Await**: `async function FetchData(): String;`
- **Pattern Matching**: `match value { 0 => 'zero', _ => 'other' }`
- **Classes & Interfaces**: Object-oriented programming support
- **Properties**: With getters and setters
- **ForEach Loops**: `for item in collection do`
- **String Interpolation**: `'Hello, ${name}!'`
- **Modern Exception Handling**: try/except/finally, `on E: Type do` clauses

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

### Exception Handling with ON Clause
```pascal
try
  raise Exception.Create('test');
except
  on E: Exception do
    WriteLn('Caught: ' + E.Message);
  else
    WriteLn('Unknown exception');
end;
```

### Generic Classes and Functions
```pascal
type
  TPair<T1, T2> = class
    First: T1;
    Second: T2;
  end;

function CreatePair<T>(x: T; y: T): TPair<T, T>;
begin
  Result := TPair<T, T>.Create;
end;

var
  pair: TPair<Integer, String>;
begin
  pair := CreatePair<Integer>(42, 'hello');
end.
```

### Anonymous Procedures & Functions
```pascal
// Anonymous procedure
var greet := procedure()
begin
  WriteLn('Hello!');
end;
greet();

// Anonymous function with parameters
var add := function(a: Integer; b: Integer): Integer
begin
  result := a + b;
end;
WriteLn(add(10, 20));  // 30
```

### Web Server
```pascal
program WebApp;
uses web;
var
  app: TServer;
begin
  app := web.createServer(8080);

  // Simple GET route
  app.get('/', procedure(req: TRequest; res: TResponse)
  begin
    res.send('Hello, Kylix Web!');
  end);

  // JSON API with path parameters
  app.get('/api/users/:id', procedure(req: TRequest; res: TResponse)
  var
    userId: String;
  begin
    userId := req.param('id');
    res.json(record id := userId; name := 'User ' + userId; end);
  end);

  // POST route with JSON body
  app.post('/api/users', procedure(req: TRequest; res: TResponse)
  var
    body: record name: String; email: String; end;
  begin
    req.json(body);
    res.status(201).json(body);
  end);

  // Middleware
  app.use(web.loggerMiddleware());

  // Static files
  app.static('/public', './static');

  app.listen();
end.
```

## Standard Library

### Web Framework (`web`)
HTTP server with routing, middleware, and request/response handling.

```pascal
uses web;

app := web.createServer(8080);
app.get('/api/users', procedure(req: TRequest; res: TResponse)
begin
  res.json(users);
end);
app.listen();
```

### Dependency Injection (`container`)
IoC container for managing dependencies with singleton, transient, and scoped lifetimes.

```pascal
uses container;

di := NewContainer;
di.RegisterSingleton('UserService', function: TUserService
begin
  result := TUserService.Create;
end);

service := di.Resolve('UserService').(TUserService);
```

### Configuration (`config`)
Load configuration from environment variables with type-safe accessors.

```pascal
uses config;

cfg := NewConfig;
cfg.SetPrefix('APP');
cfg.LoadFromEnv;

port := cfg.GetIntDefault('PORT', 8080);
debug := cfg.GetBoolDefault('DEBUG', false);
```

### Middleware (`middleware`)
Pre-built middleware for common web application needs.

```pascal
uses middleware;

app.use(NewRequestIDMiddleware.Handle);
app.use(NewLoggingMiddleware.Handle);
app.use(NewCORSMiddleware.Handle);
app.use(NewAuthMiddleware(ValidateToken).Handle);
app.use(NewRateLimitMiddleware(100, 60).Handle);
```

### Validation (`validation`)
Request validation with fluent API and common validators.

```pascal
uses validation;

validator := NewRequestValidator(req);
validator.Required(['name', 'email']);
validator.Email('email');
validator.MinLength('password', 8);

if not validator.IsValid then
  res.status(400).json(validator.Errors);
```

### ORM (`orm`)
Database ORM supporting MySQL, PostgreSQL, and SQLite with query builder and migrations.

```pascal
uses orm;

// Connect to database
dbConfig := TConnectionConfig{
  Type: DBSQLite,
  Database: './app.db'
};
db := NewDatabase(dbConfig);
orm := NewORM(db);

// Insert
data := map[string]interface{}{
  'name': 'John',
  'email': 'john@example.com'
};
id := orm.Insert('users', data);

// Query with builder
qb := orm.QueryBuilder('users');
qb.Where('age', '>', 18);
qb.OrderBy('name', 'ASC');
qb.Limit(10);
users := orm.Execute(qb);

// Find by ID
user := orm.Find('users', 1);

// Update
orm.Update('users', 
  map[string]interface{}{'id': 1},  // condition
  map[string]interface{}{'name': 'Jane'}  // data
);

// Delete
orm.Delete('users', map[string]interface{}{'id': 1});
```

### Template Engine (`template`)
HTML template rendering with layouts, partials, and custom functions.

```pascal
uses template;

engine := NewTemplateEngine;
engine.SetTemplateDir('./templates');

// Register layout
engine.RegisterLayout('main', 
  '<html><body>{{.Content}}</body></html>');

// Register partial
engine.RegisterPartial('header', '<h1>My App</h1>');

// Render with layout
view := NewView(engine);
view.With('Title', 'Home');
view.With('Message', 'Welcome!');
view.WithLayout('main');
html := view.Render('home.html');
res.HTML(html);
```

### Auto-Configuration (`autoconfig`)
Automatic configuration loading from multiple sources with environment detection.

```pascal
uses autoconfig;

config := NewAutoConfig('myapp');
config.DetectEnvironment;  // Detect from APP_ENV
config.SetConfigDir('./config');
config.AddDefaultSources;  // config.json, config.{env}.json, env vars
config.Load;

// Access configuration
port := config.GetInt('server.port');
dbHost := config.GetString('database.host');
debug := config.GetBool('app.debug');

// Environment checks
if config.IsProduction then
  // Production-specific logic
```

### File I/O (`sysutil`)
File and directory operations with a Pascal-friendly API.

```pascal
uses sysutil;

// Read/write files
content := sysutil.ReadFile('data.txt');
sysutil.WriteFile('output.txt', 'Hello, World!');
sysutil.AppendFile('log.txt', 'New line');

// File operations
if sysutil.FileExists('config.json') then
  WriteLn('Config found');

sysutil.CreateDir('new_folder');
sysutil.CopyFile('src.txt', 'dst.txt');
sysutil.DeleteFile('temp.txt');

// List files
files := sysutil.ListDir('./');
matches := sysutil.ListFiles('*.klx');

// Path utilities
fullPath := sysutil.PathJoin('dir', 'sub', 'file.txt');
dir := sysutil.PathDir('/home/user/doc.txt');
ext := sysutil.PathExt('photo.jpg');

// Line-based I/O
lines := sysutil.ReadLines('data.csv');
sysutil.WriteLines('output.csv', lines);
```

### JSON (`jsonutil`)
JSON encoding, decoding, and manipulation.

```pascal
uses jsonutil;

// Encode to JSON
jsonStr := jsonutil.JsonEncode(data);
pretty := jsonutil.JsonEncodePretty(data);

// Decode from JSON
obj := jsonutil.JsonDecodeMap('{"name": "Kylix", "version": 1}');
name := jsonutil.JsonGetString(obj, 'name');
ver := jsonutil.JsonGetInt(obj, 'version');

// Type-safe accessors
flag := jsonutil.JsonGetBool(obj, 'active');
pi := jsonutil.JsonGetFloat(obj, 'pi');
child := jsonutil.JsonGetMap(obj, 'nested');
items := jsonutil.JsonGetArray(obj, 'list');

// Validation
if jsonutil.JsonIsValid(input) then
  WriteLn('Valid JSON');

// File I/O
data := jsonutil.JsonReadFile('config.json');
jsonutil.JsonWriteFile('output.json', data);
```

### DateTime (`datetime`)
Date and time manipulation with arithmetic and formatting.

```pascal
uses datetime;

// Current time
now := datetime.Now();
WriteLn(now.FormatDateTime());  // 2024-06-15 10:30:00
WriteLn(now.FormatDate());     // 2024-06-15

// Create dates
birthday := datetime.MakeDate(1990, 5, 15);
meeting := datetime.MakeTime(2024, 12, 25, 14, 30, 0);

// Date arithmetic
nextWeek := now.AddDays(7);
nextMonth := now.AddMonths(1);
tomorrow := now.AddDays(1);

// Comparisons
days := now.DiffDays(birthday);
if now.After(deadline) then
  WriteLn('Overdue!');

// Utilities
if now.IsWeekend() then
  WriteLn('Weekend!');
if now.IsLeapYear() then
  WriteLn('Leap year');
WriteLn('Day: ' + now.DayName());
WriteLn('Month: ' + now.MonthName());

// Parsing
dt := datetime.ParseDate('2024-06-15');
dt2 := datetime.ParseDateTime('2024-06-15 10:30:00');

// Timestamps
ts := datetime.GetTimestamp();    // Unix seconds
tsMs := datetime.GetTimestampMs(); // Unix milliseconds
```

### Regular Expressions (`regex`)
Pattern matching, searching, and text manipulation.

```pascal
uses regex;

// Quick pattern checks
if regex.IsEmail('user@example.com') then
  WriteLn('Valid email');

if regex.IsNumeric('12345') then
  WriteLn('All digits');

if regex.IsURL('https://example.com') then
  WriteLn('Valid URL');

// Find and replace
match := regex.RegexFind('[0-9]+', 'Order #12345');
// match = '12345'

result := regex.RegexReplace('\s+', 'a  b  c', ' ');
// result = 'a b c'

// Compiled regex (reusable)
re := regex.RegexMustCompile('(\w+)@(\w+)');
if re.Match('user@host') then
  groups := re.Groups('user@host');
  // groups[1] = 'user', groups[2] = 'host'

// Split
parts := regex.RegexSplit(',', 'a,b,c,d');
// parts = ['a', 'b', 'c', 'd']

// Extract all numbers
nums := regex.ExtractNumbers('Room 42, Floor 3, Building 7');
// nums = ['42', '3', '7']
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
├── stdlib/             # Standard library
│   ├── web.go          # Web framework
│   ├── container.go    # Dependency injection
│   ├── config.go       # Configuration management
│   ├── middleware.go   # Middleware (CORS, auth, rate limit)
│   ├── validation.go   # Request validation
│   ├── orm.go          # ORM (MySQL, PostgreSQL, SQLite)
│   ├── template.go     # Template engine
│   ├── autoconfig.go   # Auto-configuration
│   ├── sysutil.go      # File I/O and system utilities
│   ├── jsonutil.go     # JSON encoding/decoding
│   ├── datetime.go     # Date and time operations
│   └── regex.go        # Regular expressions
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
    ├── KYLIX_TOOLS_EXPLAINED.md
    ├── WEB_FRAMEWORK.md
    ├── ORM_GUIDE.md
    └── TEMPLATE_GUIDE.md
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
- [Web Framework Guide](docs/WEB_FRAMEWORK.md) - Web server and REST API development
- [ORM Guide](docs/ORM_GUIDE.md) - Database ORM and query builder
- [Template Engine Guide](docs/TEMPLATE_GUIDE.md) - HTML template rendering
- [Phase 2 Summary](docs/PHASE2_SUMMARY.md) - IDE toolchain completion report

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

### Phase 3: Framework ✅
- ✅ Web server (based on Go net/http)
- ✅ Routing system (GET, POST, PUT, DELETE)
- ✅ Path parameters (`/users/:id` syntax)
- ✅ Middleware support (logger middleware)
- ✅ JSON request/response handling
- ✅ Static file serving
- ✅ Anonymous procedures & functions
- ✅ Enhanced VS Code extension (syntax highlighting, snippets, completions)
- ✅ Web framework documentation
- ✅ Dependency injection container
- ✅ Configuration system
- ✅ Middleware suite (CORS, Auth, Rate Limit, Request ID, Logging)
- ✅ Request validation
- ✅ ORM (MySQL, PostgreSQL, SQLite support)
- ✅ Template engine (layouts, partials, custom functions)
- ✅ Auto-configuration (multi-source config loading)

### Phase 4: Language Enhancements ✅
- ✅ Generic type parameter declarations (classes and functions)
- ✅ Exception handling with ON clause (`on E: ExceptionType do`)
- ✅ Constructor/destructor/inherited keywords
- ✅ Lambda expression parameter parsing
- ✅ Async/await code generation improvements

### Phase 5: Standard Library & Tooling ✅
- ✅ File I/O (`sysutil`) — read, write, copy, directory operations, path utilities
- ✅ JSON (`jsonutil`) — encode, decode, type-safe accessors, file I/O
- ✅ DateTime (`datetime`) — date arithmetic, formatting, parsing, comparisons
- ✅ Regular Expressions (`regex`) — match, find, replace, split, pattern helpers
- ✅ REPL improvements — readline with history (↑/↓), lexer-based detection, stderr separation
- ✅ Formatter fixes — class visibility modifiers, properties, const type annotations
- ✅ Generator stdlib wiring — sysutil, jsonutil, datetime, regex modules

## Changelog

### v1.0.1 (2026-06-02)

**Bug fix release — 4 critical fixes + 2 high-priority fixes**

- **P0**: `inherits` keyword now works correctly (was silently ignored)
- **P0**: Anonymous `procedure()` and `function()` now parseable as expressions
- **P0**: Match wildcard `_` now generates correct Go `default:` branch
- **P0**: Removed `{}` comment syntax (conflicted with match block braces)
- **P1**: Constructor calls `Dog.Create(args)` now generate `&Dog{args}`
- **P1**: Match branches now properly trigger Go import generation

See [CHANGELOG.md](CHANGELOG.md) for full release history and known issues.

### v1.0.0 (2026-06-01)

**🎉 First stable release — all 5 phases complete!**

- **Standard Library**: Added `sysutil` (File I/O), `jsonutil` (JSON), `datetime` (DateTime), `regex` (Regular Expressions)
- **REPL**: Readline support with persistent history, lexer-based multiline detection, stderr separation
- **Formatter**: Class visibility modifiers, properties output, const type annotations
- **Generics**: Type parameter declarations for classes and functions (`TList<T>`, `function Foo<T>`)
- **Exception Handling**: ON clause support (`on E: ExceptionType do`)
- **Web Framework**: DI container, config system, middleware suite, validation, ORM, template engine, auto-config
- **IDE Tools**: LSP server, VS Code extension, project management, code formatting

See [CHANGELOG.md](CHANGELOG.md) for full release history.

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

MIT License
