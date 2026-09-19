// crypto.go — Kylix stdlib crypto module: hashes, HMAC, AES-GCM, BCrypt, random.
package stdlib

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
)

func Sha256(data string) string {
	sum := sha256.Sum256([]byte(data))
	return hex.EncodeToString(sum[:])
}

func Sha512(data string) string {
	sum := sha512.Sum512([]byte(data))
	return hex.EncodeToString(sum[:])
}

func Md5(data string) string {
	sum := md5.Sum([]byte(data))
	return hex.EncodeToString(sum[:])
}

func HmacSha256(key, data string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

func aesKey(key string) []byte {
	h := sha256.Sum256([]byte(key))
	return h[:]
}

func AesEncrypt(key, plaintext string) (string, error) {
	block, err := aes.NewCipher(aesKey(key))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(append(nonce, sealed...)), nil
}

func AesDecrypt(key, ciphertext string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(aesKey(key))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	ns := gcm.NonceSize()
	if len(raw) < ns {
		return "", fmt.Errorf("crypto: ciphertext too short")
	}
	plain, err := gcm.Open(nil, raw[:ns], raw[ns:], nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// BCryptHash — DEPRECATED / Go-only: real bcrypt ($2a$), incompatible with
// the LLVM backend's PBKDF2-based BCryptHash. New code should use
// Pbkdf2Hash/Pbkdf2Compare, which produce identical envelopes on both
// backends (v0.10.0 P2 decision).
func BCryptHash(password string, cost int64) (string, error) {
	c := int(cost)
	if c == 0 {
		c = bcrypt.DefaultCost
	}
	h, err := bcrypt.GenerateFromPassword([]byte(password), c)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

// Pbkdf2 envelope format — shared verbatim with the LLVM backend
// (pkg/llvmgen/stdlib_crypto.go): "pbkdf2$sha256$<iter>$<hex_salt(16B)>$<hex_out(32B)>".
// Field 3 is the direct iteration count (BCrypt* stores a log2 cost in the
// same position — the two formats coexist; Pbkdf2Compare rejects BCrypt
// hashes because cost values fall below the minimum iteration count).
const (
	pbkdf2MinIterations = 1000
	pbkdf2MaxIterations = 1 << 24
	pbkdf2SaltLen       = 16
	pbkdf2KeyLen        = 32
)

// Pbkdf2Hash derives a PBKDF2-HMAC-SHA256 password hash in the shared
// envelope, with a random 16-byte salt (v0.10.0 P2 — the portable
// cross-backend password hash; BCryptHash is Go-only). iterations clamps to
// [1000, 2^24]; the stored value wins at verify time, so raising the default
// later needs no migration.
func Pbkdf2Hash(password string, iterations int64) string {
	it := iterations
	if it < pbkdf2MinIterations {
		it = pbkdf2MinIterations
	}
	if it > pbkdf2MaxIterations {
		it = pbkdf2MaxIterations
	}
	salt := make([]byte, pbkdf2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		// crypto/rand never fails on the supported platforms.
		panic("crypto: entropy source unavailable")
	}
	key := pbkdf2.Key([]byte(password), salt, int(it), pbkdf2KeyLen, sha256.New)
	return fmt.Sprintf("pbkdf2$sha256$%d$%s$%s", it,
		hex.EncodeToString(salt), hex.EncodeToString(key))
}

// Pbkdf2Compare verifies password against a Pbkdf2Hash envelope: parses the
// iteration count and salt, recomputes the digest, and compares in constant
// time. Malformed hashes and out-of-range iteration counts fail closed.
func Pbkdf2Compare(password, hash string) bool {
	parts := strings.Split(hash, "$")
	if len(parts) != 5 || parts[0] != "pbkdf2" || parts[1] != "sha256" {
		return false
	}
	it, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil || it < pbkdf2MinIterations || it > pbkdf2MaxIterations {
		return false
	}
	salt, err1 := hex.DecodeString(parts[3])
	want, err2 := hex.DecodeString(parts[4])
	if err1 != nil || err2 != nil || len(salt) != pbkdf2SaltLen || len(want) != pbkdf2KeyLen {
		return false
	}
	got := pbkdf2.Key([]byte(password), salt, int(it), pbkdf2KeyLen, sha256.New)
	return subtle.ConstantTimeCompare(got, want) == 1
}

func BCryptCompare(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func RandomBytes(n int64) (string, error) {
	if n <= 0 {
		return "", nil
	}
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf), nil
}

func RandomToken(n int64) (string, error) {
	if n <= 0 {
		return "", nil
	}
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
