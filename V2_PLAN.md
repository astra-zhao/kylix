# Kylix v2.0.0 开发计划

> 启动日期: 2026-06-14  
> 目标发布: 2026-07 中旬（预计 4-6 周）  
> Phase: 12  
> 前置版本: v1.5.0+ (Phase 11 技术债务清理完成)

---

## 总体目标

将 Kylix 从「可用」提升到「生产就绪」：
- 错误体验达到现代编译器水平（错误代码、恢复、建议）
- 类型系统完整（推导、泛型约束、别名）
- 工具链自完备（test/doc/bench）
- stdlib 完全 Kylix 化

---

## 任务列表（按优先级排序）

### Milestone 1: 错误体验现代化（1-2 周）

#### 1.1 错误代码系统（3 天）

**目标**：每个错误有唯一代码 + 文档链接。

**当前状态**：
```go
type Diagnostic struct {
    File    string
    Line    int
    Column  int
    Level   string // "error" or "warning"
    Message string
    Source  string
}
```

**目标状态**：
```go
type Diagnostic struct {
    File    string
    Line    int
    Column  int
    Level   string
    Code    string // "KLX001", "KLX002"
    Message string
    Source  string
    Hint    string // optional fix suggestion
}
```

**错误代码分类**：
- `KLX001-099`: 词法/语法错误
- `KLX100-199`: 类型错误
- `KLX200-299`: 语义错误（未声明变量、函数签名不匹配）
- `KLX300-399`: 接口实现错误
- `KLX400-499`: 编译器内部错误

**输出格式**：
```
error[KLX201]: undeclared variable 'userName'
  --> main.klx:10:5
   |
10 |     WriteLn(userName);
   |             ^^^^^^^^ not found in this scope
   |
   = help: did you mean 'username'?
   = note: for more information, see https://kylix-lang.org/errors/KLX201
```

**实现步骤**：
1. 扩展 `Diagnostic` 结构体
2. 定义 `pkg/compiler/errors.go` — 错误代码常量和构造函数
3. 修改所有 `Diagnostic` 创建点，传入错误代码
4. 更新 `printDiagnostics` 输出格式
5. 编写 10+ 常见错误的测试

**验收标准**：
- [ ] `Diagnostic` 包含 `Code` 和 `Hint` 字段
- [ ] 至少 20 个错误代码定义
- [ ] `printDiagnostics` 输出格式符合 rustc 风格
- [ ] 错误代码测试覆盖 100%

---

#### 1.2 错误恢复机制（4 天）

**目标**：部分错误不阻止后续编译，一次报告多个错误。

**当前问题**：第一个错误就终止编译。

**策略**：
- **类型错误**：记录但继续（假设类型为 `Any`）
- **未声明变量**：记录但继续（假设已声明）
- **语法错误**：尝试同步到下一个声明
- **致命错误**：无法恢复的立即终止（文件不存在、循环依赖）

**实现位置**：
- `parser/parser.go`: panic recovery + 同步点
- `pkg/compiler/typecheck.go`: 非致命错误继续
- `pkg/compiler/compiler.go`: 错误累积，最后统一判断

**验收标准**：
- [ ] 一次编译可报告 3+ 个错误
- [ ] parser 语法错误后可恢复到下一个 `function`/`var`/`begin`
- [ ] typecheck 错误不阻止 codegen（生成带占位符的代码）
- [ ] 新增 `TestErrorRecovery_MultipleErrors` 测试

---

#### 1.3 错误建议（Hint）系统（2 天）

**目标**：常见错误给出修复建议。

**场景**：
1. **拼写错误**：`userName` → `did you mean 'username'?`（Levenshtein 距离 ≤ 2）
2. **类型不匹配**：`var x: Integer; x := 'hello'` → `help: use StrToInt() to convert`
3. **参数数量错误**：`Add(1, 2, 3)` 但 `Add(a, b)` → `note: Add expects 2 arguments`
4. **缺少导入**：`Now()` 未找到 → `help: add 'uses datetime;' to use Now()`

**实现**：
- `pkg/compiler/suggestions.go` — 拼写纠错、上下文建议
- `Diagnostic.Hint` 字段填充
- LSP 集成：`textDocument/codeAction` 返回快速修复

**验收标准**：
- [ ] 拼写纠错命中率 > 80%（基于 stdlib 符号表）
- [ ] 4+ 种场景的建议测试
- [ ] LSP 快速修复可触发（手动测试）

