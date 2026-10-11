# 背单词示例

六个单词，三步：浏览、标记认识、复习还没掌握的。业务在 `vocab.klx`。手机壳只画卡片，并用系统 HTTP 去问 `server.klx`。单词服务不存进度，壳把已认识的 id（例如 `1,3`）每次带回去。

没有登录。要做账号和 token，用旁边的 KylixAdmin Notes 壳（`apps/android`、`apps/ios`）。逐步安装、每条命令的原文和排错在 [零基础教程](../../docs/MOBILE_TUTORIAL_CN.md)。命令表在 [MOBILE_COMMANDS.md](../../docs/MOBILE_COMMANDS.md)。

## 仓库根

```bash
go build -o kylix ./cmd/kylix/
./kylix mobile doctor
./kylix mobile check
./kylix mobile serve
./kylix mobile build android amd64
```

`check` 的最后一行是 `[vocab] e2e PASS (go)`。`serve` 停在前台，先打印 `vocab server on :8091`，端口 8091。旁边的 KylixAdmin 默认是 8090，两个可以同时开。LLVM 的同一套 curl：

```bash
./kylix mobile check --e2e-only --backend=llvm
```

最后一行是 `[vocab] e2e PASS (llvm)`。

arm64 模拟器或真机把 `amd64` 换成 `arm64`。成功时最后一行是 `android artifact OK (amd64)` 或 `android artifact OK (arm64)`，并写明 x86_64 模拟器用 amd64 那一份。设 `ANDROID_NDK_HOME` 指向 NDK r26d。

`apps/vocab/*.sh` 调用上面同一套子命令。`android/build_core.sh` 额外把 `x86_64` 当成 `amd64`。

用 Android Studio 打开 `apps/vocab/android/`（没有 `gradlew`）。模拟器里的服务器默认是 `http://10.0.2.2:8091`。先让 `serve` 留在前台。

## iOS

只能在 macOS + 完整 Xcode 上编库。非 Mac 上 `./kylix mobile build ios simulator` 直接退出，不写 `.a`。在 Mac 上：

```bash
./kylix mobile build ios simulator
cd apps/vocab/ios
xcodegen generate
open KylixVocab.xcodeproj
```

模拟器的服务器默认是 `http://127.0.0.1:8091`。`xcodegen` 用 `brew install xcodegen`，它生成工程，不参与链接 `.a`。
