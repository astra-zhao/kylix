# wasm32-unknown-wasi（纯逻辑）

v0.15 的 LLVM 目标。编译器 CLI 版本仍是 `0.14.0`。这条路径只做纯逻辑：导入表是 `wasi_snapshot_preview1` 的一个子集，**没有 DOM**。浏览器页面仍走 Go 后端的 `kylix build --wasm`（`GOOS=js`）。

## 两条 WASI 路径

| 命令 | 后端 | 产物 |
|---|---|---|
| `kylix build --wasi main.klx` | Go `GOOS=wasip1 GOARCH=wasm` | 带 Go runtime 的 `.wasm` |
| `kylix build --wasi --tinygo main.klx` | TinyGo | 较小的 `.wasm` |
| `kylix build --backend=llvm --target wasi/wasm32 main.klx` | LLVM，triple `wasm32-unknown-wasi` | 无 libc 的命令模块 |

`--backend=llvm --wasi`（不另给 target）等价于 `--target wasi/wasm32`。`--backend=llvm --wasm` 会拒绝：那是浏览器 DOM 目标，LLVM wasm 不导入 DOM。`--tinygo` 不能和 LLVM 一起用。`--gc=boehm` 在这个目标上拒绝（没有 libgc）。

别名：`wasm`、`wasm32`、`wasi`、`wasm/wasm32`、`wasi/wasm32`。Go 后端看到这些 target 会报错，并指向 `--backend=llvm` 或单独的 `--wasi`。

## 导入表

单一来源是 `internal/wasiapi.Preview1`。Go（`//go:wasmimport`，仅 `GOOS=wasip1`）和 LLVM（`wasm-import-module` / `wasm-import-name`）都按这张表发射。顺序固定，共 12 个：

`fd_write`、`fd_read`、`fd_seek`、`fd_close`、`path_open`、`clock_time_get`、`random_get`、`args_sizes_get`、`args_get`、`environ_sizes_get`、`environ_get`、`proc_exit`（noreturn）。

没有 socket、poll，也没有其余 preview1 系统调用。`@llvm.used` 把整张表留在链接后的模块里，方便核对导入节，而不只留下这个程序实际调用的那几个。

宿主（非 wasip1）的 `pkg/wasi` 仍是本机替身（`os` / `crypto/rand` / `time`），方便 `go test`。它不是导入表。

## LLVM 运行时

只在 `targetOS == wasi` 时发射，默认目标的 IR 不出现 `wasi_snapshot_preview1`、`@__kylix_wasi_` 或这条 triple（`TestWasi_HostIRUnchanged`）。

- 数据布局：`e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-n32:64-S128-ni:1:10:20`。
- 堆：8MiB bump（`@__kylix_wasi_heap`）。`@free` 是空操作，实例在 `proc_exit` 时丢掉。耗尽时向 fd 2 写一行并 `proc_exit(1)`。
- `@malloc` / `@calloc` / `@realloc` / 字符串函数 / `@printf` / `@snprintf` / `@puts` / `@exit` 是定义，不是宿主 libc 的 declare。`calloc` 用字节循环清零。wasm32 的 `llc -O0` 会把 `llvm.memset` 降成 `i32 @memset(i32, i32, i32)`，所以运行时自己提供这个符号，不链 wasi-libc。
- `@_start` 调 `main`，再 `proc_exit`。链接：`clang --target=wasm32-unknown-wasi -nostdlib -Wl,--no-entry -Wl,--export-all`，失败则回落 `wasm-ld`。不链 `-lcrypto` / `-lsqlite3` / `-lcurl` / `-lm`。默认输出名 `*.wasm`。
- `printf` 认 `%s`、`%d`/`%i`/`%u`、`%ld`/`%lld`、`%g`/`%f`/`%e`、`%%`。浮点是 6 位小数、去掉末尾的 0（`1.5` 印成 `1.5`），不是完整的 `%.17g`。
- wasm 没有 setjmp。`@setjmp` / `@_setjmp` 返回 0；`@longjmp` / `@_longjmp` 执行 `proc_exit(70)`。`try`/`except` 捕获不到，`raise` 直接退出。

`uses wasi` 在 LLVM 上只实现：`Stdout`、`Stderr`、`Getenv`、`ClockMonotonic`、`ClockWalltime`、`WasiExit`。`Args` / `Environ` / `Stdin` / `ReadFile` / `WriteFile` 是编译错误，实现在 `pkg/wasi` 的 `GOOS=wasip1` 里。

## 文件与环境（Go wasip1）

`ReadFile` / `WriteFile` 的路径相对第一个预打开目录（fd 3）。wasmtime 用 `--dir` 提供它。绝对路径不可用。`WriteFile` 是 `path_open` 的 `CREAT|TRUNC`。`Getenv` 读 `environ_get`。

## 验收

```bash
go build -o kylix ./cmd/kylix/
KYLIX=$PWD/kylix bash examples/wasi-logic/check.sh
```

脚本做两件事：

1. LLVM 编译 `examples/wasi-logic/main.klx`，确认 wasm 魔数和导入名，`wasmtime` 的 stdout 逐字是 `sum=42`、`hello wasi`、`42`、`1.5`。
2. `GOOS=wasip1 GOARCH=wasm` 编译 `examples/wasi-preview1`，在 `wasmtime --dir … --env NAME=Kylix` 下检查时钟、`random_get`、环境变量和预打开目录的读写。

CI job：`wasi-wasm32`（ubuntu-latest，apt 的 llvm/clang，wasmtime 27.0.0）。没有 wasmtime 或 llc 时，`TestWasi_HelloRoundTrip` 会跳过，所以普通 `test` job 不依赖它们。

旧示例 `examples/wasi-hello` 仍是 Go `--wasi` 路径。
