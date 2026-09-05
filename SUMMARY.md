# Kylix 编译器项目总结

[![English](https://img.shields.io/badge/lang-English-blue.svg)](README.md)
[![Official Site](https://img.shields.io/badge/official-kylix.top-4f6ef7.svg)](https://kylix.top)
[![版本](https://img.shields.io/badge/version-0.6.9-blue.svg)](CHANGELOG.md)
[![许可证](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![自举](https://img.shields.io/badge/self--hosting-IR%20%E4%B8%8D%E5%8A%A8%E7%82%B9-brightgreen.svg)](docs/SELFHOSTING_DEV_GUIDE.md)

## 项目概述

Kylix 是一个现代化的 Pascal 编译器：默认把 Kylix 源码转译为可读的 Go 代码（`go build` 编译运行）；也可以通过 **LLVM 原生后端**（`--backend=llvm`）直接产出 LLVM IR 并链接为原生二进制——**运行时完全不依赖 Go 工具链**。它结合了 Pascal 的清晰性和简洁性，同时添加了现代语言特性，并配备完整的 IDE 工具链、编辑器集成与无 Go 自举闭环。

**当前版本**：v0.6.9（2026-09-04 发布）

**项目地址**：https://github.com/astra-zhao/kylix

**官方网站**：https://kylix.top

> 🔥 **重大里程碑 (v0.6.9)**：**bootstrap 无 Go 闭环达成**——自举编译器（`src/*.klx`，9 文件）经 LLVM 后端编译为原生二进制 gen2（纯原生、无 Go），gen2 输出与 gen1 **逐字节一致（~220k 行 IR 不动点）**，gen3 ≡ gen2；stdlib IR 烘焙免手写 15.5k 行 Go 移植；51 教程 Go/LLVM 双后端全绿，自举 sweep 50 PASS + 1 SKIP。详见 [CHANGELOG.md](CHANGELOG.md)。

---

## 快速开始

### 安装

```bash
# 克隆仓库
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

## 测试状态（v0.6.9）

| 项目 | 结果 |
|------|------|
| Go 单元测试 | ✅ 16 包全绿 |
| 教程 sweep（Go 后端） | ✅ 51/51 |
| 教程 sweep（LLVM 后端） | ✅ 51/51（原生二进制） |
| 自举 sweep（无 Go） | ✅ 50 PASS + 1 SKIP |
| 自举 IR 不动点 | ✅ gen1 ≡ gen2 ≡ gen3（~220k 行逐字节） |

---

## 已完成的工作

### 编译器核心 ✅
- 词法/语法（Pratt）/AST/代码生成完整管线
- 传统 Pascal 特性：强类型、函数/过程、控制结构、record/enum、异常处理（try/except/finally）
- 现代特性：类型推断、lambda/闭包、多返回值元组解构、map[K]V、动态数组、Variant 动态类型（boxed {tag, payload} 运行时）、泛型类单态化、字符串插值、match、properties
- 面向对象：类/继承/vtable 虚方法/inherited、接口胖指针、多态基类
- 模块系统：`unit`/`uses`、多文件编译、包管理器

### 双后端 ✅
- **Go 后端**：可读代码生成 + 智能导入 + 增量编译缓存（55× 加速）
- **LLVM 后端**：完整 stdlib IR 实现（crypto AES/SHA、httpclient libcurl、websocket RFC 6455、sqlite3 数据库、JWT HS256、boot HTTP server、Variant 运行时、DWARF 调试）+ DCE 优化 + 跨平台

### KylixBoot 框架 ✅
- `[Controller]`/`[Get]`/`[Post]` 路由自动装配、`[Service]`/`[Inject]` DI、`[Required]`/`[Email]` 等字段校验、`[Authenticated]`/`[Role]` 安全守卫、`[Entity]`/`[Repository]`/`[Query]` ORM 注解、`[Body(TEntity)]` 请求体绑定、JWT 一键接入、OpenAPI 3.1 自动生成

### 自举编译器 ✅（v0.5.2 → v0.6.9）
- `src/*.klx`（token/error/ast/lexer/parser/generator/llvmgen + stdlib IR 烘焙）9 文件
- v0.5.2 构建打通 → v0.5.3 round-trip + 自繁殖 → v0.5.6 LLVM self-host 51/51 → v0.5.7 self-reproduction 不动点 → **v0.6.9 无 Go 闭环**（stdlib IR 烘焙 + emitter 补缺 20+ 项 + gen2 诞生 + IR 不动点）

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
├── examples/               # 教程（complete-tutorial 20 章节 50 示例）+ 主题示例
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
| [examples/complete-tutorial/](examples/complete-tutorial/) | 完整教程（20 章节 50 示例，README_CN.md 中文） |
| [docs/KYLIX_IDE_USER_MANUAL.md](docs/KYLIX_IDE_USER_MANUAL.md) | IDE 工具使用手册 |
| [docs/KYLIX_DEV_GUIDE.md](docs/KYLIX_DEV_GUIDE.md) | 开发指南（架构与贡献） |
| [docs/SELFHOSTING_DEV_GUIDE.md](docs/SELFHOSTING_DEV_GUIDE.md) | 自举开发指南 |
| [docs/llvm-backend.md](docs/llvm-backend.md) / [docs/llvm-performance.md](docs/llvm-performance.md) | LLVM 后端与性能 |
| [docs/WEB_FRAMEWORK.md](docs/WEB_FRAMEWORK.md) | Web 框架指南 |
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

详细更新日志见 [CHANGELOG.md](CHANGELOG.md)。

---

## 后续规划

- **v0.7.0 — Web 页面开发 + web 框架**：模板引擎 + 页面渲染 + 静态资源 + 表单/Cookie；net Winsock / regex pcre2 真实现（Windows 真机验证）
- **1.0.0**：v0.6.9 + v0.7.0 完成后发布正式版
- **后续**：跨平台 CI 稳定（Windows 真机）、自举 stdlib（.klx 编写 → bootstrap 编译 → 自包含）

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
