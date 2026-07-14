# Kylix 项目上下文

Kylix 是现代 Pascal → Go 转译器。编译器用 Go 编写，生成 Go 代码。

**重要：始终用中文回答用户，不论用户用什么语言提问，回复一律使用中文。**

## 当前状态：v4.8.0（2026-07-14）

- v4.8.0 已发布：泛型类方法 codegen + DIBasicType 多类型。修复 example21 泛型类 stub（`TStack<Integer>.Push/Pop` 从 `Pop: 0` → 与 Go 后端一致 `Pop: 30`），打通泛型类 `var x := TStack<T>.Create()` → `x.Method()` 完整链路：单态化 walk VarDecl.Value + constructor inference 处理 GenericType/CallExpression + 类字段数组 `self.Items[i]` GEP（FieldInfo.ArrayType + emitArrayIndex MemberExpression 分支）。DWARF 调试信息从单一 int64 升级为按 llvmType 发射独立 DIBasicType（double→DW_ATE_float、ptr→DW_ATE_address、i1→DW_ATE_boolean），LLDB `frame variable` 显示正确类型。LLVM 测试 249→250，教程通过率 48/48（100%），example21 从 stub → 输出正确。
- v4.7.0 已发布：静态数组下界修复 + jsonutil 嵌套对象解析。AST `ArrayType` 新增 `LowerBound` 字段，parser 记录真实下界，LLVM 后端按真实下界调整索引（不再硬编码 1）——修复 example23 段错误（`array[0..4]` 的 `0-1` 无符号下溢 → GEP 越界）。jsonutil `JsonGetMap` 从返回 null 升级为递归 `parse_flat` 解析 raw JSON 子串为 nested htab（支持任意深度嵌套对象），并修复 `skip_nested` 的 pos bug（指向 close char 之后，不再丢失 sibling 字段）。LLVM 测试 240→249，教程通过率 48/48（100%），example23 从段错误 → 输出正确。
- v4.6.0 已发布：DWARF 逐行调试升级 —— per-instruction DILocation（每条 IR 指令附 `!dbg !N` 源行号+列号+scope，按 (line,col,scope) 去重）+ DILocalVariable + `#dbg_declare` 记录（LLVM 22 语法，替代废弃的 `call @llvm.dbg.declare`）。`emitStatement`/`emitExpr` 入口 `setDbgNode` 设置源位置，`line()` 自动给指令行附加 `!dbg`。LLDB 支持按源文件行号设断点、`step`/`next` 逐行单步、`frame variable` 检视局部变量（参数/`result`/用户变量）。LLVM 测试 240→247，教程通过率 48/48（100%）无回归。
- v4.5.0 已发布：LLVM stdlib Phase 3 完成 —— 3 个 stub 模块升级为真实实现（jsonutil 递归下降解析器 / crypto AES-256-CBC+PBKDF2 / httpclient libcurl 集成）+ 进程内 IR 优化 pass 管线（DCE）+ 增量编译缓存（llc 跳过，32x 加速）+ DWARF 调试符号（`-g` flag，LLDB/GDB 函数级调试）+ 文件拆分（expr.go 1207→777、stmt.go 1081→614，回到 1000 行约束内）。LLVM 测试 198→240，教程通过率 48/48（100%）。
- v4.4.0 已发布：LLVM stdlib Phase 2 完成 —— 8 个模块（encoding/net/crypto/db/cache/jsonutil/boot/jwt/httpclient，~2000 行 IR + 60+ 单元测试）+ KylixBoot 注解方法 stub 生成 + 链式方法调用修复（`self.Repo.Name()` 类型追踪）+ 9 个关键 bug 修复（字符串比较/块作用域/ptr-nil 比较/map 后缀/...）。LLVM 教程通过率 48/48（100%，含 example33 多文件模块）。
- v4.3.0 已发布：datetime 模块 Phase 1 完整（13 API + Arena Allocator）
- v4.2.0 已发布：sysutil 模块 Phase 1（8 API）
- v4.1.0 已发布：LLVM M4 高级特性 —— Lambda/闭包（捕获变量 + 环境结构体）、`inherited` 关键字（父类方法链查找）、完整多返回值元组解构、OOP 字段/方法访问系统性修复（vtable 继承）、优化通道（`opt` + `llc -O<N>`，循环归纳达 20x 提速）。LLVM 教程通过率 27/49，01-04 章节（19 文件）与 Go 后端输出逐字节一致。
- v4.0.0 已发布：LLVM M3（异常处理/字符串插值/控制流/表达式覆盖 ✅）+ stdlib Phase 7（db/cache/http/websocket ✅）+ IDE 插件（VS Code v1.1 ✅）
- v3.3.0：KylixBoot 框架完善 —— Body 绑定 + JWT + OpenAPI 3.1 自动生成
- v3.2.0：KylixBoot 注解栈 + LLVM M2 完整 + stdlib Phase 6
- v1.5.0：stdlib `.klx` 声明文件 + 包管理器
- 所有 Go 测试通过（16 个包，LLVM 后端 250 测试）
- 教程 49/49 测试通过（Go 后端，`examples/complete-tutorial/`）
- LLVM 后端 48/48 教程编译通过（100%，01-04 章节 19 个文件与 Go 后端输出逐字节一致；example33 多文件模块经 `multifile.go` MergePrograms 合并声明后通过）
- v4.8.0 新增：泛型类方法 codegen（`TStack<T>.Create()` → `Push/Pop` 完整链路，example21 输出正确）+ 类字段数组 `self.Items[i]` GEP + DIBasicType 多类型（per-llvmType，LLDB 显示正确类型）
- v4.7.0 新增：静态数组真实 LowerBound（`array[0..N]` 不再段错误）+ jsonutil `JsonGetMap` 递归嵌套对象解析
- v4.6.0 新增：DWARF 逐行调试（per-instruction DILocation + DILocalVariable + `#dbg_declare`，LLDB 逐行单步 + `frame variable` 变量检视）
- v4.5.0 新增：进程内 IR 优化 pass（DCE，默认运行）+ 增量编译缓存（llc 跳过，32x 加速）+ DWARF 调试符号（`kylix build --backend=llvm -g`）
- 所有源文件 ≤ 1000 行

