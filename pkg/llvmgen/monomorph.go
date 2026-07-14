// monomorph.go — Generic class monomorphization for the LLVM backend.
//
// Kylix supports generic classes (Go 1.18-style):
//
//   class TBox<T>
//     Value: T;
//     function Get(): T;
//     begin result := Value; end;
//   end;
//
// LLVM IR has no parametric types, so the LLVM backend specializes each
// distinct instantiation (`TBox<Integer>`, `TBox<String>`) into a separate
// struct + vtable + method set by cloning the template ClassDecl and
// substituting the type parameter identifiers throughout fields and methods.
//
// The mangled name format is `TBase_TArg1_TArg2`, e.g. `TBox_Integer`,
// `TPair_Integer_String`.
package llvmgen

import (
	"kylix/ast"
)

// registerGenericTemplate caches a generic ClassDecl by its base name for
// later specialization. The caller is responsible for *not* emitting the
// template directly.
func (g *Generator) registerGenericTemplate(decl *ast.ClassDecl) {
	g.genericTemplates[decl.Name] = decl
}

// isGenericTemplate reports whether a ClassDecl is a generic template that
// must be deferred and specialized on demand.
func isGenericTemplate(decl *ast.ClassDecl) bool {
	return decl != nil && len(decl.TypeParams) > 0
}

