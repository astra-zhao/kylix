package llvmgen

import (
	"fmt"
	"strings"

	"kylix/ast"
)

// bootOpaqueTypes are the KylixBoot HTTP framework types lowered to opaque
// pointers in LLVM IR (v0.6.1). TRequest/TResponse carry no IR-visible
// structure for the tutorial handlers — they are passed/returned by pointer
// and only their field accesses (resp.Body) are compiled to real GEP loads.
var bootOpaqueTypes = map[string]bool{
	"TRequest":     true,
	"TResponse":    true,
	"BootRequest":  true,
	"BootResponse": true,
	// v0.10.0 P2: the db handle is opaque too — TDatabase in user-declared
	// function signatures (AdminOpen etc.) must lower to ptr, not the i64
	// fallback, or call sites type-mismatch against sqlite3_* (ptr...).
	"TDatabase": true,
	"Database":  true,
}

// stdlib_boot.go — LLVM IR for the `boot` (KylixBoot) stdlib module.
//
// The KylixBoot framework on the Go backend depends on net/http and reflect.
// The LLVM backend has no HTTP server and no RTTI, so most KylixBoot runtime
// functions are typed stubs that evaluate their arguments and return an empty
// default — enough for the tutorial examples (41-47, 49-51) to compile,
// register routes and run without crashing. BootText/BootJSON/BootHTML return
// a real {i64 status, ptr body} response handle so handler bodies lower to
// valid IR (the pre-v0.6.1 catch-all i64 stub produced a store type mismatch
// when their ptr result was stored into an i64 %result).

// bootStubReturnTypes maps each boot function to the LLVM type its call-site
// result register must have. Functions not listed default to i64.
var bootStubReturnTypes = map[string]string{
	"BootConfigGetString":  "ptr",
	"BootResolve":          "ptr",
	"BootEnforceAuth":      "ptr",
	"BootEnforceRole":      "ptr",
	"BootConfigGetInt":     "i64",
	"BootRun":              "i64",
	"BootReadJSON":         "i64",
	"BootRegisterJwtAuth":  "void",
	"BootGET":              "void",
	"BootPOST":             "void",
	"BootPUT":              "void",
	"BootDELETE":           "void",
	"BootUseLogger":        "void",
	"BootUseRecover":       "void",
	"BootUseCORS":          "void",
	"BootUseRequestID":     "void",
	"BootConfigSet":        "void",
	"BootRegisterInstance": "void",
	"BootRegisterAuth":     "void",
	"BootRegisterRoles":    "void",
	"BootNotFoundPage":     "void",
	"BootErrorPage":        "void",
	"BootPagerHTML":        "ptr",
	"BootUseCSRF":          "void",
}

// isBootHandleType reports whether name is an opaque KylixBoot handle type
// (TRequest/TResponse, v0.6.6/v0.7.0 P3) — not registered in g.classes but
// dispatched by name in the chained-method emitter. Used to register result
// slots so `result := result.Redirect(...)` resolves its receiver type.
func isBootHandleType(name string) bool {
	return name == "TRequest" || name == "TResponse" ||
		name == "BootRequest" || name == "BootResponse"
}

