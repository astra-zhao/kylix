package llvmgen

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// ndk.go — Android NDK toolchain discovery for cross-compilation (v0.14.0).
//
// Follows the FindMingwSysroot paradigm from v0.7.1 P2: looks up standard
// environment variables and default installation paths to locate the Android
// NDK's prebuilt LLVM toolchain.
//
// v0.16.0: when ANDROID_NDK_HOME or ANDROID_NDK_ROOT is set, only those
// paths are probed. A wrong pin must not silently fall through to another
// NDK, and the error names the host prebuilt directory (linux-x86_64, …)
// plus every path that was tried.

// AndroidToolchain holds the resolved NDK paths.
type AndroidToolchain struct {
	NDKRoot string // e.g. /opt/android-ndk-r26b or ~/Library/Android/sdk/ndk/26.1.10909125
	Clang   string // path to toolchains/llvm/prebuilt/<host>/bin/clang
	Sysroot string // path to toolchains/llvm/prebuilt/<host>/sysroot
}

// NdkCandidate is one directory that was probed for the host clang.
type NdkCandidate struct {
	Root     string
	Source   string   // ANDROID_NDK_HOME, ANDROID_HOME/ndk, …
	Clang    string   // clang path that was required
	Prebuilt []string // sibling directory names when the host tag is absent
}

// NdkProbe is the result of looking for the Android NDK.
type NdkProbe struct {
	HostTag   string
	Tried     []NdkCandidate
	Found     *AndroidToolchain
	EnvPinned bool // ANDROID_NDK_HOME or ANDROID_NDK_ROOT was non-empty
}

// HostPrebuiltTag is the NDK prebuilt directory name for this machine
// (linux-x86_64, darwin-x86_64, or windows-x86_64).
func HostPrebuiltTag() string {
	return hostPrebuiltTag()
}

// FindAndroidNdk searches for an Android NDK installation.
// Returns nil if no valid NDK toolchain was found.
func FindAndroidNdk() *AndroidToolchain {
	return ProbeAndroidNdk().Found
}

// ProbeAndroidNdk searches for the NDK and records every path it tried.
func ProbeAndroidNdk() NdkProbe {
	hostTag := hostPrebuiltTag()
	probe := NdkProbe{HostTag: hostTag}

	var candidates []struct{ root, source string }

	for _, env := range []string{"ANDROID_NDK_HOME", "ANDROID_NDK_ROOT"} {
		if val := os.Getenv(env); val != "" {
			probe.EnvPinned = true
			candidates = append(candidates, struct{ root, source string }{val, env})
		}
	}

	// A pinned path that does not contain the host clang is an error.
	// Do not continue into the SDK default locations: beginners who point
	// ANDROID_NDK_HOME at the wrong tree (a Windows NDK inside WSL, a typo)
	// must see that path, not a different NDK that happened to be installed.
	if !probe.EnvPinned {
		candidates = append(candidates, defaultNdkCandidates()...)
	}

	for _, c := range candidates {
		clang, sysroot, ok := clangInNDK(c.root, hostTag)
		tried := NdkCandidate{
			Root:   c.root,
			Source: c.source,
			Clang:  clang,
		}
		if !ok {
			tried.Prebuilt = listPrebuiltHosts(c.root)
		}
		probe.Tried = append(probe.Tried, tried)
		if ok && probe.Found == nil {
			probe.Found = &AndroidToolchain{
				NDKRoot: c.root,
				Clang:   clang,
				Sysroot: sysroot,
			}
		}
	}
	return probe
}

