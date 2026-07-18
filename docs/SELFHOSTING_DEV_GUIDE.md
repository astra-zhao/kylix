# 自举编译器开发指南 (Self-Hosting Development Guide)

> 版本: v5.2.0 (2026-07-18)
> 关联: [KYLIX_DEV_GUIDE.md](KYLIX_DEV_GUIDE.md) · [ROADMAP.md](../ROADMAP.md) · [TECHNICAL_DEBT.md](../TECHNICAL_DEBT.md)

本指南详细记录 Kylix 自举编译器的设计、构建流程、当前状态与后续开发方法。这是仓库中**最权威**的自举工作参考——动手改 `src/*.klx` 或 Go 后端多态 codegen 前，请先读本文档。

---

## 1. 什么是「自举编译器」

自举（self-hosting）= 用 Kylix 自己的 Pascal 方言重写 Kylix 编译器，然后用现有 Go 后端把它转译成一个**可运行的 `kylix_self` 二进制**。理想终态：`kylix_self` 能像宿主 `kylix` 一样编译任意 `.klx` 程序（round-trip），即「编译器能编译自己」。

三层概念，务必区分：

| 概念 | 说明 | v5.2.0 状态 |
|------|------|-------------|
| **宿主编译器** | 现有 Go 写的编译器（`cmd/kylix` + `generator/` + `parser/` + `ast/`...），构建产物 `/tmp/kylix_bin` | ✅ 生产可用 |
| **自举源码** | `src/*.klx`（7 文件、5250 行），用 Kylix 方言重写的编译器 | ✅ 源码完成 |
| **自举产物** | 自举源码经 Go 后端转译 → 合并 main.go → `go build` → `kylix_self` 二进制 | ✅ 构建打通（v5.2.0）|
| **round-trip** | `kylix_self` 产出的编译器（`kylix_self2`）能正确编译任意程序 | ❌ 未达成（留 v5.3）|

> **关键**：v5.2.0 达成的是「自举源码 → 可运行二进制」的**构建**打通。完整 round-trip 是 v5.3 目标——`kylix_self2` 当前构建成功但运行产出空（见 §7）。

---

## 2. 自举源码结构 (`src/*.klx`)

7 个单元文件，按依赖顺序：

| 文件 | 行数 | 职责 |
|------|------|------|
| `token.klx` | 214 | Token 类型枚举 + 关键字表（`Keywords` map） |
| `error.klx` | 91 | `TDiagnostic`/`TErrorList` 错误收集与格式化 |
| `ast.klx` | 375 | AST 节点类层次（60 个 class，三层继承） |
| `lexer.klx` | 366 | `TLexer` 词法分析器 + `NewLexer` 工厂 |
| `parser.klx` | 2423 | `TParser` Pratt 解析器 + `NewParser` 工厂（最大的文件） |
| `generator.klx` | 1702 | `TGenerator` Go 代码生成器（第二大的文件） |
| `main.klx` | 79 | 入口：ReadFile→Lex→Parse→GenerateMulti→WriteLn |
| `kylix.toml` | — | 项目配置（`name=kylix-compiler`，`main=main.klx`） |

### 2.1 类层次（`ast.klx` 核心）

```
TNode  (ast.klx:9)              ← 根基类，无字段无方法（空壳）
├── TStatement = class(TNode)   (ast.klx:12)  ← 二级基类，空壳
│   ├── TVarDecl, TConstDecl, TTypeDecl, TFunctionDecl
│   ├── TClassDecl, TInterfaceDecl, TPropertyDecl
│   ├── TBlockStatement, TAssignmentStatement, TReturnStatement
│   ├── TIfStatement, TWhileStatement, TForStatement, TForEachStatement
│   ├── TRepeatStatement, TCaseStatement, TMatchStatement
│   └── TTryStatement, TRaiseStatement, TBreakStatement, TContinueStatement, TInheritedStatement
├── TExpression = class(TNode) (ast.klx:15)  ← 二级基类，空壳
│   ├── TIdentifier, TIntegerLiteral, TFloatLiteral, TStringLiteral
│   ├── TStringInterpolation, TBooleanLiteral, TNilLiteral
│   ├── TArrayLiteral, TTupleLiteral, TPrefixExpression, TInfixExpression
│   ├── TCallExpression, TMemberExpression, TIndexExpression, TSliceExpression
│   ├── TLambdaExpression, TAwaitExpression, TTypeCastExpression, TIsExpression
│   └── TRecordType, TArrayType, TMapType, TVariantType, TEnumType, TGenericType
├── TParameter, TTypeParameter, TCaseBranch, TMatchBranch, TVariantCase
└── TProgram (ast.klx:55)      ← 含 Declarations: array of TNode, Statements: array of TStatement
```

