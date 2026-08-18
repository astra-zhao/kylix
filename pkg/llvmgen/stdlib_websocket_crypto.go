package llvmgen

import "fmt"

// stdlib_websocket_crypto.go — SHA-1 / base64 / randomness helpers used by the
// RFC 6455 WebSocket handshake (Sec-WebSocket-Accept = b64(sha1(key+GUID))).
// The Go backend uses crypto/sha1 + encoding/base64; here we emit hand-rolled
// IR (the encoding module's Base64Encode is strlen-based, and SHA-1 isn't in
// the crypto module). v6.4.0.

// wsWsGUID is RFC 6455 §4.2.2 magic string appended to Sec-WebSocket-Key.
const wsWsGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

// emitWsSha1Body emits:
//
//	void @__kylix_ws_sha1(ptr %data, i64 %len, ptr %out)
//
// Standard SHA-1: writes the 20-byte digest to %out. Words are i32 (LLVM add
// wraps mod 2^32). Message padded to a 512-bit boundary: 0x80, zeros, then the
// 64-bit big-endian bit length. State lives in memory slots so loop-carried
// values stay well-defined across SSA blocks.
func (g *Generator) emitWsSha1Body() {
	g.line("define void @__kylix_ws_sha1(ptr %data, i64 %len, ptr %out) {")
	g.line("entry:")
	// padLen = ((len + 8) / 64 + 1) * 64
	pl1 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %%len, 9", pl1))
	pl2 := g.tmp()
	g.line(fmt.Sprintf("  %s = udiv i64 %s, 64", pl2, pl1))
	pl3 := g.tmp()
	g.line(fmt.Sprintf("  %s = mul i64 %s, 64", pl3, pl2))
	pl4 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 64", pl4, pl3))
	padLen := pl4
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", buf, padLen))
	g.needMemcpy = true
	g.line(fmt.Sprintf("  call ptr @memcpy(ptr %s, ptr %%data, i64 %%len)", buf))
	// buf[len] = 0x80
	zp := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %%len", zp, buf))
	g.line(fmt.Sprintf("  store i8 -128, ptr %s", zp))
	// zero-fill buf[len+1 .. padLen-1] (pad + trailing length bytes).
	// loop z = len+1; z < padLen; z++ → buf[z] = 0
	zSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", zSlot))
	zp1 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %%len, 1", zp1))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", zp1, zSlot))
	zLoop := g.label()
	zBody := g.label()
	zDone := g.label()
	g.line(fmt.Sprintf("  br label %%%s", zLoop))
	g.line(fmt.Sprintf("%s:", zLoop))
	zv := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", zv, zSlot))
	ze := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, %s", ze, zv, padLen))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", ze, zBody, zDone))
	g.line(fmt.Sprintf("%s:", zBody))
	zpp := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", zpp, buf, zv))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", zpp))
	zn := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", zn, zv))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", zn, zSlot))
	g.line(fmt.Sprintf("  br label %%%s", zLoop))
	g.line(fmt.Sprintf("%s:", zDone))
	// bit length big-endian at buf[padLen-8 .. padLen-1]
	bitLen := g.tmp()
	g.line(fmt.Sprintf("  %s = mul i64 %%len, 8", bitLen))
	endPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", endPtr, buf, padLen))
	endM8 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 -8", endM8, endPtr))
	for i := 0; i < 8; i++ {
		sh := 8 * (7 - i)
		by := g.tmp()
		g.line(fmt.Sprintf("  %s = lshr i64 %s, %d", by, bitLen, sh))
		bc := g.tmp()
		g.line(fmt.Sprintf("  %s = trunc i64 %s to i8", bc, by))
		bp := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %d", bp, endM8, i))
		g.line(fmt.Sprintf("  store i8 %s, ptr %s", bc, bp))
	}
	// h0..h4 (i32 memory slots) + per-block initial copies
	hs := make([]string, 5)
	hInit := make([]string, 5)
	for i := range hs {
		hs[i] = g.tmp()
		g.line(fmt.Sprintf("  %s = alloca i32, align 4", hs[i]))
		hInit[i] = g.tmp()
		g.line(fmt.Sprintf("  %s = alloca i32, align 4", hInit[i]))
	}
	g.line(fmt.Sprintf("  store i32 1732584193, ptr %s", hs[0]))  // 0x67452301
	g.line(fmt.Sprintf("  store i32 -1009589776, ptr %s", hs[1])) // 0xEFCDAB89
	g.line(fmt.Sprintf("  store i32 -1732584194, ptr %s", hs[2])) // 0x98BADCFE
	g.line(fmt.Sprintf("  store i32 1009588838, ptr %s", hs[3]))  // 0x10325476
	g.line(fmt.Sprintf("  store i32 -1009637050, ptr %s", hs[4])) // 0xC3D2E1F0
	// w[80] i32 buffer
	wBuf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 320)", wBuf))

	// ---- per-block loop
	blkSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", blkSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", blkSlot))
	blkLoop := g.label()
	blkBody := g.label()
	blkDone := g.label()
	g.line(fmt.Sprintf("  br label %%%s", blkLoop))
	g.line(fmt.Sprintf("%s:", blkLoop))
	blk := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", blk, blkSlot))
	blkEnd := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, %s", blkEnd, blk, padLen))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", blkEnd, blkBody, blkDone))

	g.line(fmt.Sprintf("%s:", blkBody))
	// hInit[i] = hs[i] (snapshot for the h += block-final update)
	for i := 0; i < 5; i++ {
		hv := g.tmp()
		g.line(fmt.Sprintf("  %s = load i32, ptr %s", hv, hs[i]))
		g.line(fmt.Sprintf("  store i32 %s, ptr %s", hv, hInit[i]))
	}
	// w[0..15] = big-endian words at buf[blk + i*4]
	for i := 0; i < 16; i++ {
		bs := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 %s, %d", bs, blk, i*4))
		var acc string
		for j := 0; j < 4; j++ {
			bp := g.tmp()
			g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", bp, buf, bs))
			bb := g.tmp()
			g.line(fmt.Sprintf("  %s = load i8, ptr %s", bb, bp))
			bx := g.tmp()
			g.line(fmt.Sprintf("  %s = zext i8 %s to i32", bx, bb))
			bs2 := g.tmp()
			g.line(fmt.Sprintf("  %s = add i64 %s, 1", bs2, bs))
			bs = bs2
			if j == 0 {
				sl := g.tmp()
				g.line(fmt.Sprintf("  %s = shl i32 %s, 24", sl, bx))
				acc = sl
			} else {
				sl := g.tmp()
				g.line(fmt.Sprintf("  %s = shl i32 %s, %d", sl, bx, 8*(3-j)))
				nw := g.tmp()
				g.line(fmt.Sprintf("  %s = or i32 %s, %s", nw, acc, sl))
				acc = nw
			}
		}
		wi := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr inbounds i32, ptr %s, i64 %d", wi, wBuf, i))
		g.line(fmt.Sprintf("  store i32 %s, ptr %s", acc, wi))
	}
	// w[16..79] = rotl(w[i-3]^w[i-8]^w[i-14]^w[i-16], 1)
	for i := 16; i < 80; i++ {
		ld := func(o int) string {
			wp := g.tmp()
			g.line(fmt.Sprintf("  %s = getelementptr inbounds i32, ptr %s, i64 %d", wp, wBuf, i-o))
			wv := g.tmp()
			g.line(fmt.Sprintf("  %s = load i32, ptr %s", wv, wp))
			return wv
		}
		x1 := g.tmp()
		g.line(fmt.Sprintf("  %s = xor i32 %s, %s", x1, ld(3), ld(8)))
		x2 := g.tmp()
		g.line(fmt.Sprintf("  %s = xor i32 %s, %s", x2, x1, ld(14)))
		x3 := g.tmp()
		g.line(fmt.Sprintf("  %s = xor i32 %s, %s", x3, x2, ld(16)))
		l := g.tmp()
		g.line(fmt.Sprintf("  %s = shl i32 %s, 1", l, x3))
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = lshr i32 %s, 31", r, x3))
		ro := g.tmp()
		g.line(fmt.Sprintf("  %s = or i32 %s, %s", ro, l, r))
		wp := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr inbounds i32, ptr %s, i64 %d", wp, wBuf, i))
		g.line(fmt.Sprintf("  store i32 %s, ptr %s", ro, wp))
	}
	// 80-round loop: a..e live in hs[] memory slots.
	iSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i32, align 4", iSlot))
	g.line(fmt.Sprintf("  store i32 0, ptr %s", iSlot))
	roundLoop := g.label()
	roundBody := g.label()
	roundDone := g.label()
	g.line(fmt.Sprintf("  br label %%%s", roundLoop))
	g.line(fmt.Sprintf("%s:", roundLoop))
	ii := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", ii, iSlot))
	iiEnd := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i32 %s, 80", iiEnd, ii))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", iiEnd, roundBody, roundDone))
	g.line(fmt.Sprintf("%s:", roundBody))
	al := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", al, hs[0]))
	bl := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", bl, hs[1]))
	cl := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", cl, hs[2]))
	dl := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", dl, hs[3]))
	el := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", el, hs[4]))
	lt20 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i32 %s, 20", lt20, ii))
	lt40 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i32 %s, 40", lt40, ii))
	lt60 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i32 %s, 60", lt60, ii))
	f2 := g.tmp()
	g.line(fmt.Sprintf("  %s = xor i32 %s, %s", f2, bl, cl))
	f2v := g.tmp()
	g.line(fmt.Sprintf("  %s = xor i32 %s, %s", f2v, f2, dl))
	bc := g.tmp()
	g.line(fmt.Sprintf("  %s = and i32 %s, %s", bc, bl, cl))
	bd := g.tmp()
	g.line(fmt.Sprintf("  %s = and i32 %s, %s", bd, bl, dl))
	cd := g.tmp()
	g.line(fmt.Sprintf("  %s = and i32 %s, %s", cd, cl, dl))
	f3a := g.tmp()
	g.line(fmt.Sprintf("  %s = or i32 %s, %s", f3a, bc, bd))
	f3 := g.tmp()
	g.line(fmt.Sprintf("  %s = or i32 %s, %s", f3, f3a, cd))
	nb := g.tmp()
	g.line(fmt.Sprintf("  %s = xor i32 %s, -1", nb, bl))
	nbd := g.tmp()
	g.line(fmt.Sprintf("  %s = and i32 %s, %s", nbd, nb, dl))
	f1a := g.tmp()
	g.line(fmt.Sprintf("  %s = or i32 %s, %s", f1a, bc, nbd))
	sel1 := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i32 %s, i32 %s", sel1, lt20, f1a, f2v))
	sel2 := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i32 %s, i32 %s", sel2, lt40, sel1, f3))
	fval := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i32 %s, i32 %s", fval, lt60, sel2, f2v))
	k1 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i32 0, %d", k1, int32(0x5A827999)))
	k2 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i32 0, %d", k2, int32(0x6ED9EBA1)))
	k3 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i32 0, %d", k3, -1894007588))
	k4 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i32 0, %d", k4, -899497514))
	sk1 := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i32 %s, i32 %s", sk1, lt20, k1, k2))
	sk2 := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i32 %s, i32 %s", sk2, lt40, sk1, k3))
	kval := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i32 %s, i32 %s", kval, lt60, sk2, k4))
	wp := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i32, ptr %s, i32 %s", wp, wBuf, ii))
	wv := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", wv, wp))
	la := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i32 %s, 5", la, al))
	ra := g.tmp()
	g.line(fmt.Sprintf("  %s = lshr i32 %s, 27", ra, al))
	roa := g.tmp()
	g.line(fmt.Sprintf("  %s = or i32 %s, %s", roa, la, ra))
	t1 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i32 %s, %s", t1, roa, fval))
	t2 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i32 %s, %s", t2, t1, el))
	t3 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i32 %s, %s", t3, t2, kval))
	temp := g.tmp()
	g.line(fmt.Sprintf("  %s = add i32 %s, %s", temp, t3, wv))
	lb := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i32 %s, 30", lb, bl))
	rb := g.tmp()
	g.line(fmt.Sprintf("  %s = lshr i32 %s, 2", rb, bl))
	rob := g.tmp()
	g.line(fmt.Sprintf("  %s = or i32 %s, %s", rob, lb, rb))
	g.line(fmt.Sprintf("  store i32 %s, ptr %s", dl, hs[4]))
	g.line(fmt.Sprintf("  store i32 %s, ptr %s", cl, hs[3]))
	g.line(fmt.Sprintf("  store i32 %s, ptr %s", rob, hs[2]))
	g.line(fmt.Sprintf("  store i32 %s, ptr %s", al, hs[1]))
	g.line(fmt.Sprintf("  store i32 %s, ptr %s", temp, hs[0]))
	inext := g.tmp()
	g.line(fmt.Sprintf("  %s = add i32 %s, 1", inext, ii))
	g.line(fmt.Sprintf("  store i32 %s, ptr %s", inext, iSlot))
	g.line(fmt.Sprintf("  br label %%%s", roundLoop))

	g.line(fmt.Sprintf("%s:", roundDone))
	// h[i] += hInit[i]
	for i := 0; i < 5; i++ {
		cur := g.tmp()
		g.line(fmt.Sprintf("  %s = load i32, ptr %s", cur, hs[i]))
		init := g.tmp()
		g.line(fmt.Sprintf("  %s = load i32, ptr %s", init, hInit[i]))
		sum := g.tmp()
		g.line(fmt.Sprintf("  %s = add i32 %s, %s", sum, cur, init))
		g.line(fmt.Sprintf("  store i32 %s, ptr %s", sum, hs[i]))
	}
	blkN := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 64", blkN, blk))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", blkN, blkSlot))
	g.line(fmt.Sprintf("  br label %%%s", blkLoop))

	g.line(fmt.Sprintf("%s:", blkDone))
	// output h0..h4 big-endian
	for i := 0; i < 5; i++ {
		hv := g.tmp()
		g.line(fmt.Sprintf("  %s = load i32, ptr %s", hv, hs[i]))
		for j := 0; j < 4; j++ {
			by := g.tmp()
			g.line(fmt.Sprintf("  %s = lshr i32 %s, %d", by, hv, 8*(3-j)))
			bc := g.tmp()
			g.line(fmt.Sprintf("  %s = trunc i32 %s to i8", bc, by))
			bp := g.tmp()
			g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%out, i64 %d", bp, i*4+j))
			g.line(fmt.Sprintf("  store i8 %s, ptr %s", bc, bp))
		}
	}
	g.line("  ret void")
	g.line("}")
	g.line("")
}

