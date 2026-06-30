# Kylix 版本规划总结

## 📊 当前状态（v4.0-dev）

### 已完成的核心成果
✅ **LLVM M3 后端**：异常处理、控制流、表达式覆盖完整（73 测试）  
✅ **14/15 基础教程通过 LLVM 编译**（93% 通过率）  
✅ **stdlib Phase 7**：db/cache/http/websocket 四大模块  
✅ **VS Code 扩展 v1.1**：语法高亮 + LSP + 25 个代码片段  
✅ **16 包全绿**，350+ Go 测试，140+ stdlib 测试

### 技术债务
❌ Lambda/闭包（生成 null stub）  
❌ 多返回值元组赋值（静默注释）  
❌ inherited 关键字（未实现）  
⚠️ 无优化（比 Go 后端慢 2-5x）

---

## 🚀 v4.0.0 正式版（1 周内发布）

**目标日期**：2026-07-07  
**类型**：Major Release

### 发布清单
- [x] 文档更新（README/ROADMAP/CHANGELOG）
- [x] LLVM 使用指南（docs/llvm-backend.md）
- [ ] 教程全量验证（01-03 章节 LLVM，04-20 章节 Go）
- [ ] 交叉编译测试（linux/darwin/windows）
- [ ] 版本号提升 + git tag
- [ ] GitHub Release + 二进制发布

### 已知限制（文档化但不修复）
1. Lambda 不支持 → 使用命名函数
2. 多返回值 stub → 使用 record 类型
3. inherited 未实现 → 显式父类调用
4. 未优化 → 用 Go 后端开发

**决策**：快速发布，当前功能已达生产可用基础。

---

## 🔥 v4.1.0 — LLVM M4 高级特性（6-8 周）

**时间线**：2026-07 至 2026-08  
**主题**：LLVM 后端成熟化

### 四大优先级任务

#### 🥇 Priority 1: Lambda/闭包支持（Week 1-3）
**问题**：example15_lambda.klx 编译失败（undefined reference）

**方案**：
- 捕获变量分析（构建 capture list）
- 环境结构体生成（`%__env_f = type { i64, ptr }`）
- 函数指针降级（lambda body → 命名函数）
- 闭包结构体（`{func_ptr, env_ptr}` pair）

**交付**：
- ✅ 15+ 测试（简单捕获、多变量、嵌套、可变）
- ✅ example15_lambda.klx 通过
- ✅ 文档更新（移除 lambda 限制说明）

---

#### 🥈 Priority 2: 完整多返回值（Week 4）
**问题**：`(q, r) := DivMod(...)` 生成注释，变量未初始化

**方案**：
- 函数返回结构体类型（`%__ret_DivMod = type { i64, i64 }`）
- insertvalue 构建返回值
- extractvalue 拆包赋值

**交付**：
- ✅ 8+ 测试
- ✅ example16_multireturn.klx 正确运行（不只是编译）

---

#### 🥉 Priority 3: inherited 关键字（Week 5）
**问题**：父类方法调用不支持

**方案**：
- 类层次遍历查找父实现
- 生成直接函数调用（绕过 vtable）
- 正确传递 this 指针

**交付**：
- ✅ 6+ 测试
- ✅ 继承链教程示例通过

---

#### ⚡ Priority 4: 优化通道（Week 6）
**问题**：LLVM 代码比 Go 慢 2-5x

**方案**：
- `--llvm-opt` flag（O1/O2/O3 级别）
- 运行 LLVM opt 工具（内联、循环展开、DCE）
- 基准测试（Fibonacci、素数筛、字符串操作）

**交付**：
- ✅ `--llvm-opt=2` 可用
- ✅ 至少 30% 性能提升
- ✅ 性能报告（docs/llvm-performance.md）

---

### 成功指标
- ✅ **25+/35 教程通过 LLVM**（71%+ 覆盖率）
- ✅ **120+ LLVM 测试**（从 73 增长）
- ✅ **性能接近 Go 后端**（O2 优化下 1.5x 以内）
- ✅ **零 P0/P1 bug**

