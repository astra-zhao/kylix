// openapi.go — OpenAPI 3.1 YAML generator from KylixBoot annotations.
//
// Scans [Controller]/[Get]/[Post]/[Entity]/[Body]/[Authenticated]/[Role]
// annotations and emits a valid OpenAPI 3.1 YAML document.
package openapi

import (
	"fmt"
	"kylix/ast"
	"kylix/lexer"
	"kylix/parser"
	"os"
	"sort"
	"strings"
)

// SchemaProperty is one field in an OpenAPI object schema.
type SchemaProperty struct {
	Type      string
	Format    string
	MinLength int
	MaxLength int
	Minimum   *float64
	Maximum   *float64
	HasMinLen bool
	HasMaxLen bool
}

// Schema is an OpenAPI object schema built from a Kylix entity/validation class.
type Schema struct {
	Type       string
	Properties map[string]*SchemaProperty
	Required   []string
	propOrder  []string
}

// Operation is one HTTP operation in the document.
type Operation struct {
	OperationID string
	Tags        []string
	Summary     string
	Method      string
	Path        string
	BodySchema  string
	RequireAuth bool
	Roles       []string
}

// OpenAPIDoc is the in-memory OpenAPI document.
type OpenAPIDoc struct {
	Title   string
	Version string
	Ops     []*Operation
	Schemas map[string]*Schema
	HasAuth bool
}

// Generate parses the given .klx files and builds an OpenAPIDoc from their
// KylixBoot annotations.
func Generate(files []string, title, version string) (*OpenAPIDoc, error) {
	if title == "" {
		title = "Kylix API"
	}
	if version == "" {
		version = "1.0.0"
	}
	doc := &OpenAPIDoc{
		Title:   title,
		Version: version,
		Schemas: map[string]*Schema{},
	}
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			return nil, fmt.Errorf("cannot read %s: %w", f, err)
		}
		prog := parser.New(lexer.New(string(src))).ParseProgram()
		scanProgram(doc, prog)
	}
	return doc, nil
}

func scanProgram(doc *OpenAPIDoc, prog *ast.Program) {
	for _, decl := range prog.Declarations {
		switch d := decl.(type) {
		case *ast.TypeDecl:
			if classDecl, ok := d.Type.(*ast.ClassDecl); ok {
				classDecl.Name = d.Name
				attrs := mergeAttrs(d.Attributes, classDecl.Attributes)
				scanClass(doc, d.Name, attrs, classDecl)
			}
		case *ast.ClassDecl:
			scanClass(doc, d.Name, d.Attributes, d)
		}
	}
}

func scanClass(doc *OpenAPIDoc, name string, attrs []*ast.Attribute, class *ast.ClassDecl) {
	if findAttr(attrs, "Entity") != nil || hasValidationFields(class) {
		schema := buildSchema(class)
		if len(schema.Properties) > 0 {
			doc.Schemas[name] = schema
		}
	}
	controller := findAttr(attrs, "Controller")
	if controller == nil {
		return
	}
	basePath := attrStringArg(controller, "")
	for _, method := range class.Methods {
		if method.Body == nil {
			continue
		}
		for _, attr := range method.Attributes {
			httpMethod := routeMethodName(attr.Name)
			if httpMethod == "" {
				continue
			}
			sub := attrStringArg(attr, "/")
			op := &Operation{
				OperationID: name + "_" + method.Name,
				Tags:        []string{controllerTag(name)},
				Summary:     method.Name,
				Method:      httpMethod,
				Path:        openAPIPath(normalizePath(basePath, sub)),
			}
			for _, ma := range method.Attributes {
				switch strings.ToLower(ma.Name) {
				case "authenticated":
					op.RequireAuth = true
				case "role":
					if role := attrStringArg(ma, ""); role != "" {
						op.Roles = append(op.Roles, role)
						op.RequireAuth = true
					}
				case "body":
					op.BodySchema = attrIdentArg(ma)
				}
			}
			if op.RequireAuth {
				doc.HasAuth = true
			}
			doc.Ops = append(doc.Ops, op)
		}
	}
}