func (g *Generator) emitBootCall(funcName string, args []ast.Expression) (string, string, error) {
	switch funcName {
	case "BootText", "BootJSON", "BootHTML":
		// Real response handle {i64 status, ptr body}.
		return g.emitBootResponseCall(funcName, args)
	case "BootRun":
		// v0.6.6: real HTTP server (listen + accept + dispatch).
		if len(args) < 1 {
			return "", "", fmt.Errorf("boot.BootRun expects a port, got %d", len(args))
		}
		portReg, _, err := g.emitExpr(args[0])
		if err != nil {
			return "", "", err
		}
		g.enqueueStdlib("boot", "BootRun", "BootRun", 0)
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = call i64 @__kylix_boot_BootRun(i64 %s)", r, portReg))
		return r, "i64", nil
	case "BootRegisterJwtAuth":
		// v0.6.8: store the JWT secret as a module global so BootEnforceAuth can
		// verify `Authorization: Bearer <token>` at request time. Only string
		// literals are supported as the secret (a literal constant → module-level
		// @.str.N can be referenced from the wrapper define).
		for _, a := range args {
			if _, _, err := g.emitExpr(a); err != nil {
				return "", "", err
			}
		}
		if len(args) >= 1 {
			if lit, ok := args[0].(*ast.StringLiteral); ok {
				g.bootJwtSecretConst = g.addString(lit.Value)
			}
		}
		// BootEnforceAuth body is enqueued so the auth define is emitted when a
		// [Authenticated] route exists (emitBootWrapper calls it regardless).
		g.enqueueStdlib("boot", "BootEnforceAuth", "BootEnforceAuth", 0)
		return "0", "void", nil
	case "BootStatic":
		// v0.7.0 P2: enable static file serving under /static/. The dir argument
		// is stored into @__kylix_boot_static_dir by the define; BootRun's
		// route-miss path consults it (GET + /static/ prefix → file response).
		if len(args) < 1 {
			return "", "", fmt.Errorf("boot.BootStatic expects a directory, got %d", len(args))
		}
		dirReg, _, err := g.emitExpr(args[0])
		if err != nil {
			return "", "", err
		}
		for _, a := range args[1:] {
			if _, _, err := g.emitExpr(a); err != nil {
				return "", "", err
			}
		}
		g.enqueueStdlib("boot", "BootStatic", "BootStatic", 0)
		g.enqueueStdlib("boot", "servestatic", "servestatic", 0)
		g.line(fmt.Sprintf("  call void @__kylix_boot_BootStatic(ptr %s)", dirReg))
		return "0", "void", nil
	case "BootNotFoundPage", "BootErrorPage":
		// v0.7.0 P3: custom 404/500 HTML pages — stored into module globals
		// that BootRun's error paths consult (null = built-in response).
		if len(args) < 1 {
			return "", "", fmt.Errorf("boot.%s expects an html string, got %d", funcName, len(args))
		}
		htmlReg, _, err := g.emitExpr(args[0])
		if err != nil {
			return "", "", err
		}
		for _, a := range args[1:] {
			if _, _, err := g.emitExpr(a); err != nil {
				return "", "", err
			}
		}
		g.enqueueStdlib("boot", funcName, funcName, 0)
		g.line(fmt.Sprintf("  call void @__kylix_boot_%s(ptr %s)", funcName, htmlReg))
		return "0", "void", nil
	case "BootPagerHTML":
		// v0.9.0 P1.7: pagination navigation bar — BootPagerHTML(base, page,
		// size, total, window) → HTML string (mirrors Go boot.PagerHTML).
		if len(args) != 5 {
			return "", "", fmt.Errorf("boot.BootPagerHTML expects 5 arguments, got %d", len(args))
		}
		argRs := make([]string, 0, 5)
		for _, a := range args {
			r, _, err := g.emitExpr(a)
			if err != nil {
				return "", "", err
			}
			argRs = append(argRs, r)
		}
		g.enqueueStdlib("boot", "BootPagerHTML", "BootPagerHTML", 0)
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = call ptr @__kylix_boot_pager_html(ptr %s, i64 %s, i64 %s, i64 %s, i64 %s)",
			r, argRs[0], argRs[1], argRs[2], argRs[3], argRs[4]))
		return r, "ptr", nil
	case "BootUseCSRF":
		// v0.9.0 P1.7: arm the CSRF gate — store i1 true into
		// @__kylix_boot_csrf_enabled; BootRun's middleware chain consults it
		// between session resolve and handler dispatch.
		g.enqueueStdlib("boot", "BootUseCSRF", "BootUseCSRF", 0)
		g.enqueueStdlib("boot", "csrfcheck", "csrfcheck", 0)
		g.line("  call void @__kylix_boot_BootUseCSRF()")
		return "0", "void", nil
	default:
		retType, ok := bootStubReturnTypes[funcName]
		if !ok {
			retType = "i64"
		}
		return g.emitBootStubCall(funcName, args, retType)
	}
}

