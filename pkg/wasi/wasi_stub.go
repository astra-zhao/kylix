//go:build !wasip1

// Stub implementations for non-WASI platforms (native development + tests).
package wasi

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

// Stdout writes s to standard output.
func Stdout(s string) {
	fmt.Fprint(os.Stdout, s)
}

// Stderr writes s to standard error.
func Stderr(s string) {
	fmt.Fprint(os.Stderr, s)
}

// Stdin reads one line from standard input (blocking).
func Stdin() string {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text()
	}
	return ""
}

// Args returns the command-line arguments (excluding the program name).
func Args() []string {
	if len(os.Args) > 1 {
		return os.Args[1:]
	}
	return nil
}

// Getenv returns the value of the named environment variable.
func Getenv(name string) string {
	return os.Getenv(name)
}

// Environ returns all environment variables as "KEY=VALUE" strings.
func Environ() []string {
	return os.Environ()
}

// ClockMonotonic returns a monotonic time in nanoseconds.
func ClockMonotonic() int64 {
	return time.Now().UnixNano()
}

// ClockWalltime returns the current wall clock time in seconds since Unix epoch.
func ClockWalltime() int64 {
	return time.Now().Unix()
}

// ReadFile reads the entire contents of a file.
func ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteFile writes content to a file, creating it if necessary.
func WriteFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// WasiExit terminates the process with the given exit code.
func WasiExit(code int) {
	os.Exit(code)
}
