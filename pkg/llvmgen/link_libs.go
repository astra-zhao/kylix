package llvmgen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// stdlibLinkPlan is the system-library decision for one IR module.
// Desktop targets keep the historical -lcrypto/-lpq/-lsqlite3/-lcurl scan.
// android and ios do not link those desktop libraries.
type stdlibLinkPlan struct {
	Libs               []string
	PortableHash       bool
	SqliteAmalgamation bool
}

// planStdlibLibs decides which libraries a finished IR module needs.
// Homebrew -L/-rpath for darwin is applied by the caller; this function
// only returns the -l names, in the same order the desktop link used.
func planStdlibLibs(targetOS, ir string) (stdlibLinkPlan, error) {
	var plan stdlibLinkPlan
	switch targetOS {
	case "windows", "wasi":
		return plan, nil
	case "android", "ios":
		return planMobileLibs(targetOS, ir)
	default:
		plan.Libs = desktopStdlibLibs(targetOS, ir)
		return plan, nil
	}
}

// desktopStdlibLibs is the pre-v0.15 scan for linux and darwin. ios is not
// handled here (it used to fall into this scan and skip only -lcurl).
func desktopStdlibLibs(targetOS, ir string) []string {
	var libs []string
	if strings.Contains(ir, "@__kylix_crypto_") || (targetOS == "darwin" && strings.Contains(ir, "@SHA1")) {
		libs = append(libs, "-lcrypto")
	}
	if strings.Contains(ir, "@__kylix_db_pg_") {
		libs = append(libs, "-lpq")
	}
	// @__kylix_db_pg_ also contains the substring @__kylix_db_, so a postgres
	// program links sqlite as well. That is the existing desktop behavior.
	if strings.Contains(ir, "@__kylix_db_") || (targetOS == "darwin" && strings.Contains(ir, "@sqlite3_")) {
		libs = append(libs, "-lsqlite3")
	}
	if targetOS != "ios" && (strings.Contains(ir, "@__kylix_httpclient_") || (targetOS == "darwin" && strings.Contains(ir, "@curl_easy_"))) {
		libs = append(libs, "-lcurl")
	}
	return libs
}

func planMobileLibs(targetOS, ir string) (stdlibLinkPlan, error) {
	var plan stdlibLinkPlan
	if strings.Contains(ir, "@__kylix_httpclient_") {
		return plan, fmt.Errorf("httpclient is not available on %s: libcurl is not linked (HTTP stays in the native shell)", targetOS)
	}
	if strings.Contains(ir, "@__kylix_db_pg_") {
		return plan, fmt.Errorf("postgres (libpq) is not available on %s", targetOS)
	}
	if irCallsOpenSSL(ir) {
		return plan, fmt.Errorf("OpenSSL is not linked on %s (AES, BCrypt, and PBKDF2 need it; Sha256, Md5, and HmacSha256 use the portable hash)", targetOS)
	}
	if strings.Contains(ir, "call void @kylix_sha256") || strings.Contains(ir, "call void @kylix_md5") {
		plan.PortableHash = true
	}
	if strings.Contains(ir, "@__kylix_db_") {
		if targetOS == "ios" {
			// System libsqlite3.tbd. No Homebrew path.
			plan.Libs = append(plan.Libs, "-lsqlite3")
		} else {
			plan.SqliteAmalgamation = true
		}
	}
	return plan, nil
}

// irCallsOpenSSL reports a real OpenSSL call, not the always-emitted declare.
func irCallsOpenSSL(ir string) bool {
	needles := []string{
		"call ptr @SHA256",
		"call ptr @MD5",
		"call ptr @EVP_",
		"call i32 @EVP_",
		"call void @EVP_",
		"call i32 @RAND_bytes",
		"call i32 @PKCS5_PBKDF2_HMAC",
	}
	for _, n := range needles {
		if strings.Contains(ir, n) {
			return true
		}
	}
	return false
}

// compileSqliteAmalgamation compiles the bundled sqlite3.c for the current
// link target. Android does not link the host -lsqlite3.
func compileSqliteAmalgamation(clang string, linkArgs []string) (obj string, dir string, err error) {
	src, err := findSqliteAmalgamation()
	if err != nil {
		return "", "", err
	}
	dir, err = os.MkdirTemp("", "kylix-sqlite-")
	if err != nil {
		return "", "", err
	}
	obj = filepath.Join(dir, "sqlite3.o")
	defines := []string{
		"-DSQLITE_OMIT_LOAD_EXTENSION=1",
		"-DSQLITE_THREADSAFE=1",
	}
	if err := compileSupportC(clang, linkArgs, src, obj, defines); err != nil {
		os.RemoveAll(dir)
		return "", "", err
	}
	return obj, dir, nil
}

// findSqliteAmalgamation locates sqlite3.c. The amalgamation is not committed
// (it is about 9MB). scripts/fetch_sqlite_amalgamation.sh writes it to
// third_party/sqlite/sqlite3.c. KYLIX_SQLITE_SRC overrides that path.
func findSqliteAmalgamation() (string, error) {
	if p := os.Getenv("KYLIX_SQLITE_SRC"); p != "" {
		st, err := os.Stat(p)
		if err != nil || st.IsDir() {
			return "", fmt.Errorf("KYLIX_SQLITE_SRC=%s is not a sqlite3.c file", p)
		}
		return p, nil
	}
	candidates := []string{
		filepath.Join("third_party", "sqlite", "sqlite3.c"),
	}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates,
			filepath.Join(filepath.Dir(exe), "..", "third_party", "sqlite", "sqlite3.c"),
			filepath.Join(filepath.Dir(exe), "third_party", "sqlite", "sqlite3.c"),
		)
	}
	for _, c := range candidates {
		st, err := os.Stat(c)
		if err == nil && !st.IsDir() {
			return c, nil
		}
	}
	return "", fmt.Errorf("android sqlite needs the bundled amalgamation sqlite3.c (this target does not link -lsqlite3): run scripts/fetch_sqlite_amalgamation.sh or set KYLIX_SQLITE_SRC")
}
