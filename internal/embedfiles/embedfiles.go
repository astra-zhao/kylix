// Package embedfiles walks the directories named by a file-level
// [Embed('dir', …)] attribute (v0.12.0) and returns their contents for baking.
//
// Both backends use it, so the Go and LLVM forms bake exactly the same files
// under exactly the same names — the lookup key is the path as the program asks
// for it ("views/base.tpl"), slash-separated on every host OS.
package embedfiles

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// File is one baked file.
type File struct {
	Name    string // path as the program asks for it, slash-separated
	Content string
}

// Read collects every file under dirs, sorted by name so a build is
// reproducible. It errors when a directory is missing — pkg/compiler turns that
// into a diagnostic before codegen runs.
func Read(dirs []string) ([]File, error) {
	var files []File
	seen := map[string]bool{}
	for _, dir := range dirs {
		if seen[dir] {
			continue
		}
		seen[dir] = true
		info, err := os.Stat(dir)
		if err != nil {
			return nil, fmt.Errorf("[Embed] cannot read %s: %w", dir, err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("[Embed] %s is not a directory", dir)
		}
		err = filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if fi.IsDir() {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			files = append(files, File{Name: filepath.ToSlash(path), Content: string(data)})
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })
	return files, nil
}
