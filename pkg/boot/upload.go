// upload.go — multipart/form-data file upload for KylixBoot (v0.9.0 P1).
//
//	content, filename, ok := req.File('avatar')
//	path, ok := req.SaveFile('avatar', '/srv/uploads')
//
// The body is parsed from the cached raw bytes with mime/multipart, so
// semantics do not depend on server read order (and the LLVM boot runtime
// can mirror the same limits byte-for-byte):
//
//   - bodies larger than MaxUploadSize are rejected;
//   - SaveFile only allows whitelisted extensions and stores under the
//     sanitized base name (no path traversal).
package boot

import (
	"bytes"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

// MaxUploadSize caps the parsed multipart body (8 MB).
const MaxUploadSize = 8 << 20

// errUploadTooLarge marks oversized bodies/parts.
var errUploadTooLarge = errors.New("upload too large")

// maxMultipartFields caps the number of text fields in one form.
const maxMultipartFields = 64

// allowedUploadExts is the SaveFile whitelist (the admin platform needs
// avatars/attachments; extend deliberately, never with .html/.php/.sh).
var allowedUploadExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	".pdf": true, ".txt": true, ".csv": true, ".zip": true,
}

// File returns (content, clientFilename, ok) for the first uploaded part
// named `name`. ok=false when the body is not multipart, the part is
// missing, or the body exceeds MaxUploadSize.
func (r *Request) File(name string) ([]byte, string, bool) {
	if err := r.parseMultipart(); err != nil {
		return nil, "", false
	}
	p, ok := r.mpFiles[name]
	if !ok {
		return nil, "", false
	}
	content, err := io.ReadAll(p.body)
	if err != nil || len(content) > MaxUploadSize {
		return nil, "", false
	}
	p.body.Seek(0, io.SeekStart)
	return content, p.filename, true
}

// MultipartFiled returns a text field from a multipart form (CSRF tokens,
// entity ids — the urlencoded req.Form covers its own bodies).
func (r *Request) MultipartField(name string) string {
	if err := r.parseMultipart(); err != nil {
		return ""
	}
	return r.mpFields[name]
}

// SaveFile writes the uploaded part named `name` into dir under its
// sanitized client filename. Returns the stored path and ok=false when
// the part is missing, oversized, not whitelisted, or unwritable.
func (r *Request) SaveFile(name, dir string) (string, bool) {
	content, filename, ok := r.File(name)
	if !ok {
		return "", false
	}
	base := filepath.Base(strings.ReplaceAll(filename, "\\", "/"))
	if base == "" || base == "." || base == ".." || strings.HasPrefix(base, ".") {
		return "", false
	}
	ext := strings.ToLower(filepath.Ext(base))
	if !allowedUploadExts[ext] {
		return "", false
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", false
	}
	path := filepath.Join(dir, base)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return "", false
	}
	return path, true
}

// uploadedPart is one parsed file part.
type uploadedPart struct {
	filename string
	body     *bytes.Reader
}

// parseMultipart parses the cached raw body once, caching text fields and
// file parts on the request. Non-multipart bodies return an error and
// leave the caches empty (req.Form falls back to urlencoded).
func (r *Request) parseMultipart() error {
	if r.mpParsed {
		return r.mpErr
	}
	r.mpParsed = true
	r.mpFields = map[string]string{}
	r.mpFiles = map[string]*uploadedPart{}
	body := r.Body()
	if len(body) == 0 || len(body) > MaxUploadSize {
		r.mpErr = errUploadTooLarge
		return r.mpErr
	}
	mediaType, params, err := mime.ParseMediaType(r.Header("Content-Type"))
	if err != nil || !strings.HasPrefix(mediaType, "multipart/") {
		r.mpErr = errors.New("not multipart")
		return r.mpErr
	}
	boundary := params["boundary"]
	if boundary == "" {
		r.mpErr = errors.New("missing boundary")
		return r.mpErr
	}
	mr := multipart.NewReader(bytes.NewReader(body), boundary)
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			r.mpErr = err
			return r.mpErr
		}
		// Read fully with the shared cap so a hostile part cannot stream.
		data, err := io.ReadAll(io.LimitReader(part, MaxUploadSize+1))
		if err != nil && err != io.EOF {
			r.mpErr = err
			return r.mpErr
		}
		if len(data) > MaxUploadSize {
			r.mpErr = errUploadTooLarge
			return r.mpErr
		}
		name := part.FormName()
		if part.FileName() == "" {
			// Text field (hidden inputs, selects, text areas).
			if len(r.mpFields) < maxMultipartFields {
				r.mpFields[name] = string(data)
			}
			continue
		}
		r.mpFiles[name] = &uploadedPart{
			filename: part.FileName(),
			body:     bytes.NewReader(data),
		}
	}
}
