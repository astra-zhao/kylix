package llvmgen

import (
	"fmt"
	"strings"

	"kylix/ast"
)

// stdlib_httpclient.go — LLVM IR implementation for the `httpclient` stdlib
// module, backed by libcurl.
//
// THttpClient is a heap-allocated 32-byte handle:
//
//	%__kylix_httpclient = type { ptr curl, ptr slist, ptr baseURL, i64 timeout }
//	  offset 0:  curl     — CURL* easy handle (curl_easy_init)
//	  offset 8:  slist    — curl_slist* accumulated headers (SetHeader)
//	  offset 16: baseURL  — String ptr (reserved; NewHttpClient sets null)
//	  offset 24: timeout  — i64 seconds (default 30)
//
// The handle is opaque to Kylix (a ptr-typed local). Field access (c.BaseURL)
// is lowered in expr.go's emitMember via emitHttpclientFieldAccess; method
// calls (c.SetHeader/Get/Post) route through emitHttpclientMethodCall.
//
// libcurl declares (curl_easy_init/setopt/perform/cleanup, curl_slist_append/
// free_all) and the -lcurl link flag are wired in codegen.go/compile.go. This
// file emits only the @__kylix_httpclient_* defines + the @__kylix_http_write_cb
// write callback. The define names trigger compile.go's @__kylix_httpclient_
// scan, which appends -lcurl at link time. realloc is not in codegen.go's libc
// set, so it is declared lazily here (once per module, guarded).

const httpClientTypeName = "THttpClient"

// Handle field byte offsets (matches the struct layout above). The handle is
// treated as an opaque i8 region; field access uses getelementptr i8 with these
// offsets — no named struct type needs to be declared at module scope.
const (
	httpClientOffsetCurl    = 0
	httpClientOffsetSlist   = 8
	httpClientOffsetBaseURL = 16
	httpClientOffsetTimeout = 24
	httpClientHandleSize    = 32
)

// libcurl CURLOPT constants (from curl/easy.h). Passed as the i32 second arg
// to curl_easy_setopt; the variadic third arg's type depends on the option
// category (OBJECTPOINT→ptr, FUNCTIONPOINT→function ptr, LONG→i64).
const (
	curloptWrdata        = 10001   // CURLOPT_WRITEDATA     (OBJECTPOINT)
	curloptUrl           = 10002   // CURLOPT_URL           (OBJECTPOINT)
	curloptPostfields    = 10015   // CURLOPT_POSTFIELDS    (OBJECTPOINT)
	curloptHttpheader    = 10023   // CURLOPT_HTTPHEADER    (OBJECTPOINT)
	curloptCustomrequest = 10036   // CURLOPT_CUSTOMREQUEST (OBJECTPOINT)
	curloptTimeoutMs     = 156     // CURLOPT_TIMEOUT_MS    (LONG, milliseconds)
	curloptPost          = 47      // CURLOPT_POST          (LONG)
	curloptWritefunction = 20011   // CURLOPT_WRITEFUNCTION (FUNCTIONPOINT)
	curlInfoResponseCode = 2097154 // CURLINFO_RESPONSE_CODE (LONG; CURLINFO_LONG 0x200000 + 2)
)

// httpClientDefaultTimeout is the timeout (milliseconds) NewHttpClient stores
// when no explicit timeout is provided. Mirrors the Go backend default (10s).
// v6.1.0: unit changed from seconds to milliseconds to match the Kylix
// `NewHttpClient(baseURL, timeoutMs: Integer)` declaration; the handle stores
// milliseconds and Get/Post pass CURLOPT_TIMEOUT_MS directly.
const httpClientDefaultTimeout = 10000

// emitHttpclientCall dispatches a `httpclient.Func(args)` / bare `Func(args)`
// call. NewHttpClient and the one-shot helpers are lowered to real libcurl IR
// (the helpers share a single @__kylix_httpclient_DoRequest define); unknown
// functions fall back to an empty-string stub after evaluating args.
func (g *Generator) emitHttpclientCall(funcName string, args []ast.Expression) (string, string, error) {
	switch funcName {
	case "NewHttpClient":
		return g.emitHttpclientNewCall(args)
	case "HttpGet", "HttpPost", "HttpPut", "HttpDelete", "HttpGetJSON", "HttpPostJSON":
		return g.emitHttpclientHelperCall(funcName, args)
	case "HttpDoGet", "HttpDoPost":
		return g.emitHttpclientDoCall(funcName, args)
	default:
		for _, a := range args {
			if _, _, err := g.emitExpr(a); err != nil {
				return "", "", err
			}
		}
		emptyStr := g.addString("")
		return g.ptrTo(emptyStr, 1), "ptr", nil
	}
}

// emitHttpclientBody dispatches the deferred body emitter (called by
// emitPendingStdlib for each queued httpclient function).
func (g *Generator) emitHttpclientBody(funcName string) {
	switch funcName {
	case "NewHttpClient":
		g.emitHttpclientNewBody()
	case "SetHeader":
		g.emitHttpclientSetHeaderBody()
	case "Get":
		g.emitHttpclientGetBody()
	case "Post":
		g.emitHttpclientPostBody()
	case "DoRequest":
		g.emitHttpclientDoRequestBody()
	case "Request":
		g.emitHttpclientRequestBody()
	}
}

