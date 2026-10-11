package mobile

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"kylix/pkg/llvmgen"
)

// CheckOptions selects the host parity check, the curl e2e, or both.
type CheckOptions struct {
	Root     string
	Backend  string // go, llvm, or both. Ignored when HostOnly.
	HostOnly bool
	E2EOnly  bool
	Out      io.Writer
}

// Check runs the flashcard host parity (and C ABI) and/or the curl scenario.
// The server is stopped before Check returns, including on failure.
func Check(opts CheckOptions) error {
	if opts.HostOnly && opts.E2EOnly {
		return fmt.Errorf("kylix mobile check: --host-only and --e2e-only together do not make sense")
	}
	backend := opts.Backend
	if backend == "" {
		backend = "go"
	}
	switch backend {
	case "go", "llvm", "both":
	default:
		return fmt.Errorf("backend must be go, llvm, or both (got %q)", backend)
	}
	w := opts.Out
	if w == nil {
		w = os.Stderr
	}
	root, err := needRoot(opts.Root)
	if err != nil {
		return err
	}
	work, err := workDir(root)
	if err != nil {
		return err
	}
	defer os.RemoveAll(work)

	if !opts.E2EOnly {
		if err := hostCheck(root, work, w); err != nil {
			return err
		}
		fmt.Fprintln(w, HostPass)
	}
	if opts.HostOnly {
		return nil
	}
	if portOpen(vocabPort) {
		return fmt.Errorf("port %d is already in use; stop that process and retry", vocabPort)
	}
	parity, err := goParity(root, work)
	if err != nil {
		return err
	}
	fmt.Fprintln(w, "[vocab] expected bodies from Go parity")
	backends := []string{backend}
	if backend == "both" {
		backends = []string{"go", "llvm"}
	}
	for _, b := range backends {
		if err := e2eBackend(root, work, b, parity, w); err != nil {
			return err
		}
		if b == "llvm" {
			fmt.Fprintln(w, E2EPassLLVM)
		} else {
			fmt.Fprintln(w, E2EPassGo)
		}
	}
	return nil
}

