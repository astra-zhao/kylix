package boot

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestResponse_Send(t *testing.T) {
	res := NewResponse(200, "")
	if got := res.Send("hello"); got != res {
		t.Fatal("Send should return the receiver")
	}
	if res.Body != "hello" {
		t.Errorf("Body=%q, want hello", res.Body)
	}
	if res.ContentType != "text/plain; charset=utf-8" {
		t.Errorf("ContentType=%q", res.ContentType)
	}
	if res.Headers == nil {
		t.Error("Headers should be initialized")
	}
}

func TestResponse_StatusCode(t *testing.T) {
	res := NewResponse(200, "ok")
	if got := res.StatusCode(201); got != res {
		t.Fatal("StatusCode should return the receiver")
	}
	if res.Status != 201 {
		t.Errorf("Status=%d, want 201", res.Status)
	}
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

// ===== v0.7.0 P2: page rendering API =====

func TestRequest_Form(t *testing.T) {
	body := "name=Alice&city=New%20York"
	req := httptest.NewRequest("POST", "/submit", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	br := &Request{Request: req}
	if got := br.Form("name"); got != "Alice" {
		t.Errorf("Form(name)=%q, want Alice", got)
	}
	if got := br.Form("city"); got != "New York" {
		t.Errorf("Form(city)=%q, want 'New York' (URL-decoded)", got)
	}
	if got := br.Form("missing"); got != "" {
		t.Errorf("Form(missing)=%q, want empty", got)
	}
}

func TestRequest_FormQueryFallback(t *testing.T) {
	req := httptest.NewRequest("GET", "/page?name=Bob", nil)
	br := &Request{Request: req}
	if got := br.Form("name"); got != "Bob" {
		t.Errorf("Form(name)=%q, want Bob (query fallback)", got)
	}
}

func TestRequest_Cookie(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "abc123"})
	br := &Request{Request: req}
	if got := br.Cookie("session"); got != "abc123" {
		t.Errorf("Cookie(session)=%q, want abc123", got)
	}
	if got := br.Cookie("nope"); got != "" {
		t.Errorf("Cookie(nope)=%q, want empty", got)
	}
}

func TestResponse_Html(t *testing.T) {
	res := NewResponse(200, "")
	if got := res.Html("<h1>hi</h1>"); got != res {
		t.Fatal("Html should return the receiver")
	}
	if res.Body != "<h1>hi</h1>" {
		t.Errorf("Body=%q", res.Body)
	}
	if res.ContentType != "text/html; charset=utf-8" {
		t.Errorf("ContentType=%q", res.ContentType)
	}
}

func TestResponse_WithCookie(t *testing.T) {
	res := NewResponse(200, "x").WithCookie("session", "abc")
	res.WithCookie("theme", "dark")
	if len(res.Cookies) != 2 {
		t.Fatalf("Cookies=%v, want 2 entries", res.Cookies)
	}
	if res.Cookies[0] != "session=abc; Path=/" {
		t.Errorf("Cookies[0]=%q", res.Cookies[0])
	}
	// Cookie must reach the wire as a Set-Cookie header.
	rt := NewRouter()
	rt.GET("/c", func(r *Request) *Response { return res })
	_, _, hdr := doRequest(rt, "GET", "/c")
	if got := hdr.Get("Set-Cookie"); got != "session=abc; Path=/" {
		t.Errorf("Set-Cookie=%q", got)
	}
}

func TestStaticServing(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "style.css"), []byte("body{color:red}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "sub"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sub", "page.html"), []byte("<p>hi</p>"), 0644); err != nil {
		t.Fatal(err)
	}

	rt := NewRouter()
	rt.SetStaticDir(dir)
	rt.GET("/api", func(r *Request) *Response { return Text(200, "api") })

	// Known file with MIME type.
	code, body, hdr := doRequest(rt, "GET", "/static/style.css")
	if code != 200 || body != "body{color:red}" {
		t.Errorf("static css: code=%d body=%q", code, body)
	}
	if ct := hdr.Get("Content-Type"); ct != "text/css; charset=utf-8" {
		t.Errorf("Content-Type=%q", ct)
	}
	// Nested file.
	code, body, _ = doRequest(rt, "GET", "/static/sub/page.html")
	if code != 200 || body != "<p>hi</p>" {
		t.Errorf("static nested: code=%d body=%q", code, body)
	}
	// Missing file falls through to 404.
	code, _, _ = doRequest(rt, "GET", "/static/nope.css")
	if code != 404 {
		t.Errorf("missing static: code=%d, want 404", code)
	}
	// Route still wins over static prefix.
	code, body, _ = doRequest(rt, "GET", "/api")
	if code != 200 || body != "api" {
		t.Errorf("route: code=%d body=%q", code, body)
	}
}

func TestStaticTraversalBlocked(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "ok.txt"), []byte("ok"), 0644); err != nil {
		t.Fatal(err)
	}
	secret := filepath.Join(dir, "..", "secret.txt")
	if err := os.WriteFile(secret, []byte("SECRET"), 0644); err != nil {
		t.Fatal(err)
	}
	rt := NewRouter()
	rt.SetStaticDir(dir)
	code, body, _ := doRequest(rt, "GET", "/static/../secret.txt")
	if code == 200 || body == "SECRET" {
		t.Errorf("traversal not blocked: code=%d body=%q", code, body)
	}
}

// ===== v0.7.0 P3: Redirect + custom error pages =====

func TestResponse_Redirect(t *testing.T) {
	res := NewResponse(200, "")
	if got := res.Redirect("/login"); got != res {
		t.Fatal("Redirect should return the receiver")
	}
	if res.Status != 302 {
		t.Errorf("Status=%d, want 302", res.Status)
	}
	if res.Headers["Location"] != "/login" {
		t.Errorf("Location=%q, want /login", res.Headers["Location"])
	}
}

func TestRouter_NotFoundPage(t *testing.T) {
	r := NewRouter()
	r.SetNotFoundPage("<h1>404 custom</h1>")
	status, body, ct := doRequest(r, "GET", "/missing")
	if status != 404 {
		t.Errorf("status=%d, want 404", status)
	}
	if body != "<h1>404 custom</h1>" {
		t.Errorf("body=%q", body)
	}
	if !strings.Contains(ct.Get("Content-Type"), "text/html") {
		t.Errorf("Content-Type=%q, want text/html", ct.Get("Content-Type"))
	}
}

func TestRouter_NotFoundPage_DefaultWhenUnset(t *testing.T) {
	r := NewRouter()
	status, body, _ := doRequest(r, "GET", "/missing")
	if status != 404 || !strings.HasPrefix(body, "404 Not Found") {
		t.Errorf("status=%d body=%q, want default 404", status, body)
	}
}

func TestRouter_ErrorPage_PanicRecover(t *testing.T) {
	r := NewRouter()
	r.SetErrorPage("<h1>500 custom</h1>")
	r.GET("/boom", func(req *Request) *Response {
		panic("boom")
	})
	status, body, _ := doRequest(r, "GET", "/boom")
	if status != 500 {
		t.Errorf("status=%d, want 500", status)
	}
	if body != "<h1>500 custom</h1>" {
		t.Errorf("body=%q", body)
	}
}

func TestRouter_ErrorPage_DefaultWhenUnset(t *testing.T) {
	r := NewRouter()
	r.GET("/boom", func(req *Request) *Response {
		panic("boom")
	})
	status, body, _ := doRequest(r, "GET", "/boom")
	if status != 500 || body != "500 Internal Server Error" {
		t.Errorf("status=%d body=%q, want default 500", status, body)
	}
}
