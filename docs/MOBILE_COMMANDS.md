# `kylix mobile` 命令

当前发布版本仍是 **v0.15.0**（`kylix version` 打印 `0.15.0`）。这些子命令已经在主分支上，供背单词示例 `apps/vocab` 使用。零基础一步步怎么敲、成功和失败的原文，见 [MOBILE_TUTORIAL_CN.md](MOBILE_TUTORIAL_CN.md)。`apps/vocab` 里的 `.sh` 调用这里的同一套实现。验收记录在 [MOBILE_DX.md](MOBILE_DX.md)。

在仓库根执行。Windows 上直接运行 `kylix.exe mobile`。`apps/vocab/*.sh` 仍是 bash，cmd / PowerShell 请改用 WSL，或改用下面这些子命令。

## `kylix mobile doctor`

分四段打印，不把 libcurl / libgc 当成背单词的失败项：

| 段 | 内容 |
|---|---|
| vocab required | `go`、`llc` 或 `clang`、终端命令 `curl`、`cc`（没有只提示，不因此失败）、Windows 上写明 bash 脚本要用 WSL |
| Android NDK | `ANDROID_NDK_HOME` 或自动搜到的 NDK、主机预编译目录名（`linux-x86_64` / `darwin-x86_64` / `windows-x86_64`）、里面的 `clang` |
| iOS (macOS only) | 非 Mac：说明需要 macOS、完整 Xcode、`xcodegen`。Mac 上再看 `xcode-select -p` 和 `xcodegen`（缺了只提示） |
| not used by the vocab example | libcurl 库和 libgc。不影响背单词 |

退出码：缺 `go` / LLVM / `curl` 命令，或 `ANDROID_NDK_HOME`（或在 HOME 为空时的 `ANDROID_NDK_ROOT`）指到了没有主机 clang 的目录时为非 0。只是没装 NDK、或在 Linux 上没有 Xcode，退出码仍是 0。HOME 一旦设置，就只搜这一处：即便 `ANDROID_NDK_ROOT` 指向另一份有效 NDK（GitHub 托管镜像会预置它），也不会改用那一份。

NDK 已找到、并且这台机器没有把 libcurl 当成失败时，输出类似：

```text
=== vocab required ===
  ✓ go: /usr/bin/go
  ✓ llc: /usr/bin/llc
  ✓ clang: /usr/bin/clang
  ✓ curl: /usr/bin/curl
  ✓ cc: /usr/bin/cc
  ✓ bash: /usr/bin/bash

=== Android NDK ===
  host prebuilt directory: linux-x86_64
  ✓ NDK found: /opt/android-ndk-r26d
  ✓ clang: /opt/android-ndk-r26d/toolchains/llvm/prebuilt/linux-x86_64/bin/clang

=== iOS (macOS only) ===
  − not this host. iOS linking needs macOS, a full Xcode install, and xcodegen.

=== not used by the vocab example ===
  − libcurl does not affect the vocab example (the library, not the curl command; db / httpclient would link it)
  − libgc does not affect the vocab example (only --gc=boehm links it)

✓ mobile doctor: vocab tools are ready.
```

`ANDROID_NDK_HOME` 指错时退出码非 0，输出含 `tried:`、你写的路径，以及 `linux-x86_64` 这种预编译目录名。若那个目录里只有 `windows-x86_64`，报错会把现有目录名也打出来（WSL 里误用 Windows 版 NDK 时看这一行）。

Windows 上必有一行：`bash scripts need WSL`。

## `kylix mobile check`

默认先做 Go / LLVM 逐字节比较和 C ABI（`cc -ldl`），再对 8091 做 curl：浏览、标记 id 1、复习结果里不再有 apple，并保留脚本里原来的其余对照。结束时停掉服务，不占用端口。

```bash
kylix mobile check
kylix mobile check --e2e-only --backend=llvm
kylix mobile check --backend=both
```

成功时能看到：

```text
[vocab] host check PASS
[vocab] e2e PASS (go)
```

`--backend=llvm` 的最后一行是 `[vocab] e2e PASS (llvm)`。`--host-only` 只做到第一行。`apps/vocab/host_check.sh` 和 `apps/vocab/e2e.sh` 分别调用这两段。

## `kylix mobile serve`

在前台编译并留下单词服务，直到 Ctrl+C。端口 8091。进程自己打印：

```text
vocab server on :8091
```

```bash
kylix mobile serve
kylix mobile serve --backend=llvm
```

端口已被占用时直接失败，不会再起一个。

## `kylix mobile build android amd64|arm64`

只接受 `amd64` 或 `arm64`。源文件固定为 `stdlib/stringutil.klx`、`apps/vocab/vocab.klx`、`apps/vocab/vocab_lib.klx`，不包含 `server.klx`。

| 参数 | 模拟器 | 输出 |
|---|---|---|
| amd64 | x86_64 模拟器 | `apps/vocab/android/app/src/main/jniLibs/x86_64/libkylixvocab.so` |
| arm64 | arm64 模拟器或真机 | `apps/vocab/android/app/src/main/jniLibs/arm64-v8a/libkylixvocab.so` |

结束前检查 ELF 和 `apps/vocab/vocab_exports.list` 里的每个动态符号，并在 NDK 可用时做 JNI 链接。最后几行类似：

```text
[vocab-android] wrote .../jniLibs/x86_64/libkylixvocab.so
x86_64 emulators use the amd64 build (jniLibs/x86_64); arm64 emulators use the arm64 build (jniLibs/arm64-v8a).
android artifact OK (amd64)
```

`file` 的那一行里，amd64 含 `x86-64`，arm64 含 `ARM aarch64`。缺 NDK 时的报错和 doctor 指错路径时同一格式。`.ll` / `.o` 写在临时目录，不留在 `apps/vocab` 源码旁边。

`apps/vocab/android/build_core.sh` 调用这条命令。脚本额外把 `x86_64` 当成 `amd64` 的别名；子命令本身不接受 `x86_64`。

已有的 `.so` 可以用：

```bash
kylix mobile artifact android path/to/libkylixvocab.so amd64
```

没设置 `ANDROID_NDK_HOME` 时，符号对上之后会打印 `ANDROID_NDK_HOME is unset; skipped the JNI link`，最后一行仍是 `android artifact OK`。

## `kylix mobile build ios simulator|device`

只在 macOS 上链接，并做符号表和 Xcode SDK 冒烟（`IOSSIMULATOR` / `IOS`）。成功时最后一行是 `ios artifact OK (simulator)` 或 `ios artifact OK (device)`，输出里有 `libkylixvocab.a`。

非 macOS 上退出码非 0，不写 `.a`。开头两行与教程里已有的原文一致，并写明需要完整 Xcode 和 `xcodegen`：

```text
iOS archives must be linked on macOS with Xcode command line tools.
On this host, stop here. The Swift sources are in apps/vocab/ios/.
ios cross-link requires a macOS host with Xcode command line tools.
Install a full Xcode (xcode-select -p) and xcodegen (brew install xcodegen) on that Mac. This host does not write a .a.
```

`xcodegen` 用来生成 Xcode 工程，链接 `.a` 本身不调用它。直接 `kylix build --target ios/...` 在非 Mac 上也会在写出文件之前失败。
