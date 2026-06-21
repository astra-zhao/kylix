package wasi_test

import (
	"testing"

	"kylix/pkg/wasi"
)

// Tests run with native stub implementations (not WASI target).

func TestStdout(t *testing.T) {
	// Smoke test: just verify it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Stdout panicked: %v", r)
		}
	}()
	wasi.Stdout("test output")
}

func TestGetenv(t *testing.T) {
	v := wasi.Getenv("PATH")
	// PATH should be set in any normal environment
	if v == "" {
		t.Log("PATH is empty — may be running in restricted environment")
	}
}

func TestGetenvMissing(t *testing.T) {
	v := wasi.Getenv("KYLIX_WASI_TEST_NONEXISTENT_VAR_XYZ")
	if v != "" {
		t.Errorf("expected empty string for missing var, got %q", v)
	}
}

func TestClockMonotonic(t *testing.T) {
	t1 := wasi.ClockMonotonic()
	t2 := wasi.ClockMonotonic()
	if t2 < t1 {
		t.Errorf("monotonic clock went backwards: %d < %d", t2, t1)
	}
}

func TestClockWalltime(t *testing.T) {
	ts := wasi.ClockWalltime()
	if ts <= 0 {
		t.Errorf("unexpected walltime: %d", ts)
	}
	// Should be after 2024-01-01 (Unix: 1704067200)
	if ts < 1704067200 {
		t.Errorf("walltime too old: %d", ts)
	}
}

func TestArgs(t *testing.T) {
	args := wasi.Args()
	// In test context, Args() may return nil or the test binary args.
	// We just verify it doesn't panic.
	_ = args
}

func TestEnviron(t *testing.T) {
	env := wasi.Environ()
	if env == nil {
		t.Error("Environ() returned nil")
	}
	if len(env) == 0 {
		t.Log("Environ() returned empty list — may be restricted environment")
	}
}

func TestReadWriteFile(t *testing.T) {
	path := t.TempDir() + "/wasi_test.txt"
	content := "Hello from WASI test!"

	if err := wasi.WriteFile(path, content); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := wasi.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if got != content {
		t.Errorf("content mismatch: got %q, want %q", got, content)
	}
}

func TestReadFileNotFound(t *testing.T) {
	_, err := wasi.ReadFile("/nonexistent/path/xyz123.txt")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
