package llvmgen

import (
	"fmt"
	"kylix/ast"
)

// stdlib_regex.go — LLVM IR implementation for the regex module
//
// v0.7.1: pure hand-written character-class predicates. This replaces the
// v0.6.2 POSIX regcomp/regexec implementation, which could not work on
// Windows (UCRT has no <regex.h>) and drifted from the Go backend's RE2
// behavior. All six Is* validators are fixed patterns, so each is emitted as
// a direct byte-scan loop over small character-class helper functions — zero
// external dependencies on any platform (the same IR links everywhere, so
// Windows needs no pcre2 and the three backends agree).
//
// Pattern semantics preserved from the old POSIX ERE definitions:
//
//	IsEmail        ^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$
//	IsURL          ^https?://[^[:space:]/$.?#].[^[:space:]]*$
//	IsNumeric      ^[0-9]+$
//	IsAlpha        ^[a-zA-Z]+$
//	IsAlphaNumeric ^[a-zA-Z0-9]+$
//	IsIP           ^([0-9]{1,3}\.){3}[0-9]{1,3}$   (no 255 value-range check —
//	               999.999.999.999 matches, same as the old pattern)

// emitRegexCall generates a call to a regex validation function (IsEmail,
// IsURL, etc.) and enqueues the function body for later emission.
// All regex functions: (ptr %str) -> i1
func (g *Generator) emitRegexCall(funcName string, args []ast.Expression) (reg, typ string, err error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("regex.%s expects 1 argument, got %d", funcName, len(args))
	}
	// Emit the argument expression (String → ptr).
	argReg, argType, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
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

// emitRegexCharClassHelpers emits the (i8 c) -> i1 character-class helpers
// shared by the validators, once per module. Each is an icmp range/or chain
// — no libc calls.
func (g *Generator) emitRegexCharClassHelpers() {
	if g.stdlibEmitted["regex.classes"] {
		return
	}
	g.stdlibEmitted["regex.classes"] = true

	// [0-9]
	g.emitFn1("define i1 @__kylix_regex_isdigit(i8 %c) {", func() {
		g.retReg(g.cmpRange("48", "57"))
	})
	// [A-Za-z]
	g.emitFn1("define i1 @__kylix_regex_isalpha(i8 %c) {", func() {
		g.retReg(g.or2(g.cmpRange("65", "90"), g.cmpRange("97", "122")))
	})
	// [0-9A-Za-z] — composed from isdigit/isalpha
	g.emitFn1("define i1 @__kylix_regex_isalnum(i8 %c) {", func() {
		g.retReg(g.or2(g.call1("@__kylix_regex_isdigit"), g.call1("@__kylix_regex_isalpha")))
	})
	// POSIX [[:space:]]: space, \t \n \v \f \r
	g.emitFn1("define i1 @__kylix_regex_isspace(i8 %c) {", func() {
		for i, ch := range []string{"32", "9", "10", "11", "12", "13"} {
			g.line(fmt.Sprintf("  %%e%d = icmp eq i8 %%c, %s", i, ch))
		}
		g.line("  %x1 = or i1 %e0, %e1")
		g.line("  %x2 = or i1 %e2, %e3")
		g.line("  %x3 = or i1 %e4, %e5")
		g.line("  %x4 = or i1 %x1, %x2")
		g.retReg(g.or2("%x3", "%x4"))
	})
	// email local-part char: alnum ∪ { '.', '_', '%', '+', '-' }
	g.emitFn1("define i1 @__kylix_regex_isemailchar(i8 %c) {", func() {
		g.line("  %a = call i1 @__kylix_regex_isalnum(i8 %c)")
		g.line("  %m1 = icmp eq i8 %c, 46")
		g.line("  %m2 = icmp eq i8 %c, 95")
		g.line("  %m3 = icmp eq i8 %c, 37")
		g.line("  %m4 = icmp eq i8 %c, 43")
		g.line("  %m5 = icmp eq i8 %c, 45")
		g.line("  %x1 = or i1 %m1, %m2")
		g.line("  %x2 = or i1 %m3, %m4")
		g.line("  %x3 = or i1 %x1, %x2")
		g.line("  %x4 = or i1 %x3, %m5")
		g.retReg(g.or2("%a", "%x4"))
	})
	// host char: alnum ∪ { '.', '-' }
	g.emitFn1("define i1 @__kylix_regex_ishostchar(i8 %c) {", func() {
		g.line("  %a = call i1 @__kylix_regex_isalnum(i8 %c)")
		g.line("  %m1 = icmp eq i8 %c, 46")
		g.line("  %m2 = icmp eq i8 %c, 45")
		g.line("  %x = or i1 %m1, %m2")
		g.retReg(g.or2("%a", "%x"))
	})
	// URL first char after "://": NOT in [[:space:]] ∪ { '/', '$', '.', '?', '#' }
	g.emitFn1("define i1 @__kylix_regex_isurlstart(i8 %c) {", func() {
		g.line("  %s = call i1 @__kylix_regex_isspace(i8 %c)")
		g.line("  %m1 = icmp eq i8 %c, 47")
		g.line("  %m2 = icmp eq i8 %c, 36")
		g.line("  %m3 = icmp eq i8 %c, 46")
		g.line("  %m4 = icmp eq i8 %c, 63")
		g.line("  %m5 = icmp eq i8 %c, 35")
		g.line("  %x1 = or i1 %s, %m1")
		g.line("  %x2 = or i1 %m2, %m3")
		g.line("  %x3 = or i1 %m4, %m5")
		g.line("  %x4 = or i1 %x1, %x2")
		g.line("  %bad = or i1 %x4, %x3")
		g.line("  %ok = xor i1 %bad, true")
		g.retReg("%ok")
	})
}

