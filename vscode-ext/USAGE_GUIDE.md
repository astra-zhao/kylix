# Kylix VS Code 扩展使用指南

## 目录
1. [快速开始](#快速开始)
2. [语法高亮](#语法高亮)
3. [代码片段](#代码片段)
4. [智能补全](#智能补全)
5. [悬停信息](#悬停信息)
6. [LSP 功能](#lsp-功能)
7. [命令](#命令)
8. [快捷键](#快捷键)
9. [配置](#配置)
10. [故障排除](#故障排除)

## 快速开始

### 1. 安装扩展
参见 [INSTALL.md](INSTALL.md)

### 2. 创建第一个程序
1. 创建新文件：`File` → `New File`
2. 保存为 `hello.klx`
3. 输入 `prog` 然后按 `Tab`
4. 填写程序名称
5. 开始编写代码

### 3. 运行程序
1. 打开命令面板：`Ctrl+Shift+P`
2. 输入 `Kylix: Run`
3. 按 `Enter`

## 语法高亮

### 支持的语言元素

#### 关键字
```pascal
program, unit, uses, var, const, type, begin, end
procedure, function, class, record, array
if, then, else, while, do, for, to, repeat, until
try, except, finally, raise
and, or, not, xor, div, mod
```

#### 类型
```pascal
Integer, String, Boolean, Real, Char
TObject, TClass, TComponent
array of Integer, set of Char
```

#### 字符串
```pascal
'Single quoted string'
"Double quoted string"
"String with ${interpolation}"
'String with \'escape\' sequences'
```

#### 注释
```pascal
// Line comment
{ Block comment }
```

#### 数字
```pascal
42
3.14
1.5e-10
```

#### 匿名函数
```pascal
procedure()
begin
  // anonymous procedure
end

function(x: integer): integer
begin
  result := x * 2;
end
```

### 颜色方案

不同元素使用不同颜色：
- 🔵 **蓝色**: 关键字
- 🟢 **绿色**: 字符串
- 🟡 **黄色**: 数字
- 🟣 **紫色**: 类型
- 🔴 **红色**: 注释
- 🟠 **橙色**: 函数名
- ⚪ **白色**: 普通文本

## 代码片段

### 使用方法
1. 输入片段前缀（如 `proc`）
2. 按 `Tab` 键
3. 使用 `Tab` 在占位符之间跳转
4. 使用 `Shift+Tab` 向后跳转

### 常用片段示例

#### 程序模板
```
prog + Tab
```
生成：
```pascal
program ProgramName;

uses sysutils;

begin
  
end.
```

#### 过程定义
```
proc + Tab
```
生成：
```pascal
procedure ProcedureName(params);
begin
  
end;
```

#### 匿名函数
```
anonfunc + Tab
```
生成：
```pascal
function(params): ReturnType
begin
  
end
```

#### Web 服务器
```
web + Tab
```
生成：
```pascal
program WebServer;

uses web;

begin
  var app := web.createServer(8080);
  
  app.get('/', procedure()
  begin
    res.send('Hello from Kylix Web!');
  end);
  
  app.listen();
end.
```

### 完整片段列表

参见 [snippets/kylix.json](snippets/kylix.json)

## 智能补全

### 触发补全
- 自动触发：输入 `.` 或 `:`
- 手动触发：`Ctrl+Space`

### 补全类型

#### 内置函数
输入 `write` 然后按 `Ctrl+Space`：
```
writeln  - Write a line to console
readln   - Read a line from console
write    - Write to console
```

#### 类型
输入 `Int` 然后按 `Ctrl+Space`：
```
Integer    - 32-bit signed integer
Int64      - 64-bit signed integer
IInterface - Interface type
```

#### 关键字
输入 `beg` 然后按 `Ctrl+Space`：
```
begin  - Begin block
```

### 参数提示
输入函数名后自动显示参数信息：
```pascal
writeln(|)
        ↑
        显示：writeln(value1, value2, ...)
```

## 悬停信息

### 使用方法
将鼠标悬停在符号上

### 支持的符号

#### 内置函数
悬停在 `writeln` 上：
```
**writeln** - Write a line to console

writeln(value1, value2, ...);
```

#### 类型
悬停在 `Integer` 上：
```
**Integer** - 32-bit signed integer

Range: -2147483648 to 2147483647
```

#### 关键字
悬停在 `begin` 上：
```
**begin** - Begin a block

Used with: end
```

## LSP 功能

### 跳转到定义
- 快捷键：`F12`
- 右键菜单：`Go to Definition`
- 按住 `Ctrl` 点击符号

### 查找引用
- 快捷键：`Shift+F12`
- 右键菜单：`Find All References`

### 重命名符号
- 快捷键：`F2`
- 右键菜单：`Rename Symbol`

### 文档符号
- 快捷键：`Ctrl+Shift+O`
- 显示当前文件的所有符号

### 诊断信息
- 错误：红色波浪线
- 警告：黄色波浪线
- 信息：蓝色波浪线

悬停在波浪线上查看详细信息。

## 命令

### 可用命令

#### Kylix: Compile
编译当前文件，不运行。

**使用方法：**
1. 打开命令面板：`Ctrl+Shift+P`
2. 输入 `Kylix: Compile`
3. 按 `Enter`

**输出：**
在终端中显示编译结果。

#### Kylix: Run
编译并运行当前文件。

**使用方法：**
1. 打开命令面板：`Ctrl+Shift+P`
2. 输入 `Kylix: Run`
3. 按 `Enter`

**输出：**
在新终端中运行程序。

#### Kylix: Check
检查当前文件的语法，不生成代码。

**使用方法：**
1. 打开命令面板：`Ctrl+Shift+P`
2. 输入 `Kylix: Check`
3. 按 `Enter`

**输出：**
在终端中显示语法检查结果。

## 快捷键

### 默认快捷键

| 功能 | Windows/Linux | macOS |
|------|---------------|-------|
| 命令面板 | `Ctrl+Shift+P` | `Cmd+Shift+P` |
| 触发补全 | `Ctrl+Space` | `Ctrl+Space` |
| 切换注释 | `Ctrl+/` | `Cmd+/` |
| 跳转到定义 | `F12` | `F12` |
| 查找引用 | `Shift+F12` | `Shift+F12` |
| 重命名符号 | `F2` | `F2` |
| 文档符号 | `Ctrl+Shift+O` | `Cmd+Shift+O` |
| 代码片段 | `Tab` | `Tab` |

### 自定义快捷键

1. 打开键盘快捷方式：`Ctrl+K Ctrl+S`
2. 搜索 `kylix`
3. 点击要修改的快捷键
4. 按下新的快捷键组合

## 配置

### 访问设置
1. 打开设置：`Ctrl+,`
2. 搜索 `kylix`

### 可用设置

#### kylix.format.enable
启用/禁用代码格式化

**类型：** `boolean`  
**默认值：** `true`

```json
{
  "kylix.format.enable": true
}
```

#### kylix.lint.enable
启用/禁用代码检查

**类型：** `boolean`  
**默认值：** `true`

```json
{
  "kylix.lint.enable": true
}
```

#### kylix.compiler.path
Kylix 编译器路径

**类型：** `string`  
**默认值：** `"kylix"`

```json
{
  "kylix.compiler.path": "/usr/local/bin/kylix"
}
```

#### kylix.completion.enable
启用/禁用智能补全

**类型：** `boolean`  
**默认值：** `true`

```json
{
  "kylix.completion.enable": true
}
```

#### kylix.hover.enable
启用/禁用悬停信息

**类型：** `boolean`  
**默认值：** `true`

```json
{
  "kylix.hover.enable": true
}
```

### 工作区设置

在项目根目录创建 `.vscode/settings.json`：

```json
{
  "kylix.compiler.path": "./kylix",
  "kylix.format.enable": true,
  "kylix.completion.enable": true
}
```

## 故障排除

### 语法高亮不工作

**症状：** 代码没有颜色

**解决方案：**
1. 确保文件扩展名是 `.klx`
2. 在右下角检查语言模式
3. 点击语言模式，选择 "Kylix"
4. 重新加载窗口：`Ctrl+Shift+P` → `Developer: Reload Window`

### 补全不工作

**症状：** 输入后没有补全建议

**解决方案：**
1. 检查设置 `kylix.completion.enable` 是否为 `true`
2. 尝试手动触发：`Ctrl+Space`
3. 确保编译器路径正确
4. 查看输出面板：`Ctrl+Shift+U` → 选择 "Kylix Language Server"

### LSP 功能不工作

**症状：** 跳转到定义、查找引用等功能不可用

**解决方案：**
1. 确保 Kylix 编译器已安装
   ```bash
   kylix --version
   ```
2. 确保编译器支持 `lsp` 命令
   ```bash
   kylix lsp --help
   ```
3. 检查编译器路径设置
4. 重启 VS Code

### 代码片段不工作

**症状：** 输入前缀后按 `Tab` 没有反应

**解决方案：**
1. 确保在正确的上下文中输入
   - `proc` 需要在声明位置
   - `if` 需要在语句位置
2. 检查是否有语法错误
3. 重新加载窗口

### 命令不可用

**症状：** 命令面板中找不到 Kylix 命令

**解决方案：**
1. 确保扩展已激活
2. 打开一个 `.klx` 文件
3. 重新加载窗口

### 性能问题

**症状：** VS Code 运行缓慢

**解决方案：**
1. 禁用不必要的功能
   ```json
   {
     "kylix.completion.enable": false,
     "kylix.hover.enable": false
   }
   ```
2. 排除大文件
   ```json
   {
     "files.watcherExclude": {
       "**/large-files/**": true
     }
   }
   ```
3. 关闭其他扩展

### 查看日志

1. 打开输出面板：`Ctrl+Shift+U`
2. 选择 "Kylix Language Server"
3. 查看日志信息

### 报告问题

如果问题仍然存在：
1. 收集日志信息
2. 创建最小复现代码
3. 提交 Issue：https://github.com/your-repo/kylix/issues

## 最佳实践

### 1. 使用代码片段
- 减少重复输入
- 避免语法错误
- 保持代码风格一致

### 2. 利用智能补全
- 快速输入内置函数
- 避免拼写错误
- 学习可用的 API

### 3. 查看悬停信息
- 了解函数用法
- 查看参数说明
- 学习最佳实践

### 4. 使用 LSP 功能
- 快速导航代码
- 查找所有引用
- 安全重命名

### 5. 配置工作区设置
- 为项目定制设置
- 团队共享配置
- 保持一致性

## 进阶技巧

### 多光标编辑
- `Alt+Click` 添加光标
- `Ctrl+Alt+↑/↓` 垂直添加光标
- `Ctrl+D` 选择下一个相同单词

### 快速导航
- `Ctrl+P` 快速打开文件
- `Ctrl+G` 跳转到行号
- `Ctrl+Shift+O` 跳转到符号

### 代码折叠
- `Ctrl+Shift+[` 折叠代码块
- `Ctrl+Shift+]` 展开代码块
- `Ctrl+K Ctrl+0` 折叠所有
- `Ctrl+K Ctrl+J` 展开所有

### 搜索和替换
- `Ctrl+F` 搜索
- `Ctrl+H` 替换
- `Ctrl+Shift+F` 全局搜索
- `Ctrl+Shift+H` 全局替换

## 资源

- 📚 [Kylix 文档](https://kylix-lang.org/docs)
- 💬 [Discord 社区](https://discord.gg/kylix)
- 🐛 [问题跟踪](https://github.com/your-repo/kylix/issues)
- 📖 [VS Code 文档](https://code.visualstudio.com/docs)
