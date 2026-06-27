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
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
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
