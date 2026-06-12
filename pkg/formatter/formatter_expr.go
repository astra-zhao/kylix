// formatter_expr.go — Type and expression formatting.
package formatter

import (
	"fmt"
	"kylix/ast"
)

func (f *Formatter) formatType(typ interface{}) {
	switch t := typ.(type) {
	case *ast.Identifier:
		f.write(t.Value)
	case *ast.ArrayType:
		f.write("array")
		if t.Size != nil {
			f.write("[")
			f.formatExpression(t.Size)
			f.write("]")
		}
		f.write(" of ")
		f.formatType(t.ElementType)
	case *ast.RecordType:
		f.write("record\n")
		f.indent++
		for _, field := range t.Fields {
			f.formatVarDecl(field)
		}
		f.indent--
		f.writeIndent()
		f.write("end")
	case *ast.GenericType:
		f.write(t.Base + "<")
		for i, param := range t.TypeParams {
			if i > 0 {
				f.write(", ")
			}
			f.formatType(param)
		}
		f.write(">")
	}
}

func (f *Formatter) formatExpression(expr interface{}) {
	f.formatExpressionPrec(expr, precLowest)
}

func (f *Formatter) formatExpressionPrec(expr interface{}, parentPrec int) {
	switch e := expr.(type) {
	case *ast.Identifier:
		f.write(e.Value)
	case *ast.IntegerLiteral:
		f.write(fmt.Sprintf("%d", e.Value))
	case *ast.FloatLiteral:
		f.write(fmt.Sprintf("%g", e.Value))
	case *ast.StringLiteral:
		f.write("'" + e.Value + "'")
	case *ast.StringInterpolation:
		f.write("$'")
		for _, part := range e.Parts {
			if strPart, ok := part.(*ast.StringLiteral); ok {
				f.write(strPart.Value)
			} else {
				f.write("{")
				f.formatExpression(part)
				f.write("}")
			}
		}
		f.write("'")
	case *ast.BooleanLiteral:
		if e.Value {
			f.write("true")
		} else {
			f.write("false")
		}
	case *ast.NilLiteral:
		f.write("nil")
	case *ast.ArrayLiteral:
		f.write("[")
		for i, elem := range e.Elements {
			if i > 0 {
				f.write(", ")
			}
			f.formatExpression(elem)
		}
		f.write("]")
	case *ast.PrefixExpression:
		needParens := parentPrec > precPrefix
		if needParens {
			f.write("(")
		}
		f.write(e.Operator + " ")
		f.formatExpressionPrec(e.Right, precPrefix)
		if needParens {
			f.write(")")
		}
	case *ast.InfixExpression:
		prec := operatorPrecedence(e.Operator)
		needParens := parentPrec > prec
		if needParens {
			f.write("(")
		}
		f.formatExpressionPrec(e.Left, prec)
		f.write(" " + e.Operator + " ")
		// Use prec+1 for right operand (left-associative).
		f.formatExpressionPrec(e.Right, prec+1)
		if needParens {
			f.write(")")
		}
	case *ast.CallExpression:
		f.formatExpression(e.Function)
		f.write("(")
		for i, arg := range e.Arguments {
			if i > 0 {
				f.write(", ")
			}
			f.formatExpression(arg)
		}
		f.write(")")
	case *ast.MemberExpression:
		f.formatExpression(e.Object)
		f.write("." + e.Member)
	case *ast.IndexExpression:
		f.formatExpression(e.Left)
		f.write("[")
		f.formatExpression(e.Index)
		f.write("]")
	case *ast.LambdaExpression:
		f.write("(")
		for i, param := range e.Parameters {
			if i > 0 {
				f.write("; ")
			}
			f.write(param.Name + ": ")
			f.formatType(param.Type)
		}
		f.write(") -> ")
		switch body := e.Body.(type) {
		case ast.Expression:
			f.formatExpression(body)
		case *ast.BlockStatement:
			f.write("\n")
			f.indent++
			f.formatBlock(body)
			f.indent--
			f.writeIndent()
		}
	case *ast.AwaitExpression:
		f.write("await ")
		f.formatExpression(e.Expression)
	case *ast.IsExpression:
		f.formatExpression(e.Expression)
		f.write(" is ")
		f.formatType(e.TargetType)
	case *ast.TypeCastExpression:
		f.formatExpression(e.Expression)
		f.write(" as ")
		f.formatType(e.TargetType)
	}
}
