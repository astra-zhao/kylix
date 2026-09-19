package boot

import (
	"net/http/httptest"
	"testing"
)

func resetSecurityHooks() {
	authValidator = nil
	rolesProvider = nil
}

func newSecurityRequest(token string) *Request {
	httpReq := httptest.NewRequest("GET", "/", nil)
	if token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+token)
	}
	return &Request{Request: httpReq}
}

func TestEnforceAuth_MissingHeader(t *testing.T) {
	resetSecurityHooks()
	defer resetSecurityHooks()
	req := newSecurityRequest("")
	r := EnforceAuth(req)
	if r == nil || r.Status != 401 {
		t.Fatalf("expected 401, got %#v", r)
	}
}

func TestEnforceAuth_NoValidator(t *testing.T) {
	resetSecurityHooks()
	defer resetSecurityHooks()
	req := newSecurityRequest("anything")
	r := EnforceAuth(req)
	if r == nil || r.Status != 401 {
		t.Fatalf("expected 401, got %#v", r)
	}
}

func TestEnforceAuth_Success(t *testing.T) {
	resetSecurityHooks()
	defer resetSecurityHooks()
	RegisterAuthValidator(func(token string) (string, bool) {
		if token == "good" {
			return "alice", true
		}
		return "", false
	})
	req := newSecurityRequest("good")
	if r := EnforceAuth(req); r != nil {
		t.Fatalf("expected nil, got %#v", r)
	}
	if req.User != "alice" {
		t.Fatalf("expected alice, got %q", req.User)
	}
}

func TestEnforceRole_Forbidden(t *testing.T) {
	resetSecurityHooks()
	defer resetSecurityHooks()
	RegisterAuthValidator(func(token string) (string, bool) { return "alice", true })
	RegisterRolesProvider(func(user string) []string { return []string{"user"} })
	req := newSecurityRequest("good")
	r := EnforceRole(req, "admin")
	if r == nil || r.Status != 403 {
		t.Fatalf("expected 403, got %#v", r)
	}
}

func TestEnforceRole_Granted(t *testing.T) {
	resetSecurityHooks()
	defer resetSecurityHooks()
	RegisterAuthValidator(func(token string) (string, bool) { return "alice", true })
	RegisterRolesProvider(func(user string) []string { return []string{"admin"} })
	req := newSecurityRequest("good")
	if r := EnforceRole(req, "admin"); r != nil {
		t.Fatalf("expected nil, got %#v", r)
	}
}

func TestEnforceAuth_SessionFirst(t *testing.T) {
	resetSecurityHooks()
	defer resetSecurityHooks()
	// No Bearer header, no validator registered — a live session with a
	// non-empty __user authenticates (the HTML form-login path).
	req := newSecurityRequest("")
	req.Session = &Session{data: map[string]string{
		sessionUserKey:  "alice",
		sessionRolesKey: "admin,editor",
	}}
	if r := EnforceAuth(req); r != nil {
		t.Fatalf("expected nil, got %#v", r)
	}
	if req.User != "alice" {
		t.Fatalf("expected alice, got %q", req.User)
	}
	if len(req.Roles) != 2 || req.Roles[0] != "admin" || req.Roles[1] != "editor" {
		t.Fatalf("expected [admin editor], got %v", req.Roles)
	}
}

func TestEnforceAuth_EmptySessionUserFallsBack(t *testing.T) {
	resetSecurityHooks()
	defer resetSecurityHooks()
	// A session without __user is anonymous — the Bearer path (and its 401
	// when no header is present) must still run.
	req := newSecurityRequest("")
	req.Session = &Session{data: map[string]string{"other": "value"}}
	if r := EnforceAuth(req); r == nil || r.Status != 401 {
		t.Fatalf("expected 401, got %#v", r)
	}
	// Bearer path still works with a session present (API client logged in
	// separately in the same process).
	RegisterAuthValidator(func(token string) (string, bool) { return "bob", true })
	req2 := newSecurityRequest("good")
	req2.Session = &Session{}
	if r := EnforceAuth(req2); r != nil {
		t.Fatalf("expected nil, got %#v", r)
	}
	if req2.User != "bob" {
		t.Fatalf("expected bob, got %q", req2.User)
	}
}

func TestEnforceRole_SessionRoles(t *testing.T) {
	resetSecurityHooks()
	defer resetSecurityHooks()
	// Roles come from the session when session-first auth ran — no
	// rolesProvider needed for the HTML path.
	req := newSecurityRequest("")
	req.Session = &Session{data: map[string]string{
		sessionUserKey:  "alice",
		sessionRolesKey: "viewer,",
	}}
	if r := EnforceRole(req, "viewer"); r != nil {
		t.Fatalf("expected nil, got %#v", r)
	}
	if r := EnforceRole(req, "admin"); r == nil || r.Status != 403 {
		t.Fatalf("expected 403, got %#v", r)
	}
}
