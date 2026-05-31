package stdlib

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTemplateEngineRenderString(t *testing.T) {
	engine := NewTemplateEngine()

	// Test basic variable substitution
	result, err := engine.RenderString("Hello, {{.Name}}!", map[string]interface{}{
		"Name": "World",
	})
	if err != nil {
		t.Fatalf("RenderString failed: %v", err)
	}
	if result != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", result)
	}

	// Test with no data
	result, err = engine.RenderString("Static content", nil)
	if err != nil {
		t.Fatalf("RenderString failed: %v", err)
	}
	if result != "Static content" {
		t.Errorf("Expected 'Static content', got '%s'", result)
	}

	// Test with loop
	result, err = engine.RenderString("{{range .Items}}{{.}} {{end}}", map[string]interface{}{
		"Items": []string{"a", "b", "c"},
	})
	if err != nil {
		t.Fatalf("RenderString failed: %v", err)
	}
	expected := "a b c "
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Test with condition
	result, err = engine.RenderString("{{if .Show}}visible{{else}}hidden{{end}}", map[string]interface{}{
		"Show": true,
	})
	if err != nil {
		t.Fatalf("RenderString failed: %v", err)
	}
	if result != "visible" {
		t.Errorf("Expected 'visible', got '%s'", result)
	}

	// Test with custom function
	engine.AddFunc("greet", func(name string) string {
		return "Hello, " + name + "!"
	})
	result, err = engine.RenderString(`{{greet "Alice"}}`, nil)
	if err != nil {
		t.Fatalf("RenderString failed: %v", err)
	}
	if result != "Hello, Alice!" {
		t.Errorf("Expected 'Hello, Alice!', got '%s'", result)
	}
}

