# Kylix 技术债务与后续开发清单

> 最后更新: 2026-06-23
> 当前版本: v3.1.0
> 关联文档: [ROADMAP.md](ROADMAP.md), [CHANGELOG.md](CHANGELOG.md)

本文档记录 v3.1.0 之后的已知缺陷、功能缺口和工程质量改进项，包含修复状态追踪。

---

## ✅ v3.1.0 修复的编译器缺陷

| ID | 缺陷 | 修复内容 |
|----|------|---------|
| **KLX-C01** | `var p: TClass` 生成 `interface{}` 导致字段不可访问 | `generator/generator_types.go` 始终为类类型 emit `*TypeName` |
| **KLX-C02** | 字符串插值 `${var}` 不展开 | `lexer/lexer.go` 单引号字符串中 `${...}` emit STRING_INTERPOLATION |
| **KLX-C03** | 匿名函数 `function(x): T` 返回类型丢失 | `ast.LambdaExpression.ReturnType` + parser/generator 配套 |
| **KLX-C04** | `match` 语句生成无效 Go 代码 | 改为 tagless `switch { case _v == p: }` |
| **KLX-C05** | `uses sysutil/jsonutil/...` 在 program 中符号不可见 | `generator/generator_stdlib.go` 映射 40+ stdlib 函数 |

详见 CHANGELOG.md v3.1.0 章节。

---

## 优先级 1：正确性缺陷 🔴

### ✅ 1.1 `CompileFile` 未接入增量缓存

**已验证不需要修复。** `CompileFile` 是单文件编译路径，每次都需要重新生成（无法重用 body）。增量编译对多文件项目（`CompileProject`）有效，单文件编译本身就是全量的。

---

### ✅ 1.2 `topoSortWithFiles` 的文件路径对齐

**已验证：实际代码正确。** `progFile[prog] = files[i]` 在 parse 循环中建立，以指针为 key，topo 排序通过 `progFile[prog]` 查找，不存在对齐问题。原分析有误。

---

### ✅ 1.3 `GenerateBody` exception types 输出

**已验证：无 bug。** `BuildOutput` 中通过 `g.needsException` 判断再 snapshot，`GenerateBody` 不调用 `writeExceptionTypes`，多文件编译 exception 输出正确（经 exc_unit.klx + exc_main.klx 端到端验证）。

---

### ✅ 标准库已知缺陷 — v3.0.0-alpha 修复

**`TDateTime` +/- 运算符未实现** → ✅ 已修复（v3.0.0-alpha）
`DateAdd(dt, days)` 和 `DateSub(dt, days)` 在 `stdlib/src/datetime.klx` 中实现，替代运算符重载。

**`jsonutil` 仅支持扁平 JSON** → ✅ 已修复（v3.0.0-alpha）
`stdlib/src/jsonutil.klx` 重写为完整递归下降解析器（TJsonLexer + TJsonParser），支持任意深度嵌套。

**`external` 函数声明在文件末尾解析失败** → ✅ 已修复（v3.0.0-alpha）
`ast/ast.go` 新增 `IsExternal bool`，`parser/parser_decl.go` 识别 `EXTERNAL` 修饰词，`generator/generator_types.go` 跳过 body 生成。

---

## 优先级 6：LLVM 后端已知限制 🟠

这些是 LLVM 后端 Milestone 1 + Phase 1 后剩余的范围外项目。

### ✅ 6.0 数组未支持 → v3.1.0 修复

`pkg/llvmgen/array.go`（~200 行）：
- 静态 `array[1..N] of T` → `alloca [N x T]`
- 动态 `array of T` → `{ ptr, i64, i64 }` slice 结构体
- Pascal 1-based 索引转 LLVM 0-based
- 6 个新测试

### ✅ 6.3 无优化 Pass → v3.1.0 修复

`CompileOpts.OptLevel` + `--llvm-opt=0/1/2/3` CLI 标志；`llc -O=N`。

