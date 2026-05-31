# Kylix Web Framework

[![中文文档](https://img.shields.io/badge/lang-中文-red.svg)](#kylix-web-框架)

A lightweight, Spring Boot-inspired web framework for Kylix that compiles to Go's `net/http`.

## Features

- **Simple Routing**: GET, POST, PUT, DELETE methods
- **Path Parameters**: `/users/:id` syntax
- **Query Parameters**: Automatic parsing
- **JSON Support**: Built-in JSON request/response handling
- **Middleware**: Composable middleware chain
- **Static Files**: Serve static assets
- **Type Safety**: Strongly typed request/response objects

## Quick Start

### 1. Create a Simple Server

```pascal
program SimpleWeb;
uses web;
var
  app: TServer;
begin
  app := web.createServer(8080);
  
  app.get('/', procedure(req: TRequest; res: TResponse)
  begin
    res.send('Hello, Kylix Web!');
  end);
  
  app.listen();
end.
```

### 2. Run the Server

```bash
./kylix run simple_web.klx
```

### 3. Test the Endpoint

```bash
curl http://localhost:8080/
# Output: Hello, Kylix Web!
```

## Core Concepts

### TServer

The main server object that manages routes, middleware, and static files.

```pascal
var app: TServer;
app := web.createServer(8080);  // Port 8080
```

### Routing

Register handlers for different HTTP methods and paths:

```pascal
// GET request
app.get('/api/users', procedure(req: TRequest; res: TResponse)
begin
  res.json(users);
end);

// POST request
app.post('/api/users', procedure(req: TRequest; res: TResponse)
begin
  // Create user
  res.status(201).json(newUser);
end);

// PUT request
app.put('/api/users/:id', procedure(req: TRequest; res: TResponse)
begin
  // Update user
  res.json(updatedUser);
end);

// DELETE request
app.delete('/api/users/:id', procedure(req: TRequest; res: TResponse)
begin
  // Delete user
  res.status(204).send('');
end);
```

### Path Parameters

Extract dynamic values from URLs:

```pascal
app.get('/api/users/:id', procedure(req: TRequest; res: TResponse)
var
  userId: String;
begin
  userId := req.param('id');  // Extract :id from URL
  res.json(record id := userId; end);
end);
```

**Example**:
```bash
curl http://localhost:8080/api/users/123
# Response: {"id": "123"}
```

### Query Parameters

Access URL query string values:

```pascal
app.get('/search', procedure(req: TRequest; res: TResponse)
var
  query: String;
  page: String;
begin
  query := req.query('q');
  page := req.query('page');
  res.json(record query := query; page := page; end);
end);
```

**Example**:
```bash
curl "http://localhost:8080/search?q=kylix&page=1"
# Response: {"query": "kylix", "page": "1"}
```

### Request Headers

Read HTTP headers:

```pascal
app.get('/api/data', procedure(req: TRequest; res: TResponse)
var
  authToken: String;
begin
  authToken := req.header('Authorization');
  // Validate token...
  res.json(data);
end);
```

### JSON Handling

#### Sending JSON Responses

```pascal
app.get('/api/user', procedure(req: TRequest; res: TResponse)
var
  user: record
    id: Integer;
    name: String;
    email: String;
  end;
begin
  user.id := 1;
  user.name := 'Alice';
  user.email := 'alice@example.com';
  
  res.json(user);  // Automatically serializes to JSON
end);
```

#### Receiving JSON Requests

```pascal
app.post('/api/user', procedure(req: TRequest; res: TResponse)
var
  newUser: record
    name: String;
    email: String;
  end;
begin
  req.json(newUser);  // Automatically deserializes from JSON
  
  // Process newUser...
  
  res.status(201).json(newUser);
end);
```

**Test with curl**:
```bash
curl -X POST http://localhost:8080/api/user \
  -H "Content-Type: application/json" \
  -d '{"name": "Bob", "email": "bob@example.com"}'
```

### Response Methods

```pascal
// Send plain text
res.send('Hello World');

// Send JSON
res.json(record message := 'success'; end);

// Set status code
res.status(201).send('Created');
res.status(404).send('Not Found');
res.status(500).json(record error := 'Internal Server Error'; end);

// Set headers
res.header('Content-Type', 'text/plain');
res.header('X-Custom-Header', 'value');

// Chain methods
res.status(201)
   .header('Location', '/api/users/123')
   .json(newUser);
```

### Middleware

Middleware functions execute before route handlers:

```pascal
// Logger middleware
app.use(web.loggerMiddleware());

// Custom middleware
app.use(procedure(req: TRequest; res: TResponse)
begin
  WriteLn('Request: ', req.method(), ' ', req.path());
  // Middleware logic here
end);
```

**Built-in Middleware**:
- `web.loggerMiddleware()`: Logs all requests

### Static Files

Serve static assets (HTML, CSS, JS, images):

```pascal
// Serve files from ./static directory at /public path
app.static('/public', './static');
```

**Directory Structure**:
```
project/
├── main.klx
└── static/
    ├── index.html
    ├── style.css
    └── app.js
```

**Access**:
```bash
curl http://localhost:8080/public/index.html
curl http://localhost:8080/public/style.css
```

## Complete Example: REST API

```pascal
program RestAPI;
uses web;
var
  app: TServer;
  users: array of record
    id: Integer;
    name: String;
    email: String;
  end;
  nextId: Integer;
begin
  app := web.createServer(8080);
  nextId := 1;
  
  // Logger middleware
  app.use(web.loggerMiddleware());
  
  // GET all users
  app.get('/api/users', procedure(req: TRequest; res: TResponse)
  begin
    res.json(users);
  end);
  
  // GET user by ID
  app.get('/api/users/:id', procedure(req: TRequest; res: TResponse)
  var
    id: Integer;
    i: Integer;
    found: Boolean;
  begin
    id := StrToInt(req.param('id'));
    found := false;
    
    for i := 0 to Length(users) - 1 do
    begin
      if users[i].id = id then
      begin
        res.json(users[i]);
        found := true;
        break;
      end;
    end;
    
    if not found then
      res.status(404).json(record error := 'User not found'; end);
  end);
  
  // POST create user
  app.post('/api/users', procedure(req: TRequest; res: TResponse)
  var
    newUser: record
      name: String;
      email: String;
    end;
    createdUser: record
      id: Integer;
      name: String;
      email: String;
    end;
  begin
    req.json(newUser);
    
    createdUser.id := nextId;
    createdUser.name := newUser.name;
    createdUser.email := newUser.email;
    Inc(nextId);
    
    // Add to users array
    SetLength(users, Length(users) + 1);
    users[Length(users) - 1] := createdUser;
    
    res.status(201).json(createdUser);
  end);
  
  // PUT update user
  app.put('/api/users/:id', procedure(req: TRequest; res: TResponse)
  var
    id: Integer;
    i: Integer;
    updates: record
      name: String;
      email: String;
    end;
    found: Boolean;
  begin
    id := StrToInt(req.param('id'));
    req.json(updates);
    found := false;
    
    for i := 0 to Length(users) - 1 do
    begin
      if users[i].id = id then
      begin
        users[i].name := updates.name;
        users[i].email := updates.email;
        res.json(users[i]);
        found := true;
        break;
      end;
    end;
    
    if not found then
      res.status(404).json(record error := 'User not found'; end);
  end);
  
  // DELETE user
  app.delete('/api/users/:id', procedure(req: TRequest; res: TResponse)
  var
    id: Integer;
    i: Integer;
    j: Integer;
    newUsers: array of record
      id: Integer;
      name: String;
      email: String;
    end;
    found: Boolean;
  begin
    id := StrToInt(req.param('id'));
    found := false;
    
    SetLength(newUsers, Length(users) - 1);
    j := 0;
    
    for i := 0 to Length(users) - 1 do
    begin
      if users[i].id = id then
        found := true
      else
      begin
        newUsers[j] := users[i];
        Inc(j);
      end;
    end;
    
    if found then
    begin
      users := newUsers;
      res.status(204).send('');
    end
    else
      res.status(404).json(record error := 'User not found'; end);
  end);
  
  WriteLn('Starting REST API Server...');
  app.listen();
end.
```

**Test the API**:

```bash
# Create user
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "email": "alice@example.com"}'

# Get all users
curl http://localhost:8080/api/users

# Get user by ID
curl http://localhost:8080/api/users/1

# Update user
curl -X PUT http://localhost:8080/api/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice Smith", "email": "alice.smith@example.com"}'

# Delete user
curl -X DELETE http://localhost:8080/api/users/1
```

## API Reference

### TServer Methods

| Method | Description |
|--------|-------------|
| `web.createServer(port: Integer): TServer` | Create a new server instance |
| `app.get(path: String; handler: TRouteHandler)` | Register GET route |
| `app.post(path: String; handler: TRouteHandler)` | Register POST route |
| `app.put(path: String; handler: TRouteHandler)` | Register PUT route |
| `app.delete(path: String; handler: TRouteHandler)` | Register DELETE route |
| `app.use(middleware: TMiddleware)` | Add middleware |
| `app.static(pathPrefix: String; rootDir: String)` | Serve static files |
| `app.listen()` | Start the server |

### TRequest Methods

| Method | Description |
|--------|-------------|
| `req.path(): String` | Get request path |
| `req.method(): String` | Get HTTP method (GET, POST, etc.) |
| `req.param(name: String): String` | Get path parameter |
| `req.query(name: String): String` | Get query parameter |
| `req.header(name: String): String` | Get request header |
| `req.json(var data: Record)` | Parse JSON body into record |

### TResponse Methods

| Method | Description |
|--------|-------------|
| `res.send(body: String)` | Send text response |
| `res.json(data: Record)` | Send JSON response |
| `res.status(code: Integer): TResponse` | Set status code (chainable) |
| `res.header(name: String; value: String): TResponse` | Set response header (chainable) |

### Built-in Middleware

| Middleware | Description |
|------------|-------------|
| `web.loggerMiddleware()` | Logs request method, path, and timestamp |

## Best Practices

### 1. Organize Routes

Group related routes together:

```pascal
// User routes
app.get('/api/users', getUsers);
app.post('/api/users', createUser);
app.get('/api/users/:id', getUser);
app.put('/api/users/:id', updateUser);
app.delete('/api/users/:id', deleteUser);

// Product routes
app.get('/api/products', getProducts);
app.post('/api/products', createProduct);
```

### 2. Use Middleware for Cross-Cutting Concerns

```pascal
// Logging
app.use(web.loggerMiddleware());

// Authentication
app.use(authMiddleware);

// CORS
app.use(corsMiddleware);
```

### 3. Consistent Error Handling

```pascal
app.get('/api/users/:id', procedure(req: TRequest; res: TResponse)
begin
  try
    // Business logic
    res.json(user);
  except
    res.status(500).json(record
      error := 'Internal Server Error';
      message := 'Failed to fetch user';
    end);
  end;
end);
```

### 4. Use Status Codes Correctly

```pascal
// Success
res.status(200).json(data);           // OK
res.status(201).json(newUser);        // Created
res.status(204).send('');             // No Content

// Client Errors
res.status(400).json(error);          // Bad Request
res.status(401).json(error);          // Unauthorized
res.status(403).json(error);          // Forbidden
res.status(404).json(error);          // Not Found

// Server Errors
res.status(500).json(error);          // Internal Server Error
```

## Limitations

1. **Single-threaded**: Currently uses Go's default HTTP server (single goroutine per request)
2. **No Authentication**: Built-in auth middleware not yet available
3. **No Database ORM**: Manual data management required
4. **Basic Middleware**: Only logger middleware included
5. **No WebSocket**: Real-time features not supported yet

## Future Enhancements

- [ ] Authentication middleware (JWT, OAuth)
- [ ] Database ORM integration
- [ ] WebSocket support
- [ ] Template engine
- [ ] File upload handling
- [ ] Rate limiting middleware
- [ ] CORS middleware
- [ ] Request validation
- [ ] API documentation generator

## Examples

See the `examples/` directory for complete examples:

- `web_demo.klx` - Basic web server
- `web_simple.klx` - Simple routes
- `web_rest_api.klx` - REST API implementation
- `web_middleware.klx` - Middleware usage

## Related Documentation

- [IDE User Manual](KYLIX_IDE_USER_MANUAL.md) - CLI and editor guide
- [Developer Guide](KYLIX_DEV_GUIDE.md) - Architecture and contributing
- [Tools Explained](KYLIX_TOOLS_EXPLAINED.md) - Tool concepts

---

## Kylix Web 框架

[![English](https://img.shields.io/badge/lang-English-blue.svg)](#kylix-web-framework)

一个轻量级的、受 Spring Boot 启发的 Kylix Web 框架，编译为 Go 的 `net/http`。

### 特性

- **简单路由**：GET、POST、PUT、DELETE 方法
- **路径参数**：`/users/:id` 语法
- **查询参数**：自动解析
- **JSON 支持**：内置 JSON 请求/响应处理
- **中间件**：可组合的中间件链
- **静态文件**：提供静态资源服务
- **类型安全**：强类型的请求/响应对象

### 快速开始

```pascal
program SimpleWeb;
uses web;
var
  app: TServer;
begin
  app := web.createServer(8080);
  
  app.get('/', procedure(req: TRequest; res: TResponse)
  begin
    res.send('你好，Kylix Web！');
  end);
  
  app.listen();
end.
```

### 核心概念

- **TServer**：管理路由、中间件和静态文件的主服务器对象
- **路由**：为不同的 HTTP 方法和路径注册处理器
- **路径参数**：从 URL 中提取动态值
- **JSON 处理**：自动序列化和反序列化
- **中间件**：在路由处理器之前执行的函数

### 完整示例

参见上面的"Complete Example: REST API"部分。

### API 参考

参见上面的"API Reference"部分。

### 最佳实践

1. **组织路由**：将相关的路由分组
2. **使用中间件**：处理跨领域关注点（日志、认证等）
3. **一致的错误处理**：使用 try-except 块
4. **正确使用状态码**：200、201、404、500 等

### 限制

1. 单线程（每个请求一个 goroutine）
2. 尚无内置认证中间件
3. 无数据库 ORM
4. 基础中间件（仅日志记录器）
5. 无 WebSocket 支持

### 未来增强

- [ ] 认证中间件（JWT、OAuth）
- [ ] 数据库 ORM 集成
- [ ] WebSocket 支持
- [ ] 模板引擎
- [ ] 文件上传处理
- [ ] 速率限制中间件
- [ ] CORS 中间件
- [ ] 请求验证
- [ ] API 文档生成器
