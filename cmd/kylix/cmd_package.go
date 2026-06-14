package main

import (
	"fmt"
	"kylix/pkg/pkgmgr"
	"kylix/pkg/project"
	"os"
	"path/filepath"
)

func cmdAdd(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: kylix add <name> [<repo@version>]")
		os.Exit(1)
	}
	name := args[0]
	ref := name
	if len(args) >= 2 {
		ref = args[1]
	}

	cfg, err := project.Find(".")
	if err != nil || cfg == nil {
		fmt.Fprintln(os.Stderr, "Error: no kylix.toml found in current directory tree")
		os.Exit(1)
	}

	mgr := pkgmgr.New(cfg)
	fmt.Printf("Adding %s (%s)…\n", name, ref)
	if err := mgr.Add(name, ref); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Added %s\n", name)
}

// cmdInstall: kylix install  — install all dependencies from kylix.toml
func cmdInstall(args []string) {
	cfg, err := project.Find(".")
	if err != nil || cfg == nil {
		fmt.Fprintln(os.Stderr, "Error: no kylix.toml found in current directory tree")
		os.Exit(1)
	}

	if len(cfg.Dependencies) == 0 {
		fmt.Println("No dependencies to install.")
		return
	}

	mgr := pkgmgr.New(cfg)
	fmt.Printf("Installing %d package(s)…\n", len(cfg.Dependencies))
	if err := mgr.InstallAll(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ All packages installed")
}

// cmdRemove: kylix remove <name>
func cmdRemove(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: kylix remove <name>")
		os.Exit(1)
	}
	name := args[0]

	cfg, err := project.Find(".")
	if err != nil || cfg == nil {
		fmt.Fprintln(os.Stderr, "Error: no kylix.toml found in current directory tree")
		os.Exit(1)
	}

	mgr := pkgmgr.New(cfg)
	if err := mgr.Remove(name); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Removed %s\n", name)
}

// packageDirsFromWd returns all subdirectory paths under <wd>/packages/
// for use as PackageSearchDirs in compiler.Options.
func packageDirsFromWd(wd string) []string {
	pkgsDir := filepath.Join(wd, "packages")
	entries, err := os.ReadDir(pkgsDir)
	if err != nil {
		return nil
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, filepath.Join(pkgsDir, e.Name()))
		}
	}
	return dirs
}
