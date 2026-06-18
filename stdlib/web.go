package stdlib

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// TRequest wraps HTTP request
type TRequest struct {
	Request *http.Request
	Params  map[string]string
}

// TResponse wraps HTTP response
type TResponse struct {
	Writer     http.ResponseWriter
	StatusCode int
	Finished   bool
}

// TRouteHandler is the handler function type
type TRouteHandler func(*TRequest, *TResponse)

// TMiddleware is the middleware function type
type TMiddleware func(*TRequest, *TResponse)

// TRoute represents a route
type TRoute struct {
	Method  string
	Path    string
	Handler TRouteHandler
	Params  map[string]string
}

// TServer is the web server
type TServer struct {
	Port         int
	Routes       []TRoute
	Middlewares  []TMiddleware
	StaticRoutes []struct {
		PathPrefix string
		RootDir    string
	}
}

// NewServer creates a new server
func NewServer(port int) *TServer {
	return &TServer{
		Port:   port,
		Routes: []TRoute{},
	}
}

// Get registers a GET route
func (s *TServer) Get(path string, handler TRouteHandler) {
	s.Routes = append(s.Routes, TRoute{
		Method:  "GET",
		Path:    path,
		Handler: handler,
	})
}

// Post registers a POST route
func (s *TServer) Post(path string, handler TRouteHandler) {
	s.Routes = append(s.Routes, TRoute{
		Method:  "POST",
		Path:    path,
		Handler: handler,
	})
}

// Put registers a PUT route
func (s *TServer) Put(path string, handler TRouteHandler) {
	s.Routes = append(s.Routes, TRoute{
		Method:  "PUT",
		Path:    path,
		Handler: handler,
	})
}

// Delete registers a DELETE route
func (s *TServer) Delete(path string, handler TRouteHandler) {
	s.Routes = append(s.Routes, TRoute{
		Method:  "DELETE",
		Path:    path,
		Handler: handler,
	})
}

// Use adds a middleware
func (s *TServer) Use(middleware TMiddleware) {
	s.Middlewares = append(s.Middlewares, middleware)
}

// Static serves static files
func (s *TServer) Static(pathPrefix, rootDir string) {
	s.StaticRoutes = append(s.StaticRoutes, struct {
		PathPrefix string
		RootDir    string
	}{pathPrefix, rootDir})
}

// Listen starts the server
func (s *TServer) Listen() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		req := &TRequest{Request: r, Params: make(map[string]string)}
		res := &TResponse{Writer: w, StatusCode: 200}

		// Execute middlewares
		for _, mw := range s.Middlewares {
			mw(req, res)
			if res.Finished {
				return
			}
		}

		// Check static routes
		for _, sr := range s.StaticRoutes {
			if strings.HasPrefix(r.URL.Path, sr.PathPrefix) {
				serveStaticFile(w, r, sr.PathPrefix, sr.RootDir)
				return
			}
		}

		// Find matching route
		for _, route := range s.Routes {
			if route.Method == r.Method && matchPath(route.Path, r.URL.Path, req.Params) {
				route.Handler(req, res)
				return
			}
		}

		// 404
		w.WriteHeader(404)
		fmt.Fprint(w, "Not Found")
	})

	addr := fmt.Sprintf(":%d", s.Port)
	fmt.Printf("Kylix Web Server listening on port %d\n", s.Port)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// Request methods
func (r *TRequest) Path() string {
	return r.Request.URL.Path
}

func (r *TRequest) Method() string {
	return r.Request.Method
}

func (r *TRequest) Param(name string) string {
	return r.Params[name]
}

func (r *TRequest) Query(name string) string {
	return r.Request.URL.Query().Get(name)
}

func (r *TRequest) Header(name string) string {
	return r.Request.Header.Get(name)
}

func (r *TRequest) JSON(v interface{}) error {
	return json.NewDecoder(r.Request.Body).Decode(v)
}

func (r *TRequest) GetField(name string) string {
	// Try query parameters first
	if val := r.Query(name); val != "" {
		return val
	}
	// Try form values (for POST/PUT requests)
	return r.Request.FormValue(name)
}

// Response methods
func (r *TResponse) Status(code int) *TResponse {
	r.StatusCode = code
	return r
}

func (r *TResponse) Send(body string) {
	r.Writer.WriteHeader(r.StatusCode)
	fmt.Fprint(r.Writer, body)
	r.Finished = true
}

func (r *TResponse) JSON(v interface{}) {
	r.Writer.Header().Set("Content-Type", "application/json")
	r.Writer.WriteHeader(r.StatusCode)
	json.NewEncoder(r.Writer).Encode(v)
	r.Finished = true
}

func (r *TResponse) Header(name, value string) *TResponse {
	r.Writer.Header().Set(name, value)
	return r
}

// LoggerMiddleware creates a logging middleware
func LoggerMiddleware() TMiddleware {
	return func(req *TRequest, res *TResponse) {
		log.Printf("[%s] %s %s",
			"2024-01-01 00:00:00", // TODO: use actual time
			req.Method(),
			req.Path())
	}
}

// Helper functions
func matchPath(pattern, path string, params map[string]string) bool {
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	if len(patternParts) != len(pathParts) {
		return false
	}

	for i, part := range patternParts {
		if strings.HasPrefix(part, ":") {
			params[part[1:]] = pathParts[i]
		} else if part != pathParts[i] {
			return false
		}
	}

	return true
}

func serveStaticFile(w http.ResponseWriter, r *http.Request, pathPrefix, rootDir string) {
	filePath := filepath.Join(rootDir, strings.TrimPrefix(r.URL.Path, pathPrefix))

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		w.WriteHeader(404)
		fmt.Fprint(w, "File Not Found")
		return
	}

	http.ServeFile(w, r, filePath)
}
