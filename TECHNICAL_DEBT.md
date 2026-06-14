# Kylix 技术债务与后续开发清单

> 最后更新: 2026-06-14  
> 当前版本: v1.5.0  
> 关联文档: [ROADMAP.md](ROADMAP.md), [CHANGELOG.md](CHANGELOG.md)

本文档记录 v1.5.0 之后需要修复的已知缺陷、功能缺口和工程质量改进项。

---

## 优先级 1：正确性缺陷 🔴

这些问题会导致用户代码静默出错或产生错误的编译结果。

### 1.1 `CompileFile` 未接入增量缓存

**影响：** 项目模式（`kylix build`）使用 `CompileFile` 入口，增量编译缓存完全无效。

**位置：** `pkg/compiler/compiler.go:215`

```go
// 当前：opts.CacheDir 被传入但 CompileFile 不使用
result, err := compiler.CompileFile(mainFile, opts)
```

**修复方案：**
- `CompileFile` 添加与 `CompileProject` 相同的缓存逻辑
- 或者：项目模式改用 `CompileProject` + `FindAllKlxFiles()`

**工作量：** 2 小时

---

### 1.2 `topoSortWithFiles` 的文件路径对齐错误

**影响：** 多 unit 项目中，接口验证和 `//line` 指令会指向错误文件。

**位置：** `pkg/compiler/compiler.go:350-390`

**根本原因：**
```go
// progFile[prog] 用原始 files[i] 建立映射
for i, prog := range programs {
    if i < len(files) {
        progFile[prog] = files[i]  // BUG: 拓扑排序后 i 不再对应同一 prog
    }
}
```

**修复方案：**
```go
// 在 parse 时就建立 prog → file 映射
progFile := make(map[*ast.Program]string)
for i, file := range files {
    // ... parse ...
    progFile[program] = file
}
```

**工作量：** 1 小时

---

### 1.3 `GenerateBody` 的 exception types 输出位置不稳定

**影响：** 多文件编译时，exception 类型定义可能重复或包含其他 body 内容。

**位置：** `generator/generator.go:70-78` (`BuildOutput`)

**根本原因：** `g.output` 是全局累积状态，`snapshot := g.output.Len()` 会捕获先前 `GenerateBody` 调用的内容。

**修复方案：**
```go
func (g *Generator) WriteExceptionTypes() string {
    var b strings.Builder
    // ... 写入 b，不写入 g.output
    return b.String()
}

func (g *Generator) BuildOutput(bodies []string) string {
    // ...
    if g.needsException {
        out.WriteString(g.WriteExceptionTypes())
    }
    // ...
}
```

**工作量：** 30 分钟

---

## 优先级 2：功能缺口 🟠

用户能明显感知到的功能缺失或不可用。

### 2.1 类型检查层完全缺失

**影响：** 类型错误只能通过 Go 编译器报出，错误信息是 Go 语言风格（`cannot use int64 as string`）。

**当前流程：** parse → 接口验证 → codegen → Go 编译器报错

**应有流程：** parse → 符号解析 → 类型检查 → 接口验证 → codegen

**最小可行实现（MVP）：**
1. 赋值类型不兼容检测：`var x: Integer; x := 'hello'` → Kylix 错误
2. 未声明变量检测：`WriteLn(undeclaredVar)` → Kylix 错误
3. 函数调用参数数量检测：`foo(1, 2)` 但 `foo` 只接受 1 个参数 → Kylix 错误

**实现位置：** `pkg/compiler/typecheck.go` (新建)

**工作量：** 2 周

---

### 2.2 `kylix add` 的 git 包逻辑错误

**影响：** 已安装的包即使 ref 变更也不会更新。

**位置：** `pkg/pkgmgr/manager.go:158`

```go
// 当前逻辑（错误）：
if _, err := os.Stat(destDir); err == nil && tag == "" {
    return nil  // 无 tag 就跳过，即使 ref 变了
}
```

**正确逻辑：**
```go
// 有 tag 时才能幂等跳过；无 tag 时每次都拉最新
if _, err := os.Stat(destDir); err == nil && tag != "" {
    return nil
}
```

**工作量：** 5 分钟

---

### 2.3 多返回值的 TupleLiteral LHS 生成验证

**状态：** parser 已支持 `x, y := f()` 解析为 `TupleLiteral` LHS，但 generator 是否正确处理未验证。

**验证方法：**
```bash
grep -n "case \*ast.TupleLiteral" generator/generator_stmt.go
```

若无对应 case，需补充：
```go
case *ast.TupleLiteral:
    for i, elem := range assign.Name.(*ast.TupleLiteral).Elements {
        if i > 0 { g.write(", ") }
        g.generateExpression(elem)
    }
```

**工作量：** 1 小时（含测试）

---

### 2.4 包管理器与编译器未集成

**影响：** 用户 `kylix add utils <repo>` 后，`uses utils` 仍报"找不到 unit"。

**原因：** `pkgmgr.Manager.PackageDirs()` 已实现但从未被调用。

**修复位置：** `cmd/kylix/main.go` `cmdBuild()` 和 `pkg/compiler/compiler.go`

```go
// cmdBuild 中：
cfg, _ := project.Find(".")
mgr := pkgmgr.New(cfg)
packageUnits := mgr.PackageDirs()  // 获取 packages/*/ 路径

// CompileProject 添加参数：
func CompileProject(files []string, packageDirs []string, opts Options) (*Result, error) {
    // 从 packageDirs 加载所有 .klx 单元
}
```

