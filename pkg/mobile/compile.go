package mobile

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"kylix/pkg/compiler"
	"kylix/pkg/llvmgen"
)

// workDir is a scratch directory inside the module (kylixvocab_gen is
// gitignored at the repo root). Generated .go files must live in the module
// so `go build` can resolve kylix/stdlib. LLVM .ll/.o go in mid/ and are
// removed with the directory.
func workDir(root string) (string, error) {
	base := filepath.Join(root, "kylixvocab_gen")
	if err := os.MkdirAll(base, 0o755); err != nil {
		return "", err
	}
	return os.MkdirTemp(base, "work-")
}

func compileGo(root string, files []string, binOut string) error {
	goOut := binOut + ".gen.go"
	result, err := compiler.CompileProject(files, compiler.Options{OutputFile: goOut})
	if err != nil {
		return err
	}
	if result == nil || !result.Success {
		var b strings.Builder
		if result != nil {
			for _, d := range result.Diagnostics {
				fmt.Fprintf(&b, "%s: %s\n", d.Level, d.Message)
			}
		}
		return fmt.Errorf("go backend failed for %s\n%s", filepath.Base(binOut), b.String())
	}
	cmd := exec.Command("go", "build", "-o", binOut, goOut)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go build %s: %w\n%s", filepath.Base(binOut), err, out)
	}
	return nil
}

func compileLLVM(files []string, binOut string, opts llvmgen.CompileOpts) error {
	cleanup := func() {}
	if opts.IntermediatesDir == "" {
		dir, err := os.MkdirTemp("", "kylix-mid-")
		if err != nil {
			return err
		}
		opts.IntermediatesDir = dir
		cleanup = func() { os.RemoveAll(dir) }
	}
	defer cleanup()

	paths, err := llvmgen.FindLLVM()
	if err != nil {
		return fmt.Errorf("LLVM toolchain not found: %w", err)
	}
	if dir := filepath.Dir(binOut); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	if _, err := llvmgen.CompileFilesToNative(files, binOut, paths, opts); err != nil {
		return err
	}
	return nil
}
