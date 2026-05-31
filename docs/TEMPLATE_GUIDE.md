# Kylix 模板引擎指南 / Kylix Template Engine Guide

## 概述 / Overview

Kylix 模板引擎提供 HTML 模板渲染功能，支持：
- 变量替换（Go text/template 语法）
- 布局（Layout）系统
- 片段（Partial）复用
- 丰富的内置函数
- 模板缓存
- 自定义函数

Kylix Template Engine provides HTML template rendering with:
- Variable substitution (Go text/template syntax)
- Layout system
- Partial template reuse
- Rich built-in functions
- Template caching
- Custom functions

---

## 快速开始 / Quick Start

```pascal
uses template;

var
  engine: TTemplateEngine;

begin
  engine := NewTemplateEngine;
  engine.SetTemplateDir('./templates');

  // 简单渲染
  var html := engine.RenderString('Hello, {{.Name}}!', map[string]interface{}{
    'Name': 'World'
  });
  // 输出: Hello, World!
end.
```

---

## 核心概念 / Core Concepts

### 1. 模板引擎 / Template Engine

```pascal
var engine: TTemplateEngine;
engine := NewTemplateEngine;

// 配置
engine.SetTemplateDir('./templates');  // 设置模板目录
engine.SetCache(true);                 // 启用缓存（生产环境）
engine.SetCache(false);                // 禁用缓存（开发环境）
```

### 2. 渲染方式 / Rendering Methods

```pascal
// 方式一：渲染字符串
var html := engine.RenderString('<h1>{{.Title}}</h1>', data);

// 方式二：渲染模板文件
var html := engine.Render('page.html', data);

// 方式三：渲染文件（绝对/相对路径）
var html := engine.RenderFile('./templates/page.html', data);

// 方式四：带布局渲染
var html := engine.RenderWithLayout('main', 'page.html', data);

// 方式五：使用 View 对象
var view := NewView(engine);
view.With('Title', 'Home');
view.With('Message', 'Welcome!');
view.WithLayout('main');
var html := view.Render('home.html');
```

---

## 布局系统 / Layout System

布局定义页面的整体结构，通过 `{{.Content}}` 插入页面内容。

Layouts define the overall page structure, with `{{.Content}}` for inserting page content.

```pascal
// 注册布局
engine.RegisterLayout('main', 
  '<!DOCTYPE html>\n' +
  '<html>\n' +
  '<head>\n' +
  '  <title>{{.Title}}</title>\n' +
  '</head>\n' +
  '<body>\n' +
  '  {{include "header"}}\n' +
  '  <main>{{.Content}}</main>\n' +
  '  {{include "footer"}}\n' +
  '</body>\n' +
  '</html>'
);

// 使用布局渲染
var html := engine.RenderWithLayout('main', 'page.html', data);
```

---

## 片段系统 / Partial System

片段是可复用的模板片段，通过 `{{include "name"}}` 引用。

Partials are reusable template snippets, included with `{{include "name"}}`.

```pascal
// 注册片段
engine.RegisterPartial('header', 
  '<nav><a href="/">Home</a> | <a href="/users">Users</a></nav>'
);

engine.RegisterPartial('footer', 
  '<footer>&copy; 2026 Kylix App</footer>'
);

// 从文件加载片段
engine.LoadPartialFile('sidebar', 'partials/sidebar.html');
engine.LoadLayoutFile('admin', 'layouts/admin.html');
```

在模板中使用：
```html
{{include "header"}}
<h1>{{.Title}}</h1>
{{include "footer"}}
```

---

## 模板语法 / Template Syntax

### 变量 / Variables

```html
{{.Name}}              <!-- 输出变量 -->
{{.User.Name}}         <!-- 嵌套属性 -->
{{.User.Email}}
```

### 条件 / Conditionals

```html
{{if .LoggedIn}}
  <p>Welcome, {{.User.Name}}!</p>
{{else}}
  <p>Please log in.</p>
{{end}}

{{if and .Active .Verified}}
  <p>Account is active and verified.</p>
{{end}}

{{if or .IsAdmin .IsModerator}}
  <p>Management panel</p>
{{end}}
```

