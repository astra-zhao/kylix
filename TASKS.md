# Kylix 开发任务清单

> 最后更新: 2026-06-06
> 官网: [kylix.top](https://kylix.top)
> 关联文档: [ROADMAP.md](ROADMAP.md), [CHANGELOG.md](CHANGELOG.md)
> 当前版本: v1.1.0

---

## Phase 6: 修复关键 Bug → v1.0.2 ✅ 已完成

### 6.1 字符串插值修复 (P1) ✅

- [x] Parser — `parseStringInterpolation()` 解析 `${expression}` 段落
- [x] Generator — `writeInterpolation()` 生成 `fmt.Sprintf(format, args...)`
- [x] Import 扫描 — `scanExpressionForImports` 支持 `*ast.StringInterpolation`

### 6.2 异常类型定义 (P1) ✅

- [x] `stdlib/exceptions.go` — Exception 类型及子类型定义
- [x] Generator — `raise` → `panic(&Exception{Message: "msg"})`
- [x] Generator — `on E: Type do` → `case *Type:` 类型 switch
- [x] Auto-generate Exception 结构体和子类型内联
- [x] `scanForException` 预扫描通过

### 6.3 多返回值支持 (P2) ✅

- [x] Parser — `: (Type1, Type2)` 元组返回类型解析
- [x] Parser — `(expr1, expr2)` 元组字面量解析
- [x] Parser — `var (a, b) := expr` 解构赋值
- [x] AST — `TupleLiteral` 和 `ReturnTypes` 字段
- [x] Generator — 多返回值签名、赋值、return 生成

### 6.4 Properties 代码生成 (P2) ✅

- [x] Generator — `generatePropertyAccessors()` 生成 getter/setter

### 6.5 web_demo.klx 边界情况 (P2) ✅

- [x] Record 类型嵌套深度追踪
- [x] `parseGroupedExpression` 内存泄漏修复

---

## Phase 7: 补齐语言能力 → v1.0.3 ✅ 已完成

### 7.1 Map 类型 (P0) ✅

- [x] Token — `MAP` token + `"map"` 关键字
- [x] AST — `MapType` 节点
- [x] Parser — `parseTypeExpression()` 中 `map[K]V` 解析
- [x] Generator — `map[K]V` → Go `map[K]V`，自动初始化

### 7.2 变体类型 (P0) ✅

- [x] Token — `VARIANT` token
- [x] AST — `VariantType` + `VariantCase` 节点
- [x] Parser — `variant ... end` 解析
- [x] Generator — Go `interface` + `struct` + marker method

### 7.3 动态数组 (P0) ✅

- [x] Builtin — `append`, `SetLength`
- [x] Generator — 特殊 ExpressionStatement 处理

### 7.4 枚举类型 (P1) ✅ v1.1.0

- [x] AST — `EnumType` 节点
- [x] Parser — `tryParseEnumType()`
- [x] Generator — Go `const` + `iota`

### 7.5 多文件模块系统 (P1) ✅ v1.1.0

- [x] Parser — `unit X;` → `Program.UnitName`, `Program.IsUnit`
- [x] Compiler API — `CompileProject(files, opts)` + 拓扑排序
- [x] Generator — `GenerateMulti([]*Program)`
- [x] CLI — `kylix build a.klx b.klx` 多文件模式
- [x] CLI — `kylix run` 自动发现 .klx 文件

### 7.6 接口实现验证 (P1) — 延后

- [ ] Parser — `implements` 子句
- [ ] 编译时接口实现检查

---

## Phase 8: 编写 compiler.klx → v1.1.0 🚧 85%

### 8.1 Go 编译器后端升级 ✅

- [x] Slice 表达式: `s[a:b]` AST + Parser + Generator
- [x] 类代码生成: struct-only 方案，基类 `interface{}`，具体类 `*T`
- [x] 类字段收集: `classFields` map 用于构造函数字段名映射
- [x] 软关键字扩展: 25+ 关键字可作为标识符
- [x] 局部 var/const: `FunctionDecl.LocalDecls` 存储 + 生成 + `_ = name` 占位
- [x] Exit 语句: `exit` → `return result` (有返回值) / `return` (过程)
- [x] 构造函数: `T.Create` → `&T{}`, `T.Create(args)` → `&T{Field: arg}`
- [x] Bare method call: `self.Method` 作为 statement → `self.Method()`
- [x] Map 作为表达式: `map[K]V` 注册为 prefix parse fn → `map[K]V{}`
- [x] 空数组字面量: `[]` → `nil`
- [x] 字符串转义: `\`, `"`, `\n` 正确转义
- [x] 新内置函数: `Ord`, `Length`, `IntToStr`, `StrToInt64`, `StrToFloat`, `ReadFile`
- [x] for 循环: `for i = 0` 避免 int/int64 类型冲突
- [x] `Delete` 关键字可作为函数名
- [x] Class 字段解析安全保护 (peekTokenIs COLON)
- [x] is/as 类型分派: `is` → type assertion check，`as` → type assertion

### 8.2 Kylix 源文件 ✅

- [x] `src/token.klx` — 209 行，Token 类型枚举、关键字映射
- [x] `src/ast.klx` — 374 行，AST 节点类层次 (54 类)
- [x] `src/lexer.klx` — 366 行，词法分析器
- [x] `src/parser.klx` — 2338 行，Pratt 解析器
- [x] `src/error.klx` — 91 行，编译器错误类型
- [x] `src/generator.klx` — 221 行，Go 代码生成器 (骨架)
- [x] `src/main.klx` — 56 行，入口点（文件读取 + 错误处理 + fallback）

**7 文件联合编译: ✅ Kylix → Go 转换 + Go 编译零错误**

### 8.3 已知 Bug 🐛

| Bug | 优先级 | 描述 |
|-----|--------|------|
| **Kylix lexer tokenization** | 🔴 P0 | 有效 Pascal 关键字被识别为 tkIllegal。根因：Kylix 版 lexer 中 `IsLetter`/`IsDigit` 字符串比较逻辑或 `NextToken` 分发逻辑有 bug |
| **Single-quoted string escaping** | 🟠 P1 | Kylix `'...'` 字符串转为 Go `"..."` 时，内部单引号转义不正确 |
| **web_advanced Go syntax** | 🟡 P2 | 示例文件混入 Go 语法 `map[string]interface{}`，非 Kylix 合法语法 |

### 8.4 待完成 🟡

- [ ] 8.4a 修复 Kylix lexer tokenization bug（需要调试 Kylix 版 `lexer.klx` 和 `token.klx`）
- [ ] 8.4b 完善 `generator.klx` 骨架代码（用 `is`/`as` 实现类型分发、表达式生成）
- [ ] 8.4c 修复单引号字符串转义

---

## Phase 9: 自举验证 🚧 30%

### 9.1 Go 版编译器编译 compiler.klx ✅

- [x] Go 版编译器成功编译 7 个 .klx 文件 → Go 代码
- [x] 生成的 Go 代码零编译错误 → 自举 binary

### 9.2 自举 binary 编译 compiler.klx 自身 🟡

- [x] Binary 可运行，lexer→parser→error 管道工作
- [ ] Lexer tokenization bug 导致解析失败（无法完成完整编译）

### 9.3 两次输出 diff 验证 ⬜

- [ ] 待 lexer bug 修复

### 9.4 示例文件输出一致 ⬜

- [x] Go 版编译器: 14/15 示例通过
- [ ] Kylix 版编译器: 待 lexer 修复

### 9.5 回归测试 ✅

- [x] Go 测试全部通过
- [x] 14/15 示例文件在 Go 版编译器下通过

### 自举管道详解

```
Step 1: Go 版编译器
  7 个 .klx 文件 ──→ Go 代码 (main.go)          ✅ 零编译错误
                            ↓ go build
                      kylix_compiler (binary)      ✅ 可运行

Step 2: Kylix 编译器
  input.klx ──→ kylix_compiler ──→ 输出           🟡 Lexer bug
                                          错误信息正确输出，但 tokenization 有误
```

---

## 后续版本规划

| 版本 | 内容 | 状态 | 预计日期 |
|------|------|------|----------|
| v1.0.2 | Phase 6 完成 | ✅ | 2026-06-04 |
| v1.0.3 | Phase 7 P0 完成 | ✅ | 2026-06-05 |
| v1.1.0 | Phase 8 85% — Go 后端升级 + 7 .klx 文件 | ✅ | 2026-06-06 |
| v1.1.1 | 修复 Kylix lexer bug + 完善 generator.klx | 🟡 | ~3 天 |
| v1.2.0 | 自举验证通过 | ⬜ | ~1 周 |
| v2.0.0 | 完整自举发布 | ⬜ | ~2 周 |

---

## 关键设计决策记录

1. **类多态方案**: struct-only + `interface{}`。基类（TNode/TStatement/TExpression）用 `interface{}`，具体类用 `*T`。
   - 优点: 字段访问简单，无需 interface method forwarding
   - 缺点: 基类类型变量不能直接访问字段，需要 `as` 转换

2. **软关键字策略**: 所有 Pascal 关键字在成员位置（`.` 后面）可用作标识符
   - 通过 `isSoftKeyword()` 和 `parseMemberExpression()` 实现

3. **自举循环**: Go 版编译器必须先行完善，Kylix 版编译器才能工作
   - 当前优先修复 Go 版，让 Kylix 版逐步追赶
