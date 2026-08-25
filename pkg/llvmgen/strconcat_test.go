package llvmgen_test

import (
	"strings"
	"testing"
)

// TestStringConcat_SizedBuffer (v0.5.6) guards the fix for string concatenation
// overflowing a fixed 512-byte buffer. emitStringConcat previously did
// `malloc(512) + strcpy(lv) + strcat(rv)` for EVERY `a + b` string concat — a
// fixed 512-byte buffer that overflowed as soon as the concatenated result
// exceeded 512 bytes, corrupting the heap and leaking garbage. This broke the
// bootstrap's self-host (generator.klx builds the entire Go output by repeated
// `self.Output := self.Output + s`; once Output crossed 512 bytes the buffer
// overflowed → "strings."-like garbage / unescaped newlines → non-compilable Go).
// Now the buffer is sized strlen(lv)+strlen(rv)+1.
func TestStringConcat_SizedBuffer(t *testing.T) {
	ir := generateIR(t, `program p;
var a, b, c: String;
begin
  a := 'hello';
  b := 'world';
  c := a + b;
  WriteLn(c);
end.`)
	// The concat must size its buffer from the operand lengths, not a fixed
	// 512: look for two strlen calls (one per operand) feeding an add, then a
	// malloc of that summed size, then strcpy + strcat.
	if strings.Count(ir, "call i64 @strlen") < 2 {
		t.Errorf("expected emitStringConcat to call strlen on both operands to size the buffer\nIR:\n%s", ir)
	}
	if strings.Contains(ir, "malloc(i64 512)") {
		t.Errorf("string concat must not use a fixed malloc(512) buffer (sized via strlen now)\nIR:\n%s", ir)
	}
	// The malloc size must be a register (computed from the strlen sums), not
	// a constant — i.e. `malloc(i64 %tN)`, not `malloc(i64 <const>)`.
	if !strings.Contains(ir, "call ptr @malloc(i64 %") {
		t.Errorf("expected malloc of a computed size (register), not a constant\nIR:\n%s", ir)
	}
}