### 循环 / Loops

```html
{{range .Users}}
  <li>{{.Name}} - {{.Email}}</li>
{{end}}

{{range $index, $user := .Users}}
  <li>{{$index}}: {{$user.Name}}</li>
{{end}}

{{range .Items}}
  {{if .Active}}
    <span>{{.Name}}</span>
  {{end}}
{{else}}
  <p>No items found.</p>
{{end}}
```

---

## 内置函数 / Built-in Functions

### 字符串函数 / String Functions

```html
{{upper "hello"}}       <!-- HELLO -->
{{lower "HELLO"}}       <!-- hello -->
{{title "hello world"}} <!-- Hello World -->
{{trim "  hello  "}}    <!-- hello -->
{{replace "hello world" "world" "go"}}  <!-- hello go -->
{{split "a,b,c" ","}}   <!-- [a b c] -->
{{join .Items ", "}}    <!-- a, b, c -->
{{contains "hello" "ell"}}     <!-- true -->
{{hasPrefix "hello" "he"}}     <!-- true -->
{{hasSuffix "hello" "lo"}}     <!-- true -->
```

### 数学函数 / Math Functions

```html
{{add 3 5}}    <!-- 8 -->
{{sub 10 3}}   <!-- 7 -->
{{mul 4 5}}    <!-- 20 -->
{{div 20 4}}   <!-- 5 -->
{{mod 10 3}}   <!-- 1 -->
```

### 比较函数 / Comparison Functions

```html
{{if eq .Status "active"}}...{{end}}
{{if ne .Status "deleted"}}...{{end}}
{{if lt .Age 18}}Minor{{end}}
{{if gt .Score 90}}Excellent{{end}}
{{if le .Count 0}}Empty{{end}}
{{if ge .Score 60}}Passed{{end}}
```

### 逻辑函数 / Logical Functions

```html
{{if and .Active .Verified}}...{{end}}
{{if or .IsAdmin .IsEditor}}...{{end}}
{{if not .Deleted}}...{{end}}
```

### 工具函数 / Utility Functions

```html
{{len .Items}}           <!-- 数组/字符串长度 -->
{{len "hello"}}          <!-- 5 -->
{{default "N/A" .Name}}  <!-- 默认值 -->
{{html "<script>"}}      <!-- HTML 转义 -->
{{safe "<b>bold</b>"}}   <!-- 不转义（原始 HTML）-->
{{format "Hello %s!" .Name}}  <!-- 格式化 -->
{{toString 42}}          <!-- "42" -->
{{toInt "42"}}           <!-- 42 -->
{{urlEncode "hello world"}}  <!-- hello%20world -->
```

### 数组函数 / Array Functions

```html
{{first .Items}}   <!-- 第一个元素 -->
{{last .Items}}    <!-- 最后一个元素 -->
{{index .Items 0}} <!-- 按索引访问 -->
{{get .Map "key"}} <!-- Map 键访问 -->
```

---

## 自定义函数 / Custom Functions

```pascal
// 添加自定义函数
engine.AddFunc('formatPrice', function(price: Real): String
begin
  result := Format('$%.2f', [price]);
end);

// 在模板中使用
// {{formatPrice .Product.Price}}
```

---

## View 对象 / View Object

View 提供链式 API 来构建视图数据：

View provides a fluent API for building view data:

```pascal
var view: TView;
view := NewView(engine);

// 链式添加数据
view.With('Title', 'Dashboard')
    .With('User', currentUser)
    .With('Stats', dashboardStats);

// 批量添加
view.WithData(map[string]interface{}{
  'Title': 'Dashboard',
  'User': currentUser,
  'Stats': dashboardStats
});

// 设置布局
view.WithLayout('main');

// 渲染
var html := view.Render('dashboard.html');
```

---

## 与 Web 框架集成 / Web Framework Integration

