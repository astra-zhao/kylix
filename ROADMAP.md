# Kylix Development Roadmap

> 最后更新: 2026-06-22  
> 当前版本: v3.0.0-alpha 🚀  
> 官网: [kylix.top](https://kylix.top)  
> 目标: Kylix 成为生产级、多后端、全栈 Pascal 语言

**🚀 v3.0.0-alpha 发布！** 架构突破 — LLVM 后端 Milestone 1、包注册中心、WASI、stdlib Phase 4。  
**📍 当前重点：** v3.1 — Spring Boot 式 Web 框架 + LLVM Milestone 2 + 教程完善。

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
| **v3.0.0** | LLVM 后端 + 包注册中心 + WASI | 🚀 alpha | 2026-06-21 |
| **v3.1.0** | Spring Boot 式框架 + LLVM M2 + 教程 | 📋 规划中 | 2026-Q3 |
| **v3.2.0** | 编译器修复 + stdlib 完整可用 | 📋 规划中 | 2026-Q3 |
| **v4.0.0** | 自研运行时 + 完全脱离 Go | 📋 长期 | 2027+ |

---

## 📊 累计统计 (v3.0.0-alpha)

| 指标 | 数量 |
|------|------|
| Go 测试包 | 15 个（全部通过）|
| Go 级测试 | ~310+ |
| Kylix 级 stdlib 测试 | 117 个（10 模块）|
| 纯 Kylix stdlib 函数 | 90+ |
| CLI 命令 | 19 个 |
| 教程示例 | 29 个（27 完全工作）|
| 原生构建目标 | 5 (linux/darwin/windows × amd64/arm64) |
| WASM 目标 | 2 (Go 标准 + TinyGo) |
| WASI 目标 | 2 (Go wasip1 + TinyGo) |
| LLVM 后端 | ✅ Milestone 1（标量/控制流/函数/类）|
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

## 🚀 v3.1.0 — Spring Boot 式框架 + LLVM Milestone 2

> 预计: 2026-Q3 | 工作量: 2 个月

### 任务 1: KylixBoot 框架 — Spring Boot 式 Web 框架 ⭐⭐⭐

Kylix 版的 Spring Boot：通过**注解式声明**（Attribute）自动装配路由、依赖注入、ORM、配置。

**设计目标：**
```pascal
// 目标语法 — 声明式 Web 控制器
program UserService;
uses kylix.boot;

[Controller('/api/users')]
type
  TUserController = class
    [Inject]
    UserRepo: TUserRepository;

    [Get('/')]
    function ListUsers(req: TRequest): TResponse;
    begin
      result := req.json(UserRepo.FindAll());
    end;

    [Post('/')]
    function CreateUser(req: TRequest): TResponse;
    begin
      var user := req.body<TUser>();
      UserRepo.Save(user);
      result := req.created(user);
    end;
  end;

[Repository]
type
  TUserRepository = class
    [Query('SELECT * FROM users')]
    function FindAll(): array of TUser;

    [Query('SELECT * FROM users WHERE id = ?')]
    function FindById(id: Integer): TUser;

    procedure Save(user: TUser);
  end;

[Configuration]
type
  TAppConfig = class
    [Value('server.port', 8080)]
    Port: Integer;

    [Value('db.url')]
    DatabaseURL: String;
  end;

begin
  KylixBoot.Run(TUserController, TUserRepository, TAppConfig);
end.
```

**核心模块：**

| 模块 | 说明 |
|------|------|
| `kylix.boot/router` | 注解驱动的路由注册（[Get]/[Post]/[Put]/[Delete]）|
| `kylix.boot/di` | 依赖注入容器（[Inject]/[Component]/[Service]）|
| `kylix.boot/orm` | 声明式 ORM（[Entity]/[Repository]/[Query]）|
| `kylix.boot/config` | 配置绑定（[Configuration]/[Value]）|
| `kylix.boot/middleware` | 中间件链（[Middleware]/[Filter]）|
| `kylix.boot/validation` | 参数校验（[Required]/[Min]/[Max]/[Email]）|
| `kylix.boot/security` | 认证授权（[Authenticated]/[Role]）|
| `kylix.boot/cache` | 缓存注解（[Cacheable]/[CacheEvict]）|

**实现路径：**
1. Attribute 语法解析（parser 扩展）
2. 编译时注解处理器（annotation processor）
3. 代码生成：路由表 + DI 容器 + ORM 映射
4. 运行时：Go net/http 包装

### 任务 2: LLVM 后端 Milestone 2

- [ ] 数组 codegen（`alloca [N x T]` + GEP 索引）
- [ ] 动态数组（`array of T` → slice 结构体）
- [ ] 接口 codegen（vtable fat pointer `{ ptr vtable, ptr data }`）
- [ ] 泛型单态化（模板展开，每个实例生成独立函数）
- [ ] 异常处理（`invoke` + `landingpad`）
- [ ] LLVM 优化 Pass（`-O2` 通过 llc 参数）
- [ ] 交叉编译目标（`--backend=llvm --target=linux/amd64`）

### 任务 3: 编译器修复（v3.1 重点）

以下 bug 影响示例完整性，需修复：

| Bug | 表现 | 影响 |
|-----|------|------|
| 字符串插值 | `${var}` 不展开 | 中等 |
| 匿名函数返回类型 | `function(x):T` 生成无返回值签名 | 中等 |
| `var p: TClass` 字段访问 | 生成 `interface{}`，字段不可见 | 高 |
| `match` 代码生成 | 生成 Go 语法错误 | 高 |
| `uses` 在 `program` 中 | stdlib 函数不可直接调用 | 高 |

### 任务 4: 完整教程 v2（覆盖所有 74 特性）

- [ ] OOP 接口示例（等 bug 修复后）
- [ ] 字符串插值示例（等 bug 修复后）
- [ ] Match 模式匹配示例（等 bug 修复后）
- [ ] Lambda 返回值示例（等 bug 修复后）
- [ ] stdlib 示例（等 `uses` 在 program 中修复后）
- [ ] Web server 完整示例
- [ ] kylix test 工作流示例
- [ ] WASI 完整示例
- [ ] LLVM 后端示例

---

## 📋 v3.2.0 — 编译器健壮性 + stdlib 全量可用

> 预计: 2026-Q4 | 工作量: 6 周

### 任务 1: 编译器 Bug 全量修复

- [ ] `var p: TClass` → 生成 `*TClass` 而非 `interface{}`
- [ ] 字符串插值展开（Lexer 层处理 `${expr}`)
- [ ] 匿名函数返回类型保留
- [ ] `match` 语句完整代码生成
- [ ] `uses` 在 `program` 中正确注入符号表

