# Kylix 模板引擎指南 / Kylix Template Engine Guide

## 概述 / Overview

Kylix 模板引擎（`stdlib/template_engine.klx`，纯 Kylix 实现）提供 HTML 模板渲染功能，支持：
- 变量替换（Mustache 风格 `{{ }}`，自动 HTML 转义）
- 过滤器管道（12 个内置过滤器）
- `{{#each}}` / `{{#if}}` 块
- 布局（Layout）继承（`{{< name }}` + `{{{ content }}}`，v0.9.0）
- 片段（Partial）复用（`{{> name }}`，v0.9.0）
- **三端同源**：Go / LLVM / bootstrap 编译器行为逐字一致

Kylix Template Engine (`stdlib/template_engine.klx`, pure Kylix) provides HTML template rendering with:
- Variable substitution (Mustache-style `{{ }}`, auto HTML escaping)
- Filter pipelines (12 built-in filters)
- `{{#each}}` / `{{#if}}` blocks
- Layout inheritance (`{{< name }}` + `{{{ content }}}`, v0.9.0)
- Partial template reuse (`{{> name }}`, v0.9.0)
- **Three-end parity**: byte-identical output on Go, LLVM and bootstrap compilers

---

## 快速开始 / Quick Start

```pascal
uses template;

var
  eng: TTemplateEngine;
  out: String;
  err: error;

begin
  eng := TTemplateEngine.Create();
  eng.AddVar('name', 'World');

  // 简单渲染（推荐 RenderTemplate：返回 (输出, 错误)）
  (out, err) := RenderTemplate(eng, 'Hello, {{ name }}!');
  if err <> nil then
    WriteLn('error: ', ErrorStr(err))
  else
    WriteLn(out);
  // 输出: Hello, World!
end.
```

---

## 核心概念 / Core Concepts

### 1. 数据模型 / Data Model

全展平字符串模型——所有值在渲染前都被展平为 `Scalars['path.to.key']` 字符串查找（无嵌套 map、无 Variant 装箱）：

```pascal
eng.AddVar('title', 'Hi');              // Scalars['title'] = 'Hi'
eng.AddVar('user.name', 'Li');          // Scalars['user.name'] = 'Li'（点号路径 key）
eng.BeginList('tags'); eng.AddItem('x'); eng.EndList;
                                        // Scalars['tags.0'] = 'x' + ListLens['tags'] = 1
eng.BeginList('users'); eng.BeginItem;
eng.ItemField('name', 'Li'); eng.NextItem; eng.EndList;
                                        // Scalars['users.0.name'] = 'Li'
```

### 2. 渲染方式 / Rendering Methods

```pascal
// 方式一：RenderTemplate —— 返回 (String, error)，推荐
var (out, err) := RenderTemplate(eng, '<h1>{{ title }}</h1>');

// 方式二：RenderString —— 出错返回空串，ErrorMsg() 取错误消息
var html := eng.RenderString('<h1>{{ title }}</h1>');
if html = '' then
  WriteLn(eng.ErrorMsg());
```

---

## 布局系统 / Layout System（v0.9.0，三端同源）

`{{< name }}` 布局继承：`{{< base }}` 之后的剩余模板成为 `content`，由 base 通过 `{{{ content }}}` 插入（三重括号不转义）。`{{< layout base }}` 为等价别名。

`{{< name }}` layout inheritance: everything after the tag becomes `content`, inserted by the base via `{{{ content }}}` (triple braces = no escaping). `{{< layout base }}` is an equivalent alias.

```pascal
uses template;

var eng := TTemplateEngine.Create();
eng.AddVar('title', 'Admin');
eng.AddTemplate('base',
  '<html><title>{{ title }}</title><body>{{{ content }}}</body></html>');

// '{{< base }}' 之后的 '<h1>Dashboard</h1>' 成为 content
var (out, err) := RenderTemplate(eng, '{{< base}}<h1>{{ page }}</h1>');
// out = <html><title>Admin</title><body><h1>Dashboard</h1></body></html>
```

规则 / Rules：

- base 模板自身也可用 `{{< other }}` 继承，支持多层嵌套（Depth 上限 32，超出报 `partial recursion too deep`）；
- 引擎渲染完成后 `content` 变量被清除（下次渲染不再残留）；
- 渲染 base 时点号作用域（each 项）保持在 `{{<` 出现处的作用域。