// emitWsB64Body emits:
//
//	ptr @__kylix_ws_b64(ptr %data, i64 %len)
//
// Standard base64 (RFC 4648, with padding) of a length-bounded byte buffer —
// unlike encoding.Base64Encode this works on binary data containing NUL bytes.
// Returns a malloc'd NUL-terminated string.
func (g *Generator) emitWsB64Body() {
	tbl := g.addString("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/")
	g.line("define ptr @__kylix_ws_b64(ptr %data, i64 %len) {")
	g.line("entry:")
	ol1 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %%len, 2", ol1))
	ol2 := g.tmp()
	g.line(fmt.Sprintf("  %s = udiv i64 %s, 3", ol2, ol1))
	ol3 := g.tmp()
	g.line(fmt.Sprintf("  %s = mul i64 %s, 4", ol3, ol2))
	ol4 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", ol4, ol3))
	out := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", out, ol4))
	iSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", iSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", iSlot))
	loop := g.label()
	body := g.label()
	done := g.label()
	g.line(fmt.Sprintf("  br label %%%s", loop))
	g.line(fmt.Sprintf("%s:", loop))
	ii := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", ii, iSlot))
	ie := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, %%len", ie, ii))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", ie, body, done))
	g.line(fmt.Sprintf("%s:", body))
	rem := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %%len, %s", rem, ii))
	outIdx := g.tmp()
	g.line(fmt.Sprintf("  %s = udiv i64 %s, 3", outIdx, ii))
	outIdxMul := g.tmp()
	g.line(fmt.Sprintf("  %s = mul i64 %s, 4", outIdxMul, outIdx))
	outIdx = outIdxMul
	// a = data[i], b = data[i+1] (0 past end), c = data[i+2] (0 past end)
	readByte := func(off int64) (string, string) {
		// returns (byteReg, pastEndReg) where pastEnd is a predicate i64 0/1
		idx := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 %s, %d", idx, ii, off))
		past := g.tmp()
		g.line(fmt.Sprintf("  %s = icmp sge i64 %s, %%len", past, idx))
		pext := g.tmp()
		g.line(fmt.Sprintf("  %s = zext i1 %s to i64", pext, past))
		p := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%data, i64 %s", p, idx))
		b := g.tmp()
		g.line(fmt.Sprintf("  %s = load i8, ptr %s", b, p))
		z := g.tmp()
		g.line(fmt.Sprintf("  %s = zext i8 %s to i32", z, b))
		sel := g.tmp()
		g.line(fmt.Sprintf("  %s = select i1 %s, i32 0, i32 %s", sel, past, z))
		return sel, pext
	}
	a, _ := readByte(0)
	b, _ := readByte(1)
	c, _ := readByte(2)
	ta := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i32 %s, 16", ta, a))
	tb := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i32 %s, 8", tb, b))
	tab := g.tmp()
	g.line(fmt.Sprintf("  %s = or i32 %s, %s", tab, ta, tb))
	triple := g.tmp()
	g.line(fmt.Sprintf("  %s = or i32 %s, %s", triple, tab, c))
	// chr0..chr3; chr2/chr3 are '=' when padding
	emit := func(k int, chrReg string) {
		op := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %d", op, out, int64(k)))
		_ = op
		// outIdx is dynamic; compute op from outIdx + k
		op2 := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 %s, %d", op2, outIdx, k))
		opp := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", opp, out, op2))
		g.line(fmt.Sprintf("  store i8 %s, ptr %s", chrReg, opp))
	}
	// chr0 = alpha[(triple>>18)&0x3F]
	sh0 := g.tmp()
	g.line(fmt.Sprintf("  %s = lshr i32 %s, 18", sh0, triple))
	i0 := g.tmp()
	g.line(fmt.Sprintf("  %s = and i32 %s, 63", i0, sh0))
	c0 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [64 x i8], ptr %s, i64 0, i32 %s", c0, tbl, i0))
	v0 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", v0, c0))
	emit(0, v0)
	// chr1 = alpha[(triple>>12)&0x3F]
	sh1 := g.tmp()
	g.line(fmt.Sprintf("  %s = lshr i32 %s, 12", sh1, triple))
	i1 := g.tmp()
	g.line(fmt.Sprintf("  %s = and i32 %s, 63", i1, sh1))
	c1 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [64 x i8], ptr %s, i64 0, i32 %s", c1, tbl, i1))
	v1 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", v1, c1))
	emit(1, v1)
	// chr2 = rem<2 ? '=' : alpha[(triple>>6)&0x3F]
	sh2 := g.tmp()
	g.line(fmt.Sprintf("  %s = lshr i32 %s, 6", sh2, triple))
	i2 := g.tmp()
	g.line(fmt.Sprintf("  %s = and i32 %s, 63", i2, sh2))
	c2 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [64 x i8], ptr %s, i64 0, i32 %s", c2, tbl, i2))
	v2 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", v2, c2))
	p2 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, 2", p2, rem))
	eq2 := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i8 61, i8 %s", eq2, p2, v2))
	emit(2, eq2)
	// chr3 = rem<3 ? '=' : alpha[triple&0x3F]
	i3 := g.tmp()
	g.line(fmt.Sprintf("  %s = and i32 %s, 63", i3, triple))
	c3 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [64 x i8], ptr %s, i64 0, i32 %s", c3, tbl, i3))
	v3 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", v3, c3))
	p3 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, 3", p3, rem))
	eq3 := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i8 61, i8 %s", eq3, p3, v3))
	emit(3, eq3)
	// i += 3
	inext := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 3", inext, ii))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", inext, iSlot))
	g.line(fmt.Sprintf("  br label %%%s", loop))
	g.line(fmt.Sprintf("%s:", done))
	// out[outLen-1] = 0 (NUL terminator at (ol3) index)
	outNul := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", outNul, out, ol3))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", outNul))
	g.line(fmt.Sprintf("  ret ptr %s", out))
	g.line("}")
	g.line("")
}

