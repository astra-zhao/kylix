# stdlib.datetime — 日期时间处理模块

**状态**: ✅ Phase 1 完成 (LLVM 后端)  
**版本**: v4.3.0  
**测试覆盖**: 9/9 单元测试通过，example38 真机验证通过

## 概述

`datetime` 模块提供日期时间操作功能，基于 Unix time_t 和 libc `time.h` 实现。

## 类型

### TDateTime

不透明指针类型，内部存储 `time_t` (i64)。

```pascal
var dt := Now();
var birthday := MakeDate(1990, 5, 20);
```

## 函数

### Now() -> TDateTime

返回当前日期时间（Unix 时间戳）。

**实现**:
- 调用 `time(null)`
- `malloc(8)` 分配 TDateTime 实例
- 存储 time_t 值

**示例**:
```pascal
var now := Now();
WriteLn('Year: ' + IntToStr(now.Year()));
```

### MakeDate(year, month, day: Integer) -> TDateTime

构造指定日期（时间部分为 00:00:00）。

**参数**:
- `year`: 完整年份（如 2025）
- `month`: 月份 (1-12)
- `day`: 日期 (1-31)

**实现**:
- 分配 `struct tm` (56 字节，对齐 8)
- 用 `llvm.memset` 清零
- 设置 `tm_year` (year - 1900), `tm_mon` (month - 1), `tm_mday`
- 调用 `mktime()` 转为 time_t
- 存入 TDateTime 实例

**示例**:
```pascal
var xmas := MakeDate(2024, 12, 25);
WriteLn(xmas.FormatDate());  // 2024-12-25
```

## 方法

所有方法通过 `emitDatetimeMethodCall` 分发，接收者类型为 `TDateTime`（支持链式调用）。

### Year() -> Integer

返回年份（4 位）。

**实现**:
- 加载 time_t 值
- 调用 `localtime()` 获取 `struct tm*`
- GEP 到 `tm_year` (offset 5, i32)
- 加载并加 1900

**示例**:
```pascal
var dt := Now();
WriteLn(dt.Year());  // 2026
```

### Month() -> Integer

返回月份 (1-12)。

**实现**:
- `localtime()` → GEP `tm_mon` (offset 4, i32)
- 加载并加 1

### Day() -> Integer

返回日期 (1-31)。

**实现**:
- `localtime()` → GEP `tm_mday` (offset 3, i32)
- 直接加载（无需偏移）

### FormatDate() -> String

格式化为 `YYYY-MM-DD` 字符串。

**实现**:
- 分配 64 字节缓冲区
- 调用 `strftime(buf, 64, "%Y-%m-%d", localtime(self))`
- 返回 buf 指针

**注意**: 静态缓冲区，非线程安全。

**示例**:
```pascal
var dt := MakeDate(2025, 1, 15);
WriteLn(dt.FormatDate());  // 2025-01-15
```

### AddDays(days: Integer) -> TDateTime

返回加/减指定天数后的新 TDateTime。

**参数**:
- `days`: 天数（可为负）

**实现**:
- 加载原 time_t
- 计算 `days * 86400` (秒)
- `add i64` 得到新 time_t
- `malloc(8)` 创建新实例并存储

**示例**:
```pascal
var dt := MakeDate(2024, 12, 25);
var future := dt.AddDays(7);
WriteLn(future.FormatDate());  // 2025-01-01

var past := dt.AddDays(-7);
WriteLn(past.FormatDate());    // 2024-12-18
```

## 链式调用

支持方法链式调用（通过类型传播）：

```pascal
var dt := MakeDate(2025, 1, 1).AddDays(10);
WriteLn(dt.FormatDate());  // 2025-01-11
```

**实现细节**:
- `emitDatetimeCall` 返回类型标记为 `"TDateTime"`（不是 `"ptr"`）
- `emitVarDecl` 检测 `llvmType == "TDateTime"` 并记录到 `g.localTypes`
- `emitMethodCall` 通过 `receiverKind` 或表达式求值识别 TDateTime
- 支持变量接收者（`dt.Year()`）和链式接收者（`MakeDate().AddDays()`）

## LLVM IR 接口

### 类型
```llvm
; TDateTime 实例布局
%TDateTime = { i64 }  ; time_t (Unix 时间戳)
```

### 声明
```llvm
declare i64 @time(ptr)
declare ptr @localtime(ptr)
declare i64 @mktime(ptr)
declare i64 @strftime(ptr, i64, ptr, ptr)
declare void @llvm.memset.p0.i64(ptr nocapture writeonly, i8, i64, i1 immarg)
```

### 函数签名
```llvm
define ptr @__kylix_datetime_Now()
define ptr @__kylix_datetime_MakeDate(i64 %year, i64 %month, i64 %day)
define i64 @__kylix_datetime_Year(ptr %self)
define i64 @__kylix_datetime_Month(ptr %self)
define i64 @__kylix_datetime_Day(ptr %self)
define ptr @__kylix_datetime_FormatDate(ptr %self)
define ptr @__kylix_datetime_AddDays(ptr %self, i64 %days)
```

### 调用示例
```llvm
%dt = call ptr @__kylix_datetime_MakeDate(i64 2025, i64 1, i64 15)
%year = call i64 @__kylix_datetime_Year(ptr %dt)
%formatted = call ptr @__kylix_datetime_FormatDate(ptr %dt)
```

## Codegen 集成

