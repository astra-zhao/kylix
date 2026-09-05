# Kylix Tutorial - Complete Index

## 📚 Documentation Files

- **README.md** — English tutorial with full feature coverage
- **README_CN.md** — 中文完整版教程
- **QUICKSTART.md** — 5-minute getting started guide
- **SUMMARY.md** — Historical creation summary (see its header note)
- **INDEX.md** — This file
- **test_all.sh** — Go backend automated sweep (51/51)
- **test_all_llvm.sh** — LLVM backend automated sweep (51/51)

---

## 📂 Example Categories

### 01_basics/ (6 examples) ✅

| Example | Description | Status |
|---------|-------------|--------|
| example01_hello.klx | Hello World | ✅ |
| example02_variables.klx | Variable declarations | ✅ |
| example03_constants.klx | Constants | ✅ |
| example04_type_inference.klx | Type inference `:=` | ✅ |
| example05_operators.klx | Arithmetic, comparison, logical | ✅ |
| example06_comments.klx | Comments | ✅ |

### 02_control_flow/ (5 examples) ✅

| Example | Description | Status |
|---------|-------------|--------|
| example07_if_else.klx | If-then-else | ✅ |
| example08_while.klx | While loops | ✅ |
| example09_for_to.klx | For..to/downto | ✅ |
| example10_repeat.klx | Repeat-until | ✅ |
| example11_case.klx | Case statement | ✅ |

### 03_functions/ (4 examples) ✅

| Example | Description | Status |
|---------|-------------|--------|
| example13_functions.klx | Functions and procedures | ✅ |
| example14_recursion.klx | Recursive functions | ✅ |
| example15_lambda.klx | Lambdas and closures | ✅ |
| example16_multireturn.klx | Multiple return values | ✅ |

### 04_oop/ (4 examples) ✅

| Example | Description | Status |
|---------|-------------|--------|
| example17_class_fields.klx | Class fields | ✅ |
| example18_class_methods.klx | Class methods (self.) | ✅ |
| example19_inheritance.klx | Inheritance | ✅ |
| example40_declarative_oop.klx | Declarative OOP (`var p := TPerson.Create`) | ✅ |

### 05_generics/ (1 example) ✅

| Example | Description | Status |
|---------|-------------|--------|
| example21_generic_class.klx | Generic stack class | ✅ |

### 06_advanced_types/ (5 examples) ✅

| Example | Description | Status |
|---------|-------------|--------|
| example20_enum.klx | Enum types | ✅ |
| example22_records.klx | Record types | ✅ |
| example23_arrays.klx | Fixed and dynamic arrays | ✅ |
| example24_map.klx | Map type (hash tables) | ✅ |
| example25_string_ops.klx | String operations | ✅ |

### 07_stdlib_core/ (1 example) ✅

| Example | Description | Status |
|---------|-------------|--------|
| example29_basic_funcs.klx | Max, Min, Abs functions | ✅ |

### 08_stdlib_utils/ (4 examples) ✅

| Example | Description | Status |
|---------|-------------|--------|
| example36_sysutil.klx | File system / environment utilities | ✅ |
| example37_jsonutil.klx | JSON encode/decode | ✅ |
| example38_datetime.klx | Date and time | ✅ |
| example39_regex.klx | Regular expressions | ✅ |

### 10_exceptions/ (2 examples) ✅

| Example | Description | Status |
|---------|-------------|--------|
| example27_try_except.klx | Try-except blocks | ✅ |
| example28_finally.klx | Try-finally blocks | ✅ |

### 11_modules/ (1 example + unit) ✅

| Example | Description | Status |
|---------|-------------|--------|
| math_helper.klx | Unit definition | ✅ |
| example33_use_module.klx | Using units with `uses` | ✅ (multi-file; SKIP in no-Go bootstrap sweep) |

### 12_special_features/ (7 examples) ✅