### 6.1 接口未支持

**影响：** `class X implements IFoo` 无法通过 LLVM 后端编译。

**方案：** fat pointer（数据指针 + vtable 指针），每个接口方法生成 thunk。

**工作量：** 1–2 周（v3.2 Phase 2）

---

### 6.2 泛型无单态化

**影响：** 泛型类/函数（`TBox<T>`）无法通过 LLVM 后端编译。

**方案：** 在 codegen 前对每个具体类型参数执行 AST 克隆 + 替换（单态化）。

**工作量：** 2–3 周（v3.2 Phase 3）

---

### 6.4 不支持异常（try/catch）

**影响：** 含 `try/except/finally` 的程序无法通过 LLVM 后端编译。

**方案：** 使用 LLVM `landingpad` + `invoke` 指令，或映射到 `setjmp/longjmp`（简单方案）。

**工作量：** 1–2 周（v3.2+）

---

### ✅ 2.1 类型检查层完全缺失 → v1.5.0+

**已修复：** `pkg/compiler/typecheck.go` 实现 MVP 类型检查器：
- 未声明变量检测
- 函数调用参数数量检查
- 明显类型不兼容检测（字符串→Integer、整数→String 等）
- 7 个测试，保守策略（只报确定性错误）

---

### ✅ 2.2 `kylix add` 的 git 包逻辑错误 → v1.5.0+

**已修复：** `installGit` 逻辑修正：有 tag 才跳过（版本固定幂等），无 tag 每次重新拉取。

---

### ✅ 2.3 多返回值 TupleLiteral LHS 生成 → v1.5.0+

**已修复：** `generator_stmt.go` 的 `generateAssignment` 新增 TupleLiteral LHS 分支：
- `x, y := Pair()` 正确生成 Go: `x, y := Pair()`

---

### ✅ 2.4 包管理器与编译器未集成 → v1.5.0+

**已修复：**
- `compiler.Options` 新增 `PackageSearchDirs []string`
- `CompileProject` 自动加载 `packages/*/*.klx`
- `cmdBuild` 自动传入 `packageDirsFromWd()`

---

## 优先级 3：测试覆盖空洞 🟡

### ✅ 3.1 pkgmgr + cache 基础测试 → v1.5.0+

| 模块 | 测试文件 | 测试数 |
|------|----------|--------|
| `pkg/pkgmgr/manager.go` | `manager_test.go` | 5 |
| `pkg/compiler/cache.go` | `cache_test.go` | 5 |

---

### ✅ 3.2 parser 泛型/多返回值回归测试 → v1.5.0+

`parser/parser_regression_test.go`: 5 个测试
- `TestParseGenericInstantiation`
- `TestParseGenericTwoParams`
- `TestParseMultiReturnFunction`
- `TestParseMultiReturnAssignment`
- `TestParseTupleReturn`

---

### ✅ 3.3 LSP stdlib 加载测试 → v1.5.0+

`pkg/lsp/stdlib_test.go`: 4 个测试
- `TestStdlibKlxFilesExist`
- `TestLoadStdlibSymbols_Sysutil`
- `TestLoadStdlibSymbols_Datetime`
- `TestLoadStdlibSymbols_NoUses`

---

### ✅ 3.4 generator 多返回值回归测试 → v1.5.0+

`generator/generator_multireturn_test.go`: 4 个测试
- `TestGenerateMultiReturnFunction`
- `TestGenerateMultiReturnCall`
- `TestGenerateTupleReturnStatement`
- `TestGenerateMultiReturnNestedTuple`

---

## 优先级 4：工程质量 🟢

### ✅ 4.1 `cmd/kylix/main.go` 拆分 → v1.5.0+

763 行 → 5 个文件（最大 220 行）：
- `main.go` (159 行)
- `cmd_build.go` (197 行)
- `cmd_run.go` (118 行)
- `cmd_other.go` (220 行)
- `cmd_package.go` (96 行)