// emitBootResponseCall emits a call to a BootText/BootJSON/BootHTML define.
// Status defaults to 200, body to the string arg (BootJSON ignores it — no
// JSON serializer is needed for the tutorials, body is left empty).
func (g *Generator) emitBootResponseCall(funcName string, args []ast.Expression) (string, string, error) {
	statusReg := "200"
	if len(args) >= 1 {
		r, _, err := g.emitExpr(args[0])
		if err != nil {
			return "", "", err
		}
		statusReg = r
	}

	bodyReg := g.ptrTo(g.addString(""), 1)
	if funcName != "BootJSON" && len(args) >= 2 {
		r, rt, err := g.emitExpr(args[1])
		if err != nil {
			return "", "", err
		}
		// v0.6.8: a Variant body (e.g. BootText(200, data['name'])) must be
		// unboxed to a string before BootText strlen()s it — passing the box
		// pointer directly would read the box's tag/payload bytes as text.
		if rt == variantT {
			r, _ = g.coerceValue(r, rt, "ptr")
		}
		bodyReg = r
	}

	g.enqueueStdlib("boot", funcName, funcName, 2)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_boot_%s(i64 %s, ptr %s)", r, funcName, statusReg, bodyReg))
	return r, "ptr", nil
}

func (g *Generator) emitBootBody(funcName string) {
	switch funcName {
	case "BootText", "BootJSON", "BootHTML":
		g.emitBootResponseBody(funcName)
	case "BootStatic":
		g.emitBootStaticBody()
	case "BootNotFoundPage":
		g.emitBootErrorPageBody("BootNotFoundPage", boot404PageGlobal)
	case "BootErrorPage":
		g.emitBootErrorPageBody("BootErrorPage", boot500PageGlobal)
	case "BootPagerHTML":
		// v0.9.0 P1.7: pagination navigation bar (helpers + main define).
		g.emitBootPagerHTMLBody()
	case "formget":
		g.emitBootFormGetBody()
	case "cookieget":
		g.emitBootCookieGetBody()
	case "sessionresolve":
		// v0.9.0 P1.7c: session middleware halves (BootRun-embedded).
		g.emitBootSessionResolveBody()
	case "sessionfinish":
		g.emitBootSessionFinishBody()
	case "randhex":
		g.emitBootRandHexBody()
	case "BootUseCSRF":
		// v0.9.0 P1.7: arm the CSRF toggle global.
		g.emitBootUseCSRFBody()
	case "csrfcheck":
		// v0.9.0 P1.7: the CSRF gate itself (BootRun-embedded middleware).
		g.emitBootCsrfCheckBody()
	case "multipartparse":
		// v0.9.0 P1.7: eager multipart/form-data parse (BootRun-embedded).
		g.emitBootMultipartParseBody()
	case "reqfile":
		// v0.9.0 P1.7: req.File(name) — {content, filename, ok}.
		g.emitBootReqFileBody()
	case "reqsavefile":
		// v0.9.0 P1.7: req.SaveFile(name, dir) — {path, ok}.
		g.emitBootReqSaveFileBody()
	case "servestatic":
		g.emitBootServeStaticBody()
	case "BootGET", "BootPOST", "BootPUT", "BootDELETE":
		// v0.6.6: real route registration — store {method, path, handler} in
		// the module route table for @__kylix_boot_BootRun to dispatch.
		g.emitBootRegisterRouteBody(funcName)
	case "BootRegisterInstance", "BootUseLogger", "BootUseRecover", "BootUseCORS",
		"BootUseRequestID", "BootConfigSet", "BootRegisterAuth", "BootRegisterRoles":
		// void no-op — registrations are accepted but have no runtime effect
		// on the LLVM backend (no middleware / DI reflection).
		g.line(fmt.Sprintf("define void @__kylix_boot_%s(ptr %%a, ptr %%b) {", funcName))
		g.line("  ret void")
		g.line("}")
	case "BootRun":
		// v0.6.6: real HTTP server (listen + accept + dispatch).
		g.emitBootRunBody()
	case "readheaders":
		g.emitBootReadHeadersBody()
	case "readbody":
		// v0.6.8: read Content-Length request body bytes after the header block
		// and store the NUL-terminated body on the request handle (req[24]).
		g.emitBootReadBodyBody()
	case "parsereq":
		g.emitBootParseRequestBody()
	case "routelookup":
		g.emitBootRouteLookupBody()
	case "pathmatch":
		g.emitBootPathMatchBody()
	case "BootEnforceAuth":
		// v0.6.8: real auth guard — read `Authorization: Bearer <token>` from
		// the request headers and verify with JwtVerify(secret, token) (HS256).
		// Returns null = pass, or a 401 response handle on missing/invalid auth.
		g.emitBootEnforceAuthBody()
	case "BootEnforceRole":
		// v0.10.0 P2: real guard (replaces the always-pass stub) — checks the
		// comma-separated __roles session key, mirroring Go EnforceRole's
		// session-first path. null = pass, or a 403 response handle.
		g.emitBootEnforceRoleBody()
	case "BootReadJSON":
		g.line("define i64 @__kylix_boot_BootReadJSON(ptr %req, ptr %dst) {")
		g.line("  ret i64 0")
		g.line("}")
	default:
		// No separate define — the stub is inlined at the call site.
	}
}

