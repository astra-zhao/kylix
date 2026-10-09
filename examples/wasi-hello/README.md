# WASI Hello World

Minimal Kylix program compiled to WASI with the Go toolchain (`--wasi`, `GOOS=wasip1`).

The LLVM pure-logic target is a different command: `kylix build --backend=llvm --target wasi/wasm32`. See `examples/wasi-logic/` and `docs/WASI.md`. That path does not import DOM.

## Build

```bash
kylix build --wasi main.klx
# → hello.wasm
```

## Run with Wasmtime

```bash
# Basic run
wasmtime hello.wasm

# With environment variable
wasmtime --env NAME=Kylix hello.wasm

# With command-line args
wasmtime hello.wasm -- arg1 arg2
```

## Run with Node.js (WASI API)

```js
const { WASI } = require('wasi')
const fs = require('fs')

const wasi = new WASI({ env: { NAME: 'Kylix' } })
const wasm = await WebAssembly.compile(fs.readFileSync('./hello.wasm'))
const instance = await WebAssembly.instantiate(wasm, wasi.getImportObject())
wasi.start(instance)
```

## Expected Output

```
Hello, Kylix!
Running under WASI.
Args: 0
```
