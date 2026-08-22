# Kylix JetBrains Plugin — 安装与使用手册

Kylix 是现代化 Pascal 语言（Modern Pascal → Go / LLVM）。本插件为 **IntelliJ IDEA / GoLand** 提供 Kylix 语言支持：语法高亮 + LSP 智能编辑 + 25 个代码模板。

- **语法高亮**：TextMate 语法（与 VS Code 插件同一 grammar）
- **LSP 集成**：桥接 `kylix lsp`，提供补全 / 跳转 / 引用 / 重命名 / 格式化 / 悬停文档 / 错误波浪线
- **代码模板**：`prog` / `func` / `class` / `controller` 等 25 个缩写
- **结构感知**：`begin`/`end` 括号配对与代码折叠

---

## 1. 环境要求

| 依赖 | 版本 | 说明 |
|---|---|---|
| IntelliJ IDEA 或 GoLand | **2023.3+**（Community/Ultimate 均可） | LSP4IJ 要求 |
| [LSP4IJ](https://plugins.jetbrains.com/plugin/23057) 插件 | 任意稳定版 | **必须先安装**（见下） |
| JDK | 17+ | 构建插件用 21（本机已装） |
| Kylix 编译器 | 任意（含 `kylix lsp`） | 见「配置 Kylix 编译器」 |

> LSP4IJ 是 JetBrains Marketplace 上的开源插件（Red Hat 维护），提供 IntelliJ 平台的 LSP 客户端框架；本插件依赖它启动 `kylix lsp` 进程。安装后，**LSP 的补全/跳转/重命名/格式化等全部能力由 `kylix lsp` 服务器提供**（`pkg/lsp/`，随 Kylix 编译器分发）。

---

## 2. 安装

### 2.1 先装 LSP4IJ

`Settings (⌘,) → Plugins → Marketplace`，搜索 **`LSP4IJ`**，Install → 重启 IDE。

### 2.2 构建插件

```bash
cd jetbrains-plugin
./gradlew buildPlugin
```

> 首次构建会下载 Gradle 发行版 + IntelliJ Platform SDK + LSP4IJ 依赖（数百 MB，需要网络）。产物在 `build/distributions/Kylix-0.1.0.zip`。

### 2.3 安装到 IDE

1. `Settings (⌘,) → Plugins → ⚙（齿轮）→ Install Plugin from Disk…`
2. 选择 `build/distributions/Kylix-0.1.0.zip`
3. OK → 重启 IDE

### 2.4（可选）开发模式运行

```bash
./gradlew runIde   # 启动一个带本插件的全新 IDE 实例
```

---

## 3. 使用

### 3.1 语法高亮

打开任意 `.klx` 文件即自动高亮（TextMate grammar `source.kylix`）：关键字、类型、注解 `[...]`、字符串、注释（`//` `{ }` `(* *)`）、数字、运算符。

### 3.2 LSP 智能编辑

编辑器右下角状态栏出现 **Kylix Language Server** 即已连接。能力：

| 功能 | 快捷键 |
|---|---|
| 代码补全（关键字/内建函数/类型/文档符号） | `Ctrl+Space` / `⌘Space` |
| 跳转到定义 | `Ctrl+B` / `⌘B` |
| 查找引用 | `Ctrl+Shift+F7` / `⌥⌘F7` |
| 重命名 | `Shift+F6` |
| 格式化文档 | `Ctrl+Alt+L` / `⌥⌘L` |
| 悬停文档（WriteLn、类型、关键字说明） | 鼠标悬停 |
| 签名帮助（`(`、`,` 触发） | 输入时自动 |
| 错误/警告波浪线 | 实时诊断 |
| 文档结构视图 | `⌘7` / Structure |

### 3.3 代码模板（Live Templates）

输入缩写后按 `Tab` 展开。在 `.klx` 文件里可用：

| 缩写 | 生成 |
|---|---|
| `prog` | program 骨架（uses + begin/end） |
| `func` / `proc` | 函数 / 过程声明 |
| `class` / `record` | 类 / 记录声明 |
| `if` / `ifelse` | 条件 |
| `for` / `while` / `repeat` / `foreach` | 循环 |
| `try` / `tryfinally` / `tryon` | 异常处理 |
| `var` / `const` / `array` / `map` | 声明 |
| `writeln` / `readln` | I/O |
| `controller` | KylixBoot 控制器（含 [Controller]/[Get] 注解） |
| `routeget` / `routepost` | 路由方法 |
| `entity` | ORM 实体（[Entity]/[Column] 注解） |
| `unit` | 单元骨架（interface/implementation） |

在 `Settings → Editor → Live Templates → Kylix` 里可以查看 / 修改 / 新建模板。

### 3.4 代码折叠与括号配对

- `begin`/`end` 区段自动折叠（`⌘-`/`⌘+`）
- `begin` 与 `end`、`{` `}`、`(` `)`、`[` `]`、引号自动配对与高亮

---

## 4. 配置 Kylix 编译器

LSP 服务器通过启动 `kylix lsp` 进程工作。插件按以下顺序解析编译器路径：

1. 环境变量 **`KYLIX_PATH`**（指向 `kylix` 可执行文件）
2. `PATH` 中的 `kylix`
3. 命令 `kylix`

**stdlib 声明补全**：LSP 服务器从 `stdlib/klx/*.klx` 加载模块符号，查找顺序：

1. 环境变量 **`KYLIX_HOME`** → `$KYLIX_HOME/stdlib/klx`
2. `kylix` 可执行文件所在目录向上 5 层内的 `stdlib/klx`（本仓库 `kylix` 在根目录、`stdlib/klx` 也在根目录，天然命中）

```bash
# 推荐：把 Kylix 编译器的 bin 目录加进 PATH
export PATH="$HOME/kylix:$PATH"
# 或显式指定
export KYLIX_PATH="$HOME/kylix/kylix"
export KYLIX_HOME="$HOME/kylix"
```

---

## 5. 故障排查

| 现象 | 处理 |
|---|---|
| 无语法高亮 | 确认 TextMate 插件已启用（IDE 内置，一般无需操作）；重启 IDE |
| LSP 状态栏不出现 / 无补全 | ① 确认 `kylix lsp` 命令行可用（终端跑 `kylix lsp` 不报错）② 确认 LSP4IJ 已安装 ③ 打开 `LSP Consoles` 工具窗口看 kylix-lsp 日志 |
| 补全没有 stdlib 符号 | 设 `KYLIX_HOME`（或保证 kylix 可执行文件与 `stdlib/klx` 相对位置在 5 层内）后重启 IDE |
| `kylix: command not found` | 把 Kylix 编译器加进 `PATH`，或设 `KYLIX_PATH` |
| 安装 zip 提示版本不兼容 | 本插件要求 IDE 2023.3+；升级 IDE 或换对应版本的 SDK 重新构建 |

---

## 6. 开发

- **构建**：`./gradlew buildPlugin`（产物 `build/distributions/Kylix-*.zip`）
- **运行**：`./gradlew runIde`
- **测试**：`./gradlew test`
- **语法**：语法高亮 grammar 复用于 `vscode-ext/syntaxes/kylix.tmLanguage.json`（本目录为带 `fileTypes` 的副本）
- **LSP**：服务器为 `kylix lsp`（源码 `pkg/lsp/`，独立于本插件，两端零耦合）

---

**License**: MIT · 与 Kylix 主仓库一致。
