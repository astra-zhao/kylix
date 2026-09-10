// Package bootapi is the single source of truth for the KylixBoot API
// surface (v0.7.2). The Boot* function list was previously hand-maintained
// in two places — the Go host's stdlib heuristic (generator/generator_stdlib.go,
// which decides which calls resolve to the stdlib Go package) and the LLVM
// backend's module list (pkg/llvmgen/stdlib.go) — and the lists drifted
// apart once already: the v0.7.0 release-week `undefined: BootNotFoundPage`
// bug, where one backend compiled a program the other rejected.
//
// Both backends now import this package, so the lists cannot drift: adding
// a function here makes it resolvable on every backend at compile time.
// The only hand-maintained remnant is the dispatch side — emitBootCall
// cases and bootStubReturnTypes in pkg/llvmgen/stdlib_boot.go — guarded by
// TestBootNames_Dispatchable (generator/stdlib_boot_names_test.go).
package bootapi

// BootFunctions is the complete KylixBoot API surface, sorted
// alphabetically. Keep it sorted — diff readability on review.
var BootFunctions = []string{
	"BootConfigGetInt",
	"BootConfigGetString",
	"BootConfigSet",
	"BootDELETE",
	"BootEnforceAuth",
	"BootEnforceRole",
	"BootErrorPage",
	"BootGET",
	"BootHTML",
	"BootJSON",
	"BootNotFoundPage",
	"BootPOST",
	"BootPUT",
	"BootReadJSON",
	"BootRegisterAuth",
	"BootRegisterInstance",
	"BootRegisterJwtAuth",
	"BootRegisterRoles",
	"BootResolve",
	"BootRun",
	"BootStatic",
	"BootText",
	"BootUseCORS",
	"BootUseLogger",
	"BootUseRecover",
	"BootUseRequestID",
}

// BootRegisterJwtAuth is also exported by the jwt module's function list on
// both backends; reference this constant there so the cross-listing cannot
// drift either.
const BootRegisterJwtAuth = "BootRegisterJwtAuth"