// emitBootResponseBody emits the response handle:
//
//	%h = malloc(40); {i64 status, ptr body, ptr ctype, ptr cookie, ptr xhdrs}
//	store status @ 0; body @ 8; ctype @ 16; null cookie @ 24; null xhdrs @ 32
//
// v0.7.0 P2: extended from 16 bytes with a Content-Type slot (set by the
// constructor: text/plain / text/html / application/json) plus cookie and
// extra-headers slots that WithCookie/WithHeader fill in.
// v0.8.0 P3: the cookie slot holds a growing buffer of full
// "Set-Cookie: ...\r\n" lines (multi-cookie, mirrors the Go side's
// []string) and the xhdrs slot grows by realloc — see emitBootAppendToSlot.
func (g *Generator) emitBootResponseBody(funcName string) {
	ctype := "text/plain; charset=utf-8"
	switch funcName {
	case "BootHTML":
		ctype = "text/html; charset=utf-8"
	case "BootJSON":
		ctype = "application/json"
	}
	g.line(fmt.Sprintf("define ptr @__kylix_boot_%s(i64 %%status, ptr %%body) {", funcName))
	g.line("entry:")
	// v0.8.0 P2: the handle lives on the per-request arena — BootRun's loop
	// resets it after each request, so response handles no longer leak.
	g.needArena = true
	h := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_arena_alloc(i64 40)", h))
	s := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i64, ptr %s, i64 0", s, h))
	g.line(fmt.Sprintf("  store i64 %%status, ptr %s", s))
	b := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 8", b, h))
	g.line(fmt.Sprintf("  store ptr %%body, ptr %s", b))
	c := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 16", c, h))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", g.ptrTo(g.addString(ctype), len(ctype)), c))
	k := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 24", k, h))
	g.line("  store ptr null, ptr " + k)
	x := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 32", x, h))
	g.line("  store ptr null, ptr " + x)
	g.line(fmt.Sprintf("  ret ptr %s", h))
	g.line("}")
}

// emitBootResponseFieldAccess lowers `resp.Body` / `resp.Status` on a
// KylixBoot TResponse handle ({i64 status, ptr body}, 16 bytes) to GEP+load.
// The receiver is a call expression like `ctrl.ListUsers(nil)` whose value is
// the handle returned by BootText/BootJSON (see emitBootResponseCall).
func (g *Generator) emitBootResponseFieldAccess(obj ast.Expression, field string) (string, string, error) {
	handle, _, err := g.emitExpr(obj)
	if err != nil {
		return "", "", err
	}
	switch strings.ToLower(field) {
	case "body":
		slot := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 8", slot, handle))
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = load ptr, ptr %s", r, slot))
		return r, "ptr", nil
	case "status", "statuscode":
		slot := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr inbounds i64, ptr %s, i64 0", slot, handle))
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = load i64, ptr %s", r, slot))
		return r, "i64", nil
	default:
		// Unknown field — null fallback keeps the IR legal.
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = inttoptr i64 0 to ptr ; TResponse.%s unsupported", r, field))
		return r, "ptr", nil
	}
}

