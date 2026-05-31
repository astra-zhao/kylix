# Kylix Phase 2 增强实施计划

## 概述

本文档详细规划 Kylix 编译器的 Phase 2 增强工作，涵盖 AST 位置追踪、格式化器完善、REPL 增强、LSP 服务器完整实现和测试覆盖。

**总预计工作量**: 23-30 小时  
**实施阶段**: 7 个主要阶段  
**优先级**: P0 (核心功能) → P1 (重要功能) → P2 (增强功能)

---

## 当前状态分析

### ✅ 已完成
- ✅ Formatter 基本功能 (577 行)
- ✅ REPL 彩色输出和元命令 (319 行)
- ✅ LSP 基础框架 (500 行，仅有 completion/hover/diagnostics)

### 🔴 关键缺陷
- 🔴 **零测试** - 整个项目没有任何测试文件
- 🔴 **AST 缺少位置信息** - 大部分声明/语句节点没有位置追踪
- 🔴 **LSP 缺少符号表** - 无法实现跳转定义、查找引用等功能
- 🟡 **Formatter 缺少节点** - MatchStatement, RaiseStatement, PropertyDecl 等 6 种类型

### 📊 探索发现的关键问题

#### 1. AST 位置追踪问题
**现状**: 只有叶子表达式节点（字面量、标识符）通过 `Token` 字段携带位置信息。

**问题节点** (需要添加位置):
- 声明节点: `VarDecl`, `ConstDecl`, `TypeDecl`, `FunctionDecl`, `ClassDecl`, `InterfaceDecl`, `PropertyDecl`, `Parameter`
- 语句节点: `BlockStatement`, `AssignmentStatement`, `IfStatement`, `WhileStatement`, `ForStatement`, `ForEachStatement`, `RepeatStatement`, `CaseStatement`, `MatchStatement`, `TryStatement`, `RaiseStatement`, `BreakStatement`, `ContinueStatement`, `ReturnStatement`, `ExpressionStatement`
- 表达式节点: `CallExpression`, `MemberExpression`, `IndexExpression`, `ArrayLiteral`, `LambdaExpression`, `AwaitExpression`, `IsExpression`, `TypeCastExpression`

**总计**: 31 个 AST 节点类型需要添加位置追踪

#### 2. 解析器 Token 消费模式
发现两种主要模式：

**Pattern A: 捕获后推进** (大多数函数)
- 函数入口时 `p.curToken` 在关键字上
- 需要在 `p.nextToken()` 之前捕获 Token
- 例如: `parseIfStatement`, `parseWhileStatement`, `parseFunctionDecl`

**Pattern B: 调用者已消费关键字** (区段声明)
- 调用者在 `ParseProgram` 中已经消费了 `var`/`const`/`type`
- 函数入口时 `p.curToken` 在标识符上
- 需要重构：传递关键字 Token 或在调用前捕获
- 涉及: `parseSingleVarDecl`, `parseSingleConstDecl`, `parseSingleTypeDecl`

**Pattern C: 推进后创建** (丢失 Token)
- 先调用 `p.nextToken()`，后创建节点
- 需要在推进前捕获 Token 到局部变量
- 涉及: `BreakStatement`, `ContinueStatement`, `LambdaExpression`, `AwaitExpression`, `MemberExpression`, `IsExpression`, `TypeCastExpression`, `ExpressionStatement`, `AssignmentStatement`

#### 3. Formatter 缺失的 AST 节点
1. **MatchStatement** - 模式匹配语句
2. **RaiseStatement** - 异常抛出语句
3. **PropertyDecl** - 属性声明
4. **Visibility 修饰符** - private/public/protected/published
5. **StringInterpolation** - 字符串插值表达式
6. **LambdaExpression Block Body** - Lambda 的块语句体

#### 4. REPL 稳定性问题
1. **脆弱的语句检测**: `isCompleteStatement()` 使用简单字符串启发式
2. **错误的 begin/end 计数**: `strings.Count` 会被字符串字面量和注释欺骗
3. **无 readline 支持**: 使用原始 `bufio.Scanner`，无方向键历史导航
4. **突然退出**: `os.Exit(0)` 阻止清理
5. **缺少 :save 命令**
6. **stderr 合并**: `go run` 的 stderr 与 stdout 合并

#### 5. LSP 服务器缺失功能
**当前实现**:
- ✅ textDocument/didOpen, didChange, didClose
- ✅ textDocument/completion (仅关键字和内置函数)
- ✅ textDocument/hover (仅内置函数)
- ✅ textDocument/publishDiagnostics

**缺失功能**:
- ❌ 符号表/符号收集系统
- ❌ textDocument/definition (跳转定义)
- ❌ textDocument/references (查找引用)
- ❌ textDocument/documentSymbol (文档大纲)
- ❌ textDocument/signatureHelp (参数提示)
- ❌ textDocument/codeAction (代码操作)
- ❌ textDocument/rename (重命名)
- ❌ workspace/symbol (工作区符号搜索)
- ❌ textDocument/formatting (格式化)

**关键问题**: Server 只存储原始文本，没有 AST 缓存、符号表、作用域信息

---

## 详细实施计划

### Phase 1: AST 位置追踪 (基础工程)

**目标**: 为所有 AST 节点添加位置信息，使 LSP 功能成为可能

**预计时间**: 3-4 小时  
**优先级**: P0 (基础依赖)

#### 1.1 修改文件清单

**主要文件**:
- `ast/ast.go` - 添加 Position 结构和节点位置字段
- `parser/parser.go` - 在所有解析函数中捕获位置

**测试文件** (新增):
- `parser/parser_test.go` - 验证位置追踪正确性

#### 1.2 具体任务

**Task 1.2.1: 定义 Position 结构** (已完成 ✅)
```go
// 在 ast/ast.go 中
// token.Token 已包含 Line 和 Column，直接使用即可
```

**Task 1.2.2: 为声明节点添加位置** (部分完成 ⚠️)

需要修改的节点：
- [x] `Program` - 添加 `NameToken token.Token`
- [x] `VarDecl` - 添加 `Token token.Token` (关键字位置)
- [x] `ConstDecl` - 添加 `Token token.Token` (标识符位置)
- [x] `TypeDecl` - 添加 `Token token.Token` (标识符位置)
- [x] `FunctionDecl` - 添加 `Token token.Token` (function/procedure 关键字)
- [x] `ClassDecl` - 添加 `Token token.Token` (class 关键字)
- [x] `InterfaceDecl` - 添加 `Token token.Token` (interface 关键字)
- [x] `PropertyDecl` - 添加 `Token token.Token` (property 关键字)
- [x] `Parameter` - 添加 `Token token.Token` (参数名位置)

**Task 1.2.3: 为语句节点添加位置** (部分完成 ⚠️)

需要修改的节点：
- [x] `BlockStatement` - 添加 `Token token.Token` (begin 关键字)
- [x] `IfStatement` - 添加 `Token token.Token` (if 关键字)
- [x] `WhileStatement` - 添加 `Token token.Token` (while 关键字)
- [ ] `ForStatement` - 添加 `Token token.Token` (for 关键字)
- [ ] `ForEachStatement` - 添加 `Token token.Token` (for 关键字)
- [ ] `RepeatStatement` - 添加 `Token token.Token` (repeat 关键字)
- [ ] `CaseStatement` - 添加 `Token token.Token` (case 关键字)
- [ ] `MatchStatement` - 添加 `Token token.Token` (match 关键字)
- [ ] `TryStatement` - 添加 `Token token.Token` (try 关键字)
- [ ] `RaiseStatement` - 添加 `Token token.Token` (raise 关键字)
- [x] `BreakStatement` - 添加 `Token token.Token` (break 关键字)
- [x] `ContinueStatement` - 添加 `Token token.Token` (continue 关键字)
- [ ] `ReturnStatement` - 添加 `Token token.Token` (return 关键字)
- [ ] `AssignmentStatement` - 添加 `Token token.Token` (:= 位置)
- [ ] `ExpressionStatement` - 添加 `Token token.Token` (首 token 位置)

**Task 1.2.4: 为表达式节点添加位置** (待完成 ❌)

需要修改的节点：
- [ ] `CallExpression` - 添加 `Token token.Token` (左括号位置)
- [ ] `MemberExpression` - 添加 `Token token.Token` (点号位置)
- [ ] `IndexExpression` - 添加 `Token token.Token` (左方括号位置)
- [ ] `ArrayLiteral` - 添加 `Token token.Token` (左方括号位置)
- [ ] `LambdaExpression` - 添加 `Token token.Token` (左括号位置)
- [ ] `AwaitExpression` - 添加 `Token token.Token` (await 关键字)
- [ ] `IsExpression` - 添加 `Token token.Token` (is 关键字)
- [ ] `TypeCastExpression` - 添加 `Token token.Token` (as 关键字)

**Task 1.2.5: 修改解析器捕获位置** (部分完成 ⚠️)

对于每个解析函数，需要在适当位置添加 Token 捕获：

**Pattern A 函数** (在 nextToken 前捕获):
```go
func (p *Parser) parseIfStatement() *ast.IfStatement {
    stmt := &ast.IfStatement{Token: p.curToken} // 捕获 'if' token
    p.nextToken() // skip 'if'
    // ...
}
```

**Pattern B 函数** (传递关键字 token):
```go
// 在 ParseProgram 中
varToken := p.curToken
p.nextToken()
decl := p.parseSingleVarDecl(varToken)

// 在 parseSingleVarDecl 中
func (p *Parser) parseSingleVarDecl(varToken token.Token) *ast.VarDecl {
    decl := &ast.VarDecl{Token: varToken}
    // ...
}
```

**Pattern C 函数** (在推进前捕获到局部变量):
```go
case token.BREAK:
    tok := p.curToken
    p.nextToken()
    return &ast.BreakStatement{Token: tok}
```

