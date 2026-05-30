# Kylix 工具通俗解释

用最简单的方式解释 Kylix 各种工具的作用。

## 目录

- [kylix fmt - 代码格式化](#1️⃣-kylix-fmt---代码格式化自动排版)
- [kylix repl - 交互式编程](#2️⃣-kylix-repl---交互式编程即时计算器)
- [kylix lsp - 智能编程助手](#3️⃣-kylix-lsp---智能编程助手)
- [LSP 服务器 - 通用协议](#4️⃣-lsp-服务器---通用的智能助手协议)
- [总结对比](#-总结对比)
- [实际使用建议](#-实际使用建议)

---

## 1️⃣ kylix fmt - 代码格式化（自动排版）

**类比**：就像 Word 的"自动排版"功能

### 问题场景

你写代码时可能很随意：

```pascal
program test;
begin
var x:=10;
  if x>5 then
begin
WriteLn('big');
end;
end.
```

缩进乱七八糟，看起来很难受 😵

### fmt 的作用

运行 `kylix fmt` 后，代码自动变得整齐：

```pascal
program test;

begin
  var x := 10;
  if x > 5 then
  begin
    WriteLn('big');
  end;
end.
```

**实际好处**：

- 📖 代码更容易阅读
- 👥 团队代码风格统一
- 🔍 更容易发现错误

**目前状态**：基础版本，只检查语法，还没实现真正的格式化

---

## 2️⃣ kylix repl - 交互式编程（即时计算器）

**类比**：就像手机计算器，输入立即看到结果

### 传统方式（麻烦）

1. 创建文件 `test.klx`
2. 写代码
3. 保存文件
4. 运行 `kylix run test.klx`
5. 查看结果
6. 修改代码，重复以上步骤...

### REPL 方式（快速）

```bash
$ kylix repl
Kylix REPL v0.2.0
Type Kylix code (press Enter twice to execute, Ctrl+D to exit)

kylix> var x := 10
...    var y := 20
...    WriteLn(x + y)

30

kylix> WriteLn('Hello')

Hello

kylix>
```

**实际好处**：

- ⚡ 快速测试小段代码
- 🧪 学习语言特性时即时实验
- 🔧 调试时快速验证想法

**适合场景**：

- "这个函数怎么用？" → 立即在 REPL 里试试
- "这个算法对吗？" → 快速验证
- "我想算个东西" → 当计算器用

**目前状态**：基础版本，能用但功能简单

---

## 3️⃣ kylix lsp - 智能编程助手

**类比**：就像输入法，打字时给你提示和建议

### 没有 LSP 的痛苦

你在 VS Code 里写代码：

```pascal
begin
  Wri  // 呃... WriteLn 怎么拼来着？
```

你得自己记住所有函数名、参数、拼写...

### 有 LSP 的幸福

```pascal
begin
  Wri  // 自动弹出提示框
       ┌─────────────────────┐
       │ WriteLn(...)        │
       │ Write(...)          │
       │ WriteLn to stdout   │
       └─────────────────────┘
```

按 Tab 键自动补全：

```pascal
begin
  WriteLn('Hello!')
```

### LSP 提供的智能功能

#### 🎯 自动补全

```pascal
var myString := "hello";
myString.  // 按下 . 自动提示
           ┌──────────────┐
           │ Length()     │
           │ ToUpper()    │
           │ ToLower()    │
           └──────────────┘
```

#### 💡 悬停提示

鼠标放在函数名上，显示说明：

```pascal
WriteLn('test')
// 鼠标悬停在 WriteLn 上显示：
// WriteLn(values...)
// 输出内容到控制台并换行
```

#### ⚠️ 实时错误检查

```pascal
var x := 10
var y := x + "hello"  // 红色波浪线提示：类型不匹配
```

#### 🔗 跳转定义

```pascal
function Add(a, b: Integer): Integer;
begin
  result := a + b;
end;

begin
  var sum := Add(10, 20);  // Ctrl+点击 Add 跳转到函数定义
end.
```

**目前状态**：基础版本，能启动，但功能还不完整

---

## 4️⃣ LSP 服务器 - 通用的智能助手协议

### 为什么需要 LSP？

**以前的问题**：

```
VS Code    → 需要写一个 Kylix 插件（1000行代码）
Sublime    → 需要写一个 Kylix 插件（1000行代码）
Vim        → 需要写一个 Kylix 插件（1000行代码）
IntelliJ   → 需要写一个 Kylix 插件（1000行代码）
```

每个编辑器都要单独开发，累死了 😫

**LSP 的解决方案**：

```
VS Code    ─┐
Sublime     │
Vim         ├─→ LSP 客户端 ─→ LSP 服务器（kylix lsp）
IntelliJ   ─┘
```

只需要开发**一个** LSP 服务器，所有编辑器都能用！

### 工作流程

```
你在 VS Code 里打字
       ↓
VS Code 发送请求给 kylix lsp
       ↓
kylix lsp 分析代码，返回建议
       ↓
VS Code 显示建议给你
```

**具体例子**：

1. 你输入 `Wri`
2. VS Code 问 kylix lsp："用户输入了 Wri，有什么建议？"
3. kylix lsp 回答："有 WriteLn 和 Write，建议用 WriteLn"
4. VS Code 显示补全提示给你

### 实际好处

**对开发者**：

- ✅ 一个工具支持所有编辑器
- ✅ 节省大量开发时间
- ✅ 统一的用户体验

**对用户**：

- ✅ 可以用自己喜欢的编辑器
- ✅ 不同编辑器功能一致
- ✅ 更容易找到支持 Kylix 的编辑器

---

## 📊 总结对比

| 工具 | 作用 | 类比 | 使用频率 |
|------|------|------|----------|
| **fmt** | 代码排版 | Word 自动排版 | 低（偶尔用） |
| **repl** | 即时实验 | 计算器 | 中（学习/调试时） |
| **lsp** | 智能助手 | 输入法提示 | 高（写代码时一直用） |
| **LSP 服务器** | 通用协议 | USB 接口标准 | 用户不直接接触 |

---

## 🎯 实际使用建议

### 初学者（你现在的阶段）

1. **主要用 `kylix run`** - 编译并运行程序
2. **偶尔用 `kylix repl`** - 测试小段代码
3. **暂时不用 fmt 和 lsp** - 等功能完善

### 进阶使用

1. **安装 VS Code 扩展** - 获得智能提示
2. **使用 `kylix check`** - 检查语法错误
3. **使用 `kylix fmt`** - 保持代码整洁

### 团队协作

1. **必须用 `kylix fmt`** - 统一代码风格
2. **CI/CD 中用 `kylix check`** - 自动检查代码质量

---

## 💡 通俗理解

把 Kylix 想象成一个**写作工具**：

- `kylix new` = 创建新文档
- `kylix build` = 保存文档
- `kylix run` = 保存并预览
- `kylix check` = 拼写检查
- `kylix fmt` = **自动排版**（让文档整齐）
- `kylix repl` = **草稿纸**（快速试验）
- `kylix lsp` = **智能助手**（自动补全、语法提示）
