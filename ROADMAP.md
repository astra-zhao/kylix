# Kylix Development Roadmap

> 最后更新: 2026-06-20  
> 当前版本: v2.4.0 🎉  
> 官网: [kylix.top](https://kylix.top)  
> 目标: Kylix 语言自举（用 Kylix 写 Kylix 编译器）

**🎉 v2.4.0 已发布！** 完善与生态 — i18n 全面接入、REPL `:type` 真正推导、SetLength 修复、包管理器嵌套依赖 + lockfile、stdlib Phase 3。  
**📍 下一步：** v2.5.0 — 工具链深化（LSP 重构动作、iter 模块、方法定义修复）。

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
| **v2.5.0** | **工具链深化** | 📋 计划中 | 2026-07 |
| **v2.6.0** | **性能与优化** | 📋 计划中 | 2026-08 |
| **v3.0.0** | **LLVM 后端 + 包注册中心** | 📋 长期 | 2026-Q4 |

---

## 📊 累计统计 (v2.4.0)

| 指标 | 数量 |
|------|------|
| Go 测试包 | 13 个（全部通过）|
| Go 级测试 | ~200+ |
| Kylix 级 stdlib 测试 | 39 个（6 模块）|
| 纯 Kylix stdlib 函数 | 45 个 |
| CLI 命令 | 17 个 |
| 错误代码 (i18n) | 21 个（中英双语）|
| 原生构建目标 | 5 (linux/darwin/windows × amd64/arm64) |
| WASM 目标 | 2 (Go 标准 + TinyGo) |

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

## 🎯 v2.5.0 — 工具链深化 (2026-07)

**主题**：完成 v2.3-v2.4 "基础设施已就绪但未全面接入"的收尾。

### 1. LSP 重构动作（1 周）
- `textDocument/rename` — 跨文件重命名符号
- `textDocument/codeAction` — 提取函数、内联变量
- 复用已有 `ReferenceWalker` 做跨文件作用域

### 2. `kylix doc` 代码示例提取（3 天）
- 从 `//` 注释中提取 ` ```pascal ... ``` ` 代码块
- 在生成的 Markdown 中作为可运行示例
- 可选：自动测试提取的示例

### 3. `kylix bench` 内存分配报告（3 天）
- 新增 `--mem` 标志，报告 B/op + allocs/op
- 使用 Go `runtime.ReadMemStats`
- 输出: `BenchmarkFib  1000000  1234 ns/op  240 B/op  3 allocs/op`

### 4. `iter` 迭代器模块（5 天）
```pascal
uses iter;
var nums := [1, 2, 3, 4, 5];
var doubled := iter.Map(nums, function(x: Integer): Integer begin result := x * 2; end);
var evens := iter.Filter(doubled, ...);
var sum := iter.Reduce(evens, 0, ...);
```

### 5. 类方法外部定义修复（3 天）
- 当前：类体内声明 + 类外定义 → Go 重复方法
- 目标：支持 Pascal 风格 `function TClass.Method()` 在类体外定义

---

## 🚀 v2.6.0 — 性能与优化 (2026-08)

### 1. 并行编译（1 周）
- 独立 unit 并行 parse + generate (goroutine pool)
- `sync.WaitGroup` + worker pool
- race detector 测试
- 目标：10 文件项目 > 30% 加速

### 2. 常量传播 + 死代码消除（5 天）
- 常量折叠: `const MAX = 5; array[0..MAX-1]` → `array[0..4]`
- 移除 `return`/`raise` 后的不可达代码
- 跳过未使用的局部变量

### 3. LSP 大文件性能基准（2 天）
- 合成 10K 行 .klx 文件
- 测量 didChange → diagnostics 延迟
- 加入 CI 作为性能回归守护
- 目标：单次增量编辑 < 50ms

---

## 🔮 v3.0.0 — 架构突破 (2026-Q4)

### 1. LLVM 原生后端（2-3 月）
- Kylix → LLVM IR → 原生二进制，脱离 Go
- 更小二进制（无 Go runtime）+ 更好优化
- 保留 Go 后端作为 fallback (`--backend=go`)

### 2. 包注册中心服务端（1 月）
- `kylix.top/packages` — 可浏览的包索引
- `kylix publish` — 上传包到注册中心
- 语义版本 + 依赖解析
- GitHub 包镜像缓存

### 3. stdlib 完全 Kylix 化 Phase 4+（2 周）
- 用 Kylix 重写 `jsonutil`、`regex`、`datetime`
- 性能关键部分通过 `external` 声明保留 Go 实现

### 4. WASI 支持（2 周）
- WASM 目标的 WASI 系统调用层
- 文件 I/O、环境变量、时钟
- 支持 Cloudflare Workers / Fastly Compute 等服务端 WASM

---

## 🐛 剩余已知问题

### 类型系统
- [ ] 无泛型变体检查（协变/逆变）— v2.7+
- [ ] 多返回值未完全集成到类型推导 — v2.5

### 编译器
- [ ] 错误恢复有时在无效 AST 状态继续 — v2.5
- [ ] 无常量传播或死代码消除 — v2.6
- [ ] 类方法外部定义生成重复 Go 方法 — v2.5

### 标准库
- [ ] `TDateTime` 运算符 (+, -) 未实现 — v2.5
- [ ] `TRegex` 不支持命名捕获组 — v2.6
- [ ] `jsonutil` 仅支持扁平 JSON — v3.0

### 工具链
- [ ] `kylix doc` 不提取代码示例 — v2.5
- [ ] `kylix bench` 不报告内存分配 — v2.5
- [ ] LSP rename 重构未实现 — v2.5
- [ ] LSP code actions (extract/inline) 未实现 — v2.5

### 基础设施
- [x] ~~无 CI/CD~~ → ✅ v2.2
- [ ] 无跨平台自动化回归测试 — v2.5
- [ ] 网站 (kylix.top) 需更新 v2.x 示例 — v2.5

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
- [ ] 发布 v2.4.0 公告
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
