package llvmgen

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestPortableHash_MatchesKnownVectors(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		cc, err = exec.LookPath("clang")
	}
	if err != nil {
		t.Skip("no C compiler")
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "hash.c")
	if err := os.WriteFile(src, []byte(portableHashC), 0644); err != nil {
		t.Fatal(err)
	}
	mainSrc := filepath.Join(dir, "main.c")
	prog := `#include <stdio.h>
#include <string.h>
extern void kylix_sha256(const void *, unsigned long, unsigned char *);
extern void kylix_md5(const void *, unsigned long, unsigned char *);
static int eq(const unsigned char *p, const char *hex, int n) {
  for (int i = 0; i < n; i++) {
    unsigned int b;
    if (sscanf(hex + 2*i, "%02x", &b) != 1 || p[i] != (unsigned char)b) return 0;
  }
  return 1;
}
int main(void) {
  unsigned char o[32];
  kylix_sha256("", 0, o);
  if (!eq(o, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", 32)) return 1;
  kylix_sha256("abc", 3, o);
  if (!eq(o, "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad", 32)) return 2;
  unsigned char m[16];
  kylix_md5("abc", 3, m);
  if (!eq(m, "900150983cd24fb0d6963f7d28e17f72", 16)) return 3;
  kylix_md5("", 0, m);
  if (!eq(m, "d41d8cd98f00b204e9800998ecf8427e", 16)) return 4;
  return 0;
}
`
	if err := os.WriteFile(mainSrc, []byte(prog), 0644); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(dir, "hash_check")
	cmd := exec.Command(cc, "-O2", "-o", bin, mainSrc, src)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile: %v\n%s", err, out)
	}
	if out, err := exec.Command(bin).CombinedOutput(); err != nil {
		t.Fatalf("vectors: %v\n%s", err, out)
	}
}
