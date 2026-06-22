# Kylix 开发任务清单

> 最后更新: 2026-06-22  
> 当前版本: v3.0.0-alpha  
> 关联文档: [ROADMAP.md](ROADMAP.md) · [CHANGELOG.md](CHANGELOG.md)

---

## ✅ v3.0.0-alpha 已完成 (2026-06-21)

### stdlib Phase 4 — 纯 Kylix 化
- [x] `stdlib/src/jsonutil.klx` — 完整 JSON 解析器（嵌套支持），29 测试
- [x] `stdlib/src/regex.klx` — 字符级验证（IsEmail/IsURL/IsIPv4/IsPhone/IsDate），19 测试
- [x] `stdlib/src/datetime.klx` — FormatPattern/DateAdd/DateSub/IsLeapYear/DaysInMonth，21 测试
- [x] 更新声明文件 `stdlib/klx/regex.klx` `stdlib/klx/datetime.klx`

### 编译器 Bug 修复
- [x] `ast.FunctionDecl.IsExternal` 新字段
- [x] parser: `EXTERNAL` 修饰词识别
- [x] generator: IsExternal=true 跳过 body 生成
- [x] 8 个新测试（3 parser + 5 generator）
- [x] `typeExprName()` 辅助函数修复 `TokenLiteral()` 类型名获取 bug

### 包注册中心
- [x] `registry/` 独立 Go module（SQLite + PostgreSQL 接口）
- [x] REST API（GET/POST packages，版本列表，下载）
- [x] Bearer token 认证 middleware
- [x] Web 前端（htmx + Tailwind CSS）
- [x] `kylix publish` CLI 命令（tarball + 上传）
- [x] 7 个集成测试

### WASI 支持
- [x] `kylix build --wasi`（GOOS=wasip1 GOARCH=wasm）
- [x] `kylix build --wasi --tinygo`
- [x] `pkg/wasi/`：syscall 层（stub + wasip1 两套实现）
- [x] `stdlib/src/wasi.klx` + `stdlib/klx/wasi.klx`
- [x] `examples/wasi-hello/` + `examples/cloudflare-worker/`
- [x] 8 个单元测试

### LLVM 后端 Milestone 1
- [x] `pkg/llvmgen/codegen.go` — module/SSA/字符串常量池
- [x] `pkg/llvmgen/expr.go` — 标量类型/算术/比较/逻辑/WriteLn/IntToStr/Length
- [x] `pkg/llvmgen/stmt.go` — if/while/for/repeat/变量 alloca/函数定义
- [x] `pkg/llvmgen/class.go` — struct type/vtable/方法/GEP 字段访问/malloc 构造
- [x] `pkg/llvmgen/compile.go` — FindLLVM + CompileToNative（.ll→.o→binary）
- [x] `kylix build --backend=llvm` CLI 集成
- [x] 端到端验证：Hello World + 算术 + while 循环 → 原生二进制
- [x] 24 个单元测试（含 6 个类 codegen 测试）

### stdlib Phase 5 — HTTP 客户端
- [x] `stdlib/http_client.go`（THttpClient/HttpGet/HttpPost）
- [x] `stdlib/klx/httpclient.klx` 声明文件
- [x] `stdlib/src/httpclient.klx` 纯 Kylix 包装

### 入门教程
- [x] `examples/complete-tutorial/` — 29 个示例（27 完全工作）
- [x] `docs/GETTING_STARTED.md` + `docs/GETTING_STARTED_CN.md`
- [x] 完整 README_CN.md 中文教程

---

## 📋 v3.1.0 待完成任务

### 优先级 P0 — 编译器 Bug 修复（解锁后续所有示例）

| ID | 任务 | 影响 |
|----|------|------|
| KLX-C05 | `uses X` 在 `program` 中符号注入（strutil/mathutil/sysutil/jsonutil/datetime/regex 均受影响）| 解锁 30+ 特性 |
| KLX-C01 | `var p: TClass` 生成 `*TClass` 而非 `interface{}`（解锁 OOP 声明式变量）| 解锁 OOP 示例 |
| KLX-C04 | `match` 语句完整代码生成 | 解锁 match 示例 |
| KLX-C02 | 字符串插值 `${var}` 展开 | 解锁字符串示例 |
| KLX-C03 | 匿名函数 `function(x): T` 返回类型生成 | 解锁 lambda 示例 |

