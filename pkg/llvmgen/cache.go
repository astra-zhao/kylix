package llvmgen

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// cache.go — incremental compilation cache for the LLVM backend (v4.5.0 Phase C).
//
// Caches the compiled .o object file keyed on a hash of (source content +
// compile options + dependency unit signatures). On a cache hit, llc is
// skipped entirely and the cached .o is linked directly — a full recompile
// becomes a single clang link step.
//
// Cache key inputs (any change invalidates):
//   - Source file content (SHA256)
//   - OptLevel, DebugInfo (CompileOpts)
//   - Hash of each dependency unit file's content (for multi-file builds)
//
// Cache location: <CacheDir>/kylix-llvm-cache/<key>.o  (CacheDir defaults to
// the OS temp dir if unset). Stale entries (source changed) are simply
// overwritten; no expiry sweep is needed for correctness, though the cache
// dir may grow over time.

// CacheKey is the hash identifying a unique compilation input.
type CacheKey string

// ComputeCacheKey derives the cache key for a single-file build.
func ComputeCacheKey(srcFile string, opts CompileOpts, depHashes []string) (CacheKey, error) {
	h := sha256.New()
	src, err := os.ReadFile(srcFile)
	if err != nil {
		return "", fmt.Errorf("cache: read %s: %w", srcFile, err)
	}
	h.Write(src)
	// Options affect codegen output.
	fmt.Fprintf(h, "opt=%s|debug=%v", opts.OptLevel, opts.DebugInfo)
	// Dependency signatures (unit files merged in for multi-file builds).
	for _, dh := range depHashes {
		h.Write([]byte(dh))
	}
	return CacheKey(hex.EncodeToString(h.Sum(nil))), nil
}

// fileHash returns the hex SHA256 of a file's content (used for dep signatures).
func fileHash(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}

// DepHashes computes the content hashes of dependency unit files.
func DepHashes(paths []string) ([]string, error) {
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		h, err := fileHash(p)
		if err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, nil
}

// CacheStore resolves the on-disk object path for a cache key.
type CacheStore struct {
	dir string
}

// NewCacheStore creates a cache rooted at dir (created if missing).
// If dir is empty, defaults to os.TempDir()/kylix-llvm-cache.
func NewCacheStore(dir string) (*CacheStore, error) {
	if dir == "" {
		dir = filepath.Join(os.TempDir(), "kylix-llvm-cache")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("cache: mkdir %s: %w", dir, err)
	}
	return &CacheStore{dir: dir}, nil
}

// ObjectPath returns the cached .o path for a key (whether or not it exists).
func (c *CacheStore) ObjectPath(key CacheKey) string {
	return filepath.Join(c.dir, string(key)+".o")
}

// Get returns the cached object path if present, else "".
func (c *CacheStore) Get(key CacheKey) string {
	p := c.ObjectPath(key)
	if info, err := os.Stat(p); err == nil && !info.IsDir() {
		return p
	}
	return ""
}

// Put stores srcObj at the cache path for key (copies, leaving the original).
func (c *CacheStore) Put(key CacheKey, srcObj string) error {
	dst := c.ObjectPath(key)
	in, err := os.ReadFile(srcObj)
	if err != nil {
		return fmt.Errorf("cache: read src obj %s: %w", srcObj, err)
	}
	if err := os.WriteFile(dst, in, 0644); err != nil {
		return fmt.Errorf("cache: write %s: %w", dst, err)
	}
	return nil
}

// irCacheKey derives a cache key from the final IR text + compile options.
// Used when source-file-based keying isn't available (the AST is already
// parsed); the IR hash captures all codegen-relevant inputs.
func irCacheKey(ir string, opts CompileOpts) CacheKey {
	h := sha256.New()
	h.Write([]byte(ir))
	fmt.Fprintf(h, "opt=%s|debug=%v", opts.OptLevel, opts.DebugInfo)
	return CacheKey(hex.EncodeToString(h.Sum(nil)))
}

// defaultLLVMCache returns the process-wide cache store (lazily initialized).
// Returns nil if the cache dir cannot be created (cache disabled silently).
var defaultCacheStore *CacheStore

func defaultLLVMCache() *CacheStore {
	if defaultCacheStore != nil {
		return defaultCacheStore
	}
	store, err := NewCacheStore("")
	if err != nil {
		return nil
	}
	defaultCacheStore = store
	return store
}

// copyFile copies src to dst (used to materialize a cached .o at the path
// the clang link step expects).
func copyFile(src, dst string) error {
	in, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, in, 0644)
}
