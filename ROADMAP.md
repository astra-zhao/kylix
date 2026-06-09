# Kylix Development Roadmap

> 最后更新: 2026-06-08
> 当前版本: v1.1.4
> 官网: [kylix.top](https://kylix.top)
> 目标: Kylix 语言自举（用 Kylix 写 Kylix 编译器）

---

## 总览

本项目目标分四个阶段：

| 阶段 | 内容 | 状态 | 预计工期 |
|------|------|------|----------|
| Phase 6 | 修复关键 Bug | ✅ 完成 | ~2 周 |
| Phase 7 | 补齐语言能力 | ✅ 完成 | ~3 周 |
| Phase 8 | 编写 compiler.klx | ✅ 完成 | ~4 周 |
| Phase 9 | 自举验证 | 🚧 80% | ~2 天 |

**当前进度：Phase 8 完成，Phase 9 推进至 80%。v1.1.4 修复了类方法 receiver 生成（ClassName.MethodName → self *ClassName），126 个类方法全部正确。**

**总计剩余工期：约 2 天**

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

### ✅ v1.1.4 完成内容

| 任务 | 优先级 | 描述 | 状态 |
|------|--------|------|------|
| **字符串转义修复** | 🔴 P0 | Go generator 转义顺序修正，`\n` → 换行符 | ✅ v1.1.3 |
| **多文件自举联编** | 🔴 P0 | main.klx 读取 6 文件 + GenerateMulti | ✅ v1.1.3 |
| **软关键字扩展** | 🟠 P1 | IsIdentOrSoftKeyword 25+ tokens | ✅ v1.1.3 |
| **Prefix parse 注册** | 🟠 P1 | 17 个缺失的 prefix（exit, import 等） | ✅ v1.1.3 |
| **Class 类型 unwrap** | 🟠 P1 | TClassDecl/TInterfaceDecl 在 TTypeDecl 中展开 | ✅ v1.1.3 |
| **类方法 receiver** | 🔴 P0 | ClassName.MethodName → func (self *CN) Method | ✅ v1.1.4 |
| **软关键字方法名** | 🔴 P0 | Write/Read/New 等软关键字作为方法名 | ✅ v1.1.4 |
| **自举验证 (7 源文件)** | 🔴 P0 | 全部 7 个 .klx 源文件成功编译 | ✅ v1.1.4 |
| **多文件合并输出** | 🔴 P0 | 6 文件合并输出 135KB Go 代码 | ✅ v1.1.4 |

### 🟡 待完成

| 任务 | 优先级 | 描述 |
|------|--------|------|
| **Go 字符串转义残余** | 🔴 P0 | 引号嵌套 `""fmt""`、raw newline in strings |
| **多文件 Go 编译通过** | 🔴 P0 | 当前 135KB 输出有字符串转义语法错误 |
| **完整 diff 验证** | 🔴 P0 | Go 版输出 vs Kylix 版输出逐行对比 |
| **示例文件 Kylix 版验证** | 🟡 P2 | 用 Kylix 编译器编译 14 个示例文件 |

### 🟡 示例文件通过率

| 状态 | 数量 | 文件 |
|------|------|------|
| ✅ 通过 | 14/15 | hello, simple, types, control, classes, modern, exceptions, stdlib_demo, test_formatter, orm_example, functions, web_demo, test_map, web_fullstack |
| ❌ 失败 | 1 | web_advanced（Go 语法混入 Kylix 代码） |

---

## Phase 9: 自举验证 🚧 80%

### ✅ 已验证 (v1.1.4)

1. ✅ **Kylix → Go 编译通过**
2. ✅ **Go 代码编译通过**
3. ✅ **Binary 运行**
4. ✅ **Lexer→Parser→Error 管道**
5. ✅ **Lexer bug 修复** (v1.1.1)
6. ✅ **Generator 完善** ~1400 行 (v1.1.1)
7. ✅ **简单程序自举** (v1.1.1)
8. ✅ **Parser result 覆盖修复** 7 个函数 (v1.1.2)
9. ✅ **代码生成修复** Record/Map/Local/ReadFile (v1.1.2)
10. ✅ **7 源文件全编译** (v1.1.2)
11. ✅ **软关键字 + prefix** 25+/17 (v1.1.3)
12. ✅ **字符串转义修复** (v1.1.3)
13. ✅ **多文件自举联编** 135KB (v1.1.3)
14. ✅ **类方法 receiver** 126 个方法全部正确 (v1.1.4)

### 🟡 待完成

| 步骤 | 状态 | 描述 |
|------|------|------|
| 9.1-9.5 | ✅ | Go 编译、简单程序、7 文件、多文件、receiver 全部通过 |
| 9.6 Go 编译多文件输出 | 🟡 | 字符串转义残余（引号嵌套、raw newline）|
| 9.7 Diff 验证 | ⬜ | 待 9.6 通过 |

### 自举管道架构

```
Go 版编译器 (kylix)
    ↓ 编译 7 个 .klx 文件
Go 代码 (main.go)          ✅ 零编译错误
    ↓ go build
Kylix 编译器 (binary)       ✅ 可运行
    ↓ 运行 (逐个源文件)
输出                        ✅ 全部编译成功
    ↓ 运行 (多文件联编)
输出 (135KB)                ✅ 合并输出正确
    ↓ Go 编译
输出                        🟡 字符串转义残余
    ↓ Diff 验证
Go版 vs Kylix版             ⬜ 待 Go 编译通过
```

---

## 长期愿景（Phase 10+）

- LSP 增强（代码补全、诊断、重构）
- REPL 改进（支持 uses/class 声明）
- 标准库扩展
- 性能优化
- 包管理器

---

## 相关文档

- [CHANGELOG.md](CHANGELOG.md) — 版本发布历史
- [TASKS.md](TASKS.md) — 详细任务拆解
- [README.md](README.md) — 项目介绍