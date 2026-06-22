# Kylix v3.0.0-alpha 完全教程

欢迎使用 Kylix 完全教程！本教程通过 **29 个经过测试的可运行示例**（其中 27 个完全工作），涵盖 Kylix v3.0.0-alpha 的所有已验证特性。

## 前置要求

- Kylix 编译器 (v3.0.0-alpha 或更高版本)
- Go 1.18+ (用于运行生成的代码)

---

## 📚 教程结构

本教程包含 **29 个示例**，分为 8 个类别：

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
- `example15_lambda.klx` — 匿名过程（lambda）⚠️
- `example16_multireturn.klx` — 多返回值

**注**: `example15_lambda` 仅支持匿名过程（无返回值），带返回值的匿名函数在 v3.0.0-alpha 有已知 bug。

### 4. 面向对象 (3 个示例) - `04_oop/` ✅

- `example17_class_fields.klx` — 类字段访问（使用 `:=` 类型推导）
- `example18_class_methods.klx` — 类方法（使用 `self.Field`）
- `example19_inheritance.klx` — 类继承

**重要**: 类实例变量必须用 `:=` 类型推导声明（`var p := TPerson.Create`），否则生成的 Go 类型为 `interface{}`，字段不可访问。

### 5. 泛型 (1 个示例) - `05_generics/` ⚠️

- `example21_generic_class.klx` — 泛型栈类

**状态**: 编译通过，运行时有问题。

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

**状态**: 编译通过，运行时有问题。

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

## ⚠️ 已知限制 (v3.0.0-alpha)

以下特性在当前版本有已知问题，已记录在 ROADMAP.md，将在 v3.1/v3.2 修复：

| 特性 | 问题 | 影响 |
|------|------|------|
| **字符串插值** | `"Hello ${name}"` 不展开，输出字面字符串 | 中等 |
| **匿名函数返回值** | `function(x): T` 返回类型丢失 | 中等 |
| **类变量声明** | `var p: TClass` 生成 `interface{}`，字段不可访问 | 高 |
| **match 语句** | 生成无效 Go 代码 | 高 |
| **uses 在 program 中** | stdlib 函数（strutil/mathutil/sysutil/jsonutil/datetime/regex）不可直接调用 | 高 |

**解决方法**:
- 类变量：使用 `var p := TClass.Create` (`:=` 推导) 而不是 `var p: TClass`
- 匿名函数：只使用匿名过程（无返回值），或用命名函数
- stdlib：等待 v3.2 修复，或在 `unit` 文件中使用

---

## 📖 特性覆盖清单

### ✅ 已通过示例验证 (27/29 工作)

