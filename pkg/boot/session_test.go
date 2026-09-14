// session_test.go — Session middleware tests (v0.9.0 P1).
package boot

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// runWithSession builds a router with the Session middleware and a handler
// that echoes session state, then performs one request.
func runWithSession(t *testing.T, store *SessionStore, cookie string, fn func(req *Request) *Response) (int, string, []*http.Cookie) {
	t.Helper()
	r := NewRouter()
	r.Use(SessionWithStore(store))
	r.GET("/s", fn)
	req := httptest.NewRequest("GET", "/s", nil)
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: cookie})
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code, w.Body.String(), w.Result().Cookies()
}

func TestSession_CreateAndReuse(t *testing.T) {
	store := NewSessionStore()
	var sid string

	// 1st request: no cookie → new session + Set-Cookie.
	code, body, cookies := runWithSession(t, store, "", func(req *Request) *Response {
		if req.Session == nil {
			t.Fatal("Session should be populated")
		}
		if !req.Session.New {
			t.Error("first visit should create a new session")
		}
		sid = req.Session.ID
		return Text(200, "new:"+req.SessionGet("uid"))
	})
	if code != 200 || body != "new:" {
		t.Fatalf("got %d %q", code, body)
	}
	var setCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == SessionCookieName {
			setCookie = c
		}
	}
	if setCookie == nil || setCookie.Value != sid {
		t.Fatalf("Set-Cookie %v missing or wrong value", cookies)
	}
	if !setCookie.HttpOnly {
		t.Error("session cookie must be HttpOnly")
	}

	// 2nd request: present the cookie → same session, no new Set-Cookie.
	code, body, cookies = runWithSession(t, store, sid, func(req *Request) *Response {
		if req.Session.New {
			t.Error("existing session must be reused")
		}
		req.SessionSet("uid", "42")
		return Text(200, "uid:"+req.SessionGet("uid"))
	})
	if code != 200 || body != "uid:42" {
		t.Fatalf("got %d %q", code, body)
	}
	for _, c := range cookies {
		if c.Name == SessionCookieName {
			t.Errorf("unexpected re-send of session cookie: %v", c)
		}
	}

	// 3rd request: value persisted across requests.
	_, body, _ = runWithSession(t, store, sid, func(req *Request) *Response {
		return Text(200, "uid:"+req.SessionGet("uid"))
	})
	if body != "uid:42" {
		t.Fatalf("session value lost: %q", body)
	}
}

func TestSession_Destroy(t *testing.T) {
	store := NewSessionStore()
	var sid string
	_, _, _ = runWithSession(t, store, "", func(req *Request) *Response {
		sid = req.Session.ID
		return Text(200, "ok")
	})

	code, _, cookies := runWithSession(t, store, sid, func(req *Request) *Response {
		req.SessionDestroy()
		return Text(200, "bye")
	})
	if code != 200 {
		t.Fatalf("got %d", code)
	}
	var drop *http.Cookie
	for _, c := range cookies {
		if c.Name == SessionCookieName {
			drop = c
		}
	}
	if drop == nil || drop.MaxAge != -1 {
		t.Fatalf("expected expired Set-Cookie, got %v", cookies)
	}
	if store.Get(sid) != nil {
		t.Error("session should be removed from the store")
	}
	// Old cookie no longer maps to a session.
	_, _, _ = runWithSession(t, store, sid, func(req *Request) *Response {
		if !req.Session.New {
			t.Error("destroyed sid must not resolve to a session")
		}
		return Text(200, "ok")
	})
}

func TestSession_Remember(t *testing.T) {
	store := NewSessionStore()
	var sid string
	code, _, cookies := runWithSession(t, store, "", func(req *Request) *Response {
		sid = req.Session.ID
		req.SessionMarkRemember()
		return Text(200, "ok")
	})
	if code != 200 {
		t.Fatalf("got %d", code)
	}
	var c *http.Cookie
	for _, ck := range cookies {
		if ck.Name == SessionCookieName {
			c = ck
		}
	}
	if c == nil {
		t.Fatal("remember flip must re-send the cookie")
	}
	if !strings.Contains(c.Value, sid) && c.Value != sid {
		t.Errorf("cookie value %q != sid %q", c.Value, sid)
	}
	if c.MaxAge < int((24*time.Hour - time.Minute).Seconds()) {
		t.Errorf("remember cookie Max-Age too short: %d", c.MaxAge)
	}
}

func TestSession_Expiry(t *testing.T) {
	store := NewSessionStore()
	store.TTL = 10 * time.Millisecond
	var sid string
	_, _, _ = runWithSession(t, store, "", func(req *Request) *Response {
		sid = req.Session.ID
		return Text(200, "ok")
	})
	time.Sleep(20 * time.Millisecond)
	if store.Get(sid) != nil {
		t.Error("expired session must not resolve")
	}
	_, _, _ = runWithSession(t, store, sid, func(req *Request) *Response {
		if !req.Session.New {
			t.Error("expired sid must create a fresh session")
		}
		return Text(200, "ok")
	})
}

func TestSession_NilSafeAccessors(t *testing.T) {
	// Handlers on a router without Session() middleware must not panic.
	r := NewRouter()
	r.GET("/s", func(req *Request) *Response {
		return Text(200, "v:"+req.SessionGet("k"))
	})
	req := httptest.NewRequest("GET", "/s", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 || w.Body.String() != "v:" {
		t.Fatalf("got %d %q", w.Code, w.Body.String())
	}
}
