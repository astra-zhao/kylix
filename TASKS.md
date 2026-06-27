# Kylix 开发任务清单

> 最后更新: 2026-06-26  
> 当前版本: v3.2.0-dev  
> 关联文档: [ROADMAP.md](ROADMAP.md) · [CHANGELOG.md](CHANGELOG.md)

---

## ✅ v3.1.0 已完成 (2026-06-23)

### KylixBoot 框架核心运行时
- [x] `pkg/boot/types.go` — Request / Response / Handler / Middleware
- [x] `pkg/boot/router.go` — 路由匹配 + 路径参数 (`/users/:id`)
- [x] `pkg/boot/server.go` — HTTP 服务器 + 优雅停机
- [x] `pkg/boot/di.go` — DI 容器（Singleton / Transient / Instance + 反射 Inject）
- [x] `pkg/boot/app.go` — App + 全局快捷方式（boot.GET / POST / Use / Listen）
- [x] `pkg/boot/config.go` — 配置（环境变量回退）
- [x] `pkg/boot/middleware.go` — Logger / Recover / CORS / Auth / RateLimit / RequestID
- [x] `pkg/boot/boot_test.go` — 23/23 测试通过
- [x] `stdlib/boot_bridge.go` — 桥接为 `stdlib.BootXxx`
- [x] `stdlib/klx/boot.klx` — LSP 声明文件
- [x] generator 注册 `boot` 模块到 stdlib 分发器

### 注解语法支持
- [x] `ast.Attribute` 类型（Name + Args）
- [x] `Attributes []*Attribute` 字段加到 ClassDecl / TypeDecl / FunctionDecl / VarDecl
- [x] `parser/parser_attribute.go` — 解析 `[Name]` 和 `[Name(args...)]`
- [x] 顶层和类体内均可使用
- [x] 新示例 `example41_attributes.klx`

### 编译器修复（KLX-C01..C05）
- [x] **KLX-C01** — `var p: TClass` 生成 `*TClass` 而非 `interface{}`
  - `generator/generator_types.go` 始终为类类型 emit `*TypeName`
  - `generator/generator.go` 扫描 TypeDecl-wrapped ClassDecl 方法体的 import
  - `generator/generator_expr.go` `uses sysutil` 活跃时跳过 `os.ReadFile` 内联
  - 新示例 `example40_declarative_oop.klx`
- [x] **KLX-C02** — `lexer/lexer.go` 单引号字符串中的 `${...}` emit STRING_INTERPOLATION
- [x] **KLX-C03** — `ast.LambdaExpression` 新增 ReturnType 字段；parser 保存；generator emit 返回类型 + `var result T` + `return result`
- [x] **KLX-C04** — `match` 生成 tagless `switch { case _v == p: }`（不再是无效的 `switch _v := ... { case _v == 1: }`）
- [x] **KLX-C05** — `uses` 在 program 中符号注入
  - Generator 新增 `usedModules map[string]bool`
  - `generator/generator_stdlib.go`（~270 行）映射 stdlib 模块名 → 函数集
  - `resolveStdlibFunc()` 检查函数归属
  - `generateStdlibCall()` emit `stdlib.FuncName(...)`
  - 返回 `(T, error)` 的函数包装为具体返回类型
  - 解锁 40+ stdlib 函数在 program 文件中使用

### LLVM Milestone 2 Phase 1
- [x] `pkg/llvmgen/array.go`（~200 行）—— 静态 `array[1..N] of T` → `alloca [N x T]`
- [x] 动态 `array of T` → `{ ptr, i64, i64 }` slice 结构体
- [x] Pascal 1-based 索引转 LLVM 0-based
- [x] `pkg/llvmgen/array_test.go` —— 6 个新测试（LLVM 测试总数 30）
- [x] 编译期常量求值（`array[1..N]` 处理 `((N-1)+1)`）
- [x] `CompileOpts.OptLevel` + `--llvm-opt=0/1/2/3` CLI 标志
- [x] `llc` 传入 `-O=N`
- [x] `emitMain` 在 `main()` 中分配顶层 VarDecl
- [x] Generator 新增 `program *ast.Program` 字段

