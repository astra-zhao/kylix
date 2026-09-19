package llvmgen

import (
	"fmt"

	"kylix/ast"
)

// stdlib_boot_session.go — v0.9.0 P1.7c: server-side sessions for the LLVM
// KylixBoot runtime, mirroring pkg/boot/session.go.
//
// Storage: one module-global outer htab (@__kylix_boot_sessions) maps SID →
// per-session htab pointer (opaque value slot, cache-TTL-table precedent).
// Each per-session htab holds user key/value string pairs plus reserved
// double-underscore keys the middleware owns:
//
//	__sid  — strdup'd session ID (cookie buffer is request-arena scoped)
//	__exp  — expiry, decimal milliseconds (sliding window)
//	__rem  — present ⇒ remember-me (30-day TTL + Max-Age cookie attr)
//
// TRequest handle gains two fields (bootRequestSize 48 → 64):
//
//	48 ptr session   56 i64 flags   ; bit0 new · bit1 cookie-dirty · bit2 destroyed
//
// The middleware is BootRun-embedded (LLVM has no middleware chain):
//
//	@__kylix_boot_session_resolve(req, headers)  before the handler
//	@__kylix_boot_session_finish(req, res)       after it — flags-driven
//	                                             Set-Cookie, Go parity
const (
	bootSessionCookieName = "KYLIX_SID"
	bootSessionKeySID     = "__sid"
	bootSessionKeyExp     = "__exp"
	bootSessionKeyRem     = "__rem"
	bootSessionCSRFKey    = "__csrf"

	bootSessionsGlobal = "@__kylix_boot_sessions"

	// v0.9.0 P1.7 CSRF (mirrors pkg/boot/csrf.go): opt-in via BootUseCSRF —
	// the gate lives in BootRun between session resolve and handler dispatch.
	bootCsrfEnabledGlobal = "@__kylix_boot_csrf_enabled"
	bootCsrfHeaderNeedle  = "X-CSRF-Token: "
	bootCsrfFormField     = "_csrf"
	bootCsrfMissingMsg    = "CSRF token missing: render req.CSRFToken() into the form first"
	bootCsrfMismatchMsg   = "CSRF token mismatch"

	// Sliding TTLs (ms), matching Go DefaultSessionTTL/DefaultRememberTTL.
	bootSessionTTLms  = int64(24 * 60 * 60 * 1000)
	bootRememberTTLms = int64(30 * 24 * 60 * 60 * 1000)
	// Cookie Max-Age seconds for remember-me (Go DefaultRememberTTL).
	bootRememberSeconds = 2592000

	// req[56] flag bits — keep in sync with the doc comment above.
	bootSessFlagNew       = 1
	bootSessFlagDirty     = 2
	bootSessFlagDestroyed = 4
)

// bootDeclareSessionsGlobal emits the outer session-table global once.
// v0.10.0 P2: buffered into pendingModuleGlobals instead of written with
// g.line() — req.SessionRegenerate/SessionDestroy now reference this global
// from user-function bodies, and a direct write would land the global line
// inside the current define (invalid IR). Flushed at top level by
// emitPendingStdlib; LLVM allows forward references to globals.
func (g *Generator) bootDeclareSessionsGlobal() {
	if g.bootSessionsDeclared {
		return
	}
	g.bootSessionsDeclared = true
	g.pendingModuleGlobals = append(g.pendingModuleGlobals,
		fmt.Sprintf("%s = global ptr null", bootSessionsGlobal))
}