### stdlib.go 注册
```go
var knownStdlibModules = map[string]bool{
    "datetime": true,
    ...
}

var stdlibModuleFuncs = map[string]map[string]bool{
    "datetime": {
        "Now": true, "Today": true, "MakeDate": true, 
        "MakeTime": true, "ParseDate": true, "ParseDateTime": true,
    },
}

func (g *Generator) emitStdlibCall(module, funcName string, args []ast.Expression) (string, string, error) {
    switch module {
    case "datetime":
        return g.emitDatetimeCall(funcName, args)
    ...
    }
}

func (g *Generator) emitPendingStdlib() {
    for _, sf := range g.stdlibQueue {
        switch sf.module {
        case "datetime":
            g.emitDatetimeBody(sf.name, sf.argCount)
        ...
        }
    }
}
```

### expr.go 方法调用
```go
func (g *Generator) emitMethodCall(member *ast.MemberExpression, args []ast.Expression) (string, string, error) {
    kind, typeName := g.receiverKind(member.Object)

    // TDateTime 特殊处理
    if typeName == "TDateTime" {
        objReg, _, err := g.emitExpr(member.Object)
        if err != nil {
            return "", "", err
        }
        return g.emitDatetimeMethodCall(objReg, member.Member, args)
    }

    if kind == "" {
        // 链式调用：求值检查是否为 TDateTime
        objReg, objType, err := g.emitExpr(member.Object)
        if err != nil {
            return "", "", err
        }
        if objType == "TDateTime" {
            return g.emitDatetimeMethodCall(objReg, member.Member, args)
        }
    }
    ...
}
```

### stmt.go 变量推断
```go
func (g *Generator) emitVarDecl(s *ast.VarDecl) error {
    ...
    valReg, llvmType, err := g.emitExpr(s.Value)
    ...
    inferredClass := ""
    if llvmType == "TDateTime" {
        inferredClass = "TDateTime"
    }
    ...
    for _, name := range s.Names {
        actualLLVMType := llvmType
        if llvmType == "TDateTime" {
            suffix = "_str"
            actualLLVMType = "ptr"
        }
        allocaReg := fmt.Sprintf("%%v_%s%s", name, suffix)
        g.line(fmt.Sprintf("  %s = alloca %s, align 8", allocaReg, actualLLVMType))
        g.line(fmt.Sprintf("  store %s %s, ptr %s", actualLLVMType, valReg, allocaReg))
        g.locals[name] = allocaReg
        if inferredClass != "" {
            g.localTypes[name] = inferredClass
        }
    }
}
```

## 限制

### 当前版本 (v4.3.0)
1. **静态缓冲区**: FormatDate 使用 64 字节栈缓冲区（非线程安全）
2. **无内存管理**: TDateTime 实例通过 malloc 分配，无 GC
3. **简化 tm 结构**: 假设 `struct tm` 布局为标准 POSIX（56 字节）
4. **缺少函数**:
   - `Today()` (存根，当前等同于 Now)
   - `MakeTime(h, m, s)`
   - `ParseDate(s)` / `ParseDateTime(s)`
5. **缺少方法**:
   - `Hour()`, `Minute()`, `Second()`
   - `DayOfWeek()`, `DayOfYear()`
   - `Format(pattern)`（当前仅 `FormatDate`）

### 平台依赖
- 依赖 libc `time.h` 实现
- `struct tm` 布局可能因平台而异（当前假设 macOS/Linux）
- Windows 需验证 `mktime` / `localtime` 兼容性

## 测试

### 单元测试 (pkg/llvmgen/stdlib_datetime_test.go)
```bash
go test -v -run TestDatetime ./pkg/llvmgen
```

**覆盖**:
- `TestDatetimeNow`: time() 调用 + malloc
- `TestDatetimeMakeDate`: mktime + struct tm 初始化
- `TestDatetimeYear/Month/Day`: localtime + GEP 偏移
- `TestDatetimeFormatDate`: strftime 调用
- `TestDatetimeAddDays`: 天数算术
- `TestDatetimeChainedCalls`: 链式调用类型传播
- `TestDatetimeLibcDeclarations`: 外部声明完整性

### 集成测试
```bash
/tmp/kylix_bin build --backend=llvm -o /tmp/test \
    examples/complete-tutorial/08_stdlib_utils/example38_datetime.klx
/tmp/test
```

**预期输出**:
```
Current year: 2026
Current month: 7
Current day: 3
Christmas 2024: 2024-12-25
One week later: 2025-01-01
One week before: 2024-12-18
```

## 下一步 (Phase 2+)

### 优先级 1: 补全核心 API
- `MakeTime(hour, minute, second)`
- `Hour()`, `Minute()`, `Second()` 方法
- `Format(pattern)` 自定义格式化
- `ParseDate(str)` 字符串解析

### 优先级 2: 增强功能
- `DayOfWeek()` (0=Sunday)
- `AddMonths()`, `AddYears()`
- `Diff(other)` 时间差
- 时区支持（UTC vs Local）

### 优先级 3: 内存管理
- TDateTime 引用计数 / GC 集成
- FormatDate 动态内存分配（避免静态缓冲区）

### 优先级 4: 跨平台
- Windows `_mktime64` / `_localtime64` 适配
- `struct tm` 布局检测（编译时 offsetof）

## 相关文档
- [sysutil.md](./sysutil.md) — 文件/路径工具
- [regex.md](./regex.md) — 正则表达式验证
- [LLVM stdlib 架构](../llvm_stdlib_architecture.md)

## 版本历史

### v4.3.0 (2026-07-03)
- ✅ Phase 1 完成
- 实现 Now/MakeDate/Year/Month/Day/FormatDate/AddDays
- 支持链式调用
- 9/9 单元测试通过
- example38 真机验证通过

### 未来计划
- v4.4.0: Phase 2 (MakeTime/Hour/Minute/Second/Format)
- v4.5.0: Phase 3 (解析/时区/差值计算)
