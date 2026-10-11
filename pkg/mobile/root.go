// Package mobile implements `kylix mobile` (v0.16.0): the flashcard
// doctor's sections, the host/e2e checks, the foreground server, and the
// Android/iOS library builds. The apps/vocab shell scripts call this
// package so there is one implementation.
package mobile

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	vocabPort = 8091

	// Success lines locked by the beginner tutorial and by CI.
	HostPass    = "[vocab] host check PASS"
	E2EPassGo   = "[vocab] e2e PASS (go)"
	E2EPassLLVM = "[vocab] e2e PASS (llvm)"

	// EmulatorSentence tells a beginner which .so an emulator will load.
	EmulatorSentence = "x86_64 emulators use the amd64 build (jniLibs/x86_64); arm64 emulators use the arm64 build (jniLibs/arm64-v8a)."
)

// FindRoot walks up from start looking for the kylix module that contains
// apps/vocab. KYLIX_ROOT overrides the walk when it points at that module.
func FindRoot(start string) (string, error) {
	if env := os.Getenv("KYLIX_ROOT"); env != "" {
		if isKylixRoot(env) {
			return env, nil
		}
		return "", fmt.Errorf("KYLIX_ROOT=%s is not a kylix checkout (apps/vocab/vocab.klx missing)", env)
	}
	dir := start
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}
	for {
		if isKylixRoot(dir) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("kylix mobile: run this from a kylix checkout (apps/vocab/vocab.klx not found)")
}

func isKylixRoot(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err != nil {
		return false
	}
	_, err := os.Stat(filepath.Join(dir, "apps", "vocab", "vocab.klx"))
	return err == nil
}

func vocabCoreSources(root string) []string {
	return []string{
		filepath.Join(root, "stdlib", "stringutil.klx"),
		filepath.Join(root, "apps", "vocab", "vocab.klx"),
		filepath.Join(root, "apps", "vocab", "vocab_lib.klx"),
	}
}

func vocabParitySources(root string) []string {
	return []string{
		filepath.Join(root, "stdlib", "stringutil.klx"),
		filepath.Join(root, "apps", "vocab", "vocab.klx"),
		filepath.Join(root, "apps", "vocab", "parity.klx"),
	}
}

func vocabServerSources(root string) []string {
	return []string{
		filepath.Join(root, "stdlib", "stringutil.klx"),
		filepath.Join(root, "apps", "vocab", "vocab.klx"),
		filepath.Join(root, "apps", "vocab", "server.klx"),
	}
}

// androidABI is the only pair of names `kylix mobile build android` accepts.
type androidABI struct {
	Name        string // amd64 | arm64
	Target      string
	JniDir      string // x86_64 | arm64-v8a
	FileSubstr  string
	ClangPrefix string
}

func parseAndroidABI(abi string) (androidABI, error) {
	switch abi {
	case "amd64":
		return androidABI{
			Name:        "amd64",
			Target:      "android/amd64",
			JniDir:      "x86_64",
			FileSubstr:  "ELF 64-bit LSB shared object, x86-64",
			ClangPrefix: "x86_64-linux-android30",
		}, nil
	case "arm64":
		return androidABI{
			Name:        "arm64",
			Target:      "android/arm64",
			JniDir:      "arm64-v8a",
			FileSubstr:  "ELF 64-bit LSB shared object, ARM aarch64",
			ClangPrefix: "aarch64-linux-android30",
		}, nil
	default:
		return androidABI{}, fmt.Errorf("kylix mobile build android: ABI must be amd64 or arm64 (got %q)\n%s", abi, EmulatorSentence)
	}
}

func androidSO(root string, abi androidABI) string {
	return filepath.Join(root, "apps", "vocab", "android", "app", "src", "main", "jniLibs", abi.JniDir, "libkylixvocab.so")
}

type iosKind struct {
	Name     string
	Target   string
	SDK      string
	Triple   string
	Platform string
}

func parseIosKind(kind string) (iosKind, error) {
	switch kind {
	case "simulator", "sim":
		return iosKind{
			Name:     "simulator",
			Target:   "ios/simulator-arm64",
			SDK:      "iphonesimulator",
			Triple:   "arm64-apple-ios16.0.0-simulator",
			Platform: "IOSSIMULATOR",
		}, nil
	case "device":
		return iosKind{
			Name:     "device",
			Target:   "ios/arm64",
			SDK:      "iphoneos",
			Triple:   "arm64-apple-ios16.0.0",
			Platform: "IOS",
		}, nil
	default:
		return iosKind{}, fmt.Errorf("kylix mobile build ios: kind must be simulator or device (got %q)", kind)
	}
}

func iosArchive(root string) string {
	return filepath.Join(root, "apps", "vocab", "ios", "Sources", "CKylixVocab", "lib", "libkylixvocab.a")
}
