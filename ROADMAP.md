# Kylix Development Roadmap

> 最后更新: 2026-06-06
> 当前版本: v1.1.0
> 官网: [kylix.top](https://kylix.top)
> 目标: Kylix 语言自举（用 Kylix 写 Kylix 编译器）

---

## 总览

本项目目标分四个阶段：

| 阶段 | 内容 | 状态 | 预计工期 |
|------|------|------|----------|
| Phase 6 | 修复关键 Bug | ✅ 完成 | ~2 周 |
| Phase 7 | 补齐语言能力 | ✅ 完成 | ~3 周 |
| Phase 8 | 编写 compiler.klx | 🚧 80% | ~4 周 |
| Phase 9 | 自举验证 | ⬜ 0% | ~1 周 |

**当前进度：Phase 8 进行中，Go 编译器升级完成，7 个 .klx 源文件已编写，联合编译通过**

**总计剩余工期：约 2 周**

---

## Phase 6: 修复关键 Bug → v1.0.2 ✅ 已完成

### ✅ 全部修复

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

### ✅ 已完成

| 特性 | 严重性 | 描述 |
|------|--------|------|
| **Map 类型** | 🔴 P0 | `map[K]V` 语法，自动初始化，索引读写 |
| **变体类型** | 🔴 P0 | `variant Case: Type; end` → Go interface + struct 模式 |
| **动态数组** | 🔴 P0 | `append(arr, elem)` / `SetLength(arr, n)` 内置函数 |

---

## Phase 8: 编写 compiler.klx → v1.1.0 🚧 80%

### ✅ 已完成 — Go 编译器后端升级

| 特性 | 描述 |
|------|------|
| **枚举类型** | `(val1, val2, ...)` → Go `const` + `iota` |
| **Slice 表达式** | `s[a:b]` → Go slice 语法 |
| **Unit 文件系统** | `unit X;` 模块声明 |
| **多文件联编** | `kylix build a.klx b.klx` |
| **类代码生成** | 混合 struct/interface 方案，基类用 `interface{}` 实现多态 |
| **软关键字扩展** | 25+ 关键字可作为标识符使用 |
| **局部 var/const** | 函数体内局部声明生成 |
| **Exit 语句** | `exit` → `return result`（有返回值）或 `return`（过程） |
| **Bare method calls** | `self.Method` → `self.Method()` |
| **Map 作为表达式** | `map[K]V` 作为值生成 `map[K]V{}` |
| **新内置函数** | Ord, Length, IntToStr, StrToInt64, StrToFloat |
| **构造函数** | `T.Create` → `&T{}`，`T.Create(args)` → `&T{args...}` |

### ✅ 已完成 — Kylix 源文件

```
src/
├── token.klx           # Token 类型定义、关键字映射 (209 行)
├── ast.klx             # AST 节点类型定义 (374 行)
├── lexer.klx           # 词法分析器 (366 行)
├── parser.klx          # Pratt 语法分析器 (2338 行)
├── error.klx           # 错误类型和定位 (91 行)
├── generator.klx       # Go 代码生成器 (221 行，骨架)
└── main.klx            # 入口，串联编译器 (56 行)
```

总计 3655 行 Kylix 代码。7 文件联合编译通过（Kylix → Go 转换成功）。

### 🟡 待完成 — 自举剩余工作

| 任务 | 描述 |
|------|------|
| **generator.klx 完善** | 骨架代码需补充完整实现（类型分发、表达式生成等） |
| **Go 编译错误修复** | 生成的 Go 代码有 ~6 个类型/API 兼容问题 |
| **构造函数参数映射** | `T.Create(arg)` 需映射到正确的字段名 |
| **main.klx 完善** | 文件读取、错误处理完善 |

### 🟡 示例文件通过率

| 状态 | 数量 | 文件 |
|------|------|------|
| ✅ 通过 | 14/15 | hello, simple, types, control, classes, modern, exceptions, stdlib_demo, test_formatter, orm_example, functions, web_demo, test_map, web_fullstack |
| ❌ 失败 | 1 | web_advanced（Go 语法混入 Kylix 代码） |

---

## Phase 9: 自举验证 ⬜ 0%

验证标准：
1. `compiler.klx` 能被 Go 版编译器编译通过
2. 编译出的 `compiler` (Go binary) 能编译 `compiler.klx` 自身
3. 两次编译输出一致（字节级或语义级）
4. 输出与 Go 版编译器一致

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
