// csrf.go — Cross-Site Request Forgery protection (v0.9.0 P1).
//
// Synchronizer-token pattern on top of the Session middleware: the token
// lives in the session under a fixed key, is rendered into forms as a
// hidden field (or sent as the X-CSRF-Token header), and is verified on
// every unsafe method (POST/PUT/PATCH/DELETE).
//
// Install after Sessions():
//
//	boot.Use(boot.Sessions())
//	boot.Use(boot.CSRF())
//
// In a template:
//
//	<input type="hidden" name="_csrf" value="{{csrf}}">
//
// Handlers obtain the token with req.CSRFToken().
package boot

import (
	"crypto/subtle"
	"net/http"
)

// CSRFFormField is the form field name carrying the token.
const CSRFFormField = "_csrf"

// CSRFHeader is the request header carrying the token (AJAX flows).
const CSRFHeader = "X-CSRF-Token"

// CSRFToken returns (generating on first call) the session's CSRF token.
// Returns "" when the Session middleware is not installed.
func (r *Request) CSRFToken() string { return r.SessionCSRFToken() }

// CSRF returns middleware that rejects unsafe requests whose token does
// not match the session's. A request with no session token at all is also
// rejected — the token must have been issued via CSRFToken() first.
func CSRF() Middleware {
	return func(next Handler) Handler {
		return func(req *Request) *Response {
			switch req.Request.Method {
			case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
				return next(req)
			}
			sessionTok := req.Session.Get(csrfSessionKey)
			if sessionTok == "" {
				return Text(403, "CSRF token missing: render req.CSRFToken() into the form first")
			}
			reqTok := req.Header(CSRFHeader)
			if reqTok == "" {
				reqTok = req.Form(CSRFFormField)
			}
			if reqTok == "" || subtle.ConstantTimeCompare([]byte(reqTok), []byte(sessionTok)) != 1 {
				return Text(403, "CSRF token mismatch")
			}
			return next(req)
		}
	}
}
