// Package boot — KylixBoot framework runtime.
//
// KylixBoot is a Spring Boot-style web framework for Kylix.
// This package provides the runtime that powers declarative annotations
// like [Controller], [Get], [Inject], [Value], etc.
//
// In v3.1.0 (initial alpha), boot exposes a programmatic API:
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

// Response is a builder-style HTTP response.
type Response struct {
	Status      int
	Headers     map[string]string
	Body        string
	ContentType string
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

// ===== Handler types =====

// Handler is the canonical handler signature for KylixBoot routes.
type Handler func(*Request) *Response

// Middleware wraps a Handler with pre/post logic.
type Middleware func(Handler) Handler