// emitFn1 wraps a single-entry helper body: emits the define line, the entry
// label, the body, then the closing brace.
func (g *Generator) emitFn1(defLine string, body func()) {
	g.line(defLine)
	g.line("entry:")
	body()
	g.line("}")
	g.line("")
}

// retReg emits a `ret i1 <reg>` tail.
func (g *Generator) retReg(reg string) {
	g.line("  ret i1 " + reg)
}

// cmpRange emits a closed-range membership test and returns the result
// register: (c >= lo) && (c <= hi). LLVM IR has no register aliasing, so the
// caller consumes the returned register directly.
func (g *Generator) cmpRange(lo, hi string) string {
	ge, le, in := g.tmp(), g.tmp(), g.tmp()
	g.line(fmt.Sprintf("  %s = icmp uge i8 %%c, %s", ge, lo))
	g.line(fmt.Sprintf("  %s = icmp ule i8 %%c, %s", le, hi))
	g.line(fmt.Sprintf("  %s = and i1 %s, %s", in, ge, le))
	return in
}

// or2 emits `dst = or i1 a, b` and returns dst.
func (g *Generator) or2(a, b string) string {
	dst := g.tmp()
	g.line(fmt.Sprintf("  %s = or i1 %s, %s", dst, a, b))
	return dst
}

// call1 emits a call to a (i8) -> i1 helper on the current char and returns
// the result register.
func (g *Generator) call1(fn string) string {
	dst := g.tmp()
	g.line(fmt.Sprintf("  %s = call i1 %s(i8 %%c)", dst, fn))
	return dst
}

// emitClassScan emits a whole-string scan: len >= 1 and every byte belongs
// to classFn. Used by IsNumeric / IsAlpha / IsAlphaNumeric.
func (g *Generator) emitClassScan(fnName, classFn string) {
	g.emitRegexCharClassHelpers()
	g.line(fmt.Sprintf("define i1 @__kylix_regex_%s(ptr %%str) {", fnName))
	g.line("entry:")
	lenv := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %%str)", lenv))
	nonempty := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne i64 %s, 0", nonempty, lenv))
	scan, fail := g.label(), g.label()
	chk, cont, pass := g.label(), g.label(), g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", nonempty, scan, fail))
	g.line(scan + ":")
	ivar, nextv := g.tmp(), g.tmp()
	g.line(fmt.Sprintf("  %s = phi i64 [ 0, %%entry ], [ %s, %%%s ]", ivar, nextv, cont))
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", nextv, ivar))
	p := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %%str, i64 %s", p, ivar))
	c := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", c, p))
	ok := g.tmp()
	g.line(fmt.Sprintf("  %s = call i1 %s(i8 %s)", ok, classFn, c))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", ok, chk, fail))
	g.line(chk + ":")
	done := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, %s", done, nextv, lenv))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", done, pass, cont))
	g.line(cont + ":")
	g.line(fmt.Sprintf("  br label %%%s", scan))
	g.line(pass + ":")
	g.line("  ret i1 true")
	g.line(fail + ":")
	g.line("  ret i1 false")
	g.line("}")
	g.line("")
}