每个 bug 修复后，对应补充教程示例。

### 优先级 P1 — KylixBoot 框架

**目标**: Spring Boot 式 Web 框架，声明式注解驱动。

**阶段 1: 注解语法支持（编译器层）**
- [ ] 解析 `[Attribute]` 语法（Lexer + Parser 扩展）
- [ ] `ast.AttributeDecl` 节点
- [ ] 注解附加到类/方法/字段声明

**阶段 2: 核心容器（代码生成层）**
- [ ] `kylix.boot/di` — 依赖注入容器
  - [ ] `[Component]` / `[Service]` / `[Repository]` 注册
  - [ ] `[Inject]` 字段注入
  - [ ] 生命周期管理（singleton/scoped）
- [ ] `kylix.boot/config` — 配置绑定
  - [ ] `[Configuration]` 类注解
  - [ ] `[Value('key', default)]` 字段绑定
  - [ ] 支持 .env / YAML / 环境变量

**阶段 3: Web 层**
- [ ] `kylix.boot/router` — 路由注册
  - [ ] `[Controller('/base')]` 类注解
  - [ ] `[Get('/path')]` / `[Post]` / `[Put]` / `[Delete]` 方法注解
  - [ ] 路径参数 `[Get('/users/:id')]`
- [ ] `kylix.boot/middleware` — 中间件链
  - [ ] `[Middleware]` 注解
  - [ ] 内置：Logger / CORS / Auth / RateLimit
- [ ] `kylix.boot/validation` — 参数校验
  - [ ] `[Required]` / `[Min(n)]` / `[Max(n)]` / `[Email]` / `[Regex]`
  - [ ] 自动 400 响应

**阶段 4: 数据层**
- [ ] `kylix.boot/orm` — 声明式 ORM
  - [ ] `[Entity('table_name')]` 类注解
  - [ ] `[Column('col_name')]` 字段注解
  - [ ] `[Repository]` 类注解，自动生成 CRUD
  - [ ] `[Query('SELECT ...')]` 自定义查询
  - [ ] 支持 SQLite / PostgreSQL / MySQL
- [ ] `kylix.boot/migrate` — 数据库迁移
  - [ ] 从 `[Entity]` 自动生成迁移 SQL

**阶段 5: 安全层**
- [ ] `kylix.boot/security` — 认证授权
  - [ ] `[Authenticated]` — 要求登录
  - [ ] `[Role('admin')]` — 角色校验
  - [ ] JWT 令牌生成与验证
- [ ] `kylix.boot/cache` — 缓存
  - [ ] `[Cacheable(key, ttl)]` 方法缓存
  - [ ] `[CacheEvict(key)]` 缓存清除
  - [ ] 内存缓存 + Redis 适配

**阶段 6: 开发工具**
- [ ] `kylix boot new myapp` — 脚手架命令
- [ ] `kylix boot run` — 热重载开发服务器
- [ ] `kylix boot build` — 生产构建（静态链接）
- [ ] `kylix boot test` — 集成测试框架
- [ ] `kylix boot generate entity User` — 代码生成

### 优先级 P2 — LLVM 后端 Milestone 2

- [ ] 静态数组（`array[1..N] of T` → `[N x T]`）
- [ ] 动态数组（`array of T` → `{ ptr, len, cap }` 结构体）
- [ ] 接口 codegen（`{ ptr vtable, ptr data }` fat pointer）
- [ ] 泛型单态化（模板展开）
- [ ] 异常处理（`invoke` + `landingpad`）
- [ ] LLVM 优化 Pass（`-O0` / `-O1` / `-O2` 选项）
- [ ] 交叉编译（`--backend=llvm --target=linux/amd64`）

### 优先级 P3 — 教程完善（依赖 P0 bug 修复）

