# Kylix 项目上下文

Kylix 是现代 Pascal → Go 转译器。编译器用 Go 编写，生成 Go 代码。

**重要：始终用中文回答用户，不论用户用什么语言提问，回复一律使用中文。**

## 当前状态：v0.6.9 已发布（2026-09-04）

- v0.6.9 已发布：**bootstrap 无 Go 闭环达成（P3+P4+P4.12）**——(1) **stdlib IR 烘焙**：`scripts/extract_stdlib_ir.py` 把 host 生成的 stdlib IR 按 13 段烘焙进 `src/stdlib_ir.klx`（6.9k 行数据 + 139 签名），bootstrap 只做 call-site dispatch + wrapper 类方法（TCache/THttpClient/TDateTime）+ NewCache/Now/Today 内联——免手写 15.5k 行 Go 移植。(2) **emitter 大规模补缺 20+ 项**：数组写路径、not/负号、float 字面量、Variant 比较、调用参数类型化、dot-name 外部方法（host v0.5.4 缺口）、链式成员/receiver、record 类型系统、ClassName↔ptr coerce、epilogue 重排、嵌套循环 LoopBreak 保存恢复、alloca hoisting、构造函数 calloc、FloatToStr/StrToFloat 内置等。(3) **🎉 gen2 编译器诞生 + IR 不动点达成**：gen1（host 编译）`--emit-llvm` 9 文件 → mem2reg/llc/clang → gen2（纯原生无 Go）→ gen2 输出与 gen1 **逐字节一致（~220k 行 IR 不动点）**，gen3 ≡ gen2 行为；P4.10 破案 gen1 自举 emit"递归级联"（host emitFor 把 for 计数器绑到程序级全局槽 `@__kylix_g_i`，污染后末次迭代无限重跑→DEPTH-CUT；修复：`@__kylix_g_*` 绑定时铸造新 alloca + 出口写回）；P4.11 破案最后一个分歧（自举 emitter `array of Boolean` 写 i1 GEP/读 i64 GEP stride 不一致 → `ArrDynamic` 改 `array of Integer` 0/1）；调试探针网 72 处全清（EProg 内建保留）。(4) **🎉 教程 sweep 50/51 PASS**（`scripts/test_bootstrap_all.sh`，bootstrap-vs-host 输出逐字 diff；example33 多文件为 host 端 SKIP）；P4.12 修复最后两个已知失败：**example15**（自举无捕获 lambda——Pass 2.5 预发射 `@__lambda_N` + EmitPlainCall 查表分发 + 独立创建点计数器 `LambdaSiteCount`）、**example50**（JwtSign firstSlot alloca 提升到 entry，host + 烘焙数据双端同步）；修复后 IR 不动点复验保持（fp19 ≡ g5_all，gen3 ≡ gen2）。(5) **bootstrap 编译器坑清单**（自举开发必读，详见 CHANGELOG，共 9 条）：`or/and` 不短路、复合 `and`/`or` 条件编译出 `call @(i1)`（一律嵌套 if）、for 循环变量不遮蔽、`| tail -1` 吞 rebuild 失败、`array of Boolean` 字段数组不可用、探针删除残留悬空 if、main body 是单个 TBlockStatement 须解包。**O(n²) 性能疑虑已实测排除**（自举 emit 9 文件 ~4.8s）。详见 CHANGELOG.md
- v0.6.8 已发布（2026-08-23）：**boot server 补强 + stdlib 补全 + JetBrains 插件完善**——POST body 读取 + req.JSON 绑定 + BootRegisterJwtAuth 真校验；encoding Base64URL + httpclient JSON 嵌套；`.klx` 图标 + Run 配置 + LSP 错误高亮。详见 CHANGELOG.md
- v0.6.7 已发布：**#9 JetBrains 插件 + 安装使用手册**——(1) **`jetbrains-plugin/` 模块**（完整 Gradle Kotlin 项目，IC 2024.3 SDK）：TextMate 语法高亮（复用 vscode-ext tmLanguage，`com.intellij.textmate.bundleProvider` 扩展点）+ **LSP4IJ 桥接 `kylix lsp`**（补全/跳转/重命名/格式化全通）+ 25 个 Live Templates（prog/func/class/controller 等）。(2) **安装使用手册** `jetbrains-plugin/README.md`（环境要求 / 构建安装 / 使用 / 配置 / 故障排查）。(3) **构建验证**：`./gradlew buildPlugin` BUILD SUCCESSFUL，产出可安装 zip（1.6MB）。**ROADMAP #9 ✅**。**剩余（归入 v6.8+）**：boot server 补强（POST body / req.JSON / BootRegisterJwtAuth 真校验）、stdlib 补全（encoding Base64URL / httpclient JSON 嵌套）、net Winsock / regex pcre2（Windows）、内存管理（arena/GC）、bootstrap 无 Go 闭环。详见 CHANGELOG.md
- v0.6.6 已发布：**boot HTTP server + stdlib 补全 5 项**——(1) **boot HTTP server**（KylixBoot 应用无 Go 环境真正可用）：`Boot<M>` 写路由表（`@__kylix_boot_routes`）、`BootRun` 真体（`TcpListen/Accept` + `read_headers` + `parse_request` + `route_lookup/path_match` + TRequest handle + HTTP/1.1 响应 + 404）、`req.Param/Query/Header/Body` 内联降级（`stdlib_boot_http.go`）。(2) **stdlib 补全**：jwt claims（Verify 返回 claims map + exp 过期检查 + `JwtSubject/GetString/GetInt` + Sign extraClaims + **b64url rem==1 潜伏 bug**）、cache TTL（`PutWithTTL/Get/Sweep` + `htab_keys` + `now_ms`）、httpclient JSON（`HttpGetJSON` 返回 Variant map）、UrlEncode/UrlDecode、Variant div/mod。(3) **顺带修复**：Variant 赋值 as_str 误 coerce、emitCall 参数 variant→ptr、hashtab 门控、httpclient Request 漏 enqueue DoRequest。16 包 + 51 教程（Go+LLVM）+ self-repro 不动点全绿。**剩余（归入 v6.7+）**：#9 JetBrains 插件、net Winsock / regex pcre2（Windows）、内存管理（arena/GC）、bootstrap 无 Go 闭环。详见 CHANGELOG.md
- v0.6.5 已发布：**WS 自回环 + SHA-1 修复 + KylixRT 完善 + 性能优化**——(1) **手写 SHA-1 修复**（hs 初始常量错位/错值 + padLen 边界，websocket 不再依赖 OpenSSL）。(2) **WS 自回环分阶段 API**（WsDialConnect/WsDialFinish，LLVM+Go 双端，单进程自回环教程 example55）。(3) **KylixRT**：字符串插值 256B 溢出修复 + test/bench auto 回退 + run 错误消息 + doctor bundle 检查。(4) **性能**：DCE 单遍 + 缓存提前 + IR 确定性 + disable-verify + mem2reg——**bootstrap -O0 11.5s→0.381s（30×）**。**剩余（归入 v6.6+）**：#9 JetBrains 插件、boot HTTP server、net Winsock / regex pcre2（Windows）、内存管理（arena/GC）、stdlib 补全。详见 CHANGELOG.md
- v0.6.4 已发布：**LLVM stdlib 真实现：DbQueryRows + websocket**——(1) **DbQueryRows**（返回 `array of Variant`，每行 map-Variant box）：Variant 扩展 map 标签（`varTagMap=5`）+ `row['col']`/`rows[0]['col']` variant-map 索引 + 推断路径 slice/variant 修复（`var rows := DbQueryRows(...)`、`var v := JwtVerify(...)` 正确推断，顺带修 `var a := [1,2,3]`）+ 内联 sqlite3 prepare/step/column 循环 + htab 行构造 + 结果数组 append。(2) **websocket 完整 5 函数**（RFC 6455 客户端+服务端）：WsDial/WsAccept 握手（strcat 请求链规避 LLVM -O0 varargs spill 崩溃；**SHA-1 用 OpenSSL** 规避手写 IR bug）、WsSend/WsRecv 帧编解码（mask + 126/127 扩展 + ping/pong 自动应答）、WsClose；精确长度 recv/send helper + 带长度 base64。**剩余（归入 v6.5+）**：net Winsock / regex pcre2（需 Windows 真机）、#9 JetBrains 插件。详见 CHANGELOG.md
- v0.6.3 已发布：**jwt 双端 + 分发 B + Variant 传参修复**——(1) **jwt JwtSign/JwtVerify 真实现**（HS256：base64url(header) "." (payload) "." (HMAC-SHA256)，JwtSign 签名与 Python 逐字节一致；JwtVerify 验签版 valid / wrong-secret / tampered / malformed 全对）+ 手写 IR helper（b64url/hexdecode）。(2) **分发 B 捆绑 LLVM**：`scripts/bundle_llvm.sh` 捆绑 llc/opt + dylib 到 kylix 旁 `llvm/`，`FindLLVM` 可执行文件旁优先（编译自包含、链接用系统 clang）。(3) **Variant 传参 segfault 修复**：`Has(JwtVerify(...))` 嵌套调用根因 `isVariantType` 不认 `*ast.VariantType` → 参数 coerce 把 box `as_str`；修复 + box IR 类型校正 ptr + **闭包/inherited/virtual call 三处同类 coerce 点全修**（`closureKylixParams` / `MethodInfo.ParamKylixTypes`）。**剩余（归入 v6.4+）**：net Winsock / regex pcre2（需 Windows 真机）、DbQueryRows、websocket、#9 JetBrains 插件。详见 CHANGELOG.md
- v0.6.2 已发布：**跨平台 + KylixRT 生产化**——(1) **跨平台**：IR target triple 参数化（`CompileOpts.Target` + `tripleFor` + llc `-mtriple`）+ CLI `--target` 贯通 + 系统库链接平台化 + 修 Linux PIE 重定位（`-no-pie`）→ **Linux 51 教程 LLVM 51/51**（CI 稳定绿）；`kylix doctor` 预检 + README LLVM 前置（分发 A）。(2) **平台 API 适配**（按 targetOS）：sysutil（fopen 二进制模式 + `_access`）、datetime（`localtime_s`）、exc（`_setjmp/_longjmp` + jmpbuf 512）、net/regex（Windows stub）。(3) **stdlib sysutil 补齐**（9 函数：DirExists/CreateDir/DeleteFile/AppendFile/CopyFile/GetWorkingDir/SetWorkingDir/GetTempDir/GetFileSize）。(4) **`kylix test/bench --backend=llvm`**（KylixRT 无 Go 跑测试/基准——Kylix harness + LLVM multifile 编译 + 原生二进制）。(5) **修裸过程调用**（`Foo;` 无括号生成 call）+ Args[N] 索引 + MergePrograms 测试文件合并 3 个 LLVM bug。Windows job 尽力而为（runner LLVM 装不上）。**剩余**：net Winsock / regex pcre2 真实现、jwt 真实现、DbQueryRows、websocket（均需专门工作，记录 limitation）。
- v0.6.1 已发布：**KylixRT 核心**——`kylix run` 无 Go 工具链环境直接产出原生二进制并运行（`--backend=auto` 探测，有 Go 走 Go / 无 Go 回退 LLVM）；LLVM 后端补齐 KylixBoot 注解自动装配（`TRequest/TResponse`→opaque ptr 类型映射 + TResponse 真实句柄 `{status, body}` + `scanBootAnnotations`/`emitBootAutoWiring` 无闭包 wrapper 方案 + boot 函数表驱动 stub）；httpclient 补缺（Post 三参 + timeout 毫秒 + 8 一次性助手 + Put/Delete/StatusCode + THttpResponse）；db 补缺（DbOpen 转发 + DbExec 返回行数）；**51 教程 LLVM 编译+运行 51/51**（新 `test_all_llvm.sh` + CI `llvm-tutorials` job）。**v0.6.1 修复（2026-08-11）**：LLVM bootstrap 编译器编译含声明程序死循环——根因 1 无方法类 vtable=null 使 `is` 检查永远 false（emitVtable 发空 vtable + emitConstructor 恢复 store @X_vtable + record vtable，删除 emitClassRuntime 重复兜底）；根因 2 for/forEach `continue` 跳循环头跳过递增（continue 改跳独立 incLbl 递增块）。修复后 LLVM main 秒出编译 is/as、for+continue 程序。顺带修复 4 组潜伏回归。详见 CHANGELOG.md