---

## 片段系统 / Partial System（v0.9.0，三端同源）

`{{> name }}` 内联片段：片段从模板注册表（`AddTemplate`）查找，在当前位置展开，**继承当前 each 作用域**（`{{ . }}`、点号字段、`{{ @index }}` 均可见）。

`{{> name }}` inlines a partial looked up from the template registry (`AddTemplate`) at the current position, **inheriting the current each-scope** (`{{ . }}`, dotted fields, `{{ @index }}` all visible).

```pascal
eng.AddTemplate('row', '<li>{{ @index }}:{{ . }}</li>');
eng.BeginList('tags');
eng.AddItem('go');
eng.AddItem('llvm');
eng.EndList;

var (out, err) := RenderTemplate(eng, '<ul>{{#each tags}}{{> row}}{{/each}}</ul>');
// out = <ul><li>0:go</li><li>1:llvm</li></ul>
```

API：

| 方法 / Method | 说明 / Description |
|---|---|
| `AddTemplate(name, src)` | 注册命名模板（布局或片段共用一个注册表）/ register a named template (layouts and partials share one registry) |
| `HasTemplate(name)` | 查询是否已注册 / check registration |

错误 / Errors（`RenderTemplate` 返回 `err`，用 `ErrorStr` 取消息）：

- `unknown partial: <name>` — `{{> }}` 引用未注册片段 / unregistered partial
- `unknown layout: <name>` — `{{< }}` 引用未注册布局 / unregistered layout
- `partial recursion too deep` — 嵌套超 32 层 / nesting over 32

完整示例见 `examples/complete-tutorial/25_template_layout/example63_template_layout.klx`（三端输出逐字一致）。

---

## 模板语法 / Template Syntax

### 变量 / Variables

```html
{{ name }}              <!-- 输出变量，HTML 转义 -->
{{{ name }}}            <!-- 输出变量，原始（不转义）-->
{{ user.name }}         <!-- 点号查找（展平 key，如 'user.name'）-->
```

### 条件 / Conditionals

```html
{{#if logged_in}}
  <p>Welcome, {{ user.name }}!</p>
{{else}}
  <p>Please log in.</p>
{{/if}}
```

truthy 判定：非空、非 `'0'`、非 `'false'`。

### 循环 / Loops

```html
{{#each tags}}
  <li>{{ . }}</li>              <!-- . = 当前标量项 -->
{{/each}}

{{#each users}}
  <li>{{ @index }}: {{ name }}</li>  <!-- @index 从 0 起；点号字段查当前项 -->
{{/each}}
```

`{{#if}}`/`{{#each}}` 可嵌套；each 内嵌 if 使用当前项字段。

### 注释 / Comments

```html
{{! 会在渲染时被剔除 }}
```

---

## 过滤器 / Filters（12 个）

管道语法 `{{ name | filter[:arg] }}`，可串联：

```html
{{ name | upper }}                    <!-- HELLO -->
{{ name | lower }}                    <!-- hello -->
{{ name | capitalize }}               <!-- 首字母大写 -->
{{ name | title }}                    <!-- 每个词首字母大写 -->
{{ name | trim }}                     <!-- 去两端空白 -->
{{ name | length }}                   <!-- 长度 -->
{{ html | escape }}                   <!-- HTML 转义（{{ }} 默认已转义）-->
{{ html | raw }}                      <!-- 强制不转义 -->
{{ name | default:'N/A' }}            <!-- 空值时取默认 -->
{{ bio | truncate:80 }}               <!-- 截断到 n 字符 -->
{{ text | replace:a:b }}              <!-- 替换 a→b -->
{{ text | nl2br }}                    <!-- \n → <br> -->
```

---

## 与 Web 框架集成 / Web Framework Integration

模板引擎是纯 Kylix unit（三端同源），在 handler 里渲染出 HTML 字符串后交给 `Response.Html`（Go 端）或 TResponse 的 `Html` fluent 方法（LLVM 端）：

```pascal
uses boot, template;

var eng := TTemplateEngine.Create();

// —— Go / LLVM 通用模式 ——
procedure HomePage(req: TRequest; res: TResponse);
var out: String; err: error;
begin
  eng.AddVar('title', 'Home');
  eng.AddVar('message', 'Welcome!');
  (out, err) := RenderTemplate(eng, '{{> header}}<h1>{{ title }}</h1>');
  if err = nil then
    result := res.Html(out);      // text/html; charset=utf-8
end;
```

