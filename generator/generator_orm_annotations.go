package generator

import (
	"fmt"
	"kylix/ast"
	"strings"
)

// ormColumn maps an entity field to its database column name.
type ormColumn struct {
	FieldName string
	FieldType string
	Column    string
}

// ormEntity captures the table/column metadata declared by [Entity] annotations.
type ormEntity struct {
	ClassName string
	Table     string
	PKField   string
	PKColumn  string
	Fields    []ormColumn
}

// ormQuery describes a [Query]-annotated method on a repository class.
type ormQuery struct {
	MethodName   string
	SQL          string
	ReturnEntity string
	ReturnsList  bool
	Params       []*ast.Parameter
}

// ormRepository captures a [Repository(TEntity)] class.
type ormRepository struct {
	ClassName   string
	EntityName  string
	UserDefined map[string]bool // method names already implemented by the user
	Queries     []ormQuery
}

func (g *Generator) scanORMAnnotations(program *ast.Program) {
	// Pass 1: collect entities so repositories can reference them.
	for _, decl := range program.Declarations {
		switch d := decl.(type) {
		case *ast.TypeDecl:
			if classDecl, ok := d.Type.(*ast.ClassDecl); ok {
				classDecl.Name = d.Name
				g.scanORMEntity(d.Name, mergeAttributes(d.Attributes, classDecl.Attributes), classDecl)
			}
		case *ast.ClassDecl:
			g.scanORMEntity(d.Name, d.Attributes, d)
		}
	}
	// Pass 2: collect repositories now that all entities are known.
	for _, decl := range program.Declarations {
		switch d := decl.(type) {
		case *ast.TypeDecl:
			if classDecl, ok := d.Type.(*ast.ClassDecl); ok {
				g.scanORMRepository(d.Name, mergeAttributes(d.Attributes, classDecl.Attributes), classDecl)
			}
		case *ast.ClassDecl:
			g.scanORMRepository(d.Name, d.Attributes, d)
		}
	}
	if len(g.ormEntities) > 0 || len(g.ormRepositories) > 0 {
		g.imports["kylix/stdlib"] = true
	}
}

func (g *Generator) scanORMEntity(className string, attrs []*ast.Attribute, classDecl *ast.ClassDecl) {
	if className == "" || classDecl == nil {
		return
	}
	entityAttr := findAttribute(attrs, "Entity")
	if entityAttr == nil {
		return
	}
	table, ok := attributeStringArg(entityAttr, "")
	if !ok || table == "" {
		return // diagnostic emitted by compiler
	}
	entity := &ormEntity{ClassName: className, Table: table}
	for _, field := range classDecl.Fields {
		colName := ""
		if colAttr := findAttribute(field.Attributes, "Column"); colAttr != nil {
			if name, ok := attributeStringArg(colAttr, ""); ok {
				colName = name
			}
		}
		isPK := findAttribute(field.Attributes, "PrimaryKey") != nil
		fieldType, _ := fieldTypeName(field.Type)
		for _, name := range field.Names {
			column := colName
			if column == "" {
				column = name
			}
			entity.Fields = append(entity.Fields, ormColumn{
				FieldName: name,
				FieldType: fieldType,
				Column:    column,
			})
			if isPK && entity.PKField == "" {
				entity.PKField = name
				entity.PKColumn = column
			}
		}
	}
	if entity.PKField == "" {
		for _, c := range entity.Fields {
			if strings.EqualFold(c.FieldName, "Id") {
				entity.PKField = c.FieldName
				entity.PKColumn = c.Column
				break
			}
		}
	}
	if _, exists := g.ormEntities[className]; !exists {
		g.ormEntitiesOrder = append(g.ormEntitiesOrder, className)
	}
	g.ormEntities[className] = entity
}

func (g *Generator) scanORMRepository(className string, attrs []*ast.Attribute, classDecl *ast.ClassDecl) {
	if className == "" || classDecl == nil {
		return
	}
	repoAttr := findAttribute(attrs, "Repository")
	if repoAttr == nil {
		return
	}
	if len(repoAttr.Args) == 0 {
		return
	}
	entityIdent, ok := repoAttr.Args[0].(*ast.Identifier)
	if !ok {
		return
	}
	if _, known := g.ormEntities[entityIdent.Value]; !known {
		return
	}
	repo := ormRepository{ClassName: className, EntityName: entityIdent.Value, UserDefined: map[string]bool{}}
	for _, method := range classDecl.Methods {
		if method.Body != nil {
			repo.UserDefined[method.Name] = true
		}
		queryAttr := findAttribute(method.Attributes, "Query")
		if queryAttr == nil {
			continue
		}
		sql, ok := attributeStringArg(queryAttr, "")
		if !ok || sql == "" {
			continue
		}
		entityName, list, ok := ormQueryReturnEntity(method, g.ormEntities)
		if !ok {
			continue
		}
		if method.Body != nil {
			// User implemented the method; skip generation but keep the user copy.
			delete(repo.UserDefined, method.Name)
			repo.UserDefined[method.Name] = true
			continue
		}
		repo.Queries = append(repo.Queries, ormQuery{
			MethodName:   method.Name,
			SQL:          sql,
			ReturnEntity: entityName,
			ReturnsList:  list,
			Params:       method.Parameters,
		})
	}
	g.ormRepositories = append(g.ormRepositories, repo)
}

