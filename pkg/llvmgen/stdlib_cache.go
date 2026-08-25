package llvmgen

import (
	"fmt"
	"kylix/ast"
)

// stdlib_cache.go — LLVM IR implementation for the `cache` stdlib module.
//
// TCache is a 24-byte heap handle:
//
//	{ ptr htab @0, ptr ttl @8, i64 defaultTtlMs @16 }
//
//   - htab  : main string→string table (the value is the string form of the
//     cached Variant, as in previous versions).
//   - ttl   : parallel string→ptr table mapping each key to a malloc'd
//     `{i64 expiresAt}` record. Only keys stored with a TTL get a record.
//   - defaultTtlMs : default TTL (ms) used by Put; <= 0 = no expiry.
//
// Entries are evicted lazily: every read checks the ttl record against
// time() and deletes expired entries on access; Sweep() walks all ttl keys
// (via @__kylix_htab_keys) and evicts everything already expired. The LRU
// capacity from the Go backend's stdlib/cache.go is not replicated (the hash
// table is unbounded), matching the pre-v0.6.6 behavior.
//
// v0.6.6 adds PutWithTTL / Get / Sweep and real TTL semantics.

const cacheHandleTypeName = "TCache"

// emitCacheCall dispatches a `cache.Func(args)` / bare `Func(args)` call.
func (g *Generator) emitCacheCall(funcName string, args []ast.Expression) (string, string, error) {
	switch funcName {
	case "NewCache":
		return g.emitCacheNewCacheCall(args)
	default:
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; cache.%s not implemented", r, funcName))
		return r, "i64", nil
	}
}

// emitCacheBody dispatches the deferred body emitter.
func (g *Generator) emitCacheBody(funcName string) {
	switch funcName {
	case "NewCache":
		// NewCache is fully inlined at the call site (handle construction);
		// no separate define.
	case "put":
		g.emitCachePutBody()
	case "expired":
		g.emitCacheExpiredBody()
	case "sweep":
		g.emitCacheSweepBody()
	case "now_ms":
		g.emitCacheNowMsBody()
	}
}

// emitCacheNowMsBody — i64 @__kylix_now_ms() = wall clock in milliseconds
// (gettimeofday; tv_sec*1000 + tv_usec/1000). cache TTL expiry needs
// millisecond granularity; time() is second-granularity. Idempotent.
func (g *Generator) emitCacheNowMsBody() {
	if g.nowMsEmitted {
		return
	}
	g.nowMsEmitted = true
	g.line("define i64 @__kylix_now_ms() {")
	g.line("entry:")
	tv := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 16)", tv))
	g.line(fmt.Sprintf("  call i32 @gettimeofday(ptr %s, ptr null)", tv))
	secPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 0", secPtr, tv))
	sec := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", sec, secPtr))
	usecPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 8", usecPtr, tv))
	usec := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", usec, usecPtr))
	secMs := g.tmp()
	g.line(fmt.Sprintf("  %s = mul i64 %s, 1000", secMs, sec))
	usecMs := g.tmp()
	g.line(fmt.Sprintf("  %s = udiv i64 %s, 1000", usecMs, usec))
	ms := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %s", ms, secMs, usecMs))
	g.line(fmt.Sprintf("  ret i64 %s", ms))
	g.line("}")
	g.line("")
}

// emitCacheMethodCall dispatches a TCache instance method.
func (g *Generator) emitCacheMethodCall(receiver string, method string, args []ast.Expression) (string, string, error) {
	switch method {
	case "Put":
		return g.emitCachePutCall(receiver, args)
	case "PutWithTTL":
		return g.emitCachePutWithTTLCall(receiver, args)
	case "Get":
		return g.emitCacheGetCall(receiver, args)
	case "GetString":
		return g.emitCacheGetStringCall(receiver, args)
	case "Has":
		return g.emitCacheHasCall(receiver, args)
	case "Delete":
		return g.emitCacheDeleteCall(receiver, args)
	case "Size":
		return g.emitCacheSizeCall(receiver, args)
	case "Clear":
		return g.emitCacheClearCall(receiver, args)
	case "Sweep":
		return g.emitCacheSweepCall(receiver, args)
	default:
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; TCache.%s not implemented", r, method))
		return r, "i64", nil
	}
}

// cacheHandleField emits a `ptr` load of the handle field at byte offset off.
func (g *Generator) cacheHandleField(receiver string, off int) string {
	p := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %d", p, receiver, off))
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", r, p))
	return r
}

