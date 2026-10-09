package llvmgen

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed portable/hash.c
var portableHashC string

// compilePortableHash writes the embedded SHA-256/MD5 source and compiles it
// with the same clang, target, and sysroot the link is about to use.
// The returned directory must be removed after the link.
func compilePortableHash(clang string, linkArgs []string) (obj string, dir string, err error) {
	dir, err = os.MkdirTemp("", "kylix-hash-")
	if err != nil {
		return "", "", err
	}
	src := filepath.Join(dir, "hash.c")
	if err := os.WriteFile(src, []byte(portableHashC), 0644); err != nil {
		os.RemoveAll(dir)
		return "", "", err
	}
	obj = filepath.Join(dir, "hash.o")
	if err := compileSupportC(clang, linkArgs, src, obj, nil); err != nil {
		os.RemoveAll(dir)
		return "", "", err
	}
	return obj, dir, nil
}

// compileSupportC compiles one C file to obj using the target flags already
// chosen for the link (-fPIC so the object can land in a shared library).
func compileSupportC(clang string, linkArgs []string, src, obj string, defines []string) error {
	args := []string{"-c", "-O2", "-fPIC", "-o", obj}
	args = append(args, defines...)
	args = append(args, targetCompileArgs(linkArgs)...)
	args = append(args, src)
	cmd := exec.Command(clang, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("compile %s: %w\n%s", filepath.Base(src), err, out)
	}
	return nil
}

// targetCompileArgs copies the cross-compile flags a C support file needs
// from the clang link line. -arch is a pair; the sysroot flags are single
// arguments or a flag plus a path.
func targetCompileArgs(linkArgs []string) []string {
	var out []string
	for i := 0; i < len(linkArgs); i++ {
		a := linkArgs[i]
		switch {
		case a == "-arch" && i+1 < len(linkArgs):
			out = append(out, a, linkArgs[i+1])
			i++
		case a == "-isysroot" && i+1 < len(linkArgs):
			out = append(out, a, linkArgs[i+1])
			i++
		case strings.HasPrefix(a, "-isysroot"),
			strings.HasPrefix(a, "--sysroot"),
			strings.HasPrefix(a, "--target="):
			out = append(out, a)
		}
	}
	return out
}