- v0.6.1 已发布：**KylixRT 核心**——`kylix run` 无 Go 工具链环境直接产出原生二进制并运行（`--backend=auto` 探测，有 Go 走 Go / 无 Go 回退 LLVM）；LLVM 后端补齐 KylixBoot 注解自动装配（`TRequest/TResponse`→opaque ptr 类型映射 + TResponse 真实句柄 `{status, body}` + `scanBootAnnotations`/`emitBootAutoWiring` 无闭包 wrapper 方案 + boot 函数表驱动 stub）；httpclient 补缺（Post 三参 + timeout 毫秒 + 8 一次性助手 + Put/Delete/StatusCode + THttpResponse）；db 补缺（DbOpen 转发 + DbExec 返回行数）；**51 教程 LLVM 编译+运行 51/51**（新 `test_all_llvm.sh` + CI `llvm-tutorials` job）。**v0.6.1 修复（2026-08-11）**：LLVM bootstrap 编译器编译含声明程序死循环——根因 1 无方法类 vtable=null 使 `is` 检查永远 false（emitVtable 发空 vtable + emitConstructor 恢复 store @X_vtable + record vtable，删除 emitClassRuntime 重复兜底）；根因 2 for/forEach `continue` 跳循环头跳过递增（continue 改跳独立 incLbl 递增块）。修复后 LLVM main 秒出编译 is/as、for+continue 程序。顺带修复 4 组潜伏回归。详见 CHANGELOG.md
- v0.6.0 已发布：#7 性能 benchmark——`Result` 加 Duration/CacheHits/CacheMisses + `build --time` flag + `benchmarks/compile_time.sh` + `docs/compile-performance.md`。数据（Apple Silicon/LLVM 22）：Go 冷编译 29ms / 热 23ms / LLVM -O0 11527ms / **LLVM -O2 575ms**（opt 缩小 IR 后 llc 反而快 20×）。**#8 -O2 优化验证**：51 教程 `--llvm-opt=2` 全过，-O0 vs -O2 输出 diff 逐字节一致（`test_all_llvm.sh` 加 `LLVM_OPT`/`OUTDIR` 支持）。**klx codegen 修复（host 端）**：`var` 输出参数 → Go 指针（ast.Parameter.IsVar + parser 组延续 + 签名 `*` + body 解引用 + 调用 `&`）；as 链式确认 klx 已支持（v5.9 CastBoundary）。剩余：#9 JetBrains 插件、LLVM/klx 端 var 参数同步。
- v0.5.9 已发布：多态 gate 缺口修复 + KylixBoot 注解自动装配移植完成。宿主编译器与 bootstrap 编译器（`src/*.klx`）对同一源码的基类发射收敛为一致（`type TNode interface`；3 处根因：GenerateClassDecl interface 分支 / CollectClassTypes 无条件填充 ClassIsBaseStr / 类型表达式多态分支发 ident.Value）。KylixBoot 注解全栈移植到 `src/generator.klx`：#2 validation（GenerateValidationMethods）、#4 autowire（ScanBootAnnotations + EmitBootAutoWiring）、#5 ORM（ScanORMAnnotations + GenerateORM*Methods，`[Entity]`/`[Repository]`/`[Query]` → ToRow/FromRow + CRUD + Query 方法）。self-reproduction 不动点保持（`self_gen2 ≡ self_gen3`，7388 行逐字节一致），bootstrap 编译 51 教程 **51/51**，go test 16 包全绿。**v5.9 已知限制**：`var` 输出参数值传递（host 已修 v0.6.0，LLVM/klx 端待同步）；`(as T).Field` 链式 v5.9 已修且 v0.6.0 确认 klx 支持。详见 CHANGELOG.md
- v0.5.6 已发布：LLVM 后端 bootstrap self-host 达成 51/51（100%）。自举源码 `src/*.klx`（7 文件、5250 行）经 LLVM 后端多文件构建成原生二进制 `main_self`（**无 Go 依赖**），编译全部 51 个教程示例产出的 Go 代码能 `go build` 成功并正确运行。本轮修复 28 个 codegen bug + 移植缺口，分三类：(1) **LLVM codegen bug**（8 个）——Exit/return funcExit 出口块（v0.5.5"整数解析失败"真因）、裸无参方法调用 self.X; → emitMethodCall、strconcat 固定 512 缓冲区溢出、字符串 null 守卫（@__kylix_emptystr）、htab_get miss 返回 null、map array 值 miss zeroinitializer、ParseRepeatStatement 条件后 NextToken、GenerateExceptionTypes 移到 body 后；(2) **移植缺口**（14 个）——record/enum 值类型 + emitMapFieldIndexPut、SkipAttributes 跳过 [Attribute] 注解、stdlib 模块函数启发式 + (T,error) 包装 + boot 类型 TRequest→*stdlib.BootRequest、TLambdaExpression→Go func literal、WriteEscapedGoString 用拼接发转义（LLVM decodeKylixString 把 \\→\）、多返回 result:=(a,b)→return a,b + 解构、泛型实例化 TStack<Integer> 解析+构造+receiver [T]、validation stub IsValid()、unit interface/implementation 段标记 + forward 声明跳过；(3) **调试要点**（6 个）——lldb 函数名断点、main.ll IR 优先、in-repo sweep（kylix/stdlib 导入须 repo 解析）、无下划线文件名（Go 忽略 _ 前缀）、多文件 unit 构建（`build unit.klx main.klx`）、ParseGenericInstantiation 不消费 >（留 PeekToken=. PREC_MEMBER 让 infix loop 继续）。关键发现：generator.klx vs generator.go 是两套代码，差异是移植缺口；LLVM addString decodeKylixString 把 \\→\、\"→"（因 lexer 留 raw bytes、后端 decode），WriteEscapedGoString 须用拼接（各部分独立 decode，运行时 concat 产生正确字节）。回归 16 包 + 51 教程全绿。
- v0.5.3 已发布：自举编译器 round-trip 打通 + 自繁殖。v0.5.2 只打通「构建」；v0.5.3 打通「运行时正确性」——`kylix_self2` 能正确编译程序，且自繁殖（`kylix_self3` 同样正确）。三处修复（都在 `src/generator.klx`，宿主零改动）：`Args` builtin、条件导入（扫描 needle 拆分避自检测）、`WriteEscapedGoString` 2-char 前瞻转义保护。`self_7.go` ≡ `self_7_gen2.go`（5390 行逐字节一致，真正不动点）。
- v0.5.2 已发布：自举编译器构建打通。自举源码 `src/*.klx`（7 文件、5250 行）首次构建成可运行的 `kylix_self` 二进制。此前转译产物 `go build` 失败 208 个错误（130×`is`/`as` 在 struct 指针基类上非法、~75×子类装不进基类容器无多态、1×切片协变、3×`Args` builtin 缺失）。根因：Go 后端把所有类发射成普通 struct + 嵌入父 struct——给字段继承但无多态；`classIsBase` map 早已填充但 v0.3.1 回退后从不读取。**opt-in 修复**：仅当程序含 `is`/`as`（多态信号）时，把「有子类的基类」发射成空 interface（自举三基类 TNode/TStatement/TExpression 无字段无方法，且从不通过基类变量直接访问字段 → 空 interface 足够，无需 getter）；否则保留 struct 嵌入（字段继承，教程 example19/example40 不回归）。Parser 端 `parseIs/parseAs` 置 `program.UsesPolymorphism`，`collectClassTypes`（公共预扫描咽喉）OR 进 `g.usesPolymorphism`；`generateClassDecl`/`generateTypeExpression`/`generateTypeExpressionForCast` 按标志派发。新增 `Args` builtin（`os.Args[1:]`）；`src/parser.klx:448` 切片协变修复（`array of TStatement`→`array of TNode`）。`go build` 208→0 错误 → `kylix_self`（2.9MB）运行产出 5238 行 Go 编译器代码。go test 16 包全绿（+3 多态测试），教程 51/51 无回归。round-trip（kylix_self 产出再编译能正确编译程序）留 v5.3（自举 generator.klx codegen 保真度缺口）。
- v0.5.1 已发布：完成 Variant 运行时。补齐 v0.5.0 留的两个缺口：(A) `map[String]Variant` 真实化——htab 值槽存 Variant box 指针（不动 htab 结构，cache/string-map 不回归），`emitMapVarDecl` 检测 Variant 值类型设 `variantMaps`，`emitMapIndexGet` 走新 `htab_get_variant`（miss 返回 nilbox 全局 tag=0 → as_* 走 nil 默认）返回 `"variant"`，`emitMapIndexPut` 装箱 RHS；jsonutil `parse_flat` 改调 `value_to_variant` 让 `JsonDecodeMap` 产出真实 Variant map，`JsonGetString/Int/Float/Bool/Map/Array` 全部改 unbox（`variant_as_str/int/double/bool`）。(B) Variant 算术——`variant_add/sub/mul/div` 按标签派发（`+` 字符串拼接/双 int→int/else double），`emitInfix` 算术 stub 替换为 `emitVariantArith`，`coerceValue` 加 variant→concrete（`n := v` 解箱，`emitAssign` 内联 coercion 改调 coerceValue 统一）。新增 `as_int`/`as_bool` unbox + `@__kylix_variant_nilbox` 全局。LLVM 测试 266→274，教程 50→51（新增 example57_variant_map 双端输出逐字节一致）无回归。Variant 算术仅 LLVM（Go `interface{}`不支持运算符），`div`/`mod` Variant 留 stub。
- v0.5.0 已发布：Variant 运行时（标量 + `array of Variant`）。LLVM 后端此前把 `Variant` 静默当 `i64` 别名——`var v: Variant; v := 1.0` 截断 double、`array of Variant` 元素无类型标签、`arr[0] = 10.0` 比较位模式。v0.5.0 实现 boxed-pointer Variant（`%struct.kylix_variant = {i32 tag, i64 payload}`，tag nil/int/float/str/bool）：标量 `_var` alloca + Variant 数组 `arrayInfo.IsVariant`，赋值按类型装箱（box_int/float/str/bool），比较经 `variant_compare` 按标签派发（数值提升 double、字符串 strcmp、布尔 payload），`WriteLn(variantValue)` 按标签打印。jsonutil `JsonGetArray` 从 v4.9 字符串数组切片升级为**带类型标签的 Variant box 切片**（`value_to_variant` 窥首字符分类，数字→float box 与 Go json float64 对齐 → 双端 parity）。顺带修复 `Length(arr)` 路由（`emitArrayLength` 死代码 → 现派发 slice len word）。LLVM 测试 255→266，教程通过率 49→50（新增 example56_variant 双端输出逐字节一致）无回归。Variant 算术（`v+1`）与 `map[String]Variant` 真实化留 v5.1。
- v0.4.9 已发布：DWARF 调试信息 Phase 2 + jsonutil 嵌套数组。类方法/lambda 注册独立 DISubprogram（define 行附 `!dbg`、`self`/参数/捕获变量声明为调试局部变量），v0.4.8 泛型类方法可逐行单步 + LLDB 检视 receiver/捕获值。新增 DILexicalBlock——块内 `var` 归属正确的嵌套作用域。jsonutil `JsonGetArray` 从返回 null 的 stub 升级为真实解析器：把 JSON 数组解析为字符串数组 slice `{ptr items, i64 len, i64 cap}`（标量存文本、嵌套对象/数组存 raw 子串），新增 `JsonArrayLen`/`JsonArrayGetString`。顺手修复 `skip_nested` 丢失闭合 `]`/`}` 的 off-by-one（length `end-start` → `endAfter-start`）。LLVM 测试 250→255，教程通过率 49/49（100%）无回归。
- v0.4.8 已发布：泛型类方法 codegen + DIBasicType 多类型。修复 example21 泛型类 stub（`TStack<Integer>.Push/Pop` 从 `Pop: 0` → 与 Go 后端一致 `Pop: 30`），打通泛型类 `var x := TStack<T>.Create()` → `x.Method()` 完整链路：单态化 walk VarDecl.Value + constructor inference 处理 GenericType/CallExpression + 类字段数组 `self.Items[i]` GEP（FieldInfo.ArrayType + emitArrayIndex MemberExpression 分支）。DWARF 调试信息从单一 int64 升级为按 llvmType 发射独立 DIBasicType（double→DW_ATE_float、ptr→DW_ATE_address、i1→DW_ATE_boolean），LLDB `frame variable` 显示正确类型。LLVM 测试 249→250，教程通过率 48/48（100%），example21 从 stub → 输出正确。
- v0.4.7 已发布：静态数组下界修复 + jsonutil 嵌套对象解析。AST `ArrayType` 新增 `LowerBound` 字段，parser 记录真实下界，LLVM 后端按真实下界调整索引（不再硬编码 1）——修复 example23 段错误（`array[0..4]` 的 `0-1` 无符号下溢 → GEP 越界）。jsonutil `JsonGetMap` 从返回 null 升级为递归 `parse_flat` 解析 raw JSON 子串为 nested htab（支持任意深度嵌套对象），并修复 `skip_nested` 的 pos bug（指向 close char 之后，不再丢失 sibling 字段）。LLVM 测试 240→249，教程通过率 48/48（100%），example23 从段错误 → 输出正确。
- v0.4.6 已发布：DWARF 逐行调试升级 —— per-instruction DILocation（每条 IR 指令附 `!dbg !N` 源行号+列号+scope，按 (line,col,scope) 去重）+ DILocalVariable + `#dbg_declare` 记录（LLVM 22 语法，替代废弃的 `call @llvm.dbg.declare`）。`emitStatement`/`emitExpr` 入口 `setDbgNode` 设置源位置，`line()` 自动给指令行附加 `!dbg`。LLDB 支持按源文件行号设断点、`step`/`next` 逐行单步、`frame variable` 检视局部变量（参数/`result`/用户变量）。LLVM 测试 240→247，教程通过率 48/48（100%）无回归。
- v0.4.5 已发布：LLVM stdlib Phase 3 完成 —— 3 个 stub 模块升级为真实实现（jsonutil 递归下降解析器 / crypto AES-256-CBC+PBKDF2 / httpclient libcurl 集成）+ 进程内 IR 优化 pass 管线（DCE）+ 增量编译缓存（llc 跳过，32x 加速）+ DWARF 调试符号（`-g` flag，LLDB/GDB 函数级调试）+ 文件拆分（expr.go 1207→777、stmt.go 1081→614，回到 1000 行约束内）。LLVM 测试 198→240，教程通过率 48/48（100%）。
- v0.4.4 已发布：LLVM stdlib Phase 2 完成 —— 8 个模块（encoding/net/crypto/db/cache/jsonutil/boot/jwt/httpclient，~2000 行 IR + 60+ 单元测试）+ KylixBoot 注解方法 stub 生成 + 链式方法调用修复（`self.Repo.Name()` 类型追踪）+ 9 个关键 bug 修复（字符串比较/块作用域/ptr-nil 比较/map 后缀/...）。LLVM 教程通过率 48/48（100%，含 example33 多文件模块）。
- v0.4.3 已发布：datetime 模块 Phase 1 完整（13 API + Arena Allocator）
- v0.4.2 已发布：sysutil 模块 Phase 1（8 API）
- v0.4.1 已发布：LLVM M4 高级特性 —— Lambda/闭包（捕获变量 + 环境结构体）、`inherited` 关键字（父类方法链查找）、完整多返回值元组解构、OOP 字段/方法访问系统性修复（vtable 继承）、优化通道（`opt` + `llc -O<N>`，循环归纳达 20x 提速）。LLVM 教程通过率 27/49，01-04 章节（19 文件）与 Go 后端输出逐字节一致。
- v0.4.0 已发布：LLVM M3（异常处理/字符串插值/控制流/表达式覆盖 ✅）+ stdlib Phase 7（db/cache/http/websocket ✅）+ IDE 插件（VS Code v1.1 ✅）
- v0.3.3：KylixBoot 框架完善 —— Body 绑定 + JWT + OpenAPI 3.1 自动生成
- v0.3.2：KylixBoot 注解栈 + LLVM M2 完整 + stdlib Phase 6
- v0.1.5：stdlib `.klx` 声明文件 + 包管理器
- 所有 Go 测试通过（16 个包，LLVM 后端 274 测试）
- 教程 51/51 测试通过（Go 后端，`examples/complete-tutorial/`）
- LLVM 后端 49→51 教程编译通过（100%，含 example56_variant Variant 数组 + example57_variant_map Variant map 双端 parity；01-04 章节 19 个文件与 Go 后端输出逐字节一致；example33 多文件模块经 `multifile.go` MergePrograms 合并声明后通过）
- v0.5.6 新增：28 个 LLVM 后端 codegen bug + generator.klx 移植缺口修复，bootstrap self-host 51/51（100%）。详见 CHANGELOG.md
- v0.5.3 新增：自举编译器 round-trip 打通 + 自繁殖（`src/generator.klx` 三修：`Args` builtin + 条件导入 + `WriteEscapedGoString` 转义保护）
- v0.5.2 新增：自举编译器构建打通（多态基类 opt-in interface codegen）+ `Args` builtin + 切片协变修复
- v0.5.1 新增：map[String]Variant 真实化 + Variant 算术
- v0.5.0 新增：Variant 运行时（boxed `{tag, payload}`）
- v0.4.9 新增：类方法/lambda DISubprogram + DILexicalBlock + jsonutil `JsonGetArray`
- v0.4.8 新增：泛型类方法 codegen + DIBasicType 多类型
- v0.4.7 新增：静态数组真实 LowerBound + jsonutil `JsonGetMap` 递归嵌套对象
- v0.4.6 新增：DWARF 逐行调试（per-instruction DILocation + DILocalVariable + `#dbg_declare`）
- v0.5.5 新增：record 字段值拷贝（malloc+memcpy 深拷贝，匹配 Go 值语义）
- v0.5.4 新增：进程内 IR 优化 pass（DCE，默认运行）+ 增量编译缓存（llc 跳过，32x 加速）+ DWARF 调试符号（`kylix build --backend=llvm -g`）
- 所有源文件 ≤ 1000 行

