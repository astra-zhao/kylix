# Kylix 多端规划（H5 / Android / iOS）

> 创建: 2026-09-12（用户已确认）
> 架构取向: **共享 Kylix 核心 + 各端原生壳**（不造跨平台 UI 框架）
> 节奏: 排在 KylixAdmin（v0.10–v0.12，见 [ADMIN_PLATFORM.md](ADMIN_PLATFORM.md)）之后，v0.13.0–v0.15.0，1.0.0 gate 不变
> 技术基线核对（2026-09-12）: `tripleFor` 现有 5 个桌面 triple；`export` token 在 lexer 已有、生成器未实现；`pkg/wasi` 为 Go 侧 stub 骨架

---

## 一、总体架构：一核多壳

```
                    ┌─ H5：浏览器/微信内置浏览器（PWA）
                    ├─ Android：Kotlin 壳 + libkylix.so（JNI 桥）
共享核心（Kylix）───┤
 业务 unit + stdlib ├─ iOS：Swift 壳 + libkylix.a（C ABI 桥）
 JSON API + JWT     └─（远期）wasm：纯逻辑编译进浏览器
```

**原则：业务逻辑写一次（Kylix unit），UI 每端用该平台的正道。** 不造 UI 框架——H5 用 HTML/CSS，Android 用 Kotlin/Jetpack，iOS 用 SwiftUI；Kylix 输出**共享核心库**（数据模型、校验、业务规则、API 客户端、加解密、本地缓存），通过 C ABI + JSON 与壳交互。这是 gomobile / Kotlin Multiplatform 的同款成熟架构，并规避 iOS App Store 对纯 WebView 壳的 4.2 审核风险。

与 KylixAdmin 的关系：**KylixAdmin 后端即多端共用的 API 服务器**（JSON + JWT 已有）；`apps/shared/` 的业务 unit 同时被 admin/H5/Android/iOS 编译。

## 二、编译器侧工作清单（多端的真实成本，v0.14 主体）

1. **C ABI export 机制**（P0，一票否决项）：解析器已有 token；需 AST 节点 + 双端发射（Go 后端 `//export` 注释 + LLVM `.globl` 裸符号 + cdecl）+ 导出函数禁用内部调用约定 + `kylix_free(ptr)` 内存导出（arena reset 所有权模型——v0.8.0 P2 arena 基建正好用上）
2. **triple 矩阵扩展**：`tripleFor` 增 `aarch64-linux-android` / `x86_64-linux-android`（模拟器）/ `aarch64-apple-ios` + `--target` 贯通
3. **工具链探测**：`FindAndroidNdk()`（KYLIX_NDK_ROOT → 常见路径，链接走 NDK clang + `--sysroot`，复用 v0.7.1 mingw 三坑经验：quarantine/stack-probe/exec.Command 时序）；iOS 链接必须 xcrun clang（Apple 平台外部 ld 无法签符号）
4. **stdlib 可移植层**：平台分支从 4 个（darwin/linux/windows）扩到 6 个（+android/ios）——datetime（`localtime_r`）/sysutil/exc 先行；net（BSD socket 通用）/crypto/db 第二批
5. **CI**：Android 交叉验证 `.so` 产物形态（同 Windows COFF 验证模式）；iOS 仅 macOS runner 验 `.a` 符号表 + 模拟器链接；真机验收文档化

## 三、各端技术路线

### H5（两条路线，先易后难）

**路线 A：响应式 PWA（v0.13，零编译器改动）**
- KylixAdmin/业务后端增加移动端页面组：响应式断点复用 P4 设计系统（mobile-first 变体）+ `manifest.json` + service worker（BootStatic 直接服务）→ 可安装主屏、基础离线壳
- 认证：JWT + refresh token；登录限流（RateLimit 已有）
- 交付：`apps/admin/h5/` 模板页面组 + `docs/H5_GUIDE.md`

**路线 B：Kylix → wasm32 纯逻辑（v0.15，编译器能力）**
- LLVM 后端加 `wasm32-unknown-wasi` triple + `--target wasm`；`pkg/wasi`（现有 stub 骨架）真实现 wasi_snapshot_preview1 导入表（fd_write/clock/random）
- 边界模型：**DOM 不进 wasm**——wasm 导出纯函数（校验/计算/编解码），JS 薄胶水经 `WebAssembly.instantiate` 调用；与服务端渲染互补
- 适配点：字符串约定（线性内存 + 长度出参）；wasm 无 setjmp——exc 模块 wasm 分支改错误码返回

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
| **v0.14.0** | 编译器多端能力：export C ABI（双端）+ android/ios triple + NDK/Xcode 探测 + stdlib 可移植层第一批 | hello-core 在 Android .so / iOS .a 跑通 |
| **v0.15.0** | 示例应用 `apps/android`（Kotlin）+ `apps/ios`（SwiftUI）登录+列表 demo；CI 产物形态门禁；wasm32 triple + pkg/wasi 真实现（纯逻辑先行） | 双端真机/模拟器 demo |
| **1.0.0** | gate 不变；多端能力作为平台特性宣传 | — |

## 五、风险与诚实评估

| 风险 | 等级 | 对策 |
|---|---|---|
| iOS 纯 WebView 壳被 App Store 4.2 拒 | — | 已规避：原生壳 + 共享核心架构 |
| OpenSSL 移动端链接（体积+编译） | 🟡 | crypto 手写实现已有（SHA 全家桶）；AES 优先平台 API（Keychain/CommonCrypto）适配层 |
| JNI/Swift 桥样板代码维护 | 🟡 | 导出签名统一 JSON-in/JSON-out，桥层极薄；中期 `kylix gen jni` |
| UI 每端各写一遍 | ⚠️ 架构决定 | 明确不做跨平台 UI 框架，共享的是逻辑层 |
| wasm 异常/DOM 边界 | 🟡 | 纯逻辑先行 + 错误码返回，不承诺 DOM |
| CI 三套工具链复杂度 | 🟡 | 产物形态验证为主（不跑模拟器），真机验收文档化 |

## 六、验收标准

- [ ] 同一份业务 unit（模型+校验+API client）被 admin/H5/Android/iOS 四处编译，行为一致（diff 方法论复用）
- [ ] Android：`libkylix.so` 加载 + 导出函数调用 + 完整登录流程（真机或模拟器）
- [ ] iOS：`libkylix.a` 链接进 SwiftUI app + 模拟器完整登录流程
- [ ] H5：Lighthouse PWA 可安装性通过；弱网下降级可用
- [ ] 全量回归持续绿：16 包 + 双 sweep + bootstrap sweep + IR 不动点
