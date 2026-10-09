# Kylix 多端规划（H5 / Android / iOS）

> 创建: 2026-09-12（用户已确认）
> 架构取向: **共享 Kylix 核心 + 各端原生壳**（不造跨平台 UI 框架）
> 节奏: 排在 KylixAdmin（v0.10–v0.12，见 [ADMIN_PLATFORM.md](ADMIN_PLATFORM.md)）之后，v0.13.0–v0.15.0，1.0.0 gate 不变
> 技术基线核对（2026-09-12）: `tripleFor` 现有 5 个桌面 triple；`export` token 在 lexer 已有、生成器未实现；`pkg/wasi` 为 Go 侧 stub 骨架
> 2026-10-09：`[Export]`、移动端 triple、示例壳、CI 产物门，以及 `wasm32-unknown-wasi` 纯逻辑子集已落地。见 [EXPORT_C_ABI.md](EXPORT_C_ABI.md)、[MOBILE_APPS.md](MOBILE_APPS.md)、[WASI.md](WASI.md)。stdlib 的 android/ios 独立分支仍开放。

---

## 一、总体架构：一核多壳

```
                    ┌─ H5：浏览器/微信内置浏览器（PWA）
                    ├─ Android：Kotlin 壳 + libkylix.so（JNI 桥）
共享核心（Kylix）───┤
 业务 unit + stdlib ├─ iOS：Swift 壳 + libkylix.a（C ABI 桥）
 JSON API + JWT     └─ wasm32-wasi：纯逻辑（无 DOM，见 WASI.md）
```

**原则：业务逻辑写一次（Kylix unit），UI 每端用该平台的正道。** 不造 UI 框架——H5 用 HTML/CSS，Android 用 Kotlin/Jetpack，iOS 用 SwiftUI；Kylix 输出**共享核心库**（数据模型、校验、业务规则、API 客户端、加解密、本地缓存），通过 C ABI + JSON 与壳交互。这是 gomobile / Kotlin Multiplatform 的同款成熟架构，并规避 iOS App Store 对纯 WebView 壳的 4.2 审核风险。

与 KylixAdmin 的关系：**KylixAdmin 后端即多端共用的 API 服务器**（JSON + JWT 已有）；`apps/shared/` 的业务 unit 同时被 admin/H5/Android/iOS 编译。

## 二、编译器侧工作清单（多端的真实成本）

1. **C ABI export 机制** ✅ v0.14.0：`[Export]` / `[Export('c_symbol')]`。Go 后端发 `//export` 与 `import "C"`；LLVM 后端发未改名全局符号，并在有导出或 `--shared` 时注入 `kylix_free`（malloc → `free`，`--gc=boehm` → `GC_free`）。见 [EXPORT_C_ABI.md](EXPORT_C_ABI.md)。
2. **triple 矩阵扩展** ✅ v0.14.0：`tripleFor`（`pkg/llvmgen/compile.go`）已有 `android/arm64`（`aarch64-linux-android30`）、`android/amd64`（`x86_64-linux-android30`，别名 `android/x86_64`）、`ios/arm64`（`arm64-apple-ios16.0.0`）、`ios/simulator-arm64`（`arm64-apple-ios16.0.0-simulator`）。
3. **工具链探测** ✅ v0.14.0：`FindAndroidNdk()`（`pkg/llvmgen/ndk.go`，`ANDROID_NDK_HOME` / `ANDROID_NDK_ROOT` / SDK 路径）；iOS 链接走 macOS 上的 `xcrun` clang，非 macOS 主机直接报错。
4. **stdlib 可移植层** → **v0.15**（v0.14 未做）：`stdlib_datetime.go` / `stdlib_sysutil.go` / `stdlib_net.go` 仍只区分 `windows` 与其余平台，android/ios 走非 Windows 路径。`compile.go` 在 android 上不链 `-lcrypto`/`-lpq`/`-lsqlite3`/`-lcurl`，ios 不链 `-lcurl`。独立的 android/ios 分支，以及 net/crypto/db 在这些目标上的第二批适配，留到 v0.15。
5. **CI 产物形态门禁** ✅ v0.15：`.github/workflows/ci.yml` 的 `mobile-android`（ubuntu-latest）用 NDK r26d 跑 `apps/android/build_core.sh` 的 arm64 与 amd64，`check_artifact.sh` 要求 ELF shared object，动态符号表含 `apps/shared/mobile_exports.list`（`mc_*` 与 `kylix_free`）。`mobile-ios`（macos-15）跑 `apps/ios/build_core.sh simulator` 与 `device`：`nm` 符号表，再用 `xcrun` 把 `.a` 链进 arm64 Mach-O（模拟器平台 `IOSSIMULATOR`，设备 SDK 平台 `IOS`）。真机登录不在 CI 里：不签名、不装机、不启动模拟器。手工步骤见 [MOBILE_APPS.md](MOBILE_APPS.md)。

