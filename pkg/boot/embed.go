package boot

import "sync"

// embed.go — the embedded-file registry (v0.12.0).
//
// The compiler's file-level [Embed('views', 'static')] attribute bakes those
// directories into the binary; the generated init() registers each file here.
// Both the static handler and stdlib.ReadFile consult this registry before
// touching the filesystem, so a built binary runs with no views/ or static/
// directory beside it.
//
// It lives in boot rather than stdlib because stdlib imports boot (the reverse
// would be a cycle) and the static handler needs it too.
var (
	embedMu    sync.RWMutex
	embedFiles = map[string]string{}
)

// RegisterEmbedded stores one baked file. Called by generated code.
func RegisterEmbedded(name, content string) {
	embedMu.Lock()
	embedFiles[name] = content
	embedMu.Unlock()
}

// EmbeddedFile returns a baked file's contents, or ok=false when the program
// did not embed that path.
func EmbeddedFile(name string) (string, bool) {
	embedMu.RLock()
	c, ok := embedFiles[name]
	embedMu.RUnlock()
	return c, ok
}

// ResetEmbedded clears the registry (tests).
func ResetEmbedded() {
	embedMu.Lock()
	embedFiles = map[string]string{}
	embedMu.Unlock()
}