// cacheValueString converts a cached value's IR register to its string form.
func (g *Generator) cacheValueString(v, t string) string {
	switch t {
	case "ptr":
		return v
	case variantT:
		// A Variant box is stored by its string representation (as_str).
		return g.emitVariantAsStr(v)
	case "i64":
		buf := g.tmp()
		g.line(fmt.Sprintf("  %s = alloca [24 x i8], align 1", buf))
		bufPtr := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr inbounds [24 x i8], ptr %s, i64 0, i64 0", bufPtr, buf))
		fmtStr := g.addString("%lld")
		fmtPtr := g.ptrTo(fmtStr, 5)
		g.line(fmt.Sprintf("  call i32 (ptr, i64, ptr, ...) @snprintf(ptr %s, i64 24, ptr %s, i64 %s)", bufPtr, fmtPtr, v))
		return bufPtr
	case "double":
		buf := g.tmp()
		g.line(fmt.Sprintf("  %s = alloca [32 x i8], align 1", buf))
		bufPtr := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr inbounds [32 x i8], ptr %s, i64 0, i64 0", bufPtr, buf))
		fmtStr := g.addString("%.15g")
		fmtPtr := g.ptrTo(fmtStr, 6)
		g.line(fmt.Sprintf("  call i32 (ptr, i64, ptr, ...) @snprintf(ptr %s, i64 32, ptr %s, double %s)", bufPtr, fmtPtr, v))
		return bufPtr
	case "i1":
		trueStr := g.addString("true")
		falseStr := g.addString("false")
		truePtr := g.ptrTo(trueStr, 5)
		falsePtr := g.ptrTo(falseStr, 6)
		sel := g.tmp()
		g.line(fmt.Sprintf("  %s = select i1 %s, ptr %s, ptr %s", sel, v, truePtr, falsePtr))
		return sel
	default:
		return v
	}
}

// ---- NewCache(cap, ttlMs) -> ptr (TCache handle, inlined) ----
func (g *Generator) emitCacheNewCacheCall(args []ast.Expression) (string, string, error) {
	if len(args) != 2 {
		return "", "", fmt.Errorf("cache.NewCache expects 2 arguments, got %d", len(args))
	}
	ttlReg, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	// cap is accepted but ignored (the htab is fixed-bucket, matching prior
	// behavior); still evaluate for side effects.
	if _, _, err := g.emitExpr(args[0]); err != nil {
		return "", "", err
	}
	g.needHashtab = true
	h := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 24)", h))
	htab := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_new()", htab))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", htab, h))
	ttl := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_new()", ttl))
	ttlField := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 8", ttlField, h))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", ttl, ttlField))
	ttlMsField := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 16", ttlMsField, h))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", ttlReg, ttlMsField))
	return h, cacheHandleTypeName, nil
}

// ---- c.Put(k, v) / c.PutWithTTL(k, v, ttlMs) -> void ----
func (g *Generator) emitCachePutCall(receiver string, args []ast.Expression) (string, string, error) {
	if len(args) != 2 {
		return "", "", fmt.Errorf("TCache.Put expects 2 arguments, got %d", len(args))
	}
	kReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	vReg, vt, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	vStr := g.cacheValueString(vReg, vt)
	g.needHashtab = true
	g.enqueueStdlib("cache", "put", "put", 0)
	// ttlMs = handle->defaultTtlMs
	ttlMsField := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 16", ttlMsField, receiver))
	ttlMs := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", ttlMs, ttlMsField))
	g.line(fmt.Sprintf("  call void @__kylix_cache_put(ptr %s, ptr %s, ptr %s, i64 %s)", receiver, kReg, vStr, ttlMs))
	return "0", "void", nil
}

func (g *Generator) emitCachePutWithTTLCall(receiver string, args []ast.Expression) (string, string, error) {
	if len(args) != 3 {
		return "", "", fmt.Errorf("TCache.PutWithTTL expects 3 arguments, got %d", len(args))
	}
	kReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	vReg, vt, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	ttlReg, _, err := g.emitExpr(args[2])
	if err != nil {
		return "", "", err
	}
	vStr := g.cacheValueString(vReg, vt)
	g.needHashtab = true
	g.enqueueStdlib("cache", "put", "put", 0)
	g.line(fmt.Sprintf("  call void @__kylix_cache_put(ptr %s, ptr %s, ptr %s, i64 %s)", receiver, kReg, vStr, ttlReg))
	return "0", "void", nil
}

