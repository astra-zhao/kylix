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

// EnforceAuth verifies a Bearer token from the request and populates
// req.User on success. Returns a 401 Response on any failure; nil on success.
func EnforceAuth(req *Request) *Response {
	if req == nil || req.Request == nil {
		return JSON(401, map[string]string{"error": "unauthorized"})
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