```pascal
uses web, template;

var
  app: TServer;
  engine: TTemplateEngine;

begin
  // 初始化
  engine := NewTemplateEngine;
  engine.SetTemplateDir('./templates');
  engine.SetCache(true);

  // 注册布局和片段
  engine.RegisterLayout('main', '...');
  engine.RegisterPartial('header', '...');

  app := NewServer(8080);

  // 渲染 HTML 页面
  app.Get('/', procedure(req: TRequest; res: TResponse)
  var
    view: TView;
  begin
    view := NewView(engine);
    view.With('Title', 'Home');
    view.With('Message', 'Welcome!');
    view.WithLayout('main');

    var html := view.Render('home.html');
    res.HTML(html);
  end);

  // 列表页面
  app.Get('/users', procedure(req: TRequest; res: TResponse)
  var
    view: TView;
    users: array of map[string]interface{};
  begin
    users := getAllUsers;

    view := NewView(engine);
    view.With('Title', 'Users');
    view.With('Users', users);
    view.With('Count', length(users));
    view.WithLayout('main');

    var html := view.Render('users.html');
    res.HTML(html);
  end);

  // API 返回渲染后的 HTML 响应
  app.Get('/page', procedure(req: TRequest; res: TResponse)
  begin
    var resp := engine.RenderToResponse('page.html', data, 200);
    res.Status(resp.StatusCode);
    res.HTML(resp.Content);
  end);

  app.Listen;
end.
```

---

## 模板文件示例 / Template File Examples

### templates/home.html
```html
<div class="hero">
  <h1>{{.Title}}</h1>
  <p>{{.Message}}</p>
</div>

<div class="features">
  {{range .Features}}
  <div class="feature">
    <h3>{{.Name}}</h3>
    <p>{{.Description}}</p>
  </div>
  {{end}}
</div>
```

### templates/users.html
```html
<h1>{{.Title}} ({{.Count}})</h1>

<table>
  <thead>
    <tr>
      <th>Name</th>
      <th>Email</th>
      <th>Joined</th>
    </tr>
  </thead>
  <tbody>
    {{range .Users}}
    <tr>
      <td>{{.name}}</td>
      <td>{{.email}}</td>
      <td>{{.created_at}}</td>
    </tr>
    {{else}}
    <tr>
      <td colspan="3">No users found.</td>
    </tr>
    {{end}}
  </tbody>
</table>
```

### templates/profile.html
```html
<div class="profile">
  <h1>{{.User.name}}</h1>
  <p>Email: {{.User.email}}</p>
  <p>Age: {{default "N/A" .User.age}}</p>
  
  {{if eq .User.role "admin"}}
    <div class="admin-badge">Administrator</div>
  {{end}}
</div>
```

---

## 缓存管理 / Cache Management

```pascal
// 启用缓存（生产环境推荐）
engine.SetCache(true);

// 禁用缓存（开发环境推荐）
engine.SetCache(false);

// 手动清除缓存
engine.ClearCache;
```

---

## API 参考 / API Reference

### TTemplateEngine
- `NewTemplateEngine()`: Create engine
- `SetTemplateDir(dir)`: Set template directory
- `SetCache(enabled)`: Enable/disable caching
- `AddFunc(name, fn)`: Add custom function
- `RegisterLayout(name, content)`: Register layout
- `RegisterPartial(name, content)`: Register partial
- `LoadLayoutFile(name, path)`: Load layout from file
- `LoadPartialFile(name, path)`: Load partial from file
- `Render(name, data)`: Render template
- `RenderString(str, data)`: Render string template
- `RenderFile(path, data)`: Render file
- `RenderWithLayout(layout, name, data)`: Render with layout
- `RenderToResponse(name, data, status)`: Render to HTTP response
- `ClearCache()`: Clear template cache

### TView
- `NewView(engine)`: Create view
- `With(key, value)`: Add data (fluent)
- `WithData(map)`: Add multiple data (fluent)
- `WithLayout(name)`: Set layout (fluent)
- `Render(name)`: Render the view

### TTemplateResponse
- `Content`: Rendered HTML string
- `StatusCode`: HTTP status code
- `Headers`: Response headers map