---

### Milestone 2: 类型系统完善（1-2 周）

#### 2.1 类型推导增强（5 天）

**当前状态**：只有 `:=` 的右值推导。

**目标**：
- 函数返回值推导：`function Add(a, b: Integer) := a + b;`（省略 `: Integer`）
- 泛型实例化推导：`TBox<Integer>.Create(42)` → `TBox.Create(42)`
- 数组字面量推导：`var nums := [1, 2, 3];` → `array of Integer`

**实现**：
- `pkg/compiler/typeinfer.go` — 类型推导引擎
- AST walker 收集约束
- Hindley-Milner 算法简化版（单态化）

**验收标准**：
- [ ] 函数返回值推导测试 5+
- [ ] 数组字面量推导测试 3+
- [ ] 与现有 typecheck 集成无冲突

---

#### 2.2 泛型约束验证（3 天）

**目标**：`type TComparable<T: IComparable>`

**当前状态**：泛型语法可解析，但无约束检查。

**实现**：
- `ast.TypeParameter` 新增 `Constraint Expression`（已有，需用起来）
- typecheck 验证实例化时 `T` 实现了约束接口
- 错误代码 `KLX301: type 'String' does not implement 'IComparable'`

**验收标准**：
- [ ] `TBox<T: IFoo>` 约束验证
- [ ] 泛型约束测试 5+
- [ ] 错误提示包含缺失的方法列表

---

#### 2.3 类型别名（2 天）

**目标**：`type UserId = Integer;`

**语法**：
```pascal
type
  UserId = Integer;
  UserMap = map[UserId]User;
```

**实现**：
- parser 已支持 `TypeDecl`
- typecheck 中建立别名表：`aliases["UserId"] = "Integer"`
- 类型规范化：使用时解析到底层类型

**验收标准**：
- [ ] 类型别名解析正确
- [ ] 递归别名检测（`type A = B; type B = A;` → 错误）
- [ ] 类型别名测试 3+

---

### Milestone 3: 工具链完善（1 周）

#### 3.1 `kylix test` 命令（3 天）

**目标**：运行测试文件，输出 TAP 格式。

**测试文件约定**：
- `*_test.klx` 文件
- `procedure Test<Name>()` 函数
- `Assert(condition, message)` 内建函数

**实现**：
- `cmd/kylix/cmd_test.go` — 新命令
- `pkg/testrunner/` — 测试发现、执行、报告
- 编译测试文件 → 运行 → 捕获 Assert 失败

**输出格式**：
```
TAP version 14
1..5
ok 1 - TestAdd
ok 2 - TestSubtract
not ok 3 - TestDivide
  ---
  message: "expected 5, got 0"
  file: math_test.klx
  line: 42
  ...
```

**验收标准**：
- [ ] `kylix test` 发现并运行 `*_test.klx`
- [ ] 5+ 测试通过/失败场景
- [ ] TAP 输出可被 CI 工具解析

---

#### 3.2 `kylix doc` 命令（2 天）

**目标**：生成 Markdown 文档。

**输入**：
```pascal
// StringUtils provides string manipulation functions.
unit stringutil;

// Reverse returns the reversed string.
function Reverse(s: String): String;
```

**输出**（`stringutil.md`）：
```markdown
# stringutil

StringUtils provides string manipulation functions.

## Functions

### Reverse

```pascal
function Reverse(s: String): String
```

Reverse returns the reversed string.
```

**实现**：
- `cmd/kylix/cmd_doc.go`
- `pkg/docgen/` — AST → Markdown
- 提取注释（`//` 开头的行作为文档）

**验收标准**：
- [ ] stdlib 4 个模块文档生成测试
- [ ] 支持函数、类型、类的文档

---

#### 3.3 `kylix bench` 命令（1 天）

**目标**：性能基准测试。

**测试文件约定**：
- `procedure Bench<Name>(b: Benchmarker)` — 基准测试函数
- `b.ResetTimer()`, `b.StartTimer()`, `b.StopTimer()`

**实现**：
- `cmd/kylix/cmd_bench.go`
- 运行 N 次，计算平均时间

**输出**：
```
BenchmarkFibonacci    1000000    1234 ns/op
BenchmarkPrimeCheck    500000    2345 ns/op
```

