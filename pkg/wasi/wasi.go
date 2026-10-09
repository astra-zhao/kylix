// Package wasi is the Kylix WASI surface.
//
// Two builds:
//
//   - GOOS=wasip1: preview1_wasip1.go is a wasi_snapshot_preview1 import
//     table (fd_write, fd_read, path_open, clock_time_get, random_get,
//     args_*, environ_*, proc_exit). High-level calls go through those
//     imports. There is no os/fmt wrapper on this build.
//   - every other GOOS: wasi_stub.go talks to the host so unit tests run
//     natively. That stub is not the import table.
//
// The canonical name list is internal/wasiapi.Preview1. DOM is not imported.
// File paths on wasip1 are relative to the first preopened directory (fd 3,
// wasmtime --dir).
package wasi
