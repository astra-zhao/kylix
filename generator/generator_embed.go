package generator

import (
	"strings"

	"kylix/ast"
	"kylix/internal/embedfiles"
)

// generator_embed.go — [Embed('dir', …)] file baking (v0.12.0).
//
// A file-level attribute before the program header bakes the listed directories
// into the binary:
//
//	[Embed('views', 'static')]
//	program KylixAdmin;
//
// The generated init() registers every file's contents with the stdlib, and
// sysutil.ReadFile (plus the boot static handler on the LLVM side) consults
// that registry before touching the filesystem — so the built binary runs with
// no views/ or static/ directory beside it, and the application code does not
// change at all.
//
// Emission is gated on the attribute: a program without one produces exactly
// the code it did before, which is what keeps the bootstrap IR fixed point and
// the tutorial outputs untouched.

// embedDirs returns the directories named by the program's [Embed(...)]
// attributes, in declaration order and de-duplicated.
func embedDirs(program *ast.Program) []string {
	if program == nil {
		return nil
	}
	var dirs []string
	seen := map[string]bool{}
	for _, attr := range program.Attributes {
		if !strings.EqualFold(attr.Name, "Embed") {
			continue
		}
		for _, arg := range attr.Args {
			lit, ok := arg.(*ast.StringLiteral)
			if !ok || lit.Value == "" {
				continue
			}
			if seen[lit.Value] {
				continue
			}
			seen[lit.Value] = true
			dirs = append(dirs, lit.Value)
		}
	}
	return dirs
}

// readEmbedFiles walks the directories via internal/embedfiles, so both
// backends bake the same set.
func readEmbedFiles(dirs []string) ([]embedfiles.File, error) {
	return embedfiles.Read(dirs)
}

// hasEmbedAttributes reports whether the program bakes anything.
func hasEmbedAttributes(program *ast.Program) bool {
	return len(embedDirs(program)) > 0
}

// emitEmbeddedFiles writes the init() that registers the baked files. No-op
// when the program has no [Embed] attribute.
func (g *Generator) emitEmbeddedFiles(program *ast.Program) {
	dirs := embedDirs(program)
	if len(dirs) == 0 {
		return
	}
	files, err := readEmbedFiles(dirs)
	if err != nil {
		// pkg/compiler already reported this as a diagnostic; emitting nothing
		// keeps the generated file valid so the user sees the real error.
		return
	}
	if len(files) == 0 {
		return
	}
	g.imports["kylix/stdlib"] = true
	g.writeLine("// --- embedded files (v0.12.0 [Embed]) ---")
	g.writeLine("func init() {")
	g.indent++
	for _, f := range files {
		g.writeLine("stdlib.RegisterEmbedded(" + goStringLiteral(f.Name) + ", " + goStringLiteral(f.Content) + ")")
	}
	g.indent--
	g.writeLine("}")
}