// emitHttpclientMethodCall handles THttpClient method calls
// (c.SetHeader / c.Get / c.Post). Put/Delete/StatusCode (v6.1.0) delegate to
// the one-shot DoRequest define, extracting the body or the response code.
func (g *Generator) emitHttpclientMethodCall(receiver string, method string, args []ast.Expression) (string, string, error) {
	switch method {
	case "SetHeader":
		return g.emitHttpclientSetHeaderCall(receiver, args)
	case "Get":
		return g.emitHttpclientGetCall(receiver, args)
	case "Post":
		return g.emitHttpclientPostCall(receiver, args)
	case "Put":
		return g.emitHttpclientVerbCall(receiver, "PUT", args, true)
	case "Delete":
		return g.emitHttpclientVerbCall(receiver, "DELETE", args, false)
	case "StatusCode":
		return g.emitHttpclientStatusCodeCall(receiver, args)
	default:
		for _, a := range args {
			if _, _, err := g.emitExpr(a); err != nil {
				return "", "", err
			}
		}
		emptyStr := g.addString("")
		return g.ptrTo(emptyStr, 1), "ptr", nil
	}
}

// emitHttpclientFieldAccess lowers c.<field> on a THttpClient handle.
// BaseURL (handle offset 16) returns the stored String ptr, or empty string
// if null. Other fields are not represented on the handle and return empty
// string (best-effort, matches the prior stub behavior).
func (g *Generator) emitHttpclientFieldAccess(obj ast.Expression, field string) (string, string, error) {
	objReg, _, err := g.emitExpr(obj)
	if err != nil {
		return "", "", err
	}
	if field == "BaseURL" {
		fieldPtr := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %d", fieldPtr, objReg, httpClientOffsetBaseURL))
		baseURL := g.tmp()
		g.line(fmt.Sprintf("  %s = load ptr, ptr %s", baseURL, fieldPtr))
		// If baseURL is null, return empty string (select avoids branches).
		isNull := g.tmp()
		g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", isNull, baseURL))
		emptyStr := g.addString("")
		emptyPtr := g.ptrTo(emptyStr, 1)
		result := g.tmp()
		g.line(fmt.Sprintf("  %s = select i1 %s, ptr %s, ptr %s", result, isNull, emptyPtr, baseURL))
		return result, "ptr", nil
	}
	emptyStr := g.addString("")
	return g.ptrTo(emptyStr, 1), "ptr", nil
}

// ---- NewHttpClient: ptr @__kylix_httpclient_NewHttpClient(ptr %baseURL, i64 %timeout) ----
//
//	curl = curl_easy_init(); if null → ret null (init failed)
//	h = malloc(32); memset 0
//	h[curl]     = curl
//	h[baseURL]  = baseURL  (offset 16)
//	h[timeout]  = timeout  (offset 24)
//	ret h
//
// Matches the Go backend's NewHttpClient(baseURL, timeoutMillis) signature.
// Both args are optional at the call site (defaults: empty baseURL, 10000ms).
func (g *Generator) emitHttpclientNewCall(args []ast.Expression) (string, string, error) {
	// baseURL operand (default: empty string ptr)
	var baseURLOp string
	if len(args) >= 1 {
		r, _, err := g.emitExpr(args[0])
		if err != nil {
			return "", "", err
		}
		baseURLOp = r
	} else {
		emptyStr := g.addString("")
		baseURLOp = g.ptrTo(emptyStr, 1)
	}
	// timeout operand (default: 30 seconds)
	var timeoutOp string
	if len(args) >= 2 {
		r, _, err := g.emitExpr(args[1])
		if err != nil {
			return "", "", err
		}
		timeoutOp = r
	} else {
		timeoutOp = fmt.Sprintf("%d", httpClientDefaultTimeout)
	}
	g.enqueueStdlib("httpclient", "NewHttpClient", "NewHttpClient", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_httpclient_NewHttpClient(ptr %s, i64 %s)", r, baseURLOp, timeoutOp))
	return r, httpClientTypeName, nil
}

func (g *Generator) emitHttpclientNewBody() {
	g.line("define ptr @__kylix_httpclient_NewHttpClient(ptr %baseURL, i64 %timeout) {")
	g.line("entry:")
	curl := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @curl_easy_init()", curl))
	bad := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", bad, curl))
	failLbl := g.label()
	okLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", bad, failLbl, okLbl))
	g.line(fmt.Sprintf("%s:", failLbl))
	g.line("  ret ptr null")
	g.line(fmt.Sprintf("%s:", okLbl))
	h := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %d)", h, httpClientHandleSize))
	g.line(fmt.Sprintf("  call void @llvm.memset.p0.i64(ptr %s, i8 0, i64 %d, i1 false)", h, httpClientHandleSize))
	// store curl at offset 0
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", curl, h))
	// store baseURL at offset 16
	baseURLPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %d", baseURLPtr, h, httpClientOffsetBaseURL))
	g.line(fmt.Sprintf("  store ptr %%baseURL, ptr %s", baseURLPtr))
	// store timeout at offset 24
	timeoutPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %d", timeoutPtr, h, httpClientOffsetTimeout))
	g.line(fmt.Sprintf("  store i64 %%timeout, ptr %s", timeoutPtr))
	g.line(fmt.Sprintf("  ret ptr %s", h))
	g.line("}")
	g.line("")
}

