# Kylix C ABI 导出与多端/嵌入式集成指南

> 新增于 v0.14.0（2026-10）。Kylix 支持通过标准 C ABI 导出函数，
> 可直接编译为动态共享库（`.so` / `.dylib` / `.dll`）或静态库归档（`.a` / `.o`），
> 供 C/C++、Android（Kotlin + JNI）、iOS（Swift / Objective-C）等宿主环境直接加载调用。

- 关联文档：[MULTIPLATFORM.md](MULTIPLATFORM.md)（多端架构设计）、[API_STABILITY.md](API_STABILITY.md)（API 稳定性承诺）

---

## 一、导出语法

Kylix 使用属性注解 `[Export]` 声明对外导出的函数或过程：

```pascal
program MyCore;

// 1. 默认导出名与 Pascal 函数同名（区分大小写）
[Export]
function Add(a, b: Integer): Integer;
begin
  result := a + b;
end;

// 2. 自定义导出符号名（推荐用于跨语言命名规范或 JNI 符号）
[Export('core_calc_product')]
function Multiply(a, b: Integer): Integer;
begin
  result := a * b;
end;

// 3. 导出过程
[Export('core_log_message')]
procedure LogMsg(msg: String);
begin
  WriteLn('[KylixCore] ' + msg);
end;

begin
  // 顶层初始化逻辑：动态库加载时通过 @llvm.global_ctors 自动执行
end.
```

---

## 二、C ABI 数据交互协议与内存契约

### 1. 标量类型映射

| Kylix 类型 | C / C++ 类型 | JNI 类型 | Swift 类型 | 说明 |
|---|---|---|---|---|
| `Integer` | `int64_t` | `jlong` | `Int64` | 64 位有符号整型 |
| `Boolean` | `bool` / `int8_t` | `jboolean` | `CBool` / `Bool` | 布尔类型（0 / 1） |
| `Real` | `double` | `jdouble` | `Double` | 64 位双精度浮点数 |
| `Pointer` | `void*` | `jlong` / 指针 | `UnsafeMutableRawPointer?` | 原始指针 |

### 2. 字符串与复合结构：奉行「标量 + UTF-8 JSON 字符串」

为了规避跨语言间复杂结构体对齐（padding/alignment）和对象指针生命周期的不确定性，Kylix 推荐**复合数据一律采用 UTF-8 JSON 字符串**传输：

- 入参：传入以 null 结尾的 C 字符串指针（`const char*`）。
- 出参：返回以 null 结尾的 C 字符串指针（`const char*`）。

### 3. 内存所有权契约（`kylix_free`）

**核心原则：谁分配，谁释放。**

当 Kylix 导出的函数返回动态构造的 `String` 时，该内存在 Kylix 堆上分配。外部宿主语言在取回并拷贝/解析完毕后，**必须调用编译器自动注入的 `kylix_free` 函数释放内存**，防止跨 ABI 内存泄漏：

```c
// Kylix 编译器自动注入的标准导出函数
void kylix_free(void* p);
```

---

## 三、编译命令与产物格式

### 1. 本地动态库编译（`--shared`）

```bash
# macOS 输出 .dylib
kylix build --backend=llvm --shared -o libcore.dylib core.klx

# Linux 输出 .so
kylix build --backend=llvm --shared -o libcore.so core.klx

# Windows 输出 .dll
kylix build --backend=llvm --shared -o libcore.dll core.klx
```

### 2. 直接输出目标文件与静态库（`.o` / `.a`）

```bash
# 输出 .o 目标文件（跳过平台链接，适合接入外部构建系统）
kylix build --backend=llvm --shared -o core.o core.klx

# 输出 .a 静态库归档
kylix build --backend=llvm --shared -o libcore.a core.klx
```

---

## 四、多端交叉编译

### 1. Android（AArch64 / x86_64）

- **Target Triple**：
  - 真机设备：`--target android/arm64`（对应 `aarch64-linux-android30`）
  - 模拟器：`--target android/amd64`（对应 `x86_64-linux-android30`）
- **工具链探测**：
  Kylix 自动探测环境变量 `ANDROID_NDK_HOME`、`ANDROID_NDK_ROOT` 或 `ANDROID_HOME/ndk/*`。

```bash
# 1. 设置 NDK 路径
export ANDROID_NDK_HOME=/path/to/android-ndk

# 2. 编译 Android 动态库 (.so)
kylix build --backend=llvm --target android/arm64 --shared -o libkylixcore.so core.klx

# 3. 或直接产出 Android ELF 目标文件（无需本机安装 NDK）
kylix build --backend=llvm --target android/arm64 --shared -o core_android.o core.klx
```

### 2. iOS（AArch64 / 模拟器）

- **Target Triple**：
  - 真机设备：`--target ios/arm64`（对应 `arm64-apple-ios16.0.0`）
  - Mac 模拟器：`--target ios/simulator-arm64`（对应 `arm64-apple-ios16.0.0-simulator`）
- **工具链依赖**：
  在 macOS 上自动调用 `xcrun -sdk iphoneos clang` 链接。

```bash
# 输出 iOS 静态库（推荐用于 Swift Package 或 CocoaPods）
kylix build --backend=llvm --target ios/arm64 --shared -o libkylixcore_ios.a core.klx

# 输出 iOS 动态库
kylix build --backend=llvm --target ios/arm64 --shared -o libkylixcore_ios.dylib core.klx
```

---

## 五、各宿主调用实战示例

### 1. C / C++ 动态加载示例

```c
#include <stdio.h>
#include <dlfcn.h>
#include <stdint.h>

typedef int64_t (*AddFunc)(int64_t, int64_t);
typedef const char* (*GreetFunc)(const char*);
typedef void (*FreeFunc)(void*);

int main() {
    void* handle = dlopen("./libcore.dylib", RTLD_NOW);
    if (!handle) { return 1; }

    AddFunc add = (AddFunc)dlsym(handle, "Add");
    GreetFunc greet = (GreetFunc)dlsym(handle, "Greet");
    FreeFunc kylix_free = (FreeFunc)dlsym(handle, "kylix_free");

    // 1. 标量调用
    printf("10 + 20 = %lld\n", (long long)add(10, 20));

    // 2. 字符串与内存释放
    const char* str = greet("World");
    printf("Greeting: %s\n", str);
    kylix_free((void*)str); // 释放 Kylix 堆分配

    dlclose(handle);
    return 0;
}
```

### 2. Android Kotlin (JNI) 实战

Pascal 侧可以直接指定符合 JNI 规范的导出符号名：

```pascal
[Export('Java_com_example_app_KylixBridge_nativeAdd')]
function JniAdd(env: Pointer; thiz: Pointer; a, b: Integer): Integer;
begin
  result := a + b;
end;
```

Kotlin 侧直接声明 `external`：

```kotlin
package com.example.app

object KylixBridge {
    init {
        System.loadLibrary("kylixcore")
    }

    external fun nativeAdd(a: Long, b: Long): Long
}
```

### 3. iOS Swift 实战

通过头文件 `KylixCore.h` 桥接：

```c
// KylixCore.h
#pragma once
#include <stdint.h>

int64_t Add(int64_t a, int64_t b);
const char* Greet(const char* name);
void kylix_free(void* p);
```

Swift 侧无缝调用：

```swift
import Foundation

func testKylix() {
    let sum = Add(15, 30)
    print("Sum: \(sum)")

    if let cStr = Greet("iOS") {
        let message = String(cString: cStr)
        print("Message: \(message)")
        kylix_free(UnsafeMutableRawPointer(mutating: cStr)) // 回收内存
    }
}
```
