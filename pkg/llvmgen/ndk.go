package llvmgen

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
)

// ndk.go — Android NDK toolchain discovery for cross-compilation (v0.14.0).
//
// Follows the FindMingwSysroot paradigm from v0.7.1 P2: looks up standard
// environment variables and default installation paths to locate the Android
// NDK's prebuilt LLVM toolchain.

// AndroidToolchain holds the resolved NDK paths.
type AndroidToolchain struct {
	NDKRoot string // e.g. /opt/android-ndk-r26b or ~/Library/Android/sdk/ndk/26.1.10909125
	Clang   string // path to toolchains/llvm/prebuilt/<host>/bin/clang
	Sysroot string // path to toolchains/llvm/prebuilt/<host>/sysroot
}

// FindAndroidNdk searches for an Android NDK installation.
// Returns nil if no valid NDK toolchain was found.
func FindAndroidNdk() *AndroidToolchain {
	var candidates []string

	// 1. Explicit environment variables
	for _, env := range []string{"ANDROID_NDK_HOME", "ANDROID_NDK_ROOT"} {
		if val := os.Getenv(env); val != "" {
			candidates = append(candidates, val)
		}
	}

	// 2. Subdirectories in ANDROID_HOME / ANDROID_SDK_ROOT
	for _, env := range []string{"ANDROID_HOME", "ANDROID_SDK_ROOT"} {
		if val := os.Getenv(env); val != "" {
			ndkBase := filepath.Join(val, "ndk")
			if entries, err := os.ReadDir(ndkBase); err == nil {
				var versions []string
				for _, e := range entries {
					if e.IsDir() {
						versions = append(versions, filepath.Join(ndkBase, e.Name()))
					}
				}
				// Sort descending so the highest NDK version is probed first
				sort.Sort(sort.Reverse(sort.StringSlice(versions)))
				candidates = append(candidates, versions...)
			}
		}
	}

	// 3. Platform-specific default paths
	home, _ := os.UserHomeDir()
	if home != "" {
		switch runtime.GOOS {
		case "darwin":
			ndkBase := filepath.Join(home, "Library", "Android", "sdk", "ndk")
			if entries, err := os.ReadDir(ndkBase); err == nil {
				for _, e := range entries {
					if e.IsDir() {
						candidates = append(candidates, filepath.Join(ndkBase, e.Name()))
					}
				}
			}
		case "linux":
			ndkBase := filepath.Join(home, "Android", "Sdk", "ndk")
			if entries, err := os.ReadDir(ndkBase); err == nil {
				for _, e := range entries {
					if e.IsDir() {
						candidates = append(candidates, filepath.Join(ndkBase, e.Name()))
					}
				}
			}
			candidates = append(candidates, "/opt/android-ndk")
		}
	}

	// Probe each candidate for the prebuilt LLVM toolchain
	hostTag := hostPrebuiltTag()
	for _, root := range candidates {
		prebuiltDir := filepath.Join(root, "toolchains", "llvm", "prebuilt", hostTag)
		clangPath := filepath.Join(prebuiltDir, "bin", "clang")
		if runtime.GOOS == "windows" {
			clangPath += ".exe"
		}
		sysrootPath := filepath.Join(prebuiltDir, "sysroot")

		if st, err := os.Stat(clangPath); err == nil && !st.IsDir() {
			return &AndroidToolchain{
				NDKRoot: root,
				Clang:   clangPath,
				Sysroot: sysrootPath,
			}
		}
	}

	return nil
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
