# 移动端开发体验（v0.16.0）

> 2026-10-11 随 v0.16.0 发布。CLI 版本 `0.16.0`（`kylix version` 打印 `Kylix 0.16.0`）。发版准备 PR 不打 git 标签，标签由维护者在合并后打。命令说明见 [MOBILE_COMMANDS.md](MOBILE_COMMANDS.md)，零基础教程见 [MOBILE_TUTORIAL_CN.md](MOBILE_TUTORIAL_CN.md)。下面的摩擦记录是写计划时的旧路径，留着当验收依据。

写零基础教程时，把 `apps/vocab` 从「核心检查」走到「编出 Android `.so`」在一台 Linux 上实际跑过。Android Studio、模拟器点击、Xcode、Windows 没有跑。下面每一条都对应这次碰到的具体摩擦，不写没有见过的愿望清单。

## 现在有多绕

一个只想在模拟器里看到 apple / 苹果的人，要自己凑齐这些东西：

| 已经必须会用的 | 原因 |
|---|---|
| Git、Go、LLVM（`llc` 和 `clang`）、`curl`、`cc` | 第 3–5 章。`host_check.sh` 还要 `cc -ldl` |
| bash | `apps/vocab/*.sh` 的第一行是 `#!/usr/bin/env bash`，并用 `mktemp`、`cmp` |
| Android NDK r26d，且要自己设 `ANDROID_NDK_HOME` | `kylix doctor` 不看 NDK |
| Android Studio、JDK 17、SDK 34、CMake、Gradle | 工程没有 `gradlew`（已确认 `apps/vocab/android/gradlew` 不存在） |
| 仅 Mac：完整 Xcode、`xcodegen` | 仓库不附带 XcodeGen，doctor 也不检查 |

命令不是一条。至少要：`go build` 出 `kylix`、`host_check.sh`、`e2e.sh`（跑完会把服务杀掉）、再手动编译并留下 `vocab_server`、`build_core.sh amd64`、`check_artifact.sh`，然后才打开 **正确的** 目录 `apps/vocab/android`。旁边还有一套 Notes，目录是 `apps/android`，端口是 8090。打开错文件夹不会有一句「你打开的是登录示例」。

## 这次实际碰到的摩擦

1. **doctor 的叉号会让人停在第 3 章。** 在编写环境执行 `./kylix doctor`，libcurl 和 libgc 打了 `✗ ... needed by db / httpclient / crypto`，最后一行仍是 `✓ doctor: all required tools found.` libgc 那一行的括号是复制来的，它并不是 db/httpclient。背单词不需要这两样。doctor 也不提 NDK、Android SDK、Xcode。

2. **NDK 找不到时只有一句英文。** 把 `ANDROID_NDK_HOME` 指到不存在的目录后，原样得到：`Error: android cross-link needs Android NDK (not found): install the NDK and point ANDROID_NDK_HOME at it (it must contain toolchains/llvm/prebuilt/<host>/bin/clang)`。它不打印试过哪些路径，也不打印当前系统要的是 `linux-x86_64`、`darwin-x86_64` 还是 `windows-x86_64`（`pkg/llvmgen/ndk.go` 的 `hostPrebuiltTag`）。WSL 里误用 Windows 版 NDK 时，报错和「完全没装」长得一样。

3. **脚本是 bash，Windows 的 `kylix.exe` 帮不上这段教程。** 发布物里有 `kylix-windows-amd64.exe`，但 `host_check.sh` 使用 `cc -ldl` 和 `dlopen`。cmd / PowerShell 不能执行这些 `.sh`。教程只能把 WSL2 写成前置条件，而且 WSL 里的服务和 Windows 上的模拟器之间的 `10.0.2.2` 没有在这里实测。

4. **`e2e.sh` 成功之后，模拟器无服务可连。** 脚本最后会停掉 8091。给人看的服务要再执行另一长串 `kylix build` + `go build`。两条路径都对，但没有一条命令叫做「留下服务」。

5. **文件列表靠人记。** 手机库必须是 `stringutil.klx`、`vocab.klx`、`vocab_lib.klx`，不能带 `server.klx`。输出路径必须是 `jniLibs/x86_64` 或 `arm64-v8a`，而参数词是 `amd64` / `arm64`。`build_core.sh` 把这件事做对了，但它是每个应用一份 bash，不是编译器的子命令。

6. **ABI 装错时，失败发生在手机里。** `check_artifact.sh` 能看出 ELF 是 x86-64 还是 aarch64，但应用侧的 `UnsatisfiedLinkError` 不会说「你编的是 arm64，模拟器是 x86_64」。

7. **编译中间文件掉在源码旁边。** 在 `apps/vocab` 里构建后，目录中出现 `vocab_lib.ll`、`vocab_lib.o` 等。它们被 `*.ll` / `*.o` 忽略，不进 Git，但第一次做的人会以为把源码弄脏了。

8. **`*.txt` 忽略规则误伤 `CMakeLists.txt`。** `.gitignore` 有 `*.txt`。新增的 `apps/vocab/android/app/src/main/cpp/CMakeLists.txt` 必须 `git add -f` 才能进库。这是给改示例的人的坑，普通读者看不见，但下一个人加 JNI 文件时会丢文件。

9. **教程步骤没有进 CI。** `.github/workflows/ci.yml` 会跑 `apps/admin/e2e.sh` 和 Notes 那份 Android `.so`，不会跑 `apps/vocab/host_check.sh` 或 `apps/vocab/e2e.sh`。教程里的成功行（`[vocab] host check PASS`、`[vocab] e2e PASS`）一旦改了脚本，文档不会被 CI 打红。

