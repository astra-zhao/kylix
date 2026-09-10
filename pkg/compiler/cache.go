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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CacheVersion invalidates stale generated fragments after codegen changes.
// (Kept as a hard reset for structural cache-format changes; day-to-day
// codegen invalidation is handled by the binary self-hash below.)
const CacheVersion = 12

// CacheEntry holds the cached output for a single .klx file.
type CacheEntry struct {
	Version int `json:"version"`

	// Fingerprint fields — used to decide whether the cache is still valid.
	ModTime     time.Time `json:"mod_time"`
	Size        int64     `json:"size"`
	SrcPath     string    `json:"src_path"`
	CodegenHash string    `json:"codegen_hash,omitempty"` // v0.7.1 P4: running compiler binary self-hash

	// Cached output
	GoCode string `json:"go_code"`
}

// BuildCache manages incremental compilation state for a project.
type BuildCache struct {
	dir string // directory where cache files live

	// Hit counters (v0.6.0) — populated by Load for reporting cache efficiency
	// (kylix build --time / benchmarks). Load is called sequentially from
	// CompileProject's pre-parse loop, so no locking is needed.
	Hits   int
	Misses int
}

// NewBuildCache returns a cache that stores entries under <dir>/.kylix-cache/.
// It creates the directory if it does not exist.
func NewBuildCache(dir string) *BuildCache {
	cacheDir := filepath.Join(dir, ".kylix-cache")
	os.MkdirAll(cacheDir, 0755)
	return &BuildCache{dir: cacheDir}
}

// codegenHashOverride, when non-empty, short-circuits codegenHash — a test
// hook to simulate a rebuilt compiler binary (internal tests only).
var codegenHashOverride string

// codegenHash returns a fingerprint of the running compiler binary (v0.7.1
// P4): any recompile of the compiler — emitter, parser, typecheck changes —
// changes the hash, so cached fragments produced by the old binary are
// invalidated automatically instead of requiring a manual .kylix-cache wipe.
// The result is process-global (the binary doesn't change under a running
// process) and short: 16 hex chars are plenty to distinguish compiler builds.
// On failure it returns "" — entries recorded under "" match "" so the cache
// degrades to the pre-P4 mtime+size behavior rather than never hitting.
var cachedCodegenHash string

func codegenHash() string {
	if codegenHashOverride != "" {
		return codegenHashOverride
	}
	if cachedCodegenHash != "" {
		return cachedCodegenHash
	}
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(exe)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	cachedCodegenHash = hex.EncodeToString(sum[:8])
	return cachedCodegenHash
}

// cacheFile returns the path of the JSON file for a given source path.
func (c *BuildCache) cacheFile(srcPath string) string {
	sum := sha256.Sum256([]byte(srcPath))
	return filepath.Join(c.dir, fmt.Sprintf("%x.json", sum))
}

// Load returns the cached entry for srcPath if it is still valid (fingerprint
// matches current file stat). Returns nil when the cache is cold or stale.
// Each call bumps the Hits/Misses counters (v0.6.0).
func (c *BuildCache) Load(srcPath string) *CacheEntry {
	info, err := os.Stat(srcPath)
	if err != nil {
		c.Misses++
		return nil
	}

	data, err := os.ReadFile(c.cacheFile(srcPath))
	if err != nil {
		c.Misses++
		return nil
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		c.Misses++
		return nil
	}

	if entry.Version == CacheVersion && entry.CodegenHash == codegenHash() &&
		entry.ModTime.Equal(info.ModTime()) && entry.Size == info.Size() {
		c.Hits++
		return &entry
	}
	c.Misses++
	return nil
}

// Store writes a cache entry for srcPath with the given Go code.
func (c *BuildCache) Store(srcPath, goCode string) {
	info, err := os.Stat(srcPath)
	if err != nil {
		return
	}
	entry := CacheEntry{
		Version:     CacheVersion,
		ModTime:     info.ModTime(),
		Size:        info.Size(),
		SrcPath:     srcPath,
		CodegenHash: codegenHash(),
		GoCode:      goCode,
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
