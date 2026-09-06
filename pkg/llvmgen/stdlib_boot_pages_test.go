package llvmgen_test

import (
	"strings"
	"testing"
)

// v0.7.0 P2: page-rendering API IR — response fluent methods, static files,
// req.Form / req.Cookie.

const bootPagesProgram = `program p;
uses boot;
[Controller('/api')]
type
  TPageController = class
    [Get('/page')]
    function Page(req: TRequest): TResponse;
    begin
      result := BootHTML(200, '<h1>hi</h1>').Html('<b>x</b>');
    end;
  end;
begin
  BootStatic('./static');
  BootRun(8091);
end.`

const bootFormProgram = `program p;
uses boot;
[Controller('/api')]
type
  TFormController = class
    [Get('/f')]
    function F(req: TRequest): TResponse;
    begin
      result := BootText(200, req.Form('name') + req.Cookie('sid'));
    end;
  end;
begin
  BootRun(8092);
end.`

const bootFluentProgram = `program p;
uses boot;
[Controller('/api')]
type
  TFluentController = class
    [Get('/f')]
    function F(req: TRequest): TResponse;
    begin
      result := BootText(200, 'plain').Html('<b>x</b>').WithHeader('X-K', 'v').WithCookie('sid', 'abc').StatusCode(201);
    end;
  end;
begin
  BootRun(8093);
end.`

func TestBoot_ResponseHtmlMethod(t *testing.T) {
	ir := generateIR(t, bootPagesProgram)
	// 40-byte handle with content-type slot; Html stores body + text/html.
	assertIRContains(t, ir, "@malloc(i64 40)")
	assertIRContains(t, ir, "text/html; charset=utf-8")
	assertIRContains(t, ir, "Content-Type")
}

// v0.7.0 P2: a fluent chain whose receiver is itself a call (BootText(...) —
// callReturnKylixType must report TResponse so each chain link dispatches
// through emitBootResponseMethodCall instead of the unsupported-receiver stub,
// which returned null and crashed BootRun on the handle load).
func TestBoot_FluentChainNoStub(t *testing.T) {
	ir := generateIR(t, bootFluentProgram)
	if strings.Contains(ir, "unsupported receiver") {
		t.Fatalf("fluent chain collapsed to unsupported-receiver stub:\n%s", ir)
	}
	assertIRContains(t, ir, "@__kylix_boot_BootText(i64")
	assertIRContains(t, ir, "text/html; charset=utf-8")
	// cookie/header are runtime-formatted; the IR carries the format strings.
	assertIRContains(t, ir, "%s=%s; Path=/")
	assertIRContains(t, ir, "X-K")
}

func TestBoot_StaticServingIR(t *testing.T) {
	ir := generateIR(t, bootPagesProgram)
	assertIRContains(t, ir, "@__kylix_boot_static_dir = global ptr null")
	assertIRContains(t, ir, "define void @__kylix_boot_BootStatic(ptr %dir)")
	assertIRContains(t, ir, "define i1 @__kylix_boot_serve_static(ptr %conn, ptr %method, ptr %path)")
	assertIRContains(t, ir, "@__kylix_boot_serve_static(ptr")
	// MIME table essentials.
	assertIRContains(t, ir, ".css")
	assertIRContains(t, ir, "image/png")
	assertIRContains(t, ir, "application/octet-stream")
}

func TestBoot_FormAndCookieIR(t *testing.T) {
	ir := generateIR(t, bootFormProgram)
	assertIRContains(t, ir, "define ptr @__kylix_boot_form_get(ptr %body, ptr %name)")
	assertIRContains(t, ir, "define ptr @__kylix_boot_cookie_get(ptr %headers, ptr %name)")
	assertIRContains(t, ir, "define i64 @__kylix_boot_hexval(i8 %c)")
	assertIRContains(t, ir, "@__kylix_boot_form_get(ptr")
	assertIRContains(t, ir, "@__kylix_boot_cookie_get(ptr")
}