// emitBootEnforceAuthBody — ptr @__kylix_boot_BootEnforceAuth(ptr %req).
// v0.6.8: reads `Authorization: Bearer <token>` from the request headers and
// verifies it with jwt.JwtVerify(secret, token) (HS256). Returns null = pass,
// or a {401, "Unauthorized"} response handle when the header is absent/malformed
// or the signature fails. When no BootRegisterJwtAuth secret was configured the
// secret is an empty string, so every token fails → 401 (deny by default,
// matching the Go backend's missing-validator behavior).
func (g *Generator) emitBootEnforceAuthBody() {
	// JwtVerify needs the jwt + variant + hashtab runtimes; the 401 body is a
	// BootText response handle.
	g.enqueueStdlib("jwt", "JwtVerify", "JwtVerify", 0)
	g.enqueueStdlib("boot", "BootText", "BootText", 0)
	g.needVariantRuntime = true
	g.needHashtab = true
	// secretConst is the module-level string constant to verify against. When a
	// secret was configured (BootRegisterJwtAuth), a module global holds the
	// const pointer and we load it inside the function body at verify time;
	// otherwise pass an empty string (deny-all). The empty-string ptrTo must
	// run inside the function body (not module scope), so it is deferred below.
	secretConst := ""
	if g.bootJwtSecretConst == "" {
		emptyStr := g.addString("")
		secretConst = emptyStr
	} else {
		// Emitted here (not in emitBootGlobals) because BootRegisterJwtAuth's
		// string constant is only allocated during emitMain, which runs after
		// emitBootGlobals. emitPendingStdlib → emitBootBody → this function
		// runs after emitMain, so bootJwtSecretConst is final.
		g.line(fmt.Sprintf("@__kylix_boot_jwt_secret = global ptr %s", g.bootJwtSecretConst))
	}
	g.line("define ptr @__kylix_boot_BootEnforceAuth(ptr %req) {")
	g.line("entry:")
	// secretOp must be computed in the entry block: it emits a gep
	// instruction, and the pass label below terminates with `ret` — any
	// instruction after it would be dead-on-arrival inside the wrong block.
	secretOp := ""
	if g.bootJwtSecretConst != "" {
		// Load the global's value (the string ptr) inside the function body.
		secReg := g.tmp()
		g.line(fmt.Sprintf("  %s = load ptr, ptr @__kylix_boot_jwt_secret", secReg))
		secretOp = secReg
	} else {
		secretOp = g.ptrTo(secretConst, 1)
	}
	// v0.10.0 P2 session-first (Go pkg/boot/security.go parity): a live
	// session with a non-empty __user key passes without touching the
	// Bearer path. Session may be null in exotic non-BootRun contexts —
	// fall through to the Bearer check either way.
	sess0 := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", sess0, g.bootReqField("%req", 48)))
	sessNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", sessNull, sess0))
	userChkLbl := g.label()
	bearerLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", sessNull, bearerLbl, userChkLbl))
	g.line(fmt.Sprintf("%s:", userChkLbl))
	userKey := g.ptrTo(g.addString("__user"), 7)
	user := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_get(ptr %s, ptr %s)", user, sess0, userKey))
	userNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", userNull, user))
	userLenChkLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", userNull, bearerLbl, userLenChkLbl))
	g.line(fmt.Sprintf("%s:", userLenChkLbl))
	ulen := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %s)", ulen, user))
	uok := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ugt i64 %s, 0", uok, ulen))
	passLbl0 := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", uok, passLbl0, bearerLbl))
	g.line(fmt.Sprintf("%s:", passLbl0))
	g.line("  ret ptr null")
	g.line(fmt.Sprintf("%s:", bearerLbl))
	headers := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", headers, g.bootReqField("%req", 16)))
	// find "Authorization:" in the header block.
	authHdr := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @strstr(ptr %s, ptr %s)", authHdr, headers, g.ptrTo(g.addString("Authorization:"), 15)))
	authNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", authNull, authHdr))
	denyLbl := g.label()
	chkBearerLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", authNull, denyLbl, chkBearerLbl))
	// find "Bearer " after the header value.
	g.line(fmt.Sprintf("%s:", chkBearerLbl))
	authVal := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 14", authVal, authHdr)) // "Authorization:" = 14
	bearer := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @strstr(ptr %s, ptr %s)", bearer, authVal, g.ptrTo(g.addString("Bearer "), 8)))
	bearerNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", bearerNull, bearer))
	verifyLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", bearerNull, denyLbl, verifyLbl))
	// token starts after "Bearer " (7 chars); copy up to \r into a fresh buffer.
	g.line(fmt.Sprintf("%s:", verifyLbl))
	token := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 7", token, bearer))
	cr := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @strchr(ptr %s, i32 13)", cr, token)) // '\r'
	crNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", crNull, cr))
	tokLen := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %s)", tokLen, token))
	crAddr := g.tmp()
	g.line(fmt.Sprintf("  %s = ptrtoint ptr %s to i64", crAddr, cr))
	tokAddr := g.tmp()
	g.line(fmt.Sprintf("  %s = ptrtoint ptr %s to i64", tokAddr, token))
	crLen := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, %s", crLen, crAddr, tokAddr))
	realLen := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i64 %s, i64 %s", realLen, crNull, tokLen, crLen))
	bufSize := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", bufSize, realLen))
	tbuf := g.tmp()
	g.line(fmt.Sprintf("  %s = %s", tbuf, g.mallocCall(bufSize)))
	g.needMemcpy = true
	g.line(fmt.Sprintf("  call ptr @memcpy(ptr %s, ptr %s, i64 %s)", tbuf, token, realLen))
	termPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", termPtr, tbuf, realLen))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", termPtr))
	// JwtVerify(secret, token) → Variant box; nilbox (tag 0) = invalid. The
	// secret global holds a ptr; load it to get the actual string pointer.
	vres := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_jwt_JwtVerify(ptr %s, ptr %s)", vres, secretOp, tbuf))
	tagLoc := g.boxAddr(vres, 0)
	tag := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", tag, tagLoc))
	isNil := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 0", isNil, tag))
	passLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isNil, denyLbl, passLbl))
	// pass: ret null.
	g.line(fmt.Sprintf("%s:", passLbl))
	g.line("  ret ptr null")
	// deny: 401 "Unauthorized".
	g.line(fmt.Sprintf("%s:", denyLbl))
	denyRes := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_boot_BootText(i64 401, ptr %s)", denyRes, g.ptrTo(g.addString("Unauthorized"), 13)))
	g.line(fmt.Sprintf("  ret ptr %s", denyRes))
	g.line("}")
	g.line("")
}

