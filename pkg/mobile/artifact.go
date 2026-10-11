package mobile

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"kylix/pkg/llvmgen"
)

func verifyAndroid(root, so string, abi androidABI, w io.Writer) error {
	if _, err := os.Stat(so); err != nil {
		return fmt.Errorf("missing %s", so)
	}
	desc, err := fileDesc(so)
	if err != nil {
		return err
	}
	fmt.Fprintln(w, desc)
	if !strings.Contains(desc, abi.FileSubstr) {
		return fmt.Errorf("expected file(1) to contain %q\nthis .so is not the ABI %s. %s\nfile: %s", abi.FileSubstr, abi.Name, EmulatorSentence, desc)
	}
	blob, err := nmDefined(so, true)
	if err != nil {
		return err
	}
	fmt.Fprintln(w, filterSymLines(blob))
	if err := requireSymbols(blob, filepath.Join(root, "apps", "vocab", "vocab_exports.list")); err != nil {
		return err
	}
	probe := llvmgen.ProbeAndroidNdk()
	if probe.Found == nil {
		if probe.EnvPinned {
			return probe.MissError()
		}
		fmt.Fprintln(w, "ANDROID_NDK_HOME is unset; skipped the JNI link (symbol table already matched)")
	} else {
		if err := linkJNI(root, so, abi, probe.Found); err != nil {
			return err
		}
		fmt.Fprintf(w, "jni link OK (%s)\n", abi.Name)
	}
	return nil
}

func linkJNI(root, so string, abi androidABI, ndk *llvmgen.AndroidToolchain) error {
	clang := filepath.Join(filepath.Dir(ndk.Clang), abi.ClangPrefix+"-clang")
	if st, err := os.Stat(clang); err != nil || st.IsDir() {
		return fmt.Errorf("NDK clang not found: %s", clang)
	}
	work, err := os.MkdirTemp("", "kylix-vocab-jni-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(work)
	cmd := exec.Command(clang,
		"-shared", "-fPIC", "-Wl,--no-undefined",
		"-o", filepath.Join(work, "libvocabjni.so"),
		filepath.Join(root, "apps", "vocab", "android", "app", "src", "main", "cpp", "kylix_jni.c"),
		"-L"+filepath.Dir(so), "-lkylixvocab",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("jni link failed: %w\n%s", err, out)
	}
	return nil
}

func verifyIOS(root, archive string, kind iosKind, w io.Writer) error {
	if !runtimeDarwin() {
		return llvmgen.IosHostError()
	}
	if _, err := os.Stat(archive); err != nil {
		return fmt.Errorf("missing %s", archive)
	}
	desc, err := fileDesc(archive)
	if err != nil {
		return err
	}
	fmt.Fprintln(w, desc)
	if !strings.Contains(desc, "ar archive") {
		return fmt.Errorf("expected an ar archive\n%s", desc)
	}
	blob, err := nmDefined(archive, false)
	if err != nil {
		return err
	}
	fmt.Fprintln(w, filterSymLines(blob))
	if err := requireSymbols(blob, filepath.Join(root, "apps", "vocab", "vocab_exports.list")); err != nil {
		return err
	}
	return linkIOSSmoke(archive, kind, w)
}

func linkIOSSmoke(archive string, kind iosKind, w io.Writer) error {
	tmp, err := os.MkdirTemp("", "kylix-vocab-ios-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	src := filepath.Join(tmp, "smoke.c")
	if err := os.WriteFile(src, []byte(iosSmokeC), 0o644); err != nil {
		return err
	}
	sdkRoot, err := exec.Command("xcrun", "--sdk", kind.SDK, "--show-sdk-path").Output()
	if err != nil {
		return fmt.Errorf("xcrun sdk %s: %w", kind.SDK, err)
	}
	bin := filepath.Join(tmp, "smoke")
	cmd := exec.Command("xcrun", "--sdk", kind.SDK, "clang",
		"-arch", "arm64", "--target="+kind.Triple,
		"-isysroot", strings.TrimSpace(string(sdkRoot)),
		"-o", bin, src, archive,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ios link smoke failed: %w\n%s", err, out)
	}
	linked, err := fileDesc(bin)
	if err != nil {
		return err
	}
	fmt.Fprintln(w, linked)
	if !strings.Contains(linked, "Mach-O 64-bit executable arm64") {
		return fmt.Errorf("link smoke did not produce an arm64 Mach-O executable\n%s", linked)
	}
	build, err := exec.Command("vtool", "-show-build", bin).CombinedOutput()
	if err != nil {
		return fmt.Errorf("vtool: %w\n%s", err, build)
	}
	fmt.Fprintln(w, strings.TrimSpace(string(build)))
	text := string(build)
	if kind.Platform == "IOSSIMULATOR" {
		if !regexp.MustCompile(`platform IOSSIMULATOR([[:space:]]|$)`).MatchString(text) {
			return fmt.Errorf("expected platform IOSSIMULATOR")
		}
	} else {
		if strings.Contains(text, "IOSSIMULATOR") {
			return fmt.Errorf("device link produced a simulator binary")
		}
		if !regexp.MustCompile(`platform IOS([[:space:]]|$)`).MatchString(text) {
			return fmt.Errorf("expected platform IOS")
		}
	}
	return nil
}

func fileDesc(path string) (string, error) {
	out, err := exec.Command("file", path).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("file %s: %w\n%s", path, err, out)
	}
	return strings.TrimSpace(string(out)), nil
}

func nmDefined(path string, dynamic bool) (string, error) {
	bin := "nm"
	for _, name := range []string{"llvm-nm", "llvm-nm-18", "nm"} {
		if p, err := exec.LookPath(name); err == nil {
			bin = p
			break
		}
	}
	args := []string{}
	if dynamic {
		args = append(args, "-D", "--defined-only")
	}
	args = append(args, path)
	out, err := exec.Command(bin, args...).CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("%s: %w\n%s", bin, err, out)
	}
	return string(out), nil
}

func filterSymLines(blob string) string {
	var keep []string
	for _, line := range strings.Split(blob, "\n") {
		if strings.Contains(line, "vc_") || strings.Contains(line, "kylix_free") {
			keep = append(keep, line)
		}
	}
	return strings.Join(keep, "\n")
}

func requireSymbols(blob, listPath string) error {
	data, err := os.ReadFile(listPath)
	if err != nil {
		return err
	}
	for _, line := range strings.Split(string(data), "\n") {
		sym := strings.TrimSpace(line)
		if sym == "" || strings.HasPrefix(sym, "#") {
			continue
		}
		re := regexp.MustCompile(`(?m)(^|[^A-Za-z0-9])_?` + regexp.QuoteMeta(sym) + `($|[^A-Za-z0-9_])`)
		if !re.MatchString(blob) {
			return fmt.Errorf("missing dynamic symbol %s", sym)
		}
	}
	return nil
}

const iosSmokeC = `extern const char *vc_words_path(void);
extern const char *vc_review_path(void);
extern const char *vc_mark_path(void);
extern void kylix_free(void *);
int main(void) {
    const char *words = vc_words_path();
    const char *review = vc_review_path();
    const char *mark = vc_mark_path();
    if (words == 0 || review == 0 || mark == 0) {
        return 2;
    }
    kylix_free((void *)words);
    kylix_free((void *)review);
    kylix_free((void *)mark);
    return 0;
}
`
