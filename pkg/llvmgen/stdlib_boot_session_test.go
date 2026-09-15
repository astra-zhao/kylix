package llvmgen_test

import (
	"strings"
	"testing"
)

// v0.9.0 P1.7c: LLVM-side sessions — middleware halves embedded in BootRun,
// per-session htab on the request handle, flags-driven Set-Cookie.

const bootSessionProgram = `program p;
uses boot;
[Controller('/api')]
type
  TSessionController = class
    [Get('/s')]
    function S(req: TRequest): TResponse;
    var v: String;
    begin
      req.SessionSet('uid', '42');
      v := req.SessionGet('uid');
      req.SessionMarkRemember();
      req.SessionDestroy();
      req.SessionDelete('uid');
      result := BootText(200, v + req.SessionCSRFToken());
    end;
  end;
begin
  BootRun(8107);
end.`

// csrfProgram arms the gate and hits the protected + unprotected paths.
const bootCsrfProgram = `program p;
uses boot;
[Controller('/api')]
type
  TCsrfController = class
    [Get('/form')]
    function Form(req: TRequest): TResponse;
    [Post('/submit')]
    function Submit(req: TRequest): TResponse;
    var v: String;
    begin
      v := req.SessionCSRFToken();
      result := BootText(200, v);
    end;
  end;
begin
  BootUseCSRF();
  BootRun(8107);
end.`

// v0.9.0 P1.7: BootRun-embedded CSRF gate — opt-in global, session token
// check, header/form token sources, 403 messages (pkg/boot/csrf.go parity).
func TestBoot_CsrfIR(t *testing.T) {
	ir := generateIR(t, bootCsrfProgram)
	// The gate define exists and BootRun dispatches through it.
	assertIRContains(t, ir, "define ptr @__kylix_boot_csrf_check(ptr %req, ptr %headers)")
	assertIRContains(t, ir, "call ptr @__kylix_boot_csrf_check(ptr")
	// Opt-in toggle: global declared unconditionally, armed by BootUseCSRF.
	assertIRContains(t, ir, "@__kylix_boot_csrf_enabled = global i1 false")
	assertIRContains(t, ir, "define void @__kylix_boot_BootUseCSRF()")
	assertIRContains(t, ir, "call void @__kylix_boot_BootUseCSRF()")
	// Token sources: X-CSRF-Token header needle and _csrf form field.
	assertIRContains(t, ir, "X-CSRF-Token: ")
	assertIRContains(t, ir, "_csrf")
	// 403s route through BootText with the Go parity messages.
	assertIRContains(t, ir, "CSRF token missing: render req.CSRFToken() into the form first")
	assertIRContains(t, ir, "CSRF token mismatch")
	// Middleware chain shape: CSRF picks the response via phi, then finish.
	assertIRContains(t, ir, "phi ptr")
	if strings.Contains(ir, "unsupported receiver") {
		t.Fatalf("csrf program collapsed to unsupported-receiver stub:\n%s", ir)
	}
}

// The gate define is BootRun-embedded (always emitted, like the session
// middleware); BootUseCSRF only arms the runtime toggle — without it the
// global stays false and the store never appears.
func TestBoot_CsrfOptOutIR(t *testing.T) {
	ir := generateIR(t, bootSessionProgram)
	assertIRContains(t, ir, "@__kylix_boot_csrf_enabled = global i1 false")
	assertIRContains(t, ir, "define ptr @__kylix_boot_csrf_check")
	if strings.Contains(ir, "store i1 true, ptr @__kylix_boot_csrf_enabled") {
		t.Fatalf("CSRF armed without BootUseCSRF")
	}
}

func TestBoot_SessionMiddlewareIR(t *testing.T) {
	ir := generateIR(t, bootSessionProgram)
	// Both middleware halves are emitted and called from BootRun.
	assertIRContains(t, ir, "define void @__kylix_boot_session_resolve(ptr %req, ptr %headers)")
	assertIRContains(t, ir, "define void @__kylix_boot_session_finish(ptr %req, ptr %res)")
	assertIRContains(t, ir, "call void @__kylix_boot_session_resolve(ptr")
	assertIRContains(t, ir, "call void @__kylix_boot_session_finish(ptr")
	// Outer session table + per-session reserved keys.
	assertIRContains(t, ir, "@__kylix_boot_sessions = global ptr null")
	assertIRContains(t, ir, "__sid")
	assertIRContains(t, ir, "__exp")
	// Sliding TTLs (24h / 30d in ms) and the remember cookie Max-Age.
	assertIRContains(t, ir, "86400000")
	assertIRContains(t, ir, "2592000000")
	// Set-Cookie lines match pkg/boot.writeSessionCookie.
	assertIRContains(t, ir, "Set-Cookie: KYLIX_SID=; Path=/; Max-Age=0; HttpOnly")
	assertIRContains(t, ir, "Set-Cookie: KYLIX_SID=")
	assertIRContains(t, ir, "; Path=/; HttpOnly; Max-Age=2592000")
}

func TestBoot_SessionMethodsIR(t *testing.T) {
	ir := generateIR(t, bootSessionProgram)
	// SessionGet/Set route through the per-session htab.
	assertIRContains(t, ir, "call ptr @__kylix_htab_get(ptr")
	assertIRContains(t, ir, "call ptr @__kylix_htab_strdup(ptr")
	assertIRContains(t, ir, "call void @__kylix_htab_put(ptr")
	assertIRContains(t, ir, "call void @__kylix_htab_del(ptr")
	// CSRF tokens are minted as 16 bytes → 32 hex chars.
	assertIRContains(t, ir, "define ptr @__kylix_boot_rand_hex(i64 %nbytes)")
	assertIRContains(t, ir, "call ptr @__kylix_boot_rand_hex(i64 16)")
	assertIRContains(t, ir, "__csrf")
	if strings.Contains(ir, "unsupported receiver") {
		t.Fatalf("session methods collapsed to unsupported-receiver stub:\n%s", ir)
	}
}
