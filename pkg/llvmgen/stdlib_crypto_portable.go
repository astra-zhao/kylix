package llvmgen

// stdlib_crypto_portable.go — android/ios digest calls.
//
// Desktop bodies keep `call ptr @SHA256` / `call ptr @MD5` (OpenSSL), and that
// text must stay byte-identical. Mobile bodies call kylix_sha256 / kylix_md5
// from portable/hash.c instead, so the link does not need -lcrypto.

func (g *Generator) emitPortableHashDecl() {
	if !g.portableHashOS() || g.portableHashDecl {
		return
	}
	g.portableHashDecl = true
	g.line("declare void @kylix_sha256(ptr noundef, i64 noundef, ptr noundef)")
	g.line("declare void @kylix_md5(ptr noundef, i64 noundef, ptr noundef)")
}

// emitSHA256Digest hashes data[0:n] into out (32 bytes).
// data, n, and out are IR operands already formatted ("%%data", "%12", "96").
func (g *Generator) emitSHA256Digest(data, n, out string) {
	if g.portableHashOS() {
		g.emitPortableHashDecl()
		g.line(fmtSHA256Portable(data, n, out))
		return
	}
	g.line(fmtSHA256OpenSSL(data, n, out))
}

func (g *Generator) emitMD5Digest(data, n, out string) {
	if g.portableHashOS() {
		g.emitPortableHashDecl()
		g.line(fmtMD5Portable(data, n, out))
		return
	}
	g.line(fmtMD5OpenSSL(data, n, out))
}

func fmtSHA256OpenSSL(data, n, out string) string {
	return "  call ptr @SHA256(ptr " + data + ", i64 " + n + ", ptr " + out + ")"
}

func fmtSHA256Portable(data, n, out string) string {
	return "  call void @kylix_sha256(ptr " + data + ", i64 " + n + ", ptr " + out + ")"
}

func fmtMD5OpenSSL(data, n, out string) string {
	return "  call ptr @MD5(ptr " + data + ", i64 " + n + ", ptr " + out + ")"
}

func fmtMD5Portable(data, n, out string) string {
	return "  call void @kylix_md5(ptr " + data + ", i64 " + n + ", ptr " + out + ")"
}