**需要修改的解析函数**:
- [x] `ParseProgram` - Program.NameToken
- [x] `parseSingleVarDecl` - VarDecl.Token
- [x] `parseSingleConstDecl` - ConstDecl.Token
- [x] `parseSingleTypeDecl` - TypeDecl.Token
- [x] `parseFunctionDecl` - FunctionDecl.Token
- [x] `parseClassDecl` - ClassDecl.Token
- [x] `parseInterfaceDecl` - InterfaceDecl.Token
- [x] `parsePropertyDecl` - PropertyDecl.Token
- [x] `parseParameterList` - Parameter.Token
- [x] `parseBlockStatement` - BlockStatement.Token
- [x] `parseIfStatement` - IfStatement.Token
- [x] `parseWhileStatement` - WhileStatement.Token
- [ ] `parseForStatement` - ForStatement.Token, ForEachStatement.Token
- [ ] `parseRepeatStatement` - RepeatStatement.Token
- [ ] `parseCaseStatement` - CaseStatement.Token
- [ ] `parseMatchStatement` - MatchStatement.Token
- [ ] `parseTryStatement` - TryStatement.Token
- [ ] `parseRaiseStatement` - RaiseStatement.Token
- [ ] `parseReturnStatement` - ReturnStatement.Token
- [ ] `parseStatement` - BreakStatement.Token, ContinueStatement.Token
- [ ] `parseExpressionOrAssignment` - AssignmentStatement.Token, ExpressionStatement.Token
- [ ] `parseCallExpression` - CallExpression.Token
- [ ] `parseMemberExpression` - MemberExpression.Token
- [ ] `parseIndexExpression` - IndexExpression.Token
- [ ] `parseArrayLiteral` - ArrayLiteral.Token
- [ ] `parseGroupedExpression` - LambdaExpression.Token
- [ ] `parseAwaitExpression` - AwaitExpression.Token
- [ ] `parseIsExpression` - IsExpression.Token
- [ ] `parseAsExpression` - TypeCastExpression.Token

#### 1.3 验证方法

```bash
# 编译测试
go build ./...

# 运行现有示例确保不破坏
./kylix run examples/hello.klx
./kylix run examples/types.klx

# 添加位置追踪测试
go test ./parser -v -run TestPositionTracking
```

**验收标准**:
- ✅ 所有 31 个 AST 节点类型都有位置信息
- ✅ 所有解析函数正确捕获位置
- ✅ 现有示例仍能正常编译运行
- ✅ 位置信息准确（行号、列号正确）

---

### Phase 2: Formatter 完善

**目标**: 支持所有 AST 节点类型，处理所有示例文件

**预计时间**: 2-3 小时  
**优先级**: P1

#### 2.1 修改文件清单

**主要文件**:
- `pkg/formatter/formatter.go` - 添加缺失的节点格式化

**测试文件** (新增):
- `pkg/formatter/formatter_test.go` - 测试所有格式化功能

#### 2.2 具体任务

**Task 2.2.1: MatchStatement 格式化**

在 `formatStatement` 中添加:
```go
case *ast.MatchStatement:
    f.formatMatchStatement(s)
```

新增函数:
```go
func (f *Formatter) formatMatchStatement(stmt *ast.MatchStatement) {
    f.writeIndent()
    f.write("match ")
    f.formatExpression(stmt.Expression)
    f.write(" {\n")
    f.indent++
    
    for _, branch := range stmt.Branches {
        f.writeIndent()
        if branch.When != nil {
            f.write("when ")
            f.formatExpression(branch.When)
        } else {
            f.formatExpression(branch.Pattern)
        }
        f.write(" => ")
        
        if len(branch.Body.Statements) == 1 {
            f.formatStatement(branch.Body.Statements[0])
        } else {
            f.write("\n")
            f.indent++
            f.formatBlock(branch.Body)
            f.write(";\n")
            f.indent--
        }
        f.write(",\n")
    }
    
    f.indent--
    f.writeIndent()
    f.write("};\n")
}
```

**Task 2.2.2: RaiseStatement 格式化**

在 `formatStatement` 中添加:
```go
case *ast.RaiseStatement:
    f.writeIndent()
    f.write("raise")
    if s.Exception != nil {
        f.write(" ")
        f.formatExpression(s.Exception)
    }
    f.write(";\n")
```

**Task 2.2.3: PropertyDecl 格式化**

在 `formatClassDecl` 中添加属性格式化:
```go
// Properties
if len(decl.Properties) > 0 {
    f.writeLine("")
    for _, prop := range decl.Properties {
        f.formatPropertyDecl(prop)
    }
}
```

新增函数:
```go
func (f *Formatter) formatPropertyDecl(prop *ast.PropertyDecl) {
    f.writeIndent()
    f.write("property " + prop.Name)
    if prop.Type != nil {
        f.write(": ")
        f.formatType(prop.Type)
    }
    if prop.Getter != "" {
        f.write(" read " + prop.Getter)
    }
    if prop.Setter != "" {
        f.write(" write " + prop.Setter)
    }
    if prop.Default != nil {
        f.write(" default ")
        f.formatExpression(prop.Default)
    }
    f.write(";\n")
}
```

**Task 2.2.4: Visibility 修饰符**

修改 `formatClassDecl` 以输出可见性修饰符:
```go
func (f *Formatter) formatClassDecl(decl *ast.ClassDecl) {
    // ... 现有代码 ...
    
    f.indent++
    
    // 按可见性分组输出
    currentVisibility := token.PUBLIC
    
    // Fields
    if len(decl.Fields) > 0 {
        f.writeIndent()
        f.write("var\n")
        f.indent++
        for _, field := range decl.Fields {
            // 检查是否需要切换可见性
            if field.Visibility != currentVisibility {
                f.indent--
                f.writeLine("")
                f.writeIndent()
                f.write(strings.ToLower(field.Visibility) + "\n")
                f.indent++
                currentVisibility = field.Visibility
            }
            f.formatVarDecl(field)
        }
        f.indent--
        f.writeLine("")
    }
    
    // Methods (类似处理)
    // ...
}
```

**注意**: 需要先在 AST 中为 VarDecl, FunctionDecl, PropertyDecl 添加 `Visibility` 字段，并在解析器中设置。

**Task 2.2.5: StringInterpolation 格式化**

在 `formatExpression` 中添加:
```go
case *ast.StringInterpolation:
    f.write("'")
    for _, part := range e.Parts {
        if str, ok := part.(*ast.StringLiteral); ok {
            f.write(str.Value)
        } else {
            f.write("${")
            f.formatExpression(part)
            f.write("}")
        }
    }
    f.write("'")
```

**Task 2.2.6: LambdaExpression Block Body**

修改 `formatExpression` 中的 lambda 处理:
```go
case *ast.LambdaExpression:
    f.write("(")
    for i, param := range e.Parameters {
        if i > 0 {
            f.write("; ")
        }
        f.write(param.Name + ": ")
        f.formatType(param.Type)
    }
    f.write(") -> ")
    
    switch body := e.Body.(type) {
    case ast.Expression:
        f.formatExpression(body)
    case *ast.BlockStatement:
        f.write("\n")
        f.indent++
        f.formatBlock(body)
        f.write(";\n")
        f.indent--
    }
```

#### 2.3 验证方法

```bash
# 格式化所有示例文件
for f in examples/*.klx; do
    ./kylix fmt "$f" > /dev/null
done

# 往返测试: 格式化 → 解析 → 格式化 应该产生相同结果
./kylix fmt examples/hello.klx > /tmp/hello1.klx
./kylix fmt /tmp/hello1.klx > /tmp/hello2.klx
diff /tmp/hello1.klx /tmp/hello2.klx  # 应该无差异

# 运行格式化测试
go test ./pkg/formatter -v
```

