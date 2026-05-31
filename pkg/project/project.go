package project

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const ConfigFileName = "kylix.toml"

// Config represents a kylix.toml project configuration
type Config struct {
	Name    string
	Version string
	Main    string // entry point file, default "main.klx"
	Output  string // output directory, default "build/"
	GoMod   string // Go module name for generated code

	Path string // path to the kylix.toml file itself (not stored in file)
}

// Default returns a Config with sensible defaults
func Default() *Config {
	return &Config{
		Name:    "myapp",
		Version: "0.1.0",
		Main:    "main.klx",
		Output:  "build/",
		GoMod:   "myapp",
	}
}

// Find walks up from dir looking for kylix.toml. Returns nil if not found.
func Find(dir string) (*Config, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	for {
		cfgPath := filepath.Join(dir, ConfigFileName)
		if _, err := os.Stat(cfgPath); err == nil {
			return Load(cfgPath)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, nil // reached root, no config found
		}
		dir = parent
	}
}

// Load reads and parses a kylix.toml file
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cfg := Default()
	cfg.Path = path

	scanner := bufio.NewScanner(f)
	section := ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Section header
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.Trim(line, "[]")
			continue
		}

		// Key = value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, "\"'") // strip quotes

		switch section {
		case "project":
			switch key {
			case "name":
				cfg.Name = value
			case "version":
				cfg.Version = value
			case "main":
				cfg.Main = value
			}
		case "build":
			switch key {
			case "output":
				cfg.Output = value
			case "go_module":
				cfg.GoMod = value
			}
		}
	}

	return cfg, scanner.Err()
}

// Save writes the config to disk
func (c *Config) Save(path string) error {
	var b strings.Builder
	b.WriteString("# Kylix project configuration\n\n")
	b.WriteString("[project]\n")
	b.WriteString(fmt.Sprintf("name = \"%s\"\n", c.Name))
	b.WriteString(fmt.Sprintf("version = \"%s\"\n", c.Version))
	b.WriteString(fmt.Sprintf("main = \"%s\"\n", c.Main))
	b.WriteString("\n[build]\n")
	b.WriteString(fmt.Sprintf("output = \"%s\"\n", c.Output))
	b.WriteString(fmt.Sprintf("go_module = \"%s\"\n", c.GoMod))

	return ioutil.WriteFile(path, []byte(b.String()), 0644)
}

// ProjectDir returns the directory containing kylix.toml
func (c *Config) ProjectDir() string {
	if c.Path == "" {
		return "."
	}
	return filepath.Dir(c.Path)
}

// MainFilePath returns the absolute path to the main source file
func (c *Config) MainFilePath() string {
	dir := c.ProjectDir()
	return filepath.Join(dir, c.Main)
}

// OutputDir returns the absolute path to the output directory
func (c *Config) OutputDir() string {
	dir := c.ProjectDir()
	return filepath.Join(dir, c.Output)
}

// Init creates a new project in the given directory with a template kylix.toml
// and a hello-world main.klx
func Init(dir string, name string) (*Config, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("cannot create directory %s: %v", dir, err)
	}

	if name == "" {
		name = filepath.Base(dir)
	}

	// Convert hyphens to underscores for valid Pascal identifiers
	pascalName := strings.ReplaceAll(name, "-", "_")

	cfg := &Config{
		Name:    name,
		Version: "0.1.0",
		Main:    "main.klx",
		Output:  "build/",
		GoMod:   name,
		Path:    filepath.Join(dir, ConfigFileName),
	}

	// Write kylix.toml
	if err := cfg.Save(cfg.Path); err != nil {
		return nil, err
	}

	// Write template main.klx
	mainContent := fmt.Sprintf(`program %s;

// Entry point for the %s project

begin
  WriteLn('Hello from %s!');
end.
`, pascalName, name, name)

	mainPath := filepath.Join(dir, "main.klx")
	if err := ioutil.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		return nil, err
	}

	// Create output directory
	if err := os.MkdirAll(filepath.Join(dir, cfg.Output), 0755); err != nil {
		return nil, err
	}

	// Write .gitignore
	gitignore := "build/\n*.go\nkylix\n"
	if err := ioutil.WriteFile(filepath.Join(dir, ".gitignore"), []byte(gitignore), 0644); err != nil {
		return nil, err
	}

	return cfg, nil
}

// FindAllKlxFiles returns all .klx files in the project directory (excluding output)
func (c *Config) FindAllKlxFiles() ([]string, error) {
	var files []string
	dir := c.ProjectDir()
	outDir := c.OutputDir()

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip output dir and hidden dirs
			if path == outDir || strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".klx") {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}
