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
