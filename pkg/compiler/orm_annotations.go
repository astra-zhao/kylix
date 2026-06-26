package compiler

import (
	"fmt"
	"kylix/ast"
)

// CheckORMAnnotations validates [Entity], [Repository], and [Query] usage
// before code generation. It catches missing string args, unknown repository
// entity targets, and misplaced or unsupported [Query] return types.
func CheckORMAnnotations(programs []*ast.Program, files []string) []Diagnostic {
	entities := map[string]bool{}
	type classInfo struct {
		File      string
		ClassName string
		Attrs     []*ast.Attribute
		Class     *ast.ClassDecl
	}
	var classes []classInfo
	for i, program := range programs {
		file := ""
		if i < len(files) {
			file = files[i]
		}
		for _, decl := range program.Declarations {
			switch d := decl.(type) {
			case *ast.TypeDecl:
				if classDecl, ok := d.Type.(*ast.ClassDecl); ok {
					classes = append(classes, classInfo{
						File:      file,
						ClassName: d.Name,
						Attrs:     mergeBootAnnotationAttrs(d.Attributes, classDecl.Attributes),
						Class:     classDecl,
					})
				}
			case *ast.ClassDecl:
				classes = append(classes, classInfo{
					File:      file,
					ClassName: d.Name,
					Attrs:     d.Attributes,
					Class:     d,
				})
			}
		}
	}
	for _, c := range classes {
		if findBootAnnotation(c.Attrs, "Entity") != nil {
			entities[c.ClassName] = true
		}
	}

	var diags []Diagnostic
	for _, c := range classes {
		entityAttr := findBootAnnotation(c.Attrs, "Entity")
		repoAttr := findBootAnnotation(c.Attrs, "Repository")

		if entityAttr != nil {
			if name, ok := bootAnnotationStringArg(entityAttr, ""); !ok || name == "" {
				diags = append(diags, NewError(c.File, entityAttr.Token.Line, entityAttr.Token.Column,
					ErrInvalidORM, "[Entity] requires a string table name argument"))
			}
			if repoAttr != nil {
				diags = append(diags, NewError(c.File, repoAttr.Token.Line, repoAttr.Token.Column,
					ErrInvalidORM, fmt.Sprintf("class %s cannot be both [Entity] and [Repository]", c.ClassName)))
			}
		}

		if repoAttr != nil {
			if len(repoAttr.Args) == 0 {
				diags = append(diags, NewError(c.File, repoAttr.Token.Line, repoAttr.Token.Column,
					ErrInvalidORM, "[Repository] requires an entity class argument"))
			} else {
				ident, ok := repoAttr.Args[0].(*ast.Identifier)
				if !ok {
					diags = append(diags, NewError(c.File, repoAttr.Token.Line, repoAttr.Token.Column,
						ErrInvalidORM, "[Repository] argument must be an entity class identifier"))
				} else if !entities[ident.Value] {
					diags = append(diags, NewErrorHint(c.File, repoAttr.Token.Line, repoAttr.Token.Column,
						ErrInvalidORM,
						fmt.Sprintf("[Repository] references unknown entity %s", ident.Value),
						fmt.Sprintf("Annotate %s with [Entity('table_name')] first.", ident.Value)))
				}
			}
		}

		// Validate [Query] usage on methods.
		for _, method := range c.Class.Methods {
			queryAttr := findBootAnnotation(method.Attributes, "Query")
			if queryAttr == nil {
				continue
			}
			if repoAttr == nil {
				diags = append(diags, NewErrorHint(c.File, queryAttr.Token.Line, queryAttr.Token.Column,
					ErrInvalidORM,
					fmt.Sprintf("[Query] on %s.%s requires a [Repository] class", c.ClassName, method.Name),
					"Move the method into a class annotated with [Repository(TEntity)]."))
				continue
			}
			if sql, ok := bootAnnotationStringArg(queryAttr, ""); !ok || sql == "" {
				diags = append(diags, NewError(c.File, queryAttr.Token.Line, queryAttr.Token.Column,
					ErrInvalidORM, "[Query] requires a SQL string argument"))
				continue
			}
			if !validORMQueryReturn(method, entities) {
				diags = append(diags, NewErrorHint(c.File, queryAttr.Token.Line, queryAttr.Token.Column,
					ErrInvalidORM,
					fmt.Sprintf("[Query] %s.%s must return an [Entity] class or array of [Entity] class", c.ClassName, method.Name),
					"Use: function Foo(...): TUser; or function Foo(...): array of TUser;"))
			}
		}
	}
	return diags
}

func checkORMAnnotations(program *ast.Program, file string) []Diagnostic {
	return CheckORMAnnotations([]*ast.Program{program}, []string{file})
}

func validORMQueryReturn(method *ast.FunctionDecl, entities map[string]bool) bool {
	if method == nil || method.ReturnType == nil {
		return false
	}
	switch t := method.ReturnType.(type) {
	case *ast.Identifier:
		return entities[t.Value]
	case *ast.ArrayType:
		if ident, ok := t.ElementType.(*ast.Identifier); ok {
			return entities[ident.Value]
		}
	}
	return false
}
