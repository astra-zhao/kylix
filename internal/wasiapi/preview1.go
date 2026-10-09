// Package wasiapi is the single source of truth for the
// wasi_snapshot_preview1 import table the pure-logic wasm32 target binds.
//
// DOM, sockets, and poll are not in this table. Both pkg/wasi (Go
// //go:wasmimport, GOOS=wasip1) and the LLVM backend (wasm-import-module
// attributes) must follow Preview1; TestPreview1Imports_Dispatchable guards
// the Go source, and the LLVM emitter ranges over the same slice.
package wasiapi

// Snapshot is the WASM import module name. One module, no env/DOM imports.
const Snapshot = "wasi_snapshot_preview1"

// Import is one wasi_snapshot_preview1 function.
// Params and Results use LLVM IR types: "i32", "i64", or "ptr".
// Pointers are "ptr" here; wasm32 lowers them to i32 linear-memory offsets.
type Import struct {
	Name     string
	Params   []string
	Results  []string // empty → void
	Noreturn bool
}

// Preview1 is the pure-logic subset of wasi_snapshot_preview1.
// Order is the canonical import-table order.
var Preview1 = []Import{
	{Name: "fd_write", Params: []string{"i32", "ptr", "i32", "ptr"}, Results: []string{"i32"}},
	{Name: "fd_read", Params: []string{"i32", "ptr", "i32", "ptr"}, Results: []string{"i32"}},
	{Name: "fd_seek", Params: []string{"i32", "i64", "i32", "ptr"}, Results: []string{"i32"}},
	{Name: "fd_close", Params: []string{"i32"}, Results: []string{"i32"}},
	{Name: "path_open", Params: []string{"i32", "i32", "ptr", "i32", "i32", "i64", "i64", "i32", "ptr"}, Results: []string{"i32"}},
	{Name: "clock_time_get", Params: []string{"i32", "i64", "ptr"}, Results: []string{"i32"}},
	{Name: "random_get", Params: []string{"ptr", "i32"}, Results: []string{"i32"}},
	{Name: "args_sizes_get", Params: []string{"ptr", "ptr"}, Results: []string{"i32"}},
	{Name: "args_get", Params: []string{"ptr", "ptr"}, Results: []string{"i32"}},
	{Name: "environ_sizes_get", Params: []string{"ptr", "ptr"}, Results: []string{"i32"}},
	{Name: "environ_get", Params: []string{"ptr", "ptr"}, Results: []string{"i32"}},
	{Name: "proc_exit", Params: []string{"i32"}, Noreturn: true},
}

// Names returns the import names in table order.
func Names() []string {
	out := make([]string, len(Preview1))
	for i, im := range Preview1 {
		out[i] = im.Name
	}
	return out
}