### 教程扩展
- [x] `examples/complete-tutorial/example40_declarative_oop.klx`
- [x] `examples/complete-tutorial/example41_attributes.klx`
- [x] 教程示例 35/35 检查通过（34 个 `example*.klx` + `math_helper.klx`）

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

### ✅ v3.1.1 Hotfix 已完成 (2026-06-25)

- [x] **KLX-M01** — Unit `interface` / `implementation` parsing修复
  - `token/token.go` 新增 `IMPLEMENTATION` token 与关键字映射
  - `parser/parser.go` 将 unit 顶层 `interface` / `implementation` 作为 section marker 跳过
  - implementation 段函数现在生成为顶层 `FunctionDecl` 且保留函数体
  - `generator/generator_types.go` 跳过 bodiless forward declarations，并防御空名 interface
  - `pkg/compiler/unit_sections_test.go` 覆盖多文件 unit 构建
- [x] **KLX-G01** — 泛型类方法 receiver codegen 修复
  - `generator/generator.go` 预扫描记录 `classTypeParams`
  - `generator/generator_types.go` 生成 `func (self *TStack[T]) ...` 而不是 `func (self *TStack) ...`
  - 泛型属性 accessor receiver 同步支持类型参数
  - `generator/generator_generics_test.go` 覆盖泛型类方法 receiver
  - `example21_generic_class.klx` 改为显式 `self.Field` 访问
- [x] `examples/complete-tutorial/test_all.sh` 覆盖所有教程目录 + modules
- [x] 教程验证：35/35 通过（34 个 `example*.klx` + `math_helper.klx` unit companion）
- [x] 增量缓存加入 `CacheVersion`，避免 codegen 改动后复用旧 fragment

---

## ✅ v3.2.0-dev 已完成 (2026-06-26) — KylixBoot 注解栈

KylixBoot 在 v3.1 完成了运行时 + 注解 AST，v3.2-dev 把它们全部连接起来并加上编译期诊断。

### P0 — 注解自动装配
- [x] 编译时扫描 `[Controller('/path')]` 类，自动生成路由注册代码
- [x] `[Get]` / `[Post]` / `[Put]` / `[Delete]` 方法注解 → 注册到全局路由表
- [x] 路径合成：`Controller.path + Method.path` 完整 URL
- [x] `[Component]` / `[Service]` 类 → 自动注册到容器（singleton + `BootRegisterInstance`）
- [x] `[Inject]` 字段 → 编译时按类型生成直接赋值
- [x] 注解参数缺失/重复路径/不支持签名的编译期诊断（`KLX207`–`KLX210`）
- [x] procedure 风格 handler `procedure M(req; res)` 支持 + `Response.Send/StatusCode`

### P1 — 校验注解
- [x] `[Required]` — 字段非空
- [x] `[Min(n)]` / `[Max(n)]` — 数值边界
- [x] `[MinLen(n)]` / `[MaxLen(n)]` — 字符串长度
- [x] `[Email]` — 格式校验
- [x] 生成 `Validate()` / `IsValid()` 方法 + `KLX211 ErrInvalidValidation`

### P2 — 安全注解
- [x] `[Authenticated]` — `BootEnforceAuth` 守卫（401）
- [x] `[Role('admin')]` — `BootEnforceRole` 守卫（403，隐含认证）
- [x] 运行时 `pkg/boot/security.go`：`RegisterAuthValidator` / `RegisterRolesProvider` / `EnforceAuth` / `EnforceRole`
- [x] `KLX212 ErrInvalidSecurity`

