# Kylix VS Code Extension

Kylix 编程语言的 Visual Studio Code 扩展，提供完整的语言支持。

## 功能特性

### ✨ 语法高亮
- 关键字分类高亮（控制语句、声明、类型）
- 字符串高亮（单引号、双引号）
- 字符串插值高亮 (`${variable}`)
- 数字、注释、函数名、类型高亮
- 操作符高亮（赋值、比较、算术、逻辑）
- **匿名过程/函数高亮** (`procedure()`, `function()`)

### 📝 代码片段 (25+)
快速输入代码模板：
- `prog` - 程序模板
- `proc` - 过程定义
- `func` - 函数定义
- `anonproc` - 匿名过程
- `anonfunc` - 匿名函数
- `if` / `ifelse` - 条件语句
- `for` / `while` / `repeat` - 循环语句
- `try` / `tryfinally` - 异常处理
- `class` / `record` - 类型定义
- `web` / `route` / `routepost` / `routeparam` - Web 开发
- `writeln` / `readln` - I/O 操作

### 🧠 智能补全
- 40+ 内置函数补全（带参数提示）
- 50+ 类型补全
- 所有关键字补全
- 上下文感知

### 💡 悬停信息
- 内置函数文档
- 使用示例
- 参数说明

### 🔧 语言配置
- 括号匹配：`{}`, `[]`, `()`
- 代码折叠：`begin...end`, `procedure`, `function`, `class`, `if`, `while`, `for`, `repeat`
- 注释切换：`Ctrl+/`
- 自动缩进规则
- 自动闭合：引号、括号

### 🚀 命令
- `Kylix: Compile` - 编译当前文件
- `Kylix: Run` - 编译并运行当前文件
- `Kylix: Check` - 检查当前文件语法

### 🌐 LSP 支持
- 跳转到定义
- 查找引用
- 重命名符号
- 文档符号
- 诊断信息

## 安装

详见 [INSTALL.md](INSTALL.md)

## 使用指南

详见 [USAGE_GUIDE.md](USAGE_GUIDE.md)

## 更新日志

详见 [CHANGELOG.md](CHANGELOG.md)

## 测试

打开 `test-syntax.klx` 文件测试所有语法高亮功能。

## 配置

在 VS Code 设置中搜索 `kylix`：

- `kylix.format.enable` - 启用/禁用代码格式化（默认：true）
- `kylix.lint.enable` - 启用/禁用代码检查（默认：true）
- `kylix.compiler.path` - Kylix 编译器路径（默认：`kylix`）
- `kylix.completion.enable` - 启用/禁用智能补全（默认：true）
- `kylix.hover.enable` - 启用/禁用悬停信息（默认：true）

## 开发

### 项目结构
```
vscode-ext/
├── package.json              # 扩展清单
├── extension.js              # 扩展入口
├── language-configuration.json  # 语言配置
├── syntaxes/
│   └── kylix.tmLanguage.json   # 语法定义
├── snippets/
│   └── kylix.json              # 代码片段
└── README.md                   # 本文档
```

### 开发模式
1. 克隆仓库
2. 在 VS Code 中打开 `vscode-ext` 目录
3. 按 `F5` 启动扩展开发主机

### 打包
```bash
npm install -g @vscode/vsce
vsce package
```

## 依赖

- `vscode-languageclient`: ^8.0.0
- `@types/vscode`: ^1.60.0
- `@types/node`: ^16.0.0

## 许可证

MIT License

## 相关链接

- [Kylix 编译器](../README.md)
- [Kylix Web 框架](../docs/WEB_FRAMEWORK.md)
- [匿名函数指南](../docs/ANONYMOUS_FUNCTIONS.md)
- [VS Code 扩展开发](https://code.visualstudio.com/api)
