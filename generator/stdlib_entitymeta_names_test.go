package generator_test

// stdlib_entitymeta_names_test.go — entitymeta API surface guards (v0.11.0).
//
// The function list lives in internal/entitymetaapi, imported by both backends
// (generator/generator_stdlib.go and pkg/llvmgen/stdlib.go), so list drift is a
// compile error. What still needs a test is the *dispatch* side: the LLVM
// backend's emitEntityMetaCall switch is hand-written, and a name present in
// the list but absent from the switch would fall into the generic
// `add i64 0, 0` stub — silently wrong on the LLVM backend only.

import (
	"regexp"
	"testing"

	"kylix/internal/entitymetaapi"
)

// TestEntityMetaNames_Dispatchable: every entitymeta function must have a case
// in the LLVM backend's emitEntityMetaCall.
func TestEntityMetaNames_Dispatchable(t *testing.T) {
	src := readSrc(t, "../pkg/llvmgen/stdlib_entitymeta.go")

	caseRe := regexp.MustCompile(`case ((?:"[A-Za-z_][A-Za-z0-9_]*"(?:,\s*)?)+):`)
	handled := map[string]bool{}
	for _, m := range caseRe.FindAllStringSubmatch(src, -1) {
		for _, q := range identRe.FindAllStringSubmatch(m[1], -1) {
			handled[q[1]] = true
		}
	}

	for _, name := range entitymetaapi.EntityMetaFunctions {
		if !handled[name] {
			t.Errorf("entitymeta.%s is in internal/entitymetaapi but has no case in "+
				"pkg/llvmgen/stdlib_entitymeta.go's emitEntityMetaCall — the LLVM backend "+
				"would stub it out", name)
		}
	}
}

// TestEntityMetaNames_Sane: the list is sorted, duplicate-free, and the
// procedure/return-type tables agree with it.
func TestEntityMetaNames_Sane(t *testing.T) {
	seen := map[string]bool{}
	for i, name := range entitymetaapi.EntityMetaFunctions {
		if name == "" {
			t.Fatalf("empty name at index %d", i)
		}
		if seen[name] {
			t.Errorf("duplicate name %q", name)
		}
		seen[name] = true
		if i > 0 && entitymetaapi.EntityMetaFunctions[i-1] > name {
			t.Errorf("list is not sorted: %q comes after %q", name, entitymetaapi.EntityMetaFunctions[i-1])
		}
	}

	for _, name := range entitymetaapi.EntityMetaFunctions {
		_, isProc := entitymetaapi.EntityMetaProcedures[name]
		_, hasRet := entitymetaapi.EntityMetaReturnTypes[name]
		if isProc && hasRet {
			t.Errorf("%s is listed as both a procedure and a value-returning function", name)
		}
		if !isProc && !hasRet {
			t.Errorf("%s has no return type and is not marked as a procedure", name)
		}
	}

	// The generated wiring emits these three by constant — keep them in the list.
	for _, name := range []string{
		entitymetaapi.RegisterEntity,
		entitymetaapi.RegisterEntityField,
		entitymetaapi.SetEntityNames,
	} {
		if !seen[name] {
			t.Errorf("emitted name %q is missing from EntityMetaFunctions", name)
		}
		if !entitymetaapi.EntityMetaProcedures[name] {
			t.Errorf("%q must be marked as a procedure", name)
		}
	}
}
