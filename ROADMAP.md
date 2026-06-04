# Kylix Development Roadmap

> 最后更新: 2026-06-04
> 当前版本: v1.0.2
> 目标: Kylix 语言自举（用 Kylix 写 Kylix 编译器）

---

## 总览

本项目目标分四个阶段：

| 阶段 | 内容 | 状态 | 预计工期 |
|------|------|------|----------|
| Phase 6 | 修复关键 Bug | ✅ 完成 | ~2 周 |
| Phase 7 | 补齐语言能力 | ⬜ 0% | ~3 周 |
| Phase 8 | 编写 compiler.klx | ⬜ 0% | ~4 周 |
| Phase 9 | 自举验证 | ⬜ 0% | ~1 周 |

**当前进度：Phase 6 完成，示例通过率 13/14 (93%)**
| Phase 7 | 补齐语言能力 | ⬜ 0% | ~3 周 |
| Phase 8 | 编写 compiler.klx | ⬜ 0% | ~4 周 |
| Phase 9 | 自举验证 | ⬜ 0% | ~1 周 |

**总计剩余工期：约 9 周**

---

## Phase 6: 修复关键 Bug → v1.0.2 ✅ 已完成

### ✅ 全部修复

| Bug | 严重性 | 描述 |
|-----|--------|------|
| **字符串插值** | 🔴 P1 | Lexer→Parser→Generator 全链路修复，`$"Hello, ${name}!"` → `fmt.Sprintf("Hello, %v!", name)` |
| **异常类型** | 🔴 P1 | `raise Exception.Create('msg')` → `panic(&Exception{Message: "msg"})`，`on E: Type do` → `case *Type:` |
| **多返回值** | 🟠 P2 | `function Divide(...): (Real, Boolean)` + `result := (0, false)` → `return 0, false` |
| **Properties** | 🟠 P2 | `property Name: String read Field;` → `func (self *T) Name() string { return self.Field }` |
| **web_demo 匿名过程** | 🟠 P2 | 嵌套 record 类型的 `end` 深度追踪修复 |
| **数组范围大小** | 🟠 P2 | `array[0..2]` → 正确计算 `((2-0)+1) = 3` |
| **内存泄漏** | 🔴 P0 | `parseGroupedExpression` 多余 `nextToken()` 导致无限循环，改用 `peekToken` 检测

### 🟡 示例文件通过率

| 状态 | 数量 | 文件 |
|------|------|------|
| ✅ 通过 | 13/14 (93%) | hello, simple, types, control, classes, modern, exceptions, stdlib_demo, test_formatter, web_advanced, orm_example, functions, web_demo |
| ❌ 失败 | 1/14 (7%) | web_fullstack.klx (Go struct 字面量 `TConnectionConfig{...}` 语法 — 非 Kylix 语法) |

---

## Phase 7: 补齐语言能力

要实现编译器自举，Kylix 需要以下新能力：

### P0 - 编译器自举必需

| 特性 | 用途 | 设计要点 |
|------|------|----------|
| **Map 类型** `map[K]V` | 符号表、关键字映射、作用域管理 | 直接映射到 Go `map[K]V`；支持 `map[String]Integer`、`map[String]TToken` |
| **变体类型 / Discriminated Union** | AST 节点类型安全表示 | 语法：`type TExpr = variant IntegerLit of Integer; StringLit of String; end;` → 生成 Go interface + 具体类型 |
| **动态数组** (`append`, 可写 `len`) | Token 列表、错误列表、AST 子节点 | 映射到 Go slice，支持 `append(arr, elem)` 和 `SetLength(arr, n)` |

### P1 - 编译器质量提升

| 特性 | 用途 | 设计要点 |
|------|------|----------|
| **枚举类型** | Token 类型定义 | 语法：`type TTokenType = (tkEOF, tkIdent, tkInteger, ...);` → Go `const` + `iota` |
| **多文件模块系统** | 编译器模块化 (token.klx, lexer.klx, ...) | 扩展 `uses` 子句支持同项目多文件 |
| **接口/协议实现** | AST Visitor 模式 | 需要 `implements` 关键字正确工作 |

### P2 - 完善

| 特性 | 用途 |
|------|------|
| **单元测试框架** | 编译器回归测试 |
| **18 个未处理 token** | 清理 token 定义 |

---

## Phase 8: 编写 compiler.klx

用 Kylix 重写编译器，文件结构：

```
src/
├── token.klx           # Token 类型定义、关键字映射 (~150 行)
├── lexer.klx           # 词法分析器 (~300 行)
├── ast.klx             # AST 节点类型定义 (~400 行)
├── parser.klx          # Pratt 语法分析器 (~800 行)
├── generator.klx       # Go 代码生成器 (~600 行)
├── error.klx           # 错误类型和定位 (~100 行)
└── main.klx            # 入口，串联编译器 (~100 行)
```

总计约 2450 行 Kylix 代码。

---

## Phase 9: 自举验证

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