// emitBootSessionResolveBody — void @__kylix_boot_session_resolve(ptr %req,
// ptr %headers): the session half of the Sessions() middleware. Resolves the
// KYLIX_SID cookie to a live per-session htab (creating + minting a fresh
// SID when the cookie is absent/unknown/expired), applies the sliding expiry,
// and stores sess + flags on the request handle. Allocas carry values across
// blocks (no phis); htab_put strdups keys, but VALUES are stored verbatim —
// every value put here is strdup'd first because cookie/rand buffers are
// request-scoped.
func (g *Generator) emitBootSessionResolveBody() {
	g.needHashtab = true
	g.needArena = true
	g.enqueueStdlib("cache", "now_ms", "now_ms", 0)
	g.enqueueStdlib("boot", "randhex", "randhex", 0)
	g.bootDeclareSessionsGlobal()
	c := func(format string, args ...interface{}) { g.line(fmt.Sprintf(format, args...)) }
	c("define void @__kylix_boot_session_resolve(ptr %%req, ptr %%headers) {")
	c("entry:")
	tSlot := g.tmp()
	c("  %s = alloca ptr, align 8", tSlot)
	sidSlot := g.tmp()
	c("  %s = alloca ptr, align 8", sidSlot)
	sessSlot := g.tmp()
	c("  %s = alloca ptr, align 8", sessSlot)
	flagsSlot := g.tmp()
	c("  %s = alloca i64, align 8", flagsSlot)
	c("  store i64 0, ptr %s", flagsSlot)
	// expKey's GEP is minted in entry: it is used in renewLbl, which the
	// cookie-absent path (entry → createLbl) reaches WITHOUT passing through
	// expChkLbl — defining it in expChkLbl was an SSA domination violation
	// that llc's -disable-verify masked (the minted session then got its
	// __exp key written from a stale register on that path).
	expKey := g.ptrTo(g.addString(bootSessionKeyExp), len(bootSessionKeyExp)+1)
	// Lazy-init the outer sessions table.
	t0 := g.bootLoadPtr(bootSessionsGlobal)
	c("  store ptr %s, ptr %s", t0, tSlot)
	tNull := g.tmp()
	c("  %s = icmp eq ptr %s, null", tNull, t0)
	tinit := g.label()
	tready := g.label()
	c("  br i1 %s, label %%%s, label %%%s", tNull, tinit, tready)
	c("%s:", tinit)
	nt := g.tmp()
	c("  %s = call ptr @__kylix_htab_new()", nt)
	c("  store ptr %s, ptr %s", nt, bootSessionsGlobal)
	c("  store ptr %s, ptr %s", nt, tSlot)
	c("  br label %%%s", tready)
	// Read the session cookie (null when absent — htab-free path).
	c("%s:", tready)
	sidName := g.ptrTo(g.addString(bootSessionCookieName), len(bootSessionCookieName)+1)
	g.enqueueStdlib("boot", "cookieget", "cookieget", 0)
	sid0 := g.tmp()
	c("  %s = call ptr @__kylix_boot_cookie_get(ptr %%headers, ptr %s)", sid0, sidName)
	c("  store ptr %s, ptr %s", sid0, sidSlot)
	sidNull := g.tmp()
	c("  %s = icmp eq ptr %s, null", sidNull, sid0)
	createLbl := g.label()
	lookupLbl := g.label()
	c("  br i1 %s, label %%%s, label %%%s", sidNull, createLbl, lookupLbl)
	c("%s:", lookupLbl)
	t1 := g.tmp()
	c("  %s = load ptr, ptr %s", t1, tSlot)
	sess1 := g.tmp()
	c("  %s = call ptr @__kylix_htab_get(ptr %s, ptr %s)", sess1, t1, sid0)
	c("  store ptr %s, ptr %s", sess1, sessSlot)
	sessNull := g.tmp()
	c("  %s = icmp eq ptr %s, null", sessNull, sess1)
	mkNewLbl := g.label()
	expChkLbl := g.label()
	c("  br i1 %s, label %%%s, label %%%s", sessNull, mkNewLbl, expChkLbl)
	// Unknown SID — mint a fresh one (Go parity: store.Get miss → Create),
	// which also closes session fixation by client-chosen IDs.
	c("%s:", mkNewLbl)
	c("  store ptr null, ptr %s", sidSlot)
	c("  br label %%%s", createLbl)
	// Known session — check __exp (absent ⇒ treat as live, renew below).
	c("%s:", expChkLbl)
	sess2 := g.tmp()
	c("  %s = load ptr, ptr %s", sess2, sessSlot)
	expStr := g.tmp()
	c("  %s = call ptr @__kylix_htab_get(ptr %s, ptr %s)", expStr, sess2, expKey)
	expNull := g.tmp()
	c("  %s = icmp eq ptr %s, null", expNull, expStr)
	renewLbl := g.label()
	expParseLbl := g.label()
	c("  br i1 %s, label %%%s, label %%%s", expNull, renewLbl, expParseLbl)
	c("%s:", expParseLbl)
	exp := g.tmp()
	c("  %s = call i64 @atoll(ptr %s)", exp, expStr)
	now1 := g.tmp()
	c("  %s = call i64 @__kylix_now_ms()", now1)
	expired := g.tmp()
	c("  %s = icmp sgt i64 %s, %s", expired, now1, exp)
	staleLbl := g.label()
	c("  br i1 %s, label %%%s, label %%%s", expired, staleLbl, renewLbl)
	// Expired — drop the stored session, then create a fresh one (Go mints a
	// new random ID; the stale cookie value is never reused).
	c("%s:", staleLbl)
	t2 := g.tmp()
	c("  %s = load ptr, ptr %s", t2, tSlot)
	sidStale := g.tmp()
	c("  %s = load ptr, ptr %s", sidStale, sidSlot)
	c("  call void @__kylix_htab_del(ptr %s, ptr %s)", t2, sidStale)
	c("  store ptr null, ptr %s", sidSlot)
	c("  br label %%%s", createLbl)
	// Create path — sidSlot is null (no cookie / unknown / expired), so a
	// fresh SID is minted and the Set-Cookie on this response re-binds the
	// browser.
	c("%s:", createLbl)
	f0 := g.tmp()
	c("  %s = load i64, ptr %s", f0, flagsSlot)
	f1 := g.tmp()
	c("  %s = or i64 %s, %d", f1, f0, bootSessFlagNew)
	c("  store i64 %s, ptr %s", f1, flagsSlot)
	sess := g.tmp()
	c("  %s = call ptr @__kylix_htab_new()", sess)
	c("  store ptr %s, ptr %s", sess, sessSlot)
	curSid := g.tmp()
	c("  %s = load ptr, ptr %s", curSid, sidSlot)
	curNull := g.tmp()
	c("  %s = icmp eq ptr %s, null", curNull, curSid)
	mintLbl := g.label()
	haveSidLbl := g.label()
	c("  br i1 %s, label %%%s, label %%%s", curNull, mintLbl, haveSidLbl)
	c("%s:", mintLbl)
	ns := g.tmp()
	c("  %s = call ptr @__kylix_boot_rand_hex(i64 32)", ns)
	c("  store ptr %s, ptr %s", ns, sidSlot)
	c("  br label %%%s", haveSidLbl)
	// Register the session in the outer table (key strdup'd by htab_put) and
	// persist __sid — the value must outlive this request's buffers.
	c("%s:", haveSidLbl)
	fs := g.tmp()
	c("  %s = load ptr, ptr %s", fs, sidSlot)
	t3 := g.tmp()
	c("  %s = load ptr, ptr %s", t3, tSlot)
	c("  call void @__kylix_htab_put(ptr %s, ptr %s, ptr %s)", t3, fs, sess)
	sidKey := g.ptrTo(g.addString(bootSessionKeySID), len(bootSessionKeySID)+1)
	dupSid := g.tmp()
	c("  %s = call ptr @__kylix_htab_strdup(ptr %s)", dupSid, fs)
	c("  call void @__kylix_htab_put(ptr %s, ptr %s, ptr %s)", sess, sidKey, dupSid)
	c("  br label %%%s", renewLbl)
	// Sliding expiry — shared tail for the create and renew paths.
	c("%s:", renewLbl)
	sess3 := g.tmp()
	c("  %s = load ptr, ptr %s", sess3, sessSlot)
	remKey := g.ptrTo(g.addString(bootSessionKeyRem), len(bootSessionKeyRem)+1)
	remStr := g.tmp()
	c("  %s = call ptr @__kylix_htab_get(ptr %s, ptr %s)", remStr, sess3, remKey)
	isRem := g.tmp()
	c("  %s = icmp ne ptr %s, null", isRem, remStr)
	ttl := g.tmp()
	c("  %s = select i1 %s, i64 %d, i64 %d", ttl, isRem, bootRememberTTLms, bootSessionTTLms)
	now2 := g.tmp()
	c("  %s = call i64 @__kylix_now_ms()", now2)
	newExp := g.tmp()
	c("  %s = add i64 %s, %s", newExp, now2, ttl)
	expBuf := g.bootIntToStr(newExp)
	dupExp := g.tmp()
	c("  %s = call ptr @__kylix_htab_strdup(ptr %s)", dupExp, expBuf)
	c("  call void @__kylix_htab_put(ptr %s, ptr %s, ptr %s)", sess3, expKey, dupExp)
	// Publish on the request handle: sess at 48, flags at 56.
	c("  store ptr %s, ptr %s", sess3, g.bootReqField("%req", 48))
	flags := g.tmp()
	c("  %s = load i64, ptr %s", flags, flagsSlot)
	c("  store i64 %s, ptr %s", flags, g.bootReqField("%req", 56))
	c("  ret void")
	c("}")
	c("")
}

