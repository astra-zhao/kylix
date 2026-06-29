package stdlib

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestServer starts an httptest server that handles GET/POST/PUT/DELETE
// and echoes the method + body so tests can assert on them.
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, r.Method+"|"+string(body))
	})
	mux.HandleFunc("/created", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		io.WriteString(w, "created")
	})
	return httptest.NewServer(mux)
}

func TestHttp_Put(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	body, err := HttpPut(srv.URL, "text/plain", "data")
	if err != nil {
		t.Fatalf("HttpPut failed: %v", err)
	}
	if body != "PUT|data" {
		t.Errorf("HttpPut body = %q, want PUT|data", body)
	}
}

func TestHttp_Delete(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	body, err := HttpDelete(srv.URL)
	if err != nil {
		t.Fatalf("HttpDelete failed: %v", err)
	}
	if body != "DELETE|" {
		t.Errorf("HttpDelete body = %q, want DELETE|", body)
	}
}

func TestHttp_PostJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"echo":"ok"}`)
	}))
	defer srv.Close()

	m, err := HttpPostJSON(srv.URL, `{"x":1}`)
	if err != nil {
		t.Fatalf("HttpPostJSON failed: %v", err)
	}
	if m["echo"] != "ok" {
		t.Errorf("echo = %v, want ok", m["echo"])
	}
}

func TestHttp_DoGetReturnsStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/created" {
			w.WriteHeader(http.StatusCreated)
			io.WriteString(w, "created")
			return
		}
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "ok")
	}))
	defer srv.Close()

	resp, err := HttpDoGet(srv.URL + "/created")
	if err != nil {
		t.Fatalf("HttpDoGet failed: %v", err)
	}
	if resp.Status != 201 {
		t.Errorf("status = %d, want 201", resp.Status)
	}
	if resp.Body != "created" {
		t.Errorf("body = %q, want created", resp.Body)
	}
}

func TestHttp_DoPostReturnsStatus(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	resp, err := HttpDoPost(srv.URL, "text/plain", "payload")
	if err != nil {
		t.Fatalf("HttpDoPost failed: %v", err)
	}
	if resp.Status != 200 {
		t.Errorf("status = %d, want 200", resp.Status)
	}
	if resp.Body != "POST|payload" {
		t.Errorf("body = %q, want POST|payload", resp.Body)
	}
}

func TestHttpClient_PutDelete(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	c := NewHttpClient(srv.URL, 5000)
	c.SetHeader("X-Test", "1")

	putBody, err := c.Put("/", "text/plain", "update")
	if err != nil {
		t.Fatalf("client.Put failed: %v", err)
	}
	if !strings.HasPrefix(putBody, "PUT|") {
		t.Errorf("Put body = %q, want PUT|...", putBody)
	}

	delBody, err := c.Delete("/")
	if err != nil {
		t.Fatalf("client.Delete failed: %v", err)
	}
	if !strings.HasPrefix(delBody, "DELETE") {
		t.Errorf("Delete body = %q, want DELETE...", delBody)
	}
}