### 2.2 多态模式（自举源码的核心 OOP 风格）

自举源码用经典 Pascal OOP 多态——**从不**通过基类变量直接访问字段，一律先 `is`/`as` 下转：

```pascal
// generator.klx:238-253（典型）
decl := prog.Declarations[i];          // decl: TNode（基类）
if decl is TClassDecl then             // is 判断
begin
  cd := decl as TClassDecl;            // as 下转到具体类
  self.GenerateClassDecl(cd);          // 才访问 cd 的字段/方法
end;
```

- `is` / `as` 在 generator.klx 中有 **~95 处**（parser/ast/lexer/error/main 中为零）。最大两处分发链：`GenerateStatement`（:871-1011，~15 路 TStatement）、`GenerateExpression`（:1296-1459，~15 路 TExpression）。
- **三个基类无字段无方法**（ast.klx:9-16 是空壳），且源码从不 `node.Field` / `stmt.Field` 形式直接访问基类字段——这决定了它们可发射成**空 Go interface**（见 §4）。

### 2.3 异构集合（多态容器的根因）

需要基类成为 interface 的根本原因——`array of TBase` 持有不同子类实例：

| 字段 | 位置 | 实际存入 |
|------|------|---------|
| `TProgram.Declarations: array of TNode` | ast.klx:61 | TVarDecl/TConstDecl/TTypeDecl/TFunctionDecl/TClassDecl/TInterfaceDecl |
| `TFunctionDecl.LocalDecls: array of TNode` | ast.klx:97 | TVarDecl/TConstDecl |
| `TClassDecl.Properties: array of TNode` | ast.klx:111 | TPropertyDecl |
| `TProgram.Statements: array of TStatement` | ast.klx:62 | TBlockStatement + 各 statement |
| `TBlockStatement.Statements: array of TStatement` | ast.klx:137 | 全部 TStatement 子类 |
| `TCallExpression.Arguments: array of TExpression` | ast.klx:300 | 各 TExpression 子类 |
| （其余 `array of TExpression`） | 多处 | Values/AdditionalPatterns/ReturnTypes/Parts/Elements/TypeParams |

### 2.4 构造与数组风格

- 所有类用无参 `TClass.Create()` 再逐字段赋值：`prog := TProgram.Create(); prog.UnitName := ...`（parser.klx:173+）。
- 无 `SetLength`，统一 `arr := []` 初始化 + 语句风格 `append(arr, elem)`（mutating，编译器内建）或表达式风格 `self.Errors := append(self.Errors, d)`（error.klx:37）。

---

## 3. 构建流程（端到端）

### 3.1 一键自举构建

```bash
# 1. 构建宿主编译器
go build -o /tmp/kylix_bin ./cmd/kylix/

# 2. 复制自举源码到干净目录（避免污染仓库；src/build/ 是 gitignored 产物）
rm -rf /tmp/kxsrc && cp -r src /tmp/kxsrc && rm -rf /tmp/kxsrc/build /tmp/kxsrc/main.go

# 3. 用宿主编译器转译自举源码（Go 后端，多文件合并）
(cd /tmp/kxsrc && /tmp/kylix_bin build --backend=go *.klx)
#   → 产出 /tmp/kxsrc/main.go（合并的 8000+ 行 Go）

# 4. 编译自举产物
mkdir -p /tmp/kxself_mod && cp /tmp/kxsrc/main.go /tmp/kxself_mod/main.go
printf 'module kylixself\n\ngo 1.21\n' > /tmp/kxself_mod/go.mod
(cd /tmp/kxself_mod && go build -o /tmp/kylix_self ./main.go)
#   → /tmp/kylix_self（~2.9MB）

# 5. 运行自举编译器（无参 → 默认读 6 文件：token/error/ast/lexer/parser/generator）
(cd /tmp/kxsrc && /tmp/kylix_self | head)
#   → stdout 产出 ~5238 行 Go 编译器代码
```

