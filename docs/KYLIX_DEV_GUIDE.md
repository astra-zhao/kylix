# Kylix 开发指南

本指南面向希望了解 Kylix 内部实现、贡献代码或扩展语言功能的开发者。

## 目录

- [项目架构](#项目架构)
- [编译器工作流程](#编译器工作流程)
- [代码组织](#代码组织)
- [开发环境](#开发环境)
- [核心模块详解](#核心模块详解)
- [添加语言特性](#添加语言特性)
- [测试](#测试)
- [贡献指南](#贡献指南)
- [路线图](#路线图)

> 📌 **自举编译器**：开发自举功能（`src/*.klx` / 多态 codegen / self-reproduction）前请先读 [SELFHOSTING_DEV_GUIDE.md](SELFHOSTING_DEV_GUIDE.md) —— 自举源码结构、opt-in 多态基类→interface 规则、构建流程、v5.3+ round-trip / v5.7 self-reproduction 不动点的权威参考。

---

## 项目架构

Kylix 是一个 **双后端编译器**：默认把 Kylix 代码转译为 Go 代码（Go 工具链执行），也可经 **LLVM 后端** 直接生成原生二进制（`--backend=llvm`，无 Go 依赖）。编译器自身源码（`src/*.klx`）可用 LLVM 后端自举，且已达成无 Go 闭环（v0.6.9）。

```
┌─────────────┐
│  Kylix 源码  │  (.klx 文件)
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Lexer     │  词法分析：源代码 → Token 流
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Parser    │  语法分析：Token 流 → AST
└──────┬──────┘
       │
       ▼
┌─────────────┐
│     AST     │  抽象语法树
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ Generator   │  代码生成：AST → Go 代码
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Go 源码    │  (.go 文件)
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Go 工具链   │  编译/运行 Go 代码
└─────────────┘
```

LLVM 后端（`pkg/llvmgen/`）从同一棵 AST 分叉：

```
┌─────────────┐
│     AST     │
└──────┬──────┘
       │  --backend=llvm
       ▼
┌─────────────┐
│ LLVMGenerator │  代码生成：AST → LLVM IR (.ll)
└──────┬──────┘
       │
       ▼
┌──────────────────────┐
│ opt / llc / clang    │  优化 → 目标文件 → 原生二进制
└──────────────────────┘
```

### 设计原则

1. **简单优先**：保持编译器小巧、可理解
2. **渐进增强**：先实现核心功能，再逐步添加高级特性
3. **错误友好**：提供清晰的错误信息和位置
4. **可测试性**：每个模块都有单元测试

---

## 编译器工作流程

### 1. 词法分析 (Lexer)

**位置**: `lexer/lexer.go`

将源代码字符串转换为 Token 序列。

**输入**:
```pascal
var x := 42;
```

**输出**:
```
Token{Type: VAR, Literal: "var", Line: 1, Column: 1}
Token{Type: IDENT, Literal: "x", Line: 1, Column: 5}
Token{Type: ASSIGN_OP, Literal: ":=", Line: 1, Column: 7}
Token{Type: INT, Literal: "42", Line: 1, Column: 10}
Token{Type: SEMICOLON, Literal: ";", Line: 1, Column: 12}
```

**关键数据结构**:
```go
type Token struct {
    Type    TokenType  // Token 类型（VAR, IDENT, INT 等）
    Literal string     // 原始文本
    Line    int        // 行号
    Column  int        // 列号
}
```

### 2. 语法分析 (Parser)

**位置**: `parser/parser.go`

使用 **Pratt 解析算法** 将 Token 流转换为 AST。

**输入**: Token 序列

**输出**:
```go
&ast.VarDecl{
    Names: []string{"x"},
    Value: &ast.IntegerLiteral{Value: 42},
    Inferred: true,  // 使用 :=
}
```

**Pratt 解析的核心思想**:
- 每个 Token 类型都有对应的 **前缀解析函数** 或 **中缀解析函数**
- 通过优先级表处理运算符优先级
- 递归下降构建 AST

**关键数据结构**:
```go
type Parser struct {
    l              *lexer.Lexer
    curToken       token.Token
    peekToken      token.Token
    prefixParseFns map[token.TokenType]prefixParseFn
    infixParseFns  map[token.TokenType]infixParseFn
}
```

### 3. 代码生成 (Generator)

**位置**: `generator/generator.go`

遍历 AST，生成等效的 Go 代码。

**输入**: AST

**输出**:
```go
var x = 42
```

**关键策略**:
- **类型映射**: Kylix 类型 → Go 类型（Integer → int64, String → string）
- **内置函数映射**: WriteLn → fmt.Println, Length → len
- **智能导入**: 只导入实际使用的包
- **函数返回值**: 自动处理 `result` 变量

---

## 代码组织

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
├── pkg/llvmgen/            # LLVM 原生后端（IR 生成 + stdlib IR + Variant + DWARF）
│   ├── codegen.go
│   ├── compile.go
│   └── stdlib_*.go
│
├── src/                    # 自举编译器源码（.klx，9 文件）
│   └── stdlib_ir.klx       # 烘焙 stdlib IR 数据（自动生成）
│
├── token/                  # Token 定义
│   └── token.go
│
├── examples/               # 示例代码
│   ├── hello.klx
│   ├── types.klx
│   └── ...
│
├── vscode-ext/             # VS Code 扩展
├── jetbrains-plugin/       # JetBrains 插件（TextMate + LSP4IJ）
│   ├── extension.js
│   ├── package.json
│   └── syntaxes/
│
├── docs/                   # 文档
│   ├── KYLIX_IDE_USER_MANUAL.md
│   └── KYLIX_DEV_GUIDE.md
│
└── go.mod                  # Go 模块定义
```

---

## 开发环境

### 前置要求

- Go 1.18+
- Git
- VS Code（推荐）

### 搭建开发环境

```bash
# 1. 克隆仓库
git clone <repository-url>
cd kylix

# 2. 构建编译器
go build -o kylix cmd/kylix/main.go

# 3. 运行测试
go test ./...

# 4. 安装到 PATH（可选）
export PATH=$PATH:$(pwd)
```

### VS Code 配置

创建 `.vscode/settings.json`:
```json
{
  "go.toolsManagement.autoUpdate": true,
  "editor.formatOnSave": true,
  "go.formatTool": "gofmt"
}
```

创建 `.vscode/launch.json`:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Kylix Compiler",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/cmd/kylix/main.go",
      "args": ["build", "examples/hello.klx"]
    }
  ]
}
```

---

## 核心模块详解

### Token 模块 (`token/token.go`)

定义所有 Token 类型和关键字映射。

**添加新关键字**:
```go
// 1. 定义 Token 类型
const (
    // ... 现有定义
    NEW_KEYWORD = "new_keyword"
)

// 2. 添加到关键字映射
var keywords = map[string]TokenType{
    // ... 现有关键字
    "new_keyword": NEW_KEYWORD,
}
```

### AST 模块 (`ast/ast.go`)

定义所有 AST 节点类型。

**添加新节点**:
```go
// 1. 定义节点结构
type NewStatement struct {
    // 字段
    Value Expression
}

// 2. 实现 Statement 接口
func (n *NewStatement) statementNode() {}
func (n *NewStatement) TokenLiteral() string { return "new" }
```

### Parser 模块 (`parser/parser.go`)

使用 Pratt 解析算法。

**添加新语法规则**:

```go
// 1. 注册解析函数
func New(l *lexer.Lexer) *Parser {
    p := &Parser{l: l}
    
    // 注册前缀解析函数（用于表达式开头）
    p.registerPrefix(token.NEW_TOKEN, p.parseNewExpression)
    
    // 注册中缀解析函数（用于运算符）
    p.registerInfix(token.NEW_OPERATOR, p.parseNewOperator)
    
    return p
}

// 2. 实现解析函数
func (p *Parser) parseNewExpression() ast.Expression {
    // 解析逻辑
    return &ast.NewExpression{...}
}

// 3. 如果是运算符，设置优先级
var precedences = map[token.TokenType]int{
    // ... 现有优先级
    token.NEW_OPERATOR: PRODUCT,  // 或自定义优先级
}
```

### Generator 模块 (`generator/generator.go`)

AST 遍历和 Go 代码生成。

**添加新节点生成**:
```go
func (g *Generator) generateStatement(stmt ast.Statement) {
    switch s := stmt.(type) {
    // ... 现有 case
    
    case *ast.NewStatement:
        g.generateNewStatement(s)
    }
}

func (g *Generator) generateNewStatement(stmt *ast.NewStatement) {
    // 生成 Go 代码
    g.writeLine("go code here")
}
```

---

## 添加语言特性

### 示例：添加三元运算符

**目标语法**:
```pascal
var max := if a > b then a else b;
```

**步骤**:

#### 1. 添加 Token（如果需要）

三元运算符使用现有关键字 `if`, `then`, `else`，无需新 Token。

#### 2. 添加 AST 节点

在 `ast/ast.go`:
```go
type TernaryExpression struct {
    Condition   Expression
    Consequence Expression
    Alternative Expression
}

func (t *TernaryExpression) expressionNode() {}
func (t *TernaryExpression) TokenLiteral() string { return "if" }
```

#### 3. 修改 Parser

在 `parser/parser.go`:

修改 `parseIfStatement` 以区分语句和表达式：
```go
func (p *Parser) parseIfStatement() ast.Node {
    p.nextToken() // skip 'if'
    
    condition := p.parseExpression(LOWEST)
    
    if !p.expectPeek(token.THEN) {
        return nil
    }
    
    // 检查是表达式还是语句
    if p.peekTokenIs(token.SEMICOLON) || p.peekTokenIs(token.END) {
        // 这是表达式
        consequence := p.parseExpression(LOWEST)
        
        if !p.expectPeek(token.ELSE) {
            return nil
        }
        
        alternative := p.parseExpression(LOWEST)
        
        return &ast.TernaryExpression{
            Condition:   condition,
            Consequence: consequence,
            Alternative: alternative,
        }
    } else {
        // 这是语句（现有逻辑）
        // ...
    }
}
```

#### 4. 修改 Generator

在 `generator/generator.go`:
```go
func (g *Generator) generateExpression(expr ast.Expression) {
    switch e := expr.(type) {
    // ... 现有 case
    
    case *ast.TernaryExpression:
        g.generateTernaryExpression(e)
    }
}

func (g *Generator) generateTernaryExpression(expr *ast.TernaryExpression) {
    // 生成 Go 的三元表达式
    // Go 没有三元运算符，需要转换为函数调用
    g.write("func() interface{} { if ")
    g.generateExpression(expr.Condition)
    g.write(" { return ")
    g.generateExpression(expr.Consequence)
    g.write(" }; return ")
    g.generateExpression(expr.Alternative)
    g.write(" }()")
}
```

#### 5. 测试

创建 `examples/ternary.klx`:
```pascal
program ternary;

begin
  var a := 10;
  var b := 20;
  var max := if a > b then a else b;
  WriteLn('Max: ', max);
end.
```

运行测试：
```bash
./kylix examples/ternary.klx
# 输出：Max: 20
```

---

## 测试

### 运行所有测试

```bash
go test ./...
```

### 测试特定模块

```bash
# 测试 lexer
go test ./lexer/...

# 测试 parser
go test ./parser/...

# 测试 generator
go test ./generator/...
```

### 添加单元测试

创建 `lexer/lexer_test.go`:
```go
package lexer

import (
    "testing"
    "kylix/token"
)

func TestNewFeature(t *testing.T) {
    input := `var x := 42;`
    l := New(input)
    
    tests := []struct {
        expectedType    token.TokenType
        expectedLiteral string
    }{
        {token.VAR, "var"},
        {token.IDENT, "x"},
        {token.ASSIGN_OP, ":="},
        {token.INT, "42"},
        {token.SEMICOLON, ";"},
    }
    
    for i, tt := range tests {
        tok := l.NextToken()
        
        if tok.Type != tt.expectedType {
            t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
                i, tt.expectedType, tok.Type)
        }
        
        if tok.Literal != tt.expectedLiteral {
            t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
                i, tt.expectedLiteral, tok.Literal)
        }
    }
}
```

### 集成测试

创建端到端测试：
```bash
#!/bin/bash

# test_examples.sh

for file in examples/*.klx; do
    echo "Testing $file..."
    ./kylix "$file" || exit 1
done

echo "All tests passed!"
```

---

## 贡献指南

### 提交代码

1. **Fork 仓库**
2. **创建功能分支**:
   ```bash
   git checkout -b feature/my-feature
   ```
3. **编写代码和测试**
4. **运行测试**:
   ```bash
   go test ./...
   ```
5. **提交**:
   ```bash
   git commit -am "Add feature: ..."
   ```
6. **推送并创建 PR**

### 代码规范

- 使用 `gofmt` 格式化代码
- 添加注释说明复杂逻辑
- 保持函数简短（< 50 行）
- 添加单元测试

### 报告 Bug

提交 Issue 时包含：
- Kylix 版本
- 操作系统
- 最小复现代码
- 期望行为 vs 实际行为
- 错误信息

### 功能建议

提交 Issue 时说明：
- 功能描述
- 使用场景
- 语法示例（如果是语言特性）
- 实现思路（可选）

---

## 路线图

早期 Phase 1–5（基础编译器、IDE 工具、Web 框架、语言增强、标准库）已全部完成 —— 泛型、异常、接口、多态、模式匹配、lambda/async、完整 stdlib（文件 I/O、网络、JSON、日期时间、正则、加密、db、websocket、jwt）均已落地。当前状态 **v0.6.9**（bootstrap 无 Go 闭环达成，gen1 ≡ gen2 ≡ gen3 逐字节 IR 不动点），下一站 v0.7.0（web 页面开发 + web 框架）。

完整路线图见 [ROADMAP.md](ROADMAP.md)，版本历史见 [CHANGELOG.md](CHANGELOG.md)。

---

## 常见问题

### Q: 为什么选择编译到 Go 而不是直接编译为机器码？

**A:** 
1. **快速实现**：Go 工具链成熟，无需自己实现后端
2. **性能**：Go 编译的代码性能优秀
3. **生态系统**：可以直接使用 Go 的标准库和第三方包
4. **简单**：编译器保持小巧，易于理解和维护

需要脱离 Go 工具链时，LLVM 后端（`--backend=llvm`）直接产出原生二进制——无 Go 运行时依赖，且支持 DWARF 调试、`--target` 交叉编译与 `-O2` 优化通道。参见 [llvm-backend.md](llvm-backend.md)。

### Q: Pratt 解析算法是什么？

**A:** Pratt 解析（也称"自顶向下运算符优先级解析"）是一种优雅的解析算法：
- 每个运算符都有优先级和解析函数
- 通过递归和优先级表处理复杂表达式
- 代码简洁，易于扩展

参考：[Simple Top-Down Parsing in Python](http://effbot.org/zone/simple-top-down-parsing.htm)

### Q: 如何调试编译器？

**A:** 使用 VS Code 的 Go 调试器：
```json
{
  "name": "Debug Compiler",
  "type": "go",
  "request": "launch",
  "mode": "auto",
  "program": "${workspaceFolder}/cmd/kylix/main.go",
  "args": ["build", "test.klx"]
}
```

在代码中设置断点，按 F5 启动调试。

### Q: 生成的 Go 代码性能如何？

**A:** 与手写 Go 代码性能相近。可以通过以下方式优化：
- 避免不必要的类型转换
- 使用内联函数
- 减少临时变量

### Q: 如何添加新的内置函数？

**A:** 在 `generator/generator.go` 的 `mapBuiltinFunction` 中添加映射：
```go
func (g *Generator) mapBuiltinFunction(name string) string {
    builtinMap := map[string]string{
        // ... 现有映射
        "NewFunction": "go.package.FunctionName",
    }
    
    if goFunc, ok := builtinMap[name]; ok {
        // 自动添加导入
        if strings.HasPrefix(goFunc, "go.package.") {
            g.imports["go.package"] = true
        }
        return goFunc
    }
    return name
}
```

---

## 参考资料

- [Writing an Interpreter in Go](https://interpreterbook.com/) - Monkey 语言实现
- [Crafting Interpreters](https://craftinginterpreters.com/) - 编译器实现教程
- [Pratt Parsing](https://matklad.github.io/2020/04/13/simple-but-powerful-pratt-parsing.html) - Pratt 解析详解
- [LSP Specification](https://microsoft.github.io/language-server-protocol/) - LSP 协议规范

---

## 联系

- **项目仓库**: [GitHub](https://github.com/astra-zhao/kylix)
- **问题反馈**: [Issues](https://github.com/astra-zhao/kylix/issues)
- **讨论**: [Discussions](https://github.com/astra-zhao/kylix/discussions)

感谢你的贡献！🎉
