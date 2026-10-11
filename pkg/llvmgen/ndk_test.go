package llvmgen_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"kylix/pkg/llvmgen"
)

func TestProbeAndroidNdk_WrongPinDoesNotFallThrough(t *testing.T) {
	bad := t.TempDir()
	// A Windows-shaped prebuilt tree must not satisfy a Linux/macOS host.
	other := "windows-x86_64"
	if runtime.GOOS == "windows" {
		other = "linux-x86_64"
	}
	if err := os.MkdirAll(filepath.Join(bad, "toolchains", "llvm", "prebuilt", other, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ANDROID_NDK_HOME", bad)
	t.Setenv("ANDROID_NDK_ROOT", "")

	probe := llvmgen.ProbeAndroidNdk()
	if !probe.EnvPinned {
		t.Fatal("expected a pinned search")
	}
	if probe.Found != nil {
		t.Fatalf("wrong ANDROID_NDK_HOME was accepted: %+v", probe.Found)
	}
	if probe.HostTag != llvmgen.HostPrebuiltTag() {
		t.Fatalf("host tag %q", probe.HostTag)
	}
	msg := probe.MissError().Error()
	if !strings.Contains(msg, bad) {
		t.Fatalf("miss error omitted the tried path:\n%s", msg)
	}
	if !strings.Contains(msg, probe.HostTag) {
		t.Fatalf("miss error omitted host tag:\n%s", msg)
	}
	if !strings.Contains(msg, other) {
		t.Fatalf("miss error should name the prebuilt dir that is present (%s):\n%s", other, msg)
	}
}

func TestProbeAndroidNdk_FindsHostClang(t *testing.T) {
	root := t.TempDir()
	tag := llvmgen.HostPrebuiltTag()
	clang := filepath.Join(root, "toolchains", "llvm", "prebuilt", tag, "bin", "clang")
	if runtime.GOOS == "windows" {
		clang += ".exe"
	}
	if err := os.MkdirAll(filepath.Dir(clang), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(clang, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ANDROID_NDK_HOME", root)
	t.Setenv("ANDROID_NDK_ROOT", "")

	probe := llvmgen.ProbeAndroidNdk()
	if probe.Found == nil {
		t.Fatalf("expected NDK, miss:\n%v", probe.MissError())
	}
	if probe.Found.NDKRoot != root {
		t.Fatalf("root %q", probe.Found.NDKRoot)
	}
	if probe.Found.Clang != clang {
		t.Fatalf("clang %q", probe.Found.Clang)
	}
	if probe.MissError() != nil {
		t.Fatal(probe.MissError())
	}
}

func TestIosHostError_NamesMacXcodeAndXcodegen(t *testing.T) {
	msg := llvmgen.IosHostError().Error()
	for _, s := range []string{
		"iOS archives must be linked on macOS with Xcode command line tools.",
		"On this host, stop here. The Swift sources are in apps/vocab/ios/.",
		"ios cross-link requires a macOS host with Xcode command line tools",
		"xcodegen",
		"does not write a .a",
	} {
		if !strings.Contains(msg, s) {
			t.Errorf("missing %q\n%s", s, msg)
		}
	}
}

func TestIntermediateBase_LeavesSourceDir(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "apps", "vocab", "vocab_lib.klx")
	out := filepath.Join(root, "jniLibs", "x86_64", "libkylixvocab.so")
	mid := filepath.Join(root, "mid")

	got, err := llvmgen.IntermediateBase(src, out, mid)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(mid, "vocab_lib")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}

	got, err = llvmgen.IntermediateBase(src, out, "")
	if err != nil {
		t.Fatal(err)
	}
	want = filepath.Join(root, "jniLibs", "x86_64", "libkylixvocab")
	if got != want {
		t.Fatalf("follow -o: got %q want %q", got, want)
	}

	same, err := llvmgen.IntermediateBase(src, filepath.Join(filepath.Dir(src), "vocab_lib"), "")
	if err != nil {
		t.Fatal(err)
	}
	if same != strings.TrimSuffix(src, ".klx") {
		t.Fatalf("same directory should stay beside the source, got %q", same)
	}
}
