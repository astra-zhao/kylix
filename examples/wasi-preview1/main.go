// GOOS=wasip1 smoke for pkg/wasi's wasi_snapshot_preview1 import table.
//
//	GOOS=wasip1 GOARCH=wasm go build -o preview1.wasm .
//	wasmtime --dir . --env NAME=Kylix preview1.wasm
package main

import "kylix/pkg/wasi"

func main() {
	wasi.Stdout("preview1 ok\n")
	if wasi.ClockWalltime() < 1700000000 {
		wasi.Stderr("wall clock looks unset\n")
		wasi.WasiExit(2)
	}
	if wasi.ClockMonotonic() <= 0 {
		wasi.Stderr("monotonic clock is zero\n")
		wasi.WasiExit(3)
	}
	buf, err := wasi.RandomBytes(8)
	if err != nil || len(buf) != 8 {
		wasi.Stderr("random_get failed\n")
		wasi.WasiExit(4)
	}
	wasi.Stdout("NAME=" + wasi.Getenv("NAME") + "\n")
	body, err := wasi.ReadFile("note.txt")
	if err != nil {
		wasi.Stderr("read: " + err.Error() + "\n")
		wasi.WasiExit(5)
	}
	wasi.Stdout(body)
	if err := wasi.WriteFile("out.txt", "stored\n"); err != nil {
		wasi.Stderr("write: " + err.Error() + "\n")
		wasi.WasiExit(6)
	}
	again, err := wasi.ReadFile("out.txt")
	if err != nil || again != "stored\n" {
		wasi.Stderr("rewrite mismatch\n")
		wasi.WasiExit(7)
	}
	wasi.Stdout("stored ok\n")
}
