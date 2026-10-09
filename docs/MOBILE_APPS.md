# 多端示例应用（v0.15.0 进行中）

> 2026-10-09。CLI 版本仍是 0.14.0，本指南描述的是未发版的第一项。
> 规划原文：[MULTIPLATFORM.md](MULTIPLATFORM.md) 第三、四、六节；C ABI：[EXPORT_C_ABI.md](EXPORT_C_ABI.md)。

同一份 Kylix 业务单元跑在 admin（H5 就是这个二进制）和 Android / iOS 壳上。壳只负责界面和 HTTP。

## 为什么 HTTP 不进 Kylix 核心

`pkg/llvmgen/compile.go` 在 `android` 上整段跳过 `-lcurl`、`-lcrypto`、`-lpq`、`-lsqlite3`，在 `ios` 上跳过 `-lcurl`。核心里若调用 `httpclient`，这两个目标链不出来。把 libcurl 和 OpenSSL 静态链进移动端二进制，是规划第五节标黄的体积和工具链风险；OkHttp 与 URLSession 已经做了 TLS。

所以：

| 层 | 放什么 |
|---|---|
| `apps/shared/mobilecore.klx` | 校验、JSON 请求体、JSON 响应判定、`Bearer` 头、路径、access / refresh TTL |
| KylixAdmin `controllers/api.klx` | 真正查库、`DoLogin`、`JwtSign` |
| Android / iOS | 画登录页和列表，发 HTTP |

JWT refresh 已接上。登录同时发 24 小时 access token（`McAccessTTL` = 86400，`typ=access`）和 30 天 refresh token（`McRefreshTTL` = 2592000，`typ=refresh` 加 `jti`）。壳把两个 token 放在内存里（进程结束即丢），access 到期前 60 秒或列表返回 401 时调 `POST /api/refresh`，只尝试一次；刷新失败才回到登录页。

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

导出符号（`kylixcore.h`）：`mc_validate_login`、`mc_login_request`、`mc_parse_login`、`mc_refresh_request`、`mc_parse_refresh`、`mc_list_request`、`mc_parse_list`、`mc_auth_header`、`mc_login_path`、`mc_refresh_path`、`mc_notes_path`、`kylix_free`。返回的字符串都经过 `s + ''`，是 malloc 出来的，调用方复制后必须 `kylix_free`。模块常量不能 free。`Integer` 是 `int64_t`。

## JSON API

KylixAdmin 增加两条路由，HTML 登录不变。

`POST /api/login`

```json
{"username":"admin","password":"Admin@123"}
```

成功（200），登录和刷新是同一种 body：

```json
{"ok":true,"token":"...","refresh_token":"...","username":"admin","display_name":"...","expires_in":86400,"refresh_expires_in":2592000}
```

失败：400 校验、401 口令错误、429 限流。body 形如 `{"ok":false,"error":"..."}`。经 `mc_parse_login` 后登录失败的 `relogin` 是 false。

`POST /api/refresh`，body `{"refresh_token":"..."}`。成功时轮换：`api_refresh` 里该用户只留新的 `jti`，旧 refresh token 再提交得到 401。access token 不能拿来刷新，refresh token 不能当 Bearer 去拉 Notes。刷新失败经 `mc_parse_refresh` 后 `relogin` 是 true。

每个用户名同时只有一条 refresh 记录。第二次登录会使上一次登录发出的 refresh token 失效。多设备各自持有 refresh token 不在这个示例里。

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

期望末行含 `26 scenarios`。S26 在限流场景 S25 之前，避免把登录预算打满。postgres 四形态沿用原来的 `KYADMIN_DSN` 开关。

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
4. access token 过期或被换成坏值时，壳会用 refresh token 换一对新的，列表仍在。把 refresh token 也清掉，或等它过期，才会回到登录页。

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

## 已知边界

- 控制字符（码点 < 32）在 JSON 转义里变成空格。词法器没有可用的 `Chr`，示例的 note 正文按单行处理。
- 反斜杠和引号会转义成 `\\` 与 `\"`。这两字节是 `'\' + '\'` 和 `'\' + '"'` 拼出来的：源码字面量 `'\\'` / `'\"'` 在 Go 后端是两个字节，在 LLVM 后端被 `decodeKylixString` 收成一个字节。
- 公开函数至少有一个参数。宿主 Go 后端对跨单元零参调用、且用在参数位置时会丢掉括号。C 导出的 `mc_login_path` / `mc_refresh_path` / `mc_notes_path` 是零参的，只在 LLVM 库里。
- 每个用户名同时只留一条 refresh 记录。第二次登录会使上一张 refresh token 失效。壳不把 token 写入磁盘。
- 登录成功仍会 `Set-Cookie`。壳不保存这张 cookie，之后只送 Bearer。
- wasm、CI 上的 `.so`/`.a` 形态门、stdlib 的 android/ios 平台分支，都不在这一项里。