## 关键文档

- [ROADMAP.md](ROADMAP.md) — 开发路线图（直到 v4.0）
- [TECHNICAL_DEBT.md](TECHNICAL_DEBT.md) — 已知问题与改进积压
- [TASKS.md](TASKS.md) — 详细任务分解
- [CHANGELOG.md](CHANGELOG.md) — 版本历史

## 架构

- `token/token.go` — Token 类型定义和关键字映射
- `lexer/lexer.go` — 词法分析器（字符 → token 流）
- `ast/ast.go` — AST 节点定义（接口 + 具体类型）
- `parser/parser.go` — Pratt 解析器核心；`parser_decl.go` 声明；`parser_stmt.go` 语句；`parser_expr.go` 表达式
- `generator/generator.go` — 生成器核心 + 预扫描；`generator_types.go` 类型/函数代码生成；`generator_stmt.go` 语句代码生成；`generator_expr.go` 表达式代码生成
- `generator/generator_boot_annotations.go` — KylixBoot 注解扫描 + 自动装配代码生成
- `generator/generator_validation_annotations.go` — 字段校验注解代码生成（`[Required]`/`[Email]` 等）
- `cmd/kylix/main.go` — CLI 入口（版本 3.3.0）
- `pkg/compiler/` — 编译 API + 增量缓存
- `pkg/compiler/annotations.go` — KylixBoot 注解诊断（KLX207–KLX214）
- `pkg/openapi/openapi.go` — OpenAPI 3.1 YAML 生成器
- `pkg/pkgmgr/` — 包管理器（add/install/remove）
- `pkg/repl/` — 交互式 REPL
- `pkg/lsp/` — Language Server Protocol
- `stdlib/` — Go 标准库封装（web, orm, template, exceptions, jwt 等）
- `stdlib/klx/` — LSP 补全用的 Kylix 声明文件
- `pkg/llvmgen/` — LLVM 后端代码生成器（原生二进制）
  - `codegen.go` — Generator 核心 + 字符串常量池 + 调试符号
  - `compile.go` — 编译管线（AST → IR → .o → binary）
  - `expr.go` — 表达式 codegen（算术/比较/调用/WriteLn）
  - `expr_access.go` — 成员/方法/接口/闭包访问 codegen
  - `stmt.go` — 语句 codegen（赋值/return/变量声明）
  - `stmt_flow.go` — 控制流 codegen（if/while/for/case/match/try）
  - `class.go` — 类/vtable/构造/方法 codegen
  - `stdlib_*.go` — 标准库模块 IR 实现（encoding/net/crypto/db/cache/jsonutil/boot/jwt/httpclient/sysutil/datetime）
  - `debug.go` — DWARF 调试符号生成（`-g` flag）：per-instruction DILocation + DILocalVariable + `#dbg_declare`（v4.6.0 逐行调试）+ per-llvmType DIBasicType（v4.8.0 类型精度）
  - `passes.go` — IR 优化 pass 管线（DCE + ConstantFold）
  - `cache.go` — 增量编译缓存（SHA256 键控 .o 复用）

