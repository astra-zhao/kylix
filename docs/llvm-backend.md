# LLVM Backend Guide

> **Status**: Production-ready (v4.0.0 M3)  
> **Tutorial Coverage**: 14/15 basic tutorials compile to native binary (93%)

The LLVM backend generates native machine code directly from Kylix source, bypassing the Go toolchain entirely. This enables standalone executables with no Go runtime dependency.

---

## Quick Start

### Prerequisites

Install LLVM toolchain (llc + clang):

```bash
# macOS
brew install llvm

# Ubuntu/Debian
sudo apt-get install llvm clang

# Arch Linux
sudo pacman -S llvm clang
```

Verify installation:
```bash
llc --version    # LLVM static compiler
clang --version  # C compiler/linker
```

### Compile with LLVM Backend

```bash
# Basic compilation
kylix build --backend=llvm program.klx

# With optimization
kylix build --backend=llvm --llvm-opt program.klx

# Run immediately
kylix run --backend=llvm program.klx
```

### Generated Files

```
program.klx → program.ll (LLVM IR) → program.o (object file) → program (executable)
```

Keep intermediate files for debugging:
```bash
kylix build --backend=llvm --keep-ir program.klx
# Produces: program.ll (human-readable LLVM IR)
```

---

## Supported Features (M3)

### ✅ Fully Supported

#### Exception Handling
- `try...except...finally` blocks
- `raise` statements (with and without argument)
- `on E: ExceptionType do` type-specific handlers
- Nested try blocks
- Bare `raise` (re-throw inside handler)

```pascal
try
  raise Exception.Create('error');
except
  on E: MyCustomError do
    WriteLn('Custom: ', E.Message);
  on E: Exception do
    WriteLn('Generic: ', E.Message);
finally
  WriteLn('Cleanup');
end;
```

**Implementation**: Global exception slot + setjmp/longjmp with type ID propagation. Runtime subtype checking via `@__kylix_is_subtype`.

#### Control Flow
- `break` / `continue` (all loop types)
- `case...of` (switch on integers)
- `match` (pattern matching with guards)
- `for...in` (foreach over strings/arrays)
- `if...then...else`
- `while...do`
- `for...to/downto`
- `repeat...until`

#### Data Types
- Primitives: Integer (i64), Real (double), Boolean (i1), String (ptr)
- Records (struct types)
- Arrays (heap-allocated with length prefix)
- Classes (heap-allocated with vtable)
- Interfaces (vtable + data pointer pair)
- Generics (TBox<Integer> → mangled struct types)

#### Type Conversions
Automatic insertion in assignments:
- `Boolean ↔ Integer` (zext / icmp ne 0)
- `Integer ↔ Real` (sitofp / fptosi)

#### Built-in Functions
- `WriteLn(...)` — 0, 1, or multiple arguments (buffer + puts)
- `Write(...)` — no newline variant
- `ReadLn()` — read from stdin
- `Length(s)` — string/array length
- `IntToStr(n)` — integer to string conversion

#### String Operations
- String literals
- String interpolation: `"Value: ${x}"`
- Concatenation (via `+` operator)

### ⚠️ Partial Support

#### Arrays
- **Array literals**: `[1, 2, 3]` — basic heap allocation
- **Array indexing**: `arr[i]` — read/write supported
- **Slicing**: `arr[lo..hi]` — returns base pointer (incomplete)

### ⚠️ Partial Support

#### Lambda/Closures (M4 ✅)
- **Block-bodied** lambdas fully supported: `var f := procedure(x: T) begin ... end;`
- **Captured variables** supported via environment struct (by-value snapshot)
- **Expression-bodied** lambda (`function(x): T -> expr`) — not supported (parser limitation)

#### Async/Await
- **Current**: `await expr` executes synchronously (await keyword ignored)
- **Status**: Async runtime requires coroutine infrastructure (long-term)

#### Advanced OOP (M4 ✅)
- **`inherited` keyword**: supported as a statement (`inherited;` and `inherited Method(args);`)
- **`inherited` as expression** (`result := inherited F(x)`): not supported (parser limitation)