### 3.2 为什么复制到 /tmp

`src/build/` 是 gitignored 的构建产物目录；根目录 `main.go` 也被 `.gitignore`（`:43 /main.go`）忽略——它们若残留在仓库根会污染 `go test ./...` 的根包（v5.2.0 开发期间踩过：根 `main.go` 导致 `kylix` 根包构建失败，误判为回归）。始终在 `/tmp` 副本里生成与构建。

### 3.3 多文件合并机制

`kylix build --backend=go *.klx` 走 `compiler.CompileProject`（pkg/compiler/compiler.go:391），它把 7 个 `.klx` 解析成 7 个 `*ast.Program`，`topoSortWithFiles` 按依赖排序，`generator.GenerateBody` 逐文件生成 body，最后 `gen.BuildOutput(bodies)` 合并成一个 `main.go`（`package main`）。**注意**：这条路径不走 `GenerateMulti`，而是 `CollectClassTypes` + `GenerateBody` + `BuildOutput`——所以多态标志必须在 `CollectClassTypes`（公共预扫描咽喉）里设置，而非 `GenerateMulti`（见 §4）。

---

## 4. Go 后端多态 codegen（v5.2.0 核心）

### 4.1 历史与 opt-in 设计

- **v3.1.0 前**：基类发射成 `interface{}` → 多态可行，但 `var p: TClass; p.Field` 字段不可访问（KLX-C01）。
- **v3.1.0 回退**：所有类发射成普通 struct + 嵌入父 struct，类型一律 `*ClassName` → 字段继承可用，但**无多态**；`classIsBase` map 从此变死代码。
- **v5.2.0 opt-in**：仅当程序含 `is`/`as` 时，把「有子类的基类」发射成空 interface；否则保留 struct 嵌入。自举（含 `is`/`as`）→ interface 通过；教程 example19/example40（继承 + 字段访问、无 `is`/`as`）→ struct 不回归。

### 4.2 标志传播链路

```
parser/parser_expr.go  parseIsExpression/parseAsExpression
        │  p.usesPolymorphism = true
        ▼
parser/parser.go  ParseProgram 末尾
        │  program.UsesPolymorphism = p.usesPolymorphism
        ▼
ast/ast.go  Program struct 新增 UsesPolymorphism bool
        ▼
generator/generator.go  collectClassTypes（所有预扫描路径的公共咽喉）
        │  g.usesPolymorphism = g.usesPolymorphism || program.UsesPolymorphism
        │  （一处覆盖 Generate / GenerateMulti / CompileProject / CollectClassTypes）
        ▼
generator/generator_types.go
        generateClassDecl          — 基类→interface，具体类→struct（父嵌入条件化）
        generateTypeExpression    — 基类→TName，具体类→*TName
        generateTypeExpressionForCast — 基类→TName（断言合法）
```

> **为什么在 `collectClassTypes` 设标志而非 `Generate`/`GenerateMulti`？** 因为多文件 `CompileProject` 走 `CollectClassTypes`+`GenerateBody`+`BuildOutput`，根本不调 `GenerateMulti`。`collectClassTypes` 是所有预扫描路径的唯一入口，在此 OR 标志可一处覆盖全部 codegen 路径。

### 4.3 codegen 规则

**`generateClassDecl`**（generator_types.go:37+）：
```go
if g.usesPolymorphism && g.classIsBase[decl.Name] {
    // 基类（有子类）→ 空 interface。跳过 struct 体与方法循环
    // （interface 不能有 `func (self *TName)` 接收者方法体）。
    // 仍保留 g.classTypes[decl.Name]=true 供类型识别。
    → type TName[TypeParams] interface {}
    return
}
// 具体类 → struct；父嵌入条件化：
if decl.Parent != "" && !(g.usesPolymorphism && g.classIsBase[decl.Parent]) {
    嵌入父 struct    // 父是 interface 时不嵌入（基类无字段可继承）
}
```

