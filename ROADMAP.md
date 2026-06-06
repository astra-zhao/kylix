# Kylix Development Roadmap

> 最后更新: 2026-06-06
> 当前版本: v1.1.1
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
| Phase 9 | 自举验证 | 🚧 40% | ~1 周 |

**当前进度：Phase 8 完成，lexer bug 已修复，generator.klx 已完善，简单程序自举通过，复杂源文件自举待完善**

**总计剩余工期：约 1 周**

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

## Phase 8: 编写 compiler.klx → v1.1.1 ✅ 已完成

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

### ✅ v1.1.1 完成内容

| 任务 | 优先级 | 描述 | 状态 |
|------|--------|------|------|
| **Kylix lexer tokenization bug** | 🔴 P0 | 两个根因：(1) `LookupIdent` 单值 map 查找返回零值；(2) `TParser.Create` 未初始化 token 状态 | ✅ 已修复 |
| **generator.klx 完善** | 🔴 P0 | 221 行 → ~1350 行，is/as 类型分发，15+ 语句类型，20+ 表达式类型 | ✅ 已完成 |
| **自举验证 (简单程序)** | 🔴 P0 | `program hello; begin WriteLn(42); end.` → 合法 Go 代码 | ✅ 已通过 |

### 🟡 待完成

| 任务 | 优先级 | 描述 |
|------|--------|------|
| **复杂源文件自举** | 🔴 P0 | token.klx 等复杂文件编译有局部变量和参数类型问题 |
| **Kylix AST TFunctionDecl.LocalDecls** | 🟠 P1 | Kylix AST 缺少此字段，局部 var/const 无法传递给生成器 |
| **单引号字符串转义** | 🟠 P1 | Kylix `'...'` → Go `"..."` 转义 |
| **完整 diff 验证** | 🟡 P2 | Go 版输出 vs Kylix 版输出逐行对比 |

### 🟡 示例文件通过率

| 状态 | 数量 | 文件 |
|------|------|------|
| ✅ 通过 | 14/15 | hello, simple, types, control, classes, modern, exceptions, stdlib_demo, test_formatter, orm_example, functions, web_demo, test_map, web_fullstack |
| ❌ 失败 | 1 | web_advanced（Go 语法混入 Kylix 代码） |

---

## Phase 9: 自举验证 🚧 40%

### ✅ 已验证 (v1.1.1)

1. ✅ **Kylix → Go 编译通过** — Go 版编译器成功编译 7 个 .klx 文件
2. ✅ **Go 代码编译通过** — 生成的 Go 代码零编译错误
3. ✅ **Binary 运行** — Kylix 编译器 binary 可以运行
4. ✅ **Lexer→Parser→Error 管道** — 全链路工作，错误信息正确输出
5. ✅ **Lexer bug 修复** — 两个根因均已修复
6. ✅ **Generator 完善** — 221 行骨架 → ~1350 行完整实现
7. ✅ **简单程序自举** — `program hello; begin WriteLn(42); end.` → 合法 Go 代码

### 🟡 待完成

| 步骤 | 状态 | 描述 |
|------|------|------|
| 9.1 Go 版编译器编译 compiler.klx | ✅ | 已完成 |
| 9.2 编译出的 binary 编译简单程序 | ✅ | v1.1.1 已验证 |
| 9.3 编译出的 binary 编译 compiler.klx 自身 | 🟡 | 复杂源文件有局部变量/参数类型问题 |
| 9.4 两次输出 diff 验证 | ⬜ | 待 9.3 通过 |
| 9.5 示例文件输出一致 | ⬜ | 待 9.3 通过 |
| 9.6 回归测试 | ✅ | 14/15 示例通过 |

### 自举管道架构

```
Go 版编译器 (kylix)
    ↓ 编译 7 个 .klx 文件
Go 代码 (main.go)          ✅ 零编译错误
    ↓ go build
Kylix 编译器 (binary)       ✅ 可运行
    ↓ 运行 (简单程序)
输出                        ✅ 合法 Go 代码

    ↓ 运行 (复杂源文件)
输出                        🟡 局部变量/参数类型问题
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