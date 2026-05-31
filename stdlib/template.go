package stdlib

import (
	"bytes"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"text/template"
)

// TemplateEngine provides HTML template rendering
type TemplateEngine struct {
	templates   map[string]*template.Template
	layouts     map[string]string
	partials    map[string]string
	funcMap     template.FuncMap
	templateDir string
	mu          sync.RWMutex
	cache       bool
}

// NewTemplateEngine creates a new template engine
func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{
		templates: make(map[string]*template.Template),
		layouts:   make(map[string]string),
		partials:  make(map[string]string),
		funcMap:   defaultFuncMap(),
		cache:     true,
	}
}

// SetTemplateDir sets the directory for template files
func (t *TemplateEngine) SetTemplateDir(dir string) *TemplateEngine {
	t.templateDir = dir
	return t
}

// SetCache enables or disables template caching
func (t *TemplateEngine) SetCache(cache bool) *TemplateEngine {
	t.cache = cache
	return t
}

// AddFunc adds a custom function to the template engine
func (t *TemplateEngine) AddFunc(name string, fn interface{}) *TemplateEngine {
	t.funcMap[name] = fn
	return t
}

// RegisterLayout registers a layout template
func (t *TemplateEngine) RegisterLayout(name, content string) *TemplateEngine {
	t.layouts[name] = content
	return t
}

// RegisterPartial registers a partial template
func (t *TemplateEngine) RegisterPartial(name, content string) *TemplateEngine {
	t.partials[name] = content
	return t
}

// LoadLayoutFile loads a layout from a file
func (t *TemplateEngine) LoadLayoutFile(name, filename string) error {
	content, err := t.readFile(filename)
	if err != nil {
		return fmt.Errorf("failed to load layout file %s: %w", filename, err)
	}
	t.layouts[name] = content
	return nil
}

// LoadPartialFile loads a partial from a file
func (t *TemplateEngine) LoadPartialFile(name, filename string) error {
	content, err := t.readFile(filename)
	if err != nil {
		return fmt.Errorf("failed to load partial file %s: %w", filename, err)
	}
	t.partials[name] = content
	return nil
}

// Render renders a template with the given data
func (t *TemplateEngine) Render(templateName string, data interface{}) (string, error) {
	// Check cache first
	if t.cache {
		t.mu.RLock()
		if tmpl, ok := t.templates[templateName]; ok {
			t.mu.RUnlock()
			return t.executeTemplate(tmpl, data)
		}
		t.mu.RUnlock()
	}

	// Load and compile template
	tmpl, err := t.compileTemplate(templateName)
	if err != nil {
		return "", err
	}

	// Cache if enabled
	if t.cache {
		t.mu.Lock()
		t.templates[templateName] = tmpl
		t.mu.Unlock()
	}

	return t.executeTemplate(tmpl, data)
}

// RenderString renders a template string with the given data
func (t *TemplateEngine) RenderString(templateStr string, data interface{}) (string, error) {
	tmpl, err := template.New("inline").Funcs(t.funcMap).Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	return t.executeTemplate(tmpl, data)
}

// RenderFile renders a template file with the given data
func (t *TemplateEngine) RenderFile(filename string, data interface{}) (string, error) {
	content, err := t.readFile(filename)
	if err != nil {
		return "", err
	}

	return t.RenderString(content, data)
}

// RenderWithLayout renders a template with a layout
func (t *TemplateEngine) RenderWithLayout(layoutName, templateName string, data interface{}) (string, error) {
	// Render the main template content
	content, err := t.Render(templateName, data)
	if err != nil {
		return "", err
	}

	// Get the layout
	layout, ok := t.layouts[layoutName]
	if !ok {
		return "", fmt.Errorf("layout %s not found", layoutName)
	}

	// Replace {{.Content}} in layout with rendered content
	layoutData := map[string]interface{}{
		"Content": content,
	}

	// Merge with original data if it's a map
	if dataMap, ok := data.(map[string]interface{}); ok {
		for k, v := range dataMap {
			layoutData[k] = v
		}
	}

	return t.RenderString(layout, layoutData)
}

// compileTemplate compiles a template with all partials
func (t *TemplateEngine) compileTemplate(name string) (*template.Template, error) {
	// Check if it's a file or inline content
	var content string
	if strings.HasSuffix(name, ".html") || strings.HasSuffix(name, ".tmpl") {
		var err error
		content, err = t.readFile(name)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("template %s not found", name)
	}

	// Preprocess includes
	processed, err := t.preprocessIncludes(content)
	if err != nil {
		return nil, err
	}

	// Parse template
	tmpl, err := template.New(name).Funcs(t.funcMap).Parse(processed)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	return tmpl, nil
}

// preprocessIncludes replaces {{include "name"}} with partial content
func (t *TemplateEngine) preprocessIncludes(content string) (string, error) {
	// Match {{include "name"}}
	re := regexp.MustCompile(`\{\{include\s+"([^"]+)"\}\}`)

	result := re.ReplaceAllStringFunc(content, func(match string) string {
		submatch := re.FindStringSubmatch(match)
		if len(submatch) < 2 {
			return match
		}

		name := submatch[1]
		partial, ok := t.partials[name]
		if !ok {
			return fmt.Sprintf("<!-- Partial %s not found -->", name)
		}

		return partial
	})

	return result, nil
}

