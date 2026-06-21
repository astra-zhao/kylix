# Kylix Development Roadmap

> 最后更新: 2026-06-21  
> 当前版本: v3.0.0-alpha 🚀  
> 官网: [kylix.top](https://kylix.top)  
> 目标: Kylix 语言自举（用 Kylix 写 Kylix 编译器）

**🚀 v3.0.0-alpha 发布！** 架构突破 — LLVM 原生后端（最小可用子集）、包注册中心、WASI 支持、stdlib Phase 4。  
**📍 下一步：** LLVM 后端继续完善（类与接口 vtable、泛型单态化、-O2）。

---

## 总览

| 阶段 | 内容 | 状态 | 版本 |
|------|------|------|------|
| Phase 6 | 修复关键 Bug | ✅ 完成 | v1.0.2 |
| Phase 7 | 补齐语言能力 | ✅ 完成 | v1.0.3 |
| Phase 8 | 编写 compiler.klx | ✅ 完成 | v1.1.2 |
| Phase 9 | 自举验证 | ✅ 完成 | v1.2.0 |
| Phase 10 | v2.0 准备：核心特性 | ✅ 完成 | v1.3.0–v1.5.0 |
| Phase 11 | v2.0 准备：工程质量 | ✅ 完成 | v1.6.0–v2.0.0 |
| Phase 12 | v2.0.0 发布：生产级编译器 | ✅ 完成 | v2.0.0 |
| **v2.1.0** | **增强类型系统 + stdlib Phase 1** | **✅ 完成** | **v2.1.0** |
| **v2.2.0** | **工程质量 + stdlib Phase 2** | **✅ 完成** | **v2.2.0** |
| **v2.3.0** | **开发者体验 (LSP/REPL/Debug/WASM)** | **✅ 完成** | **v2.3.0** |
| **v2.4.0** | **完善与生态 (i18n/推导/SetLength/包管理)** | **✅ 完成** | **v2.4.0** |
| **v2.5.0** | **工具链深化 (LSP/doc/bench/iter/方法修复)** | **✅ 完成** | **v2.5.0** |
| **v2.6.0** | **性能与优化 (并行编译/DCE/LSP基准)** | **✅ 完成** | **v2.6.0** |
| **v3.0.0** | **LLVM 后端 + 包注册中心** | 🚀 alpha | 2026-06-21 |

---

## 📊 累计统计 (v3.0.0-alpha)

| 指标 | 数量 |
|------|------|
| Go 测试包 | 15 个（全部通过）|
| Go 级测试 | ~310+ |
| Kylix 级 stdlib 测试 | 117 个（10 模块）|
| 纯 Kylix stdlib 函数 | 90+ |
| CLI 命令 | 19 个 |
| 错误代码 (i18n) | 21 个（中英双语）|
| 原生构建目标 | 5 (linux/darwin/windows × amd64/arm64) |
| WASM 目标 | 2 (Go 标准 + TinyGo) |
| WASI 目标 | 2 (Go wasip1 + TinyGo) |
| LLVM 后端 | ✅ Milestone 1（标量/控制流/函数/类）|
| 包注册中心 | ✅ REST API + Web 前端 |
| 并行编译 | ✅ (goroutine pool, race detector) |
| 死代码消除 | ✅ (递归 DCE) |

---

## v2.1.0 ✅ (2026-06-19) — 增强类型系统 + stdlib Phase 1

| # | 任务 | 状态 |
|---|------|------|
| 1 | 多参数泛型约束 (`TMap<K: IComparable, V: IHashable>`) | ✅ |
| 2 | 类→接口实现映射验证（方法名 + 签名） | ✅ |
| 3 | 增强类型推导（Boolean、array of T、nil、not） | ✅ |
| 4 | stdlib Phase 1: `strutil` (8 fn) + `mathutil` (12 fn) | ✅ |

---

## v2.2.0 ✅ (2026-06-19) — 工程质量 + stdlib Phase 2

| # | 任务 | 状态 |
|---|------|------|
| 1 | GitHub Actions CI/CD (ci.yml + release.yml) | ✅ |
| 2 | 泛型约束方法签名验证（参数类型 + 返回类型） | ✅ |
| 3 | 包级类型检查 (`CheckProject`，跨文件 uses 解析) | ✅ |
| 4 | 增量编译启用 (`BuildCache` 接入 cmdBuild) | ✅ |
| 5 | stdlib Phase 2: `arrayutil` (8 fn) + `collections` (6 fn) | ✅ |

---

## v2.3.0 ✅ (2026-06-19) — 开发者体验

| # | 任务 | 状态 |
|---|------|------|
| 1 | LSP 增量同步 (textDocumentSync 2，版本检查，range 编辑) | ✅ |
| 2 | REPL Tab 补全 + `:load` + `:type` | ✅ |
| 3 | kylix test: Setup/Teardown + `--filter` | ✅ |
| 4 | i18n 框架 (21 codes × 2 languages) | ✅ |
| 5 | Delve 调试器集成 (`kylix debug`) | ✅ |
| 6 | WebAssembly 后端 (`--wasm` + `--tinygo`) | ✅ |

---

## v2.4.0 ✅ (2026-06-20) — 完善与生态

| # | 任务 | 状态 |
|---|------|------|
| 1 | i18n 全面接入 typecheck（消息 + hint 中文化） | ✅ |
| 2 | REPL `:type` 真正类型推导 (`compiler.InferType`) | ✅ |
| 3 | SetLength 修复 (Go 泛型 `__kylixSetLength[T any]`) | ✅ |
| 5 | 包管理器: 嵌套依赖解析 + `kylix.lock` | ✅ |
| 6 | stdlib Phase 3: `stringbuilder` (5 fn) + `resulttype` (6 fn) | ✅ |

---

## v2.5.0 ✅ (2026-06-20) — 工具链深化

| # | 任务 | 状态 |
|---|------|------|
| 1 | LSP 跨文件 rename + 上下文 codeAction | ✅ |
| 2 | `kylix doc` 代码示例提取 (fenced code blocks) | ✅ |
| 3 | `kylix bench --mem` 内存分配报告 (B/op + allocs/op) | ✅ |
| 4 | `iter` 迭代器模块 (9 函数: Contains/Count/Unique/Reverse/Concat/Slice/Sum/Min/Max) | ✅ |
| 5 | 类方法外部定义修复 (forward declaration 不再重复生成) | ✅ |

**注**: Map/Filter/Reduce 延后 — Kylix 当前不支持函数类型参数。改为提供不依赖回调的 9 个实用函数。

---

## v2.6.0 ✅ (2026-06-20) — 性能与优化

| # | 任务 | 状态 |
|---|------|------|
| 1 | 并行编译 (goroutine pool parse, race detector 通过, CPU 113%) | ✅ |
| 2 | 死代码消除 (递归 DCE, return/raise/Exit/break/continue 后截断) | ✅ |
| 3 | LSP 大文件性能基准 (500 函数 1.2ms parse, 1.0ms 增量编辑) | ✅ |

---

## 🔮 v3.0.0 — 架构突破 (2026-Q4)

### 1. LLVM 原生后端（2-3 月）
- [x] `pkg/llvmgen/codegen.go` — 生成器核心架构（module/function/block/SSA）
- [x] `pkg/llvmgen/expr.go` — 基础类型表达式 (i64/i1/double/ptr)，算术/比较/逻辑运算，WriteLn/IntToStr/Length
- [x] `pkg/llvmgen/stmt.go` — 控制流 (if/while/for/repeat)，函数定义/调用，变量 alloca
- [x] `pkg/llvmgen/compile.go` — 完整管道：AST → .ll → .o → native binary（llc + clang）
- [x] `kylix build --backend=llvm` — CLI 集成，bypassGo codegen 路径
- [x] 端到端验证：Hello World + 整数算术 + while 循环，原生二进制运行正确
- [x] 18 个单元测试（IR 生成正确性）
- [ ] 类与接口 codegen（vtable / fat pointer）— 下一步
- [ ] 泛型单态化 — 下一步
- [ ] LLVM 优化 Pass (-O2 / LTO) — 下一步
- [ ] 交叉编译 (linux/windows/arm64) — 下一步

### 2. 包注册中心服务端（1 月）
- [x] `registry/` 子目录，独立 Go module
- [x] SQLite 数据库层（Store 接口，可切换 PostgreSQL）
- [x] REST API：GET /packages, POST /packages, GET /packages/:name/versions, GET /dl
- [x] API token 认证（Bearer token middleware）
- [x] Web 前端：htmx + Tailwind CSS，首页 + 包详情页
- [x] `kylix publish` CLI 命令（tarball 打包 + 上传）
- [x] 7 个集成测试（端到端：publish → list → search → download）

### 3. stdlib 完全 Kylix 化 Phase 4+（2 周）
- [x] 任务 3.1: `jsonutil` — 纯 Kylix 解码器，支持嵌套 JSON，29 测试
- [x] 任务 3.2: `regex` — 纯 Kylix 验证函数（IsEmail/IsURL/IsIPv4/IsPhone/IsDate），19 测试
- [x] 任务 3.3: `datetime` — FormatPattern/DateAdd/DateSub/IsLeapYear/DaysInMonth，21 测试
- 性能关键部分通过 `external` 声明保留 Go 实现

### 4. WASI 支持（2 周）
- [x] `--wasi` 编译选项 (`GOOS=wasip1 GOARCH=wasm` via Go 1.21+)
- [x] TinyGo `-target=wasi` 支持（更小二进制）
- [x] `pkg/wasi/` — 系统调用层（Stdout/Stdin/Getenv/Args/Clock/File I/O）
- [x] Build-tag 分离：wasip1 原生实现 + 非 WASI stub（可本地测试）
- [x] `stdlib/src/wasi.klx` — 纯 Kylix 高层包装（WriteLn/ReadLine/HasArg/ElapsedMs）
- [x] `stdlib/klx/wasi.klx` — LSP 声明文件
- [x] `examples/wasi-hello/` — Wasmtime/Node.js 运行示例
- [x] `examples/cloudflare-worker/` — Cloudflare Workers HTTP handler 示例
- [x] 8 个单元测试（pkg/wasi）

---

## 🐛 剩余已知问题

### 类型系统
- [ ] 无泛型变体检查（协变/逆变）— v2.7+
- [ ] 多返回值未完全集成到类型推导 — v2.5

### 编译器
- [x] ~~错误恢复有时在无效 AST 状态继续~~ → ✅ v2.2 (错误恢复)
- [x] ~~无常量传播或死代码消除~~ → ✅ v2.6 (DCE)
- [x] ~~类方法外部定义生成重复 Go 方法~~ → ✅ v2.5
- [x] ~~`external` 函数声明解析失败~~ → ✅ v3.0.0-alpha

### 标准库
- [x] ~~`TDateTime` 运算符 (+, -) 未实现~~ → ✅ v3.0.0-alpha (DateAdd/DateSub)
- [x] ~~`jsonutil` 仅支持扁平 JSON~~ → ✅ v3.0.0-alpha (嵌套解析)
- [ ] `TRegex` 不支持命名捕获组 — v3.1

### 工具链
- [x] ~~`kylix doc` 不提取代码示例~~ → ✅ v2.5
- [x] ~~`kylix bench` 不报告内存分配~~ → ✅ v2.5
- [x] ~~LSP rename 重构未实现~~ → ✅ v2.5 (跨文件)
- [x] ~~LSP code actions 未实现~~ → ✅ v2.5 (rename + extract)

### LLVM 后端（Milestone 1 已知限制）
- [ ] 不支持接口（vtable fat pointer）— Milestone 2
- [ ] 不支持泛型单态化 — Milestone 2
- [ ] 无优化 Pass（-O0）— Milestone 2
- [ ] 不支持异常（try/catch）— Milestone 3
- [ ] 不支持数组、record — Milestone 2

### 基础设施
- [x] ~~无 CI/CD~~ → ✅ v2.2
- [ ] 无跨平台自动化回归测试 — v2.5
- [ ] 网站 (kylix.top) 需更新 v3.x 示例 — v3.1

---

## 🔮 v3.1 — 下一步里程碑

### LLVM 后端 Milestone 2
- [ ] 接口 codegen（vtable / fat pointer）
- [ ] 泛型单态化
- [ ] LLVM 优化 Pass (-O2 / LTO)
- [ ] 交叉编译支持 (linux/windows/arm64)

### 包注册中心部署
- [ ] 部署到 kylix.top/packages
- [ ] 域名 + TLS + PostgreSQL 生产配置
- [ ] 搜索索引优化

### stdlib Phase 6
- [ ] `net` 模块（TCP/UDP/DNS）
- [ ] `crypto` 模块（hash/HMAC/AES）
- [ ] `encoding` 模块（base64/hex/CSV）

---

## 📝 文档缺口

- [ ] 新手入门教程 — v2.5
- [ ] 泛型约束使用指南 — v2.5
- [ ] 测试最佳实践指南 — v2.5
- [ ] 性能优化指南 — v2.6
- [ ] Delphi/FreePascal 迁移指南 — v2.6
- [ ] LSP 配置 (VS Code / Neovim) — v2.5

---

## 🎓 社区与生态

### 短期 (v2.5)
- [ ] 发布 v2.6.0 公告
- [ ] 创建 Discord/Slack 社区
- [ ] 设置 GitHub Discussions

### 中期 (v3.0)
- [ ] 包注册中心 (kylix.top/packages)
- [ ] 示例项目库
- [ ] 视频教程 / 截屏

### 长期 (v3.0 后)
- [ ] 会议演讲 / 工作坊
- [ ] 企业赞助 / 基金会

---

## 历史阶段存档

### Phase 6 → v1.0.2 ✅
字符串插值、异常类型、多返回值、Properties、数组范围、内存泄漏修复。

### Phase 7 → v1.0.3 ✅
Map 类型、变体类型、动态数组。

### Phase 8 → v1.1.2 ✅
枚举类型、Slice 表达式、Unit 文件系统、多文件联编。7 个 Kylix 源文件编写。

### Phase 9 → v1.2.0 ✅
自举验证通过 — Go 参考编译器与 Kylix 自举编译器语义等价，15/15 示例通过。

### Phase 10 → v1.3.0–v1.5.0 ✅
接口验证、Kylix 层错误报告 (`//line`)、真实泛型、多返回值覆盖、LSP 诊断、增量编译 (55×)、stdlib `.klx` 声明 + 包管理器。

### Phase 11 → v2.0.0 ✅
技术债务清理：CompileFile 缓存、topoSort 路径对齐、GenerateBody exception、类型检查 MVP、包管理器集成、38 个新测试。

### Phase 12 → v2.0.0 ✅
错误代码体系 (KLX001–499)、错误恢复、拼写建议、类型推导、泛型约束、`kylix test/doc/bench`、26 个测试、1360 LOC。

---

**License**: MIT
