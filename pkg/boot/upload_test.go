// upload_test.go — multipart upload tests (v0.9.0 P1).
package boot

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// buildMultipart creates a multipart body with the given text fields and
// file fields (name → filename, content).
func buildMultipart(t *testing.T, fields map[string]string, files map[string][2]string) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			t.Fatal(err)
		}
	}
	for name, fv := range files {
		fw, err := w.CreateFormFile(name, fv[0])
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write([]byte(fv[1])); err != nil {
			t.Fatal(err)
		}
	}
	w.Close()
	return &buf, w.FormDataContentType()
}

func TestUpload_FileAndField(t *testing.T) {
	body, ct := buildMultipart(t,
		map[string]string{"_csrf": "tok123", "title": "hello"},
		map[string][2]string{"doc": {"notes.txt", "file content here"}},
	)
	req := &Request{Request: httptest.NewRequest("POST", "/upload", body)}
	req.Request.Header.Set("Content-Type", ct)

	content, filename, ok := req.File("doc")
	if !ok || filename != "notes.txt" || string(content) != "file content here" {
		t.Fatalf("file=%q ok=%v content=%q", filename, ok, content)
	}
	// Text fields readable through both accessors.
	if got := req.MultipartField("_csrf"); got != "tok123" {
		t.Errorf("multipart field _csrf=%q", got)
	}
	if got := req.Form("title"); got != "hello" {
		t.Errorf("Form(title)=%q", got)
	}
	if got := req.Form("_csrf"); got != "tok123" {
		t.Errorf("Form(_csrf) fallback=%q", got)
	}
	// Missing part → ok=false, no panic.
	if _, _, ok := req.File("nope"); ok {
		t.Error("missing part must not be ok")
	}
}

func TestUpload_SaveFile(t *testing.T) {
	dir := t.TempDir()
	body, ct := buildMultipart(t, nil,
		map[string][2]string{
			"avatar": {"me.PNG", "pngbytes"},
			"evil":   {"../../etc/passwd", "x"},   // traversal in name
			"script": {"shell.sh", "#!/bin/sh\n"}, // not whitelisted
		},
	)
	req := &Request{Request: httptest.NewRequest("POST", "/u", body)}
	req.Request.Header.Set("Content-Type", ct)

	path, ok := req.SaveFile("avatar", dir)
	if !ok || filepath.Base(path) != "me.PNG" {
		t.Fatalf("save=%q ok=%v", path, ok)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "pngbytes" {
		t.Fatalf("saved content %q err=%v", data, err)
	}

	if _, ok := req.SaveFile("evil", dir); ok {
		t.Error("path-traversal filename must be rejected")
	}
	if _, ok := req.SaveFile("script", dir); ok {
		t.Error("non-whitelisted extension must be rejected")
	}
	if _, ok := req.SaveFile("nope", dir); ok {
		t.Error("missing part must not save")
	}
}

func TestUpload_RejectsOversize(t *testing.T) {
	big := bytes.Repeat([]byte("a"), MaxUploadSize+1)
	body, ct := buildMultipart(t, nil, map[string][2]string{"f": {"big.bin", string(big)}})
	req := &Request{Request: httptest.NewRequest("POST", "/u", body)}
	req.Request.Header.Set("Content-Type", ct)
	if _, _, ok := req.File("f"); ok {
		t.Error("oversized part must be rejected")
	}
}

func TestUpload_NonMultipartNoCrash(t *testing.T) {
	req := &Request{Request: httptest.NewRequest("POST", "/u", nil)}
	req.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if _, _, ok := req.File("f"); ok {
		t.Error("non-multipart must not produce a file")
	}
	if v := req.MultipartField("x"); v != "" {
		t.Error("non-multipart must not produce fields")
	}
}
