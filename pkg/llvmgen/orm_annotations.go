package llvmgen

import (
	"fmt"
	"strings"

	"kylix/ast"
	"kylix/internal/entitymetaapi"
)

// orm_annotations.go — [Entity] metadata scanning for the LLVM backend
// (v0.11.0).
//
// The Go backend has generated ORM methods (ToRow/FromRow/FindAll/…) from
// [Entity] annotations since v0.3.x; the LLVM backend has never implemented
// them (method calls on those names fall back to stubs in class.go). What the
// LLVM backend needs for the CRUD engine is not the methods but the *metadata*:
// table, primary key, and the per-column kind/label/flags.
//
// This scanner collects exactly that and emitEntityMetaWiring() writes the
// registration sequence at the top of main(), mirroring
// generator/generator_entitymeta.go so both backends hand the pure-Kylix CRUD
// engine identical strings. The encoding itself lives in
// internal/entitymetaapi — shared, so it cannot drift.

// llvmEntityField is one column of an [Entity] class.
type llvmEntityField struct {
	Entry string // "<column>|<kind>|<label>|<flags>"
}

// llvmEntity is one [Entity]-annotated class.
type llvmEntity struct {
	Table  string
	Meta   string // "<pk>|<label>|<flags>"
	Fields []llvmEntityField
}

// entityMetaMaxTables / entityMetaMaxFields bound the emitted registrations.
// The Go backend has no such limit, so exceeding it here would silently give
// the two backends different metadata — reject at compile time instead.
const (
	entityMetaMaxTables = 64
	entityMetaMaxFields = 512
)

// scanEntityMeta collects [Entity] classes from a parsed program. Called next
// to scanBootAnnotations during emitProgram.
func (g *Generator) scanEntityMeta(prog *ast.Program) error {
	for _, decl := range prog.Declarations {
		switch d := decl.(type) {
		case *ast.TypeDecl:
			if cd, ok := d.Type.(*ast.ClassDecl); ok {
				cd.Name = d.Name
				if err := g.scanEntityMetaClass(d.Name, mergeAttributes(d.Attributes, cd.Attributes), cd); err != nil {
					return err
				}
			}
		case *ast.ClassDecl:
			if err := g.scanEntityMetaClass(d.Name, d.Attributes, d); err != nil {
				return err
			}
		}
	}
	if len(g.entityMeta) > entityMetaMaxTables {
		return fmt.Errorf("too many [Entity] classes: %d (limit %d)", len(g.entityMeta), entityMetaMaxTables)
	}
	fields := 0
	for _, e := range g.entityMeta {
		fields += len(e.Fields)
	}
	if fields > entityMetaMaxFields {
		return fmt.Errorf("too many [Entity] columns: %d (limit %d)", fields, entityMetaMaxFields)
	}
	return nil
}

func (g *Generator) scanEntityMetaClass(className string, attrs []*ast.Attribute, cd *ast.ClassDecl) error {
	if className == "" || cd == nil {
		return nil
	}
	entityAttr := entitymetaapi.FindAttribute(attrs, "Entity")
	if entityAttr == nil {
		return nil
	}
	table, ok := entitymetaapi.StringArg(entityAttr)
	if !ok || table == "" {
		return nil // diagnostic emitted by pkg/compiler
	}

	label := table
	if l, ok := entitymetaapi.StringArg(entitymetaapi.FindAttribute(attrs, "Label")); ok && l != "" {
		label = l
	}
	flags := ""
	if entitymetaapi.FindAttribute(attrs, "ReadOnly") != nil {
		flags = "readonly"
	}

	entity := llvmEntity{Table: table, Meta: strings.Join([]string{"", label, flags}, "|")}
	pkColumn := ""
	for _, field := range cd.Fields {
		column := ""
		if c, ok := entitymetaapi.StringArg(entitymetaapi.FindAttribute(field.Attributes, "Column")); ok && c != "" {
			column = c
		}
		fieldLabel, _ := entitymetaapi.StringArg(entitymetaapi.FindAttribute(field.Attributes, "Label"))
		isPK := entitymetaapi.FindAttribute(field.Attributes, "PrimaryKey") != nil
		fieldType, _ := fieldTypeName(field.Type)
		for _, name := range field.Names {
			col := column
			if col == "" {
				col = name
			}
			kind := entitymetaapi.Kind(name, fieldType)
			lbl := fieldLabel
			if lbl == "" {
				lbl = name
			}
			entity.Fields = append(entity.Fields, llvmEntityField{
				Entry: strings.Join([]string{col, kind, lbl, entitymetaapi.FieldFlags(field.Attributes, kind)}, "|"),
			})
			if isPK && pkColumn == "" {
				pkColumn = col
			}
		}
	}
	if pkColumn == "" {
		// Same fallback as the Go scanner: a field named Id (any case).
		for i, field := range cd.Fields {
			for _, name := range field.Names {
				if strings.EqualFold(name, "Id") && i < len(entity.Fields) {
					pkColumn = strings.SplitN(entity.Fields[i].Entry, "|", 2)[0]
				}
			}
			if pkColumn != "" {
				break
			}
		}
	}
	// Rebuild Meta now that the primary key is known.
	entity.Meta = strings.Join([]string{pkColumn, label, flags}, "|")
	g.entityMeta = append(g.entityMeta, entity)
	return nil
}

// emitEntityMetaWiring writes the registration calls at the current position
// (the top of main, next to emitBootAutoWiring). No-op when the program has no
// [Entity] classes — that gating is what keeps the bootstrap IR fixed point.
func (g *Generator) emitEntityMetaWiring() {
	if len(g.entityMeta) == 0 {
		return
	}
	g.pendingModuleGlobals = append(g.pendingModuleGlobals, entityMetaGlobals()...)
	g.line("  ; --- entity metadata (v0.11.0 CRUD engine) ---")
	names := make([]string, 0, len(g.entityMeta))
	for _, entity := range g.entityMeta {
		names = append(names, entity.Table)
		g.enqueueStdlib("entitymeta", "RegisterEntity", "RegisterEntity", 2)
		g.line(fmt.Sprintf("  call void @__kylix_entitymeta_RegisterEntity(ptr %s, ptr %s)",
			g.addString(entity.Table), g.addString(entity.Meta)))
		for _, field := range entity.Fields {
			g.enqueueStdlib("entitymeta", "RegisterEntityField", "RegisterEntityField", 2)
			g.line(fmt.Sprintf("  call void @__kylix_entitymeta_RegisterEntityField(ptr %s, ptr %s)",
				g.addString(entity.Table), g.addString(field.Entry)))
		}
	}
	g.enqueueStdlib("entitymeta", "SetEntityNames", "SetEntityNames", 1)
	g.line(fmt.Sprintf("  call void @__kylix_entitymeta_SetEntityNames(ptr %s)", g.addString(strings.Join(names, ","))))
}
