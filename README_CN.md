# Kylix - 现代化 Pascal 语言

[![English](https://img.shields.io/badge/lang-English-blue.svg)](README.md)
[![Official Site](https://img.shields.io/badge/official-kylix.top-4f6ef7.svg)](https://kylix.top)
[![版本](https://img.shields.io/badge/version-1.1.5-blue.svg)](CHANGELOG.md)
[![许可证](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![自举进度](https://img.shields.io/badge/self--hosting-90%25-brightgreen.svg)](ROADMAP.md)

Kylix 是 Pascal 语言的现代化重制版，编译目标为 Go。它将 Pascal 的清晰简洁与现代化语言特性相结合，并配备完整的 IDE 工具链和编辑器集成。

> 🌐 **官方网站**: [https://kylix.top](https://kylix.top) — 交互式文档、在线示例和完整功能展示。
>
> 🔥 **重大里程碑 (v1.1.5)**: 自举编译器生成的多文件 Go 代码（136KB）编译零错误通过——Kylix 编译器可以编译自身！详见 [ROADMAP.md](ROADMAP.md)。

## 特性

### Pascal 核心特性
- 强类型与类型推导
- 过程和函数
- 控制结构（if, while, for, case, repeat）
- 记录和数组
- 异常处理

### 现代语言特性
- **类型推导**: `var x := 42;`
- **Lambda 表达式**: `var square = (x: Integer) -> x * x;`
- **泛型**: 声明类型参数: `TList<T>`, `function Foo<T>(x: T): T`
- **泛型类型引用**: `TList<Integer>`, `TPair<String, Integer>`
- **Async/Await**: `async function FetchData(): String;`
- **模式匹配**: `match value { 0 => 'zero', _ => 'other' }`
- **类和接口**: 面向对象编程支持
- **属性**: 带 getter/setter
- **ForEach 循环**: `for item in collection do`
- **字符串插值**: `'Hello, ${name}!'`
- **现代化异常处理**: try/except/finally, `on E: Type do` 子句

## 安装

```bash
# 克隆仓库
git clone https://github.com/astra-zhao/kylix.git
cd kylix

# 编译
go build -o kylix cmd/kylix/main.go

# 添加到 PATH（可选）
export PATH=$PATH:$(pwd)
```

## 快速上手

```bash
# 创建新项目
./kylix new myapp
cd myapp

# 编译并运行
./kylix run

# 语法检查
./kylix check

# 格式化代码
./kylix fmt
```

## CLI 命令

```bash
kylix new <name>       # 创建新项目
kylix build            # 编译项目或文件
kylix run              # 编译并运行
kylix check            # 语法检查（不生成代码）
kylix fmt              # 格式化源代码
kylix repl             # 交互式 REPL
kylix lsp              # 启动 LSP 服务器（用于编辑器）
kylix version          # 显示版本
kylix help             # 显示帮助
```

## 示例

### Hello World
```pascal
program Hello;
begin
  WriteLn('Hello, Kylix World!');
end.
```

### 类型推导
```pascal
var count := 42;
var message := '推断为字符串';
var ratio := 3.14;
```

### 函数
```pascal
function Add(a: Integer; b: Integer): Integer;
begin
  result := a + b;
end;

// Lambda
var square := (x: Integer) -> x * x;
```

### 类
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
    WriteLn(Name, ' 发出声音');
  end;
end;

class Dog inherits Animal
public
  procedure Speak; override;
  begin
    WriteLn(Name, ' 汪汪叫！');
  end;
end;
```

### 模式匹配
```pascal
match value {
  0 => 'zero',
  1, 2, 3 => 'small',
  when value > 100 => 'large',
  _ => 'other'
};
```

### 泛型类和函数
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
```

### 异常处理
```pascal
try
  raise Exception.Create('test');
except
  on E: Exception do
    WriteLn('捕获: ' + E.Message);
  else
    WriteLn('未知异常');
end;
```

## 语言参考

### 类型
- `Integer` — 64位整数（映射到 Go 的 `int64`）
- `Real` — 64位浮点数（映射到 Go 的 `float64`）
- `Boolean` — 布尔值
- `String` — 字符串
- `Char` — 单字符（映射到 Go 的 `byte`）

### 运算符
- 算术: `+`, `-`, `*`, `/`, `div`, `mod`
- 比较: `=`, `<>`, `<`, `>`, `<=`, `>=`
- 逻辑: `and`, `or`, `not`, `xor`
- 赋值: `:=`, `=`

### 控制结构
- `if/then/else`
- `while/do`
- `for/to/downto`
- `for/in`（foreach）
- `repeat/until`
- `case/of`
- `match`（模式匹配）
- `try/except/finally`

## 项目结构

```
kylix/
├── cmd/kylix/          # CLI 入口
├── pkg/
│   ├── compiler/       # 编译 API
│   ├── project/        # 项目管理 (kylix.toml)
│   ├── lsp/            # Language Server Protocol 服务器
│   └── repl/           # 交互式 REPL
├── stdlib/             # 标准库
│   ├── web.go          # Web 框架
│   ├── container.go    # 依赖注入
│   ├── config.go       # 配置管理
│   ├── middleware.go   # 中间件
│   ├── validation.go   # 请求验证
│   ├── orm.go          # ORM
│   ├── template.go     # 模板引擎
│   ├── autoconfig.go   # 自动配置
│   ├── sysutil.go      # 文件 I/O 和系统工具
│   ├── jsonutil.go     # JSON 编解码
│   ├── datetime.go     # 日期时间操作
│   └── regex.go        # 正则表达式
├── token/              # Token 定义
├── lexer/              # 词法分析器
├── ast/                # 抽象语法树
├── parser/             # 解析器 (Pratt 解析)
├── generator/          # Go 代码生成器
├── src/                # 自举编译器源码 (.klx 文件)
│   ├── token.klx       # Token 类型 (209 行)
│   ├── ast.klx         # AST 节点层次, 54 个类 (374 行)
│   ├── lexer.klx       # 词法分析器 (366 行)
│   ├── parser.klx      # Pratt 解析器 (2338 行)
│   ├── error.klx       # 错误/诊断类型 (91 行)
│   ├── generator.klx   # Go 代码生成器 (~1500 行)
│   └── main.klx        # 入口点, 支持多文件编译
├── examples/           # 示例程序
├── vscode-ext/         # VS Code 扩展
└── docs/               # 文档
```

## 标准库

### Web 框架 (`web`)
HTTP 服务器，支持路由、中间件和请求/响应处理。

### 依赖注入 (`container`)
IoC 容器，支持 singleton、transient 和 scoped 生命周期。

### 配置 (`config`)
从环境变量加载配置，带类型安全访问器。

### 中间件 (`middleware`)
预置中间件：CORS、Auth、Rate Limit、Request ID、Logging。

### 验证 (`validation`)
请求验证，支持 fluent API 和常用验证器。

### ORM (`orm`)
数据库 ORM，支持 MySQL、PostgreSQL、SQLite。

### 模板引擎 (`template`)
HTML 模板渲染，支持布局、局部模板和自定义函数。

### 文件 I/O (`sysutil`)
文件和目录操作，Pascal 风格 API。

### JSON (`jsonutil`)
JSON 编解码和操作。

### 日期时间 (`datetime`)
日期时间操作，支持算术运算和格式化。

### 正则表达式 (`regex`)
模式匹配、搜索和文本操作。

## 编辑器集成

### VS Code
`vscode-ext/` 目录包含完整的 VS Code 扩展：
- 语法高亮
- 语言配置（括号、注释、折叠）
- LSP 客户端集成

### 其他编辑器
Kylix LSP 支持任何带 LSP 客户端的编辑器：
```json
{
  "command": ["kylix", "lsp"],
  "filetypes": ["kylix"]
}
```

## 路线图

### 第一阶段：编译器核心 ✅
词法分析器、解析器、AST、Go 代码生成、基本语言特性、现代特性

### 第二阶段：IDE 工具 ✅
CLI、项目管理、LSP、VS Code 扩展、REPL、文档

### 第三阶段：Web 框架 ✅
HTTP 服务器、路由、中间件、JSON、DI、配置、ORM、模板引擎

### 第四阶段：语言增强 ✅
泛型、异常 ON 子句、构造器/析构器、Lambda、Async/Await

### 第五阶段：标准库与工具 ✅
sysutil、jsonutil、datetime、regex、REPL 改进、格式化器

### 第六-七阶段：Bug 修复与语言能力 ✅
字符串插值、异常类型、多返回值、Properties、Map、Variant、动态数组、枚举、多文件模块

### 第八阶段：自举编译器 ✅
7 个 Kylix 源文件、类代码生成、软关键字、is/as 类型分发

### 第九阶段：自举验证 🚧 90%
- ✅ 多文件自举联编
- ✅ 自举 Go 输出编译零错误
- 🟡 Diff 验证进行中

## 更新日志

### v1.1.5 (2026-06-08) — 多文件 Go 编译通过 🎉

自举多文件 Go 输出（136KB）编译零错误，运行正常。主要修复：字符串转义、基类类型映射、枚举类型声明、内置函数、多参数解析。

### v1.1.0-v1.1.4 (2026-06-06~08) — 自举引导

Lexer bug 修复、generator 骨架完善、parser result 覆盖修复、代码生成改进（Record 类型、Map 初始化、局部变量）、软关键字扩展、多文件自举、类方法 receiver 生成。

详见 [CHANGELOG.md](CHANGELOG.md)。

## 文档

- [IDE 用户手册](docs/KYLIX_IDE_USER_MANUAL.md)
- [开发者指南](docs/KYLIX_DEV_GUIDE.md)
- [工具详解](docs/KYLIX_TOOLS_EXPLAINED.md)
- [Web 框架指南](docs/WEB_FRAMEWORK.md)
- [ORM 指南](docs/ORM_GUIDE.md)
- [模板引擎指南](docs/TEMPLATE_GUIDE.md)
- [项目总结 (中文)](SUMMARY.md)

## 贡献

欢迎贡献代码！请提交 Issue 和 Pull Request。

## 许可证

MIT License
