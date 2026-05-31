package stdlib

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Environment represents the application environment
type Environment string

const (
	EnvDevelopment Environment = "development"
	EnvTesting     Environment = "testing"
	EnvStaging     Environment = "staging"
	EnvProduction  Environment = "production"
)

// AutoConfig provides automatic configuration loading from multiple sources
type AutoConfig struct {
	data        map[string]interface{}
	environment Environment
	appName     string
	configDir   string
	prefix      string
	sources     []ConfigSource
	loaded      bool
}

// ConfigSource represents a configuration source
type ConfigSource struct {
	Name     string
	Type     string // "file", "env", "flag"
	Priority int    // Higher priority overrides lower
	Path     string // File path or env prefix
}

// NewAutoConfig creates a new auto-configuration instance
func NewAutoConfig(appName string) *AutoConfig {
	return &AutoConfig{
		data:        make(map[string]interface{}),
		environment: EnvDevelopment,
		appName:     appName,
		configDir:   ".",
		prefix:      strings.ToUpper(appName),
		sources:     make([]ConfigSource, 0),
	}
}

// SetEnvironment sets the application environment
func (ac *AutoConfig) SetEnvironment(env Environment) *AutoConfig {
	ac.environment = env
	return ac
}

// DetectEnvironment detects the environment from APP_ENV or environment variable
func (ac *AutoConfig) DetectEnvironment() *AutoConfig {
	if env := os.Getenv("APP_ENV"); env != "" {
		switch strings.ToLower(env) {
		case "development", "dev":
			ac.environment = EnvDevelopment
		case "testing", "test":
			ac.environment = EnvTesting
		case "staging", "stage":
			ac.environment = EnvStaging
		case "production", "prod":
			ac.environment = EnvProduction
		}
	}
	return ac
}

// SetConfigDir sets the configuration directory
func (ac *AutoConfig) SetConfigDir(dir string) *AutoConfig {
	ac.configDir = dir
	return ac
}

// SetPrefix sets the environment variable prefix
func (ac *AutoConfig) SetPrefix(prefix string) *AutoConfig {
	ac.prefix = strings.ToUpper(prefix)
	return ac
}

// AddSource adds a configuration source
func (ac *AutoConfig) AddSource(name, sourceType string, priority int, path string) *AutoConfig {
	ac.sources = append(ac.sources, ConfigSource{
		Name:     name,
		Type:     sourceType,
		Priority: priority,
		Path:     path,
	})
	return ac
}

// AddDefaultSources adds default configuration sources
// Priority order (lowest to highest): defaults -> config file -> env-specific file -> environment variables
func (ac *AutoConfig) AddDefaultSources() *AutoConfig {
	// Base config file
	ac.AddSource("config.json", "file", 10, filepath.Join(ac.configDir, "config.json"))

	// Environment-specific config file
	envFile := fmt.Sprintf("config.%s.json", ac.environment)
	ac.AddSource(envFile, "file", 20, filepath.Join(ac.configDir, envFile))

	// Environment variables
	ac.AddSource("env", "env", 30, ac.prefix)

	return ac
}

// Load loads configuration from all sources
func (ac *AutoConfig) Load() error {
	// Sort sources by priority (lowest first, so higher priority overrides)
	for i := 0; i < len(ac.sources); i++ {
		for j := i + 1; j < len(ac.sources); j++ {
			if ac.sources[i].Priority > ac.sources[j].Priority {
				ac.sources[i], ac.sources[j] = ac.sources[j], ac.sources[i]
			}
		}
	}

	// Load each source
	for _, source := range ac.sources {
		switch source.Type {
		case "file":
			if err := ac.loadFile(source.Path); err != nil {
				// File not found is okay, other errors are not
				if !os.IsNotExist(err) {
					return fmt.Errorf("failed to load %s: %w", source.Name, err)
				}
			}
		case "env":
			ac.loadEnvironment(source.Path)
		}
	}

	ac.loaded = true
	return nil
}

// loadFile loads configuration from a JSON file
func (ac *AutoConfig) loadFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Merge data (higher priority overrides)
	ac.mergeData(ac.data, data)

	return nil
}

