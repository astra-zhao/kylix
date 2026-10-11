# 用 Kylix 做 Android 和 iOS（零基础教程）

这篇教程带你做出一个很小的「背单词」应用：六个英文词，能浏览、能点「认识」、能只复习还没掌握的。

你不需要会 Pascal，也不需要先读编译器内部文档。会复制命令、会看终端里的字就可以。

下面的步骤仍然按脚本写。这些脚本现在会去调用 `kylix mobile`（doctor / check / serve / build）。命令说明在 [MOBILE_COMMANDS.md](MOBILE_COMMANDS.md)。这一篇等 v0.16.0 发版后再整篇改写，当前发布版本仍是 v0.15.0。

每一章都按这个顺序写：

1. 这一步在做什么、为什么做。
2. 在哪个文件夹里、用哪种系统。
3. 可以整段复制的命令。
4. 成功时长什么样、怎样算做完。
5. 这一章的常见问题。
6. 自检清单。

文末有术语表。某个词第一次出现时，正文里也会用白话解释一次。

## 哪些步骤是在真机器上跑过的

编写这篇教程时，检查环境是一台 **Linux**（`uname` 显示 `Linux`，bash 5.2，Go 1.25.0，LLVM 18.1.3，NDK r26d 在 `/opt/android-ndk-r26d`）。

| 步骤 | 状态 |
|---|---|
| 第 3 章：`go build`、`./kylix version`、`./kylix doctor` | 已在这台 Linux 上跑通。没有在一台空电脑上重装 Go / LLVM |
| 第 4 章：`bash apps/vocab/host_check.sh` | 已跑通，最后一行是 `[vocab] host check PASS` |
| 第 5 章：`bash apps/vocab/e2e.sh`、`VOCAB_BACKEND=llvm`、留下单词服务、三条 `curl` | 已跑通。服务日志里有 `vocab server on :8091` 和 `KylixBoot started on http://localhost:8091` |
| 第 6 章：Android `amd64` 与 `arm64` 两份 `.so`，以及 `check_artifact.sh` | 已跑通，最后一行分别是 `android artifact OK (amd64)` 和 `android artifact OK (arm64)` |
| 第 8 章：在 Linux 上执行 `apps/vocab/ios/build_core.sh simulator` | 已跑过，它会失败并打印必须在 macOS 上链接。**没有**在 Mac 上编出 `.a`，**没有**打开 Xcode |
| 第 7 章 Android Studio、模拟器点击；第 8 章 Mac 上的模拟器点击 | **没有**在编写环境里点击。下面的按钮名称按 Android Studio / Xcode 的英文界面写 |
| Windows、WSL | **没有**在 Windows 上跑。第 0 章根据脚本实际内容说明：这些 `.sh` 是 bash，Windows 自带的命令提示符跑不了 |

