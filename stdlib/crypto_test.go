package stdlib

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestCrypto_Sha256(t *testing.T) {
	got := Sha256("abc")
	want := "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if got != want {
		t.Fatalf("Sha256=%q, want %q", got, want)
	}
}

func TestCrypto_Sha512Length(t *testing.T) {
	if len(Sha512("")) != 128 {
		t.Fatalf("Sha512 hex length=%d, want 128", len(Sha512("")))
	}
}

func TestCrypto_Md5(t *testing.T) {
	got := Md5("abc")
	want := "900150983cd24fb0d6963f7d28e17f72"
	if got != want {
		t.Fatalf("Md5=%q, want %q", got, want)
	}
}

func TestCrypto_HmacSha256(t *testing.T) {
	got := HmacSha256("key", "The quick brown fox jumps over the lazy dog")
	want := "f7bc83f430538424b13298e6aa6fb143ef4d59a14946175997479dbc2d1a3cd8"
	if got != want {
		t.Fatalf("HmacSha256=%q, want %q", got, want)
	}
}

func TestCrypto_AesRoundTrip(t *testing.T) {
	plain := "secret message"
	enc, err := AesEncrypt("password", plain)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := base64.StdEncoding.DecodeString(enc); err != nil {
		t.Fatalf("AES ciphertext should be base64: %v", err)
	}
	dec, err := AesDecrypt("password", enc)
	if err != nil {
		t.Fatal(err)
	}
	if dec != plain {
		t.Fatalf("round trip mismatch: got %q", dec)
	}
}

func TestCrypto_AesWrongKey(t *testing.T) {
	enc, _ := AesEncrypt("k1", "hi")
	if _, err := AesDecrypt("k2", enc); err == nil {
		t.Fatal("expected decryption error with wrong key")
	}
}

func TestCrypto_BCryptRoundTrip(t *testing.T) {
	h, err := BCryptHash("hunter2", 4)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(h, "$2") {
		t.Fatalf("BCrypt hash should start with $2, got %q", h)
	}
	if !BCryptCompare("hunter2", h) {
		t.Fatal("BCryptCompare should accept the original password")
	}
	if BCryptCompare("wrong", h) {
		t.Fatal("BCryptCompare should reject a wrong password")
	}
}

func TestCrypto_RandomBytesUnique(t *testing.T) {
	a, err := RandomBytes(16)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := RandomBytes(16)
	if a == "" || a == b {
		t.Fatal("RandomBytes should return unique non-empty values")
	}
}

func TestCrypto_RandomTokenURLSafe(t *testing.T) {
	tok, err := RandomToken(24)
	if err != nil {
		t.Fatal(err)
	}
	if strings.ContainsAny(tok, "+/=") {
		t.Fatalf("RandomToken should be URL-safe, got %q", tok)
	}
}