func buildSchema(class *ast.ClassDecl) *Schema {
	s := &Schema{Type: "object", Properties: map[string]*SchemaProperty{}}
	for _, field := range class.Fields {
		if len(field.Names) == 0 {
			continue
		}
		prop := &SchemaProperty{Type: kyToOASType(field.Type)}
		for _, attr := range field.Attributes {
			switch strings.ToLower(attr.Name) {
			case "required":
				for _, n := range field.Names {
					s.Required = append(s.Required, n)
				}
			case "email":
				prop.Format = "email"
			case "minlen":
				if v := attrIntArg(attr); v > 0 {
					prop.MinLength = v
					prop.HasMinLen = true
				}
			case "maxlen":
				if v := attrIntArg(attr); v > 0 {
					prop.MaxLength = v
					prop.HasMaxLen = true
				}
			case "min":
				if v := attrFloat64Arg(attr); v != 0 {
					f := v
					prop.Minimum = &f
				}
			case "max":
				if v := attrFloat64Arg(attr); v != 0 {
					f := v
					prop.Maximum = &f
				}
			}
		}
		for _, n := range field.Names {
			if _, exists := s.Properties[n]; !exists {
				s.propOrder = append(s.propOrder, n)
			}
			s.Properties[n] = prop
		}
	}
	return s
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func mergeAttrs(a, b []*ast.Attribute) []*ast.Attribute {
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 {
		return a
	}
	out := make([]*ast.Attribute, 0, len(a)+len(b))
	return append(append(out, a...), b...)
}

func findAttr(attrs []*ast.Attribute, name string) *ast.Attribute {
	for _, a := range attrs {
		if strings.EqualFold(a.Name, name) {
			return a
		}
	}
	return nil
}

func attrStringArg(attr *ast.Attribute, fallback string) string {
	if attr == nil || len(attr.Args) == 0 {
		return fallback
	}
	if s, ok := attr.Args[0].(*ast.StringLiteral); ok {
		return s.Value
	}
	return fallback
}

func attrIdentArg(attr *ast.Attribute) string {
	if attr == nil || len(attr.Args) == 0 {
		return ""
	}
	if id, ok := attr.Args[0].(*ast.Identifier); ok {
		return id.Value
	}
	return ""
}

func attrIntArg(attr *ast.Attribute) int {
	if attr == nil || len(attr.Args) == 0 {
		return 0
	}
	if n, ok := attr.Args[0].(*ast.IntegerLiteral); ok {
		return int(n.Value)
	}
	return 0
}

func attrFloat64Arg(attr *ast.Attribute) float64 {
	if attr == nil || len(attr.Args) == 0 {
		return 0
	}
	switch v := attr.Args[0].(type) {
	case *ast.IntegerLiteral:
		return float64(v.Value)
	case *ast.FloatLiteral:
		return v.Value
	}
	return 0
}

func hasValidationFields(class *ast.ClassDecl) bool {
	if class == nil {
		return false
	}
	for _, field := range class.Fields {
		for _, attr := range field.Attributes {
			switch strings.ToLower(attr.Name) {
			case "required", "email", "min", "max", "minlen", "maxlen":
				return true
			}
		}
	}
	return false
}

func routeMethodName(name string) string {
	switch strings.ToLower(name) {
	case "get":
		return "get"
	case "post":
		return "post"
	case "put":
		return "put"
	case "delete":
		return "delete"
	}
	return ""
}

func normalizePath(base, sub string) string {
	base = strings.TrimSpace(base)
	sub = strings.TrimSpace(sub)
	if base == "" {
		if sub == "" || sub == "/" {
			return "/"
		}
		if strings.HasPrefix(sub, "/") {
			return sub
		}
		return "/" + sub
	}
	if sub == "" || sub == "/" {
		if strings.HasPrefix(base, "/") {
			return strings.TrimRight(base, "/")
		}
		return "/" + strings.TrimRight(base, "/")
	}
	return "/" + strings.Trim(strings.TrimRight(base, "/")+"/"+strings.TrimLeft(sub, "/"), "/")
}

// openAPIPath converts Kylix :param style to OpenAPI {param} style.
func openAPIPath(p string) string {
	parts := strings.Split(p, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[i] = "{" + part[1:] + "}"
		}
	}
	return strings.Join(parts, "/")
}

func controllerTag(className string) string {
	name := className
	if strings.HasPrefix(name, "T") && len(name) > 1 {
		name = name[1:]
	}
	if strings.HasSuffix(name, "Controller") {
		name = name[:len(name)-10]
	}
	return strings.ToLower(name)
}

func kyToOASType(expr ast.Expression) string {
	if expr == nil {
		return "string"
	}
	if id, ok := expr.(*ast.Identifier); ok {
		switch strings.ToLower(id.Value) {
		case "integer", "int64", "int":
			return "integer"
		case "real", "float", "double":
			return "number"
		case "boolean", "bool":
			return "boolean"
		}
	}
	return "string"
}