// emitBootEnforceRoleBody — ptr @__kylix_boot_BootEnforceRole(ptr %req,
// ptr %role). v0.10.0 P2: reads the comma-separated __roles session key and
// passes when any token equals %role (Go pkg/boot/security.go EnforceRole
// session path parity; no JWT roles fallback on the LLVM side — documented).
// null = pass; otherwise a {403, "Forbidden"} response handle. Assumes the
// caller ran BootEnforceAuth first (the route guard prologue emits both).
func (g *Generator) emitBootEnforceRoleBody() {
	g.needHashtab = true
	g.enqueueStdlib("boot", "BootText", "BootText", 0)
	c := func(format string, args ...interface{}) { g.line(fmt.Sprintf(format, args...)) }
	c("define ptr @__kylix_boot_BootEnforceRole(ptr %%req, ptr %%role) {")
	c("entry:")
	sess := g.tmp()
	c("  %s = load ptr, ptr %s", sess, g.bootReqField("%req", 48))
	sessNull := g.tmp()
	c("  %s = icmp eq ptr %s, null", sessNull, sess)
	chkKeyLbl := g.label()
	denyLbl := g.label()
	c("  br i1 %s, label %%%s, label %%%s", sessNull, denyLbl, chkKeyLbl)
	// roles = session["__roles"] — absent ⇒ deny.
	c("%s:", chkKeyLbl)
	rolesKey := g.ptrTo(g.addString("__roles"), 8)
	roles := g.tmp()
	c("  %s = call ptr @__kylix_htab_get(ptr %s, ptr %s)", roles, sess, rolesKey)
	rolesNull := g.tmp()
	c("  %s = icmp eq ptr %s, null", rolesNull, roles)
	loopInitLbl := g.label()
	c("  br i1 %s, label %%%s, label %%%s", rolesNull, denyLbl, loopInitLbl)
	// Walk the comma-separated tokens; match when a token equals %role.
	c("%s:", loopInitLbl)
	pSlot := g.tmp()
	c("  %s = alloca ptr, align 8", pSlot)
	c("  store ptr %s, ptr %s", roles, pSlot)
	loopCondLbl := g.label()
	c("  br label %%%s", loopCondLbl)
	c("%s:", loopCondLbl)
	p := g.tmp()
	c("  %s = load ptr, ptr %s", p, pSlot)
	comma := g.tmp()
	c("  %s = call ptr @strchr(ptr %s, i32 44)", comma, p) // ','
	commaNull := g.tmp()
	c("  %s = icmp eq ptr %s, null", commaNull, comma)
	// Token length: up to the comma, or strlen for the last token.
	pAddr := g.tmp()
	c("  %s = ptrtoint ptr %s to i64", pAddr, p)
	commaAddr := g.tmp()
	c("  %s = ptrtoint ptr %s to i64", commaAddr, comma)
	diff := g.tmp()
	c("  %s = sub i64 %s, %s", diff, commaAddr, pAddr)
	pLen := g.tmp()
	c("  %s = call i64 @strlen(ptr %s)", pLen, p)
	tokLen := g.tmp()
	c("  %s = select i1 %s, i64 %s, i64 %s", tokLen, commaNull, pLen, diff)
	rLen := g.tmp()
	c("  %s = call i64 @strlen(ptr %%role)", rLen)
	lenEq := g.tmp()
	c("  %s = icmp eq i64 %s, %s", lenEq, tokLen, rLen)
	tokCmpLbl := g.label()
	tokNextLbl := g.label()
	c("  br i1 %s, label %%%s, label %%%s", lenEq, tokCmpLbl, tokNextLbl)
	c("%s:", tokCmpLbl)
	m := g.tmp()
	c("  %s = call i32 @strncmp(ptr %s, ptr %%role, i64 %s)", m, p, tokLen)
	mEq := g.tmp()
	c("  %s = icmp eq i32 %s, 0", mEq, m)
	passLbl := g.label()
	c("  br i1 %s, label %%%s, label %%%s", mEq, passLbl, tokNextLbl)
	c("%s:", tokNextLbl)
	c("  br i1 %s, label %%%s, label %%%s", commaNull, denyLbl, loopCondLbl)
	c("%s:", passLbl)
	c("  ret ptr null")
	c("%s:", denyLbl)
	denyRes := g.tmp()
	c("  %s = call ptr @__kylix_boot_BootText(i64 403, ptr %s)", denyRes,
		g.ptrTo(g.addString("Forbidden"), 10))
	c("  ret ptr %s", denyRes)
	c("}")
	c("")
}

