# API 稳定性承诺（v0.12.0 → 1.0.0）

> v0.9.0 起进入 1.0.0-rc 打磨阶段。本文档划定 **1.0.0 将冻结的公开 API 面**，
> 以及冻结后的破坏性变更 / 弃用流程约定。1.0.0 发布时本文档即生效。

## 承诺分级

| 级别 | 范围 | 冻结后的变更规则 |
|------|------|------------------|
| **冻结（Stable）** | 下文三张清单内的全部面 | 破坏性变更禁止；只能**向后兼容地新增**。确需破坏时走下文弃用流程，且至少保留一个大版本（v1.x） |
| **实验（Experimental）** | 标注 experimental 的功能（如 `--backend=llvm` 帮助文本自述） | 可在小版本中破坏，但必须在 CHANGELOG 显著标注 |
| **内部（Internal）** | 下文"不冻结范围"内的所有面 | 随时变更，无通知义务 |

## 冻结范围

### 1. CLI 面（`kylix` 可执行文件）

子命令（`cmd/kylix/main.go`）：

```
build   run   check   test   doc   bench   doctor   debug
new     fmt   repl    lsp    add   install remove(rm)  publish
version(-v/--version)   help(-h/--help)
```

flag 面（全部经 Go `flag` 包定义，`-flag` 与 `--flag` 双形式等价；新增一律
向后兼容，已有语义不改）：

| 子命令 | flags |
|--------|-------|
| `build` | `-o` `-v` `-g` `--time` `--backend go\|llvm` `--llvm-opt 0..3` `--target os/arch` `--gc boehm` `--wasm` `--wasi` `--tinygo` |
| `run` | `-keep` `-v` `-g` `--backend auto\|go\|llvm` `--llvm-opt 0..3` `--gc boehm` |
| `check` | `--syntax` |
| `test` | `-v` `-tap` `-dir` `-filter` `-backend auto\|go\|llvm` |
| `bench` | `-count N` `-mem` |
| `doc` | `-out DIR` `-stdout` `-openapi` `-title` `-api-version` |
| `debug` | `-headless` `-port N` `-keep` |
| `publish` | `-version` `-registry URL` `-token T`（或 `KYLIX_TOKEN` 环境变量） |

约定细节同样冻结：`--backend auto` 的探测顺序（有 Go 走 Go，否则 LLVM 回退）、
`kylix run` 无 Go 环境产出原生二进制的行为、`kylix doctor` 的退出码语义
（缺必需工具非零，advisory 项不失败）。

### 2. stdlib 声明面（Kylix 源码 import 的单元）

**内建/绑定模块**（`stdlib/klx/*.klx`，24 个 LSP 声明文件描述的行为面）：
`boot` `cache` `config` `container` `crypto` `datetime` `db` `encoding`
`entitymeta` `exceptions` `httpclient` `jsonutil` `jwt` `middleware` `net`
`orm` `regex` `sysutil` `template` `validation` `wasi` `web` `websocket`
`autoconfig`

`db` 模块（v0.12.0 增补）：`DbOpenPg(dsn)`（postgres 独立入口点）、`DbLastError(db)`、
`DbSetMaxOpenConns/DbSetMaxIdleConns/DbSetConnMaxLifetime` 一并冻结。**`?` 占位符的语义**
（sqlite 原样、postgres 由 db 层改写为 `$n`）与**参数求值次数**（每个实参恰好求值一次）也是契约的一部分。

`entitymeta`（v0.11.0）是 `[Entity]` 注解元数据的运行期读取面，冻结的是
**编码格式与访问器语义**（`EntityMetaOf` → `"pk|label|flags"`、
`EntityFieldAt` → `"column|kind|label|flags"`、越界/未知返回 `""`），
注册函数由编译器发射、非用户调用面。

**纯 Kylix 多文件单元**（三端同源，`stdlib/*.klx`）：
`stringutil.klx` `regex_engine.klx` `template_engine.klx`

冻结的是**函数/类型签名与文档化行为**（字节语义、错误返回约定等，见各教程
与 `docs/*.md`）。三端（Go 后端 / LLVM 后端 / bootstrap）对同一签名的输出
以教程 parity 为准——端间差异按 [TECHNICAL_DEBT.md](../TECHNICAL_DEBT.md)
记录的限制处理，不算契约变更。