// RenderYAML produces an OpenAPI 3.1 YAML document from doc.
func RenderYAML(doc *OpenAPIDoc) string {
	var b strings.Builder

	b.WriteString("openapi: \"3.1.0\"\n")
	b.WriteString("info:\n")
	b.WriteString(fmt.Sprintf("  title: %s\n", yamlStr(doc.Title)))
	b.WriteString(fmt.Sprintf("  version: %s\n", yamlStr(doc.Version)))

	// Group ops by path, preserving insertion order.
	type pathEntry struct {
		path string
		ops  []*Operation
	}
	pathIndex := map[string]int{}
	var pathOrder []pathEntry
	for _, op := range doc.Ops {
		if i, ok := pathIndex[op.Path]; ok {
			pathOrder[i].ops = append(pathOrder[i].ops, op)
		} else {
			pathIndex[op.Path] = len(pathOrder)
			pathOrder = append(pathOrder, pathEntry{path: op.Path, ops: []*Operation{op}})
		}
	}

	if len(pathOrder) > 0 {
		b.WriteString("paths:\n")
		for _, pe := range pathOrder {
			b.WriteString(fmt.Sprintf("  %s:\n", pe.path))
			for _, op := range pe.ops {
				b.WriteString(fmt.Sprintf("    %s:\n", op.Method))
				b.WriteString(fmt.Sprintf("      operationId: %s\n", op.OperationID))
				if len(op.Tags) > 0 {
					b.WriteString("      tags:\n")
					for _, tag := range op.Tags {
						b.WriteString(fmt.Sprintf("        - %s\n", tag))
					}
				}
				b.WriteString(fmt.Sprintf("      summary: %s\n", op.Summary))
				if op.RequireAuth {
					b.WriteString("      security:\n")
					b.WriteString("        - BearerAuth: []\n")
					for _, role := range op.Roles {
						b.WriteString(fmt.Sprintf("        # required role: %s\n", role))
					}
				}
				if op.BodySchema != "" {
					b.WriteString("      requestBody:\n")
					b.WriteString("        required: true\n")
					b.WriteString("        content:\n")
					b.WriteString("          application/json:\n")
					b.WriteString("            schema:\n")
					b.WriteString(fmt.Sprintf("              $ref: '#/components/schemas/%s'\n", op.BodySchema))
				}
				b.WriteString("      responses:\n")
				b.WriteString("        \"200\":\n")
				b.WriteString("          description: OK\n")
			}
		}
	}

	hasComponents := len(doc.Schemas) > 0 || doc.HasAuth
	if hasComponents {
		b.WriteString("components:\n")
		if len(doc.Schemas) > 0 {
			b.WriteString("  schemas:\n")
			names := make([]string, 0, len(doc.Schemas))
			for n := range doc.Schemas {
				names = append(names, n)
			}
			sort.Strings(names)
			for _, schemaName := range names {
				schema := doc.Schemas[schemaName]
				b.WriteString(fmt.Sprintf("    %s:\n", schemaName))
				b.WriteString(fmt.Sprintf("      type: %s\n", schema.Type))
				if len(schema.Required) > 0 {
					b.WriteString("      required:\n")
					for _, r := range schema.Required {
						b.WriteString(fmt.Sprintf("        - %s\n", r))
					}
				}
				if len(schema.propOrder) > 0 {
					b.WriteString("      properties:\n")
					for _, propName := range schema.propOrder {
						prop := schema.Properties[propName]
						b.WriteString(fmt.Sprintf("        %s:\n", propName))
						b.WriteString(fmt.Sprintf("          type: %s\n", prop.Type))
						if prop.Format != "" {
							b.WriteString(fmt.Sprintf("          format: %s\n", prop.Format))
						}
						if prop.HasMinLen {
							b.WriteString(fmt.Sprintf("          minLength: %d\n", prop.MinLength))
						}
						if prop.HasMaxLen {
							b.WriteString(fmt.Sprintf("          maxLength: %d\n", prop.MaxLength))
						}
						if prop.Minimum != nil {
							b.WriteString(fmt.Sprintf("          minimum: %g\n", *prop.Minimum))
						}
						if prop.Maximum != nil {
							b.WriteString(fmt.Sprintf("          maximum: %g\n", *prop.Maximum))
						}
					}
				}
			}
		}
		if doc.HasAuth {
			b.WriteString("  securitySchemes:\n")
			b.WriteString("    BearerAuth:\n")
			b.WriteString("      type: http\n")
			b.WriteString("      scheme: bearer\n")
			b.WriteString("      bearerFormat: JWT\n")
		}
	}

	return b.String()
}

func yamlStr(s string) string {
	if strings.ContainsAny(s, ":{}[]|>&*!,'\"") {
		return fmt.Sprintf("%q", s)
	}
	return s
}
