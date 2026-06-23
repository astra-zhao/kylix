package boot

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// helper: make a request and run it through the router
func doRequest(r *Router, method, path string) (int, string, http.Header) {
	req := httptest.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body), resp.Header
}

// ===== Router tests =====

func TestRouter_BasicGET(t *testing.T) {
	r := NewRouter()
	r.GET("/hello", func(req *Request) *Response {
		return Text(200, "world")
	})

	status, body, _ := doRequest(r, "GET", "/hello")
	if status != 200 || body != "world" {
		t.Errorf("got status=%d body=%q, want 200 'world'", status, body)
	}
}

func TestRouter_PathParam(t *testing.T) {
	r := NewRouter()
	r.GET("/users/:id", func(req *Request) *Response {
		return Text(200, "user:"+req.Param("id"))
	})

	status, body, _ := doRequest(r, "GET", "/users/42")
	if status != 200 || body != "user:42" {
		t.Errorf("got status=%d body=%q", status, body)
	}
}

func TestRouter_MultipleParams(t *testing.T) {
	r := NewRouter()
	r.GET("/users/:uid/posts/:pid", func(req *Request) *Response {
		return Text(200, req.Param("uid")+"-"+req.Param("pid"))
	})

	status, body, _ := doRequest(r, "GET", "/users/7/posts/abc")
	if status != 200 || body != "7-abc" {
		t.Errorf("got status=%d body=%q", status, body)
	}
}

func TestRouter_NotFound(t *testing.T) {
	r := NewRouter()
	r.GET("/foo", func(req *Request) *Response { return Text(200, "foo") })

	status, _, _ := doRequest(r, "GET", "/bar")
	if status != 404 {
		t.Errorf("got status=%d, want 404", status)
	}
}

func TestRouter_MethodMismatch(t *testing.T) {
	r := NewRouter()
	r.GET("/x", func(req *Request) *Response { return Text(200, "ok") })

	status, _, _ := doRequest(r, "POST", "/x")
	if status != 404 {
		t.Errorf("got status=%d, want 404 (POST to GET route)", status)
	}
}

func TestRouter_AllMethods(t *testing.T) {
	r := NewRouter()
	r.GET("/m", func(req *Request) *Response { return Text(200, "G") })
	r.POST("/m", func(req *Request) *Response { return Text(200, "P") })
	r.PUT("/m", func(req *Request) *Response { return Text(200, "U") })
	r.DELETE("/m", func(req *Request) *Response { return Text(200, "D") })
	r.PATCH("/m", func(req *Request) *Response { return Text(200, "A") })

	for _, c := range []struct{ method, want string }{
		{"GET", "G"}, {"POST", "P"}, {"PUT", "U"}, {"DELETE", "D"}, {"PATCH", "A"},
	} {
		_, body, _ := doRequest(r, c.method, "/m")
		if body != c.want {
			t.Errorf("%s: got %q want %q", c.method, body, c.want)
		}
	}
}

// ===== Response helpers =====

func TestJSON_Response(t *testing.T) {
	r := NewRouter()
	r.GET("/api", func(req *Request) *Response {
		return JSON(200, map[string]interface{}{"name": "Kylix", "ver": 3})
	})

	status, body, header := doRequest(r, "GET", "/api")
	if status != 200 {
		t.Errorf("status=%d", status)
	}
	if !strings.Contains(header.Get("Content-Type"), "application/json") {
		t.Errorf("missing JSON content-type: %q", header.Get("Content-Type"))
	}
	var got map[string]interface{}
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got["name"] != "Kylix" {
		t.Errorf("got name=%v", got["name"])
	}
}

// ===== Middleware =====

func TestMiddleware_Chain(t *testing.T) {
	r := NewRouter()
	r.Use(func(next Handler) Handler {
		return func(req *Request) *Response {
			resp := next(req)
			resp.WithHeader("X-One", "1")
			return resp
		}
	})
	r.Use(func(next Handler) Handler {
		return func(req *Request) *Response {
			resp := next(req)
			resp.WithHeader("X-Two", "2")
			return resp
		}
	})
	r.GET("/x", func(req *Request) *Response { return Text(200, "ok") })

	_, _, header := doRequest(r, "GET", "/x")
	if header.Get("X-One") != "1" || header.Get("X-Two") != "2" {
		t.Errorf("middleware headers missing: %v", header)
	}
}

func TestMiddleware_Recover(t *testing.T) {
	r := NewRouter()
	r.Use(Recover())
	r.GET("/panic", func(req *Request) *Response { panic("kaboom") })

	status, body, _ := doRequest(r, "GET", "/panic")
	if status != 500 || !strings.Contains(body, "Internal") {
		t.Errorf("got status=%d body=%q", status, body)
	}
}

func TestMiddleware_CORS(t *testing.T) {
	r := NewRouter()
	r.Use(CORS())
	r.GET("/x", func(req *Request) *Response { return Text(200, "ok") })

	_, _, header := doRequest(r, "GET", "/x")
	if header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("missing CORS header: %v", header)
	}
}

func TestMiddleware_Auth_Success(t *testing.T) {
	r := NewRouter()
	validator := func(t string) (string, bool) {
		if t == "secret123" {
			return "user42", true
		}
		return "", false
	}
	r.GET("/secure", func(req *Request) *Response { return Text(200, "secret data") }, Auth(validator))

	req := httptest.NewRequest("GET", "/secure", nil)
	req.Header.Set("Authorization", "Bearer secret123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("got status=%d, want 200", w.Code)
	}
}