// emitBootSessionFinishBody — void @__kylix_boot_session_finish(ptr %req,
// ptr %res): the response half of the middleware. Flags ≠ 0 ⇒ append the
// session cookie to the response's cookie slot: destroyed clears the cookie
// (Max-Age=0), remember adds Max-Age=2592000, plain re-send is Path=/;
// HttpOnly — the exact lines pkg/boot.writeSessionCookie produces.
func (g *Generator) emitBootSessionFinishBody() {
	g.needHashtab = true
	g.needArena = true
	c := func(format string, args ...interface{}) { g.line(fmt.Sprintf(format, args...)) }
	c("define void @__kylix_boot_session_finish(ptr %%req, ptr %%res) {")
	c("entry:")
	flags := g.tmp()
	c("  %s = load i64, ptr %s", flags, g.bootReqField("%req", 56))
	need := g.tmp()
	c("  %s = icmp ne i64 %s, 0", need, flags)
	emitLbl := g.label()
	doneLbl := g.label()
	c("  br i1 %s, label %%%s, label %%%s", need, emitLbl, doneLbl)
	c("%s:", doneLbl)
	c("  ret void")
	c("%s:", emitLbl)
	df := g.tmp()
	c("  %s = and i64 %s, %d", df, flags, bootSessFlagDestroyed)
	destroyed := g.tmp()
	c("  %s = icmp ne i64 %s, 0", destroyed, df)
	clearLbl := g.label()
	setLbl := g.label()
	c("  br i1 %s, label %%%s, label %%%s", destroyed, clearLbl, setLbl)
	// Destroyed: clear-cookie line is a constant — append verbatim.
	c("%s:", clearLbl)
	clearLine := "Set-Cookie: " + bootSessionCookieName + "=; Path=/; Max-Age=0; HttpOnly\r\n"
	g.emitBootAppendToSlot("%res", 24, g.ptrTo(g.addString(clearLine), len(clearLine)+1), fmt.Sprintf("%d", len(clearLine)+1))
	c("  ret void")
	// Set path: assemble "Set-Cookie: KYLIX_SID=<sid><attrs>\r\n" — attrs
	// selected on the persisted __rem key so a MarkRemember inside this
	// request's handler takes effect on the very response.
	c("%s:", setLbl)
	sess := g.tmp()
	c("  %s = load ptr, ptr %s", sess, g.bootReqField("%req", 48))
	sessNull := g.tmp()
	c("  %s = icmp eq ptr %s, null", sessNull, sess)
	noSessLbl := g.label()
	haveSessLbl := g.label()
	c("  br i1 %s, label %%%s, label %%%s", sessNull, noSessLbl, haveSessLbl)
	c("%s:", noSessLbl)
	c("  ret void")
	c("%s:", haveSessLbl)
	sidKey := g.ptrTo(g.addString(bootSessionKeySID), len(bootSessionKeySID)+1)
	sid := g.tmp()
	c("  %s = call ptr @__kylix_htab_get(ptr %s, ptr %s)", sid, sess, sidKey)
	remKey := g.ptrTo(g.addString(bootSessionKeyRem), len(bootSessionKeyRem)+1)
	remStr := g.tmp()
	c("  %s = call ptr @__kylix_htab_get(ptr %s, ptr %s)", remStr, sess, remKey)
	isRem := g.tmp()
	c("  %s = icmp ne ptr %s, null", isRem, remStr)
	attrsPlain := "; Path=/; HttpOnly\r\n"
	attrsRem := "; Path=/; HttpOnly; Max-Age=" + fmt.Sprintf("%d", bootRememberSeconds) + "\r\n"
	attrs := g.tmp()
	c("  %s = select i1 %s, ptr %s, ptr %s", attrs, isRem,
		g.ptrTo(g.addString(attrsRem), len(attrsRem)+1),
		g.ptrTo(g.addString(attrsPlain), len(attrsPlain)+1))
	// Buffer cap: prefix 22 + sid (64 hex) + attrs (≤37) + slack.
	prefix := "Set-Cookie: " + bootSessionCookieName + "="
	bufCap := g.tmp()
	sidLen := g.tmp()
	c("  %s = call i64 @strlen(ptr %s)", sidLen, sid)
	c("  %s = add i64 %s, %d", bufCap, sidLen, len(prefix)+len(attrsRem)+16)
	buf := g.tmp()
	c("  %s = call ptr @__kylix_arena_alloc(i64 %s)", buf, bufCap)
	c("  store i8 0, ptr %s", buf)
	c("  call ptr @strcpy(ptr %s, ptr %s)", buf, g.ptrTo(g.addString(prefix), len(prefix)+1))
	c("  call ptr @strcat(ptr %s, ptr %s)", buf, sid)
	c("  call ptr @strcat(ptr %s, ptr %s)", buf, attrs)
	g.emitBootAppendToSlot("%res", 24, buf, bufCap)
	c("  ret void")
	c("}")
	c("")
}