// ---- SetHeader: void @__kylix_httpclient_SetHeader(ptr %self, ptr %k, ptr %v) ----
//
//	hdr = malloc(strlen(k)+strlen(v)+3); strcpy(hdr,k); strcat(hdr,": "); strcat(hdr,v)
//	slist = curl_slist_append(self.slist, hdr); self.slist = slist
//	curl_easy_setopt(self.curl, HTTPHEADER, slist)
//	free(hdr)  // curl_slist_append copies the string
func (g *Generator) emitHttpclientSetHeaderCall(receiver string, args []ast.Expression) (string, string, error) {
	if len(args) != 2 {
		return "", "", fmt.Errorf("THttpClient.SetHeader expects 2 arguments, got %d", len(args))
	}
	kReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	vReg, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("httpclient", "SetHeader", "SetHeader", 0)
	g.line(fmt.Sprintf("  call void @__kylix_httpclient_SetHeader(ptr %s, ptr %s, ptr %s)", receiver, kReg, vReg))
	return "0", "void", nil
}

func (g *Generator) emitHttpclientSetHeaderBody() {
	g.line("define void @__kylix_httpclient_SetHeader(ptr %self, ptr %k, ptr %v) {")
	g.line("entry:")
	klen := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %%k)", klen))
	vlen := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %%v)", vlen))
	sum := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %s", sum, klen, vlen))
	total := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 3", total, sum)) // ": " + null
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", buf, total))
	g.line(fmt.Sprintf("  call ptr @strcpy(ptr %s, ptr %%k)", buf))
	sepStr := g.addString(": ")
	sepPtr := g.ptrTo(sepStr, 3)
	g.line(fmt.Sprintf("  call ptr @strcat(ptr %s, ptr %s)", buf, sepPtr))
	g.line(fmt.Sprintf("  call ptr @strcat(ptr %s, ptr %%v)", buf))
	// self.slist (offset 8) = curl_slist_append(self.slist, buf)
	slistPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%self, i64 %d", slistPtr, httpClientOffsetSlist))
	oldSlist := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", oldSlist, slistPtr))
	newSlist := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @curl_slist_append(ptr %s, ptr %s)", newSlist, oldSlist, buf))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", newSlist, slistPtr))
	// curl_easy_setopt(self.curl, HTTPHEADER, slist)
	curlPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%self, i64 %d", curlPtr, httpClientOffsetCurl))
	curl := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", curl, curlPtr))
	g.line(fmt.Sprintf("  call i32 (ptr, i32, ...) @curl_easy_setopt(ptr %s, i32 %d, ptr %s)", curl, curloptHttpheader, newSlist))
	// slist copied the string; free our scratch buffer
	g.line(fmt.Sprintf("  call void @free(ptr %s)", buf))
	g.line("  ret void")
	g.line("}")
	g.line("")
}

// emitHttpclientWriteCallbackBody emits the libcurl write callback that
// accumulates received bytes into a growable response buffer. userdata points
// to a 24-byte struct { ptr data; i64 len; i64 cap } allocated by the caller
// (Get/Post) on its stack. Idempotent — guarded via stdlibEmitted so it emits
// at most once per module (shared by Get and Post).
//
//	define i64 @__kylix_http_write_cb(ptr %data, i64 %size, i64 %nmemb, ptr %userdata)
//
// Returns size*nmemb on success (libcurl contract) or 0 on realloc failure
// (which aborts the transfer). The response buffer is grown with realloc to
// exactly the needed size (curLen + realsize + 1 for the null terminator);
// this is simple and correct, at the cost of one realloc per chunk.
func (g *Generator) emitHttpclientWriteCallbackBody() {
	const key = "httpclient.__write_cb"
	if g.stdlibEmitted[key] {
		return
	}
	g.stdlibEmitted[key] = true
	// realloc is not in codegen.go's libc declare set; declare it locally.
	// LLVM permits module-scope declares anywhere; the guard above ensures it
	// fires at most once.
	g.line("declare ptr @realloc(ptr noundef, i64 noundef)")
	g.line("define i64 @__kylix_http_write_cb(ptr %data, i64 %size, i64 %nmemb, ptr %userdata) {")
	g.line("entry:")
	realsize := g.tmp()
	g.line(fmt.Sprintf("  %s = mul i64 %%size, %%nmemb", realsize))
	// load cur{Data,Len,Cap} from userdata
	dataField := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%userdata, i64 0", dataField))
	curData := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", curData, dataField))
	lenField := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%userdata, i64 8", lenField))
	curLen := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curLen, lenField))
	capField := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%userdata, i64 16", capField))
	curCap := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curCap, capField))
	// needed = curLen + realsize + 1
	add1 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %s", add1, curLen, realsize))
	needed := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", needed, add1))
	// if needed <= curCap → copy directly, else grow
	fits := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sle i64 %s, %s", fits, needed, curCap))
	growLbl := g.label()
	copyLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", fits, copyLbl, growLbl))
	// grow: newData = realloc(curData, needed)
	g.line(fmt.Sprintf("%s:", growLbl))
	newData := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @realloc(ptr %s, i64 %s)", newData, curData, needed))
	isNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", isNull, newData))
	failLbl := g.label()
	okLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isNull, failLbl, okLbl))
	g.line(fmt.Sprintf("%s:", failLbl))
	g.line("  ret i64 0") // realloc failed → abort transfer
	g.line(fmt.Sprintf("%s:", okLbl))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", newData, dataField))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", needed, capField))
	g.line(fmt.Sprintf("  br label %%%s", copyLbl))
	// copy: reload data/len (data may have changed in grow), append bytes
	g.line(fmt.Sprintf("%s:", copyLbl))
	d2 := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", d2, dataField))
	l2 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", l2, lenField))
	dst := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", dst, d2, l2))
	g.line(fmt.Sprintf("  call ptr @memcpy(ptr %s, ptr %%data, i64 %s)", dst, realsize))
	l3 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %s", l3, l2, realsize))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", l3, lenField))
	term := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", term, d2, l3))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", term))
	g.line(fmt.Sprintf("  ret i64 %s", realsize))
	g.line("}")
	g.line("")
}

