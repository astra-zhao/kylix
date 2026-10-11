# 用 Kylix 做 Android 和 iOS（零基础教程）

这篇教程带你做出一个很小的「背单词」应用：六个英文词，能浏览、能点「认识」、能只复习还没掌握的。

你不需要会 Pascal，也不需要先读编译器内部。会把命令复制到终端、会看终端里打出来的字，就可以开始。

当前版本是 **v0.16.0**（2026-10-11）。`./kylix version` 打印 `Kylix 0.16.0`。命令的逐项说明在 [MOBILE_COMMANDS.md](MOBILE_COMMANDS.md)。

主练习是 `apps/vocab/`。登录和 Notes 是另一套示例，放在文末的进阶里，只告诉你去哪看，不在这里重做。

## 最短的一条路

工具装好之后，在仓库根（能看到 `README_CN.md` 的那个文件夹）只需要这四条：

```bash
./kylix mobile doctor
./kylix mobile check
./kylix mobile serve
./kylix mobile build android amd64
```

| 命令 | 它替你做的事 | 成功时看这一行 |
|---|---|---|
| `doctor` | 检查 Go、clang、curl，以及 Android NDK 在不在 | `✓ mobile doctor: vocab tools are ready.` |
| `check` | 不打开手机，把浏览、认识、复习走一遍 | `[vocab] e2e PASS (go)` |
| `serve` | 把单词服务留在这个窗口，给模拟器用 | `vocab server on :8091` |
| `build android amd64` | 编出 x86_64 模拟器要的库 | `android artifact OK (amd64)` |

arm64 模拟器或真机把最后一条的 `amd64` 换成 `arm64`。iOS 要另外一台 Mac，见第 8 章。

后面每一章解释这些词，并给出这台 Linux 上的原文。你的文件夹路径会和这里的 `/workspace`、`/tmp/ndk/...` 不同，成功行的英文要一致。

## 哪些步骤是在这台机器上跑过的

编写时的机器：`uname -s` 打印 `Linux`，bash 5.2.21，`go version` 是 `go version go1.25.0 linux/amd64`，系统 clang 是 `Ubuntu clang version 18.1.3 (1ubuntu1)`。这台机器没有单独的 `llc` 命令。NDK 是 r26d，放在 `/tmp/ndk/android-ndk-r26d`。NDK 里的 clang 第一行是：

```text
Android (11349228, +pgo, +bolt, +lto, -mlgo, based on r487747e) clang version 17.0.2 (https://android.googlesource.com/toolchain/llvm-project d9f89f4d16663d5012e5c09495f3b30ece3d2362)
```

| 步骤 | 这台 Linux |
|---|---|
| `go build -o kylix`、`./kylix version` | 跑过，打印 `Kylix 0.16.0` |
| `./kylix mobile doctor`（NDK 指对） | 跑过，退出码 0 |
| 不设 `ANDROID_NDK_HOME` / `ANDROID_NDK_ROOT` 再跑 doctor | 跑过，退出码仍是 0 |
| HOME 指到不存在的目录，同时 ROOT 指着一份好的 NDK | 跑过，退出码 1 |
| HOME 指到只有 `windows-x86_64` 的目录，ROOT 仍指着好的 NDK | 跑过，退出码 1 |
| `./kylix mobile check` | 跑过，最后一行 `[vocab] e2e PASS (go)` |
| `./kylix mobile check --e2e-only --backend=llvm` | 跑过，最后一行 `[vocab] e2e PASS (llvm)` |
| `./kylix mobile check --host-only` | 跑过，最后一行 `[vocab] host check PASS`，没有 e2e 那一行 |
| `./kylix mobile serve`，以及下面三条 curl | 跑过。`Content-Type` 是 `text/plain; charset=utf-8` |
| 在跑着 serve 的时候再跑 `check --e2e-only` | 跑过，退出码 1 |
| 对 serve 那个进程组发 SIGINT（终端里的 Ctrl+C） | 跑过，8091 不再接受连接 |
| `build android amd64` 和 `arm64` | 跑过 |
| `build android x86_64` | 跑过，退出码 1。子命令不接受这个词 |
| 不设 NDK 环境变量就 `build android amd64` | 跑过，退出码 1 |
| `artifact android`（设了 NDK，以及两个变量都空着） | 跑过，退出码都是 0 |
| `build ios simulator` 和 `build ios device` | 跑过，两段失败原文相同，没有写出 `.a` |
| Android Studio、模拟器点击、Mac 上的 Xcode、Windows | **没有**跑。第 7、8 章里写着「没点过」的部分来自工程文件和程序源码，不是这台机器的点击记录 |

---

## 目录