**基础语法**:
- ✅ Hello World
- ✅ 变量声明 (var)
- ✅ 常量 (const)
- ✅ 类型推导 (var x := value)
- ✅ 所有运算符 (+, -, *, /, div, mod, and, or, not, =, <>, <, >, <=, >=)
- ✅ 注释 (//, (* *))

**控制流**:
- ✅ if/then/else
- ✅ while/do
- ✅ for..to/downto
- ✅ repeat..until
- ✅ case/of

**函数**:
- ✅ function (带返回值)
- ✅ procedure (无返回值)
- ✅ 递归
- ✅ 多返回值 (a, b) := Func()
- ✅ 匿名过程 (无返回值 lambda)

**面向对象**:
- ✅ class 定义
- ✅ 字段访问 (`:=` 推导)
- ✅ 方法 (`self.Field`)
- ✅ 继承 (class(TParent))

**类型系统**:
- ✅ record
- ✅ array[1..N]
- ✅ map[K]V
- ✅ enum (TColor = (Red, Green, Blue))

**异常处理**:
- ✅ try/except
- ✅ try/finally
- ✅ raise

**模块系统**:
- ✅ unit 定义
- ✅ uses 导入

**字符串**:
- ✅ 拼接 (+)
- ✅ IntToStr / Length
- ✅ 切片 s[from:to]
- ✅ 比较 (=, <>)

### ⏳ 待补充示例（等 bug 修复）

**面向对象进阶**:
- ❌ 接口 interface (等编译器修复)
- ❌ 属性 property
- ❌ 构造器/析构器
- ❌ public/private/protected

**泛型进阶**:
- ❌ 泛型函数
- ❌ 泛型约束 T: IComparable
- ❌ 多参数泛型

**高级类型**:
- ❌ Variant
- ❌ 字符串插值 `${expr}` (已知 bug)

**函数式**:
- ❌ 带返回值的匿名函数 (已知 bug)

**模式匹配**:
- ❌ match 表达式 (已知 bug)

**标准库** (等 `uses` 在 program 修复):
- ❌ strutil（Reverse/ToUpper/StartsWith...）
- ❌ mathutil（Abs/Max/Min/Pow/IsPrime...）
- ❌ arrayutil（Sum/MinValue/MaxValue...）
- ❌ sysutil（ReadFile/WriteFile/FileExists...）
- ❌ jsonutil（JsonDecodeMap/JsonGetString...）
- ❌ datetime（Now/MakeDate/FormatPattern...）
- ❌ regex（IsEmail/IsURL/IsNumeric...）
- ❌ httpclient（HttpGet/HttpPost/THttpClient...）
- ❌ wasi（WriteLn/ReadLine/GetEnv/Args...）

**异步编程**:
- ❌ async function
- ❌ await

**工具链**:
- ❌ kylix test 示例
- ❌ kylix doc 示例
- ❌ --wasi 编译示例
- ❌ --backend=llvm 示例

---

## 💡 学习路径建议

### 🟢 初学者 (第1-2天)
1. **01_basics** — 全部 6 个示例
2. **02_control_flow** — if/while/for/repeat/case
3. **03_functions** — 函数基础（跳过 lambda）

### 🟡 进阶 (第3-5天)
4. **04_oop** — 类与继承（3 个示例）
5. **06_advanced_types** — record/array/map/enum/string
6. **10_exceptions** — 异常处理

### 🔴 高级 (第6-10天)
7. **03_functions** — 多返回值 + 递归
8. **11_modules** — 模块化编程
9. **05_generics** — 泛型编程（需等修复）
10. **stdlib** — 标准库应用（需等 v3.2 修复）

---

## 🔧 常见问题

**Q: 示例运行失败怎么办？**

1. 确认 Kylix 版本：`kylix version` (需 v3.0.0-alpha+)
2. 检查 Go 环境：`go version` (需 Go 1.18+)
3. 确保在项目根目录运行（有 go.mod 的目录）
4. 清理生成文件：`rm *.go`

**Q: stdlib 函数找不到？**

stdlib 函数（strutil/mathutil 等）在 `program` 文件中 `uses X` 后无法直接调用，这是 v3.0.0-alpha 已知 bug。等待 v3.2 修复，或在 `unit` 文件中使用。

**Q: 类字段访问报错 "type interface{} has no field"？**

使用 `:=` 类型推导声明类变量：
```pascal
// 错误
var p: TPerson;
p := TPerson.Create;  // p 类型为 interface{}

// 正确
var p := TPerson.Create;  // p 类型为 *TPerson
```

**Q: 匿名函数返回值丢失？**

v3.0.0-alpha 的匿名函数返回类型不生成，只能使用匿名过程（无返回值）：
```pascal
// 可以
var greet := procedure(name: String)
begin
  WriteLn('Hello ' + name);
end;

// 不行（返回类型丢失）
var sq := function(x: Integer): Integer
begin
  result := x * x;  // result 未定义
end;
```

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

**最后更新**: 2026-06-22  
**示例状态**: 27/29 完全工作 (93.1%)  
**特性覆盖**: 35/74 特性点 (47.3%)