**`generateTypeExpression`** 类分支（:524+）：
- `g.usesPolymorphism && g.classIsBase[typeName]` → `TName`（interface，**不带 `*`**）
- `g.classTypes[typeName]` → `*TName`（现状，字段继承）
- else → `mapType(...)`

**`generateTypeExpressionForCast`**（:619+）：基类（poly）→ `TName`，保证 `x.(TBase)` 在 interface 上合法。

### 4.4 适用范围与边界

✅ **支持**：「基类无字段无方法 + 多态靠 `is`/`as`」模式（自举即此）。
❌ **不支持**（留 v5.3）：
- 基类含字段且通过基类变量访问 → 空 interface 上字段不可访问，会崩。需 getter 转发或 vtable。
- 基类有虚方法需分派 → 空 interface 无方法签名。需 interface 方法 + 具体类实现。
- **program-level 标志过宽**：含 `is`/`as` 的程序会把**所有**「有子类的基类」都变 interface。混合程序（部分基类需字段继承、部分需多态）会误伤。需 per-base 检测：仅对实际承载子类实例/作断言操作数的基类发射 interface。

---

## 5. `Args` builtin 与切片协变

### 5.1 `Args`（main.klx 命令行参数）

main.klx 用 `Length(Args)` / `Args[i]` 读命令行参数。v5.2.0 在 `mapBuiltinFunction`（generator_types.go:660 builtinMap）加：
```go
"Args": "os.Args[1:]",
```
- `Args` → `os.Args[1:]`；`Length(Args)` → `len(os.Args[1:])`；`Args[i]` → `os.Args[1:][i]`。
- Go 允许 int64 索引切片（已实测 `go build` exit=0），故自举的 `for i := 0 to Length(Args)-1 do Args[i]` 直接合法。
- import 开关已有 `case strings.HasPrefix(goFunc, "os."): g.imports["os"]=true`（:712），自动注册 `os` 导入。

### 5.2 切片协变修复

`src/parser.klx:448` 原为 `var localDecls: array of TStatement;`，但赋值给 `decl.LocalDecls`（`array of TNode`，ast.klx:97）。**Go 切片不变式**：`[]TStatement` 与 `[]TNode` 即使元素兼容也不可互赋。改为 `array of TNode` 后：localDecls 是 `[]TNode`（interface），append `*TVarDecl`/`*TConstDecl`（实现 TNode）合法，赋值同型合法。

> 注：自举编译器自身的 `MapType`（generator.klx:1654）已硬编码三基类→`interface{}`，所以此源码改动不影响自举输出语义——它只影响宿主编译器转译自举源码时的类型发射。

---

## 6. 测试与验证

### 6.1 单元测试

`generator/generator_polymorphism_test.go`（v5.2.0 新增，3 个测试）：
- `TestPolymorphism_BaseClassBecomesInterface`：含 `is`/`as` 程序 → 基类发射 `interface{}`、`array of TBase`→`[]TBase`、`x.(*TSub)` 断言在 interface 上。
- `TestPolymorphism_NoIsAs_KeepsStructInheritance`：**回归**——无 `is`/`as` 继承程序仍发射 struct + 嵌入 + `*TBase`，字段访问保留（守护 example19/example40）。
- `TestPolymorphism_AsBuiltinArgs`：`Args`→`os.Args[1:]` + `os` 导入。

```bash
go test ./generator/ -run 'TestPolymorphism' -v
go test $(go list ./... | grep -v '/examples')    # 16 包全绿
```

### 6.2 自举构建验收

```bash
# 完整自举构建（§3.1）后，错误数应为 0
(cd /tmp/kxself_mod && go build -gcflags="-e" -o /tmp/kylix_self ./main.go 2>&1 | wc -l)
# 期望: 0（无错误输出）。开发期复现错误时用 -gcflags="-e" 列全量错误
# （Go 默认只显示前 10 个就 "too many errors"，-e 列全部）
```

### 6.3 教程回归（opt-in 不破坏继承教程）

```bash
go build -o /tmp/kylix_bin ./cmd/kylix/
KYLIX=/tmp/kylix_bin bash examples/complete-tutorial/test_all.sh 2>&1 | tail -4
# 期望: Results: 51/51 passed, 0 failed
```

