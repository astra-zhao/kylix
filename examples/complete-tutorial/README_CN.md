# Kylix 完全教程（中文）

欢迎使用 Kylix 完全教程！本教程涵盖 Kylix **v0.6.9** 的全部可用特性，全部经过测试、可直接运行——**20 个章节共 50 个编号示例**，外加 `math_helper.klx` 单元配套文件与 `test.klx` 冒烟文件。

测试状态（v0.6.9）：

- ✅ Go 后端 sweep：**51/51**（`examples/complete-tutorial/test_all.sh`）
- ✅ LLVM 后端 sweep：**51/51**（`examples/complete-tutorial/test_all_llvm.sh`，原生二进制，运行时无 Go）
- ✅ 自举（无 Go）sweep：**50 PASS + 1 SKIP**（`scripts/test_bootstrap_all.sh`；example33 多文件为 host 端验证）
- ✅ 编译器本身用 Kylix 编写，达到 IR 不动点（gen1 ≡ gen2 ≡ gen3）

## Kylix 是什么？

Kylix 是现代 Pascal 编译器：默认转译为可读的 Go 代码（`go build` 编译运行）；加 `--backend=llvm` 则直接产出 LLVM IR 并链接为原生二进制——运行时完全不需要 Go 工具链。

## 前置要求

- Kylix 编译器（v0.6.9 或更高版本）
- **Go 1.18+**（Go 后端）或 **LLVM**（`llc`/`clang`，原生后端）二者其一。可用 `kylix doctor` 预检 LLVM 环境。
- 原生后端按需链接系统库（openssl、sqlite3、curl）。

---

## 📚 教程结构

### 1. 基础语法 (6 个示例) - `01_basics/`

- `example01_hello.klx` — Hello World
- `example02_variables.klx` — 变量声明与类型
- `example03_constants.klx` — 常量
- `example04_type_inference.klx` — 类型推导 `:=`
- `example05_operators.klx` — 算术、比较、逻辑运算符
- `example06_comments.klx` — 单行注释

### 2. 控制流 (5 个示例) - `02_control_flow/`

- `example07_if_else.klx` — If-then-else 条件语句
- `example08_while.klx` — While 循环
- `example09_for_to.klx` — For..to 和 for..downto 循环
- `example10_repeat.klx` — Repeat-until 循环
- `example11_case.klx` — Case 语句

### 3. 函数 (4 个示例) - `03_functions/`

- `example13_functions.klx` — 函数与过程
- `example14_recursion.klx` — 递归函数
- `example15_lambda.klx` — Lambda 与闭包
- `example16_multireturn.klx` — 多返回值

### 4. 面向对象 (4 个示例) - `04_oop/`

- `example17_class_fields.klx` — 类字段访问
- `example18_class_methods.klx` — 类方法（`self.Field`）
- `example19_inheritance.klx` — 类继承
- `example40_declarative_oop.klx` — `var p := TPerson.Create` 声明式模式 + 继承

### 5. 泛型 (1 个示例) - `05_generics/`

- `example21_generic_class.klx` — 泛型栈类

### 6. 高级类型 (5 个示例) - `06_advanced_types/`

- `example20_enum.klx` — 枚举类型
- `example22_records.klx` — 记录类型
- `example23_arrays.klx` — 固定数组与动态数组
- `example24_map.klx` — Map 类型（哈希表）
- `example25_string_ops.klx` — 字符串操作

### 7. 核心函数 (1 个示例) - `07_stdlib_core/`

- `example29_basic_funcs.klx` — Max, Min, Abs 函数

### 8. stdlib 工具库 (4 个示例) - `08_stdlib_utils/`

- `example36_sysutil.klx` — 文件系统 / 环境工具
- `example37_jsonutil.klx` — JSON 编解码
- `example38_datetime.klx` — 日期时间
- `example39_regex.klx` — 正则表达式

### 9. 异常处理 (2 个示例) - `10_exceptions/`

- `example27_try_except.klx` — Try-except 块
- `example28_finally.klx` — Try-finally 和 try-except-finally

### 10. 模块 (1 个示例 + 单元) - `11_modules/`

- `math_helper.klx` — 单元定义（`unit`/`interface`/`implementation`）
- `example33_use_module.klx` — 使用 `uses` 导入单元

### 11. 特殊特性 (7 个示例) - `12_special_features/`

- `example41_attributes.klx` — `[Attribute]` 注解语法（`[Controller]`、`[Get]`、`[Inject]`、`[Entity]`）
- `example42_kylixboot_autowire.klx` — KylixBoot `[Controller]` + `[Get]` 自动路由注册
- `example43_kylixboot_di.klx` — KylixBoot `[Service]` + `[Inject]` DI 自动装配
- `example44_kylixboot_proc_handler.klx` — 过程式路由处理器
- `example45_validation_annotations.klx` — `[Required]`/`[Email]`/`[Min]`/`[MinLen]` 字段校验
- `example46_security_annotations.klx` — `[Authenticated]`/`[Role]` 路由安全守卫
- `example47_orm_annotations.klx` — ORM 注解（`[Entity]`/`[Repository]`/`[Query]`）

