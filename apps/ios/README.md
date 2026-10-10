# KylixAdmin iOS 壳

SwiftUI + Swift Package 薄壳：登录，然后列出 Notes。业务协议在 `apps/shared/mobilecore.klx`，HTTP 用 URLSession。静态库 `libkylixcore.a` 只在 macOS 上能链出来。

第一次做请看 [小白教程](../../docs/MOBILE_TUTORIAL_CN.md)。命令和边界的全文在 [docs/MOBILE_APPS.md](../../docs/MOBILE_APPS.md)。

```bash
# 在 Mac 上，仓库根目录
bash apps/ios/build_core.sh simulator
cd apps/ios && xcodegen generate && open KylixAdmin.xcodeproj
```