func ormQueryReturnEntity(method *ast.FunctionDecl, entities map[string]*ormEntity) (string, bool, bool) {
	if method == nil || method.ReturnType == nil {
		return "", false, false
	}
	switch t := method.ReturnType.(type) {
	case *ast.Identifier:
		if _, ok := entities[t.Value]; ok {
			return t.Value, false, true
		}
	case *ast.ArrayType:
		if ident, ok := t.ElementType.(*ast.Identifier); ok {
			if _, known := entities[ident.Value]; known {
				return ident.Value, true, true
			}
		}
	}
	return "", false, false
}

// ── Emission ─────────────────────────────────────────────────────────────────

func (g *Generator) generateORMEntityMethods(className string, typeParams []*ast.TypeParameter, classDecl *ast.ClassDecl) {
	entity := g.ormEntities[className]
	if entity == nil {
		return
	}
	hasToRow := classHasMethod(classDecl, "ToRow")
	hasFromRow := classHasMethod(classDecl, "FromRow")
	if !hasToRow {
		g.emitEntityToRow(className, typeParams, entity)
	}
	if !hasFromRow {
		g.emitEntityFromRow(className, typeParams, entity)
	}
}

func (g *Generator) emitEntityToRow(className string, typeParams []*ast.TypeParameter, entity *ormEntity) {
	g.write("func (self *")
	g.writeClassReceiverType(className, typeParams)
	g.writeLine(") ToRow() map[string]interface{} {")
	g.indent++
	g.writeLine("return map[string]interface{}{")
	g.indent++
	for _, col := range entity.Fields {
		g.writeLine(fmt.Sprintf("%q: self.%s,", col.Column, col.FieldName))
	}
	g.indent--
	g.writeLine("}")
	g.indent--
	g.writeLine("}")
	g.writeLine("")
}

func (g *Generator) emitEntityFromRow(className string, typeParams []*ast.TypeParameter, entity *ormEntity) {
	g.write("func (self *")
	g.writeClassReceiverType(className, typeParams)
	g.writeLine(") FromRow(row map[string]interface{}) {")
	g.indent++
	for _, col := range entity.Fields {
		goType := ormGoType(col.FieldType)
		if goType == "" {
			continue
		}
		g.writeLine(fmt.Sprintf("if v, ok := row[%q].(%s); ok {", col.Column, goType))
		g.indent++
		g.writeLine(fmt.Sprintf("self.%s = v", col.FieldName))
		g.indent--
		g.writeLine("}")
	}
	g.indent--
	g.writeLine("}")
	g.writeLine("")
}

func (g *Generator) generateORMRepositoryMethods(className string, typeParams []*ast.TypeParameter, classDecl *ast.ClassDecl) {
	var repo *ormRepository
	for i := range g.ormRepositories {
		if g.ormRepositories[i].ClassName == className {
			repo = &g.ormRepositories[i]
			break
		}
	}
	if repo == nil {
		return
	}
	entity := g.ormEntities[repo.EntityName]
	if entity == nil {
		return
	}
	if !repo.UserDefined["FindAll"] && !classHasMethod(classDecl, "FindAll") {
		g.emitRepoFindAll(className, typeParams, entity)
	}
	if !repo.UserDefined["FindById"] && !classHasMethod(classDecl, "FindById") {
		g.emitRepoFindById(className, typeParams, entity)
	}
	if !repo.UserDefined["Save"] && !classHasMethod(classDecl, "Save") {
		g.emitRepoSave(className, typeParams, entity)
	}
	if !repo.UserDefined["DeleteById"] && !classHasMethod(classDecl, "DeleteById") {
		g.emitRepoDeleteById(className, typeParams, entity)
	}
	for _, q := range repo.Queries {
		g.emitRepoQuery(className, typeParams, entity, q)
	}
}

func (g *Generator) emitRepoFindAll(className string, typeParams []*ast.TypeParameter, entity *ormEntity) {
	g.write("func (self *")
	g.writeClassReceiverType(className, typeParams)
	g.writeLine(fmt.Sprintf(") FindAll(orm *stdlib.ORM) []*%s {", entity.ClassName))
	g.indent++
	g.writeLine(fmt.Sprintf("rows, err := orm.FindAll(%q)", entity.Table))
	g.writeLine("if err != nil { return nil }")
	g.writeLine(fmt.Sprintf("out := make([]*%s, 0, len(rows))", entity.ClassName))
	g.writeLine("for _, row := range rows {")
	g.indent++
	g.writeLine(fmt.Sprintf("e := &%s{}", entity.ClassName))
	g.writeLine("e.FromRow(row)")
	g.writeLine("out = append(out, e)")
	g.indent--
	g.writeLine("}")
	g.writeLine("return out")
	g.indent--
	g.writeLine("}")
	g.writeLine("")
}