### 任务 2: stdlib 全量可用（program 中直接调用）

- [ ] strutil 函数可在 program 中用（Reverse/ToUpper/StartsWith...）
- [ ] mathutil 函数可在 program 中用（Abs/Max/Min/Pow/IsPrime...）
- [ ] arrayutil 函数可在 program 中用
- [ ] sysutil 函数可在 program 中用（ReadFile/WriteFile/FileExists...）
- [ ] jsonutil 函数可在 program 中用（JsonDecodeMap/JsonGetString...）
- [ ] datetime 函数可在 program 中用（Now/MakeDate/FormatPattern...）
- [ ] regex 函数可在 program 中用（IsEmail/IsURL/IsNumeric...）

### 任务 3: stdlib Phase 6 — 网络与加密

- [ ] `net` — TCP/UDP 客户端，DNS 查询
- [ ] `crypto` — SHA256/MD5/HMAC/AES 加密
- [ ] `encoding` — Base64/Hex/CSV/URL 编解码
- [ ] `os` — 进程管理、信号、环境变量

### 任务 4: 包注册中心正式上线

- [ ] 部署到 kylix.top/packages（PostgreSQL + TLS）
- [ ] 搜索索引与全文检索
- [ ] 包统计仪表板（下载量/依赖图）
- [ ] GitHub Actions 自动发布集成

---

## 📋 v4.0.0 — 完全独立运行时（脱离 Go）

> 预计: 2027 | 工作量: 6+ 月

### 目标

Kylix 彻底脱离 Go 工具链，成为完全自主的编译型语言：

```
Kylix 源码 (.klx)
    ↓  kylix compile
LLVM IR (.ll)
    ↓  llc / lld
原生二进制
```

### 任务 1: 自研运行时 (KylixRT)

- [ ] 垃圾回收器（引用计数 + 标记清除）
- [ ] 字符串运行时（引用计数字符串，零拷贝切片）
- [ ] 动态数组运行时（增长策略）
- [ ] 接口运行时（fat pointer + 类型擦除）
- [ ] 异常运行时（stack unwinding via LLVM EH）
- [ ] 并发原语（goroutine 等价 → 协程）

### 任务 2: LLVM 后端 Milestone 3 — 完整 Kylix 语言

- [ ] 泛型完整实现（约束 + 单态化 + 特化）
- [ ] 字符串插值 codegen（`${expr}` → LLVM IR）
- [ ] 闭包 codegen（捕获变量的内存布局）
- [ ] async/await codegen（stackful 协程）
- [ ] 完整 Pascal 运行时（Set 类型、string 类型、Real）

### 任务 3: 自举编译器 v2.0

- [ ] 用 Kylix 重写 LLVM 后端（`generator_llvm.klx`）
- [ ] 编译器能编译自己（完整自举）
- [ ] 性能对比：Kylix v4 vs Go 参考编译器

### 任务 4: 完整 IDE 支持

- [ ] DAP 调试适配器（KylixRT 原生调试）
- [ ] VS Code 扩展 v2（语义高亮 + 类型推导显示）
- [ ] IntelliJ 插件
- [ ] Language Server 重写（更快的增量分析）

---

## 🐛 已知问题跟踪

### 编译器层

| ID | 问题 | 严重度 | 目标版本 |
|----|------|--------|---------|
| KLX-C01 | `var p: TClass` 生成 `interface{}` 导致字段不可访问 | 高 | v3.2 |
| KLX-C02 | 字符串插值 `${var}` 不展开 | 中 | v3.2 |
| KLX-C03 | 匿名函数 `function(x):T` 返回类型丢失 | 中 | v3.2 |
| KLX-C04 | `match` 语句生成无效 Go 代码 | 高 | v3.2 |
| KLX-C05 | `uses` 模块在 `program` 中符号不可见 | 高 | v3.2 |

### LLVM 后端层（Milestone 1 已知限制）

| ID | 问题 | 目标版本 |
|----|------|---------|
| KLX-L01 | 不支持数组（array of T / array[1..N]）| v3.1 |
| KLX-L02 | 不支持接口（interface）| v3.1 |
| KLX-L03 | 不支持泛型单态化 | v3.1 |
| KLX-L04 | 不支持异常处理（try/except）| v3.1 |
| KLX-L05 | 无优化 Pass（-O0 等效）| v3.1 |

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
