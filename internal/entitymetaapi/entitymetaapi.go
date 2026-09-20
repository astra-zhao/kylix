// Package entitymetaapi is the single source of truth for the entity-metadata
// (CRUD engine) API surface introduced in v0.11.0.
//
// The Go host's stdlib heuristic (generator/generator_stdlib.go) and the LLVM
// backend's module list (pkg/llvmgen/stdlib.go) both import this package, so a
// function added here is resolvable on every backend at compile time and the
// two lists cannot drift — the same arrangement internal/bootapi uses for the
// KylixBoot surface (v0.7.2, after the `undefined: BootNotFoundPage` bug).
//
// The dispatch side (pkg/llvmgen/stdlib_entitymeta.go) stays hand-written and
// is guarded by TestEntityMetaNames_Dispatchable in generator.
package entitymetaapi

// EntityMetaFunctions is the complete entitymeta module surface, sorted
// alphabetically. Keep it sorted — diff readability on review.
var EntityMetaFunctions = []string{
	"EntityFieldAt",
	"EntityFieldCount",
	"EntityMetaOf",
	"EntityNames",
	"RegisterEntity",
	"RegisterEntityField",
	"SetEntityNames",
}

// The registration procedures, referenced by name so a rename surfaces as a
// compile error in both emitters.
const (
	RegisterEntity      = "RegisterEntity"
	RegisterEntityField = "RegisterEntityField"
	SetEntityNames      = "SetEntityNames"
)

// EntityMetaProcedures are the subset that return no value. The host generator
// emits a bare call statement for these instead of an assignment.
var EntityMetaProcedures = map[string]bool{
	RegisterEntity:      true,
	RegisterEntityField: true,
	SetEntityNames:      true,
}

// FunctionSet returns EntityMetaFunctions as a set, for the backends' module
// function tables.
func FunctionSet() map[string]bool {
	s := make(map[string]bool, len(EntityMetaFunctions))
	for _, n := range EntityMetaFunctions {
		s[n] = true
	}
	return s
}

// EntityMetaReturnTypes maps the value-returning functions to their concrete
// Go return type.
var EntityMetaReturnTypes = map[string]string{
	"EntityMetaOf":     "string",
	"EntityFieldCount": "int64",
	"EntityFieldAt":    "string",
	"EntityNames":      "string",
}