func TestTemplateEngineRenderFile(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create a template file
	tmplPath := filepath.Join(tmpDir, "test.html")
	err := os.WriteFile(tmplPath, []byte("<h1>{{.Title}}</h1><p>{{.Body}}</p>"), 0644)
	if err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	engine := NewTemplateEngine()
	engine.SetTemplateDir(tmpDir)

	result, err := engine.RenderFile("test.html", map[string]interface{}{
		"Title": "Test Page",
		"Body":  "This is a test.",
	})
	if err != nil {
		t.Fatalf("RenderFile failed: %v", err)
	}

	expected := "<h1>Test Page</h1><p>This is a test.</p>"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestTemplateEngineLayout(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create a page template
	err := os.WriteFile(filepath.Join(tmpDir, "page.html"), []byte("<h1>{{.Title}}</h1>"), 0644)
	if err != nil {
		t.Fatalf("Failed to create page template: %v", err)
	}

	engine := NewTemplateEngine()
	engine.SetTemplateDir(tmpDir)

	// Register layout
	engine.RegisterLayout("main", `<html><body>{{.Content}}</body></html>`)

	result, err := engine.RenderWithLayout("main", "page.html", map[string]interface{}{
		"Title": "My Page",
	})
	if err != nil {
		t.Fatalf("RenderWithLayout failed: %v", err)
	}

	expected := "<html><body><h1>My Page</h1></body></html>"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestTemplateEnginePartials(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create a template that uses partials
	err := os.WriteFile(filepath.Join(tmpDir, "page.html"),
		[]byte(`{{include "header"}}<h1>{{.Title}}</h1>{{include "footer"}}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	engine := NewTemplateEngine()
	engine.SetTemplateDir(tmpDir)
	engine.RegisterPartial("header", "<header>Header</header>")
	engine.RegisterPartial("footer", "<footer>Footer</footer>")

	result, err := engine.Render("page.html", map[string]interface{}{
		"Title": "Test",
	})
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	expected := "<header>Header</header><h1>Test</h1><footer>Footer</footer>"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestTemplateEngineCaching(t *testing.T) {
	tmpDir := t.TempDir()
	tmplPath := filepath.Join(tmpDir, "cached.html")

	engine := NewTemplateEngine()
	engine.SetTemplateDir(tmpDir)
	engine.SetCache(true)

	// Write first version
	os.WriteFile(tmplPath, []byte("Version 1: {{.Name}}"), 0644)
	result1, err := engine.Render("cached.html", map[string]interface{}{"Name": "Alice"})
	if err != nil {
		t.Fatalf("First render failed: %v", err)
	}
	if result1 != "Version 1: Alice" {
		t.Errorf("Expected 'Version 1: Alice', got '%s'", result1)
	}

	// Update file - should still get cached version
	os.WriteFile(tmplPath, []byte("Version 2: {{.Name}}"), 0644)
	result2, err := engine.Render("cached.html", map[string]interface{}{"Name": "Bob"})
	if err != nil {
		t.Fatalf("Second render failed: %v", err)
	}
	if result2 != "Version 1: Bob" {
		t.Errorf("Expected cached 'Version 1: Bob', got '%s'", result2)
	}

	// Clear cache and re-render
	engine.ClearCache()
	result3, err := engine.Render("cached.html", map[string]interface{}{"Name": "Bob"})
	if err != nil {
		t.Fatalf("Third render failed: %v", err)
	}
	if result3 != "Version 2: Bob" {
		t.Errorf("Expected 'Version 2: Bob', got '%s'", result3)
	}
}

func TestDefaultFuncMap(t *testing.T) {
	engine := NewTemplateEngine()

	tests := []struct {
		name     string
		template string
		data     interface{}
		expected string
	}{
		{"upper", `{{upper "hello"}}`, nil, "HELLO"},
		{"lower", `{{lower "HELLO"}}`, nil, "hello"},
		{"trim", `{{trim "  hello  "}}`, nil, "hello"},
		{"add", `{{add 3 5}}`, nil, "8"},
		{"sub", `{{sub 10 3}}`, nil, "7"},
		{"mul", `{{mul 4 5}}`, nil, "20"},
		{"div", `{{div 20 4}}`, nil, "5"},
		{"div_zero", `{{div 20 0}}`, nil, "0"},
		{"mod", `{{mod 10 3}}`, nil, "1"},
		{"len_string", `{{len "hello"}}`, nil, "5"},
		{"default", `{{default "fallback" .Missing}}`, map[string]interface{}{"Missing": ""}, "fallback"},
		{"html_escape", `{{html "<script>"}}`, nil, "&lt;script&gt;"},
		{"replace", `{{replace "hello world" "world" "go"}}`, nil, "hello go"},
		{"contains_true", `{{if contains "hello world" "world"}}yes{{end}}`, nil, "yes"},
		{"contains_false", `{{if contains "hello" "world"}}yes{{end}}`, nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.RenderString(tt.template, tt.data)
			if err != nil {
				t.Fatalf("RenderString failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestView(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "view.html"),
		[]byte("{{.Title}} - {{.Message}}"), 0644)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	engine := NewTemplateEngine()
	engine.SetTemplateDir(tmpDir)

	view := NewView(engine)
	view.With("Title", "Test")
	view.With("Message", "Hello World")

	result, err := view.Render("view.html")
	if err != nil {
		t.Fatalf("View.Render failed: %v", err)
	}
	expected := "Test - Hello World"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Test WithData
	view2 := NewView(engine)
	view2.WithData(map[string]interface{}{
		"Title":   "Batch",
		"Message": "Multiple",
	})
	result2, err := view2.Render("view.html")
	if err != nil {
		t.Fatalf("View.Render failed: %v", err)
	}
	expected2 := "Batch - Multiple"
	if result2 != expected2 {
		t.Errorf("Expected '%s', got '%s'", expected2, result2)
	}

	// Test WithLayout
	engine.RegisterLayout("wrapper", "<wrap>{{.Content}}</wrap>")
	view3 := NewView(engine)
	view3.With("Title", "Layout").With("Message", "Test")
	view3.WithLayout("wrapper")
	result3, err := view3.Render("view.html")
	if err != nil {
		t.Fatalf("View.Render with layout failed: %v", err)
	}
	expected3 := "<wrap>Layout - Test</wrap>"
	if result3 != expected3 {
		t.Errorf("Expected '%s', got '%s'", expected3, result3)
	}
}

func TestTemplateResponse(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "resp.html"),
		[]byte("<h1>{{.Title}}</h1>"), 0644)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	engine := NewTemplateEngine()
	engine.SetTemplateDir(tmpDir)

	resp, err := engine.RenderToResponse("resp.html", map[string]interface{}{
		"Title": "Hello",
	}, 200)
	if err != nil {
		t.Fatalf("RenderToResponse failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	if resp.Content != "<h1>Hello</h1>" {
		t.Errorf("Expected '<h1>Hello</h1>', got '%s'", resp.Content)
	}
	if resp.Headers["Content-Type"] != "text/html; charset=utf-8" {
		t.Errorf("Expected HTML content type, got '%s'", resp.Headers["Content-Type"])
	}
}

func TestAutoConfig(t *testing.T) {
	// Test basic config operations
	ac := NewAutoConfig("testapp")

	// Set and get
	ac.Set("app.name", "My App")
	ac.Set("app.version", "1.0.0")
	ac.Set("server.port", 8080)
	ac.Set("server.debug", true)

	if name := ac.GetString("app.name"); name != "My App" {
		t.Errorf("Expected 'My App', got '%s'", name)
	}
	if port := ac.GetInt("server.port"); port != 8080 {
		t.Errorf("Expected 8080, got %d", port)
	}
	if debug := ac.GetBool("server.debug"); !debug {
		t.Error("Expected true, got false")
	}

	// Has
	if !ac.Has("app.name") {
		t.Error("Expected app.name to exist")
	}
	if ac.Has("nonexistent") {
		t.Error("Expected nonexistent to not exist")
	}

	// Defaults
	if val := ac.GetStringDefault("nonexistent", "default"); val != "default" {
		t.Errorf("Expected 'default', got '%s'", val)
	}
	if val := ac.GetIntDefault("nonexistent", 42); val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}
	if val := ac.GetBoolDefault("nonexistent", true); !val {
		t.Error("Expected true default")
	}
}

func TestAutoConfigEnvironment(t *testing.T) {
	ac := NewAutoConfig("testapp")

	// Test environment detection
	os.Setenv("APP_ENV", "production")
	defer os.Unsetenv("APP_ENV")

	ac.DetectEnvironment()
	if !ac.IsProduction() {
		t.Error("Expected production environment")
	}
	if ac.IsDevelopment() {
		t.Error("Expected not development")
	}

	// Test other environments
	ac2 := NewAutoConfig("testapp")
	ac2.SetEnvironment(EnvDevelopment)
	if !ac2.IsDevelopment() {
		t.Error("Expected development")
	}

	ac3 := NewAutoConfig("testapp")
	ac3.SetEnvironment(EnvTesting)
	if !ac3.IsTesting() {
		t.Error("Expected testing")
	}

	ac4 := NewAutoConfig("testapp")
	ac4.SetEnvironment(EnvStaging)
	if !ac4.IsStaging() {
		t.Error("Expected staging")
	}
}

func TestAutoConfigFromEnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("TESTAPP_SERVER_PORT", "9090")
	os.Setenv("TESTAPP_SERVER_HOST", "localhost")
	os.Setenv("TESTAPP_APP_DEBUG", "true")
	os.Setenv("TESTAPP_APP_VERSION", "2.0.0")
	defer func() {
		os.Unsetenv("TESTAPP_SERVER_PORT")
		os.Unsetenv("TESTAPP_SERVER_HOST")
		os.Unsetenv("TESTAPP_APP_DEBUG")
		os.Unsetenv("TESTAPP_APP_VERSION")
	}()

	ac := NewAutoConfig("testapp")
	ac.AddSource("env", "env", 10, "TESTAPP")
	err := ac.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if port := ac.GetInt("server.port"); port != 9090 {
		t.Errorf("Expected 9090, got %d", port)
	}
	if host := ac.GetString("server.host"); host != "localhost" {
		t.Errorf("Expected 'localhost', got '%s'", host)
	}
	if debug := ac.GetBool("app.debug"); !debug {
		t.Error("Expected app.debug=true")
	}
}

func TestAutoConfigFromFile(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")
	configContent := `{
		"server": {
			"port": 3000,
			"host": "0.0.0.0"
		},
		"database": {
			"type": "sqlite",
			"path": "./app.db"
		}
	}`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	ac := NewAutoConfig("testapp")
	ac.SetConfigDir(tmpDir)
	ac.AddSource("config.json", "file", 10, configFile)
	err = ac.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if port := ac.GetInt("server.port"); port != 3000 {
		t.Errorf("Expected 3000, got %d", port)
	}
	if host := ac.GetString("server.host"); host != "0.0.0.0" {
		t.Errorf("Expected '0.0.0.0', got '%s'", host)
	}
	if dbType := ac.GetString("database.type"); dbType != "sqlite" {
		t.Errorf("Expected 'sqlite', got '%s'", dbType)
	}
}

func TestAutoConfigValidate(t *testing.T) {
	ac := NewAutoConfig("testapp")
	ac.Set("app.name", "Test")
	ac.Set("server.port", 8080)

	// Should pass
	err := ac.Validate([]string{"app.name", "server.port"})
	if err != nil {
		t.Errorf("Expected validation to pass, got: %v", err)
	}

	// Should fail
	err = ac.Validate([]string{"app.name", "missing.key"})
	if err == nil {
		t.Error("Expected validation to fail")
	}
	if !strings.Contains(err.Error(), "missing.key") {
		t.Errorf("Expected error to mention 'missing.key', got: %v", err)
	}
}

func TestAutoConfigToJSON(t *testing.T) {
	ac := NewAutoConfig("testapp")
	ac.Set("app.name", "Test")
	ac.Set("server.port", 8080)

	jsonStr, err := ac.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}
	if !strings.Contains(jsonStr, "Test") {
		t.Errorf("Expected JSON to contain 'Test', got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, "8080") {
		t.Errorf("Expected JSON to contain '8080', got: %s", jsonStr)
	}
}

func TestAppConfig(t *testing.T) {
	// Create temp config files
	tmpDir := t.TempDir()

	baseConfig := `{
		"server": {"port": 8080, "host": "localhost"},
		"database": {"path": "./app.db"}
	}`
	prodConfig := `{
		"server": {"port": 443, "host": "0.0.0.0"}
	}`

	os.WriteFile(filepath.Join(tmpDir, "config.json"), []byte(baseConfig), 0644)
	os.WriteFile(filepath.Join(tmpDir, "config.production.json"), []byte(prodConfig), 0644)

	// Test with development environment
	os.Setenv("APP_ENV", "development")

	ac := NewAppConfig("myapp")
	ac.config.SetConfigDir(tmpDir)
	ac.config.AddDefaultSources()
	err := ac.config.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if port := ac.GetInt("server.port"); port != 8080 {
		t.Errorf("Expected 8080, got %d", port)
	}

	dbConfig := ac.GetDatabaseConfig()
	if dbConfig["path"] != "./app.db" {
		t.Errorf("Expected './app.db', got '%v'", dbConfig["path"])
	}

	// Test with production environment - need fresh instance
	os.Setenv("APP_ENV", "production")
	ac2 := NewAppConfig("myapp")
	ac2.config.SetConfigDir(tmpDir)
	ac2.config.DetectEnvironment() // Must detect before adding sources
	ac2.config.AddDefaultSources()
	err = ac2.config.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if port := ac2.GetInt("server.port"); port != 443 {
		t.Errorf("Expected 443 (production), got %d", port)
	}
	if host := ac2.GetString("server.host"); host != "0.0.0.0" {
		t.Errorf("Expected '0.0.0.0' (production), got '%s'", host)
	}
	os.Unsetenv("APP_ENV")
}
