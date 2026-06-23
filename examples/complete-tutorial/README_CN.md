# Kylix v3.1.0 完全教程

欢迎使用 Kylix 完全教程！本教程通过 **34 个经过测试的可运行示例**（其中 32 个完全工作，~94%），涵盖 Kylix v3.1.0 的所有已验证特性。

## 前置要求

- Kylix 编译器 (v3.1.0 或更高版本)
- Go 1.18+ (用于运行生成的代码)

---

## 🎉 v3.1.0 修复亮点

v3.0.0-alpha 的多数限制在 v3.1.0 已修复：

- ✅ `var p: TClass` 字段访问（KLX-C01）
- ✅ 字符串插值 `${var}` 展开（KLX-C02）
- ✅ Lambda 返回类型保留（KLX-C03）
- ✅ `match` 语句 codegen（KLX-C04）
- ✅ `uses sysutil/jsonutil/datetime/...` 在 program 中可用（KLX-C05）

新增：注解语法 `[Name]` + KylixBoot 框架（Spring Boot 式）。

---

## 📚 教程结构

本教程包含 **34 个示例**，分为 11 个类别：

### 1. 基础语法 (6 个示例) - `01_basics/` ✅

- `example01_hello.klx` — Hello World
- `example02_variables.klx` — 变量声明与类型
- `example03_constants.klx` — 常量
- `example04_type_inference.klx` — 类型推导 `:=`
- `example05_operators.klx` — 算术、比较、逻辑运算符
- `example06_comments.klx` — 单行注释

### 2. 控制流 (5 个示例) - `02_control_flow/` ✅

- `example07_if_else.klx` — If-then-else 条件语句
- `example08_while.klx` — While 循环
- `example09_for_to.klx` — For..to 和 for..downto 循环
- `example10_repeat.klx` — Repeat-until 循环
- `example11_case.klx` — Case 语句

### 3. 函数 (4 个示例) - `03_functions/` ✅

- `example13_functions.klx` — 函数与过程
- `example14_recursion.klx` — 递归函数
- `example15_lambda.klx` — 匿名过程（lambda）—— v3.1.0 修复了返回类型
- `example16_multireturn.klx` — 多返回值

### 4. 面向对象 (3 个示例) - `04_oop/` ✅

- `example17_class_fields.klx` — 类字段访问
- `example18_class_methods.klx` — 类方法（`self.Field`）
- `example19_inheritance.klx` — 类继承

v3.1.0 起 `var p: TPerson;` 也直接生成 `*TPerson`（KLX-C01 修复），不必再用 `:=` 推导。

### 5. 泛型 (1 个示例) - `05_generics/` ⚠️

- `example21_generic_class.klx` — 泛型栈类

**状态**: 编译通过，运行时有问题（KLX-G01，v3.2 修复）。

### 6. 高级类型 (5 个示例) - `06_advanced_types/` ✅

- `example20_enum.klx` — 枚举类型
- `example22_records.klx` — 记录类型
- `example23_arrays.klx` — 固定数组
- `example24_map.klx` — Map 类型（哈希表）
- `example25_string_ops.klx` — 字符串操作

### 7. 核心函数 (1 个示例) - `07_stdlib_core/` ✅

- `example29_basic_funcs.klx` — Max, Min, Abs 函数

### 8. 异常处理 (2 个示例) - `10_exceptions/` ✅

- `example27_try_except.klx` — Try-except 块
- `example28_finally.klx` — Try-finally 和 try-except-finally

### 9. 模块 (2 个示例) - `11_modules/` ⚠️

- `math_helper.klx` — 单元定义
- `example33_use_module.klx` — 使用 `uses` 导入单元

**状态**: 部分场景有问题（KLX-M01，v3.2 修复）。

### 10. 声明式 OOP (1 个示例) - 新增 v3.1.0 ✅

- `example40_declarative_oop.klx` — `var p := TPerson.Create` 模式 + 继承（KLX-C01 修复演示）

### 11. 特殊特性 (1 个示例) - 新增 v3.1.0 ✅

- `example41_attributes.klx` — `[Attribute]` 注解语法（`[Controller]`、`[Get]`、`[Inject]`、`[Entity]`）

---

## 🚀 如何运行示例

### 单个文件

```bash
cd examples/complete-tutorial/01_basics
kylix build example01_hello.klx
go run example01_hello.go
```

### 多文件（模块）

```bash
cd examples/complete-tutorial/11_modules
kylix build math_helper.klx example33_use_module.klx
go run main.go
```

### 批量测试所有示例

```bash
cd examples/complete-tutorial
for f in **/*.klx; do
  [[ "$f" == *"_test"* ]] && continue
  kylix build "$f" && go run $(basename "$f" .klx).go && echo "✓ $f"
done
```

---

