# 用 Kylix 做 Android 和 iOS（小白教程）

> 面向第一次把 Kylix 接到手机上的读者。会用命令行即可。不要求会写 Pascal，也不要求先读完编译器内部文档。
>
> 主练习是一个能在电脑上先跑通的背单词小应用：六个单词，浏览、标记认识、再复习还没掌握的。仓库里另有一套登录 + Notes 壳，用来看账号和 token，放在[附录](#13-附录登录和-notes)。
>
> 接口和 CI 清单仍是 [多端示例说明](MOBILE_APPS.md)。为什么这样拆见 [多端规划](MULTIPLATFORM.md)。导出语法见 [C ABI 指南](EXPORT_C_ABI.md)。
>
> CI 只检查编出来的 `.so` / `.a`，**不启动模拟器**。背单词的 HTTP 流程可以在本机用 curl 走完（第 5 章）。第 7、9 章是照着壳的默认地址和按钮写的手工步骤。

---

## 目录

1. [你最后会看到什么](#1-你最后会看到什么)
2. [一张图：逻辑写一次，界面各端自己画](#2-一张图逻辑写一次界面各端自己画)
3. [准备电脑](#3-准备电脑)
4. [不碰手机，先证明单词核心是好的](#4-不碰手机先证明单词核心是好的)
5. [把单词服务跑起来，用 curl 走完三步](#5-把单词服务跑起来用-curl-走完三步)
6. [编出 Android 用的 .so](#6-编出-android-用的-so)
7. [在 Android 模拟器里背单词](#7-在-android-模拟器里背单词)
8. [编出 iOS 用的 .a](#8-编出-ios-用的-a)
9. [在 iOS 模拟器里背单词](#9-在-ios-模拟器里背单词)
10. [深一层：Export 和谁来释放内存](#10-深一层export-和谁来释放内存)
11. [深一层：手机上的标准库边界](#11-深一层手机上的标准库边界)
12. [排障](#12-排障)
13. [附录：登录和 Notes](#13-附录登录和-notes)
14. [接下来读什么](#14-接下来读什么)

---

## 1. 你最后会看到什么

背单词示例在 `apps/vocab/`。它不是应用商店里的完整 App，只够把这条链路走通：

1. 电脑上跑一个很小的单词服务（端口 **8091**）。
2. 服务给出六张卡片。你标记「认识」之后，复习列表里那张就不见了。
3. Android 或 iOS 壳画这张卡片。壳把「已经认识哪些」交给 Kylix 编出来的库，库给出路径和 JSON，壳再用系统自带的 HTTP 发给单词服务。

内置的六个词：

| id | 英文 | 中文 |
|---|---|---|
| 1 | apple | 苹果 |
| 2 | book | 书 |
| 3 | water | 水 |
| 4 | friend | 朋友 |
| 5 | time | 时间 |
| 6 | happy | 高兴 |

「已经认识」是一串 id，例如 `1,3`。单词服务**不保存**这串字，每次请求由壳带回去。所以不需要登录，也没有数据库。

要做账号、口令和 token，用附录里的 KylixAdmin Notes，不要把登录硬塞进这个单词核心。

---

## 2. 一张图：逻辑写一次，界面各端自己画

```text
                    ┌─ 电脑：apps/vocab/server.klx（BootRun，端口 8091）
  同一份 Kylix ─────┤
  apps/vocab/       ├─ Android：Kotlin 画卡片，OkHttp 发 HTTP
  vocab.klx         │     加载 libkylixvocab.so
                    └─ iOS：SwiftUI 画卡片，URLSession 发 HTTP
                          链接 libkylixvocab.a
```

| 层 | 目录 | 它负责 |
|---|---|---|
| 共享逻辑 | `apps/vocab/vocab.klx` | 六个词、拼 JSON、标记认识、筛出还没掌握的 |
| 导出包装 | `apps/vocab/vocab_lib.klx` | `[Export]` 变成 C 函数。只编进手机库 |
| 单词服务 | `apps/vocab/server.klx` | 桌面上的 HTTP。**不**编进手机库 |
| Android 壳 | `apps/vocab/android/` | 卡片、四个按钮、OkHttp |
| iOS 壳 | `apps/vocab/ios/` | 卡片、四个按钮、URLSession |

HTTP **不**写进 `vocab.klx`。android / ios 这两个编译目标会拒绝 `httpclient`，也不会去链 libcurl。TLS 用手机系统已经有的那一套。

壳上的四个按钮：

| 按钮 | 壳做的事 |
|---|---|
| 全部单词 | `GET /api/words?known=1,3` |
| 复习未掌握 | `GET /api/review?known=1,3` |
| 认识 | `POST /api/mark`，body 是 `{"known":"1","id":"2"}` |
| 下一张 | 只在已经拿到的卡片里换一张，不再发请求 |

路径和 POST body 都由库返回。壳不自己拼这些字符串。

登录 + Notes 是另一份核心 `apps/shared/mobilecore.klx`，壳在 `apps/android/` 和 `apps/ios/`，后台是 KylixAdmin（默认端口 8090）。两套可以同时开。见第 13 章。

---

## 3. 准备电脑

在仓库根目录工作。下面把这个目录写成「仓库根」。

| 你要做的事 | 需要 | 怎么确认 |
|---|---|---|
| 编译 Kylix 自己 | Go | `go version` 有输出 |
| 第 4、5 章（核心和 curl） | `llc` 和 `clang` | 装好 LLVM 后，`kylix doctor` 里能看到它们 |
| Android 的 `.so` | Android NDK | 设 `ANDROID_NDK_HOME`（或 `ANDROID_NDK_ROOT`）。CI 用的是 **NDK r26d** |
| 打开 Android 工程 | Android Studio，JDK 17 | 工程是 AGP 8.5.2、Kotlin 1.9.24、minSdk 24。仓库里**没有** `gradlew` |
| iOS 的 `.a` 和模拟器 | macOS + Xcode | 不是 Mac 时，`apps/vocab/ios/build_core.sh` 会直接退出 |
| 生成 Xcode 工程 | 本机有 `xcodegen` 命令 | 输入文件是 `apps/vocab/ios/project.yml` |

先编出编译器（产物放在仓库根的 `kylix`，脚本默认就找这个名字）：

```bash
cd /path/to/kylix
go build -o kylix ./cmd/kylix/
./kylix version
```

应打印 `Kylix 0.15.0`。

`FindAndroidNdk` 认环境变量，也认 `ANDROID_HOME/ndk/` 或 `ANDROID_SDK_ROOT/ndk/` 下面的版本目录。NDK 根目录里要有 `toolchains/llvm/prebuilt/<主机>/bin/clang`。找不到时，链接会停在这句：

```text
android cross-link needs Android NDK (not found): install the NDK and point ANDROID_NDK_HOME at it
```

---

## 4. 不碰手机，先证明单词核心是好的

这一步不需要 NDK，也不需要 Mac。它证明两件事：同一份 `vocab.klx` 用 Go 后端和 LLVM 后端打出来的字一样；宿主上能 `dlopen` 那些 C 函数，并走完「浏览 → 把 1 标成认识 → 复习列表里没有 apple」。

```bash
bash apps/vocab/host_check.sh
```

脚本若发现仓库根没有可执行的 `kylix`，会自己 `go build -o kylix ./cmd/kylix/`。成功时最后一行是：

```text
[vocab] host check PASS
```

中间还有 `[vocab] parity PASS` 和 `c abi PASS`。这**不是**模拟器。

---

## 5. 把单词服务跑起来，用 curl 走完三步

`e2e.sh` 自己编译服务、在 8091 监听、发请求、把 body 和第 4 章那份核心输出逐字节对比，然后关掉服务。

```bash
bash apps/vocab/e2e.sh
```

成功的最后一行：

```text
[vocab] e2e PASS (go)
```

它做的事就是这三步（你也可以对着一个已经开着的服务自己 curl）：

```bash
curl -sS http://127.0.0.1:8091/api/words
curl -sS -X POST http://127.0.0.1:8091/api/mark \
  -H 'Content-Type: application/json' \
  --data '{"known":"","id":"1"}'
curl -sS 'http://127.0.0.1:8091/api/review?known=1'
```

第一次的 body 里有 `"left":6` 和 `apple`。标记之后 `"known":"1"`、`"left":5`。复习那次的 body **没有** `apple`，仍有 `book`。id 不是 1 到 6 时，服务返回 **400**，body 是 `{"ok":false,"error":"Unknown word"}`。

再把同一套 curl 打到 LLVM 编出来的服务上（更慢，和 Go 应对上同一份 JSON）：

```bash
VOCAB_BACKEND=llvm bash apps/vocab/e2e.sh
```

成功的最后一行是 `[vocab] e2e PASS (llvm)`。8091 上如果已经有进程，脚本会停下来让你先关掉它。

给模拟器留一个一直开着的服务时，在仓库根：

```bash
mkdir -p kylixvocab_gen
cd apps/vocab
../../kylix build --backend=go -o ../../kylixvocab_gen/server.go \
  ../../stdlib/stringutil.klx vocab.klx server.klx
cd ../..
go build -o vocab_server ./kylixvocab_gen
./vocab_server
```

成功时先看到：

```text
vocab server on :8091
```

接着是 Boot 的启动行。这个终端先别关。`kylixvocab_gen/` 和 `vocab_server` 是本地产物，不要提交。

响应是 `BootText`，`Content-Type` 实际是 `text/plain`。客户端只解析 body，不看这个头。不要改成 `BootJSON`：LLVM 那条调用会把 body 丢掉。

---

## 6. 编出 Android 用的 `.so`

在**另一个**终端，回到仓库根。先看你的模拟器是哪种 CPU：

| 模拟器 / 手机 | 命令 | 产物 |
|---|---|---|
| x86_64 模拟器（电脑上最常见） | `bash apps/vocab/android/build_core.sh amd64` | `apps/vocab/android/app/src/main/jniLibs/x86_64/libkylixvocab.so` |
| arm64 模拟器或真机 | `bash apps/vocab/android/build_core.sh arm64` | `apps/vocab/android/app/src/main/jniLibs/arm64-v8a/libkylixvocab.so` |

`amd64` 和 `x86_64` 是同一个参数。两种 ABI 可以各编一次，两个目录互不覆盖。

```bash
export ANDROID_NDK_HOME=/path/to/android-ndk-r26d
bash apps/vocab/android/build_core.sh amd64
```

脚本实际执行的是（在 `apps/vocab/` 里）：

```bash
kylix build --backend=llvm --target android/amd64 --shared \
  -o <仓库>/apps/vocab/android/app/src/main/jniLibs/x86_64/libkylixvocab.so \
  ../../stdlib/stringutil.klx vocab.klx vocab_lib.klx
```

注意文件列表里**没有** `server.klx`。手机库不含 HTTP 服务。`arm64` 时 `--target` 是 `android/arm64`。成功的最后一行是：

```text
[vocab-android] wrote .../libkylixvocab.so
```

检查形态（不启动模拟器，也不 `dlopen`，因为宿主机的动态链接器打不开 Android 的 `.so`）。设了 `ANDROID_NDK_HOME` 时，还会用 NDK 的 clang 把 JNI 桥链上这份库，缺符号会在这里失败：

```bash
bash apps/vocab/android/check_artifact.sh \
  apps/vocab/android/app/src/main/jniLibs/x86_64/libkylixvocab.so amd64
```

`file` 的输出里要有 `ELF 64-bit LSB shared object, x86-64`。arm64 则是 `ARM aarch64`。动态符号表要含 `apps/vocab/vocab_exports.list` 里的每个名字（`vc_*` 和 `kylix_free`）。成功的最后一行是 `android artifact OK (amd64)`。

---

## 7. 在 Android 模拟器里背单词

1. 第 5 章的 `./vocab_server` 仍在跑，端口 8091。
2. 第 6 章的 `.so` 已经在和模拟器 CPU 对应的 `jniLibs` 目录里。
3. 用 Android Studio 打开目录 `apps/vocab/android/`（不是打开某一个 `.kt` 文件）。等 Gradle 同步完。仓库不带 `gradlew`，用 Android Studio 自带的 Gradle。
4. 建一个模拟器。x86_64 模拟器要先跑过 `build_core.sh amd64`；arm64 模拟器要先跑过 `build_core.sh arm64`。
5. 点 Run，装上 `dev.kylix.vocab`。

界面上服务器一栏的默认字是 `http://10.0.2.2:8091`。`10.0.2.2` 是官方模拟器看宿主机的地址，不要改成 `127.0.0.1`（那是模拟器自己）。真机改成电脑的局域网地址，手机和电脑在同一网络。清单里这个示例允许明文 HTTP（`usesCleartextTraffic`），不要拿到公网上去用。

点 **全部单词**。第一张应是 `apple` / `苹果`，进度是「全部 · 还剩 6 / 6」。

点 **认识**。进度变成「还剩 5 / 6」，这张下面出现「已认识」。点 **下一张** 只换卡片，不访问网络。

点 **复习未掌握**。apple 不再出现，第一张变成 book。杀掉 App 再打开：已认识的 id 在普通 SharedPreferences 里，复习列表仍不含 apple。这不是密码，所以没有用 Keystore。

---

## 8. 编出 iOS 用的 `.a`

**只能在 Mac 上做。** 在 Linux 或 Windows 上执行会退出码 1，并打印：

```text
iOS archives must be linked on macOS with Xcode command line tools.
On this host, stop here. The Swift sources are in apps/vocab/ios/.
```

编译器自己的报错是：`ios cross-link requires a macOS host with Xcode command line tools`。

在 Mac 上、仓库根：

```bash
go build -o kylix ./cmd/kylix/
bash apps/vocab/ios/build_core.sh simulator
```

Apple silicon 模拟器用 `simulator`（`--target ios/simulator-arm64`）。真机用 `bash apps/vocab/ios/build_core.sh device`（`--target ios/arm64`）。两次写的是**同一个**文件：

```text
apps/vocab/ios/Sources/CKylixVocab/lib/libkylixvocab.a
```

这个 `.a` 在 `.gitignore` 里，要在本机现编。先编模拟器再编真机，会把模拟器那份盖掉。要跑模拟器时，最后一次必须是 `simulator`。链接的源文件同样是 `stringutil.klx`、`vocab.klx`、`vocab_lib.klx`，没有 `server.klx`。

成功的最后一行：

```text
[vocab-ios] wrote .../libkylixvocab.a
```

Mac 上可以再做同款检查（不安装到手机）：

```bash
bash apps/vocab/ios/check_artifact.sh \
  apps/vocab/ios/Sources/CKylixVocab/lib/libkylixvocab.a simulator
```

`nm` 里要有 `vocab_exports.list` 的符号（Mach-O 上带前导 `_`）。然后用模拟器 SDK 把一个很小的 C `main` 链成 arm64 Mach-O，`vtool` 的平台是 `IOSSIMULATOR`。`device` 参数则是 `IOS`，那次链接不签名、不装到手机。

---

## 9. 在 iOS 模拟器里背单词

1. 单词服务仍在这台 Mac 的 8091 端口（第 5 章的 `./vocab_server`）。
2. 刚刚跑的是 `build_core.sh simulator`，不是 `device`。
3. 生成并打开工程（需要本机已有 `xcodegen`）：

```bash
cd apps/vocab/ios
xcodegen generate
open KylixVocab.xcodeproj
```

`project.yml` 里的 scheme 叫 **KylixVocab**，bundle id 是 `dev.kylix.vocab`，最低系统 iOS 16，签名样式是 Automatic。选一台 Apple silicon 模拟器，Run。

服务器默认是 `http://127.0.0.1:8091`。模拟器和 Mac 共用这台机器的网络，所以是 `127.0.0.1`，不是 Android 的 `10.0.2.2`。按钮和 Android 相同：全部单词、认识、下一张、复习未掌握。看到 apple 消失，就和 Android 同一套含义。`Info.plist` 打开了本地网络和任意 HTTP 加载，只给这个示例。

已认识的 id 在 UserDefaults（`vocab.known`）。杀掉进程再打开，复习列表仍不含刚刚标记的词。

真机（脚本不做安装）：

1. `bash apps/vocab/ios/build_core.sh device`（会覆盖模拟器的 `.a`）。
2. 再 `xcodegen generate`，用本机 Apple ID 签名，选已连接的 iPhone，Run。
3. 服务器改成这台 Mac 的局域网地址，端口仍是 8091。

---

## 10. 深一层：Export 和谁来释放内存

壳不能直接调用 Pascal 函数名。`vocab_lib.klx` 用 `[Export('符号名')]` 导出 C 符号。头文件是 `apps/vocab/ios/Sources/CKylixVocab/include/kylixvocab.h`，Android 的 JNI 桥读的是同一批名字。

| C 函数 | 作用 |
|---|---|
| `vc_words_query` / `vc_review_query` | 返回 `/api/words?known=1,3` 这种路径。id 会先排成升序、去重 |
| `vc_mark_path` | 返回 `/api/mark` |
| `vc_mark_request` | 拼出 POST body |
| `vc_parse` | 看 HTTP 状态和 body。200 且带 `"ok":true` 就原样留下，否则收成错误 JSON |
| `vc_browse` / `vc_review` / `vc_apply_mark` | 同一份牌组 JSON。服务进程直接调 Kylix 函数；宿主检查通过 C 符号再调一次，用来对上 |
| `kylix_free` | 释放上面那些函数返回的字符串 |

返回的字符串在 Kylix 里经过 `s + ''`，是新 malloc 的。壳复制完必须 `kylix_free`。模块里的常量字符串不能 free。`Integer` 在 C 侧是 `int64_t`。

路径函数在 Kylix 里写成带一个用不到的参数（`VcWordsPath(0)`），因为宿主 Go 后端对「跨单元、零参数、又用在参数位置」的调用会丢掉括号。C 导出本身是零参数的，只在 LLVM 编出来的库里用。

自己写一个新导出时，语法和编译开关在 [EXPORT_C_ABI.md](EXPORT_C_ABI.md)。编进手机库的命令仍是第 6、8 章的 `build_core.sh`，不要把 `vocab_lib.klx` 或 `server.klx` 加进别的程序「顺便」里：Go 后端看到 `[Export]` 会发 `//export` 和 cgo。

---

## 11. 深一层：手机上的标准库边界

背单词的手机库只用了 `stringutil` 和 `vocab`。没有在手机库里打开数据库或发 HTTP。`server.klx` 用的 `boot` 是桌面进程，不在 `build_core.sh` 的文件列表里。

如果你以后把别的 Kylix 程序编到 android / ios，这些边界已经在编译器里：

| 在 android / ios 上 | 实际行为 |
|---|---|
| `Sha256` / `Md5` / `HmacSha256` | 走可移植实现，不链 OpenSSL |
| `AesEncrypt`、`BCrypt*`、`Pbkdf2*` | 编译期报错。没有 OpenSSL，也没有接到 Keystore / CommonCrypto |
| `httpclient`、libpq | 链接前拒绝，不链 `-lcurl` / `-lpq` |
| SQLite | iOS 链系统 `libsqlite3`。Android 要先有 amalgamation：`scripts/fetch_sqlite_amalgamation.sh` 或 `KYLIX_SQLITE_SRC`，源码不在 git 里 |
| 时间 | `localtime_r`，不走 Windows 的 `localtime_s` |
| 临时目录 | Android：`TMPDIR`，否则 `/data/local/tmp`。iOS：`TMPDIR`，否则 `confstr`，不用 `/tmp` |
| `SO_REUSEADDR` | Android 用 Linux 常量；iOS 用 Darwin 常量 |

桌面（linux / macOS / Windows）仍走原来的 OpenSSL 和 libcurl。不要把桌面那条链接方式抄到手机上。

想只验证标准库、不打开壳：

```bash
bash examples/mobile-stdlib/check.sh android
# Mac 上：
bash examples/mobile-stdlib/check.sh ios
```

---

## 12. 排障

| 现象 | 先看 |
|---|---|
| `port 8091 is already in use` | 已经有单词服务。停掉它，或直接用那个进程 |
| `PARITY DIFF` | Go 和 LLVM 打出的 JSON 不一致。先看 diff 里是不是转义或中文被换成了空格 |
| `android cross-link needs Android NDK` | `ANDROID_NDK_HOME` 没指到含 `toolchains/llvm/prebuilt/.../clang` 的 NDK 根 |
| Android 安装后一打开就崩在加载库 | `.so` 的 ABI 和模拟器不一致。x86_64 模拟器要 `amd64` 那份 |
| Android 模拟器连不上，curl 却可以 | 服务器地址应是 `http://10.0.2.2:8091`，不是 `127.0.0.1` |
| iOS 脚本马上退出 | 当前不是 macOS，或没装 Xcode 命令行工具 |
| 模拟器里的 `.a` 是设备版 | 最后一次 `build_core.sh` 若是 `device`，会盖掉 `simulator` 的产物 |
| 复习列表仍有刚刚认识的词 | 壳没把新的 `known` 存下来，或请求打到了另一个端口 |
| 想在手机库里 `uses httpclient` | 会被拒绝。HTTP 留在 OkHttp / URLSession |
| 把 `server.klx` 加进 `build_core.sh` | 不要。手机库只链 `vocab.klx` 和 `vocab_lib.klx` |

---

## 13. 附录：登录和 Notes

背单词没有账号。需要登录、access / refresh token、以及一张从数据库读出来的列表时，用仓库里的另一套示例。业务单元是 `apps/shared/mobilecore.klx`，壳是 `apps/android/` 和 `apps/ios/`，后台是 KylixAdmin。字段和 CI 的全文在 [MOBILE_APPS.md](MOBILE_APPS.md)。

先证明那份核心（同样不启动模拟器）：

```bash
bash apps/shared/host_check.sh
```

最后一行应是 `[core] host check PASS`。

**必须在 `apps/admin/` 里编译后台。** `main.klx` 有 `[Embed('views', 'static')]`，这两个目录相对「你执行 kylix 的当前目录」。在仓库根执行会报 `error[KLX213]: [Embed] cannot read views`。

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

成功时端口默认 **8090**（和背单词的 8091 不冲突）：

```text
[kyadmin] dialect=sqlite db=.../.kylixadmin/admin.db port=8090
🚀 KylixBoot started on http://localhost:8090
```

不设 `KYADMIN_PASSWORD` 时口令是 `Admin@123`，启动会警告。浏览器打开 http://localhost:8090 ，用 `admin` / `Admin@123` 登录，在 **Notes**（`/admin/notes`）新建一条。壳里的 JSON 把码点小于 32 的控制字符换成空格。

Android 库：

```bash
export ANDROID_NDK_HOME=/path/to/android-ndk-r26d
bash apps/android/build_core.sh amd64
bash apps/android/check_artifact.sh \
  apps/android/app/src/main/jniLibs/x86_64/libkylixlogic.so amd64
```

用 Android Studio 打开 `apps/android/`。模拟器里的服务器默认是 `http://10.0.2.2:8090`，用户名 `admin`，密码 `Admin@123`，点 **Sign in**。空列表时的英文是 `No notes yet...`，回浏览器新建后再登录。会话在 EncryptedSharedPreferences 里；Keystore 建主密钥失败时只留在本进程内存，不会因此崩溃。

iOS 只能在 Mac 上：

```bash
bash apps/ios/build_core.sh simulator
cd apps/ios
xcodegen generate
open KylixAdmin.xcodeproj
```

模拟器的服务器默认是 `http://127.0.0.1:8090`。不是 Mac 时 `apps/ios/build_core.sh` 会直接退出。

登录成功的 body 里有两张 token：`token` 是 24 小时的 access，拿去 `GET /api/notes`；`refresh_token` 是 30 天的 refresh，只能拿去 `POST /api/refresh` 或 `POST /api/logout`。同一用户最多 8 张 refresh。壳在 access 到期前 60 秒，或列表返回 401 时，用 refresh 换一对新的，只试一次。这些规则在 `mobilecore.klx` 和 `controllers/api.klx`，不在背单词那份核心里。

---

## 14. 接下来读什么

| 想知道 | 读 |
|---|---|
| 背单词目录和命令的短表 | [apps/vocab/README.md](../apps/vocab/README.md) |
| Notes 每个 JSON 字段、CI 在查什么 | [MOBILE_APPS.md](MOBILE_APPS.md) |
| 为什么不把 UI 框架做进 Kylix | [MULTIPLATFORM.md](MULTIPLATFORM.md) |
| 自己加 `[Export]` | [EXPORT_C_ABI.md](EXPORT_C_ABI.md) |
| 后台怎么加一张新表 | [ADMIN_DEV_GUIDE_CN.md](ADMIN_DEV_GUIDE_CN.md) |
| 浏览器装到主屏 | [H5_GUIDE.md](H5_GUIDE.md) |
| 语言本身从零学 | [TUTORIAL_FOR_BEGINNERS_CN.md](TUTORIAL_FOR_BEGINNERS_CN.md) |
