# KylixAdmin Android 壳

Kotlin + JNI 薄壳：登录，然后列出 Notes。业务协议在 `apps/shared/mobilecore.klx`，HTTP 用 OkHttp。

完整步骤（NDK、模拟器、和 iOS 同一套验收）见 [docs/MOBILE_APPS.md](../../docs/MOBILE_APPS.md)。

```bash
# 仓库根目录，NDK 已安装
bash apps/android/build_core.sh arm64
# 然后用 Android Studio 打开 apps/android/
```
