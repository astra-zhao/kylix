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
