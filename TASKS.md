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

### 额外修复

- [x] 数组范围 `array[0..2]` 大小正确计算
- [x] `parseGroupedExpression` 死循环修复

---

## Phase 7: 补齐语言能力 → v1.0.3 ✅ 已完成

### 7.1 Map 类型 (P0) ✅

- [x] Token — `MAP` token + `"map"` 关键字
- [x] AST — `MapType` 节点 (`KeyType`, `ValueType`)
- [x] Parser — `parseTypeExpression()` 中 `map[K]V` 解析
- [x] Generator — `map[K]V` → Go `map[K]V`，自动初始化 `{}`
- [x] 索引读写（复用已有 `IndexExpression`）
- [x] 示例: `examples/test_map.klx`

### 7.2 变体类型 / Discriminated Union (P0) ✅

- [x] Token — `VARIANT` token + `"variant"` 关键字
- [x] AST — `VariantType` + `VariantCase` 节点
- [x] Parser — `variant CaseName: Type; ... end` 解析
- [x] Generator — Go `interface` + 具体 `struct` + marker method
- [x] 命名规则: `TExpr_IntLit`, `TExpr_StrLit` 等

### 7.3 动态数组 (P0) ✅

- [x] Builtin — `append`, `SetLength` 注册到 builtinMap
- [x] Generator — `append(arr, elem)` → `arr = append(arr, elem)`
- [x] Generator — `SetLength(arr, n)` → `arr = arr[:n]`
- [x] 作为 ExpressionStatement 特殊处理，自动赋值

### 7.4 枚举类型 (P1) ✅ v1.1.0

- [x] AST — `EnumType` 节点
- [x] Parser — `tryParseEnumType()` 解析 `(val1, val2);` 语法
- [x] Generator — Go `const` + `iota`

### 7.5 多文件模块系统 (P1) ✅ v1.1.0

- [x] Parser — `unit X;` 声明 → `Program.UnitName`, `Program.IsUnit`
- [x] Compiler API — `CompileProject(files, opts)` + 拓扑排序
- [x] Generator — `GenerateMulti([]*Program)` 多文件编译
- [x] CLI — `kylix build a.klx b.klx` 多文件模式
- [x] CLI — `kylix run` 自动发现 .klx 文件

### 7.6 接口实现验证 (P1) — 延后

- [ ] Parser — `implements` 子句
- [ ] 编译时接口实现检查

---

## Phase 8: 编写 compiler.klx → v1.1.0 🚧 80%

### 8.1 Go 编译器后端升级 ✅

- [x] Slice 表达式: `s[a:b]` AST + Parser + Generator
- [x] 类代码生成: 混合 struct/interface 方案，基类 `interface{}`，具体类 `*T`
- [x] 软关键字扩展: 25+ 关键字可作为标识符
- [x] 局部 var/const: `FunctionDecl.LocalDecls` 存储 + 生成
- [x] Exit 语句: `exit` → `return result` (有返回值) / `return` (过程)
- [x] 构造函数: `T.Create` → `&T{}`, `T.Create(args)` → `&T{args...}`
- [x] Bare method call: `self.Method` → `self.Method()`
- [x] Map 作为表达式: `map[K]V` 注册为 prefix parse fn
- [x] 空数组字面量: `[]` → `nil`
- [x] 字符串转义: `\`, `"`, `\n` 正确转义
- [x] 新内置函数: `Ord`, `Length`, `IntToStr`, `StrToInt64`, `StrToFloat`
- [x] for 循环: `for i = 0` 避免类型冲突
- [x] `Delete` 关键字可作为函数名
- [x] Class 字段解析安全保护 (peekTokenIs COLON)

### 8.2 Kylix 源文件 ✅

- [x] `src/token.klx` — 209 行，Token 类型枚举、关键字映射
- [x] `src/ast.klx` — 374 行，AST 节点类层次 (54 类)
- [x] `src/lexer.klx` — 366 行，词法分析器
- [x] `src/parser.klx` — 2338 行，Pratt 解析器
- [x] `src/error.klx` — 91 行，编译器错误类型
- [x] `src/generator.klx` — 221 行，Go 代码生成器 (骨架)
- [x] `src/main.klx` — 56 行，入口点

**7 文件联合编译通过（Kylix → Go 转换成功）**

### 8.3 待完成 🟡

- [ ] 8.3a 修复生成 Go 代码中的 ~6 个类型/API 兼容问题
- [ ] 8.3b 完善 `generator.klx` 骨架代码
- [ ] 8.3c 构造函数参数字段名映射
- [ ] 8.3d 完善 `main.klx`（文件读取、错误处理）

---

## Phase 9: 自举验证 ⬜ 0%

- [ ] 9.1 Go 版编译器编译 compiler.klx → 自举 binary
- [ ] 9.2 自举 binary 编译 compiler.klx 自身
- [ ] 9.3 两次输出 diff 验证
- [ ] 9.4 示例文件输出一致
- [ ] 9.5 回归测试

---

## 版本规划

| 版本 | 内容 | 日期 |
|------|------|----------|
| v1.0.2 | ✅ Phase 6 完成 | 2026-06-04 |
| v1.0.3 | ✅ Phase 7 P0 完成 | 2026-06-05 |
| v1.1.0 | 🚧 Phase 8 80% — Go 后端升级 + .klx 源文件 | 2026-06-06 |
| v1.2.0 | Phase 8 收尾 — generator/main 完善，Go 编译修复 | ~1 周 |
| v2.0.0 | Phase 9 完成 — 自举验证通过 | ~2 周 |