// emitBootRandHexBody — ptr @__kylix_boot_rand_hex(i64 %nbytes): 2*nbytes hex
// chars + NUL in a malloc'd buffer (Go randomID: crypto/rand → hex). Unix
// reads /dev/urandom; a failed open falls back to a time-derived fill (and
// Windows, which has no urandom, uses it outright — debt: pseudo-random).
// nbytes clamps to 64 (the fixed byte scratch buffer).
func (g *Generator) emitBootRandHexBody() {
	c := func(format string, args ...interface{}) { g.line(fmt.Sprintf(format, args...)) }
	c("define ptr @__kylix_boot_rand_hex(i64 %%nbytes) {")
	c("entry:")
	// Clamp nbytes to 64.
	tooBig := g.tmp()
	c("  %s = icmp sgt i64 %%nbytes, 64", tooBig)
	n := g.tmp()
	c("  %s = select i1 %s, i64 64, i64 %%nbytes", n, tooBig)
	// Scratch byte buffer (alloca is fine: 64B, and the buffer is consumed
	// before return).
	byteBuf := g.tmp()
	c("  %s = alloca [64 x i8], align 1", byteBuf)
	byteBuf0 := g.tmp()
	c("  %s = getelementptr inbounds [64 x i8], ptr %s, i64 0, i64 0", byteBuf0, byteBuf)
	useURandom := g.targetOS != "windows"
	hexEncLbl := g.label()
	if useURandom {
		urandomPath := g.ptrTo(g.addString("/dev/urandom"), 13)
		urandomMode := g.ptrTo(g.addString("rb"), 3)
		fb := g.tmp()
		c("  %s = call ptr @fopen(ptr %s, ptr %s)", fb, urandomPath, urandomMode)
		fbNull := g.tmp()
		c("  %s = icmp eq ptr %s, null", fbNull, fb)
		rdLbl := g.label()
		timeFillLbl := g.label()
		c("  br i1 %s, label %%%s, label %%%s", fbNull, timeFillLbl, rdLbl)
		c("%s:", rdLbl)
		got := g.tmp()
		c("  %s = call i64 @fread(ptr %s, i64 1, i64 %s, ptr %s)", got, byteBuf0, n, fb)
		c("  call i32 @fclose(ptr %s)", fb)
		full := g.tmp()
		c("  %s = icmp eq i64 %s, %s", full, got, n)
		c("  br i1 %s, label %%%s, label %%%s", full, hexEncLbl, timeFillLbl)
		// Fallback (open failed / short read) + Windows: time-derived bytes.
		c("%s:", timeFillLbl)
		g.emitBootRandHexTimeFill(c, byteBuf0, n)
		c("  br label %%%s", hexEncLbl)
	} else {
		g.emitBootRandHexTimeFill(c, byteBuf0, n)
		c("  br label %%%s", hexEncLbl)
	}
	// Hex-encode into a fresh malloc'd buffer.
	c("%s:", hexEncLbl)
	out := g.tmp()
	twice := g.tmp()
	c("  %s = shl i64 %s, 1", twice, n)
	outLen := g.tmp()
	c("  %s = add i64 %s, 1", outLen, twice)
	c("  %s = %s", out, g.mallocCall(outLen))
	hexTable := g.ptrTo(g.addString("0123456789abcdef"), 17)
	iSlot := g.tmp()
	c("  %s = alloca i64, align 8", iSlot)
	c("  store i64 0, ptr %s", iSlot)
	condLbl := g.label()
	bodyLbl := g.label()
	doneLbl := g.label()
	c("  br label %%%s", condLbl)
	c("%s:", condLbl)
	i := g.tmp()
	c("  %s = load i64, ptr %s", i, iSlot)
	more := g.tmp()
	c("  %s = icmp slt i64 %s, %s", more, i, n)
	c("  br i1 %s, label %%%s, label %%%s", more, bodyLbl, doneLbl)
	c("%s:", bodyLbl)
	bp := g.tmp()
	c("  %s = getelementptr inbounds i8, ptr %s, i64 %s", bp, byteBuf0, i)
	b8 := g.tmp()
	c("  %s = load i8, ptr %s", b8, bp)
	b32 := g.tmp()
	c("  %s = zext i8 %s to i32", b32, b8)
	hi := g.tmp()
	c("  %s = lshr i32 %s, 4", hi, b32)
	hi64 := g.tmp()
	c("  %s = zext i32 %s to i64", hi64, hi)
	lo := g.tmp()
	c("  %s = and i32 %s, 15", lo, b32)
	lo64 := g.tmp()
	c("  %s = zext i32 %s to i64", lo64, lo)
	twiceI := g.tmp()
	c("  %s = shl i64 %s, 1", twiceI, i)
	o0 := g.tmp()
	c("  %s = getelementptr inbounds i8, ptr %s, i64 %s", o0, out, twiceI)
	o1 := g.tmp()
	c("  %s = add i64 %s, 1", o1, twiceI)
	o1p := g.tmp()
	c("  %s = getelementptr inbounds i8, ptr %s, i64 %s", o1p, out, o1)
	hc0 := g.tmp()
	c("  %s = getelementptr inbounds i8, ptr %s, i64 %s", hc0, hexTable, hi64)
	ch0 := g.tmp()
	c("  %s = load i8, ptr %s", ch0, hc0)
	c("  store i8 %s, ptr %s", ch0, o0)
	hc1 := g.tmp()
	c("  %s = getelementptr inbounds i8, ptr %s, i64 %s", hc1, hexTable, lo64)
	ch1 := g.tmp()
	c("  %s = load i8, ptr %s", ch1, hc1)
	c("  store i8 %s, ptr %s", ch1, o1p)
	iNext := g.tmp()
	c("  %s = add i64 %s, 1", iNext, i)
	c("  store i64 %s, ptr %s", iNext, iSlot)
	c("  br label %%%s", condLbl)
	c("%s:", doneLbl)
	end := g.tmp()
	c("  %s = getelementptr inbounds i8, ptr %s, i64 %s", end, out, twice)
	c("  store i8 0, ptr %s", end)
	c("  ret ptr %s", out)
	c("}")
	c("")
}

