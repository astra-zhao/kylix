// router.go — KylixBoot HTTP router with path parameters and middleware.
package boot

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// Route represents a registered handler.
type Route struct {
	Method      string
	Pattern     string
	Segments    []routeSegment
	Handler     Handler
	Middlewares []Middleware
}

type routeSegment struct {
	literal string // empty if param
	param   string // empty if literal
}

// parseSegments breaks "/users/:id/posts/:pid" into segments.
func parseSegments(pattern string) []routeSegment {
	parts := strings.Split(strings.Trim(pattern, "/"), "/")
	if len(parts) == 1 && parts[0] == "" {
		return nil
	}
	out := make([]routeSegment, len(parts))
	for i, p := range parts {
		if strings.HasPrefix(p, ":") {
			out[i] = routeSegment{param: p[1:]}
		} else {
			out[i] = routeSegment{literal: p}
		}
	}
	return out
}

// match attempts to match an incoming path to this route, extracting params.
func (r *Route) match(method, path string) (map[string]string, bool) {
	if !strings.EqualFold(r.Method, method) {
		return nil, false
	}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 1 && parts[0] == "" {
		parts = nil
	}
	if len(parts) != len(r.Segments) {
		return nil, false
	}
	params := map[string]string{}
	for i, seg := range r.Segments {
		if seg.param != "" {
			params[seg.param] = parts[i]
		} else if seg.literal != parts[i] {
			return nil, false
		}
	}
	return params, true
}

// Router is the registry of routes + middleware chain.
type Router struct {
	mu          sync.RWMutex
	routes      []*Route
	middlewares []Middleware
	notFound    Handler
}

// NewRouter creates an empty router.
func NewRouter() *Router {
	return &Router{
		notFound: func(req *Request) *Response {
			return Text(404, "404 Not Found: "+req.Request.URL.Path)
		},
	}
}

// Use adds a global middleware to the chain.
func (r *Router) Use(mw Middleware) {
	r.mu.Lock()
	r.middlewares = append(r.middlewares, mw)
	r.mu.Unlock()
}

// Handle registers a route for the given method + pattern.
func (r *Router) Handle(method, pattern string, h Handler, mws ...Middleware) {
	r.mu.Lock()
	r.routes = append(r.routes, &Route{
		Method:      strings.ToUpper(method),
		Pattern:     pattern,
		Segments:    parseSegments(pattern),
		Handler:     h,
		Middlewares: mws,
	})
	r.mu.Unlock()
}

// HTTP method shortcuts.
func (r *Router) GET(p string, h Handler, mws ...Middleware)    { r.Handle("GET", p, h, mws...) }
func (r *Router) POST(p string, h Handler, mws ...Middleware)   { r.Handle("POST", p, h, mws...) }
func (r *Router) PUT(p string, h Handler, mws ...Middleware)    { r.Handle("PUT", p, h, mws...) }
func (r *Router) DELETE(p string, h Handler, mws ...Middleware) { r.Handle("DELETE", p, h, mws...) }
func (r *Router) PATCH(p string, h Handler, mws ...Middleware)  { r.Handle("PATCH", p, h, mws...) }

// SetNotFound overrides the default 404 handler.
func (r *Router) SetNotFound(h Handler) {
	r.notFound = h
}

// dispatch finds the matching route and runs the full middleware chain.
func (r *Router) dispatch(req *Request) *Response {
	r.mu.RLock()
	routes := r.routes
	middlewares := r.middlewares
	notFound := r.notFound
	r.mu.RUnlock()

	for _, route := range routes {
		if params, ok := route.match(req.Request.Method, req.Request.URL.Path); ok {
			req.Params = params
			// Compose: global middlewares + per-route middlewares + handler
			h := route.Handler
			for i := len(route.Middlewares) - 1; i >= 0; i-- {
				h = route.Middlewares[i](h)
			}
			for i := len(middlewares) - 1; i >= 0; i-- {
				h = middlewares[i](h)
			}
			return h(req)
		}
	}
	return notFound(req)
}

// ServeHTTP implements http.Handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	bootReq := &Request{Request: req, Params: map[string]string{}}
	resp := r.dispatch(bootReq)

	if resp == nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	if resp.ContentType != "" {
		w.Header().Set("Content-Type", resp.ContentType)
	}
	for k, v := range resp.Headers {
		w.Header().Set(k, v)
	}
	if resp.Status == 0 {
		resp.Status = 200
	}
	w.WriteHeader(resp.Status)
	fmt.Fprint(w, resp.Body)
}
