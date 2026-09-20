package generator

import (
	"kylix/ast"
	"kylix/internal/entitymetaapi"
	"strings"
)

// generator_entitymeta.go — CRUD metadata emission (v0.11.0).
//
// The [Entity] scanner (generator_orm_annotations.go) already collects table,
// primary key and column names for the ORM codegen. This file turns the same
// annotations — plus the CRUD annotations [Label]/[Searchable]/[Hidden]/
// [Nullable]/[ReadOnly] and the validation annotations — into a registration
// sequence emitted at the top of main():
//
//	stdlib.RegisterEntity("users", "id", "Users", "")
//	stdlib.RegisterEntityField("users", "username", "text", "Username", "required,searchable")
//	...
//
// The pure-Kylix CRUD engine reads it back at runtime through the entitymeta
// module (stdlib/entitymeta.go on the Go side, pkg/llvmgen/stdlib_entitymeta.go
// on the LLVM side). Both backends emit the same sequence, so the engine sees
// identical metadata whichever backend compiled it.
//
// Emission is gated on the program actually declaring an [Entity] class: a
// program without one gets byte-identical output to before this feature, which
// is what keeps the bootstrap IR fixed point intact.

// entityMetaKind / entityMetaFieldFlags delegate to internal/entitymetaapi so
// the Go and LLVM emitters cannot drift on the encoding.
func entityMetaKind(fieldName, fieldType string) string {
	return entitymetaapi.Kind(fieldName, fieldType)
}

func entityMetaFieldFlags(attrs []*ast.Attribute, kind string) string {
	return entitymetaapi.FieldFlags(attrs, kind)
}

// hasEntityMeta reports whether any [Entity] class was scanned.
func (g *Generator) hasEntityMeta() bool {
	return len(g.ormEntitiesOrder) > 0
}

// emitEntityMetaWiring writes the registration calls at the current position
// (the top of main, next to emitBootAutoWiring). No-op when the program has no
// [Entity] classes — see the gating note at the top of this file.
func (g *Generator) emitEntityMetaWiring() {
	if !g.hasEntityMeta() {
		return
	}
	g.imports["kylix/stdlib"] = true
	g.writeLine("// --- entity metadata (v0.11.0 CRUD engine) ---")
	var names []string
	for _, className := range g.ormEntitiesOrder {
		entity := g.ormEntities[className]
		if entity == nil {
			continue
		}
		names = append(names, entity.Table)
		g.writeLine("stdlib." + entitymetaapi.RegisterEntity + "(" +
			goStringLiteral(entity.Table) + ", " +
			goStringLiteral(joinMeta(entity.PKColumn, entity.Label, entity.Flags)) + ")")
		for _, field := range entity.Fields {
			g.writeLine("stdlib." + entitymetaapi.RegisterEntityField + "(" +
				goStringLiteral(entity.Table) + ", " +
				goStringLiteral(joinMeta(field.Column, field.Kind, field.Label, field.Flags)) + ")")
		}
	}
	g.writeLine("stdlib." + entitymetaapi.SetEntityNames + "(" + goStringLiteral(strings.Join(names, ",")) + ")")
}

// joinMeta assembles a metadata string from its parts. Both backends call this
// so the separators are defined once.
func joinMeta(parts ...string) string {
	return strings.Join(parts, "|")
}

// goStringLiteral renders a Go string literal for the emitted registration
// calls. Labels are validated at compile time (pkg/compiler rejects quotes and
// control characters), so a straight quote-escape is enough here.
func goStringLiteral(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '"':
			b.WriteString("\\\"")
		case '\\':
			b.WriteString("\\\\")
		case '\n':
			b.WriteString("\\n")
		case '\r':
			b.WriteString("\\r")
		default:
			b.WriteByte(c)
		}
	}
	b.WriteByte('"')
	return b.String()
}
