# Kylix Development Roadmap

> 最后更新: 2026-06-14
> 当前版本: v1.5.0
> 官网: [kylix.top](https://kylix.top)
> 目标: Kylix 语言自举（用 Kylix 写 Kylix 编译器）

---

## 总览

本项目目标分九个阶段：

| 阶段 | 内容 | 状态 | 版本 |
|------|------|------|------|
| Phase 6 | 修复关键 Bug | ✅ 完成 | v1.0.2 |
| Phase 7 | 补齐语言能力 | ✅ 完成 | v1.0.3 |
| Phase 8 | 编写 compiler.klx | ✅ 完成 | v1.1.2 |
| Phase 9 | 自举验证 | ✅ 完成 | v1.2.0 |
| **Phase 10 (P0-P2)** | **v2.0 准备：核心特性** | **✅ 完成** | **v1.3.0–v1.5.0** |
| Phase 11 | v2.0 准备：工程质量 | 🚧 进行中 | v1.6.0 |
| Phase 12 | v2.0.0 发布：生产级编译器 | 📋 计划中 | v2.0.0 |

**当前进度：Phase 10 完成！v1.5.0 已具备增量编译、LSP 实时诊断、stdlib 声明文件、包管理器。**

**Phase 11 目标：修复已知正确性缺陷，补齐测试覆盖，为 v2.0.0 生产级发布铺平道路。**

---

## Phase 10: v2.0 准备 — 核心特性 ✅ 完成

### P0: 接口验证 + Kylix 层错误报告 + 真实泛型 → v1.3.0 ✅

| 特性 | 描述 |
|------|------|
| **接口实现验证** | `class Foo implements IBar` 编译时检查，缺失方法报 Kylix 错误 |
| **Kylix 层错误映射** | `//line` directives 让 Go 编译器错误指向 `.klx` 文件 |
| **真实泛型代码生成** | `TBox<T>` → Go 1.18+ `type TBox[T any] struct` |

### P1: 多返回值全场景覆盖 → v1.3.1 ✅

| 语法 | 状态 |
|------|------|
| `return (b, a)` | ✅ 已修复 |
| `x, y := Swap(3, 7)` | ✅ 已修复 |
| `var a, b := Pair()` | ✅ 已修复 |

### P2: LSP 实时诊断 → v1.3.2 ✅

LSP 在 `didOpen`/`didChange` 时推送 Kylix 层诊断（parse 错误 + 接口验证）。

### P2: 增量编译 → v1.4.0 ✅

| 场景 | 时间 | 加速比 |
|------|------|--------|
| 冷构建（无缓存） | 444 ms | — |
| 全缓存 | 8 ms | **55×** |
| 单文件变更 | 6 ms | **74×** |

**实现：**
- `.kylix-cache/` — SHA-256 keyed JSON 缓存
- `CompileProject` 跳过未变更文件的 parse+generate
- `generator.GenerateBody` / `BuildOutput` 支持独立生成

### P3: stdlib Kylix 化 + 包管理器 → v1.5.0 ✅

**stdlib `.klx` 声明文件：**
- `stdlib/klx/sysutil.klx` — 文件 I/O、路径操作、环境变量
- `stdlib/klx/datetime.klx` — TDateTime 30+ 方法
- `stdlib/klx/regex.klx` — TRegex + IsEmail/IsURL 验证
- `stdlib/klx/jsonutil.klx` — JSON 编解码、文件 I/O

**包管理器：**
```bash
kylix add utils github.com/alice/utils@v1.0.0
kylix install
kylix remove utils
```

---

## Phase 11: v2.0 准备 — 工程质量 🚧 进行中

### 11.1 修复正确性缺陷（P1）

详见 [TECHNICAL_DEBT.md](TECHNICAL_DEBT.md) 优先级 1。

| 缺陷 | 影响 | 工作量 |
|------|------|--------|
| `CompileFile` 未接入缓存 | 项目模式增量编译无效 | 2h |
| `topoSortWithFiles` 路径对齐错误 | `//line` 指向错误文件 | 1h |
| `GenerateBody` exception 输出不稳定 | 多文件编译异常类型重复 | 30min |

### 11.2 补齐功能缺口（P2）

| 功能 | 状态 | 工作量 |
|------|------|--------|
| 类型检查层（最小 MVP） | 📋 未开始 | 2 周 |
| 包管理器集成到编译器 | 📋 未开始 | 3h |
| `kylix add` git 逻辑修复 | 📋 未开始 | 5min |

### 11.3 测试覆盖（P3）

| 模块 | 当前测试数 | 目标 |
|------|-----------|------|
| `pkg/pkgmgr` | 0 | 5+ |
| `pkg/compiler/cache.go` | 0 | 3+ |
| `pkg/lsp` stdlib 加载 | 0 | 2+ |
| `stdlib/klx/*.klx` 可解析性 | 0 | 1 |
| parser 泛型/多返回值 | 0 | 3+ |

**目标：** 关键路径测试覆盖率达到 60%+

---

## Phase 12: v2.0.0 发布 — 生产级编译器 📋 计划中

### 12.1 类型系统完善

- [ ] 类型推导：`var x := 42` 推导出 `Integer`
- [ ] 泛型约束：`<T: IComparable>` 约束类型参数
- [ ] 类型别名：`type UserId = Integer;` 生成真实类型别名

### 12.2 错误体验

- [ ] Kylix 错误代码（`E001: 未声明变量`）
- [ ] 错误恢复：parse 错误不阻塞后续 semantic 检查
- [ ] 建议修复：`did you mean 'userName'?`

### 12.3 工具链

- [ ] `kylix fmt --check` CI 模式
- [ ] `kylix test` 测试运行器
- [ ] `kylix doc` 文档生成器

### 12.4 stdlib 完全 Kylix 化

- [ ] 用 `.klx` 重写 `web.go`、`orm.go`（保留 Go cgo 桥接层）
- [ ] stdlib 文档（LSP hover 显示）

---

## Phase 6: 修复关键 Bug → v1.0.2 ✅ 已完成

| Bug | 严重性 | 描述 |
|-----|--------|------|
| **字符串插值** | 🔴 P1 | Lexer→Parser→Generator 全链路修复 |
| **异常类型** | 🔴 P1 | `raise Exception.Create('msg')` → `panic(&Exception{Message: "msg"})` |
| **多返回值** | 🟠 P2 | `function Divide(...): (Real, Boolean)` + `result := (0, false)` → `return 0, false` |
| **Properties** | 🟠 P2 | `property Name: String read Field;` → getter/setter 方法 |
| **web_demo 匿名过程** | 🟠 P2 | 嵌套 record 类型的 `end` 深度追踪修复 |
| **数组范围大小** | 🟠 P2 | `array[0..2]` → 正确计算 `((2-0)+1) = 3` |
| **内存泄漏** | 🔴 P0 | `parseGroupedExpression` 死循环修复 |

---

## Phase 7: 补齐语言能力 → v1.0.3 ✅ 已完成

| 特性 | 严重性 | 描述 |
|------|--------|------|
| **Map 类型** | 🔴 P0 | `map[K]V` 语法，自动初始化，索引读写 |
| **变体类型** | 🔴 P0 | `variant Case: Type; end` → Go interface + struct 模式 |
| **动态数组** | 🔴 P0 | `append(arr, elem)` / `SetLength(arr, n)` 内置函数 |

---

## Phase 8: 编写 compiler.klx → v1.1.2 ✅ 已完成

### ✅ 已完成 — Go 编译器后端升级

| 特性 | 描述 |
|------|------|
| **枚举类型** | `(val1, val2, ...)` → Go `const` + `iota` |
| **Slice 表达式** | `s[a:b]` → Go slice 语法 |
| **Unit 文件系统** | `unit X;` 模块声明 + `uses` 跨文件引用 |
| **多文件联编** | `kylix build a.klx b.klx` + 拓扑排序 |
| **类代码生成** | struct-only 方案，基类用 `interface{}` 实现多态，具体类用 `*T` 指针 |
| **软关键字扩展** | 25+ 关键字可作为标识符（Default, DownTo, When, Dynamic 等） |
| **局部 var/const** | 函数体内局部声明生成，含 `_ = name` 占位避免 Go 未使用变量报错 |
| **Exit 语句** | `exit` → `return result`（有返回值）或 `return`（过程） |
| **Bare method calls** | `self.Method` 作为 statement → `self.Method()` |
| **Map 作为表达式** | `map[K]V` 注册为 prefix parse fn，生成 `map[K]V{}` |
| **空数组** | `[]` → `nil`，兼容任意 Go slice 类型 |
| **字符串转义** | `\`, `"`, `\n` 正确转义 |
| **for 循环** | `for i = 0`（不用 `:=`）避免 int/int64 类型冲突 |
| **构造函数** | `T.Create` → `&T{}`，`T.Create(args)` → `&T{Field: arg}` 使用类字段名映射 |
| **新内置函数** | Ord, Length, IntToStr, StrToInt64, StrToFloat, ReadFile |
| **is/as 类型分派** | `is` → type assertion check，`as` → type assertion，基类 `interface{}` 多态下工作正常 |

### ✅ 已完成 — Kylix 源文件

```
src/
├── token.klx           # Token 类型枚举、关键字映射 (209 行)
├── ast.klx             # AST 节点类层次，54 个类 (374 行)
├── lexer.klx           # 词法分析器，字符→Token 流 (366 行)
├── parser.klx          # Pratt 解析器，Token→AST (2338 行)
├── error.klx           # 编译器错误/诊断类型 (91 行)
├── generator.klx       # Go 代码生成器，完整实现 ~1350 行
└── main.klx            # 入口，串联 lexer→parser→generator (56 行)
```

总计 ~4800 行 Kylix 代码。7 文件联合编译通过（Kylix → Go 转换 + Go 编译零错误）。

### ✅ v1.1.5 完成内容

| 任务 | 优先级 | 描述 | 状态 |
|------|--------|------|------|
| **字符串转义修复** | 🔴 P0 | Go generator 转义顺序修正，`\n` → 换行符 | ✅ v1.1.3 |
| **多文件自举联编** | 🔴 P0 | main.klx 读取 6 文件 + GenerateMulti | ✅ v1.1.3 |
| **软关键字扩展** | 🟠 P1 | IsIdentOrSoftKeyword 25+ tokens | ✅ v1.1.3 |
| **Prefix parse 注册** | 🟠 P1 | 17 个缺失的 prefix（exit, import 等） | ✅ v1.1.3 |
| **Class 类型 unwrap** | 🟠 P1 | TClassDecl/TInterfaceDecl 在 TTypeDecl 中展开 | ✅ v1.1.3 |
| **类方法 receiver** | 🔴 P0 | ClassName.MethodName → func (self *CN) Method | ✅ v1.1.4 |
| **软关键字方法名** | 🔴 P0 | Write/Read/New 等软关键字作为方法名 | ✅ v1.1.4 |
| **字符串转义 (Go输出)** | 🔴 P0 | WriteEscapedGoString 转义 \ 和 " | ✅ v1.1.5 |
| **基类类型映射** | 🔴 P0 | TNode/TStatement/TExpression → interface{} | ✅ v1.1.5 |
| **枚举类型声明** | 🔴 P0 | 生成 type Name int 声明 | ✅ v1.1.5 |
| **内置函数完善** | 🔴 P0 | StrToInt64/StrToFloat/append/Exit/Create | ✅ v1.1.5 |
| **多参数解析** | 🔴 P0 | a, b: Type 多变量参数声明 | ✅ v1.1.5 |
| **多文件 Go 编译通过** | 🔴 P0 | 136KB 输出零错误，binary 运行正常 | ✅ v1.1.5 |

### 🟡 待完成

| 任务 | 优先级 | 描述 |
|------|--------|------|
| **完整 diff 验证** | 🔴 P0 | Go 版输出 vs Kylix 版输出逐行对比 |
| **示例文件 Kylix 版验证** | 🟡 P2 | 用 Kylix 编译器编译 14 个示例文件 |

### 🟡 示例文件通过率

| 状态 | 数量 | 文件 |
|------|------|------|
| ✅ 通过 | 14/15 | hello, simple, types, control, classes, modern, exceptions, stdlib_demo, test_formatter, orm_example, functions, web_demo, test_map, web_fullstack |
| ❌ 失败 | 1 | web_advanced（Go 语法混入 Kylix 代码） |

---

## Phase 9: 自举验证 ✅ 完成

### ✅ Diff 验证通过 (v1.2.0)

| 维度 | Go 参考版 | Kylix 自举版 | 结果 |
|------|----------|-------------|------|
| 函数数量 | 136 | 136 | ✅ 一致 |
| 类型定义 | 66 | 66 | ✅ 一致 |
| 常量块 | 10 | 10 | ✅ 一致 |
| 函数签名差异 | — | — | 3 个（格式，语义等价） |
| Go 编译 | ✅ | ✅ | 两者都能编译 |
| 运行时行为 | ✅ | ✅ | 语义等价 |

### 🟡 待完成

| 任务 | 优先级 | 描述 |
|------|--------|------|
| **示例文件 Kylix 版验证** | 🟡 P2 | 用 Kylix 编译器编译 14 个示例文件 |
| **web_advanced 修复** | 🟡 P2 | 清理混入的 Go 语法 |

### 自举完成！

```
Go 版编译器 (kylix)
    ↓ 编译 7 个 .klx 文件
Go 代码 (main.go)          ✅ 零编译错误
    ↓ go build
Kylix 编译器 (binary)       ✅ 可运行
    ↓ 运行 (多文件联编)
输出 (141KB)                ✅ 合并输出正确
    ↓ Go 编译
输出                        ✅ 零错误，运行正常
    ↓ Diff 验证
Go版 vs Kylix版             ✅ 语义等价！
```

---

## v1.2.3 (2026-06-12) — 代码重构：每文件不超过 1000 行 ✅

### 目标
强制执行每个源文件最多 1000 行的限制，提升可读性和可维护性。无行为变更。

### 拆分结果

| 文件 | 原始行数 | 拆分后文件 | 最大行数 |
|------|---------|-----------|---------|
| `parser/parser.go` | 2271 | `parser.go` + `parser_decl.go` + `parser_stmt.go` + `parser_expr.go` | 685 |
| `generator/generator.go` | 1979 | `generator.go` + `generator_types.go` + `generator_stmt.go` + `generator_expr.go` | 631 |
| `pkg/lsp/server.go` | 1238 | `server.go` + `handler_completion.go` + `handler_navigation.go` | 523 |
| `stdlib/orm.go` | 964 | `orm.go` + `orm_query.go` + `orm_migrate.go` | 410 |
| `pkg/formatter/formatter.go` | 897 | `formatter.go` + `formatter_stmt.go` + `formatter_expr.go` | 396 |

### 规则落地
CLAUDE.md 新增约束：**每个源文件不超过 1000 行，按功能职责拆分**。

---

## v2.0.0 规划 — 生产级自举编译器

> 当前状态：自举完成（v1.2.x），代码库整洁（v1.2.3）。v2.0 聚焦让 Kylix 从"能用"走向"好用"。

### 优先级功能清单

#### P0 — 编译器诊断质量（影响每次开发体验）

| 特性 | 描述 |
|------|------|
| **Kylix 层错误报告** | 目前错误来自 Go 编译器，用户看到 Go 语法报错。需在 Kylix 层捕获并报告行列号、友好信息 |
| **未使用变量检测** | 目前靠 Go 的 `_ = varName` hack 抑制，应在 Kylix 层实现 |
| **类型不匹配提示** | 赋值/调用时类型不兼容给出明确提示，而非 Go 类型断言 panic |

#### P1 — 语言完整性（影响能写什么程序）

| 特性 | 描述 |
|------|------|
| **接口实现编译时验证** | `class Foo implements IBar` 时检查 Foo 是否实现了 IBar 的所有方法 |
| **泛型代码生成实现** | 当前泛型是空壳，生成 `interface{}`；需生成真实 Go 泛型 `[T any]` |
| **多返回值完整支持** | 目前 `result := (a, b)` 有场景不覆盖；统一为 Go 标准多返回值模式 |
| **字符串内置方法** | `s.Length`、`s.ToUpper()`、`s.Contains(sub)` 等方法语法 |

#### P2 — 工具链（影响日常工作流）

| 特性 | 描述 |
|------|------|
| **`kylix build --target=<os>/<arch>`** | 跨平台交叉编译，调用 Go 的 `GOOS/GOARCH` |
| **增量编译** | 仅重新编译变更的 unit，加速大项目构建 |
| **LSP 诊断增强** | 实时在编辑器内显示 Kylix 层错误（目前 LSP 只做语法分析） |
| **REPL 支持 class/uses** | 当前 REPL 不支持类和跨文件声明 |

#### P3 — stdlib 迁移（长期目标）

| 特性 | 描述 |
|------|------|
| **stdlib 逐步 Kylix 化** | 将 `web.go`、`orm.go` 等逐步用 Kylix 重写，成为语言自举的一部分 |
| **包管理器** | `kylix add <package>` 命令，管理第三方依赖 |

### 建议启动顺序

1. **P0: Kylix 层错误报告** — 最高价值，直接改善每次编写代码的体验
2. **P1: 接口实现验证** — 语法已支持，只差语义检查，工作量小收益高
3. **P1: 泛型真实实现** — 让现有泛型语法真正可用
4. **P2: 跨平台编译** — 一行 Go 调用，实现成本极低

---

## 相关文档

- [CHANGELOG.md](CHANGELOG.md) — 版本发布历史
- [TASKS.md](TASKS.md) — 详细任务拆解
- [README.md](README.md) — 项目介绍