func needRoot(root string) (string, error) {
	if root != "" {
		if !isKylixRoot(root) {
			return "", fmt.Errorf("kylix mobile: %s is not a kylix checkout", root)
		}
		return root, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return FindRoot(wd)
}

func hostCheck(root, work string, w io.Writer) error {
	fmt.Fprintln(w, "[vocab] parity Go vs LLVM")
	goBin := filepath.Join(work, "parity_go")
	if err := compileGo(root, vocabParitySources(root), goBin); err != nil {
		return err
	}
	goTxt := filepath.Join(work, "go.txt")
	if err := captureTo(goBin, goTxt); err != nil {
		return err
	}
	goOut, err := os.ReadFile(goTxt)
	if err != nil {
		return err
	}
	llBin := filepath.Join(work, "parity_ll")
	if err := compileLLVM(vocabParitySources(root), llBin, llvmOpts(work)); err != nil {
		return err
	}
	llTxt := filepath.Join(work, "ll.txt")
	if err := captureTo(llBin, llTxt); err != nil {
		return err
	}
	llOut, err := os.ReadFile(llTxt)
	if err != nil {
		return err
	}
	if string(goOut) != string(llOut) {
		return fmt.Errorf("PARITY DIFF\n--- go\n%s\n--- llvm\n%s", goOut, llOut)
	}
	fmt.Fprintln(w, "[vocab] parity PASS")

	fmt.Fprintln(w, "[vocab] C ABI .so")
	so := filepath.Join(work, "libkylixvocab.so")
	if err := compileLLVM(vocabCoreSources(root), so, sharedOpts(work, "")); err != nil {
		return err
	}
	hostBin := filepath.Join(work, "c_host_check")
	cc := exec.Command("cc", "-o", hostBin, filepath.Join(root, "apps", "vocab", "c_host_check.c"), "-ldl")
	if out, err := cc.CombinedOutput(); err != nil {
		return fmt.Errorf("cc -ldl failed (install a C compiler): %w\n%s", err, out)
	}
	run := exec.Command(hostBin, so)
	run.Stdout = w
	run.Stderr = w
	if err := run.Run(); err != nil {
		return fmt.Errorf("c_host_check: %w", err)
	}
	return nil
}

func goParity(root, work string) (string, error) {
	bin := filepath.Join(work, "parity_go_e2e")
	if err := compileGo(root, vocabParitySources(root), bin); err != nil {
		return "", err
	}
	path, err := runCapture(bin)
	if err != nil {
		return "", err
	}
	body, err := os.ReadFile(path)
	return string(body), err
}

func e2eBackend(root, work, backend, parity string, w io.Writer) error {
	fmt.Fprintf(w, "[vocab] e2e backend=%s\n", backend)
	expect := map[string]string{}
	for _, key := range []string{"browse", "mark", "review", "again", "both", "bad"} {
		val, err := parityLine(parity, key)
		if err != nil {
			return err
		}
		expect[key] = val
	}
	bin := filepath.Join(work, "vocab_server_"+backend)
	var err error
	if backend == "go" {
		err = compileGo(root, vocabServerSources(root), bin)
	} else {
		err = compileLLVM(vocabServerSources(root), bin, llvmOpts(work))
	}
	if err != nil {
		return err
	}
	logPath := filepath.Join(work, "server_"+backend+".log")
	logf, err := os.Create(logPath)
	if err != nil {
		return err
	}
	defer logf.Close()
	cmd := exec.Command(bin)
	cmd.Stdout = logf
	cmd.Stderr = logf
	if err := cmd.Start(); err != nil {
		return err
	}
	exited := make(chan error, 1)
	go func() { exited <- cmd.Wait() }()
	defer func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}()

	if err := waitReady(cmd, exited, logPath); err != nil {
		return err
	}
	if err := curlSame(work, "browse", expect["browse"], "http://127.0.0.1:8091/api/words"); err != nil {
		return withServerLog(logPath, err)
	}
	if err := curlSame(work, "empty", expect["browse"], "http://127.0.0.1:8091/api/words?known="); err != nil {
		return withServerLog(logPath, err)
	}
	if err := curlSame(work, "both", expect["both"], "http://127.0.0.1:8091/api/words?known=3,1,1"); err != nil {
		return withServerLog(logPath, err)
	}
	if err := curlPostSame(work, "mark", expect["mark"], `{"known":"","id":"1"}`); err != nil {
		return withServerLog(logPath, err)
	}
	if err := curlSame(work, "review", expect["review"], "http://127.0.0.1:8091/api/review?known=1"); err != nil {
		return withServerLog(logPath, err)
	}
	if err := curlPostSame(work, "again", expect["again"], `{"known":"1","id":"1"}`); err != nil {
		return withServerLog(logPath, err)
	}
	badPath := filepath.Join(work, "bad.got")
	code, err := curlCode(badPath, "-X", "POST", "http://127.0.0.1:8091/api/mark",
		"-H", "Content-Type: application/json", "--data", `{"known":"1","id":"9"}`)
	if err != nil {
		return err
	}
	if code != "400" {
		return fmt.Errorf("mark of unknown id status=%s", code)
	}
	gotBad, err := os.ReadFile(badPath)
	if err != nil {
		return err
	}
	if string(gotBad) != expect["bad"] {
		return fmt.Errorf("DIFF bad\n--- want\n%s\n--- got\n%s", expect["bad"], gotBad)
	}
	browse, err := os.ReadFile(filepath.Join(work, "browse.got"))
	if err != nil {
		return err
	}
	review, err := os.ReadFile(filepath.Join(work, "review.got"))
	if err != nil {
		return err
	}
	if !strings.Contains(string(browse), "apple") {
		return fmt.Errorf("browse body lost apple")
	}
	if strings.Contains(string(review), "apple") {
		return fmt.Errorf("review still contains apple")
	}
	return nil
}

func withServerLog(logPath string, err error) error {
	log, readErr := os.ReadFile(logPath)
	if readErr != nil || len(log) == 0 {
		return err
	}
	return fmt.Errorf("%w\nserver log:\n%s", err, log)
}