## 已完成阶段

### Phase 6–10 → v1.0.2–v1.5.0
- 字符串插值、异常类型、多返回值、属性
- Map 类型、Variant 类型、动态数组
- 枚举、切片、单元文件系统、多文件编译
- 自举验证完成（Self-hosted compiler）
- 接口验证、Kylix 层错误报告、真正的泛型（Go 1.18+）
- 增量编译（55× 加速）
- stdlib `.klx` 声明 + 包管理器

### v3.1.x → KylixBoot 框架 + LLVM M2 Phase 1
- `[Controller]`/`[Get]`/`[Post]` 路由自动装配
- `[Service]`/`[Component]`/`[Inject]` DI 自动装配
- `[Required]`/`[Email]`/`[Min]`/`[Max]`/`[MinLen]`/`[MaxLen]` 字段校验
- `[Authenticated]`/`[Role]` 路由安全守卫
- `[Entity]`/`[Column]`/`[PrimaryKey]`/`[Repository]`/`[Query]` ORM 注解
- 注解诊断 KLX207–KLX213

### v3.2.0 → LLVM M2 完整 + stdlib Phase 6
- LLVM 后端 M2：接口胖指针、成员/方法分发、泛型类单态化
- stdlib `net`（TCP/UDP/DNS）、`crypto`（SHA/AES/BCrypt）、`encoding`（Base64/Hex/CSV）
- 注解栈全部完成，教程 42/42

### v3.3.0 → KylixBoot 框架完善（2026-06-28）
- `[Body(TEntity)]`：POST/PUT 路由的 JSON 请求体自动绑定 + IsValid()/Validate() 校验
- `jwt` stdlib：JwtSign/JwtVerify/JwtSubject + BootRegisterJwtAuth 一键接入 `[Authenticated]`
- `kylix doc --openapi`：从注解自动生成 OpenAPI 3.1 YAML（路径、schema、安全方案）
- 错误码修正：ErrBodyBinding 从 KLX301（冲突）改为 KLX214
- 教程 45/45 通过（新增 14_body_binding、15_jwt、16_openapi）

## 下一步：v3.3.0 收尾

**已完成 ✅**
- 类型检查层 MVP：`pkg/compiler/typecheck.go`（862 行）完整实现
- 包管理器编译器集成：`CompileProject` 自动发现 `packages/*/` 并去重
- 测试覆盖提升：新增 `packages_test.go`，所有关键包已有测试

**剩余工作**
- CompileFile 单文件模式的跨单元依赖自动解析（可选，非阻塞）
- 文档更新：tutorial README 提及包管理器用法
- 性能优化：大型项目的增量编译缓存验证

**v4.0 规划**
- LLVM M3：完整类型系统 + 优化通道
- stdlib Phase 7：http client/server + 数据库连接池
- IDE 插件：VSCode/JetBrains 语法高亮 + 跳转

## 关键约束

- Go 后端保持不变（Kylix → Go → binary）
- AST 节点使用 class（不用 variant records）
- **未经用户明确许可，绝不 commit/push**
- **每个源文件不超过 1000 行**：大文件按功能拆分
- build=`go build -o /tmp/kylix_bin ./cmd/kylix/ && KYLIX=/tmp/kylix_bin bash examples/complete-tutorial/test_all.sh 2>&1 | tail -8`
- test=`go test $(go list ./... | grep -v '/examples') 2>&1 | grep -E "^ok|FAIL"`

## 已知问题（v3.3.0）

详见 [TECHNICAL_DEBT.md](TECHNICAL_DEBT.md)。最优先修复的 3 项：
1. 包管理器未集成到编译器搜索路径（2.4）
2. `topoSortWithFiles` 文件路径对齐 bug（1.2）
3. `pkg/pkgmgr` + `pkg/compiler/cache` 零测试覆盖（3.1）

## 教程结构（examples/complete-tutorial/）

| 目录 | 示例数 | 状态 |
|------|--------|------|
| 01_basics ~ 11_modules | 32 | ✅ 全部通过 |
| 12_special_features | 7 | ✅ v3.2.0 |
| 13_stdlib_phase6 | 1 | ✅ v3.2.0 |
| 14_body_binding | 1 | ✅ v3.3.0 |
| 15_jwt | 1 | ✅ v3.3.0 |
| 16_openapi | 1 | ✅ v3.3.0 |
| 17_database | 1 | ✅ v4.0 |
| 18_cache | 1 | ✅ v4.0 |
| 19_http | 1 | ✅ v4.0 |
| 20_websocket | 1 | ✅ v4.0 |
| **合计** | **48 文件** | **49/49 通过** |
