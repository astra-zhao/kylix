package llvmgen_test

import (
	"testing"
)

// stdlib_boot_http tests — verify the v6.6.0 KylixBoot HTTP server lowers to
// real IR: route registration into @__kylix_boot_routes, a BootRun loop using
// BSD sockets, and inline req.Param/Query/Header lookups.

const bootServerProgram = `program p;
uses boot;
[Controller('/api')]
type
  TApiController = class
    [Get('/hello')]
    function Hello(req: TRequest): TResponse;
    begin
      result := BootText(200, 'hi');
    end;
  end;
begin
  BootRun(8080);
end.`

func TestBoot_RunEmitsRealServer(t *testing.T) {
	ir := generateIR(t, bootServerProgram)
	// BootRun is a real call + define (not the add i64 0,0 stub), backed by
	// BSD sockets (net TcpListen/TcpAccept) and the HTTP helpers.
	assertIRContains(t, ir, "call i64 @__kylix_boot_BootRun(i64")
	assertIRContains(t, ir, "define i64 @__kylix_boot_BootRun(i64 %port)")
	assertIRContains(t, ir, "call ptr @__kylix_net_TcpListen(i64")
	assertIRContains(t, ir, "define ptr @__kylix_boot_read_headers(ptr %conn)")
	assertIRContains(t, ir, "define ptr @__kylix_boot_route_lookup(ptr %method, ptr %path, ptr %req)")
}

func TestBoot_RouteTableGlobals(t *testing.T) {
	ir := generateIR(t, bootServerProgram)
	// Route table + counter are declared and written by Boot<M>.
	assertIRContains(t, ir, "@__kylix_boot_routes = global [64 x { ptr, ptr, ptr }] zeroinitializer")
	assertIRContains(t, ir, "@__kylix_boot_nroutes = global i64 0")
	assertIRContains(t, ir, "define void @__kylix_boot_BootGET(ptr %path, ptr %wrapper)")
	assertIRContains(t, ir, "getelementptr inbounds [64 x { ptr, ptr, ptr }], ptr @__kylix_boot_routes")
}

func TestBoot_ReqParamAndQueryInlined(t *testing.T) {
	ir := generateIR(t, `program p;
uses boot;
[Controller('/api')]
type
  TApiController = class
    [Get('/users/:id')]
    function User(req: TRequest): TResponse;
    begin
      result := BootText(200, req.Param('id') + '?' + req.Query('page'));
    end;
  end;
begin
  WriteLn('ok');
end.`)
	// req.Param/Query/Header lower to inline strcmp/strchr lookups on the
	// request handle (never a stub).
	assertIRContains(t, ir, "call i32 @strcmp(ptr")
	assertIRContains(t, ir, "call ptr @strchr(ptr")
	if contains(t, ir, "TRequest.Param stub") || contains(t, ir, "TRequest.Query stub") {
		t.Errorf("TRequest methods still routed to stubs\nIR:\n%s", ir)
	}
}

func contains(t *testing.T, s, sub string) bool {
	t.Helper()
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// ---- v6.8.0: POST body read / req.JSON / BootRegisterJwtAuth ----

func TestBoot_ReadBodyReal(t *testing.T) {
	ir := generateIR(t, `program p;
uses boot;
[Controller('/api')]
type
  TApiController = class
    [Post('/echo')]
    function Echo(req: TRequest): TResponse;
    begin
      result := BootText(200, req.Body());
    end;
  end;
begin
  BootRun(8080);
end.`)
	// BootRun must call read_body (per Content-Length) before the handler so
	// req.Body() returns the real POST payload, not a constant null.
	assertIRContains(t, ir, "call void @__kylix_boot_read_body(ptr %t")
	assertIRContains(t, ir, "define void @__kylix_boot_read_body(ptr %conn, ptr %headers, ptr %req)")
	// Content-Length parse + recv.
	assertIRContains(t, ir, "call ptr @strstr(ptr %headers, ptr")
	assertIRContains(t, ir, "Content-Length:")
	assertIRContains(t, ir, "call i64 @recv(i32")
	if contains(t, ir, "store ptr null, ptr %req") {
		t.Errorf("req body still hardcoded null\nIR:\n%s", ir)
	}
}

func TestBoot_ReqJSONVariantMap(t *testing.T) {
	ir := generateIR(t, `program p;
uses boot;
[Controller('/api')]
type
  TApiController = class
    [Post('/json')]
    function J(req: TRequest): TResponse;
    var
      data: Variant;
    begin
      data := req.JSON();
      result := BootText(200, data['name']);
    end;
  end;
begin
  WriteLn('ok');
end.`)
	// req.JSON() parses the request body into a Variant map (JsonDecodeMap →
	// box_map) and variant-map lookup reads keys.
	assertIRContains(t, ir, "call ptr @__kylix_json_JsonDecodeMap(ptr")
	assertIRContains(t, ir, "call ptr @__kylix_variant_box_map(ptr")
	assertIRContains(t, ir, "call ptr @__kylix_variant_map_get(ptr")
	// BootText body must unbox the Variant (as_str) — passing the box ptr
	// directly would strlen the box bytes.
	assertIRContains(t, ir, "call ptr @__kylix_variant_as_str(ptr")
	if contains(t, ir, "TRequest.JSON stub") {
		t.Errorf("req.JSON still routed to stub\nIR:\n%s", ir)
	}
}

func TestBoot_RegisterJwtAuthReal(t *testing.T) {
	ir := generateIR(t, `program p;
uses boot;
[Controller('/api')]
type
  TApiController = class
    [Get('/admin')]
    [Authenticated]
    function A(req: TRequest): TResponse;
    begin
      result := BootText(200, 'ok');
    end;
  end;
begin
  BootRegisterJwtAuth('s3cret');
  BootRun(8080);
end.`)
	// BootRegisterJwtAuth stores the secret as a module global.
	assertIRContains(t, ir, "@__kylix_boot_jwt_secret = global ptr")
	// BootEnforceAuth is a real define (not `ret ptr null`), reads the
	// Authorization: Bearer header and verifies with JwtVerify (HS256).
	assertIRContains(t, ir, "define ptr @__kylix_boot_BootEnforceAuth(ptr %req)")
	assertIRContains(t, ir, "Authorization:")
	assertIRContains(t, ir, "Bearer ")
	assertIRContains(t, ir, "call ptr @__kylix_jwt_JwtVerify(ptr")
	// Deny path returns a 401 response handle.
	assertIRContains(t, ir, "call ptr @__kylix_boot_BootText(i64 401")
	// A real pass-branch still exists (ret null), but it must coexist with the
	// auth machinery — proving this is not the old unconditional pass stub.
	assertIRContains(t, ir, "ret ptr null")
}