修复 P0 后立即补充以下示例：

- [ ] `example26_string_interp.klx` — 字符串插值 `${var}`
- [ ] `example15_lambda_fn.klx` — 带返回值的匿名函数
- [ ] `example34_match.klx` — match 模式匹配
- [ ] `example35_interface.klx` — 接口声明与实现
- [ ] `example36_strutil.klx` — strutil 模块
- [ ] `example37_mathutil.klx` — mathutil 模块
- [ ] `example38_arrayutil.klx` — arrayutil 模块
- [ ] `example39_sysutil.klx` — 文件 I/O
- [ ] `example40_jsonutil.klx` — JSON 解析
- [ ] `example41_datetime.klx` — 日期时间
- [ ] `example42_regex.klx` — 正则表达式验证
- [ ] `example43_httpclient.klx` — HTTP 客户端
- [ ] `example44_web_server.klx` — Web 服务器
- [ ] `example45_kylix_test.klx` — 测试框架
- [ ] `example46_wasi.klx` — WASI 编译示例
- [ ] `example47_llvm.klx` — LLVM 后端示例
- [ ] `example48_async.klx` — async/await

### 优先级 P4 — 包注册中心部署

- [ ] 部署到 kylix.top/packages（VPS + PostgreSQL + TLS）
- [ ] 域名配置 packages.kylix.top
- [ ] GitHub Actions 自动发布 workflow
- [ ] 搜索与分类功能

---

## 📋 v3.2.0 待完成任务

### 编译器健壮性全面修复

- [ ] 所有 KLX-C0x bug（见 ROADMAP.md 清单）
- [ ] 类型检查覆盖率提升至 80%+
- [ ] 错误恢复：所有解析错误后继续报告下一个错误
- [ ] 更好的错误信息（含建议修复）

### stdlib 全量可用

stdlib 所有模块在 `program` 文件中 `uses X` 后可直接调用，无需命名空间前缀。

- [ ] strutil（8 函数）
- [ ] mathutil（12 函数）
- [ ] arrayutil（8 函数）
- [ ] collections（TIntList）
- [ ] sysutil（ReadFile/WriteFile/FileExists 等）
- [ ] jsonutil（JsonDecodeMap/JsonEncode 等）
- [ ] datetime（Now/MakeDate/FormatPattern 等）
- [ ] regex（IsEmail/IsURL/IsNumeric 等）
- [ ] httpclient（HttpGet/HttpPost/THttpClient）
- [ ] wasi（WriteLn/ReadLine/GetEnv/Args）

### stdlib Phase 6 — 网络与加密

- [ ] `net` — TCP/UDP 客户端，HTTP 代理
- [ ] `crypto` — SHA256/MD5/HMAC/AES/BCrypt
- [ ] `encoding` — Base64/Hex/CSV/URL/JSON-Lines
- [ ] `os` — 进程/信号/管道

---

## 📋 v4.0.0 长期任务

详见 ROADMAP.md v4.0.0 章节。核心：
- [ ] 自研运行时 KylixRT（GC + 字符串 + 动态数组）
- [ ] LLVM 后端 Milestone 3（完整 Kylix 语言）
- [ ] 自举编译器 v2.0（用 Kylix 写 LLVM 后端）
- [ ] 完整 IDE 支持（DAP + VS Code v2 + IntelliJ）

---

## 关键设计决策

1. **KylixBoot 注解实现**: 第一版通过代码生成器处理（编译时元编程），不依赖运行时反射。注解 `[Get('/')]` 在编译时展开为路由注册代码。

2. **bug 修复优先级原则**: 解锁最多特性的 bug 优先。`uses` 符号注入修复后可解锁 10+ stdlib 模块示例，优先级最高。

3. **教程示例原则**: 每个示例必须 `kylix build + go run` 全程无错，不允许"仅供参考"的示例进入 `examples/complete-tutorial/`。

4. **v4.0 脱离 Go 时机**: 等 LLVM 后端能编译完整 stdlib + 自身 LLVM 后端代码后，再启动 v4.0 工作。预计 2027 年。