// emitBootRandHexTimeFill fills byteBuf with time-derived pseudo bytes — the
// fallback when /dev/urandom is unavailable (or on Windows). seed = now_ms;
// byte i = seed >> ((i&7)*8). Not cryptographic — debt note above.
func (g *Generator) emitBootRandHexTimeFill(c func(string, ...interface{}), byteBuf0, n string) {
	g.enqueueStdlib("cache", "now_ms", "now_ms", 0)
	seed := g.tmp()
	c("  %s = call i64 @__kylix_now_ms()", seed)
	iSlot := g.tmp()
	c("  %s = alloca i64, align 8", iSlot)
	c("  store i64 0, ptr %s", iSlot)
	condLbl := g.label()
	bodyLbl := g.label()
	doneLbl := g.label()
	c("  br label %%%s", condLbl)
	c("%s:", condLbl)
	i := g.tmp()
	c("  %s = load i64, ptr %s", i, iSlot)
	more := g.tmp()
	c("  %s = icmp slt i64 %s, %s", more, i, n)
	c("  br i1 %s, label %%%s, label %%%s", more, bodyLbl, doneLbl)
	c("%s:", bodyLbl)
	iAnd7 := g.tmp()
	c("  %s = and i64 %s, 7", iAnd7, i)
	shift := g.tmp()
	c("  %s = shl i64 %s, 3", shift, iAnd7)
	shr := g.tmp()
	c("  %s = lshr i64 %s, %s", shr, seed, shift)
	b := g.tmp()
	c("  %s = trunc i64 %s to i8", b, shr)
	bp := g.tmp()
	c("  %s = getelementptr inbounds i8, ptr %s, i64 %s", bp, byteBuf0, i)
	c("  store i8 %s, ptr %s", b, bp)
	iNext := g.tmp()
	c("  %s = add i64 %s, 1", iNext, i)
	c("  store i64 %s, ptr %s", iNext, iSlot)
	c("  br label %%%s", condLbl)
	c("%s:", doneLbl)
}

// ---- req.Session* accessors (called inline at the handler's call site) ----
//
// The session pointer at req[48] is guaranteed non-null: BootRun runs
// @__kylix_boot_session_resolve before dispatching to every handler, and
// resolve always creates a session.

// emitBootReqSessionGet — req.SessionGet(key): htab lookup, "" on miss.
func (g *Generator) emitBootReqSessionGet(req string, args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("TRequest.SessionGet expects 1 argument, got %d", len(args))
	}
	keyReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.needHashtab = true
	sess := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", sess, g.bootReqField(req, 48)))
	v := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_get(ptr %s, ptr %s)", v, sess, keyReg))
	vNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", vNull, v))
	empty := g.ptrTo(g.addString(""), 1)
	sel := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, ptr %s, ptr %s", sel, vNull, empty, v))
	return sel, "ptr", nil
}

// emitBootReqSessionSet — req.SessionSet(key, value): strdup the value (htab
// stores value pointers verbatim; the caller's string may be request-scoped),
// then put. No cookie re-send — matching Go, where Set does not dirty the
// cookie.
func (g *Generator) emitBootReqSessionSet(req string, args []ast.Expression) (string, string, error) {
	if len(args) != 2 {
		return "", "", fmt.Errorf("TRequest.SessionSet expects 2 arguments, got %d", len(args))
	}
	keyReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	valReg, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	g.needHashtab = true
	sess := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", sess, g.bootReqField(req, 48)))
	dup := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_strdup(ptr %s)", dup, valReg))
	g.line(fmt.Sprintf("  call void @__kylix_htab_put(ptr %s, ptr %s, ptr %s)", sess, keyReg, dup))
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 0, 0", r))
	return r, "i64", nil
}

