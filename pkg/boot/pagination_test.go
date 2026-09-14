// pagination_test.go — Page/Pageable tests (v0.9.0 P1).
package boot

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPageableFrom(t *testing.T) {
	cases := []struct {
		query      string
		page, size int
	}{
		{"", 1, DefaultPageSize},
		{"?page=3&size=10", 3, 10},
		{"?page=0", 1, DefaultPageSize},   // clamp low
		{"?page=-2", 1, DefaultPageSize},  // clamp low
		{"?size=999", 1, MaxPageSize},     // clamp high
		{"?size=0", 1, DefaultPageSize},   // fallback
		{"?page=abc", 1, DefaultPageSize}, // fallback
		{"?page=5&size=abc", 5, DefaultPageSize},
	}
	for _, c := range cases {
		req := &Request{Request: httptest.NewRequest("GET", "/list"+c.query, nil)}
		pg := req.Pageable()
		if pg.Page != c.page || pg.Size != c.size {
			t.Errorf("%q → page=%d size=%d, want %d/%d", c.query, pg.Page, pg.Size, c.page, c.size)
		}
		if pg.Offset != (c.page-1)*c.size {
			t.Errorf("%q → offset=%d, want %d", c.query, pg.Offset, (c.page-1)*c.size)
		}
	}
}

func TestNewPage(t *testing.T) {
	// 95 items total, size 20 → 5 pages; page 2 has prev+next.
	items := make([]interface{}, 20)
	p := NewPage(items, 2, 20, 95)
	if p.Pages != 5 || !p.HasPrev || !p.HasNext {
		t.Errorf("pages=%d prev=%v next=%v", p.Pages, p.HasPrev, p.HasNext)
	}
	if p.PrevPage() != 1 || p.NextPage() != 3 {
		t.Errorf("prev=%d next=%d", p.PrevPage(), p.NextPage())
	}
	// First page: no prev.
	p = NewPage(items, 1, 20, 95)
	if p.HasPrev || p.PrevPage() != 1 {
		t.Error("page 1 must not have prev")
	}
	// Last page: no next; NextPage clamps.
	p = NewPage(items, 5, 20, 95)
	if p.HasNext || p.NextPage() != 5 {
		t.Error("page 5 must not have next")
	}
	// Exact division: 100/20 → 5 pages.
	p = NewPage(items, 1, 20, 100)
	if p.Pages != 5 {
		t.Errorf("100/20 pages = %d, want 5", p.Pages)
	}
	// Zero total → 1 page, no crash.
	p = NewPage(nil, 1, 20, 0)
	if p.Pages != 1 || p.Items == nil {
		t.Errorf("empty page: pages=%d items=%v", p.Pages, p.Items)
	}
}

func TestHTMLPager(t *testing.T) {
	items := make([]interface{}, 10)
	p := NewPage(items, 2, 10, 95) // 10 pages
	html := p.HTMLPager("/admin/users?size=10", 2)
	for _, want := range []string{
		`<nav class="pager">`,
		`class="pager-prev" href="/admin/users?page=1&amp;size=10"`,
		`class="pager-current">2<`,
		`class="pager-link" href="/admin/users?page=1&amp;size=10"`,
		`class="pager-link" href="/admin/users?page=10&amp;size=10"`,
		`class="pager-ellipsis"`,
		`class="pager-next" href="/admin/users?page=3&amp;size=10"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("pager missing %q in:\n%s", want, html)
		}
	}
	// Single page → disabled arrows, no numbered links beyond current.
	p = NewPage(items, 1, 20, 15)
	html = p.HTMLPager("/x", 2)
	if !strings.Contains(html, "pager-disabled") {
		t.Error("single page should disable arrows")
	}
	if strings.Contains(html, "pager-ellipsis") {
		t.Error("single page should have no ellipsis")
	}
}

func TestPageableEndToEnd(t *testing.T) {
	r := NewRouter()
	r.GET("/list", func(req *Request) *Response {
		pg := req.Pageable()
		total := int64(55)
		items := make([]interface{}, 0, pg.Size)
		_ = items
		page := NewPage(nil, pg.Page, pg.Size, total)
		return Text(200, page.HTMLPager("/list", 1))
	})
	req := httptest.NewRequest("GET", "/list?page=3&size=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	body := w.Body.String()
	if !strings.Contains(body, "pager-current\">3<") {
		t.Errorf("page 3 not current:\n%s", body)
	}
	_ = http.StatusOK
}
