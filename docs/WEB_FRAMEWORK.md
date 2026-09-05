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

1. **Single-threaded**: The LLVM boot server (`BootRun`) serves one connection at a time; the Go backend uses Go's default HTTP server (one goroutine per request)
2. **Auth scope**: JWT auth is wired via `BootRegisterJwtAuth` (HS256, v0.6.8 real verification); OAuth is not built in
3. **WebSocket**: available as a separate stdlib module (`WsDial`/`WsAccept`/`WsSend`/`WsRecv`/`WsClose`, RFC 6455, v0.6.4+), not yet integrated into the web middleware chain

## Future Enhancements

- [x] ~~Authentication middleware (JWT)~~ ✅ v0.3.3 (`BootRegisterJwtAuth`) — real HS256 verification since v0.6.8
- [x] ~~Database ORM integration~~ ✅ v0.5.9 (`[Entity]`/`[Repository]` annotations)
- [x] ~~Template engine~~ ✅ available (`template` module, `res.HTML`)
- [x] ~~CORS middleware~~ ✅ `boot.CORS()`
- [x] ~~Request validation~~ ✅ v0.3.2 (`[Required]`/`[Email]`/`[Min]`...)
- [x] ~~API documentation generator~~ ✅ v0.3.3 (`kylix doc --openapi`, OpenAPI 3.1)
- [ ] OAuth support
- [ ] File upload handling
- [ ] Rate limiting middleware

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

1. 单线程——LLVM 端 boot server（`BootRun`）一次服务一个连接；Go 端用 Go 默认 HTTP server（每请求一个 goroutine）
2. 认证范围——JWT 经 `BootRegisterJwtAuth` 接入（HS256，v0.6.8 起真校验）；OAuth 未内置
3. WebSocket——作为独立 stdlib 模块提供（`WsDial`/`WsAccept`/`WsSend`/`WsRecv`/`WsClose`，RFC 6455，v0.6.4+），尚未并入 web 中间件链

### 未来增强

- [x] ~~认证中间件（JWT、OAuth）~~ ✅ v0.3.3 已完成
- [ ] 数据库 ORM 集成（v4.0）
- [ ] WebSocket 支持
- [x] ~~API 文档生成器~~ ✅ v0.3.3 已完成（OpenAPI 3.1）
- [ ] 速率限制中间件
- [ ] CORS 中间件
- [x] ~~请求验证~~ ✅ v0.3.2 已完成（[Required]/[Email]/[Min] 等）
- [ ] 文件上传处理

---

## KylixBoot 框架（v0.3.2+）

v0.3.2 引入了 **KylixBoot**，Spring Boot 风格的注解驱动 Web 框架，通过编译器代码生成实现零运行时反射。

### 自动路由装配

```pascal
program UserAPI;
uses boot;

[Controller('/api/users')]
type
  TUserController = class
    [Get('/')]
    function ListUsers(req: TRequest): TResponse;
    begin
      result := BootJSON(200, nil);
    end;

    [Post('/')]
    function CreateUser(req: TRequest): TResponse;
    begin
      result := BootText(201, 'created');
    end;
  end;

begin
  BootRun(8080);
end.
```

### 请求体绑定（v0.3.3）

`[Body(TEntity)]` 注解自动绑定并验证 JSON 请求体：

```pascal
[Entity('users')]
type
  TCreateUser = class
    [Required]
    [Email]
    Email: String;
    [Required]
    [MinLen(8)]
    Password: String;
  end;

[Controller('/api')]
type
  TUserController = class
    [Post('/users')]
    [Body(TCreateUser)]
    function CreateUser(req: TRequest): TResponse;
    begin
      // 编译器自动生成绑定 + 验证代码
      result := BootText(201, 'created');
    end;
  end;
```

编译器生成的 Go 代码：
```go
stdlib.BootPOST("/api/users", func(req *stdlib.BootRequest) *stdlib.BootResponse {
    var __body TCreateUser
    if err := stdlib.BootReadJSON(req, &__body); err != nil {
        return stdlib.BootJSON(400, map[string]string{"error": "invalid JSON"})
    }
    if !__body.IsValid() {
        return stdlib.BootJSON(400, __body.Validate())
    }
    return __kylix_ctrl_TUserController.CreateUser(req)
})
```

### JWT 认证（v0.3.3）

```pascal
uses boot, jwt;

[Controller('/api')]
type
  TAuthController = class
    [Post('/login')]
    function Login(req: TRequest): TResponse;
    begin
      var token := JwtSign('my-secret', 'user42', 3600, nil);
      result := BootText(200, token);
    end;

    [Get('/me')]
    [Authenticated]
    function Me(req: TRequest): TResponse;
    begin
      result := BootText(200, 'authenticated!');
    end;
  end;

begin
  BootRegisterJwtAuth('my-secret');  // 一键接入 [Authenticated]
  BootRun(8080);
end.
```

### OpenAPI 3.1 自动生成（v0.3.3）