// emitBootReqSessionDelete — req.SessionDelete(key).
func (g *Generator) emitBootReqSessionDelete(req string, args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("TRequest.SessionDelete expects 1 argument, got %d", len(args))
	}
	keyReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.needHashtab = true
	sess := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", sess, g.bootReqField(req, 48)))
	g.line(fmt.Sprintf("  call void @__kylix_htab_del(ptr %s, ptr %s)", sess, keyReg))
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 0, 0", r))
	return r, "i64", nil
}

// emitBootReqSessionMarkRemember — req.SessionMarkRemember(): persist the
// remember marker (read by the sliding TTL and the finish cookie attrs) and
// flag the cookie dirty so finish re-sends it with Max-Age.
func (g *Generator) emitBootReqSessionMarkRemember(req string, args []ast.Expression) (string, string, error) {
	if len(args) != 0 {
		return "", "", fmt.Errorf("TRequest.SessionMarkRemember expects 0 arguments, got %d", len(args))
	}
	g.needHashtab = true
	sess := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", sess, g.bootReqField(req, 48)))
	remKey := g.ptrTo(g.addString(bootSessionKeyRem), len(bootSessionKeyRem)+1)
	one := g.ptrTo(g.addString("1"), 2)
	g.line(fmt.Sprintf("  call void @__kylix_htab_put(ptr %s, ptr %s, ptr %s)", sess, remKey, one))
	flagsSlot := g.bootReqField(req, 56)
	f0 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", f0, flagsSlot))
	f1 := g.tmp()
	g.line(fmt.Sprintf("  %s = or i64 %s, %d", f1, f0, bootSessFlagDirty))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", f1, flagsSlot))
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 0, 0", r))
	return r, "i64", nil
}

// emitBootReqSessionDestroy — req.SessionDestroy(): drop the stored session
// (via the persisted __sid value) from the outer table and flag destroyed so
// finish sends the clear-cookie line. The per-session htab itself is not
// freed (htab has no free — same leak profile as expired sessions).
func (g *Generator) emitBootReqSessionDestroy(req string, args []ast.Expression) (string, string, error) {
	if len(args) != 0 {
		return "", "", fmt.Errorf("TRequest.SessionDestroy expects 0 arguments, got %d", len(args))
	}
	g.needHashtab = true
	sess := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", sess, g.bootReqField(req, 48)))
	sidKey := g.ptrTo(g.addString(bootSessionKeySID), len(bootSessionKeySID)+1)
	sid := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_get(ptr %s, ptr %s)", sid, sess, sidKey))
	sidNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", sidNull, sid))
	joinLbl := g.label()
	delLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", sidNull, joinLbl, delLbl))
	g.line(fmt.Sprintf("%s:", delLbl))
	tbl := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", tbl, bootSessionsGlobal))
	tblNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", tblNull, tbl))
	doDelLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", tblNull, joinLbl, doDelLbl))
	g.line(fmt.Sprintf("%s:", doDelLbl))
	g.line(fmt.Sprintf("  call void @__kylix_htab_del(ptr %s, ptr %s)", tbl, sid))
	g.line(fmt.Sprintf("  br label %%%s", joinLbl))
	g.line(fmt.Sprintf("%s:", joinLbl))
	flagsSlot := g.bootReqField(req, 56)
	f0 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", f0, flagsSlot))
	f1 := g.tmp()
	g.line(fmt.Sprintf("  %s = or i64 %s, %d", f1, f0, bootSessFlagDestroyed))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", f1, flagsSlot))
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 0, 0", r))
	return r, "i64", nil
}

// emitBootReqSessionCSRFToken — req.SessionCSRFToken(): the persisted __csrf
// token, minted (16 random bytes → 32 hex chars) and stored on first call —
// the value the CSRF middleware validates forms against (P1.7 CSRF slice).
func (g *Generator) emitBootReqSessionCSRFToken(req string, args []ast.Expression) (string, string, error) {
	if len(args) != 0 {
		return "", "", fmt.Errorf("TRequest.SessionCSRFToken expects 0 arguments, got %d", len(args))
	}
	g.needHashtab = true
	sess := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", sess, g.bootReqField(req, 48)))
	csrfKey := g.ptrTo(g.addString(bootSessionCSRFKey), len(bootSessionCSRFKey)+1)
	tok0 := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_get(ptr %s, ptr %s)", tok0, sess, csrfKey))
	tokNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", tokNull, tok0))
	tokSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", tokSlot))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", tok0, tokSlot))
	genLbl := g.label()
	haveLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", tokNull, genLbl, haveLbl))
	g.line(fmt.Sprintf("%s:", genLbl))
	minted := g.tmp()
	g.enqueueStdlib("boot", "randhex", "randhex", 0)
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_boot_rand_hex(i64 16)", minted))
	dup := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_strdup(ptr %s)", dup, minted))
	g.line(fmt.Sprintf("  call void @__kylix_htab_put(ptr %s, ptr %s, ptr %s)", sess, csrfKey, dup))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", minted, tokSlot))
	g.line(fmt.Sprintf("  br label %%%s", haveLbl))
	g.line(fmt.Sprintf("%s:", haveLbl))
	tok := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", tok, tokSlot))
	return tok, "ptr", nil
}