// emitWsRandBody emits:
//
//	void @__kylix_ws_rand(ptr %buf, i64 %n)
//
// Fills buf with random bytes. macOS: arc4random_buf; Linux: getrandom;
// fallback (Windows / unknown): deterministic bytes from a counter (enough for
// the handshake to be well-formed, not for security).
func (g *Generator) emitWsRandBody() {
	g.line("define void @__kylix_ws_rand(ptr %buf, i64 %n) {")
	g.line("entry:")
	if g.targetOS == "darwin" {
		g.line("  call void @arc4random_buf(ptr %buf, i64 %n)")
	} else if g.targetOS == "linux" {
		// getrandom(buf, n, 0) returns ssize_t; ignore result.
		g.line("  call i64 @getrandom(ptr %buf, i64 %n, i32 0)")
	} else {
		// deterministic fill: buf[i] = (i*7+11)&0xFF
		iSlot := g.tmp()
		g.line(fmt.Sprintf("  %s = alloca i64, align 8", iSlot))
		g.line(fmt.Sprintf("  store i64 0, ptr %s", iSlot))
		loop := g.label()
		body := g.label()
		done := g.label()
		g.line(fmt.Sprintf("  br label %%%s", loop))
		g.line(fmt.Sprintf("%s:", loop))
		ii := g.tmp()
		g.line(fmt.Sprintf("  %s = load i64, ptr %s", ii, iSlot))
		ie := g.tmp()
		g.line(fmt.Sprintf("  %s = icmp slt i64 %s, %%n", ie, ii))
		g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", ie, body, done))
		g.line(fmt.Sprintf("%s:", body))
		mul := g.tmp()
		g.line(fmt.Sprintf("  %s = mul i64 %s, 7", mul, ii))
		add := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 %s, 11", add, mul))
		tr := g.tmp()
		g.line(fmt.Sprintf("  %s = trunc i64 %s to i8", tr, add))
		pp := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%buf, i64 %s", pp, ii))
		g.line(fmt.Sprintf("  store i8 %s, ptr %s", tr, pp))
		nn := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 %s, 1", nn, ii))
		g.line(fmt.Sprintf("  store i64 %s, ptr %s", nn, iSlot))
		g.line(fmt.Sprintf("  br label %%%s", loop))
		g.line(fmt.Sprintf("%s:", done))
	}
	g.line("  ret void")
	g.line("}")
	g.line("")
}
