# 背单词示例

六个单词，三步：浏览、标记认识、复习还没掌握的。业务在 `vocab.klx`。手机壳只画卡片，并用系统 HTTP 去问 `server.klx`。单词服务不存进度，壳把已认识的 id（例如 `1,3`）每次带回去。

没有登录。要做账号和 token，用旁边的 KylixAdmin Notes 壳（`apps/android`、`apps/ios`），说明在 [小白教程](../../docs/MOBILE_TUTORIAL_CN.md) 附录。

## 本机先跑通（不需要手机）

在仓库根：

```bash
go build -o kylix ./cmd/kylix/
bash apps/vocab/host_check.sh
bash apps/vocab/e2e.sh
```

`host_check.sh` 最后一行是 `[vocab] host check PASS`（Go 与 LLVM 输出逐字节相同，并且 `dlopen` 走完浏览 / 认识 / 复习）。

`e2e.sh` 在 8091 端口拉起单词服务，用 curl 对上同一份 JSON。最后一行是 `[vocab] e2e PASS (go)`。旁边的 KylixAdmin 默认是 8090，两个可以同时开。

LLVM 后端的同一套 curl：

```bash
VOCAB_BACKEND=llvm bash apps/vocab/e2e.sh
```

## Android 模拟器

```bash
export ANDROID_NDK_HOME=/path/to/android-ndk-r26d
bash apps/vocab/android/build_core.sh amd64
bash apps/vocab/android/check_artifact.sh \
  apps/vocab/android/app/src/main/jniLibs/x86_64/libkylixvocab.so amd64
```

arm64 模拟器或真机把 `amd64` 换成 `arm64`。用 Android Studio 打开 `apps/vocab/android/`（没有 `gradlew`）。先在仓库根把单词服务留在前台（8091）：

```bash
mkdir -p kylixvocab_gen
cd apps/vocab
../../kylix build --backend=go -o ../../kylixvocab_gen/server.go \
  ../../stdlib/stringutil.klx vocab.klx server.klx
cd ../..
go build -o vocab_server ./kylixvocab_gen
./vocab_server
```

成功时先看到 `vocab server on :8091`。模拟器里的服务器默认是 `http://10.0.2.2:8091`。不要写成 `127.0.0.1`。点「全部单词」看到 apple / 苹果，点「认识」，再点「复习未掌握」，apple 不再出现。`kylixvocab_gen/` 和 `vocab_server` 是本地产物，不要提交。

## iOS 模拟器

只能在 macOS + Xcode 上编库：

```bash
bash apps/vocab/ios/build_core.sh simulator
bash apps/vocab/ios/check_artifact.sh \
  apps/vocab/ios/Sources/CKylixVocab/lib/libkylixvocab.a simulator
cd apps/vocab/ios
xcodegen generate
open KylixVocab.xcodeproj
```

模拟器的服务器默认是 `http://127.0.0.1:8091`。单词服务要跑在这台 Mac 上。不是 Mac 时 `build_core.sh` 直接退出，不会写出一个不能用的库。