// ---- Get: ptr @__kylix_httpclient_Get(ptr %self, ptr %url) ----
//
//	respbuf = alloca [24 x i8]; memset 0   // {data,len,cap} for write_cb
//	curl = self.curl; timeout = self.timeout
//	setopt(URL, url); setopt(WRITEFUNCTION, @__kylix_http_write_cb)
//	setopt(WRITEDATA, respbuf); setopt(TIMEOUT, timeout)
//	curl_easy_perform(curl)
//	data = respbuf.data; if null → ret empty string; else ret data
func (g *Generator) emitHttpclientGetCall(receiver string, args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("THttpClient.Get expects 1 argument, got %d", len(args))
	}
	urlReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("httpclient", "Get", "Get", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_httpclient_Get(ptr %s, ptr %s)", r, receiver, urlReg))
	return r, "ptr", nil
}

func (g *Generator) emitHttpclientGetBody() {
	// Emit the shared write callback first (idempotent).
	g.emitHttpclientWriteCallbackBody()
	g.line("define ptr @__kylix_httpclient_Get(ptr %self, ptr %url) {")
	g.line("entry:")
	// respbuf: 24-byte {ptr data; i64 len; i64 cap} on the stack, zeroed.
	resp := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [24 x i8], align 8", resp))
	g.line(fmt.Sprintf("  call void @llvm.memset.p0.i64(ptr %s, i8 0, i64 24, i1 false)", resp))
	// curl = self.curl (offset 0)
	curlPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%self, i64 %d", curlPtr, httpClientOffsetCurl))
	curl := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", curl, curlPtr))
	// timeout = self.timeout (offset 24)
	timeoutPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%self, i64 %d", timeoutPtr, httpClientOffsetTimeout))
	timeout := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", timeout, timeoutPtr))
	// setopt: URL, WRITEFUNCTION, WRITEDATA, TIMEOUT_MS (handle stores ms; v6.1.0)
	g.line(fmt.Sprintf("  call i32 (ptr, i32, ...) @curl_easy_setopt(ptr %s, i32 %d, ptr %%url)", curl, curloptUrl))
	g.line(fmt.Sprintf("  call i32 (ptr, i32, ...) @curl_easy_setopt(ptr %s, i32 %d, ptr @__kylix_http_write_cb)", curl, curloptWritefunction))
	g.line(fmt.Sprintf("  call i32 (ptr, i32, ...) @curl_easy_setopt(ptr %s, i32 %d, ptr %s)", curl, curloptWrdata, resp))
	g.line(fmt.Sprintf("  call i32 (ptr, i32, ...) @curl_easy_setopt(ptr %s, i32 %d, i64 %s)", curl, curloptTimeoutMs, timeout))
	// perform
	g.line(fmt.Sprintf("  call i32 @curl_easy_perform(ptr %s)", curl))
	// load resp data; null → empty string
	dataField := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 0", dataField, resp))
	data := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", data, dataField))
	isNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", isNull, data))
	emptyStr := g.addString("")
	emptyPtr := g.ptrTo(emptyStr, 1)
	result := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, ptr %s, ptr %s", result, isNull, emptyPtr, data))
	g.line(fmt.Sprintf("  ret ptr %s", result))
	g.line("}")
	g.line("")
}

