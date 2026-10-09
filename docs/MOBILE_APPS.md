# 多端示例应用（v0.15.0）

> 2026-10-09。CLI 版本 `0.15.0`。
> 规划原文：[MULTIPLATFORM.md](MULTIPLATFORM.md) 第三、四、六节；C ABI：[EXPORT_C_ABI.md](EXPORT_C_ABI.md)。

同一份 Kylix 业务单元跑在 admin（H5 就是这个二进制）和 Android / iOS 壳上。壳只负责界面和 HTTP。

## 为什么 HTTP 不进 Kylix 核心

`pkg/llvmgen/compile.go` 在 android 与 ios 上拒绝 `httpclient` 和 libpq，不链 `-lcurl`、`-lpq`、`-lcrypto`。SHA-256/MD5 走可移植实现。sqlite：ios 链系统 `libsqlite3`，android 编译 amalgamation，不链宿主 `-lsqlite3`。把 libcurl 和 OpenSSL 静态链进移动端二进制，是规划第五节标黄的体积和工具链风险；OkHttp 与 URLSession 已经做了 TLS。

所以：

| 层 | 放什么 |
|---|---|
| `apps/shared/mobilecore.klx` | 校验、JSON 请求体、JSON 响应判定、`Bearer` 头、路径、access / refresh TTL |
| KylixAdmin `controllers/api.klx` | 真正查库、`DoLogin`、`JwtSign` |
| Android / iOS | 画登录页和列表，发 HTTP |

JWT refresh 已接上。登录同时发 24 小时 access token（`McAccessTTL` = 86400，`typ=access`）和 30 天 refresh token（`McRefreshTTL` = 2592000，`typ=refresh` 加 `jti`）。同一用户可以同时持有多张 refresh token（每张一个 `jti`）。壳把服务器地址和两个 token 写入安全存储（Android `EncryptedSharedPreferences`，iOS Keychain），冷启动时恢复，access 到期前 60 秒或列表返回 401 时调 `POST /api/refresh`，只尝试一次；刷新失败或登出才清掉本地会话并回到登录页。

## 目录

```
apps/shared/mobilecore.klx       业务单元（admin 与两端都编译它）
apps/shared/mobilecore_lib.klx   [Export] 包装，只给移动端库
apps/shared/parity.klx           宿主 Go/LLVM 逐字对比
apps/shared/host_check.sh        parity + dlopen
apps/android/                    Kotlin + JNI + OkHttp
apps/ios/                        SwiftUI + Swift Package + URLSession
```

`mobilecore_lib.klx` 不进 admin 的文件列表。Go 后端看到 `[Export]` 会发 `//export` 和 cgo，admin 不需要那样。

导出符号（`kylixcore.h`）：`mc_validate_login`、`mc_login_request`、`mc_parse_login`、`mc_refresh_request`、`mc_parse_refresh`、`mc_list_request`、`mc_parse_list`、`mc_auth_header`、`mc_login_path`、`mc_refresh_path`、`mc_logout_path`、`mc_notes_path`、`kylix_free`。返回的字符串都经过 `s + ''`，是 malloc 出来的，调用方复制后必须 `kylix_free`。模块常量不能 free。`Integer` 是 `int64_t`。

## JSON API

KylixAdmin 增加 JSON 路由，HTML 登录不变。

`POST /api/login`

```json
{"username":"admin","password":"Admin@123"}
```

成功（200），登录和刷新是同一种 body：

```json
{"ok":true,"token":"...","refresh_token":"...","username":"admin","display_name":"...","expires_in":86400,"refresh_expires_in":2592000}
```

失败：400 校验、401 口令错误、429 限流。body 形如 `{"ok":false,"error":"..."}`。经 `mc_parse_login` 后登录失败的 `relogin` 是 false。

`POST /api/refresh`，body `{"refresh_token":"..."}`。成功时只轮换提交上来的那张：删掉它的 `jti`，再插入新行。另一台设备的 refresh token 仍然有效。被轮换的旧 token 再提交得到 401。access token 不能拿来刷新，refresh token 不能当 Bearer 去拉 Notes。刷新失败经 `mc_parse_refresh` 后 `relogin` 是 true。