**验收标准**:
- ✅ 所有 examples/*.klx 都能正确格式化
- ✅ 格式化输出符合 Pascal 风格规范
- ✅ 往返测试通过
- ✅ 所有新增的格式化函数都有测试覆盖

---

### Phase 3: REPL 稳定性改进

**目标**: 更智能的多行检测，更好的用户体验

**预计时间**: 2-3 小时  
**优先级**: P1

#### 3.1 修改文件清单

**主要文件**:
- `pkg/repl/repl.go` - 改进多行检测和错误处理
- `go.mod` - 添加 readline 依赖

**测试文件** (新增):
- `pkg/repl/repl_test.go` - 测试 REPL 功能

#### 3.2 具体任务

**Task 3.2.1: 基于 AST 的多行检测**

替换简单的 begin/end 计数:
```go
func isCompleteCode(code string) (complete bool, err error) {
    // 尝试解析当前输入
    l := lexer.New(code)
    p := parser.New(l)
    p.ParseProgram()
    
    errors := p.Errors()
    
    // 如果没有错误，代码完整
    if len(errors) == 0 {
        return true, nil
    }
    
    // 检查是否是"unexpected EOF"错误
    for _, e := range errors {
        if strings.Contains(e, "unexpected EOF") ||
           strings.Contains(e, "expected") {
            return false, nil // 需要更多输入
        }
    }
    
    // 其他错误是真正的语法错误
    return true, fmt.Errorf("syntax error")
}
```

**Task 3.2.2: 改进 begin/end 计数**

如果保留计数方法，使用词法分析:
```go
func countBeginEnd(code string) (begins, ends int) {
    l := lexer.New(code)
    for {
        tok := l.NextToken()
        if tok.Type == token.EOF {
            break
        }
        // 跳过注释和字符串
        if tok.Type == token.COMMENT {
            continue
        }
        if tok.Type == token.STRING {
            continue
        }
        if tok.Type == token.BEGIN {
            begins++
        }
        if tok.Type == token.END {
            ends++
        }
    }
    return
}
```

**Task 3.2.3: 添加 readline 支持**

安装依赖:
```bash
go get github.com/chzyer/readline
```

修改 REPL 主循环:
```go
import "github.com/chzyer/readline"

func Start() error {
    rl, err := readline.NewEx(&readline.Config{
        Prompt:          colorGreen + prompt + colorReset,
        HistoryFile:     filepath.Join(os.TempDir(), "kylix_history"),
        AutoComplete:    nil, // 可以后续添加
        InterruptPrompt: "^C",
        EOFPrompt:       "exit",
    })
    if err != nil {
        return err
    }
    defer rl.Close()

    for {
        line, err := rl.Readline()
        if err == readline.ErrInterrupt {
            continue
        }
        if err == io.EOF {
            break
        }
        
        // ... 处理输入
    }
    
    return nil
}
```

**Task 3.2.4: 改进错误处理**

替换 `os.Exit(0)`:
```go
case ":quit", ":q", ":exit":
    fmt.Fprintln(r.out, colorCyan+"Goodbye!"+colorReset)
    return nil // 返回而不是退出
```

**Task 3.2.5: 添加 :save 命令**

```go
case ":save":
    if len(parts) < 2 {
        fmt.Fprintln(r.out, colorRed+"Usage: :save <filename>"+colorReset)
        return
    }
    r.saveToFile(parts[1])
```

新增函数:
```go
func (r *REPL) saveToFile(filename string) {
    var content strings.Builder
    content.WriteString("program saved;\n\n")
    
    for _, decl := range r.declarations {
        content.WriteString(decl)
        content.WriteString("\n")
    }
    
    content.WriteString("\nbegin\n")
    for _, cmd := range r.history {
        if !strings.HasPrefix(cmd, ":") && !isDeclaration(cmd) {
            content.WriteString("  ")
            content.WriteString(cmd)
            content.WriteString("\n")
        }
    }
    content.WriteString("end.\n")
    
    err := os.WriteFile(filename, []byte(content.String()), 0644)
    if err != nil {
        fmt.Fprintf(r.out, colorRed+"Error saving: %v"+colorReset+"\n", err)
    } else {
        fmt.Fprintf(r.out, colorGreen+"✓ Saved to %s"+colorReset+"\n", filename)
    }
}
```

**Task 3.2.6: 改进输出格式**

分离 stdout 和 stderr:
```go
var stderrBuf bytes.Buffer
cmd.Stdout = r.out
cmd.Stderr = &stderrBuf
err := cmd.Run()

if stderrBuf.Len() > 0 {
    fmt.Fprintln(r.out, colorRed+"Errors:"+colorReset)
    fmt.Fprint(r.out, stderrBuf.String())
}

if err != nil && stderrBuf.Len() == 0 {
    fmt.Fprintf(r.out, colorRed+"Runtime error: %v"+colorReset+"\n", err)
}
```

#### 3.3 验证方法

```bash
# 测试多行检测
echo -e "var x := 10\nbegin\nWriteLn(x)\nend." | ./kylix repl

# 测试 readline (方向键历史)
./kylix repl
# 输入: var x := 10
# 输入: WriteLn(x)
# 按上箭头应该能恢复之前的输入

# 测试 :save 命令
./kylix repl
# 输入: var x := 10
# 输入: WriteLn(x)
# 输入: :save test.klx
# 检查 test.klx 文件

# 运行 REPL 测试
go test ./pkg/repl -v
```

**验收标准**:
- ✅ 多行检测准确，不会被字符串/注释欺骗
- ✅ 方向键可以浏览历史
- ✅ :quit 优雅退出
- ✅ :save 可以保存会话
- ✅ stderr 和 stdout 分离
- ✅ 所有新功能都有测试覆盖

---

### Phase 4: LSP 符号表 (核心功能)

**目标**: 构建符号收集和缓存系统，为所有 LSP 功能奠定基础

**预计时间**: 4-5 小时  
**优先级**: P0 (基础依赖)

#### 4.1 修改文件清单

**新增文件**:
- `pkg/lsp/symbols.go` - 符号表定义和收集器
- `pkg/lsp/document.go` - 文档状态管理
- `pkg/lsp/handlers_definition.go` - 跳转定义处理器
- `pkg/lsp/handlers_references.go` - 查找引用处理器
- `pkg/lsp/handlers_document_symbol.go` - 文档大纲处理器
- `pkg/lsp/handlers_signature_help.go` - 参数提示处理器
- `pkg/lsp/handlers_rename.go` - 重命名处理器
- `pkg/lsp/handlers_code_action.go` - 代码操作处理器
- `pkg/lsp/handlers_formatting.go` - 格式化处理器
- `pkg/lsp/handlers_workspace_symbol.go` - 工作区符号搜索处理器

**修改文件**:
- `pkg/lsp/server.go` - 集成符号表，更新 capabilities

**测试文件** (新增):
- `pkg/lsp/symbols_test.go` - 符号收集测试
- `pkg/lsp/server_test.go` - LSP 功能集成测试

#### 4.2 具体任务

**Task 4.2.1: 定义符号类型** (symbols.go)

```go
package lsp

type SymbolKind int

const (
    SymbolVariable SymbolKind = iota
    SymbolConstant
    SymbolType
    SymbolFunction
    SymbolProcedure
    SymbolClass
    SymbolInterface
    SymbolMethod
    SymbolField
    SymbolProperty
    SymbolParameter
)

type Symbol struct {
    Name       string
    Kind       SymbolKind
    Detail     string      // 类型签名或描述
    Location   Location    // URI + Range
    Scope      *Scope      // 父作用域
    Children   []*Symbol   // 子符号 (类的成员等)
}

type Scope struct {
    Parent   *Scope
    Symbols  map[string]*Symbol
    Children []*Scope
}

type SymbolTable struct {
    Root   *Scope
    AllSymbols []*Symbol // 扁平列表，便于搜索
}
```

**Task 4.2.2: 构建符号收集器** (symbols.go)

```go
func CollectSymbols(program *ast.Program, uri string) *SymbolTable {
    collector := &symbolCollector{
        uri: uri,
        currentScope: &Scope{
            Symbols: make(map[string]*Symbol),
        },
    }
    
    collector.collectProgram(program)
    
    return &SymbolTable{
        Root: collector.rootScope,
        AllSymbols: collector.allSymbols,
    }
}

type symbolCollector struct {
    uri string
    rootScope *Scope
    currentScope *Scope
    allSymbols []*Symbol
}

func (c *symbolCollector) collectProgram(program *ast.Program) {
    c.rootScope = c.currentScope
    
    // 收集全局声明
    for _, decl := range program.Declarations {
        c.collectDeclaration(decl)
    }
    
    // 收集主程序块中的声明
    for _, stmt := range program.Statements {
        c.collectStatement(stmt)
    }
}

func (c *symbolCollector) collectDeclaration(node ast.Node) {
    switch d := node.(type) {
    case *ast.VarDecl:
        for _, name := range d.Names {
            symbol := &Symbol{
                Name: name,
                Kind: SymbolVariable,
                Detail: c.formatType(d.Type),
                Location: Location{
                    URI: c.uri,
                    Range: c.tokenToRange(d.Token),
                },
                Scope: c.currentScope,
            }
            c.addSymbol(symbol)
        }
    
    case *ast.FunctionDecl:
        symbol := &Symbol{
            Name: d.Name,
            Kind: SymbolFunction,
            Detail: c.formatFunctionSignature(d),
            Location: Location{
                URI: c.uri,
                Range: c.tokenToRange(d.Token),
            },
            Scope: c.currentScope,
        }
        
        // 创建函数作用域
        funcScope := &Scope{
            Parent: c.currentScope,
            Symbols: make(map[string]*Symbol),
        }
        c.currentScope.Children = append(c.currentScope.Children, funcScope)
        
        // 收集参数
        for _, param := range d.Parameters {
            paramSymbol := &Symbol{
                Name: param.Name,
                Kind: SymbolParameter,
                Detail: c.formatType(param.Type),
                Location: Location{
                    URI: c.uri,
                    Range: c.tokenToRange(param.Token),
                },
                Scope: funcScope,
            }
            funcScope.Symbols[param.Name] = paramSymbol
            c.allSymbols = append(c.allSymbols, paramSymbol)
        }
        
        // 切换到函数作用域，收集函数体
        oldScope := c.currentScope
        c.currentScope = funcScope
        if d.Body != nil {
            c.collectBlock(d.Body)
        }
        c.currentScope = oldScope
        
        c.addSymbol(symbol)
    
    case *ast.ClassDecl:
        symbol := &Symbol{
            Name: d.Name,
            Kind: SymbolClass,
            Detail: c.formatClassDetail(d),
            Location: Location{
                URI: c.uri,
                Range: c.tokenToRange(d.Token),
            },
            Scope: c.currentScope,
        }
        
        // 创建类作用域
        classScope := &Scope{
            Parent: c.currentScope,
            Symbols: make(map[string]*Symbol),
        }
        
        // 收集字段
        for _, field := range d.Fields {
            for _, name := range field.Names {
                fieldSymbol := &Symbol{
                    Name: name,
                    Kind: SymbolField,
                    Detail: c.formatType(field.Type),
                    Location: Location{
                        URI: c.uri,
                        Range: c.tokenToRange(field.Token),
                    },
                    Scope: classScope,
                }
                classScope.Symbols[name] = fieldSymbol
                c.allSymbols = append(c.allSymbols, fieldSymbol)
                symbol.Children = append(symbol.Children, fieldSymbol)
            }
        }
        
        // 收集方法
        for _, method := range d.Methods {
            methodSymbol := c.collectMethod(method, classScope)
            symbol.Children = append(symbol.Children, methodSymbol)
        }
        
        c.addSymbol(symbol)
    
    // ... 处理其他声明类型
    }
}

func (c *symbolCollector) addSymbol(symbol *Symbol) {
    c.currentScope.Symbols[symbol.Name] = symbol
    c.allSymbols = append(c.allSymbols, symbol)
}

func (c *symbolCollector) tokenToRange(tok token.Token) Range {
    return Range{
        Start: Position{Line: tok.Line - 1, Character: tok.Column - 1},
        End: Position{Line: tok.Line - 1, Character: tok.Column - 1 + len(tok.Literal)},
    }
}

func (c *symbolCollector) formatType(expr ast.Expression) string {
    if expr == nil {
        return ""
    }
    // ... 格式化类型表达式
}

func (c *symbolCollector) formatFunctionSignature(decl *ast.FunctionDecl) string {
    // ... 格式化函数签名
}

func (c *symbolCollector) formatClassDetail(decl *ast.ClassDecl) string {
    // ... 格式化类描述
}
```

**Task 4.2.3: 定义 Document 状态** (document.go)

```go
package lsp

type Document struct {
    URI         string
    Text        string
    Lines       []string       // 按行分割的文本
    AST         *ast.Program
    Symbols     *SymbolTable
    Diagnostics []Diagnostic
}

type DocumentStore struct {
    docs map[string]*Document
    mu   sync.RWMutex
}

func NewDocumentStore() *DocumentStore {
    return &DocumentStore{
        docs: make(map[string]*Document),
    }
}

func (ds *DocumentStore) Update(uri, text string) *Document {
    ds.mu.Lock()
    defer ds.mu.Unlock()
    
    // 解析为 AST
    l := lexer.New(text)
    p := parser.New(l)
    ast := p.ParseProgram()
    
    // 收集符号
    symbols := CollectSymbols(ast, uri)
    
    // 生成诊断
    diagnostics := []Diagnostic{}
    for _, errMsg := range p.Errors() {
        line, col := parseLocation(errMsg)
        diagnostics = append(diagnostics, Diagnostic{
            Range: Range{
                Start: Position{Line: line, Character: col},
                End: Position{Line: line, Character: col + 1},
            },
            Severity: 1,
            Message: errMsg,
        })
    }
    
    doc := &Document{
        URI: uri,
        Text: text,
        Lines: strings.Split(text, "\n"),
        AST: ast,
        Symbols: symbols,
        Diagnostics: diagnostics,
    }
    
    ds.docs[uri] = doc
    return doc
}

func (ds *DocumentStore) Get(uri string) *Document {
    ds.mu.RLock()
    defer ds.mu.RUnlock()
    return ds.docs[uri]
}

func (ds *DocumentStore) Delete(uri string) {
    ds.mu.Lock()
    defer ds.mu.Unlock()
    delete(ds.docs, uri)
}
```

**Task 4.2.4: 修改 Server 结构** (server.go)

```go
type Server struct {
    in  io.Reader
    out io.Writer
    docs *DocumentStore // 替换 map[string]string
}

func New(in io.Reader, out io.Writer) *Server {
    return &Server{
        in: in,
        out: out,
        docs: NewDocumentStore(),
    }
}

func (s *Server) handleDidOpen(msg *Message) {
    var params DidOpenParams
    json.Unmarshal(msg.Params, &params)
    
    doc := s.docs.Update(params.TextDocument.URI, params.TextDocument.Text)
    s.publishDiagnostics(doc)
}

func (s *Server) handleDidChange(msg *Message) {
    var params DidChangeParams
    json.Unmarshal(msg.Params, &params)
    
    for _, change := range params.ContentChanges {
        doc := s.docs.Update(params.TextDocument.URI, change.Text)
        s.publishDiagnostics(doc)
    }
}

func (s *Server) publishDiagnostics(doc *Document) {
    s.writeMessage(&Message{
        Method: "textDocument/publishDiagnostics",
        Params: mustMarshal(PublishDiagnosticsParams{
            URI: doc.URI,
            Diagnostics: doc.Diagnostics,
        }),
    })
}
```

**Task 4.2.5: 更新 initialize 响应** (server.go)

```go
func (s *Server) handleInitialize(msg *Message) *Message {
    return &Message{
        ID: msg.ID,
        Result: map[string]interface{}{
            "capabilities": map[string]interface{}{
                "textDocumentSync": 1,
                "completionProvider": map[string]interface{}{
                    "triggerCharacters": []string{".", ":"},
                },
                "hoverProvider": true,
                "definitionProvider": true,
                "referencesProvider": true,
                "documentSymbolProvider": true,
                "signatureHelpProvider": map[string]interface{}{
                    "triggerCharacters": []string{"(", ","},
                },
                "codeActionProvider": true,
                "renameProvider": true,
                "documentFormattingProvider": true,
                "workspaceSymbolProvider": true,
            },
            "serverInfo": map[string]interface{}{
                "name": "kylix-lsp",
                "version": "0.3.0",
            },
        },
    }
}
```

#### 4.3 验证方法

```bash
# 编译测试
go build ./pkg/lsp/...

# 符号收集测试
go test ./pkg/lsp -v -run TestCollectSymbols

# 手动测试 (在 VS Code 中)
# 1. 启动 kylix lsp
# 2. 打开 .klx 文件
# 3. 检查诊断是否正确显示
```

**验收标准**:
- ✅ 符号表正确收集所有声明
- ✅ 作用域嵌套正确
- ✅ 文档状态缓存工作正常
- ✅ 诊断信息实时更新
- ✅ 所有符号收集函数都有测试覆盖

---

### Phase 5: LSP 功能实现

**目标**: 实现完整的 LSP 功能集

**预计时间**: 6-8 小时  
**优先级**: P0

#### 5.1 修改文件清单

**新增文件**: 见 Phase 4 列表

**修改文件**:
- `pkg/lsp/server.go` - 路由新的 LSP 方法
- `pkg/lsp/handlers_*.go` - 各个处理器实现

#### 5.2 具体任务

**Task 5.2.1: textDocument/documentSymbol** (handlers_document_symbol.go)

```go
func (s *Server) handleDocumentSymbol(msg *Message) *Message {
    var params DocumentSymbolParams
    json.Unmarshal(msg.Params, &params)
    
    doc := s.docs.Get(params.TextDocument.URI)
    if doc == nil {
        return &Message{ID: msg.ID, Result: []DocumentSymbol{}}
    }
    
    symbols := s.convertToDocumentSymbols(doc.Symbols.Root)
    return &Message{ID: msg.ID, Result: symbols}
}

func (s *Server) convertToDocumentSymbols(scope *Scope) []DocumentSymbol {
    var result []DocumentSymbol
    
    for _, symbol := range scope.Symbols {
        ds := DocumentSymbol{
            Name: symbol.Name,
            Kind: s.symbolKindToLSP(symbol.Kind),
            Range: symbol.Location.Range,
            SelectionRange: symbol.Location.Range,
            Detail: symbol.Detail,
        }
        
        // 递归处理子符号
        if len(symbol.Children) > 0 {
            childScope := s.findScopeForSymbol(symbol)
            if childScope != nil {
                ds.Children = s.convertToDocumentSymbols(childScope)
            }
        }
        
        result = append(result, ds)
    }
    
    return result
}

func (s *Server) symbolKindToLSP(kind SymbolKind) int {
    switch kind {
    case SymbolVariable:
        return 13 // Variable
    case SymbolConstant:
        return 14 // Constant
    case SymbolFunction:
        return 12 // Function
    case SymbolClass:
        return 5 // Class
    case SymbolInterface:
        return 11 // Interface
    case SymbolMethod:
        return 6 // Method
    case SymbolField:
        return 8 // Field
    case SymbolProperty:
        return 7 // Property
    default:
        return 1 // File
    }
}
```

**Task 5.2.2: textDocument/definition** (handlers_definition.go)

```go
func (s *Server) handleDefinition(msg *Message) *Message {
    var params DefinitionParams
    json.Unmarshal(msg.Params, &params)
    
    doc := s.docs.Get(params.TextDocument.URI)
    if doc == nil {
        return &Message{ID: msg.ID, Result: nil}
    }
    
    // 获取光标位置的标识符
    identifier := s.getIdentifierAtPosition(doc, params.Position)
    if identifier == "" {
        return &Message{ID: msg.ID, Result: nil}
    }
    
    // 在符号表中查找定义
    symbol := s.findSymbol(doc.Symbols, identifier, params.Position)
    if symbol == nil {
        return &Message{ID: msg.ID, Result: nil}
    }
    
    return &Message{
        ID: msg.ID,
        Result: Location{
            URI: symbol.Location.URI,
            Range: symbol.Location.Range,
        },
    }
}

func (s *Server) getIdentifierAtPosition(doc *Document, pos Position) string {
    if pos.Line >= len(doc.Lines) {
        return ""
    }
    
    line := doc.Lines[pos.Line]
    if pos.Character >= len(line) {
        return ""
    }
    
    // 找到标识符边界
    start := pos.Character
    for start > 0 && isIdentChar(line[start-1]) {
        start--
    }
    
    end := pos.Character
    for end < len(line) && isIdentChar(line[end]) {
        end++
    }
    
    if start == end {
        return ""
    }
    
    return line[start:end]
}

func (s *Server) findSymbol(table *SymbolTable, name string, pos Position) *Symbol {
    // 从当前位置的作用域向上查找
    scope := s.findScopeAtPosition(table, pos)
    
    for scope != nil {
        if symbol, ok := scope.Symbols[name]; ok {
            return symbol
        }
        scope = scope.Parent
    }
    
    return nil
}

func (s *Server) findScopeAtPosition(table *SymbolTable, pos Position) *Scope {
    // 查找包含该位置的最内层作用域
    // ... 实现作用域查找逻辑
}
```

**Task 5.2.3: textDocument/references** (handlers_references.go)

```go
func (s *Server) handleReferences(msg *Message) *Message {
    var params ReferenceParams
    json.Unmarshal(msg.Params, &params)
    
    doc := s.docs.Get(params.TextDocument.URI)
    if doc == nil {
        return &Message{ID: msg.ID, Result: []Location{}}
    }
    
    // 获取目标符号
    identifier := s.getIdentifierAtPosition(doc, params.Position)
    if identifier == "" {
        return &Message{ID: msg.ID, Result: []Location{}}
    }
    
    // 遍历 AST 查找所有引用
    references := s.findAllReferences(doc.AST, identifier, doc.URI)
    
    return &Message{ID: msg.ID, Result: references}
}

func (s *Server) findAllReferences(program *ast.Program, name string, uri string) []Location {
    var locations []Location
    
    // 使用 AST 遍历器查找所有匹配的名称
    walker := &referenceWalker{
        name: name,
        uri: uri,
        locations: &locations,
    }
    
    walker.walkProgram(program)
    
    return locations
}

type referenceWalker struct {
    name string
    uri string
    locations *[]Location
}

func (w *referenceWalker) walkProgram(program *ast.Program) {
    for _, decl := range program.Declarations {
        w.walkDeclaration(decl)
    }
    for _, stmt := range program.Statements {
        w.walkStatement(stmt)
    }
}

func (w *referenceWalker) walkIdentifier(ident *ast.Identifier) {
    if ident.Value == w.name {
        *w.locations = append(*w.locations, Location{
            URI: w.uri,
            Range: Range{
                Start: Position{Line: ident.Token.Line - 1, Character: ident.Token.Column - 1},
                End: Position{Line: ident.Token.Line - 1, Character: ident.Token.Column - 1 + len(ident.Value)},
            },
        })
    }
}

// ... 实现 walkDeclaration, walkStatement, walkExpression 等方法
```

**Task 5.2.4: 增强 textDocument/completion** (server.go)

```go
func (s *Server) handleCompletion(msg *Message) *Message {
    var params CompletionParams
    json.Unmarshal(msg.Params, &params)
    
    doc := s.docs.Get(params.TextDocument.URI)
    
    items := []CompletionItem{}
    
    // 1. 关键字补全
    keywords := []string{
        "program", "var", "const", "type", "function", "procedure",
        "begin", "end", "if", "then", "else", "while", "for", "to",
        // ... 更多关键字
    }
    for _, kw := range keywords {
        items = append(items, CompletionItem{
            Label: kw,
            Kind: 14, // Keyword
        })
    }
    
    // 2. 内置函数补全
    builtins := []CompletionItem{
        {Label: "WriteLn", Kind: 3, Detail: "procedure", InsertText: "WriteLn("},
        {Label: "ReadLn", Kind: 3, Detail: "procedure", InsertText: "ReadLn("},
        // ... 更多内置函数
    }
    items = append(items, builtins...)
    
    // 3. 用户定义符号补全 (新增)
    if doc != nil {
        for _, symbol := range doc.Symbols.AllSymbols {
            items = append(items, CompletionItem{
                Label: symbol.Name,
                Kind: s.symbolKindToCompletionKind(symbol.Kind),
                Detail: symbol.Detail,
            })
        }
    }
    
    return &Message{ID: msg.ID, Result: items}
}

func (s *Server) symbolKindToCompletionKind(kind SymbolKind) int {
    switch kind {
    case SymbolVariable:
        return 6 // Variable
    case SymbolFunction:
        return 3 // Function
    case SymbolClass:
        return 7 // Class
    default:
        return 1 // Text
    }
}
```

**Task 5.2.5: textDocument/signatureHelp** (handlers_signature_help.go)

```go
func (s *Server) handleSignatureHelp(msg *Message) *Message {
    var params SignatureHelpParams
    json.Unmarshal(msg.Params, &params)
    
    doc := s.docs.Get(params.TextDocument.URI)
    if doc == nil {
        return &Message{ID: msg.ID, Result: nil}
    }
    
    // 检测光标是否在函数调用内
    callInfo := s.detectFunctionCall(doc, params.Position)
    if callInfo == nil {
        return &Message{ID: msg.ID, Result: nil}
    }
    
    // 查找函数签名
    symbol := s.findSymbol(doc.Symbols, callInfo.FunctionName, params.Position)
    if symbol == nil || symbol.Kind != SymbolFunction {
        return &Message{ID: msg.ID, Result: nil}
    }
    
    // 解析函数签名
    signature := s.parseFunctionSignature(symbol.Detail)
    
    return &Message{
        ID: msg.ID,
        Result: SignatureHelp{
            Signatures: []SignatureInformation{
                {
                    Label: signature.Label,
                    Parameters: signature.Parameters,
                },
            },
            ActiveSignature: 0,
            ActiveParameter: callInfo.ActiveParameter,
        },
    }
}

type FunctionCallInfo struct {
    FunctionName string
    ActiveParameter int
}

func (s *Server) detectFunctionCall(doc *Document, pos Position) *FunctionCallInfo {
    // 从光标位置向前扫描，查找函数名和参数位置
    // ... 实现函数调用检测逻辑
}

func (s *Server) parseFunctionSignature(detail string) *SignatureInformation {
    // 解析 "function Name(param1: Type1; param2: Type2): ReturnType"
    // ... 实现签名解析
}
```

**Task 5.2.6: textDocument/rename** (handlers_rename.go)

```go
func (s *Server) handleRename(msg *Message) *Message {
    var params RenameParams
    json.Unmarshal(msg.Params, &params)
    
    doc := s.docs.Get(params.TextDocument.URI)
    if doc == nil {
        return &Message{ID: msg.ID, Result: nil}
    }
    
    // 查找所有引用
    references := s.findAllReferences(doc.AST, params.Position, doc.URI)
    
    // 生成编辑
    edits := []TextEdit{}
    for _, ref := range references {
        edits = append(edits, TextEdit{
            Range: ref.Range,
            NewText: params.NewName,
        })
    }
    
    return &Message{
        ID: msg.ID,
        Result: WorkspaceEdit{
            Changes: map[string][]TextEdit{
                doc.URI: edits,
            },
        },
    }
}
```

**Task 5.2.7: textDocument/codeAction** (handlers_code_action.go)

```go
func (s *Server) handleCodeAction(msg *Message) *Message {
    var params CodeActionParams
    json.Unmarshal(msg.Params, &params)
    
    doc := s.docs.Get(params.TextDocument.URI)
    if doc == nil {
        return &Message{ID: msg.ID, Result: []CodeAction{}}
    }
    
    actions := []CodeAction{}
    
    // 示例: 提取变量
    actions = append(actions, CodeAction{
        Title: "Extract Variable",
        Kind: "refactor.extract",
        Command: Command{
            Title: "Extract Variable",
            Command: "kylix.extractVariable",
            Arguments: []interface{}{doc.URI, params.Range},
        },
    })
    
    // 示例: 提取函数
    actions = append(actions, CodeAction{
        Title: "Extract Function",
        Kind: "refactor.extract",
        Command: Command{
            Title: "Extract Function",
            Command: "kylix.extractFunction",
            Arguments: []interface{}{doc.URI, params.Range},
        },
    })
    
    return &Message{ID: msg.ID, Result: actions}
}
```

**Task 5.2.8: textDocument/formatting** (handlers_formatting.go)

```go
func (s *Server) handleFormatting(msg *Message) *Message {
    var params DocumentFormattingParams
    json.Unmarshal(msg.Params, &params)
    
    doc := s.docs.Get(params.TextDocument.URI)
    if doc == nil {
        return &Message{ID: msg.ID, Result: []TextEdit{}}
    }
    
    // 使用现有的 formatter
    formatter := formatter.New()
    formatted := formatter.Format(doc.AST)
    
    // 计算整个文档的范围
    lastLine := len(doc.Lines) - 1
    lastChar := len(doc.Lines[lastLine])
    
    // 返回单个 TextEdit 替换整个文档
    return &Message{
        ID: msg.ID,
        Result: []TextEdit{
            {
                Range: Range{
                    Start: Position{Line: 0, Character: 0},
                    End: Position{Line: lastLine, Character: lastChar},
                },
                NewText: formatted,
            },
        },
    }
}
```

**Task 5.2.9: workspace/symbol** (handlers_workspace_symbol.go)

```go
func (s *Server) handleWorkspaceSymbol(msg *Message) *Message {
    var params WorkspaceSymbolParams
    json.Unmarshal(msg.Params, &params)
    
    query := strings.ToLower(params.Query)
    var symbols []SymbolInformation
    
    // 搜索所有打开的文档
    s.docs.mu.RLock()
    defer s.docs.mu.RUnlock()
    
    for _, doc := range s.docs.docs {
        for _, symbol := range doc.Symbols.AllSymbols {
            if strings.Contains(strings.ToLower(symbol.Name), query) {
                symbols = append(symbols, SymbolInformation{
                    Name: symbol.Name,
                    Kind: s.symbolKindToLSP(symbol.Kind),
                    Location: symbol.Location,
                })
            }
        }
    }
    
    return &Message{ID: msg.ID, Result: symbols}
}
```

**Task 5.2.10: 路由新的 LSP 方法** (server.go)

在 `handleMessage` 中添加:
```go
func (s *Server) handleMessage(msg *Message) *Message {
    switch msg.Method {
    // ... 现有方法 ...
    
    case "textDocument/definition":
        return s.handleDefinition(msg)
    case "textDocument/references":
        return s.handleReferences(msg)
    case "textDocument/documentSymbol":
        return s.handleDocumentSymbol(msg)
    case "textDocument/signatureHelp":
        return s.handleSignatureHelp(msg)
    case "textDocument/rename":
        return s.handleRename(msg)
    case "textDocument/codeAction":
        return s.handleCodeAction(msg)
    case "textDocument/formatting":
        return s.handleFormatting(msg)
    case "workspace/symbol":
        return s.handleWorkspaceSymbol(msg)
    
    default:
        // ... 错误处理
    }
}
```

#### 5.3 验证方法

```bash
# 编译测试
go build ./pkg/lsp/...

# 单元测试
go test ./pkg/lsp -v

# VS Code 集成测试
# 1. 安装 VS Code 扩展
# 2. 打开 .klx 文件
# 3. 测试每个功能:
#    - F12 跳转定义
#    - Shift+F12 查找引用
#    - Ctrl+Shift+O 文档大纲
#    - Ctrl+Space 智能补全
#    - 鼠标悬停查看提示
#    - Ctrl+Shift+Space 参数提示
#    - F2 重命名
#    - Shift+Alt+F 格式化
```

**验收标准**:
- ✅ 所有 9 个 LSP 方法都能正常工作
- ✅ 在 VS Code 中测试通过
- ✅ 所有处理器都有单元测试
- ✅ 错误处理完善（null 安全、边界情况）

---

### Phase 6: 单元测试

**目标**: 为核心组件添加测试覆盖

**预计时间**: 4-5 小时  
**优先级**: P0

#### 6.1 测试文件清单

**新增测试文件**:
- `lexer/lexer_test.go`
- `parser/parser_test.go`
- `generator/generator_test.go`
- `pkg/formatter/formatter_test.go`
- `pkg/lsp/symbols_test.go`
- `pkg/lsp/server_test.go`

**新增测试数据**:
- `testdata/simple.klx`
- `testdata/functions.klx`
- `testdata/classes.klx`
- `testdata/modern.klx`

#### 6.2 具体任务

**Task 6.2.1: Lexer 测试** (lexer/lexer_test.go)

```go
package lexer

import (
    "testing"
    "kylix/token"
)

func TestNextToken_BasicTokens(t *testing.T) {
    input := `program test; var x: Integer := 42;`
    l := New(input)
    
    tests := []struct {
        expectedType    token.TokenType
        expectedLiteral string
    }{
        {token.PROGRAM, "program"},
        {token.IDENT, "test"},
        {token.SEMICOLON, ";"},
        {token.VAR, "var"},
        {token.IDENT, "x"},
        {token.COLON, ":"},
        {token.IDENT, "Integer"},
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

func TestNextToken_StringLiterals(t *testing.T) {
    input := `'single' "double"`
    l := New(input)
    
    tok := l.NextToken()
    if tok.Type != token.STRING || tok.Literal != "single" {
        t.Errorf("expected STRING 'single', got %s %s", tok.Type, tok.Literal)
    }
    
    tok = l.NextToken()
    if tok.Type != token.STRING || tok.Literal != "double" {
        t.Errorf("expected STRING \"double\", got %s %s", tok.Type, tok.Literal)
    }
}

func TestNextToken_Comments(t *testing.T) {
    input := `// line comment
var x; // inline comment
{ block comment }
(* another block *)`
    l := New(input)
    
    // 应该跳过所有注释
    tok := l.NextToken()
    if tok.Type != token.VAR {
        t.Errorf("expected VAR, got %s", tok.Type)
    }
}

func TestNextToken_PositionTracking(t *testing.T) {
    input := `var x := 10;
var y := 20;`
    l := New(input)
    
    tok := l.NextToken() // var
    if tok.Line != 1 || tok.Column != 1 {
        t.Errorf("expected line 1, col 1, got line %d, col %d", tok.Line, tok.Column)
    }
    
    tok = l.NextToken() // x
    if tok.Line != 1 || tok.Column != 5 {
        t.Errorf("expected line 1, col 5, got line %d, col %d", tok.Line, tok.Column)
    }
    
    // 跳到第二行
    for i := 0; i < 5; i++ {
        l.NextToken()
    }
    
    tok = l.NextToken() // var (第二行)
    if tok.Line != 2 || tok.Column != 1 {
        t.Errorf("expected line 2, col 1, got line %d, col %d", tok.Line, tok.Column)
    }
}

// ... 更多测试用例
```

**Task 6.2.2: Parser 测试** (parser/parser_test.go)

```go
package parser

import (
    "testing"
    "kylix/ast"
    "kylix/lexer"
)

func TestVarDeclaration(t *testing.T) {
    input := `program test;
var
  x: Integer := 42;
  y: String := 'hello';
begin
end.`
    
    l := lexer.New(input)
    p := New(l)
    program := p.ParseProgram()
    
    if len(p.Errors()) > 0 {
        t.Fatalf("parser has errors: %v", p.Errors())
    }
    
    if len(program.Declarations) != 2 {
        t.Fatalf("expected 2 declarations, got %d", len(program.Declarations))
    }
    
    // 检查第一个声明
    varDecl, ok := program.Declarations[0].(*ast.VarDecl)
    if !ok {
        t.Fatalf("expected VarDecl, got %T", program.Declarations[0])
    }
    
    if varDecl.Names[0] != "x" {
        t.Errorf("expected name 'x', got %s", varDecl.Names[0])
    }
    
    if varDecl.Token.Line != 3 {
        t.Errorf("expected line 3, got %d", varDecl.Token.Line)
    }
}

func TestFunctionDeclaration(t *testing.T) {
    input := `program test;
function Add(a: Integer; b: Integer): Integer;
begin
  result := a + b;
end;
begin
end.`
    
    l := lexer.New(input)
    p := New(l)
    program := p.ParseProgram()
    
    if len(p.Errors()) > 0 {
        t.Fatalf("parser has errors: %v", p.Errors())
    }
    
    funcDecl, ok := program.Declarations[0].(*ast.FunctionDecl)
    if !ok {
        t.Fatalf("expected FunctionDecl, got %T", program.Declarations[0])
    }
    
    if funcDecl.Name != "Add" {
        t.Errorf("expected name 'Add', got %s", funcDecl.Name)
    }
    
    if len(funcDecl.Parameters) != 2 {
        t.Errorf("expected 2 parameters, got %d", len(funcDecl.Parameters))
    }
}

func TestIfStatement(t *testing.T) {
    input := `program test;
begin
  if x > 10 then
  begin
    WriteLn('big');
  end
  else
  begin
    WriteLn('small');
  end;
end.`
    
    l := lexer.New(input)
    p := New(l)
    program := p.ParseProgram()
    
    if len(p.Errors()) > 0 {
        t.Fatalf("parser has errors: %v", p.Errors())
    }
    
    ifStmt, ok := program.Statements[0].(*ast.IfStatement)
    if !ok {
        t.Fatalf("expected IfStatement, got %T", program.Statements[0])
    }
    
    if ifStmt.Alternative == nil {
        t.Error("expected alternative branch")
    }
}

// ... 更多测试用例
```

**Task 6.2.3: Generator 测试** (generator/generator_test.go)

```go
package generator

import (
    "strings"
    "testing"
    "kylix/ast"
    "kylix/lexer"
    "kylix/parser"
)

func TestGenerate_HelloWorld(t *testing.T) {
    input := `program hello;
begin
  WriteLn('Hello, World!');
end.`
    
    l := lexer.New(input)
    p := parser.New(l)
    program := p.ParseProgram()
    
    gen := New()
    output := gen.Generate(program)
    
    // 检查生成的 Go 代码
    if !strings.Contains(output, "package main") {
        t.Error("expected 'package main'")
    }
    
    if !strings.Contains(output, "fmt.Println") {
        t.Error("expected 'fmt.Println'")
    }
    
    if !strings.Contains(output, "Hello, World!") {
        t.Error("expected 'Hello, World!'")
    }
}

func TestGenerate_FunctionDeclaration(t *testing.T) {
    input := `program test;
function Add(a: Integer; b: Integer): Integer;
begin
  result := a + b;
end;
begin
end.`
    
    l := lexer.New(input)
    p := parser.New(l)
    program := p.ParseProgram()
    
    gen := New()
    output := gen.Generate(program)
    
    // 检查函数签名
    if !strings.Contains(output, "func Add(a int64, b int64) int64") {
        t.Error("expected function signature")
    }
    
    // 检查返回值
    if !strings.Contains(output, "return result") {
        t.Error("expected 'return result'")
    }
}

func TestGenerate_TypeMapping(t *testing.T) {
    tests := []struct {
        kylixType string
        goType    string
    }{
        {"Integer", "int64"},
        {"Real", "float64"},
        {"Boolean", "bool"},
        {"String", "string"},
        {"Char", "byte"},
    }
    
    for _, tt := range tests {
        input := fmt.Sprintf(`program test;
var x: %s;
begin
end.`, tt.kylixType)
        
        l := lexer.New(input)
        p := parser.New(l)
        program := p.ParseProgram()
        
        gen := New()
        output := gen.Generate(program)
        
        if !strings.Contains(output, tt.goType) {
            t.Errorf("expected %s for %s, output: %s", tt.goType, tt.kylixType, output)
        }
    }
}

// ... 更多测试用例
```

**Task 6.2.4: Formatter 测试** (pkg/formatter/formatter_test.go)

```go
package formatter

import (
    "strings"
    "testing"
    "kylix/lexer"
    "kylix/parser"
)

func TestFormat_Indentation(t *testing.T) {
    input := `program test;
begin
if x > 0 then
begin
WriteLn('positive');
end;
end.`
    
    l := lexer.New(input)
    p := parser.New(l)
    program := p.ParseProgram()
    
    f := New()
    output := f.Format(program)
    
    // 检查缩进
    lines := strings.Split(output, "\n")
    
    // "begin" 应该在第 0 级缩进
    if !strings.HasPrefix(lines[1], "begin") {
        t.Error("expected 'begin' at indent level 0")
    }
    
    // "if" 应该在第 1 级缩进 (2 空格)
    if !strings.HasPrefix(lines[2], "  if") {
        t.Error("expected 'if' at indent level 1")
    }
    
    // 内部 "begin" 应该在第 2 级缩进 (4 空格)
    if !strings.HasPrefix(lines[3], "    begin") {
        t.Error("expected inner 'begin' at indent level 2")
    }
}

func TestFormat_BlankLines(t *testing.T) {
    input := `program test;
function Add(a: Integer; b: Integer): Integer;
begin
result := a + b;
end;
function Sub(a: Integer; b: Integer): Integer;
begin
result := a - b;
end;
begin
end.`
    
    l := lexer.New(input)
    p := parser.New(l)
    program := p.ParseProgram()
    
    f := New()
    output := f.Format(program)
    
    // 函数之间应该有空行
    if !strings.Contains(output, "end;\n\nfunction") {
        t.Error("expected blank line between functions")
    }
}

func TestFormat_OperatorSpacing(t *testing.T) {
    input := `program test;
begin
x:=10+20*3;
end.`
    
    l := lexer.New(input)
    p := parser.New(l)
    program := p.ParseProgram()
    
    f := New()
    output := f.Format(program)
    
    // 操作符周围应该有空格
    if !strings.Contains(output, "x := 10 + 20 * 3") {
        t.Error("expected spaces around operators")
    }
}

func TestFormat_RoundTrip(t *testing.T) {
    input := `program test;
var
  x: Integer := 42;
  y: String := 'hello';

function Add(a: Integer; b: Integer): Integer;
begin
  result := a + b;
end;

begin
  var z := Add(x, 10);
  WriteLn(z);
end.`
    
    // 第一次格式化
    l := lexer.New(input)
    p := parser.New(l)
    program := p.ParseProgram()
    f := New()
    output1 := f.Format(program)
    
    // 第二次格式化
    l2 := lexer.New(output1)
    p2 := parser.New(l2)
    program2 := p2.ParseProgram()
    f2 := New()
    output2 := f2.Format(program2)
    
    // 两次输出应该相同
    if output1 != output2 {
        t.Error("round-trip formatting produced different results")
        t.Logf("First:\n%s", output1)
        t.Logf("Second:\n%s", output2)
    }
}

// ... 更多测试用例
```

**Task 6.2.5: LSP 符号测试** (pkg/lsp/symbols_test.go)

```go
package lsp

import (
    "testing"
    "kylix/lexer"
    "kylix/parser"
)

func TestCollectSymbols_Variables(t *testing.T) {
    input := `program test;
var
  x: Integer := 42;
  y: String := 'hello';
begin
end.`
    
    l := lexer.New(input)
    p := parser.New(l)
    program := p.ParseProgram()
    
    symbols := CollectSymbols(program, "file:///test.klx")
    
    if len(symbols.AllSymbols) != 2 {
        t.Fatalf("expected 2 symbols, got %d", len(symbols.AllSymbols))
    }
    
    // 查找 x
    xSymbol := findSymbolByName(symbols, "x")
    if xSymbol == nil {
        t.Fatal("symbol 'x' not found")
    }
    
    if xSymbol.Kind != SymbolVariable {
        t.Errorf("expected SymbolVariable, got %d", xSymbol.Kind)
    }
    
    if xSymbol.Detail != "Integer" {
        t.Errorf("expected detail 'Integer', got %s", xSymbol.Detail)
    }
}

func TestCollectSymbols_Functions(t *testing.T) {
    input := `program test;
function Add(a: Integer; b: Integer): Integer;
begin
  result := a + b;
end;
begin
end.`
    
    l := lexer.New(input)
    p := parser.New(l)
    program := p.ParseProgram()
    
    symbols := CollectSymbols(program, "file:///test.klx")
    
    // 应该有 Add 函数和两个参数
    if len(symbols.AllSymbols) < 3 {
        t.Fatalf("expected at least 3 symbols, got %d", len(symbols.AllSymbols))
    }
    
    addSymbol := findSymbolByName(symbols, "Add")
    if addSymbol == nil {
        t.Fatal("symbol 'Add' not found")
    }
    
    if addSymbol.Kind != SymbolFunction {
        t.Errorf("expected SymbolFunction, got %d", addSymbol.Kind)
    }
}

func TestCollectSymbols_Classes(t *testing.T) {
    input := `program test;
class Animal
var
  Name: String;
  Age: Integer;
procedure Speak;
begin
end;
begin
end.`
    
    l := lexer.New(input)
    p := parser.New(l)
    program := p.ParseProgram()
    
    symbols := CollectSymbols(program, "file:///test.klx")
    
    animalSymbol := findSymbolByName(symbols, "Animal")
    if animalSymbol == nil {
        t.Fatal("symbol 'Animal' not found")
    }
    
    if animalSymbol.Kind != SymbolClass {
        t.Errorf("expected SymbolClass, got %d", animalSymbol.Kind)
    }
    
    // 检查子符号
    if len(animalSymbol.Children) != 3 {
        t.Errorf("expected 3 children (2 fields + 1 method), got %d", len(animalSymbol.Children))
    }
}

func findSymbolByName(table *SymbolTable, name string) *Symbol {
    for _, symbol := range table.AllSymbols {
        if symbol.Name == name {
            return symbol
        }
    }
    return nil
}

// ... 更多测试用例
```

**Task 6.2.6: LSP 服务器测试** (pkg/lsp/server_test.go)

```go
package lsp

import (
    "bytes"
    "encoding/json"
    "testing"
)

func TestHandleInitialize(t *testing.T) {
    server := New(nil, nil)
    
    msg := &Message{
        ID: intPtr(1),
        Method: "initialize",
    }
    
    response := server.handleInitialize(msg)
    
    if response.ID == nil || *response.ID != 1 {
        t.Error("expected response ID 1")
    }
    
    // 检查 capabilities
    result, ok := response.Result.(map[string]interface{})
    if !ok {
        t.Fatal("expected map result")
    }
    
    capabilities, ok := result["capabilities"].(map[string]interface{})
    if !ok {
        t.Fatal("expected capabilities map")
    }
    
    // 检查关键功能
    requiredCapabilities := []string{
        "definitionProvider",
        "referencesProvider",
        "documentSymbolProvider",
        "completionProvider",
        "hoverProvider",
    }
    
    for _, cap := range requiredCapabilities {
        if _, ok := capabilities[cap]; !ok {
            t.Errorf("missing capability: %s", cap)
        }
    }
}

func TestHandleDidOpen(t *testing.T) {
    var output bytes.Buffer
    server := New(nil, &output)
    
    params := DidOpenParams{
        TextDocument: TextDocumentItem{
            URI: "file:///test.klx",
            Text: `program test;
var x: Integer := 42;
begin
end.`,
        },
    }
    
    paramsJSON, _ := json.Marshal(params)
    msg := &Message{
        Method: "textDocument/didOpen",
        Params: paramsJSON,
    }
    
    server.handleDidOpen(msg)
    
    // 检查文档是否被缓存
    doc := server.docs.Get("file:///test.klx")
    if doc == nil {
        t.Fatal("document not cached")
    }
    
    if doc.AST == nil {
        t.Error("AST not parsed")
    }
    
    if doc.Symbols == nil {
        t.Error("symbols not collected")
    }
}

func TestHandleDefinition(t *testing.T) {
    var output bytes.Buffer
    server := New(nil, &output)
    
    // 先打开文档
    params := DidOpenParams{
        TextDocument: TextDocumentItem{
            URI: "file:///test.klx",
            Text: `program test;
var x: Integer := 42;
begin
  WriteLn(x);
end.`,
        },
    }
    
    paramsJSON, _ := json.Marshal(params)
    server.handleDidOpen(&Message{
        Method: "textDocument/didOpen",
        Params: paramsJSON,
    })
    
    // 测试跳转定义
    defParams := DefinitionParams{
        TextDocument: TextDocumentIdentifier{URI: "file:///test.klx"},
        Position: Position{Line: 3, Character: 10}, // "x" in WriteLn(x)
    }
    
    defParamsJSON, _ := json.Marshal(defParams)
    response := server.handleDefinition(&Message{
        ID: intPtr(2),
        Method: "textDocument/definition",
        Params: defParamsJSON,
    })
    
    if response.Result == nil {
        t.Fatal("expected definition result")
    }
    
    location, ok := response.Result.(Location)
    if !ok {
        t.Fatal("expected Location result")
    }
    
    // 应该跳转到 var x 的声明
    if location.Range.Start.Line != 1 {
        t.Errorf("expected line 1, got %d", location.Range.Start.Line)
    }
}

func intPtr(i int) *int {
    return &i
}

// ... 更多测试用例
```

#### 6.3 验证方法

```bash
# 运行所有测试
go test ./... -v

# 运行特定包测试
go test ./lexer -v
go test ./parser -v
go test ./generator -v
go test ./pkg/formatter -v
go test ./pkg/lsp -v

# 运行覆盖率报告
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# 检查测试覆盖目标
# lexer: > 90%
# parser: > 85%
# generator: > 80%
# formatter: > 85%
# lsp: > 75%
```

**验收标准**:
- ✅ 所有包都有测试文件
- ✅ 总体测试覆盖率 > 75%
- ✅ 所有关键路径都有测试
- ✅ 测试通过 `go test ./...`
- ✅ 测试运行时间 < 30 秒

---

### Phase 7: 文档更新

**目标**: 详细记录所有新功能

**预计时间**: 2-3 小时  
**优先级**: P1

#### 7.1 修改文件清单

**主要文档**:
- `docs/KYLIX_TOOLS_EXPLAINED.md` - 添加新功能说明
- `docs/KYLIX_IDE_USER_MANUAL.md` - 添加 LSP 功能使用指南
- `README.md` - 更新功能列表
- `SUMMARY.md` - 更新中文文档

#### 7.2 具体任务

**Task 7.2.1: Formatter 新功能文档**

在 `KYLIX_TOOLS_EXPLAINED.md` 中添加:

```markdown
## kylix fmt 详细功能

### 支持的语句类型

kylix fmt 支持格式化所有 Kylix 语句类型：

#### 基本语句
- 变量声明 (`var`)
- 常量声明 (`const`)
- 类型声明 (`type`)
- 赋值语句 (`:=`)
- 表达式语句

#### 控制流语句
- `if`/`then`/`else`
- `while`/`do`
- `for`/`to`/`downto`
- `for`/`in` (foreach)
- `repeat`/`until`
- `case`/`of`
- `match` (模式匹配)

#### 异常处理
- `try`/`except`/`finally`
- `raise`

#### 其他
- `return`
- `break`
- `continue`

### 格式化规则

#### 缩进
- 使用 2 个空格缩进
- `begin`/`end` 块内容缩进
- 类成员缩进

#### 空行
- 函数之间添加空行
- 类型声明之间添加空行
- 主程序块前添加空行

#### 操作符空格
- 二元操作符前后添加空格: `x := 10 + 20`
- 一元操作符后不添加空格: `-x`, `not flag`
- 逗号后添加空格: `Add(a, b, c)`
- 冒号后添加空格: `x: Integer`

#### 示例

**格式化前**:
```pascal
program test;
begin
var x:=10;
if x>5 then
begin
WriteLn('big');
end;
end.
```

**格式化后**:
```pascal
program test;

begin
  var x := 10;
  if x > 5 then
  begin
    WriteLn('big');
  end;
end.
```
```

**Task 7.2.2: REPL 新功能文档**

在 `KYLIX_TOOLS_EXPLAINED.md` 中添加:

```markdown
## kylix repl 详细功能

### 元命令

#### :help, :h, :?
显示帮助信息，列出所有可用命令。

#### :quit, :q, :exit
优雅退出 REPL，保存历史记录。

#### :clear, :c
清屏。

#### :history
显示命令历史，包括编号和内容。

```
kylix> :history
  1: var x := 10
  2: var y := 20
  3: WriteLn(x + y)
```

#### :decls, :declarations
显示当前会话中累积的所有声明。

```
kylix> :decls

Accumulated Declarations:
var x := 10
var y := 20
function Add(a, b: Integer): Integer;
```

#### :reset
清空所有累积的声明，重新开始。

#### :save <filename>
保存当前会话到文件。

```
kylix> :save session.klx
✓ Saved to session.klx
```

生成的文件包含完整的 Kylix 程序结构。

#### :version, :v
显示 Kylix REPL 版本信息。

### 多行输入

REPL 自动检测多行代码：

```
kylix> function Add(a, b: Integer): Integer;
...    begin
...      result := a + b;
...    end;
✓ Declaration added
```

使用 begin/end 配对检测，智能处理嵌套。

### 声明累积

变量、函数、类型声明会在会话中累积：

```
kylix> var x := 10
kylix> var y := 20
kylix> WriteLn(x + y)
30
```

每次执行时，所有声明都会包含在生成的程序中。

### 方向键历史导航

- **上箭头**: 浏览上一条命令
- **下箭头**: 浏览下一条命令
- **左/右箭头**: 编辑当前行
- **Home/End**: 移动到行首/行尾
- **Ctrl+A/E**: 移动到行首/行尾
- **Ctrl+K**: 删除到行尾
```

**Task 7.2.3: LSP 功能完整文档**

在 `KYLIX_IDE_USER_MANUAL.md` 中添加:

```markdown
## LSP 功能使用指南

### 安装 VS Code 扩展

1. 打开 VS Code
2. 按 `Ctrl+Shift+X` 打开扩展面板
3. 搜索 "Kylix"
4. 点击安装

或从源码安装：

```bash
cd vscode-ext
npm install
npm run compile
# 在 VS Code 中按 F5 启动扩展开发
```

### 功能详解

#### 跳转定义 (Go to Definition)

**快捷键**: `F12` 或 `Ctrl+Click`

跳转到变量、函数、类的定义位置。

```pascal
function Add(a, b: Integer): Integer;
begin
  result := a + b;
end;

begin
  var sum := Add(10, 20);  // Ctrl+Click "Add" 跳转到上面
end.
```

#### 查找引用 (Find All References)

**快捷键**: `Shift+F12`

查找符号的所有使用位置。

```pascal
var counter := 0;  // Shift+F12 显示所有使用 counter 的位置

begin
  counter := counter + 1;
  WriteLn(counter);
end.
```

#### 文档大纲 (Document Outline)

**快捷键**: `Ctrl+Shift+O`

显示当前文件的所有符号（变量、函数、类等）。

支持按类型筛选和搜索。

#### 智能补全 (IntelliSense)

**快捷键**: `Ctrl+Space`

提供上下文感知的代码补全：

- **关键字补全**: `begin`, `end`, `if`, `while` 等
- **内置函数**: `WriteLn`, `ReadLn`, `Length` 等
- **用户定义符号**: 当前文件中的所有变量、函数、类
- **类型补全**: `Integer`, `String`, `Boolean` 等

#### 悬停提示 (Hover)

将鼠标悬停在符号上显示详细信息：

- **变量**: 显示类型和初始值
- **函数**: 显示完整签名和返回类型
- **类**: 显示继承关系和成员列表

#### 参数提示 (Signature Help)

**快捷键**: `Ctrl+Shift+Space`

在函数调用时显示参数信息：

```pascal
function Add(a: Integer; b: Integer): Integer;

begin
  Add(  // 显示: Add(a: Integer; b: Integer): Integer
```

当前参数会高亮显示。

#### 重命名 (Rename Symbol)

**快捷键**: `F2`

重命名符号，自动更新所有引用：

```pascal
var oldName := 10;  // F2 输入 newName
begin
  WriteLn(oldName);  // 自动更新为 newName
end.
```

#### 代码格式化 (Format Document)

**快捷键**: `Shift+Alt+F`

自动格式化整个文档，应用一致的代码风格。

#### 实时诊断 (Diagnostics)

自动检测语法错误，显示红色波浪线：

```pascal
var x := 10
begin  // 错误: 缺少分号
  WriteLn(x);
end.
```

悬停显示详细错误信息。

### 配置选项

在 VS Code 设置中搜索 "kylix"：

- `kylix.compiler.path`: Kylix 编译器路径（默认: `kylix`）
- `kylix.lsp.enabled`: 启用/禁用 LSP（默认: `true`）
- `kylix.format.indentSize`: 缩进大小（默认: `2`）

### 故障排除

#### LSP 未启动

1. 检查 Kylix 编译器是否已安装并在 PATH 中
2. 在 VS Code 输出面板查看 "Kylix Language Server" 日志
3. 尝试重启 VS Code

#### 补全不工作

1. 确保文件扩展名是 `.klx`
2. 检查文件是否有语法错误
3. 尝试手动触发补全 (`Ctrl+Space`)

#### 格式化不工作

1. 确保文件可以成功解析（无语法错误）
2. 检查是否设置了默认格式化器
3. 在 settings.json 中添加:
   ```json
   "[kylix]": {
     "editor.defaultFormatter": "kylix.kylix-vscode"
   }
   ```
```

**Task 7.2.4: README.md 更新**

更新功能列表：

```markdown
## Features

### Core Pascal Features
- ✅ Strong typing with type inference
- ✅ Procedures and functions
- ✅ Control structures (if, while, for, case, repeat)
- ✅ Records and arrays
- ✅ Exception handling

### Modern Additions
- ✅ **Type Inference**: `var x := 42;`
- ✅ **Lambda Expressions**: `var square = (x: Integer) -> x * x;`
- ✅ **Generics**: `TList<T>`
- ✅ **Async/Await**: `async function FetchData(): String;`
- ✅ **Pattern Matching**: `match value { 0 => 'zero', _ => 'other' }`
- ✅ **Classes & Interfaces**: Object-oriented programming support
- ✅ **Properties**: With getters and setters
- ✅ **ForEach Loops**: `for item in collection do`
- ✅ **String Interpolation**: `'Hello, ${name}!'`
- ✅ **Modern Exception Handling**: try/except/finally

### IDE Tools
- ✅ **kylix fmt**: Code formatter with consistent style
- ✅ **kylix repl**: Interactive REPL with history and meta-commands
- ✅ **kylix lsp**: Full LSP server with:
  - Go to Definition (F12)
  - Find All References (Shift+F12)
  - Document Outline (Ctrl+Shift+O)
  - IntelliSense Completion (Ctrl+Space)
  - Hover Information
  - Signature Help (Ctrl+Shift+Space)
  - Rename Symbol (F2)
  - Code Formatting (Shift+Alt+F)
  - Real-time Diagnostics
- ✅ **VS Code Extension**: Syntax highlighting and LSP integration

### Test Coverage
- ✅ Comprehensive unit tests for all core components
- ✅ > 75% overall test coverage
- ✅ Integration tests for LSP features
```

**Task 7.2.5: SUMMARY.md 更新**

更新中文文档，添加：

```markdown
## Phase 2 完成状态

### 已完成的功能

#### 1. 代码格式化器 (kylix fmt)
- ✅ 支持所有语句类型（包括 match, raise, property）
- ✅ 智能缩进（2 空格）
- ✅ 自动空行
- ✅ 操作符空格
- ✅ 往返测试（格式化 → 解析 → 格式化 = 相同结果）

#### 2. 交互式 REPL (kylix repl)
- ✅ 彩色输出（ANSI 颜色）
- ✅ 元命令（:help, :quit, :clear, :history, :decls, :reset, :save, :version）
- ✅ 基于 AST 的多行检测
- ✅ 方向键历史导航（readline 支持）
- ✅ 声明累积（var, function, type 在会话中保持）
- ✅ 优雅退出
- ✅ stderr 和 stdout 分离

#### 3. LSP 服务器 (kylix lsp)
- ✅ 符号表系统（收集所有声明和作用域）
- ✅ 文档状态缓存（AST + 符号 + 诊断）
- ✅ 跳转定义 (textDocument/definition)
- ✅ 查找引用 (textDocument/references)
- ✅ 文档大纲 (textDocument/documentSymbol)
- ✅ 智能补全 (textDocument/completion)
- ✅ 悬停提示 (textDocument/hover)
- ✅ 参数提示 (textDocument/signatureHelp)
- ✅ 重命名 (textDocument/rename)
- ✅ 代码操作 (textDocument/codeAction)
- ✅ 格式化 (textDocument/formatting)
- ✅ 工作区符号搜索 (workspace/symbol)

#### 4. VS Code 扩展
- ✅ 语法高亮
- ✅ 语言配置（括号、注释、折叠）
- ✅ LSP 客户端集成

#### 5. 测试覆盖
- ✅ Lexer 测试（> 90% 覆盖率）
- ✅ Parser 测试（> 85% 覆盖率）
- ✅ Generator 测试（> 80% 覆盖率）
- ✅ Formatter 测试（> 85% 覆盖率）
- ✅ LSP 测试（> 75% 覆盖率）
- ✅ 总计 > 75% 覆盖率

### 性能指标

- 编译速度: ~100ms（小型程序）
- LSP 响应时间: < 50ms
- 格式化速度: ~50ms（1000 行文件）
- 测试运行时间: < 30 秒

### 已知限制

1. StringInterpolation 格式化需要 Lexer 增强
2. Visibility 修饰符需要 AST 扩展
3. 跨文件符号查找尚未实现
4. 代码操作（重构）功能基础
```

#### 7.3 验证方法

```bash
# 检查文档完整性
ls -lh docs/

# 检查 Markdown 语法
# 可以使用在线工具或 VS Code 插件

# 检查链接有效性
grep -r "http" docs/ | grep -v "github.com"
```

**验收标准**:
- ✅ 所有新功能都有详细说明
- ✅ 包含使用示例
- ✅ 中英文文档同步更新
- ✅ 所有链接有效
- ✅ 文档结构清晰，易于导航

---

## 时间估算与里程碑

### 总时间估算

| Phase | 任务 | 预计时间 | 优先级 | 状态 |
|-------|------|----------|--------|------|
| 1 | AST 位置追踪 | 3-4 小时 | P0 | ⚠️ 进行中 |
| 2 | Formatter 完善 | 2-3 小时 | P1 | ❌ 待开始 |
| 3 | REPL 稳定性 | 2-3 小时 | P1 | ❌ 待开始 |
| 4 | LSP 符号表 | 4-5 小时 | P0 | ❌ 待开始 |
| 5 | LSP 功能实现 | 6-8 小时 | P0 | ❌ 待开始 |
| 6 | 单元测试 | 4-5 小时 | P0 | ❌ 待开始 |
| 7 | 文档更新 | 2-3 小时 | P1 | ❌ 待开始 |
| **总计** | | **23-31 小时** | | |

### 里程碑

**Milestone 1**: AST 位置追踪完成 (Phase 1)
- 所有 AST 节点都有位置信息
- 解析器正确捕获位置
- 现有示例仍能运行

**Milestone 2**: 工具完善 (Phase 2 + 3)
- Formatter 支持所有节点类型
- REPL 稳定且功能丰富
- 所有示例文件都能正确格式化

**Milestone 3**: LSP 核心功能 (Phase 4 + 5)
- 符号表系统完成
- 所有 9 个 LSP 方法工作正常
- 在 VS Code 中测试通过

**Milestone 4**: 测试和文档 (Phase 6 + 7)
- 测试覆盖率 > 75%
- 所有文档更新完成
- 项目可以发布 v0.3.0

---

## 风险与应对

### 风险 1: AST 位置追踪工作量大

**影响**: 可能超出预计时间

**应对**:
- 优先处理关键节点（声明、语句）
- 表达式节点可以分批添加
- 先实现基本功能，后续迭代完善

### 风险 2: LSP 功能复杂度高

**影响**: 某些功能可能需要更多时间

**应对**:
- 按优先级实现（definition > references > others）
- 可以先实现简化版本
- 复杂功能（如 codeAction）可以后续迭代

### 风险 3: 测试覆盖难以达标

**影响**: 可能无法达到 75% 覆盖率

**应对**:
- 优先测试关键路径
- 使用表驱动测试减少重复代码
- 可以调整目标为 70%

---

## 依赖关系

```
Phase 1 (AST 位置) 
    ↓
Phase 4 (LSP 符号表) ← Phase 2 (Formatter)
    ↓                    ↑
Phase 5 (LSP 功能)      |
    ↓                    |
Phase 6 (测试) ← Phase 3 (REPL)
    ↓
Phase 7 (文档)
```

**关键路径**: Phase 1 → Phase 4 → Phase 5 → Phase 6

**并行任务**:
- Phase 2 和 Phase 3 可以并行
- Phase 6 的部分测试可以提前编写

---

## 验收标准

### 总体标准

1. ✅ 所有 Phase 的任务都完成
2. ✅ 所有测试通过 (`go test ./...`)
3. ✅ 测试覆盖率 > 75%
4. ✅ 所有示例文件都能正确编译和运行
5. ✅ 在 VS Code 中测试所有 LSP 功能
6. ✅ 文档完整且准确

### 功能验收

#### Formatter
- [ ] 所有 examples/*.klx 都能正确格式化
- [ ] 往返测试通过
- [ ] 格式化输出符合 Pascal 风格

#### REPL
- [ ] 多行检测准确
- [ ] 所有元命令工作正常
- [ ] 方向键历史导航可用
- [ ] :save 命令可以保存文件

#### LSP
- [ ] 跳转定义 (F12) 工作正常
- [ ] 查找引用 (Shift+F12) 工作正常
- [ ] 文档大纲 (Ctrl+Shift+O) 显示所有符号
- [ ] 智能补全 (Ctrl+Space) 包含用户符号
- [ ] 悬停提示显示类型信息
- [ ] 参数提示 (Ctrl+Shift+Space) 工作正常
- [ ] 重命名 (F2) 更新所有引用
- [ ] 格式化 (Shift+Alt+F) 工作正常
- [ ] 实时诊断显示语法错误

---

## 总结

本计划详细规划了 Kylix Phase 2 的所有增强工作，涵盖：

1. **AST 位置追踪** - 为 LSP 功能奠定基础
2. **Formatter 完善** - 支持所有 AST 节点类型
3. **REPL 稳定性** - 改进多行检测和用户体验
4. **LSP 符号表** - 构建符号收集和缓存系统
5. **LSP 功能实现** - 实现 9 个完整的 LSP 方法
6. **单元测试** - 达到 75%+ 测试覆盖率
7. **文档更新** - 详细记录所有新功能

**总预计工作量**: 23-31 小时

**预期成果**: Kylix 编译器达到 v0.3.0 版本，具备完整的 IDE 工具链和编辑器集成能力。

---

**下一步**: 请审核本计划，确认后我将按顺序开始实施。如有任何修改意见或优先级调整，请告知。
