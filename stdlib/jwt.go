// jwt.go — Kylix stdlib JWT module.
//
// Implements HS256 (HMAC-SHA256) JWT signing and verification using only the
// Go standard library. No external dependencies.
//
// Supported functions:
//   JwtSign(secret, subject, expiresIn, extraClaims) → token string
//   JwtVerify(secret, token) → (claims map[string]interface{}, ok bool)
//   JwtSubject(claims) → string
//   JwtGetString(claims, key) → string
//   JwtGetInt(claims, key) → int64
package stdlib

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"kylix/pkg/boot"
	"strings"
	"time"
)

// JwtSign creates a signed HS256 JWT token.
//
// Parameters:
//   - secret: signing secret
//   - subject: "sub" claim (usually user ID or username)
//   - expiresIn: token lifetime in seconds (0 = no expiry)
//   - extraClaims: additional claims merged into the payload
func JwtSign(secret, subject string, expiresIn int64, extraClaims map[string]interface{}) (string, error) {
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	hb, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	headerEnc := base64.RawURLEncoding.EncodeToString(hb)

	payload := make(map[string]interface{})
	for k, v := range extraClaims {
		payload[k] = v
	}
	payload["sub"] = subject
	payload["iat"] = time.Now().Unix()
	if expiresIn > 0 {
		payload["exp"] = time.Now().Unix() + expiresIn
	}
	pb, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	payloadEnc := base64.RawURLEncoding.EncodeToString(pb)

	signingInput := headerEnc + "." + payloadEnc
	sig := jwtHmacSha256(secret, signingInput)
	return signingInput + "." + sig, nil
}

// JwtVerify validates an HS256 JWT and returns the decoded claims.
// Returns (nil, false) if the token is invalid, malformed, or expired.
func JwtVerify(secret, token string) (map[string]interface{}, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, false
	}
	signingInput := parts[0] + "." + parts[1]
	expected := jwtHmacSha256(secret, signingInput)
	if !hmac.Equal([]byte(parts[2]), []byte(expected)) {
		return nil, false
	}
	pb, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, false
	}
	var claims map[string]interface{}
	if err := json.Unmarshal(pb, &claims); err != nil {
		return nil, false
	}
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, false // expired
		}
	}
	return claims, true
}

// JwtSubject returns the "sub" claim from decoded JWT claims.
func JwtSubject(claims map[string]interface{}) string {
	if claims == nil {
		return ""
	}
	s, _ := claims["sub"].(string)
	return s
}

// JwtGetString returns a string claim by key from decoded JWT claims.
func JwtGetString(claims map[string]interface{}, key string) string {
	if claims == nil {
		return ""
	}
	s, _ := claims[key].(string)
	return s
}

// JwtGetInt returns a numeric claim by key as int64 from decoded JWT claims.
func JwtGetInt(claims map[string]interface{}, key string) int64 {
	if claims == nil {
		return 0
	}
	f, _ := claims[key].(float64)
	return int64(f)
}

func jwtHmacSha256(secret, data string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// BootRegisterJwtAuth registers a JWT-based auth validator with KylixBoot.
// After calling this once with your secret, all routes annotated with
// [Authenticated] will automatically verify Bearer JWT tokens.
//
// The "sub" claim becomes req.User; all claims are stored in req.JwtClaims.
//
// Usage in Kylix startup code:
//
//	BootRegisterJwtAuth('my-secret-key');
func BootRegisterJwtAuth(secret string) {
	boot.RegisterAuthValidator(func(token string) (string, bool) {
		claims, ok := JwtVerify(secret, token)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("%v", JwtSubject(claims)), true
	})
}
