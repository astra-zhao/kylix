# 用 Kylix 做 Android 和 iOS（小白教程）

> 面向第一次把 Kylix 接到手机上的读者。会用命令行即可。不要求会写 Pascal，也不要求先读完编译器内部文档。
>
> 仓库里已经有三份**参考**，但它们不是一门课：[多端示例说明](MOBILE_APPS.md)（接口、目录、验收清单）、[多端规划](MULTIPLATFORM.md)（为什么这样拆）、[C ABI 指南](EXPORT_C_ABI.md)（导出语法）。本文按「先跑通、再看懂」往下走，每一步用的都是仓库里现成的路径和命令。
>
> CI 只检查编出来的 `.so` / `.a`，**不启动模拟器、不登录**。第 6、8 章是照着壳的默认地址和界面文案写的手工步骤。

---

## 目录

1. [你最后会看到什么](#1-你最后会看到什么)
2. [一张图：逻辑写一次，界面各端自己画](#2-一张图逻辑写一次界面各端自己画)
3. [准备电脑](#3-准备电脑)
4. [不碰手机，先证明核心是好的](#4-不碰手机先证明核心是好的)
5. [把后台跑起来，并建一条 Note](#5-把后台跑起来并建一条-note)
6. [编出 Android 用的 .so](#6-编出-android-用的-so)
7. [在 Android 模拟器里登录](#7-在-android-模拟器里登录)
8. [编出 iOS 用的 .a](#8-编出-ios-用的-a)
9. [在 iOS 模拟器里登录](#9-在-ios-模拟器里登录)
10. [深一层：Export 和谁来释放内存](#10-深一层export-和谁来释放内存)
11. [深一层：登录之后的两张 token](#11-深一层登录之后的两张-token)
12. [深一层：手机上的标准库边界](#12-深一层手机上的标准库边界)
13. [排障](#13-排障)
14. [接下来读什么](#14-接下来读什么)

---

## 1. 你最后会看到什么

仓库自带一个很小的示例，不是应用商店里的完整 App：

1. 电脑上跑 **KylixAdmin**（后台，也是 API 服务器），浏览器能登录。
2. Android 模拟器或 iOS 模拟器里打开壳，填同一个账号，看到 Notes 列表。
3. 列表里的字不是壳自己编的。壳把用户名和密码交给 Kylix 编出来的库，库给出 JSON 和路径，壳再用系统自带的 HTTP（Android 是 OkHttp，iOS 是 URLSession）发给后台。

默认账号在种子数据里：用户名 `admin`。不设置 `KYADMIN_PASSWORD` 时口令是 `Admin@123`，启动时会打一行警告。这只适合本机。

空列表时界面上的英文是：

```text
No notes yet.
Create one in the KylixAdmin web console (Notes), then sign in again.
```

所以第 5 章要先在网页里建一条 Note，手机上才看得到内容。

---

## 2. 一张图：逻辑写一次，界面各端自己画

```text
                    ┌─ 浏览器：就是 KylixAdmin 这一个程序（PWA）
                    ├─ Android：Kotlin 画界面，OkHttp 发 HTTP
  同一份 Kylix ─────┤     加载 libkylixlogic.so
  apps/shared/      ├─ iOS：SwiftUI 画界面，URLSession 发 HTTP
  mobilecore.klx    │     链接 libkylixcore.a
                    └─ 后台：apps/admin 查数据库、签发 token
```

三层各自放什么：

| 层 | 目录 | 它负责 |
|---|---|---|
| 共享逻辑 | `apps/shared/mobilecore.klx` | 校验、拼 JSON、判断「要不要重新登录」、路径、token 有效期 |
| 导出包装 | `apps/shared/mobilecore_lib.klx` | 把上面的函数用 `[Export]` 变成 C 函数。只编进手机库，不编进 admin |
| 后台 | `apps/admin/` | 真正查库、核对口令、发 token。网页登录还在 |
| Android 壳 | `apps/android/` | 输入框、列表、OkHttp |
| iOS 壳 | `apps/ios/` | 输入框、列表、URLSession |

HTTP **不**写进 `mobilecore.klx`。android / ios 这两个编译目标会拒绝 `httpclient`，也不会去链 libcurl。TLS 用手机系统已经有的那一套。

H5 不是再编一次。手机浏览器打开的就是第 5 章跑起来的那个 KylixAdmin。安装到主屏的做法在 [H5 指南](H5_GUIDE.md)，和本文的原生壳是两条路。

---

## 3. 准备电脑

在仓库根目录工作。下面把这个目录写成「仓库根」。

| 你要做的事 | 需要 | 怎么确认 |
|---|---|---|
| 编译 Kylix 自己 | Go | `go version` 有输出 |
| 编手机库（LLVM 后端） | `llc` 和 `clang` | 装好 LLVM 后，`kylix doctor` 里能看到它们 |
| Android 的 `.so` | Android NDK | 设 `ANDROID_NDK_HOME`（或 `ANDROID_NDK_ROOT`）。CI 用的是 **NDK r26d** |
| 打开 Android 工程 | Android Studio，JDK 17 | 工程是 AGP 8.5.2、Kotlin 1.9.24、minSdk 24。仓库里**没有** `gradlew` |
| iOS 的 `.a` 和模拟器 | macOS + Xcode | 不是 Mac 时，`apps/ios/build_core.sh` 会直接退出 |
| 生成 Xcode 工程 | 本机有 `xcodegen` 命令 | 输入文件是 `apps/ios/project.yml`。仓库不负责安装 XcodeGen |

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

## 4. 不碰手机，先证明核心是好的

这一步不需要 NDK，也不需要 Mac。它只证明：同一份 `mobilecore.klx` 用 Go 后端和 LLVM 后端打出来的字一样，并且宿主上能 `dlopen` 那些 C 函数。

```bash
bash apps/shared/host_check.sh
```

脚本若发现仓库根没有可执行的 `kylix`，会自己 `go build -o kylix ./cmd/kylix/`。成功时最后一行是：

```text
[core] host check PASS
```

中间还有 `[core] parity PASS`。这**不是**模拟器登录。

---

## 5. 把后台跑起来，并建一条 Note

手机壳要连的就是这个进程。命令与 [KylixAdmin 小白指南](ADMIN_DEV_GUIDE_CN.md) 第 2 章相同，这里收成能直接抄的一份。

**必须在 `apps/admin/` 里编译。** `main.klx` 有 `[Embed('views', 'static')]`，这两个目录相对「你执行 kylix 的当前目录」。在仓库根执行会报 `error[KLX213]: [Embed] cannot read views`。

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

成功时有两行，端口默认 8090：

```text
[kyadmin] dialect=sqlite db=.../.kylixadmin/admin.db port=8090
🚀 KylixBoot started on http://localhost:8090
```

不设 `KYADMIN_PASSWORD` 时还会警告默认口令正在使用。浏览器打开 http://localhost:8090 ，用 `admin` / `Admin@123` 登录。

左侧菜单进入 **Notes**（地址是 `/admin/notes`），新建一条。标题和正文用普通单行文字即可。壳里的 JSON 把码点小于 32 的控制字符换成空格。

这个进程先别关。后面的模拟器都连它。

换端口：`KYADMIN_PORT=9000 ./kylixadmin`。换了的话，壳里的服务器地址也要改成同一个端口。

JWT 密钥是环境变量 `KYADMIN_JWT_SECRET`。不设时用 `kylix-admin-dev-secret`，只适合本机。手机和电脑必须用**同一次启动**的后台，否则 token 对不上。

---

## 6. 编出 Android 用的 `.so`

在**另一个**终端，回到仓库根。先看你的模拟器是哪种 CPU：

| 模拟器 / 手机 | 命令 | 产物 |
|---|---|---|
| x86_64 模拟器（电脑上最常见） | `bash apps/android/build_core.sh amd64` | `apps/android/app/src/main/jniLibs/x86_64/libkylixlogic.so` |
| arm64 模拟器或真机 | `bash apps/android/build_core.sh arm64` | `apps/android/app/src/main/jniLibs/arm64-v8a/libkylixlogic.so` |

`amd64` 和 `x86_64` 是同一个参数。两种 ABI 可以各编一次，两个目录互不覆盖。工程的 `abiFilters` 就是 `arm64-v8a` 和 `x86_64`。

```bash
export ANDROID_NDK_HOME=/path/to/android-ndk-r26d
bash apps/android/build_core.sh amd64
```

脚本实际执行的是（在 `apps/shared/` 里）：

```bash
kylix build --backend=llvm --target android/amd64 --shared \
  -o <仓库>/apps/android/app/src/main/jniLibs/x86_64/libkylixlogic.so \
  ../../stdlib/stringutil.klx mobilecore.klx mobilecore_lib.klx
```

`arm64` 时 `--target` 是 `android/arm64`。成功的最后一行是：

```text
[android] wrote .../libkylixlogic.so
```

检查形态（不启动模拟器，也不 `dlopen`，因为宿主机的动态链接器打不开 Android 的 `.so`）：

```bash
bash apps/android/check_artifact.sh \
  apps/android/app/src/main/jniLibs/x86_64/libkylixlogic.so amd64
```

`file` 的输出里要有 `ELF 64-bit LSB shared object, x86-64`。arm64 则是 `ARM aarch64`。动态符号表要含 `apps/shared/mobile_exports.list` 里的每个名字（`mc_*` 和 `kylix_free`）。

还没有 NDK、只想看目标文件时（这条不产生壳能加载的 `.so`）：

```bash
cd apps/shared
../kylix build --backend=llvm --target android/arm64 \
  -o /tmp/mobilecore_android.o \
  ../../stdlib/stringutil.klx mobilecore.klx mobilecore_lib.klx
file /tmp/mobilecore_android.o
```

---

## 7. 在 Android 模拟器里登录

1. 第 5 章的 `./kylixadmin` 仍在跑，端口 8090。
2. 第 6 章的 `.so` 已经在和模拟器 CPU 对应的 `jniLibs` 目录里。
3. 用 Android Studio 打开目录 `apps/android/`（不是打开某一个 `.kt` 文件）。等 Gradle 同步完。仓库不带 `gradlew`，用 Android Studio 自带的 Gradle。
4. 建一个模拟器。x86_64 模拟器要先跑过 `build_core.sh amd64`；arm64 模拟器要先跑过 `build_core.sh arm64`。
5. 点 Run，装上 `dev.kylix.admin`。

界面上服务器一栏的默认字是 `http://10.0.2.2:8090`。`10.0.2.2` 是官方模拟器看宿主机的地址，不要改成 `127.0.0.1`（那是模拟器自己）。真机改成电脑的局域网地址，手机和电脑在同一网络。清单里这个示例允许明文 HTTP（`usesCleartextTraffic`），不要拿到公网上去用。

用户名默认 `admin`，密码填 `Admin@123`，点 **Sign in**。

- 成功：登录框消失，出现 Notes。你在网页里建过的那条应在列表里。
- 库是空的：看到 `No notes yet...`。回浏览器在 Notes 里新建，再在壳里重新登录。
- 一直转圈或 network error：后台没开、端口不对，或模拟器地址不是 `10.0.2.2`。

杀掉 App 再打开：Keystore 可用时，会话从 EncryptedSharedPreferences 恢复，应直接在列表页。Keystore 建主密钥失败时，会话只留在本进程内存里，冷启动要重新登录，程序不会因此崩溃。

再开一台模拟器、用同一账号登录：两边都能看列表。一边点登出，另一边仍然有效。原因在第 11 章。

---

## 8. 编出 iOS 用的 `.a`

**只能在 Mac 上做。** 在 Linux 或 Windows 上执行会退出码 1，并打印：

```text
iOS archives must be linked on macOS with Xcode command line tools.
On this host, stop here. The Swift sources are in apps/ios/.
```

编译器自己的报错是：`ios cross-link requires a macOS host with Xcode command line tools`。

在 Mac 上、仓库根：

```bash
go build -o kylix ./cmd/kylix/
bash apps/ios/build_core.sh simulator
```

Apple silicon 模拟器用 `simulator`（`--target ios/simulator-arm64`）。真机用 `bash apps/ios/build_core.sh device`（`--target ios/arm64`）。两次写的是**同一个**文件：

```text
apps/ios/Sources/CKylixCore/lib/libkylixcore.a
```

这个 `.a` 在 `.gitignore` 里，要在本机现编。先编模拟器再编真机，会把模拟器那份盖掉。要跑模拟器时，最后一次必须是 `simulator`。

成功的最后一行：

```text
[ios] wrote .../libkylixcore.a
```

Mac 上可以再做 CI 同款检查（不安装到手机）：

```bash
bash apps/ios/check_artifact.sh \
  apps/ios/Sources/CKylixCore/lib/libkylixcore.a simulator
```

`nm` 里要有 `mobile_exports.list` 的符号（Mach-O 上带前导 `_`）。然后用模拟器 SDK 把一个很小的 C `main` 链成 arm64 Mach-O，`vtool` 的平台是 `IOSSIMULATOR`。`device` 参数则是 `IOS`，那次链接不签名、不装到手机。

没有 Xcode、只想看 `.o` 时，把第 6 章那条 android 命令的 `--target` 换成 `ios/simulator-arm64` 或 `ios/arm64`，输出名用 `.o`。

---

## 9. 在 iOS 模拟器里登录

1. 后台仍在这台 Mac 的 8090 端口。
2. 刚刚跑的是 `build_core.sh simulator`，不是 `device`。
3. 生成并打开工程（需要本机已有 `xcodegen`）：

```bash
cd apps/ios
xcodegen generate
open KylixAdmin.xcodeproj
```

`project.yml` 里的 scheme 叫 **KylixAdmin**，bundle id 是 `dev.kylix.admin`，最低系统 iOS 16，签名样式是 Automatic。选一台 Apple silicon 模拟器，Run。

服务器默认是 `http://127.0.0.1:8090`。模拟器和 Mac 共用这台机器的网络，所以是 `127.0.0.1`，不是 Android 的 `10.0.2.2`。用户名 `admin`，密码 `Admin@123`，点 **Sign in**。

看到 Notes，或空列表那句英文，就和 Android 同一套含义。`Info.plist` 打开了本地网络和任意 HTTP 加载，只给这个示例。

真机（CI 不做）：

1. `bash apps/ios/build_core.sh device`（会覆盖模拟器的 `.a`）。
2. 再 `xcodegen generate`，用本机 Apple ID 签名，选已连接的 iPhone，Run。
3. 服务器改成这台 Mac 的局域网地址。账号仍是 `admin` / `Admin@123`。
4. 杀掉进程再打开：Keychain（service 名 `dev.kylix.admin`）里的 refresh token 应恢复会话。

---

## 10. 深一层：Export 和谁来释放内存

壳不能直接调用 Pascal 函数名。`mobilecore_lib.klx` 用 `[Export('符号名')]` 导出 C 符号。头文件是 `apps/ios/Sources/CKylixCore/include/kylixcore.h`，Android 的 JNI 桥读的是同一批名字。

| C 函数 | 作用 |
|---|---|
| `mc_validate_login` | 登录前检查用户名、密码是否为空、是否太长 |
| `mc_login_request` | 拼出 `POST /api/login` 的 JSON |
| `mc_parse_login` | 看 HTTP 状态和 body，决定成功还是失败 |
| `mc_refresh_request` | 拼出刷新和登出共用的 JSON：`{"refresh_token":"..."}` |
| `mc_parse_refresh` | 刷新失败时 `relogin` 为 true |
| `mc_list_request` / `mc_auth_header` | 列表请求和 `Authorization: Bearer ...` |
| `mc_parse_list` | 401 变成 `relogin:true`；403 不刷新 |
| `mc_login_path` 等四个 path | 返回 `/api/login`、`/api/refresh`、`/api/logout`、`/api/notes` |
| `kylix_free` | 释放上面那些函数返回的字符串 |

返回的字符串在 Kylix 里经过 `s + ''`，是新 malloc 的。壳复制完必须 `kylix_free`。模块里的常量字符串不能 free。`Integer` 在 C 侧是 `int64_t`。

四个 path 函数在 Kylix 里写成带一个用不到的参数（`McLoginPath(0)`），因为宿主 Go 后端对「跨单元、零参数、又用在参数位置」的调用会丢掉括号。C 导出本身是零参数的，只在 LLVM 编出来的库里用。

自己写一个新导出时，语法和编译开关在 [EXPORT_C_ABI.md](EXPORT_C_ABI.md)。编进手机库的命令仍是第 6、8 章的 `build_core.sh`，不要把 `mobilecore_lib.klx` 加进 admin 的文件列表：Go 后端看到 `[Export]` 会发 `//export` 和 cgo，后台不需要。

---

## 11. 深一层：登录之后的两张 token

`POST /api/login` 的成功 body（登录和刷新是同一种）：

```json
{"ok":true,"token":"...","refresh_token":"...","username":"admin","display_name":"...","expires_in":86400,"refresh_expires_in":2592000}
```

- `token`：access，24 小时（`McAccessTTL` = 86400），类型 `typ=access`。拿它放进 `Authorization: Bearer` 去 `GET /api/notes`。
- `refresh_token`：30 天（`McRefreshTTL` = 2592000），类型 `typ=refresh`，带 `jti`。只能拿去 `POST /api/refresh` 或 `POST /api/logout`。不能当 Bearer。

壳在 access 到期前 60 秒，或列表返回 401 时，用 refresh token 换一对新的，只试一次。刷新失败或登出才清本地存储并回到登录页。

服务器表 `api_refresh` 是每个 `jti` 一行，不是每个用户一行。同一用户最多 **8** 行。第 9 次登录会挤掉最老的一台设备。刷新先删掉自己那一行再插入，所以人已经在上限上时，刷新不会挤掉另一台。

`POST /api/logout` 只删提交的那一张。空 token 是 400；无效或已经轮换过的 token 仍返回 200 `{"ok":true}`，避免壳卡在登出。壳发出请求后就清本地，网络失败也清。离线登出删不掉服务器那一行。

响应是 `BootText`，`Content-Type` 实际是 `text/plain`。客户端只解析 body，不看这个头。`/api` 与 `/api/` 跳过 CSRF；网页表单登录仍然要 CSRF。

---

## 12. 深一层：手机上的标准库边界

示例壳的核心只用了 `stringutil` 和 `mobilecore`，没有在手机库里打开数据库或发 HTTP。如果你以后把别的 Kylix 程序编到 android / ios，这些边界已经在编译器里：

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

这同样不登录。

---

## 13. 排障

| 现象 | 先看 |
|---|---|
| `[Embed] cannot read views` | 编译 admin 时当前目录不是 `apps/admin/` |
| `go: go.mod file not found` | 生成的 Go 文件放在了仓库外面。应放在仓库里的 `kylixadmin_gen/` |
| `android cross-link needs Android NDK` | `ANDROID_NDK_HOME` 没指到含 `toolchains/llvm/prebuilt/.../clang` 的 NDK 根 |
| Android 安装后一点登录就崩在加载库 | `.so` 的 ABI 和模拟器不一致。x86_64 模拟器要 `amd64` 那份 |
| Android 模拟器连不上，浏览器却可以 | 服务器地址应是 `http://10.0.2.2:8090`，不是 `127.0.0.1` |
| iOS 脚本马上退出 | 当前不是 macOS，或没装 Xcode 命令行工具 |
| 模拟器里的 `.a` 是设备版 | 最后一次 `build_core.sh` 若是 `device`，会盖掉 `simulator` 的产物 |
| 登录 401，网页却能进 | 壳和网页必须打同一进程。口令是启动时的 `KYADMIN_PASSWORD`，没设就是 `Admin@123` |
| 列表 401 然后被踢回登录 | access 坏了且 refresh 也失败。看后台是否重启过（密钥或库变了） |
| 空列表 | 正常。到网页 `/admin/notes` 新建后再登录 |
| 第 9 台设备登不上、最老的那台掉线 | 每个用户最多 8 条 refresh |
| 想在手机库里 `uses httpclient` | 会被拒绝。HTTP 留在 OkHttp / URLSession |

登录限流（20 次失败 / 15 分钟 / IP）在后台，壳连续试错口令也会撞上 429。

---

## 14. 接下来读什么

| 想知道 | 读 |
|---|---|
| 每个 JSON 字段、CI 在查什么、已知边界的全文 | [MOBILE_APPS.md](MOBILE_APPS.md) |
| 为什么不把 UI 框架做进 Kylix | [MULTIPLATFORM.md](MULTIPLATFORM.md) |
| 自己加 `[Export]` | [EXPORT_C_ABI.md](EXPORT_C_ABI.md) |
| 后台怎么加一张新表 | [ADMIN_DEV_GUIDE_CN.md](ADMIN_DEV_GUIDE_CN.md) |
| 浏览器装到主屏 | [H5_GUIDE.md](H5_GUIDE.md) |
| 语言本身从零学 | [TUTORIAL_FOR_BEGINNERS_CN.md](TUTORIAL_FOR_BEGINNERS_CN.md) |
