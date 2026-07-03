package llvmgen

import (
	"fmt"
	"kylix/ast"
)

// stdlib_regex.go — LLVM IR implementation for regex module
//
// Uses POSIX regex (regcomp/regexec from libc) to implement pattern validation
// helpers. Each IsXxx function compiles a hardcoded pattern, matches the input
// string, frees the regex, and returns i1 (true/false).
//
// POSIX pattern differences from Go regexp:
//   - \d → [0-9]
//   - \s → [[:space:]]
//   - No native support for PCRE features; use POSIX Extended (REG_EXTENDED)

// emitRegexCall generates a call to a regex validation function (IsEmail, IsURL,
// etc.) and enqueues the function body for later emission if not already done.
// All regex functions: (ptr %str) -> i1
func (g *Generator) emitRegexCall(funcName string, args []ast.Expression) (reg, typ string, err error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("regex.%s expects 1 argument, got %d", funcName, len(args))
	}
	// Emit the argument expression (should be String → ptr)
	argReg, argType, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	// Coerce to ptr if needed (best-effort; type mismatches caught by llc)
	_ = argType

	fn := fmt.Sprintf("@__kylix_regex_%s", funcName)
	key := "regex." + funcName
	if !g.stdlibEmitted[key] {
		g.stdlibEmitted[key] = true
		g.stdlibQueue = append(g.stdlibQueue, stdlibFunc{
			module: "regex", name: funcName, key: key, argCount: 0,
		})
	}
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i1 %s(ptr %s)", r, fn, argReg))
	return r, "i1", nil
}

// regexPatterns maps function names to POSIX-compatible regex patterns.
// POSIX ERE (Extended Regular Expressions) used via REG_EXTENDED flag.
var regexPatterns = map[string]string{
	"IsEmail":        `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`,
	"IsURL":          `^https?://[^[:space:]/$.?#].[^[:space:]]*$`,
	"IsNumeric":      `^[0-9]+$`,
	"IsAlpha":        `^[a-zA-Z]+$`,
	"IsAlphaNumeric": `^[a-zA-Z0-9]+$`,
	"IsIP":           `^([0-9]{1,3}\.){3}[0-9]{1,3}$`,
}

// emitRegexBody dispatches the body emitter for a regex function. Called by
// emitPendingStdlib for each queued regex function.
func (g *Generator) emitRegexBody(funcName string) {
	switch funcName {
	case "IsEmail":
		g.emitRegexIsEmail()
	case "IsURL":
		g.emitRegexIsURL()
	case "IsNumeric":
		g.emitRegexIsNumeric()
	case "IsAlpha":
		g.emitRegexIsAlpha()
	case "IsAlphaNumeric":
		g.emitRegexIsAlphaNumeric()
	case "IsIP":
		g.emitRegexIsIP()
	default:
		g.line(fmt.Sprintf("; ERROR: unsupported regex function: %s", funcName))
	}
}

// emitRegexIsEmail emits the LLVM IR body for IsEmail(str) -> i1.
func (g *Generator) emitRegexIsEmail() {
	g.emitRegexValidator("IsEmail", regexPatterns["IsEmail"])
}

func (g *Generator) emitRegexIsURL() {
	g.emitRegexValidator("IsURL", regexPatterns["IsURL"])
}

func (g *Generator) emitRegexIsNumeric() {
	g.emitRegexValidator("IsNumeric", regexPatterns["IsNumeric"])
}

func (g *Generator) emitRegexIsAlpha() {
	g.emitRegexValidator("IsAlpha", regexPatterns["IsAlpha"])
}

func (g *Generator) emitRegexIsAlphaNumeric() {
	g.emitRegexValidator("IsAlphaNumeric", regexPatterns["IsAlphaNumeric"])
}

func (g *Generator) emitRegexIsIP() {
	g.emitRegexValidator("IsIP", regexPatterns["IsIP"])
}

// emitRegexValidator generates a validation function body:
//   1. alloca regex_t (opaque struct, ~64 bytes on most platforms)
//   2. addString(pattern) → getelementptr
//   3. regcomp(&regex, pattern_ptr, REG_EXTENDED | REG_NOSUB)
//      - REG_EXTENDED = 1 (use ERE)
//      - REG_NOSUB = 8 (don't fill in pmatch[], faster)
//   4. if regcomp != 0 → ret false (pattern compile error)
//   5. regexec(&regex, str, 0, null, 0)
//   6. regfree(&regex)
//   7. ret (regexec == 0) as i1
func (g *Generator) emitRegexValidator(funcName, pattern string) {
	g.line(fmt.Sprintf("define i1 @__kylix_regex_%s(ptr %%str) {", funcName))
	g.line("entry:")

	// alloca regex_t (opaque %struct.regex_t, assume 64 bytes)
	regexSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [64 x i8], align 8", regexSlot))

	// Get pattern string pointer (after entry:, so ptrTo writes inside function)
	patStr := g.addString(pattern)
	patPtr := g.ptrTo(patStr, len(pattern)+1)

	// regcomp: int regcomp(regex_t *preg, const char *pattern, int cflags)
	// cflags = REG_EXTENDED(1) | REG_NOSUB(8) = 9
	compResult := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @regcomp(ptr %s, ptr %s, i32 9)", compResult, regexSlot, patPtr))

	// if regcomp != 0 → pattern error, ret false
	compOk := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 0", compOk, compResult))
	okBlk := g.label()
	failBlk := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", compOk, okBlk, failBlk))

	// fail block: ret false
	g.line(failBlk + ":")
	g.line("  ret i1 false")

	// ok block: regexec
	g.line(okBlk + ":")
	// int regexec(const regex_t *preg, const char *string, size_t nmatch,
	//             regmatch_t pmatch[], int eflags)
	// nmatch=0, pmatch=null (REG_NOSUB), eflags=0
	execResult := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @regexec(ptr %s, ptr %%str, i64 0, ptr null, i32 0)", execResult, regexSlot))

	// regfree (no return value)
	g.line(fmt.Sprintf("  call void @regfree(ptr %s)", regexSlot))

	// ret (regexec == 0) as i1
	match := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 0", match, execResult))
	g.line(fmt.Sprintf("  ret i1 %s", match))
	g.line("}")
	g.line("")
}
