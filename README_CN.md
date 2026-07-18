# Kylix - 现代 Pascal 语言

[![Official Site](https://img.shields.io/badge/official-kylix.top-4f6ef7.svg)](https://kylix.top)
[![English](https://img.shields.io/badge/lang-English-blue.svg)](README.md)
[![版本](https://img.shields.io/badge/version-5.1.0-blue.svg)](CHANGELOG.md)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![自举](https://img.shields.io/badge/self--hosting-100%25-brightgreen.svg)](ROADMAP.md)

Kylix 是 Pascal 语言的现代化重构,设计为编译到 Go。它将 Pascal 的清晰简洁与现代语言特性结合,并提供完整的 IDE 工具链和编辑器集成。

> 🌐 **官网**: [https://kylix.top](https://kylix.top) — 交互式文档、实时示例和完整功能展示。
>
> 🎉 **v5.1.0**: 完成 Variant 运行时 — `map[String]Variant` 真实化（htab 值槽存 Variant box，`m['pi']=3.14` 按 `variant_compare` 标签派发，JsonDecodeMap 产出真实 Variant map，JsonGet\* 全部 unbox）+ Variant 算术（`variant_add/sub/mul/div` 按标签派发）+ `n := v` 解箱。LLVM 测试 274，教程通过率 **51/51 (100%)**。详见 [CHANGELOG.md](CHANGELOG.md)。

> 🎉 **v5.0.0**: Variant 运行时 — boxed-pointer `{tag, payload}` 动态值（标量 `var v: Variant` + `array of Variant`），赋值按类型装箱、比较按标签派发、`WriteLn` 按标签打印；jsonutil `JsonGetArray` 升级为类型标签化 Variant 切片（与 Go json float64 对齐 → 双端 parity）；顺带修复 `Length(arr)` 路由。LLVM 测试 266，教程通过率 **50/50 (100%)**。

## 特性

### 核心 Pascal 特性
- 强类型并支持类型推导
- 过程和函数
- 控制结构 (if、while、for、case、repeat)
- 记录和数组
- 异常处理

### 现代化扩展
- **类型推导**: `var x := 42;` — 从字面量推导出 Integer
- **泛型约束**: `TBox<T: IComparable>` — 验证类型参数
- **Lambda 表达式**: `var square = (x: Integer) -> x * x;`
- **泛型**: 声明类型参数: `TList<T>`、`function Foo<T>(x: T): T`
- **泛型类型引用**: `TList<Integer>`、`TPair<String, Integer>`
- **Async/Await**: `async function FetchData(): String;`
- **模式匹配**: `match value { 0 => 'zero', _ => 'other' }`
- **类与接口**: 面向对象编程支持
- **属性**: 带有 getter 和 setter
- **ForEach 循环**: `for item in collection do`
- **字符串插值**: `'Hello, ${name}!'`
- **现代异常处理**: try/except/finally,`on E: Type do` 子句

### 完整工具链 (v2.0.0+)
- **测试**: `kylix test` — 发现并运行 `*_test.klx` 中的 `Test*` 过程
- **基准测试**: `kylix bench` — 衡量 `Bench*` 过程的性能
- **文档生成**: `kylix doc` — 从 `//` 文档注释生成 Markdown
- **类型检查**: 含错误代码 (KLX001–499)、错误恢复、"你是否想用?" 建议
- **LSP 服务器**: 完整 IDE 支持 — 补全、悬停、诊断、签名帮助、增量同步 (v2.3.0)
- **包管理器**: `kylix add`、`kylix remove`、`kylix publish` 管理和发布包
- **REPL**: 多行输入、Tab 补全、`:load`/`:type` 元命令 (v2.3.0)
- **调试器**: `kylix debug` 集成 Delve (v2.3.0)
- **WebAssembly**: `kylix build --wasm` 编译为 .wasm (v2.3.0)
- **WASI**: `kylix build --wasi` 编译为 WASI 目标 (v3.0.0-alpha)
- **LLVM 后端**: `kylix build --backend=llvm` 原生代码，绕过 Go 工具链。**48/48 教程编译到原生二进制 (100%)**，支持逐行 DWARF 调试（`-g`，LLDB 逐行单步 + 变量检视，含类方法/lambda DISubprogram + 块作用域 DILexicalBlock）、泛型类方法（TStack<T>.Push/Pop）、静态数组真实下界、jsonutil 嵌套对象/数组解析（JsonGetMap/JsonGetArray/JsonArrayLen/JsonArrayGetString）。
- **KylixBoot 框架**: Spring Boot 式注解驱动的 Web 应用 (v3.1.0)
- **注解自动装配**: `[Controller]`/`[Get]`/`[Post]`/`[Put]`/`[Delete]` 自动路由注册 (v3.2.0)
- **依赖注入**: `[Service]`/`[Component]`/`[Inject]` 编译期自动装配 (v3.2.0)
- **Procedure 风格 handler**: `procedure M(req; res)` 与 function 风格并存 (v3.2.0)
- **校验注解**: `[Required]`/`[Email]`/`[Min]`/`[Max]`/`[MinLen]`/`[MaxLen]` → `Validate()`/`IsValid()` (v3.2.0)
- **安全注解**: `[Authenticated]`/`[Role('admin')]` 按路由守卫 (v3.2.0)
- **ORM 注解**: `[Entity]`/`[Column]`/`[PrimaryKey]`/`[Repository]`/`[Query]` 声明式数据层 (v3.2.0)
- **注解诊断**: KLX207–KLX213 框架契约错误 (v3.2.0)
- **国际化**: 通过 `KYLIX_LANG=zh` 切换中文错误消息 (v2.3.0)

## 安装

```bash
# 克隆仓库
git clone https://github.com/astra-zhao/kylix.git
cd kylix

# 编译编译器
go build -o kylix cmd/kylix/main.go

# 加入 PATH (可选)
export PATH=$PATH:$(pwd)
```

## 快速开始

```bash
# 创建新项目
./kylix new myapp
cd myapp

# 编译并运行
./kylix run

# 检查语法
./kylix check

# 格式化代码
./kylix fmt
```

## CLI 命令

```bash
kylix new <name>       # 创建新项目
kylix build            # 编译项目或文件
kylix build --wasm     # 编译为 WebAssembly
kylix build --wasi     # 编译为 WASI (wasip1/wasm, Go 1.21+)
kylix build --backend=llvm  # 通过 LLVM 原生后端编译
kylix build --backend=llvm --llvm-opt=2  # 启用 LLVM 优化（-O2）
kylix run              # 编译并运行
kylix check            # 项目级类型检查 (跨文件)
kylix fmt              # 格式化源文件
kylix test             # 运行测试 (*_test.klx)
kylix bench            # 运行基准测试 (*_bench.klx)
kylix doc              # 从注释生成 Markdown 文档
kylix debug            # 启动 Delve 调试器 (需安装 dlv)
kylix repl             # 交互式 REPL
kylix lsp              # 启动 LSP 服务 (供编辑器)
kylix add <pkg>        # 添加依赖包
kylix remove <pkg>     # 删除依赖包
kylix publish          # 发布包到注册中心
kylix version          # 显示版本
kylix help             # 显示帮助
```

## 示例

### Hello World
```pascal
program Hello;
begin
  WriteLn('Hello, Kylix World!');
end.
```

### 类型推导
```pascal
var count := 42;
var message := 'Inferred as string';
var ratio := 3.14;
```

### 函数
```pascal
function Add(a: Integer; b: Integer): Integer;
begin
  result := a + b;
end;

// Lambda
var square := (x: Integer) -> x * x;
```

### 类
```pascal
class Animal
private
  var Name: String;
public
  constructor Create(name: String);
  begin
    Name := name;
  end;

  procedure Speak; virtual;
  begin
    WriteLn(Name, ' makes a sound');
  end;
end;

class Dog inherits Animal
public
  procedure Speak; override;
  begin
    WriteLn(Name, ' barks!');
  end;
end;
```

### 模式匹配
```pascal
match value {
  0 => 'zero',
  1, 2, 3 => 'small',
  when value > 100 => 'large',
  _ => 'other'
};
```

### Async/Await
```pascal
async function FetchData(url: String): String;
begin
  // 异步操作
  result := 'Data from ' + url;
end;

var data := await FetchData('http://example.com');
```

### 异常处理 (ON 子句)
```pascal
try
  raise Exception.Create('test');
except
  on E: Exception do
    WriteLn('Caught: ' + E.Message);
  else
    WriteLn('Unknown exception');
end;
```

### 泛型类与泛型函数
```pascal
type
  TPair<T1, T2> = class
    First: T1;
    Second: T2;
  end;

function CreatePair<T>(x: T; y: T): TPair<T, T>;
begin
  Result := TPair<T, T>.Create;
end;

var
  pair: TPair<Integer, String>;
begin
  pair := CreatePair<Integer>(42, 'hello');
end.
```

### 匿名过程与匿名函数
```pascal
// 匿名过程
var greet := procedure()
begin
  WriteLn('Hello!');
end;
greet();

// 带参数的匿名函数
var add := function(a: Integer; b: Integer): Integer
begin
  result := a + b;
end;
WriteLn(add(10, 20));  // 30
```

### Web 服务器
```pascal
program WebApp;
uses web;
var
  app: TServer;
begin
  app := web.createServer(8080);

  // GET 路由
  app.get('/', procedure(req: TRequest; res: TResponse)
  begin
    res.send('Hello, Kylix Web!');
  end);

  // 带路径参数的 JSON API
  app.get('/api/users/:id', procedure(req: TRequest; res: TResponse)
  var
    userId: String;
  begin
    userId := req.param('id');
    res.json(record id := userId; name := 'User ' + userId; end);
  end);

  // 处理 JSON body 的 POST 路由
  app.post('/api/users', procedure(req: TRequest; res: TResponse)
  var
    body: record name: String; email: String; end;
  begin
    req.json(body);
    res.status(201).json(body);
  end);

  // 中间件
  app.use(web.loggerMiddleware());

  // 静态文件
  app.static('/public', './static');

  app.listen();
end.
```

### 测试 (v2.0.0+)
```pascal
// math_test.klx
unit math_test;
uses math;

procedure Setup;
begin
  // 每个测试前运行
end;

procedure Teardown;
begin
  // 每个测试后运行 (defer)
end;

procedure TestAdd;
begin
  Assert(Add(2, 3) = 5, 'expected 2+3=5');
end;
```

```bash
$ kylix test --filter Add math_test.klx
  ok  TestAdd
1 passed, 0 failed (filter: "Add")
```

## 标准库

### Web 框架 (`web`)
HTTP 服务器,支持路由、中间件、请求/响应处理。

```pascal
uses web;

app := web.createServer(8080);
app.get('/api/users', procedure(req: TRequest; res: TResponse)
begin
  res.json(users);
end);
app.listen();
```

### 依赖注入 (`container`)
IoC 容器,支持 singleton、transient、scoped 生命周期。

```pascal
uses container;

di := NewContainer;
di.RegisterSingleton('UserService', function: TUserService
begin
  result := TUserService.Create;
end);

service := di.Resolve('UserService').(TUserService);
```

### 配置 (`config`)
从环境变量加载配置,提供类型安全的访问器。

```pascal
uses config;

cfg := NewConfig;
cfg.SetPrefix('APP');
cfg.LoadFromEnv;

port := cfg.GetIntDefault('PORT', 8080);
debug := cfg.GetBoolDefault('DEBUG', false);
```

### 中间件 (`middleware`)
为常见 Web 应用预置的中间件。

```pascal
uses middleware;

app.use(NewRequestIDMiddleware.Handle);
app.use(NewLoggingMiddleware.Handle);
app.use(NewCORSMiddleware.Handle);
app.use(NewAuthMiddleware(ValidateToken).Handle);
app.use(NewRateLimitMiddleware(100, 60).Handle);
```

### 验证 (`validation`)
请求验证,提供流式 API 和常用验证器。

```pascal
uses validation;

validator := NewRequestValidator(req);
validator.Required(['name', 'email']);
validator.Email('email');
validator.MinLength('password', 8);

if not validator.IsValid then
  res.status(400).json(validator.Errors);
```

### ORM (`orm`)
数据库 ORM,支持 MySQL、PostgreSQL、SQLite,提供查询构造器和迁移工具。

```pascal
uses orm;

// 连接数据库
dbConfig := TConnectionConfig{
  Type: DBSQLite,
  Database: './app.db'
};
db := NewDatabase(dbConfig);
orm := NewORM(db);

// 插入
data := map[string]interface{}{
  'name': 'John',
  'email': 'john@example.com'
};
id := orm.Insert('users', data);

// 用查询构造器查询
qb := orm.QueryBuilder('users');
qb.Where('age', '>', 18);
qb.OrderBy('name', 'ASC');
qb.Limit(10);
users := orm.Execute(qb);

// 按 ID 查找
user := orm.Find('users', 1);

// 更新
orm.Update('users',
  map[string]interface{}{'id': 1},  // 条件
  map[string]interface{}{'name': 'Jane'}  // 数据
);

// 删除
orm.Delete('users', map[string]interface{}{'id': 1});
```

### 模板引擎 (`template`)
HTML 模板渲染,支持 layout、partial、自定义函数。

```pascal
uses template;

engine := NewTemplateEngine;
engine.SetTemplateDir('./templates');

// 注册 layout
engine.RegisterLayout('main',
  '<html><body>{{.Content}}</body></html>');

// 注册 partial
engine.RegisterPartial('header', '<h1>My App</h1>');

// 渲染
view := NewView(engine);
view.With('Title', 'Home');
view.With('Message', 'Welcome!');
view.WithLayout('main');
html := view.Render('home.html');
res.HTML(html);
```

### 自动配置 (`autoconfig`)
从多个源加载配置,带环境检测。

```pascal
uses autoconfig;

config := NewAutoConfig('myapp');
config.DetectEnvironment;       // 从 APP_ENV 检测
config.SetConfigDir('./config');
config.AddDefaultSources;       // config.json、config.{env}.json、env vars
config.Load;

// 访问配置
port := config.GetInt('server.port');
dbHost := config.GetString('database.host');
debug := config.GetBool('app.debug');

// 环境检查
if config.IsProduction then
  // 生产环境特定逻辑
```

### 文件 I/O (`sysutil`)
文件和目录操作,提供 Pascal 风格的 API。

```pascal
uses sysutil;

// 读/写文件
content := sysutil.ReadFile('data.txt');
sysutil.WriteFile('output.txt', 'Hello, World!');
sysutil.AppendFile('log.txt', 'New line');

// 文件操作
if sysutil.FileExists('config.json') then
  WriteLn('Config found');

sysutil.CreateDir('new_folder');
sysutil.CopyFile('src.txt', 'dst.txt');
sysutil.DeleteFile('temp.txt');

// 列文件
files := sysutil.ListDir('./');
matches := sysutil.ListFiles('*.klx');

// 路径工具
fullPath := sysutil.PathJoin('dir', 'sub', 'file.txt');
dir := sysutil.PathDir('/home/user/doc.txt');
ext := sysutil.PathExt('photo.jpg');

// 行级 I/O
lines := sysutil.ReadLines('data.csv');
sysutil.WriteLines('output.csv', lines);
```

### JSON (`jsonutil`)
JSON 编解码与操作。

```pascal
uses jsonutil;

// 编码为 JSON
jsonStr := jsonutil.JsonEncode(data);
pretty := jsonutil.JsonEncodePretty(data);

// 解析 JSON
obj := jsonutil.JsonDecodeMap('{"name": "Kylix", "version": 1}');
name := jsonutil.JsonGetString(obj, 'name');
ver := jsonutil.JsonGetInt(obj, 'version');

// 类型安全访问器
flag := jsonutil.JsonGetBool(obj, 'active');
pi := jsonutil.JsonGetFloat(obj, 'pi');
child := jsonutil.JsonGetMap(obj, 'nested');
items := jsonutil.JsonGetArray(obj, 'list');

// 验证
if jsonutil.JsonIsValid(input) then
  WriteLn('Valid JSON');

// 文件 I/O
data := jsonutil.JsonReadFile('config.json');
jsonutil.JsonWriteFile('output.json', data);
```

### 日期时间 (`datetime`)
日期与时间操作,支持算术和格式化。

```pascal
uses datetime;

// 当前时间
now := datetime.Now();
WriteLn(now.FormatDateTime());  // 2024-06-15 10:30:00
WriteLn(now.FormatDate());      // 2024-06-15

// 创建日期
birthday := datetime.MakeDate(1990, 5, 15);
meeting := datetime.MakeTime(2024, 12, 25, 14, 30, 0);

// 日期算术
nextWeek := now.AddDays(7);
nextMonth := now.AddMonths(1);
tomorrow := now.AddDays(1);

// 比较
days := now.DiffDays(birthday);
if now.After(deadline) then
  WriteLn('Overdue!');

// 工具方法
if now.IsWeekend() then
  WriteLn('Weekend!');
if now.IsLeapYear() then
  WriteLn('Leap year');
WriteLn('Day: ' + now.DayName());
WriteLn('Month: ' + now.MonthName());

// 解析
dt := datetime.ParseDate('2024-06-15');
dt2 := datetime.ParseDateTime('2024-06-15 10:30:00');

// 时间戳
ts := datetime.GetTimestamp();    // Unix 秒
tsMs := datetime.GetTimestampMs(); // Unix 毫秒
```

### 正则表达式 (`regex`)
模式匹配、查找、文本替换。

```pascal
uses regex;

// 快速模式检查
if regex.IsEmail('user@example.com') then
  WriteLn('Valid email');

if regex.IsNumeric('12345') then
  WriteLn('All digits');

if regex.IsURL('https://example.com') then
  WriteLn('Valid URL');

// 查找与替换
match := regex.RegexFind('[0-9]+', 'Order #12345');
// match = '12345'

result := regex.RegexReplace('\s+', 'a  b  c', ' ');
// result = 'a b c'

// 编译后的 regex (可重用)
re := regex.RegexMustCompile('(\w+)@(\w+)');
if re.Match('user@host') then
  groups := re.Groups('user@host');
  // groups[1] = 'user', groups[2] = 'host'

// 分割
parts := regex.RegexSplit(',', 'a,b,c,d');
// parts = ['a', 'b', 'c', 'd']

// 提取所有数字
nums := regex.ExtractNumbers('Room 42, Floor 3, Building 7');
// nums = ['42', '3', '7']
```

### HTTP 客户端 (`httpclient`) — v3.0.0-alpha

一键 HTTP 辅助函数和可复用客户端，支持自定义 Header。

```pascal
uses httpclient;

// 一键 GET
body := HttpGet('https://api.example.com/data');

// 一键 POST，返回 JSON 响应
resp := HttpGetJSON('https://api.example.com/items');

// 可复用客户端，带自定义 Header
client := NewHttpClient('https://api.example.com');
client.SetHeader('Authorization', 'Bearer ' + token);
body := client.Get('/users');
WriteLn(IntToStr(client.StatusCode()));
```

### WASI (`wasi`) — v3.0.0-alpha

适用于 WASI 运行时（Wasmtime、Node.js、Cloudflare Workers）的可移植系统接口。

```pascal
uses wasi;

begin
  WriteLn('Hello from WASI!');
  WriteLn('Arg count: ' + IntToStr(ArgCount()));
  WriteLn('First arg: ' + Arg(0));

  var val := GetEnvOrDefault('PORT', '8080');
  WriteLn('PORT=' + val);

  var ms := ElapsedMs();
  WriteLn('Elapsed: ' + IntToStr(ms) + ' ms');
end.
```

### KylixBoot 框架 (`boot`) — v3.1.0+

Spring Boot 式 Web 框架：带路径参数的路由、DI 容器、优雅停机、环境变量配置和内置中间件。

```pascal
program HelloBoot;
uses boot;

begin
  // 通过全局快捷方式注册路由
  boot.GET('/', procedure(req: TRequest; res: TResponse)
  begin
    res.Send('Hello, KylixBoot!');
  end);

  boot.GET('/users/:id', procedure(req: TRequest; res: TResponse)
  begin
    res.JSON(record id := req.Param('id'); end);
  end);

  // 内置中间件
  boot.Use(boot.Logger());
  boot.Use(boot.Recover());
  boot.Use(boot.CORS());

  // 内置优雅停机
  boot.Listen(':8080');
end.
```

容器支持：

```pascal
// DI 容器
container := boot.NewContainer();
container.RegisterSingleton('UserService', TUserService);
container.RegisterTransient('Request', TRequestScope);

// 基于反射的注入
container.Inject(controller);
```

23 个单元测试位于 `pkg/boot/`；声明文件在 `stdlib/klx/boot.klx`。

### KylixBoot 注解 — v3.2.0+

编译器现在会直接从注解自动装配整套栈，你不再需要手动调用 `boot.GET`：

```pascal
program App;
uses boot, orm;

[Entity('users')]
type
  TUser = class
    [PrimaryKey]
    Id: Integer;
    [Column('email')]
    Email: String;
  end;

[Repository(TUser)]
type
  TUserRepository = class
    [Query('SELECT * FROM users WHERE email = ?')]
    function ByEmail(email: String): TUser;
  end;

[Service]
type
  TUserService = class
    [Inject]
    Repo: TUserRepository;
    function Greeting(): String;
    begin result := 'hello from service'; end;
  end;

[Controller('/api')]
type
  TUserController = class
    [Inject]
    UserService: TUserService;

    [Get('/hello')]
    [Authenticated]
    function Hello(req: TRequest): TResponse;
    begin
      result := BootText(200, self.UserService.Greeting());
    end;

    [Get('/users')]
    [Role('admin')]
    function Users(req: TRequest): TResponse;
    begin
      result := BootJSON(200, record ok := true; end);
    end;
  end;

begin
  WriteLn('KylixBoot annotations ready');
end.
```

编译器会在你的 `main()` 之前生成：

- `[Service]`/`[Component]` 类的单例 + `BootRegisterInstance`
- `TUser.ToRow()` / `FromRow()` 字段映射助手
- `TUserRepository.FindAll/FindById/Save/DeleteById` + `ByEmail` 查询体
- 强类型 `[Inject]` 字段赋值
- 通过 `BootGET`/`BootPOST` 注册的路由闭包，并为 `[Authenticated]`/`[Role]` 注入 `BootEnforceAuth`/`BootEnforceRole` 守卫

非法注解会在 codegen 之前以 `KLX207`–`KLX213` 报错（重复路由、非法参数、不支持的 handler 签名、缺失注入目标、非法 validation/security/ORM 用法）。

可运行示例见 `examples/complete-tutorial/12_special_features/example42..47`。

### 注解语法 (`[Attribute]`) — v3.1.0+

注解为类、类型、函数和字段附加元数据，是声明式 API（路由注册、ORM 映射、DI、校验）的基础。

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

    [Post('/'), Authenticated]
    function CreateUser(req: TRequest): TResponse;
    begin
      var user := req.Body<TUser>();
      UserRepo.Save(user);
      result := req.Created(user);
    end;
  end;

[Entity('users')]
type
  TUser = class
    [Column('id'), PrimaryKey]
    Id: Integer;
    [Required, Email]
    Email: String;
  end;
```

注解在 AST 层解析（`ast.Attribute`），附加到 `ClassDecl`/`TypeDecl`/`FunctionDecl`/`VarDecl`，为 v3.2 的自动路由注册 + ORM 代码生成做准备。

### 纯 Kylix stdlib (v2.1+)

四个模块完全用 Kylix 自身实现:

| 模块 | 函数数 | 内容 |
|------|--------|------|
| `strutil` | 8 | Reverse、StartsWith、EndsWith、Contains、PadLeft、PadRight... |
| `mathutil` | 12 | Abs、Min、Max、Pow、Gcd、Lcm、Factorial、IsPrime... |
| `arrayutil` | 8 | Sum、Product、MinValue、MaxValue、Contains、IndexOf、Reverse... |
| `collections` | TIntList 类 | Count、Get、Add、Clear、IsEmpty、Sum |

```pascal
uses strutil, mathutil;

WriteLn(Reverse('hello'));      // olleh
WriteLn(IsPrime(17));            // true
WriteLn(Pow(2, 10));             // 1024
WriteLn(PadLeft('42', 5, '0'));  // 00042
```

## 语言参考

### 类型
- `Integer` - 64 位整数 (映射到 Go 的 `int64`)
- `Real` - 64 位浮点 (映射到 Go 的 `float64`)
- `Boolean` - 布尔值
- `String` - 字符串
- `Char` - 单字符 (映射到 Go 的 `byte`)

### 操作符
- 算术: `+`、`-`、`*`、`/`、`div`、`mod`
- 比较: `=`、`<>`、`<`、`>`、`<=`、`>=`
- 逻辑: `and`、`or`、`not`、`xor`
- 赋值: `:=`、`=`

### 控制结构
- `if/then/else`
- `while/do`
- `for/to/downto`
- `for/in` (foreach)
- `repeat/until`
- `case/of`
- `match` (模式匹配)
- `try/except/finally`

### 声明
- `var` - 变量声明
- `const` - 常量声明
- `type` - 类型声明
- `function` - 带返回值的函数
- `procedure` - 不带返回值的过程
- `class` - 类声明
- `interface` - 接口声明

## 项目结构

```
kylix/
├── cmd/kylix/          # CLI 入口
│   ├── main.go             # 命令分发
│   ├── cmd_build.go        # build / WASM / WASI / LLVM
│   ├── cmd_run.go          # run
│   ├── cmd_other.go        # check / fmt / new
│   ├── cmd_package.go      # add / install / remove / publish
│   ├── cmd_testcmd.go      # test (Setup/Teardown/--filter)
│   ├── cmd_bench.go        # bench
│   ├── cmd_doc.go          # doc
│   └── cmd_debug.go        # debug (Delve)
├── pkg/
│   ├── compiler/       # 编译 API + 增量缓存
│   ├── project/        # 项目管理 (kylix.toml)
│   ├── pkgmgr/         # 包管理器 (add/install/remove/publish)
│   ├── llvmgen/        # LLVM 原生后端 (v3.0.0-alpha)
│   │   ├── codegen.go      # 生成器核心、SSA、字符串常量池
│   │   ├── expr.go         # 表达式 codegen
│   │   ├── stmt.go         # 语句 codegen
│   │   ├── class.go        # 类/vtable codegen
│   │   └── compile.go      # 完整管道：AST → binary
│   ├── wasi/           # WASI 系统调用层 (v3.0.0-alpha)
│   │   ├── wasi.go         # 包文档
│   │   ├── wasi_stub.go    # 非 WASI stub（本地测试）
│   │   └── wasi_wasip1.go  # WASI 原生实现
│   ├── formatter/      # 源代码格式化
│   ├── lsp/            # Language Server Protocol (含增量同步)
│   ├── repl/           # 交互式 REPL (Tab 补全)
│   ├── testrunner/     # 测试与基准测试
│   ├── docgen/         # 文档生成器
│   └── i18n/           # 错误信息国际化 (中文/英文)
├── registry/           # 包注册中心服务端 (v3.0.0-alpha)
│   ├── internal/
│   │   ├── api/        # REST API 处理器
│   │   ├── auth/       # Bearer token 认证
│   │   ├── db/         # SQLite/PostgreSQL 存储
│   │   └── models/     # 数据模型
│   └── web/templates/  # htmx + Tailwind CSS 前端
├── stdlib/             # 标准库
│   ├── web.go              # Web 框架
│   ├── orm.go              # 数据库连接 + 事务
│   ├── orm_query.go        # QueryBuilder 流式 API
│   ├── orm_migrate.go      # ORM CRUD + 迁移
│   ├── http_client.go      # HTTP 客户端 (v3.0.0-alpha)
│   ├── klx/                # LSP 自动补全声明文件
│   │   ├── sysutil.klx, datetime.klx, regex.klx
│   │   ├── jsonutil.klx, httpclient.klx, wasi.klx
│   └── src/                # 纯 Kylix stdlib 实现
│       ├── strutil.klx, mathutil.klx, arrayutil.klx
│       ├── collections.klx, stringbuilder.klx, resulttype.klx, iter.klx
│       ├── jsonutil.klx, regex.klx, datetime.klx  # Phase 4 (v3.0.0-alpha)
│       ├── httpclient.klx, wasi.klx               # Phase 5 (v3.0.0-alpha)
├── token/              # token 定义
├── lexer/              # 词法分析
├── ast/                # AST 节点
├── parser/             # Pratt parser (按职责拆分)
│   ├── parser.go           # 核心
│   ├── parser_decl.go      # 声明
│   ├── parser_stmt.go      # 语句
│   └── parser_expr.go      # 表达式
├── generator/          # Go 代码生成器
│   ├── generator.go        # 核心 + pre-scan
│   ├── generator_types.go  # 类/接口/泛型
│   ├── generator_stmt.go   # 语句
│   └── generator_expr.go   # 表达式
├── src/                # 自举编译器源码 (.klx)
├── examples/           # 示例程序
│   ├── wasi-hello/         # WASI 示例 (Wasmtime/Node.js)
│   └── cloudflare-worker/  # Cloudflare Workers 示例
├── vscode-ext/         # VS Code 扩展
└── docs/               # 文档
```

## 编辑器集成

### VS Code
`vscode-ext/` 目录包含完整的 VS Code 扩展:
- 语法高亮
- 语言配置 (括号、注释、折叠)
- LSP 客户端集成

```bash
cd vscode-ext
npm install
# 在 VS Code 按 F5 启动扩展
```

### 其他编辑器
Kylix LSP 支持任何带 LSP 客户端的编辑器:
```json
{
  "command": ["kylix", "lsp"],
  "filetypes": ["kylix"]
}
```

## 文档

- [IDE 用户手册](docs/KYLIX_IDE_USER_MANUAL.md) - 完整的 CLI 与编辑器指南
- [开发者指南](docs/KYLIX_DEV_GUIDE.md) - 架构、内部机制、贡献指南
- [工具说明](docs/KYLIX_TOOLS_EXPLAINED.md) - 适合新手的工具说明
- [Web 框架指南](docs/WEB_FRAMEWORK.md) - Web 服务器与 REST API 开发
- [ORM 指南](docs/ORM_GUIDE.md) - 数据库 ORM 与查询构造器
- [模板引擎指南](docs/TEMPLATE_GUIDE.md) - HTML 模板渲染

## 路线图

### Phase 1: 转译器 ✅
- ✅ 词法分析与语法分析
- ✅ AST 生成
- ✅ Go 代码生成
- ✅ 基础语言特性
- ✅ 现代特性 (lambda、async、模式匹配)

### Phase 2: IDE 工具 ✅
- ✅ CLI 工具链 (new、build、run、check、fmt、repl、lsp)
- ✅ 项目管理 (kylix.toml)
- ✅ LSP 服务,支持补全和悬停
- ✅ VS Code 扩展(语法高亮)
- ✅ 交互式 REPL
- ✅ 完整文档

### Phase 3: 框架 ✅
- ✅ Web 服务器 (基于 Go net/http)
- ✅ 路由系统 (GET、POST、PUT、DELETE)
- ✅ 路径参数 (`/users/:id`)
- ✅ 中间件支持
- ✅ JSON 请求/响应
- ✅ 静态文件服务
- ✅ 匿名过程与匿名函数
- ✅ 增强的 VS Code 扩展
- ✅ 依赖注入容器
- ✅ 配置系统
- ✅ 中间件套件 (CORS、Auth、Rate Limit、Request ID、Logging)
- ✅ 请求验证
- ✅ ORM (MySQL、PostgreSQL、SQLite)
- ✅ 模板引擎
- ✅ 自动配置

### Phase 4: 语言增强 ✅
- ✅ 泛型类型参数声明 (类与函数)
- ✅ 异常处理 ON 子句 (`on E: ExceptionType do`)
- ✅ Constructor/destructor/inherited 关键字
- ✅ Lambda 表达式参数解析
- ✅ Async/await 代码生成改进

### Phase 5: 标准库与工具链 ✅
- ✅ 文件 I/O (`sysutil`)
- ✅ JSON (`jsonutil`)
- ✅ 日期时间 (`datetime`)
- ✅ 正则表达式 (`regex`)
- ✅ REPL 改进 (readline 历史)
- ✅ Formatter 修复
- ✅ Generator stdlib 串接

### Phase 6-7: Bug 修复与语言能力 ✅
- ✅ 字符串插值
- ✅ 异常类型 (ON 子句)
- ✅ 多返回值 (`function Div(a,b: Integer): (Integer, Integer)`)
- ✅ 属性代码生成 (getter/setter)
- ✅ Map 类型、Variant 类型、动态数组
- ✅ 枚举类型
- ✅ 多文件模块系统 (`unit X;`、`uses X;`)

### Phase 8: 自举编译器 ✅
- ✅ 7 个 Kylix 源文件 (token、ast、lexer、parser、error、generator、main)
- ✅ 类代码生成
- ✅ 软关键字 (25+ 关键字可作标识符)
- ✅ is/as 类型分发
- ✅ 局部变量声明、构造器、内置函数

### Phase 9: 自举验证 ✅
- ✅ 多文件自举编译
- ✅ 自举生成的 Go 输出零错误编译运行
- ✅ Diff 验证: Go 参考实现 vs Kylix 自举 — 语义等价
- ✅ 15/15 示例在两种编译器上通过

### Phase 10-12 (v2.0.0+) — 生产级编译器 ✅
- ✅ 错误代码体系 (KLX001–499) + 智能建议
- ✅ 类型推导 (`var x := 42`)
- ✅ 泛型约束验证 (`T: IComparable`)
- ✅ 完整测试框架 (`kylix test`)
- ✅ 文档生成器 (`kylix doc`)
- ✅ 性能基准 (`kylix bench`)

### v2.1.0 — 增强类型系统 ✅
- ✅ 多参数泛型约束 (`TMap<K: IComparable, V: IHashable>`)
- ✅ 类→接口实现映射验证 (含方法签名)
- ✅ 增强类型推导 (Boolean、array of T、nil、not 等)
- ✅ stdlib Phase 1 (`strutil`、`mathutil`)

### v2.2.0 — 工程质量 ✅
- ✅ GitHub Actions CI/CD
- ✅ 泛型约束方法签名验证
- ✅ 包级类型检查 (`CheckProject`)
- ✅ 增量编译启用 (BuildCache)
- ✅ stdlib Phase 2 (`arrayutil`、`collections`)

### v2.3.0 — 开发者体验 ✅
- ✅ LSP 增量同步 (textDocumentSync 升级到 Incremental)
- ✅ REPL Tab 补全 + `:load` + `:type`
- ✅ kylix test 高级功能 (Setup/Teardown/--filter)
- ✅ 错误信息国际化 i18n (中英双语)
- ✅ Delve 调试器集成 (`kylix debug`)
- ✅ WebAssembly 后端 (`--wasm`、`--tinygo`)

### v2.4.0–v2.6.0 — 完善、生态与性能 ✅
- ✅ i18n 全面接入、REPL `:type` 真正推导、SetLength 修复
- ✅ 包管理器嵌套依赖 + lockfile、stdlib Phase 3
- ✅ LSP 跨文件 rename、`kylix doc` 代码示例、`kylix bench --mem`、iter 模块
- ✅ 并行编译 (goroutine pool)、死代码消除、LSP 大文件性能基准

### v3.0.0-alpha — 架构突破 ✅
- ✅ LLVM 原生后端 Milestone 1（标量类型、控制流、函数、类/vtable）
- ✅ WASI 支持（`--wasi`、`--tinygo`、`pkg/wasi/`、`stdlib/src/wasi.klx`）
- ✅ 包注册中心服务端（`registry/`、REST API、htmx 前端、`kylix publish`）
- ✅ stdlib Phase 4：纯 Kylix jsonutil（嵌套 JSON）、regex、datetime（DateAdd/DateSub）
- ✅ `external` 函数声明解析修复
- ✅ HTTP 客户端 stdlib（`httpclient`）

### v3.1.0 — KylixBoot + 编译器修复 + LLVM 数组 ✅
- ✅ KylixBoot 框架（`pkg/boot/`，约 700 行，23 测试）—— 路由、DI、中间件、优雅停机
- ✅ 注解语法 `[Name]` / `[Name(args)]`，作用于类、类型、函数和字段
- ✅ KLX-C01 修复：`var p: TClass` 现在生成 `*TClass`（不再是 `interface{}`）
- ✅ KLX-C02 修复：单引号字符串中的 `${...}` 正确生成 STRING_INTERPOLATION
- ✅ KLX-C03 修复：lambda / 匿名函数返回类型保留
- ✅ KLX-C04 修复：match 语句 codegen 生成有效 Go 代码
- ✅ KLX-C05 修复：`uses sysutil/jsonutil/...` 在 program 文件中注入 stdlib 符号（40+ 函数）
- ✅ LLVM Milestone 2 Phase 1：静态 + 动态数组，`--llvm-opt=N`
- ✅ 教程扩展 `example40_declarative_oop.klx` 和 `example41_attributes.klx`（32/34 示例通过）

### v3.2.0-dev：KylixBoot 注解栈 ✅（进行中）
- ✅ 从 `[Controller]`/`[Get]`/`[Post]`/`[Put]`/`[Delete]` 注解自动注册路由
- ✅ DI 自动装配：`[Service]`/`[Component]`/`[Inject]`
- ✅ Procedure 风格路由 handler（`procedure M(req; res)`）
- ✅ 校验注解：`[Required]`/`[Email]`/`[Min]`/`[Max]`/`[MinLen]`/`[MaxLen]` → `Validate()`/`IsValid()`
- ✅ 安全注解：`[Authenticated]`/`[Role('admin')]` 按路由守卫
- ✅ ORM 注解：`[Entity]`/`[Column]`/`[PrimaryKey]`/`[Repository]`/`[Query]`
- ✅ 注解诊断：KLX207（重复路由）… KLX213（非法 ORM）
- ✅ 教程：41/41 示例通过（新增 6 个 KylixBoot 注解示例）
- ✅ LLVM Milestone 2 Phase 2 —— 接口 fat pointer + 成员访问 + 方法分发 + is/as
- ✅ LLVM Milestone 2 Phase 3 —— 泛型类单态化（模板克隆 + 类型参数替换）
- ✅ 包注册中心部署脚手架（Dockerfile / docker-compose / CI — DNS + TLS 后 `make up` 即可上线）
- ✅ stdlib Phase 6：net / crypto / encoding

## 跨平台编译

Kylix 编译为 Go 源码,然后利用 Go 内置的交叉编译产生原生二进制 — 无虚拟机,目标机器无需安装运行时。

### 工作原理

```
你的 .klx 文件
    ↓  kylix build  (Pascal → Go 转译)
生成的 .go 文件
    ↓  go build     (Go → 原生二进制)
原生可执行文件
```

### 不同平台构建

```bash
# Linux (Intel/AMD)
kylix build --target=linux/amd64 main.klx

# Windows (Intel/AMD)
kylix build --target=windows/amd64 main.klx

# macOS Apple Silicon (M1/M2/M3)
kylix build --target=darwin/arm64 main.klx

# macOS Intel
kylix build --target=darwin/amd64 main.klx

# Linux ARM (树莓派、ARM 云)
kylix build --target=linux/arm64 main.klx

# WebAssembly (浏览器/Node.js)
kylix build --wasm main.klx           # 标准 Go (~3 MB)
kylix build --wasm --tinygo main.klx  # TinyGo (~30 KB)
```

所有交叉编译都在本地完成 — 无需远程构建服务器。

最终二进制无外部依赖,终端用户无需安装 Go 或 Kylix 即可运行。

### 支持的目标

| 系统 | 架构 | `--target` 值 |
|----|-------------|-----------------|
| Linux | x86-64 | `linux/amd64` |
| Linux | ARM64 | `linux/arm64` |
| Windows | x86-64 | `windows/amd64` |
| macOS | x86-64 | `darwin/amd64` |
| macOS | Apple Silicon | `darwin/arm64` |
| WebAssembly | wasm | `--wasm` (含可选 `--tinygo`) |
| WASI | wasip1/wasm | `--wasi` (含可选 `--tinygo`) |

### LLVM 原生后端 (v3.0.0-alpha，v3.1.0 扩展)

Kylix 现在有实验性 LLVM 后端，直接从 AST 生成原生二进制，绕过 Go 工具链。

```bash
# 通过 LLVM 编译（需安装 llc + clang）
kylix build --backend=llvm main.klx

# 启用 LLVM 优化（-O0 / -O1 / -O2 / -O3）
kylix build --backend=llvm --llvm-opt=2 main.klx
```

管道：AST → LLVM IR (`.ll`) → 目标文件 (`.o`) → 原生二进制（via `llc` + `clang`）。Go 后端仍为默认。

**Milestone 1 + Phase 1 (v3.1.0) 支持：**
- 所有标量类型、算术/比较/逻辑运算、控制流、函数
- 类（vtable 虚函数分发）
- **静态数组**（`array[1..N] of T` → `alloca [N x T]`）
- **动态数组**（`array of T` → `{ ptr, i64, i64 }` slice 结构体）
- Pascal 1-based 索引自动转换为 LLVM 0-based
- 通过 `--llvm-opt=N` 启用 LLVM 优化 Pass（`llc -O=N`）

```pascal
program Arrays;
var
  fixed: array[1..5] of Integer;
  dyn: array of Integer;
  i: Integer;
begin
  for i := 1 to 5 do fixed[i] := i * i;
  WriteLn(fixed[3]);  // 9
end.
```

接口、泛型和异常计划在 Milestone 2 Phase 2-3（v3.2）中实现。

---

## 更新日志

完整版本历史请见 [CHANGELOG.md](CHANGELOG.md)。最近更新:

### v3.1.0 (2026-06-23)
KylixBoot 框架（路由/DI/中间件，23 测试）、注解语法 `[Name]`、5 个编译器修复（KLX-C01..C05：类变量类型、字符串插值、lambda 返回值、match codegen、uses 符号注入）、LLVM Milestone 2 Phase 1（静态 + 动态数组、`--llvm-opt=N`）。

### v3.0.0-alpha (2026-06-21)
架构突破 — LLVM 原生后端 Milestone 1、WASI 支持、包注册中心服务端、stdlib Phase 4（纯 Kylix jsonutil/regex/datetime）、`external` 解析修复、HTTP 客户端 stdlib。

### v2.6.0 (2026-06-20)
性能与优化 — 并行编译 (goroutine pool)、死代码消除、LSP 大文件性能基准。

### v2.5.0 (2026-06-20)
工具链深化 — LSP 跨文件 rename + codeAction、`kylix doc` 代码示例提取、`kylix bench --mem`、iter 模块、类方法外部定义修复。

### v2.4.0 (2026-06-20)
完善与生态 — i18n 全面接入、REPL `:type` 真正推导、SetLength 修复、包管理器嵌套依赖 + lockfile、stdlib Phase 3。

### v2.3.0 (2026-06-19)
开发者体验全面提升 — LSP 增量同步、REPL Tab 补全、测试 fixtures + filter、i18n 框架、Delve 调试器、WebAssembly 后端。

### v2.0.0 (2026-06-17)
🎉 生产级首发 — 错误代码体系、类型推导、泛型约束、`kylix test/doc/bench` 完整工具链。

## 贡献

欢迎贡献!请随时提交 issue 和 pull request。

## 许可证

MIT License
