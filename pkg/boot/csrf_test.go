// csrf_test.go — CSRF middleware tests (v0.9.0 P1).
package boot

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// csrfRouter wires Sessions + CSRF with a per-request hook.
func csrfRouter(hook func(req *Request) *Response) *Router {
	r := NewRouter()
	r.Use(Sessions())
	r.Use(CSRF())
	r.GET("/form", func(req *Request) *Response {
		return Text(200, "token:"+req.CSRFToken())
	})
	r.POST("/submit", hook)
	return r
}

func doCsrf(r *Router, method, target, body string, cookies []*http.Cookie) (int, string) {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if method == "POST" && body != "" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code, w.Body.String()
}

func TestCSRF_GetExempt(t *testing.T) {
	r := csrfRouter(func(req *Request) *Response { return Text(200, "ok") })
	code, _ := doCsrf(r, "POST", "/submit", "", nil)
	if code != 403 {
		t.Fatalf("POST without session/token: got %d, want 403", code)
	}
	code, _ = doCsrf(r, "GET", "/form", "", nil)
	if code != 200 {
		t.Fatalf("GET must be exempt: got %d", code)
	}
}

func TestCSRF_TokenRoundtrip(t *testing.T) {
	r := csrfRouter(func(req *Request) *Response { return Text(200, "accepted") })

	// Visit /form: session created + token issued; capture cookie + token.
	req := httptest.NewRequest("GET", "/form", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var sid *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == SessionCookieName {
			sid = c
		}
	}
	if sid == nil {
		t.Fatal("no session cookie issued")
	}
	token := strings.TrimPrefix(w.Body.String(), "token:")
	if len(token) != 32 {
		t.Fatalf("token = %q, want 32 hex chars", token)
	}

	// Same session, POST with token in the form → accepted.
	code, body := doCsrf(r, "POST", "/submit", CSRFFormField+"="+token, []*http.Cookie{sid})
	if code != 200 || body != "accepted" {
		t.Fatalf("valid token rejected: %d %q", code, body)
	}

	// Token also accepted via header.
	code, _ = doCsrf(r, "POST", "/submit", "", []*http.Cookie{sid})
	if code != 403 {
		t.Fatalf("headerless+formless POST must 403, got %d", code)
	}
}

func TestCSRF_WrongAndForeignToken(t *testing.T) {
	r := csrfRouter(func(req *Request) *Response { return Text(200, "accepted") })

	// Establish a session and a token.
	req := httptest.NewRequest("GET", "/form", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var sid *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == SessionCookieName {
			sid = c
		}
	}
	token := strings.TrimPrefix(w.Body.String(), "token:")

	// Right session, wrong token → 403.
	code, _ := doCsrf(r, "POST", "/submit", CSRFFormField+"="+"00000000000000000000000000000000", []*http.Cookie{sid})
	if code != 403 {
		t.Fatalf("wrong token must 403, got %d", code)
	}

	// Token from a DIFFERENT session must not validate (session-bound).
	other := NewSessionStore()
	s2 := other.Create(false)
	s2.Set(csrfSessionKey, "ffffffffffffffffffffffffffffffff")
	code, _ = doCsrf(r, "POST", "/submit", CSRFFormField+"=ffffffffffffffffffffffffffffffff", []*http.Cookie{sid})
	if code != 403 {
		t.Fatalf("foreign token must 403, got %d", code)
	}
	_ = token
}

func TestCSRF_TokenStableAcrossRequests(t *testing.T) {
	r := NewRouter()
	r.Use(Sessions())
	r.GET("/f", func(req *Request) *Response {
		return Text(200, req.CSRFToken())
	})
	req := httptest.NewRequest("GET", "/f", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	tok1 := w.Body.String()
	var sid *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == SessionCookieName {
			sid = c
		}
	}

	req2 := httptest.NewRequest("GET", "/f", nil)
	req2.AddCookie(sid)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Body.String() != tok1 {
		t.Errorf("token changed between requests: %q vs %q", tok1, w2.Body.String())
	}
}