## 关键文档

- [ROADMAP.md](ROADMAP.md) — 开发路线图（直到 v4.0）
- [TECHNICAL_DEBT.md](TECHNICAL_DEBT.md) — 已知问题与改进积压
- [TASKS.md](TASKS.md) — 详细任务分解
- [CHANGELOG.md](CHANGELOG.md) — 版本历史

## 架构

- `token/token.go` — Token 类型定义和关键字映射
- `lexer/lexer.go` — 词法分析器（字符 → token 流）
- `ast/ast.go` — AST 节点定义（接口 + 具体类型）
- `parser/parser.go` — Pratt 解析器核心；`parser_decl.go` 声明；`parser_stmt.go` 语句；`parser_expr.go` 表达式
- `generator/generator.go` — 生成器核心 + 预扫描；`generator_types.go` 类型/函数代码生成；`generator_stmt.go` 语句代码生成；`generator_expr.go` 表达式代码生成
- `generator/generator_boot_annotations.go` — KylixBoot 注解扫描 + 自动装配代码生成
- `generator/generator_validation_annotations.go` — 字段校验注解代码生成（`[Required]`/`[Email]` 等）
- `cmd/kylix/main.go` — CLI 入口（版本 0.3.3）
- `pkg/compiler/` — 编译 API + 增量缓存
- `pkg/compiler/annotations.go` — KylixBoot 注解诊断（KLX207–KLX214）
- `pkg/openapi/openapi.go` — OpenAPI 3.1 YAML 生成器
- `pkg/pkgmgr/` — 包管理器（add/install/remove）
- `pkg/repl/` — 交互式 REPL
- `pkg/lsp/` — Language Server Protocol
- `stdlib/` — Go 标准库封装（web, orm, template, exceptions, jwt 等）
- `stdlib/klx/` — LSP 补全用的 Kylix 声明文件
- `pkg/llvmgen/` — LLVM 后端代码生成器（原生二进制）
  - `codegen.go` — Generator 核心 + 字符串常量池 + 调试符号
  - `compile.go` — 编译管线（AST → IR → .o → binary）
  - `expr.go` — 表达式 codegen（算术/比较/调用/WriteLn）
  - `expr_access.go` — 成员/方法/接口/闭包访问 codegen
  - `stmt.go` — 语句 codegen（赋值/return/变量声明）
  - `stmt_flow.go` — 控制流 codegen（if/while/for/case/match/try）
  - `class.go` — 类/vtable/构造/方法 codegen
  - `stdlib_*.go` — 标准库模块 IR 实现（encoding/net/crypto/db/cache/jsonutil/boot/jwt/httpclient/sysutil/datetime）
  - `variant.go` — Variant 运行时（v0.5.0）：boxed `{i32 tag, i64 payload}` + box/unbox(as_double/as_str/as_int/as_bool)/compare/arith(add/sub/mul/div)/print helpers + nilbox 全局 + call-site 装箱/比较/算术/unbox
  - `debug.go` — DWARF 调试符号生成（`-g` flag）：per-instruction DILocation + DILocalVariable + `#dbg_declare`（v0.4.6 逐行调试）+ per-llvmType DIBasicType（v0.4.8 类型精度）
  - `passes.go` — IR 优化 pass 管线（DCE + ConstantFold）
  - `cache.go` — 增量编译缓存（SHA256 键控 .o 复用）