```bash
# 从源码生成 openapi.yaml
kylix doc --openapi --title "My API" --api-version 0.1.0 main.klx

# 输出到标准输出
kylix doc --openapi --stdout main.klx
```

生成结果示例：
```yaml
openapi: "0.3.1"
info:
  title: My API
  version: 0.1.0
paths:
  /api/users:
    post:
      security:
        - BearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/TCreateUser'
components:
  schemas:
    TCreateUser:
      type: object
      required: [Email, Password]
      properties:
        Email: {type: string, format: email}
        Password: {type: string, minLength: 8}
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
```

### 支持的注解一览

| 注解 | 作用 | 级别 |
|------|------|------|
| `[Controller('/path')]` | 定义 HTTP 控制器基路径 | class |
| `[Get('/path')]` | GET 路由 | method |
| `[Post('/path')]` | POST 路由 | method |
| `[Put('/path')]` | PUT 路由 | method |
| `[Delete('/path')]` | DELETE 路由 | method |
| `[Body(TEntity)]` | JSON 请求体绑定 + 验证 | method |
| `[Authenticated]` | 要求已登录 | method |
| `[Role('admin')]` | 要求指定角色 | method |
| `[Service]` | 注册为 DI 服务 | class |
| `[Inject]` | 注入依赖 | field |
| `[Required]` | 字段必填 | field |
| `[Email]` | Email 格式验证 | field |
| `[Min(n)]` / `[Max(n)]` | 数值范围 | field |
| `[MinLen(n)]` / `[MaxLen(n)]` | 字符串长度 | field |
| `[Entity('table')]` | ORM 实体映射 | class |
| `[Column('name')]` | 列名映射 | field |
| `[PrimaryKey]` | 主键 | field |

---

## v0.7.0 Web 页面开发规划（HTML 页面 + 框架补强）

KylixBoot 已有 REST API（JSON/路由/中间件/静态文件）。v0.7.0 补齐 **HTML 页面开发**能力，让 Kylix 能开发完整 web 应用（页面 + API），框架 API 对齐 Spring Boot/Go `html/template` 风格。

### 1. 模板引擎（HTML 渲染）

- **Go 端**：复用 `html/template`（自动 HTML 转义，防 XSS）。
- **LLVM 端**：手写轻量模板渲染（`{{var}}` 替换 + `{{for}}` 循环 + `{{if}}` 条件），字符串替换 + 循环展开 IR。
- **API**：
  ```pascal
  res.HTML(status, 'page.html');              // 渲染模板（无数据）
  res.Render('page.html', data);              // 模板 + 数据（map[String]Variant）
  ```
- **模板语法**（mustache/Go template 子集）：
  - `{{title}}` 变量替换
  - `{{user.name}}` 字段/键访问（嵌套）
  - `{{for item in items}}...{{end}}` 循环
  - `{{if cond}}...{{else}}...{{end}}` 条件

### 2. 表单处理

```pascal
var name := req.Form('name');     // POST application/x-www-form-urlencoded
var age  := req.FormInt('age');   // 数字字段
var ok   := req.FormBool('agree');// 复选框
var f    := req.FormFile('avatar');// 文件上传（可选）
```

### 3. Cookie 与会话

```pascal
var sid := req.Cookie('sid');                  // 读 Cookie
res.SetCookie('sid', 'abc123', 3600);          // 写 Cookie（maxAge 秒）
```
简单会话（签名 Cookie）可作为后续增强。

### 4. 静态资源与项目结构约定

```
app/
├── main.klx          # BootRun(8080) + 路由注册
├── views/            # HTML 模板（res.Render('index.html') 默认从这找）
├── static/           # css/js/img（app.Static('/static', 'static/')）
└── controllers/      # [Controller] 类（可选）
```

### 5. 页面辅助 API

- `res.Redirect('/login')` — 302 重定向
- `res.SetHeader(name, value)` — 自定义响应头
- `res.Status(status)` 链式（已有）

### 6. 教程（新增 example58_web_page）

完整 web 页面应用：首页（模板 `{{for}}` 渲染列表）+ 表单提交页（`req.Form` + 回显）+ 登录/登出（Cookie + `[Authenticated]`）+ 静态资源（style.css）。双后端（Go + LLVM）行为一致。

### 7. 实现顺序

1. **Go 端模板**：`html/template` 封装（`res.HTML/Render`）→ 快
2. **表单/Cookie**：`req.Form*` / `res.SetCookie`（Go 端先）
3. **模板语法定稿**：`{{var}}`/`{{for}}`/`{{if}}` 子集
4. **LLVM 端模板渲染**：手写 IR（字符串替换 + 循环展开）
5. **教程 + 文档**：example58 + 本文档补全

> 完成后 KylixBoot = REST API + HTML 页面 + 静态资源 + 表单 + Cookie，成为完整 web 开发框架。全部 v0.6.9/v0.7.0 规划完成后发布 1.0.0。