0. [先看你的电脑](#0-先看你的电脑)
1. [做完之后你手里有什么](#1-做完之后你手里有什么)
2. [程序分成几块](#2-程序分成几块)
3. [装上工具，编出 kylix，跑 doctor](#3-装上工具编出-kylix跑-doctor)
4. [不打开手机：kylix mobile check](#4-不打开手机kylix-mobile-check)
5. [把服务留在前台：kylix mobile serve](#5-把服务留在前台kylix-mobile-serve)
6. [NDK 和两种模拟器：build android](#6-ndk-和两种模拟器build-android)
7. [在 Android Studio 里打开](#7-在-android-studio-里打开)
8. [iOS：这台 Linux 上会停住](#8-ios这台-linux-上会停住)
9. [排错：对着原文看](#9-排错对着原文看)
10. [进阶](#10-进阶)

---

## 0. 先看你的电脑

终端是一个打字、按回车、看文字结果的窗口。命令是交给电脑的一行字，例如 `go version`。

仓库是克隆下来的这一整份 Kylix 源代码。仓库根是里面能直接看到 `README_CN.md` 和 `apps` 的那个文件夹。后面没写 `cd` 的命令，都在仓库根执行。

当前目录是终端现在站着的文件夹。站错了，命令会说找不到文件。`pwd` 可以打印当前目录。

| 你的电脑 | 怎么走这篇教程 |
|---|---|
| Linux（Ubuntu 等） | 从第 3 章起，命令按原文复制。编写时跑通的就是这种 |
| macOS | 第 3 章用 macOS 那一小节装工具。第 4–7 章的 `./kylix mobile ...` 与 Linux 相同。第 8 章只有 Mac 能编出 iOS 的库 |
| Windows | 用 **WSL2 里的 Ubuntu** 做第 3–6 章。WSL 是装在 Windows 里的一套真正的 Linux。装好之后，开始菜单里的 **Ubuntu** 才是后面要用的终端 |

`apps/vocab/*.sh` 是 bash 脚本。Windows 自带的命令提示符和 PowerShell 跑不了它们。`kylix mobile` 本身是程序，不是 `.sh`。这篇教程的动手步骤用 `./kylix mobile`，所以在 Ubuntu 窗口里做即可。

这台机器没有打开 Windows。程序在 Windows 上的 doctor 会多打一行（这句话写在程序里，这里没有在 Windows 上执行）：

```text
  − bash scripts need WSL (Ubuntu). Do not run apps/vocab/*.sh from cmd or PowerShell. kylix mobile itself runs on Windows.
```

意思是：bash 脚本请用 WSL；`kylix mobile` 这个程序可以在 Windows 上跑。你仍然建议在 Ubuntu 里做第 3–6 章，因为 NDK 要下 Linux 版，见第 6 章。

### 安装 WSL2（在 Windows 里做一次）

1. 开始菜单里找到 **PowerShell**，右键，选 **以管理员身份运行**。
2. 输入 `wsl --install`，回车。
3. 按提示重启。重启后 Ubuntu 会让你设用户名和密码。密码输入时屏幕上没有圆点，输完回车即可。
4. 以后做第 3–6 章，打开开始菜单里的 **Ubuntu**。

在 Ubuntu 里：

```bash
uname -s
```

应打印 `Linux`。

Android Studio 装在 Windows 上（第 7 章），不要装进 WSL。单词服务如果跑在 WSL 里，Windows 上的模拟器能不能连上，这次没有实测。第 7 章写了怎样用浏览器判断。

---

## 1. 做完之后你手里有什么

你会得到这些：

1. 电脑上一个小服务，地址是 `http://127.0.0.1:8091`。它发出六张单词卡片。
2. 一条检查：同一份单词逻辑，两种编译方式打出来的结果一致，并且 curl 能走完浏览、认识、复习。
3. 一份 Android 要加载的库 `libkylixvocab.so`。有 Android Studio 时，可以把它装进模拟器。

内置的六个词：

| 编号 id | 英文 | 中文 |
|---|---|---|
| 1 | apple | 苹果 |
| 2 | book | 书 |
| 3 | water | 水 |
| 4 | friend | 朋友 |
| 5 | time | 时间 |
| 6 | happy | 高兴 |

**id** 是这张卡片的编号。应用记住「已经认识」时，存的是编号串，例如 `1,3` 表示认识了 apple 和 water。单词服务不把这串字写进数据库，每次由手机带回去。所以这个例子不需要注册账号。

这不是应用商店里的完整产品。它用来把「一份逻辑、手机只负责画和上网」走通。

---

## 2. 程序分成几块

```text
你写一次的单词规则                 谁来用它
apps/vocab/vocab.klx          →   电脑上的单词服务（第 5 章，带上 server.klx）
                              →   手机里的库（第 6、8 章，不带 server.klx）

手机上的画面和上网                 不放进上面那份规则
apps/vocab/android/               Android 的按钮，用 OkHttp 上网
apps/vocab/ios/                   iPhone 的按钮，用 URLSession 上网
```

用到的词：

- **HTTP**：手机和电脑之间传文字的一种方式。浏览器打开网页用的就是它。
- **GET**：去读取。浏览单词是 GET。
- **POST**：去提交。点「认识」是 POST。
- **JSON**：一种写数据的文本，大括号、引号和冒号。例如 `{"known":"1","id":"2"}`。
- **壳**：只负责按钮和上网的那一层。Android 壳用 Kotlin 写，iOS 壳用 Swift 写。
- **核心**：单词规则，文件是 `vocab.klx`。手机库里没有自己的 HTTP 代码。
- **`.so`**：Android 运行时加载的共享库，文件名是 `libkylixvocab.so`。
- **`.a`**：iOS 这份示例使用的静态库，文件名是 `libkylixvocab.a`。只能在 Mac 上链出来。
- **端口**：同一台电脑上区分不同服务的数字。背单词是 **8091**。旁边那套登录后台是 **8090**。两个可以同时开。
- **127.0.0.1**：这台电脑自己。在 Android 模拟器里，它指的是模拟器自己，不是你的单词服务。

四个按钮和请求：

| 按钮 | 壳发出的请求 |
|---|---|
| 全部单词 | `GET /api/words?known=1,3` |
| 复习未掌握 | `GET /api/review?known=1,3` |
| 认识 | `POST /api/mark`，正文 `{"known":"1","id":"2"}` |
| 下一张 | 不再上网，只在已经拿到的卡片里换一张 |

`known=` 后面是那串编号。路径和 POST 正文由 Kylix 库算出来。

仓库里还有一套登录 + Notes，文件夹是 `apps/android/` 和 `apps/ios/`（路径里没有 `vocab`）。背单词要打开的是 `apps/vocab/android` 和 `apps/vocab/ios`。

---

## 3. 装上工具，编出 kylix，跑 doctor

这一步得到仓库根的 `kylix` 程序，并用 `kylix mobile doctor` 看缺什么。

**doctor** 是自带的检查。它看四段：背单词必需的工具、Android NDK、iOS（只有 Mac 才继续查）、这个例子用不到的库。最后一行决定能不能往下做。

三种记号：

| 记号 | 意思 |
|---|---|
| `✓` | 这一项找到了 |
| `−` | 只是说明。最后一行仍可以是 ready |
| `✗` | 这一项不行。最后一行会是 missing，退出码是 1 |

**PATH** 是一串文件夹名单。你只打 `go`，系统按这个名单去找名为 `go` 的程序。名单里没有，就会说 `command not found`。

**后端** 在这里的意思是：同一份 `.klx` 可以走两条路。`go` 先变成 Go 再编译；`llvm` 直接变成机器码。`check` 要证明两条路的结果一致。有 `clang` 就可以走 llvm。`llc` 是 LLVM 里的另一个程序，没有它时 doctor 打减号，改用 clang。

### 3.1 下载仓库

在你准备用来工作的终端里（Windows 用户：Ubuntu 窗口）：

```bash
cd ~
git clone https://github.com/astra-zhao/kylix.git
cd kylix
ls README_CN.md apps
```

`cd` 是进入文件夹。`~` 是你的用户主目录。`ls` 是列出文件。能看到 `README_CN.md` 和 `apps`，当前目录就是仓库根。

已经有这份文件夹时，不必再克隆。`cd` 进去，再 `ls README_CN.md`。编写时仓库已经在这台机器上，没有重新执行 `git clone`。

### 3.2 安装 Go、clang、curl、C 编译器

#### Linux，或 WSL 里的 Ubuntu

```bash
sudo apt update
sudo apt install -y golang-go git llvm clang build-essential curl unzip
```

`sudo` 会问 Linux 密码。`build-essential` 里有 `cc`。`llvm` 里有 `llc`。这台编写机器当时没装 `llc`，只有 clang，doctor 仍然 ready。上面的 `apt` 没有在这台已经装好 Go、clang、curl 的机器上重跑。你装上 `llvm` 之后，`llc` 那一行会是勾，而不是减号。下面改用官方 Go 压缩包的几行，这台机器也没有执行（这里的 `go` 已经是 1.25.0）。

若发行版自带的 Go 比 1.25 旧，到 https://go.dev/dl/ 下载 Linux 的 `go1.25.0.linux-amd64.tar.gz`（或更新的 1.25），然后：

```bash
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.25.0.linux-amd64.tar.gz
echo 'export PATH=/usr/local/go/bin:$PATH' >> ~/.bashrc
source ~/.bashrc
```

`source ~/.bashrc` 让刚才写进配置的 PATH 立刻生效。

#### macOS

先安装 Homebrew（https://brew.sh 首页的那一行）。然后：

```bash
brew install go llvm git curl
echo 'export PATH="$(brew --prefix llvm)/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
xcode-select --install
```

最后一行安装命令行工具，里面有 `cc`。已经装过会说已安装。

Homebrew 的 LLVM 默认不进 PATH。上面把 `llvm` 的 `bin` 加进了 `~/.zshrc`。编写环境是 Linux，**没有**执行 `brew`。

### 3.3 看版本，并编出 kylix

仓库根：

```bash
go version
clang --version
curl --version
cc --version
go build -o kylix ./cmd/kylix/
./kylix version
```

这台机器上，`go version` 的整行是 `go version go1.25.0 linux/amd64`。`clang --version` 和 `cc --version` 的第一行都是 `Ubuntu clang version 18.1.3 (1ubuntu1)`（这里的 `cc` 就是 clang）。`llc --version` 是 `command not found`。`./kylix version` 打印：

```text
Kylix 0.16.0
```

`go.mod` 写的是 Go 1.25.0。请用 1.25.0 或更新的 Go。文件名是 `kylix`，就在仓库根。它已经在 `.gitignore` 里，不用提交。

### 3.4 跑 doctor

NDK 还没装也可以跑。没装时最后一行仍是 ready，第 4 章不需要 NDK。第 6 章才要它。

这台机器已经把 `ANDROID_NDK_HOME` 设成 `/tmp/ndk/android-ndk-r26d`。仓库根执行 `./kylix mobile doctor`，退出码 0，原文是：

```text
=== vocab required ===
  ✓ go: /usr/bin/go
  − llc: not found (IR→object via clang -x ir)
  ✓ clang: /usr/bin/clang
  ✓ curl: /usr/bin/curl
  ✓ cc: /usr/bin/cc
  ✓ bash: /usr/bin/bash

=== Android NDK ===
  host prebuilt directory: linux-x86_64
  ✓ NDK found: /tmp/ndk/android-ndk-r26d
  ✓ clang: /tmp/ndk/android-ndk-r26d/toolchains/llvm/prebuilt/linux-x86_64/bin/clang

=== iOS (macOS only) ===
  − not this host. iOS linking needs macOS, a full Xcode install, and xcodegen.

=== not used by the vocab example ===
  − libcurl does not affect the vocab example (the library, not the curl command; db / httpclient would link it)
  − libgc does not affect the vocab example (only --gc=boehm links it)

✓ mobile doctor: vocab tools are ready.
```

怎么读：

- `− llc`：这台机器没有 `llc`。括号里写着用 clang 把中间结果变成目标文件。有 `✓ clang` 就可以做第 4 章。
- `host prebuilt directory: linux-x86_64`：NDK 压缩包里，给这台电脑用的编译器放在这个名字的文件夹。macOS 上是 `darwin-x86_64`，Windows 上是 `windows-x86_64`。
- iOS 那一行减号：这台不是 Mac。做 Android 可以继续。做 iOS 要换 Mac，见第 8 章。
- libcurl / libgc 那两行减号：它们是库，不是终端里的 `curl` 命令。背单词不链它们。看到减号，继续即可。

最后一行是 `✓ mobile doctor: vocab tools are ready.` 就可以做第 4 章。若最后一行是 `✗ mobile doctor: missing required tooling (see above).`，往上找带 `✗` 的那一项，按第 9 章处理。

### 这一章做完时

- [ ] `./kylix version` 打印 `Kylix 0.16.0`
- [ ] `./kylix mobile doctor` 的最后一行是 `vocab tools are ready`

---

## 4. 不打开手机：kylix mobile check

`check` 做两段，然后自己关掉服务，不占着 8091：

1. 用 Go 和 LLVM 各编译一次单词核心，比较输出是否一致。再做成一份电脑能直接打开的 `.so`，用 C 去调用浏览、把 1 标成认识、复习时不再出现 apple。
2. 在 8091 上拉起单词服务，用 curl 对同一份结果，然后停掉服务。

**C ABI** 的意思是：函数用 C 语言能调用的名字露出来。Kotlin 和 Swift 叫的是 `vc_browse` 这种名字。进阶里有对照表。

仓库根：

```bash
./kylix mobile check
```

这台机器的原文（退出码 0）：

```text
[vocab] parity Go vs LLVM
[vocab] parity PASS
[vocab] C ABI .so
ok words_path
ok review_path
ok mark_path
ok words_query
ok review_query
ok browse_apple
ok browse_zh
ok browse_left
ok browse_known
ok mark_known
ok mark_left
ok mark_apple_known
ok review_hides_apple
ok review_book
ok review_left
ok mark_again
ok mark_bad
ok mark_req
ok parse_keep
ok parse_fail
c abi PASS
[vocab] host check PASS
[vocab] expected bodies from Go parity
[vocab] e2e backend=go
[vocab] e2e PASS (go)
```

`review_hides_apple` 表示编号 1 已经认识之后，复习结果里不再含有 apple。最后一行必须是 `[vocab] e2e PASS (go)`。

这还不是手机模拟器。服务在命令结束时已经停掉，所以还不能给第 7 章用。下一章才把它留在前台。

再用 LLVM 编出来的服务做同一套 curl（这台机器跑过，退出码 0）：

```bash
./kylix mobile check --e2e-only --backend=llvm
```

```text
[vocab] expected bodies from Go parity
[vocab] e2e backend=llvm
[vocab] e2e PASS (llvm)
```

### 这一章做完时

- [ ] 命令在仓库根
- [ ] `./kylix mobile check` 的最后一行是 `[vocab] e2e PASS (go)`
- [ ] 我知道模拟器还没打开，8091 也没有留着

---

## 5. 把服务留在前台：kylix mobile serve

**serve** 编译单词服务，并让它一直占着这个窗口，直到你按 Ctrl+C。模拟器要连的就是它。

新开一个终端，进入仓库根：

```bash
./kylix mobile serve
```

这台机器上，起来之后的两行是（日期会不同）：

```text
vocab server on :8091
2026/10/11 02:36:15 🚀 KylixBoot started on http://localhost:8091
```

看到这两行之后，窗口停在这里是正常的。服务要一直占着它。

另开一个终端，仍在仓库根，自己发三次请求。这台机器的正文如下。

第一次，浏览全部：

```bash
curl -sS -D - http://127.0.0.1:8091/api/words -o /tmp/vocab-words.txt
```

响应头里有：

```text
HTTP/1.1 200 OK
Content-Type: text/plain; charset=utf-8
```

`Content-Type` 是 `text/plain`，不是 `application/json`。正文是 JSON。壳只读正文。`/tmp/vocab-words.txt` 的内容是：

```text
{"ok":true,"known":"","total":6,"left":6,"cards":[{"id":1,"en":"apple","zh":"苹果","known":false},{"id":2,"en":"book","zh":"书","known":false},{"id":3,"en":"water","zh":"水","known":false},{"id":4,"en":"friend","zh":"朋友","known":false},{"id":5,"en":"time","zh":"时间","known":false},{"id":6,"en":"happy","zh":"高兴","known":false}]}
```

里面有 `"left":6`、`apple`、`"known":""`。

第二次，把编号 1 标成认识：

```bash
curl -sS -X POST http://127.0.0.1:8091/api/mark \
  -H 'Content-Type: application/json' \
  --data '{"known":"","id":"1"}'
```

这台机器打出：

```text
{"ok":true,"known":"1","total":6,"left":5,"cards":[{"id":1,"en":"apple","zh":"苹果","known":true},{"id":2,"en":"book","zh":"书","known":false},{"id":3,"en":"water","zh":"水","known":false},{"id":4,"en":"friend","zh":"朋友","known":false},{"id":5,"en":"time","zh":"时间","known":false},{"id":6,"en":"happy","zh":"高兴","known":false}]}
```

`"known":"1"`，`"left":5`，apple 那一项是 `"known":true`。

第三次，复习还没掌握的（带上 `known=1`）：

```bash
curl -sS 'http://127.0.0.1:8091/api/review?known=1'
```

```text
{"ok":true,"known":"1","total":6,"left":5,"cards":[{"id":2,"en":"book","zh":"书","known":false},{"id":3,"en":"water","zh":"水","known":false},{"id":4,"en":"friend","zh":"朋友","known":false},{"id":5,"en":"time","zh":"时间","known":false},{"id":6,"en":"happy","zh":"高兴","known":false}]}
```

有 book，没有 apple。

要停掉服务，回到跑着 `serve` 的那个窗口，按 **Ctrl+C**。这台机器上对那个进程组发送 SIGINT 之后，`curl http://127.0.0.1:8091/api/words` 不再连得上。

停掉之后，仓库根可能留下 `kylixvocab_gen/work-数字/`，里面是刚才编出来的服务程序。这个目录在 `.gitignore` 里，不用提交。想清掉就删掉整个 `kylixvocab_gen`。

第 7 章要点模拟器时，把 `serve` 重新开着，窗口停在 `vocab server on :8091`。

### 这一章做完时

- [ ] 有一个终端停在 `vocab server on :8091`
- [ ] 第一次 curl 能看到 apple 和 `"left":6`
- [ ] 标记之后的复习 curl 看不到 apple

---

## 6. NDK 和两种模拟器：build android

这一步编出 `libkylixvocab.so`，放进 Android 工程指定的文件夹。Android Studio 只负责打包，不会替你生成这份 Kylix 库。

### NDK 是什么

**NDK**（Native Development Kit）是 Google 提供的一套 Android C 编译器。Kylix 用它把库链接成手机能加载的 `.so`。它和 Android Studio 向导里下载的 Android SDK 是两套东西。SDK 用来编译 Kotlin 壳；NDK 用来链接这份 `.so`。

**ANDROID_NDK_HOME** 是一个环境变量，告诉程序「NDK 解压在哪个文件夹」。环境变量是终端记住的一个名字和对应的文字。

**ANDROID_NDK_ROOT** 是另一个名字，有些安装脚本会设置它。规则是：

- HOME 一旦有内容，就只搜 HOME 这一处。ROOT 即使指着另一份好的 NDK，也不会拿来用。
- HOME 是空的时候，才把 ROOT 当作唯一的搜索路径。
- 两个都空着，才去看几个默认位置（例如 `/opt/android-ndk`）。这时 doctor 的退出码仍是 0，因为第 4 章不需要 NDK。`build android` 会失败。

GitHub 的 Ubuntu 镜像会预置 ROOT。你在自己电脑上如果两个都设了，以 HOME 为准。指错时的原文在第 9 章。

### ABI：模拟器和库必须是同一种 CPU

**ABI** 是 CPU 种类。库和模拟器不一致时，应用一打开就会因为加载 `.so` 失败而闪退。

| 你要跑的设备 | 命令里写 | 库出现的文件夹 | `file` 那一行里要有 |
|---|---|---|---|
| 电脑上最常见的 x86_64 模拟器 | `amd64` | `jniLibs/x86_64/` | `x86-64` |
| arm64 模拟器，或多数真机 | `arm64` | `jniLibs/arm64-v8a/` | `ARM aarch64` |

命令里写 `amd64`，给的是 x86_64 模拟器。写 `x86_64` 会被拒绝，并告诉你改用 `amd64`。两种可以各编一次，两个文件夹互不覆盖。模拟器是哪一种，就至少要有对应的那一份。

**triple** 是编译器内部用的全名。你写 `android/amd64`，工具把它变成给 Android x86_64 用的那一串。你现在只要选对 `amd64` 或 `arm64`。

### 6.1 下载 NDK r26d

请用和你的终端一致的压缩包。WSL 的 Ubuntu 下 **linux.zip**。把 Windows 的 zip 解压进 Ubuntu，clang 会在 `windows-x86_64` 里，Linux 上的 Kylix 找不到它。

| 终端 | 下载 | 解压后 clang 在 |
|---|---|---|
| Linux，或 WSL 的 Ubuntu | https://dl.google.com/android/repository/android-ndk-r26d-linux.zip | `toolchains/llvm/prebuilt/linux-x86_64/bin/clang` |
| macOS | https://dl.google.com/android/repository/android-ndk-r26d-darwin.dmg | `toolchains/llvm/prebuilt/darwin-x86_64/bin/clang` |

在 Linux 或 WSL 里：

```bash
cd ~
curl -L -o ndk.zip https://dl.google.com/android/repository/android-ndk-r26d-linux.zip
unzip -q ndk.zip
ls "$HOME/android-ndk-r26d/toolchains/llvm/prebuilt/linux-x86_64/bin/clang"
```

`ls` 能打印出 clang 的路径，才算解压对了。macOS 打开 dmg，把 `android-ndk-r26d` 拖到主目录，再检查 `darwin-x86_64` 那条路径。编写环境没有重新下载，用的是已经放在 `/tmp/ndk/android-ndk-r26d` 的 r26d。

### 6.2 设置 ANDROID_NDK_HOME

Linux / WSL 写进 `~/.bashrc`，macOS 写进 `~/.zshrc`。把路径换成你的文件夹。macOS 把 `linux-x86_64` 改成 `darwin-x86_64`。

```bash
echo 'export ANDROID_NDK_HOME="$HOME/android-ndk-r26d"' >> ~/.bashrc
source ~/.bashrc
echo "$ANDROID_NDK_HOME"
ls "$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/linux-x86_64/bin/clang"
```

`echo` 打印你设的路径，`ls` 打印 clang。编写时没有改 `~/.bashrc`，是在当前终端里导出 `ANDROID_NDK_HOME=/tmp/ndk/android-ndk-r26d`，并用 `ls` 确认了那个 clang。然后回到仓库根再跑一次：

```bash
./kylix mobile doctor
```

Android 那一段应是 `✓ NDK found:` 加上你的路径，并且 clang 路径里含有 `linux-x86_64`（Mac 上是 `darwin-x86_64`）。最后一行仍是 `vocab tools are ready`。

### 6.3 编译

仓库根。x86_64 模拟器：

```bash
./kylix mobile build android amd64
```

这台机器的原文（退出码 0）。左边的地址每次编译可以不同，符号名字和最后三行要在：

```text
[vocab-android] android/amd64 → /workspace/apps/vocab/android/app/src/main/jniLibs/x86_64/libkylixvocab.so
/workspace/apps/vocab/android/app/src/main/jniLibs/x86_64/libkylixvocab.so: ELF 64-bit LSB shared object, x86-64, version 1 (SYSV), dynamically linked, not stripped
00000000000082f0 T kylix_free
00000000000081b0 T vc_apply_mark
00000000000080f0 T vc_browse
0000000000008070 T vc_mark_path
0000000000008170 T vc_mark_request
00000000000081f0 T vc_parse
0000000000008130 T vc_review
0000000000008050 T vc_review_path
00000000000080c0 T vc_review_query
0000000000008030 T vc_words_path
0000000000008090 T vc_words_query
jni link OK (amd64)
[vocab-android] wrote /workspace/apps/vocab/android/app/src/main/jniLibs/x86_64/libkylixvocab.so
x86_64 emulators use the amd64 build (jniLibs/x86_64); arm64 emulators use the arm64 build (jniLibs/arm64-v8a).
android artifact OK (amd64)
```

`/workspace` 是这台机器的仓库根。你的 `wrote` 路径以你的仓库根开头，以 `jniLibs/x86_64/libkylixvocab.so` 结尾。

中间的 `T vc_browse` 这类行是动态符号。`jni link OK` 表示 NDK 把很小的 JNI 桥和这份库链上了。**JNI** 是 Android 让 Kotlin 调用 C 函数的通道。

那句英文的意思是：x86_64 模拟器用 amd64 这份（`jniLibs/x86_64`）；arm64 模拟器用 arm64 这份（`jniLibs/arm64-v8a`）。

arm64 模拟器或真机：

```bash
./kylix mobile build android arm64
```

这台机器的原文（退出码 0）：

```text
[vocab-android] android/arm64 → /workspace/apps/vocab/android/app/src/main/jniLibs/arm64-v8a/libkylixvocab.so
/workspace/apps/vocab/android/app/src/main/jniLibs/arm64-v8a/libkylixvocab.so: ELF 64-bit LSB shared object, ARM aarch64, version 1 (SYSV), dynamically linked, not stripped
0000000000007460 T kylix_free
0000000000007314 T vc_apply_mark
0000000000007270 T vc_browse
00000000000071e4 T vc_mark_path
00000000000072dc T vc_mark_request
000000000000734c T vc_parse
00000000000072a4 T vc_review
00000000000071b8 T vc_review_path
0000000000007240 T vc_review_query
000000000000718c T vc_words_path
0000000000007210 T vc_words_query
jni link OK (arm64)
[vocab-android] wrote /workspace/apps/vocab/android/app/src/main/jniLibs/arm64-v8a/libkylixvocab.so
x86_64 emulators use the amd64 build (jniLibs/x86_64); arm64 emulators use the arm64 build (jniLibs/arm64-v8a).
android artifact OK (arm64)
```

手机库的源文件是 `stdlib/stringutil.klx`、`apps/vocab/vocab.klx`、`apps/vocab/vocab_lib.klx`。里面没有 `server.klx`。手机库不含那个网站。`.ll` / `.o` 写在临时目录里，不留在 `apps/vocab` 源码旁边。`.so` 在 `.gitignore` 里，不用提交。

还没有 Android Studio 时，第 4–6 章做完，核心和库就已经在了。第 7 章可以以后再点。

### 这一章做完时

- [ ] `echo "$ANDROID_NDK_HOME"` 不是空行，对应系统的 `bin/clang` 存在
- [ ] 模拟器是 x86_64 时，最后一行是 `android artifact OK (amd64)`，`file` 那一行含 `x86-64`
- [ ] 模拟器是 arm64 时，最后一行是 `android artifact OK (arm64)`，`file` 那一行含 `ARM aarch64`

---

## 7. 在 Android Studio 里打开

这一章把壳装进模拟器，连上第 5 章那个还开着的单词服务。

**模拟器（emulator）** 是在电脑里跑的一台假 Android 手机。它不是第 8 章的 iOS Simulator。

**Gradle** 是 Android 工程用来下载依赖、编译壳、打出安装包的构建系统。`apps/vocab/android/gradlew` 在仓库里不存在（这台机器上 `ls` 确认过）。用 Android Studio 自己的 Gradle，不写 `./gradlew`。

**这一章的点击没有在编写环境里做。** 按钮名称来自 `activity_main.xml` 和 `AndroidManifest.xml`，按 Android Studio 默认英文界面写。菜单是中文时，看括号里的意思。

开始之前：

- 第 5 章的终端仍停在 `vocab server on :8091`。
- 第 6 章的 `.so` 已经在和模拟器 CPU 相同的 `jniLibs` 文件夹里。

### 7.1 安装 Android Studio

1. 打开 https://developer.android.com/studio ，下载与你的系统对应的安装包。Windows 用户装在 Windows 上。
2. 安装时用默认选项。第一次启动的向导里选 **Standard**，许可证页面点 **Accept**，等它下完 SDK。
3. 向导下载的是 Android SDK。第 6 章的 NDK 仍要保留。

### 7.2 打开正确的文件夹

1. 欢迎页点 **Open**（打开）。已经在工程里时，用菜单 **File → Open**。
2. 选中 `apps/vocab/android`。这个文件夹里有 `settings.gradle.kts`。
3. 若出现 **Trust and Open**（信任并打开），点信任。
4. 右下角开始同步。第一次会下载依赖，需要网络。等到结束、没有红色报错。

同步时如果弹出这些窗口：

| 提示 | 你点什么 |
|---|---|
| 缺少 Android SDK Platform 34 | 点安装 / Accept。工程的 `compileSdk` 是 34 |
| 选择 JDK，或 Gradle JDK | 选 **Embedded JDK** 或标着 17 的那一项。工程按 JDK 17 写的 |
| 要不要用 Gradle Wrapper | 选 Studio 建议的默认项。仓库里没有现成的 `gradlew` |

同步说缺 NDK 或 CMake 时：打开 **File → Settings**（macOS 是 **Android Studio → Settings**）→ **Languages & Frameworks → Android SDK** → **SDK Tools**，勾选 **NDK (Side by side)** 和 **CMake**，点 Apply。Studio 用这份 NDK 编译很小的 JNI 桥。Kylix 的 `.so` 仍是第 6 章用 `ANDROID_NDK_HOME` 编的那份。

若同步或编译说缺 `libkylixvocab.so`，回到第 6 章，看 `wrote` 的路径是 `x86_64` 还是 `arm64-v8a`，和报错里的文件夹是否同一个。

### 7.3 建立模拟器

1. 右上角手机形状的图标 **Device Manager**（设备管理器）。找不到时用菜单 **Tools → Device Manager**。
2. 点 **Create Device**（创建虚拟设备）。
3. 选一台手机，例如 **Pixel 6**，点 **Next**。
4. 选系统镜像。你编的是 `amd64`，这里选名字里带 **x86_64** 的镜像（API 34 即可）。你只编了 `arm64`，这里选 **arm64-v8a**。旁边若是 **Download**，先下载再 Next。
5. 点 **Finish**。在设备列表里点播放三角形，等它开机到桌面。

### 7.4 运行

1. 窗口上方绿色三角形 **Run**。调试安装用 Studio 自带的 debug 签名。
2. 若弹出设备列表，选刚才的模拟器。
3. 等到模拟器里出现标题 **背单词**。启动器上的名字是 **Kylix Vocab**（写在 `AndroidManifest.xml` 的 `android:label`）。应用 id 是 `dev.kylix.vocab`。

界面上「Server URL」的默认文字来自布局文件：

```text
http://10.0.2.2:8091
```

**10.0.2.2** 是官方 Android 模拟器看「跑着模拟器的那台电脑」的地址。在模拟器里，`127.0.0.1` 是模拟器自己。保持 `10.0.2.2`，除非你是真机。

真机（同一 Wi-Fi）：把地址改成电脑的局域网 IP，端口仍是 8091。这个示例允许明文 HTTP（`usesCleartextTraffic`），只适合你自己的电脑。

然后按这个顺序点：

1. **全部单词**。第一张应是 `apple` / `苹果`。进度是「还剩 6 / 6」。
2. **认识**。进度变成还剩 5。这张下面出现已认识。
3. **下一张**。只换卡片，不再上网。
4. **复习未掌握**。apple 不再作为卡片出现，能看到 book。

关掉应用再打开，复习列表仍不含 apple。编号存在应用自己的本地存储里。

这些点击没有在编写环境里做过。卡片上的字和第 5 章 curl 的 JSON 是同一份字段：`en`、`zh`、`left`。

### 7.5 Windows + WSL

这次没有在 Windows 上实测。请先用 **Windows 自己的浏览器** 打开 `http://127.0.0.1:8091/api/words`。

- 浏览器里能看到带 `apple` 的 JSON：再在模拟器里用 `http://10.0.2.2:8091`。较新的 WSL2 上，Windows 的 `127.0.0.1` 会转到 Ubuntu 里的服务。
- 浏览器打不开，但 Ubuntu 里的 curl 可以：服务只在 WSL 里听得见。到 Windows 的「WSL 设置」确认 localhost 转发是开的。先把 Windows 浏览器打通，再点模拟器。

Android Studio 打开 WSL 里的工程时，路径类似 `\\wsl$\Ubuntu\home\<你的用户名>\kylix\apps\vocab\android`。

### 这一章做完时

- [ ] 打开的文件夹以 `apps/vocab/android` 结尾，里面有 `settings.gradle.kts`
- [ ] 模拟器 CPU 和第 6 章的 `amd64` 或 `arm64` 一致
- [ ] `serve` 仍停在 8091
- [ ] 模拟器里看到 apple / 苹果；点认识后再点复习，apple 不在复习卡片里

---

## 8. iOS：这台 Linux 上会停住

iOS 的库必须在 **macOS** 上链接。需要三样东西：

1. 一台 Mac。
2. **完整的 Xcode**（App Store 里的应用，不只是命令行工具）。`xcode-select -p` 用来确认它在哪。
3. **xcodegen**（`brew install xcodegen`）。它根据 `apps/vocab/ios/project.yml` 生成 Xcode 工程。链接 `.a` 本身不调用 xcodegen；没有它，就没有 `KylixVocab.xcodeproj` 可打开。

在 Mac 上，doctor 的 iOS 段会去看 `xcode-select -p` 和 `xcodegen`。缺了是减号，不因此让 doctor 失败。编库失败的原因会写在 `build ios` 的输出里。这段是程序里的行为，这台 Linux 没有 Mac 可对照。

### 8.1 在这台 Linux 上实际看到的

仓库根，这两条的输出相同，退出码都是 1：

```bash
./kylix mobile build ios simulator
./kylix mobile build ios device
```

```text
iOS archives must be linked on macOS with Xcode command line tools.
On this host, stop here. The Swift sources are in apps/vocab/ios/.
ios cross-link requires a macOS host with Xcode command line tools.
Install a full Xcode (xcode-select -p) and xcodegen (brew install xcodegen) on that Mac. This host does not write a .a.
```

这台机器上 `find apps/vocab -name 'libkylixvocab.a'` 没有结果。看到这四行，表示程序停对了。Swift 源码还在 `apps/vocab/ios/`，换到 Mac 上再编。

直接 `kylix build --target ios/simulator-arm64` 在非 Mac 上也会在写出文件之前失败，原因同一段：先判断主机是不是 macOS。

### 8.2 到了 Mac 上再做（这里没有执行）

下面从「在 Mac 上」开始的命令和点击，**没有**在编写环境里执行。成功时程序在链接完成之后打印的最后一行是 `ios artifact OK (simulator)` 或 `ios artifact OK (device)`，输出里有 `libkylixvocab.a`。这是源码里的打印，不是这台 Linux 的运行记录。

1. 从 App Store 安装完整 Xcode，打开一次，接受协议。
2. 终端：

```bash
sudo xcode-select -s /Applications/Xcode.app
xcode-select -p
brew install xcodegen
xcodegen --version
```

`xcode-select -p` 应打印 `/Applications/Xcode.app/Contents/Developer`。`xcodegen --version` 有版本号即可。

3. 第 3–5 章在这台 Mac 上做完，`./kylix mobile serve` 停在 8091。`clang` 要在 PATH 里。
4. 仓库根编库：

```bash
./kylix mobile build ios simulator
```

真机用 `./kylix mobile build ios device`。两次写的是**同一个文件** `apps/vocab/ios/Sources/CKylixVocab/lib/libkylixvocab.a`。后一次盖掉前一次。要跑模拟器，最后一次是 `simulator`。这个 `.a` 不进 Git。

5. 生成并打开工程：

```bash
cd apps/vocab/ios
xcodegen generate
open KylixVocab.xcodeproj
```

打开的是 `KylixVocab.xcodeproj`。`project.yml` 里目标名是 **KylixVocab**，bundle id 是 `dev.kylix.vocab`，最低系统是 iOS 16。

在 Xcode 里：

1. 左侧点蓝色工程图标，选中目标 **KylixVocab**。
2. **Signing & Capabilities** 里 **Team** 选你的 Apple ID。没有账号时：菜单 **Xcode → Settings → Accounts**，左下角 **+**，用 Apple ID 登录，再回到 Team 里选中它。模拟器一般用免费账号即可。
3. 窗口上方 scheme 选 **KylixVocab**，设备选一台 **iPhone 模拟器**（系统 16 或更高）。
4. 按左上角三角形 **Run**（或 `Cmd+R`）。

模拟器和 Mac 共用网络，所以服务器一栏保持：

```text
http://127.0.0.1:8091
```

`10.0.2.2` 是 Android 模拟器的地址，在这里不适用。按钮和 Android 相同：全部单词、认识、下一张、复习未掌握。这些名字写在 `VocabApp.swift` 里。

真机：先 `./kylix mobile build ios device`，再 Run，把服务器改成这台 Mac 的局域网地址。手机要信任你的开发者证书。这一步没有在编写环境里做过。

### 这一章做完时

- [ ] 不是 Mac 时，我看到了上面四行英文，并且没有 `.a`
- [ ] 到了 Mac 之后，`build ios simulator` 的最后一行是 `ios artifact OK (simulator)`，然后才打开 Xcode

---

## 9. 排错：对着原文看

先看最后一行，再往上找 `✗`。下面每一段都是这台 Linux 上的原文。路径里的 `/tmp/...` 和 `/workspace` 换成你自己的。

### 最后一行是 ready，但 NDK 是减号

两个环境变量都空着时，`./kylix mobile doctor` 的退出码是 0。这台机器上 Android 那一段是：

```text
=== Android NDK ===
  host prebuilt directory: linux-x86_64
  − NDK: not found (set ANDROID_NDK_HOME to build the Android .so). The flashcard host check does not need it.
    android cross-link needs Android NDK (not found): install the NDK and point ANDROID_NDK_HOME at it (it must contain toolchains/llvm/prebuilt/linux-x86_64/bin/clang)
    host prebuilt directory: linux-x86_64
    tried:
      /opt/android-ndk (/opt/android-ndk)
        expected clang: /opt/android-ndk/toolchains/llvm/prebuilt/linux-x86_64/bin/clang
```

`tried` 里的 `/opt/android-ndk` 是默认位置，这台机器上那个文件夹不存在。第 4 章可以继续。要做第 6 章，按 6.1、6.2 设上 HOME，再跑 doctor，直到出现 `✓ NDK found:`。

这时如果直接构建，退出码是 1，原文是：

```text
android cross-link needs Android NDK (not found): install the NDK and point ANDROID_NDK_HOME at it (it must contain toolchains/llvm/prebuilt/linux-x86_64/bin/clang)
host prebuilt directory: linux-x86_64
tried:
  /opt/android-ndk (/opt/android-ndk)
    expected clang: /opt/android-ndk/toolchains/llvm/prebuilt/linux-x86_64/bin/clang
```

命令是 `env -u ANDROID_NDK_HOME -u ANDROID_NDK_ROOT ./kylix mobile build android amd64`。你平时只要看到同样的 “not found” 和 `linux-x86_64`，就回去设 HOME。

### HOME 指错了，ROOT 还指着一份好的 NDK

这是最容易看走眼的一种。这台机器上 HOME 是 `/tmp/kylix-ndk-missing`（这个文件夹不存在），ROOT 是 `/tmp/ndk/android-ndk-r26d`（里面有 clang）。`./kylix mobile doctor` 的退出码是 **1**，不会改用 ROOT。Android 段和最后一行：

```text
=== Android NDK ===
  host prebuilt directory: linux-x86_64
  ✗ NDK: ANDROID_NDK_HOME or ANDROID_NDK_ROOT does not contain the host clang
    android cross-link needs Android NDK (not found): install the NDK and point ANDROID_NDK_HOME at it (it must contain toolchains/llvm/prebuilt/linux-x86_64/bin/clang)
    host prebuilt directory: linux-x86_64
    tried:
      /tmp/kylix-ndk-missing (ANDROID_NDK_HOME)
        expected clang: /tmp/kylix-ndk-missing/toolchains/llvm/prebuilt/linux-x86_64/bin/clang
    ANDROID_NDK_ROOT (/tmp/ndk/android-ndk-r26d) was not searched because ANDROID_NDK_HOME is set

✗ mobile doctor: missing required tooling (see above).
```

做法：`echo "$ANDROID_NDK_HOME"` 看它是不是你解压的那个文件夹。改 HOME，`source ~/.bashrc`，新开一个终端再跑 doctor。那一行 “was not searched” 的意思是：ROOT 被故意放下了，因为 HOME 已经设置。把 HOME 改对，或者在你确定要只用 ROOT 时把 HOME 清空（`unset ANDROID_NDK_HOME`）。两个都设着的时候，以 HOME 为准。

### 目录里只有 windows-x86_64

WSL 里用了 Windows 版 NDK 时，预编译目录的名字对不上。这台机器上造了一个只有 `windows-x86_64` 的文件夹作为 HOME，ROOT 仍指着好的 Linux NDK。doctor 退出码 1，关键几行是：

```text
    tried:
      /tmp/kylix-ndk-windows (ANDROID_NDK_HOME)
        expected clang: /tmp/kylix-ndk-windows/toolchains/llvm/prebuilt/linux-x86_64/bin/clang
        prebuilt directories present: windows-x86_64
    ANDROID_NDK_ROOT (/tmp/ndk/android-ndk-r26d) was not searched because ANDROID_NDK_HOME is set
```

`expected clang` 要的是 `linux-x86_64`。`prebuilt directories present` 写着实际有的是 `windows-x86_64`。重新下载 **linux.zip**，让 HOME 指到新解压的文件夹。Mac 上如果看到 `windows-x86_64` 或只有别的名字、没有 `darwin-x86_64`，同样是下错了系统的压缩包。

### 把 x86_64 写进了子命令

```bash
./kylix mobile build android x86_64
```

退出码 1：

```text
kylix mobile build android: ABI must be amd64 or arm64 (got "x86_64")
x86_64 emulators use the amd64 build (jniLibs/x86_64); arm64 emulators use the arm64 build (jniLibs/arm64-v8a).
```

x86_64 模拟器使用 `./kylix mobile build android amd64`。

### 8091 已经被占用

`serve` 还开着时再跑 `./kylix mobile check --e2e-only`，退出码 1，整段输出是：

```text
port 8091 is already in use; stop that process and retry
```

回到停在 `vocab server on :8091` 的窗口，按 Ctrl+C，然后再跑 check。若你就是要给模拟器用，保持 serve，不要再跑会占用端口的 check。

### 当前目录不是仓库

在 `/tmp` 里执行仓库根的 `kylix`，退出码 1：

```text
kylix mobile: run this from a kylix checkout (apps/vocab/vocab.klx not found)
```

`cd` 到能 `ls README_CN.md` 的目录再跑。

### 其他常见行

| 你看到的 | 做法 |
|---|---|
| `go: command not found` | 按 3.2 安装 Go，或把 `/usr/local/go/bin` 加进 PATH，再 `go version` |
| `clang: not found`，或最后一行是 missing，上面有 `✗ clang` | 安装 `clang`。只有 `− llc` 而 `✓ clang` 在，可以继续 |
| `✗ curl: not found (the curl command, used by mobile check)` | 安装终端命令 `curl`。这和下面 libcurl 那一行减号不是一回事 |
| `cc -ldl failed (install a C compiler)` | Linux 安装 `build-essential`。macOS 执行 `xcode-select --install` |
| `PARITY DIFF` 然后两大段文字 | Go 和 LLVM 的输出不一致。先停在第 4 章，把输出留下来再查 |
| `✗ mobile doctor: missing required tooling (see above).` | 往上找 `✗`。减号那几行（llc、iOS、libcurl、libgc、没装 NDK）单独出现时，最后一行仍可以是 ready |
| Linux 上 iOS 四行 “must be linked on macOS” | 正常。换 Mac，见第 8 章 |
| 模拟器里一打开就闪退，Logcat 有 `UnsatisfiedLinkError` 和 `libkylixvocab.so` | 库的 CPU 和模拟器不一致。x86_64 模拟器要第 6 章的 `amd64`。这一条点击没有在编写环境里复现，按 ABI 表处理 |
| 模拟器里 `Failed to connect` | `serve` 没开，或地址写成了模拟器自己的 `127.0.0.1`。Android 模拟器用 `http://10.0.2.2:8091`。先在跑服务的系统上 curl 第 5 章的第一条 |

符号已经对上、但这次没设 HOME 时，检查已有的 `.so` 仍会成功，并多一行说明跳过了 JNI。这台机器上 `env -u ANDROID_NDK_HOME -u ANDROID_NDK_ROOT ./kylix mobile artifact android apps/vocab/android/app/src/main/jniLibs/x86_64/libkylixvocab.so amd64` 的结尾是：

```text
ANDROID_NDK_HOME is unset; skipped the JNI link (symbol table already matched)
x86_64 emulators use the amd64 build (jniLibs/x86_64); arm64 emulators use the arm64 build (jniLibs/arm64-v8a).
android artifact OK (amd64)
```

库的符号是对的。要连 JNI 桥再链一次，设上 HOME，用第 6 章的 `build android` 重编。

---

## 10. 进阶

日常做到第 6 章或第 7 章就可以停。下面这些是同一套东西的另一面，需要时再看。

### 脚本和子命令是同一条路

`apps/vocab/host_check.sh` 最后执行的是 `kylix mobile check --host-only`。`e2e.sh` 执行 `kylix mobile check --e2e-only --backend=...`，`--backend` 来自环境变量 `VOCAB_BACKEND`（不设就是 `go`）。`apps/vocab/android/build_core.sh` 执行 `build android`。脚本另外接受 `x86_64`，并把它换成 `amd64`；子命令本身不接受 `x86_64`，第 9 章有原文。

只做前半段、不拉起 8091：

```bash
./kylix mobile check --host-only
```

这台机器上，它在 `c abi PASS` 之后结束，最后一行是 `[vocab] host check PASS`，没有 `e2e PASS`。

已经编好的 `.so` 可以再查一次（这台机器跑过，设了 NDK 时结尾有 `jni link OK (amd64)`，最后一行是 `android artifact OK (amd64)`）：

```bash
./kylix mobile artifact android \
  apps/vocab/android/app/src/main/jniLibs/x86_64/libkylixvocab.so amd64
```

### 壳调用的 C 名字

壳不调用 `VcDeckJSON` 这种 Kylix 名字。`apps/vocab/vocab_lib.klx` 里写了 `[Export('vc_browse')]` 这类标注。第 6 章 `nm` 打出来的就是这些名字，外加 `kylix_free`。头文件在 `apps/vocab/ios/Sources/CKylixVocab/include/kylixvocab.h`。

| C 函数 | 作用 |
|---|---|
| `vc_words_query` / `vc_review_query` | 返回 `/api/words?known=1,3` 这种路径 |
| `vc_mark_path` | 返回 `/api/mark` |
| `vc_mark_request` | 拼出 POST 的 JSON |
| `vc_parse` | 看 HTTP 状态和正文 |
| `vc_browse` / `vc_review` / `vc_apply_mark` | 同一份卡片 JSON |
| `kylix_free` | 释放上面那些函数返回的字符串 |

这些函数返回的字符串是新分配的。壳把文字抄进自己的字符串之后调用 `kylix_free`。Android 的 `kylix_jni.c` 和 iOS 的 `KylixVocab.swift` 已经在复制之后调用了。你写界面时不用自己再调一次。

| 你写的目标 | 实际机器 |
|---|---|
| `android/amd64` | x86_64 模拟器 |
| `android/arm64` | arm64 手机或 arm64 模拟器 |
| `ios/simulator-arm64` | Mac 上的 iOS 模拟器 |
| `ios/arm64` | 真 iPhone |

手机目标上，`httpclient` 和连接 Postgres 的库会在链接前被拒绝。AES、BCrypt、Pbkdf2 会报错。背单词的手机库没有用到它们。细节在 [EXPORT_C_ABI.md](EXPORT_C_ABI.md) 和 [MOBILE_APPS.md](MOBILE_APPS.md)。

### 登录和 Notes 在另一套目录

背单词没有账号。需要登录和一张列表时，用 `apps/shared/mobilecore.klx`，壳在 `apps/android/` 和 `apps/ios/`（路径里没有 `vocab`），后台是 KylixAdmin，默认端口 **8090**。字段、token 和编译命令在 [MOBILE_APPS.md](MOBILE_APPS.md)。这次重写没有重跑那一套后台编译。

### 术语

| 词 | 白话 |
|---|---|
| 终端 | 打命令、看文字输出的窗口 |
| 仓库根 | 克隆下来的 `kylix` 文件夹，里面有 `README_CN.md` |
| PATH | 系统寻找 `go`、`clang` 的文件夹名单 |
| WSL | 装在 Windows 里的 Linux。第 3–6 章在这里做 |
| doctor | `./kylix mobile doctor`。看工具、NDK、iOS、用不到的库。最后一行算数 |
| 后端 | `go` 或 `llvm`，两条编译路线。`check` 要求它们的结果一致 |
| 端口 8091 | 单词服务。Notes 后台是 8090 |
| 127.0.0.1 | 这台电脑自己。Android 模拟器里它是模拟器自己 |
| 10.0.2.2 | Android 模拟器访问电脑的地址 |
| NDK | 编 Android `.so` 的工具包。用 `ANDROID_NDK_HOME` 指出位置 |
| ABI | CPU 种类。`amd64` 给 x86_64 模拟器，`arm64` 给 arm64 |
| `.so` / `.a` | Android 的库 / 只能在 Mac 上链接的 iOS 库 |
| JNI | Android 上 Kotlin 调用 C 函数的通道 |
| xcodegen | 在 Mac 上根据 `project.yml` 生成 Xcode 工程 |
| Gradle | Android Studio 用来同步工程的构建系统。本仓库没有 `gradlew` |

界面为什么不写进 Kylix，见 [MULTIPLATFORM.md](MULTIPLATFORM.md)。这组命令的验收记录在 [MOBILE_DX.md](MOBILE_DX.md)。
