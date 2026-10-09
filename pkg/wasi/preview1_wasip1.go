//go:build wasip1

package wasi

import "unsafe"

// wasi_snapshot_preview1 import table.
// Each //go:wasmimport line is checked against internal/wasiapi.Preview1
// (TestPreview1Imports_Dispatchable). Signatures are the wasm32 ABI:
// pointers are linear-memory offsets (unsafe.Pointer on GOARCH=wasm).

//go:wasmimport wasi_snapshot_preview1 fd_write
func fdWrite(fd int32, iovs unsafe.Pointer, iovsLen int32, nwritten unsafe.Pointer) int32

//go:wasmimport wasi_snapshot_preview1 fd_read
func fdRead(fd int32, iovs unsafe.Pointer, iovsLen int32, nread unsafe.Pointer) int32

//go:wasmimport wasi_snapshot_preview1 fd_seek
func fdSeek(fd int32, offset int64, whence int32, newOffset unsafe.Pointer) int32

//go:wasmimport wasi_snapshot_preview1 fd_close
func fdClose(fd int32) int32

//go:wasmimport wasi_snapshot_preview1 path_open
func pathOpen(fd int32, dirflags int32, path unsafe.Pointer, pathLen int32, oflags int32, rightsBase int64, rightsInheriting int64, fdflags int32, opened unsafe.Pointer) int32

//go:wasmimport wasi_snapshot_preview1 clock_time_get
func clockTimeGet(clockID int32, precision int64, time unsafe.Pointer) int32

//go:wasmimport wasi_snapshot_preview1 random_get
func randomGet(buf unsafe.Pointer, bufLen int32) int32

//go:wasmimport wasi_snapshot_preview1 args_sizes_get
func argsSizesGet(argc unsafe.Pointer, argvBufSize unsafe.Pointer) int32

//go:wasmimport wasi_snapshot_preview1 args_get
func argsGet(argv unsafe.Pointer, argvBuf unsafe.Pointer) int32

//go:wasmimport wasi_snapshot_preview1 environ_sizes_get
func environSizesGet(count unsafe.Pointer, bufSize unsafe.Pointer) int32

//go:wasmimport wasi_snapshot_preview1 environ_get
func environGet(environ unsafe.Pointer, environBuf unsafe.Pointer) int32

//go:wasmimport wasi_snapshot_preview1 proc_exit
func procExit(code int32)

// iovec is the in-memory preview1 ciovec (two u32s, no padding).
type iovec struct {
	buf uint32
	len uint32
}

func writeBytes(fd int32, b []byte) int32 {
	if len(b) == 0 {
		return ErrnoSuccess
	}
	off := 0
	for off < len(b) {
		chunk := b[off:]
		vec := iovec{
			buf: uint32(uintptr(unsafe.Pointer(&chunk[0]))),
			len: uint32(len(chunk)),
		}
		var nwritten uint32
		errno := fdWrite(fd, unsafe.Pointer(&vec), 1, unsafe.Pointer(&nwritten))
		if errno != ErrnoSuccess {
			return errno
		}
		if nwritten == 0 {
			return ErrnoIo
		}
		off += int(nwritten)
	}
	return ErrnoSuccess
}

func readBytes(fd int32, b []byte) (int, int32) {
	if len(b) == 0 {
		return 0, ErrnoSuccess
	}
	vec := iovec{
		buf: uint32(uintptr(unsafe.Pointer(&b[0]))),
		len: uint32(len(b)),
	}
	var nread uint32
	errno := fdRead(fd, unsafe.Pointer(&vec), 1, unsafe.Pointer(&nread))
	if errno != ErrnoSuccess {
		return 0, errno
	}
	return int(nread), ErrnoSuccess
}

func loadCounted(sizes, get func(unsafe.Pointer, unsafe.Pointer) int32) ([]uint32, []byte, int32) {
	var count, sz uint32
	if e := sizes(unsafe.Pointer(&count), unsafe.Pointer(&sz)); e != ErrnoSuccess {
		return nil, nil, e
	}
	ptrs := make([]uint32, count)
	buf := make([]byte, sz)
	var pptr, bptr unsafe.Pointer
	if count > 0 {
		pptr = unsafe.Pointer(&ptrs[0])
	}
	if sz > 0 {
		bptr = unsafe.Pointer(&buf[0])
	}
	if e := get(pptr, bptr); e != ErrnoSuccess {
		return nil, nil, e
	}
	return ptrs, buf, ErrnoSuccess
}

func stringData(s string) (unsafe.Pointer, int32) {
	if s == "" {
		return unsafe.Pointer(&_emptyByte), 0
	}
	return unsafe.Pointer(unsafe.StringData(s)), int32(len(s))
}

var _emptyByte byte

func openAt(path string, oflags int32, rights int64) (int32, int32) {
	p, n := stringData(path)
	var out int32
	errno := pathOpen(FDPreopen, LookupSymlinkFollow, p, n, oflags, rights, 0, 0, unsafe.Pointer(&out))
	return out, errno
}

func readFileFD(fd int32) ([]byte, int32) {
	var size int64
	var scratch int64
	if e := fdSeek(fd, 0, WhenceEnd, unsafe.Pointer(&size)); e != ErrnoSuccess {
		return nil, e
	}
	if size < 0 {
		size = 0
	}
	if e := fdSeek(fd, 0, WhenceSet, unsafe.Pointer(&scratch)); e != ErrnoSuccess {
		return nil, e
	}
	if size == 0 {
		return []byte{}, ErrnoSuccess
	}
	buf := make([]byte, size)
	got := 0
	for got < len(buf) {
		n, e := readBytes(fd, buf[got:])
		if e != ErrnoSuccess {
			return buf[:got], e
		}
		if n == 0 {
			break
		}
		got += n
	}
	return buf[:got], ErrnoSuccess
}
