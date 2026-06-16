// runner.go — Kylix test runner.
//
// Test file convention:
//   - File name ends with _test.klx
//   - Contains procedures named Test<Something>() with no parameters
//   - Uses built-in Assert(condition, message) to check expectations
//
// Strategy:
//  1. Parse _test.klx to discover Test* procedures
//  2. Compile _test.klx to Go with the generator
//  3. Inject Assert() + a dispatch main() into the generated Go
//  4. Run `go run harness.go <TestName>` for each test
//  5. Report TAP output
package testrunner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"kylix/ast"
	"kylix/generator"
	"kylix/lexer"
	"kylix/parser"
)

// TestCase is a single discovered test procedure.
type TestCase struct {
	Name string
	File string
}

// TestResult holds the outcome of one test.
type TestResult struct {
	TestCase
	Passed  bool
	Message string
}

// Runner discovers and runs Kylix tests.
type Runner struct {
	Verbose bool
}

// New returns a Runner.
func New(verbose bool) *Runner {
	return &Runner{Verbose: verbose}
}

// DiscoverFile returns Test* procedures found in a single _test.klx file.
func (r *Runner) DiscoverFile(path string) ([]TestCase, error) {
	if !strings.HasSuffix(path, "_test.klx") {
		return nil, nil
	}
	return discoverInFile(path)
}

// Discover walks dir and returns all Test* procedures in *_test.klx files.
func (r *Runner) Discover(dir string) ([]TestCase, error) {
	var cases []TestCase
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}
		if !strings.HasSuffix(path, "_test.klx") {
			return nil
		}
		found, ferr := discoverInFile(path)
		if ferr != nil {
			return ferr
		}
		cases = append(cases, found...)
		return nil
	})
	return cases, err
}

// discoverInFile parses a .klx file and extracts Test* procedure names.
func discoverInFile(path string) ([]TestCase, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	l := lexer.New(string(src))
	p := parser.New(l)
	prog := p.ParseProgram()

	var cases []TestCase
	for _, decl := range prog.Declarations {
		fd, ok := decl.(*ast.FunctionDecl)
		if !ok {
			continue
		}
		if strings.HasPrefix(fd.Name, "Test") && len(fd.Parameters) == 0 {
			cases = append(cases, TestCase{Name: fd.Name, File: path})
		}
	}
	return cases, nil
}

// Run executes each test case and returns results.
// Cases are grouped by file: each file is compiled once, then each Test* is run.
func (r *Runner) Run(cases []TestCase) []TestResult {
	byFile := make(map[string][]string)
	var fileOrder []string
	for _, tc := range cases {
		if _, seen := byFile[tc.File]; !seen {
			fileOrder = append(fileOrder, tc.File)
		}
		byFile[tc.File] = append(byFile[tc.File], tc.Name)
	}

	var results []TestResult
	for _, file := range fileOrder {
		results = append(results, r.runFile(file, byFile[file])...)
	}
	return results
}

// runFile compiles a _test.klx with a test harness and runs each Test*.
func (r *Runner) runFile(file string, names []string) []TestResult {
	tmpDir, err := os.MkdirTemp("", "kylix-test-*")
	if err != nil {
		return failAll(file, names, err.Error())
	}
	defer os.RemoveAll(tmpDir)

	harnessPath, err := buildHarness(file, names, tmpDir)
	if err != nil {
		return failAll(file, names, fmt.Sprintf("compile: %v", err))
	}

	var results []TestResult
	for _, name := range names {
		tc := TestCase{Name: name, File: file}
		cmd := exec.Command("go", "run", harnessPath, name)
		out, runErr := cmd.CombinedOutput()
		outStr := strings.TrimSpace(string(out))
		if runErr != nil {
			msg := outStr
			if msg == "" {
				msg = runErr.Error()
			}
			results = append(results, TestResult{TestCase: tc, Passed: false, Message: msg})
		} else {
			results = append(results, TestResult{TestCase: tc, Passed: true})
		}
	}
	return results
}

