/* Portable SHA-256 and MD5 for android/ios.
 *
 * Desktop LLVM targets keep calling OpenSSL. These two symbols are compiled
 * into the mobile link so Sha256 / Md5 / HmacSha256 do not pull -lcrypto.
 * Public domain implementation of FIPS 180-4 and RFC 1321.
 */
#include <stddef.h>
#include <stdint.h>
#include <string.h>

static uint32_t rotr(uint32_t x, uint32_t n) {
	return (x >> n) | (x << (32u - n));
}

static uint32_t rotl(uint32_t x, uint32_t n) {
	return (x << n) | (x >> (32u - n));
}

static uint32_t load_be32(const unsigned char *p) {
	return ((uint32_t)p[0] << 24) | ((uint32_t)p[1] << 16) |
	       ((uint32_t)p[2] << 8) | (uint32_t)p[3];
}

static void store_be32(unsigned char *p, uint32_t v) {
	p[0] = (unsigned char)(v >> 24);
	p[1] = (unsigned char)(v >> 16);
	p[2] = (unsigned char)(v >> 8);
	p[3] = (unsigned char)v;
}

static uint32_t load_le32(const unsigned char *p) {
	return (uint32_t)p[0] | ((uint32_t)p[1] << 8) |
	       ((uint32_t)p[2] << 16) | ((uint32_t)p[3] << 24);
}

static void store_le32(unsigned char *p, uint32_t v) {
	p[0] = (unsigned char)v;
	p[1] = (unsigned char)(v >> 8);
	p[2] = (unsigned char)(v >> 16);
	p[3] = (unsigned char)(v >> 24);
}

static const uint32_t sha256_k[64] = {
	0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1,
	0x923f82a4, 0xab1c5ed5, 0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3,
	0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174, 0xe49b69c1, 0xefbe4786,
	0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
	0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147,
	0x06ca6351, 0x14292967, 0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13,
	0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85, 0xa2bfe8a1, 0xa81a664b,
	0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
	0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a,
	0x5b9cca4f, 0x682e6ff3, 0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208,
	0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2,
};

static void sha256_block(uint32_t s[8], const unsigned char blk[64]) {
	uint32_t w[64];
	uint32_t a, b, c, d, e, f, g, h;
	int i;

	for (i = 0; i < 16; i++) {
		w[i] = load_be32(blk + (i * 4));
	}
	for (i = 16; i < 64; i++) {
		uint32_t s0 = rotr(w[i - 15], 7) ^ rotr(w[i - 15], 18) ^ (w[i - 15] >> 3);
		uint32_t s1 = rotr(w[i - 2], 17) ^ rotr(w[i - 2], 19) ^ (w[i - 2] >> 10);
		w[i] = w[i - 16] + s0 + w[i - 7] + s1;
	}
	a = s[0];
	b = s[1];
	c = s[2];
	d = s[3];
	e = s[4];
	f = s[5];
	g = s[6];
	h = s[7];
	for (i = 0; i < 64; i++) {
		uint32_t S1 = rotr(e, 6) ^ rotr(e, 11) ^ rotr(e, 25);
		uint32_t ch = (e & f) ^ ((~e) & g);
		uint32_t t1 = h + S1 + ch + sha256_k[i] + w[i];
		uint32_t S0 = rotr(a, 2) ^ rotr(a, 13) ^ rotr(a, 22);
		uint32_t maj = (a & b) ^ (a & c) ^ (b & c);
		uint32_t t2 = S0 + maj;
		h = g;
		g = f;
		f = e;
		e = d + t1;
		d = c;
		c = b;
		b = a;
		a = t1 + t2;
	}
	s[0] += a;
	s[1] += b;
	s[2] += c;
	s[3] += d;
	s[4] += e;
	s[5] += f;
	s[6] += g;
	s[7] += h;
}

void kylix_sha256(const void *data, size_t len, unsigned char out[32]) {
	uint32_t s[8] = {
		0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a,
		0x510e527f, 0x9b05688c, 0x1f83d9ab, 0x5be0cd19,
	};
	const unsigned char *p = (const unsigned char *)data;
	unsigned char blk[64];
	size_t off = 0;
	size_t rem;
	uint64_t bits;
	int i;

	if (p == NULL) {
		len = 0;
	}
	while (len - off >= 64) {
		sha256_block(s, p + off);
		off += 64;
	}
	rem = len - off;
	memset(blk, 0, sizeof blk);
	if (rem != 0) {
		memcpy(blk, p + off, rem);
	}
	blk[rem] = 0x80;
	if (rem >= 56) {
		sha256_block(s, blk);
		memset(blk, 0, sizeof blk);
	}
	bits = (uint64_t)len * 8u;
	for (i = 0; i < 8; i++) {
		blk[63 - i] = (unsigned char)(bits >> (8 * i));
	}
	sha256_block(s, blk);
	for (i = 0; i < 8; i++) {
		store_be32(out + (i * 4), s[i]);
	}
}