## 三、各端技术路线

### H5（两条路线，先易后难）

**路线 A：响应式 PWA（v0.13，零编译器改动）**
- KylixAdmin 增强移动端体验：同 URL + 响应式 CSS/JS 增强（非独立 h5/ 页面组——双倍维护成本不值得），`manifest.json` + service worker（BootStatic 服务）→ 可安装主屏
- 认证：session-first（PWA 走 cookie）；JWT refresh 由 v0.15 原生壳消费（`POST /api/refresh`，见 [MOBILE_APPS.md](MOBILE_APPS.md)）；登录限流（应用层查 login_logs，非中间件——LLVM 端无中间件链）
- 交付：✅ v0.13.0（2026-09-24）+ `docs/H5_GUIDE.md`

**路线 B：Kylix → wasm32 纯逻辑（v0.15，编译器能力）** ✅
- LLVM triple `wasm32-unknown-wasi`（`--backend=llvm --target wasi/wasm32`，别名 `wasm` / `wasm32` / `wasi`）。`internal/wasiapi.Preview1` 是 12 个 `wasi_snapshot_preview1` 导入的单一来源；Go `GOOS=wasip1` 用 `//go:wasmimport`，LLVM 用 import attribute。宿主非 wasip1 仍是本机替身。
- 边界模型：**DOM 不进 wasm**。`--wasm` 仍是 Go `GOOS=js`。这条目标不链 libcrypto / sqlite / curl / libgc。
- 适配点：8MiB bump 堆，`@free` 为空操作；wasm 无 setjmp——`@setjmp` 返回 0，`longjmp` / 未捕获异常 `proc_exit(70)`，不是错误码改写。LLVM `uses wasi` 只有 Stdout/Stderr/Getenv/时钟/WasiExit；文件与参数在 `pkg/wasi` 的 wasip1 实现里，路径相对预打开 fd 3。详见 [WASI.md](WASI.md)。

### Android（v0.14 编译器能力 + v0.15 示例应用）

| 步骤 | 内容 |
|---|---|
| 交叉编译 | `aarch64-linux-android` triple；`FindAndroidNdk` 探测；NDK clang 链接 `.so` |
| C ABI 导出 | 同一套 export 机制（一次实现多端受益） |
| 桥接 | JNI 薄壳（~200 行 Kotlin + 一个 JNI `.c`）：String↔jstring + JSON 进出；中期可选 `kylix gen jni` 从导出签名生成绑定 |
| stdlib 移植 | net ✓；datetime/sysutil/exc 平台分支；crypto（OpenSSL 静态 or 手写实现已有）；db（随包 `sqlite3.c` amalgamation）；httpclient（libcurl 静态 or 平台 API 适配层） |
| 交付 | `apps/android/`（Kotlin 示例：登录 + 列表，连 KylixAdmin API，核心逻辑跑 libkylix.so）+ Gradle 集成文档 |

### iOS（v0.14 + v0.15，依赖 macOS 构建）

| 步骤 | 内容 |
|---|---|
| 交叉编译 | `aarch64-apple-ios` triple（llc 原生支持）→ `.o` → `ar` 产 `libkylix.a`；链接走 xcrun clang |
| C ABI 导出 | 同 Android |
| 桥接 | Swift Package 薄壳：`module.modulemap` 暴露 C 头 + Swift 包装层（String↔UnsafePointer、JSON 进出） |
| stdlib 移植 | iOS **系统自带 libsqlite3.tbd**（零打包成本）；net/crypto 同 Android 策略 |
| 交付 | `apps/ios/`（SwiftUI 示例：登录 + 列表）+ 真机/模拟器运行文档（签名需用户 Apple ID） |

