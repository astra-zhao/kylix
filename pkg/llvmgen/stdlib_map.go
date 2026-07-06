package llvmgen

import (
	"fmt"

	"kylix/ast"
)

// stdlib_map.go — language-level map[K]V support for the LLVM backend.
//
// A map variable's alloca holds a ptr to an @__kylix_htab_* table (the same
// runtime used by the cache stdlib module). This file handles:
//   - Variable declaration: map[K]V → htab_new() init
//   - Index read  m[k]  → htab_get
//   - Index assign m[k] := v → htab_put
//
// Currently only map[String]Integer (and map[String]String) are supported —
// the hash table stores string→string, so Integer values are stringified on
// store and parsed back on load. For the tutorial examples (example24) which
// only read/write and pass values to WriteLn/IntToStr, this is sufficient.

// emitMapVarDecl declares a map[K]V local: alloca a ptr slot and initialize
// it with htab_new().
func (g *Generator) emitMapVarDecl(name string, mapT *ast.MapType) error {
	g.needHashtab = true
	allocaReg := g.freshVarReg(name, "_map")
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", allocaReg))
	// Initialize with htab_new()
	tbl := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_new()", tbl))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", tbl, allocaReg))
	g.locals[name] = allocaReg
	g.mapVars[name] = true
	return nil
}

// emitMapIndexGet handles m[key] for a map variable → htab_get.
// Returns the value as a ptr (String) — callers that need Integer must
// parse via atoll, but for WriteLn/concatenation a String is fine.
func (g *Generator) emitMapIndexGet(idx *ast.IndexExpression) (string, string, error) {
	leftIdent, ok := idx.Left.(*ast.Identifier)
	if !ok {
		return "", "", fmt.Errorf("map index target must be an identifier")
	}
	tblSlot, ok := g.locals[leftIdent.Value]
	if !ok {
		return "", "", fmt.Errorf("undefined map variable: %s", leftIdent.Value)
	}
	tbl := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", tbl, tblSlot))
	keyReg, _, err := g.emitExpr(idx.Index)
	if err != nil {
		return "", "", err
	}
	g.needHashtab = true
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_get(ptr %s, ptr %s)", r, tbl, keyReg))
	return r, "ptr", nil
}

// emitMapIndexPut handles m[key] := value for a map variable → htab_put.
// The value is coerced to a String ptr (Integer → IntToStr → ptr).
func (g *Generator) emitMapIndexPut(idx *ast.IndexExpression, valReg string, valType string) error {
	leftIdent, ok := idx.Left.(*ast.Identifier)
	if !ok {
		return fmt.Errorf("map index target must be an identifier")
	}
	tblSlot, ok := g.locals[leftIdent.Value]
	if !ok {
		return fmt.Errorf("undefined map variable: %s", leftIdent.Value)
	}
	tbl := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", tbl, tblSlot))
	keyReg, _, err := g.emitExpr(idx.Index)
	if err != nil {
		return err
	}
	// Coerce value to ptr (String). Integer values need IntToStr.
	vPtr := valReg
	if valType != "ptr" {
		vPtr = g.emitIntToStrReg(valReg)
	}
	g.needHashtab = true
	g.line(fmt.Sprintf("  call void @__kylix_htab_put(ptr %s, ptr %s, ptr %s)", tbl, keyReg, vPtr))
	return nil
}

// emitIntToStrReg emits an IntToStr conversion inline (snprintf "%lld" → ptr).
// Used by map put when the value is an Integer — the hash table stores
// string→string, so we stringify integers on store.
func (g *Generator) emitIntToStrReg(valReg string) string {
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 32)", buf))
	fmtStr := g.addString("%lld")
	fmtPtr := g.ptrTo(fmtStr, 5)
	g.line(fmt.Sprintf("  call i32 (ptr, i64, ptr, ...) @snprintf(ptr %s, i64 32, ptr %s, i64 %s)", buf, fmtPtr, valReg))
	return buf
}
