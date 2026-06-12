// generator_expr.go — Expression code generation.
package generator

import (
	"fmt"
	"kylix/ast"
	"strings"
)

func (g *Generator) generateExpression(expr ast.Expression) {
	switch e := expr.(type) {
	case *ast.Identifier:
		// Apply name substitution map (e.g., ON clause variable E → e).
		if mapped, ok := g.nameMap[e.Value]; ok {
			g.write(mapped)
		} else {
			g.write(g.mapBuiltinFunction(e.Value))
		}
	case *ast.IntegerLiteral:
		g.write(fmt.Sprintf("%d", e.Value))
	case *ast.FloatLiteral:
		g.write(fmt.Sprintf("%f", e.Value))
	case *ast.StringLiteral:
		g.write(`"` + escapeGoString(e.Value) + `"`)
	case *ast.StringInterpolation:
		g.writeInterpolation(e)
	case *ast.BooleanLiteral:
		if e.Value {
			g.write("true")
		} else {
			g.write("false")
		}
	case *ast.NilLiteral:
		g.write("nil")
	case *ast.ArrayLiteral:
		if len(e.Elements) == 0 {
			g.write("nil")
		} else {
			g.write("[]interface{}{")
			for i, elem := range e.Elements {
				if i > 0 {
					g.write(", ")
				}
				g.generateExpression(elem)
			}
			g.write("}")
		}
	case *ast.PrefixExpression:
		op := e.Operator
		if op == "not" {
			op = "!"
		}
		g.write("(" + op)
		g.generateExpression(e.Right)
		g.write(")")
	case *ast.InfixExpression:
		g.write("(")
		g.generateExpression(e.Left)
		g.write(" " + mapOperator(e.Operator) + " ")
		g.generateExpression(e.Right)
		g.write(")")
	case *ast.CallExpression:
		g.generateCallExpression(e)
	case *ast.MemberExpression:
		// ClassName.Create without args → &ClassName{}
		if e.Member == "Create" {
			if ident, ok := e.Object.(*ast.Identifier); ok {
				g.write("&" + ident.Value + "{}")
				return
			}
		}
		g.generateExpression(e.Object)
		g.write("." + e.Member)
	case *ast.IndexExpression:
		g.generateExpression(e.Left)
		g.write("[")
		g.generateExpression(e.Index)
		g.write("]")
	case *ast.SliceExpression:
		g.generateExpression(e.Left)
		g.write("[")
		if e.Low != nil {
			g.generateExpression(e.Low)
		}
		g.write(":")
		if e.High != nil {
			g.generateExpression(e.High)
		}
		g.write("]")
	case *ast.LambdaExpression:
		g.generateLambdaExpression(e)
	case *ast.AwaitExpression:
		// await expr → <-expr (channel receive)
		g.write("<-")
		g.generateExpression(e.Expression)
	case *ast.IsExpression:
		g.write("func() bool { _, ok := ")
		g.generateExpression(e.Expression)
		g.write(".(")
		g.generateTypeExpressionForCast(e.TargetType)
		g.write("); return ok }()")
	case *ast.TypeCastExpression:
		g.generateExpression(e.Expression)
		g.write(".(")
		g.generateTypeExpressionForCast(e.TargetType)
		g.write(")")
	case *ast.TupleLiteral:
		for i, elem := range e.Elements {
			if i > 0 {
				g.write(", ")
			}
			g.generateExpression(elem)
		}
	case *ast.MapType:
		// Map type used as a value: map[K]V{}
		g.write("map[")
		if e.KeyType != nil {
			g.generateTypeExpression(e.KeyType)
		} else {
			g.write("string")
		}
		g.write("]")
		if e.ValueType != nil {
			g.generateTypeExpression(e.ValueType)
		} else {
			g.write("interface{}")
		}
		g.write("{}")
	}
}