func (g *Generator) emitRepoFindById(className string, typeParams []*ast.TypeParameter, entity *ormEntity) {
	g.write("func (self *")
	g.writeClassReceiverType(className, typeParams)
	g.writeLine(fmt.Sprintf(") FindById(orm *stdlib.ORM, id int64) *%s {", entity.ClassName))
	g.indent++
	g.writeLine(fmt.Sprintf("row, err := orm.Find(%q, id)", entity.Table))
	g.writeLine("if err != nil || row == nil { return nil }")
	g.writeLine(fmt.Sprintf("e := &%s{}", entity.ClassName))
	g.writeLine("e.FromRow(row)")
	g.writeLine("return e")
	g.indent--
	g.writeLine("}")
	g.writeLine("")
}

func (g *Generator) emitRepoSave(className string, typeParams []*ast.TypeParameter, entity *ormEntity) {
	pkField := entity.PKField
	if pkField == "" {
		pkField = "Id"
	}
	pkColumn := entity.PKColumn
	if pkColumn == "" {
		pkColumn = "id"
	}
	g.write("func (self *")
	g.writeClassReceiverType(className, typeParams)
	g.writeLine(fmt.Sprintf(") Save(orm *stdlib.ORM, e *%s) int64 {", entity.ClassName))
	g.indent++
	g.writeLine("if e == nil { return 0 }")
	g.writeLine(fmt.Sprintf("if e.%s == 0 {", pkField))
	g.indent++
	g.writeLine(fmt.Sprintf("id, _ := orm.Insert(%q, e.ToRow())", entity.Table))
	g.writeLine(fmt.Sprintf("e.%s = id", pkField))
	g.writeLine("return id")
	g.indent--
	g.writeLine("}")
	g.writeLine(fmt.Sprintf("orm.Update(%q, map[string]interface{}{%q: e.%s}, e.ToRow())", entity.Table, pkColumn, pkField))
	g.writeLine(fmt.Sprintf("return e.%s", pkField))
	g.indent--
	g.writeLine("}")
	g.writeLine("")
}

func (g *Generator) emitRepoDeleteById(className string, typeParams []*ast.TypeParameter, entity *ormEntity) {
	pkColumn := entity.PKColumn
	if pkColumn == "" {
		pkColumn = "id"
	}
	g.write("func (self *")
	g.writeClassReceiverType(className, typeParams)
	g.writeLine(") DeleteById(orm *stdlib.ORM, id int64) int64 {")
	g.indent++
	g.writeLine(fmt.Sprintf("n, _ := orm.Delete(%q, map[string]interface{}{%q: id})", entity.Table, pkColumn))
	g.writeLine("return n")
	g.indent--
	g.writeLine("}")
	g.writeLine("")
}

func (g *Generator) emitRepoQuery(className string, typeParams []*ast.TypeParameter, entity *ormEntity, q ormQuery) {
	g.write("func (self *")
	g.writeClassReceiverType(className, typeParams)
	g.write(fmt.Sprintf(") %s(orm *stdlib.ORM", q.MethodName))
	for _, p := range q.Params {
		g.write(", ")
		g.write(p.Name + " ")
		if p.Type != nil {
			g.generateTypeExpression(p.Type)
		} else {
			g.write("interface{}")
		}
	}
	if q.ReturnsList {
		g.writeLine(fmt.Sprintf(") []*%s {", entity.ClassName))
	} else {
		g.writeLine(fmt.Sprintf(") *%s {", entity.ClassName))
	}
	g.indent++
	args := ormQueryArgList(q.Params)
	if q.ReturnsList {
		g.writeLine(fmt.Sprintf("rows, err := orm.QueryAll(%q%s)", q.SQL, args))
		g.writeLine("if err != nil { return nil }")
		g.writeLine(fmt.Sprintf("out := make([]*%s, 0, len(rows))", entity.ClassName))
		g.writeLine("for _, row := range rows {")
		g.indent++
		g.writeLine(fmt.Sprintf("e := &%s{}", entity.ClassName))
		g.writeLine("e.FromRow(row)")
		g.writeLine("out = append(out, e)")
		g.indent--
		g.writeLine("}")
		g.writeLine("return out")
	} else {
		g.writeLine(fmt.Sprintf("row, err := orm.Query(%q%s)", q.SQL, args))
		g.writeLine("if err != nil || row == nil { return nil }")
		g.writeLine(fmt.Sprintf("e := &%s{}", entity.ClassName))
		g.writeLine("e.FromRow(row)")
		g.writeLine("return e")
	}
	g.indent--
	g.writeLine("}")
	g.writeLine("")
}

func ormQueryArgList(params []*ast.Parameter) string {
	if len(params) == 0 {
		return ""
	}
	parts := make([]string, 0, len(params))
	for _, p := range params {
		parts = append(parts, p.Name)
	}
	return ", " + strings.Join(parts, ", ")
}

func ormGoType(kylixType string) string {
	switch kylixType {
	case "String", "string":
		return "string"
	case "Integer", "Int", "int", "int64", "LongInt", "Cardinal", "Word", "Byte":
		return "int64"
	case "Real", "Float", "Double", "float64":
		return "float64"
	case "Boolean", "Bool", "bool":
		return "bool"
	}
	return ""
}
