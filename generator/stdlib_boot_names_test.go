package generator_test

// stdlib_boot_names_test.go — Boot* API surface guards (v0.7.1 P4, reworked
// in v0.7.2). The function list itself lives in internal/bootapi as the
// single source of truth imported by both backends (generator_stdlib.go and
// pkg/llvmgen/stdlib.go), so list drift is now a compile error and the old
// host-vs-LLVM source-parsing comparison is gone.
//
// What still needs a test: the *dispatch* side. emitBootCall cases and
// bootStubReturnTypes (pkg/llvmgen/stdlib_boot.go) are hand-maintained — a
// name present in bootapi but absent from both still links yet types as i64
// by accident. TestBootNames_Dispatchable catches that.

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"kylix/internal/bootapi"
)

var identRe = regexp.MustCompile(`"([A-Za-z_][A-Za-z0-9_]*)"`)

// extractList finds `anchor` in src, skips whitespace and // comments to the
// first bracket ('(' or '{'), and returns every quoted identifier in the
// balanced literal.
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

// TestBootNames_Sane guards the single source of truth itself: no
// duplicates, every name carries the Boot prefix, and the jwt cross-listing
// constant is part of the surface.
func TestBootNames_Sane(t *testing.T) {
	seen := map[string]bool{}
	for _, n := range bootapi.BootFunctions {
		if seen[n] {
			t.Fatalf("duplicate name in bootapi.BootFunctions: %q", n)
		}
		seen[n] = true
		if !strings.HasPrefix(n, "Boot") {
			t.Fatalf("bootapi.BootFunctions entry without Boot prefix: %q", n)
		}
	}
	if len(seen) < 20 {
		t.Fatalf("boot surface suspiciously small (%d names) — list truncated?", len(seen))
	}
	if !seen[bootapi.BootRegisterJwtAuth] {
		t.Fatalf("jwt cross-listing constant %q missing from bootapi.BootFunctions", bootapi.BootRegisterJwtAuth)
	}
	sorted := append([]string(nil), bootapi.BootFunctions...)
	sort.Strings(sorted)
	for i := range sorted {
		if sorted[i] != bootapi.BootFunctions[i] {
			t.Fatalf("bootapi.BootFunctions must stay sorted alphabetically; first disorder at %q (got %q, want %q)",
				bootapi.BootFunctions[i], bootapi.BootFunctions[i], sorted[i])
		}
	}
}

// TestBootNames_Dispatchable: every name in the single-source boot surface
// must be handled by the LLVM backend — either an explicit case in
// emitBootCall or a bootStubReturnTypes entry (the generic stub path keys
// its return type on it; a name absent from both still links but types as
// i64 by accident).
func TestBootNames_Dispatchable(t *testing.T) {
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
	for _, name := range bootapi.BootFunctions {
		if !handled[name] && !stubs[name] {
			unhandled = append(unhandled, name)
		}
	}
	if len(unhandled) > 0 {
		t.Fatalf("boot names in internal/bootapi with no emitBootCall case and no bootStubReturnTypes entry: %v\n"+
			"add a case in pkg/llvmgen/stdlib_boot.go or a bootStubReturnTypes entry (return type defaults to i64)", unhandled)
	}
}