// ---- Post: ptr @__kylix_httpclient_Post(ptr %self, ptr %url, ptr %ct, ptr %body) ----
//
//	Same as Get, plus:
//	  if ct != "" → slist = curl_slist_append(null, "Content-Type: "+ct); setopt(HTTPHEADER, slist)
//	  setopt(POST, 1); setopt(POSTFIELDS, body)
//
// v6.1.0: signature matched to the Kylix declaration Post(path, contentType,
// body). The 2-arg form Post(url, body) is still accepted (contentType empty).
func (g *Generator) emitHttpclientPostCall(receiver string, args []ast.Expression) (string, string, error) {
	if len(args) != 2 && len(args) != 3 {
		return "", "", fmt.Errorf("THttpClient.Post expects 2 or 3 arguments (path, body) or (path, contentType, body), got %d", len(args))
	}
	urlReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	var ctOp, bodyOp string
	if len(args) == 3 {
		ctReg, _, err := g.emitExpr(args[1])
		if err != nil {
			return "", "", err
		}
		bodyReg, _, err := g.emitExpr(args[2])
		if err != nil {
			return "", "", err
		}
		ctOp, bodyOp = ctReg, bodyReg
	} else {
		emptyStr := g.addString("")
		ctOp = g.ptrTo(emptyStr, 1)
		bodyReg, _, err := g.emitExpr(args[1])
		if err != nil {
			return "", "", err
		}
		bodyOp = bodyReg
	}
	g.enqueueStdlib("httpclient", "Post", "Post", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_httpclient_Post(ptr %s, ptr %s, ptr %s, ptr %s)", r, receiver, urlReg, ctOp, bodyOp))
	return r, "ptr", nil
}

func (g *Generator) emitHttpclientPostBody() {
	g.emitHttpclientWriteCallbackBody()
	g.line("define ptr @__kylix_httpclient_Post(ptr %self, ptr %url, ptr %ct, ptr %body) {")
	g.line("entry:")
	// Content-Type header (skip when empty)
	ctLen := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %%ct)", ctLen))
	ctZero := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, 0", ctZero, ctLen))
	noCtLbl := g.label()
	withCtLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", ctZero, noCtLbl, withCtLbl))
	g.line(fmt.Sprintf("%s:", withCtLbl))
	ctSum := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 15", ctSum, ctLen)) // "Content-Type: " (14) + ct + null
	hdrBuf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", hdrBuf, ctSum))
	ctPrefix := g.addString("Content-Type: ")
	g.line(fmt.Sprintf("  call ptr @strcpy(ptr %s, ptr %s)", hdrBuf, g.ptrTo(ctPrefix, 15)))
	g.line(fmt.Sprintf("  call ptr @strcat(ptr %s, ptr %%ct)", hdrBuf))
	ctSlist := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @curl_slist_append(ptr null, ptr %s)", ctSlist, hdrBuf))
	g.line(fmt.Sprintf("  call void @free(ptr %s)", hdrBuf))
	// curl handle (needed below for setopt)
	curlPtr2 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%self, i64 %d", curlPtr2, httpClientOffsetCurl))
	curl2 := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", curl2, curlPtr2))
	g.line(fmt.Sprintf("  call i32 (ptr, i32, ...) @curl_easy_setopt(ptr %s, i32 %d, ptr %s)", curl2, curloptHttpheader, ctSlist))
	g.line(fmt.Sprintf("  br label %%%s", noCtLbl))
	g.line(fmt.Sprintf("%s:", noCtLbl))
	resp := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [24 x i8], align 8", resp))
	g.line(fmt.Sprintf("  call void @llvm.memset.p0.i64(ptr %s, i8 0, i64 24, i1 false)", resp))
	curlPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%self, i64 %d", curlPtr, httpClientOffsetCurl))
	curl := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", curl, curlPtr))
	timeoutPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%self, i64 %d", timeoutPtr, httpClientOffsetTimeout))
	timeout := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", timeout, timeoutPtr))
	// setopt: URL, WRITEFUNCTION, WRITEDATA, TIMEOUT_MS, POST, POSTFIELDS
	g.line(fmt.Sprintf("  call i32 (ptr, i32, ...) @curl_easy_setopt(ptr %s, i32 %d, ptr %%url)", curl, curloptUrl))
	g.line(fmt.Sprintf("  call i32 (ptr, i32, ...) @curl_easy_setopt(ptr %s, i32 %d, ptr @__kylix_http_write_cb)", curl, curloptWritefunction))
	g.line(fmt.Sprintf("  call i32 (ptr, i32, ...) @curl_easy_setopt(ptr %s, i32 %d, ptr %s)", curl, curloptWrdata, resp))
	g.line(fmt.Sprintf("  call i32 (ptr, i32, ...) @curl_easy_setopt(ptr %s, i32 %d, i64 %s)", curl, curloptTimeoutMs, timeout))
	g.line(fmt.Sprintf("  call i32 (ptr, i32, ...) @curl_easy_setopt(ptr %s, i32 %d, i64 1)", curl, curloptPost))
	g.line(fmt.Sprintf("  call i32 (ptr, i32, ...) @curl_easy_setopt(ptr %s, i32 %d, ptr %%body)", curl, curloptPostfields))
	g.line(fmt.Sprintf("  call i32 @curl_easy_perform(ptr %s)", curl))
	dataField := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 0", dataField, resp))
	data := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", data, dataField))
	isNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", isNull, data))
	emptyStr := g.addString("")
	emptyPtr := g.ptrTo(emptyStr, 1)
	result := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, ptr %s, ptr %s", result, isNull, emptyPtr, data))
	g.line(fmt.Sprintf("  ret ptr %s", result))
	g.line("}")
	g.line("")
}

