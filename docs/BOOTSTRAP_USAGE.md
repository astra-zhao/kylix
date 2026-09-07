# Bootstrap Compiler (main_self) Usage

The bootstrap compiler is the Kylix compiler compiled **from its own Kylix
source** (`src/*.klx`, 9 files) through the LLVM backend — a pure native
binary with **no Go dependency at runtime** (v0.6.9 no-Go closed loop;
gen1 ≡ gen2 ≡ gen3 byte-identical IR fixed point).

## What's in the archive

- `main_self` — the bootstrap compiler (native binary)
- `USAGE.md` — this file

## Requirements

- Shared libraries: `libcrypto` (OpenSSL 3), `libsqlite3`, `libcurl` — the
  compiler itself links them for its stdlib IR baking (on Linux:
  `libssl-dev libsqlite3-dev libcurl4-openssl-dev`).

## Usage

The bootstrap compiler reads `.klx` files and writes the transpiled output
to **stdout**:

```bash
# Kylix → Go code (then build with the Go toolchain)
./main_self program.klx > program.go
go build -o program program.go

# Multi-file (unit + main):
./main_self myunit.klx main.klx > main.go

# Kylix → LLVM IR (v0.6.9): the no-Go path
./main_self --emit-llvm program.klx > program.ll
llc -filetype=obj program.ll -o program.o
clang program.o -o program
```

## Rebuilding the compiler from itself (fixed point)

```bash
# gen1: host-built compiler produces the bootstrap as LLVM IR
kylix build --backend=llvm -o main_self \
  src/token.klx src/error.klx src/ast.klx src/lexer.klx src/parser.klx \
  src/generator.klx src/stdlib_ir.klx src/llvmgen.klx src/main.klx

# gen2: the bootstrap compiles its own source to IR, link with no Go:
./main_self --emit-llvm src/token.klx ... src/main.klx > gen2.ll
opt -passes=mem2reg gen2.ll -o gen2_opt.ll
llc -filetype=obj gen2_opt.ll -o gen2.o
clang gen2.o -o main_self_gen2

# gen1 ≡ gen2: re-emitting with gen2 yields byte-identical IR (fixed point)
```

See `CHANGELOG.md` (v0.6.9) for the full story and
`scripts/test_bootstrap_all.sh` for the tutorial sweep harness.
