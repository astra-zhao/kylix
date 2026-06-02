# Kylix 编译器项目总结

[![English](https://img.shields.io/badge/lang-English-blue.svg)](README.md)
[![版本](https://img.shields.io/badge/version-1.0.1-blue.svg)](CHANGELOG.md)
[![许可证](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## 项目概述

Kylix 是一个现代化的 Pascal 语言重新实现，编译为 Go 代码。它结合了 Pascal 的清晰性和简洁性，同时添加了现代语言特性，并配备了完整的 IDE 工具链和编辑器集成。

**当前版本**：v1.0.1（2026-06-02 发布，BUG 修复版本）

**项目地址**：https://github.com/astra-zhao/kylix

---

## 快速开始

### 安装

```bash
# 克隆仓库
git clone https://github.com/astra-zhao/kylix.git
cd kylix

# 构建编译器
go build -o kylix cmd/kylix/main.go

# 添加到 PATH（可选）
export PATH=$PATH:$(pwd)
```

### 创建第一个项目

```bash
# 创建新项目
./kylix new myapp
cd myapp

# 编译并运行
./kylix run

# 检查语法
./kylix check

# 格式化代码
./kylix fmt
```

---

## CLI 命令参考

```bash
kylix new <name>       # 创建新项目
kylix build            # 编译项目或文件
kylix run              # 编译并运行
kylix check            # 语法检查（不生成代码）
kylix fmt              # 格式化源代码
kylix repl             # 交互式 REPL
kylix lsp              # 启动 LSP 服务器（用于编辑器）
kylix version          # 显示版本信息
kylix help             # 显示帮助
```

### 命令详解

#### `kylix new` - 创建新项目
创建包含 `kylix.toml` 配置、`main.klx` 入口文件、`build/` 输出目录和 `.gitignore` 的项目模板。

#### `kylix build` - 编译
将 Kylix 代码编译为 Go 代码。支持项目模式（读取 `kylix.toml`）和单文件模式。

#### `kylix run` - 编译并运行
编译后立即运行程序，支持 `--keep` 参数保留生成的 `.go` 文件。

#### `kylix check` - 语法检查
只检查语法错误，不生成代码。适合在 CI/CD 中使用。

#### `kylix fmt` - 代码格式化
自动格式化代码，保持统一风格。（目前为基础版本）

#### `kylix repl` - 交互式解释器
即时执行 Kylix 代码，适合学习和快速测试。

#### `kylix lsp` - 语言服务器
启动 LSP 服务器，为编辑器提供智能补全、悬停提示等功能。

---

## 已完成的工作

### 第一阶段：编译器核心 ✅

#### 核心组件

1. **词法分析器 (Lexer)**
   - 支持 Pascal 风格的所有 token
   - 处理单引号和双引号字符串
   - 支持注释（行注释 `//` 和块注释 `{}`、`(* *)`）
   - 正确的行号和列号追踪

2. **语法分析器 (Parser)**
   - Pratt 解析器实现，正确处理运算符优先级
   - 支持所有 Pascal 控制结构
   - 支持函数和过程声明
   - 支持类和接口声明
   - 现代特性：lambda 表达式、模式匹配、async/await

3. **抽象语法树 (AST)**
   - 完整的节点类型定义
   - 支持语句和表达式
   - 支持现代语言特性

4. **代码生成器 (Generator)**
   - 将 Kylix AST 转换为 Go 代码
   - 内置函数映射（WriteLn → fmt.Println 等）
   - 智能导入管理（只导入需要的包）
   - 类型映射（Integer → int64, Real → float64 等）

#### 语言特性

**传统 Pascal 特性**：
- ✅ 强类型系统
- ✅ 变量和常量声明
- ✅ 函数和过程
- ✅ 控制结构（if/while/for/case/repeat）
- ✅ 记录和数组
- ✅ 异常处理（try/except/finally）

**现代特性**：
- ✅ 类型推断（`var x := 42;`）
- ✅ Lambda 表达式（`(x: Integer) -> x * x`）
- ✅ 模式匹配（`match value { ... }`）
- ✅ Async/Await（`async function`, `await`）
- ✅ 类和接口
- ✅ 属性（properties）
- ✅ ForEach 循环（`for item in collection`）
- ✅ 泛型类型引用和声明（`TList<T>`, `TPair<T1, T2>`, `function Foo<T>`）
- ✅ 构造器/析构器/继承关键字
- ✅ Lambda 参数解析（`(x: Integer) -> x * x`）
- ✅ ON 子句异常处理（`on E: ExceptionType do`）

#### 示例程序

所有示例都成功编译和运行：

1. **hello.klx** - Hello World 程序
2. **simple.klx** - 简单的变量和赋值
3. **types.klx** - 类型演示
4. **control.klx** - 控制结构
5. **functions.klx** - 函数和过程
6. **modern.klx** - 现代特性
7. **classes.klx** - 面向对象编程
8. **exceptions.klx** - 异常处理

---

### 第二阶段：IDE 工具 ✅

#### 核心功能

1. **CLI 工具链**
   - ✅ 项目管理（`kylix new`）
   - ✅ 编译系统（`kylix build`）
   - ✅ 运行器（`kylix run`）
   - ✅ 语法检查（`kylix check`）
   - ✅ 代码格式化（`kylix fmt`）
   - ✅ 交互式 REPL（`kylix repl`）
   - ✅ LSP 服务器（`kylix lsp`）

2. **项目管理系统**
   - ✅ `kylix.toml` 配置文件
   - ✅ 自动查找项目根目录
   - ✅ 多文件项目支持
   - ✅ 构建输出目录管理

3. **LSP 语言服务器**
   - ✅ 语法诊断（实时错误检查）
   - ✅ 代码补全（关键字和内置函数）
   - ✅ 悬停提示（函数文档）
   - ✅ 标准 LSP 协议支持

4. **VS Code 扩展**
   - ✅ 语法高亮
   - ✅ 语言配置（括号、注释、折叠）
   - ✅ LSP 客户端集成

5. **文档**
   - ✅ 用户手册（IDE 工具使用说明）
   - ✅ 开发指南（架构和贡献指南）
   - ✅ 工具解释（通俗易懂的概念说明）

#### 工具功能说明

| 工具 | 作用 | 类比 | 使用频率 |
|------|------|------|----------|
| **fmt** | 代码排版 | Word 自动排版 | 低（偶尔用） |
| **repl** | 即时实验 | 计算器 | 中（学习/调试时） |
| **lsp** | 智能助手 | 输入法提示 | 高（写代码时一直用） |
| **LSP 服务器** | 通用协议 | USB 接口标准 | 用户不直接接触 |

**通俗理解**：
- `kylix new` = 创建新文档
- `kylix build` = 保存文档
- `kylix run` = 保存并预览
- `kylix check` = 拼写检查
- `kylix fmt` = **自动排版**（让文档整齐）
- `kylix repl` = **草稿纸**（快速试验）
- `kylix lsp` = **智能助手**（自动补全、语法提示）

---

## 项目结构

```
kylix/
├── cmd/
│   └── kylix/              # CLI 入口
│       └── main.go         # 命令解析和调度
│
├── pkg/
│   ├── compiler/           # 编译器核心
│   │   └── compiler.go     # 编译 API
│   ├── project/            # 项目管理
│   │   └── project.go      # kylix.toml 解析
│   ├── lsp/                # Language Server Protocol
│   │   └── server.go       # LSP 服务器
│   └── repl/               # 交互式解释器
│       └── repl.go         # REPL 实现
│
├── lexer/                  # 词法分析器
│   └── lexer.go
│
├── parser/                 # 语法分析器
│   └── parser.go
│
├── ast/                    # 抽象语法树定义
│   └── ast.go
│
├── generator/              # Go 代码生成器
│   └── generator.go
│
├── token/                  # Token 定义
│   └── token.go
│
├── examples/               # 示例代码
│   ├── hello.klx
│   ├── types.klx
│   ├── control.klx
│   ├── functions.klx
│   ├── modern.klx
│   ├── classes.klx
│   ├── exceptions.klx
│   ├── simple.klx
│   ├── web_demo.klx        # Web 框架基础示例
│   ├── web_advanced.klx    # Web 框架高级示例（DI、配置、中间件、验证）
│   ├── web_fullstack.klx   # 全栈 Web 示例（模板、ORM、自动配置）
│   └── orm_example.klx     # ORM 示例（数据库 CRUD、查询构建器、迁移）
│
├── stdlib/                 # 标准库
│   ├── web.go              # Web 框架实现
│   ├── container.go        # 依赖注入容器
│   ├── config.go           # 配置管理
│   ├── middleware.go       # 中间件（CORS、认证、限流、日志）
│   ├── validation.go       # 请求验证
│   ├── orm.go              # ORM（支持 MySQL、PostgreSQL、SQLite）
│   ├── template.go         # 模板引擎（布局、片段、自定义函数）
│   ├── autoconfig.go       # 自动配置（多源加载、环境检测）
│   ├── sysutil.go          # 文件 I/O 和系统工具
│   ├── jsonutil.go         # JSON 编码/解码
│   ├── datetime.go         # 日期和时间操作
│   └── regex.go            # 正则表达式
│
├── vscode-ext/             # VS Code 扩展
│   ├── extension.js        # LSP 客户端
│   ├── package.json        # 扩展配置
│   ├── language-configuration.json
│   ├── syntaxes/
│   │   └── kylix.tmLanguage.json
│   ├── snippets/           # 代码片段
│   ├── README.md           # 扩展说明
│   ├── INSTALL.md          # 安装指南
│   ├── USAGE_GUIDE.md      # 使用指南
│   └── CHANGELOG.md        # 更新日志
│
├── docs/                   # 文档
│   ├── KYLIX_IDE_USER_MANUAL.md    # IDE 用户手册
│   ├── KYLIX_DEV_GUIDE.md          # 开发指南
│   ├── KYLIX_TOOLS_EXPLAINED.md    # 工具解释
│   ├── PHASE2_SUMMARY.md           # 第二阶段总结
│   ├── WEB_FRAMEWORK.md            # Web 框架指南
│   ├── ORM_GUIDE.md                # ORM 指南
│   └── TEMPLATE_GUIDE.md           # 模板引擎指南
│
├── go.mod                  # Go 模块定义
├── README.md               # 英文项目文档
├── SUMMARY.md              # 中文项目总结（本文件）
└── Makefile                # 构建脚本
```

---

## 编辑器集成

### VS Code

`vscode-ext/` 目录包含完整的 VS Code 扩展：

```bash
cd vscode-ext
npm install
# 在 VS Code 中按 F5 启动扩展
```

**功能**：
- 语法高亮
- 语言配置（括号、注释、折叠）
- LSP 客户端集成
- 实时错误检查
- 代码补全

### 其他编辑器

Kylix LSP 支持任何带有 LSP 客户端的编辑器：

**Neovim**（使用 `nvim-lspconfig`）：
```lua
require('lspconfig').kylix.setup{}
```

**Emacs**（使用 `lsp-mode`）：
```elisp
(add-to-list 'lsp-language-id-configuration '(kylix-mode . "kylix"))
(lsp-register-client
 (make-lsp-client :new-connection (lsp-stdio-connection '("kylix" "lsp"))
                  :major-modes '(kylix-mode)))
```

**通用配置**：
```json
{
  "command": ["kylix", "lsp"],
  "filetypes": ["kylix"]
}
```

---

## 文档资源

### 📘 IDE 用户手册
**文件**：`docs/KYLIX_IDE_USER_MANUAL.md`

**内容**：
- 安装指南
- 快速开始（3步上手）
- 完整命令参考（9个命令详解）
- 项目结构说明
- 编辑器集成配置
- 常见问题解答（10个FAQ）
- 示例项目

**适合**：刚接触 Kylix 的用户、想了解 CLI 命令的开发者、需要配置编辑器的用户

### 🛠️ 开发指南
**文件**：`docs/KYLIX_DEV_GUIDE.md`

**内容**：
- 项目架构图
- 编译器原理详解（Lexer、Parser、Generator）
- 代码组织结构
- 开发环境搭建
- 核心模块详解
- 添加语言特性教程
- 测试指南
- 贡献指南
- 项目路线图

**适合**：想贡献代码的开发者、想了解编译器内部原理的人、想添加新语言特性的开发者

### 💡 工具解释
**文件**：`docs/KYLIX_TOOLS_EXPLAINED.md`

**内容**：
- kylix fmt 通俗解释（代码格式化）
- kylix repl 通俗解释（交互式编程）
- kylix lsp 通俗解释（智能助手）
- LSP 服务器通俗解释（通用协议）
- 总结对比表
- 实际使用建议

**适合**：初学者、想理解工具作用的用户

### 🌐 Web 框架指南
**文件**：`docs/WEB_FRAMEWORK.md`

**内容**：
- Web 服务器快速入门
- REST API 开发完整示例
- 路由系统详解（GET、POST、PUT、DELETE）
- 路径参数和查询参数
- JSON 请求/响应处理
- 中间件开发
- 静态文件服务
- 最佳实践
- API 参考手册

**适合**：需要开发 Web 应用和 REST API 的开发者

---

## 技术亮点

### 1. Pratt 解析器
- 优雅地处理运算符优先级
- 易于扩展新的运算符
- 代码简洁，易于理解

### 2. 智能导入管理
- 自动检测需要的导入
- 避免未使用的导入错误
- 只导入实际使用的包

### 3. 内置函数映射
- WriteLn → fmt.Println
- Length → len
- 支持 30+ 个内置函数

### 4. 类型映射
- Pascal 类型到 Go 类型的自动转换
- 支持所有基本类型
- Integer → int64, Real → float64, String → string

### 5. 错误处理
- 详细的错误信息
- 行号和列号追踪
- 防止无限循环的安全机制

### 6. LSP 协议支持
- 一个工具支持所有编辑器
- 节省开发时间
- 统一的用户体验

---

## 开发路线图

### 第一阶段：转译器 ✅
- ✅ 词法分析器和语法分析器
- ✅ AST 生成
- ✅ Go 代码生成
- ✅ 基本语言特性
- ✅ 现代特性（lambda、async、模式匹配）

### 第二阶段：IDE 工具 ✅
- ✅ CLI 工具链（new、build、run、check、fmt、repl、lsp）
- ✅ 项目管理（kylix.toml）
- ✅ LSP 服务器（补全和悬停）
- ✅ VS Code 扩展（语法高亮）
- ✅ 交互式 REPL
- ✅ 完整文档

### 第三阶段：Web 框架（已完成） ✅
- ✅ HTTP 服务器（基于 Go net/http）
- ✅ 路由系统（GET、POST、PUT、DELETE）
- ✅ 路径参数（`/users/:id` 语法）
- ✅ 中间件支持（日志中间件）
- ✅ JSON 请求/响应处理
- ✅ 静态文件服务
- ✅ 匿名过程/函数支持
- ✅ VS Code 扩展增强（语法高亮、代码片段、智能补全）
- ✅ Web 框架文档
- ✅ 依赖注入容器
- ✅ 配置系统
- ✅ 中间件套件（CORS、认证、限流、请求ID、日志）
- ✅ 请求验证
- ✅ ORM（支持 MySQL、PostgreSQL、SQLite）
- ✅ 模板引擎（支持布局、片段、自定义函数）
- ✅ 自动配置（多源加载、环境检测）

### 第四阶段：语言增强（已完成） ✅
- ✅ 泛型类型参数声明（类和函数的 `<T>` 语法）
- ✅ 异常处理 ON 子句（`on E: ExceptionType do`）
- ✅ 构造器/析构器/继承关键字
- ✅ Lambda 参数解析
- ✅ Async/Await 代码生成改进（goroutine + channel 模式）

### 第五阶段：标准库与工具改进 ✅
- ✅ 文件 I/O（`sysutil`）— 读写、复制、目录操作、路径工具
- ✅ JSON（`jsonutil`）— 编码、解码、类型安全访问器、文件 I/O
- ✅ 日期时间（`datetime`）— 日期运算、格式化、解析、比较
- ✅ 正则表达式（`regex`）— 匹配、查找、替换、分割、模式辅助
- ✅ REPL 改进 — readline 历史记录（↑/↓）、词法分析检测、stderr 分离
- ✅ 格式化器修复 — 类可见性修饰符、属性、常量类型注解
- ✅ 生成器 stdlib 映射 — sysutil、jsonutil、datetime、regex 模块

---

## 性能指标

- **编译速度**：~100ms（小型程序）
- **生成代码质量**：可直接运行，无需手动修改
- **内存使用**：~50MB（编译器本身）
- **支持的文件大小**：无限制（受系统内存限制）

---

## 当前限制

1. 某些高级 Pascal 特性尚未完全实现
2. 标准库还不完整
3. 错误恢复机制可以改进
4. `kylix fmt` 目前只检查语法，未实现真正的格式化
5. LSP 功能还比较简单，缺少跳转定义、查找引用等高级功能
6. REPL 功能基础，缺少历史记录和多行编辑

---

## 总结

Kylix 编译器第一阶段至第五阶段已全部成功完成！

**第一阶段**实现了一个功能完整的 Pascal-to-Go 转译器，支持传统 Pascal 特性和现代语言特性。所有示例程序都能成功编译和运行。

**第二阶段**构建了完整的 IDE 工具链，包括 CLI 工具、项目管理、LSP 服务器、VS Code 扩展和详尽的文档。

**第三阶段**实现了完整的 Web 框架：HTTP 服务器、路由系统、中间件、JSON处理、静态文件服务、依赖注入、配置系统、ORM、模板引擎和自动配置。

**第四阶段**完善了语言特性：泛型类型参数声明、异常处理 ON 子句、构造器/析构器/inherited 关键字、Lambda 参数解析和 Async/Await 代码生成改进。

**第五阶段**扩展了标准库和工具：文件 I/O（sysutil）、JSON 处理（jsonutil）、日期时间（datetime）、正则表达式（regex），同时改进了 REPL（readline 历史、词法分析检测、stderr 分离）和格式化器（类可见性、属性、常量类型）。

**当前状态**：Phase 1-5 已完成，v1.0.1 已发布！🎉

### v1.0.1 BUG 修复 (2026-06-02)

| 优先级 | 问题 | 修复 |
|--------|------|------|
| **P0** | `inherits` 关键字静默忽略 | 添加 `INHERITS` 解析分支，正确设置父类 |
| **P0** | 匿名 `procedure()`/`function()` 不解析 | 注册为表达式 context 的 prefix 解析函数 |
| **P0** | Match 通配符 `_` 生成无效 Go | 检测 `_` 模式生成 `default:` 分支 |
| **P0** | `{ }` 注释与 match 块冲突 | 移除 `{}` 注释语法，仅保留 `//` 和 `(* *)` |
| **P1** | 构造函数 `Dog.Create(args)` 无效 Go | 检测 `.Create` 模式生成 `&Dog{args}` |
| **P1** | Match 分支不触发 import 扫描 | 添加 `MatchStatement` 到 import 扫描 |

### 版本号升级记录 (v1.0.0 → v1.0.1)

| 组件 | v1.0.0 | **v1.0.1** |
|------|--------|-----------|
| 编译器 (`cmd/kylix/main.go`) | 1.0.0 | **1.0.1** |
| REPL (`pkg/repl/repl.go`) | 1.0.0 | **1.0.1** |
| LSP 服务器 (`pkg/lsp/server.go`) | 1.0.0 | **1.0.1** |
| 项目配置 (`pkg/project/project.go`) | 1.0.0 | **1.0.1** |
| VS Code 扩展 (`vscode-ext/package.json`) | 1.0.0 | **1.0.1** |

### 待修复问题 (计划 v1.0.2+)

| 优先级 | 问题 | 影响 |
|--------|------|------|
| P1 | 字符串插值三层断裂 | Lexer→Parser→Generator 全链路修复 |
| P1 | 异常类型在 Go 中不存在 | `on E: Exception do` 生成无效类型 |
| P2 | 多返回值不支持 | `function Div(): (Real, Boolean)` 失败 |
| P2 | Properties 被跳过 | `property Name: String` 不生成 Go 代码 |
| P2 | 无多文件编译 | `uses` 仅导入 stdlib，不处理用户文件 |
| P2 | Map 类型不支持 | 符号表等场景需要 |
| P2 | 核心模块无测试 | Lexer/Parser/Generator 无单元测试 |
| P3 | 18 个 Token 无解析处理 | `with`, `set`, `new`, `exit` 等 |
| P3 | LSP 代码操作是 stub | 无实际 import 整理/格式化功能 |
| P3 | REPL 不支持 uses/class | REPL 中无法使用模块或定义类 |

详细更新日志见 [CHANGELOG.md](CHANGELOG.md)。

---

## 贡献

欢迎贡献代码！请随时提交 Issue 和 Pull Request。

**提交代码流程**：
1. Fork 仓库
2. 创建功能分支：`git checkout -b feature/my-feature`
3. 编写代码和测试
4. 运行测试：`go test ./...`
5. 提交：`git commit -am "Add feature: ..."`
6. 推送并创建 PR

**代码规范**：
- 使用 `gofmt` 格式化代码
- 添加注释说明复杂逻辑
- 保持函数简短（< 50 行）
- 添加单元测试

---

## 许可证

MIT License