---

### ✅ 4.2 `stdlib/klx/*.klx` 可解析性测试 → v1.5.0+

`stdlib/klx_test.go`: `TestKlxDeclarationsAreParseable`
发现并修复了 `jsonutil.klx` 的 `Map<K,V>` → `map[K]V` 语法错误。

---

### ✅ 4.3 `ioutil` 废弃替换 → v1.5.0+

`pkg/compiler/compiler.go` 和 `pkg/project/project.go` 全部替换为 `os.ReadFile`/`os.WriteFile`。

---

## 优先级 5：设计层面的长期债务 ⚪

这些项需要架构重构，适合 Phase 12 处理。

### 5.1 缺少符号解析器（name resolver）

**影响：** LSP 补全只能依赖 stdlib/klx 声明文件，用户自己的 unit 无法跨文件补全。

**长期方案：** `pkg/resolver/` — 建立全局符号表，跨文件符号解析。

**工作量：** 3 周

---

### 5.2 `Generator` 全局状态不可重入

**影响：** 增量编译将来要并行化时，当前 `Generator` 无法并行调用。

**根本原因：** `g.output` 是全局累积状态，`GenerateBody` 通过 snapshot 绕开。

**长期方案：** 拆分为全局状态（类型表、imports）和 per-unit 状态（当前输出）。

**工作量：** 1 周

---

### 5.3 错误位置信息从字符串反向解析

**影响：** `parseLocation(errMsg)` 用 `fmt.Sscanf` 从错误字符串提取行列号，脆弱。

**长期方案：** `Diagnostic` 直接传递 `token.Token`，而非序列化再解析。

**工作量：** 2 天

---

## 当前状态总结（2026-06-23）

### 新增已知缺陷（v3.1.0 引入或残留）

| ID | 问题 | 严重度 | 目标 |
|----|------|--------|------|
| **KLX-G01** | `example21_generic_class` 编译通过但运行时异常（泛型实例化路径）| 中 | v3.2 |
| **KLX-M01** | `example33_use_module` 多文件 unit 编译路径在某些场景失败 | 中 | v3.2 |

### Phase 11 完成度（v1.5.0–v2.0.0）

| 优先级 | 项数 | 已完成 | 完成率 |
|--------|------|--------|--------|
| P1（正确性缺陷） | 3 | 3（已验证或确认无 bug） | 100% |
| P2（功能缺口） | 4 | 4 | 100% |
| P3（测试覆盖） | 4 | 4 | 100% |
| P4（工程质量） | 3 | 3 | 100% |
| P5（设计债务） | 3 | 0 | 0%（长期）|

### v3.0.0-alpha 新增修复

| 项目 | 状态 |
|------|------|
| TDateTime +/- 运算符 | ✅ DateAdd/DateSub |
| jsonutil 仅支持扁平 JSON | ✅ 嵌套解析器 |
| external 函数解析失败 | ✅ IsExternal 字段 |

### v3.1.0 新增修复

| 项目 | 状态 |
|------|------|
| KLX-C01 `var p: TClass` 字段访问 | ✅ 生成 `*TClass` |
| KLX-C02 字符串插值 | ✅ STRING_INTERPOLATION token |
| KLX-C03 lambda 返回类型 | ✅ ReturnType 字段 |
| KLX-C04 match codegen | ✅ tagless switch |
| KLX-C05 uses 符号注入 | ✅ generator_stdlib.go |
| LLVM 数组（静态 + 动态）| ✅ Milestone 2 Phase 1 |
| LLVM 优化 Pass | ✅ `--llvm-opt=N` |

### LLVM 后端剩余限制

| 项目 | 状态 |
|------|------|
| 接口（vtable fat pointer）| 🔲 Milestone 2 Phase 2 (v3.2) |
| 泛型单态化 | 🔲 Milestone 2 Phase 3 (v3.2) |
| 异常（try/catch）| 🔲 v3.2+ |