布局/片段配合：把公共片段注册进 `AddTemplate`，页面模板用 `{{< base }}` 继承布局、用 `{{> name }}` 引用片段——见上文布局/片段两节。

完整的可运行 server（BootRun + 模板 + Redirect + 404/500 错误页）见 `examples/complete-tutorial/22_web_pages/example60_web_framework.klx`；web 框架 API 详见 [WEB_FRAMEWORK.md](WEB_FRAMEWORK.md)。

---

## 模板示例 / Template Examples

### 页面模板（布局继承 + 片段）
```html
{{< layout base}}
<div class="hero">
  <h1>{{ title }}</h1>
  <p>{{ message }}</p>
</div>

<div class="features">
  {{#each features}}
  <div class="feature">
    <h3>{{ name }}</h3>
    <p>{{ description }}</p>
  </div>
  {{/each}}
</div>
```

### 列表页（空列表用 if 判断）
```html
<h1>{{ title }} ({{ count }})</h1>
<table>
  <tbody>
  {{#if has_users}}
    {{#each users}}
    <tr><td>{{ name }}</td><td>{{ email }}</td></tr>
    {{/each}}
  {{else}}
    <tr><td>No users found.</td></tr>
  {{/if}}
  </tbody>
</table>
```

```pascal
// 空列表条件用显式变量：
eng.AddVar('count', IntToStr(n));
if n > 0 then
  eng.AddVar('has_users', '1')
else
  eng.AddVar('has_users', '');
```

> `<name>#len` 是内部约定：`AddListLen`/`SetContext` map 里携带 `users#len` key 时，`{{#each users}}` 可直接探测长度而无需逐项构建。它不被 `{{ }}`/`{{#if}}` 插值解析。

### 个人页（过滤器）
```html
<div class="profile">
  <h1>{{ user.name | capitalize }}</h1>
  <p>Email: {{ user.email }}</p>
  <p>Bio: {{ user.bio | default:'N/A' | truncate:120 }}</p>
  {{#if user.is_admin}}
    <div class="admin-badge">Administrator</div>
  {{/if}}
</div>
```

完整可运行示例：
- `examples/complete-tutorial/22_web_pages/example59_template.klx`（基础语法 + 过滤器）
- `examples/complete-tutorial/25_template_layout/example63_template_layout.klx`（布局 + 片段，v0.9.0）

---

---

## API 参考 / API Reference

> `stdlib/template_engine.klx` 纯 Kylix 实现，Go / LLVM / bootstrap 三端同源。
> Pure-Kylix unit — identical behavior on Go, LLVM and bootstrap backends.

### TTemplateEngine（类 / class）
- `Create()`: 构造引擎 / construct engine
- `AddVar(name, value)`: 添加字符串变量 / add string variable
- `AddInt(name, value)`: 添加整型变量 / add integer variable
- `AddVariant(name, value)`: 添加 Variant 变量 / add Variant variable
- `SetContext(m)`: 批量设置变量 map / bulk-set variables from a map
- `AddListLen(name, count)`: 声明列表长度（配合 `{{#each}}` 驱动） / declare list length for `{{#each}}`
- `BeginList(name)` / `AddItem(value)` / `BeginItem` / `ItemField(key, value)` / `NextItem` / `EndList`: 构建列表上下文 / build list context
- `ListLen(name)`: 查询列表长度 / query list length
- `AddTemplate(name, src)`: 注册命名模板（布局/片段）/ register named template (layout or partial)
- `HasTemplate(name)`: 查询是否已注册 / check registration
- `RenderString(tpl)`: 渲染（出错时返回空串，`ErrorMsg()` 取消息）/ render; on error returns empty string, see `ErrorMsg()`
- `ErrorMsg()`: 最近一次渲染的错误消息 / last render error message

### 模块级函数 / module-level functions
- `RenderTemplate(eng, tpl): (String, error)`: 渲染并返回 (输出, 错误) —— 推荐入口 / render returning (output, error) — recommended entry point
- `ErrorStr(err)`: 取错误消息 / extract error message（stdlib 内建）

> 与 Web 框架集成见上文「与 Web 框架集成 / Web Framework Integration」章节（Go 端 `Response.Html` / LLVM 端 TResponse handle）。