// loadEnvironment loads configuration from environment variables
func (ac *AutoConfig) loadEnvironment(prefix string) {
	prefix = strings.ToUpper(prefix) + "_"

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		if !strings.HasPrefix(key, prefix) {
			continue
		}

		// Remove prefix and convert to lowercase with dots
		configKey := strings.TrimPrefix(key, prefix)
		configKey = strings.ToLower(configKey)
		configKey = strings.ReplaceAll(configKey, "_", ".")

		// Parse value to appropriate type
		parsedValue := ac.parseValue(value)

		// Set in data map
		ac.setNestedValue(configKey, parsedValue)
	}
}

// parseValue parses a string value to the appropriate type
func (ac *AutoConfig) parseValue(value string) interface{} {
	// Try boolean
	if strings.ToLower(value) == "true" {
		return true
	}
	if strings.ToLower(value) == "false" {
		return false
	}

	// Try integer
	if i, err := strconv.ParseInt(value, 10, 64); err == nil {
		return i
	}

	// Try float
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return f
	}

	// Try duration (e.g., "5s", "10m", "1h")
	if d, err := time.ParseDuration(value); err == nil {
		return d.String()
	}

	// Default to string
	return value
}

// mergeData merges source data into target
func (ac *AutoConfig) mergeData(target, source map[string]interface{}) {
	for key, value := range source {
		if sourceMap, ok := value.(map[string]interface{}); ok {
			if targetMap, ok := target[key].(map[string]interface{}); ok {
				ac.mergeData(targetMap, sourceMap)
				continue
			}
		}
		target[key] = value
	}
}

// setNestedValue sets a nested value using dot notation (e.g., "database.host")
func (ac *AutoConfig) setNestedValue(key string, value interface{}) {
	parts := strings.Split(key, ".")
	current := ac.data

	for i, part := range parts {
		if i == len(parts)-1 {
			current[part] = value
			return
		}

		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			next := make(map[string]interface{})
			current[part] = next
			current = next
		}
	}
}

// getNestedValue gets a nested value using dot notation
func (ac *AutoConfig) getNestedValue(key string) (interface{}, bool) {
	parts := strings.Split(key, ".")
	current := ac.data

	for i, part := range parts {
		if i == len(parts)-1 {
			val, ok := current[part]
			return val, ok
		}

		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return nil, false
		}
	}

	return nil, false
}

// Get retrieves a configuration value
func (ac *AutoConfig) Get(key string) interface{} {
	val, _ := ac.getNestedValue(key)
	return val
}

// GetString retrieves a string configuration value
func (ac *AutoConfig) GetString(key string) string {
	val := ac.Get(key)
	if val == nil {
		return ""
	}
	return fmt.Sprintf("%v", val)
}

// GetStringDefault retrieves a string with a default value
func (ac *AutoConfig) GetStringDefault(key, defaultValue string) string {
	val := ac.GetString(key)
	if val == "" {
		return defaultValue
	}
	return val
}

// GetInt retrieves an integer configuration value
func (ac *AutoConfig) GetInt(key string) int {
	val := ac.Get(key)
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		i, _ := strconv.Atoi(v)
		return i
	default:
		return 0
	}
}

// GetIntDefault retrieves an integer with a default value
func (ac *AutoConfig) GetIntDefault(key string, defaultValue int) int {
	val := ac.GetInt(key)
	if val == 0 {
		return defaultValue
	}
	return val
}

// GetBool retrieves a boolean configuration value
func (ac *AutoConfig) GetBool(key string) bool {
	val := ac.Get(key)
	if val == nil {
		return false
	}
	switch v := val.(type) {
	case bool:
		return v
	case string:
		return strings.ToLower(v) == "true"
	default:
		return false
	}
}

// GetBoolDefault retrieves a boolean with a default value
func (ac *AutoConfig) GetBoolDefault(key string, defaultValue bool) bool {
	val := ac.Get(key)
	if val == nil {
		return defaultValue
	}
	return ac.GetBool(key)
}

// GetFloat retrieves a float configuration value
func (ac *AutoConfig) GetFloat(key string) float64 {
	val := ac.Get(key)
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	default:
		return 0
	}
}

// GetFloatDefault retrieves a float with a default value
func (ac *AutoConfig) GetFloatDefault(key string, defaultValue float64) float64 {
	val := ac.GetFloat(key)
	if val == 0 {
		return defaultValue
	}
	return val
}

// GetDuration retrieves a duration configuration value
func (ac *AutoConfig) GetDuration(key string) time.Duration {
	val := ac.GetString(key)
	if val == "" {
		return 0
	}
	d, _ := time.ParseDuration(val)
	return d
}