example19_inheritance.klx、example40_declarative_oop.klx 用继承 + 字段访问、**无** `is`/`as` → opt-in 保证它们基类仍是 struct 嵌入 → 字段继承保留 → 不回归。

---

## 7. 当前状态与 v5.3 目标

### 7.1 v5.2.0 已达成

- 自举源码 `src/*.klx` → `go build` **208 个错误 → 0** → `kylix_self`（2.9MB）运行产出 5238 行 Go 编译器代码。
- go test 16 包全绿，教程 51/51 无回归。

### 7.2 v5.3 目标：完整 round-trip

**问题**：`kylix_self` 产出的编译器（`kylix_self2`）构建成功（2.2MB）但运行产出空——自举 `generator.klx` 自身是**简化重实现**，codegen 不完整：
- `generator.klx:206-208` 注释明确：`ClassTypes`/`ClassIsBase` map 未填充，依赖 `MapType` 的 fallback 逻辑（硬编码三基类→`interface{}`，其余→`*TName`）。
- 带参构造 `&TFoo{Field: arg}` 因 `ClassFields` 未填充会退化。
- 其余 codegen 分支（语句/表达式/类型）可能不完整。

**v5.3 工作方向**：补全自举 `generator.klx` 的 codegen，使 `kylix_self2` 输出与宿主 `kylix` **逐字节一致**。验收：`kylix_self2` 能正确编译 `hello.klx` 并产出可运行 Go 代码。

### 7.3 检查清单（改 `src/*.klx` 或多态 codegen 前）

- [ ] 读本文档 §2（源码结构）、§4（opt-in 规则）、§7.2（已知缺口）
- [ ] 在 `/tmp` 副本里构建（§3.2），勿污染仓库根
- [ ] 用 `go build -gcflags="-e"` 复现全量错误
- [ ] 改完跑 §6 三套验证（单元测试 + 自举构建 + 教程回归）
- [ ] 确认无游离 `main.go` 残留在仓库根（会导致根包构建失败）
- [ ] 文件行数 ≤1000（项目约束）

---

## 8. 常见问题

**Q: 自举产物 `kylix_self` 和宿主 `kylix` 有何区别？**
A: 宿主是 Go 写的完整编译器；`kylix_self` 是自举源码（Kylix 方言）经 Go 后端转译出的等价编译器。功能上 `kylix_self` 受限于 `generator.klx` 的 codegen 完整度（v5.2.0 构建通过但 codegen 不完整，见 §7.2）。

**Q: 为什么不直接用 `GenerateMulti` 而要走 `CompileProject`？**
A: `kylix build *.klx` 调 `CompileProject`（增量缓存 + topo 排序），它不调 `GenerateMulti`。这是 v5.2.0 把多态标志放进 `collectClassTypes`（而非 `GenerateMulti`）的原因——`collectClassTypes` 是所有路径的公共咽喉。

**Q: 修改 `src/*.klx` 后怎么验证不影响宿主编译器？**
A: `src/*.klx` 只是被编译的对象，不参与宿主编译器构建。但要跑 §6.1 教程回归——改 Go 后端多态 codegen 时必须确认不破坏无 `is`/`as` 的继承教程。

**Q: round-trip 为什么这么难？**
A: 自举编译器要「能编译自己」，要求 `generator.klx` 的 codegen 与宿主 `generator/` 包**语义完全一致**。v5.2.0 只打通了「宿主能编译自举源码」（单向），反向（自举产物能编译自举源码）需 `generator.klx` codegen 补全——这是 v5.3 的核心工作。

---

## 9. 相关文档

- [KYLIX_DEV_GUIDE.md](KYLIX_DEV_GUIDE.md) — 通用开发指南（架构/工作流程/贡献）
- [llvm-backend.md](llvm-backend.md) — LLVM 后端文档（自举目前用 Go 后端，但 LLVM 后端是 ROADMAP 终态之一）
- [../ROADMAP.md](../ROADMAP.md) — 版本路线图（v5.2.0 自举行）
- [../TECHNICAL_DEBT.md](../TECHNICAL_DEBT.md) — v5.2.0 残留限制详表
- [../CHANGELOG.md](../CHANGELOG.md) — v5.2.0 变更记录
