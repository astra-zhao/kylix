# 更新日志

## [1.1.0] - 2026-06-30

### ✨ 新增功能

#### 语法高亮增强
- ✅ **KylixBoot 注解高亮**：`[Controller]`/`[Get]`/`[Post]`/`[Body]`/`[Authenticated]`/`[Entity]`/`[Required]` 等注解中的名称、字符串参数、数字参数独立着色
- ✅ **stdlib 函数高亮**：新增 Phase 6/7 标准库函数识别（`JwtSign`/`DbOpen`/`NewCache`/`HttpGet`/`Sha256`/`BootRun` 等 30+ 函数）

#### 命令与快捷键
- ✅ `Kylix: Compile File` 命令（`Ctrl+Shift+B` / `Cmd+Shift+B`）——在集成终端编译当前 `.klx`
- ✅ `Kylix: Run File` 命令（`F5`）——编译并运行当前文件
- ✅ 编辑器标题栏运行按钮 + 右键菜单（编译/运行）

#### 编译器路径解析
- ✅ `kylix.compiler.path` 配置项优先
- ✅ `KYLIX_PATH` 环境变量回退
- ✅ 默认 `kylix`（假定在 PATH 上）

#### 状态栏指示
- ✅ 状态栏显示 LSP 连接状态（启动中 / 就绪 / 失败）
- ✅ 启动失败时给出明确的路径配置提示

### 🐛 修复
- 修复 LSP 服务器重复启动 bug：旧版本手动 `spawn` 进程后 `LanguageClient` 又启动了第二个，现在由 `LanguageClient` 统一管理进程生命周期
- 修复文件路径含空格时编译命令出错（路径加引号转义）

---

## [1.0.0] - 2026-05-31

### ✨ 新增功能

#### 语法高亮
- ✅ 完整的 Kylix 语法高亮支持
- ✅ 关键字分类高亮（控制语句、声明、类型等）
- ✅ 字符串高亮（单引号、双引号）
- ✅ **字符串插值高亮** (`${variable}`)
- ✅ 数字高亮
- ✅ 注释高亮（行注释 `//` 和块注释 `{}`）
- ✅ 函数名高亮
- ✅ 类型高亮（50+ 内置类型）
- ✅ 操作符高亮
- ✅ **匿名过程/函数高亮** (`procedure()`, `function()`)

#### 代码片段
新增 25+ 个代码片段：
- `prog` - 程序模板
- `proc` - 过程定义
- `func` - 函数定义
- `anonproc` - 匿名过程
- `anonfunc` - 匿名函数
- `if` - if 语句
- `ifelse` - if-else 语句
- `for` - for 循环
- `while` - while 循环
- `repeat` - repeat-until 循环
- `try` - try-except 块
- `tryfinally` - try-finally 块
- `class` - 类定义
- `record` - 记录定义
- `web` - Web 服务器模板
- `route` - GET 路由处理器
- `routepost` - POST 路由处理器
- `routeparam` - 带参数的路由
- `var` - 变量声明
- `const` - 常量声明
- `array` - 数组声明
- `writeln` - 输出到控制台
- `readln` - 从控制台读取

#### 智能补全
- ✅ 40+ 内置函数补全
  - I/O: `writeln`, `readln`, `write`, `read`
  - 字符串: `length`, `copy`, `delete`, `insert`, `pos`, `uppercase`, `lowercase`, `trim`
  - 转换: `inttostr`, `strtoint`, `floattostr`, `strtofloat`, `str`, `val`
  - 格式化: `format`
  - 数学: `abs`, `sqr`, `sqrt`, `sin`, `cos`, `tan`, `arctan`, `ln`, `exp`, `round`, `trunc`, `random`
  - 数组: `high`, `low`, `inc`, `dec`, `succ`, `pred`, `ord`, `chr`
  - 对象: `assigned`, `free`, `create`
- ✅ 50+ 类型补全
- ✅ 所有关键字补全
- ✅ 参数提示

#### 悬停信息
- ✅ 内置函数文档
- ✅ 使用示例
- ✅ Markdown 格式

#### 语言配置
- ✅ 括号匹配：`{}`, `[]`, `()`
- ✅ 代码折叠：`begin...end`, `procedure`, `function`, `class`, `if`, `while`, `for`, `repeat`
- ✅ 注释切换：`Ctrl+/`
- ✅ 自动缩进规则
- ✅ 自动闭合：引号、括号
- ✅ 单词模式

#### 命令
- ✅ `Kylix: Compile` - 编译当前文件
- ✅ `Kylix: Run` - 编译并运行当前文件
- ✅ `Kylix: Check` - 检查当前文件语法

#### LSP 集成
- ✅ 语言服务器客户端
- ✅ 跳转到定义
- ✅ 查找引用
- ✅ 重命名符号
- ✅ 文档符号
- ✅ 诊断信息

#### 配置选项
- ✅ `kylix.format.enable` - 启用/禁用代码格式化
- ✅ `kylix.lint.enable` - 启用/禁用代码检查
- ✅ `kylix.compiler.path` - Kylix 编译器路径
- ✅ `kylix.completion.enable` - 启用/禁用智能补全
- ✅ `kylix.hover.enable` - 启用/禁用悬停信息

### 📚 文档
- ✅ README.md - 完整的扩展文档
- ✅ INSTALL.md - 快速安装指南
- ✅ USAGE_GUIDE.md - 详细使用指南
- ✅ CHANGELOG.md - 更新日志（本文件）
- ✅ test-syntax.klx - 语法高亮测试文件

### 🔧 技术细节
- 基于 TextMate 语法定义
- 使用 vscode-languageclient 8.0.0
- 支持 VS Code 1.60.0+
- 完整的 TypeScript 类型定义

### 🎨 主题支持
- 支持所有 VS Code 主题
- 推荐主题：One Dark Pro, GitHub Theme, Dracula

### 🐛 已知问题
- 格式化功能尚未完全实现
- 某些复杂的匿名函数语法可能需要改进

### 🔮 未来计划
- [ ] 实现代码格式化
- [ ] 添加更多代码片段
- [ ] 改进类型推断
- [ ] 添加重构支持
- [ ] 添加调试支持
- [ ] 添加单元测试
- [ ] 添加性能分析工具

### 📦 依赖
- `vscode-languageclient`: ^8.0.0
- `@types/vscode`: ^1.60.0
- `@types/node`: ^16.0.0

### 🙏 致谢
- 感谢 Kylix 编译器团队
- 感谢所有测试用户
- 感谢 VS Code 扩展开发社区

---

## [0.1.0] - 2026-05-30

### 初始版本
- 基础语法高亮
- LSP 集成
- 基本代码片段

---

## 版本说明

### 版本号规则
- **主版本号** (X.0.0): 不兼容的 API 更改
- **次版本号** (0.X.0): 向后兼容的功能新增
- **修订号** (0.0.X): 向后兼容的问题修复

### 发布频率
- 主版本：每年 1-2 次
- 次版本：每月 1-2 次
- 修订版：每周（如有需要）

### 支持政策
- 当前版本：完整支持
- 前一个次版本：安全修复
- 更早版本：不再支持