## 四、版本路线（衔接已定规划）

| 版本 | 内容 | 交付标志 |
|---|---|---|
| v0.9.0–v0.12.0 | 不变（1.0.0-rc 打磨 + KylixAdmin P1–P5） | admin 平台完成 |
| **v0.13.0** | H5 路线 A：PWA 移动页面组 + manifest/SW + refresh token + H5_GUIDE | 手机浏览器可安装使用 admin 移动版 |
| **v0.14.0** | 编译器多端能力：export C ABI（双端）+ android/ios triple + NDK/Xcode 探测。stdlib 可移植层与 CI 产物门未纳入本版（见第二节第 4、5 条） | `[Export]` + `--shared` + 四个移动端 triple；指南 `EXPORT_C_ABI.md` |
| **v0.15.0** | 示例应用与 CI 产物门 ✅；wasm32-unknown-wasi + `pkg/wasi` preview1 子集 ✅（[WASI.md](WASI.md)）；stdlib android/ios 平台分支仍开放 | 双端真机/模拟器登录仍是手工步骤 |
| **1.0.0** | gate 不变；多端能力作为平台特性宣传 | — |

## 五、风险与诚实评估

| 风险 | 等级 | 对策 |
|---|---|---|
| iOS 纯 WebView 壳被 App Store 4.2 拒 | — | 已规避：原生壳 + 共享核心架构 |
| OpenSSL 移动端链接（体积+编译） | 🟡 | crypto 手写实现已有（SHA 全家桶）；AES 优先平台 API（Keychain/CommonCrypto）适配层 |
| JNI/Swift 桥样板代码维护 | 🟡 | 导出签名统一 JSON-in/JSON-out，桥层极薄；中期 `kylix gen jni` |
| UI 每端各写一遍 | ⚠️ 架构决定 | 明确不做跨平台 UI 框架，共享的是逻辑层 |
| wasm 异常/DOM 边界 | 🟡 | 纯逻辑子集已落地，不承诺 DOM。setjmp 不捕获（`proc_exit(70)`）；堆不回收。见 [WASI.md](WASI.md) |
| CI 三套工具链复杂度 | 🟡 | 产物形态验证为主（不跑模拟器），真机验收文档化。Android `.so`、iOS `.a` 与 wasm32 `wasi-wasm32` 门禁已进 `ci.yml` |

## 六、验收标准

- [x] 同一份业务 unit（`apps/shared/mobilecore.klx`：校验 + JSON API 协议）被 admin 编译（H5 即该二进制）并由 Android/iOS 以 C ABI 链接。宿主 Go/LLVM 输出逐字一致（`apps/shared/host_check.sh`）。HTTP 传输不在这份 unit 里——见 [MOBILE_APPS.md](MOBILE_APPS.md)
- [x] Android 产物门：CI 交叉链接 `libkylixlogic.so`（arm64 与 amd64），`file` 为 ELF shared object，动态符号含 `mc_*` / `kylix_free`
- [x] iOS 产物门：CI（macos-15）检查 `libkylixcore.a` 符号表，并把归档链进模拟器与 iphoneos Mach-O。入库环境不是 macOS，这条由 CI 跑
- [x] wasm32：`examples/wasi-logic/check.sh` 用 LLVM `wasm32-unknown-wasi` 链出模块，wasmtime stdout 为 `sum=42` / `hello wasi` / `42` / `1.5`，导入模块是 `wasi_snapshot_preview1`。Go `GOOS=wasip1` 的 `pkg/wasi` 在 `wasmtime --dir` 下读写预打开目录。详见 [WASI.md](WASI.md)
- [ ] 壳里的完整登录（Android 模拟器、iOS 模拟器、签名后的真机）仍是手工步骤，见 [MOBILE_APPS.md](MOBILE_APPS.md)
- [ ] H5：Lighthouse PWA 可安装性通过；弱网下降级可用
- [ ] 全量回归持续绿：16 包 + 双 sweep + bootstrap sweep + IR 不动点