---

## Known Limitations

### 1. Expression-bodied Lambda
**Limitation**: `var sq := function(x: Integer): Integer -> x * x;` is not parsed (the parser doesn't recognize `->` after a return type in `procedure`/`function` lambdas).

**Workaround**: Use block-bodied lambdas:
```pascal
var sq := function(x: Integer): Integer;
begin
  result := x * x;
end;
WriteLn(sq(5));  // Works
```

### 2. Multi-Return Tuple Destructuring (M4 ✅)
**Status**: Fully supported since v4.1.0.
```pascal
function DivMod(n, d: Integer): (Integer, Integer);
begin
  result := (n div d, n mod d);
end;

var q, r: Integer;
(q, r) := DivMod(17, 5);  // Works: q=3, r=2
```
### 3. Optimization Level
**Status**: ✅ Implemented in v4.1.0. `--llvm-opt=1/2/3` runs the standalone `opt` tool (IR-level: mem2reg, inlining, loop induction, DCE) before `llc -O<N>` (codegen-level). The `opt` step is where the biggest wins come from — Kylix's alloca/load/store style benefits greatly from mem2reg.

**Benchmark highlights** (see [docs/llvm-performance.md](llvm-performance.md)):
- `loop_sum` (100M iterations): **20× speedup** at O2 (loop induced to closed-form)
- `fib(35)` recursive: 1.7× speedup at O3
- 01–04 tutorials produce byte-identical output optimized vs unoptimized

**Usage**:
```bash
kylix build --backend=llvm --llvm-opt=2 program.klx   # recommended
```

---

## Performance

### Compilation Speed
- **LLVM backend**: ~100-200ms per file (lexer → parser → LLVM IR → llc → clang)
- **Go backend**: ~80-150ms per file (lexer → parser → Go codegen → go build)

LLVM backend is ~30% slower due to additional llc + clang steps, but produces standalone binaries.

### Runtime Performance (Preliminary)
Benchmarked on basic arithmetic (10M iterations):

| Backend | Time | Binary Size |
|---------|------|-------------|
| Go      | 0.45s | 2.1 MB |
| LLVM (unoptimized) | 1.2s | 16 KB |
| LLVM (--llvm-opt) | 0.6s | 16 KB |

**Note**: Optimized LLVM code approaches Go performance while producing **100x smaller binaries** (no Go runtime).

---

## Troubleshooting

### Error: "llc: command not found"
**Solution**: Install LLVM toolchain (see Prerequisites).

### Error: "llc failed: exit status 1"
**Cause**: Invalid LLVM IR generated (compiler bug).

**Debug**:
```bash
kylix build --backend=llvm --keep-ir program.klx
cat program.ll  # Inspect LLVM IR
llc program.ll  # Run llc manually to see detailed error
```

**Common Issues**:
1. **Type mismatch**: `store i64 %t0, ptr %v_x_bool` — storing wrong type. File a bug report.
2. **Undefined symbol**: `call i64 @undefined` — function not declared. Check for typos or missing imports.
3. **SSA dominance error**: `Instruction does not dominate all uses` — compiler bug in control flow. File a bug report.

### Error: "clang: undefined reference to `___something`"
**Cause**: Missing C runtime function.

**Solution**: Link with libc explicitly:
```bash
clang program.o -o program -lc
```

---

## Implementation Notes

### Exception Handling Strategy

**Route C** (chosen): Global slot + setjmp/longjmp

```c
// Global exception slot (in LLVM IR)
@__kylix_exc_obj = global ptr null
@__kylix_exc_typeid = global i32 0

// Try block
%jmp_buf = alloca [37 x i64]
%jmp_ret = call i32 @setjmp(ptr %jmp_buf)
if (%jmp_ret == 0) {
  // try body
} else {
  // except handler: check type ID, match on-clauses
}

// Raise
store ptr %exc_obj, ptr @__kylix_exc_obj
store i32 %type_id, ptr @__kylix_exc_typeid
call void @longjmp(ptr %jmp_buf, i32 1)
```

**Advantages**:
- No platform-specific unwinding ABI
- Portable across all LLVM targets
- Simple runtime (no libunwind dependency)

**Trade-offs**:
- `finally` block duplicated 3x (normal/exception/reraise paths)
- Slightly larger code size vs. table-driven unwinding

### Type System Mapping

| Kylix Type | LLVM Type | Notes |
|------------|-----------|-------|
| Integer | i64 | Signed 64-bit |
| Real | double | IEEE 754 double precision |
| Boolean | i1 | Single bit (extended to i64 for I/O) |
| String | ptr | Pointer to null-terminated char array |
| Record | %TRecordName = type { ... } | Named struct |
| Class | ptr | Pointer to heap struct with vtable |
| Array | ptr | Pointer to heap buffer (i64 length + elements) |
| Interface | { ptr, ptr } | Vtable pointer + data pointer |

---

## Roadmap

### M4 (v4.1.0) — Advanced Features ✅
- [x] Lambda/closure support (captured variables in environment struct)
- [x] Multi-return tuple destructuring (`(q, r) := func()`)
- [x] `inherited` keyword (parent method dispatch)
- [x] Optimization passes (`opt` + `llc -O<N>`, see [llvm-performance.md](llvm-performance.md))
- [x] OOP field access & method dispatch (vtable inheritance)
- **Result**: 01–04 tutorials (19 files) produce byte-identical output to the Go backend; 27/49 tutorials compile via LLVM.

#### Known limitations in v4.1.0
- Expression-bodied lambda (`function(x): T -> expr`) — parser doesn't recognize `->` after a return type; use block-bodied lambdas.
- `inherited` as an expression (`result := inherited F(x)`) — parser treats `inherited` as a statement only.
- stdlib-heavy tutorials (08, 13–20) still require the Go toolchain — LLVM stdlib support is v4.2.0.

### M5 (v5.0.0) — Go Independence
- [ ] Self-hosting: Kylix compiler written in Kylix
- [ ] Custom runtime KylixRT (GC + string/array/map)
- [ ] stdlib rewritten in pure Kylix (remove `stdlib/*.go`)
- **Goal**: Zero Go dependency, standalone binaries

---

## Contributing

### Adding LLVM Backend Tests

Tests live in `pkg/llvmgen/*_test.go`. Each test compiles a Kylix snippet and asserts on generated IR.

**Example**:
```go
func TestMyFeature(t *testing.T) {
    ir := generateExcIR(t, `program p;
    begin
      // Your Kylix code here
    end.`)
    
    if !strings.Contains(ir, "expected IR pattern") {
        t.Errorf("IR missing expected pattern\nFull IR:\n%s", ir)
    }
}
```

Run tests:
```bash
go test ./pkg/llvmgen/ -v
```

### Reporting Bugs

If LLVM backend produces invalid IR or crashes:

1. **Minimize the reproducer** (smallest Kylix program that triggers the bug)
2. **Attach generated IR** (`kylix build --backend=llvm --keep-ir`)
3. **Include llc error output**
4. **File issue on GitHub** with label `llvm-backend`

---

## FAQ

**Q: Can I ship binaries compiled with LLVM backend?**  
A: Yes, starting from v4.0.0 M3. They are standalone executables with no external dependencies (besides libc).

**Q: Why is the binary so small compared to Go?**  
A: No Go runtime. LLVM binary contains only your code + minimal C runtime. Go bunaries include GC, scheduler, goroutines (adds ~1-2 MB).

**Q: Will LLVM backend replace Go backend?**  
A: No. Both backends coexist long-term. Use Go backend for rapid development (better tooling), LLVM for production deployment (smaller, faster).

**Q: Can I mix Go and Kylix code with LLVM backend?**  
A: No. LLVM backend produces pure native code. To call Go code, use Go backend and CGo.

**Q: What about Windows?**  
A: LLVM backend works on Windows with clang installed. Use `clang-cl` or MinGW toolchain. Tested on Windows 10/11.

---

**Last Updated**: 2026-06-30 (v4.0.0 M3)  
**Maintainer**: Kylix Core Team  
**Feedback**: https://github.com/astra-zhao/kylix/issues
