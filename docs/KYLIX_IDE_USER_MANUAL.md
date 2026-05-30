# Kylix IDE Tool 使用手册

Kylix 是一个现代化的 Pascal 编译器，将 Kylix 代码编译为 Go 语言。本手册介绍 Kylix IDE 工具的所有功能。

## 目录

- [安装](#安装)
- [快速开始](#快速开始)
- [命令参考](#命令参考)
- [项目结构](#项目结构)
- [编辑器集成](#编辑器集成)
- [常见问题](#常见问题)

---

## 安装

### 从源码构建

```bash
# 克隆仓库
git clone <repository-url>
cd kylix

# 构建编译器
go build -o kylix cmd/kylix/main.go

# 添加到 PATH（可选）
export PATH=$PATH:/path/to/kylix
```

### 验证安装

```bash
kylix version
```

输出示例：
```
Kylix 0.2.0
```

---

## 快速开始

### 1. 创建新项目

```bash
kylix new myproject
```

这将创建以下结构：
```
myproject/
├── kylix.toml          # 项目配置
├── main.klx            # 主程序文件
├── build/              # 编译输出目录
└── .gitignore          # Git 忽略规则
```

### 2. 进入项目目录

```bash
cd myproject
```

### 3. 编译并运行

```bash
kylix run
```

输出：
```
Hello from myproject!
```

---

## 命令参考

### kylix new - 创建新项目

创建一个新的 Kylix 项目，包含基本的项目结构和配置文件。

**语法：**
```bash
kylix new <project-name>
```

**示例：**
```bash
kylix new hello-world
cd hello-world
kylix run
```

**生成的文件：**
- `kylix.toml` - 项目配置文件
- `main.klx` - 主程序入口
- `build/` - 编译输出目录
- `.gitignore` - Git 忽略规则

---

### kylix build - 编译项目

将 Kylix 代码编译为 Go 代码。

**语法：**
```bash
# 编译当前项目
kylix build

# 编译单个文件
kylix build <file.klx>

# 指定输出文件
kylix build -o output.go <file.klx>
```

**选项：**
- `-o, --output <file>` - 指定输出文件路径
- `-v, --verbose` - 显示详细编译信息

**示例：**
```bash
# 编译项目
kylix build
# 输出：✓ Built myproject → build/myproject.go

# 编译单个文件
kylix build calculator.klx
# 输出：✓ Compiled calculator.klx → calculator.go

# 指定输出
kylix build -o app.go main.klx
```

---

### kylix run - 编译并运行

编译并立即运行 Kylix 程序。

**语法：**
```bash
# 运行当前项目
kylix run

# 运行单个文件
kylix run <file.klx>

# 保留生成的 .go 文件
kylix run --keep <file.klx>
```

**选项：**
- `--keep` - 保留生成的 .go 文件（默认删除）
- `-v, --verbose` - 显示详细信息

**示例：**
```bash
# 运行项目
kylix run
# 输出：Hello from myproject!

# 运行单个文件
kylix run hello.klx
# 输出：Hello, World!
```

---

### kylix check - 语法检查

检查 Kylix 代码的语法错误，不生成代码。

**语法：**
```bash
# 检查当前项目所有文件
kylix check

# 检查指定文件
kylix check <file1.klx> <file2.klx>
```

**示例：**
```bash
kylix check
# 输出：
# ✓ main.klx
# ✓ utils.klx
# All files OK
```

**错误示例：**
```bash
kylix check error.klx
# 输出：
# ✗ error.klx:5:3: expected ';' but found 'end'
#   end.
#   ^
```

---

### kylix fmt - 代码格式化

格式化 Kylix 源代码（基础格式化）。

**语法：**
```bash
# 格式化当前项目所有文件
kylix fmt

# 格式化指定文件
kylix fmt <file1.klx> <file2.klx>
```

**示例：**
```bash
kylix fmt
# 输出：
# ✓ main.klx
# Formatted 1 file(s)
```

---

### kylix repl - 交互式解释器

启动 Kylix 交互式 REPL 环境。

**语法：**
```bash
kylix repl
```

**使用方法：**
```
Kylix REPL v0.2.0
Type Kylix code (press Enter twice to execute, Ctrl+D to exit)

kylix> WriteLn('Hello!')
Hello!

kylix> var x := 10
...    var y := 20
...    WriteLn(x + y)
30

kylix> ^D
```

**特性：**
- 单行语句直接执行
- 多行代码按两次 Enter 执行
- 支持 `var` 声明、`WriteLn` 等内置函数
- 自动包装为完整的 program

---

### kylix lsp - 语言服务器

启动 LSP (Language Server Protocol) 服务器，用于编辑器集成。

**语法：**
```bash
kylix lsp
```

**说明：**
此命令通常由编辑器插件自动调用，不需要手动运行。

---

### kylix version - 版本信息

显示 Kylix 编译器版本。

**语法：**
```bash
kylix version
kylix -v
kylix --version
```

---

### kylix help - 帮助信息

显示帮助信息。

**语法：**
```bash
kylix help
kylix -h
kylix --help
```

---

## 项目结构

### kylix.toml 配置文件

每个 Kylix 项目都有一个 `kylix.toml` 配置文件：

```toml
# Kylix project configuration

[project]
name = "myproject"        # 项目名称
version = "0.1.0"         # 版本号
main = "main.klx"         # 主程序入口

[build]
output = "build/"         # 编译输出目录
go_module = "myproject"   # Go 模块名
```

### 标准目录结构

```
myproject/
├── kylix.toml          # 项目配置
├── main.klx            # 主程序入口
├── src/                # 源代码目录（可选）
│   ├── utils.klx
│   └── types.klx
├── build/              # 编译输出（自动生成）
│   └── myproject.go
└── .gitignore          # Git 忽略规则
```

---

## 编辑器集成

### VS Code

#### 安装扩展

1. 进入 `vscode-ext` 目录：
```bash
cd vscode-ext
```

2. 安装依赖：
```bash
npm install
```

3. 在 VS Code 中：
   - 按 `F5` 启动扩展开发主机
   - 或者打包安装：
   ```bash
   npm install -g vsce
   vsce package
   code --install-extension kylix-vscode-0.2.0.vsix
   ```

#### 功能特性

- **语法高亮**：关键字、字符串、数字、注释等
- **语法检查**：实时显示语法错误
- **代码补全**：关键字和内置函数
- **悬停提示**：函数和类型的文档
- **跳转定义**：跳转到函数定义（基础支持）

#### 配置选项

在 VS Code 设置中搜索 "kylix"：

- `kylix.compiler.path` - Kylix 编译器路径（默认：`kylix`）
- `kylix.lsp.enabled` - 是否启用 LSP（默认：`true`）

### 其他编辑器

Kylix LSP 服务器支持标准的 LSP 协议，可以集成到任何支持 LSP 的编辑器：

- **Neovim**：使用 `nvim-lspconfig`
- **Emacs**：使用 `lsp-mode`
- **Sublime Text**：使用 `LSP` 插件

配置示例（通用）：
```json
{
  "command": ["kylix", "lsp"],
  "filetypes": ["kylix"]
}
```

---

## 常见问题

### Q: 编译时报错 "cannot read kylix.toml"

**A:** 确保在项目根目录下运行命令，且 `kylix.toml` 文件存在。

```bash
# 检查当前目录
pwd

# 查看项目文件
ls -la
```

### Q: 运行时报错 "undefined: result"

**A:** 函数返回值使用 `result` 关键字，不需要额外声明：

```pascal
function Add(a: Integer; b: Integer): Integer;
begin
  result := a + b;  // ✓ 正确
end;
```

### Q: 如何在函数中返回多个值？

**A:** Kylix 目前不支持多返回值，可以使用记录类型：

```pascal
type
  TPoint = record
    X, Y: Integer;
  end;

function GetPoint(): TPoint;
begin
  result.X := 10;
  result.Y := 20;
end;
```

### Q: 如何调试程序？

**A:** 由于 Kylix 编译为 Go，可以使用 Go 的调试工具：

```bash
# 保留生成的 .go 文件
kylix run --keep main.klx

# 使用 Go 调试器
go run main.go
# 或者
dlv debug main.go
```

### Q: 如何添加第三方库？

**A:** 在 `build/` 目录下的 `go.mod` 中添加依赖：

```bash
cd build
go get github.com/some/package
```

然后在代码中使用（需要后续版本支持）

### Q: 项目名可以使用中文吗？

**A:** 不可以。项目名必须是有效的 Pascal 标识符，只能包含字母、数字和下划线，且不能以数字开头。

```bash
# ✗ 错误
kylix new 我的项目

# ✓ 正确
kylix new myproject
kylix new my_project
```

### Q: 如何查看所有可用的内置函数？

**A:** 在 VS Code 中输入函数名会自动提示，或查看语言参考文档。常用函数包括：

- `WriteLn(...)` - 输出并换行
- `Write(...)` - 输出
- `ReadLn(...)` - 读取输入
- `Length(str)` - 获取字符串长度
- `IntToStr(n)` - 整数转字符串
- `StrToInt(s)` - 字符串转整数
- `Copy(str, start, length)` - 复制字符串
- `UpperCase(str)` - 转大写
- `LowerCase(str)` - 转小写
- `Sqrt(n)` - 平方根
- `Abs(n)` - 绝对值
- `Round(n)` - 四舍五入

---

## 示例项目

### Hello World

```bash
kylix new hello
cd hello
kylix run
```

**main.klx:**
```pascal
program hello;

begin
  WriteLn('Hello, World!');
end.
```

### 计算器

```bash
kylix new calculator
cd calculator
```

**main.klx:**
```pascal
program calculator;

function Add(a: Integer; b: Integer): Integer;
begin
  result := a + b;
end;

function Multiply(a: Integer; b: Integer): Integer;
begin
  result := a * b;
end;

begin
  var sum := Add(10, 20);
  var product := Multiply(5, 7);
  
  WriteLn('10 + 20 = ', sum);
  WriteLn('5 * 7 = ', product);
end.
```

**运行：**
```bash
kylix run
# 输出：
# 10 + 20 = 30
# 5 * 7 = 35
```

### 数组操作

```pascal
program arrays;

function SumArray(arr: array of Integer): Integer;
var
  i: Integer;
begin
  result := 0;
  for i := 0 to Length(arr) - 1 do
  begin
    result := result + arr[i];
  end;
end;

begin
  var numbers := [1, 2, 3, 4, 5];
  var total := SumArray(numbers);
  WriteLn('Sum: ', total);
end.
```

---

## 下一步

- 阅读 [Kylix 开发指南](KYILIX_DEV_GUIDE.md) 了解如何贡献代码
- 查看 [示例代码](../examples/) 学习更多用法
- 访问项目仓库获取最新版本
