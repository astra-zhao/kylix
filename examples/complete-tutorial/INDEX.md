# Kylix Tutorial - Complete Index

## 📚 Documentation Files

- **README.md** — English tutorial with full feature coverage
- **README_CN.md** — Chinese tutorial (最新完整版)
- **QUICKSTART.md** — 5-minute getting started guide
- **SUMMARY.md** — Creation summary and metrics
- **INDEX.md** — This file
- **test_all.sh** — Automated test script

---

## 📂 Example Categories

### 01_basics/ (6 examples) ✅ 全部通过

| Example | Description | Status |
|---------|-------------|--------|
| example01_hello.klx | Hello World | ✅ |
| example02_variables.klx | Variable declarations | ✅ |
| example03_constants.klx | Constants | ✅ |
| example04_type_inference.klx | Type inference `:=` | ✅ |
| example05_operators.klx | Arithmetic, comparison, logical | ✅ |
| example06_comments.klx | Comments | ✅ |

### 02_control_flow/ (5 examples) ✅ 全部通过

| Example | Description | Status |
|---------|-------------|--------|
| example07_if_else.klx | If-then-else | ✅ |
| example08_while.klx | While loops | ✅ |
| example09_for_to.klx | For..to/downto | ✅ |
| example10_repeat.klx | Repeat-until | ✅ |
| example11_case.klx | Case statement | ✅ |

### 03_functions/ (4 examples) ✅ 全部通过

| Example | Description | Status |
|---------|-------------|--------|
| example13_functions.klx | Functions and procedures | ✅ |
| example14_recursion.klx | Recursive functions | ✅ |
| example15_lambda.klx | Anonymous procedures (lambda) | ✅ ⚠️ 仅过程 |
| example16_multireturn.klx | Multiple return values | ✅ |

### 04_oop/ (3 examples) ✅ 全部通过

| Example | Description | Status |
|---------|-------------|--------|
| example17_class_fields.klx | Class fields (with := inference) | ✅ |
| example18_class_methods.klx | Class methods (self.) | ✅ |
| example19_inheritance.klx | Inheritance | ✅ |

### 05_generics/ (1 example) ⚠️ 编译通过，运行时问题

| Example | Description | Status |
|---------|-------------|--------|
| example21_generic_class.klx | Generic stack class | ⚠️ |

### 06_advanced_types/ (5 examples) ✅ 全部通过

| Example | Description | Status |
|---------|-------------|--------|
| example20_enum.klx | Enum types | ✅ |
| example22_records.klx | Record types | ✅ |
| example23_arrays.klx | Fixed arrays | ✅ |
| example24_map.klx | Map type (hash tables) | ✅ |
| example25_string_ops.klx | String operations | ✅ |

### 07_stdlib_core/ (1 example) ✅ 全部通过

| Example | Description | Status |
|---------|-------------|--------|
| example29_basic_funcs.klx | Max, Min, Abs functions | ✅ |

### 10_exceptions/ (2 examples) ✅ 全部通过

| Example | Description | Status |
|---------|-------------|--------|
| example27_try_except.klx | Try-except blocks | ✅ |
| example28_finally.klx | Try-finally blocks | ✅ |

### 11_modules/ (2 examples) ⚠️ 编译通过，运行时问题

| Example | Description | Status |
|---------|-------------|--------|
| math_helper.klx | Unit definition | ✅ |
| example33_use_module.klx | Using units with `uses` | ⚠️ |

---

## 📊 Statistics

- **Total examples**: 29 (含 1 个 test.klx)
- **Fully working**: 27 (93.1%)
- **Compile but runtime issues**: 2 (泛型, 模块)
- **Categories**: 8
- **Feature coverage**: 35/74 language features (47.3%)

---

## ⚠️ Known Limitations (v3.0.0-alpha)

**Will be fixed in v3.1/v3.2:**

1. **String interpolation** — `${var}` doesn't expand
2. **Anonymous function return types** — `function(x): T` loses return type
3. **Class variable type** — `var p: TClass` generates `interface{}`
4. **Match statement** — generates invalid Go code
5. **Uses in program** — stdlib functions not directly callable

**Workarounds:**
- Classes: Use `var p := TClass.Create` (`:=` inference)
- Lambda: Only use anonymous procedures (no return value)
- Stdlib: Wait for v3.2 fix, or use in `unit` files

---

## 🎯 Next Steps

**For learners:**
1. Start with `01_basics/` — foundation
2. Progress to `02_control_flow/` and `03_functions/`
3. Explore `04_oop/` and `06_advanced_types/`
4. Advanced: `10_exceptions/` and `11_modules/`

**For contributors:**
- Add missing OOP examples (interfaces, properties)
- Add stdlib examples (strutil, mathutil, sysutil, jsonutil, datetime, regex)
- Add Web server example
- Add WASI example
- Add LLVM backend example

See [ROADMAP.md](../../ROADMAP.md) and [TASKS.md](../../TASKS.md) for planned work.

---

**Last updated**: 2026-06-22  
**Tutorial version**: v2.0 (29 examples, 27 working)
