// Package boot — KylixBoot framework runtime.
//
// KylixBoot is a Spring Boot-style web framework for Kylix.
// This package provides the runtime that powers declarative annotations
// like [Controller], [Get], [Inject], [Value], etc.
//
// In v0.3.1 (initial alpha), boot exposes a programmatic API:
//
//	boot.GET("/users", handleListUsers)
//	boot.POST("/users", handleCreateUser)
//	boot.Run(8080)
//
// Future versions will add annotation-driven auto-registration where
// classes tagged with [Controller] and methods tagged with [Get]/[Post]
// are automatically registered at startup via compile-time code generation.
package boot

import (
	"encoding/json"
	"net/http"
	"net/url"
)

// ===== Request / Response =====

// Request wraps an HTTP request with helpers for params, query, body parsing.
type Request struct {
	Request *http.Request
	Params  map[string]string
	body    []byte
	User    string
	Roles   []string
}

// Param returns a URL path parameter value (e.g. "/users/:id" → req.Param("id")).
func (r *Request) Param(name string) string {
	if r.Params == nil {
		return ""
	}
	return r.Params[name]
}

// Query returns a URL query string value (?name=value).
func (r *Request) Query(name string) string {
	return r.Request.URL.Query().Get(name)
}

// Header returns a request header value.
func (r *Request) Header(name string) string {
	return r.Request.Header.Get(name)
}

// Body returns the raw request body bytes.
func (r *Request) Body() []byte {
	if r.body != nil {
		return r.body
	}
	if r.Request.Body == nil {
		return nil
	}
	buf := make([]byte, 0, 512)
	tmp := make([]byte, 512)
	for {
		n, err := r.Request.Body.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if err != nil {
			break
		}
	}
	r.body = buf
	return buf
}

// JSON parses the request body as JSON into the given pointer.
func (r *Request) JSON(out interface{}) error {
	return json.Unmarshal(r.Body(), out)
}

// Form returns a value from an application/x-www-form-urlencoded body
// (HTML <form method="POST"> submissions). Falls back to the URL query
// string when the body is not form-encoded or the key is absent.
// v0.7.0 P2.
func (r *Request) Form(name string) string {
	body := string(r.Body())
	if vals, err := url.ParseQuery(body); err == nil {
		if v := vals.Get(name); v != "" {
			return v
		}
	}
	return r.Query(name)
}

// Cookie returns a request cookie value ("Cookie: a=1; b=2" headers),
// or "" when the cookie is absent. v0.7.0 P2.
func (r *Request) Cookie(name string) string {
	if r.Request == nil {
		return ""
	}
	c, err := r.Request.Cookie(name)
	if err != nil {
		return ""
	}
	return c.Value
}

// Response is a builder-style HTTP response.
type Response struct {
	Status      int
	Headers     map[string]string
	Body        string
	ContentType string
	Cookies     []string // raw Set-Cookie values (v0.7.0 P2)
}

// NewResponse creates a basic Response with given status and body.
func NewResponse(status int, body string) *Response {
	return &Response{Status: status, Body: body, Headers: map[string]string{}}
}

// JSON creates a JSON response from any serializable value.
func JSON(status int, value interface{}) *Response {
	data, err := json.Marshal(value)
	if err != nil {
		return &Response{Status: 500, Body: `{"error":"marshal failed"}`, ContentType: "application/json"}
	}
	return &Response{
		Status:      status,
		Body:        string(data),
		ContentType: "application/json",
		Headers:     map[string]string{},
	}
}

// Text creates a plain text response.
func Text(status int, body string) *Response {
	return &Response{
		Status:      status,
		Body:        body,
		ContentType: "text/plain; charset=utf-8",
		Headers:     map[string]string{},
	}
}

// HTML creates an HTML response.
func HTML(status int, body string) *Response {
	return &Response{
		Status:      status,
		Body:        body,
		ContentType: "text/html; charset=utf-8",
		Headers:     map[string]string{},
	}
}

// WithHeader adds a response header (fluent API).
func (r *Response) WithHeader(key, value string) *Response {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

// Send mutates the response body as plain text.
func (r *Response) Send(body string) *Response {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Body = body
	if r.ContentType == "" {
		r.ContentType = "text/plain; charset=utf-8"
	}
	return r
}

// StatusCode mutates the response status code.
func (r *Response) StatusCode(status int) *Response {
	r.Status = status
	return r
}

// Html mutates the response body to an HTML page (v0.7.0 P2).
// Typical use with the template engine:
//
//	Result := resp.Html(eng.RenderString('home'));
func (r *Response) Html(body string) *Response {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Body = body
	r.ContentType = "text/html; charset=utf-8"
	return r
}

// WithCookie appends a Set-Cookie header (fluent API, v0.7.0 P2).
// Sent as "name=value; Path=/".
func (r *Response) WithCookie(name, value string) *Response {
	r.Cookies = append(r.Cookies, name+"="+value+"; Path=/")
	return r
}

// Redirect returns a 302 response pointing at url (fluent API, v0.7.0 P3).
// Typical handler use: Result := res.Redirect('/login');
func (r *Response) Redirect(url string) *Response {
	r.Status = 302
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers["Location"] = url
	return r
}

// ===== Handler types =====

// Handler is the canonical handler signature for KylixBoot routes.
type Handler func(*Request) *Response

// Middleware wraps a Handler with pre/post logic.
type Middleware func(Handler) Handler
