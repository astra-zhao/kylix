// security.go — Process-wide auth + role hooks for KylixBoot annotation guards.
//
// The compiler emits route closures like:
//
//	if r := boot.EnforceAuth(req); r != nil { return r }
//	if r := boot.EnforceRole(req, "admin"); r != nil { return r }
//
// before invoking the controller method. Applications register validators
// once at startup; missing hooks cause the guards to reject with 401/403.
package boot

import "strings"

// Session keys the session-first auth model reads (v0.10.0 P2). The LLVM
// backend hardcodes the same literals in emitBootEnforceAuthBody /
// BootEnforceRole — keep the three in sync. Applications write them at
// login (see the session-first contract in docs/ADMIN_PLATFORM.md).
const (
	sessionUserKey  = "__user"  // logged-in username; empty/absent = anonymous
	sessionRolesKey = "__roles" // comma-separated role names
)

var (
	authValidator func(token string) (user string, ok bool)
	rolesProvider func(user string) []string
)

// RegisterAuthValidator sets the global token → user validator used by
// [Authenticated] and [Role] annotation guards.
func RegisterAuthValidator(v func(token string) (string, bool)) { authValidator = v }

// RegisterRolesProvider sets the global user → roles provider used by
// [Role] annotation guards.
func RegisterRolesProvider(p func(user string) []string) { rolesProvider = p }

// EnforceAuth authenticates the request, session-first (v0.10.0 P2): a live
// session with a non-empty __user key passes (req.User/Roles populated from
// __user/__roles) — the HTML form-login path. Otherwise the Bearer/JWT API
// path runs unchanged: the Authorization header is verified with the
// registered authValidator. Returns a 401 Response on any failure; nil on
// success. The LLVM backend mirrors this order in emitBootEnforceAuthBody.
func EnforceAuth(req *Request) *Response {
	if req == nil || req.Request == nil {
		return JSON(401, map[string]string{"error": "unauthorized"})
	}
	if req.Session != nil {
		if user := req.Session.Get(sessionUserKey); user != "" {
			req.User = user
			req.Roles = splitSessionRoles(req.Session.Get(sessionRolesKey))
			return nil
		}
	}
	header := req.Header("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return JSON(401, map[string]string{"error": "missing or invalid Authorization header"})
	}
	token := strings.TrimPrefix(header, "Bearer ")
	if authValidator == nil {
		return JSON(401, map[string]string{"error": "auth validator not registered"})
	}
	user, ok := authValidator(token)
	if !ok {
		return JSON(401, map[string]string{"error": "invalid token"})
	}
	req.User = user
	if rolesProvider != nil {
		req.Roles = rolesProvider(user)
	}
	return nil
}

// splitSessionRoles parses the comma-separated __roles session value.
func splitSessionRoles(v string) []string {
	var roles []string
	for _, r := range strings.Split(v, ",") {
		if r = strings.TrimSpace(r); r != "" {
			roles = append(roles, r)
		}
	}
	return roles
}

// EnforceRole returns a 403 Response unless req.User has the given role.
// Callers should invoke EnforceAuth first; if req.User is empty, EnforceAuth
// is called transparently and any 401 propagates.
func EnforceRole(req *Request, role string) *Response {
	if req == nil {
		return JSON(403, map[string]string{"error": "forbidden"})
	}
	if req.User == "" {
		if r := EnforceAuth(req); r != nil {
			return r
		}
	}
	for _, r := range req.Roles {
		if r == role {
			return nil
		}
	}
	if rolesProvider != nil && len(req.Roles) == 0 {
		req.Roles = rolesProvider(req.User)
		for _, r := range req.Roles {
			if r == role {
				return nil
			}
		}
	}
	return JSON(403, map[string]string{"error": "missing required role: " + role})
}
