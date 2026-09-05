# Kylix 上手指南 —— 写给第一次接触 Kylix 的你

> 适用版本：v0.6.9（2026-09-04） · 无需任何 Pascal 或编译器知识，会用命令行即可跟着走完全文。**本文所有示例均经过实测可运行。**

---

## 目录

1. [Kylix 是什么？](#1-kylix-是什么)
2. [安装：两条路任选一条](#2-安装两条路任选一条)
3. [第一个程序：Hello, World!](#3-第一个程序hello-world)
4. [变量与类型：让编译器帮你算](#4-变量与类型让编译器帮你算)
5. [条件与循环：程序会做决定了](#5-条件与循环程序会做决定了)
6. [函数：把活儿打包](#6-函数把活儿打包)
7. [集合类型：数组、Map、Record、枚举](#7-集合类型数组maprecord枚举)
8. [面向对象：类与继承](#8-面向对象类与继承)
9. [异常处理：出错了别崩](#9-异常处理出错了别崩)
10. [标准库：站在巨人的肩膀上](#10-标准库站在巨人的肩膀上)
11. [写一个 Web 服务：KylixBoot](#11-写一个-web-服务kylixboot)
12. [进阶：双后端与自举（它是怎么炼成的）](#12-进阶双后端与自举它是怎么炼成的)
13. [命令速查表](#13-命令速查表)
14. [常见问题](#14-常见问题)
15. [下一步去哪](#15-下一步去哪)

---

## 1. Kylix 是什么？

**一句话**：Kylix 是一门**现代 Pascal 语言**，写出来的代码可以编译成**真正的原生二进制**（不依赖虚拟机、不依赖解释器），也能转译成 Go 代码借助 Go 生态。

你写 Kylix（`.klx` 文件），编译器帮你产出可执行程序：

```
 你写的代码              Kylix 编译器               最终产物
┌──────────────┐      ┌───────────────┐      ┌─────────────────┐
│  hello.klx   │      │     kylix     │      │  hello (二进制)  │
│              │ ───► │   编译器       │ ───► │  ./hello 直接跑  │
│ WriteLn('Hi')│      │  (Go / LLVM)  │      │  不需要任何依赖   │
└──────────────┘      └───────────────┘      └─────────────────┘
```

它有三个最打动人的特点：

| 特点 | 意思是 |
|------|--------|
| **简单** | Pascal 语系，`begin...end` 结构清晰，初学者一晚上能读懂基本语法 |
| **快** | 编译快（毫秒级），运行快（原生机器码） |
| **自举** | 编译器自己能编译自己——v0.6.9 起编译器本体可以**完全不依赖 Go** 构建 |

> 🤔 **为什么叫 Kylix？** Kylix 是古希腊的一种双耳酒杯——优雅、实用、经典，正如 Pascal 语言的气质。

---

## 2. 安装：两条路任选一条

Kylix 的编译器本体用 Go 写（这部分叫 **host 编译器**），但它产出的程序走两条后端：

```
                      ┌── 后端 A：Go 后端 ──► 生成 Go 代码 → go build → 二进制
 hello.klx ─► kylix ──┤                      （需要本机装 Go）
                      │
                      └── 后端 B：LLVM 后端 ─► 生成 LLVM IR → llc/clang → 二进制
                                             （需要本机装 LLVM，不需要 Go）
```

### 从源码构建编译器（5 分钟）

前置：装好 [Go 1.21+](https://go.dev/dl/) 和 [LLVM](https://releases.llvm.org/)（macOS: `brew install llvm`；Ubuntu: `apt install llvm clang`）。

```bash
git clone https://github.com/astra-zhao/kylix.git
cd kylix
go build -o /usr/local/bin/kylix ./cmd/kylix/
kylix --version        # 应显示 kylix version 0.6.9
kylix doctor           # 体检：检查 Go/LLVM 环境是否就绪
```

> 💡 `kylix doctor` 是你的环境体检医生——Go、LLVM、链接库（openssl/sqlite3/curl）有没有装、版本对不对，一条命令全知道。

---

## 3. 第一个程序：Hello, World!

新建 `hello.klx`：

```pascal
program Hello;

begin
  WriteLn('Hello, World!');
end.
```

注意三个 Pascal 传统：

- `program Hello;` 声明程序名（分号结尾）
- 代码体夹在 `begin ... end.` 之间（**end 后面是句号**，表示程序结束）
- 每条语句以分号 `;` 结尾

编译并运行：

```bash
kylix run hello.klx          # 一条命令：自动选后端、编译、运行
# 输出：Hello, World!
```

`kylix run` 很聪明（叫 **auto 后端**）：本机有 Go 就走 Go 后端，没有 Go 就自动切 LLVM 后端——两种情况你都不用操心。

如果想分步：

```bash
kylix build hello.klx        # 编译 → 产出二进制 hello
./hello                      # 自己运行
```

---

## 4. 变量与类型：让编译器帮你算

Kylix 支持两种声明方式——**显式类型**和**类型推断**：

```pascal
program Variables;

const
  MAX = 100;                     // 常量区：编译期确定，不可改

begin
  var age: Integer;              // 显式：明确告诉编译器
  age := 25;

  var name := 'Astra';           // 推断：编译器自己看出是 String
  var pi := 3.14;                // 推断为 Float
  var ok := true;                // 推断为 Boolean

  WriteLn(name + ' is ' + IntToStr(age) + ' years old');
  WriteLn('age = ${age}');       // 字符串插值：${变量} 会被替换
end.
```

常用类型速览：

| 类型 | 装什么 | 例子 |
|------|--------|------|
| `Integer` | 整数 | `42` |
| `Float` | 小数 | `3.14` |
| `String` | 文本 | `'hello'` |
| `Boolean` | 真/假 | `true` |
| `Variant` | 什么都能装（动态类型） | `var v := 1; v := 'one';` |

> 💡 字符串插值的完整形态是 `'${表达式}'`——比手动 `+ IntToStr(...)` 拼接清爽得多。

---

## 5. 条件与循环：程序会做决定了

```pascal
program ControlFlow;

begin
  var score := 87;

  // if / else if / else
  if score >= 90 then
    WriteLn('优秀')
  else if score >= 60 then
    WriteLn('及格')
  else
    WriteLn('不及格');

  // while 循环
  var i := 0;
  while i < 3 do
  begin
    WriteLn('while: ' + IntToStr(i));
    i := i + 1;
  end;

  // for 循环：正着数、倒着数
  for j := 1 to 5 do
    WriteLn('for: ' + IntToStr(j));

  for k := 5 downto 1 do
    WriteLn('倒数: ' + IntToStr(k));

  // repeat...until（先执行后判断，至少跑一次）
  var n := 0;
  repeat
    n := n + 1;
  until n >= 3;

  // case 多路分支（多个值用逗号并列）
  var day := 6;
  case day of
    1, 2, 3, 4, 5: WriteLn('工作日');
    6, 7:          WriteLn('周末');
  end;
end.
```

四种循环怎么选：

```
 需要知道次数？          需要先判断条件？         至少执行一次？
     │                     │                     │
     ▼                     ▼                     ▼
  for ... to            while ... do          repeat ... until
 （1 to 5, downto）    （可能一次都不跑）      （先跑再说）
```

---

## 6. 函数：把活儿打包

```pascal
program Functions;

// 函数：有返回值
function Add(a: Integer; b: Integer): Integer;
begin
  result := a + b;      // result 是内置的返回值变量
end;

// 多返回值：一个函数同时算出商和余数！
function DivMod(a: Integer; b: Integer): (Integer, Integer);
begin
  result := (a div b, a mod b);
end;

// 过程：没有返回值
procedure Greet(name: String);
begin
  WriteLn('Hi, ' + name + '!');
end;

var
  q, r: Integer;        // 多返回值接收变量（放程序级声明区）

begin
  Greet('Kylix');
  WriteLn(Add(1, 2));            // 3

  (q, r) := DivMod(17, 5);       // 解构接收：q=3, r=2
  WriteLn(IntToStr(q) + ' 余 ' + IntToStr(r));
end.
```

匿名函数（lambda）——赋给变量、当函数用：

```pascal
program LambdaDemo;

begin
  var dbl := procedure(x: Integer)
  begin
    WriteLn(x * 2);
  end;

  dbl(21);          // 42
end.
```

> 💡 **多返回值**是 Kylix 相对传统 Pascal 的重要增强——不用再为"返回两个值"定义 Record 或用 var 参数了。

---

## 7. 集合类型：数组、Map、Record、枚举

```pascal
program Collections;

type
  // Record：把相关的数据打包（类似 C 的 struct）
  TPoint = record
    X: Integer;
    Y: Integer;
  end;

  // 枚举：一组命名常量
  TColor = (Red, Green, Blue);

var
  fixed: array[0..4] of Integer;   // 静态数组：定长（放声明区）
  p: TPoint;
  c: TColor;

begin
  // 动态数组：自动扩容（推断式声明）
  var nums := [1, 2, 3];
  WriteLn(Length(nums));            // 3（Length 取长度）

  // 遍历数组（forEach）
  for n in nums do
    WriteLn('n = ', n);

  // Map：键值对
  var m: map[String]Integer;
  m['apple'] := 5;
  m['pear'] := 3;
  WriteLn(m['apple']);              // 5

  // Record：按字段访问
  p.X := 3;
  p.Y := 4;
  WriteLn(p.X + p.Y);               // 7

  // 枚举：按名字比较
  c := Green;
  if c = Green then
    WriteLn('是绿色');
end.
```

它们的关系一张图：

```
       一个变量装一个值        一个变量装一堆值
      ┌──────────────┐    ┌────────────────────────────┐
      │ Integer      │    │ 数组：一排同类型（按下标取）    │
      │ Float        │    │ Map：键值对（按名字取）        │
      │ String       │    │ Record：不同类型打包（按字段取）│
      │ Boolean      │    │ 枚举：有限个命名常量            │
      │ 枚举/Record  │    │ Variant：动态装箱（啥都行）    │
      └──────────────┘    └────────────────────────────┘
```

---

## 8. 面向对象：类与继承

```pascal
program OOP;

type
  TAnimal = class
  public
    Name: String;
    constructor Create(n: String);
    procedure Speak; virtual;      // virtual = 允许子类重写
  end;

  TDog = class(TAnimal)            // 继承 TAnimal
  public
    procedure Speak; override;     // 重写父类方法
  end;

constructor TAnimal.Create(n: String);
begin
  self.Name := n;
end;

procedure TAnimal.Speak;
begin
  WriteLn(self.Name + ' makes a sound');
end;

procedure TDog.Speak;
begin
  WriteLn(self.Name + ' says: Woof!');
end;

begin
  var a := TAnimal.Create('Generic');
  var d := TDog.Create('Rex');
  a.Speak;              // Generic makes a sound
  d.Speak;              // Rex says: Woof!

  // 类型判断与转换
  if a is TAnimal then WriteLn('a 是 TAnimal');
end.
```

类相关的完整能力：`is`/`as` 类型判断与转换、接口（interface）、泛型类 `TStack<Integer>`、属性（property）、`inherited` 调父类。全部教程见 `examples/complete-tutorial/04_oop/`。

---

## 9. 异常处理：出错了别崩

```pascal
program Exceptions;

begin
  try
    var x := 10 div 0;         // 会抛异常
    WriteLn(IntToStr(x));
  except
    on E: Exception do
      WriteLn('捕获异常: ' + E.Message);
  end;
end.
```

`try / except / finally` 三件套与主流语言语义一致：`except` 接住错误，`finally` 无论成败都执行（善后清理）。

---

## 10. 标准库：站在巨人的肩膀上

Kylix 自带一套覆盖常见场景的标准库（`uses` 引入即用）：

```
┌─────────────────────────── Kylix stdlib ───────────────────────────┐
│  sysutil    文件/目录：ReadFile, WriteFile, DirExists, CopyFile...  │
│  jsonutil   JSON 编解码：JsonEncode/Decode, JsonGetString...       │
│  datetime   时间：Now, Today, FormatDate, AddDays, Year/Month/Day  │
│  encoding   编码：Base64/Base64URL, Hex, CSV                       │
│  crypto     安全：SHA256, AES-256-CBC, HMAC, BCrypt                │
│  jwt        令牌：JwtSign / JwtVerify（HS256，双端实现）            │
│  cache      缓存：NewCache, Put/Get（带 TTL 过期）                  │
│  db         数据库：DbOpen, DbExec, DbQueryRows（SQLite）          │
│  httpclient HTTP 客户端：HttpGet/Post/HttpPut + JSON 直解           │
│  net        网络：TCP/UDP/DNS（LLVM 端 WebSocket 完整实现）          │
│  regex      正则：RegexMatch / RegexReplace                        │
└─────────────────────────────────────────────────────────────────────┘
```

例子——读文件 + 解析 JSON：

```pascal
program StdlibDemo;

uses sysutil, jsonutil;

begin
  // 读文件
  var content := sysutil.ReadFile('config.json');

  // JSON → Variant map，链式取值
  var cfg := JsonDecodeMap(content);
  WriteLn('name = ' + JsonGetString(cfg, 'name'));
end.
```

> 💡 全部 stdlib 函数清单见 `stdlib/klx/`（给 LSP 补全用的声明文件，也是最好的 API 速查表）。

---

## 11. 写一个 Web 服务：KylixBoot

Kylix 内置注解驱动的 Web 框架 **KylixBoot**（类似 Spring Boot 的体验）——写好类和方法，打上注解，路由/校验/DI/文档全自动：

```pascal
program MyApp;

uses boot;

[Entity('users')]
type
  TCreateUser = class
    [Required]
    [Email]
    Email: String;
    [Required]
    [MinLen(8)]
    Password: String;
  end;

[Controller('/api')]
type
  TUserController = class
    [Post('/users')]
    [Body(TCreateUser)]            // 请求体自动绑定 + 自动校验
    function CreateUser(req: TRequest): TResponse;
    begin
      result := BootText(201, 'user created');
    end;

    [Get('/me')]
    function Me(req: TRequest): TResponse;
    begin
      result := BootText(200, 'hello, kylix');
    end;
  end;

begin
  WriteLn('KylixBoot app');
end.
```

编译运行后即可 HTTP 访问（POST `/api/users`、GET `/api/me`）。

KylixBoot 注解全家福：

```
 路由与文档                数据校验                  依赖与安全
┌─────────────────┐    ┌──────────────────┐    ┌──────────────────┐
│ [Controller]    │    │ [Required]       │    │ [Service]        │
│ [Get] [Post]    │    │ [Email]          │    │ [Component]      │
│ [Put] [Delete]  │    │ [Min] [Max]      │    │ [Inject]         │
│ [Body(TEntity)] │    │ [MinLen][MaxLen] │    │ [Authenticated]  │
│                 │    │                  │    │ [Role('admin')]  │
│ 自动生成 OpenAPI │    │ 自动 400 拦截     │    │ 自动 JWT 校验     │
└─────────────────┘    └──────────────────┘    └──────────────────┘
 [Entity]/[Repository] 注解还能自动生成数据库 CRUD（ORM）
```

`kylix doc --openapi` 一条命令从注解生成 OpenAPI 3.1 文档。

---

## 12. 进阶：双后端与自举（它是怎么炼成的）

这一节不需要你会——但看懂了你会明白 Kylix 的独特之处。

### 双后端

同一份 `.klx` 源码，两条编译路径，输出**行为一致**的程序：

```
                        ┌────────────────────────────────────┐
                        │            kylix 编译器             │
                        │  (词法 → 语法 → AST → 代码生成)      │
                        └──────────────┬─────────────────────┘
                    Go 后端 │                        │ LLVM 后端
                            ▼                        ▼
                 ┌────────────────────┐    ┌────────────────────┐
                 │  生成 Go 源码 .go   │    │  生成 LLVM IR .ll  │
                 │  → go build        │    │  → llc → clang     │
                 │  → 原生二进制        │    │  → 原生二进制        │
                 └────────────────────┘    └────────────────────┘
                 依赖：Go 工具链               依赖：LLVM 工具链
                 优势：编译极快(毫秒级)         优势：无 Go 也能跑
                                               (KylixRT / 交叉编译 --target)
```

51 个官方教程全部通过**两种后端**编译运行且输出一致——这是 Kylix 每次发版的回归基线。

### 自举（Bootstrap）：编译器自己造自己

v0.6.9 的里程碑：**Kylix 编译器本体就是用 Kylix 写的**（`src/*.klx`，9 个文件）。它能编译出自己，而且迭代三代输出完全一致（数学上的"不动点"）：

```
 第 1 代 (gen1)               第 2 代 (gen2)                第 3 代 (gen3)
 host 编译器(Go)              gen1 编译 src/*.klx            gen2 编译 src/*.klx
 编译 src/*.klx               产出 LLVM IR → llc → 二进制     同样流程
      │                            │                            │
      ▼                            ▼                            ▼
 产出 LLVM IR ──逐字节对比──► 产出 LLVM IR ──逐字节对比──► 产出 LLVM IR
      └───── fp ≡ g5_all ≡ g6_all：三代输出完全一致 = 稳定不动点 ─────┘

 关键：gen2 是纯原生二进制，构建它不需要 Go！
 → Kylix 从此可以脱离 Go 工具链自我繁殖 🎉
```

> 🐛 调试这个故事花了几十轮。最有趣的两个 bug：for 循环计数器被误绑到全局变量槽（污染后死循环）；布尔数组写入按 1 字节、读取按 8 字节算偏移（数据错乱）。细节见 CHANGELOG 的 P4.10/P4.11。

### 没有 Go 环境的机器怎么办？—— KylixRT

`kylix run --backend=auto` 的探测逻辑：

```
 kylix run hello.klx
        │
        ▼
   本机有 Go 吗？
    │        │
   有│        │没有
    ▼        ▼
 Go 后端    LLVM 后端（KylixRT 路径）
 (最快)     找到 llvm/ 工具链 → IR → llc → clang → 二进制 → 运行
            甚至自带 `kylix test / kylix bench --backend=llvm`
            （没有 Go 也能跑测试和基准）
```

---

## 13. 命令速查表

```bash
kylix build <file.klx>            # 编译（默认 Go 后端）
kylix build <file.klx> --backend=llvm        # 用 LLVM 后端
kylix build <file.klx> --target=linux-arm64  # 交叉编译
kylix build <file.klx> --llvm-opt=2          # LLVM -O2 优化
kylix build <file.klx> -g                    # 带 DWARF 调试符号（lldb 可逐行调试）
kylix run <file.klx>              # 编译并运行（auto 后端）
kylix test [目录]                  # 跑测试（可 --backend=llvm）
kylix bench                       # 基准测试
kylix doc --openapi               # 从 KylixBoot 注解生成 OpenAPI 3.1
kylix doctor                      # 环境体检
kylix lsp                         # 语言服务器（IDE 集成用）
kylix pkg add <pkg>               # 包管理器
kylix repl                        # 交互式解释器
```

工具链全家福：

```
┌─────────────┬──────────────────────────────────────────┐
│ 编译器       │ kylix（Go 编写，支持自举）                  │
│ 双后端       │ Go（默认）/ LLVM（--backend=llvm）          │
│ 运行时       │ KylixRT（无 Go 环境一键编译运行）            │
│ Web 框架     │ KylixBoot（注解驱动 + OpenAPI）             │
│ IDE         │ VS Code 插件 / JetBrains 插件（LSP4IJ）     │
│ LSP         │ 补全/跳转/重命名/格式化/悬停文档             │
│ 调试         │ DWARF + LLDB/GDB 逐行调试                  │
└─────────────┴──────────────────────────────────────────┘
```

---

## 14. 常见问题

**Q: 我需要先学 Pascal 吗？**
A: 不需要。本文就是全部入门所需；Pascal 老手会觉得 `begin...end`、`:=`、`div/mod` 都是老朋友。

**Q: `=` 和 `:=` 的区别？**
A: `:=` 是赋值（`a := 1`），`=` 是比较（`if a = 1 then`）。这是 Pascal 传统，写两天就习惯了。

**Q: 报错 "llvm: command not found"？**
A: 装 LLVM（`brew install llvm` / `apt install llvm clang`），或先用默认 Go 后端。`kylix doctor` 会给具体建议。

**Q: 想用老写法 `Begin/END` 大小写混着来可以吗？**
A: 可以，Kylix 对关键字大小写不敏感（惯例是小写）。

**Q: 教程在哪？跑不通怎么办？**
A: `examples/complete-tutorial/` 有 51 个由浅入深的示例（01_basics 到 21_variant，每个可独立编译运行）。跑不通先 `kylix doctor`。

**Q: 生产能用吗？**
A: v0.6.9 已达成编译器自举闭环 + 双后端全教程回归，但仍在 pre-1.0 快速迭代中（下一步 v0.7.0 是 Web 页面框架）。适合学习、原型与内部工具；上生产请锁定版本并自行评估。

---

## 15. 下一步去哪

```
 你在这里
    │
    ▼
[本文：入门] ──► [51 个官方教程] ──► [进阶专题文档] ──► [参与开发]
                 examples/           docs/
                 complete-tutorial/  ├─ GETTING_STARTED.md     环境与工具链
                 01_basics     ▲     ├─ llvm-backend.md         LLVM 后端内幕
                 02_control_flow│    ├─ WEB_FRAMEWORK.md       KylixBoot 详解
                 03_functions   │    ├─ ORM_GUIDE.md            ORM 注解
                 04_oop         │    ├─ TEMPLATE_GUIDE.md       模板引擎
                 ...            │    ├─ compile-performance.md  性能数据
                 21_variant ────┘    └─ SELFHOSTING_DEV_GUIDE.md 自举开发指南
```

| 你想… | 去看 |
|-------|------|
| 系统学语法 | `examples/complete-tutorial/` 按目录顺序刷 |
| 查某个 stdlib 函数 | `stdlib/klx/` 声明文件 |
| 了解 Web 开发 | `docs/WEB_FRAMEWORK.md` |
| 深入编译器实现 | `docs/llvm-backend.md` + `docs/SELFHOSTING_DEV_GUIDE.md` |
| 官网 | [kylix.top](https://kylix.top) |

欢迎来到 Kylix 的世界——`WriteLn('Have fun!');` 🎉
