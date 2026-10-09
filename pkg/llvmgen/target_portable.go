package llvmgen

// target_portable.go — android/ios are their own targets, not "whatever is
// left after windows". Desktop IR (linux/darwin/windows) stays on the paths
// those targets already emitted. v0.15.

// portableHashOS is true when Sha256/Md5/HmacSha256 must not call OpenSSL.
// The digest comes from portable/hash.c, linked only for these targets.
func (g *Generator) portableHashOS() bool {
	return g.targetOS == "android" || g.targetOS == "ios"
}

// netReuseConsts returns SOL_SOCKET and SO_REUSEADDR.
//
// Linux and Android bionic use the Linux ABI (1, 2). Darwin — macOS and
// iOS — and Winsock use 0xffff / 4. Android must not inherit the BSD default:
// setsockopt would target a bogus level and reuse would silently do nothing.
func (g *Generator) netReuseConsts() (solSocket, soReuse string) {
	switch g.targetOS {
	case "linux", "android":
		return "1", "2"
	case "darwin", "ios", "windows":
		return "65535", "4"
	default:
		return "65535", "4"
	}
}

// localtimeSymbol is the reentrant broken-down-time call for this target.
// Windows UCRT has localtime_s (arguments reversed). bionic and Darwin have
// POSIX localtime_r and do not have localtime_s. The LP64 struct tm prefix
// (tm_sec at 0 through tm_wday at 24, 4-byte ints) is the same on glibc,
// bionic, and Darwin, so the existing 56-byte buffer and field offsets apply.
func (g *Generator) localtimeSymbol() string {
	switch g.targetOS {
	case "windows":
		return "localtime_s"
	case "android", "ios", "darwin", "linux":
		return "localtime_r"
	default:
		return "localtime_r"
	}
}