/* RFC 1321 MD5. T[i] = floor(2^32 * abs(sin(i+1))). */
static const uint32_t md5_t[64] = {
	0xd76aa478, 0xe8c7b756, 0x242070db, 0xc1bdceee, 0xf57c0faf, 0x4787c62a,
	0xa8304613, 0xfd469501, 0x698098d8, 0x8b44f7af, 0xffff5bb1, 0x895cd7be,
	0x6b901122, 0xfd987193, 0xa679438e, 0x49b40821, 0xf61e2562, 0xc040b340,
	0x265e5a51, 0xe9b6c7aa, 0xd62f105d, 0x02441453, 0xd8a1e681, 0xe7d3fbc8,
	0x21e1cde6, 0xc33707d6, 0xf4d50d87, 0x455a14ed, 0xa9e3e905, 0xfcefa3f8,
	0x676f02d9, 0x8d2a4c8a, 0xfffa3942, 0x8771f681, 0x6d9d6122, 0xfde5380c,
	0xa4beea44, 0x4bdecfa9, 0xf6bb4b60, 0xbebfbc70, 0x289b7ec6, 0xeaa127fa,
	0xd4ef3085, 0x04881d05, 0xd9d4d039, 0xe6db99e5, 0x1fa27cf8, 0xc4ac5665,
	0xf4292244, 0x432aff97, 0xab9423a7, 0xfc93a039, 0x655b59c3, 0x8f0ccc92,
	0xffeff47d, 0x85845dd1, 0x6fa87e4f, 0xfe2ce6e0, 0xa3014314, 0x4e0811a1,
	0xf7537e82, 0xbd3af235, 0x2ad7d2bb, 0xeb86d391,
};

static const unsigned char md5_s[64] = {
	7, 12, 17, 22, 7, 12, 17, 22, 7, 12, 17, 22, 7, 12, 17, 22,
	5, 9, 14, 20, 5, 9, 14, 20, 5, 9, 14, 20, 5, 9, 14, 20,
	4, 11, 16, 23, 4, 11, 16, 23, 4, 11, 16, 23, 4, 11, 16, 23,
	6, 10, 15, 21, 6, 10, 15, 21, 6, 10, 15, 21, 6, 10, 15, 21,
};

static void md5_block(uint32_t st[4], const unsigned char blk[64]) {
	uint32_t m[16];
	uint32_t a, b, c, d;
	int i;

	for (i = 0; i < 16; i++) {
		m[i] = load_le32(blk + (i * 4));
	}
	a = st[0];
	b = st[1];
	c = st[2];
	d = st[3];
	for (i = 0; i < 64; i++) {
		uint32_t f;
		int g;
		uint32_t t;
		if (i < 16) {
			f = (b & c) | ((~b) & d);
			g = i;
		} else if (i < 32) {
			f = (d & b) | ((~d) & c);
			g = (5 * i + 1) % 16;
		} else if (i < 48) {
			f = b ^ c ^ d;
			g = (3 * i + 5) % 16;
		} else {
			f = c ^ (b | (~d));
			g = (7 * i) % 16;
		}
		t = d;
		d = c;
		c = b;
		b = b + rotl(a + f + md5_t[i] + m[g], md5_s[i]);
		a = t;
	}
	st[0] += a;
	st[1] += b;
	st[2] += c;
	st[3] += d;
}

void kylix_md5(const void *data, size_t len, unsigned char out[16]) {
	uint32_t st[4] = {0x67452301, 0xefcdab89, 0x98badcfe, 0x10325476};
	const unsigned char *p = (const unsigned char *)data;
	unsigned char blk[64];
	size_t off = 0;
	size_t rem;
	uint64_t bits;
	int i;

	if (p == NULL) {
		len = 0;
	}
	while (len - off >= 64) {
		md5_block(st, p + off);
		off += 64;
	}
	rem = len - off;
	memset(blk, 0, sizeof blk);
	if (rem != 0) {
		memcpy(blk, p + off, rem);
	}
	blk[rem] = 0x80;
	if (rem >= 56) {
		md5_block(st, blk);
		memset(blk, 0, sizeof blk);
	}
	bits = (uint64_t)len * 8u;
	for (i = 0; i < 8; i++) {
		blk[56 + i] = (unsigned char)(bits >> (8 * i));
	}
	md5_block(st, blk);
	for (i = 0; i < 4; i++) {
		store_le32(out + (i * 4), st[i]);
	}
}
