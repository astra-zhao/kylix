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
	// v0.8.0 P2: the handle comes from the per-request arena.
	assertIRContains(t, ir, "@__kylix_arena_alloc(i64 40)")
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

// v0.8.0 P3: repeated WithCookie calls append full "Set-Cookie: ...\r\n"
// lines to a growing buffer (mirroring the Go side's []string) instead of
// overwriting a single cookie slot; the response assembler appends the
// buffer verbatim.
func TestBoot_MultiCookieAppend(t *testing.T) {
	ir := generateIR(t, `program p;
uses boot;
[Controller('/api')]
type
  TCookieController = class
    [Get('/c')]
    function C(req: TRequest): TResponse;
    begin
      result := BootText(200, 'ok').WithCookie('a', '1').WithCookie('b', '2');
    end;
  end;
begin
  BootRun(8094);
end.`)
	assertIRContains(t, ir, "Set-Cookie: %s=%s; Path=/\\0D\\0A")
	// appends copy old+entry into a fresh arena block, not a slot overwrite
	assertIRContains(t, ir, "call ptr @memcpy(ptr")
}

// v0.8.0 P3: WithHeader/Redirect grow the xhdrs buffer by realloc — the
// fixed 1024-byte first-use allocation is gone.
func TestBoot_XhdrsReallocGrowth(t *testing.T) {
	ir := generateIR(t, bootFluentProgram)
	if strings.Contains(ir, "malloc(i64 1024)") {
		t.Errorf("xhdrs still uses a fixed 1024-byte buffer\nIR:\n%s", ir)
	}
	// v0.8.0 P2: the growable buffer is arena-resident (copy-append, no realloc)
	assertIRContains(t, ir, "call ptr @memcpy(ptr")
}

// v0.9.0 P1.7: FileBytes responds with in-memory bytes as an attachment —
// status 200, MIME by filename extension, Content-Disposition header appended
// to the xhdrs slot.
func TestBoot_FileBytesIR(t *testing.T) {
	ir := generateIR(t, `program p;
uses boot;
[Controller('/api')]
type
  TDlController = class
    [Get('/csv')]
    function Csv(req: TRequest): TResponse;
    begin
      result := BootText(200, '').FileBytes('id,name', 'report.csv');
    end;
  end;
begin
  BootRun(8095);
end.`)
	assertIRContains(t, ir, `Content-Disposition: attachment; filename=\22%s\22`)
	// .csv resolves in the MIME table (v0.9.0 P1.7 — matches Go mimeFor)
	assertIRContains(t, ir, ".csv")
	assertIRContains(t, ir, "text/csv; charset=utf-8")
}

// v0.9.0 P1.7: Download reads the file via fopen/fseek/ftell/fread and
// degrades to a 404 "file not found" text response when fopen fails.
func TestBoot_DownloadIR(t *testing.T) {
	ir := generateIR(t, `program p;
uses boot;
[Controller('/api')]
type
  TDlController = class
    [Get('/dl')]
    function Dl(req: TRequest): TResponse;
    begin
      result := BootText(200, '').Download('/srv/report.pdf', 'report.pdf');
    end;
  end;
begin
  BootRun(8096);
end.`)
	assertIRContains(t, ir, "call ptr @fopen(ptr")
	assertIRContains(t, ir, "call i64 @fread(ptr")
	assertIRContains(t, ir, "file not found")
	assertIRContains(t, ir, "call i32 (ptr, i64, ptr, ...) @snprintf")
}

// v0.9.0 P1.7: BootPagerHTML — module-level pagination renderer. The nav-bar
// assembler plus its put/puti/url/htmlattr/link helpers must all be present,
// and the arena buffer must be zero-initialized (strcat cursor semantics on
// recycled per-request arena memory).
func TestBoot_PagerHTMLIR(t *testing.T) {
	ir := generateIR(t, `program p;
uses boot;
[Controller('/api')]
type
  TPgController = class
    [Get('/list')]
    function List(req: TRequest): TResponse;
    var nav: String;
    begin
      nav := BootPagerHTML('/api/list', 2, 10, 95, 2);
      result := BootHTML(200, nav);
    end;
  end;
begin
  BootRun(8094);
end.`)
	assertIRContains(t, ir, "define ptr @__kylix_boot_pager_html(ptr %base, i64 %page, i64 %size, i64 %total, i64 %window)")
	assertIRContains(t, ir, "define ptr @__kylix_boot_pager_put(ptr %cur, ptr %s)")
	assertIRContains(t, ir, "define ptr @__kylix_boot_pager_puti(ptr %cur, i64 %n)")
	assertIRContains(t, ir, "define ptr @__kylix_boot_pager_url(ptr %base, i64 %n)")
	assertIRContains(t, ir, "define ptr @__kylix_boot_htmlattr(ptr %s)")
	assertIRContains(t, ir, "define ptr @__kylix_boot_pager_link(ptr %cur, ptr %href, i64 %n)")
	// pager window fragments (quotes escape as \22, UTF-8 ‹ › … as \E2\80\B9
	// etc. — matches pkg/boot/pagination.go exactly)
	assertIRContains(t, ir, `<span class=\22pager-current\22>`)
	assertIRContains(t, ir, `<span class=\22pager-ellipsis\22>`)
	assertIRContains(t, ir, `pager-disabled`)
	// attribute escaping for double-quoted hrefs
	assertIRContains(t, ir, "&amp;")
	assertIRContains(t, ir, "&quot;")
	// arena buffer must be zeroed before the first strcat append
	assertIRContains(t, ir, "call ptr @__kylix_arena_alloc(i64 %t")
	assertIRContains(t, ir, "define ptr @__kylix_boot_pager_puti(ptr %cur, i64 %n)")
}

// v0.9.0 P1.7: req.PageNum / req.PageSize — scalar pagination parse over the
// raw query string with clamp-to-default / clamp-to-max select chains.
func TestBoot_PageNumPageSizeIR(t *testing.T) {
	ir := generateIR(t, `program p;
uses boot;
[Controller('/api')]
type
  TPgController = class
    [Get('/list')]
    function List(req: TRequest): TResponse;
    var pg, sz: Integer;
    begin
      pg := req.PageNum(1);
      sz := req.PageSize(20, 100);
      result := BootText(200, 'ok');
    end;
  end;
begin
  BootRun(8099);
end.`)
	// both parses reuse the extracted emitBootQueryGet core
	assertIRContains(t, ir, "call ptr @strchr(ptr %t")
	// clamping: PageNum selects def when < 1; PageSize clamps to max too
	assertIRContains(t, ir, "icmp slt i64")
	assertIRContains(t, ir, "icmp sgt i64")
	assertIRContains(t, ir, "select i1")
	if strings.Contains(ir, "unsupported receiver") {
		t.Fatalf("PageNum/PageSize collapsed to unsupported-receiver stub:\n%s", ir)
	}
}