func (g *Generator) emitRegexIsNumeric()      { g.emitClassScan("IsNumeric", "@__kylix_regex_isdigit") }
func (g *Generator) emitRegexIsAlpha()        { g.emitClassScan("IsAlpha", "@__kylix_regex_isalpha") }
func (g *Generator) emitRegexIsAlphaNumeric() { g.emitClassScan("IsAlphaNumeric", "@__kylix_regex_isalnum") }

// emitPrefixCmp emits byte-exact comparisons of str[0..len(bytes)) against a
// literal prefix (given as byte values) and returns the AND-accumulated
// result register. Caller guarantees strlen >= len(bytes).
func (g *Generator) emitPrefixCmp(bytes []byte) string {
	andAcc := ""
	for i, b := range bytes {
		p := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %%str, i64 %d", p, i))
		c := g.tmp()
		g.line(fmt.Sprintf("  %s = load i8, ptr %s", c, p))
		eq := g.tmp()
		g.line(fmt.Sprintf("  %s = icmp eq i8 %s, %d", eq, c, b))
		if andAcc == "" {
			andAcc = eq
			continue
		}
		nand := g.tmp()
		g.line(fmt.Sprintf("  %s = and i1 %s, %s", nand, andAcc, eq))
		andAcc = nand
	}
	return andAcc
}

// emitRegexIsURL emits IsURL(str): ^https?://[^[:space:]/$.?#].[^[:space:]]*$
// — match the "https://" (8) or "http://" (7) prefix, require the first char
// after it to be a url-start char, then require every remaining char to be
// non-space.
func (g *Generator) emitRegexIsURL() {
	g.emitRegexCharClassHelpers()
	g.line("define i1 @__kylix_regex_IsURL(ptr %str) {")
	g.line("entry:")
	lenv := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %%str)", lenv))
	ge8 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp uge i64 %s, 8", ge8, lenv))
	ck8, ck7 := g.label(), g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", ge8, ck8, ck7))

	g.line(ck8 + ":")
	all8 := g.emitPrefixCmp([]byte("https://"))
	nphi := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", all8, nphi, ck7))

	g.line(ck7 + ":")
	ge7 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp uge i64 %s, 7", ge7, lenv))
	ck7b, fail := g.label(), g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", ge7, ck7b, fail))
	g.line(ck7b + ":")
	all7 := g.emitPrefixCmp([]byte("http://"))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", all7, nphi, fail))

	g.line(nphi + ":")
	n := g.tmp()
	g.line(fmt.Sprintf("  %s = phi i64 [ 8, %%%s ], [ 7, %%%s ]", n, ck8, ck7b))
	fp := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %%str, i64 %s", fp, n))
	fc := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", fc, fp))
	fok := g.tmp()
	g.line(fmt.Sprintf("  %s = call i1 @__kylix_regex_isurlstart(i8 %s)", fok, fc))
	rstart := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", rstart, n))
	restscan, fail2 := g.label(), g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", fok, restscan, fail2))

	rchk, rcont, pass := g.label(), g.label(), g.label()
	g.line(restscan + ":")
	ri, rnext := g.tmp(), g.tmp()
	g.line(fmt.Sprintf("  %s = phi i64 [ %s, %%%s ], [ %s, %%%s ]", ri, rstart, nphi, rnext, rcont))
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", rnext, ri))
	rp := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %%str, i64 %s", rp, ri))
	rc := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", rc, rp))
	sp := g.tmp()
	g.line(fmt.Sprintf("  %s = call i1 @__kylix_regex_isspace(i8 %s)", sp, rc))
	nsp := g.tmp()
	g.line(fmt.Sprintf("  %s = xor i1 %s, true", nsp, sp))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", nsp, rchk, fail2))
	g.line(rchk + ":")
	rdone := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, %s", rdone, rnext, lenv))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", rdone, pass, rcont))
	g.line(rcont + ":")
	g.line(fmt.Sprintf("  br label %%%s", restscan))

	g.line(pass + ":")
	g.line("  ret i1 true")
	g.line(fail + ":")
	g.line("  ret i1 false")
	g.line(fail2 + ":")
	g.line("  ret i1 false")
	g.line("}")
	g.line("")
}

