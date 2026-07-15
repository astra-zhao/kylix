# Kylix Development Roadmap

> 最后更新: 2026-07-15  
> 当前版本: v4.9.0 ✅  
> 官网: [kylix.top](https://kylix.top)  
> 目标: Kylix 成为生产级、多后端、全栈 Pascal 语言

**✅ v4.9.0 已发布！** DWARF 调试信息 Phase 2 —— 类方法/lambda 注册独立 DISubprogram（define 行附 `!dbg`、`self`/参数/捕获变量声明为调试局部变量，v4.8.0 泛型类方法可逐行单步）+ DILexicalBlock（块内 `var` 归属正确的嵌套作用域）+ jsonutil `JsonGetArray` 从返回 null 的 stub 升级为真实解析器（字符串数组 slice `{ptr,i64,i64}` + `JsonArrayLen`/`JsonArrayGetString`）+ 顺手修复 `skip_nested` 丢失闭合 `]`/`}` 的 off-by-one。LLVM 测试 250→**255**，教程通过率 **49/49 (100%)** 无回归。详见 [CHANGELOG.md](CHANGELOG.md)。

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
| **v4.0.0** | LLVM M3（异常/控制流/表达式）+ stdlib Phase 7 + VS Code v1.1 | ✅ 完成 | 2026-07-01 |
| **v4.1.0** | LLVM M4 高级特性（闭包/多返回值/inherited/优化） | ✅ 完成 | 2026-07-02 |
| **v4.2.0** | LLVM stdlib Phase 1 (sysutil 模块) | ✅ 完成 | 2026-07-03 |
| **v4.3.0** | LLVM stdlib Phase 1 (datetime 模块 + Arena Allocator) | ✅ 完成 | 2026-07-03 |
| **v4.4.0** | LLVM stdlib Phase 2 (8 模块 + KylixBoot 注解支持) | ✅ 完成 | 2026-07-07 |
| **v4.5.0** | LLVM stdlib Phase 3 + 优化 pass + 增量缓存 + DWARF 调试符号 | ✅ 完成 | 2026-07-08 |
| **v4.6.0** | DWARF 逐行调试（per-instruction DILocation + DILocalVariable） | ✅ 完成 | 2026-07-10 |
| **v4.7.0** | 静态数组下界修复 + jsonutil 嵌套对象解析 | ✅ 完成 | 2026-07-10 |
| **v4.8.0** | 泛型类方法 codegen + 类字段数组 GEP + DIBasicType 多类型 | ✅ 完成 | 2026-07-14 |
| **v4.9.0** | DWARF Phase 2（类方法/lambda DISubprogram + DILexicalBlock）+ jsonutil 嵌套数组 | ✅ 完成 | 2026-07-15 |
| **v5.0.0** | 自研运行时 KylixRT + Variant 运行时 + JetBrains 插件 + 自举编译器 | 📋 长期 | 2027+ |

---

## 📊 累计统计 (v4.0-dev)

| 指标 | 数量 |
|------|------|
| Go 测试包 | 16 个（全部通过）|
| Go 级测试 | ~350+ |
| Kylix 级 stdlib 测试 | 140+ 个（14 模块）|
| 纯 Kylix stdlib 函数 | 110+ |
| CLI 命令 | 19 个 |
| 教程示例 | 55+ 个（含 stdlib Phase 7 新增）|
| 原生构建目标 | 5 (linux/darwin/windows × amd64/arm64) |
| WASM 目标 | 2 (Go 标准 + TinyGo) |
| WASI 目标 | 2 (Go wasip1 + TinyGo) |
| LLVM 后端 | ✅ Milestone 4（stdlib Phase 3 完成，3 模块真实化）|
| LLVM 测试 | 255 个（含 stdlib 60+ 个 + debug/DCE/cache 25 个 + 数组下界 2 个 + DIBasicType 1 个 + 方法/lambda subprogram + lexical block + JsonGetArray）|
| LLVM 教程编译通过率 | 48/48（100%，含 example33 多文件模块；example23 修复后正确运行；example21 泛型类输出正确）|
| LLVM 增量缓存 | ✅ v4.5.0（llc 跳过，32x 加速）|
| LLVM 调试符号 | ✅ v4.6.0/v4.8.0/v4.9.0（DWARF `-g` flag，逐行单步 + 变量检视 + per-llvmType DIBasicType + 方法/lambda DISubprogram + DILexicalBlock）|
| LLVM 静态数组 | ✅ v4.7.0（真实 LowerBound，array[0..N]/array[1..N]/array[5..N] 均正确）|
| LLVM 泛型类 | ✅ v4.8.0（TStack<T>.Create() → Push/Pop 完整 codegen，example21 输出正确）|
| LLVM jsonutil | ✅ v4.7.0/v4.9.0（嵌套对象 JsonGetMap + 嵌套数组 JsonGetArray/JsonArrayLen/JsonArrayGetString）|
| KylixBoot 测试 | 23 个 |
| 包注册中心 | ✅ REST API + Web 前端 |
| VS Code 扩展 | ✅ v1.1（语法高亮 + LSP + 代码片段 25 个）|

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

## ✅ v4.0.0 — LLVM M3 + stdlib Phase 7 + IDE 插件

> 发布: 2026-07-01 | 状态: 已发布

### 目标

三条并行主线：LLVM 后端成熟化、stdlib 扩展、开发者工具链。Go 后端持续保留；脱离 Go 是 v5.0+ 的长期目标。

### 主线 1: LLVM 后端 Milestone 3

- [x] 异常处理 codegen（try/except/finally + raise → setjmp/longjmp + 全局异常槽 + on E: Type 子类型匹配）✅ v4.0 M3
- [x] 字符串插值 codegen（`${expr}` → malloc 缓冲 + strcat/snprintf）✅ v4.0 M3
- [x] 控制流语句补全（break/continue/case/match/foreach + 循环标签保存/恢复）✅ v4.0 M3
- [x] 表达式覆盖提升（WriteLn 多参数/零参数、ArrayLiteral、SliceExpression、TupleLiteral、AwaitExpression）✅ v4.0 M3
- [x] 关键 bug 修复（多变量声明、类型自动转换 i1↔i64↔double、__kylix_is_subtype SSA dominance）✅ v4.0 M3
- [x] 元组 LHS 赋值 stub（多返回值 `(q, r) := func()` 降级为注释，IR 仍合法）✅ v4.0 M3
- [x] **14/15 基础教程通过 LLVM 编译到原生二进制**（example15_lambda 因闭包架构限制预期失败）✅ v4.0 M3

> 闭包、完整多返回值、inherited、优化通道属 v4.1.0（M4），见下文。

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
- [x] **VS Code 代码片段**：25 个片段（program/unit、function/procedure、class/record、控制流、try/except、WriteLn、KylixBoot controller/routes、ORM entity）✅ v4.0
- [ ] **JetBrains 插件**：IntelliJ / GoLand 支持
- [ ] LSP 增强：补全精度提升、重构支持（改名、提取函数）
- [ ] DAP 调试适配器（配合 VS Code 断点调试）

---

## ✅ v4.1.0 — LLVM M4 高级特性

> 发布: 2026-07-02 | 状态: 已发布 | 详细计划见 [docs/v4.1.0-plan.md](docs/v4.1.0-plan.md)

### 目标

LLVM 后端达到与 Go 后端的功能对等（常见用例）。**实际交付：27/49 教程通过 LLVM 编译，01-04 章节（19文件）与 Go 后端输出逐字节一致。**

### Priority 1: 闭包/Lambda 支持 ✅

当前 lambda 生成 null ptr stub，导致 example15_lambda 编译失败。

- [x] 捕获变量分析（AST walker 构建 capture list）
- [x] 环境结构体生成（`%__env_f = type { i64, ptr }`）
- [x] 函数指针降级（lambda body → 命名函数 `@__lambda_f(ptr %env, args)`）
- [x] 闭包结构体（`{func_ptr, env_ptr}` pair）+ 调用约定
- [x] 多变量捕获、块体闭包（表达式体 lambda 暂不支持，见已知限制）
- [x] example15_lambda.klx 通过

### Priority 2: 完整多返回值 ✅

当前 `(q, r) := func()` 生成静默注释，变量未初始化。

- [x] 函数返回结构体类型 + insertvalue/extractvalue
- [x] 元组解构赋值正确工作

### Priority 3: inherited 关键字 ✅

父类方法调用不支持。

- [x] 类层次遍历查找方法实际定义类（`DefiningClass` 字段）
- [x] 生成直接函数调用（绕过 vtable）
- [x] 正确传递 self 指针
- [x] 多层继承链测试通过

### 额外交付（原计划外）：04_oop 系统性修复 ✅

- [x] vtable 继承（子类 vtable 含父类方法槽位）
- [x] vtable 函数指针按 `DefiningClass` 生成
- [x] 虚方法调用 void 返回类型签名修复
- [x] `self.Field` 访问崩溃修复（self 参数 vs alloca 变量的 load 语义区分）
- [x] 显式类型变量赋值类型推断修复
- [x] 04_oop 章节 4/4 教程通过，与 Go 后端逐字节一致

### Priority 4: 优化通道 ✅

LLVM 代码比 Go 慢 2–5x，无优化。

- [x] `--llvm-opt` flag（O1/O2/O3 级别）
- [x] `opt` 工具集成（IR 级优化：mem2reg/内联/循环归纳/DCE）+ `llc -O<N>`（codegen 级）
- [x] 基准测试（fib 递归、loop_sum 循环求和、primes 素数筛）
- [x] **实测：loop_sum 20x 提速（循环归纳为闭式常量），fib 1.7x 提速** — 远超原定 1.5x 目标

### 成功指标（实际达成）

| 指标 | v4.0.0 | v4.1.0 目标 | v4.1.0 实际 |
|------|--------|-------------|-------------|
| LLVM 教程通过率 | 14/15 (93%，基础章节) | 25+/35 (71%+) | 27/49，01-04 章节 100% |
| Lambda 支持 | ❌ | ✅ | ✅ |
| 多返回值 | Stub | ✅ | ✅ |
| inherited | ❌ | ✅ | ✅ |
| OOP 字段/方法访问 | 部分崩溃 | — | ✅ 系统性修复 |
| 优化性能 | 2–5x 慢于 Go | 1.5x 以内 | loop_sum 20x / fib 1.7x 提速 |

### 已知限制（v4.1.0）

- 表达式体 lambda（`function(x): T -> expr`）— parser 不识别返回类型后的 `->`
- `inherited` 作表达式（`result := inherited F(x)`）— parser 仅支持语句形式
- stdlib 重度教程（08、13–20 章节）仍需 Go 工具链 — LLVM stdlib 属 v4.2.0 范围

---

## 📋 v4.2.0 — LLVM 工具链深化 + stdlib Phase 8

> 预计: 2026 Q4 | 工作量: 2–3 月

### 目标

LLVM 后端工具链成熟化，补充 stdlib 常用模块，提升开发者体验。

### 主线 1: LLVM 工具链

- [x] **增量编译**：缓存 LLVM IR 每模块，只重编变更文件，链接预编译 .o（✅ v4.5.0）
- [x] **调试符号**：发出 DWARF 调试信息，支持 GDB/LLDB 逐行单步 + 变量检视（`kylix build --backend=llvm -g`）（✅ v4.6.0，per-instruction DILocation + DILocalVariable）
- [ ] **交叉编译**：无需本机安装 LLVM 的目标构建（预编译 IR + 链接器）
- [x] **LLVM stdlib Phase 1**：用 LLVM 编译核心 stdlib 模块（sysutil/datetime，纯 Kylix 无 Go wrapper）（✅ v4.2.0/v4.3.0）

### 主线 2: stdlib Phase 8

- [ ] `logging` 模块：结构化日志（leveled + JSON 输出）
- [ ] `profiling` 模块：CPU/内存 profiling
- [ ] `reflection` 模块：运行时类型信息（基础 RTTI）
- [ ] `httpserver` 模块：高性能 HTTP 服务器（与 web/KylixBoot 分工明确）
- [ ] `cache` Redis 适配器（远程缓存）

### 主线 3: IDE 与生态

- [ ] **JetBrains 插件**：IntelliJ / GoLand 支持
- [ ] LSP 增强：补全精度提升、重构支持
- [ ] DAP 调试适配器（配合 VS Code 断点调试）
- [ ] 包注册中心部署上线（kylix.top/packages）

---

### 长期愿景（v5.0+）

> 预计: 2027+ | 工作量: 6–12 月 | 完全脱离 Go 依赖

自研运行时 KylixRT，完全脱离 Go 工具链：

```
.klx → LLVM IR → 原生二进制（无 Go 依赖）
```

**前提**：LLVM 后端能编译全部 Kylix stdlib（M3/M4 完成后评估）。

#### 核心工作

- [ ] **自研运行时 KylixRT**
  - 垃圾回收器（标记-清扫 或 集成 Boehm GC）
  - 字符串/动态数组/映射（纯 C 实现，替代 Go string/slice/map）
  - 协程库或线程池（替代 goroutine）
- [ ] **stdlib 纯 Kylix 重写**
  - 移除所有 `stdlib/*.go` 包装文件
  - 用 Kylix + C FFI 重写核心功能
- [ ] **自举编译器**
  - Kylix 编译器用 Kylix 重写（当前是 Go）
  - `kylix compile kylix_compiler.klx --backend=llvm`
  - 生成的编译器可编译自己

#### 里程碑标准

- ✅ LLVM 后端可编译所有 stdlib 模块
- ✅ LLVM 后端可编译 Kylix 编译器自身
- ✅ 生成的二进制零 Go 依赖（完全独立）

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
| KLX-G01 | `example21_generic_class` 运行时异常（泛型实例化） | 中 | ✅ v3.1.1 |
| KLX-M01 | `example33_use_module` 多文件 unit 编译问题 | 中 | ✅ v3.1.1（Go）/ ✅ v4.4.0（LLVM） |

### LLVM 后端层

| ID | 问题 | 状态 |
|----|------|------|
| KLX-L01 | 数组 codegen（array of T / array[1..N]）| ✅ v3.1.0 (Phase 1) |
| KLX-L02 | 接口 fat pointer | ✅ v3.2.0 (Phase 2) |
| KLX-L03 | 泛型单态化 | ✅ v3.2.0 (Phase 3) |
| KLX-L04 | 异常处理（try/except）| ✅ v4.0.0 (M3) |
| KLX-L05 | 优化 Pass | ✅ v3.1.0（`--llvm-opt=N`）|

---

## 🎓 社区与生态

### 近期 (v4.0)
- [ ] 发布 v4.0.0 公告（LLVM M3 生产可用）
- [ ] VS Code 扩展发布到 Marketplace
- [ ] 完整教程（55+ 示例，含 Phase 7）
- [ ] LLVM 后端使用指南（docs/llvm-backend.md）

### 中期 (v4.1–v4.2)
- [ ] 包注册中心 kylix.top/packages 上线
- [ ] GitHub Actions Kylix 模板
- [ ] JetBrains 插件（IntelliJ/GoLand）
- [ ] 企业级项目模板库

### 长期 (v5.0+)
- [ ] 会议演讲 / 工作坊
- [ ] 企业赞助 / 基金会
- [ ] 学术论文（类型系统 + LLVM 后端设计）

---

## 设计决策记录

1. **KylixBoot 框架策略**: 第一阶段用代码生成实现注解语义（无运行时反射），第二阶段等 LLVM 后端成熟后迁移到编译时属性处理。

2. **LLVM 后端分阶段策略**: M1（基础）→ M2（完整语言）→ M3（异常/控制流/表达式）→ M4（闭包/多返回值/优化）。Go 后端始终保留作 fallback 与快速开发路径。

3. **LLVM 异常处理路线决策（v4.0 M3）**: 选择路线 C（全局异常槽 + setjmp/longjmp + 类型 ID），而非 Itanium C++ EH ABI（invoke/landingpad/resume + __cxa_*）。理由：手写 IR 文本下 C++ ABI 极易错位，且需链接 libc++abi 违反"仅依赖 libc"约束。setjmp/longjmp 全是 call/load/store/br/icmp，IR 字符串拼接可行；finally 用代码复制（3 份）保证确定性。

4. **双后端并存策略**: Go 后端（默认，工具链成熟、快速开发）与 LLVM 后端（opt-in，原生二进制、体积小）长期并存。不追求单一后端淘汰，按场景选用。

5. **v5.0 脱离 Go 的时机**: 等 LLVM 后端能编译 Kylix 标准库（所有 stdlib）+ 编译器自身之后再正式宣布 Go 独立。不提前承诺。

6. **编译器 bug 修复优先级**: `uses` 在 program 中的符号注入 > `var p: TClass` 类型推导 > 字符串插值 > match codegen > lambda 返回类型。