- `src/stdlib_ir.klx` — **v0.6.9**：host 生成的 stdlib IR 烘焙数据（13 段 + 139 签名，**AUTO-GENERATED，勿手改**；再生：`python3 scripts/extract_stdlib_ir.py > src/stdlib_ir.klx`，依赖 /tmp/stdir_cover/cover.ll + 教程 .ll）
- `scripts/extract_stdlib_ir.py` — stdlib IR 烘焙提取器（设计文档见文件头）
- `scripts/test_bootstrap_all.sh` — **v0.6.9**：51 教程 bootstrap-vs-host 全量回归（`--emit-llvm` → llc → clang → 运行 → 输出 diff；按 IR 扫描自动加 -lcrypto/-lsqlite3/-lcurl）

## 已完成阶段

### Phase 6–10 → v0.1.0–v0.1.5
- 字符串插值、异常类型、多返回值、属性
- Map 类型、Variant 类型、动态数组
- 枚举、切片、单元文件系统、多文件编译
- 自举验证完成（Self-hosted compiler）
- 接口验证、Kylix 层错误报告、真正的泛型（Go 1.18+）
- 增量编译（55× 加速）
- stdlib `.klx` 声明 + 包管理器

### v3.1.x → KylixBoot 框架 + LLVM M2 Phase 1
- `[Controller]`/`[Get]`/`[Post]` 路由自动装配
- `[Service]`/`[Component]`/`[Inject]` DI 自动装配
- `[Required]`/`[Email]`/`[Min]`/`[Max]`/`[MinLen]`/`[MaxLen]` 字段校验
- `[Authenticated]`/`[Role]` 路由安全守卫
- `[Entity]`/`[Column]`/`[PrimaryKey]`/`[Repository]`/`[Query]` ORM 注解
- 注解诊断 KLX207–KLX213

