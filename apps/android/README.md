# KylixAdmin Android 壳

Kotlin + JNI 薄壳：登录，然后列出 Notes。业务协议在 `apps/shared/mobilecore.klx`，HTTP 用 OkHttp。

第一次做请看 [小白教程](../../docs/MOBILE_TUTORIAL_CN.md)。命令和边界的全文在 [docs/MOBILE_APPS.md](../../docs/MOBILE_APPS.md)。

```bash
# 仓库根目录，NDK 已安装
bash apps/android/build_core.sh arm64
# 然后用 Android Studio 打开 apps/android/
```