### P0 — ORM 注解
- [x] `[Entity('table_name')]` 类注解
- [x] `[Column('col_name')]` / `[PrimaryKey]` 字段注解
- [x] 生成 `ToRow()` / `FromRow()` 映射方法
- [x] `[Repository(TEntity)]` → 自动生成 CRUD（`FindAll` / `FindById` / `Save` / `DeleteById`）
- [x] `[Query('SELECT ...')]` 方法注解 → 参数化 SQL（单行 / 数组返回）
- [x] `KLX213 ErrInvalidORM`

### 教程与测试
- [x] 新增 `example42..47`（route / DI / procedure / validation / security / ORM）
- [x] 教程 41/41 通过
- [x] generator / pkg/compiler / pkg/boot 全部测试通过

---

## 📋 v3.2.0 剩余任务

### 优先级 P1 — LLVM Milestone 2 Phase 2 (接口) ✅ 已完成 (2026-06-27)

- [x] 接口 codegen：`%IFoo_iface = type { ptr, ptr }` fat pointer + `%IFoo_vtable` 类型
- [x] 每个接口在实现类上 emit 独立的 `@TFoo_IFoo_vtable` 常量
- [x] `MemberExpression` 字段访问 codegen（之前完全缺失）
- [x] `obj.Method(args)` 直接分发（class）+ 通过 vtable 槽位间接分发（interface）
- [x] `iface := obj` / `iface := obj as IFoo` 装箱
- [x] `obj is IFoo` / `obj as IFoo` 表达式
- [x] `pkg/llvmgen/interface_test.go` — 8 个 IR fragment 测试

### 优先级 P1 — LLVM Milestone 2 Phase 3 (泛型) ✅ 已完成 (2026-06-27)

- [x] 泛型类模板注册：`ClassDecl.TypeParams` 非空时挂到 `genericTemplates`，跳过直接 emit
- [x] AST collector：遍历 program 收集所有 `*ast.GenericType` 实例化点
- [x] 名称改写：`TBox<Integer>` → `TBox_Integer`，`TPair<Integer, String>` → `TPair_Integer_String`
- [x] AST 克隆 + 类型参数替换（支持 `Identifier` / `GenericType` / `ArrayType.ElementType`）
- [x] 特化类复用 `emitClassDecl`，自动得到结构体 + vtable + 方法
- [x] `var x: TBox<Integer>` / `TBox<Integer>.Create` / 裸 `TFoo.Create` 全部路由到 `emitConstructor(mangled)`
- [x] `pkg/llvmgen/generics_test.go` — 6 个 IR fragment 测试



- [ ] 泛型单态化：每个 `TBox<Integer>` / `TBox<String>` 生成独立函数/类型
- [ ] AST 克隆 + 类型参数替换
- [ ] 与现有 monomorphization 表配合

### 优先级 P2 — 包注册中心部署 ✅ 已交付部署脚手架

- [x] `registry/deploy/` 脚手架（Dockerfile / docker-compose / nginx.conf / .env.example / Makefile / README）
- [x] `.github/workflows/registry.yml` — CI 镜像构建 + 推送
- [x] DNS + VPS 后 `make up` 即可上线

### 优先级 P2 — stdlib Phase 6 ✅ 已完成 (2026-06-27)

- [x] `net` — TCP/UDP 客户端、DNS 查询
- [ ] `crypto` — SHA256/MD5/HMAC/AES/BCrypt
- [ ] `encoding` — Base64/Hex/CSV/URL/JSON-Lines
- [ ] `os` — 进程/信号/管道

### 优先级 P3 — 残留 Bug（v3.1 未覆盖）

- [x] **KLX-G01** — `example21_generic_class` 运行时异常 → ✅ v3.1.1 修复（泛型 receiver + 显式 self 字段访问）
- [x] **KLX-M01** — `example33_use_module` 多文件 unit 编译路径问题 → ✅ v3.1.1 修复（unit section marker 解析 + forward declaration 跳过）

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