主练习是 `apps/vocab/`。登录和 Notes 在[附录](#10-附录登录和-notes)，那是另一套已经在仓库里的示例。附录里的后台编译命令这次没有重新执行。

---

## 目录

0. [先看你的电脑属于哪一种](#0-先看你的电脑属于哪一种)
1. [做完之后你手里有什么](#1-做完之后你手里有什么)
2. [程序分成几块](#2-程序分成几块)
3. [安装工具，并编出 kylix](#3-安装工具并编出-kylix)
4. [不打开手机，先检查单词核心](#4-不打开手机先检查单词核心)
5. [用 curl 走完三步，并让服务一直开着](#5-用-curl-走完三步并让服务一直开着)
6. [安装 NDK，编出 Android 用的库](#6-安装-ndk编出-android-用的库)
7. [在 Android Studio 里打开，并在模拟器里点](#7-在-android-studio-里打开并在模拟器里点)
8. [iOS：只有 Mac 能编库、能开模拟器](#8-ios只有-mac-能编库能开模拟器)
9. [Export、kylix_free、triple 是什么](#9-exportkylix_freetriple-是什么)
10. [附录：登录和 Notes](#10-附录登录和-notes)
11. [术语表](#11-术语表)

---

## 0. 先看你的电脑属于哪一种

### 这一步在做什么

先确认你后面要开的是哪一种「终端」。终端就是一个可以打字、按回车、让电脑执行命令的窗口。命令是一行交给电脑的文字，例如 `go version`。

仓库是这一整份 Kylix 源代码，克隆下来之后是一个文件夹。后面说的「仓库根」就是这个文件夹，里面能直接看到 `README_CN.md` 和 `apps`。

当前目录是终端「现在站在哪个文件夹」。命令会在当前目录里找文件。站错文件夹，命令就会说找不到文件。

### 三种系统，后面的命令不一样

| 你的电脑 | 这篇教程怎么走 |
|---|---|
| Linux（Ubuntu 等） | 从第 3 章起，命令原样复制。编写时跑通的就是这种 |
| macOS | 第 3 章的安装命令用下面的 macOS 小节。第 4–7 章的 `bash ...` 与 Linux 相同。第 8 章只有 Mac 能做 |
| Windows | 自带的「命令提示符」和 PowerShell **不能**运行 `apps/vocab/*.sh`。这些脚本第一行是 `#!/usr/bin/env bash`，里面用了 `mktemp`、`cmp`、`cc`。请先安装 **WSL2 里的 Ubuntu**，然后在 Ubuntu 窗口里把电脑当成 Linux 来做第 3–6 章 |

**WSL** 是「Windows Subsystem for Linux」：在 Windows 里装一个真正的 Linux 环境。装好之后你会有一个名为 Ubuntu 的窗口，那个窗口才是终端。

这次编写没有打开 Windows。下面的 WSL 步骤来自微软当前的安装方式，以及仓库里脚本的真实写法。

#### Windows：安装 WSL2（在 Windows 里做，只做一次）

1. 在开始菜单里找到 **PowerShell**，右键，选 **以管理员身份运行**。
2. 输入下面这一行，回车：

```powershell
wsl --install
```

3. 按提示重启电脑。
4. 重启后会打开 Ubuntu，让你设一个用户名和密码。密码输入时屏幕上没有圆点，这是正常的，输完回车即可。
5. 以后要做第 3–6 章，都打开开始菜单里的 **Ubuntu**，不要打开「命令提示符」。

在 Ubuntu 里确认：

```bash
uname -s
```

应打印 `Linux`。打印出来才算进入了后面各章要用的环境。

Android Studio 仍然装在 Windows 上（第 7 章）。单词服务如果跑在 WSL 里，Windows 上的模拟器能不能连上，这次没有实测。第 7 章写了怎样用浏览器判断。

### 这一章的常见问题

| 你看到的 | 原因和做法 |
|---|---|
| PowerShell 里 `wsl` 不是内部或外部命令 | Windows 版本太旧，或功能没装上。到微软文档搜索「安装 WSL」，按「Windows 10/11」那一页做完再回到这里 |
| Ubuntu 窗口一闪就关 | 第一次安装没完成。再以管理员打开 PowerShell，执行 `wsl --install -d Ubuntu` |
| 你在命令提示符里粘贴 `bash apps/vocab/host_check.sh`，提示找不到 `bash` | 你还在 Windows 自己的终端里。改开 Ubuntu |

### 自检清单

- [ ] 我知道自己后面要用 Linux、macOS，还是 WSL 里的 Ubuntu
- [ ] 我能打开一个终端，并且知道它不是记事本
- [ ] Windows 用户已经在 Ubuntu 里看到 `uname -s` 打印 `Linux`

---

## 1. 做完之后你手里有什么

### 这一步在做什么

先看终点，避免做到一半不知道自己在装什么。

你会得到三样东西：

1. 电脑上一个小网站，地址是 `http://127.0.0.1:8091`。它发出六张单词卡片。
2. 一个检查：同一份单词逻辑，用两种编译方式打出来的文字完全一样。
3. （如果你有 Android Studio）模拟器里一个标题为「背单词」的应用。点「认识」之后，再点「复习未掌握」，刚才那个词不再出现。

内置的六个词：

| 编号 id | 英文 | 中文 |
|---|---|---|
| 1 | apple | 苹果 |
| 2 | book | 书 |
| 3 | water | 水 |
| 4 | friend | 朋友 |
| 5 | time | 时间 |
| 6 | happy | 高兴 |

**id** 就是这张卡片的编号，用数字表示。应用记住「已经认识」时，存的是编号串，例如 `1,3` 表示认识了 apple 和 water。单词服务不把这串字写进数据库，每次由手机带回去。所以这个例子不需要注册账号。

这不是应用商店里的完整产品。它用来把「一份逻辑、手机只负责画和上网」这条路走通。

### 自检清单

- [ ] 我能说出六个词里第一个是 apple / 苹果
- [ ] 我知道「认识」之后，复习列表里这个词会消失
- [ ] 我知道这个例子不用登录

---

## 2. 程序分成几块

### 这一步在做什么

看清每个文件夹是干什么的。后面打开错文件夹，是最常见的空转。

```text
你写一次的单词规则          谁来用它
apps/vocab/vocab.klx   →   电脑上的单词服务（第 5 章）
                        →   手机里的一份库（第 6、8 章）

手机上的画面和上网          不放进上面那份规则
apps/vocab/android/         Android 的按钮，用 OkHttp 上网
apps/vocab/ios/             iPhone 的按钮，用 URLSession 上网
```

用到的词：

- **HTTP**：手机和电脑之间传文字的一种方式。你在浏览器地址栏打开网页，用的就是它。
- **GET**：去读取。浏览单词是 GET。
- **POST**：去提交。点「认识」是 POST。
- **JSON**：一种写数据的文本格式，大括号、引号和冒号。例如 `{"known":"1","id":"2"}`。
- **壳**：只负责按钮和上网的那一层。Android 壳用 Kotlin 写，iOS 壳用 Swift 写。
- **核心**：单词规则。文件是 `vocab.klx`。手机库里没有自己的 HTTP 代码。
- **`.so`**：Android 要加载的共享库，一份编好的机器码。文件名是 `libkylixvocab.so`。可以把它想成「手机运行时才打开的零件」。
- **`.a`**：iOS 这份示例使用的静态库，文件名是 `libkylixvocab.a`。链接应用时把机器码拷进去。
- **OkHttp / URLSession**：Android 和 iOS 系统侧已经提供的上网工具。Kylix 核心不自己上网。

四个按钮对应的请求：

| 按钮 | 壳发出的请求 |
|---|---|
| 全部单词 | `GET /api/words?known=1,3` |
| 复习未掌握 | `GET /api/review?known=1,3` |
| 认识 | `POST /api/mark`，正文 `{"known":"1","id":"2"}` |
| 下一张 | 不再上网，只在已经拿到的卡片里换一张 |

`known=` 后面就是那串编号。路径和 POST 正文由 Kylix 库算出来，壳不去手写这些字符串。

仓库里还有另一套「登录 + Notes」，文件夹是 `apps/android/` 和 `apps/ios/`（注意没有 `vocab`）。那一套连的是 8090 端口的后台。背单词在 `apps/vocab/`，端口是 **8091**。两套可以同时开，不要打开错目录。

### 这一章的常见问题

| 你看到的 | 原因和做法 |
|---|---|
| 打开了 `apps/android`，里面是登录框 | 那是 Notes 示例。背单词要打开 `apps/vocab/android` |
| 想在 `vocab.klx` 里写上网代码 | 手机编译目标会拒绝名为 `httpclient` 的库。上网留在壳里 |

### 自检清单

- [ ] 我能指出背单词的 Android 工程在 `apps/vocab/android`
- [ ] 我知道 8091 是单词服务，8090 是另一套后台
- [ ] 我知道「下一张」不上网

---

## 3. 安装工具，并编出 kylix

### 这一步在做什么

装上编译 Kylix 自己、以及编译单词核心所需要的程序，然后在仓库根得到一个名为 `kylix` 的可执行文件。

**为什么：** 后面的脚本会在仓库根寻找这个文件。找不到就先自己编译一次，那也需要 Go。

要用到的程序：

| 名字 | 它是什么 | 背单词哪一步要用 |
|---|---|---|
| Git | 把仓库下载下来的工具 | 还没有源码时 |
| Go | 用来编译出 `kylix` 这个程序，也用来编译单词服务 | 第 3、5 章 |
| LLVM 里的 `llc` 和 `clang` | 把 Kylix 编成电脑或手机能跑的机器码的工具 | 第 4、6 章 |
| `curl` | 在终端里访问网址的命令 | 第 5 章 |
| `cc` | 系统里的 C 编译器，脚本用它做一次加载检查 | 第 4 章 |

**PATH** 是一串文件夹名单。你在终端里只打 `go`，系统就按这个名单去找名为 `go` 的程序。名单里没有，就会说 `command not found`（找不到命令）。

**后端** 在这里的意思是：同一份 `.klx` 可以编成两种结果。`--backend=go` 先变成 Go 再编译；`--backend=llvm` 直接变成机器码。第 4 章要证明两种结果打印的字一模一样。

### 3.1 下载仓库

在你准备用来工作的终端里（Windows 用户：Ubuntu 窗口）：

```bash
cd ~
git clone https://github.com/astra-zhao/kylix.git
cd kylix
ls README_CN.md apps
```

`cd` 是进入某个文件夹。`~` 是你的用户主目录。`ls` 是列出文件。

成功时，`ls` 能看到 `README_CN.md`，也能看到 `apps`。你现在的当前目录就是仓库根。后面没写 `cd` 的命令，都在仓库根执行。

已经有这份文件夹时，不必再克隆。用 `cd` 进入它，再执行 `ls README_CN.md`。

### 3.2 安装 Go、LLVM、curl、C 编译器

#### Linux，或 WSL 里的 Ubuntu

下面是 Ubuntu / Debian 的安装命令。编写环境里这些包已经装好，**没有**在空机器上重跑 `apt`。装好之后的版本检查是跑过的。

```bash
sudo apt update
sudo apt install -y golang-go git llvm clang build-essential curl unzip
```

`sudo` 会问你的 Linux 密码。`build-essential` 里包含 `cc`。

若发行版自带的 Go 很旧，改到 https://go.dev/dl/ 下载 Linux 的 `go1.25.0.linux-amd64.tar.gz`（或更新的 1.25 补丁版本），然后：

```bash
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.25.0.linux-amd64.tar.gz
echo 'export PATH=/usr/local/go/bin:$PATH' >> ~/.bashrc
source ~/.bashrc
```

`source ~/.bashrc` 是让刚才写进配置文件的 PATH 立刻生效。新开的终端也会自动读这个文件。

#### macOS

先安装 Homebrew（https://brew.sh 首页的那一行安装命令）。然后：

```bash
brew install go llvm git curl
echo 'export PATH="$(brew --prefix llvm)/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
xcode-select --install
```

最后一行会弹出窗口安装「命令行工具」，里面有 `cc`。已经装过会说已安装，那就可以。

Homebrew 的 LLVM **不会**自动出现在 PATH 里。所以上面把 `llvm` 的 `bin` 加进了 `~/.zshrc`。这是 macOS 上 `llc` 找不到时的第一处检查。编写环境是 Linux，**没有**执行 `brew`。

#### 不要在 Windows 的 cmd 里装一套来跑这些 `.sh`

Windows 版的 `kylix.exe` 存在（发布页上的 `kylix-windows-amd64.exe`），`kylix.exe doctor` 可以在 cmd 里运行。背单词的 `host_check.sh`、`e2e.sh`、`build_core.sh` 仍是 bash。请在第 0 章的 Ubuntu 里安装 Linux 版工具，而不是只在 cmd 里安装 Go。

### 3.3 检查版本

在仓库根：

```bash
go version
llc --version
clang --version
curl --version
cc --version
```

编写环境的实际第一行是：

```text
go version go1.25.0 linux/amd64
Ubuntu LLVM version 18.1.3
Ubuntu clang version 18.1.3
```

你的 LLVM 版本号可以不同。这个仓库的 `go.mod` 写的是 `go 1.25.0`，请用 1.25.0 或更新的 Go。`llc` 和 `clang` 只要有输出、而不是 `command not found`。

### 3.4 编译出仓库根的 `kylix`

仍在仓库根（能看到 `README_CN.md` 的那个目录）：

```bash
go build -o kylix ./cmd/kylix/
./kylix version
```

成功时打印：

```text
Kylix 0.15.0
```

文件就在当前目录，名字是 `kylix`。脚本默认找的就是它，不必再复制到系统目录。

### 3.5 跑 doctor

**doctor** 是 Kylix 自带的检查，看看 Go 和 LLVM 在不在。

```bash
./kylix doctor
```

编写环境里，最后一行是：

```text
✓ doctor: all required tools found.
```

中间可能出现这两行叉号（编写环境就是这样）：

```text
✗ curl: not found / not linkable (needed by db / httpclient / crypto)
✗ boehm gc (libgc, for --gc=boehm): not found / not linkable (needed by db / httpclient / crypto)
```

这里的 `curl` 指的是 **库** libcurl，不是终端命令 `curl`。背单词的手机库不链接它。只要最后一行是 `all required tools found`，就可以做第 4 章。`boehm gc` 是另一种内存回收，这个例子不用。

doctor **不会**检查 Android NDK。NDK 要到第 6 章才装。这是当前工具的缺口，不是你漏了一步。

### 这一章的常见问题

| 你看到的 | 原因和做法 |
|---|---|
| `go: command not found` | PATH 里没有 Go。Linux 检查 `ls /usr/local/go/bin/go`。有这个文件就执行 `export PATH=/usr/local/go/bin:$PATH`，再 `go version` |
| `llc: command not found` | LLVM 没装，或 macOS 上 brew 的 llvm 没进 PATH。按 3.2 的对应系统重做，然后新开一个终端再试 |
| `./kylix: Permission denied` | 执行 `chmod +x kylix`，再 `./kylix version` |
| `go: go.mod file not found` | 当前目录不是仓库根。先 `cd` 到能 `ls README_CN.md` 的目录 |
| doctor 里有 `✗ curl`，最后一行仍是 all required tools found | 可以继续。那一行不是终端命令 curl |
| doctor 最后一行是 `✗ doctor: missing required tooling` | 往上找带 `✗` 的 `go` 或 `llc` 或 `clang`，按 3.2 装上再跑 |

### 自检清单

- [ ] `go version` 有一行版本号
- [ ] `llc --version` 有一行版本号
- [ ] 仓库根有文件 `kylix`，`./kylix version` 打印 `Kylix 0.15.0`
- [ ] `./kylix doctor` 的最后一行是 `✓ doctor: all required tools found.`

---

## 4. 不打开手机，先检查单词核心

### 这一步在做什么

运行 `apps/vocab/host_check.sh`。它做两件事：

1. 用 Go 后端和 LLVM 后端各编译一次单词核心，比较打印出来的字是否逐字节相同。
2. 把核心做成一份电脑能直接打开的 `.so`，用 C 语言去调用「浏览、把 1 标成认识、复习时不再出现 apple」。

**为什么先做这个：** 手机模拟器更重。核心如果在这里就对不上，后面的界面没有意义。

**C ABI** 在这里的意思是：函数用 C 语言能调用的名字和参数形式露出来。Kotlin 和 Swift 不能直接叫 Kylix 的函数名，它们叫的是这些 C 名字，例如 `vc_browse`。第 9 章再展开。

**在哪个目录：** 仓库根。

**哪种系统：** Linux、macOS、WSL 里的 Ubuntu。命令相同。

```bash
bash apps/vocab/host_check.sh
```

第一次会自己使用仓库根的 `kylix`。成功时最后几行是：

```text
[vocab] parity PASS
c abi PASS
[vocab] host check PASS
```

中间还有 `ok browse_apple`、`ok review_hides_apple` 这类行。`review_hides_apple` 表示：编号 1 已经认识之后，复习结果里不再含有 apple。

这不是手机模拟器。

### 这一章的常见问题

| 你看到的 | 原因和做法 |
|---|---|
| `bash: apps/vocab/host_check.sh: No such file or directory` | 当前目录不是仓库根。执行 `pwd` 看路径，`cd` 到克隆下来的 `kylix` 文件夹 |
| `llc: command not found`，或 `LLVM toolchain not found: no llc and no clang` | 回到第 3 章把 `llc` 和 `clang` 放进 PATH |
| `cc: command not found` 或链接报错 `-ldl` | Linux 安装 `build-essential`（里面有 `cc`）。macOS 执行 `xcode-select --install` |
| `PARITY DIFF` 然后一大段 diff | Go 和 LLVM 打出的文字不一致。把 diff 保留下来再查，不要继续做手机。编写环境这次没有出现 diff |
| Windows cmd 说 `bash` 不是命令 | 改到 WSL 的 Ubuntu 里，并且 `cd` 到仓库 |

### 自检清单

- [ ] 命令是在仓库根执行的
- [ ] 最后一行是 `[vocab] host check PASS`
- [ ] 我知道这还没有打开模拟器

---

## 5. 用 curl 走完三步，并让服务一直开着

### 这一步在做什么

先让脚本自己启动单词服务、用 curl 对答案、再关掉服务。然后再按手动步骤把服务留在一个终端里，供第 7 章的模拟器使用。

**端口** 是一台电脑上区分不同服务的数字。单词服务用 **8091**。浏览器或 curl 访问的是 `http://127.0.0.1:8091`。`127.0.0.1` 的意思是「这台电脑自己」。

**curl** 是终端里的浏览器：把网址的正文打到屏幕上，不画按钮。

### 5.1 自动走完（服务最后会被关掉）

仓库根：

```bash
bash apps/vocab/e2e.sh
```

成功的最后一行：

```text
[vocab] e2e PASS (go)
```

再用 LLVM 编出来的服务做同一套请求（更慢一点，答案应相同）：

```bash
VOCAB_BACKEND=llvm bash apps/vocab/e2e.sh
```

成功的最后一行：

```text
[vocab] e2e PASS (llvm)
```

这两条都在编写环境里跑通了。脚本结束时服务已经停掉，所以还不能给模拟器用。

### 5.2 把服务留在前台

新开一个终端，进入仓库根，执行：

```bash
mkdir -p kylixvocab_gen
cd apps/vocab
../../kylix build --backend=go -o ../../kylixvocab_gen/server.go \
  ../../stdlib/stringutil.klx vocab.klx server.klx
cd ../..
go build -o vocab_server ./kylixvocab_gen
./vocab_server
```

`mkdir -p` 是建立文件夹，已经有也不会报错。第一条 `kylix build` 必须在 `apps/vocab` 里执行，因为后面的文件名是相对这个目录写的。`go build` 回到仓库根执行。

编写环境里，这个编译成功，启动后的日志是：

```text
vocab server on :8091
2026/10/11 01:20:45 🚀 KylixBoot started on http://localhost:8091
```

日期会不同。出现 `started on http://localhost:8091` 之后，这个终端就停在这里。这是正常的：服务要一直占着窗口。要停掉它，在这个窗口按 `Ctrl+C`。

`kylixvocab_gen/` 和 `vocab_server` 是编出来的文件，不要提交进 Git。

### 5.3 另开一个终端，自己发三次请求

仓库根，服务仍开着：

```bash
curl -sS http://127.0.0.1:8091/api/words
echo
curl -sS -X POST http://127.0.0.1:8091/api/mark \
  -H 'Content-Type: application/json' \
  --data '{"known":"","id":"1"}'
echo
curl -sS 'http://127.0.0.1:8091/api/review?known=1'
echo
```

编写环境里三次正文分别满足：

1. 第一次含有 `"left":6` 和 `apple`，`"known":""`。
2. 第二次含有 `"known":"1"` 和 `"left":5`，apple 那一项里有 `"known":true`。
3. 第三次含有 `book`，不含 `apple`。

正文很长，是一整行 JSON。用眼睛搜 `apple` 和 `left` 即可。

`Content-Type` 实际会是 `text/plain`，不是 `application/json`。壳只读正文。这是当前服务的行为，不是 curl 用错了。

### 这一章的常见问题

| 你看到的 | 原因和做法 |
|---|---|
| `port 8091 is already in use; stop that process and retry` | 已经有一个单词服务。找到那个还停在 `started on http://localhost:8091` 的终端，按 `Ctrl+C`，然后再跑 |
| `curl: (7) Failed to connect to 127.0.0.1 port 8091` | 服务没开，或开在另一个终端里又被关掉了。回到 5.2，确认窗口里有 `started on http://localhost:8091` |
| `error[KLX213]: [Embed] cannot read views` | 你在用编译后台的方法编译单词服务，或者当前目录错了。单词服务用 5.2 的命令，不要加 `[Embed]` 那条后台命令 |
| 第三次结果里还有 apple | `known` 没带上 `1`。对照命令里的引号，是 `'http://127.0.0.1:8091/api/review?known=1'` |

### 自检清单

- [ ] `bash apps/vocab/e2e.sh` 的最后一行是 `[vocab] e2e PASS (go)`
- [ ] 有一个终端停在 `KylixBoot started on http://localhost:8091`，没有按 Ctrl+C
- [ ] 另一个终端里，第一次 curl 能看到 apple，标记之后的复习 curl 看不到 apple

---

## 6. 安装 NDK，编出 Android 用的库

### 这一步在做什么

编出 `libkylixvocab.so`，放进 Android 工程指定的文件夹。Android Studio 后面只负责打包，不会替你生成这份 Kylix 库。

**NDK**（Native Development Kit）是 Google 提供的一套 Android C 编译器。Kylix 用它把库链接成手机能加载的 `.so`。没有它，编译会停在下面这句（编写环境里故意指错目录，原样得到的）：

```text
Error: android cross-link needs Android NDK (not found): install the NDK and point ANDROID_NDK_HOME at it (it must contain toolchains/llvm/prebuilt/<host>/bin/clang)
```

**ANDROID_NDK_HOME** 是一个环境变量：告诉程序「NDK 解压在哪个文件夹」。环境变量是终端记住的一个名字和对应的路径。

**ABI** 是 CPU 种类。模拟器和库必须是同一种，否则应用一打开就崩。

| 你的模拟器 | 命令里的词 | 库出现的文件夹 |
|---|---|---|
| 电脑上最常见的 x86_64 模拟器 | `amd64` | `app/src/main/jniLibs/x86_64/` |
| arm64 模拟器或多数真机 | `arm64` | `app/src/main/jniLibs/arm64-v8a/` |

**triple** 是编译器内部用的「给哪种系统、哪种 CPU」的全名。你在命令里写 `--target android/amd64`，工具会把它变成 `x86_64-linux-android30`。你现在只要选对 `amd64` 或 `arm64`。第 9 章有对照表。

### 6.1 下载 NDK r26d

仓库的 Android 检查用的是 **NDK r26d**。请用和你的终端一致的压缩包：

| 终端 | 下载地址 | 解压后里面的 clang 在 |
|---|---|---|
| Linux，或 WSL 的 Ubuntu | https://dl.google.com/android/repository/android-ndk-r26d-linux.zip | `toolchains/llvm/prebuilt/linux-x86_64/bin/clang` |
| macOS | https://dl.google.com/android/repository/android-ndk-r26d-darwin.dmg | `toolchains/llvm/prebuilt/darwin-x86_64/bin/clang` |
| Windows 自己的资源管理器（给 cmd 用，这篇教程的 bash 不用它） | https://dl.google.com/android/repository/android-ndk-r26d-windows.zip | `toolchains/llvm/prebuilt/windows-x86_64/bin/clang.exe` |

这三个地址在编写时用请求头检查过，都返回 200。编写环境没有重新下载，用的是已经放在 `/opt/android-ndk-r26d` 的同一版本。

WSL 里必须下 **linux.zip**。把 Windows 的 zip 解压进 Ubuntu，clang 会在 `windows-x86_64` 里，Linux 上的 Kylix 找不到它，报错仍是上面那句 NDK not found。

在 Linux 或 WSL 里：

```bash
cd ~
curl -L -o ndk.zip https://dl.google.com/android/repository/android-ndk-r26d-linux.zip
unzip -q ndk.zip
ls "$HOME/android-ndk-r26d/toolchains/llvm/prebuilt/linux-x86_64/bin/clang"
```

`ls` 能打印出 `clang` 的路径，才算解压对了。压缩包解开后的文件夹名就是 `android-ndk-r26d`。

macOS 打开 dmg，把里面的 `android-ndk-r26d` 拖到主目录。然后检查 `darwin-x86_64` 那条路径。

### 6.2 设置 ANDROID_NDK_HOME

把下面两行里的路径换成你的实际文件夹。Linux / WSL 写进 `~/.bashrc`，macOS 写进 `~/.zshrc`。

```bash
echo 'export ANDROID_NDK_HOME="$HOME/android-ndk-r26d"' >> ~/.bashrc
source ~/.bashrc
echo "$ANDROID_NDK_HOME"
ls "$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/linux-x86_64/bin/clang"
```

macOS 把文件名改成 `~/.zshrc`，并把 `linux-x86_64` 改成 `darwin-x86_64`。

`echo` 应打印你设的路径，`ls` 应打印 clang，而不是 `No such file`。

编写环境的 clang 版本第一行是：

```text
Android (11349228, +pgo, +bolt, +lto, -mlgo, based on r487747e) clang version 17.0.2
```

你的版本号只要来自这个 NDK 即可。

Kylix 还会在 `ANDROID_HOME/ndk/` 或 `ANDROID_SDK_ROOT/ndk/` 下面找版本号文件夹（Linux 上也会看 `~/Android/Sdk/ndk` 和 `/opt/android-ndk`）。新手请仍然设置 `ANDROID_NDK_HOME`，这样你知道它用的是哪一份。

### 6.3 编译 .so

回到仓库根。x86_64 模拟器：

```bash
bash apps/vocab/android/build_core.sh amd64
bash apps/vocab/android/check_artifact.sh \
  apps/vocab/android/app/src/main/jniLibs/x86_64/libkylixvocab.so amd64
```

arm64 模拟器或真机，把两处 `amd64` 换成 `arm64`，并把 `check_artifact.sh` 的文件路径换成：

```text
apps/vocab/android/app/src/main/jniLibs/arm64-v8a/libkylixvocab.so
```

成功时，输出里有一行 `[vocab-android] wrote .../libkylixvocab.so`，还有一句哪个模拟器用哪一份（x86_64 用 amd64，arm64 用 arm64）。最后一行是：

```text
android artifact OK (amd64)
```

`build_core.sh` 已经做了原来 `check_artifact.sh` 的 ELF 和符号检查。再单独跑检查脚本时，最后一行仍是：

```text
android artifact OK (amd64)
```

或 `android artifact OK (arm64)`。编写环境两种都跑通了。`file` 的那一行里，amd64 含有 `ELF 64-bit LSB shared object, x86-64`，arm64 含有 `ARM aarch64`。

设了 `ANDROID_NDK_HOME` 时，检查脚本还会把 JNI 桥和这份库链一次。**JNI** 是 Android 让 Kotlin 调用 C 函数的通道。链成功时会多一行 `jni link OK`。

脚本编译的文件只有 `stringutil.klx`、`vocab.klx`、`vocab_lib.klx`。里面没有 `server.klx`。手机库不含那个网站。

两种 ABI 可以各编一次，两个文件夹互不覆盖。模拟器是哪一种，就至少要有对应的那一份。

### 这一章的常见问题

| 你看到的 | 原因和做法 |
|---|---|
| `android cross-link needs Android NDK (not found)` | `echo "$ANDROID_NDK_HOME"` 是空的，或指到没有 `toolchains/llvm/prebuilt/<你的系统>/bin/clang` 的目录。按 6.2 的 `ls` 检查。WSL 不要用 Windows 版 NDK |
| `ANDROID_NDK_HOME is unset; skipped the JNI link` 然后仍有 `android artifact OK` | 符号表已经对上，但这次检查没链 JNI。把环境变量设上，重新开终端再跑 `check_artifact.sh` |
| `missing dynamic symbol vc_browse` | 编出来的不是这份背单词库，或文件下错了。看 `build_core.sh` 打印的 `wrote` 路径是不是你检查的那个文件 |
| 命令在 `apps/vocab` 里直接跑 `build_core.sh amd64` 找不到脚本 | 脚本在 `apps/vocab/android/build_core.sh`。请在仓库根用上面的 `bash apps/vocab/android/...` |

### 自检清单

- [ ] `echo "$ANDROID_NDK_HOME"` 不是空行
- [ ] 该目录下对应系统的 `bin/clang` 存在
- [ ] 模拟器是 x86_64 时，`check_artifact.sh ... amd64` 的最后一行是 `android artifact OK (amd64)`
- [ ] 我知道这份 `.so` 还没有装进手机，只是放进了工程文件夹

---

## 7. 在 Android Studio 里打开，并在模拟器里点

### 这一步在做什么

用 Android Studio 把壳装进模拟器，连上第 5 章那个还开着的单词服务。

**模拟器（emulator）** 是在电脑里跑的一台假 Android 手机。它不是 iOS 的 Simulator。名字相近，软件不同。

**Gradle** 是 Android 工程用来下载依赖、调用 CMake、打出安装包的构建系统。这个仓库**没有**自带 `gradlew`，所以不写 `./gradlew` 那条命令。请用 Android Studio 自己带的 Gradle。编写环境确认了 `apps/vocab/android/gradlew` 不存在。

**这一章的点击没有在编写环境里做。** 按钮名称按 Android Studio 默认英文界面写。如果你的菜单是中文，看括号里的意思。

开始之前：

- 第 5.2 节的终端仍停在 `started on http://localhost:8091`。
- 第 6 章的 `.so` 已经放进和模拟器 CPU 相同的 `jniLibs` 文件夹。

### 7.1 安装 Android Studio

1. 打开 https://developer.android.com/studio ，下载与你的系统对应的安装包。Windows 用户装在 Windows 上，不要装进 WSL。
2. 安装时一路用默认选项。第一次启动的向导里选 **Standard**，许可证页面点 **Accept**，等它下完 SDK。
3. 向导要下的是 Android SDK。它和大段 NDK zip 是两套东西。第 6 章的 NDK 仍要保留。

### 7.2 打开正确的文件夹

1. 欢迎页点 **Open**（打开）。已经在工程里时，用菜单 **File → Open**。
2. 选中文件夹 `apps/vocab/android`。判断标准：这个文件夹里能看到 `settings.gradle.kts`。不要选仓库根，不要选某一个 `.kt` 文件。
3. 若出现 **Trust and Open**（信任并打开），点信任。这是你自己克隆的源码。
4. 右下角开始 **Gradle Sync**。第一次会下载依赖，需要网络，可能要几分钟。等到同步结束、没有红色报错。

同步时如果弹出这些窗口：

| 提示 | 你点什么 |
|---|---|
| 缺少 Android SDK Platform 34，或要安装 SDK | 点安装 / Accept。工程的 `compileSdk` 是 34 |
| 选择 JDK，或 Gradle JDK | 选 **Embedded JDK** 或标着 17 的那一项。工程按 JDK 17 写的 |
| 要不要用 Gradle Wrapper | 选 Studio 建议的默认项（让它下载 wrapper）。仓库里没有现成的 `gradlew` |

同步失败时打开 **File → Settings**（macOS 是 **Android Studio → Settings**）→ **Languages & Frameworks → Android SDK**。记下 **Android SDK Location**。若同步说缺 NDK 或 CMake：在同一页的 **SDK Tools** 选项卡里勾选 **NDK (Side by side)** 和 **CMake**，点 Apply。Studio 用这份 NDK 来编译很小的 JNI 桥；Kylix 的 `.so` 仍是第 6 章用 `ANDROID_NDK_HOME` 编的那份。

若同步或编译说：

```text
Missing .../jniLibs/.../libkylixvocab.so
```

回到第 6 章，确认 `wrote` 的路径和报错里的 `x86_64` 或 `arm64-v8a` 是同一个文件夹。

### 7.3 建立模拟器

1. 右上角手机形状的图标 **Device Manager**（设备管理器）。找不到时用菜单 **Tools → Device Manager**。
2. 点 **Create Device**（创建虚拟设备）。
3. 选一台手机，例如 **Pixel 6**，点 **Next**。
4. 选一个 **x86_64** 的系统镜像（名字里带 `x86_64`，API 34 即可）。旁边若是 **Download**，先下载再 Next。你若只编了 `arm64` 的 `.so`，这里要选 **arm64-v8a** 镜像，不要选 x86_64。
5. 点 **Finish**。
6. 在设备列表里点这一台的播放三角形，等它开机到桌面。工具栏的设备下拉框里要能选中它。

### 7.4 运行应用

1. 窗口上方绿色三角形 **Run**。不要点 **Build → Generate Signed Bundle / APK**。那个是上架用的签名包，模拟器不需要。调试安装用 Studio 自带的 debug 签名，不会问你密钥库。
2. 若弹出设备列表，选刚才的模拟器。
3. 等到模拟器里出现标题 **背单词**。应用 id 是 `dev.kylix.vocab`。

界面上「Server URL」默认是：

```text
http://10.0.2.2:8091
```

**10.0.2.2** 是官方 Android 模拟器看「跑着模拟器的那台电脑」的地址。在模拟器里，`127.0.0.1` 是模拟器自己，不是你的单词服务。所以这里保持 `10.0.2.2`，除非你是真机。

真机（同一 Wi-Fi）：把地址改成电脑的局域网 IP，端口仍是 8091。这个示例允许明文 HTTP，只适合你自己的电脑，不要放到公网。

然后：

1. 点 **全部单词**。第一张应是 `apple`，下面是 `苹果`。进度里有「还剩 6 / 6」。
2. 点 **认识**。进度变成还剩 5。这张下面出现「已认识」。
3. 点 **下一张**。只换卡片。
4. 点 **复习未掌握**。apple 不再作为卡片出现，能看到 book。

关掉应用再打开，复习列表仍不含 apple。编号存在应用自己的本地存储里，不是账号。

### 7.5 Windows + WSL 时怎样判断服务能不能被模拟器看见

这次没有在 Windows 上实测。请先用 **Windows 自己的浏览器** 打开：

```text
http://127.0.0.1:8091/api/words
```

- 浏览器里能看到带 `apple` 的 JSON：再在模拟器里用 `http://10.0.2.2:8091`。模拟器的 `10.0.2.2` 指 Windows，Windows 的 `127.0.0.1` 在较新的 WSL2 上会转到 Ubuntu 里的服务。
- 浏览器打不开，但 Ubuntu 里的 `curl http://127.0.0.1:8091/api/words` 可以：服务只在 WSL 里听得见。到 Windows 的「WSL 设置」确认 localhost 转发是开的，或把仓库放在 Ubuntu 家目录里重新按第 5.2 节启动。不要把模拟器地址改成 Ubuntu 里 `hostname -I` 的地址除非你知道模拟器能路由到那个网段；先把 Windows 浏览器打通。

Android Studio 打开 WSL 里的工程时，路径类似 `\\wsl$\Ubuntu\home\<你的用户名>\kylix\apps\vocab\android`。在 Open 对话框的地址栏可以粘贴这个路径。

### 这一章的常见问题

| 你看到的 | 原因和做法 |
|---|---|
| 一打开应用就闪退，Logcat 里有 `UnsatisfiedLinkError` 和 `libkylixvocab.so` | 库的 CPU 和模拟器不一致。x86_64 模拟器需要第 6 章的 `amd64`。编完对应的那一份再 Run |
| 点按钮后界面写 `Failed to connect` 或 `network error` | 服务没开、端口不是 8091，或模拟器地址写成了 `127.0.0.1`。先在跑服务的那台系统上 `curl` 第 5.3 节的第一条 |
| Gradle Sync 一直转圈后失败，提示 SDK 或 license | SDK Tools 里接受协议，装上 Platform 34。公司网络可能要代理，这篇教程不写代理参数 |
| 找不到 `gradlew` | 正常。用 Android Studio 的 Run，不要自己造一个 wrapper 命令来替代第 6 章 |
| 点了 Generate Signed Bundle，问 keystore | 关掉那个向导，改点绿色 Run |

### 自检清单

- [ ] 打开的文件夹里面有 `settings.gradle.kts`，路径以 `apps/vocab/android` 结尾
- [ ] Gradle Sync 没有红色错误
- [ ] 模拟器 CPU 和第 6 章的 `amd64` 或 `arm64` 一致
- [ ] 服务终端仍停在 8091
- [ ] 模拟器里看到 apple / 苹果；点认识后再点复习，apple 不在复习卡片里

若你还没有 Android Studio，第 4–6 章的自检已经通过，也算把「核心和库」做完。第 7 章可以以后再点。

---

## 8. iOS：只有 Mac 能编库、能开模拟器

### 这一步在做什么

在 Mac 上编出 `libkylixvocab.a`，用 Xcode 把 Swift 壳装进 **Simulator**。

iOS Simulator 是 Mac 上的假 iPhone。它和 Android 模拟器不是同一个软件。Android 的 `10.0.2.2` 在这里不适用。模拟器和 Mac 共用网络，所以地址是 `http://127.0.0.1:8091`。

**不是 Mac 就做不了这一章的后半。** 在 Linux 或 WSL 里执行构建脚本，会立刻失败。编写环境跑过，退出码是 1，原文是：

```text
iOS archives must be linked on macOS with Xcode command line tools.
On this host, stop here. The Swift sources are in apps/vocab/ios/.
```

编译器自己的英文是：`ios cross-link requires a macOS host with Xcode command line tools`。看到这两句，表示脚本判断正确，不是仓库坏了。Swift 源码还在，换到 Mac 上再编。

下面从「在 Mac 上」开始的点击，**没有**在编写环境里执行。

### 8.1 在 Mac 上准备

1. 从 App Store 安装 **Xcode**（完整应用，不只是命令行工具）。打开一次，接受协议。若要装额外组件，等它下完。
2. 终端里执行：

```bash
sudo xcode-select -s /Applications/Xcode.app
xcode-select -p
```

应打印 `/Applications/Xcode.app/Contents/Developer`。

3. 安装 XcodeGen（仓库不附带它）。有 Homebrew 时：

```bash
brew install xcodegen
xcodegen --version
```

有版本号即可。

4. 第 3–5 章在这台 Mac 上做完，单词服务开在 8091。第 3 章的 `llc` 要在 PATH 里。

### 8.2 编库

仓库根：

```bash
bash apps/vocab/ios/build_core.sh simulator
bash apps/vocab/ios/check_artifact.sh \
  apps/vocab/ios/Sources/CKylixVocab/lib/libkylixvocab.a simulator
```

成功时输出里有 `[vocab-ios] wrote .../libkylixvocab.a`。最后一行是 `ios artifact OK (simulator)`（构建脚本已经做了原来检查脚本的符号和 SDK 链接）。

`check_artifact.sh` 的最后一行应是 `ios artifact OK (simulator)`。它会用 Xcode 的 SDK 把一个很小的 C 程序链上这份库。编写环境没有 Mac，所以这行成功输出没有在这里见到。

`simulator` 给 Apple 芯片的 Mac 模拟器（编译目标名是 `ios/simulator-arm64`）。真机用：

```bash
bash apps/vocab/ios/build_core.sh device
```

两次写的是**同一个文件** `libkylixvocab.a`。后一次会盖掉前一次。要跑模拟器，最后一次必须是 `simulator`。这个 `.a` 不进 Git，每台 Mac 自己编。

### 8.3 在 Xcode 里打开

终端：

```bash
cd apps/vocab/ios
xcodegen generate
open KylixVocab.xcodeproj
```

`xcodegen generate` 根据 `project.yml` 生成 Xcode 工程。打开的是 `KylixVocab.xcodeproj`，不是某一个 `.swift` 文件。

在 Xcode 里：

1. 左侧点最上面的蓝色工程图标。中间选中目标 **KylixVocab**（不要只选项目不选目标）。
2. 打开 **Signing & Capabilities**。**Team** 选你的 Apple ID。没有账号时：菜单 **Xcode → Settings → Accounts**，点左下角 **+**，选 Apple ID，登录免费账号，再回到 Team 下拉框选中它。
3. 若出现 “Signing for KylixVocab requires a development team”，就是第 2 步的 Team 还空着。
4. 窗口最上方，scheme 选 **KylixVocab**。旁边的设备选一台 **iPhone 模拟器**（系统 16 或更高），不要选 **Any iOS Device**。
5. 按左上角三角形 **Run**（或按 `Cmd+R`）。

应用起来之后，服务器一栏保持：

```text
http://127.0.0.1:8091
```

按钮和 Android 相同：全部单词、认识、下一张、复习未掌握。看到 apple 消失，就和 Android 是同一套含义。

真机：先 `build_core.sh device`，再重新 Run，把服务器改成这台 Mac 的局域网地址。手机要信任你的开发者证书（设置里的描述文件）。这一步同样没有在编写环境里做过。

### 这一章的常见问题

| 你看到的 | 原因和做法 |
|---|---|
| Linux 上脚本立刻打印必须在 macOS 上链接 | 换一台 Mac。不要在 Linux 上改脚本强行链接 |
| `xcodegen: command not found` | 按 8.1 安装。没有它就没有 `KylixVocab.xcodeproj` |
| `xcode-select: error: tool 'xcodebuild' requires Xcode` | 只装了命令行工具。从 App Store 装完整 Xcode，再执行 8.1 的 `xcode-select -s` |
| 模拟器里连不上，Mac 上 curl 可以 | 地址应是 `http://127.0.0.1:8091`，不是 `10.0.2.2` |
| 编译报找不到 `libkylixvocab` | 8.2 没成功，或最后一次编的是 `device` 而你在跑模拟器。重新 `build_core.sh simulator` |
| 要登录付费开发者才能跑模拟器 | 模拟器一般用免费 Apple ID 即可。付费账号是真机和上架才需要。Team 下拉框里选 Personal Team |

### 自检清单

- [ ] 不是 Mac 时，我已经看到「必须在 macOS 上链接」并停在这里
- [ ] Mac 上 `build_core.sh simulator` 的输出含有 `wrote` 和 `libkylixvocab.a`，最后一行是 `ios artifact OK (simulator)`
- [ ] Xcode 的目标是模拟器，Team 不是空的
- [ ] 单词服务在这台 Mac 的 8091 上
- [ ] 模拟器里复习列表不含刚点过「认识」的词

---

## 9. Export、kylix_free、triple 是什么

这一章不增加新的操作。它把前面已经用过的三个词说完整。可以在第 4 章通过之后再读。

### Export 和 C ABI

壳不能调用 `VcDeckJSON` 这种 Kylix 名字。`apps/vocab/vocab_lib.klx` 里写了 `[Export('vc_browse')]` 这类标注，编译后库里就有一个 C 函数 `vc_browse`。

Android 的 `kylix_jni.c` 和 iOS 的 `kylixvocab.h` 调用的是同一批名字。头文件在：

```text
apps/vocab/ios/Sources/CKylixVocab/include/kylixvocab.h
```

| C 函数 | 作用 |
|---|---|
| `vc_words_query` / `vc_review_query` | 返回 `/api/words?known=1,3` 这种路径。编号会先排成从小到大、去掉重复 |
| `vc_mark_path` | 返回 `/api/mark` |
| `vc_mark_request` | 拼出 POST 的 JSON |
| `vc_parse` | 看 HTTP 状态和正文。成功就留下原文，失败就收成带 `error` 的 JSON |
| `vc_browse` / `vc_review` / `vc_apply_mark` | 同一份卡片 JSON。网站直接调用 Kylix 函数；第 4 章通过这些 C 名字再调一次，用来对上 |
| `kylix_free` | 释放上面那些函数返回的字符串 |

### kylix_free

这些函数返回的字符串是新分配的一块内存（源码里写成 `s + ''`）。壳把文字抄进自己的字符串之后，必须调用 `kylix_free` 把那块内存还回去。不还，会一直占着。库里的固定文字不能拿去 free。

你写界面时不用自己调用它：Android 的 `kylix_jni.c` 和 iOS 的 `KylixVocab.swift` 已经在复制之后调用了。

Kylix 里的 `Integer` 到了 C 这一侧是 64 位整数 `int64_t`。

### triple

| 你写的 `--target` | 实际机器 |
|---|---|
| `android/amd64` | 给 x86_64 模拟器用的 Android 库 |
| `android/arm64` | 给 arm64 手机或 arm64 模拟器用的 Android 库 |
| `ios/simulator-arm64` | 给 Mac 上的 iOS 模拟器 |
| `ios/arm64` | 给真 iPhone |

对应关系已经写进 `build_core.sh`，按第 6、8 章的参数选即可。

路径函数在 Kylix 里写成带一个用不到的参数（`VcWordsPath(0)`），因为 Go 后端对「跨文件、没有参数、又写在别的函数参数位置」的调用会丢掉括号。C 那边看到的 `vc_words_path` 是没有参数的。

自己要加新的导出函数时，语法在 [EXPORT_C_ABI.md](EXPORT_C_ABI.md)。编进手机的命令仍是第 6、8 章的脚本。不要把 `server.klx` 加进那条文件列表。

### 手机上不要用的库

背单词的手机库只用了字符串工具和 `vocab.klx`。下面这些在 android / ios 目标上的行为已经固定：

| 在手机目标上 | 会发生什么 |
|---|---|
| `httpclient`、连接 Postgres 的库 | 链接前被拒绝 |
| AES、BCrypt、Pbkdf2 | 编译直接报错。这个例子用不到 |
| SHA-256 / MD5 | 有一份不依赖 OpenSSL 的实现。背单词没用到 |

桌面程序仍可以使用原来的 OpenSSL 和 libcurl。不要把桌面的链接方式抄到第 6 章的命令里。

### 自检清单

- [ ] 我能说出壳调用的是 `vc_` 开头的 C 函数，不是 `VcDeckJSON`
- [ ] 我知道返回的字符串由 JNI / Swift 包装调用 `kylix_free`
- [ ] 我知道 `server.klx` 不进手机库

---

## 10. 附录：登录和 Notes

背单词没有账号。需要登录、口令，以及从数据库读出一张列表时，用仓库里的另一套示例。

- 规则在 `apps/shared/mobilecore.klx`
- Android 壳在 `apps/android/`（不是 `apps/vocab/android`）
- iOS 壳在 `apps/ios/`
- 后台是 KylixAdmin，默认端口 **8090**

字段说明在 [MOBILE_APPS.md](MOBILE_APPS.md)。这一附录的长编译，这次重写时**没有**重新执行。命令与已经合并的示例一致。

先检查那份核心（不打开模拟器）：

```bash
bash apps/shared/host_check.sh
```

成功时最后一行是 `[core] host check PASS`。

后台必须在 `apps/admin/` 里编译。`main.klx` 要读旁边的 `views` 和 `static`。在仓库根编译会报：

```text
error[KLX213]: [Embed] cannot read views
```

```bash
cd apps/admin
mkdir -p ../../kylixadmin_gen

../../kylix build --backend=go -o ../../kylixadmin_gen/main.go \
  ../../stdlib/stringutil.klx ../../stdlib/template_engine.klx \
  entities/admin_entities.klx \
  lib/dialect.klx lib/migrate.klx lib/admindb.klx lib/adminsec.klx lib/audit.klx \
  lib/crud.klx lib/crudrender.klx lib/crudhooks.klx lib/adminpage.klx ../shared/mobilecore.klx \
  controllers/entity.klx controllers/dashboard.klx \
  controllers/profile.klx controllers/theme.klx controllers/api.klx \
  main.klx

cd ../..
go build -o kylixadmin ./kylixadmin_gen
./kylixadmin
```

成功时能看到端口 8090。不设置环境变量 `KYADMIN_PASSWORD` 时，口令是 `Admin@123`，启动时会有一行警告。浏览器打开 http://localhost:8090 ，用户名 `admin`，用这个口令登录。左侧进入 **Notes**（地址 `/admin/notes`）新建一条。正文请用普通单行文字。

**token** 是登录成功之后，服务器发给壳的一串字，用来证明「这是刚才那个用户」，而不是把口令存在手机上反复发送。这一套有两张：

- `token`：大约 24 小时，拿去读取 Notes。
- `refresh_token`：大约 30 天，只拿去换新的一对，或登出。

同一用户最多留 8 张 refresh。细节在 `docs/MOBILE_APPS.md`。背单词不发 token。

Android 库：

```bash
bash apps/android/build_core.sh amd64
bash apps/android/check_artifact.sh \
  apps/android/app/src/main/jniLibs/x86_64/libkylixlogic.so amd64
```

用 Android Studio 打开 `apps/android/`。模拟器里的服务器默认是 `http://10.0.2.2:8090`。空列表时的英文是 `No notes yet...`，回浏览器新建后再登录。

iOS 只能在 Mac 上：`bash apps/ios/build_core.sh simulator`，然后在 `apps/ios` 里 `xcodegen generate` 并打开 `KylixAdmin.xcodeproj`。模拟器地址是 `http://127.0.0.1:8090`。

### 自检清单

- [ ] 我没有把 Notes 工程和 `apps/vocab` 弄混
- [ ] 我知道 token 只属于这一附录，背单词用的是 `known` 编号串

---

## 11. 术语表

| 词 | 白话 |
|---|---|
| 终端 | 打命令、看文字输出的窗口 |
| 仓库根 | 克隆下来的 `kylix` 文件夹，里面有 `README_CN.md` |
| 当前目录 | 终端现在所在的文件夹。`pwd` 可以打印它 |
| PATH | 系统寻找 `go`、`llc` 这类命令的文件夹名单 |
| bash | 这篇教程里 `.sh` 脚本使用的命令语言。Windows 要在 WSL 的 Ubuntu 里用 |
| WSL | 装在 Windows 里的 Linux。教程脚本在这里跑 |
| Go | 用来把 Kylix 编译器自己编出来的语言工具 |
| LLVM / llc / clang | 把程序变成机器码的工具。`llc` 和 `clang` 都要在 PATH 里 |
| 后端 | `--backend=go` 或 `--backend=llvm`，两种编译路线 |
| doctor | `./kylix doctor`，检查 Go 和 LLVM 在不在。不检查 NDK |
| 端口 | 同一台电脑上区分服务的数字。单词是 8091，Notes 后台是 8090 |
| 127.0.0.1 | 「这台电脑自己」。在 Android 模拟器里它是模拟器自己 |
| 10.0.2.2 | Android 模拟器访问电脑的专用地址 |
| curl | 在终端里取回网址正文 |
| JSON | 带大括号和引号的数据文本 |
| HTTP GET / POST | 读取 / 提交 |
| 核心 | `vocab.klx` 里的单词规则 |
| 壳 | 按钮和上网。Android 在 `apps/vocab/android`，iOS 在 `apps/vocab/ios` |
| NDK | 编 Android `.so` 用的 Google 工具包。用 `ANDROID_NDK_HOME` 指出位置 |
| `.so` | Android 运行时加载的库 `libkylixvocab.so` |
| `.a` | iOS 示例使用的库 `libkylixvocab.a`，只能在 Mac 上链接 |
| ABI / amd64 / arm64 | CPU 种类。库和模拟器必须一致。命令里的 `amd64` 就是 x86_64 |
| triple | `--target` 后面那串的内部全名，例如给模拟器用的 Android x86_64 |
| JNI | Android 上 Kotlin 调用 C 函数的通道 |
| C ABI / Export | 把 Kylix 函数变成 `vc_browse` 这种 C 名字 |
| kylix_free | 壳抄完字符串之后，把 Kylix 分配的那块内存还回去 |
| 模拟器 emulator | Android Studio 里的假 Android 手机 |
| Simulator | Xcode 里的假 iPhone。只有 Mac 有 |
| Gradle | Android Studio 用来同步工程的构建系统。本仓库没有 `gradlew` |
| CMake | 用来把很小的 JNI C 文件编进应用。缺 `.so` 时它会报 Missing |
| token | 附录里登录之后的凭证。背单词不用 |
| id / known | 单词编号，以及已经认识的编号串，例如 `1,3` |

更细的接口说明在 [MOBILE_APPS.md](MOBILE_APPS.md)。为什么界面不写进 Kylix，在 [MULTIPLATFORM.md](MULTIPLATFORM.md)。把移动端步骤收成更少的命令，是计划中的 v0.16.0，见 [MOBILE_DX.md](MOBILE_DX.md)，现在还没有那条命令。
