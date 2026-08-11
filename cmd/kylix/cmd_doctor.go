package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"kylix/pkg/llvmgen"
)

// cmdDoctor diagnoses the Kylix toolchain: Go backend, LLVM backend
// (llc/clang/opt), and the native stdlib system libraries (sqlite3/curl/
// openssl). v6.2.0 — the KylixRT "preflight" for distribution (LLVM backend
// needs llc + clang; the Go backend needs go).
func cmdDoctor(args []string) {
	fs := flag.NewFlagSet("doctor", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Printf(`USAGE: kylix doctor

Diagnose the Kylix toolchain:
  - Go backend:   needs go on PATH (kylix run / build --backend=go)
  - LLVM backend: needs llc + clang (+ optional opt for --llvm-opt)
                  macOS: brew install llvm    Linux: apt install llvm clang
  - stdlib native libs: sqlite3 / curl / openssl (db / httpclient / crypto)

Exit code is non-zero if any required tool is missing.
`)
	}
	_ = fs.Parse(args)

	fail := false

	fmt.Println("=== Go toolchain ===")
	if p, err := exec.LookPath("go"); err == nil {
		fmt.Printf("  ✓ go: %s\n", p)
	} else {
		fmt.Println("  ✗ go: not found — Go backend unavailable (kylix run will fall back to LLVM)")
		fail = true
	}

	fmt.Println("\n=== LLVM toolchain ===")
	llvm, err := llvmgen.FindLLVM()
	if err != nil {
		fmt.Printf("  ✗ %v\n", err)
		fmt.Println("    Hint: brew install llvm (macOS) or apt install llvm clang (Linux)")
		fail = true
	} else {
		if llvm.LLC != "" {
			fmt.Printf("  ✓ llc: %s\n", llvm.LLC)
		} else {
			fmt.Println("  ✗ llc: not found")
			fail = true
		}
		if llvm.Clang != "" {
			fmt.Printf("  ✓ clang: %s\n", llvm.Clang)
		} else {
			fmt.Println("  ✗ clang: not found")
			fail = true
		}
		if llvm.Opt != "" {
			fmt.Printf("  ✓ opt: %s (optional, for --llvm-opt)\n", llvm.Opt)
		} else {
			fmt.Println("  − opt: not found (optional; --llvm-opt falls back to llc -O)")
		}
		if out, err := exec.Command(llvm.LLC, "--version").Output(); err == nil {
			fmt.Printf("  version: %s\n", strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0])
		}
	}

	fmt.Println("\n=== stdlib native system libraries ===")
	clang := "clang"
	if llvm != nil && llvm.Clang != "" {
		clang = llvm.Clang
	}
	checkLib(clang, "sqlite3", "-lsqlite3", "sqlite")
	checkLib(clang, "curl", "-lcurl", "curl")
	checkLib(clang, "openssl (libcrypto)", "-lcrypto", "openssl")

	fmt.Println()
	if fail {
		fmt.Println("✗ doctor: missing required tooling (see above).")
		os.Exit(1)
	}
	fmt.Println("✓ doctor: all required tools found.")
}

// checkLib probes whether a system library links, by compiling a trivial C
// program with the platform C compiler. Homebrew brewName (openssl/sqlite/curl)
// is retried with the Homebrew -L paths, which are not on clang's default
// search path on macOS.
func checkLib(cc, name, linkFlag, brewName string) {
	probe := func(extra ...string) bool {
		args := append([]string{"-x", "c", "-", "-o", "/tmp/kylix_libprobe"}, extra...)
		args = append(args, linkFlag)
		cmd := exec.Command(cc, args...)
		cmd.Stdin = strings.NewReader("int main(void){return 0;}\n")
		if err := cmd.Run(); err == nil {
			os.Remove("/tmp/kylix_libprobe")
			return true
		}
		return false
	}
	if probe() {
		fmt.Printf("  ✓ %s (linkable)\n", name)
		return
	}
	for _, dir := range []string{
		"/opt/homebrew/opt/" + brewName + "/lib",
		"/usr/local/opt/" + brewName + "/lib",
	} {
		if probe("-L", dir) {
			fmt.Printf("  ✓ %s (linkable, %s)\n", name, dir)
			return
		}
	}
	fmt.Printf("  ✗ %s: not found / not linkable (needed by db / httpclient / crypto)\n", name)
}
