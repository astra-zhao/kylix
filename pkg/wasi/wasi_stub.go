//go:build !wasip1

// Host stand-in for native tests and `go test`. The wasi_snapshot_preview1
// import table itself lives in preview1_wasip1.go (GOOS=wasip1) and
// internal/wasiapi.Preview1 — this file does not pretend to be that table.
package wasi

import (
	"bufio"
	"crypto/rand"
	"os"
	"time"
)

// Stdout writes s to the host standard output.
func Stdout(s string) {
	_, _ = os.Stdout.Write([]byte(s))
}

// Stderr writes s to the host standard error.
func Stderr(s string) {
	_, _ = os.Stderr.Write([]byte(s))
}

// Stdin reads one line from the host standard input.
func Stdin() string {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text()
	}
	return ""
}

// Args returns the host command-line arguments, excluding the program name.
func Args() []string {
	if len(os.Args) > 1 {
		return os.Args[1:]
	}
	return nil
}

// Getenv returns the host environment variable.
func Getenv(name string) string {
	return os.Getenv(name)
}

// Environ returns the host environment as "KEY=VALUE" strings.
func Environ() []string {
	env := os.Environ()
	if env == nil {
		return []string{}
	}
	return env
}

// ClockMonotonic returns a host monotonic timestamp in nanoseconds.
func ClockMonotonic() int64 {
	return time.Now().UnixNano()
}

// ClockWalltime returns host wall time in Unix seconds.
func ClockWalltime() int64 {
	return time.Now().Unix()
}

// RandomBytes reads n bytes from the host CSPRNG.
func RandomBytes(n int) ([]byte, error) {
	if n <= 0 {
		return nil, nil
	}
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	return buf, nil
}

// ReadFile reads a host file.
func ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteFile writes a host file.
func WriteFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// WasiExit terminates the host process.
func WasiExit(code int) {
	os.Exit(code)
}