// emitRegexIsEmail emits IsEmail(str): local@host.tld —
//
//	local: [0, at) non-empty, all email chars
//	host:  (at, end) all host chars; ld = index of last '.' in host;
//	       require ld >= 1 (non-empty part before the final dot) and the
//	       final label [ld+1, hostlen) is 2+ alpha chars.
func (g *Generator) emitRegexIsEmail() {
	g.emitRegexCharClassHelpers()
	g.line("define i1 @__kylix_regex_IsEmail(ptr %str) {")
	g.line("entry:")
	at := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @strchr(ptr %%str, i32 64)", at))
	hasAt := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne ptr %s, null", hasAt, at))
	hostpre, fail := g.label(), g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", hasAt, hostpre, fail))

	g.line(hostpre + ":")
	si := g.tmp()
	g.line(fmt.Sprintf("  %s = ptrtoint ptr %%str to i64", si))
	ai := g.tmp()
	g.line(fmt.Sprintf("  %s = ptrtoint ptr %s to i64", ai, at))
	locLen := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, %s", locLen, ai, si))
	locOk := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne i64 %s, 0", locOk, locLen))
	locscan, loccont, fail2 := g.label(), g.label(), g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", locOk, locscan, fail2))

	g.line(locscan + ":")
	li, lnext := g.tmp(), g.tmp()
	g.line(fmt.Sprintf("  %s = phi i64 [ 0, %%%s ], [ %s, %%%s ]", li, hostpre, lnext, loccont))
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", lnext, li))
	lp := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %%str, i64 %s", lp, li))
	lc := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", lc, lp))
	lok := g.tmp()
	g.line(fmt.Sprintf("  %s = call i1 @__kylix_regex_isemailchar(i8 %s)", lok, lc))
	locchk, fail3 := g.label(), g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", lok, locchk, fail3))
	g.line(locchk + ":")
	ldone := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, %s", ldone, lnext, locLen))
	hostscan := g.label()
	// host starts right after '@' — computed here (locchk) so the hostscan
	// phis sit at the top of their block
	hp := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %s, i64 1", hp, at))
	hlen := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %s)", hlen, hp))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", ldone, hostscan, loccont))
	g.line(loccont + ":")
	g.line(fmt.Sprintf("  br label %%%s", locscan))

	g.line(hostscan + ":")
	hcont := g.label()
	hnewld := g.tmp()
	hi, hnext := g.tmp(), g.tmp()
	g.line(fmt.Sprintf("  %s = phi i64 [ 0, %%%s ], [ %s, %%%s ]", hi, locchk, hnext, hcont))
	hld := g.tmp()
	g.line(fmt.Sprintf("  %s = phi i64 [ -1, %%%s ], [ %s, %%%s ]", hld, locchk, hnewld, hcont))
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", hnext, hi))
	hcp := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %s, i64 %s", hcp, hp, hi))
	hc := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", hc, hcp))
	hok := g.tmp()
	g.line(fmt.Sprintf("  %s = call i1 @__kylix_regex_ishostchar(i8 %s)", hok, hc))
	hchk, fail4 := g.label(), g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", hok, hchk, fail4))

	g.line(hchk + ":")
	isDot := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 46", isDot, hc))
	g.line(fmt.Sprintf("  %s = select i1 %s, i64 %s, i64 %s", hnewld, isDot, hi, hld))
	hdone := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, %s", hdone, hnext, hlen))
	tldpre := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", hdone, tldpre, hcont))
	g.line(hcont + ":")
	g.line(fmt.Sprintf("  br label %%%s", hostscan))

	g.line(tldpre + ":")
	ldGe1 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sge i64 %s, 1", ldGe1, hnewld))
	tldLen := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, %s", tldLen, hlen, hnewld))
	tldM1 := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, 1", tldM1, tldLen))
	tldGe2 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sge i64 %s, 2", tldGe2, tldM1))
	shapeOk := g.tmp()
	g.line(fmt.Sprintf("  %s = and i1 %s, %s", shapeOk, ldGe1, tldGe2))
	tldscan, tcont, fail5 := g.label(), g.label(), g.label()
	ti0 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", ti0, hnewld))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", shapeOk, tldscan, fail5))

	g.line(tldscan + ":")
	ti, tnext := g.tmp(), g.tmp()
	g.line(fmt.Sprintf("  %s = phi i64 [ %s, %%%s ], [ %s, %%%s ]", ti, ti0, tldpre, tnext, tcont))
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", tnext, ti))
	tp := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %s, i64 %s", tp, hp, ti))
	tc := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", tc, tp))
	tok := g.tmp()
	g.line(fmt.Sprintf("  %s = call i1 @__kylix_regex_isalpha(i8 %s)", tok, tc))
	tchk, fail6 := g.label(), g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", tok, tchk, fail6))
	g.line(tchk + ":")
	tdone := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, %s", tdone, tnext, hlen))
	pass := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", tdone, pass, tcont))
	g.line(tcont + ":")
	g.line(fmt.Sprintf("  br label %%%s", tldscan))

	g.line(pass + ":")
	g.line("  ret i1 true")
	for _, f := range []string{fail, fail2, fail3, fail4, fail5, fail6} {
		g.line(f + ":")
		g.line("  ret i1 false")
	}
	g.line("}")
	g.line("")
}

