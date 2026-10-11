package mobile

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"kylix/pkg/llvmgen"
)

func TestDoctor_ReadyWithoutLibcurl(t *testing.T) {
	var buf bytes.Buffer
	code := Doctor(Toolchain{
		GOOS: "linux",
		LookPath: func(name string) (string, error) {
			return "/usr/bin/" + name, nil
		},
		NDK: llvmgen.NdkProbe{
			HostTag: "linux-x86_64",
			Found: &llvmgen.AndroidToolchain{
				NDKRoot: "/opt/android-ndk-r26d",
				Clang:   "/opt/android-ndk-r26d/toolchains/llvm/prebuilt/linux-x86_64/bin/clang",
			},
		},
		Clang: "/usr/bin/clang",
		LLC:   "/usr/bin/llc",
	}, &buf)
	out := buf.String()
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, out)
	}
	for _, s := range []string{
		"NDK found:",
		"/opt/android-ndk-r26d",
		"linux-x86_64",
		"libcurl does not affect the vocab example",
		"libgc does not affect the vocab example",
		"✓ mobile doctor: vocab tools are ready.",
	} {
		if !strings.Contains(out, s) {
			t.Errorf("missing %q\n%s", s, out)
		}
	}
	if strings.Contains(out, "needed by db / httpclient / crypto") {
		t.Fatalf("libcurl/libgc must not use the old required-lib failure line:\n%s", out)
	}
}

func TestDoctor_WrongNDKPinFails(t *testing.T) {
	bad := t.TempDir()
	t.Setenv("ANDROID_NDK_HOME", bad)
	t.Setenv("ANDROID_NDK_ROOT", "")
	probe := llvmgen.ProbeAndroidNdk()
	var buf bytes.Buffer
	code := Doctor(Toolchain{
		GOOS: "linux",
		LookPath: func(name string) (string, error) {
			return "/usr/bin/" + name, nil
		},
		NDK:   probe,
		Clang: "/usr/bin/clang",
	}, &buf)
	out := buf.String()
	if code == 0 {
		t.Fatalf("expected non-zero exit\n%s", out)
	}
	if !strings.Contains(out, bad) || !strings.Contains(out, probe.HostTag) {
		t.Fatalf("expected tried path and host tag %s\n%s", probe.HostTag, out)
	}
}

