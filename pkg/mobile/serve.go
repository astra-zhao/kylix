package mobile

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Serve compiles the flashcard server and leaves it in the foreground on
// port 8091 until the process is interrupted. The server prints
// "vocab server on :8091" itself.
func Serve(root, backend string) error {
	if backend == "" {
		backend = "go"
	}
	if backend != "go" && backend != "llvm" {
		return fmt.Errorf("backend must be go or llvm (got %q)", backend)
	}
	root, err := needRoot(root)
	if err != nil {
		return err
	}
	if portOpen(vocabPort) {
		return fmt.Errorf("port %d is already in use; stop that process and retry", vocabPort)
	}
	work, err := workDir(root)
	if err != nil {
		return err
	}
	defer os.RemoveAll(work)
	bin := filepath.Join(work, "vocab_server")
	if backend == "go" {
		err = compileGo(root, vocabServerSources(root), bin)
	} else {
		err = compileLLVM(vocabServerSources(root), bin, llvmOpts(work))
	}
	if err != nil {
		return err
	}
	cmd := exec.Command(bin)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil && !stoppedBySignal(err) {
		return err
	}
	return nil
}
