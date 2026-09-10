package generator_test

// stdlib_boot_names_test.go — "three name lists" consistency check (v0.7.1
// P4). The Boot* API surface is declared in two independent hand-maintained
// lists — the Go host's stdlib heuristic (generator_stdlib.go, decides which
// calls resolve to the stdlib Go package) and the LLVM backend's module list
// (pkg/llvmgen/stdlib.go) — and dispatched in emitBootCall /
// bootStubReturnTypes (pkg/llvmgen/stdlib_boot.go). When the lists drift
// apart, one backend compiles a program the other rejects with
// `undefined: BootNotFoundPage` (the exact v0.7.0 release-week bug). These
// tests parse the source lists and fail on any asymmetric drift.
//
// Parsing the source text (instead of reflecting on the unexported maps)
// keeps the check honest: regexes against hand-edited list literals are
// exactly the drift surface being guarded.

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

var identRe = regexp.MustCompile(`"([A-Za-z_][A-Za-z0-9_]*)"`)

// extractList finds `anchor` in src, skips whitespace and // comments to the
// first bracket ('(' for strToSet, '{' for map literals), and returns every
// quoted identifier in the balanced literal.
func extractList(t *testing.T, src, anchor string) map[string]bool {
	t.Helper()
	i := strings.Index(src, anchor)
	if i < 0 {
		t.Fatalf("anchor %q not found — list moved or renamed, update this test", anchor)
	}
	rest := src[i+len(anchor):]
	// The opening bracket: already the anchor's last byte when the anchor
	// includes it ("boot": {), otherwise the first non-space/non-comment
	// character after the anchor (strToSet).
	var open byte
	start := -1
	if c := anchor[len(anchor)-1]; c == '(' || c == '{' {
		open, start = c, 0
	} else {
		for j := 0; j < len(rest); j++ {
			if strings.HasPrefix(rest[j:], "//") {
				for j < len(rest) && rest[j] != '\n' {
					j++
				}
				continue
			}
			c := rest[j]
			if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
				continue
			}
			if c == '(' || c == '{' {
				open, start = c, j
				break
			}
			t.Fatalf("expected ( or { after %q, got %q", anchor, string(c))
		}
	}
	if start < 0 {
		t.Fatalf("no bracket literal after %q", anchor)
	}
	close := map[byte]byte{'(': ')', '{': '}'}[open]
	depth := 1
	j := start + 1
	for j < len(rest) && depth > 0 {
		switch rest[j] {
		case open:
			depth++
		case close:
			depth--
		}
		j++
	}
	if depth != 0 {
		t.Fatalf("unbalanced literal after %q", anchor)
	}
	names := map[string]bool{}
	for _, m := range identRe.FindAllStringSubmatch(rest[start:j], -1) {
		names[m[1]] = true
	}
	return names
}

func readSrc(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v (test must run from the generator/ dir)", path, err)
	}
	return string(data)
}

// TestBootNameLists_Converged: the Go host and LLVM backend boot lists must
// be identical sets. Only the boot module is compared — the other modules'
// heuristics legitimately differ between backends.
func TestBootNameLists_Converged(t *testing.T) {
	goHost := extractList(t, readSrc(t, "generator_stdlib.go"), `"boot": strToSet`)
	llvm := extractList(t, readSrc(t, "../pkg/llvmgen/stdlib.go"), `"boot": {`)

	var missingGo, missingLLVM []string
	for name := range llvm {
		if !goHost[name] {
			missingGo = append(missingGo, name)
		}
	}
	for name := range goHost {
		if !llvm[name] {
			missingLLVM = append(missingLLVM, name)
		}
	}
	if len(missingGo) > 0 || len(missingLLVM) > 0 {
		t.Fatalf("boot name lists drifted (the v0.7.0 `undefined: Boot*` bug class):\n"+
			"  in LLVM list, missing from Go host generator_stdlib.go: %v\n"+
			"  in Go host list, missing from pkg/llvmgen/stdlib.go:    %v\n"+
			"add the names to BOTH lists — go build passes on one backend and fails on the other otherwise",
			missingGo, missingLLVM)
	}
	if len(goHost) < 20 {
		t.Fatalf("boot list suspiciously small (%d names) — anchor matched the wrong literal?", len(goHost))
	}
}

// TestBootNames_Dispatchable: every boot name in the LLVM list must be
// handled by the LLVM backend — either an explicit case in emitBootCall or a
// bootStubReturnTypes entry (the generic stub path keys its return type on
// it; a name absent from both still links but types as i64 by accident).
func TestBootNames_Dispatchable(t *testing.T) {
	llvm := extractList(t, readSrc(t, "../pkg/llvmgen/stdlib.go"), `"boot": {`)
	boot := readSrc(t, "../pkg/llvmgen/stdlib_boot.go")

	stubs := extractList(t, boot, "bootStubReturnTypes = map[string]string{")
	// case labels: `case "A", "B":` — capture all quoted idents on case lines.
	caseRe := regexp.MustCompile(`case ((?:"[A-Za-z_][A-Za-z0-9_]*"(?:,\s*)?)+):`)
	handled := map[string]bool{}
	for _, m := range caseRe.FindAllStringSubmatch(boot, -1) {
		for _, id := range identRe.FindAllStringSubmatch(m[1], -1) {
			handled[id[1]] = true
		}
	}

	var unhandled []string
	for name := range llvm {
		if !handled[name] && !stubs[name] {
			unhandled = append(unhandled, name)
		}
	}
	if len(unhandled) > 0 {
		t.Fatalf("boot names in the LLVM stdlib list with no emitBootCall case and no bootStubReturnTypes entry: %v\n"+
			"add a case in stdlib_boot.go or a bootStubReturnTypes entry (return type defaults to i64)", unhandled)
	}
}