func TestDoctor_WrongHomeIgnoresValidRoot(t *testing.T) {
	bad := t.TempDir()
	good := t.TempDir()
	tag := llvmgen.HostPrebuiltTag()
	clang := filepath.Join(good, "toolchains", "llvm", "prebuilt", tag, "bin", "clang")
	if runtime.GOOS == "windows" {
		clang += ".exe"
	}
	if err := os.MkdirAll(filepath.Dir(clang), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(clang, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ANDROID_NDK_HOME", bad)
	t.Setenv("ANDROID_NDK_ROOT", good)
	var buf bytes.Buffer
	code := Doctor(Toolchain{
		GOOS: "linux",
		LookPath: func(name string) (string, error) {
			return "/usr/bin/" + name, nil
		},
		NDK:   llvmgen.ProbeAndroidNdk(),
		Clang: "/usr/bin/clang",
	}, &buf)
	out := buf.String()
	if code == 0 {
		t.Fatalf("expected non-zero exit while ROOT is valid\n%s", out)
	}
	if strings.Contains(out, "NDK found:") || !strings.Contains(out, bad) || !strings.Contains(out, tag) {
		t.Fatalf("doctor should report the bad HOME, not the ROOT NDK:\n%s", out)
	}
}

func TestDoctor_WindowsMentionsWSL(t *testing.T) {
	var buf bytes.Buffer
	code := Doctor(Toolchain{
		GOOS: "windows",
		LookPath: func(name string) (string, error) {
			if name == "bash" || name == "cc" {
				return "", os.ErrNotExist
			}
			return `C:\tools\` + name + ".exe", nil
		},
		NDK: llvmgen.NdkProbe{
			HostTag: "windows-x86_64",
			Found:   &llvmgen.AndroidToolchain{NDKRoot: `C:\ndk`, Clang: `C:\ndk\clang.exe`},
		},
		Clang: `C:\llvm\clang.exe`,
	}, &buf)
	out := buf.String()
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, out)
	}
	if !strings.Contains(out, "bash scripts need WSL") {
		t.Fatalf("missing WSL guidance\n%s", out)
	}
}

func TestSources_PhoneLibraryOmitsServer(t *testing.T) {
	for _, src := range vocabCoreSources("/repo") {
		if strings.HasSuffix(src, "server.klx") {
			t.Fatalf("phone library must not link server.klx: %v", vocabCoreSources("/repo"))
		}
	}
	joined := strings.Join(vocabCoreSources("/repo"), " ")
	for _, name := range []string{"stringutil.klx", "vocab.klx", "vocab_lib.klx"} {
		if !strings.Contains(joined, name) {
			t.Fatalf("missing %s in %s", name, joined)
		}
	}
}

func TestParseAndroidABI(t *testing.T) {
	amd, err := parseAndroidABI("amd64")
	if err != nil || amd.JniDir != "x86_64" {
		t.Fatalf("amd64: %+v %v", amd, err)
	}
	arm, err := parseAndroidABI("arm64")
	if err != nil || arm.JniDir != "arm64-v8a" {
		t.Fatalf("arm64: %+v %v", arm, err)
	}
	_, err = parseAndroidABI("x86_64")
	if err == nil || !strings.Contains(err.Error(), EmulatorSentence) {
		t.Fatalf("x86_64 should be rejected with the emulator sentence, got %v", err)
	}
}

func TestBuildIOS_NonMacWritesNothing(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("macOS is allowed to link")
	}
	err := BuildIOS("", "simulator", io.Discard)
	if err == nil {
		t.Fatal("expected an error")
	}
	msg := err.Error()
	for _, s := range []string{"macOS", "Xcode", "xcodegen", "does not write a .a"} {
		if !strings.Contains(msg, s) {
			t.Errorf("missing %q\n%s", s, msg)
		}
	}
}

func TestBuildAndroid_WrongNDKBeforeCompile(t *testing.T) {
	bad := t.TempDir()
	t.Setenv("ANDROID_NDK_HOME", bad)
	t.Setenv("ANDROID_NDK_ROOT", "")
	err := BuildAndroid("", "amd64", io.Discard)
	if err == nil {
		t.Fatal("expected NDK miss")
	}
	msg := err.Error()
	if !strings.Contains(msg, bad) || !strings.Contains(msg, llvmgen.HostPrebuiltTag()) {
		t.Fatalf("miss error should name the path and host tag:\n%s", msg)
	}
}

func TestRequireSymbols(t *testing.T) {
	dir := t.TempDir()
	list := filepath.Join(dir, "exports.list")
	if err := os.WriteFile(list, []byte("# comment\n\nvc_browse\nkylix_free\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	blob := "0000000000001120 T vc_browse\n0000000000001200 T _kylix_free\n"
	if err := requireSymbols(blob, list); err != nil {
		t.Fatal(err)
	}
	if err := requireSymbols("T other\n", list); err == nil {
		t.Fatal("expected a missing symbol")
	}
}

func TestParityLine(t *testing.T) {
	got, err := parityLine("browse={\"id\":1}\nmark=ok\n", "browse")
	if err != nil || got != `{"id":1}` {
		t.Fatalf("got %q err %v", got, err)
	}
}

func TestGitignore_CMakeListsNotIgnored(t *testing.T) {
	root, err := FindRoot("")
	if err != nil {
		t.Fatal(err)
	}
	// Not under an ignored directory: a parent ignore would hide CMakeLists.txt
	// no matter what the !CMakeLists.txt exception says.
	dir := filepath.Join(root, "apps", "vocab", "android", "ignore-probe")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	cmake := filepath.Join(dir, "CMakeLists.txt")
	other := filepath.Join(dir, "scratch.txt")
	if err := os.WriteFile(cmake, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(other, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if ignored(root, cmake) {
		t.Fatalf("CMakeLists.txt is ignored: %s", cmake)
	}
	if !ignored(root, other) {
		t.Fatalf("scratch.txt should still be ignored: %s", other)
	}
}

func ignored(root, path string) bool {
	cmd := exec.Command("git", "check-ignore", "-q", path)
	cmd.Dir = root
	err := cmd.Run()
	return err == nil
}
