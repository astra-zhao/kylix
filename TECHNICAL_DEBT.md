# Kylix 技术债务与后续开发清单

> 最后更新: 2026-07-22
> 当前版本: v5.4.0 已发布
> 关联文档: [ROADMAP.md](ROADMAP.md), [CHANGELOG.md](CHANGELOG.md)

本文档记录 v3.1.0 之后的已知缺陷、功能缺口和工程质量改进项，包含修复状态追踪。

---

## ✅ v5.4.0 修复：LLVM 后端自举编译打通（类层次 RTTI + 全局变量 + record + 外部方法 + 20+ 运行时修复）

**症状**：v5.3.0 在 Go 后端达成自举不动点（`kylix_self2` ≡ `kylix_self3` 逐字节），但 LLVM 后端无法编译自举源码——`kylix build --backend=llvm src/*.klx` 失败，暴露整套类层次多态 + 类型系统缺失。

**根因**：LLVM 后端在自举源码（5250 行重多态）上暴露 20+ 缺口——类型系统（LLVMType 不感知 class/array）、函数 array 参数（fallback i64）、is/as 只支持 interface、异构 array of TBase（元素 i64）、无 collectClassTypes/classIsBase、无全局变量支持、无 record 支持、无外部方法支持。

**修复**（详见 CHANGELOG v5.4.0）：类型系统（llvmTypeOfExpr）、全局变量（collectGlobals + IsMerged 窄化）、record 支持（emitRecordDecl）、外部方法（@ClassName_Method + self）、类型推断（exprKylixType 递归 + auto-declare 按 RHS 类型 + 局部遮蔽全局）、is/as 运行时（vtable 边表 + class_is_a + null guard）、map 值类型化、builtin（Args/Ord/StrToFloat/LowerCase/ReadFile/append）、Boolean 比较 + 条件 coerce + G14 转义解码 + llc -O0 + emitConstructor call Create。

**验证**：`kylix build --backend=llvm src/*.klx` → 原生二进制 127KB → 运行 hello.klx exit 0 产出 Go 代码。回归 16 包 + 51 教程全绿。

### 🟠 v5.5 自举 parser 深层 bug（LLVM 后端运行时）

v5.4 让 LLVM 自举二进制能运行产出 Go 代码，但自举 parser（2400 行 parser.klx）编译成 LLVM 后有运行时 bug：

| 限制 | 影响 | 修复方向 | 状态 |
|------|------|---------|------|
| 整数解析失败 | `WriteLn(42)` → "no prefix parse function for 0" | 追踪 parseInteger 的 LLVM IR（整数 token 解析逻辑） | 🟠 v5.5 |
| 字符串参数未传递 | `fmt.Println()` 缺参数 | 追踪 GenerateCallExpression args 循环 + GenerateExpression(TStringLiteral) | 🟠 v5.5 |

---

## ✅ v5.3.0 修复：自举编译器 round-trip 打通 + 自繁殖

**症状**：v5.2.0 打通自举「构建」（`src/*.klx` → `kylix_self`），但 `kylix_self2`（自举产出的编译器）构建成功却运行产出空/错误——自举 `generator.klx` 的 codegen 保真度缺口。

**诊断**：自举 7 文件产出 `self_7.go` 全量 `go build` 仅 4 个错误（`Args`×3 + `os`×1）。自举源码不实际用 match/try/async/插值/lambda/注解/stdlib 模块（只在 token.klx 关键字表，是数据非代码）。故 G1–G27 大缺口不阻断自举自编译 round-trip。

**修复**（三处，都在 `src/generator.klx`，宿主 `generator/` 包零改动）：
- **`Args` builtin**：`MapBuiltinFunction` 加 `Args`→`os.Args[1:]`（main.klx 命令行参数）。
- **条件导入**：`GenerateImports` 硬编码 fmt/strings/strconv → 末尾组装（`CollectImports` 手写 `StrContains` 扫描 body 前缀设 `Need*` 标志，`BuildImportBlock` 按标志构造）。扫描 needle 拆分 `'fmt'+'.'` 避免编译器扫描自己输出时自检测（CollectImports 字面量 `"fmt."` 永远命中 → math/rand 假阳性未使用导入）。
- **字符串转义（G14）**：`WriteEscapedGoString` 逐字符转义把 `\n` 反斜杠加倍成 `\\n`（字面 backslash-n 而非真换行）→ 2-char 前瞻，`\n`/`\t`/`\r` 序列不加倍反斜杠。这是 round-trip 关键运行时 bug——gen1（宿主编译）的 `'\n'` 由宿主正确发射，但 gen2（自举编译）经自举 WriteEscapedGoString 发成字面 `\n` 致 `invalid character U+005C`。