`POST /api/logout`，body 与刷新相同（`mc_refresh_request`）。验签成功且 `typ=refresh` 时只删除这一张 `jti`。token 为空是 400；token 无效或已经轮换过仍返回 200 `{"ok":true}`，这样壳不会卡在登出。壳在请求之后清本地存储，网络失败也清。离线登出删不掉服务器上的那一行，要等这张 token 再被提交，或被下面的上限挤掉。

### 多设备模型

`api_refresh` 是 `(jti TEXT PRIMARY KEY, username TEXT NOT NULL, created_at INTEGER NOT NULL)`。一次登录或一次成功的刷新插入一行，不按用户名整表删。

上限是每个用户名 8 行（`ApiRefreshCap`）。发新 token 之前，先删掉该用户 `created_at` 早于 refresh TTL（30 天）的行，再按 `created_at, jti` 删最老的，直到不足 8 行。刷新会先删掉自己那一行再插入，所以人已经在上限上时，刷新不会挤掉另一台设备。

旧库如果第一列还是 `username`（第一版示例的主键），启动时 `DROP` 再按新结构建表。那几行 refresh 不能拆成多设备会话，这些设备需要重新登录。E2E 每次用新库，不经过这条迁移。

没有行锁。同一张 refresh token 被两个请求同时刷新时，两边都可能通过查找并各插入一个后继 `jti`。过期但从未再提交的行，要等到该用户下次登录或刷新才会被清掉；单独的 401 不删行。`jti` 仍是用户名、unix 秒和进程内计数，不是随机数（LLVM 端没有 `RandomToken`）。

`GET /api/notes`，头必须是 `Authorization: Bearer <token>`（LLVM 的 `req.Header` 大小写敏感）。

成功（200）：`{"ok":true,"items":[{"id","title","body","done":true|false}, ...]}`，按 `id` 排序。空表是 `"items":[]`。

401（无 token、坏 token、过期、拿 refresh token 当 access）的 body 经 `mc_parse_list` 变成 `relogin:true`。壳先刷新；刷新也失败才回到登录。403 是没有 `notes.read`，不刷新。

响应是 `BootText`，`Content-Type` 为 `text/plain`。LLVM 的 `BootJSON` 会丢掉 body，所以两端都发手写 JSON；客户端只解析 body。请求体用 `req.BodyText()`（Go `string(Body())`，LLVM 与 `req.Body` 同一次 load）。bootstrap 编译器的 `src/llvmgen.klx` 还不认 `BodyText`；admin CI 用宿主编译器。

`/api` 与 `/api/` 跳过 CSRF。`/apiv2` 不跳过。HTML 表单仍要 `_csrf`。宿主 Go 与宿主 LLVM 都有这条豁免。**bootstrap 烘焙的 `src/stdlib_ir.klx` 还没有**，要等下次重烘；admin 的 CI 用宿主编译器。

secret：环境变量 `KYADMIN_JWT_SECRET`。未设置时用 `kylix-admin-dev-secret`，进程启动会打一行警告。只适合本机。

默认账号 `admin` / `Admin@123`，端口 `KYADMIN_PORT`，缺省 8090。

## 本仓库能自动验证的

在仓库根目录（需要 Go、以及 LLVM 后端用的 `llc`/`clang`）：

```bash
bash apps/shared/host_check.sh
```

它做两件事：`parity.klx` 的 Go 后端与 LLVM 后端 stdout 逐字一致；再把 `mobilecore_lib.klx` 编成宿主 `.so`，用 `apps/shared/c_host_check.c` `dlopen` 并 `kylix_free`。这不是模拟器。

admin 双端（较重，会起服务）：

```bash
go build -o /tmp/kylix_bin ./cmd/kylix/
KYLIX=/tmp/kylix_bin bash apps/admin/e2e.sh
```

期望末行含 `28 scenarios`。S26、S27、S28 在限流场景 S25 之前，避免把登录预算打满。S28 检查两张 refresh token 同时有效、轮换其中一张不影响另一张、登出只作废被提交的那张。postgres 四形态沿用原来的 `KYADMIN_DSN` 开关。