func TestMiddleware_Auth_Fail(t *testing.T) {
	r := NewRouter()
	validator := func(t string) (string, bool) { return "", false }
	r.GET("/secure", func(req *Request) *Response { return Text(200, "secret") }, Auth(validator))

	status, _, _ := doRequest(r, "GET", "/secure")
	if status != 401 {
		t.Errorf("got status=%d, want 401", status)
	}
}

func TestMiddleware_RequestID(t *testing.T) {
	r := NewRouter()
	r.Use(RequestID())
	r.GET("/x", func(req *Request) *Response { return Text(200, "ok") })

	_, _, header := doRequest(r, "GET", "/x")
	if !strings.HasPrefix(header.Get("X-Request-ID"), "req-") {
		t.Errorf("missing/invalid X-Request-ID: %q", header.Get("X-Request-ID"))
	}
}

// ===== Container (DI) =====

func TestContainer_Singleton(t *testing.T) {
	c := NewContainer()
	calls := 0
	c.Register("MyService", func(*Container) interface{} {
		calls++
		return "instance"
	})

	a := c.Resolve("MyService")
	b := c.Resolve("MyService")
	if calls != 1 {
		t.Errorf("factory called %d times, want 1 (singleton)", calls)
	}
	if a != b {
		t.Error("singleton: got different instances")
	}
}

func TestContainer_Transient(t *testing.T) {
	c := NewContainer()
	calls := 0
	c.RegisterTransient("Counter", func(*Container) interface{} {
		calls++
		return calls
	})
	c.Resolve("Counter")
	c.Resolve("Counter")
	c.Resolve("Counter")
	if calls != 3 {
		t.Errorf("transient factory called %d times, want 3", calls)
	}
}

func TestContainer_Instance(t *testing.T) {
	c := NewContainer()
	c.RegisterInstance("Greeting", "hello")
	if c.Resolve("Greeting").(string) != "hello" {
		t.Error("instance not retrieved correctly")
	}
}

func TestContainer_TryResolve(t *testing.T) {
	c := NewContainer()
	c.RegisterInstance("X", 1)
	_, ok := c.TryResolve("X")
	if !ok {
		t.Error("expected TryResolve to succeed")
	}
	_, ok = c.TryResolve("Missing")
	if ok {
		t.Error("expected TryResolve to fail for missing")
	}
}

func TestContainer_Inject(t *testing.T) {
	c := NewContainer()
	c.RegisterInstance("Greeting", "hello")
	c.RegisterInstance("Count", 42)

	type Target struct {
		Greeting string
		Count    int
		Other    string // not registered
	}
	target := &Target{}
	if err := c.Inject(target); err != nil {
		t.Fatalf("inject error: %v", err)
	}
	if target.Greeting != "hello" {
		t.Errorf("Greeting=%q", target.Greeting)
	}
	if target.Count != 42 {
		t.Errorf("Count=%d", target.Count)
	}
	if target.Other != "" {
		t.Errorf("Other should be empty, got %q", target.Other)
	}
}

// ===== Config =====

func TestConfig_SetGet(t *testing.T) {
	c := NewConfig()
	c.Set("app.name", "MyApp")
	c.Set("server.port", 8080)

	if c.StringDefault("app.name", "") != "MyApp" {
		t.Errorf("app.name wrong")
	}
	if c.IntDefault("server.port", 0) != 8080 {
		t.Errorf("server.port wrong")
	}
	if c.StringDefault("missing", "fallback") != "fallback" {
		t.Errorf("fallback wrong")
	}
}

func TestConfig_EnvFallback(t *testing.T) {
	c := NewConfig()
	t.Setenv("APP_TITLE", "FromEnv")
	if c.StringDefault("app.title", "default") != "FromEnv" {
		t.Errorf("env fallback failed")
	}
}

func TestConfig_BoolDefault(t *testing.T) {
	c := NewConfig()
	c.Set("debug", "true")
	if !c.BoolDefault("debug", false) {
		t.Error("debug=true should parse as true")
	}
	c.Set("flag", false)
	if c.BoolDefault("flag", true) {
		t.Error("flag=false should be false")
	}
}

// ===== Smoke test: end-to-end through HTTP =====

func TestApp_EndToEnd(t *testing.T) {
	app := NewApp()
	app.Config.Set("app.name", "TestApp")
	app.Router.Use(RequestID())
	app.Router.GET("/info", func(req *Request) *Response {
		return JSON(200, map[string]string{
			"name": app.Config.StringDefault("app.name", "?"),
		})
	})

	req := httptest.NewRequest("GET", "/info", nil)
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("status=%d", w.Code)
	}
	var got map[string]string
	json.Unmarshal(w.Body.Bytes(), &got)
	if got["name"] != "TestApp" {
		t.Errorf("got name=%v", got["name"])
	}
}

// Quick check that the Server struct compiles & starts (we don't actually run it).
func TestServer_Construction(t *testing.T) {
	s := NewServer(8080)
	if s.Router == nil {
		t.Error("Router is nil")
	}
	if s.Addr != ":8080" {
		t.Errorf("Addr=%q", s.Addr)
	}
	_ = time.Now() // silence import
}
