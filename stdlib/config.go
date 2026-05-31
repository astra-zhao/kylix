package stdlib

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds configuration values
type Config struct {
	values   map[string]string
	prefix   string
	defaults map[string]string
}

// NewConfig creates a new configuration instance
func NewConfig() *Config {
	return &Config{
		values:   make(map[string]string),
		defaults: make(map[string]string),
	}
}

// SetPrefix sets the environment variable prefix
func (c *Config) SetPrefix(prefix string) {
	c.prefix = prefix
}

// LoadFromEnv loads configuration from environment variables
func (c *Config) LoadFromEnv() {
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		// Check if the key matches the prefix
		if c.prefix != "" {
			if !strings.HasPrefix(key, c.prefix+"_") {
				continue
			}
			// Remove prefix and convert to lowercase with dots
			key = strings.ToLower(strings.TrimPrefix(key, c.prefix+"_"))
			key = strings.ReplaceAll(key, "_", ".")
		} else {
			key = strings.ToLower(key)
		}

		c.values[key] = value
	}
}

// Set sets a configuration value
func (c *Config) Set(key string, value interface{}) {
	c.values[key] = fmt.Sprintf("%v", value)
}

// SetDefault sets a default value for a key
func (c *Config) SetDefault(key string, value interface{}) {
	c.defaults[key] = fmt.Sprintf("%v", value)
}

// Get retrieves a string value
func (c *Config) Get(key string) string {
	if value, exists := c.values[key]; exists {
		return value
	}
	if value, exists := c.defaults[key]; exists {
		return value
	}
	return ""
}

// GetString is an alias for Get
func (c *Config) GetString(key string) string {
	return c.Get(key)
}

// GetInt retrieves an integer value
func (c *Config) GetInt(key string) int {
	str := c.Get(key)
	if str == "" {
		return 0
	}
	value, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return value
}

// GetIntDefault retrieves an integer value with a default
func (c *Config) GetIntDefault(key string, defaultValue int) int {
	str := c.Get(key)
	if str == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(str)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetFloat retrieves a float64 value
func (c *Config) GetFloat(key string) float64 {
	str := c.Get(key)
	if str == "" {
		return 0.0
	}
	value, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0.0
	}
	return value
}

// GetFloatDefault retrieves a float64 value with a default
func (c *Config) GetFloatDefault(key string, defaultValue float64) float64 {
	str := c.Get(key)
	if str == "" {
		return defaultValue
	}
	value, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetBool retrieves a boolean value
func (c *Config) GetBool(key string) bool {
	str := strings.ToLower(c.Get(key))
	return str == "true" || str == "1" || str == "yes" || str == "on"
}

// GetBoolDefault retrieves a boolean value with a default
func (c *Config) GetBoolDefault(key string, defaultValue bool) bool {
	str := c.Get(key)
	if str == "" {
		return defaultValue
	}
	return c.GetBool(key)
}

// GetDuration retrieves a duration value
func (c *Config) GetDuration(key string) time.Duration {
	str := c.Get(key)
	if str == "" {
		return 0
	}
	duration, err := time.ParseDuration(str)
	if err != nil {
		return 0
	}
	return duration
}

// GetDurationDefault retrieves a duration value with a default
func (c *Config) GetDurationDefault(key string, defaultValue time.Duration) time.Duration {
	str := c.Get(key)
	if str == "" {
		return defaultValue
	}
	duration, err := time.ParseDuration(str)
	if err != nil {
		return defaultValue
	}
	return duration
}

// GetStringSlice retrieves a string slice value (comma-separated)
func (c *Config) GetStringSlice(key string) []string {
	str := c.Get(key)
	if str == "" {
		return []string{}
	}
	parts := strings.Split(str, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// Has checks if a key exists
func (c *Config) Has(key string) bool {
	_, exists := c.values[key]
	if !exists {
		_, exists = c.defaults[key]
	}
	return exists
}

// Delete removes a key
func (c *Config) Delete(key string) {
	delete(c.values, key)
}

// Clear removes all values (but not defaults)
func (c *Config) Clear() {
	c.values = make(map[string]string)
}

// Keys returns all configuration keys
func (c *Config) Keys() []string {
	keysMap := make(map[string]bool)
	for k := range c.values {
		keysMap[k] = true
	}
	for k := range c.defaults {
		keysMap[k] = true
	}

	keys := make([]string, 0, len(keysMap))
	for k := range keysMap {
		keys = append(keys, k)
	}
	return keys
}

// All returns all configuration values
func (c *Config) All() map[string]string {
	result := make(map[string]string)

	// Add defaults first
	for k, v := range c.defaults {
		result[k] = v
	}

	// Override with actual values
	for k, v := range c.values {
		result[k] = v
	}

	return result
}

// MustGet retrieves a string value or panics if not found
func (c *Config) MustGet(key string) string {
	value := c.Get(key)
	if value == "" {
		panic(fmt.Sprintf("configuration key '%s' is required but not set", key))
	}
	return value
}

// MustGetInt retrieves an integer value or panics if not found or invalid
func (c *Config) MustGetInt(key string) int {
	str := c.Get(key)
	if str == "" {
		panic(fmt.Sprintf("configuration key '%s' is required but not set", key))
	}
	value, err := strconv.Atoi(str)
	if err != nil {
		panic(fmt.Sprintf("configuration key '%s' must be an integer, got '%s'", key, str))
	}
	return value
}

// Load loads configuration from environment and applies defaults
func (c *Config) Load() {
	c.LoadFromEnv()
}
