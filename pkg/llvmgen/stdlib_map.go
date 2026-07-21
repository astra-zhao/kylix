package llvmgen

import (
	"fmt"
	"strings"

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
	isVariant := isVariantTypeExpr(mapT.ValueType)
	if isVariant {
		// v5.1.0: map[String]Variant — the htab's value slots hold Variant
		// box pointers (not C strings). Reads return "variant", writes box.
		g.needVariantRuntime = true
	}
	allocaReg := g.freshVarReg(name, "_map")
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", allocaReg))
	// Initialize with htab_new()
	tbl := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_new()", tbl))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", tbl, allocaReg))
	g.locals[name] = allocaReg
	g.mapVars[name] = true
	if isVariant {
		g.variantMaps[name] = true
	}
	return nil
}

// emitMapIndexGet handles m[key] for a map variable → htab_get.
// For Variant-valued maps, uses htab_get_variant (returns a box ptr or the
// nil-box on miss) and returns the "variant" pseudo-type so downstream
// comparisons/printing dispatch on the tag. Otherwise returns the value as
// a ptr (String) — callers that need Integer must parse via atoll.
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
	if g.variantMaps[leftIdent.Value] {
		g.needVariantRuntime = true
		g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_get_variant(ptr %s, ptr %s)", r, tbl, keyReg))
		return r, variantT, nil
	}
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_get(ptr %s, ptr %s)", r, tbl, keyReg))
	mr, mt2 := g.mapReadResult(r, idx.Left)
	return mr, mt2, nil
}

// mapReadResult converts a raw htab_get result (ptr) to the map's value type
// when the value isn't a string (e.g. Integer/enum → atoll → i64). v5.4.0.
func (g *Generator) mapReadResult(rawReg string, left ast.Expression) (string, string) {
	vt := g.mapValueKylixType(left)
	if g.isIntegerLikeType(vt) {
		g.needAtoll = true
		conv := g.tmp()
		g.line(fmt.Sprintf("  %s = call i64 @atoll(ptr %s)", conv, rawReg))
		return conv, "i64"
	}
	return rawReg, "ptr"
}

// isIntegerLikeType reports whether a Kylix type name is an integer/enum type
// (so map values of that type should be atoll-converted on read). v5.4.0.
func (g *Generator) isIntegerLikeType(t string) bool {
	switch strings.ToLower(t) {
	case "integer", "int", "int64", "longint", "word", "cardinal", "byte":
		return true
	}
	if t == "" {
		return false
	}
	if _, isClass := g.classes[t]; isClass {
		return false
	}
	if g.records[t] {
		return false
	}
	if strings.HasPrefix(t, "T") {
		return true // enum (T-prefixed, not class/record)
	}
	return false
}

// mapValueKylixType resolves the value type name for a map index expression's
// target (local/param map via globalMapValueTypes, or class field map).
// v5.4.0.
func (g *Generator) mapValueKylixType(left ast.Expression) string {
	if left == nil {
		return ""
	}
	if ident, ok := left.(*ast.Identifier); ok {
		if vt, ok := g.globalMapValueTypes[ident.Value]; ok {
			return vt
		}
		return "String"
	}
	if member, ok := left.(*ast.MemberExpression); ok {
		kind, tn := g.receiverKind(member.Object)
		if kind == "class" {
			if info, ok := g.classes[tn]; ok {
				for _, f := range info.Fields {
					if f.Name == member.Member && f.MapType != nil {
						return typeExprName(f.MapType.ValueType)
					}
				}
			}
		}
	}
	return "String"
}

// emitMapFieldIndexGet handles obj.Field[key] for a map-typed class field → load
// the htab handle from the field slot, then htab_get. v5.4.0.
//
// Array-valued maps (map[K]array of T): the LLVM htab stores string→string, so
// it cannot represent slice values. The bootstrap only reads (never writes)
// these maps (e.g. TGenerator.ClassFields), so they're always empty — return a
// zero slice, the correct "absent" value, so Length(fields)=0 and indexing is
// skipped.
func (g *Generator) emitMapFieldIndexGet(typeName, objReg, fieldName string, mt *ast.MapType, idx *ast.IndexExpression) (string, string, error) {
	if mt != nil {
		if _, ok := mt.ValueType.(*ast.ArrayType); ok {
			// v5.4.0: return a zero slice SSA struct value (not an alloca ptr)
			// so downstream `store {ptr,i64,i64} %v` is well-typed.
			z := g.tmp()
			g.line(fmt.Sprintf("  %s = insertvalue { ptr, i64, i64 } undef, ptr null, 0", z))
			return z, "{ ptr, i64, i64 }", nil
		}
	}
	fieldAddr, _, err := g.emitFieldStore(typeName, objReg, fieldName)
	if err != nil {
		return "", "", err
	}
	tbl := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", tbl, fieldAddr))
	keyReg, _, err := g.emitExpr(idx.Index)
	if err != nil {
		return "", "", err
	}
	g.needHashtab = true
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_get(ptr %s, ptr %s)", r, tbl, keyReg))
	// v5.4.0: typed map field reads — convert to the value type (e.g. enum→i64).
	if mt != nil && g.isIntegerLikeType(typeExprName(mt.ValueType)) {
		g.needAtoll = true
		conv := g.tmp()
		g.line(fmt.Sprintf("  %s = call i64 @atoll(ptr %s)", conv, r))
		return conv, "i64", nil
	}
	return r, "ptr", nil
}

// emitMapIndexPut handles m[key] := value for a map variable → htab_put.
// For Variant-valued maps, boxes the RHS into a Variant (the value slot holds
// a box pointer). Otherwise coerces the value to a String ptr (Integer →
// IntToStr → ptr), since the hash table stores string→string.
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
	vPtr := valReg
	if g.variantMaps[leftIdent.Value] {
		// Box the RHS into a Variant; the slot stores the box pointer.
		vPtr = g.emitVariantBox(valReg, valType)
	} else if valType != "ptr" {
		// String map: coerce Integer values to a String ptr.
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