// executeTemplate executes a template and returns the result
func (t *TemplateEngine) executeTemplate(tmpl *template.Template, data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	return buf.String(), nil
}

// readFile reads a file from the template directory
func (t *TemplateEngine) readFile(filename string) (string, error) {
	var path string
	if filepath.IsAbs(filename) {
		path = filename
	} else if t.templateDir != "" {
		path = filepath.Join(t.templateDir, filename)
	} else {
		path = filename
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", path, err)
	}

	return string(content), nil
}

// ClearCache clears the template cache
func (t *TemplateEngine) ClearCache() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.templates = make(map[string]*template.Template)
}

// defaultFuncMap returns the default template functions
func defaultFuncMap() template.FuncMap {
	return template.FuncMap{
		// HTML escaping
		"html": func(s string) string {
			return html.EscapeString(s)
		},

		// String functions
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": strings.Title,
		"trim":  strings.TrimSpace,

		// String manipulation
		"replace": strings.ReplaceAll,
		"split":   strings.Split,
		"join":    strings.Join,
		"contains": strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,

		// Math
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"mul": func(a, b int) int { return a * b },
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"mod": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a % b
		},

		// Comparison
		"eq": func(a, b interface{}) bool { return a == b },
		"ne": func(a, b interface{}) bool { return a != b },
		"lt": func(a, b int) bool { return a < b },
		"le": func(a, b int) bool { return a <= b },
		"gt": func(a, b int) bool { return a > b },
		"ge": func(a, b int) bool { return a >= b },

		// Logical
		"and": func(a, b bool) bool { return a && b },
		"or":  func(a, b bool) bool { return a || b },
		"not": func(a bool) bool { return !a },

		// Type conversion
		"toString": func(v interface{}) string {
			return fmt.Sprintf("%v", v)
		},
		"toInt": func(v interface{}) int {
			switch val := v.(type) {
			case int:
				return val
			case int64:
				return int(val)
			case float64:
				return int(val)
			case string:
				var i int
				fmt.Sscanf(val, "%d", &i)
				return i
			default:
				return 0
			}
		},

		// Formatting
		"format": func(format string, args ...interface{}) string {
			return fmt.Sprintf(format, args...)
		},

		// Safe HTML (no escaping)
		"safe": func(s string) string {
			return s
		},

		// Default value
		"default": func(def, val interface{}) interface{} {
			if val == nil || val == "" || val == 0 {
				return def
			}
			return val
		},

		// Slice/array functions
		"len": func(v interface{}) int {
			switch val := v.(type) {
			case string:
				return len(val)
			case []interface{}:
				return len(val)
			case []string:
				return len(val)
			case []int:
				return len(val)
			case map[string]interface{}:
				return len(val)
			default:
				return 0
			}
		},

		// First/Last
		"first": func(arr []interface{}) interface{} {
			if len(arr) == 0 {
				return nil
			}
			return arr[0]
		},
		"last": func(arr []interface{}) interface{} {
			if len(arr) == 0 {
				return nil
			}
			return arr[len(arr)-1]
		},

		// Index access
		"index": func(arr []interface{}, i int) interface{} {
			if i < 0 || i >= len(arr) {
				return nil
			}
			return arr[i]
		},

		// Map access
		"get": func(m map[string]interface{}, key string) interface{} {
			return m[key]
		},

		// Date formatting
		"dateFormat": func(layout string, v interface{}) string {
			switch val := v.(type) {
			case string:
				return val
			default:
				return fmt.Sprintf("%v", val)
			}
		},

		// JSON encoding
		"json": func(v interface{}) string {
			return fmt.Sprintf("%v", v)
		},

		// URL encoding
		"urlEncode": func(s string) string {
			return strings.ReplaceAll(strings.ReplaceAll(s, " ", "%20"), "/", "%2F")
		},
	}
}

// View is a convenience wrapper for rendering views
type View struct {
	engine *TemplateEngine
	data   map[string]interface{}
	layout string
}

// NewView creates a new view with the given engine
func NewView(engine *TemplateEngine) *View {
	return &View{
		engine: engine,
		data:   make(map[string]interface{}),
	}
}

// With adds data to the view
func (v *View) With(key string, value interface{}) *View {
	v.data[key] = value
	return v
}

// WithData adds multiple data items to the view
func (v *View) WithData(data map[string]interface{}) *View {
	for k, val := range data {
		v.data[k] = val
	}
	return v
}

// WithLayout sets the layout for the view
func (v *View) WithLayout(layout string) *View {
	v.layout = layout
	return v
}

// Render renders the view
func (v *View) Render(templateName string) (string, error) {
	if v.layout != "" {
		return v.engine.RenderWithLayout(v.layout, templateName, v.data)
	}
	return v.engine.Render(templateName, v.data)
}

// TemplateResponse represents an HTTP response with a rendered template
type TemplateResponse struct {
	Content    string
	StatusCode int
	Headers    map[string]string
}

// RenderToResponse renders a template and returns a TemplateResponse
func (t *TemplateEngine) RenderToResponse(templateName string, data interface{}, statusCode int) (*TemplateResponse, error) {
	content, err := t.Render(templateName, data)
	if err != nil {
		return nil, err
	}

	return &TemplateResponse{
		Content:    content,
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "text/html; charset=utf-8",
		},
	}, nil
}
