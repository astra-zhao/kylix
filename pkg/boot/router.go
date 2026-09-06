// router.go — KylixBoot HTTP router with path parameters and middleware.
package boot

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
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
	StaticDir   string // directory served under /static/ (v0.7.0 P2)
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

// SetStaticDir enables static file serving under the /static/ URL prefix
// (e.g. SetStaticDir("./static") serves ./static/style.css at
// /static/style.css). URL paths are cleaned to prevent ".." traversal.
// v0.7.0 P2.
func (r *Router) SetStaticDir(dir string) {
	r.StaticDir = dir
}

// staticPrefix is the reserved URL prefix for static files.
const staticPrefix = "/static/"

// mimeFor returns the Content-Type for a file extension ("" = unknown →
// application/octet-stream).
func mimeFor(ext string) string {
	switch strings.ToLower(ext) {
	case ".html", ".htm":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js", ".mjs":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".txt":
		return "text/plain; charset=utf-8"
	case ".xml":
		return "application/xml"
	case ".pdf":
		return "application/pdf"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	default:
		return "application/octet-stream"
	}
}

// serveStatic maps a /static/... URL path to a file under StaticDir and
// returns the file response, or nil when the file does not exist.
func (r *Router) serveStatic(urlPath string) *Response {
	if r.StaticDir == "" || !strings.HasPrefix(urlPath, staticPrefix) {
		return nil
	}
	rel := strings.TrimPrefix(urlPath, staticPrefix)
	rel = path.Clean("/" + rel) // leading "/" forces absolute clean → no ".." escape
	if rel == "/" || rel == "" {
		return nil
	}
	full := filepath.Join(r.StaticDir, rel)
	data, err := os.ReadFile(full)
	if err != nil {
		return nil
	}
	return &Response{
		Status:      200,
		Body:        string(data),
		ContentType: mimeFor(filepath.Ext(full)),
		Headers:     map[string]string{},
	}
}

// dispatch finds the matching route and runs the full middleware chain.
func (r *Router) dispatch(req *Request) *Response {
	r.mu.RLock()
	routes := r.routes
	middlewares := r.middlewares
	notFound := r.notFound
	staticDir := r.StaticDir
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
	// Route miss → static files (v0.7.0 P2) → notFound.
	if staticDir != "" && req.Request.Method == http.MethodGet {
		if resp := r.serveStatic(req.Request.URL.Path); resp != nil {
			return resp
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
	for _, c := range resp.Cookies {
		w.Header().Add("Set-Cookie", c)
	}
	if resp.Status == 0 {
		resp.Status = 200
	}
	w.WriteHeader(resp.Status)
	fmt.Fprint(w, resp.Body)
}