// emitRegexIsIP emits IsIP(str): ^([0-9]{1,3}\.){3}[0-9]{1,3}$ — four digit
// groups of length 1..3 separated by exactly three dots. A single
// state-machine loop tracks (position, group number, digits-in-group).
func (g *Generator) emitRegexIsIP() {
	g.emitRegexCharClassHelpers()
	g.line("define i1 @__kylix_regex_IsIP(ptr %str) {")
	g.line("entry:")
	lenv := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %%str)", lenv))
	loop, fail := g.label(), g.label()
	ni, ng, nd := g.tmp(), g.tmp(), g.tmp()
	g.line(fmt.Sprintf("  br label %%%s", loop))

	g.line(loop + ":")
	iv, gv, dv := g.tmp(), g.tmp(), g.tmp()
	g.line(fmt.Sprintf("  %s = phi i64 [ 0, %%entry ], [ %s, %%step ]", iv, ni))
	g.line(fmt.Sprintf("  %s = phi i64 [ 1, %%entry ], [ %s, %%step ]", gv, ng))
	g.line(fmt.Sprintf("  %s = phi i64 [ 0, %%entry ], [ %s, %%step ]", dv, nd))
	iend := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, %s", iend, iv, lenv))
	final, body := g.label(), g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", iend, final, body))

	g.line(body + ":")
	bp := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %%str, i64 %s", bp, iv))
	c := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", c, bp))
	isDig := g.tmp()
	g.line(fmt.Sprintf("  %s = call i1 @__kylix_regex_isdigit(i8 %s)", isDig, c))
	dig, dotchk := g.label(), g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isDig, dig, dotchk))

	g.line(dig + ":")
	d1 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", d1, dv))
	dOk := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ule i64 %s, 3", dOk, d1))
	dig2, fail2 := g.label(), g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", dOk, dig2, fail2))
	g.line(dig2 + ":")
	di1 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", di1, iv))
	g.line("  br label %step")

	g.line(dotchk + ":")
	isDot := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 46", isDot, c))
	dot, fail3 := g.label(), g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isDot, dot, fail3))

	g.line(dot + ":")
	dNe0 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne i64 %s, 0", dNe0, dv))
	gLt4 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne i64 %s, 4", gLt4, gv))
	dAndG := g.tmp()
	g.line(fmt.Sprintf("  %s = and i1 %s, %s", dAndG, dNe0, gLt4))
	dot2, fail4 := g.label(), g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", dAndG, dot2, fail4))
	g.line(dot2 + ":")
	doi := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", doi, iv))
	ng1 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", ng1, gv))
	g.line("  br label %step")

	g.line("step:")
	g.line(fmt.Sprintf("  %s = phi i64 [ %s, %%%s ], [ %s, %%%s ]", ni, di1, dig2, doi, dot2))
	g.line(fmt.Sprintf("  %s = phi i64 [ %s, %%%s ], [ 0, %%%s ]", nd, d1, dig2, dot2))
	g.line(fmt.Sprintf("  %s = phi i64 [ %s, %%%s ], [ %s, %%%s ]", ng, gv, dig2, ng1, dot2))
	g.line(fmt.Sprintf("  br label %%%s", loop))

	g.line(final + ":")
	dGe1 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne i64 %s, 0", dGe1, nd))
	gEq4 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, 4", gEq4, ng))
	okRes := g.tmp()
	g.line(fmt.Sprintf("  %s = and i1 %s, %s", okRes, dGe1, gEq4))
	pass := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", okRes, pass, fail))

	g.line(pass + ":")
	g.line("  ret i1 true")
	for _, f := range []string{fail, fail2, fail3, fail4} {
		g.line(f + ":")
		g.line("  ret i1 false")
	}
	g.line("}")
	g.line("")
}