**验证**：`kylix_self2` 编译 hello.klx → 可运行 Go 输出 `Hello, World!`；`kylix_self3`（kylix_self2 重新编译 src/*.klx）自繁殖同样正确；`self_7.go` 与 `self_7_gen2.go` 均 5390 行（接近不动点）。宿主 go test / 教程无回归（host 代码未改）。

---

## 🟠 v5.4 自举编译器通用 codegen 缺口（不阻断自举自编译）

v5.3 让自举编译器能正确编译**自举源码自身用到的特性子集**。自举源码不用的特性仍 stub/缺失——这些只影响「自举编译器编译用了这些特性的*其它*程序」，留 v5.4。详见 v5.3 规划期的 `src/generator.klx` 缺口清单（G1–G27）。

| 缺口 | 描述 | generator.klx 位置 | 宿主对应 | 状态 |
|------|------|---------------------|---------|------|
| G2 | 字符串插值 `` `Hello ${name}` `` 未实现 | GenerateExpression 无 TStringInterpolation 分支 | generator_expr.go:25-26 + generator.go:400-429 | 🟠 v5.4 |
| G7 | match 语句用 `switch expr`（Go case 不接受模式） | :1190-1223 | generator_stmt.go:348-405 | 🟠 v5.4 |
| G8 | try/except 缺 FinallyBlock/re-raise/nameMap | :1225-1272 | generator_stmt.go:408-490 | 🟠 v5.4 |
| G11 | lambda/await/tuple 表达式未处理 | GenerateExpression 缺分支 | generator_expr.go:101-124,276-324 | 🟠 v5.4 |
| G15 | async 函数未实现 | GenerateFunctionDecl 无 IsAsync 分支 | generator_types.go:334-336,408-455 | 🟠 v5.4 |
| G16 | variant 类型未处理 | GenerateTypeExpression 缺 TVariantType | generator_types.go:241-264 | 🟠 v5.4 |
| G6 | 多返回值链断裂 | GenerateFunctionDecl/Assignment/Return | generator_stmt.go:175-206,506-524 | 🟠 v5.4 |
| G3 | stdlib 模块派发（uses web/orm/jsonutil/...）未实现 | 无 resolveStdlibFunc | generator_stdlib.go | 🟠 v5.4 |
| G22-G24 | KylixBoot/Validation/ORM 注解自动装配未实现 | 无 | generator_*_annotations.go | 🟠 v5.4 |
| G1/G5/G25 | CollectClassTypes 空（ClassIsBase/ClassFields 未填充） | :199-208 | generator.go:438-472 | 🟠 v5.4（自举源码不用多态构造/基类，故自编译不触发；但编译其它有多态程序时会触发） |
| G4 | MapBuiltinFunction 仅映射 13/31 项（缺 math/rand/Copy/Pos/Trim 等） | :1671-1701 | generator_types.go:660-727 | 🟠 v5.4 |
| G10 | TFloatLiteral 硬编码 `0.0` | :1326 | generator_expr.go:22 | 🟠 v5.4 |
| G17 | GenerateFunctionDecl/ClassDecl 不跳过 Body==nil 前向声明 | :408-412,763 | generator_types.go:91-93,325-327 | 🟠 v5.4 |
| G19 | 无 //line 指令 | 全文 | generator.go:391-396 | 🟠 v5.4（调试体验，非正确性） |

---

## 🟠 v5.2.0 自举编译器残留限制

v5.2.0 打通自举编译器**构建**（`src/*.klx` → `kylix_self` 可运行二进制，208→0 错误），但完整 round-trip（自举编译器能正确编译任意程序）未达成，留 v5.3。

| 限制 | 影响 | 修复方向 | 状态 |
|------|------|---------|------|
| 自举 generator.klx codegen 保真度缺口 | `kylix_self` 产出的 Go 编译器（`kylix_self2`）构建成功但运行产出空——自举 generator.klx 自身是简化重实现，ClassIsBase/ClassFields 未填充（generator.klx:206-208 注释）、其余 codegen 路径不完整 | 补全自举 generator.klx 的 codegen（ClassIsBase/ClassFields 收集、类型表达式、语句/表达式全分支）使 `kylix_self2` 输出与宿主 kylix 逐字节一致 | 🟠 待 v5.3 |
| 基类含字段 + 多态 | opt-in interface 方案仅支持「基类无字段无方法 + 多态靠 is/as」。若程序基类有字段且通过基类变量访问，或基类有虚方法需分派，会崩 | getter 转发（interface 方法 + 具体类实现）/ vtable；或 per-base 决策 | 🟠 待 v5.3 |
| program-level 多态标志过宽 | 含 `is`/`as` 的程序会把所有「有子类的基类」都变 interface；混合程序（部分基类需字段继承、部分需多态）会误伤 | per-base 检测：仅对实际承载子类实例/作断言操作数的基类发射 interface | 🟠 待 v5.3 |
| `Args` builtin 无变量名守卫 | `mapBuiltinFunction("Args")` 无 var 守卫，若用户程序声明 `var Args` 会被改写成 `os.Args[1:]` | 全仓库无此用例；可加 var-name 守卫或在标识符 codegen 查 var 表 | ⚪ 文档化为限制 |

### ✅ v5.2.0 修复：自举编译器构建打通

**症状**：自举源码 `src/*.klx`（7 文件 5250 行）经 Go 后端转译后 `go build` 失败 208 个错误，无法生成可运行的 `kylix_self`。

**根因**：Go 后端把所有类发射成普通 struct + 嵌入父 struct（给字段继承但无多态）。自举源码用经典 Pascal OOP 多态：`decl: TNode; decl := prog.Declarations[i]; if decl is TClassDecl then cd := decl as TClassDecl`——异构 `array of TNode` 持有子类实例 + `is`/`as` 下转。struct 指针基类上 `x.(*TSub)` 非法、子类装不进 `[]*TBase`。`classIsBase` map 早已在 `collectClassTypes` 填充但 v3.1.0（KLX-C01）回退后从不读取（死代码）。

**修复**（opt-in，避免回归教程 example19/example40 的字段继承）：
- Parser 端 `parseIsExpression`/`parseAsExpression` 置 `p.usesPolymorphism=true` → `program.UsesPolymorphism`（新增 `ast.Program.UsesPolymorphism bool`）。
- `collectClassTypes`（所有预扫描路径的公共咽喉）OR 进 `g.usesPolymorphism`，一处覆盖 Generate/GenerateMulti/CompileProject/CollectClassTypes。
- `generateClassDecl`：`g.usesPolymorphism && g.classIsBase[name]` → `type TName interface{}`（跳过 struct 体与方法循环）；具体类父嵌入加条件 `!(poly && classIsBase[parent])`。
- `generateTypeExpression`/`generateTypeExpressionForCast`：基类（poly）→ `TName`（interface，不带 `*`）；具体类 → `*TName`。
- `mapBuiltinFunction` 加 `Args`→`os.Args[1:]`（main.klx 命令行参数）。
- `src/parser.klx:448` 切片协变修复（`array of TStatement`→`array of TNode`，Go 切片不变式）。

**验证**：`go build` 208→0 错误 → `kylix_self`（2.9MB）运行产出 5238 行 Go 编译器代码。go test 16 包全绿（+3 多态测试），教程 51/51 无回归。

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

### ✅ 6.1 接口未支持 → v3.2.0 M2 Phase 2 修复

`pkg/llvmgen/interface.go`（~230 行）：fat pointer（`{ ptr vtable, ptr data }`）、每接口方法 thunk、`is`/`as` 断言。23 个测试。

### ✅ 6.2 泛型无单态化 → v3.2.0 M2 Phase 3 修复

`pkg/llvmgen/monomorph.go`（~270 行）：`collectInstantiations` AST walker + 类型参数替换 + mangling（`TBox<Integer>` → `TBox_Integer`）。6 个 IR 测试。

### ✅ 6.4 不支持异常（try/catch）→ v4.0 M3 修复

**已修复：** `pkg/llvmgen/exc.go` + `stmt.go` 的 `emitTry`/`emitRaise` 实现完整 Pascal 异常语义：
- 路线 C：全局异常槽 + setjmp/longjmp 携带类型信息（避开 Itanium C++ EH ABI）
- try/except/finally + on E: Type do 类型化捕获 + raise + 裸 raise 重抛 + 嵌套 try
- 注入 Exception class + `@__kylix_is_subtype` 运行时子类型匹配
- finally 复制 3 份保证确定性执行
- 20 个 IR 片段测试

附带修复：字符串插值 codegen、带参构造 `T.Create(args)`、类字段继承（子类 struct 包含父类字段）。

### ✅ 6.5 文件大小约束违反 → v4.5.0 修复

**已修复：** expr.go(1207行) / stmt.go(1081行) 超过 1000 行约束。

拆分结果：
- `expr.go` 1207→**777 行**（核心表达式 codegen）
- `expr_access.go` **440 行**（新，成员/方法/接口/闭包访问）
- `stmt.go` 1081→**614 行**（核心语句 codegen）
- `stmt_flow.go` **484 行**（新，控制流 if/while/for/case/match/try）

### ✅ 6.6 无优化深化（DCE/内联/循环优化）→ v4.5.0 修复

**已修复：** `pkg/llvmgen/passes.go`（126 行）—— 进程内 IR 优化 pass 管线：
- DeadCodeElim (DCE)：删除未被引用的 `%tN` 临时寄存器定义（纯指令），词边界检查防止 `%t1` 误匹配 `%t10`
- ConstantFold：MVP 结构钩子（未来扩展用）
- 默认运行（-O0 时自动），`--llvm-opt` 时跳过（外部 opt 跑更强 pass）
- 字符串常量去重：`addString` 按内容去重，两个 `"hello"` 共享一个 `@.str.N`

### ✅ 6.7 无增量编译（每次全量 llc）→ v4.5.0 修复

**已修复：** `pkg/llvmgen/cache.go`（149 行）—— 按 IR 内容 + opts 的 SHA256 缓存 `.o`：
- 缓存命中时直接复用 `.o`，跳过 llc
- 实测：example01 二次构建 **0.939s → 0.029s（32x 加速）**
- best-effort：缓存失败非致命（静默降级到全量编译）

### ✅ 6.8 无调试符号（LLDB/GDB 不可用）→ v4.5.0 函数级 / v4.6.0 逐行

**已修复：** `pkg/llvmgen/debug.go` —— DWARF 调试符号：
- `-g` flag：`kylix build --backend=llvm -g` 发出 DWARF 调试信息
- metadata：`!llvm.dbg.cu` + `DICompileUnit` + `DIFile` + `DISubprogram`（每个用户函数 + main）
- v4.5.0 函数级：函数级断点、backtrace 显示函数+行号、源文件关联
- v4.6.0 逐行：per-instruction DILocation（每条 IR 指令附 `!dbg !N` 源行号+列号+scope）+ DILocalVariable + `#dbg_declare` 记录
  - LLDB 支持：按源文件行号设断点、`step`/`next` 逐行单步、`frame variable` 检视局部变量（参数/`result`/用户变量）
  - LLVM 22 适配：用 `#dbg_declare` 记录语法替代废弃的 `call @llvm.dbg.declare` intrinsic
  - `emitStatement`/`emitExpr` 入口 `setDbgNode` 设置源位置；`line()` 自动给指令行附加 `!dbg`
- -g 与 -O 互斥：`-g` 强制 `-O0`（优化重排指令会使调试信息误导）
- 已知残留（v4.7.0）：stdlib 预生成 IR 无源码行号（不附 DILocation）；DIBasicType 单一化（double/ptr 值格式化可能不精确）；类方法/lambda 未给 DISubprogram；无 DILexicalBlock（块作用域归函数级）

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

## 当前状态总结（2026-06-25）

### 新增已知缺陷（v3.1.0 引入或残留）

| ID | 问题 | 严重度 | 目标 |
|----|------|--------|------|
| **KLX-G01** | `example21_generic_class` 泛型类 receiver codegen 错误 | 中 | ✅ v3.1.1 |
| **KLX-M01** | `example33_use_module` 多文件 unit 编译路径失败 | 中 | ✅ v3.1.1 |

### Phase 11 完成度（v1.5.0–v2.0.0）

| 优先级 | 项数 | 已完成 | 完成率 |
|--------|------|--------|--------|
| P1（正确性缺陷） | 3 | 3（已验证或确认无 bug） | 100% |
| P2（功能缺口） | 4 | 4 | 100% |
| P3（测试覆盖） | 4 | 4 | 100% |
| P4（工程质量） | 3 | 3 | 100% |
| P5（设计债务） | 3 | 0 | 0%（长期）|

### v3.1.1 新增修复

| 项目 | 状态 |
|------|------|
| KLX-G01 泛型类方法 receiver | ✅ 生成 `*TStack[T]`，教程示例可运行 |
| KLX-M01 Unit interface/implementation | ✅ 正确生成 implementation 函数体，跳过 forward declarations |
| 教程测试覆盖 | ✅ `test_all.sh` 覆盖所有目录，35/35 通过 |
| 增量缓存 codegen 失效 | ✅ `CacheVersion` 防止复用旧 fragment |

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

---

## ✅ v4.7.0 修复：example23_arrays 静态数组段错误

**症状**：`example23_arrays.klx`（静态 `array[0..N] of T`）经 LLVM 后端编译运行时段错误（exit 139），不带 `-g` 也复现。

**排查**：v4.5.0（HEAD）与 v4.6.0 生成的 IR **逐字节一致**（`diff` 无差异），且两个版本都段错误——确认为 v4.5.0 预先存在的 bug，非 v4.6.0 引入。

**根因**：`pkg/llvmgen/array.go` 第 69 行硬编码 `LowerBound: 1`（Pascal 默认），但 `array[0..4]` 的下界是 0。`emitArrayIndex` 第 106 行 `sub i64 idx, LowerBound` 把 `0 - 1 = -1`（i64 无符号下溢成 0xFFFFFFFFFFFFFFFF），GEP 越界访问非法内存 → 段错误。

**修复**（v4.7.0）：
- `ast/ast.go`：`ArrayType` 新增 `LowerBound Expression` 字段
- `parser/parser_expr.go`：DOTDOT 分支设置 `arrayType.LowerBound = lowerBound`（解析时记录，不再丢弃）
- `pkg/llvmgen/array.go`：`emitArrayVarDecl` 用 `evalConstInt(arr.LowerBound)` 计算真实下界，传给 `arrayInfo.LowerBound`
- `emitArrayIndex` 已有 `if LowerBound != 0` 分支（LowerBound=0 走 `add idx, 0` 不调整），逻辑正确——只需 `arrayInfo.LowerBound` 正确

**验证**：example23 LLVM 编译运行输出 `numbers[0]=10 ... numbers[4]=50` + `Names: 1.Alice 2.Bob 3.Charlie`，与 Go 后端一致。

**附注**：example21_generic_class 也用 `array[0..99]`，但因泛型类方法是 stub（`unsupported receiver`），输出全是 0，掩盖了下界 bug。修复后 example21 仍是 stub 输出（泛型类方法未实现，独立问题）。

---

## 🟠 v4.6.0 DWARF 调试信息残留限制

| 限制 | 影响 | 修复方向 | 状态 |
|------|------|---------|------|
| stdlib 预生成 IR 无 DILocation | 进入 stdlib 函数时无可单步源码 | 给 stdlib IR 注入合成 DILocation（或文档化为预期） | 🟠 待 v5.0 |
| DIBasicType 单一化（int64） | double/ptr/string 局部变量值格式化不精确 | 按 llvmType 发射独立 DIBasicType（DW_ATE_float/address/...） | ✅ v4.8.0 已修复 |
| 类方法/lambda 无 DISubprogram | OOP 方法/闭包体内不可逐行单步 | emitMethod/emitLambda 注册 subprogram + setDbgScope | ✅ v4.9.0 已修复 |
| 无 DILexicalBlock | 块级 `begin var x; end` 变量归函数级 scope | 为 BlockStatement 发射 DILexicalBlock 作 scope | ✅ v4.9.0 已修复 |
| 块作用域变量 location list 范围窄 | 块内变量在块外 PC（如 WriteLn 调用点）报 `<variable not available>` | 把块内 alloca 提升到 entry block（结构重构） | 🟠 待 v5.0 |

---

## 🟠 v4.7.0 jsonutil 残留限制

| 限制 | 影响 | 修复方向 | 状态 |
|------|------|---------|------|
| ~~`JsonGetArray` 返回 null~~ | ~~array of Variant 需 Variant 运行时~~ | ~~v4.9.0 字符串数组 slice；v5.0.0 类型标签 Variant box~~ | ✅ v5.0.0 已修复（类型标签 Variant box 切片） |
| ~~`array of Variant` 数值比较（`arr[0] = 1.0`）~~ | ~~v4.9.0 字符串数组版不支持 Variant 数值索引~~ | ~~Variant 运行时（类型标签 + union + dispatch）~~ | ✅ v5.0.0 已修复（variant_compare 按标签派发） |
| Variant 算术（`v + 1`、`v * 2`） | v5.0.0 未实现 Variant 运算符重载 | 运行时按标签算术派发 | ✅ v5.1.0 已修复（variant_add/sub/mul/div，LLVM-only） |
| ~~`map[String]Variant` 真实化~~ | ~~htab 值槽是字符串，Variant map 值需宽化~~ | ~~htab 变体值槽（结构重写）或 Variant-valued map~~ | ✅ v5.1.0 已修复（htab 值槽存 box ptr，不动结构） |
| `div`/`mod` Variant | v5.1.0 算术只支持 `+,-,*,/`，`div`/`mod` 留 stub | 整数除/模按标签派发 | 🟠 待 v5.2 |
| Variant 算术 Go 不支持 | Go `interface{}`不支持运算符，Variant 算术仅 LLVM | Go 端 type-switch codegen 或文档化 | ⚪ 文档化为限制 |
| jsonutil 嵌套递归深度无界 | 超深 JSON 可能栈溢出 | 教程用例 2-3 层可接受；超深场景文档化为限制 | ⚪ 文档化为限制 |
| null 打印分歧 | LLVM nil box → "nil"，Go `nil` → "<nil>" | 统一 null 语义（v5.2） | ⚪ 文档化为限制 |
| ~~example21 泛型类方法 stub~~ | ~~`TStack<T>.Push/Pop` 输出 `Pop: 0`（unsupported receiver）~~ | ~~泛型类单态化后的方法 codegen~~ | ✅ v4.8.0 已修复 |

---

## ✅ v5.1.0 修复：完成 Variant 运行时（map[String]Variant 真实化 + Variant 算术）

**症状**：v5.0.0 让标量 + `array of Variant` 成为真 Variant，但 `map[String]Variant` 仍是 htab 字符串（`m['pi']` 返回 C 串 → `m['pi']=3.14` 是 string-vs-double），Variant 算术（`v+1`）发 stub。

**修复**：
- **map[String]Variant 真实化**（htab 值槽类型无关，仅 map codegen 层 + jsonutil 改动，cache/string-map 不回归）：`stdlib_map.go` `emitMapVarDecl` 检测 Variant 值类型设 `variantMaps`；`emitMapIndexGet` 走 `htab_get_variant` 返回 `"variant"`；`emitMapIndexPut` 装箱 RHS。`stdlib_hashtab.go` 新增 `htab_get_variant`（miss 返回 `@__kylix_variant_nilbox` 全局 tag=0 → as_* 走 nil 默认，与「missing key 返回默认」契约一致）。`parse_flat` 改调 `value_to_variant` 让 JsonDecodeMap 产出 Variant map；`JsonGetString/Int/Float/Bool/Map/Array` 全部 unbox via `htab_get_variant` + `variant_as_*`。新增 `as_int`/`as_bool` 主体 + nilbox 全局。
- **Variant 算术**：`variant_add`（either str→拼接/both int→int/else double）、`variant_sub`/`variant_mul`（both int→int else double）、`variant_div`（始终 double）；`emitInfix` 算术 stub 替换为 `emitVariantArith`；`coerceValue` 加 variant→concrete（`n := v` 解箱），`emitAssign` 内联 coercion 改调 `coerceValue` 统一处理。

**验证**：example57_variant_map 双端输出逐字节一致（`m['pi']=3.14`/`m['flag']=true`/`WriteLn(m['name'])`）。LLVM 测试 266→274（+8 Variant map/算术），教程 50→51。

**限制**：Variant 算术仅 LLVM（Go `interface{}`不支持运算符，无共享教程，IR 测试覆盖）。`div`/`mod` Variant 留 stub。box 内存不释放（no-GC 一致）。

---

## ✅ v5.0.0 修复：Variant 运行时（标量 + `array of Variant`）

**症状**：LLVM 后端把 `Variant` 静默当 `i64` 别名——`LLVMType("Variant")` 走 `default → "i64"`，`var v: Variant; v := 1.0` 把 double 截断成 i64，`array of Variant` 元素槽是裸 i64（无类型标签），`arr[0] = 10.0` 比较的是位模式/数据指针而非数值。v4.9.0 的 `JsonGetArray` 退而产出「C 字符串指针切片」并文档化「完整 Variant 运行时留 v5.0」。

**修复**（新文件 `pkg/llvmgen/variant.go` + expr/stmt/array/codegen/jsonutil 接线）：
- **boxed-pointer 表示**：`%struct.kylix_variant = { i32 tag, i64 payload }`（16 字节，tag 0=nil/1=int/2=float/3=str/4=bool）。Variant 值是 `ptr`（指向堆 box）；存储槽（`_var` alloca / Variant 数组元素槽）是 `ptr`。
- **运行时 helpers**（受 `variantRuntimeEmitted`/`needVariantRuntime` 守卫，`emitProgram` 末尾发射）：`box_int/float/str/bool`、`as_double`、`as_str`、`compare(ptr a,ptr b)→i32`（-1/0/1，按标签派发：数值提升 double、字符串 strcmp、布尔 payload、异类型按 tag 距离）、`print`/`println`（按标签 puts/printf）。
- **合成类型 "variant"**：`emitExpr` 对 Variant box 值返回伪类型字符串 `"variant"`，让 Variant 身份贯穿 `emitInfix`（`+` guard 前插比较分支）/`emitWriteLn` 等而无需平行类型表。逐点审计现有 switch。
- **标量 + 数组**：`emitVarDeclSingle` Variant 分支（`_var` suffix，generic fallback 前）；`arrayInfo.IsVariant` + `emitArrayIndex` asLValue/rvalue 类型分流；`emitAssign` 装箱（arr[i] 路径 + 标量 `_var` 路径，coercion 前）。
- **JsonGetArray 升级**：`emitJsonParseArray` 改调 `value_to_variant`（窥首字符分类 → box_str/float/bool，数字全 float box 与 Go json 的 float64 对齐）；`JsonArrayGetString` unbox via `as_str`；`JsonArrayLen` 不变。
- **Length(arr) 路由修复**：`emitCall` Length 分支先查 `arrayInfo` → `emitArrayLength`（slice len word / 静态常量），否则 strlen。此前 `emitArrayLength` 是死代码、`Length(arr)` 对数组返回 0（v3.1.0 起遗漏）。

**验证**：example56_variant 双后端输出逐字节一致（`arr[0]=10.0` 数值比较、`WriteLn(arr[i])` 按标签打印、`Length(arr)` 正确、字符串/布尔数组）。LLVM 测试 255→266（+11 Variant 测试），教程 49→50。

**务实范围**：Variant 算术（`v+1`）与 `map[String]Variant` 真实化留 v5.1（box 内存不释放与现有 no-GC 一致）。

---

## ✅ v4.9.0 修复：类方法/lambda DISubprogram + DILexicalBlock + jsonutil 嵌套数组

### 类方法/lambda DISubprogram

**症状**：v4.6.0 只给用户函数和 main 注册 DISubprogram，类方法（`emitMethod`）和 lambda（`emitLambdaFunc`）无调试元数据——OOP 方法不可逐行单步、`frame variable` 不显示 `self`/参数/捕获变量。

**修复**：`emitMethod`/`emitLambdaFunc` 复用 `emitFunctionDecl` 模式——`registerSubprogram` + `defineLineWithDbg` + `setDbgScope`/`setDbgNode` + `self`/参数/`result`/捕获变量 `emitDbgDeclare` + 退出 `setDbgScope(0)`/`clearDbgPos` + 合成 ret 前 `clearDbgPos`。stub 方法（`Body==nil`）不发调试信息。`nodeToken` 加 `FunctionDecl` 分支。

**验证**：LLDB 方法体内 `break`/`step`/`frame variable` 显示 `(long) self = ...`、参数值、捕获变量值；backtrace 显示 `__lambda_0` 帧。

### DILexicalBlock

**症状**：v4.6.0–v4.8.0 块内 `var` 声明的变量全归函数级 subprogram scope，块作用域语义在调试信息中丢失。

**修复**：`dbgMeta` 加 `lexBlocks` + `registerLexicalBlock()`；`emitBlockScoped` 进出块时切/恢复 `curScope`；`emitDbgMetadata` 发射 `!DILexicalBlock(scope: !parent, ...)`；`retainedNodes` 只含 subprogram 级变量（块作用域变量经 lexical block scope 链可达）。

**已知限制**：块作用域变量的 alloca 跟随 VarDecl 在块的基本块发射（非 entry block），LLVM location-list 生成器据此限制变量可用范围——LLDB 在块外（如 `WriteLn(y)` 调用点）可能报 `<variable not available>`。这是 v4.6.0 起 alloca-in-block 设计的固有行为（v4.8.0 同样存在，非 v4.9.0 回归）；DILexicalBlock 仍正确反映 scope 树。完整修复需把块内 alloca 提升到 entry block（留 v5.0）。

### jsonutil JsonGetArray 嵌套数组 + skip_nested off-by-one

**症状**：`JsonGetArray` 是 `ret ptr null` stub。同时 `skip_nested` 的 raw 子串丢失闭合 `]`/`}`（`length = end - start`，end 指向 close char 但 memcpy 不含它）——嵌套对象因 `parse_flat` 容忍未暴露，数组因 `parse_array` 遇缺 `]` 无限循环 → OOM/SIGKILL。

**修复**：`skip_nested` `length = endAfter - start`（含闭合 char）；`JsonGetArray` 升级为解析 raw 数组子串为字符串数组 slice `{ptr,i64,i64}`（标量存文本、嵌套对象/数组存 raw 子串）；新增 `parse_array` 状态机（growable buffer）+ `JsonArrayLen`/`JsonArrayGetString` 访问器；`emitAssign` 加 `_dyn` slice 结构 copy 分支；`sliceArgPtr` 从 ident 直接取 alloca。

**验证**：`{"items":["apple","banana","cherry"]}` → `JsonArrayLen=3` + 三字符串；`{"nums":[10,20,30]}` 数字数组；`{"users":[{...},{...}]}` 嵌套对象数组；缺失键 → 0 长度。

**务实范围**：完整 Variant 运行时（`array of Variant` 的 `arr[0] = 1.0` 数值比较）是 v5.0 级工作。v4.9.0 字符串数组版覆盖字符串/数字数组读取 80% 用例，不引入 Variant。

---

## ✅ v4.8.0 修复：example21 泛型类方法 codegen

**症状**：`example21_generic_class.klx` LLVM 后端输出 `Stack count: 0 / Pop: 0`（应为 `Stack count: 3 / Pop: 30`）。Go 后端正确。

**根因（三个）**：
1. 单态化未触发：`collectInstantiations` 的 `visitStmtForGenerics` VarDecl case 只 walk `s.Type`，不 walk `s.Value`。`var intStack := TStack<Integer>.Create()` 是类型推断，GenericType 从未被 walk。
2. constructor inference 不处理 CallExpression/GenericType：`emitVarDecl` 只识别 `TFoo.Create`（MemberExpression + Identifier），不处理 `TStack<Integer>.Create()`（CallExpression + GenericType）。
3. 类字段数组 GEP 未实现：`self.Items[i]` 的 Left 是 MemberExpression，`emitArrayIndex` 只处理 Identifier。

**修复**：
- `monomorph.go`：VarDecl case 加 walk s.Value
- `stmt.go`：`constructorClassName` 辅助函数处理 MemberExpression/CallExpression + Identifier/GenericType
- `class.go`：`FieldInfo.ArrayType` 字段 + 数组字段 LLVM 类型 `[N x T]`
- `array.go`：`emitArrayIndex` MemberExpression Left 分支 + `emitStaticArrayGEP`

**验证**：example21 LLVM 输出与 Go 后端逐字节一致（`Stack count: 3 / Pop: 30 / Pop: 20 / Stack count: 1 / Pop: World`）。

**附注**：example21 的静态数组字段 `array[0..99]` 现在也受益于 v4.7.0 的 LowerBound 修复（下界 0 → 不调整索引）。两个修复协同使泛型栈类完整工作。