// GetDurationDefault retrieves a duration with a default value
func (ac *AutoConfig) GetDurationDefault(key string, defaultValue time.Duration) time.Duration {
	d := ac.GetDuration(key)
	if d == 0 {
		return defaultValue
	}
	return d
}

// GetStringSlice retrieves a string slice configuration value
func (ac *AutoConfig) GetStringSlice(key string) []string {
	val := ac.Get(key)
	if val == nil {
		return nil
	}

	switch v := val.(type) {
	case []interface{}:
		result := make([]string, len(v))
		for i, item := range v {
			result[i] = fmt.Sprintf("%v", item)
		}
		return result
	case []string:
		return v
	case string:
		// Try comma-separated
		return strings.Split(v, ",")
	default:
		return nil
	}
}

// Has checks if a configuration key exists
func (ac *AutoConfig) Has(key string) bool {
	_, ok := ac.getNestedValue(key)
	return ok
}

// Set sets a configuration value
func (ac *AutoConfig) Set(key string, value interface{}) {
	ac.setNestedValue(key, value)
}

// GetEnvironment returns the current environment
func (ac *AutoConfig) GetEnvironment() Environment {
	return ac.environment
}

// IsDevelopment returns true if in development environment
func (ac *AutoConfig) IsDevelopment() bool {
	return ac.environment == EnvDevelopment
}

// IsTesting returns true if in testing environment
func (ac *AutoConfig) IsTesting() bool {
	return ac.environment == EnvTesting
}

// IsStaging returns true if in staging environment
func (ac *AutoConfig) IsStaging() bool {
	return ac.environment == EnvStaging
}

// IsProduction returns true if in production environment
func (ac *AutoConfig) IsProduction() bool {
	return ac.environment == EnvProduction
}

// All returns all configuration data
func (ac *AutoConfig) All() map[string]interface{} {
	return ac.data
}

// ToJSON returns the configuration as JSON
func (ac *AutoConfig) ToJSON() (string, error) {
	data, err := json.MarshalIndent(ac.data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Validate validates required configuration keys
func (ac *AutoConfig) Validate(requiredKeys []string) error {
	missing := make([]string, 0)
	for _, key := range requiredKeys {
		if !ac.Has(key) {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required configuration keys: %s", strings.Join(missing, ", "))
	}

	return nil
}

// AppConfig is a convenience wrapper that provides common application configuration
type AppConfig struct {
	config *AutoConfig
}

// NewAppConfig creates a new application configuration
func NewAppConfig(appName string) *AppConfig {
	return &AppConfig{
		config: NewAutoConfig(appName),
	}
}

// Load loads the application configuration
func (ac *AppConfig) Load() error {
	ac.config.DetectEnvironment()
	ac.config.AddDefaultSources()
	return ac.config.Load()
}

// Get retrieves a configuration value
func (ac *AppConfig) Get(key string) interface{} {
	return ac.config.Get(key)
}

// GetString retrieves a string value
func (ac *AppConfig) GetString(key string) string {
	return ac.config.GetString(key)
}

// GetInt retrieves an integer value
func (ac *AppConfig) GetInt(key string) int {
	return ac.config.GetInt(key)
}

// GetBool retrieves a boolean value
func (ac *AppConfig) GetBool(key string) bool {
	return ac.config.GetBool(key)
}

// IsDevelopment returns true if in development environment
func (ac *AppConfig) IsDevelopment() bool {
	return ac.config.IsDevelopment()
}

// IsProduction returns true if in production environment
func (ac *AppConfig) IsProduction() bool {
	return ac.config.IsProduction()
}

// GetDatabaseConfig returns database configuration
func (ac *AppConfig) GetDatabaseConfig() map[string]interface{} {
	if db := ac.config.Get("database"); db != nil {
		if dbMap, ok := db.(map[string]interface{}); ok {
			return dbMap
		}
	}
	return make(map[string]interface{})
}

// GetServerConfig returns server configuration
func (ac *AppConfig) GetServerConfig() map[string]interface{} {
	if server := ac.config.Get("server"); server != nil {
		if serverMap, ok := server.(map[string]interface{}); ok {
			return serverMap
		}
	}
	return make(map[string]interface{})
}

// GetLogConfig returns logging configuration
func (ac *AppConfig) GetLogConfig() map[string]interface{} {
	if log := ac.config.Get("log"); log != nil {
		if logMap, ok := log.(map[string]interface{}); ok {
			return logMap
		}
	}
	return make(map[string]interface{})
}
