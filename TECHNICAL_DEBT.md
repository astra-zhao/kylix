# Kylix 技术债务与后续开发清单

> 最后更新: 2026-06-14
> 当前版本: v1.5.0+（post-Phase-11 修复批次）
> 关联文档: [ROADMAP.md](ROADMAP.md), [CHANGELOG.md](CHANGELOG.md)

本文档记录 v1.5.0 之后的已知缺陷、功能缺口和工程质量改进项，包含修复状态追踪。

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

## 优先级 2：功能缺口 🟠

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

## 当前状态总结（2026-06-14）

### Phase 11 完成度

| 优先级 | 项数 | 已完成 | 完成率 |
|--------|------|--------|--------|
| P1（正确性缺陷） | 3 | 3（已验证或确认无 bug） | 100% |
| P2（功能缺口） | 4 | 4 | 100% |
| P3（测试覆盖） | 4 | 4 | 100% |
| P4（工程质量） | 3 | 3 | 100% |
| P5（设计债务） | 3 | 0 | 0%（Phase 12）|

**Phase 11 完成！Phase 12（v2.0.0 发布准备）可以启动。**

### 新增测试统计

本批次新增 **38 个测试**（152 → 190+ 总测试数）：
- `pkg/compiler/compiler_test.go`: 5（接口验证）
- `pkg/compiler/cache_test.go`: 5（增量编译缓存）
- `pkg/compiler/typecheck_test.go`: 7（类型检查）
- `pkg/pkgmgr/manager_test.go`: 5（包管理器）
- `stdlib/klx_test.go`: 4（stdlib 声明文件）
- `parser/parser_regression_test.go`: 5（泛型/多返回值）
- `pkg/lsp/stdlib_test.go`: 4（LSP stdlib 加载）
- `generator/generator_multireturn_test.go`: 4（多返回值生成）

---

## Phase 12 目标（v2.0.0）

参见 [ROADMAP.md](ROADMAP.md) Phase 12 章节：
- 类型系统完善（类型推导、泛型约束、类型别名）
- 错误体验改进（错误代码、错误恢复、建议修复）
- 工具链（`kylix test`、`kylix doc`）
- stdlib 完全 Kylix 化