// ---- One-shot helpers (HttpGet/HttpPost/HttpPut/HttpDelete/HttpGetJSON/
// HttpPostJSON) — real libcurl IR via the shared DoRequest define (v6.1.0).
// HttpGetJSON/HttpPostJSON parse the response body (a JSON object) into a
// Variant map (JsonDecodeMap → box_map) so callers can do m['key'] with
// variant-map indexing; the other helpers return the raw body string. v6.6.0.

func (g *Generator) emitHttpclientHelperCall(funcName string, args []ast.Expression) (string, string, error) {
	var method string
	switch funcName {
	case "HttpGet", "HttpGetJSON":
		method = "GET"
	case "HttpPost", "HttpPostJSON":
		method = "POST"
	case "HttpPut":
		method = "PUT"
	case "HttpDelete":
		method = "DELETE"
	}
	if len(args) < 1 {
		return "", "", fmt.Errorf("httpclient.%s expects a URL argument", funcName)
	}
	urlReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	emptyStr := g.addString("")
	emptyPtr := g.ptrTo(emptyStr, 1)
	ctOp, bodyOp := emptyPtr, emptyPtr
	switch funcName {
	case "HttpPost", "HttpPut":
		if len(args) < 3 {
			return "", "", fmt.Errorf("httpclient.%s expects (url, contentType, body), got %d", funcName, len(args))
		}
		ctReg, _, err := g.emitExpr(args[1])
		if err != nil {
			return "", "", err
		}
		bodyReg, _, err := g.emitExpr(args[2])
		if err != nil {
			return "", "", err
		}
		ctOp, bodyOp = ctReg, bodyReg
	case "HttpPostJSON":
		if len(args) < 2 {
			return "", "", fmt.Errorf("httpclient.HttpPostJSON expects (url, body), got %d", len(args))
		}
		bodyReg, _, err := g.emitExpr(args[1])
		if err != nil {
			return "", "", err
		}
		bodyOp = bodyReg
	}
	g.enqueueStdlib("httpclient", "Request", "Request", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_httpclient_Request(ptr %s, ptr %s, ptr %s, ptr %s, i64 %d)",
		r, g.addString(method), urlReg, bodyOp, ctOp, httpClientDefaultTimeout))
	// v6.6.0: HttpGetJSON/HttpPostJSON parse the response body (a JSON object)
	// into a Variant map. JsonDecodeMap → parse_flat already boxes every value
	// via value_to_variant, so box_map gives a map[String]Variant.
	if funcName == "HttpGetJSON" || funcName == "HttpPostJSON" {
		g.enqueueStdlib("jsonutil", "JsonDecodeMap", "JsonDecodeMap", 0)
		g.needVariantRuntime = true
		htab := g.tmp()
		g.line(fmt.Sprintf("  %s = call ptr @__kylix_json_JsonDecodeMap(ptr %s)", htab, r))
		box := g.tmp()
		g.line(fmt.Sprintf("  %s = call ptr @__kylix_variant_box_map(ptr %s)", box, htab))
		return box, variantT, nil
	}
	return r, "ptr", nil
}

// ---- HttpDoGet / HttpDoPost: {i64 status, ptr body} THttpResponse handle.

func (g *Generator) emitHttpclientDoCall(funcName string, args []ast.Expression) (string, string, error) {
	method := "GET"
	if funcName == "HttpDoPost" {
		method = "POST"
	}
	if len(args) < 1 {
		return "", "", fmt.Errorf("httpclient.%s expects a URL argument", funcName)
	}
	urlReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	emptyStr := g.addString("")
	emptyPtr := g.ptrTo(emptyStr, 1)
	ctOp, bodyOp := emptyPtr, emptyPtr
	if funcName == "HttpDoPost" {
		if len(args) >= 3 {
			ctReg, _, err := g.emitExpr(args[1])
			if err != nil {
				return "", "", err
			}
			bodyReg, _, err := g.emitExpr(args[2])
			if err != nil {
				return "", "", err
			}
			ctOp, bodyOp = ctReg, bodyReg
		}
	}
	g.enqueueStdlib("httpclient", "DoRequest", "DoRequest", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call { i64, ptr } @__kylix_httpclient_DoRequest(ptr %s, ptr %s, ptr %s, ptr %s, i64 %d)",
		r, g.addString(method), urlReg, bodyOp, ctOp, httpClientDefaultTimeout))
	return r, "THttpResponse", nil
}