// emitCachePutBody — shared TTL-aware store. TTL record (malloc'd {i64
// expiresAt}) is only created when ttlMs > 0.
func (g *Generator) emitCachePutBody() {
	g.enqueueStdlib("cache", "now_ms", "now_ms", 0)
	g.line("define void @__kylix_cache_put(ptr %h, ptr %k, ptr %v, i64 %ttlMs) {")
	g.line("entry:")
	htab := g.cacheHandleField("%h", 0)
	g.line(fmt.Sprintf("  call void @__kylix_htab_put(ptr %s, ptr %%k, ptr %%v)", htab))
	pos := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sgt i64 %%ttlMs, 0", pos))
	storeLbl := g.label()
	doneLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", pos, storeLbl, doneLbl))
	g.line(fmt.Sprintf("%s:", storeLbl))
	ttl := g.cacheHandleField("%h", 8)
	now := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_now_ms()", now))
	exp := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %%ttlMs", exp, now))
	rec := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 8)", rec))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", exp, rec))
	g.line(fmt.Sprintf("  call void @__kylix_htab_put(ptr %s, ptr %%k, ptr %s)", ttl, rec))
	g.line(fmt.Sprintf("  br label %%%s", doneLbl))
	g.line(fmt.Sprintf("%s:", doneLbl))
	g.line("  ret void")
	g.line("}")
	g.line("")
}

// ---- c.Get(k) -> Variant ----
func (g *Generator) emitCacheGetCall(receiver string, args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("TCache.Get expects 1 argument, got %d", len(args))
	}
	kReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.needHashtab = true
	g.needVariantRuntime = true
	g.enqueueStdlib("cache", "expired", "expired", 0)
	resSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", resSlot))
	expired := g.tmp()
	g.line(fmt.Sprintf("  %s = call i1 @__kylix_cache_expired(ptr %s, ptr %s)", expired, receiver, kReg))
	expiredLbl := g.label()
	valueLbl := g.label()
	doneLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", expired, expiredLbl, valueLbl))
	g.line(fmt.Sprintf("%s:", expiredLbl))
	g.cacheStoreNilbox(resSlot)
	g.line(fmt.Sprintf("  br label %%%s", doneLbl))
	g.line(fmt.Sprintf("%s:", valueLbl))
	htab := g.cacheHandleField(receiver, 0)
	v := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_get(ptr %s, ptr %s)", v, htab, kReg))
	isNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", isNull, v))
	nilLbl := g.label()
	boxLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isNull, nilLbl, boxLbl))
	g.line(fmt.Sprintf("%s:", nilLbl))
	g.cacheStoreNilbox(resSlot)
	g.line(fmt.Sprintf("  br label %%%s", doneLbl))
	g.line(fmt.Sprintf("%s:", boxLbl))
	boxed := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_variant_box_str(ptr %s)", boxed, v))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", boxed, resSlot))
	g.line(fmt.Sprintf("  br label %%%s", doneLbl))
	g.line(fmt.Sprintf("%s:", doneLbl))
	res := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", res, resSlot))
	return res, variantT, nil
}

// ---- c.GetString(k) -> String ----
func (g *Generator) emitCacheGetStringCall(receiver string, args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("TCache.GetString expects 1 argument, got %d", len(args))
	}
	kReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.needHashtab = true
	g.enqueueStdlib("cache", "expired", "expired", 0)
	resSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", resSlot))
	expired := g.tmp()
	g.line(fmt.Sprintf("  %s = call i1 @__kylix_cache_expired(ptr %s, ptr %s)", expired, receiver, kReg))
	expiredLbl := g.label()
	valueLbl := g.label()
	doneLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", expired, expiredLbl, valueLbl))
	g.line(fmt.Sprintf("%s:", expiredLbl))
	emptyStr := g.addString("")
	emptyPtr := g.ptrTo(emptyStr, 1)
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", emptyPtr, resSlot))
	g.line(fmt.Sprintf("  br label %%%s", doneLbl))
	g.line(fmt.Sprintf("%s:", valueLbl))
	htab := g.cacheHandleField(receiver, 0)
	v := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_get(ptr %s, ptr %s)", v, htab, kReg))
	gv := g.nullGuardString(v)
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", gv, resSlot))
	g.line(fmt.Sprintf("  br label %%%s", doneLbl))
	g.line(fmt.Sprintf("%s:", doneLbl))
	res := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", res, resSlot))
	return res, "ptr", nil
}

