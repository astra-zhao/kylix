// stdlib_boot_auth_test.go — v0.10.0 P2: session-first auth guards,
// SessionRegenerate lowering, and the Pbkdf2 password-hash IR.
package llvmgen_test

import (
	"strings"
	"testing"
)

const authGuardProgram = `program p;
uses boot;
[Controller('/api')]
type
  TApiController = class
    [Authenticated]
    [Role('admin')]
    [Get('/secret')]
    function Secret(req: TRequest): TResponse;
    begin
      result := BootText(200, 'classified');
    end;
  end;
begin
  BootRun(8080);
end.`

func TestBoot_EnforceAuth_SessionFirst(t *testing.T) {
	ir := generateIR(t, authGuardProgram)
	body := defineBody(ir, "define ptr @__kylix_boot_BootEnforceAuth(")
	if body == "" {
		t.Fatal("BootEnforceAuth define missing")
	}
	// String literals live in the module constant pool; inside the define,
	// assert order structurally: the __user htab lookup precedes the
	// Authorization header search (session-first order, Go
	// pkg/boot/security.go parity).
	if !strings.Contains(ir, `c"__user\00"`) {
		t.Fatal("__user session key constant missing")
	}
	getIdx := strings.Index(body, "call ptr @__kylix_htab_get")
	if getIdx < 0 {
		t.Fatal("session lookup missing from BootEnforceAuth body")
	}
	// The "Authorization:" literal is also a constant-pool string; the Bearer
	// path is located by its strstr call (the first one in the body).
	authIdx := strings.Index(body, "call ptr @strstr(")
	if authIdx < 0 {
		t.Fatal("Bearer path missing from BootEnforceAuth body")
	}
	if getIdx > authIdx {
		t.Fatalf("session check must precede the Bearer path (get@%d > auth@%d)", getIdx, authIdx)
	}
}

func TestBoot_EnforceRole_RealBody(t *testing.T) {
	ir := generateIR(t, authGuardProgram)
	body := defineBody(ir, "define ptr @__kylix_boot_BootEnforceRole(")
	if body == "" {
		t.Fatal("BootEnforceRole define missing")
	}
	// v0.10.0 P2: the always-pass stub is gone; the body reads __roles and
	// can deny with a 403.
	if !strings.Contains(ir, `c"__roles\00"`) {
		t.Fatal("__roles session key constant missing")
	}
	if !strings.Contains(ir, `c"Forbidden\00"`) {
		t.Fatal("403 Forbidden message constant missing")
	}
	if !strings.Contains(body, "call ptr @__kylix_htab_get") {
		t.Fatal("session lookup missing from BootEnforceRole body")
	}
	stub := "define ptr @__kylix_boot_BootEnforceRole(ptr %req, ptr %role) {\n  ret ptr null\n}"
	if strings.Contains(ir, stub) {
		t.Fatal("BootEnforceRole still lowered to the always-pass stub")
	}
}

func TestBoot_SessionRegenerateLowered(t *testing.T) {
	ir := generateIR(t, `program p;
uses boot;
[Controller('/auth')]
type
  TAuthController = class
    [Post('/login')]
    function Login(req: TRequest): TResponse;
    begin
      req.SessionRegenerate();
      req.SessionSet('user', 'admin');
      result := BootText(200, 'ok');
    end;
  end;
begin
  BootRun(8080);
end.`)
	// Inline regenerate sequence: mint a fresh SID, move the outer-table
	// entry, re-persist __sid, flag the cookie dirty.
	assertIRContains(t, ir, "call ptr @__kylix_boot_rand_hex(i64 32)")
	assertIRContains(t, ir, "call void @__kylix_htab_del(ptr")
	assertIRContains(t, ir, "call ptr @__kylix_htab_strdup(")
}

const pbkdf2Program = `program p;
uses crypto;
begin
  var h := Pbkdf2Hash('hunter2', 210000);
  if Pbkdf2Compare('hunter2', h) then
    WriteLn('ok');
end.`

func TestCrypto_Pbkdf2HashIR(t *testing.T) {
	ir := generateIR(t, pbkdf2Program)
	assertIRContains(t, ir, "define ptr @__kylix_crypto_Pbkdf2Hash(ptr %password, i64 %iter)")
	assertIRContains(t, ir, "call i32 @RAND_bytes(ptr")
	assertIRContains(t, ir, "call i32 @PKCS5_PBKDF2_HMAC(")
	assertIRContains(t, ir, `pbkdf2$sha256$%lld$%s$%s`)
	assertIRContains(t, ir, "call ptr @__kylix_crypto_Pbkdf2Hash(ptr")
	// Iteration clamp bounds.
	assertIRContains(t, ir, "icmp slt i64 %iter, 1000")
	assertIRContains(t, ir, "icmp sgt i64")
}

func TestCrypto_Pbkdf2CompareIR(t *testing.T) {
	ir := generateIR(t, pbkdf2Program)
	assertIRContains(t, ir, "define i1 @__kylix_crypto_Pbkdf2Compare(ptr %password, ptr %hash)")
	assertIRContains(t, ir, `pbkdf2$sha256$%lld$%32[^$]$%64[^$]`)
	assertIRContains(t, ir, "call i32 (ptr, ptr, ...) @sscanf(")
	assertIRContains(t, ir, "call i32 @PKCS5_PBKDF2_HMAC(")
	// Fail closed on out-of-range iterations.
	assertIRContains(t, ir, "icmp sge i64")
	assertIRContains(t, ir, "icmp sle i64")
	// Constant-time XOR-accumulate compare (no early exit).
	assertIRContains(t, ir, "xor i8")
	assertIRContains(t, ir, "icmp eq i8")
}