// emitHttpclientVerbCall lowers c.Put(path, contentType, body) / c.Delete(path)
// to a DoRequest on the client's own handle, extracting the body string.
func (g *Generator) emitHttpclientVerbCall(receiver, verb string, args []ast.Expression, hasBody bool) (string, string, error) {
	if len(args) < 1 {
		return "", "", fmt.Errorf("THttpClient.%s expects a path argument", verb)
	}
	urlReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	emptyStr := g.addString("")
	emptyPtr := g.ptrTo(emptyStr, 1)
	ctOp, bodyOp := emptyPtr, emptyPtr
	if hasBody {
		if len(args) < 3 {
			return "", "", fmt.Errorf("THttpClient.Put expects (path, contentType, body), got %d", len(args))
		}
		ctReg, _, err := g.emitExpr(args[1])
		if err != nil {
			return "", "", err
		}
		bodyReg, _, err := g.emitExpr(args[2])
		if err != nil {
			return "", "", err
		}
		ctOp, bodyOp = ctReg, bodyReg
	}
	timeoutPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %d", timeoutPtr, receiver, httpClientOffsetTimeout))
	timeout := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", timeout, timeoutPtr))
	g.enqueueStdlib("httpclient", "DoRequest", "DoRequest", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call { i64, ptr } @__kylix_httpclient_DoRequest(ptr %s, ptr %s, ptr %s, ptr %s, i64 %s)",
		r, g.addString(verb), urlReg, bodyOp, ctOp, timeout))
	body := g.tmp()
	g.line(fmt.Sprintf("  %s = extractvalue { i64, ptr } %s, 1", body, r))
	return body, "ptr", nil
}

// emitHttpclientStatusCodeCall lowers c.StatusCode(path) to a GET via
// DoRequest, extracting the HTTP response code.
func (g *Generator) emitHttpclientStatusCodeCall(receiver string, args []ast.Expression) (string, string, error) {
	if len(args) < 1 {
		return "", "", fmt.Errorf("THttpClient.StatusCode expects a path argument")
	}
	urlReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	emptyStr := g.addString("")
	emptyPtr := g.ptrTo(emptyStr, 1)
	timeoutPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %d", timeoutPtr, receiver, httpClientOffsetTimeout))
	timeout := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", timeout, timeoutPtr))
	g.enqueueStdlib("httpclient", "DoRequest", "DoRequest", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call { i64, ptr } @__kylix_httpclient_DoRequest(ptr %s, ptr %s, ptr %s, ptr %s, i64 %s)",
		r, g.addString("GET"), urlReg, emptyPtr, emptyPtr, timeout))
	code := g.tmp()
	g.line(fmt.Sprintf("  %s = extractvalue { i64, ptr } %s, 0", code, r))
	return code, "i64", nil
}

// ---- DoRequest: {i64, ptr} @__kylix_httpclient_DoRequest(ptr %method, ptr %url, ptr %body, ptr %ct, i64 %timeoutMs)
//
//	curl = curl_easy_init()
//	[optional] "Content-Type: "+ct → slist → setopt(HTTPHEADER)
//	setopt(URL, url); setopt(CUSTOMREQUEST, method); setopt(WRITEFUNCTION, write_cb)
//	setopt(WRITEDATA, resp); setopt(TIMEOUT_MS, timeoutMs)
//	[optional] setopt(POSTFIELDS, body)
//	perform; getinfo(RESPONSE_CODE); body = write_cb buffer
//	ret {status, body}