// ---- c.Has(k) -> Boolean ----
func (g *Generator) emitCacheHasCall(receiver string, args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("TCache.Has expects 1 argument, got %d", len(args))
	}
	kReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.needHashtab = true
	g.enqueueStdlib("cache", "expired", "expired", 0)
	resSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i1, align 1", resSlot))
	expired := g.tmp()
	g.line(fmt.Sprintf("  %s = call i1 @__kylix_cache_expired(ptr %s, ptr %s)", expired, receiver, kReg))
	expiredLbl := g.label()
	valueLbl := g.label()
	doneLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", expired, expiredLbl, valueLbl))
	g.line(fmt.Sprintf("%s:", expiredLbl))
	g.line(fmt.Sprintf("  store i1 false, ptr %s", resSlot))
	g.line(fmt.Sprintf("  br label %%%s", doneLbl))
	g.line(fmt.Sprintf("%s:", valueLbl))
	htab := g.cacheHandleField(receiver, 0)
	hv := g.tmp()
	g.line(fmt.Sprintf("  %s = call i1 @__kylix_htab_has(ptr %s, ptr %s)", hv, htab, kReg))
	g.line(fmt.Sprintf("  store i1 %s, ptr %s", hv, resSlot))
	g.line(fmt.Sprintf("  br label %%%s", doneLbl))
	g.line(fmt.Sprintf("%s:", doneLbl))
	res := g.tmp()
	g.line(fmt.Sprintf("  %s = load i1, ptr %s", res, resSlot))
	return res, "i1", nil
}

// ---- c.Delete(k) -> void ----
func (g *Generator) emitCacheDeleteCall(receiver string, args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("TCache.Delete expects 1 argument, got %d", len(args))
	}
	kReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.needHashtab = true
	htab := g.cacheHandleField(receiver, 0)
	ttl := g.cacheHandleField(receiver, 8)
	g.line(fmt.Sprintf("  call void @__kylix_htab_del(ptr %s, ptr %s)", htab, kReg))
	g.line(fmt.Sprintf("  call void @__kylix_htab_del(ptr %s, ptr %s)", ttl, kReg))
	return "0", "void", nil
}