func (g *Generator) emitBootStubCall(funcName string, args []ast.Expression, retType string) (string, string, error) {
	// Evaluate all arguments for side effects.
	for _, a := range args {
		if _, _, err := g.emitExpr(a); err != nil {
			return "", "", err
		}
	}
	if retType == "ptr" {
		emptyStr := g.addString("")
		return g.ptrTo(emptyStr, 1), "ptr", nil
	}
	if retType == "void" {
		return "0", "void", nil
	}
	if retType == "i1" {
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i1 0, 0 ; boot.%s stub", r, funcName))
		return r, "i1", nil
	}
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 0, 0 ; boot.%s stub", r, funcName))
	return r, "i64", nil
}

// emitBootRegisterRouteBody — v0.6.6: `void @__kylix_boot_Boot<M>(ptr %path,
// ptr %wrapper)` writes {method, path, handler} into the module route table
// (@__kylix_boot_routes / @__kylix_boot_nroutes), which @__kylix_boot_BootRun
// dispatches against. Capacity 64; overflow is dropped.
func (g *Generator) emitBootRegisterRouteBody(method string) {
	methodLit := strings.TrimPrefix(method, "Boot")
	methodPtr := g.addString(methodLit)
	g.line(fmt.Sprintf("define void @__kylix_boot_%s(ptr %%path, ptr %%wrapper) {", method))
	g.line("entry:")
	n := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr @__kylix_boot_nroutes", n))
	full := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sge i64 %s, 64", full, n))
	storeLbl := g.label()
	doneLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", full, doneLbl, storeLbl))
	g.line(fmt.Sprintf("%s:", storeLbl))
	slot := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [64 x { ptr, ptr, ptr }], ptr @__kylix_boot_routes, i64 0, i64 %s", slot, n))
	mField := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { ptr, ptr, ptr }, ptr %s, i32 0, i32 0", mField, slot))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", methodPtr, mField))
	pField := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { ptr, ptr, ptr }, ptr %s, i32 0, i32 1", pField, slot))
	g.line(fmt.Sprintf("  store ptr %%path, ptr %s", pField))
	hField := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { ptr, ptr, ptr }, ptr %s, i32 0, i32 2", hField, slot))
	g.line(fmt.Sprintf("  store ptr %%wrapper, ptr %s", hField))
	nextN := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", nextN, n))
	g.line(fmt.Sprintf("  store i64 %s, ptr @__kylix_boot_nroutes", nextN))
	g.line(fmt.Sprintf("  br label %%%s", doneLbl))
	g.line(fmt.Sprintf("%s:", doneLbl))
	g.line("  ret void")
	g.line("}")
	g.line("")
}
