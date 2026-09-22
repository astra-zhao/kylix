# Kylix 编译器项目总结

[![English](https://img.shields.io/badge/lang-English-blue.svg)](README.md)
[![Official Site](https://img.shields.io/badge/official-kylix.top-4f6ef7.svg)](https://kylix.top)
[![版本](https://img.shields.io/badge/version-0.12.0-blue.svg)](CHANGELOG.md)
[![许可证](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![自举](https://img.shields.io/badge/self--hosting-IR%20%E4%B8%8D%E5%8A%A8%E7%82%B9-brightgreen.svg)](docs/SELFHOSTING_DEV_GUIDE.md)

## 项目概述

Kylix 是一个现代化的 Pascal 编译器：默认把 Kylix 源码转译为可读的 Go 代码（`go build` 编译运行）；也可以通过 **LLVM 原生后端**（`--backend=llvm`）直接产出 LLVM IR 并链接为原生二进制——**运行时完全不依赖 Go 工具链**。它结合了 Pascal 的清晰性和简洁性，同时添加了现代语言特性，并配备完整的 IDE 工具链、编辑器集成与无 Go 自举闭环。

**当前版本**：v0.12.0（KylixAdmin P5）

**项目地址**：https://github.com/astra-zhao/kylix

**官方网站**：https://kylix.top

> 🔥 **重大里程碑 (v0.12.0)**：**KylixAdmin 走向可移植与单文件部署**——**纯 Kylix SQL 方言层**（`apps/admin/lib/dialect.klx`）+ **LLVM 端 libpq 后端**（`pkg/llvmgen/stdlib_db_pg.go`，`PQexecParams` + `PQftype` OID 分派；`DbOpenPg` 独立入口点，只有用到 pg 的程序才链 `-lpq`）；**四形态 E2E 逐字一致**（23 场景 × sqlite/postgres × Go/LLVM）；**`[Entity]` 驱动建表与增量迁移**（`lib/migrate.klx` + `schema_migrations` + 内省 ADD COLUMN）；**`[Embed('views','static')]` 编译器语言特性**（程序头属性 → 两端烘焙 → `ReadFile`/`BootStatic` 内嵌优先）实现**单二进制自包含**；部署文档 [docs/ADMIN_DEPLOY.md](docs/ADMIN_DEPLOY.md) + 自包含 CI 门 + release 资产。PoC 前置验证：C collation 与 ILIKE 是强制项、连接泄漏第 101 次撞上限。IR 不动点保持。详见 [CHANGELOG.md](CHANGELOG.md)。
>
> 🔥 **重大里程碑 (v0.11.0)**：**KylixAdmin 进入生产力阶段**——**通用 CRUD 引擎**（`[Entity]` 元数据由编译器**两端**发射，新 stdlib 模块 `entitymeta`；LLVM 后端此前对 ORM/校验注解零支持）：写一个带注解的类即得到列表（搜索/排序/分页）、表单（含校验）、删除、审计、菜单项与权限点；全部业务表迁入引擎（`main.klx` 775 → 103 行）；**仪表盘**（统计卡 + 零依赖整数几何 SVG 图）+ **个人中心**（改密/头像）+ **UI 设计系统**（设计令牌/组件/服务端渲染三态主题/响应式）；顺带修复 LLVM cookie 解析器不跳 `;` 后空格（第二个及之后的 cookie 读不到）与 `DbQueryScalar` 遇 NULL 列崩溃；**双端 E2E 22 场景**逐字 diff。IR 不动点保持。详见 [CHANGELOG.md](CHANGELOG.md) 与 [docs/ADMIN_CRUD_GUIDE.md](docs/ADMIN_CRUD_GUIDE.md)。
>
> 🔥 **重大里程碑 (v0.10.0)**：**KylixAdmin 启航**——**LLVM 后端 Boehm GC**（`--gc=boehm` opt-in：~110 处用户数据分配点路由 `GC_malloc`，默认 malloc 模式 IR 逐字节不变，GC 教程 maxRSS 62MB→22MB，issue #1 闭环）；**KylixAdmin P2 认证与 RBAC**（纯 Kylix `apps/admin/`：PBKDF2-HMAC-SHA256 口令哈希、SessionRegenerate 防固定、session-first `[Authenticated]`、`[Role]` 守卫真体化、失败锁定、Remember-me，4 控制器 15 路由）；**双端 E2E**（`apps/admin/e2e.sh`，12 curl 场景 Go/LLVM 归一化逐字 diff，已入 CI）；htab GC 混搭分配器 bug 修复。**CI 11 job 全绿**。v0.6.9 的 IR 不动点保持（gen1 ≡ gen2，26.7 万行逐字节）。详见 [CHANGELOG.md](CHANGELOG.md)。

---

## 快速开始

### 安装

```bash
# 方式一：下载预编译二进制（GitHub Release，5 平台 + bootstrap tarball）
gh release download v0.12.0

# 方式二：源码构建
git clone https://github.com/astra-zhao/kylix.git
cd kylix

# 构建编译器
go build -o kylix ./cmd/kylix/

# 添加到 PATH（可选）
export PATH=$PATH:$(pwd)

# LLVM 原生后端预检（可选，需 llc/clang）
./kylix doctor
```

### 第一个程序

```bash
cat > hello.klx << 'EOF'
program Hello;
begin
  WriteLn('Hello, Kylix!');
end.
EOF

./kylix run hello.klx            # 自动探测：有 Go 走 Go，无 Go 回退 LLVM
./kylix build --backend=llvm hello.klx   # 强制原生二进制（无 Go 依赖）
```

### 创建完整项目

```bash
./kylix new myapp
cd myapp
../kylix run      # 编译并运行 main.klx
```

---

## CLI 命令参考

```bash
kylix new <name>        # 创建新项目
kylix build [file]      # 编译项目或文件（--backend=llvm / --llvm-opt=N / --target / -g / --emit-llvm）
kylix run [file]        # 编译并运行（--backend=auto 自动探测 Go/LLVM）
kylix check [file]      # 语法检查（不生成代码）
kylix fmt [file]        # 格式化源代码
kylix test              # 运行测试（--backend=llvm 无 Go 可用）
kylix bench             # 运行基准（--backend=llvm 无 Go 可用）
kylix doc --openapi     # 从注解生成 OpenAPI 3.1 YAML
kylix debug             # 调试辅助
kylix add/install/remove # 包管理器
kylix doctor            # 环境预检（go/llc/clang/opt/sqlite3/curl/openssl）
kylix repl              # 交互式 REPL
kylix lsp               # 启动 LSP 服务器（用于编辑器）
kylix version           # 显示版本信息
```

---

## 架构与双后端

```
                 .klx 源码
                    │
          ┌─────────┴─────────┐
          ▼                   ▼
   Go 后端（默认）        LLVM 后端（--backend=llvm）
   Kylix → Go 代码        Kylix → LLVM IR → llc → .o
   → go build             → clang 链接原生二进制
          │                   │
          ▼                   ▼
      可执行文件           原生二进制（运行时无 Go）
```

- **Go 后端**：转译为可读 Go 代码，29ms 冷编译，调试信息/工具链生态直接复用
- **LLVM 后端**：原生二进制 + DWARF 逐行调试（`-g`，LLDB 单步/检视）+ `-O2` 优化（575ms）+ 跨平台 target triple（Linux/macOS/Windows amd64/arm64）
- **自举编译器**：编译器本身用 Kylix 编写（`src/*.klx`，9 文件），经 LLVM 后端编译为原生二进制后可再编译任意 Kylix 程序——**全程无 Go**

---

## 测试状态（v0.12.0）

| 项目 | 结果 |
|------|------|
| Go 单元测试 | ✅ 17 包全绿 |
| 教程 sweep（Go 后端） | ✅ 58/58（57 示例，多文件模块记 2 项） |
| 教程 sweep（LLVM 后端） | ✅ 58/58（含 example60 server E2E + example64 GC parity） |
| 自举 sweep（无 Go） | ✅ 57 PASS + 1 SKIP（example60 E2E） |
| 自举 IR 不动点 | ✅ gen1 ≡ gen2（26.7 万行逐字节） |
| KylixAdmin E2E | ✅ **四形态** 23 场景逐字一致（sqlite×{Go,LLVM} ≡ postgres×{Go,LLVM}，`apps/admin/e2e.sh`，两个 CI job） |
| KylixAdmin 迁移/部署 | ✅ 增量迁移（3 列→10 列，sqlite/pg）· 空目录自包含冒烟（`migrate_check.sh`/`deploy_check.sh`） |
| CI | ✅ 12 job（三平台 + arm64 + selfrepro + perf-gate + admin-e2e + admin-e2e-pg） |

---

## 已完成的工作

### 编译器核心 ✅
- 词法/语法（Pratt）/AST/代码生成完整管线
- 传统 Pascal 特性：强类型、函数/过程、控制结构、record/enum、异常处理（try/except/finally）
- 现代特性：类型推断、lambda/闭包、多返回值元组解构、map[K]V、动态数组、Variant 动态类型（boxed {tag, payload} 运行时）、泛型类单态化、字符串插值、match、properties、**error 类型**（v0.7.0：`(T, error)` 多返回 + `error('msg')` 构造 + ErrorStr）
- 面向对象：类/继承/vtable 虚方法/inherited、接口胖指针、多态基类
- 模块系统：`unit`/`uses`、多文件编译、包管理器

### 双后端 ✅
- **Go 后端**：可读代码生成 + 智能导入 + 增量编译缓存（55× 加速）+ 纯 Kylix 模板引擎（v0.7.0）
- **LLVM 后端**：完整 stdlib IR 实现（crypto AES/SHA/PBKDF2、httpclient libcurl、websocket RFC 6455、sqlite3 数据库、JWT HS256、boot HTTP server、Variant 运行时、DWARF 调试）+ DCE 优化 + 跨平台（含 Windows Winsock 真实现 + llvm-mingw 交叉链接，v0.7.1）+ **Boehm GC**（v0.10.0，`--gc=boehm` opt-in）
- **纯 Kylix stdlib unit（三端同源）**：regex 回溯引擎（v0.7.1）、stringutil 20 函数（v0.8.0）、template_engine 含 layout/partials（v0.9.0）

### KylixBoot 框架 ✅
- `[Controller]`/`[Get]`/`[Post]` 路由自动装配、`[Service]`/`[Inject]` DI、`[Required]`/`[Email]` 等字段校验、`[Authenticated]`/`[Role]` 安全守卫、`[Entity]`/`[Repository]`/`[Query]` ORM 注解、`[Body(TEntity)]` 请求体绑定、JWT 一键接入、OpenAPI 3.1 自动生成
- **v0.9.0 框架补齐**：服务端 Session + CSRF、multipart 文件上传、`TResponse.Download/FileBytes/CSV`、分页 BootPagerHTML、模板 layout/partials
- **v0.11.0 CRUD 元数据**：`[Entity]` 注解 → 编译器两端发射 `entitymeta` 注册序列（列表/表单/校验/权限全自动），新增注解 `[Label]`/`[Searchable]`/`[Hidden]`/`[Nullable]`/`[Default]`/`[ReadOnly]`（见 [docs/ADMIN_CRUD_GUIDE.md](docs/ADMIN_CRUD_GUIDE.md)）

### 自举编译器 ✅（v0.5.2 → v0.9.0 bootstrap boot server）
- `src/*.klx`（token/error/ast/lexer/parser/generator/llvmgen + stdlib IR 烘焙）9 文件
- v0.5.2 构建打通 → v0.5.3 round-trip + 自繁殖 → v0.5.6 LLVM self-host 51/51 → v0.5.7 self-reproduction 不动点 → **v0.6.9 无 Go 闭环**（stdlib IR 烘焙 + emitter 补缺 20+ 项 + gen2 诞生 + IR 不动点）→ v0.7.0 模板引擎三端同源 + Release 工作流（预编译二进制 + bootstrap tarball）→ **v0.9.0 bootstrap boot server 落地**（example60 E2E 解除 SKIP）+ stdlib IR 重烘链路闭环（139→179 签名）

### IDE 工具链 ✅
- CLI 完整命令集（new/build/run/check/fmt/test/bench/doc/debug/add/install/remove/doctor/repl/lsp）
- **VS Code 扩展**（语法高亮、LSP 集成、代码片段、实时诊断）
- **JetBrains 插件**（v0.6.7：TextMate 高亮 + LSP4IJ 桥接 + 25 Live Templates + Run 配置）
- LSP 服务器（补全/悬停/跳转/重命名/格式化，任何 LSP 编辑器可接入）

---

## 项目结构

```
kylix/
├── cmd/kylix/              # CLI 入口
├── token/ lexer/ ast/ parser/  # 前端（token → lexer → Pratt parser → AST）
├── generator/              # Go 后端代码生成器（+ KylixBoot 注解扫描/装配）
├── pkg/
│   ├── compiler/           # 编译 API + 增量缓存 + 注解诊断 + 类型检查
│   ├── llvmgen/            # LLVM 后端（IR 生成 + stdlib IR 实现 + Variant + DWARF + 优化 pass）
│   ├── lsp/ repl/ pkgmgr/ openapi/  # LSP / REPL / 包管理 / OpenAPI
├── stdlib/                 # Go 标准库封装（web, orm, db, cache, jwt, crypto, ...）
│   └── klx/                # LSP 补全用 Kylix 声明文件
├── src/                    # 自举编译器源码（.klx，9 文件 + stdlib_ir.klx 烘焙数据）
├── scripts/                # 测试 sweep / stdlib IR 提取 / LLVM 捆绑脚本
├── examples/               # 教程（complete-tutorial 26 章节 57 编号示例）+ 主题示例
├── vscode-ext/             # VS Code 扩展
├── jetbrains-plugin/       # JetBrains 插件（Gradle Kotlin）
├── html/                   # 官网页面
├── docs/                   # 文档（入门/开发/自举/LLVM/性能/Web 框架/小白教程）
└── benchmarks/             # 编译性能基准
```

---

## 编辑器集成

### VS Code

```bash
cd vscode-ext && npm install
# 在 VS Code 中按 F5 启动扩展开发主机，或打包安装
```

功能：语法高亮、实时错误检查、代码补全、悬停提示、LSP 客户端集成。

### JetBrains（IntelliJ IDEA / GoLand）

```bash
cd jetbrains-plugin && ./gradlew buildPlugin
# 安装 build/distributions/Kylix-*.zip（Settings → Plugins → Install from Disk）
```

功能：TextMate 语法高亮、LSP4IJ 桥接（补全/跳转/重命名/格式化/错误高亮）、25 个 Live Templates、Kylix Run 配置。详见 `jetbrains-plugin/README.md`。

### 其他编辑器（Neovim / Emacs / Sublime）

```json
{ "command": ["kylix", "lsp"], "filetypes": ["kylix"] }
```

---

## 文档资源

| 文档 | 内容 |
|------|------|
| [README.md](README.md) / [README_CN.md](README_CN.md) | 项目主文档（英文/中文） |
| [docs/GETTING_STARTED_CN.md](docs/GETTING_STARTED_CN.md) | 快速入门（中文） |
| [docs/TUTORIAL_FOR_BEGINNERS_CN.md](docs/TUTORIAL_FOR_BEGINNERS_CN.md) | 小白友好入门教程（由浅入深 + ASCII 图解） |
| [examples/complete-tutorial/](examples/complete-tutorial/) | 完整教程（25 章节 56 编号示例，README_CN.md 中文） |
| [docs/KYLIX_IDE_USER_MANUAL.md](docs/KYLIX_IDE_USER_MANUAL.md) | IDE 工具使用手册 |
| [docs/KYLIX_DEV_GUIDE.md](docs/KYLIX_DEV_GUIDE.md) | 开发指南（架构与贡献） |
| [docs/SELFHOSTING_DEV_GUIDE.md](docs/SELFHOSTING_DEV_GUIDE.md) | 自举开发指南 |
| [docs/llvm-backend.md](docs/llvm-backend.md) / [docs/llvm-performance.md](docs/llvm-performance.md) | LLVM 后端与性能 |
| [docs/WEB_FRAMEWORK.md](docs/WEB_FRAMEWORK.md) | Web 框架指南 |
| [docs/API_STABILITY.md](docs/API_STABILITY.md) | 1.0.0 API 稳定性冻结承诺 |
| [CHANGELOG.md](CHANGELOG.md) | 完整版本历史 |

---

## 版本里程碑

| 版本 | 日期 | 核心内容 |
|------|------|----------|
| v0.1.x | 2026-06 | 编译器核心 + CLI 工具链 + Web 框架雏形 |
| v0.3.x | 2026-06 | LLVM M1/M2 + KylixBoot 注解栈 + stdlib Phase 6 |
| v0.4.x | 2026-07 | LLVM stdlib 全模块（jsonutil/crypto/httpclient/db/websocket）+ DWARF 逐行调试 + 泛型/lambda/闭包 |
| v0.5.x | 2026-07/08 | **自举全线打通**：构建 → round-trip → LLVM self-host 51/51 → self-reproduction 不动点；Variant 运行时；多态 gate；KylixBoot 注解移植 |
| v0.6.0-0.6.2 | 2026-08 | 性能基准（LLVM -O2 575ms）+ KylixRT（`kylix run` 无 Go 单二进制）+ 跨平台（Linux 51/51）+ var 参数 |
| v0.6.3-0.6.5 | 2026-08 | jwt 双端真实现 + 分发 B（捆绑 LLVM）+ DbQueryRows + websocket + 手写 SHA-1 + 30× 性能优化 |
| v0.6.6-0.6.8 | 2026-08 | boot HTTP server（无 Go 真可用）+ stdlib 补全 + **JetBrains 插件（ROADMAP #9 ✅）** |
| **v0.6.9** | **2026-09-04** | **bootstrap 无 Go 闭环：stdlib IR 烘焙 + gen2 诞生 + IR 不动点（gen1 ≡ gen2 ≡ gen3）+ 教程三 sweep 全绿** |
| v0.7.0 | 2026-09-06 | web 页面开发 + web 框架：error 类型（三端）、纯 Kylix 模板引擎、页面渲染 API、example60 server E2E |
| v0.7.1-0.7.2 | 2026-09-10 | Windows 一等公民（Winsock/regex 引擎/交叉链接）+ CI 全绿 + selfrepro 改走 --emit-llvm IR 链 |
| v0.8.0 | 2026-09-11 | 自举 stdlib：纯 Kylix stringutil 三端同源 + per-request arena + htab magic 校验加固 |
| **v0.9.0** | **2026-09-13** | **1.0.0-rc 打磨：KylixBoot 补齐（Session/CSRF/上传/Download/分页/layout）+ bootstrap boot server（example60 E2E）+ 重烘闭环 + CI 三平台全绿 + 性能门禁 + API 稳定性冻结** |
| **v0.10.0** | **2026-09-19** | **KylixAdmin P0+P2：LLVM 后端 Boehm GC（`--gc=boehm`，issue #1）+ 后台认证与 RBAC（PBKDF2/Session/锁定/RBAC/审计，15 路由，双端 E2E 入 CI）+ htab GC bug 修复** |
| **v0.11.0** | **2026-09-21** | **KylixAdmin P3+P4：`[Entity]` 元数据驱动的通用 CRUD 引擎（编译器两端发射 `entitymeta`）+ 仪表盘（零依赖 SVG）+ 个人中心 + UI 设计系统（三态主题/响应式）；LLVM cookie 解析器与 `DbQueryScalar` NULL 修复；双端 E2E 22 场景** |
| **v0.12.0** | **2026-09-22** | **KylixAdmin P5：纯 Kylix 方言层 + LLVM 端 libpq 后端（sqlite/postgres 同一份源码，四形态 E2E 逐字一致）+ `[Entity]` 驱动建表与增量迁移 + `[Embed]` 单二进制自包含 + 部署文档/资产** |

详细更新日志见 [CHANGELOG.md](CHANGELOG.md)。

---

## 后续规划

- **v0.13.0 — H5 多端路线 A**（见 [docs/MULTIPLATFORM.md](docs/MULTIPLATFORM.md)）：响应式 PWA 移动页面组（复用 P4 设计系统）+ manifest/service worker + JWT refresh token + 登录限流 + `docs/H5_GUIDE.md`。KylixAdmin P0–P5 已全部完成（v0.10.0/v0.11.0/v0.12.0）
- **v0.13.0–v0.15.0 — 多端平台**（见 [docs/MULTIPLATFORM.md](docs/MULTIPLATFORM.md)）：H5 PWA → 编译器多端能力（C ABI export + android/ios triple）→ 示例应用（Kotlin+JNI / SwiftUI 壳）+ wasm32
- **1.0.0**：v0.7.1–v0.15.0 gate 全过后发布正式版（KylixAdmin + 多端为旗舰 showcase）

完整路线图见 [ROADMAP.md](ROADMAP.md)。

---

## 贡献

欢迎贡献代码！请随时提交 Issue 和 Pull Request。

**提交代码流程**：
1. Fork 仓库
2. 创建功能分支：`git checkout -b feature/my-feature`
3. 编写代码和测试
4. 运行测试：`go test ./...` + `bash examples/complete-tutorial/test_all.sh`
5. 提交：`git commit -am "Add feature: ..."`
6. 推送并创建 PR

**代码规范**：
- 使用 `gofmt` 格式化代码；每个源文件 ≤ 1000 行
- 添加注释说明复杂逻辑
- 添加单元测试

---

## 许可证

MIT License