**验收标准**：
- [ ] 基准测试发现和运行
- [ ] 输出格式兼容 Go bench

---

### Milestone 4: stdlib Kylix 化（1-2 周）

#### 4.1 stdlib Go 包装器移除（5 天）

**当前**：`stdlib/*.go` 用 Go 实现，Kylix 调用。

**目标**：用 Kylix 重写核心库。

**优先级**：
1. `sysutil` — 最常用（GetEnv, ReadFile, WriteFile）
2. `datetime` — 日期时间
3. `jsonutil` — JSON 处理
4. `regex` — 正则表达式

**实现策略**：
- Kylix 代码调用 Go stdlib（通过 cgo 或生成的 Go wrapper）
- `stdlib/impl/` 存放 Kylix 实现
- 保留 Go 实现作为 fallback

**验收标准**：
- [ ] `sysutil.klx` → `stdlib/impl/sysutil.klx` 可编译
- [ ] 4 个模块的 Kylix 实现 + 测试

---

### Milestone 5: 性能优化（1 周）

#### 5.1 并行编译（3 天）

**目标**：多文件项目并行编译。

**实现**：
- `CompileProject` 的 parse 阶段并行化
- topoSort 后，无依赖的 unit 并行 codegen
- 使用 `sync.WaitGroup` + worker pool

**验收标准**：
- [ ] 10 文件项目编译时间减少 > 30%
- [ ] 并发安全测试（race detector）

---

#### 5.2 增量链接（2 天）

**目标**：跳过未变更 unit 的 Go 编译。

**实现**：
- 缓存每个 unit 的 `.go` + `.o` 文件
- 检查 mtime，未变则跳过 `go build`

**验收标准**：
- [ ] 二次构建时间 < 10% 首次构建

---

## 发布检查清单

### 代码质量
- [ ] 所有新功能有测试（目标：新增 50+ 测试）
- [ ] 全套测试通过（240+ 测试）
- [ ] `go vet` 和 `golangci-lint` 通过
- [ ] 代码覆盖率 > 70%

### 文档
- [ ] CHANGELOG.md 更新 v2.0.0 条目
- [ ] README.md 更新功能列表
- [ ] 错误代码文档（docs/errors/）
- [ ] stdlib 文档（自动生成 + 人工审核）

### 性能
- [ ] 编译速度基准测试（vs v1.5.0）
- [ ] 内存占用基准测试

### 用户体验
- [ ] 10 个错误场景的端到端测试
- [ ] LSP 快速修复手动测试
- [ ] `kylix test/doc/bench` 示例运行

### 发布
- [ ] Git tag `v2.0.0`
- [ ] GitHub Release 发布说明
- [ ] 二进制构建（Linux/macOS/Windows）

---

## 时间表

| Week | Milestone | 关键产出 |
|------|-----------|----------|
| W1 | M1.1-1.2 | 错误代码系统 + 错误恢复 |
| W2 | M1.3 + M2.1 | 错误建议 + 类型推导 |
| W3 | M2.2-2.3 + M3.1 | 泛型约束 + 类型别名 + kylix test |
| W4 | M3.2-3.3 + M4.1 | kylix doc/bench + stdlib Kylix 化启动 |
| W5 | M4.1 + M5 | stdlib 完成 + 性能优化 |
| W6 | 集成测试 + 发布准备 | 文档、基准测试、发布 |

---

## 风险与缓解

| 风险 | 影响 | 概率 | 缓解 |
|------|------|------|------|
| 错误恢复导致级联错误 | 高 | 中 | 保守策略，充分测试 |
| 类型推导算法复杂度爆炸 | 中 | 低 | 限制推导深度，启发式算法 |
| stdlib Kylix 化性能下降 | 高 | 中 | 保留 Go 实现作为 fallback |
| 并行编译竞态条件 | 高 | 中 | race detector + 单元测试 |

---

## v2.0.0 成功标准

1. **错误体验现代化**：错误代码 + 恢复 + 建议，用户一次编译可看到 3+ 个错误
2. **类型系统完整**：类型推导 + 泛型约束 + 类型别名
3. **工具链自完备**：`kylix test/doc/bench` 可用
4. **性能提升**：编译速度提升 > 30%
5. **测试覆盖**：新增 50+ 测试，总测试数 240+

**最终目标**：Kylix v2.0.0 成为「生产就绪」的现代 Pascal 编译器。
