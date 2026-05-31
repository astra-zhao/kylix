# Kylix VS Code 扩展安装指南

## 方法一：开发模式安装（推荐用于开发）

### 1. 打开 VS Code
```bash
code /Users/astra/Documents/ai/learn/kylix/vscode-ext
```

### 2. 安装依赖
在终端中运行：
```bash
cd /Users/astra/Documents/ai/learn/kylix/vscode-ext
npm install
```

### 3. 启动扩展开发主机
- 在 VS Code 中按 `F5`
- 会打开一个新的 VS Code 窗口（扩展开发主机）

### 4. 测试扩展
在新窗口中：
- 打开任意 `.klx` 文件
- 验证语法高亮、代码片段等功能

## 方法二：打包安装（用于正式发布）

### 1. 安装 vsce 工具
```bash
npm install -g @vscode/vsce
```

### 2. 打包扩展
```bash
cd /Users/astra/Documents/ai/learn/kylix/vscode-ext
vsce package
```
这会生成 `kylix-1.0.0.vsix` 文件

### 3. 安装扩展
```bash
code --install-extension kylix-1.0.0.vsix
```

### 4. 重新加载 VS Code
- 按 `Ctrl+Shift+P` (macOS: `Cmd+Shift+P`)
- 输入 `Developer: Reload Window`
- 按 Enter

## 方法三：从 VS Code 界面安装

### 1. 打开 VS Code

### 2. 打开命令面板
按 `Ctrl+Shift+P` (macOS: `Cmd+Shift+P`)

### 3. 选择安装命令
输入并选择：
```
Extensions: Install from VSIX...
```

### 4. 选择文件
- 浏览到 `kylix-1.0.0.vsix`
- 点击 "Install"

### 5. 重新加载窗口
按 `Ctrl+Shift+P` → `Developer: Reload Window`

## 验证安装

### 1. 创建测试文件
```bash
cat > test.klx << 'EOF'
program Test;
begin
  writeln('Hello, Kylix!');
end.
EOF
```

### 2. 在 VS Code 中打开
```bash
code test.klx
```

### 3. 检查功能
- ✅ 语法高亮（关键字、字符串、注释）
- ✅ 输入 `prog` 然后按 Tab（代码片段）
- ✅ 输入 `write` 然后按 Ctrl+Space（补全）
- ✅ 将鼠标悬停在 `writeln` 上（悬停信息）

## 卸载扩展

### 1. 打开扩展面板
按 `Ctrl+Shift+X`

### 2. 找到 Kylix 扩展

### 3. 点击卸载按钮

### 4. 重新加载窗口

## 常见问题

### Q: 扩展没有激活？
**A:** 确保文件扩展名是 `.klx`，或者在右下角手动选择 "Kylix" 语言模式

### Q: 代码片段不工作？
**A:** 检查是否在正确的上下文中输入前缀，例如 `proc` 需要在声明位置使用

### Q: LSP 功能不工作？
**A:** 确保 Kylix 编译器已安装并在 PATH 中，运行 `kylix --version` 验证

### Q: 语法高亮颜色不对？
**A:** 这取决于你的 VS Code 主题，可以尝试切换主题

## 更新扩展

### 1. 卸载旧版本

### 2. 安装新版本（使用上述任一方法）

### 3. 重新加载窗口

## 开发模式调试

如果需要调试扩展：

### 1. 打开扩展项目
```bash
code /Users/astra/Documents/ai/learn/kylix/vscode-ext
```

### 2. 设置断点
- 在 `extension.js` 中设置断点

### 3. 按 F5 启动调试
- 会打开一个新的 VS Code 窗口
- 在新窗口中使用扩展
- 原窗口会停在断点处

### 4. 查看调试控制台
- 在原窗口查看变量、调用栈等

## 性能优化

如果扩展运行缓慢：

### 1. 禁用不必要的功能
```json
{
  "kylix.completion.enable": false,
  "kylix.hover.enable": false
}
```

### 2. 排除大文件
```json
{
  "files.exclude": {
    "**/*.klx": false
  },
  "files.watcherExclude": {
    "**/large-files/**": true
  }
}
```

## 联系支持

- 📧 Email: support@kylix-lang.org
- 🐛 Issues: https://github.com/your-repo/kylix/issues
- 💬 Discord: https://discord.gg/kylix
- 📚 Documentation: https://kylix-lang.org/docs
