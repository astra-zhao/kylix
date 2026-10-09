package llvmgen_test

import (
	"strings"
	"testing"

	"kylix/lexer"
	"kylix/parser"
	"kylix/pkg/llvmgen"
)

const mobileTimeSrc = `program p;
uses datetime;
begin
  var now := Now();
  WriteLn(IntToStr(now.Year()));
end.`

const mobileTempSrc = `program p;
uses sysutil;
begin
  WriteLn(GetTempDir());
end.`

const mobileHashSrc = `program p;
uses crypto;
begin
  WriteLn(Sha256('abc'));
  WriteLn(Md5('abc'));
  WriteLn(HmacSha256('key', 'data'));
end.`

const mobileListenSrc = `program p;
uses net;
begin
  var l := TcpListen(9090);
end.`

func TestMobile_DatetimeUsesLocaltimeR(t *testing.T) {
	for _, target := range []string{"android/arm64", "android/amd64", "ios/arm64", "ios/simulator-arm64"} {
		ir := generateIRForTarget(t, mobileTimeSrc, target)
		if !strings.Contains(ir, "call ptr @localtime_r") {
			t.Errorf("%s: datetime should call localtime_r", target)
		}
		if strings.Contains(ir, "call i32 @localtime_s") {
			t.Errorf("%s: localtime_s is the Windows call", target)
		}
	}
}

func TestMobile_GetTempDir(t *testing.T) {
	and := generateIRForTarget(t, mobileTempSrc, "android/arm64")
	if !strings.Contains(and, "/data/local/tmp") {
		t.Fatal("android GetTempDir should fall back to /data/local/tmp")
	}
	if strings.Contains(and, "c\"/tmp\\00\"") {
		t.Fatal("android GetTempDir must not fall back to /tmp")
	}
	ios := generateIRForTarget(t, mobileTempSrc, "ios/arm64")
	if !strings.Contains(ios, "call i64 @confstr(i32 65537,") {
		t.Fatal("ios GetTempDir should call confstr(_CS_DARWIN_USER_TEMP_DIR)")
	}
	if strings.Contains(ios, "/tmp") {
		t.Fatal("ios GetTempDir must not use /tmp")
	}
	host := generateIR(t, mobileTempSrc)
	if !strings.Contains(host, "/tmp") {
		t.Fatal("host GetTempDir should keep the /tmp fallback")
	}
	if strings.Contains(host, "/data/local/tmp") || strings.Contains(host, "@confstr") {
		t.Fatal("host GetTempDir IR changed")
	}
}

func TestMobile_HashIsPortable(t *testing.T) {
	for _, target := range []string{"android/arm64", "ios/simulator-arm64"} {
		ir := generateIRForTarget(t, mobileHashSrc, target)
		if strings.Count(ir, "call void @kylix_sha256") < 3 {
			t.Errorf("%s: Sha256 + HMAC inner/outer should call kylix_sha256, IR:\n%s", target, ir)
		}
		if !strings.Contains(ir, "call void @kylix_md5") {
			t.Errorf("%s: Md5 should call kylix_md5", target)
		}
		if strings.Contains(ir, "call ptr @SHA256") || strings.Contains(ir, "call ptr @MD5") {
			t.Errorf("%s: mobile hash still calls OpenSSL", target)
		}
	}
	host := generateIR(t, mobileHashSrc)
	if !strings.Contains(host, "call ptr @SHA256") || !strings.Contains(host, "call ptr @MD5") {
		t.Fatal("host hash should keep OpenSSL calls")
	}
	if strings.Contains(host, "@kylix_sha256") || strings.Contains(host, "@kylix_md5") {
		t.Fatal("host IR must not mention the portable hash")
	}
}

func TestMobile_OpenSSLCryptoRejected(t *testing.T) {
	src := `program p;
uses crypto;
begin
  var enc := AesEncrypt('mykey', 'hello');
end.`
	for _, target := range []string{"android/arm64", "ios/arm64"} {
		l := generateIRForTargetMaybe(t, src, target)
		if l.err == nil || !strings.Contains(l.err.Error(), "OpenSSL") {
			t.Fatalf("%s: AesEncrypt err = %v", target, l.err)
		}
	}
	if ir := generateIR(t, src); !strings.Contains(ir, "call ptr @EVP_CIPHER_CTX_new") {
		t.Fatal("host AesEncrypt should still emit OpenSSL")
	}
}

func TestMobile_SocketReuseConsts(t *testing.T) {
	and := generateIRForTarget(t, mobileListenSrc, "android/amd64")
	if !strings.Contains(and, "call i32 @setsockopt(i32 ") || !strings.Contains(and, ", i32 1, i32 2, ptr %one, i32 4)") {
		t.Fatalf("android reuse should use Linux SOL_SOCKET/SO_REUSEADDR:\n%s", and)
	}
	ios := generateIRForTarget(t, mobileListenSrc, "ios/arm64")
	if !strings.Contains(ios, ", i32 65535, i32 4, ptr %one, i32 4)") {
		t.Fatalf("ios reuse should use Darwin constants:\n%s", ios)
	}
	lin := generateIRForTarget(t, mobileListenSrc, "linux/amd64")
	if !strings.Contains(lin, ", i32 1, i32 2, ptr %one, i32 4)") {
		t.Fatal("linux reuse constants changed")
	}
	dar := generateIRForTarget(t, mobileListenSrc, "darwin/arm64")
	if !strings.Contains(dar, ", i32 65535, i32 4, ptr %one, i32 4)") {
		t.Fatal("darwin reuse constants changed")
	}
}

func TestMobile_WsRand(t *testing.T) {
	src := `program p;
uses websocket;
begin
  var c := WsDial('127.0.0.1:8080', '/');
end.`
	and := generateIRForTarget(t, src, "android/arm64")
	if !strings.Contains(and, "call i64 @getrandom") {
		t.Fatal("android websocket rand should use getrandom")
	}
	ios := generateIRForTarget(t, src, "ios/arm64")
	if !strings.Contains(ios, "call void @arc4random_buf") {
		t.Fatal("ios websocket rand should use arc4random_buf")
	}
	lin := generateIRForTarget(t, src, "linux/amd64")
	if !strings.Contains(lin, "call i64 @getrandom") || strings.Contains(lin, "call void @arc4random_buf") {
		t.Fatal("linux websocket rand changed")
	}
}

type irResult struct {
	ir  string
	err error
}

func generateIRForTargetMaybe(t *testing.T, src, target string) irResult {
	t.Helper()
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	ir, err := llvmgen.GenerateWithOpts(prog, "", llvmgen.CompileOpts{Target: target})
	return irResult{ir: ir, err: err}
}