**工作量：** 3 小时

---

## 优先级 3：测试覆盖空洞 🟡

关键路径缺少回归测试，重构时容易引入 bug。

### 3.1 关键模块无测试

| 模块 | 测试文件 | 测试数 | 风险 |
|------|----------|--------|------|
| `pkg/pkgmgr/manager.go` | 不存在 | 0 | 🔴 高 — git clone 逻辑未验证 |
| `pkg/compiler/cache.go` | 不存在 | 0 | 🔴 高 — 缓存失效逻辑未测试 |
| `pkg/lsp/document.go` `loadStdlibSymbols` | 不存在 | 0 | 🟠 中 — LSP 补全可能不工作 |
| `stdlib/klx/*.klx` 声明文件 | 不存在 | 0 | 🟠 中 — 语法错误无法发现 |
| `parser` 多返回值 / 泛型 | `parser_test.go` 不覆盖 | 0 | 🟠 中 — 优先级表变动易退化 |

**修复计划：**
1. `pkg/pkgmgr/manager_test.go` — 测试 Add/Install/Remove（用本地 tmpdir 模拟包）
2. `pkg/compiler/cache_test.go` — 测试 Load/Store/Invalidate
3. `pkg/lsp/symbols_test.go` — 测试 stdlib/klx 加载
4. `stdlib/klx_test.go` — 用 parser 解析所有 `.klx` 声明文件
5. `parser_test.go` — 补充 `TestParseGenericInstantiation`、`TestParseMultiReturn`

**工作量：** 1 周

---

## 优先级 4：工程质量 🟢

不影响功能，但提升代码可维护性。

### 4.1 `cmd/kylix/main.go` 超过 1000 行限制

**当前：** 744 行（接近上限）

**修复方案：** 拆分为 `cmd/kylix/cmd_*.go`
- `cmd_build.go` — `cmdBuild()`
- `cmd_run.go` — `cmdRun()`
- `cmd_package.go` — `cmdAdd()`, `cmdInstall()`, `cmdRemove()`
- `cmd_repl.go` — `cmdRepl()`
- `cmd_lsp.go` — `cmdLsp()`

**工作量：** 2 小时

---

### 4.2 `stdlib/klx/*.klx` 未验证可解析

**风险：** 手写的声明文件可能有语法错误（如 `Map<String, Variant>` 泛型语法）。

**修复方案：**
```go
// stdlib/klx_test.go
func TestKlxDeclarationsAreParseable(t *testing.T) {
    files := []string{"sysutil.klx", "datetime.klx", "regex.klx", "jsonutil.klx"}
    for _, file := range files {
        src, _ := os.ReadFile(filepath.Join("klx", file))
        l := lexer.New(string(src))
        p := parser.New(l)
        prog := p.ParseProgram()
        if len(p.Errors()) > 0 {
            t.Errorf("%s has parse errors: %v", file, p.Errors())
        }
    }
}
```

**工作量：** 30 分钟

---

### 4.3 `ioutil` 已废弃（Go 1.16+）

**影响：** `compiler.go` 和 `project.go` 使用 `io/ioutil` 产生编译警告。

**修复：**
```go
// 替换：
ioutil.ReadFile → os.ReadFile
ioutil.WriteFile → os.WriteFile
```

**工作量：** 15 分钟

---

## 优先级 5：设计层面的长期债务 ⚪

需要架构重构，但不阻塞短期功能开发。

### 5.1 缺少符号解析器（name resolver）

**影响：** LSP 补全只能依赖 stdlib/klx 声明文件，用户自己的 unit 无法跨文件补全。

**长期方案：**
1. `pkg/resolver/` — 符号解析器，建立全局符号表
2. 跨文件符号解析：`uses mylib` 时加载 `mylib.klx` 的符号定义
3. LSP 集成：补全时查询全局符号表

**工作量：** 3 周

---

### 5.2 `Generator` 全局状态不可重入

**影响：** 增量编译将来要并行化时（每个 unit 独立生成），当前 `Generator` 无法并行调用。

**根本原因：** `g.output` 和多个 map 字段是全局累积状态。

**长期方案：**
```go
type Generator struct {
    // 全局状态（多个 unit 共享）
    classTypes   map[string]bool
    imports      map[string]bool
    
    // per-unit 状态（每次 Generate 重置）
    currentOutput strings.Builder
    currentFile   string
}
```

**工作量：** 1 周

---

### 5.3 错误位置信息从字符串反向解析

**问题：** `parseLocation(errMsg string)` 用 `fmt.Sscanf` 从错误字符串提取行列号，脆弱且低效。

**长期方案：** `Diagnostic` 直接传递 `token.Token`，而非序列化再解析。

**工作量：** 2 天

---

## 最值得优先动手的三件事

按收益/成本比排序：

1. **修 2.4**（packages 目录加入编译路径）  
   → 5 行改动，让包管理器真正可用，用户立即能用

2. **修 1.2**（topo 排序文件路径对齐）  
   → 正确性 bug，多 unit 项目必现，影响 `//line` 映射

3. **补 3.1**（pkgmgr + cache 基础测试）  
   → 为后续重构提供安全网，避免引入回归
