package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPackageSearchDirs verifies that CompileProject automatically discovers
// and includes .klx files from directories listed in opts.PackageSearchDirs.
func TestPackageSearchDirs(t *testing.T) {
	// Setup: create temp project with packages/ subdirectory
	tmpDir := t.TempDir()
	pkgDir := filepath.Join(tmpDir, "packages", "mypkg")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write a unit file in packages/mypkg/
	unitSrc := `unit mypkg;
interface
function Hello(): String;
implementation
function Hello(): String;
begin
  result := 'from package';
end;
end.`
	unitFile := filepath.Join(pkgDir, "mypkg.klx")
	if err := os.WriteFile(unitFile, []byte(unitSrc), 0644); err != nil {
		t.Fatal(err)
	}

	// Write main program that uses mypkg
	mainSrc := `program Main;
uses mypkg;
begin
  WriteLn(Hello());
end.`
	mainFile := filepath.Join(tmpDir, "main.klx")
	if err := os.WriteFile(mainFile, []byte(mainSrc), 0644); err != nil {
		t.Fatal(err)
	}

	// Compile with PackageSearchDirs pointing to packages/mypkg/
	opts := Options{
		PackageSearchDirs: []string{pkgDir},
		OutputFile:        filepath.Join(tmpDir, "main.go"),
	}
	result, err := CompileProject([]string{mainFile}, opts)
	if err != nil {
		t.Fatalf("CompileProject failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("compilation failed: %+v", result.Diagnostics)
	}

	// Verify the generated Go code includes the Hello function from the package
	if !strings.Contains(result.GoCode, "func Hello()") {
		t.Errorf("generated code missing Hello() function from package:\n%s", result.GoCode)
	}
}

// TestPackageSearchDirsDedup verifies that files explicitly passed to CompileProject
// are not re-added when they also appear in PackageSearchDirs.
func TestPackageSearchDirsDedup(t *testing.T) {
	tmpDir := t.TempDir()
	pkgDir := filepath.Join(tmpDir, "packages", "foo")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatal(err)
	}

	unitSrc := `unit foo;
interface
function Bar(): Integer;
implementation
function Bar(): Integer;
begin
  result := 42;
end;
end.`
	unitFile := filepath.Join(pkgDir, "foo.klx")
	if err := os.WriteFile(unitFile, []byte(unitSrc), 0644); err != nil {
		t.Fatal(err)
	}

	mainSrc := `program Main;
uses foo;
begin
  WriteLn(Bar());
end.`
	mainFile := filepath.Join(tmpDir, "main.klx")
	if err := os.WriteFile(mainFile, []byte(mainSrc), 0644); err != nil {
		t.Fatal(err)
	}

	// Pass unitFile explicitly AND via PackageSearchDirs — should deduplicate
	opts := Options{
		PackageSearchDirs: []string{pkgDir},
		OutputFile:        filepath.Join(tmpDir, "main.go"),
	}
	result, err := CompileProject([]string{mainFile, unitFile}, opts)
	if err != nil {
		t.Fatalf("CompileProject failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("compilation failed: %+v", result.Diagnostics)
	}

	// Count occurrences of "func Bar()" — should appear exactly once
	count := strings.Count(result.GoCode, "func Bar()")
	if count != 1 {
		t.Errorf("expected Bar() to appear once, got %d times:\n%s", count, result.GoCode)
	}
}