### v0.3.2 → LLVM M2 完整 + stdlib Phase 6
- LLVM 后端 M2：接口胖指针、成员/方法分发、泛型类单态化
- stdlib `net`（TCP/UDP/DNS）、`crypto`（SHA/AES/BCrypt）、`encoding`（Base64/Hex/CSV）
- 注解栈全部完成，教程 42/42

### v0.3.3 → KylixBoot 框架完善（2026-06-28）
- `[Body(TEntity)]`：POST/PUT 路由的 JSON 请求体自动绑定 + IsValid()/Validate() 校验
- `jwt` stdlib：JwtSign/JwtVerify/JwtSubject + BootRegisterJwtAuth 一键接入 `[Authenticated]`
- `kylix doc --openapi`：从注解自动生成 OpenAPI 3.1 YAML（路径、schema、安全方案）
- 错误码修正：ErrBodyBinding 从 KLX301（冲突）改为 KLX214
- 教程 45/45 通过（新增 14_body_binding、15_jwt、16_openapi）

## 下一步：v0.3.3 收尾

**已完成 ✅**
- 类型检查层 MVP：`pkg/compiler/typecheck.go`（862 行）完整实现
- 包管理器编译器集成：`CompileProject` 自动发现 `packages/*/` 并去重
- 测试覆盖提升：新增 `packages_test.go`，所有关键包已有测试

