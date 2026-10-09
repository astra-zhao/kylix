//go:build wasip1

package wasi

import (
	"errors"
	"unsafe"
)

// High-level API on top of the wasi_snapshot_preview1 import table.
// This file does not import os or fmt: stdout is fd_write, clocks are
// clock_time_get, files go through path_open on the preopened directory (fd 3).

// Errno is a wasi_snapshot_preview1 error code.
type Errno int32

func (e Errno) Error() string {
	return "wasi " + ErrnoName(int32(e))
}

func errnoErr(code int32) error {
	if code == ErrnoSuccess {
		return nil
	}
	return Errno(code)
}

// Stdout writes s to file descriptor 1.
func Stdout(s string) {
	_ = writeBytes(FDStdout, []byte(s))
}

// Stderr writes s to file descriptor 2.
func Stderr(s string) {
	_ = writeBytes(FDStderr, []byte(s))
}

// Stdin reads one line from file descriptor 0, without the trailing newline.
func Stdin() string {
	var b []byte
	tmp := make([]byte, 1)
	for {
		n, errno := readBytes(FDStdin, tmp)
		if errno != ErrnoSuccess || n == 0 {
			break
		}
		if tmp[0] == '\n' {
			break
		}
		if tmp[0] != '\r' {
			b = append(b, tmp[0])
		}
	}
	return string(b)
}

// Args returns command-line arguments, excluding the program name.
func Args() []string {
	ptrs, buf, errno := loadCounted(argsSizesGet, argsGet)
	if errno != ErrnoSuccess || len(buf) == 0 {
		return nil
	}
	base := uint32(uintptr(unsafe.Pointer(&buf[0])))
	return DecodeIndexedStrings(base, ptrs, buf, true)
}

// Getenv returns the value of an environment variable, or "" if unset.
func Getenv(name string) string {
	for _, kv := range Environ() {
		k, v, ok := SplitEnv(kv)
		if ok && k == name {
			return v
		}
	}
	return ""
}

// Environ returns environment entries as "KEY=VALUE" strings.
func Environ() []string {
	ptrs, buf, errno := loadCounted(environSizesGet, environGet)
	if errno != ErrnoSuccess || len(buf) == 0 {
		return []string{}
	}
	base := uint32(uintptr(unsafe.Pointer(&buf[0])))
	out := DecodeIndexedStrings(base, ptrs, buf, false)
	if out == nil {
		return []string{}
	}
	return out
}

// ClockMonotonic returns clock_time_get(CLOCK_MONOTONIC) in nanoseconds.
func ClockMonotonic() int64 {
	return clockNanos(ClockIDMonotonic)
}

// ClockWalltime returns clock_time_get(CLOCK_REALTIME) in Unix seconds.
func ClockWalltime() int64 {
	return clockNanos(ClockIDRealtime) / 1_000_000_000
}

func clockNanos(id int32) int64 {
	var ns uint64
	if clockTimeGet(id, 1, unsafe.Pointer(&ns)) != ErrnoSuccess {
		return 0
	}
	return int64(ns)
}

// RandomBytes fills n bytes from random_get. n <= 0 yields nil.
func RandomBytes(n int) ([]byte, error) {
	if n <= 0 {
		return nil, nil
	}
	if n > 1<<20 {
		return nil, errors.New("wasi: RandomBytes larger than 1MiB")
	}
	buf := make([]byte, n)
	errno := randomGet(unsafe.Pointer(&buf[0]), int32(n))
	if errno != ErrnoSuccess {
		return nil, errnoErr(errno)
	}
	return buf, nil
}

// ReadFile reads path relative to the first preopened directory (fd 3).
func ReadFile(path string) (string, error) {
	rights := RightFDRead | RightFDSeek | RightFDTell | RightFDFilestatGet
	fd, errno := openAt(path, 0, rights)
	if errno != ErrnoSuccess {
		return "", errnoErr(errno)
	}
	defer fdClose(fd)
	b, errno := readFileFD(fd)
	if errno != ErrnoSuccess {
		return "", errnoErr(errno)
	}
	return string(b), nil
}

// WriteFile writes content via path_open(CREAT|TRUNC) on the preopened directory.
func WriteFile(path, content string) error {
	rights := RightFDWrite | RightFDSeek | RightFDTell | RightFDFilestatGet | RightPathCreateFile
	fd, errno := openAt(path, OFlagCreat|OFlagTrunc, rights)
	if errno != ErrnoSuccess {
		return errnoErr(errno)
	}
	defer fdClose(fd)
	return errnoErr(writeBytes(fd, []byte(content)))
}

// WasiExit terminates the instance via proc_exit.
func WasiExit(code int) {
	procExit(int32(code))
	// proc_exit does not return. If a host stub ever did, do not continue.
	for {
	}
}