func (g *Generator) emitHttpclientDoRequestBody() {
	g.emitHttpclientWriteCallbackBody()
	// curl_easy_getinfo is not in codegen.go's declare set — declare lazily.
	g.line("declare i32 @curl_easy_getinfo(ptr noundef, i32 noundef, ptr noundef)")
	g.line("define { i64, ptr } @__kylix_httpclient_DoRequest(ptr %method, ptr %url, ptr %body, ptr %ct, i64 %timeoutMs) {")
	g.line("entry:")
	curl := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @curl_easy_init()", curl))
	// Content-Type header (skip when ct is empty)
	ctLen := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %%ct)", ctLen))
	ctZero := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, 0", ctZero, ctLen))
	noCtLbl := g.label()
	withCtLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", ctZero, noCtLbl, withCtLbl))
	g.line(fmt.Sprintf("%s:", withCtLbl))
	ctSum := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 15", ctSum, ctLen)) // "Content-Type: " (14) + ct + null
	hdrBuf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", hdrBuf, ctSum))
	ctPrefix := g.addString("Content-Type: ")
	g.line(fmt.Sprintf("  call ptr @strcpy(ptr %s, ptr %s)", hdrBuf, g.ptrTo(ctPrefix, 15)))
	g.line(fmt.Sprintf("  call ptr @strcat(ptr %s, ptr %%ct)", hdrBuf))
	ctSlist := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @curl_slist_append(ptr null, ptr %s)", ctSlist, hdrBuf))
	g.line(fmt.Sprintf("  call void @free(ptr %s)", hdrBuf))
	g.line(fmt.Sprintf("  call i32 (ptr, i32, ...) @curl_easy_setopt(ptr %s, i32 %d, ptr %s)", curl, curloptHttpheader, ctSlist))
	g.line(fmt.Sprintf("  br label %%%s", noCtLbl))
	g.line(fmt.Sprintf("%s:", noCtLbl))
	resp := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [24 x i8], align 8", resp))
	g.line(fmt.Sprintf("  call void @llvm.memset.p0.i64(ptr %s, i8 0, i64 24, i1 false)", resp))
	g.line(fmt.Sprintf("  call i32 (ptr, i32, ...) @curl_easy_setopt(ptr %s, i32 %d, ptr %%url)", curl, curloptUrl))
	g.line(fmt.Sprintf("  call i32 (ptr, i32, ...) @curl_easy_setopt(ptr %s, i32 %d, ptr %%method)", curl, curloptCustomrequest))
	g.line(fmt.Sprintf("  call i32 (ptr, i32, ...) @curl_easy_setopt(ptr %s, i32 %d, ptr @__kylix_http_write_cb)", curl, curloptWritefunction))
	g.line(fmt.Sprintf("  call i32 (ptr, i32, ...) @curl_easy_setopt(ptr %s, i32 %d, ptr %s)", curl, curloptWrdata, resp))
	g.line(fmt.Sprintf("  call i32 (ptr, i32, ...) @curl_easy_setopt(ptr %s, i32 %d, i64 %%timeoutMs)", curl, curloptTimeoutMs))
	// POST/PUT payload (skip when body is empty)
	bodyLen := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %%body)", bodyLen))
	bodyZero := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, 0", bodyZero, bodyLen))
	noBodyLbl := g.label()
	withBodyLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", bodyZero, noBodyLbl, withBodyLbl))
	g.line(fmt.Sprintf("%s:", withBodyLbl))
	g.line(fmt.Sprintf("  call i32 (ptr, i32, ...) @curl_easy_setopt(ptr %s, i32 %d, ptr %%body)", curl, curloptPostfields))
	g.line(fmt.Sprintf("  br label %%%s", noBodyLbl))
	g.line(fmt.Sprintf("%s:", noBodyLbl))
	g.line(fmt.Sprintf("  call i32 @curl_easy_perform(ptr %s)", curl))
	// response code
	codeSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", codeSlot))
	g.line(fmt.Sprintf("  call i32 @curl_easy_getinfo(ptr %s, i32 %d, ptr %s)", curl, curlInfoResponseCode, codeSlot))
	code := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", code, codeSlot))
	// response body
	dataField := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 0", dataField, resp))
	data := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", data, dataField))
	isNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", isNull, data))
	emptyStr := g.addString("")
	emptyPtr := g.ptrTo(emptyStr, 1)
	bodyVal := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, ptr %s, ptr %s", bodyVal, isNull, emptyPtr, data))
	// pack {i64, ptr}
	res1 := g.tmp()
	g.line(fmt.Sprintf("  %s = insertvalue { i64, ptr } undef, i64 %s, 0", res1, code))
	res2 := g.tmp()
	g.line(fmt.Sprintf("  %s = insertvalue { i64, ptr } %s, ptr %s, 1", res2, res1, bodyVal))
	g.line(fmt.Sprintf("  ret { i64, ptr } %s", res2))
	g.line("}")
	g.line("")
}

// ---- Request: ptr @__kylix_httpclient_Request(ptr %method, ptr %url, ptr %body, ptr %ct, i64 %timeoutMs)
//
//	Thin wrapper over DoRequest returning just the body string.

func (g *Generator) emitHttpclientRequestBody() {
	// The Request define is a thin wrapper over DoRequest — ensure that body
	// is emitted too (one-shot helpers HttpGet/HttpPost/... go through here).
	// v6.6.0: previously DoRequest could end up undefined when only a one-shot
	// helper was used.
	g.enqueueStdlib("httpclient", "DoRequest", "DoRequest", 0)
	g.line("define ptr @__kylix_httpclient_Request(ptr %method, ptr %url, ptr %body, ptr %ct, i64 %timeoutMs) {")
	g.line("entry:")
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call { i64, ptr } @__kylix_httpclient_DoRequest(ptr %%method, ptr %%url, ptr %%body, ptr %%ct, i64 %%timeoutMs)", r))
	b := g.tmp()
	g.line(fmt.Sprintf("  %s = extractvalue { i64, ptr } %s, 1", b, r))
	g.line(fmt.Sprintf("  ret ptr %s", b))
	g.line("}")
	g.line("")
}

// emitHttpclientResponseFieldAccess lowers `resp.Status` / `resp.Body` on a
// THttpResponse ({i64 status, ptr body}) to extractvalue.
func (g *Generator) emitHttpclientResponseFieldAccess(obj ast.Expression, field string) (string, string, error) {
	r, _, err := g.emitExpr(obj)
	if err != nil {
		return "", "", err
	}
	switch strings.ToLower(field) {
	case "status":
		v := g.tmp()
		g.line(fmt.Sprintf("  %s = extractvalue { i64, ptr } %s, 0", v, r))
		return v, "i64", nil
	case "body":
		v := g.tmp()
		g.line(fmt.Sprintf("  %s = extractvalue { i64, ptr } %s, 1", v, r))
		return v, "ptr", nil
	default:
		v := g.tmp()
		g.line(fmt.Sprintf("  %s = inttoptr i64 0 to ptr ; THttpResponse.%s unsupported", v, field))
		return v, "ptr", nil
	}
}