**剩余工作**
- CompileFile 单文件模式的跨单元依赖自动解析（可选，非阻塞）
- 文档更新：tutorial README 提及包管理器用法
- 性能优化：大型项目的增量编译缓存验证

**v4.0 规划**
- LLVM M3：完整类型系统 + 优化通道
- stdlib Phase 7：http client/server + 数据库连接池
- IDE 插件：VSCode/JetBrains 语法高亮 + 跳转

## 关键约束

- Go 后端保持不变（Kylix → Go → binary）
- AST 节点使用 class（不用 variant records）
- **未经用户明确许可，绝不 commit/push**
- **每个源文件不超过 1000 行**：大文件按功能拆分
- build=`go build -o /tmp/kylix_bin ./cmd/kylix/ && KYLIX=/tmp/kylix_bin bash examples/complete-tutorial/test_all.sh 2>&1 | tail -8`
- test=`go test $(go list ./... | grep -v '/examples') 2>&1 | grep -E "^ok|FAIL"`
- bootstrap（v0.6.9）= `go build -o /tmp/kylix_bin ./cmd/kylix/ && /tmp/kylix_bin build --backend=llvm -o /tmp/main_self_p2_dbg src/token.klx src/error.klx src/ast.klx src/lexer.klx src/parser.klx src/generator.klx src/stdlib_ir.klx src/llvmgen.klx src/main.klx`（**验证输出必须含 `✓ Built`**——管道 tail 会吞 llc 失败）
- bootstrap sweep（v0.6.9）= `bash scripts/test_bootstrap_all.sh 2>&1 | grep -E "^(FAIL|DIFF|SKIP|KNOWN)|==="`（期望 50 PASS + 1 SKIP）
- 不动点验证（v0.6.9）= gen1 `--emit-llvm` 9 文件 → fp.ll → `opt -passes=mem2reg` → `llc` → `clang -L/opt/homebrew/opt/openssl@3/lib -lcrypto -lsqlite3 -lcurl` → gen2 → gen2 `--emit-llvm` 同 9 文件 → cmp fp.ll（逐字节一致即不动点）

