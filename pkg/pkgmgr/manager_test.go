package pkgmgr_test

import (
	"os"
	"path/filepath"
	"testing"

	"kylix/pkg/pkgmgr"
	"kylix/pkg/project"
)

// makeTestProject creates a temp dir with a minimal kylix.toml and returns the config.
func makeTestProject(t *testing.T) *project.Config {
	t.Helper()
	dir := t.TempDir()
	cfg := &project.Config{
		Name:         "testapp",
		Version:      "1.0.0",
		Main:         "main.klx",
		Output:       "build/",
		GoMod:        "testapp",
		Dependencies: make(map[string]string),
		Path:         filepath.Join(dir, "kylix.toml"),
	}
	if err := cfg.Save(cfg.Path); err != nil {
		t.Fatal(err)
	}
	return cfg
}

// makeLocalPkg creates a temp dir with a .klx unit file and returns its path.
func makeLocalPkg(t *testing.T, unitName string) string {
	t.Helper()
	dir := t.TempDir()
	src := "unit " + unitName + ";\nfunction Hello(): String;\nbegin result := 'hi'; end;\n"
	if err := os.WriteFile(filepath.Join(dir, unitName+".klx"), []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestAdd_LocalPath(t *testing.T) {
	cfg := makeTestProject(t)
	pkgDir := makeLocalPkg(t, "utils")

	mgr := pkgmgr.New(cfg)
	if err := mgr.Add("utils", pkgDir); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// kylix.toml must have the dependency
	if cfg.Dependencies["utils"] != pkgDir {
		t.Errorf("expected dependency utils=%q, got %q", pkgDir, cfg.Dependencies["utils"])
	}

	// packages/utils must be a symlink pointing to pkgDir
	linkPath := filepath.Join(cfg.ProjectDir(), "packages", "utils")
	target, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("symlink not created: %v", err)
	}
	if target != pkgDir {
		t.Errorf("symlink target = %q, want %q", target, pkgDir)
	}
}

func TestRemove(t *testing.T) {
	cfg := makeTestProject(t)
	pkgDir := makeLocalPkg(t, "utils")

	mgr := pkgmgr.New(cfg)
	if err := mgr.Add("utils", pkgDir); err != nil {
		t.Fatal(err)
	}

	if err := mgr.Remove("utils"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if _, ok := cfg.Dependencies["utils"]; ok {
		t.Error("dependency still present after Remove")
	}

	linkPath := filepath.Join(cfg.ProjectDir(), "packages", "utils")
	if _, err := os.Lstat(linkPath); err == nil {
		t.Error("symlink still exists after Remove")
	}
}

func TestList(t *testing.T) {
	cfg := makeTestProject(t)
	pkgDir := makeLocalPkg(t, "mylib")

	mgr := pkgmgr.New(cfg)
	if err := mgr.Add("mylib", pkgDir); err != nil {
		t.Fatal(err)
	}

	pkgs, err := mgr.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].Name != "mylib" {
		t.Errorf("expected name=mylib, got %q", pkgs[0].Name)
	}
}

func TestPackageDirs(t *testing.T) {
	cfg := makeTestProject(t)
	pkgDir := makeLocalPkg(t, "alpha")

	mgr := pkgmgr.New(cfg)
	if err := mgr.Add("alpha", pkgDir); err != nil {
		t.Fatal(err)
	}

	dirs := mgr.PackageDirs()
	if len(dirs) != 1 {
		t.Fatalf("expected 1 package dir, got %d", len(dirs))
	}
}

func TestInstallAll(t *testing.T) {
	cfg := makeTestProject(t)
	a := makeLocalPkg(t, "pkga")
	b := makeLocalPkg(t, "pkgb")
	cfg.Dependencies["pkga"] = a
	cfg.Dependencies["pkgb"] = b

	mgr := pkgmgr.New(cfg)
	if err := mgr.InstallAll(); err != nil {
		t.Fatalf("InstallAll failed: %v", err)
	}

	for _, name := range []string{"pkga", "pkgb"} {
		link := filepath.Join(cfg.ProjectDir(), "packages", name)
		if _, err := os.Lstat(link); err != nil {
			t.Errorf("package %s not installed: %v", name, err)
		}
	}
}
