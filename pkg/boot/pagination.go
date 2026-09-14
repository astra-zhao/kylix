// pagination.go — Page/Pageable abstraction for KylixBoot (v0.9.0 P1).
//
// Pageable parses ?page=&size= from the request (clamped), Page carries
// one page of results plus total metadata, and Page.HTMLPager renders a
// navigation bar for templates. Pairs with stdlib QueryBuilder:
//
//	qb := orm.NewQueryBuilder('users');
//	qb.Where('role', '=', role);
//	(cq, args) := qb.BuildCount();
//	total := db.ScalarInt(cq, args);
//	qb.Page(pg.Page, pg.Size);
//	(q, args) := qb.BuildSelect();
//	rows := DbQueryRows(q, args);
//	page := boot.NewPage(rows, pg.Page, pg.Size, total);
package boot

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Default/max page sizes for Pageable parsing.
const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Pageable is a parsed pagination request.
type Pageable struct {
	Page   int // 1-based
	Size   int // items per page
	Offset int // (Page-1)*Size, ready for QueryBuilder.Offset
}

// PageableFrom parses ?page= and ?size= with lenient fallbacks:
// unparsable or < 1 → default, size > MaxPageSize → MaxPageSize.
func PageableFrom(query url.Values) Pageable {
	page := atoiQuery(query.Get("page"), 1, 1<<30)
	size := atoiQuery(query.Get("size"), DefaultPageSize, MaxPageSize)
	return Pageable{Page: page, Size: size, Offset: (page - 1) * size}
}

// Pageable parses pagination parameters from the request query string.
func (r *Request) Pageable() Pageable {
	return PageableFrom(r.Request.URL.Query())
}

// PageNum returns the ?page= value clamped to >= 1, falling back to def when
// absent/unparsable (v0.9.0 P1.7 — scalar form used by the LLVM backend too,
// where the Pageable record cannot cross the handle boundary).
func (r *Request) PageNum(def int) int {
	return atoiQuery(r.Request.URL.Query().Get("page"), def, 1<<30)
}

// PageSize returns the ?size= value clamped to 1..max, falling back to def
// when absent/unparsable (v0.9.0 P1.7 — scalar form, see PageNum).
func (r *Request) PageSize(def, max int) int {
	return atoiQuery(r.Request.URL.Query().Get("size"), def, max)
}

// atoiQuery parses s leniently (the boot runtime never errors on query
// input): unparsable or < 1 → fallback, > hi → hi.
func atoiQuery(s string, fallback, hi int) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n < 1 {
		n = fallback
	}
	if n > hi {
		n = hi
	}
	return n
}

// Page is one page of results with metadata. Items holds the raw rows
// (records/arrays/strings — whatever the caller fetched).
type Page struct {
	Items   []interface{}
	Page    int
	Size    int
	Total   int64
	Pages   int
	HasPrev bool
	HasNext bool
}

// NewPage assembles a Page from fetched items and the total row count.
func NewPage(items []interface{}, page, size int, total int64) *Page {
	pages := 0
	if size > 0 {
		pages = int((total + int64(size) - 1) / int64(size))
	}
	if pages < 1 {
		pages = 1
	}
	if items == nil {
		items = []interface{}{}
	}
	return &Page{
		Items:   items,
		Page:    page,
		Size:    size,
		Total:   total,
		Pages:   pages,
		HasPrev: page > 1,
		HasNext: page < pages,
	}
}

// PrevPage / NextPage return the adjacent page numbers (clamped to 1..Pages).
func (p *Page) PrevPage() int {
	if p.Page <= 1 {
		return 1
	}
	return p.Page - 1
}

func (p *Page) NextPage() int {
	if p.Page >= p.Pages {
		return p.Pages
	}
	return p.Page + 1
}

// PageURL builds "base?page=N" preserving other query parameters of the
// current request (size, search filters, ...).
func (p *Page) PageURL(base string, page int) string {
	u, err := url.Parse(base)
	if err != nil {
		return fmt.Sprintf("%s?page=%d", base, page)
	}
	q := u.Query()
	q.Set("page", strconv.Itoa(page))
	u.RawQuery = q.Encode()
	return u.String()
}

// HTMLPager renders a minimal navigation bar (no external CSS needed —
// the admin design system styles ul/pager classes):
//
//	<nav class="pager">‹ Prev [1] [2] [3] … [9] Next ›</nav>
//
// window keeps at most `window` numbered links around the current page.
func (p *Page) HTMLPager(baseURL string, window int) string {
	var b strings.Builder
	b.WriteString(`<nav class="pager">`)
	if p.HasPrev {
		b.WriteString(`<a class="pager-prev" href="` + htmlAttr(p.PageURL(baseURL, p.PrevPage())) + `">‹ Prev</a>`)
	} else {
		b.WriteString(`<span class="pager-prev pager-disabled">‹ Prev</span>`)
	}
	lo := p.Page - window
	if lo < 1 {
		lo = 1
	}
	hi := p.Page + window
	if hi > p.Pages {
		hi = p.Pages
	}
	if lo > 1 {
		b.WriteString(pageLink(p.PageURL(baseURL, 1), 1))
		if lo > 2 {
			b.WriteString(`<span class="pager-ellipsis">…</span>`)
		}
	}
	for n := lo; n <= hi; n++ {
		if n == p.Page {
			b.WriteString(`<span class="pager-current">` + strconv.Itoa(n) + `</span>`)
		} else {
			b.WriteString(pageLink(p.PageURL(baseURL, n), n))
		}
	}
	if hi < p.Pages {
		if hi < p.Pages-1 {
			b.WriteString(`<span class="pager-ellipsis">…</span>`)
		}
		b.WriteString(pageLink(p.PageURL(baseURL, p.Pages), p.Pages))
	}
	if p.HasNext {
		b.WriteString(`<a class="pager-next" href="` + htmlAttr(p.PageURL(baseURL, p.NextPage())) + `">Next ›</a>`)
	} else {
		b.WriteString(`<span class="pager-next pager-disabled">Next ›</span>`)
	}
	b.WriteString(`</nav>`)
	return b.String()
}

func pageLink(href string, n int) string {
	return `<a class="pager-link" href="` + htmlAttr(href) + `">` + strconv.Itoa(n) + `</a>`
}

// PagerHTML renders the same navigation bar as Page.HTMLPager without
// assembling a Page first (v0.9.0 P1.7 — module-level form callable from the
// LLVM backend, where no Page object crosses the handle boundary).
func PagerHTML(baseURL string, page, size int, total int64, window int) string {
	return NewPage(nil, page, size, total).HTMLPager(baseURL, window)
}

// htmlAttr escapes the few characters that matter inside a
// double-quoted attribute value (URLs from page numbers cannot contain
// them in practice, but base URLs come from the caller).
func htmlAttr(s string) string {
	r := strings.NewReplacer("&", "&amp;", `"`, "&quot;", "<", "&lt;", ">", "&gt;")
	return r.Replace(s)
}
