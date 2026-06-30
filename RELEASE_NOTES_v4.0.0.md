# Kylix v4.0.0 Release Notes

**发布日期**: 2026-07-01  
**版本类型**: Major Release  
**代号**: LLVM M3 + stdlib Phase 7

---

## 🎉 概览

v4.0.0 是 Kylix 的重大版本，标志着 **LLVM 后端达到生产可用基础覆盖**，同时新增 stdlib Phase 7 四大模块和 VS Code 代码片段。LLVM 后端现在可以编译真实 Pascal 程序到原生二进制，无需 Go 工具链。

### 核心数字
- ✅ **14/15 基础教程通过 LLVM 编译到原生二进制**（93%）
- ✅ **49/49 教程通过 Go 后端测试**（100%）
- ✅ **16 个 Go 包全部测试通过**
- ✅ **73 个 LLVM 后端测试**
- ✅ **350+ Go 测试，140+ stdlib 测试**

---

## ✨ 新增功能

### 1. LLVM 后端 M3 — 完整异常处理

完整实现 Pascal 异常语义，路线 C（全局异常槽 + setjmp/longjmp + 类型 ID）：

```pascal
try
  raise TFooError.Create('bad');
except
  on E: TFooError do
    WriteLn('Custom: ', E.Message);
  on E: Exception do
    WriteLn('Generic: ', E.Message);
finally
  WriteLn('Cleanup');
end;
```

**支持**：
- `try/except/finally` 三段式
- `raise` 语句（带参数 + 裸 raise 重抛）
- `on E: Type do` 类型化捕获（运行时子类型匹配）
- 嵌套 try（栈式 handler 链）
- finally 确定性执行（3 路径代码复制）

**实现细节**：注入 Exception class + `@__kylix_is_subtype` 运行时函数 + 全局异常槽 4 个 global。避开 Itanium C++ EH ABI（手写 IR 下不可行 + 需 libc++abi 依赖）。

### 2. LLVM 后端 M3 — 控制流补全

补全所有缺失的控制流语句：

- `break` / `continue`（所有循环类型，嵌套正确）
- `case...of`（LLVM switch 指令）
- `match`（icmp eq 链 + OR 多模式 + when 守卫）
- `for...in`（foreach，strlen bound + getelementptr）

### 3. LLVM 后端 M3 — 表达式覆盖提升

- `WriteLn(a, b, c)` 多参数（512B buffer + strcat/snprintf）
- `WriteLn()` 零参数（空行）
- `ArrayLiteral`（malloc heap buffer + length prefix）
- `SliceExpression`、`TupleLiteral`、`AwaitExpression`（基础覆盖）

### 4. 关键 Bug 修复

- **多变量声明**：`var a, b: Boolean` 现在为所有变量正确 alloca
- **类型自动转换**：i1↔i64（zext/icmp）、i64↔double（sitofp/fptosi）
- **SSA dominance**：`__kylix_is_subtype` phi 节点修复
- **元组 LHS 赋值**：`(q, r) := func()` 降级为注释（IR 仍合法）

### 5. stdlib Phase 7 — 四大模块

#### `db` 模块（数据库便捷封装）
```pascal
uses db;
var db := DbOpenSQLite(':memory:');
DbExec(db, 'INSERT INTO users VALUES (?, ?)', 'alice', 30);
var count := DbQueryScalar(db, 'SELECT COUNT(*) FROM users');
```
支持 SQLite/MySQL/PostgreSQL，参数化查询防注入。

#### `cache` 模块（LRU 缓存）
```pascal
uses cache;
var c := NewCache(4, 0);
c.Put('key', 'value');
WriteLn(c.GetString('key'));
```
线程安全，O(1) 操作，TTL 过期 + Sweep 惰性回收。

#### `http` 模块增强
新增 `HttpPut`/`HttpDelete`/`HttpPostJSON` + `THttpResponse`（Status+Body）响应对象。

#### `websocket` 模块（RFC 6455）
纯 stdlib 实现 WebSocket 客户端 + 服务端，握手/文本帧/ping 自动 pong/close。

### 6. VS Code 代码片段（25 个）

之前 CHANGELOG 声称有 25+ 片段但文件缺失，现已补全：program/unit、function/procedure、class/record、控制流、try/except、WriteLn、KylixBoot controller/routes、ORM entity。

---

## 📊 测试覆盖

| 维度 | 数量 | 状态 |
|------|------|------|
| Go 测试包 | 16 | ✅ 全部通过 |
| Go 测试 | 350+ | ✅ |
| LLVM 后端测试 | 73 | ✅ |
| stdlib 测试 | 140+ | ✅ |
| Go 后端教程 | 49/49 | ✅ 100% |
| LLVM 后端教程（01-03） | 14/15 | ✅ 93% |

---

## ⚠️ 已知限制（不影响发布）

以下限制有文档化的 workaround，将在 **v4.1.0（LLVM M4）** 修复：

1. **Lambda/闭包**（LLVM 后端）— example15_lambda 编译失败
   - Workaround: 使用命名函数
2. **多返回值元组赋值**（LLVM 后端）— `(q, r) := func()` 静默 stub
   - Workaround: 使用 record 返回类型
3. **inherited 关键字**（LLVM 后端）— 未实现
   - Workaround: 显式父类方法调用
4. **LLVM 优化**（LLVM 后端）— 无优化，比 Go 慢 2-5x
   - Workaround: 用 Go 后端开发

详见 `docs/llvm-backend.md`。

---

## 📦 安装

### 从源码编译
```bash
git clone https://github.com/astra-zhao/kylix.git
cd kylix
go build -o kylix cmd/kylix/main.go
```

### 预编译二进制（GitHub Release）
- `kylix-linux-amd64`
- `kylix-darwin-amd64`（Intel Mac）
- `kylix-darwin-arm64`（Apple Silicon）
- `kylix-windows-amd64.exe`

### 使用 LLVM 后端（需 llc + clang）
```bash
brew install llvm          # macOS
sudo apt install llvm clang  # Linux
kylix build --backend=llvm program.klx
```

---

## 📚 文档

- [CHANGELOG.md](CHANGELOG.md) — 完整版本历史
- [ROADMAP.md](ROADMAP.md) — 开发路线图（v4.1.0/v4.2.0/v5.0.0）
- [VERSION_PLAN.md](VERSION_PLAN.md) — 版本规划总结
- [docs/llvm-backend.md](docs/llvm-backend.md) — LLVM 后端使用指南
- [docs/v4.1.0-plan.md](docs/v4.1.0-plan.md) — v4.1.0 详细计划

---

## 🙏 致谢

感谢所有贡献者。LLVM M3 异常处理采用了路线 C（setjmp/longjmp）方案，避免了 C++ EH ABI 的复杂性，使手写 IR 文本实现成为可能。

---

## 🚀 下一步

v4.1.0（LLVM M4）将解决本次的已知限制：
- 闭包/Lambda 支持（最高优先级）
- 完整多返回值
- inherited 关键字
- 优化通道（`--llvm-opt`）

预计 6-8 周。详见 `docs/v4.1.0-plan.md`。

---

**完整更新日志**: [CHANGELOG.md](CHANGELOG.md)  
**问题反馈**: https://github.com/astra-zhao/kylix/issues
