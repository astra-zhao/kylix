# Kylix Development Roadmap

> 最后更新: 2026-06-29  
> 当前版本: v3.3.0 ✅  
> 官网: [kylix.top](https://kylix.top)  
> 目标: Kylix 成为生产级、多后端、全栈 Pascal 语言

**✅ v3.3.0 已发布！** Body Binding + JWT + OpenAPI 3.1 + 包管理器集成 + 类型检查器 MVP，教程 45/45 通过，16 包全部测试通过。

---

## 总览

| 阶段 | 内容 | 状态 | 版本 |
|------|------|------|------|
| Phase 1-5 | 转译器 + IDE + 框架 + 语言增强 + stdlib | ✅ 完成 | v1.0.0 |
| Phase 6-9 | Bug修复 + 语言能力 + 自举 + 自举验证 | ✅ 完成 | v1.0.2–v1.2.0 |
| Phase 10-12 | v2.0 核心特性 + 工程质量 + 生产级发布 | ✅ 完成 | v1.3.0–v2.0.0 |
| **v2.1.0** | 增强类型系统 + stdlib Phase 1 | ✅ 完成 | v2.1.0 |
| **v2.2.0** | 工程质量 + stdlib Phase 2 | ✅ 完成 | v2.2.0 |
| **v2.3.0** | 开发者体验 (LSP/REPL/Debug/WASM) | ✅ 完成 | v2.3.0 |
| **v2.4.0** | 完善与生态 (i18n/推导/SetLength/包管理) | ✅ 完成 | v2.4.0 |
| **v2.5.0** | 工具链深化 (LSP/doc/bench/iter/方法修复) | ✅ 完成 | v2.5.0 |
| **v2.6.0** | 性能与优化 (并行编译/DCE/LSP基准) | ✅ 完成 | v2.6.0 |
| **v3.0.0** | LLVM 后端 + 包注册中心 + WASI | ✅ alpha | 2026-06-21 |
| **v3.1.0** | KylixBoot 框架 + 注解语法 + LLVM 数组 + 编译器修复 | ✅ 完成 | 2026-06-23 |
| **v3.2.0** | KylixBoot 注解栈（路由/DI/校验/安全/ORM）+ 诊断 | ✅ 完成 | 2026-06-26 |
| **v3.3.0** | Body Binding + JWT + OpenAPI + 包管理器集成 + 类型检查器 | ✅ 完成 | 2026-06-29 |
| **v4.0.0** | LLVM M3 + stdlib Phase 7 + IDE 插件（脱离 Go 长期目标） | 📋 规划中 | 2026 Q3 |

---

## 📊 累计统计 (v3.2.0-dev)

| 指标 | 数量 |
|------|------|
| Go 测试包 | 15+ 个（全部通过）|
| Go 级测试 | ~330+ |
| Kylix 级 stdlib 测试 | 117 个（10 模块）|
| 纯 Kylix stdlib 函数 | 90+ |
| CLI 命令 | 19 个 |
| 教程示例 | 34 个（32 完全工作，~94%）|
| 原生构建目标 | 5 (linux/darwin/windows × amd64/arm64) |
| WASM 目标 | 2 (Go 标准 + TinyGo) |
| WASI 目标 | 2 (Go wasip1 + TinyGo) |
| LLVM 后端 | ✅ Milestone 2 Phase 1（数组 + 优化）|
| LLVM 测试 | 30 个（含 6 个数组测试）|
| KylixBoot 测试 | 23 个 |
| 包注册中心 | ✅ REST API + Web 前端 |

---

## v3.0.0-alpha ✅ (2026-06-21) — 架构突破

已在 TASKS.md 和 CHANGELOG.md 详细记录。核心交付：
- LLVM 原生后端 Milestone 1（24 测试）
- 包注册中心服务端（registry/ + kylix publish）
- WASI 支持（--wasi flag + pkg/wasi/）
- stdlib Phase 4 纯 Kylix 化（jsonutil/regex/datetime）
- 编译器 bug 修复（external 解析）
- 入门教程 29 个示例

---

## ✅ v3.1.0 (2026-06-23) — KylixBoot 框架 + 注解语法 + LLVM 数组 + 编译器修复

详见 TASKS.md 和 CHANGELOG.md。核心交付：

### 已完成
- [x] **KylixBoot 框架核心运行时** —— `pkg/boot/`（~700 行，23 测试）
  - Router（路径参数 `/users/:id`）、Server（优雅停机）
  - DI 容器（Singleton/Transient/Instance + 反射 Inject）
  - 全局快捷方式（`boot.GET`、`boot.POST`、`boot.Use`、`boot.Listen`）
  - Config（环境变量回退）
  - 内置中间件：Logger / Recover / CORS / Auth / RateLimit / RequestID
  - 桥接 `stdlib/boot_bridge.go` 重新导出
  - LSP 声明文件 `stdlib/klx/boot.klx`
- [x] **注解语法 `[Name]` / `[Name(args)]`** —— `ast.Attribute` + `parser/parser_attribute.go`
  - 作用于 `ClassDecl`、`TypeDecl`、`FunctionDecl`、`VarDecl`
  - 顶层和类体内均可使用
  - 新示例 `example41_attributes.klx`
- [x] **KLX-C01 修复** —— `var p: TClass` 现在生成 `*TClass` 而非 `interface{}`
- [x] **KLX-C02 修复** —— 单引号字符串中 `${...}` 正确生成 STRING_INTERPOLATION
- [x] **KLX-C03 修复** —— lambda/匿名函数返回类型保留
- [x] **KLX-C04 修复** —— match 语句生成 tagless `switch { case _v == p: }`（不再生成无效 Go）
- [x] **KLX-C05 修复** —— `uses sysutil/jsonutil/datetime/regex/httpclient` 在 program 中注入符号
  - `generator/generator_stdlib.go`（~270 行）映射 40+ stdlib 函数
- [x] **LLVM Milestone 2 Phase 1** —— 数组 + 优化
  - `pkg/llvmgen/array.go`（~200 行）：静态 `array[1..N] of T` → `alloca [N x T]`
  - 动态 `array of T` → `{ ptr, i64, i64 }` slice 结构体
  - Pascal 1-based 索引自动转 LLVM 0-based
  - 编译期常量求值（`array[1..N]` 处理 `((N-1)+1)`）
  - `--llvm-opt=0/1/2/3` CLI 标志（`llc -O=N`）
  - 6 个新测试（总数 30）
- [x] **教程扩展** —— `example40_declarative_oop.klx` + `example41_attributes.klx`，32/34 通过（~94%）

---

## 📋 v3.2.0 — 自动装配 + ORM 注解 + LLVM Milestone 2 Phase 2/3

> 预计: 2026-Q3 | 工作量: 8 周

### P0 — 注解处理器自动装配

KylixBoot 框架的注解需要自动绑定到 DI/路由层（v3.1 完成了 AST + 运行时，v3.2 把它们连接起来）。

- [ ] 编译时扫描 `[Controller('/path')]` 类 → 自动调用 `boot.GET('/path/...', handler)`
- [ ] `[Get('/sub')]` / `[Post]` / `[Put]` / `[Delete]` 方法注解 → 注册到全局路由表
- [ ] `[Inject]` 字段 → 编译时生成 DI 容器 `Resolve(typeOf(Field))` 调用
- [ ] `[Component]` / `[Service]` 类 → 自动注册到容器
- [ ] 错误处理：注解参数缺失/重复路径的编译期诊断

### P0 — ORM 注解

声明式数据访问层，对接现有 `stdlib/orm`。

- [ ] `[Entity('table_name')]` 类注解 → 自动表映射
- [ ] `[Column('col')]` / `[PrimaryKey]` / `[AutoIncrement]` 字段注解
- [ ] `[Repository]` 类注解 → 自动生成 CRUD 方法
- [ ] `[Query('SELECT ...')]` 方法注解 → 编译为参数化 SQL
- [ ] 支持 SQLite / PostgreSQL / MySQL
- [ ] 从 `[Entity]` 自动推导迁移 SQL

### P1 — LLVM Milestone 2 Phase 2 (接口 fat pointer)

- [ ] 接口 codegen：`{ ptr vtable, ptr data }` fat pointer
- [ ] 每个接口方法生成 thunk
- [ ] 接口断言 `obj is IFoo` / `obj as IFoo` 的 LLVM 指令
- [ ] 工作量：1–2 周

### P1 — LLVM Milestone 2 Phase 3 (泛型单态化)

- [ ] 模板展开：每个 `TBox<Integer>`、`TBox<String>` 生成独立函数/类型
- [ ] AST 克隆 + 类型参数替换
- [ ] 工作量：2–3 周

### P1 — 校验注解

- [ ] `[Required]` —— 字段非空校验
- [ ] `[Min(n)]` / `[Max(n)]` —— 数值边界
- [ ] `[MinLen(n)]` / `[MaxLen(n)]` —— 字符串长度
- [ ] `[Email]` / `[Regex(pattern)]` —— 格式校验
- [ ] 自动 400 响应

### P2 — 安全注解

- [ ] `[Authenticated]` —— 要求登录
- [ ] `[Role('admin')]` —— 角色校验
- [ ] JWT 令牌生成与验证

### P2 — 包注册中心部署 ✅ 脚手架已交付

- [x] `registry/deploy/` 脚手架（Dockerfile / docker-compose / nginx.conf / Makefile / CI workflow）
- [x] DNS + TLS 后 `make up` 即可上线
- [ ] 搜索索引、全文检索
- [ ] 包统计仪表板
- [ ] GitHub Actions 自动发布 workflow

### P2 — stdlib Phase 6 (网络/加密/编码) ✅ 已完成

- [ ] `net` — TCP/UDP 客户端、HTTP 代理、DNS 查询
- [ ] `crypto` — SHA256/MD5/HMAC/AES/BCrypt
- [ ] `encoding` — Base64/Hex/CSV/URL/JSON-Lines
- [ ] `os` — 进程管理、信号、管道、环境变量

## 📋 v4.0.0 — LLVM M3 + stdlib Phase 7 + IDE 插件

> 预计: 2026 Q3 | 工作量: 2–3 月

### 目标

三条并行主线：LLVM 后端成熟化、stdlib 扩展、开发者工具链。Go 后端持续保留；脱离 Go 是 v5.0+ 的长期目标。

### 主线 1: LLVM 后端 Milestone 3

- [x] 异常处理 codegen（try/except/finally + raise → setjmp/longjmp + 全局异常槽）✅ v4.0 M3
- [x] 字符串插值 codegen（`${expr}` → malloc 缓冲 + strcat/snprintf）✅ v4.0 M3
- [ ] 闭包 codegen（捕获变量的内存布局）
- [ ] 全量自举测试（Go 后端 vs LLVM 后端输出比对）
- [ ] 优化通道：循环展开、内联、常量折叠

### 主线 2: stdlib Phase 7

- [x] `http` 模块：HTTP 客户端（GET/POST/PUT/DELETE，连接池，超时，响应对象）✅ v4.0 Phase 7
- [x] `websocket` 模块：WebSocket 客户端 + 服务端（RFC 6455，纯 stdlib，ping/pong 自动应答）✅ v4.0 Phase 7
- [ ] `httpserver` 模块：高性能 HTTP 服务器（纯 stdlib，不依赖 net/http wrapper）
- [x] `db` 模块：数据库便捷封装 + 连接池（SQLite/MySQL/PostgreSQL，参数化查询）✅ v4.0 Phase 7
- [x] `cache` 模块：内存 LRU 缓存 + TTL（线程安全，Sweep 惰性回收）✅ v4.0 Phase 7
- [ ] `cache` Redis 适配器（远程缓存，依赖 redis client）
- [ ] `websocket` 模块：WebSocket 服务端支持

### 主线 3: IDE 插件

- [x] **VS Code 扩展 v1.1**：语法高亮（含 KylixBoot 注解 + stdlib 函数）+ LSP 集成 + 编译/运行命令 + 快捷键 + 状态栏 + 编译器路径解析 ✅ v4.0
- [ ] **JetBrains 插件**：IntelliJ / GoLand 支持
- [ ] LSP 增强：补全精度提升、重构支持（改名、提取函数）
- [ ] DAP 调试适配器（配合 VS Code 断点调试）

### 长期愿景（v5.0+）

自研运行时 KylixRT，完全脱离 Go 工具链：
```
.klx → LLVM IR → 原生二进制（无 Go 依赖）
```
前提：LLVM 后端能编译全部 Kylix stdlib（M3 完成后评估）。

---

## 🐛 已知问题跟踪

### 编译器层

| ID | 问题 | 严重度 | 状态 |
|----|------|--------|------|
| KLX-C01 | `var p: TClass` 生成 `interface{}` 导致字段不可访问 | 高 | ✅ v3.1.0 |
| KLX-C02 | 字符串插值 `${var}` 不展开 | 中 | ✅ v3.1.0 |
| KLX-C03 | 匿名函数 `function(x):T` 返回类型丢失 | 中 | ✅ v3.1.0 |
| KLX-C04 | `match` 语句生成无效 Go 代码 | 高 | ✅ v3.1.0 |
| KLX-C05 | `uses` 模块在 `program` 中符号不可见 | 高 | ✅ v3.1.0 |
| KLX-G01 | `example21_generic_class` 运行时异常（泛型实例化） | 中 | 🔲 v3.2 |
| KLX-M01 | `example33_use_module` 多文件 unit 编译问题 | 中 | 🔲 v3.2 |

### LLVM 后端层

| ID | 问题 | 状态 |
|----|------|------|
| KLX-L01 | 数组 codegen（array of T / array[1..N]）| ✅ v3.1.0 (Phase 1) |
| KLX-L02 | 接口 fat pointer | 🔲 v3.2 (Phase 2) |
| KLX-L03 | 泛型单态化 | 🔲 v3.2 (Phase 3) |
| KLX-L04 | 异常处理（try/except）| 🔲 v3.2+ |
| KLX-L05 | 优化 Pass | ✅ v3.1.0（`--llvm-opt=N`）|

---

## 🎓 社区与生态

### 近期 (v3.1)
- [ ] 发布 v3.0.0-alpha 公告
- [ ] 创建 Discord 社区
- [ ] 完整教程 v2（29→50+ 示例）
- [ ] Spring Boot 式框架预览

### 中期 (v3.2)
- [ ] 包注册中心 kylix.top/packages 上线
- [ ] GitHub Actions Kylix 模板
- [ ] 企业级项目模板库

### 长期 (v4.0)
- [ ] 会议演讲 / 工作坊
- [ ] 企业赞助 / 基金会
- [ ] 学术论文（类型系统 + LLVM 后端设计）

---

## 设计决策记录

1. **KylixBoot 框架策略**: 第一阶段用代码生成实现注解语义（无运行时反射），第二阶段等 LLVM 后端成熟后迁移到编译时属性处理。

2. **LLVM 后端分阶段策略**: M1（基础）→ M2（完整语言）→ M3（优化）三阶段，Go 后端始终保留作 fallback。

3. **v4.0 脱离 Go 的时机**: 等 LLVM 后端能编译 Kylix 标准库（所有 stdlib）之后再正式宣布 Go 独立。不提前承诺。

4. **编译器 bug 修复优先级**: `uses` 在 program 中的符号注入 > `var p: TClass` 类型推导 > 字符串插值 > match codegen > lambda 返回类型。
