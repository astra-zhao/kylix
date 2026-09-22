package llvmgen

import (
	"fmt"
	"strings"

	"kylix/ast"
	"kylix/internal/embedfiles"
)

// embed.go — [Embed('dir', …)] file baking on the LLVM backend (v0.12.0).
//
// The Go backend registers baked files with pkg/boot; here the file set is
// known at compile time, so it becomes two parallel constant arrays plus a
// linear-scan lookup — no registration calls at run time:
//
//	@__kylix_embed_names [N x ptr]   file paths, as the program asks for them
//	@__kylix_embed_data  [N x ptr]   their contents
//
// ReadFile and the static handler call @__kylix_embed_get first and fall back
// to the filesystem, so a built binary needs no views/ or static/ directory.
//
// Emission is gated on the attribute: a program without one emits none of this
// (the bootstrap IR fixed point and the tutorial outputs stay untouched).

// embedFiles collects the program's baked files, sorted by path so the emitted
// IR is deterministic.
func (g *Generator) embedFiles() []embedfiles.File {
	if g.program == nil {
		return nil
	}
	var dirs []string
	for _, attr := range g.program.Attributes {
		if !strings.EqualFold(attr.Name, "Embed") {
			continue
		}
		for _, arg := range attr.Args {
			if lit, ok := arg.(*ast.StringLiteral); ok && lit.Value != "" {
				dirs = append(dirs, lit.Value)
			}
		}
	}
	if len(dirs) == 0 {
		return nil
	}
	files, err := embedfiles.Read(dirs)
	if err != nil {
		// pkg/compiler already reported this as a diagnostic.
		return nil
	}
	return files
}

// emitEmbedGlobals declares the two module-level arrays and the file count.
// They start zeroed and are filled by @__kylix_embed_init: the pointer-to-string
// conversions are body instructions (getelementptr), so they cannot appear in a
// global initialiser.
func (g *Generator) emitEmbedGlobals(files []embedfiles.File) {
	n := len(files)
	g.line(fmt.Sprintf("@__kylix_embed_count = global i64 %d", n))
	g.line(fmt.Sprintf("@__kylix_embed_names = global [%d x ptr] zeroinitializer", n))
	g.line(fmt.Sprintf("@__kylix_embed_data = global [%d x ptr] zeroinitializer", n))
}

// emitEmbedInitBody fills the arrays. Emitted at module level and called at the
// top of main, the same shape the boot route table uses.
func (g *Generator) emitEmbedInitBody() {
	files := g.embedFileList
	if len(files) == 0 {
		return
	}
	n := len(files)
	g.line("define void @__kylix_embed_init() {")
	g.line("entry:")
	for i, f := range files {
		np := g.ptrTo(g.addString(f.Name), len(f.Name)+1)
		slot := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr inbounds [%d x ptr], ptr @__kylix_embed_names, i64 0, i64 %d", slot, n, i))
		g.line(fmt.Sprintf("  store ptr %s, ptr %s", np, slot))
		dp := g.ptrTo(g.addString(f.Content), len(f.Content)+1)
		dslot := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr inbounds [%d x ptr], ptr @__kylix_embed_data, i64 0, i64 %d", dslot, n, i))
		g.line(fmt.Sprintf("  store ptr %s, ptr %s", dp, dslot))
	}
	g.line("  ret void")
	g.line("}")
	g.line("")
}

// emitEmbedGetBody emits the lookup: a linear scan over the name array,
// returning the matching content pointer or null.
func (g *Generator) emitEmbedGetBody(files []embedfiles.File) {
	if len(files) == 0 {
		return
	}
	n := len(files)
	g.line("define ptr @__kylix_embed_get(ptr %name) {")
	g.line("entry:")
	iSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", iSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", iSlot))
	loop := g.label()
	body := g.label()
	hit := g.label()
	miss := g.label()
	g.line(fmt.Sprintf("  br label %%%s", loop))
	g.line(fmt.Sprintf("%s:", loop))
	i := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", i, iSlot))
	more := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, %d", more, i, n))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", more, body, miss))
	g.line(fmt.Sprintf("%s:", body))
	nslot := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [%d x ptr], ptr @__kylix_embed_names, i64 0, i64 %s", nslot, n, i))
	nptr := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", nptr, nslot))
	cmp := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @strcmp(ptr %s, ptr %%name)", cmp, nptr))
	eq := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 0", eq, cmp))
	next := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", eq, hit, next))
	g.line(fmt.Sprintf("%s:", next))
	iNext := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", iNext, i))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", iNext, iSlot))
	g.line(fmt.Sprintf("  br label %%%s", loop))
	g.line(fmt.Sprintf("%s:", hit))
	dslot := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [%d x ptr], ptr @__kylix_embed_data, i64 0, i64 %s", dslot, n, i))
	dptr := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", dptr, dslot))
	g.line(fmt.Sprintf("  ret ptr %s", dptr))
	g.line(fmt.Sprintf("%s:", miss))
	g.line("  ret ptr null")
	g.line("}")
	g.line("")
}

// emitEmbeddedFiles emits the arrays and the lookup for the program's [Embed]
// directories. No-op without the attribute, so a program that does not bake
// anything produces exactly the IR it did before.
func (g *Generator) emitEmbeddedFiles() {
	files := g.embedFiles()
	if len(files) == 0 {
		return
	}
	g.embedFileList = files
	g.emitEmbedGlobals(files)
	g.emitEmbedGetBody(files)
	g.hasEmbedded = true
}