### 12. stdlib Phase 6 (1 个示例) - `13_stdlib_phase6/`

- `example48_phase6_net_crypto_encoding.klx` — SHA-256、Base64、BCrypt、CSV、HMAC、MD5

### 13. 请求体绑定 (1 个示例) - `14_body_binding/`

- `example49_body_binding.klx` — `[Body(TEntity)]` JSON 请求体绑定 + `Validate()`/`IsValid()` 校验

### 14. JWT 认证 (1 个示例) - `15_jwt/`

- `example50_jwt_auth.klx` — `JwtSign`/`JwtVerify` + `BootRegisterJwtAuth` 接入 `[Authenticated]` 路由守卫

### 15. OpenAPI / Swagger (1 个示例) - `16_openapi/`

- `example51_openapi.klx` — `[Controller]`/`[Get]`/`[Post]`/`[Body]`/`[Authenticated]`/`[Role]` → `kylix doc --openapi` 生成 OpenAPI 3.1 YAML

### 16. 数据库 (1 个示例) - `17_database/`

- `example52_database.klx` — SQLite 内存库：`DbOpenSQLite`/`DbExec`/`DbQueryScalar`/`DbQueryRows`、参数化查询

### 17. 缓存 (1 个示例) - `18_cache/`

- `example53_cache.klx` — 线程安全 LRU 缓存：`NewCache`/`Put`/`GetString`/`Has`/`Delete`/`Size`/`Clear`

### 18. HTTP 客户端 (1 个示例) - `19_http/`

- `example54_http.klx` — `THttpClient` GET/POST/PUT/DELETE、一次性助手函数、`THttpResponse`（status+body）

### 19. WebSocket (1 个示例) - `20_websocket/`

- `example55_websocket.klx` — RFC 6455 WebSocket 客户端/服务端（`WsDial`/`WsAccept`/`WsSend`/`WsRecv`/`WsClose`），纯 stdlib

### 20. Variant (2 个示例) - `21_variant/`

- `example56_variant.klx` — Variant 标量与数组（带类型标签的运行时值）
- `example57_variant_map.klx` — `map[String]Variant` 类型标签 map（`row['col']` 式访问）

---

## 🚀 如何运行示例

### 单个文件

```bash
cd examples/complete-tutorial/01_basics
kylix build example01_hello.klx
go run example01_hello.go
```

或用 LLVM 原生后端（运行时无 Go）：

```bash
kylix run example01_hello.klx                    # 自动探测后端
kylix build --backend=llvm example01_hello.klx   # 强制原生二进制
```

### 多文件（模块）

```bash
cd examples/complete-tutorial/11_modules
kylix build math_helper.klx example33_use_module.klx
go run main.go
```

### 批量运行某类示例

```bash
cd examples/complete-tutorial/02_control_flow
for f in example*.klx; do
  echo "=== $f ==="
  kylix build "$f"
  go run "${f%.klx}.go"
  echo ""
done
```

### 全量回归

```bash
# Go 后端（51/51）
KYLIX=/path/to/kylix bash examples/complete-tutorial/test_all.sh

# LLVM 后端（51/51，原生二进制）
bash examples/complete-tutorial/test_all_llvm.sh
```

---

## 📖 语言特性速查

### 变量与类型

```pascal
var x: Integer;           // 整数
var name: String;         // 字符串
var pi: Real;             // 浮点
var active: Boolean;      // 布尔

var count := 42;          // 类型推导
```

### 控制流

```pascal
if x > 5 then
  WriteLn('Greater')
else
  WriteLn('Not greater');

while i < 10 do
begin
  i := i + 1;
end;

for i := 1 to 10 do
  WriteLn(i);

repeat
  WriteLn(i);
  i := i - 1;
until i <= 0;

case day of
  1: WriteLn('Monday');
  6, 7: WriteLn('Weekend');
end;
```

### 函数

```pascal
function Add(a: Integer; b: Integer): Integer;
begin
  result := a + b;
end;

// 多返回值
function DivMod(a: Integer; b: Integer): (Integer, Integer);
begin
  result := (a div b, a mod b);
end;

var q, r: Integer;
(q, r) := DivMod(17, 5);
```

### 数组与集合

```pascal
var numbers: array[0..9] of Integer;
numbers[0] := 42;

var scores: map[String]Integer;
scores['Alice'] := 95;
WriteLn(scores['Alice']);
```

### 记录

```pascal
type
  TPoint = record
    X: Real;
    Y: Real;
  end;

var point: TPoint;
point.X := 10.5;
point.Y := 20.3;
```

### 注解

```pascal
[Controller('/api/users')]
type
  TUserController = class
    [Inject]
    UserRepo: TUserRepository;

    [Get('/')]
    function ListUsers(req: TRequest): TResponse;
    begin
      result := req.JSON(UserRepo.FindAll());
    end;
  end;
```

