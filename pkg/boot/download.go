// download.go — File download / CSV export for KylixBoot (v0.9.0 P1).
//
//	resp.Download('/srv/data/report.pdf', 'report.pdf');
//	resp.CSV([][]string{{'id', 'name'}, {'1', 'alice'}}, 'users.csv');
//
// Content-Type is inferred from the file extension (same table as static
// serving); Content-Disposition is attachment with a quoted filename.
package boot

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Download responds with the file at path as an attachment named filename.
// On read failure it turns into a 404 (missing) / 500 (unreadable) text
// response — handlers rarely need to branch on this.
func (r *Response) Download(path, filename string) *Response {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return r.StatusCode(404).WithHeader("Content-Type", "text/plain; charset=utf-8").Send("file not found")
		}
		return r.StatusCode(500).WithHeader("Content-Type", "text/plain; charset=utf-8").Send("cannot read file")
	}
	return r.FileBytes(data, filename)
}

// FileBytes responds with in-memory bytes as an attachment (CSV blob the
// handler built itself, generated PDFs, ...).
func (r *Response) FileBytes(data []byte, filename string) *Response {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	ct := mimeFor(strings.ToLower(filepath.Ext(filename)))
	if ct == "" {
		ct = "application/octet-stream"
	}
	r.Status = 200
	r.ContentType = ct
	r.Body = string(data)
	r.Headers["Content-Disposition"] =
		fmt.Sprintf("attachment; filename=%q", filepath.Base(filename))
	return r
}

// CSV encodes rows as RFC 4180 CSV and responds with it as an attachment.
// The first row, when present, is the header line.
func (r *Response) CSV(rows [][]string, filename string) *Response {
	var b strings.Builder
	w := csv.NewWriter(&b)
	if err := w.WriteAll(rows); err != nil {
		return r.StatusCode(500).Send("csv encode failed")
	}
	w.Flush()
	return r.FileBytes([]byte(b.String()), filename)
}
