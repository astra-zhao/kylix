# Kylix 入门教程 — 从简到难

> 本教程包含 5 个经过测试的 Kylix 示例，从 Hello World 到类与对象，逐步介绍 Kylix 语言的核心特性。
> 
> **环境要求**: 已安装 Kylix 编译器（v3.0.0-alpha 或更高版本）
> 
> **运行方式**: `kylix build example.klx && go run example.go`

---

## 示例 1: Hello World — 第一个程序

最简单的 Kylix 程序，输出 "Hello, Kylix!"。

```pascal
program HelloWorld;
begin
  WriteLn('Hello, Kylix!');
end.
```

**运行**:
```bash
kylix build example1_hello.klx
go run example1_hello.go
```

**输出**:
```
Hello, Kylix!
```

**说明**:
- `program` 关键字声明程序名
- `begin...end.` 包裹主程序块
- `WriteLn()` 输出一行文本

---

## 示例 2: 变量与类型 — 数据存储

演示 Kylix 的基本数据类型：`String`、`Integer`、`Real`、`Boolean`。

```pascal
program Variables;
var
  name: String;
  age: Integer;
  score: Real;
  passed: Boolean;
begin
  name := 'Alice';
  age := 25;
  score := 89.5;
  passed := score >= 60.0;

  WriteLn('Name: ' + name);
  WriteLn('Age: ' + IntToStr(age));
  if passed then
    WriteLn('Status: Passed')
  else
    WriteLn('Status: Failed');
end.
```

**输出**:
```
Name: Alice
Age: 25
Status: Passed
```

**说明**:
- `var` 声明变量
- `:=` 赋值运算符
- `IntToStr()` 将整数转为字符串（用于拼接）
- `if...then...else` 条件判断

---

## 示例 3: 函数与过程 — 代码复用

函数返回值，过程不返回值。演示递归（Factorial）和参数传递。

```pascal
program FunctionDemo;

function Add(a: Integer; b: Integer): Integer;
begin
  result := a + b;
end;

function Factorial(n: Integer): Integer;
begin
  if n <= 1 then
    result := 1
  else
    result := n * Factorial(n - 1);
end;

procedure Greet(name: String);
begin
  WriteLn('Hello, ', name, '!');
end;

var
  x, y, sum: Integer;
  fact: Integer;
begin
  x := 10;
  y := 20;
  sum := Add(x, y);
  WriteLn(IntToStr(x), ' + ', IntToStr(y), ' = ', IntToStr(sum));

  fact := Factorial(5);
  WriteLn('5! = ', IntToStr(fact));

  Greet('Kylix User');
end.
```

**输出**:
```
10 + 20 = 30
5! = 120
Hello, Kylix User!
```

**说明**:
- `function` 有返回值（用 `result` 赋值）
- `procedure` 无返回值
- 函数可以递归调用自己（`Factorial`）

---

## 示例 4: 循环 — 重复执行

演示 `while` 循环和嵌套循环。

```pascal
program LoopsDemo;
var
  i, sum: Integer;
  j: Integer;
begin
  // While loop - 求和 1 到 5
  WriteLn('=== For Loop ===');
  sum := 0;
  i := 1;
  while i <= 5 do
  begin
    sum := sum + i;
    i := i + 1;
  end;
  WriteLn('Sum 1-5: ' + IntToStr(sum));

  // Nested loops - 乘法表
  WriteLn('=== Multiplication Table (3x3) ===');
  i := 1;
  while i <= 3 do
  begin
    j := 1;
    while j <= 3 do
    begin
      WriteLn(IntToStr(i) + ' x ' + IntToStr(j) + ' = ' + IntToStr(i * j));
      j := j + 1;
    end;
    i := i + 1;
  end;
end.
```

**输出**:
```
=== For Loop ===
Sum 1-5: 15
=== Multiplication Table (3x3) ===
1 x 1 = 1
1 x 2 = 2
1 x 3 = 3
2 x 1 = 2
2 x 2 = 4
2 x 3 = 6
3 x 1 = 3
3 x 2 = 6
3 x 3 = 9
```

**说明**:
- `while...do` 循环（条件为真时执行）
- `begin...end` 包裹多条语句
- 嵌套循环：外层控制行，内层控制列

---

## 示例 5: 类与对象 — 面向对象编程

演示类的定义、创建对象、访问字段。

```pascal
program ClassDemo;

type
  TPerson = class
  public
    Name: String;
    Age: Integer;
  end;

var
  person: TPerson;
begin
  person := TPerson.Create;
  person.Name := 'Bob';
  person.Age := 30;
  WriteLn('Person: ' + person.Name + ', Age: ' + IntToStr(person.Age));
end.
```

**输出**:
```
Person: Bob, Age: 30
```

**说明**:
- `type...class` 定义类
- `public` 声明公开字段
- `TPerson.Create` 创建对象实例
- `.` 操作符访问字段

---

## 下一步学习

**中级主题**:
- 数组与记录：`array of Integer`、`record...end`
- 异常处理：`try...except...finally`
- 泛型：`TList<T>`
- 接口：`interface...end`

**高级主题**:
- Web 服务器：`uses web; app := createServer(8080);`
- JSON 处理：`uses jsonutil; obj := ParseJSON(str);`
- 文件 I/O：`uses sysutil; content := ReadFile('data.txt');`
- WASI 编译：`kylix build --wasi main.klx`
- LLVM 后端：`kylix build --backend=llvm main.klx`

**完整文档**: [README_CN.md](../README_CN.md) | [官网](https://kylix.top) | [更新日志](../CHANGELOG.md)

---

## 常见问题

**Q: 如何查看版本？**
```bash
kylix version
```

**Q: 如何格式化代码？**
```bash
kylix fmt myfile.klx
```

**Q: 如何运行测试？**
```bash
kylix test myfile_test.klx
```

**Q: 编译出错怎么办？**
- 检查语法：每个语句后面要有分号 `;`
- 函数返回值：用 `result :=` 而不是 `return`
- 变量先声明：`var x: Integer;` 在 `begin` 之前
- 字符串拼接：用 `+` 连接字符串

---

**版本**: 适用于 Kylix v3.0.0-alpha  
**最后更新**: 2026-06-21