## CI 产物门

`.github/workflows/ci.yml` 有两个 job，只检查库的形态，不登录、不启动模拟器。

| Job | Runner | 做什么 |
|---|---|---|
| `mobile-android` | `ubuntu-latest` | 安装 NDK r26d（`ANDROID_NDK_HOME`），`build_core.sh arm64` 与 `amd64`，再 `check_artifact.sh` |
| `mobile-ios` | `macos-15` | `brew install llvm` 提供 `llc`，`build_core.sh simulator` 与 `device`，再 `check_artifact.sh` |

Android 门：`file` 必须是 `ELF 64-bit LSB shared object`（arm64 为 ARM aarch64，amd64 为 x86-64）。`llvm-nm -D`（或 `nm -D`）的动态符号表必须含 `apps/shared/mobile_exports.list` 里的每个名字（`mc_*` 与 `kylix_free`）。不 `dlopen`：Android 的 linker 不是宿主机的。

iOS 门：归档必须是 `ar archive`，`nm` 里要有同一份符号（Mach-O 上带前导 `_`）。然后用对应 SDK 把一个调用 `mc_login_path` / `mc_notes_path` / `mc_logout_path` / `kylix_free` 的 C `main` 链成 arm64 Mach-O。`vtool -show-build` 对模拟器必须是 `platform IOSSIMULATOR`，对 device 必须是 `platform IOS`。device 这次链接不签名，也不安装到手机。`build_core.sh` 两次写的是同一个 `libkylixcore.a`，所以 CI 先查模拟器再覆盖成 device。

本机有 NDK 时可以复跑 Android 门：

```bash
go build -o kylix ./cmd/kylix/
export ANDROID_NDK_HOME=/opt/android-ndk-r26d
KYLIX=$PWD/kylix bash apps/android/build_core.sh arm64
bash apps/android/check_artifact.sh apps/android/app/src/main/jniLibs/arm64-v8a/libkylixlogic.so arm64
KYLIX=$PWD/kylix bash apps/android/build_core.sh amd64
bash apps/android/check_artifact.sh apps/android/app/src/main/jniLibs/x86_64/libkylixlogic.so amd64
```

iOS 门只能在 macOS + Xcode 上跑，命令与 CI 相同（`KYLIX=$PWD/kylix`）。

相关单测：

```bash
go test ./pkg/boot/ ./pkg/llvmgen/ -count=1 -timeout 180s
```

## Android（需要 NDK）

本环境没有 NDK，下面的链接步骤没有在这里跑过。

```bash
# 设备或 arm64 模拟器
export ANDROID_NDK_HOME="$HOME/Android/Sdk/ndk/<version>"
bash apps/android/build_core.sh arm64
# x86_64 模拟器
bash apps/android/build_core.sh amd64
```

产物：`apps/android/app/src/main/jniLibs/<abi>/libkylixlogic.so`。用 Android Studio 打开 `apps/android/`（AGP 8.5.2，minSdk 24，abi `arm64-v8a` 与 `x86_64`）。

没有 NDK 时可以只出目标文件（不链接）：

```bash
cd apps/shared
kylix build --backend=llvm --target android/arm64 \
  -o /tmp/mobilecore_android.o \
  ../../stdlib/stringutil.klx mobilecore.klx mobilecore_lib.klx
file /tmp/mobilecore_android.o
```

模拟器验收：

1. 先在宿主机跑 admin：`KYADMIN_PORT=8090`（见 [ADMIN_DEV_GUIDE_CN.md](ADMIN_DEV_GUIDE_CN.md)）。
2. 壳里的服务器默认是 `http://10.0.2.2:8090`（模拟器看宿主机的别名）。真机改成电脑的局域网地址。清单允许明文 HTTP。
3. 用户 `admin`，密码 `Admin@123`。登录后应看到 Notes；库被 E2E 清过就是空列表，文案会提示去网页后台建一条。
4. access token 过期或被换成坏值时，壳会用 refresh token 换一对新的，列表仍在。登出、刷新失败，或 refresh token 过期，才会回到登录页。
5. 杀掉进程再打开：应仍在登录后的列表（Keystore 可用时）。再开一台模拟器登录同一账号，两边的列表都在；其中一边刷新或登出，另一边仍能拉列表。Keystore 建主密钥失败时，会话只留在本进程内存里，冷启动要重新登录。

