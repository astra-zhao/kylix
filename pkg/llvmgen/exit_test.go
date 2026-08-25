package llvmgen_test

import (
	"strings"
	"testing"
)

// blockBody returns the IR instructions belonging to the basic block labeled
// `label` (up to the next block label or the function's closing brace),
// excluding the label line itself.
func blockBody(lines []string, label string) string {
	var out []string
	started := false
	for _, ln := range lines {
		s := strings.TrimSpace(ln)
		if s == label+":" {
			started = true
			continue
		}
		if !started {
			continue
		}
		if s == "}" {
			break
		}
		if strings.HasSuffix(s, ":") && (strings.HasPrefix(s, "lbl") || s == "entry") {
			break // next block
		}
		out = append(out, ln)
	}
	return strings.Join(out, "\n")
}

// firstBrTargetAfter returns the label branched to by the first `br label %X`
// line at or after lines[start], or "" if none.
func firstBrTargetAfter(lines []string, start int) string {
	for j := start; j < len(lines); j++ {
		ln := lines[j]
		if !strings.Contains(ln, "br label") {
			continue
		}
		// `  br label %lblN`
		i := strings.Index(ln, "%")
		if i < 0 {
			continue
		}
		return strings.TrimPrefix(ln[i:], "%") // lblN
	}
	return ""
}

// TestExit_EarlyReturnDoesNotFallThrough (v0.5.6) guards the fix for the bug
// where `Exit` inside `if … then begin …; result := A; Exit; end;` was silently
// dropped: the codegen emitted no terminator for `Exit`, so the then-block
// fell through to the statements after the `if` and overwrote `result` (e.g.
// the bootstrap parser parsed `x := 1` into a TAssignmentStatement that was
// then overwritten by the fall-through expression-statement builder, emitting
// bare `x`). With the fix, `Exit` branches to the function's single exit block
// (the block holding `ret %result`), so the fall-through `result := 200` is
// reachable only via the if's false branch.
func TestExit_EarlyReturnDoesNotFallThrough(t *testing.T) {
	ir := generateIR(t, `program p;
function Pick(x: Integer): Integer;
begin
  if x > 0 then
  begin
    result := 100;
    Exit;
  end;
  result := 200;
end;
var n: Integer;
begin
  n := Pick(5);
  WriteLn(n);
end.`)

	lines := strings.Split(ir, "\n")

	// Find the Exit-path assignment constant `add i64 0, 100`.
	idx100 := -1
	for i, ln := range lines {
		if strings.Contains(ln, "add i64 0, 100") {
			idx100 = i
			break
		}
	}
	if idx100 == -1 {
		t.Fatalf("expected `add i64 0, 100` in IR\nIR:\n%s", ir)
	}

	// The next `br label %X` after the 100 assignment is the Exit branch's
	// target. With the fix it must be the exit (ret) block, NOT the
	// `result := 200` block — i.e. the targeted block must contain `ret i64`
	// and must NOT contain `add i64 0, 200` (which would mean the then-block
	// fell through into the 200 assignment).
	target := firstBrTargetAfter(lines, idx100+1)
	if target == "" {
		t.Fatalf("expected a `br label` after `add i64 0, 100` (the Exit branch)\nIR:\n%s", ir)
	}
	body := blockBody(lines, target)
	if !strings.Contains(body, "ret i64") {
		t.Errorf("Exit branch should target the exit block (containing `ret i64`), but block %s body is:\n%s\nIR:\n%s", target, body, ir)
	}
	if strings.Contains(body, "add i64 0, 200") {
		t.Errorf("Exit branch targets the `result := 200` block — `Exit` was dropped so the then-block falls through into result := 200 (overwriting the assignment)\ntarget block %s body:\n%s\nIR:\n%s", target, body, ir)
	}

	// The function must have exactly one `ret` (single exit block).
	if c := strings.Count(ir, "ret i64"); c != 1 {
		t.Errorf("expected exactly one `ret i64` (single exit block), got %d\nIR:\n%s", c, ir)
	}
}

// TestExit_BareProcedureReturn ensures `Exit;` in a void procedure branches to
// the procedure's exit block (`ret void`) rather than being a no-op. The Exit
// path's branch must target a block holding `ret void`, not the fall-through
// `WriteLn(name)` block.
func TestExit_BareProcedureReturn(t *testing.T) {
	ir := generateIR(t, `program p;
procedure Greet(name: String);
begin
  if Length(name) = 0 then
  begin
    Exit;
  end;
  WriteLn(name);
end;
begin
  Greet('hi');
end.`)
	// `ret void` should appear exactly once (single exit block), and since the
	// Exit path must reach it, the branch (br label) must exist.
	if c := strings.Count(ir, "ret void"); c != 1 {
		t.Errorf("expected exactly one `ret void` (single exit block), got %d\nIR:\n%s", c, ir)
	}
	if !strings.Contains(ir, "br label") {
		t.Errorf("expected Exit to emit a `br label` to the exit block\nIR:\n%s", ir)
	}
}