### 3. 语言语法面

README「语言特性」清单所列全部构造：类型系统（强类型 + 推断 / record /
map / Variant / 动态数组 / 泛型 / is / as）、控制流（if/while/for/repeat/
case/forEach/match）、OOP（class/interface/继承/virtual/property）、
多返回 + 解构、lambda、error 类型、try/except、字符串插值与切片、
注解（KylixBoot 全家）、unit/uses、`*_test.klx` / `*_bench.klx` 约定。

**注解面**（v0.11.0 起一并冻结）：`[Entity]` `[Column]` `[PrimaryKey]`
`[Repository]` `[Query]` `[Controller]` `[Service]` `[Component]` `[Inject]`
`[Get]/[Post]/[Put]/[Delete]` `[Authenticated]` `[Role]` `[Body]`
`[Required]` `[Email]` `[Min]` `[Max]` `[MinLen]` `[MaxLen]`
`[Label]` `[Searchable]` `[Hidden]` `[Nullable]` `[Default]` `[ReadOnly]` `[Unique]`
（后七项为 v0.11.0/v0.12.0 新增的 CRUD 注解，语义见
[docs/ADMIN_CRUD_GUIDE.md](ADMIN_CRUD_GUIDE.md)）。

**文件级属性**（v0.12.0）：`[Embed('dir', …)]` 写在 `program`/`unit` 头之前，把目录烘焙进二进制；
`ReadFile` 与静态资源服务的**查找顺序**（内嵌优先、磁盘回落）随之冻结。

**TRequest/TResponse 方法面**（v0.11.0 增补）：`req.Path()` 与既有
`Param/Query/Header/Body/Form/Cookie/Session*/File/SaveFile/PageNum` 同级冻结；
`req.Form` 的**回退链差异**（Go 含 URL query 回退、LLVM 只查 body）按
TECHNICAL_DEBT 记录的限制处理，不构成契约——应用代码应 GET 用 `Query`、
POST 用 `Form`。

关键字集合（`token/token.go`）与内置类型名（`Integer Real Boolean String
Char Variant`）冻结——新增关键字属于破坏性变更（可能撞用户标识符），须走
弃用流程。

## 不冻结范围（Internal）

- **Go 内部 API**：`lexer/ parser/ ast/ generator/ pkg/llvmgen/ pkg/compiler/
  pkg/lsp/ pkg/repl/` 等包的导出符号——它们是编译器实现细节，不是用户契约。
- **自举源码** `src/*.klx` 及 `src/stdlib_ir.klx`（AUTO-GENERATED，勿手改）
  ——bootstrap 编译器的私有实现。
- **生成的 Go 代码形态**：Go 后端发射的 `.go` 文件结构无兼容承诺（用户契约
  是 Kylix 源码语义，不是中间产物）。
- **`.kylix-cache/` 缓存格式**、**包管理器 registry 响应格式**（registry
  本身 1.0.0 前仍可能调整）。
- **未实现/stub 行为**：TECHNICAL_DEBT 记录的限制（Windows target 下
  websocket typed stub、UDP/DnsLookup stub、bootstrap 端 boot server 部分
  stub 等）——补全实现不算破坏（目标是行为收敛到 Go 端语义）。

## 弃用流程（冻结后）

1. **标记**：CHANGELOG「Deprecated」段 + 编译器 diagnostic（KLX 前缀警告）
   + 文档标注。
2. **保留**：至少整个下一个大版本（v1.x 系列内不删除）。
3. **移除**：仅在 major 版本（v2.0.0）。

新增 API 不需要流程；修改冻结 API 的**行为**（同签名不同语义）视同破坏，
按上述流程处理。

## 审查机制

- 每个 release 前对照本文档过一遍 diff：CLI flag / stdlib 签名 / 关键字
  三处是否有未记录的变更（人工 checklist，暂不自动化）。
- `go test` 16 包 + 三 sweep（Go/LLVM/bootstrap 教程 parity）是行为契约的
  回归网；教程输出逐字一致即契约未被破坏。
