// config.go — Configuration management for KylixBoot.
//
// Sources (in precedence order):
//   1. Programmatic Set() calls
//   2. Environment variables (KEY → app.key, KYLIX_X_Y → x.y)
//   3. Defaults
package boot

import (
	"os"
	"strconv"
	"strings"
	"sync"
)

// Config stores key/value application settings.
type Config struct {
	mu     sync.RWMutex
	values map[string]interface{}
}

// NewConfig creates an empty config (env vars merged on first access).
func NewConfig() *Config {
	return &Config{values: map[string]interface{}{}}
}

// Set explicitly stores a value (highest precedence).
func (c *Config) Set(key string, value interface{}) {
	c.mu.Lock()
	c.values[key] = value
	c.mu.Unlock()
}

// Get returns (value, true) if set, otherwise checks env vars.
func (c *Config) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	v, ok := c.values[key]
	c.mu.RUnlock()
	if ok {
		return v, true
	}
	// Translate "app.port" → APP_PORT env var.
	envKey := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
	if env := os.Getenv(envKey); env != "" {
		return env, true
	}
	return nil, false
}

// StringDefault returns the value as string, or the fallback.
func (c *Config) StringDefault(key, fallback string) string {
	v, ok := c.Get(key)
	if !ok {
		return fallback
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmtAny(v)
}

// IntDefault returns the value as int, or the fallback.
func (c *Config) IntDefault(key string, fallback int) int {
	v, ok := c.Get(key)
	if !ok {
		return fallback
	}
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	case string:
		if n, err := strconv.Atoi(x); err == nil {
			return n
		}
	}
	return fallback
}

// BoolDefault returns the value as bool, or the fallback.
func (c *Config) BoolDefault(key string, fallback bool) bool {
	v, ok := c.Get(key)
	if !ok {
		return fallback
	}
	switch x := v.(type) {
	case bool:
		return x
	case string:
		switch strings.ToLower(x) {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off":
			return false
		}
	}
	return fallback
}

func fmtAny(v interface{}) string {
	switch x := v.(type) {
	case string:
		return x
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case bool:
		if x {
			return "true"
		}
		return "false"
	}
	return ""
}
