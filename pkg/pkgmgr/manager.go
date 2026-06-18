// manager.go — Kylix package manager.
//
// Packages are git repositories containing .klx unit files.
// They are installed into <projectDir>/packages/<name>/.
//
// kylix.toml [dependencies] format:
//
//	http    = "github.com/kylix-lang/http@v0.1.0"
//	myutils = "github.com/alice/myutils"          (latest)
//	local   = "./local_pkg"                        (local path)
package pkgmgr

import (
	"fmt"
	"io/fs"
	"kylix/pkg/project"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const PackagesDir = "packages"

// Package describes an installed package.
type Package struct {
	Name  string
	Ref   string   // original ref from kylix.toml
	Dir   string   // local install directory
	Units []string // .klx unit file paths
}

// Manager handles package installation for a project.
type Manager struct {
	cfg     *project.Config
	pkgsDir string
}

// New returns a Manager for the given project config.
func New(cfg *project.Config) *Manager {
	return &Manager{
		cfg:     cfg,
		pkgsDir: filepath.Join(cfg.ProjectDir(), PackagesDir),
	}
}

// Add adds a package to kylix.toml and installs it.
//
//	ref examples: "github.com/kylix-lang/http@v0.1.0"  "./local/path"
func (m *Manager) Add(name, ref string) error {
	if m.cfg.Dependencies == nil {
		m.cfg.Dependencies = make(map[string]string)
	}
	m.cfg.Dependencies[name] = ref

	if err := m.Install(name, ref); err != nil {
		return err
	}

	return m.cfg.Save(m.cfg.Path)
}

// Install installs a single package by ref.
func (m *Manager) Install(name, ref string) error {
	if err := os.MkdirAll(m.pkgsDir, 0755); err != nil {
		return err
	}

	destDir := filepath.Join(m.pkgsDir, name)

	// Local path
	if strings.HasPrefix(ref, ".") || strings.HasPrefix(ref, "/") {
		return m.installLocal(name, ref, destDir)
	}

	// Git URL with optional @version tag
	return m.installGit(name, ref, destDir)
}

// InstallAll installs all dependencies listed in kylix.toml.
func (m *Manager) InstallAll() error {
	for name, ref := range m.cfg.Dependencies {
		fmt.Printf("  installing %s (%s)…\n", name, ref)
		if err := m.Install(name, ref); err != nil {
			return fmt.Errorf("failed to install %s: %v", name, err)
		}
	}
	return nil
}

// Remove removes a package from kylix.toml and deletes its directory.
func (m *Manager) Remove(name string) error {
	delete(m.cfg.Dependencies, name)
	dir := filepath.Join(m.pkgsDir, name)
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	return m.cfg.Save(m.cfg.Path)
}

// List returns all installed packages.
func (m *Manager) List() ([]*Package, error) {
	var pkgs []*Package
	for name, ref := range m.cfg.Dependencies {
		dir := filepath.Join(m.pkgsDir, name)
		units, _ := findUnits(dir)
		pkgs = append(pkgs, &Package{
			Name:  name,
			Ref:   ref,
			Dir:   dir,
			Units: units,
		})
	}
	return pkgs, nil
}

// PackageDirs returns all package .klx directories, for use in compilation.
func (m *Manager) PackageDirs() []string {
	var dirs []string
	entries, err := os.ReadDir(m.pkgsDir)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		p := filepath.Join(m.pkgsDir, e.Name())
		// Follow symlinks: os.ReadDir DirEntry.IsDir() is false for symlinks.
		info, err := os.Stat(p)
		if err == nil && info.IsDir() {
			dirs = append(dirs, p)
		}
	}
	return dirs
}

// ── internal helpers ──────────────────────────────────────────────────────────

func (m *Manager) installLocal(name, ref, destDir string) error {
	var abs string
	if filepath.IsAbs(ref) {
		abs = ref
	} else {
		abs, _ = filepath.Abs(filepath.Join(m.cfg.ProjectDir(), ref))
	}
	if _, err := os.Stat(abs); err != nil {
		return fmt.Errorf("local package path %q not found", abs)
	}
	// Symlink for local dev convenience.
	os.Remove(destDir)
	return os.Symlink(abs, destDir)
}

func (m *Manager) installGit(name, ref, destDir string) error {
	// Parse ref: "github.com/user/repo@tag" or just URL
	url, tag := splitRef(ref)
	if !strings.Contains(url, "://") {
		url = "https://" + url
	}

	// If already installed and version is pinned, skip (idempotent).
	// Without a tag we always pull to get the latest.
	if _, err := os.Stat(destDir); err == nil && tag != "" {
		return nil
	}

	// Remove stale install.
	os.RemoveAll(destDir)

	args := []string{"clone", "--depth=1"}
	if tag != "" {
		args = append(args, "--branch", tag)
	}
	args = append(args, url, destDir)

	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone %s failed: %v", url, err)
	}
	return nil
}

func splitRef(ref string) (url, tag string) {
	if idx := strings.LastIndex(ref, "@"); idx > 0 {
		return ref[:idx], ref[idx+1:]
	}
	return ref, ""
}

func findUnits(dir string) ([]string, error) {
	var units []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if strings.HasSuffix(path, ".klx") {
			units = append(units, path)
		}
		return nil
	})
	return units, err
}
