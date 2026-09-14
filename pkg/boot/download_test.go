// download_test.go — Download/CSV tests (v0.9.0 P1).
package boot

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResponse_Download(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.txt")
	if err := os.WriteFile(path, []byte("hello download"), 0o644); err != nil {
		t.Fatal(err)
	}

	res := NewResponse(200, "").Download(path, "report.txt")
	if res.Status != 200 || res.Body != "hello download" {
		t.Errorf("status=%d body=%q", res.Status, res.Body)
	}
	if res.ContentType != "text/plain; charset=utf-8" {
		t.Errorf("content-type=%q", res.ContentType)
	}
	cd := res.Headers["Content-Disposition"]
	if cd != `attachment; filename="report.txt"` {
		t.Errorf("disposition=%q", cd)
	}

	// Missing file → 404.
	res = NewResponse(200, "").Download(filepath.Join(dir, "nope.txt"), "x.txt")
	if res.Status != 404 {
		t.Errorf("missing file: got %d, want 404", res.Status)
	}
}

func TestResponse_DownloadPathTraversal(t *testing.T) {
	// filename is always sanitized to its base name.
	res := NewResponse(200, "").FileBytes([]byte("x"), "../../etc/passwd")
	if !strings.Contains(res.Headers["Content-Disposition"], `filename="passwd"`) {
		t.Errorf("filename not basenamed: %q", res.Headers["Content-Disposition"])
	}
}

func TestResponse_CSV(t *testing.T) {
	res := NewResponse(200, "").CSV([][]string{
		{"id", "name"},
		{"1", "alice, jr"},
		{"2", `bob "the builder"`},
	}, "users.csv")
	if res.ContentType != "text/csv; charset=utf-8" {
		t.Errorf("content-type=%q", res.ContentType)
	}
	if res.Headers["Content-Disposition"] != `attachment; filename="users.csv"` {
		t.Errorf("disposition=%q", res.Headers["Content-Disposition"])
	}
	want := "id,name\n1,\"alice, jr\"\n2,\"bob \"\"the builder\"\"\"\n"
	if res.Body != want {
		t.Errorf("csv=\n%q\nwant\n%q", res.Body, want)
	}
}

func TestDownload_RouteEndToEnd(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.csv")
	os.WriteFile(path, []byte("a,b\n1,2\n"), 0o644)

	r := NewRouter()
	r.GET("/export", func(req *Request) *Response {
		return NewResponse(200, "").Download(path, "export.csv")
	})
	req := httptest.NewRequest("GET", "/export", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 || !strings.Contains(w.Header().Get("Content-Disposition"), "attachment") {
		t.Errorf("code=%d disposition=%q", w.Code, w.Header().Get("Content-Disposition"))
	}
}
