# Kylix - Modern Pascal Language

[![Official Site](https://img.shields.io/badge/official-kylix.top-4f6ef7.svg)](https://kylix.top)
[![中文文档](https://img.shields.io/badge/lang-中文-red.svg)](SUMMARY.md)
[![Version](https://img.shields.io/badge/version-3.1.0-blue.svg)](CHANGELOG.md)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Self-Hosting](https://img.shields.io/badge/self--hosting-100%25-brightgreen.svg)](ROADMAP.md)

Kylix is a modern reimagining of Pascal, designed to compile to Go. It combines the clarity and simplicity of Pascal with modern language features, and ships with a full IDE toolchain and editor integrations.

> 🌐 **Official Website**: [https://kylix.top](https://kylix.top) — interactive docs, live examples, and the full feature showcase.
>
> 🎉 **v3.1.0 Release**: KylixBoot framework (Spring Boot–style annotations + DI + router), `[Attribute]` syntax, LLVM arrays + optimization passes, and 5 critical compiler fixes (KLX-C01..C05). See [CHANGELOG.md](CHANGELOG.md) for details.

## Features

### Core Pascal Features
- Strong typing with type inference
- Procedures and functions
- Control structures (if, while, for, case, repeat)
- Records and arrays
- Exception handling

### Modern Additions
- **Type Inference**: `var x := 42;` — infers Integer from literal
- **Generic Constraints**: `TBox<T: IComparable>` — validate type parameters
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

### Complete Toolchain (v2.0.0)
- **Testing**: `kylix test` — discover and run `Test*` procedures in `*_test.klx` files
- **Benchmarking**: `kylix bench` — measure performance of `Bench*` procedures
- **Documentation**: `kylix doc` — generate Markdown from `//` doc comments
- **Type Checking**: Enhanced with error codes (KLX001–499), recovery, and "did you mean?" suggestions
- **LSP Server**: Full IDE support with completion, hover, diagnostics, and signature help
- **Package Manager**: `kylix add`, `kylix remove`, `kylix publish` for dependency management
- **WASI**: `kylix build --wasi` — compile to WebAssembly System Interface (v3.0.0-alpha)
- **LLVM Backend**: `kylix build --backend=llvm` — native code without Go toolchain (v3.0.0-alpha)
- **KylixBoot Framework**: Spring Boot–style annotation-driven web apps (v3.1.0)
- **Annotation Syntax**: `[Controller]`, `[Get]`, `[Inject]`, `[Entity]` (v3.1.0)

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
kylix build --wasm     # Compile to WebAssembly
kylix build --wasi     # Compile to WASI (wasip1/wasm, Go 1.21+)
kylix build --backend=llvm  # Compile via LLVM native backend
kylix build --backend=llvm --llvm-opt=2  # Optimized LLVM build (-O2)
kylix run              # Compile and run
kylix check            # Syntax check (no code generation)
kylix fmt              # Format source files
kylix repl             # Interactive REPL
kylix lsp              # Start LSP server (for editors)
kylix publish          # Publish package to registry
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

### HTTP Client (`httpclient`) — v3.0.0-alpha

One-shot HTTP helpers and a reusable client with header support.

```pascal
uses httpclient;

// One-shot GET
body := HttpGet('https://api.example.com/data');

// One-shot POST with JSON response
resp := HttpGetJSON('https://api.example.com/items');

// Reusable client with custom headers
client := NewHttpClient('https://api.example.com');
client.SetHeader('Authorization', 'Bearer ' + token);
body := client.Get('/users');
WriteLn(IntToStr(client.StatusCode()));
```

### WASI (`wasi`) — v3.0.0-alpha

Portable system interface for WASI runtimes (Wasmtime, Node.js, Cloudflare Workers).

```pascal
uses wasi;

begin
  WriteLn('Hello from WASI!');
  WriteLn('Arg count: ' + IntToStr(ArgCount()));
  WriteLn('First arg: ' + Arg(0));

  var val := GetEnvOrDefault('PORT', '8080');
  WriteLn('PORT=' + val);

  var ms := ElapsedMs();
  WriteLn('Elapsed: ' + IntToStr(ms) + ' ms');
end.
```

### KylixBoot Framework (`boot`) — v3.1.0+

Spring Boot–style web framework: router with path params, DI container, graceful shutdown, env-var configuration, and built-in middleware.

```pascal
program HelloBoot;
uses boot;

begin
  // Register routes via global shortcuts
  boot.GET('/', procedure(req: TRequest; res: TResponse)
  begin
    res.Send('Hello, KylixBoot!');
  end);

  boot.GET('/users/:id', procedure(req: TRequest; res: TResponse)
  begin
    res.JSON(record id := req.Param('id'); end);
  end);

  // Built-in middleware
  boot.Use(boot.Logger());
  boot.Use(boot.Recover());
  boot.Use(boot.CORS());

  // Graceful shutdown built in
  boot.Listen(':8080');
end.
```

Container support:

```pascal
// DI container
container := boot.NewContainer();
container.RegisterSingleton('UserService', TUserService);
container.RegisterTransient('Request', TRequestScope);

// Reflection-based injection
container.Inject(controller);
```

23 unit tests in `pkg/boot/`; declarations in `stdlib/klx/boot.klx`.

### Annotation Syntax (`[Attribute]`) — v3.1.0+

Annotations attach metadata to classes, types, functions, and fields — the foundation of declarative APIs (route registration, ORM mapping, DI, validation).

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

    [Post('/'), Authenticated]
    function CreateUser(req: TRequest): TResponse;
    begin
      var user := req.Body<TUser>();
      UserRepo.Save(user);
      result := req.Created(user);
    end;
  end;

[Entity('users')]
type
  TUser = class
    [Column('id'), PrimaryKey]
    Id: Integer;
    [Required, Email]
    Email: String;
  end;
```

Annotations parse at the AST level (`ast.Attribute`), attach to `ClassDecl`/`TypeDecl`/`FunctionDecl`/`VarDecl`, and are designed for v3.2's auto-route registration + ORM code generation.

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
│   ├── pkgmgr/         # Package manager (add/install/remove/publish)
│   ├── llvmgen/        # LLVM native backend (v3.0.0-alpha)
│   │   ├── codegen.go      # Generator core, SSA, string pool
│   │   ├── expr.go         # Expression codegen
│   │   ├── stmt.go         # Statement codegen
│   │   ├── class.go        # Class/vtable codegen
│   │   └── compile.go      # Full pipeline: AST → binary
│   ├── wasi/           # WASI system interface (v3.0.0-alpha)
│   │   ├── wasi.go         # Package doc
│   │   ├── wasi_stub.go    # Non-WASI stub (local testing)
│   │   └── wasi_wasip1.go  # WASI native implementation
│   ├── formatter/      # Source formatter
│   │   ├── formatter.go       # core + declaration formatting
│   │   ├── formatter_stmt.go  # statement formatting
│   │   └── formatter_expr.go  # expression + type formatting
│   ├── lsp/            # Language Server Protocol
│   │   ├── server.go              # JSON-RPC transport + dispatch
│   │   ├── handler_completion.go  # completion + hover
│   │   └── handler_navigation.go  # definition, refs, rename, formatting
│   └── repl/           # Interactive REPL
├── registry/           # Package registry server (v3.0.0-alpha)
│   ├── internal/
│   │   ├── api/        # REST API handlers
│   │   ├── auth/       # Bearer token auth
│   │   ├── db/         # SQLite/PostgreSQL store
│   │   └── models/     # Data models
│   └── web/templates/  # htmx + Tailwind CSS frontend
├── stdlib/             # Standard library
│   ├── web.go          # Web framework
│   ├── orm.go          # Database connection + transaction
│   ├── orm_query.go    # QueryBuilder fluent API
│   ├── orm_migrate.go  # ORM CRUD + MigrationManager
│   ├── http_client.go  # THttpClient + one-shot helpers (v3.0.0-alpha)
│   ├── klx/            # LSP declaration files
│   │   ├── sysutil.klx
│   │   ├── datetime.klx
│   │   ├── regex.klx
│   │   ├── jsonutil.klx
│   │   ├── httpclient.klx
│   │   └── wasi.klx
│   └── src/            # Pure Kylix stdlib implementations
│       ├── strutil.klx, mathutil.klx, arrayutil.klx
│       ├── collections.klx, stringbuilder.klx, resulttype.klx, iter.klx
│       ├── jsonutil.klx, regex.klx, datetime.klx  # Phase 4 (v3.0.0-alpha)
│       ├── httpclient.klx, wasi.klx               # Phase 5 (v3.0.0-alpha)
│       └── ...
├── token/              # Token definitions
├── lexer/              # Lexical analyzer
├── ast/                # AST node definitions
├── parser/             # Pratt parser (split by concern)
│   ├── parser.go       # core: Parser struct, ParseProgram
│   ├── parser_decl.go  # var, const, type, function, class declarations
│   ├── parser_stmt.go  # if, for, while, try, match, raise, ...
│   └── parser_expr.go  # expressions, literals, types
├── generator/          # Go code generator (split by concern)
│   ├── generator.go        # core: Generate/GenerateMulti, pre-scan
│   ├── generator_types.go  # class, interface, variant, enum, function codegen
│   ├── generator_stmt.go   # statement codegen
│   └── generator_expr.go   # expression codegen
├── src/                # Self-hosting compiler source (.klx)
│   ├── token.klx
│   ├── ast.klx
│   ├── lexer.klx
│   ├── parser.klx      # TParser — all methods (2423 lines, not yet split)
│   ├── generator.klx   # TGenerator — all methods (1702 lines, class body)
│   ├── error.klx
│   └── main.klx
├── examples/           # Example programs
│   ├── wasi-hello/     # WASI Hello World (Wasmtime/Node.js)
│   └── cloudflare-worker/  # Cloudflare Workers HTTP handler
├── vscode-ext/         # VS Code extension
└── docs/               # Documentation
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

### Phase 6-7: Bug Fixes & Language Capabilities ✅
- ✅ String interpolation (Lexer→Parser→Generator)
- ✅ Exception types with ON clause
- ✅ Multi-value return (`function Div(a,b: Integer): (Integer, Integer)`)
- ✅ Properties code generation (getter/setter)
- ✅ Map type (`map[K]V`), Variant type, Dynamic arrays
- ✅ Enum types (`type TColor = (Red, Green, Blue)`)
- ✅ Multi-file module system (`unit X;`, `uses X;`)

### Phase 8: Self-Hosting Compiler ✅
- ✅ 7 Kylix source files written (token, ast, lexer, parser, error, generator, main)
- ✅ Class code generation (struct + interface{} polymorphism)
- ✅ Soft keywords (25+ keywords usable as identifiers)
- ✅ is/as type dispatch for expressions and statements
- ✅ Local variable declarations, constructor, builtins

### Phase 9: Bootstrap Verification ✅ Complete
- ✅ Multi-file bootstrap compilation
- ✅ Self-hosted Go output compiles and runs with zero errors
- ✅ Diff verification: Go reference vs Kylix self-hosted — semantically equivalent
- ✅ 15/15 examples pass on both compilers

### v2.1.0–v2.6.0: Engineering Quality & stdlib ✅
- ✅ Enhanced type system (multi-param generics, interface mapping, type inference)
- ✅ GitHub Actions CI/CD, incremental compilation (55×), dead code elimination
- ✅ LSP incremental sync, REPL Tab completion, i18n error messages
- ✅ stdlib Phase 1–3: strutil, mathutil, arrayutil, collections, stringbuilder, resulttype, iter
- ✅ Delve debugger integration, WebAssembly backend (`--wasm`)
- ✅ Parallel compilation, LSP large-file performance benchmarks

### v3.0.0-alpha: Architecture Breakthrough ✅
- ✅ LLVM native backend — Milestone 1 (scalar types, control flow, functions, classes)
- ✅ WASI support (`--wasi`, `--tinygo`, `pkg/wasi/`, `stdlib/src/wasi.klx`)
- ✅ Package registry server (`registry/`, REST API, htmx frontend, `kylix publish`)
- ✅ stdlib Phase 4: pure Kylix jsonutil (nested JSON), regex, datetime (DateAdd/DateSub)
- ✅ `external` function declaration parsing fixed
- ✅ HTTP client stdlib (`httpclient`)

### v3.1.0: KylixBoot + Compiler Fixes + LLVM Arrays ✅
- ✅ KylixBoot framework (`pkg/boot/`, ~700 lines, 23 tests) — router, DI, middleware, graceful shutdown
- ✅ Annotation syntax `[Name]` / `[Name(args)]` on classes, types, functions, fields
- ✅ KLX-C01 fix: `var p: TClass` now emits `*TClass` instead of `interface{}`
- ✅ KLX-C02 fix: single-quoted strings with `${...}` produce STRING_INTERPOLATION
- ✅ KLX-C03 fix: lambda/anonymous-function return types preserved
- ✅ KLX-C04 fix: match statement codegen now produces valid Go
- ✅ KLX-C05 fix: `uses sysutil/jsonutil/...` in program files now injects stdlib symbols (40+ functions)
- ✅ LLVM Milestone 2 Phase 1: static + dynamic arrays, `--llvm-opt=N`
- ✅ Tutorial expanded with `example40_declarative_oop.klx` and `example41_attributes.klx` (32/34 examples pass)

### v3.2.0: Auto-Wiring + ORM + LLVM M2 Phase 2 🔲
- 🔲 Auto-route registration from `[Controller]`/`[Get]` annotations (DI integration)
- 🔲 ORM annotations: `[Entity]` / `[Repository]` / `[Query]`
- 🔲 LLVM Milestone 2 Phase 2 — interface fat pointer
- 🔲 LLVM Milestone 2 Phase 3 — generic monomorphization
- 🔲 Validation annotations `[Required]` / `[Min]` / `[Email]`
- 🔲 Security annotations `[Authenticated]` / `[Role]`
- 🔲 Registry deployment to kylix.top/packages
- 🔲 stdlib Phase 6: net / crypto / encoding

## Cross-Platform Compilation

Kylix compiles to Go source, then uses Go's built-in cross-compilation to produce native binaries for any platform — no virtual machine, no runtime installation required on the target machine.

### How it works

```
your .klx file
    ↓  kylix build  (transpile Pascal → Go)
generated .go file
    ↓  go build     (compile Go → native binary)
native executable
```

### Build for different platforms

```bash
# Linux (Intel/AMD)
kylix build --target=linux/amd64 main.klx

# Windows (Intel/AMD)
kylix build --target=windows/amd64 main.klx

# macOS Apple Silicon (M1/M2/M3)
kylix build --target=darwin/arm64 main.klx

# macOS Intel
kylix build --target=darwin/amd64 main.klx

# Linux ARM (Raspberry Pi, cloud ARM)
kylix build --target=linux/arm64 main.klx
```

All cross-compilation runs on your local machine — no remote build servers needed.

The final binary has no external dependencies. End users do not need Go or Kylix installed to run it.

### LLVM Native Backend (v3.0.0-alpha, expanded in v3.1.0)

Kylix has an experimental LLVM backend that generates native binaries directly from the AST, bypassing the Go toolchain entirely.

```bash
# Compile via LLVM (requires llc + clang installed)
kylix build --backend=llvm main.klx

# Apply LLVM optimization passes (-O0 / -O1 / -O2 / -O3)
kylix build --backend=llvm --llvm-opt=2 main.klx
```

Pipeline: AST → LLVM IR (`.ll`) → object file (`.o`) → native binary via `llc` + `clang`. The Go backend remains the default for production use.

**Milestone 1 + Phase 1 (v3.1.0) support:**
- All scalar types, arithmetic/comparison/logic, control flow, functions
- Classes with vtable dispatch
- **Static arrays** (`array[1..N] of T` → `alloca [N x T]`)
- **Dynamic arrays** (`array of T` → `{ ptr, i64, i64 }` slice struct)
- Pascal 1-based indexing automatically converted to LLVM 0-based
- LLVM optimization passes via `--llvm-opt=N` (`llc -O=N`)

```pascal
program Arrays;
var
  fixed: array[1..5] of Integer;
  dyn: array of Integer;
  i: Integer;
begin
  for i := 1 to 5 do fixed[i] := i * i;
  WriteLn(fixed[3]);  // 9
end.
```

Generics, interfaces, and exceptions land in Milestone 2 (Phase 2-3, v3.2).

### WASI Target (v3.0.0-alpha)

Compile Kylix programs to run in WebAssembly System Interface runtimes (Wasmtime, Node.js, Cloudflare Workers):

```bash
# Standard WASI via Go 1.21+ (GOOS=wasip1 GOARCH=wasm)
kylix build --wasi main.klx

# Size-optimized via TinyGo (~30 KB vs ~3 MB)
kylix build --wasi --tinygo main.klx
```

Use the `wasi` stdlib module for portable I/O:

```pascal
uses wasi;

begin
  WriteLn('Running in WASI!');
  WriteLn('Args: ' + IntToStr(ArgCount()));
end.
```

### Supported targets

| OS | Architecture | `--target` value |
|----|-------------|-----------------|
| Linux | x86-64 | `linux/amd64` |
| Linux | ARM64 | `linux/arm64` |
| Windows | x86-64 | `windows/amd64` |
| macOS | x86-64 | `darwin/amd64` |
| macOS | Apple Silicon | `darwin/arm64` |
| WebAssembly | wasm | `--wasm` (optionally `--tinygo`) |
| WASI | wasip1/wasm | `--wasi` (optionally `--tinygo`) |

---

## Changelog

### v3.1.0 (2026-06-23) — KylixBoot Framework + Compiler Fixes + LLVM Arrays

KylixBoot framework (router/DI/middleware, 23 tests), annotation syntax `[Name]`, 5 compiler fixes (KLX-C01..C05: typed class vars, string interpolation, lambda return types, match codegen, uses-symbol injection), LLVM Milestone 2 Phase 1 (static + dynamic arrays, `--llvm-opt=N`).

### v3.0.0-alpha (2026-06-21) — Architecture Breakthrough 🚀

LLVM native backend (Milestone 1), WASI target, package registry server, stdlib Phase 4 (pure Kylix jsonutil/regex/datetime), `external` parsing fix, HTTP client stdlib.

### v2.6.0 (2026-06-20) — Performance & Optimization

Parallel compilation (goroutine pool), dead code elimination, LSP large-file performance benchmarks.

### v2.5.0 (2026-06-20) — Toolchain Deepening

LSP cross-file rename + code actions, `kylix doc` code example extraction, `kylix bench --mem`, `iter` module, class method external definition fix.

### v2.4.0 (2026-06-20) — Polish & Ecosystem

i18n fully integrated, REPL `:type` real inference, SetLength fixed, package manager nested deps + lockfile, stdlib Phase 3.

### v2.3.0 (2026-06-19) — Developer Experience

LSP incremental sync, REPL Tab/load/type, test fixtures + filter, i18n framework, Delve debug, WebAssembly backend.

### v2.2.0 (2026-06-19) — Engineering Quality

GitHub Actions CI, generic signature verification, project-level type checking, incremental compilation activated, stdlib Phase 2.

### v2.1.0 (2026-06-19) — Type System + stdlib Phase 1

Multi-param generic constraints, class→interface mapping, enhanced type inference, strutil + mathutil.

### v2.0.0 (2026-06-17) — Production-Ready Release

Error codes (KLX001–499), type inference, generic constraints, `kylix test/doc/bench` toolchain.

See [CHANGELOG.md](CHANGELOG.md) for full release history.

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

MIT License
