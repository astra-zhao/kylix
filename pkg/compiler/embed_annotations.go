package compiler

import (
	"fmt"
	"os"

	"kylix/ast"
)

// CheckEmbedAnnotations validates the file-level [Embed('dir', …)] attribute
// (v0.12.0). The generator bakes those directories into the binary, so a typo
// would otherwise surface as a build that silently lacks its templates — the
// check turns it into a compile error with the path in it.
func CheckEmbedAnnotations(programs []*ast.Program, files []string) []Diagnostic {
	var diags []Diagnostic
	for i, program := range programs {
		file := ""
		if i < len(files) {
			file = files[i]
		}
		for _, attr := range program.Attributes {
			if attr.Name != "Embed" && attr.Name != "embed" {
				continue
			}
			if len(attr.Args) == 0 {
				diags = append(diags, NewErrorHint(file, attr.Token.Line, attr.Token.Column,
					ErrInvalidORM,
					"[Embed] requires at least one directory argument",
					"Use: [Embed('views', 'static')]"))
				continue
			}
			for _, arg := range attr.Args {
				lit, ok := arg.(*ast.StringLiteral)
				if !ok || lit.Value == "" {
					diags = append(diags, NewError(file, attr.Token.Line, attr.Token.Column,
						ErrInvalidORM, "[Embed] arguments must be non-empty string literals"))
					continue
				}
				info, err := os.Stat(lit.Value)
				if err != nil {
					diags = append(diags, NewErrorHint(file, attr.Token.Line, attr.Token.Column,
						ErrInvalidORM,
						fmt.Sprintf("[Embed] cannot read %s", lit.Value),
						"Paths are resolved relative to the directory the compiler runs in."))
					continue
				}
				if !info.IsDir() {
					diags = append(diags, NewError(file, attr.Token.Line, attr.Token.Column,
						ErrInvalidORM, fmt.Sprintf("[Embed] %s is not a directory", lit.Value)))
				}
			}
		}
	}
	return diags
}
