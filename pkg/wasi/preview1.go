package wasi

import "encoding/binary"

// Pure logic for the wasi_snapshot_preview1 ABI. No os, no fmt, no syscall.
// The host stub and the wasip1 import bindings both sit on top of these
// layouts. Preview1 linear memory is wasm32: every iovec field is a
// little-endian u32 (8 bytes per entry, no padding).

const (
	ErrnoSuccess    int32 = 0
	ErrnoBadf       int32 = 8
	ErrnoFault      int32 = 21
	ErrnoInval      int32 = 28
	ErrnoIo         int32 = 29
	ErrnoNoent      int32 = 43
	ErrnoNosys      int32 = 51
	ErrnoNotdir     int32 = 53
	ErrnoAcces      int32 = 2
	ErrnoExist      int32 = 20
	ErrnoIsdir      int32 = 31
	ErrnoNotcapable int32 = 76
)

const (
	ClockIDRealtime  int32 = 0
	ClockIDMonotonic int32 = 1
)

const (
	FDStdin   int32 = 0
	FDStdout  int32 = 1
	FDStderr  int32 = 2
	FDPreopen int32 = 3 // first preopened directory (wasmtime --dir)
)

const (
	WhenceSet int32 = 0
	WhenceCur int32 = 1
	WhenceEnd int32 = 2
)

const (
	OFlagCreat int32 = 1 << 0
	OFlagExcl  int32 = 1 << 2
	OFlagTrunc int32 = 1 << 3
)

const (
	LookupSymlinkFollow int32 = 1 << 0
)

const (
	RightFDRead         int64 = 1 << 1
	RightFDSeek         int64 = 1 << 2
	RightFDTell         int64 = 1 << 5
	RightFDWrite        int64 = 1 << 6
	RightPathCreateFile int64 = 1 << 10
	RightFDFilestatGet  int64 = 1 << 21
)

// CIVec is one ciovec/iovec entry: linear-memory offset plus length.
type CIVec struct {
	Off uint32
	Len uint32
}

// CIVecSize is the encoded size of one preview1 iovec (two u32s).
const CIVecSize = 8

// EncodeCIVecs returns the little-endian preview1 iovec array.
func EncodeCIVecs(vecs []CIVec) []byte {
	dst := make([]byte, CIVecSize*len(vecs))
	for i, v := range vecs {
		binary.LittleEndian.PutUint32(dst[i*CIVecSize:], v.Off)
		binary.LittleEndian.PutUint32(dst[i*CIVecSize+4:], v.Len)
	}
	return dst
}

// DecodeIndexedStrings splits a WASI args_get / environ_get buffer.
// ptrs are absolute linear-memory offsets; base is the offset of buf[0].
// dropFirst omits entry 0 (WASI argv[0], the program name).
func DecodeIndexedStrings(base uint32, ptrs []uint32, buf []byte, dropFirst bool) []string {
	if len(buf) == 0 || len(ptrs) == 0 {
		return nil
	}
	out := make([]string, 0, len(ptrs))
	for i, p := range ptrs {
		if dropFirst && i == 0 {
			continue
		}
		if p < base {
			continue
		}
		off := int(p - base)
		if off < 0 || off >= len(buf) {
			continue
		}
		end := off
		for end < len(buf) && buf[end] != 0 {
			end++
		}
		out = append(out, string(buf[off:end]))
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// SplitEnv splits a WASI "KEY=VALUE" entry. The key must be non-empty.
func SplitEnv(kv string) (key, val string, ok bool) {
	for i := 0; i < len(kv); i++ {
		if kv[i] == '=' {
			if i == 0 {
				return "", "", false
			}
			return kv[:i], kv[i+1:], true
		}
	}
	return "", "", false
}

// ErrnoName returns the preview1 name for a known errno, or "ESUCCESS"/"EUNKNOWN".
func ErrnoName(code int32) string {
	if code == ErrnoSuccess {
		return "ESUCCESS"
	}
	if name, ok := errnoNames[code]; ok {
		return name
	}
	return "EUNKNOWN"
}

var errnoNames = map[int32]string{
	ErrnoAcces:      "EACCES",
	ErrnoBadf:       "EBADF",
	ErrnoExist:      "EEXIST",
	ErrnoFault:      "EFAULT",
	ErrnoInval:      "EINVAL",
	ErrnoIo:         "EIO",
	ErrnoIsdir:      "EISDIR",
	ErrnoNoent:      "ENOENT",
	ErrnoNosys:      "ENOSYS",
	ErrnoNotdir:     "ENOTDIR",
	ErrnoNotcapable: "ENOTCAPABLE",
}