// ---- c.Size() -> Integer ----
func (g *Generator) emitCacheSizeCall(receiver string, args []ast.Expression) (string, string, error) {
	if len(args) != 0 {
		return "", "", fmt.Errorf("TCache.Size expects 0 arguments, got %d", len(args))
	}
	g.needHashtab = true
	htab := g.cacheHandleField(receiver, 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_htab_size(ptr %s)", r, htab))
	return r, "i64", nil
}

// ---- c.Clear() -> void ----
func (g *Generator) emitCacheClearCall(receiver string, args []ast.Expression) (string, string, error) {
	if len(args) != 0 {
		return "", "", fmt.Errorf("TCache.Clear expects 0 arguments, got %d", len(args))
	}
	g.needHashtab = true
	htab := g.cacheHandleField(receiver, 0)
	ttl := g.cacheHandleField(receiver, 8)
	g.line(fmt.Sprintf("  call void @__kylix_htab_clear(ptr %s)", htab))
	g.line(fmt.Sprintf("  call void @__kylix_htab_clear(ptr %s)", ttl))
	return "0", "void", nil
}

// ---- c.Sweep() -> Integer ----
func (g *Generator) emitCacheSweepCall(receiver string, args []ast.Expression) (string, string, error) {
	if len(args) != 0 {
		return "", "", fmt.Errorf("TCache.Sweep expects 0 arguments, got %d", len(args))
	}
	g.needHashtab = true
	g.enqueueStdlib("cache", "sweep", "sweep", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_cache_sweep(ptr %s)", r, receiver))
	return r, "i64", nil
}

// emitCacheSweepBody — walk every TTL-tracked key (via htab_keys) and delete
// entries whose expiresAt has passed. Returns the number evicted.
func (g *Generator) emitCacheSweepBody() {
	g.enqueueStdlib("cache", "now_ms", "now_ms", 0)
	g.line("define i64 @__kylix_cache_sweep(ptr %h) {")
	g.line("entry:")
	ttl := g.cacheHandleField("%h", 8)
	keysAgg := g.tmp()
	g.line(fmt.Sprintf("  %s = call { ptr, i64 } @__kylix_htab_keys(ptr %s)", keysAgg, ttl))
	items := g.tmp()
	g.line(fmt.Sprintf("  %s = extractvalue { ptr, i64 } %s, 0", items, keysAgg))
	n := g.tmp()
	g.line(fmt.Sprintf("  %s = extractvalue { ptr, i64 } %s, 1", n, keysAgg))
	now := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_now_ms()", now))
	iSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", iSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", iSlot))
	rmSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", rmSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", rmSlot))
	condLbl := g.label()
	bodyLbl := g.label()
	exitLbl := g.label()
	g.line(fmt.Sprintf("  br label %%%s", condLbl))
	g.line(fmt.Sprintf("%s:", condLbl))
	curI := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curI, iSlot))
	done := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sge i64 %s, %s", done, curI, n))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", done, exitLbl, bodyLbl))
	g.line(fmt.Sprintf("%s:", bodyLbl))
	keyPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds ptr, ptr %s, i64 %s", keyPtr, items, curI))
	k := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", k, keyPtr))
	expPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_get(ptr %s, ptr %s)", expPtr, ttl, k))
	isExp := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", isExp, expPtr))
	chkLbl := g.label()
	skipLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isExp, skipLbl, chkLbl))
	g.line(fmt.Sprintf("%s:", chkLbl))
	expVal := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", expVal, expPtr))
	isOld := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, %s", isOld, expVal, now))
	delLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isOld, delLbl, skipLbl))
	g.line(fmt.Sprintf("%s:", delLbl))
	htab := g.cacheHandleField("%h", 0)
	g.line(fmt.Sprintf("  call void @__kylix_htab_del(ptr %s, ptr %s)", htab, k))
	g.line(fmt.Sprintf("  call void @__kylix_htab_del(ptr %s, ptr %s)", ttl, k))
	curRm := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curRm, rmSlot))
	rmNext := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", rmNext, curRm))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", rmNext, rmSlot))
	g.line(fmt.Sprintf("  br label %%%s", skipLbl))
	g.line(fmt.Sprintf("%s:", skipLbl))
	iNext := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", iNext, curI))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", iNext, iSlot))
	g.line(fmt.Sprintf("  br label %%%s", condLbl))
	g.line(fmt.Sprintf("%s:", exitLbl))
	rmOut := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", rmOut, rmSlot))
	g.line(fmt.Sprintf("  ret i64 %s", rmOut))
	g.line("}")
	g.line("")
}

// emitCacheExpiredBody — `i1 @__kylix_cache_expired(ptr %h, ptr %k)`. Returns
// true (and evicts the entry) if the key has a TTL record whose expiresAt has
// passed; false otherwise.
func (g *Generator) emitCacheExpiredBody() {
	g.enqueueStdlib("cache", "now_ms", "now_ms", 0)
	g.line("define i1 @__kylix_cache_expired(ptr %h, ptr %k) {")
	g.line("entry:")
	ttl := g.cacheHandleField("%h", 8)
	expPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_get(ptr %s, ptr %%k)", expPtr, ttl))
	hasExp := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", hasExp, expPtr))
	falseLbl := g.label()
	chkLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", hasExp, falseLbl, chkLbl))
	g.line(fmt.Sprintf("%s:", chkLbl))
	expVal := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", expVal, expPtr))
	now := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_now_ms()", now))
	isOld := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, %s", isOld, expVal, now))
	delLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isOld, delLbl, falseLbl))
	g.line(fmt.Sprintf("%s:", delLbl))
	htab := g.cacheHandleField("%h", 0)
	g.line(fmt.Sprintf("  call void @__kylix_htab_del(ptr %s, ptr %%k)", htab))
	g.line(fmt.Sprintf("  call void @__kylix_htab_del(ptr %s, ptr %%k)", ttl))
	g.line("  ret i1 true")
	g.line(fmt.Sprintf("%s:", falseLbl))
	g.line("  ret i1 false")
	g.line("}")
	g.line("")
}

// cacheStoreNilbox stores the nilbox address into a ptr result slot.
func (g *Generator) cacheStoreNilbox(slot string) {
	nb := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { i32, i64 }, ptr @__kylix_variant_nilbox, i32 0, i32 0", nb))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", nb, slot))
}
