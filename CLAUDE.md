# Kylix 项目上下文

Kylix 是现代 Pascal → Go 转译器。编译器用 Go 编写，生成 Go 代码。

**重要：始终用中文回答用户。**

## 当前状态：v3.3.0（2026-06-28）

- v3.3.0：KylixBoot 框架完善 —— Body 绑定 + JWT + OpenAPI 3.1 自动生成
- v3.2.0：KylixBoot 注解栈 + LLVM M2 完整 + stdlib Phase 6
- v3.1.x：接口验证、Kylix 层错误报告、真正的泛型、增量编译（55× 加速）
- v1.5.0：stdlib `.klx` 声明文件 + 包管理器
- 所有 Go 测试通过（16 个包）
- 教程 45/45 测试通过（`examples/complete-tutorial/`）
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
| **合计** | **44 文件** | **45/45 通过** |
