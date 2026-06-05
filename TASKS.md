# Kylix 开发任务清单

> 最后更新: 2026-06-05
> 官网: [kylix.top](https://kylix.top)
> 关联文档: [ROADMAP.md](ROADMAP.md), [CHANGELOG.md](CHANGELOG.md)
> 当前版本: v1.0.3

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

### 7.4 枚举类型 (P1) — 延后

- [ ] AST — `EnumType` 节点
- [ ] Parser — `type T = (val1, val2);` 语法
- [ ] Generator — Go `const` + `iota`

### 7.5 多文件模块系统 (P1) — 延后

- [ ] Compiler API — 多文件编译
- [ ] 跨文件符号可见性
- [ ] CLI — `kylix build` 项目编译

### 7.6 接口实现验证 (P1) — 延后

- [ ] Parser — `implements` 子句
- [ ] 编译时接口实现检查

---

## Phase 8: 编写 compiler.klx → v2.0.0

- [ ] 8.1 token.klx (~150 行)
- [ ] 8.2 lexer.klx (~300 行)
- [ ] 8.3 ast.klx (~400 行)
- [ ] 8.4 parser.klx (~800 行)
- [ ] 8.5 generator.klx (~600 行)
- [ ] 8.6 error.klx (~100 行)
- [ ] 8.7 main.klx (~100 行)

---

## Phase 9: 自举验证

- [ ] 9.1 Go 版编译器编译 compiler.klx
- [ ] 9.2 编译出的 binary 编译 compiler.klx 自身
- [ ] 9.3 两次输出 diff 验证
- [ ] 9.4 示例文件输出一致
- [ ] 9.5 回归测试

---

## 版本规划

| 版本 | 内容 | 预计日期 |
|------|------|----------|
| v1.0.2 | ✅ Phase 6 完成 | 2026-06-04 |
| v1.0.3 | ✅ Phase 7 P0 完成（Map、变体、动态数组） | 2026-06-05 |
| v1.1.0 | Phase 7 P1（枚举、多文件模块、接口验证） | ~2 周 |
| v2.0.0 | Phase 8-9 完成 | ~6 周 |
