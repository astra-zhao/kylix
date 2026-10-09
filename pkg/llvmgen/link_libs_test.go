package llvmgen

import (
	"strings"
	"testing"
)

func TestPlanStdlibLibs_DesktopCrypto(t *testing.T) {
	ir := "define ptr @__kylix_crypto_Sha256(ptr %a) {\n  call ptr @SHA256(ptr %a, i64 %n, ptr %o)\n}\n"
	got, err := planStdlibLibs("linux", ir)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(got.Libs, " ") != "-lcrypto" {
		t.Fatalf("linux libs = %v", got.Libs)
	}
	if got.PortableHash || got.SqliteAmalgamation {
		t.Fatalf("linux plan = %+v", got)
	}
}

func TestPlanStdlibLibs_DarwinDeclaresSqlite(t *testing.T) {
	// Darwin links -lsqlite3 whenever the always-emitted sqlite declare is
	// present. Linux does not.
	ir := "declare i32 @sqlite3_open(ptr noundef, ptr noundef)\n"
	darwin, err := planStdlibLibs("darwin", ir)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(darwin.Libs, " ") != "-lsqlite3" {
		t.Fatalf("darwin libs = %v", darwin.Libs)
	}
	linux, err := planStdlibLibs("linux", ir)
	if err != nil {
		t.Fatal(err)
	}
	if len(linux.Libs) != 0 {
		t.Fatalf("linux libs = %v, want none", linux.Libs)
	}
}

func TestPlanStdlibLibs_MobileDoesNotLinkDesktopLibs(t *testing.T) {
	hello := "define i32 @main() { ret i32 0 }\ndeclare ptr @SHA256(ptr, i64, ptr)\ndeclare i32 @sqlite3_open(ptr, ptr)\n"
	for _, osName := range []string{"android", "ios"} {
		got, err := planStdlibLibs(osName, hello)
		if err != nil {
			t.Fatal(osName, err)
		}
		if len(got.Libs) != 0 || got.PortableHash || got.SqliteAmalgamation {
			t.Fatalf("%s hello plan = %+v", osName, got)
		}
	}
}

func TestPlanStdlibLibs_MobileHashAndSQLite(t *testing.T) {
	ir := "declare void @kylix_sha256(ptr, i64, ptr)\n  call void @kylix_sha256(ptr %a, i64 %n, ptr %o)\n  call ptr @__kylix_db_DbOpenSQLite(ptr %p)\n"
	and, err := planStdlibLibs("android", ir)
	if err != nil {
		t.Fatal(err)
	}
	if !and.PortableHash || !and.SqliteAmalgamation || len(and.Libs) != 0 {
		t.Fatalf("android plan = %+v", and)
	}
	ios, err := planStdlibLibs("ios", ir)
	if err != nil {
		t.Fatal(err)
	}
	if !ios.PortableHash || ios.SqliteAmalgamation || strings.Join(ios.Libs, " ") != "-lsqlite3" {
		t.Fatalf("ios plan = %+v", ios)
	}
}

func TestPlanStdlibLibs_MobileRejectsDesktopStacks(t *testing.T) {
	cases := []struct {
		ir   string
		want string
	}{
		{"call ptr @__kylix_httpclient_Get(ptr %s, ptr %u)\n", "httpclient"},
		{"define ptr @__kylix_db_pg_open(ptr %dsn)\n", "postgres"},
		{"call ptr @EVP_CIPHER_CTX_new()\n", "OpenSSL"},
		{"call ptr @SHA256(ptr %a, i64 %n, ptr %o)\n", "OpenSSL"},
	}
	for _, osName := range []string{"android", "ios"} {
		for _, tc := range cases {
			_, err := planStdlibLibs(osName, tc.ir)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("%s ir %q: got %v, want substring %q", osName, tc.ir, err, tc.want)
			}
		}
	}
}

func TestTargetCompileArgs(t *testing.T) {
	in := []string{
		"-o", "app", "app.o",
		"--target=aarch64-linux-android30",
		"--sysroot=/opt/ndk/sysroot",
		"-arch", "arm64",
		"-isysroot", "/Sdk",
		"-shared",
	}
	got := strings.Join(targetCompileArgs(in), " ")
	want := "--target=aarch64-linux-android30 --sysroot=/opt/ndk/sysroot -arch arm64 -isysroot /Sdk"
	if got != want {
		t.Fatalf("got %q\nwant %q", got, want)
	}
}
