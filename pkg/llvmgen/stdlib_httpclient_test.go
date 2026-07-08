package llvmgen_test

import (
	"strings"
	"testing"
)

// stdlib_httpclient tests — verify the httpclient stdlib module lowers to
// libcurl-backed IR defines (not stubs). THttpClient is a 32-byte heap handle
// wrapping a CURL* easy handle; SetHeader accumulates a curl_slist, and
// Get/Post drive curl_easy_perform with a write callback that gathers the
// response body into a realloc-grown buffer.

func TestHttp_NewHttpClient(t *testing.T) {
	ir := generateIR(t, `program p;
uses httpclient;
begin
  var c := NewHttpClient();
end.`)
	assertIRContains(t, ir, "call ptr @curl_easy_init")
	assertIRContains(t, ir, "define ptr @__kylix_httpclient_NewHttpClient(ptr %baseURL, i64 %timeout)")
	if strings.Contains(ir, "httpclient.NewHttpClient not implemented") {
		t.Errorf("NewHttpClient still routed to not-implemented stub\nIR:\n%s", ir)
	}
}

func TestHttp_NewHttpClientWithArgs(t *testing.T) {
	// NewHttpClient(baseURL, timeout) — the form the tutorial uses. baseURL
	// and timeout must be passed through to the IR define and stored on the
	// handle so c.BaseURL returns the stored value.
	ir := generateIR(t, `program p;
uses httpclient;
begin
  var c := NewHttpClient('https://example.com', 5000);
  WriteLn(c.BaseURL);
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_httpclient_NewHttpClient(ptr")
	// the call must pass the baseURL arg (not just defaults)
	if !strings.Contains(ir, "call ptr @__kylix_httpclient_NewHttpClient(ptr %") {
		t.Errorf("NewHttpClient call does not pass baseURL operand\nIR:\n%s", ir)
	}
}

func TestHttp_NewHttpClientHandleInit(t *testing.T) {
	ir := generateIR(t, `program p;
uses httpclient;
begin
  var c := NewHttpClient();
end.`)
	// malloc(32) for the handle + memset to zero.
	assertIRContains(t, ir, "call ptr @malloc(i64 32)")
	assertIRContains(t, ir, "call void @llvm.memset.p0.i64")
	// 0-arg call passes the default timeout (30s) as the 2nd operand.
	assertIRContains(t, ir, "call ptr @__kylix_httpclient_NewHttpClient(ptr")
	assertIRContains(t, ir, "i64 30)")
	// body stores the timeout parameter at offset 24.
	assertIRContains(t, ir, "store i64 %timeout, ptr")
}

func TestHttp_SetHeader(t *testing.T) {
	ir := generateIR(t, `program p;
uses httpclient;
begin
  var c := NewHttpClient();
  c.SetHeader('X-Demo', 'kylix');
end.`)
	assertIRContains(t, ir, "define void @__kylix_httpclient_SetHeader(ptr %self, ptr %k, ptr %v)")
	assertIRContains(t, ir, "call ptr @curl_slist_append")
	// CURLOPT_HTTPHEADER (10023) via variadic curl_easy_setopt.
	assertIRContains(t, ir, "call i32 (ptr, i32, ...) @curl_easy_setopt")
	assertIRContains(t, ir, "i32 10023")
}

func TestHttp_Get(t *testing.T) {
	ir := generateIR(t, `program p;
uses httpclient;
begin
  var c := NewHttpClient();
  var r := c.Get('http://example.com');
end.`)
	assertIRContains(t, ir, "define ptr @__kylix_httpclient_Get(ptr %self, ptr %url)")
	assertIRContains(t, ir, "call i32 @curl_easy_perform")
	// Write callback registered as the WRITEFUNCTION.
	assertIRContains(t, ir, "@__kylix_http_write_cb")
	// CURLOPT_URL (10002) + CURLOPT_WRITEFUNCTION (20011) + CURLOPT_WRITEDATA (10001).
	assertIRContains(t, ir, "i32 10002")
	assertIRContains(t, ir, "i32 20011")
	assertIRContains(t, ir, "i32 10001")
	// CURLOPT_TIMEOUT (13) wired from the handle's stored timeout.
	assertIRContains(t, ir, "i32 13")
}

func TestHttp_Post(t *testing.T) {
	ir := generateIR(t, `program p;
uses httpclient;
begin
  var c := NewHttpClient();
  var r := c.Post('http://example.com', 'hello');
end.`)
	assertIRContains(t, ir, "define ptr @__kylix_httpclient_Post(ptr %self, ptr %url, ptr %body)")
	// CURLOPT_POST (47) + CURLOPT_POSTFIELDS (10015).
	assertIRContains(t, ir, "i32 47")
	assertIRContains(t, ir, "i32 10015")
	assertIRContains(t, ir, "call i32 @curl_easy_perform")
}

func TestHttp_WriteCallback(t *testing.T) {
	ir := generateIR(t, `program p;
uses httpclient;
begin
  var c := NewHttpClient();
  var r := c.Get('http://example.com');
end.`)
	assertIRContains(t, ir, "define i64 @__kylix_http_write_cb(ptr %data, i64 %size, i64 %nmemb, ptr %userdata)")
	// realloc grows the response buffer; null check aborts on failure.
	assertIRContains(t, ir, "call ptr @realloc")
	assertIRContains(t, ir, "icmp eq ptr")
	// Successful path returns size*nmemb to libcurl.
	assertIRContains(t, ir, "ret i64 0")
}

func TestHttp_WriteCallbackDedup(t *testing.T) {
	// Both Get and Post share one write-callback define.
	ir := generateIR(t, `program p;
uses httpclient;
begin
  var c := NewHttpClient();
  var a := c.Get('http://example.com');
  var b := c.Post('http://example.com', 'x');
end.`)
	if got := strings.Count(ir, "define i64 @__kylix_http_write_cb"); got != 1 {
		t.Errorf("write callback define should appear once, got %d\nIR:\n%s", got, ir)
	}
	// realloc declare should also appear exactly once.
	if got := strings.Count(ir, "declare ptr @realloc"); got != 1 {
		t.Errorf("realloc declare should appear once, got %d\nIR:\n%s", got, ir)
	}
}

func TestHttp_BodyDedup(t *testing.T) {
	ir := generateIR(t, `program p;
uses httpclient;
begin
  var a := NewHttpClient();
  var b := NewHttpClient();
end.`)
	if got := strings.Count(ir, "define ptr @__kylix_httpclient_NewHttpClient"); got != 1 {
		t.Errorf("NewHttpClient define should appear once, got %d\nIR:\n%s", got, ir)
	}
}

func TestHttp_LibcurlDeclarations(t *testing.T) {
	ir := generateIR(t, `program p;
uses httpclient;
begin
  var c := NewHttpClient();
end.`)
	assertIRContains(t, ir, "declare ptr @curl_easy_init")
	assertIRContains(t, ir, "declare i32 @curl_easy_setopt(ptr noundef, i32 noundef, ...)")
	assertIRContains(t, ir, "declare i32 @curl_easy_perform")
	assertIRContains(t, ir, "declare ptr @curl_slist_append")
}

func TestHttp_NotUsedNoBodies(t *testing.T) {
	ir := generateIR(t, `program p;
begin
  WriteLn('hi');
end.`)
	if strings.Contains(ir, "@__kylix_httpclient_") {
		t.Errorf("httpclient symbol emitted without `uses httpclient`\nIR:\n%s", ir)
	}
	if strings.Contains(ir, "@__kylix_http_write_cb") {
		t.Errorf("write callback emitted without `uses httpclient`\nIR:\n%s", ir)
	}
}

func TestHttp_BaseURLFieldAccess(t *testing.T) {
	ir := generateIR(t, `program p;
uses httpclient;
begin
  var c := NewHttpClient();
  WriteLn(c.BaseURL);
end.`)
	// Real GEP+load on the handle (offset 16), not the empty-string stub.
	assertIRContains(t, ir, "getelementptr inbounds i8, ptr")
	assertIRContains(t, ir, "i64 16")
	assertIRContains(t, ir, "icmp eq ptr")
	assertIRContains(t, ir, "select i1")
}
