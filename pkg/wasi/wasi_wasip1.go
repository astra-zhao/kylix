//go:build wasip1

// Native WASI implementations using Go 1.21+ wasip1 syscalls.
// When compiled with GOOS=wasip1, the standard library provides
// os, bufio, and time with WASI syscall backends.
package wasi

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

// Stdout writes s to standard output (fd 1).
func Stdout(s string) {
	fmt.Fprint(os.Stdout, s)
}

// Stderr writes s to standard error (fd 2).
func Stderr(s string) {
	fmt.Fprint(os.Stderr, s)
}

// Stdin reads one line from standard input (fd 0).
func Stdin() string {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text()
	}
	return ""
}

// Args returns the command-line arguments (excluding program name).
// Under WASI, arguments are passed via args_get/args_sizes_get syscalls.
func Args() []string {
	if len(os.Args) > 1 {
		return os.Args[1:]
	}
	return nil
}

// Getenv returns the value of an environment variable.
// Under WASI, env vars are passed via environ_get/environ_sizes_get.
func Getenv(name string) string {
	return os.Getenv(name)
}

// Environ returns all environment variables as "KEY=VALUE" strings.
func Environ() []string {
	return os.Environ()
}

// ClockMonotonic returns a monotonic time in nanoseconds.
// Uses WASI clock_time_get(CLOCK_MONOTONIC).
func ClockMonotonic() int64 {
	return time.Now().UnixNano()
}

// ClockWalltime returns wall clock time in seconds since Unix epoch.
// Uses WASI clock_time_get(CLOCK_REALTIME).
func ClockWalltime() int64 {
	return time.Now().Unix()
}

// ReadFile reads file contents via WASI fd_read.
func ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteFile writes content to a file via WASI fd_write.
func WriteFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// WasiExit terminates the WASI process via proc_exit.
func WasiExit(code int) {
	os.Exit(code)
}
