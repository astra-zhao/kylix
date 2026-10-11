package mobile

import (
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"

	"kylix/pkg/llvmgen"
)

// Toolchain is the slice of the machine `kylix mobile doctor` inspects.
// Tests fill it in; LiveToolchain reads the real host.
type Toolchain struct {
	GOOS     string
	LookPath func(string) (string, error)
	NDK      llvmgen.NdkProbe
	LLVMErr  error
	LLC      string
	Clang    string
	Run      func(name string, args ...string) (string, error)
}

// LiveToolchain probes the current machine.
func LiveToolchain() Toolchain {
	t := Toolchain{
		GOOS:     runtime.GOOS,
		LookPath: exec.LookPath,
		NDK:      llvmgen.ProbeAndroidNdk(),
		Run: func(name string, args ...string) (string, error) {
			out, err := exec.Command(name, args...).Output()
			return string(out), err
		},
	}
	llvm, err := llvmgen.FindLLVM()
	if err != nil {
		t.LLVMErr = err
		return t
	}
	t.LLC = llvm.LLC
	t.Clang = llvm.Clang
	return t
}

// Doctor writes the four sections and returns the process exit code.
// libcurl and libgc never fail this command: the flashcard example does not link them.
// A pinned but unusable Android NDK does fail it.
func Doctor(t Toolchain, w io.Writer) int {
	if t.GOOS == "" {
		t.GOOS = runtime.GOOS
	}
	if t.LookPath == nil {
		t.LookPath = exec.LookPath
	}
	fail := false

	fmt.Fprintln(w, "=== vocab required ===")
	if p, err := t.LookPath("go"); err != nil {
		fmt.Fprintln(w, "  ✗ go: not found")
		fail = true
	} else {
		fmt.Fprintf(w, "  ✓ go: %s\n", p)
	}
	switch {
	case t.LLVMErr != nil:
		fmt.Fprintf(w, "  ✗ llvm: %v\n", t.LLVMErr)
		fail = true
	case t.Clang == "" && t.LLC == "":
		fmt.Fprintln(w, "  ✗ llvm: no llc and no clang")
		fail = true
	default:
		if t.LLC != "" {
			fmt.Fprintf(w, "  ✓ llc: %s\n", t.LLC)
		} else {
			fmt.Fprintln(w, "  − llc: not found (IR→object via clang -x ir)")
		}
		if t.Clang != "" {
			fmt.Fprintf(w, "  ✓ clang: %s\n", t.Clang)
		} else {
			fmt.Fprintln(w, "  ✗ clang: not found")
			fail = true
		}
	}
	if p, err := t.LookPath("curl"); err != nil {
		fmt.Fprintln(w, "  ✗ curl: not found (the curl command, used by mobile check)")
		fail = true
	} else {
		fmt.Fprintf(w, "  ✓ curl: %s\n", p)
	}
	if p, err := t.LookPath("cc"); err != nil {
		fmt.Fprintln(w, "  − cc: not found (host check links c_host_check.c with cc -ldl)")
	} else {
		fmt.Fprintf(w, "  ✓ cc: %s\n", p)
	}
	if t.GOOS == "windows" {
		fmt.Fprintln(w, "  − bash scripts need WSL (Ubuntu). Do not run apps/vocab/*.sh from cmd or PowerShell. kylix mobile itself runs on Windows.")
	} else if p, err := t.LookPath("bash"); err != nil {
		fmt.Fprintln(w, "  − bash: not found (kylix mobile does not need the .sh scripts)")
	} else {
		fmt.Fprintf(w, "  ✓ bash: %s\n", p)
	}

	fmt.Fprintln(w, "\n=== Android NDK ===")
	tag := t.NDK.HostTag
	if tag == "" {
		tag = llvmgen.HostPrebuiltTag()
	}
	fmt.Fprintf(w, "  host prebuilt directory: %s\n", tag)
	switch {
	case t.NDK.Found != nil:
		fmt.Fprintf(w, "  ✓ NDK found: %s\n", t.NDK.Found.NDKRoot)
		fmt.Fprintf(w, "  ✓ clang: %s\n", t.NDK.Found.Clang)
	case t.NDK.EnvPinned:
		fmt.Fprintln(w, "  ✗ NDK: ANDROID_NDK_HOME or ANDROID_NDK_ROOT does not contain the host clang")
		fmt.Fprintln(w, indent(t.NDK.MissError().Error()))
		fail = true
	default:
		fmt.Fprintln(w, "  − NDK: not found (set ANDROID_NDK_HOME to build the Android .so). The flashcard host check does not need it.")
		if msg := t.NDK.MissError(); msg != nil {
			fmt.Fprintln(w, indent(msg.Error()))
		}
	}

	fmt.Fprintln(w, "\n=== iOS (macOS only) ===")
	if t.GOOS != "darwin" {
		fmt.Fprintln(w, "  − not this host. iOS linking needs macOS, a full Xcode install, and xcodegen.")
	} else {
		reportXcode(t, w)
	}

	fmt.Fprintln(w, "\n=== not used by the vocab example ===")
	fmt.Fprintln(w, "  − libcurl does not affect the vocab example (the library, not the curl command; db / httpclient would link it)")
	fmt.Fprintln(w, "  − libgc does not affect the vocab example (only --gc=boehm links it)")

	fmt.Fprintln(w)
	if fail {
		fmt.Fprintln(w, "✗ mobile doctor: missing required tooling (see above).")
		return 1
	}
	fmt.Fprintln(w, "✓ mobile doctor: vocab tools are ready.")
	return 0
}

func reportXcode(t Toolchain, w io.Writer) {
	run := t.Run
	if run == nil {
		fmt.Fprintln(w, "  − xcode-select: not checked")
	} else if out, err := run("xcode-select", "-p"); err != nil {
		fmt.Fprintln(w, "  − xcode-select: not found. iOS builds need a full Xcode install.")
	} else {
		fmt.Fprintf(w, "  ✓ xcode-select: %s\n", strings.TrimSpace(out))
	}
	if p, err := t.LookPath("xcodegen"); err != nil {
		fmt.Fprintln(w, "  − xcodegen: not found (needed to generate the Xcode project: brew install xcodegen)")
	} else {
		fmt.Fprintf(w, "  ✓ xcodegen: %s\n", p)
	}
}

func indent(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = "    " + line
	}
	return strings.Join(lines, "\n")
}