10. **iOS 失败信息只有两行，且 XcodeGen 无检查。** Linux 上 `build_core.sh simulator` 的原文是「必须在 macOS 上链接」。它不说明下一步在 Mac 上要先有 `xcodegen`。doctor 同样不检查。

11. **HTTP 正文是 JSON，头却是 text/plain。** `server.klx` 使用 `BootText`，因为 LLVM 的 `BootJSON` 会把正文丢掉。curl 能对上正文；用「只认 application/json」的客户端的人会对不上。这不是教程能单独改掉的，它是编译器行为。

## v0.16.0 做了什么

下列各项已随 2026-10-11 的 v0.16.0 发布。目标：在教程支持的系统上，小白用少量命令走到「核心已证明、服务开着、`.so` 放对文件夹」，并且报错里写明缺的是哪一个工具、到哪条路径去找。不在这一版把界面框架做进 Kylix，也不把 `httpclient` 链进手机。

实现语言优先用 Go 子命令，这样 Windows 的 `kylix.exe` 和 Linux / macOS 是同一条命令。现有 `.sh` 可以留着，改成调用子命令，避免两套逻辑。

### 1. `kylix mobile doctor`

- **对着哪条摩擦：** 1、2、3、10。
- **行为：** 分三列打印：背单词必需（`go`、`llc` 或可用的 `clang`、`curl` 命令、bash 或「本机是 Windows 因此请用 WSL」）、Android（`ANDROID_NDK_HOME` 或自动搜到的 NDK、预编译目录名、`clang` 是否存在）、仅 macOS 的 iOS（`xcode-select -p`、`xcodegen`）。libcurl / libgc 放进「这个例子不用」而不是 `✗` 加错误的括号。
- **验收：** 在现在这台缺 libcurl 的 Linux 上，退出码 0，并明确写 NDK 已找到、libcurl 不影响背单词。把 `ANDROID_NDK_HOME` 指错时，退出码非 0，输出里含有试过的路径和 `linux-x86_64`。windows-latest 的 CI 上运行时，明确写「bash 脚本请用 WSL」或改由子命令完成检查，而不是让用户猜。

### 2. `kylix mobile serve` 与 `kylix mobile check`

- **对着哪条摩擦：** 4、9。
- **行为：** `check` 做现在的 Go/LLVM 逐字节比较和 curl 三步（浏览、标记 id 1、复习不含 apple），成功行保持稳定。`serve` 在前台监听 8091，打印 `vocab server on :8091`，直到 Ctrl+C。不要在 `check` 结束时还占用端口。
- **验收：** CI 增加一步，运行 `kylix mobile check`（或等价地跑 `apps/vocab/host_check.sh` 与 `apps/vocab/e2e.sh`），期望最后一行分别为 `[vocab] host check PASS` 和 `[vocab] e2e PASS (go)`。LLVM 那次 `e2e` 至少在 Linux job 上跑一次。

### 3. `kylix mobile build android`

- **对着哪条摩擦：** 5、6。
- **行为：** 参数只接受 `amd64` 或 `arm64`。自己带上正确的三个 `.klx`，输出到 `apps/vocab/android/app/src/main/jniLibs/<对应目录>/libkylixvocab.so`，然后做现在 `check_artifact.sh` 的 ELF 和符号检查。结束时用一句话写明：x86_64 模拟器用 amd64 这份，arm64 模拟器用 arm64 这份。
- **验收：** 在带 NDK r26d 的 Linux 上，两种 ABI 的 `file` 输出分别含 `x86-64` 和 `ARM aarch64`，动态符号含 `vocab_exports.list` 里的每一个名字。缺 NDK 时走第 1 条的错误格式。

### 4. 中间文件不要留在源码目录

- **对着哪条摩擦：** 7。
- **行为：** `llc` 的 `.ll` / `.o` 默认写到临时目录或 `kylixvocab_gen/` 一类已经忽略的目录。
- **验收：** 在 `apps/vocab` 下执行 mobile build / check 之后，`git status` 不出现新的 `.ll` 或 `.o`。

### 5. `.gitignore` 不要忽略 `CMakeLists.txt`

- **对着哪条摩擦：** 8。
- **行为：** 把 `*.txt` 收窄，或对 `CMakeLists.txt` 取反。已经进库的文件不受影响，但新文件不必 `git add -f`。
- **验收：** 在一个临时目录放 `CMakeLists.txt`，`git check-ignore` 不把它当成忽略文件。现有确实该忽略的 `*.txt` 产物仍被忽略（若有测试或文档依赖，先列出来再改）。

### 6. iOS 仍只在 Mac 上链接，但失败要能看懂

- **对着哪条摩擦：** 10。
- **行为：** 非 macOS 上 `kylix mobile build ios` 退出码非 0，说明需要 macOS、完整 Xcode，以及 `xcodegen`。不尝试在 Linux 上造出一个不能用的 `.a`。
- **验收：** Linux CI 上这条命令失败且输出含 `macOS` 和 `Xcode`。macos runner 上 `simulator` 与 `device` 仍做现有的符号和 SDK 链接冒烟（可以继续针对 `apps/ios` 的 Notes 库，vocab 的 `.a` 至少有一条）。

### 明确不放进 v0.16.0

- 不把 OkHttp / URLSession 换成 Kylix 的 `httpclient`。
- 不新增图形界面框架。
- 不承诺在 CI 里启动 Android 模拟器或 iOS Simulator 并点击「认识」。那仍是教程里的手工步骤，直到有单独的设备实验室。
- 不在 v0.16.0 修改 `BootJSON` 丢正文的行为，除非单独立项；教程继续说明正文是 JSON、头是 `text/plain`。若要修，验收是 Go 与 LLVM 的同一请求正文逐字节一致，且现有 admin E2E 不回归。
