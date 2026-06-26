package generator

import (
	"fmt"
	"kylix/ast"
	"strings"
)

// validationField captures a single annotated field for validation codegen.
type validationField struct {
	ClassName  string
	FieldName  string
	FieldType  string
	Attributes []*ast.Attribute
	SourceLine int
}

// validationAttrNames are the field annotations recognised by the validator.
var validationAttrNames = map[string]bool{
	"required": true,
	"email":    true,
	"min":      true,
	"max":      true,
	"minlen":   true,
	"maxlen":   true,
}

func (g *Generator) scanValidationAnnotations(program *ast.Program) {
	for _, decl := range program.Declarations {
		switch d := decl.(type) {
		case *ast.TypeDecl:
			if classDecl, ok := d.Type.(*ast.ClassDecl); ok {
				classDecl.Name = d.Name
				g.scanValidationClass(d.Name, classDecl)
			}
		case *ast.ClassDecl:
			g.scanValidationClass(d.Name, d)
		}
	}
}

func (g *Generator) scanValidationClass(className string, classDecl *ast.ClassDecl) {
	if className == "" || classDecl == nil {
		return
	}
	var fields []validationField
	for _, field := range classDecl.Fields {
		attrs := collectValidationAttrs(field.Attributes)
		if len(attrs) == 0 {
			continue
		}
		fieldType, _ := fieldTypeName(field.Type)
		for _, name := range field.Names {
			fields = append(fields, validationField{
				ClassName:  className,
				FieldName:  name,
				FieldType:  fieldType,
				Attributes: attrs,
				SourceLine: field.Token.Line,
			})
			for _, attr := range attrs {
				switch strings.ToLower(attr.Name) {
				case "required":
					if isStringFieldType(fieldType) {
						g.imports["strings"] = true
					}
				case "email":
					g.imports["regexp"] = true
				}
			}
		}
	}
	if len(fields) == 0 {
		return
	}
	if _, exists := g.validationFields[className]; !exists {
		g.validatedOrder = append(g.validatedOrder, className)
	}
	g.validationFields[className] = append(g.validationFields[className], fields...)
}

func collectValidationAttrs(attrs []*ast.Attribute) []*ast.Attribute {
	var out []*ast.Attribute
	for _, attr := range attrs {
		if validationAttrNames[strings.ToLower(attr.Name)] {
			out = append(out, attr)
		}
	}
	return out
}

// classHasMethod reports whether classDecl already defines a method with the
// given name; used to avoid clobbering user-defined Validate/IsValid.
func classHasMethod(classDecl *ast.ClassDecl, name string) bool {
	if classDecl == nil {
		return false
	}
	for _, m := range classDecl.Methods {
		if strings.EqualFold(m.Name, name) {
			return true
		}
	}
	return false
}

// validationEmailRegex is the same conservative pattern used by stdlib.
const validationEmailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

// generateValidationMethods emits Validate() and IsValid() for a class when
// any of its fields carry validation annotations. It is invoked by class codegen
// once per class. Skipped when the user already defined Validate or IsValid.
func (g *Generator) generateValidationMethods(className string, typeParams []*ast.TypeParameter, classDecl *ast.ClassDecl) {
	fields := g.validationFields[className]
	if len(fields) == 0 {
		return
	}
	hasValidate := classHasMethod(classDecl, "Validate")
	hasIsValid := classHasMethod(classDecl, "IsValid")
	if hasValidate && hasIsValid {
		return
	}

	if !hasValidate {
		g.emitValidateMethod(className, typeParams, fields)
	}
	if !hasIsValid {
		g.emitIsValidMethod(className, typeParams)
	}
}

func (g *Generator) emitValidateMethod(className string, typeParams []*ast.TypeParameter, fields []validationField) {
	g.write("func (self *")
	g.writeClassReceiverType(className, typeParams)
	g.writeLine(") Validate() map[string]string {")
	g.indent++
	g.writeLine("errors := map[string]string{}")

	needsStrings := false
	needsRegexp := false

	for _, field := range fields {
		for _, attr := range field.Attributes {
			switch strings.ToLower(attr.Name) {
			case "required":
				if isStringFieldType(field.FieldType) {
					needsStrings = true
					g.writeLine(fmt.Sprintf("if strings.TrimSpace(self.%s) == \"\" {", field.FieldName))
					g.indent++
					g.writeLine(fmt.Sprintf("errors[%q] = \"is required\"", field.FieldName))
					g.indent--
					g.writeLine("}")
				} else {
					g.writeLine(fmt.Sprintf("if self.%s == 0 {", field.FieldName))
					g.indent++
					g.writeLine(fmt.Sprintf("errors[%q] = \"is required\"", field.FieldName))
					g.indent--
					g.writeLine("}")
				}
			case "email":
				needsRegexp = true
				g.writeLine(fmt.Sprintf("if self.%s != \"\" && !regexp.MustCompile(%q).MatchString(self.%s) {", field.FieldName, validationEmailRegex, field.FieldName))
				g.indent++
				g.writeLine(fmt.Sprintf("errors[%q] = \"must be a valid email address\"", field.FieldName))
				g.indent--
				g.writeLine("}")
			case "minlen":
				if n, ok := attributeIntArg(attr); ok {
					g.writeLine(fmt.Sprintf("if len(self.%s) < %d {", field.FieldName, n))
					g.indent++
					g.writeLine(fmt.Sprintf("errors[%q] = \"must be at least %d characters\"", field.FieldName, n))
					g.indent--
					g.writeLine("}")
				}
			case "maxlen":
				if n, ok := attributeIntArg(attr); ok {
					g.writeLine(fmt.Sprintf("if len(self.%s) > %d {", field.FieldName, n))
					g.indent++
					g.writeLine(fmt.Sprintf("errors[%q] = \"must be at most %d characters\"", field.FieldName, n))
					g.indent--
					g.writeLine("}")
				}
			case "min":
				if n, ok := attributeIntArg(attr); ok {
					g.writeLine(fmt.Sprintf("if self.%s < %d {", field.FieldName, n))
					g.indent++
					g.writeLine(fmt.Sprintf("errors[%q] = \"must be at least %d\"", field.FieldName, n))
					g.indent--
					g.writeLine("}")
				}
			case "max":
				if n, ok := attributeIntArg(attr); ok {
					g.writeLine(fmt.Sprintf("if self.%s > %d {", field.FieldName, n))
					g.indent++
					g.writeLine(fmt.Sprintf("errors[%q] = \"must be at most %d\"", field.FieldName, n))
					g.indent--
					g.writeLine("}")
				}
			}
		}
	}

	g.writeLine("return errors")
	g.indent--
	g.writeLine("}")
	g.writeLine("")

	if needsStrings {
		g.imports["strings"] = true
	}
	if needsRegexp {
		g.imports["regexp"] = true
	}
}

func (g *Generator) emitIsValidMethod(className string, typeParams []*ast.TypeParameter) {
	g.write("func (self *")
	g.writeClassReceiverType(className, typeParams)
	g.writeLine(") IsValid() bool {")
	g.indent++
	g.writeLine("return len(self.Validate()) == 0")
	g.indent--
	g.writeLine("}")
	g.writeLine("")
}

func isStringFieldType(t string) bool {
	switch t {
	case "String", "string":
		return true
	}
	return false
}

func attributeIntArg(attr *ast.Attribute) (int64, bool) {
	if attr == nil || len(attr.Args) == 0 {
		return 0, false
	}
	if lit, ok := attr.Args[0].(*ast.IntegerLiteral); ok {
		return lit.Value, true
	}
	return 0, false
}
