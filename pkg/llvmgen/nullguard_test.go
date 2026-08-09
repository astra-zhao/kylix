package llvmgen_test

import (
	"strings"
	"testing"
)

// TestNullGuard_StringCompare (v5.6.0) guards the fix for the bootstrap segfault
// at `cd.Parent <> ”` in GenerateClassDecl. A parentless class has cd.Parent
// unset → null ptr in the LLVM backend (Go backend uses ""). strcmp(null, ...)
// segfaulted (exit 139 / EXC_BAD_ACCESS in _platform_strcmp). String comparison
// and concatenation now null-guard their operands, normalizing null → the
// @__kylix_emptystr global before calling strcmp/strlen/strcpy/strcat.
func TestNullGuard_StringCompare(t *testing.T) {
	ir := generateIR(t, `program p;
type
  TThing = class
    Name: String;
  end;
var t: TThing;
begin
  t := TThing.Create;
  if t.Name <> '' then WriteLn('named') else WriteLn('unnamed');
end.`)
	// t.Name is unset (null). The `<> ''` comparison must not pass null straight
	// to strcmp — it should be routed through the null guard (icmp eq, null +
	// select @__kylix_emptystr) before strcmp.
	if !strings.Contains(ir, "@__kylix_emptystr") {
		t.Errorf("expected null-guarded string comparison to reference @__kylix_emptystr\nIR:\n%s", ir)
	}
	// No raw `strcmp(ptr null` — the null literal must be normalized away.
	// (The strcmp call's operands are now guarded registers, not the raw
	// field load.)
	if strings.Contains(ir, "call i32 @strcmp(ptr null") {
		t.Errorf("string comparison must not pass null directly to strcmp\nIR:\n%s", ir)
	}
}
