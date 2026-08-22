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