func waitReady(cmd *exec.Cmd, exited <-chan error, logPath string) error {
	for i := 0; i < 20; i++ {
		// A bare TCP dial is not enough: the LLVM boot server can reset the
		// process when the probe closes before an HTTP request. curl matches
		// apps/vocab/e2e.sh.
		if exec.Command("curl", "-sf", "-o", "/dev/null", "--max-time", "1", "http://127.0.0.1:8091/api/words").Run() == nil {
			return nil
		}
		select {
		case err := <-exited:
			log, _ := os.ReadFile(logPath)
			return fmt.Errorf("server exited: %v\n%s", err, log)
		case <-time.After(250 * time.Millisecond):
		}
	}
	if !portOpen(vocabPort) {
		log, _ := os.ReadFile(logPath)
		_ = cmd.Process.Kill()
		return fmt.Errorf("server did not accept port %d\n%s", vocabPort, log)
	}
	return nil
}

func curlSame(work, name, want, url string) error {
	got := filepath.Join(work, name+".got")
	if _, err := curlCode(got, url); err != nil {
		return fmt.Errorf("curl %s: %w", name, err)
	}
	body, err := os.ReadFile(got)
	if err != nil {
		return err
	}
	if string(body) != want {
		return fmt.Errorf("DIFF %s\n--- want\n%s\n--- got\n%s", name, want, body)
	}
	return nil
}

func curlPostSame(work, name, want, data string) error {
	got := filepath.Join(work, name+".got")
	if _, err := curlCode(got, "-X", "POST", "http://127.0.0.1:8091/api/mark",
		"-H", "Content-Type: application/json", "--data", data); err != nil {
		return fmt.Errorf("curl %s: %w", name, err)
	}
	body, err := os.ReadFile(got)
	if err != nil {
		return err
	}
	if string(body) != want {
		return fmt.Errorf("DIFF %s\n--- want\n%s\n--- got\n%s", name, want, body)
	}
	return nil
}

func curlCode(outPath string, args ...string) (string, error) {
	all := append([]string{"-sS", "--max-time", "5", "-o", outPath, "-w", "%{http_code}"}, args...)
	cmd := exec.Command("curl", all...)
	out, err := cmd.CombinedOutput()
	code := strings.TrimSpace(string(out))
	// -w writes the status after any curl diagnostic on stderr mixed into
	// CombinedOutput. Keep the last line, which is the status.
	if i := strings.LastIndex(code, "\n"); i >= 0 {
		code = strings.TrimSpace(code[i+1:])
	}
	if err != nil {
		return code, fmt.Errorf("%w\n%s", err, out)
	}
	if code != "400" && !strings.HasPrefix(code, "2") {
		return code, fmt.Errorf("HTTP %s", code)
	}
	return code, nil
}

func parityLine(parity, key string) (string, error) {
	prefix := key + "="
	for _, line := range strings.Split(parity, "\n") {
		if strings.HasPrefix(line, prefix) {
			return strings.TrimPrefix(line, prefix), nil
		}
	}
	return "", fmt.Errorf("missing parity line %s", key)
}

func portOpen(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 300*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func runCapture(bin string) (string, error) {
	path := bin + ".txt"
	if err := captureTo(bin, path); err != nil {
		return "", err
	}
	return path, nil
}

func captureTo(bin, path string) error {
	out, err := exec.Command(bin).CombinedOutput()
	if werr := os.WriteFile(path, out, 0o644); werr != nil && err == nil {
		return werr
	}
	if err != nil {
		return fmt.Errorf("%s: %w\n%s", filepath.Base(bin), err, out)
	}
	return nil
}

func llvmOpts(work string) llvmgen.CompileOpts {
	return llvmgen.CompileOpts{IntermediatesDir: filepath.Join(work, "mid")}
}

func sharedOpts(work, target string) llvmgen.CompileOpts {
	return llvmgen.CompileOpts{
		IntermediatesDir: filepath.Join(work, "mid"),
		Shared:           true,
		Target:           target,
	}
}