| Example | Description | Status |
|---------|-------------|--------|
| example41_attributes.klx | `[Attribute]` annotation syntax | ✅ |
| example42_kylixboot_autowire.klx | KylixBoot `[Controller]` + `[Get]` auto routing | ✅ |
| example43_kylixboot_di.klx | KylixBoot `[Service]` + `[Inject]` DI | ✅ |
| example44_kylixboot_proc_handler.klx | Procedure-style route handler | ✅ |
| example45_validation_annotations.klx | `[Required]`/`[Email]`/`[Min]`/`[MinLen]` validation | ✅ |
| example46_security_annotations.klx | `[Authenticated]`/`[Role]` security guards | ✅ |
| example47_orm_annotations.klx | `[Entity]`/`[Repository]`/`[Query]` ORM | ✅ |

### 13_stdlib_phase6/ (1 example) ✅

| Example | Description | Status |
|---------|-------------|--------|
| example48_phase6_net_crypto_encoding.klx | SHA-256, Base64, BCrypt, CSV, HMAC, MD5 | ✅ |

### 14_body_binding/ (1 example) ✅

| Example | Description | Status |
|---------|-------------|--------|
| example49_body_binding.klx | `[Body(TEntity)]` JSON body binding + validation | ✅ |

### 15_jwt/ (1 example) ✅

| Example | Description | Status |
|---------|-------------|--------|
| example50_jwt_auth.klx | `JwtSign`/`JwtVerify` + `BootRegisterJwtAuth` | ✅ |

### 16_openapi/ (1 example) ✅

| Example | Description | Status |
|---------|-------------|--------|
| example51_openapi.klx | Annotations → OpenAPI 3.1 YAML (`kylix doc --openapi`) | ✅ |

### 17_database/ (1 example) ✅

| Example | Description | Status |
|---------|-------------|--------|
| example52_database.klx | SQLite: `DbOpenSQLite`/`DbExec`/`DbQueryScalar`/`DbQueryRows` | ✅ |

### 18_cache/ (1 example) ✅

| Example | Description | Status |
|---------|-------------|--------|
| example53_cache.klx | LRU cache (`NewCache`/`Put`/`Get`/`Has`/`Delete`/TTL) | ✅ |

### 19_http/ (1 example) ✅

| Example | Description | Status |
|---------|-------------|--------|
| example54_http.klx | `THttpClient` GET/POST/PUT/DELETE + one-shot helpers | ✅ |

### 20_websocket/ (1 example) ✅

| Example | Description | Status |
|---------|-------------|--------|
| example55_websocket.klx | RFC 6455 WebSocket client/server | ✅ |

### 21_variant/ (2 examples) ✅

| Example | Description | Status |
|---------|-------------|--------|
| example56_variant.klx | Variant scalars and arrays (type-tagged) | ✅ |
| example57_variant_map.klx | `map[String]Variant` type-tagged map | ✅ |

---

## 📊 Statistics

- **Numbered examples**: 50 across 20 chapters (plus `math_helper.klx` unit and `test.klx` smoke file)
- **Go backend sweep**: 51/51 (`test_all.sh`)
- **LLVM backend sweep**: 51/51 (`test_all_llvm.sh`, native binaries)
- **No-Go bootstrap sweep**: 50 PASS + 1 SKIP (`scripts/test_bootstrap_all.sh`; example33 verified host-side)
- **Compiler self-hosting**: IR fixed point reached (gen1 ≡ gen2 ≡ gen3)

---

## 🎯 Learning Path

**For learners:**
1. Start with `01_basics/` — foundation
2. Progress to `02_control_flow/` and `03_functions/`
3. Explore `04_oop/` and `06_advanced_types/`
4. Advanced: `10_exceptions/`, `11_modules/`, `12_special_features/`
5. stdlib in practice: chapters 13-20 (crypto / JWT / OpenAPI / database / cache / HTTP / WebSocket)
6. Dynamic typing: `21_variant/`

**For contributors:**
- Add interface / property examples
- Add generic functions / constraints examples
- Add a `kylix test` workflow example
- See [ROADMAP.md](../../ROADMAP.md) and [TASKS.md](../../TASKS.md) for planned work.

---

**Last updated**: 2026-09-04  
**Tutorial version**: v0.6.9 (50 numbered examples, all passing)