### 异常处理

```pascal
try
begin
  result := SafeDivide(10, 0);
end
except
begin
  WriteLn('Error occurred');
end
finally
begin
  WriteLn('Cleanup');
end
end;
```

### 模块（单元）

```pascal
// math_helper.klx
unit MathHelper;

interface
function Square(x: Integer): Integer;

implementation
function Square(x: Integer): Integer;
begin
  result := x * x;
end;
end.

// main.klx
program Main;
uses MathHelper;
begin
  WriteLn(Square(5));
end.
```

---

## ⚠️ 已知限制 (v0.6.9)

- **Windows**：`net`（Winsock）与 `regex`（pcre2）为 stub——真实现需 Windows 真机验证，已列入规划。
- **Variant 算术**（`v + 1`）仅 LLVM 后端可用（Go 的 `interface{}` 不支持运算符）；比较与打印双端一致。
- **example33**（多文件 unit 构建）为 host 端验证；无 Go 自举 sweep 报 SKIP。

---

## 💡 学习路径建议

### 🟢 初学者（第 1-2 天）
1. **01_basics** — 全部 6 个示例
2. **02_control_flow** — if/while/for/repeat/case
3. **03_functions** — 函数基础

### 🟡 进阶（第 3-5 天）
4. **04_oop** — 类与继承
5. **06_advanced_types** — enum/record/array/map/string
6. **10_exceptions** — 异常处理
7. **08_stdlib_utils** — sysutil/jsonutil/datetime/regex

### 🔴 高级（第 6-10 天）
8. **03_functions** — 多返回值 + lambda
9. **11_modules** — 模块化编程
10. **12_special_features** — 注解 + KylixBoot 框架
11. **13-20 章节** — stdlib 实战（crypto/JWT/OpenAPI/数据库/缓存/HTTP/WebSocket）
12. **21_variant** — Variant 动态类型

---

## 🔧 常见问题

**Q: 示例运行失败怎么办？**

1. 确认 Kylix 版本：`kylix version`（需 v0.6.9+）
2. 检查 Go 环境：`go version`（需 Go 1.18+）；或改用 `kylix run`（自动回退 LLVM）
3. 清理生成文件：`rm *.go`

**Q: 如何不装 Go 运行示例？**

安装 LLVM（`llc`/`clang`）后执行 `kylix doctor` 预检，然后 `kylix run example01_hello.klx`——自动走 LLVM 原生后端。

**Q: 如何贡献新示例？**

1. 在对应章节目录创建 `exampleNN_*.klx` 文件
2. 确保可运行：`kylix build && go run`
3. 在 README.md / README_CN.md 补充描述
4. 提交 PR

---

## 📚 更多资源

- **官方网站**: [kylix.top](https://kylix.top)
- **完整文档**: [README.md](../../README.md) · [README_CN.md](../../README_CN.md)
- **小白入门教程**: [docs/TUTORIAL_FOR_BEGINNERS_CN.md](../../docs/TUTORIAL_FOR_BEGINNERS_CN.md)
- **更新日志**: [CHANGELOG.md](../../CHANGELOG.md)
- **开发路线**: [ROADMAP.md](../../ROADMAP.md)
- **快速入门**: [docs/GETTING_STARTED_CN.md](../../docs/GETTING_STARTED_CN.md)

---

## 📊 分类汇总

| 分类 | 示例数 | 状态 |
|------|--------|------|
| 基础语法 | 6 | ✅ 全部通过 |
| 控制流 | 5 | ✅ 全部通过 |
| 函数（含 lambda） | 4 | ✅ 全部通过 |
| 面向对象（含声明式） | 4 | ✅ 全部通过 |
| 泛型 | 1 | ✅ 通过 |
| 高级类型 | 5 | ✅ 全部通过 |
| 核心函数 | 1 | ✅ 通过 |
| stdlib 工具库 | 4 | ✅ 全部通过 |
| 异常处理 | 2 | ✅ 全部通过 |
| 模块（unit/uses） | 1 + 单元 | ✅ 通过 |
| 注解 / KylixBoot / 校验 / 安全 / ORM | 7 | ✅ 全部通过 |
| stdlib Phase 6（crypto / encoding） | 1 | ✅ 通过 |
| 请求体绑定 | 1 | ✅ 通过 |
| JWT 认证 | 1 | ✅ 通过 |
| OpenAPI / Swagger | 1 | ✅ 通过 |
| 数据库（SQLite） | 1 | ✅ 通过 |
| 缓存（LRU） | 1 | ✅ 通过 |
| HTTP 客户端 | 1 | ✅ 通过 |
| WebSocket | 1 | ✅ 通过 |
| Variant（标量/数组 + map） | 2 | ✅ 全部通过 |
| **合计** | **50 + 单元 + 冒烟** | **Go 51/51 · LLVM 51/51 · 自举 50+1 SKIP** |

---

**最后更新**: 2026-09-04  
**版本**: v0.6.9

用 Kylix 愉快编码！🚀
