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
│   └── autoconfig.go   # Auto-configuration
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

### Phase 3: Framework (In Progress)
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

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

MIT License