// MissError describes a failed NDK search. The text includes the host
// prebuilt directory name and every path that was tried.
func (p NdkProbe) MissError() error {
	if p.Found != nil {
		return nil
	}
	var b strings.Builder
	fmt.Fprintf(&b, "android cross-link needs Android NDK (not found): install the NDK and point ANDROID_NDK_HOME at it (it must contain toolchains/llvm/prebuilt/%s/bin/clang)\n", p.HostTag)
	fmt.Fprintf(&b, "host prebuilt directory: %s\n", p.HostTag)
	if len(p.Tried) == 0 {
		b.WriteString("tried: (none — set ANDROID_NDK_HOME)\n")
	} else {
		b.WriteString("tried:\n")
		for _, c := range p.Tried {
			fmt.Fprintf(&b, "  %s (%s)\n", c.Root, c.Source)
			fmt.Fprintf(&b, "    expected clang: %s\n", c.Clang)
			if len(c.Prebuilt) > 0 {
				fmt.Fprintf(&b, "    prebuilt directories present: %s\n", strings.Join(c.Prebuilt, ", "))
			}
		}
	}
	return fmt.Errorf("%s", strings.TrimRight(b.String(), "\n"))
}

func defaultNdkCandidates() []struct{ root, source string } {
	var candidates []struct{ root, source string }

	for _, env := range []string{"ANDROID_HOME", "ANDROID_SDK_ROOT"} {
		if val := os.Getenv(env); val != "" {
			for _, root := range ndkVersionDirs(filepath.Join(val, "ndk")) {
				candidates = append(candidates, struct{ root, source string }{root, env + "/ndk"})
			}
		}
	}

	home, _ := os.UserHomeDir()
	if home != "" {
		switch runtime.GOOS {
		case "darwin":
			for _, root := range ndkVersionDirs(filepath.Join(home, "Library", "Android", "sdk", "ndk")) {
				candidates = append(candidates, struct{ root, source string }{root, "~/Library/Android/sdk/ndk"})
			}
		case "linux":
			for _, root := range ndkVersionDirs(filepath.Join(home, "Android", "Sdk", "ndk")) {
				candidates = append(candidates, struct{ root, source string }{root, "~/Android/Sdk/ndk"})
			}
			candidates = append(candidates, struct{ root, source string }{"/opt/android-ndk", "/opt/android-ndk"})
		}
	}
	return candidates
}

// ndkVersionDirs lists version directories under an SDK ndk/ folder, highest
// name first (the previous search order).
func ndkVersionDirs(ndkBase string) []string {
	entries, err := os.ReadDir(ndkBase)
	if err != nil {
		return nil
	}
	var versions []string
	for _, e := range entries {
		if e.IsDir() {
			versions = append(versions, filepath.Join(ndkBase, e.Name()))
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(versions)))
	return versions
}

func clangInNDK(root, hostTag string) (clang, sysroot string, ok bool) {
	prebuiltDir := filepath.Join(root, "toolchains", "llvm", "prebuilt", hostTag)
	clang = filepath.Join(prebuiltDir, "bin", "clang")
	if runtime.GOOS == "windows" {
		clang += ".exe"
	}
	sysroot = filepath.Join(prebuiltDir, "sysroot")
	st, err := os.Stat(clang)
	if err != nil || st.IsDir() {
		return clang, sysroot, false
	}
	return clang, sysroot, true
}

func listPrebuiltHosts(root string) []string {
	dir := filepath.Join(root, "toolchains", "llvm", "prebuilt")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names
}

// hostPrebuiltTag returns the prebuilt directory name for the current host.
func hostPrebuiltTag() string {
	switch runtime.GOOS {
	case "darwin":
		return "darwin-x86_64" // NDK universal/x86_64 runs natively or via Rosetta
	case "linux":
		return "linux-x86_64"
	case "windows":
		return "windows-x86_64"
	default:
		return "linux-x86_64"
	}
}

// IosHostError is the non-macOS failure for an iOS link. The first lines are
// the sentences the beginner tutorial already quotes; the rest names Xcode
// and xcodegen. No .a is written when this is returned before codegen.
func IosHostError() error {
	return fmt.Errorf("%s", strings.TrimSpace(`
iOS archives must be linked on macOS with Xcode command line tools.
On this host, stop here. The Swift sources are in apps/vocab/ios/.
ios cross-link requires a macOS host with Xcode command line tools.
Install a full Xcode (xcode-select -p) and xcodegen (brew install xcodegen) on that Mac. This host does not write a .a.
`))
}
