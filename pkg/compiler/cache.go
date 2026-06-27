// cache.go — Incremental compilation cache.
//
// Each .klx file is identified by its absolute path. The cache entry stores
// the file's mtime + size (the "fingerprint") along with the parsed AST
// serialised as the generated Go source fragment.  On a subsequent build, if
// the fingerprint is unchanged the parse + generate step is skipped.
//
// Cache location: <projectRoot>/.kylix-cache/  (one JSON file per source file)
// The cache is keyed by the SHA-256 of the absolute source path so filenames
// with special characters are safe.
package compiler

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CacheVersion invalidates stale generated fragments after codegen changes.
const CacheVersion = 9

// CacheEntry holds the cached output for a single .klx file.
type CacheEntry struct {
	Version int `json:"version"`

	// Fingerprint fields — used to decide whether the cache is still valid.
	ModTime time.Time `json:"mod_time"`
	Size    int64     `json:"size"`
	SrcPath string    `json:"src_path"`

	// Cached output
	GoCode string `json:"go_code"`
}

// BuildCache manages incremental compilation state for a project.
type BuildCache struct {
	dir string // directory where cache files live
}

// NewBuildCache returns a cache that stores entries under <dir>/.kylix-cache/.
// It creates the directory if it does not exist.
func NewBuildCache(dir string) *BuildCache {
	cacheDir := filepath.Join(dir, ".kylix-cache")
	os.MkdirAll(cacheDir, 0755)
	return &BuildCache{dir: cacheDir}
}

// cacheFile returns the path of the JSON file for a given source path.
func (c *BuildCache) cacheFile(srcPath string) string {
	sum := sha256.Sum256([]byte(srcPath))
	return filepath.Join(c.dir, fmt.Sprintf("%x.json", sum))
}

// Load returns the cached entry for srcPath if it is still valid (fingerprint
// matches current file stat). Returns nil when the cache is cold or stale.
func (c *BuildCache) Load(srcPath string) *CacheEntry {
	info, err := os.Stat(srcPath)
	if err != nil {
		return nil
	}

	data, err := os.ReadFile(c.cacheFile(srcPath))
	if err != nil {
		return nil
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil
	}

	if entry.Version == CacheVersion && entry.ModTime.Equal(info.ModTime()) && entry.Size == info.Size() {
		return &entry
	}
	return nil
}

// Store writes a cache entry for srcPath with the given Go code.
func (c *BuildCache) Store(srcPath, goCode string) {
	info, err := os.Stat(srcPath)
	if err != nil {
		return
	}
	entry := CacheEntry{
		Version: CacheVersion,
		ModTime: info.ModTime(),
		Size:    info.Size(),
		SrcPath: srcPath,
		GoCode:  goCode,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}
	os.WriteFile(c.cacheFile(srcPath), data, 0644)
}

// Invalidate removes the cache entry for srcPath (e.g. after a semantic error).
func (c *BuildCache) Invalidate(srcPath string) {
	os.Remove(c.cacheFile(srcPath))
}