### 时间表
| 周数 | 里程碑 | 可交付物 |
|------|--------|----------|
| 1-3  | Lambda 完整实现 | 环境结构 + 闭包调用 + 15 测试 |
| 4    | 多返回值 | 元组拆包正确工作 |
| 5    | inherited | 父类方法调用 |
| 6    | 优化 | --llvm-opt + 基准 |
| 7    | 缓冲/打磨 | Bug 修复 + 文档 |
| 8    | 发布准备 | 测试 + CHANGELOG + tag |

---

## 🔮 v4.2.0+ 未来展望（3-6 个月后）

### 可能的方向
1. **LLVM stdlib Phase 1** — 用 LLVM 编译核心 stdlib 模块
   - strutil/mathutil/sysutil（纯 Kylix，无 Go wrapper）
   - 减少对 Go stdlib 的依赖

2. **增量编译** — 缓存 LLVM IR 每个模块
   - 只重新编译变更文件
   - 链接预编译 .o 文件

3. **调试符号** — 发出 DWARF 调试信息
   - 支持 GDB/LLDB 单步调试
   - `kylix build --backend=llvm -g`

4. **交叉编译** — 无需安装 LLVM 的目标构建
   - 预编译 IR，仅打包链接器
   - `--target=aarch64-linux-gnu`

---

## 🎯 v5.0.0 终极目标（长期，6-12 个月）

**愿景**：完全脱离 Go 依赖

### 核心工作
1. **自研运行时 KylixRT**
   - GC（标记-清扫或 Boehm GC）
   - 字符串/数组/映射（纯 C 实现）
   - 协程库或线程池（替代 goroutine）

2. **stdlib 纯 Kylix 重写**
   - 移除所有 `stdlib/*.go` 文件
   - 用 Kylix + C FFI 重写

3. **自举编译器**
   - Kylix 编译器用 Kylix 重写
   - `kylix compile kylix_compiler.klx --backend=llvm`
   - 生成的编译器可编译自己

### 里程碑
- ✅ LLVM 后端编译所有 stdlib 模块
- ✅ LLVM 后端编译 Kylix 编译器自身
- ✅ 生成的二进制零 Go 依赖

---

## 📈 优先级建议

### 近期（2-4 周）
1. ✅ **完成 v4.0.0 发布**（测试 + 文档）
2. 🔥 **启动 v4.1.0 Lambda 开发**（最高价值）

### 中期（2-3 个月）
3. 🔥 **v4.1.0 其他特性**（多返回值 + inherited + 优化）
4. 📚 **stdlib Phase 8**（可选：logging/profiling/reflection）

### 长期（6-12 个月）
5. 🚀 **v5.0.0 自举准备**（自研运行时 + stdlib 重写）

---

## 💡 立即可做的事

### 本周任务
1. **完成 v4.0.0 发布检查清单**
   - 运行教程全量测试
   - 验证交叉编译
   - 打 git tag

2. **开始 Lambda 设计**
   - 研究闭包实现（LLVM IR 示例）
   - 编写捕获变量分析器原型
   - 创建 lambda_test.go 框架

3. **监控社区反馈**
   - GitHub Issues 中的 bug 报告
   - 用户对 LLVM 后端的使用反馈

---

**创建日期**：2026-06-30  
**维护者**：Kylix Core Team  
**状态**：Active Planning

---

## 📞 需要决策的问题

1. **v4.0.0 发布时间**：
   - [ ] 快速发布（2-3 天，仅文档）
   - [ ] 稳妥发布（1 周，含测试验证）

2. **v4.1.0 人力投入**：
   - [ ] 专职 1 人（8 周完成）
   - [ ] 兼职或社区贡献（12+ 周）

3. **优先级调整**：
   - [ ] 按当前计划（Lambda → 多返回值 → inherited → 优化）
   - [ ] 先做优化（快速提升性能吸引用户）
   - [ ] 先做多返回值（相对简单，快速成果）

**建议**：保持当前优先级，Lambda 是最明显的缺失特性。