// buildHarness compiles klxFile to Go, injects Assert() and a dispatch main().
func buildHarness(klxFile string, names []string, tmpDir string) (string, error) {
	src, err := os.ReadFile(klxFile)
	if err != nil {
		return "", err
	}

	l := lexer.New(string(src))
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		return "", fmt.Errorf("parse error in %s: %s", klxFile, errs[0])
	}

	// Generate Go body (unit mode — no main function)
	g := generator.New()
	goBody := g.GenerateBody(prog)

	var sb strings.Builder

	sb.WriteString("package main\n\n")
	sb.WriteString("import (\n\t\"fmt\"\n\t\"os\"\n\t\"runtime/debug\"\n)\n\n")

	// Inject Assert as a Go panic-based function
	sb.WriteString("// Assert is the Kylix test assertion built-in.\n")
	sb.WriteString("func Assert(cond bool, msg string) {\n")
	sb.WriteString("\tif !cond {\n\t\tpanic(\"FAIL: \" + msg)\n\t}\n}\n\n")

	// Suppress unused import warnings from generated code
	sb.WriteString("var _ = fmt.Sprintf\n\n")

	// Insert generated procedure bodies
	sb.WriteString(goBody)
	sb.WriteString("\n")

	// Dispatch main: run one test by name, panic on failure
	sb.WriteString("func main() {\n")
	sb.WriteString("\tif len(os.Args) < 2 {\n")
	sb.WriteString("\t\tfmt.Fprintln(os.Stderr, \"usage: harness <TestName>\")\n")
	sb.WriteString("\t\tos.Exit(1)\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\tname := os.Args[1]\n")
	sb.WriteString("\tdefer func() {\n")
	sb.WriteString("\t\tif rec := recover(); rec != nil {\n")
	sb.WriteString("\t\t\tfmt.Fprintln(os.Stderr, rec)\n")
	sb.WriteString("\t\t\t_ = debug.Stack()\n")
	sb.WriteString("\t\t\tos.Exit(1)\n")
	sb.WriteString("\t\t}\n")
	sb.WriteString("\t}()\n")
	sb.WriteString("\tswitch name {\n")
	for _, n := range names {
		sb.WriteString(fmt.Sprintf("\tcase %q:\n\t\t%s()\n", n, n))
	}
	sb.WriteString("\tdefault:\n")
	sb.WriteString("\t\tfmt.Fprintf(os.Stderr, \"unknown test: %s\\n\", name)\n")
	sb.WriteString("\t\tos.Exit(1)\n")
	sb.WriteString("\t}\n}\n")

	outPath := filepath.Join(tmpDir, "harness.go")
	if err := os.WriteFile(outPath, []byte(sb.String()), 0644); err != nil {
		return "", err
	}
	return outPath, nil
}

// ── TAP output ────────────────────────────────────────────────────────────────

// PrintTAP writes TAP version 14 to stdout.
func PrintTAP(results []TestResult) {
	fmt.Printf("TAP version 14\n1..%d\n", len(results))
	for i, r := range results {
		num := i + 1
		if r.Passed {
			fmt.Printf("ok %d - %s\n", num, r.Name)
		} else {
			fmt.Printf("not ok %d - %s\n", num, r.Name)
			if r.Message != "" {
				fmt.Printf("  ---\n  message: %q\n  file: %s\n  ...\n",
					r.Message, r.File)
			}
		}
	}
}

// Summary returns pass/fail counts.
func Summary(results []TestResult) (passed, failed int) {
	for _, r := range results {
		if r.Passed {
			passed++
		} else {
			failed++
		}
	}
	return
}

// failAll creates failure results for all names in a file.
func failAll(file string, names []string, msg string) []TestResult {
	results := make([]TestResult, len(names))
	for i, n := range names {
		results[i] = TestResult{
			TestCase: TestCase{Name: n, File: file},
			Passed:   false,
			Message:  msg,
		}
	}
	return results
}