## iOS（需要 macOS + Xcode）

非 Darwin 上 `apps/ios/build_core.sh` 直接退出 1，并说明原因。链接必须在 Mac 上。

```bash
bash apps/ios/build_core.sh simulator   # Apple silicon 模拟器
# 或
bash apps/ios/build_core.sh device
cd apps/ios
xcodegen generate
open KylixAdmin.xcodeproj
```

静态库写到 `apps/ios/Sources/CKylixCore/lib/libkylixcore.a`（已 gitignore）。Swift Package 用 `-lkylixcore`。模拟器服务器默认 `http://127.0.0.1:8090`。`Info.plist` 打开了本地网络与任意加载，仅供这个示例。签名用本机 Apple ID。

没有 Xcode 时，在 Mac 上也可以只出 `.o`：把上面的 android 命令里的 `--target` 换成 `ios/simulator-arm64` 或 `ios/arm64`，输出名用 `.o`。

### 真机验收（CI 不做）

CI 的 iphoneos 链接只证明 `.a` 能链进设备 SDK 的 Mach-O。装到手机要本机签名：

1. 在 Mac 上 `bash apps/ios/build_core.sh device`（覆盖模拟器那份 `libkylixcore.a`）。
2. `cd apps/ios && xcodegen generate && open KylixAdmin.xcodeproj`。
3. 选一台已连接的 iPhone，用本机 Apple ID 签名，Run。
4. 服务器改成这台 Mac 的局域网地址（模拟器才是 `http://127.0.0.1:8090`）。用户 `admin`，密码 `Admin@123`。登录后应看到 Notes。
5. 杀掉进程再打开：Keychain 里的 refresh token 应恢复会话。另一台设备或模拟器上的同一账号在这一台刷新或登出之后仍然有效（每 `jti` 一行）。
6. 模拟器完整登录同样不在 CI 里：`build_core.sh simulator` 之后用 Xcode 的模拟器目标 Run，服务器用 `http://127.0.0.1:8090`。

## 已知边界

- 控制字符（码点 < 32）在 JSON 转义里变成空格。词法器没有可用的 `Chr`，示例的 note 正文按单行处理。
- 反斜杠和引号会转义成 `\\` 与 `\"`。这两字节是 `'\' + '\'` 和 `'\' + '"'` 拼出来的：源码字面量 `'\\'` / `'\"'` 在 Go 后端是两个字节，在 LLVM 后端被 `decodeKylixString` 收成一个字节。
- 公开函数至少有一个参数。宿主 Go 后端对跨单元零参调用、且用在参数位置时会丢掉括号。C 导出的 `mc_login_path` / `mc_refresh_path` / `mc_logout_path` / `mc_notes_path` 是零参的，只在 LLVM 库里。
- 每个用户名最多 8 条 refresh 记录。第 9 次登录会挤掉最老的一台设备。壳把 token 放进 EncryptedSharedPreferences（Android，`androidx.security:security-crypto` 1.1.0-alpha06）或 Keychain（iOS，service `dev.kylix.admin`）。Keystore 失败时 Android 退回内存，不崩溃。模拟器/真机上的冷启动本环境没有跑。
- 登录成功仍会 `Set-Cookie`。壳不保存这张 cookie，之后只送 Bearer。
- wasm32-unknown-wasi 见 [WASI.md](WASI.md)。stdlib 的 android/ios 分支已落地（`examples/mobile-stdlib/check.sh`）：哈希不链 OpenSSL，ios sqlite 用系统库，android sqlite 要先跑 `scripts/fetch_sqlite_amalgamation.sh`。AES、httpclient、libpq 在移动端会报错。CI 的 `.so` / `.a` 形态门之外会再链这份探针，不跑壳里的登录。