## 已知问题（v0.3.3）

详见 [TECHNICAL_DEBT.md](TECHNICAL_DEBT.md)。最优先修复的 3 项：
1. 包管理器未集成到编译器搜索路径（2.4）
2. `topoSortWithFiles` 文件路径对齐 bug（1.2）
3. `pkg/pkgmgr` + `pkg/compiler/cache` 零测试覆盖（3.1）

## 教程结构（examples/complete-tutorial/）

| 目录 | 示例数 | 状态 |
|------|--------|------|
| 01_basics ~ 11_modules | 32 | ✅ 全部通过 |
| 12_special_features | 7 | ✅ v0.3.2 |
| 13_stdlib_phase6 | 1 | ✅ v0.3.2 |
| 14_body_binding | 1 | ✅ v0.3.3 |
| 15_jwt | 1 | ✅ v0.3.3 |
| 16_openapi | 1 | ✅ v0.3.3 |
| 17_database | 1 | ✅ v4.0 |
| 18_cache | 1 | ✅ v4.0 |
| 19_http | 1 | ✅ v4.0 |
| 20_websocket | 1 | ✅ v4.0 |
| 21_variant | 2 | ✅ v5.0/v5.1 |
| **合计** | **50 文件** | **51/51 通过** |

## v0.5.7 里程碑：LLVM 后端 self-reproduction 不动点（2026-07-29）

