package llvmgen

import (
	"regexp"
	"strings"
)

// passes.go — process-in-LLVM IR optimization passes (v4.5.0 Phase C).
//
// Post-generation text transforms applied to the complete IR module BEFORE it
// is written to .ll and fed to llc. They complement the external `opt --O<N>`
// pipeline (which runs LLVM's own passes on the same IR):
//
//   - opt handles SSA-level optimizations (mem2reg, inline, loop induction,
//     full DCE) when --llvm-opt is set — the gold standard.
//   - these passes handle cheap, always-safe IR-text cleanups that don't
//     require the full LLVM pass infrastructure and run by default (no flag).
//     They reduce IR size and binary size for the common -O0 case.
//
// Conservative: only deletes instructions PROVEN dead; never touches
// instructions whose result might be observed (stores, calls, rets, branches).

// PassPipeline runs a sequence of IR-text optimization passes.
type PassPipeline struct {
	passes []irPass
}

type irPass interface {
	name() string
	run(ir string) string
}

// DefaultPassPipeline returns the standard pass set: ConstantFold then DCE.
func DefaultPassPipeline() *PassPipeline {
	return &PassPipeline{
		passes: []irPass{constantFoldPass{}, deadCodeElimPass{}},
	}
}

func (p *PassPipeline) Run(ir string) string {
	for _, pass := range p.passes {
		ir = pass.run(ir)
	}
	return ir
}

// ---- ConstantFold pass ----
//
// Structural hook for constant folding. The MVP performs no transformation
// (DCE below handles the common "add i64 0, 0" dead arithmetic the generator
// emits for unsupported-receiver stubs). Future use-rewriting + dead-def
// deletion can land here without changing the pipeline structure.
type constantFoldPass struct{}

func (constantFoldPass) name() string         { return "constantfold" }
func (constantFoldPass) run(ir string) string { return ir }

// ---- DeadCodeElim (DCE) pass ----
//
// Removes definitions of SSA temporary registers (%tN) that are never
// referenced anywhere else in the module. A %tN is dead if:
//
//  1. Defined by a pure instruction (add/sub/mul/and/or/xor/shl/icmp/load/
//     getelementptr/ptrtoint/inttoptr/zext/sext/trunc/select/insertvalue/
//     extractvalue) — NOT call/store (side effects).
//  2. The token "%tN" appears exactly once in the whole module (its own
//     definition) — no other instruction reads it.
type deadCodeElimPass struct{}

func (deadCodeElimPass) name() string { return "dce" }

var pureDefRe = regexp.MustCompile(`^  %(t\d+) = (add|sub|mul|udiv|sdiv|urem|srem|and|or|xor|shl|lshr|ashr|icmp|fcmp|load|getelementptr|ptrtoint|inttoptr|zext|sext|trunc|bitcast|select|insertvalue|extractvalue|freeze) `)

func (deadCodeElimPass) run(ir string) string {
	lines := strings.Split(ir, "\n")
	type def struct {
		reg string
		idx int
	}
	var defs []def
	for i, ln := range lines {
		if m := pureDefRe.FindStringSubmatch(ln); m != nil {
			defs = append(defs, def{reg: m[1], idx: i})
		}
	}
	if len(defs) == 0 {
		return ir
	}
	dead := make(map[int]bool)
	for _, d := range defs {
		if countOccurrences(ir, "%"+d.reg) == 1 {
			dead[d.idx] = true
		}
	}
	if len(dead) == 0 {
		return ir
	}
	out := make([]string, 0, len(lines)-len(dead))
	for i, ln := range lines {
		if !dead[i] {
			out = append(out, ln)
		}
	}
	return strings.Join(out, "\n")
}

// countOccurrences counts non-overlapping occurrences of needle in haystack,
// treating the match as a "word": the char following the match must not be a
// digit (so "%t1" doesn't match inside "%t10").
func countOccurrences(haystack, needle string) int {
	count := 0
	for {
		idx := strings.Index(haystack, needle)
		if idx < 0 {
			break
		}
		end := idx + len(needle)
		if end >= len(haystack) || !isDigitByte(haystack[end]) {
			count++
		}
		haystack = haystack[end:]
	}
	return count
}

func isDigitByte(b byte) bool { return b >= '0' && b <= '9' }
