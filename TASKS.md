# Kylix 开发任务清单

> 最后更新: 2026-06-21
> 官网: [kylix.top](https://kylix.top)
> 关联文档: [ROADMAP.md](ROADMAP.md), [CHANGELOG.md](CHANGELOG.md)
> 当前版本: v3.0.0-alpha

---

## v3.0.0-alpha ✅ 已完成 (2026-06-21)

### 任务 1: stdlib Phase 4 — 纯 Kylix 化 ✅

- [x] `stdlib/src/jsonutil.klx` — 完整 JSON 解析器（TJsonLexer + TJsonParser），支持嵌套，29 测试
- [x] `stdlib/src/regex.klx` — 纯 Kylix 字符级验证函数（IsEmail/IsURL/IsIPv4/IsPhone/IsDate），19 测试
- [x] `stdlib/src/datetime.klx` — FormatPattern/DateAdd/DateSub/IsLeapYear/DaysInMonth，21 测试
- [x] 更新 `stdlib/klx/regex.klx` 和 `stdlib/klx/datetime.klx` 声明文件

### 任务 2: 编译器 Bug 修复 — `external` 函数声明解析 ✅

- [x] `ast/ast.go`: `FunctionDecl.IsExternal` 新字段
- [x] `parser/parser_decl.go`: `EXTERNAL` 修饰词识别（行 253）
- [x] `generator/generator_types.go`: IsExternal=true 时跳过函数体
- [x] 8 个新测试（3 parser + 5 generator）

### 任务 3: 包注册中心服务端 ✅

- [x] `registry/` — 独立 Go module
- [x] SQLite 数据库层（Store 接口，可切换 PostgreSQL）
- [x] REST API（GET/POST /api/v1/packages，版本列表，下载）
- [x] Bearer token 认证 + middleware
- [x] Web 前端：htmx + Tailwind CSS（首页 + 包详情页）
- [x] `kylix publish` CLI 命令（tarball 打包 + 上传，支持 KYLIX_TOKEN）
- [x] `pkg/pkgmgr/http.go` HTTP 辅助函数
- [x] 7 个集成测试

### 任务 4: WASI 支持 ✅

- [x] `kylix build --wasi`（Go 1.21+ GOOS=wasip1 GOARCH=wasm）
- [x] `kylix build --wasi --tinygo`（TinyGo -target=wasi）
- [x] `pkg/wasi/`: wasi.go + wasi_stub.go + wasi_wasip1.go
- [x] `stdlib/src/wasi.klx` + `stdlib/klx/wasi.klx`
- [x] `examples/wasi-hello/` + `examples/cloudflare-worker/`
- [x] 8 个单元测试（pkg/wasi）

### 任务 5: LLVM 原生后端 Milestone 1 ✅

- [x] `pkg/llvmgen/codegen.go` — 生成器核心（module/function/block/SSA）
- [x] `pkg/llvmgen/expr.go` — 表达式 codegen（i64/i1/double/ptr，算术/比较/逻辑，WriteLn/IntToStr/Length）
- [x] `pkg/llvmgen/stmt.go` — 语句 codegen（if/while/for/repeat，变量 alloca，函数定义）
- [x] `pkg/llvmgen/class.go` — 类 codegen（%ClassName struct，vtable，GEP 字段访问，虚函数分发，malloc 构造）
- [x] `pkg/llvmgen/compile.go` — 完整管道（FindLLVM + CompileToNative）
- [x] `kylix build --backend=llvm` CLI 集成
- [x] 端到端验证：Hello World + 算术 + while 循环 → 原生二进制
- [x] 24 个单元测试（含 6 个类 codegen 测试）

### 任务 6: stdlib HTTP 客户端 ✅

- [x] `stdlib/http_client.go` — THttpClient（Get/Post/StatusCode/SetHeader），一键函数
- [x] `stdlib/klx/httpclient.klx` — LSP 声明文件
- [x] `stdlib/src/httpclient.klx` — 纯 Kylix 包装

---

## v3.1 — 下一步任务

### LLVM 后端 Milestone 2
- [ ] 接口 codegen（vtable fat pointer）
- [ ] 泛型单态化
- [ ] LLVM 优化 Pass (-O2 / LTO)
- [ ] 交叉编译支持 (linux/windows/arm64)

### 包注册中心部署
- [ ] 部署到 kylix.top/packages（PostgreSQL + TLS）
- [ ] 搜索索引优化

### stdlib Phase 6
- [ ] `net` 模块（TCP/UDP）
- [ ] `crypto` 模块（hash/HMAC）
- [ ] `encoding` 模块（base64/hex/CSV）

---

## 关键设计决策记录

1. **LLVM 后端策略**: Milestone 1 优先最小可用集（标量/控制流/函数/类），Milestone 2 补全接口/泛型/异常。Go 后端仍为默认，LLVM 为可选。

2. **stdlib 纯 Kylix 化策略**: 性能关键路径（JsonEncode/JsonEncodePretty、regexp）保留 `external` Go 实现；逻辑层（解析、验证、格式化）全部用纯 Kylix 实现。

3. **包注册中心独立模块**: `registry/` 是独立 Go module，避免污染主编译器依赖树。SQLite 用于开发/单机部署，Store 接口支持切换 PostgreSQL。

4. **WASI build-tag 分离**: `wasi_stub.go`（`//go:build !wasip1`）允许在原生系统运行单元测试，无需 WASI 运行时。

---