- v0.5.7 已发布：LLVM 后端 bootstrap self-reproduction 达成。main_self（LLVM 编译的 bootstrap）编译 src/*.klx → self_gen.go → go build → main_self2（Go 编译的 bootstrap）→ 编译 51 教程全过 → main_self2 编译 src/*.klx → self_gen2.go → go build → main_self3 → 编译 51 教程全过 → **self_gen.go ≡ self_gen2.go（逐字节一致，真正不动点）**。修复 3 个 bug：(1) stdlib 启发式排除 Go builtins（append 被误映射为 stdlib.append）；(2) ClassTypes/UserFuncs 从 map 改为 String（Go 后端 nil map 写 panic）；(3) WriteEscapedGoString 用单反斜杠 '\' 替代双 '\\'(LLVM decodeKylixString 解码不一致)。回归 16 包 + 51 教程全绿。

## 后续开发规划（v0.5.9+）

- **v0.5.9 — 多态 gate + KylixBoot 注解自动装配** ✅（2026-08-08 发布）：#2 validation + #3 多态 gate + #4 KylixBoot autowire + #5 ORM 注解全部移植到 `src/generator.klx`，宿主与 bootstrap 行为收敛一致，不动点保持。
- **v0.6.0 — 性能 benchmark（#7）+ LLVM -O2 验证（#8）+ var 参数 host 支持** ✅（2026-08-10）：Result/BuildCache 统计字段 + `build --time` + `benchmarks/compile_time.sh` + `docs/compile-performance.md`。数据：Go 冷 29ms/热 23ms、LLVM -O0 11527ms、**-O2 575ms**。**#8**：51 教程 `--llvm-opt=2` 全过 + 输出与 -O0 逐字节一致。**var 参数**（host Go 后端）：指针传递 + 读写解引用 + 调用传址。剩余：#9 JetBrains 插件、var 参数 LLVM/klx 端同步（bootstrap 不用 var 参数，self-reproduction 不受影响）。
- **v0.6.1 — KylixRT 核心** ✅（2026-08-10）：`kylix run` 无 Go 单二进制（auto 探测）+ LLVM boot 注解自动装配 + httpclient/db 补缺 + **51 教程 LLVM 51/51** + bootstrap LLVM -O0/-O2 编译修复。详见 CHANGELOG。
- **v0.6.2 — 跨平台 + KylixRT 生产化** ✅（2026-08-11）：target triple 参数化 + CLI --target + 系统库链接 + **Linux LLVM 51/51** + `kylix doctor` + 平台 API 适配（sysutil/datetime/exc 真实，net/regex stub）+ sysutil 补齐（9 函数）+ **test/bench --backend=llvm**（无 Go 跑测试/基准）+ 裸过程调用/Args[N]/MergePrograms 3 个 LLVM bug。**剩余 limitation**：net Winsock / regex pcre2 / jwt / DbQueryRows / websocket（需专门工作）、Windows 真机验证（CI runner 装不上 LLVM）。
- **v0.6.3 — jwt 双端 + 分发 B + Variant 传参修复** ✅（2026-08-16 发布）：JwtSign/JwtVerify 真实现（HS256，签名与 Python 逐字节一致、验签 valid/wrong-secret/tampered/malformed 全对）+ `bundle_llvm.sh` 捆绑 LLVM（FindLLVM 可执行文件旁优先）+ **Variant 嵌套调用 segfault 修复**（`Has(JwtVerify(...))`：isVariantType 识别 `*ast.VariantType` + box IR 类型校正 ptr + 闭包/inherited/virtual call 三处同类 coerce 点全修）。16 包 + 51 教程（Go+LLVM）全绿。详见 CHANGELOG。
- **v0.6.4 — LLVM stdlib 真实现：DbQueryRows + websocket** ✅（2026-08-18 发布）：DbQueryRows（Variant map 标签 + variant-map 索引 + 推断路径修复 + 内联 sqlite3 行循环）；websocket 完整 5 函数（RFC 6455 握手 + 帧 + ping/pong，strcat 请求链规避 varargs spill 崩溃、SHA-1 用 OpenSSL）。16 包 + 51 教程（Go+LLVM）全绿。详见 CHANGELOG。
- **v0.6.5 — WS 自回环 + SHA-1 修复 + KylixRT 完善 + 性能优化** ✅（2026-08-20 发布）：手写 SHA-1 修复（websocket 去 OpenSSL）+ 分阶段握手 API（单进程自回环）+ 字符串插值溢出修复 + test/bench auto 回退 + 性能五项（bootstrap -O0 30× 提速）。16 包 + 51 教程（Go+LLVM）+ self-repro 不动点全绿。详见 CHANGELOG。
- **v0.6.6 — boot HTTP server + stdlib 补全 5 项** ✅（2026-08-21 发布）：boot server（`Boot<M>` 路由表 + `BootRun` 真体 + `req.Param/Query/Header/Body` 降级，KylixBoot 无 Go 真正可用）；stdlib 补全（jwt claims / cache TTL / httpclient JSON / UrlEncode / Variant div-mod）；顺带修复（Variant 赋值 as_str 误 coerce、b64url rem==1、hashtab 门控、Request enqueue DoRequest）。16 包 + 51 教程（Go+LLVM）+ self-repro 全绿。详见 CHANGELOG。
- **v0.6.7 — #9 JetBrains 插件 + 安装使用手册** ✅（2026-08-22 发布）：`jetbrains-plugin/` Gradle Kotlin 模块（IC 2024.3 SDK）：TextMate 语法高亮（复用 vscode-ext tmLanguage）+ LSP4IJ 桥 `kylix lsp`（补全/跳转/重命名/格式化）+ 25 个 Live Templates + `README.md` 安装使用手册；`./gradlew buildPlugin` 产出可安装 zip。**ROADMAP #9 ✅**。详见 CHANGELOG。
- **v0.6.8 — boot server 补强 + stdlib 补全 + JetBrains 插件完善** ✅（2026-08-23 发布）：boot server（POST body 读取 + `req.JSON` 绑定 + `BootRegisterJwtAuth` 真校验 + BootText Variant coerce + 401 reason）；stdlib（encoding Base64URL + httpclient JSON 嵌套对象 + JsonGetMap map-box 直通）；JetBrains 插件（`.klx` 图标 + Run 配置 + LSP undefined warning / SymbolTable 补全）。16 包 + 51 教程（Go+LLVM）+ boot 端到端 + buildPlugin 全绿。详见 CHANGELOG。
- **v0.6.9 — bootstrap 无 Go 闭环** ✅（2026-09-04 发布）：内存管理（arena 推广 ✅）、LLVM 端 Variant 嵌套链式索引 coerce 收尾 ✅、**bootstrap 无 Go 闭环达成**——stdlib IR 烘焙（`scripts/extract_stdlib_ir.py` → `src/stdlib_ir.klx`）+ emitter 补缺 20+ 项 + **gen2 诞生**（自举 IR 过 llc 链接成编译器）+ **IR 不动点**（gen1 ≡ gen2 ≡ gen3，~220k 行逐字节）+ 教程 sweep 50/51（example15 lambda / example50 jwt 于 P4.12 收尾修复）；O(n²) 性能疑虑实测排除。详见 CHANGELOG。
- **v0.7.0 — web 页面开发 + web 框架**：新增 **HTML 页面开发能力**（模板引擎 + 页面渲染 + 静态资源 + 表单/Cookie）+ 一套 **Kylix web 框架**（在 KylixBoot REST API 基础上补页面侧）+ **net Winsock / regex pcre2 真实现**（需 Windows 真机验证）。web 框架具体方案见文档（模板语法、res.HTML/Render、req.Form、static/ 约定、页面教程）。
- **1.0.0**：v0.6.9 + v0.7.0 全部完成后发布正式版 1.0.0。
- **后续**：跨平台（Linux/Windows/ARM64 CI 稳定）、自举 stdlib（.klx 编写→bootstrap 编译→自包含）。