// collectInstantiations walks the program AST to find every concrete
// instantiation of a registered generic class and specializes each one.
// Idempotent: repeated instantiations of the same mangled type are emitted
// only the first time.
func (g *Generator) collectInstantiations(prog *ast.Program) error {
	for _, decl := range prog.Declarations {
		if err := g.visitDeclForGenerics(decl); err != nil {
			return err
		}
	}
	for _, stmt := range prog.Statements {
		if err := g.visitStmtForGenerics(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) visitDeclForGenerics(node ast.Node) error {
	switch d := node.(type) {
	case *ast.VarDecl:
		if err := g.maybeSpecializeType(d.Type); err != nil {
			return err
		}
	case *ast.FunctionDecl:
		for _, p := range d.Parameters {
			if err := g.maybeSpecializeType(p.Type); err != nil {
				return err
			}
		}
		if d.ReturnType != nil {
			if err := g.maybeSpecializeType(d.ReturnType); err != nil {
				return err
			}
		}
		if d.Body != nil {
			for _, stmt := range d.Body.Statements {
				if err := g.visitStmtForGenerics(stmt); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (g *Generator) visitStmtForGenerics(node ast.Statement) error {
	switch s := node.(type) {
	case *ast.VarDecl:
		if err := g.maybeSpecializeType(s.Type); err != nil {
			return err
		}
		// v4.8.0: type-inferred vars (var x := TBox<Integer>.Create()) carry
		// the GenericType in s.Value, not s.Type. Walk the initializer so
		// monomorphization triggers before emitMain references the
		// specialized class. Without this, example21's TStack<Integer>.Create()
		// never specialized and method calls resolved to "unsupported receiver".
		if s.Value != nil {
			if err := g.visitExprForGenerics(s.Value); err != nil {
				return err
			}
		}
	case *ast.AssignmentStatement:
		if err := g.visitExprForGenerics(s.Value); err != nil {
			return err
		}
	case *ast.ExpressionStatement:
		if err := g.visitExprForGenerics(s.Expression); err != nil {
			return err
		}
	case *ast.BlockStatement:
		for _, ss := range s.Statements {
			if err := g.visitStmtForGenerics(ss); err != nil {
				return err
			}
		}
	case *ast.IfStatement:
		if err := g.visitStmtForGenerics(s.Consequence); err != nil {
			return err
		}
		if s.Alternative != nil {
			if err := g.visitStmtForGenerics(s.Alternative); err != nil {
				return err
			}
		}
	case *ast.WhileStatement:
		if err := g.visitStmtForGenerics(s.Body); err != nil {
			return err
		}
	case *ast.ForStatement:
		if err := g.visitStmtForGenerics(s.Body); err != nil {
			return err
		}
	case *ast.RepeatStatement:
		if err := g.visitStmtForGenerics(s.Body); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) visitExprForGenerics(expr ast.Expression) error {
	switch e := expr.(type) {
	case *ast.GenericType:
		return g.maybeSpecializeType(e)
	case *ast.MemberExpression:
		return g.visitExprForGenerics(e.Object)
	case *ast.CallExpression:
		if err := g.visitExprForGenerics(e.Function); err != nil {
			return err
		}
		for _, a := range e.Arguments {
			if err := g.visitExprForGenerics(a); err != nil {
				return err
			}
		}
	}
	return nil
}

// maybeSpecializeType triggers monomorphization if expr is a GenericType
// referencing a registered template.
func (g *Generator) maybeSpecializeType(expr ast.Expression) error {
	gt, ok := expr.(*ast.GenericType)
	if !ok {
		return nil
	}
	tmpl, ok := g.genericTemplates[gt.Base]
	if !ok {
		return nil
	}
	mangled := mangleGeneric(gt.Base, gt.TypeParams)
	if mangled == "" || g.instantiations[mangled] {
		return nil
	}
	return g.specialize(tmpl, mangled, gt.TypeParams)
}

// mangleGeneric returns a unique name for one (base, typeArgs) pair, or "" if
// any arg can't be reduced to a simple identifier (e.g. nested generics —
// deferred to a future slice).
func mangleGeneric(base string, args []ast.Expression) string {
	out := base
	for _, a := range args {
		ident, ok := a.(*ast.Identifier)
		if !ok {
			return ""
		}
		out += "_" + ident.Value
	}
	return out
}

// specialize clones the generic template, substitutes type parameters, and
// emits the resulting concrete class through the existing emitClassDecl path.
func (g *Generator) specialize(tmpl *ast.ClassDecl, mangled string, typeArgs []ast.Expression) error {
	if len(tmpl.TypeParams) != len(typeArgs) {
		return nil // arity mismatch — let the Go-backend diagnostic catch it
	}
	subst := map[string]ast.Expression{}
	for i, tp := range tmpl.TypeParams {
		subst[tp.Name] = typeArgs[i]
	}

	clone := &ast.ClassDecl{
		Token:      tmpl.Token,
		Name:       mangled,
		Parent:     tmpl.Parent,
		Interfaces: append([]string(nil), tmpl.Interfaces...),
		Visibility: tmpl.Visibility,
	}
	// Substitute field types.
	for _, f := range tmpl.Fields {
		clone.Fields = append(clone.Fields, &ast.VarDecl{
			Token: f.Token,
			Names: append([]string(nil), f.Names...),
			Type:  substituteType(f.Type, subst),
			Value: f.Value,
		})
	}
	// Substitute method param and return types.
	for _, m := range tmpl.Methods {
		nm := &ast.FunctionDecl{
			Token:       m.Token,
			Name:        m.Name,
			Body:        m.Body,
			LocalDecls:  m.LocalDecls,
			IsAsync:     m.IsAsync,
			IsExternal:  m.IsExternal,
			ReturnType:  substituteType(m.ReturnType, subst),
		}
		for _, p := range m.Parameters {
			nm.Parameters = append(nm.Parameters, &ast.Parameter{
				Token: p.Token,
				Name:  p.Name,
				Type:  substituteType(p.Type, subst),
			})
		}
		clone.Methods = append(clone.Methods, nm)
	}
	// Properties carry types too — substitute minimally.
	for _, p := range tmpl.Properties {
		clone.Properties = append(clone.Properties, &ast.PropertyDecl{
			Token:   p.Token,
			Name:    p.Name,
			Type:    substituteType(p.Type, subst),
			Getter:  p.Getter,
			Setter:  p.Setter,
			Default: p.Default,
		})
	}

	g.instantiations[mangled] = true
	return g.emitClassDecl(clone)
}

// substituteType returns a copy of expr with any type-parameter identifiers
// replaced by their concrete arguments. Supports identifiers, generic types
// (recursing into their args), and array types.
func substituteType(expr ast.Expression, subst map[string]ast.Expression) ast.Expression {
	if expr == nil {
		return nil
	}
	switch t := expr.(type) {
	case *ast.Identifier:
		if repl, ok := subst[t.Value]; ok {
			return repl
		}
		return t
	case *ast.GenericType:
		newArgs := make([]ast.Expression, len(t.TypeParams))
		for i, a := range t.TypeParams {
			newArgs[i] = substituteType(a, subst)
		}
		return &ast.GenericType{Base: t.Base, TypeParams: newArgs}
	case *ast.ArrayType:
		return &ast.ArrayType{
			ElementType: substituteType(t.ElementType, subst),
			Size:        t.Size,
			Dynamic:     t.Dynamic,
		}
	}
	return expr
}