// emitBootReqSessionRegenerate — req.SessionRegenerate() (v0.10.0 P2,
// session-fixation defense at login): mints a fresh SID, moves the session
// entry in the outer table from the old SID to the new one, re-persists
// __sid, and flags the cookie dirty so finish re-sends it. Mirrors Go
// pkg/boot SessionRegenerate. The per-session htab (and thus all values,
// including __user/__roles) carries over untouched.
func (g *Generator) emitBootReqSessionRegenerate(req string, args []ast.Expression) (string, string, error) {
	if len(args) != 0 {
		return "", "", fmt.Errorf("TRequest.SessionRegenerate expects 0 arguments, got %d", len(args))
	}
	g.needHashtab = true
	g.enqueueStdlib("boot", "randhex", "randhex", 0)
	g.bootDeclareSessionsGlobal()
	sess := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", sess, g.bootReqField(req, 48)))
	sessNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", sessNull, sess))
	doneLbl := g.label()
	regenLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", sessNull, doneLbl, regenLbl))
	g.line(fmt.Sprintf("%s:", regenLbl))
	// oldSid = session["__sid"] (null when the session was never persisted).
	sidKey := g.ptrTo(g.addString(bootSessionKeySID), len(bootSessionKeySID)+1)
	oldSid := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_get(ptr %s, ptr %s)", oldSid, sess, sidKey))
	oldNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", oldNull, oldSid))
	newSid := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_boot_rand_hex(i64 32)", newSid))
	delLbl := g.label()
	putLbl := g.label()
	delDoLbl := g.label()
	putDoLbl := g.label()
	finishLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", oldNull, putLbl, delLbl))
	g.line(fmt.Sprintf("%s:", delLbl))
	tbl := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", tbl, bootSessionsGlobal))
	tblNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", tblNull, tbl))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", tblNull, putLbl, delDoLbl))
	g.line(fmt.Sprintf("%s:", delDoLbl))
	g.line(fmt.Sprintf("  call void @__kylix_htab_del(ptr %s, ptr %s)", tbl, oldSid))
	g.line(fmt.Sprintf("  br label %%%s", putLbl))
	g.line(fmt.Sprintf("%s:", putLbl))
	tbl2 := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", tbl2, bootSessionsGlobal))
	tbl2Null := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", tbl2Null, tbl2))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", tbl2Null, finishLbl, putDoLbl))
	g.line(fmt.Sprintf("%s:", putDoLbl))
	g.line(fmt.Sprintf("  call void @__kylix_htab_put(ptr %s, ptr %s, ptr %s)", tbl2, newSid, sess))
	dupSid := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_strdup(ptr %s)", dupSid, newSid))
	g.line(fmt.Sprintf("  call void @__kylix_htab_put(ptr %s, ptr %s, ptr %s)", sess, sidKey, dupSid))
	g.line(fmt.Sprintf("  br label %%%s", finishLbl))
	g.line(fmt.Sprintf("%s:", finishLbl))
	flagsSlot := g.bootReqField(req, 56)
	f0 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", f0, flagsSlot))
	f1 := g.tmp()
	g.line(fmt.Sprintf("  %s = or i64 %s, %d", f1, f0, bootSessFlagDirty))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", f1, flagsSlot))
	g.line(fmt.Sprintf("  br label %%%s", doneLbl))
	g.line(fmt.Sprintf("%s:", doneLbl))
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 0, 0", r))
	return r, "i64", nil
}

// ---- v0.9.0 P1.7 CSRF (pkg/boot/csrf.go parity) ----
//
// Synchronizer-token gate over the session table: unsafe methods (anything
// but GET — the boot server only routes GET/POST/PUT/DELETE) must present the
// session's __csrf token via the X-CSRF-Token header or the _csrf form field.
// Opt-in via BootUseCSRF: the BootRun gate loads @__kylix_boot_csrf_enabled
// so programs without the call keep POSTs working.

// bootDeclareCsrfGlobal emits the csrf toggle global once (module level).
func (g *Generator) bootDeclareCsrfGlobal() {
	if g.bootCsrfEnabledDeclared {
		return
	}
	g.bootCsrfEnabledDeclared = true
	g.line(fmt.Sprintf("%s = global i1 false", bootCsrfEnabledGlobal))
}

// emitBootUseCSRFBody — void @__kylix_boot_BootUseCSRF(): arm the gate.
func (g *Generator) emitBootUseCSRFBody() {
	g.bootDeclareCsrfGlobal()
	g.line("define void @__kylix_boot_BootUseCSRF() {")
	g.line("entry:")
	g.line(fmt.Sprintf("  store i1 true, ptr %s", bootCsrfEnabledGlobal))
	g.line("  ret void")
	g.line("}")
	g.line("")
}