## ⚠️ 已知限制 (v3.1.0)

多数 v3.0 已知问题在 v3.1 已修复。剩余的：

| ID | 特性 | 问题 | 目标 |
|----|------|------|------|
| **KLX-G01** | `example21_generic_class` | 编译通过但运行时异常 | v3.2 |
| **KLX-M01** | `example33_use_module` | 多文件 unit 编译边缘场景失败 | v3.2 |
| **LLVM Phase 2** | 接口 / 泛型 / 异常 | LLVM 后端尚未支持 | v3.2 (Phase 2/3) |

---

## 📖 特性覆盖清单

### ✅ 已通过示例验证 (32/34 工作)

**基础语法**:
- ✅ Hello World、变量声明、常量、类型推导、运算符、注释

**控制流**:
- ✅ if / while / for / repeat / case

**函数**:
- ✅ function / procedure / 递归 / 多返回值 / 匿名过程 / **lambda 带返回值（v3.1）**

**面向对象**:
- ✅ class / 字段（含 `var p: TClass`，v3.1） / `self.Field` / 继承

**类型系统**:
- ✅ record / array[1..N] / map[K]V / enum

**异常处理**:
- ✅ try / except / finally / raise

**模块系统**:
- ✅ unit 定义 / uses 导入

**字符串**:
- ✅ 拼接、IntToStr / Length、切片、比较、**插值 `${var}`（v3.1）**

**新特性 (v3.1)**:
- ✅ **注解语法 `[Name]` / `[Name(args)]`**
- ✅ **KylixBoot 框架运行时** —— `uses boot` 提供声明式 Web 应用
- ✅ **uses sysutil/jsonutil/datetime/regex/httpclient 在 program 中可用**（40+ stdlib 函数解锁）
- ✅ **match 语句完整 codegen**

### ⏳ 待补充示例

- ❌ 接口 interface（等 LLVM Phase 2）
- ❌ 泛型函数 / 多参数泛型 / 约束
- ❌ Variant 类型
- ❌ async / await
- ❌ kylix test 完整工作流示例
- ❌ kylix doc 示例
- ❌ --wasi 编译示例
- ❌ --backend=llvm + --llvm-opt 示例

---

## 💡 学习路径建议

### 🟢 初学者 (第1-2天)
1. **01_basics** — 全部 6 个示例
2. **02_control_flow** — if/while/for/repeat/case
3. **03_functions** — 函数基础

### 🟡 进阶 (第3-5天)
4. **04_oop** — 类与继承（3 个示例）
5. **06_advanced_types** — record/array/map/enum/string
6. **10_exceptions** — 异常处理

### 🔴 高级 (第6-10天)
7. **03_functions** — 多返回值 + lambda
8. **11_modules** — 模块化编程
9. **04_oop + 声明式 OOP** — `example40_declarative_oop.klx`
10. **特殊特性** — `example41_attributes.klx`（注解 + KylixBoot 预览）

---

## 🔧 常见问题

**Q: 示例运行失败怎么办？**

1. 确认 Kylix 版本：`kylix version` (需 v3.1.0+)
2. 检查 Go 环境：`go version` (需 Go 1.18+)
3. 确保在项目根目录运行（有 go.mod 的目录）
4. 清理生成文件：`rm *.go`

**Q: 类字段访问报错？**

v3.1.0 已修复 KLX-C01。`var p: TPerson;` 和 `var p := TPerson.Create;` 都正确生成 `*TPerson`。

**Q: 匿名函数返回值丢失？**

v3.1.0 已修复 KLX-C03，lambda 返回类型保留：

```pascal
var sq := function(x: Integer): Integer
begin
  result := x * x;  // 现在正确工作
end;
```

**Q: stdlib 函数找不到？**

v3.1.0 已修复 KLX-C05。在 `program` 中 `uses sysutil/jsonutil/datetime/regex/httpclient` 后可直接调用 40+ stdlib 函数。

**Q: 如何贡献新示例？**

1. 在对应分类目录创建 `.klx` 文件
2. 确保代码可运行：`kylix build && go run`
3. 在 README_CN.md 添加描述
4. 提交 PR

---

## 📚 更多资源

- **官方网站**: [kylix.top](https://kylix.top)
- **完整文档**: [README.md](../../README.md) · [README_CN.md](../../README_CN.md)
- **更新日志**: [CHANGELOG.md](../../CHANGELOG.md)
- **开发路线**: [ROADMAP.md](../../ROADMAP.md)
- **任务清单**: [TASKS.md](../../TASKS.md)
- **快速入门**: [docs/GETTING_STARTED_CN.md](../../docs/GETTING_STARTED_CN.md)

---

**最后更新**: 2026-06-23  
**示例状态**: 32/34 完全工作 (~94%)  
**版本**: v3.1.0
