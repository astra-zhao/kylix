package llvmgen

import (
	"fmt"
	"kylix/ast"
)

// stdlib_cache.go — LLVM IR implementation for the `cache` stdlib module.
//
// TCache is an opaque ptr handle wrapping the internal hash-table runtime
// (see stdlib_hashtab.go). All cache methods lower to a single call into the
// appropriate @__kylix_htab_* helper — cache is essentially a typed facade
// over the generic string→string hash table.
//
//   NewCache(cap, ttl)  -> ptr (TCache)     htab_new()  (cap/ttl currently
//                                                 ignored — the underlying
//                                                 table is fixed-bucket)
//   c.Put(k, v)         -> void             htab_put
//   c.GetString(k)      -> ptr (String)     htab_get
//   c.Has(k)            -> i1               htab_has
//   c.Delete(k)         -> void             htab_del
//   c.Size()            -> i64              htab_size
//   c.Clear()           -> void             htab_clear
//
// TTL and LRU eviction from the Go backend's stdlib/cache.go are NOT
// replicated here (the hash table is unbounded and never expires entries).
// The tutorial examples only exercise basic put/get/has/delete/clear, which
// this implementation covers.

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
		g.emitCacheNewCacheBody()
	}
}

// emitCacheMethodCall dispatches a TCache instance method.
func (g *Generator) emitCacheMethodCall(receiver string, method string, args []ast.Expression) (string, string, error) {
	switch method {
	case "Put":
		return g.emitCachePutCall(receiver, args)
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
	default:
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; TCache.%s not implemented", r, method))
		return r, "i64", nil
	}
}

// ---- NewCache(cap, ttl) -> ptr ----
//
//	t = htab_new(); ret t  (cap/ttl args accepted but ignored)
func (g *Generator) emitCacheNewCacheCall(args []ast.Expression) (string, string, error) {
	if len(args) != 2 {
		return "", "", fmt.Errorf("cache.NewCache expects 2 arguments, got %d", len(args))
	}
	// Evaluate args for side effects (e.g. a function call passed as cap),
	// but the values themselves are unused — the underlying table is fixed.
	for _, a := range args {
		if _, _, err := g.emitExpr(a); err != nil {
			return "", "", err
		}
	}
	g.enqueueStdlib("cache", "NewCache", "NewCache", 0)
	// Also ensure the hash-table runtime is emitted (once, idempotent).
	g.needHashtab = true
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_new()", r))
	return r, cacheHandleTypeName, nil
}

func (g *Generator) emitCacheNewCacheBody() {
	// NewCache is fully inlined at the call site (htab_new), so there's no
	// separate @__kylix_cache_NewCache define to emit. This body emitter is a
	// no-op kept for dispatch symmetry with other stdlib modules.
}

// ---- c.Put(k, v) -> void ----
func (g *Generator) emitCachePutCall(receiver string, args []ast.Expression) (string, string, error) {
	if len(args) != 2 {
		return "", "", fmt.Errorf("TCache.Put expects 2 arguments, got %d", len(args))
	}
	kReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	vReg, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	g.needHashtab = true
	g.line(fmt.Sprintf("  call void @__kylix_htab_put(ptr %s, ptr %s, ptr %s)", receiver, kReg, vReg))
	return "0", "void", nil
}

// ---- c.GetString(k) -> ptr ----
func (g *Generator) emitCacheGetStringCall(receiver string, args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("TCache.GetString expects 1 argument, got %d", len(args))
	}
	kReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.needHashtab = true
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_get(ptr %s, ptr %s)", r, receiver, kReg))
	return r, "ptr", nil
}

// ---- c.Has(k) -> i1 ----
func (g *Generator) emitCacheHasCall(receiver string, args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("TCache.Has expects 1 argument, got %d", len(args))
	}
	kReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.needHashtab = true
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i1 @__kylix_htab_has(ptr %s, ptr %s)", r, receiver, kReg))
	return r, "i1", nil
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
	g.line(fmt.Sprintf("  call void @__kylix_htab_del(ptr %s, ptr %s)", receiver, kReg))
	return "0", "void", nil
}

// ---- c.Size() -> i64 ----
func (g *Generator) emitCacheSizeCall(receiver string, args []ast.Expression) (string, string, error) {
	if len(args) != 0 {
		return "", "", fmt.Errorf("TCache.Size expects 0 arguments, got %d", len(args))
	}
	g.needHashtab = true
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_htab_size(ptr %s)", r, receiver))
	return r, "i64", nil
}

// ---- c.Clear() -> void ----
func (g *Generator) emitCacheClearCall(receiver string, args []ast.Expression) (string, string, error) {
	if len(args) != 0 {
		return "", "", fmt.Errorf("TCache.Clear expects 0 arguments, got %d", len(args))
	}
	g.needHashtab = true
	g.line(fmt.Sprintf("  call void @__kylix_htab_clear(ptr %s)", receiver))
	return "0", "void", nil
}