// emitBootCsrfCheckBody — ptr @__kylix_boot_csrf_check(ptr %req,
// ptr %headers): null = pass (safe method, gate off, or token match);
// non-null = a 403 BootText response handle the caller must send INSTEAD of
// dispatching to the handler. Token compare is strcmp here — Go uses
// subtle.ConstantTimeCompare (debt: timing side channel, low risk for a
// per-session random token).
func (g *Generator) emitBootCsrfCheckBody() {
	c := func(format string, args ...interface{}) { g.line(fmt.Sprintf(format, args...)) }
	g.needHashtab = true
	g.enqueueStdlib("boot", "BootText", "BootText", 2)
	g.enqueueStdlib("boot", "formget", "formget", 0)
	g.bootDeclareCsrfGlobal()
	c("define ptr @__kylix_boot_csrf_check(ptr %%req, ptr %%headers) {")
	c("entry:")
	passLbl := g.label()
	chkMethodLbl := g.label()
	en := g.tmp()
	c("  %s = load i8, ptr %s", en, bootCsrfEnabledGlobal)
	enOn := g.tmp()
	c("  %s = icmp ne i8 %s, 0", enOn, en)
	c("  br i1 %s, label %%%s, label %%%s", enOn, chkMethodLbl, passLbl)
	// Gate off or safe method → pass.
	c("%s:", passLbl)
	c("  ret ptr null")
	c("%s:", chkMethodLbl)
	method := g.tmp()
	c("  %s = load ptr, ptr %s", method, g.bootReqField("%req", 0))
	mCmp := g.tmp()
	c("  %s = call i32 @strcmp(ptr %s, ptr %s)", mCmp, method,
		g.ptrTo(g.addString("GET"), 4))
	isGet := g.tmp()
	c("  %s = icmp eq i32 %s, 0", isGet, mCmp)
	chkSessLbl := g.label()
	c("  br i1 %s, label %%%s, label %%%s", isGet, passLbl, chkSessLbl)
	// Session token — issued by req.CSRFToken(); absent ⇒ 403 missing.
	c("%s:", chkSessLbl)
	sess := g.tmp()
	c("  %s = load ptr, ptr %s", sess, g.bootReqField("%req", 48))
	sessTok := g.tmp()
	csrfKey := g.ptrTo(g.addString(bootSessionCSRFKey), len(bootSessionCSRFKey)+1)
	c("  %s = call ptr @__kylix_htab_get(ptr %s, ptr %s)", sessTok, sess, csrfKey)
	sessTokNull := g.tmp()
	c("  %s = icmp eq ptr %s, null", sessTokNull, sessTok)
	chkTokLbl := g.label()
	missingLbl := g.label()
	c("  br i1 %s, label %%%s, label %%%s", sessTokNull, missingLbl, chkTokLbl)
	// Request token: X-CSRF-Token header first, _csrf form field fallback —
	// both funnel through a slot into the shared compare block.
	chkTokJoinLbl := g.label()
	chkTokCmpLbl := g.label()
	c("%s:", chkTokLbl)
	tokSlot := g.tmp()
	c("  %s = alloca ptr, align 8", tokSlot)
	c("  store ptr null, ptr %s", tokSlot)
	needle := g.ptrTo(g.addString(bootCsrfHeaderNeedle), len(bootCsrfHeaderNeedle)+1)
	hp := g.tmp()
	c("  %s = call ptr @strstr(ptr %%headers, ptr %s)", hp, needle)
	hpNull := g.tmp()
	c("  %s = icmp ne ptr %s, null", hpNull, hp)
	hdrValLbl := g.label()
	formChkLbl := g.label()
	c("  br i1 %s, label %%%s, label %%%s", hpNull, hdrValLbl, formChkLbl)
	c("%s:", hdrValLbl)
	hv := g.tmp()
	c("  %s = getelementptr inbounds i8, ptr %s, i64 %d", hv, hp, len(bootCsrfHeaderNeedle))
	// Copy the header value out (it runs to the '\r' line terminator — same
	// trim idiom as emitBootReqHeader) so strcmp sees a C string.
	cr := g.tmp()
	c("  %s = call ptr @strchr(ptr %s, i32 13)", cr, hv)
	crAddr := g.tmp()
	c("  %s = ptrtoint ptr %s to i64", crAddr, cr)
	hvAddr := g.tmp()
	c("  %s = ptrtoint ptr %s to i64", hvAddr, hv)
	tokLen := g.tmp()
	c("  %s = sub i64 %s, %s", tokLen, crAddr, hvAddr)
	g.needMemcpy = true
	hbuf := g.tmp()
	c("  %s = %s", hbuf, g.mallocCall(tokLen))
	c("  call ptr @memcpy(ptr %s, ptr %s, i64 %s)", hbuf, hv, tokLen)
	hterm := g.tmp()
	c("  %s = getelementptr inbounds i8, ptr %s, i64 %s", hterm, hbuf, tokLen)
	c("  store i8 0, ptr %s", hterm)
	c("  store ptr %s, ptr %s", hbuf, tokSlot)
	c("  br label %%%s", chkTokJoinLbl)
	c("%s:", formChkLbl)
	body := g.tmp()
	c("  %s = load ptr, ptr %s", body, g.bootReqField("%req", 24))
	fv := g.tmp()
	csrfField := g.ptrTo(g.addString(bootCsrfFormField), len(bootCsrfFormField)+1)
	c("  %s = call ptr @__kylix_boot_form_get(ptr %s, ptr %s)", fv, body, csrfField)
	c("  store ptr %s, ptr %s", fv, tokSlot)
	c("  br label %%%s", chkTokJoinLbl)
	c("%s:", chkTokJoinLbl)
	reqTok := g.tmp()
	c("  %s = load ptr, ptr %s", reqTok, tokSlot)
	reqTokNull := g.tmp()
	c("  %s = icmp eq ptr %s, null", reqTokNull, reqTok)
	mismatchLbl := g.label()
	c("  br i1 %s, label %%%s, label %%%s", reqTokNull, mismatchLbl, chkTokCmpLbl)
	c("%s:", chkTokCmpLbl)
	tcmp := g.tmp()
	c("  %s = call i32 @strcmp(ptr %s, ptr %s)", tcmp, reqTok, sessTok)
	tokEq := g.tmp()
	c("  %s = icmp eq i32 %s, 0", tokEq, tcmp)
	c("  br i1 %s, label %%%s, label %%%s", tokEq, passLbl, mismatchLbl)
	// 403s — BootText handles, sent by BootRun's normal response assembly.
	c("%s:", missingLbl)
	miss := g.tmp()
	c("  %s = call ptr @__kylix_boot_BootText(i64 403, ptr %s)", miss,
		g.ptrTo(g.addString(bootCsrfMissingMsg), len(bootCsrfMissingMsg)+1))
	c("  ret ptr %s", miss)
	c("%s:", mismatchLbl)
	mism := g.tmp()
	c("  %s = call ptr @__kylix_boot_BootText(i64 403, ptr %s)", mism,
		g.ptrTo(g.addString(bootCsrfMismatchMsg), len(bootCsrfMismatchMsg)+1))
	c("  ret ptr %s", mism)
	c("}")
	c("")
}
