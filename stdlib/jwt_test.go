package stdlib

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"kylix/pkg/boot"
)

func TestJwt_SignVerifyRoundTrip(t *testing.T) {
	token, err := JwtSign("secret", "alice", 3600, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(strings.Split(token, ".")) != 3 {
		t.Fatalf("expected 3 parts, got %q", token)
	}
	claims, ok := JwtVerify("secret", token)
	if !ok {
		t.Fatal("expected valid token")
	}
	if JwtSubject(claims) != "alice" {
		t.Fatalf("sub=%q, want alice", JwtSubject(claims))
	}
}

func TestJwt_WrongSecretRejected(t *testing.T) {
	token, _ := JwtSign("secret", "alice", 3600, nil)
	_, ok := JwtVerify("wrong-secret", token)
	if ok {
		t.Fatal("expected rejection with wrong secret")
	}
}

func TestJwt_ExpiredRejected(t *testing.T) {
	token, err := JwtSign("secret", "alice", 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, ok := JwtVerify("secret", token)
	if !ok {
		t.Fatal("fresh token should be valid")
	}
	time.Sleep(2 * time.Second)
	_, ok = JwtVerify("secret", token)
	if ok {
		t.Fatal("expired token should be rejected")
	}
}

func TestJwt_NoExpiry(t *testing.T) {
	token, err := JwtSign("secret", "bob", 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	claims, ok := JwtVerify("secret", token)
	if !ok {
		t.Fatal("token with no expiry should be valid")
	}
	if _, hasExp := claims["exp"]; hasExp {
		t.Fatal("token with expiresIn=0 should not have exp claim")
	}
}

func TestJwt_ExtraClaims(t *testing.T) {
	extra := map[string]interface{}{"role": "admin", "org": "acme"}
	token, err := JwtSign("secret", "charlie", 3600, extra)
	if err != nil {
		t.Fatal(err)
	}
	claims, ok := JwtVerify("secret", token)
	if !ok {
		t.Fatal("expected valid token")
	}
	if JwtGetString(claims, "role") != "admin" {
		t.Fatalf("role=%q, want admin", JwtGetString(claims, "role"))
	}
	if JwtGetString(claims, "org") != "acme" {
		t.Fatalf("org=%q, want acme", JwtGetString(claims, "org"))
	}
}

func TestJwt_MalformedTokenRejected(t *testing.T) {
	for _, tc := range []string{"", "abc", "a.b", "a.b.c.d"} {
		_, ok := JwtVerify("secret", tc)
		if ok {
			t.Fatalf("malformed token %q should be rejected", tc)
		}
	}
}

func TestJwt_GetInt(t *testing.T) {
	extra := map[string]interface{}{"age": float64(30)}
	token, _ := JwtSign("secret", "dave", 3600, extra)
	claims, _ := JwtVerify("secret", token)
	if JwtGetInt(claims, "age") != 30 {
		t.Fatalf("age=%d, want 30", JwtGetInt(claims, "age"))
	}
}

func TestJwt_BootRegisterJwtAuth(t *testing.T) {
	defer boot.RegisterAuthValidator(nil) // reset after test

	BootRegisterJwtAuth("test-secret")
	token, _ := JwtSign("test-secret", "user42", 3600, nil)

	// Verify via boot.EnforceAuth — calls the validator we just registered.
	httpReq := httptest.NewRequest("GET", "/", nil)
	httpReq.Header.Set("Authorization", "Bearer "+token)
	req := &boot.Request{Request: httpReq}

	if resp := boot.EnforceAuth(req); resp != nil {
		t.Fatalf("expected auth ok (nil), got status %d", resp.Status)
	}
	if req.User != "user42" {
		t.Fatalf("req.User=%q, want user42", req.User)
	}

	// Invalid token should 401.
	badHTTP := httptest.NewRequest("GET", "/", nil)
	badHTTP.Header.Set("Authorization", "Bearer bad.token.here")
	if resp := boot.EnforceAuth(&boot.Request{Request: badHTTP}); resp == nil {
		t.Fatal("expected 401 for invalid JWT")
	}
}