// generateCallExpression handles special built-in function call rewrites.
func (g *Generator) generateCallExpression(e *ast.CallExpression) {
	// ClassName.Create(args) → &ClassName{Field: arg, ...}
	if member, ok := e.Function.(*ast.MemberExpression); ok && member.Member == "Create" {
		if ident, ok := member.Object.(*ast.Identifier); ok {
			g.write("&" + ident.Value + "{")
			fields := g.classFields[ident.Value]
			for i, arg := range e.Arguments {
				if i > 0 {
					g.write(", ")
				}
				if i < len(fields) {
					g.write(fields[i] + ": ")
				}
				g.generateExpression(arg)
			}
			g.write("}")
			return
		}
	}

	if ident, ok := e.Function.(*ast.Identifier); ok {
		switch ident.Value {
		case "Ord":
			if len(e.Arguments) == 1 {
				// Ord(s) → func() int { if len(s)==0 { return 0 }; return int(s[0]) }()
				g.write("func() int { if len(")
				g.generateExpression(e.Arguments[0])
				g.write(") == 0 { return 0 }; return int(")
				g.generateExpression(e.Arguments[0])
				g.write("[0]) }()")
				return
			}
		case "Length":
			if len(e.Arguments) == 1 {
				g.write("int64(len(")
				g.generateExpression(e.Arguments[0])
				g.write("))")
				return
			}
		case "IntToStr":
			if len(e.Arguments) == 1 {
				g.imports["fmt"] = true
				g.write(`fmt.Sprintf("%d", `)
				g.generateExpression(e.Arguments[0])
				g.write(")")
				return
			}
		case "StrToInt64":
			if len(e.Arguments) == 1 {
				g.imports["strconv"] = true
				g.write("func() int64 { v, _ := strconv.ParseInt(")
				g.generateExpression(e.Arguments[0])
				g.write(", 10, 64); return v }()")
				return
			}
		case "StrToFloat":
			if len(e.Arguments) == 1 {
				g.imports["strconv"] = true
				g.write("func() float64 { v, _ := strconv.ParseFloat(")
				g.generateExpression(e.Arguments[0])
				g.write(", 64); return v }()")
				return
			}
		case "ReadFile":
			if len(e.Arguments) == 1 {
				g.imports["os"] = true
				g.write("func() string { data, _ := os.ReadFile(")
				g.generateExpression(e.Arguments[0])
				g.write("); return string(data) }()")
				return
			}
		}
	}

	// Generic call
	g.generateExpression(e.Function)
	g.write("(")
	for i, arg := range e.Arguments {
		if i > 0 {
			g.write(", ")
		}
		g.generateExpression(arg)
	}
	g.write(")")
}

func (g *Generator) generateLambdaExpression(e *ast.LambdaExpression) {
	g.write("func(")
	for i, param := range e.Parameters {
		if i > 0 {
			g.write(", ")
		}
		g.write(param.Name)
		if param.Type != nil {
			g.write(" ")
			g.generateTypeExpression(param.Type)
		}
	}
	g.write(") ")
	switch body := e.Body.(type) {
	case *ast.BlockStatement:
		g.writeLine("{")
		g.indent++
		for _, s := range body.Statements {
			g.generateStatement(s)
		}
		g.indent--
		g.write("}")
	case ast.Expression:
		g.writeLine("{")
		g.indent++
		g.write("return ")
		g.generateExpression(body)
		g.write("\n")
		g.indent--
		g.write("}")
	}
}

// mapOperator converts Kylix infix operators to Go equivalents.
func mapOperator(op string) string {
	switch op {
	case "and":
		return "&&"
	case "or":
		return "||"
	case "xor":
		return "^"
	case "div":
		return "/"
	case "mod":
		return "%"
	case "<>":
		return "!="
	case "=":
		return "=="
	}
	return op
}

// escapeGoString escapes a Kylix string value for use inside Go double-quoted literals.
// Strategy: protect \n, \t, \r with markers → escape \ and " → restore markers.
func escapeGoString(s string) string {
	s = strings.ReplaceAll(s, `\n`, "\x00n")
	s = strings.ReplaceAll(s, `\t`, "\x00t")
	s = strings.ReplaceAll(s, `\r`, "\x00r")
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\x00n", `\n`)
	s = strings.ReplaceAll(s, "\x00t", `\t`)
	s = strings.ReplaceAll(s, "\x00r", `\r`)
	return s
}
