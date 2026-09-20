package stdlib

// entitymeta.go — runtime registry for [Entity]-annotation metadata (v0.11.0).
//
// The compiler scans [Entity]/[Column]/[PrimaryKey] (plus the CRUD annotations
// [Label]/[Searchable]/[Hidden]/[Nullable]/[ReadOnly] and the validation
// annotations) and emits a registration sequence at the top of main(). The
// pure-Kylix CRUD engine (apps/admin/lib/crud.klx) reads it back through
// EntityMetaOf / EntityFieldCount / EntityFieldAt / EntityNames and renders
// list/form pages generically — so a new business entity needs nothing but an
// annotated class.
//
// Both backends must observe identical behaviour: the LLVM backend implements
// the same surface as IR over two fixed-size global arrays
// (pkg/llvmgen/stdlib_entitymeta.go). To keep the IR side free of string
// building, codegen pre-joins the pieces and the runtime stores them verbatim:
//
//	RegisterEntity(table, "<pk>|<label>|<flags>")
//	RegisterEntityField(table, "<column>|<kind>|<label>|<flags>")
//	SetEntityNames("users,roles,login_logs")
//
// Labels may not contain '|' or a newline; pkg/compiler.CheckORMAnnotations
// rejects that at compile time, so no escaping is needed at runtime.

// entityMetaEntry is one registered entity: its metadata string plus the
// columns registered for it, in declaration order.
type entityMetaEntry struct {
	Table  string
	Meta   string
	Fields []string
}

var (
	entityMetaTables []*entityMetaEntry
	entityMetaIndex  = map[string]*entityMetaEntry{}
	entityMetaNames  string
)

// RegisterEntity registers (or replaces) one entity. Calling it again for the
// same table replaces the previous entry, so repeated runs in one process stay
// deterministic.
func RegisterEntity(table, meta string) {
	e := &entityMetaEntry{Table: table, Meta: meta}
	if _, ok := entityMetaIndex[table]; !ok {
		entityMetaTables = append(entityMetaTables, e)
	} else {
		for i, prev := range entityMetaTables {
			if prev.Table == table {
				entityMetaTables[i] = e
				break
			}
		}
	}
	entityMetaIndex[table] = e
}

// RegisterEntityField appends one column entry to an already-registered entity.
func RegisterEntityField(table, entry string) {
	e := entityMetaIndex[table]
	if e == nil {
		return
	}
	e.Fields = append(e.Fields, entry)
}

// SetEntityNames stores the comma-joined table list (registration order) that
// EntityNames returns.
func SetEntityNames(csv string) {
	entityMetaNames = csv
}

// EntityMetaOf returns "<pk>|<label>|<flags>" for a table, or "" when the table
// is not registered.
func EntityMetaOf(table string) string {
	e := entityMetaIndex[table]
	if e == nil {
		return ""
	}
	return e.Meta
}

// EntityFieldCount returns the number of registered columns, or 0 when the
// table is unknown.
func EntityFieldCount(table string) int64 {
	e := entityMetaIndex[table]
	if e == nil {
		return 0
	}
	return int64(len(e.Fields))
}

// EntityFieldAt returns "<column>|<kind>|<label>|<flags>" for the i-th column
// (declaration order), or "" when the table or index is out of range.
func EntityFieldAt(table string, i int64) string {
	e := entityMetaIndex[table]
	if e == nil || i < 0 || i >= int64(len(e.Fields)) {
		return ""
	}
	return e.Fields[i]
}

// EntityNames returns the registered table names joined by ",", in
// registration order. The CRUD engine uses it to build the sidebar and to
// reject unknown :entity path segments.
func EntityNames() string {
	return entityMetaNames
}
