package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"kylix/registry/internal/auth"
	"kylix/registry/internal/db"
	"kylix/registry/internal/models"
)

// Handler holds all API dependencies.
type Handler struct {
	store   db.Store
	authSvc *auth.Service
}

func NewHandler(store db.Store, authSvc *auth.Service) *Handler {
	return &Handler{store: store, authSvc: authSvc}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

// HandlePackages routes GET /api/v1/packages (list/search) and POST (publish).
func (h *Handler) HandlePackages(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listPackages(w, r)
	case http.MethodPost:
		h.authSvc.RequireAuth(h.publishPackage)(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// GET /api/v1/packages?q=<query>&limit=20&offset=0
func (h *Handler) listPackages(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	limit := queryInt(r, "limit", 20)
	offset := queryInt(r, "offset", 0)

	if limit > 100 {
		limit = 100
	}

	packages, err := h.store.ListPackages(q, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if packages == nil {
		packages = []*models.Package{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"packages": packages,
		"count":    len(packages),
		"limit":    limit,
		"offset":   offset,
	})
}

// POST /api/v1/packages  (requires Bearer token)
// Body: {"name":"...", "description":"...", "repo_url":"...", "version":"...", "tarball_url":"...", "dependencies":"..."}
func (h *Handler) publishPackage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name         string `json:"name"`
		Description  string `json:"description"`
		RepoURL      string `json:"repo_url"`
		Version      string `json:"version"`
		TarballURL   string `json:"tarball_url"`
		Dependencies string `json:"dependencies"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Version == "" {
		writeError(w, http.StatusBadRequest, "version is required")
		return
	}
	if req.TarballURL == "" {
		writeError(w, http.StatusBadRequest, "tarball_url is required")
		return
	}

	// Get or create package
	owner := tokenOwner(r)
	pkg, err := h.store.GetPackage(req.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if pkg == nil {
		pkg = &models.Package{
			Name:        req.Name,
			Owner:       owner,
			Description: req.Description,
			RepoURL:     req.RepoURL,
		}
		if err := h.store.CreatePackage(pkg); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create package")
			return
		}
	}

	// Check version uniqueness
	existing, err := h.store.GetVersion(pkg.ID, req.Version)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if existing != nil {
		writeError(w, http.StatusConflict, fmt.Sprintf("version %s already exists", req.Version))
		return
	}

	ver := &models.Version{
		PackageID:    pkg.ID,
		Version:      req.Version,
		TarballURL:   req.TarballURL,
		Dependencies: req.Dependencies,
		PublishedAt:  time.Now(),
	}
	if err := h.store.CreateVersion(ver); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to publish version")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"message": "published",
		"package": pkg.Name,
		"version": ver.Version,
	})
}

// HandlePackageDetail routes /api/v1/packages/:name and sub-paths.
func (h *Handler) HandlePackageDetail(w http.ResponseWriter, r *http.Request) {
	// Path: /api/v1/packages/<name>[/versions][/<version>/dl]
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/packages/")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusBadRequest, "package name required")
		return
	}

	name := parts[0]

	switch {
	case len(parts) == 1:
		// GET /api/v1/packages/:name
		h.getPackage(w, r, name)
	case len(parts) == 2 && parts[1] == "versions":
		// GET /api/v1/packages/:name/versions
		h.listVersions(w, r, name)
	case len(parts) == 3 && parts[2] == "dl":
		// GET /api/v1/packages/:name/:version/dl
		h.downloadVersion(w, r, name, parts[1])
	default:
		writeError(w, http.StatusNotFound, "endpoint not found")
	}
}

// GET /api/v1/packages/:name
func (h *Handler) getPackage(w http.ResponseWriter, r *http.Request, name string) {
	pkg, err := h.store.GetPackage(name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if pkg == nil {
		writeError(w, http.StatusNotFound, "package not found")
		return
	}

	versions, err := h.store.ListVersions(pkg.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if versions == nil {
		versions = []*models.Version{}
	}

	downloads, _ := h.store.GetDownloadCount(pkg.ID)

	writeJSON(w, http.StatusOK, map[string]any{
		"package":   pkg,
		"versions":  versions,
		"downloads": downloads,
	})
}

// GET /api/v1/packages/:name/versions
func (h *Handler) listVersions(w http.ResponseWriter, r *http.Request, name string) {
	pkg, err := h.store.GetPackage(name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if pkg == nil {
		writeError(w, http.StatusNotFound, "package not found")
		return
	}

	versions, err := h.store.ListVersions(pkg.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if versions == nil {
		versions = []*models.Version{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"package":  name,
		"versions": versions,
	})
}

// GET /api/v1/packages/:name/:version/dl
func (h *Handler) downloadVersion(w http.ResponseWriter, r *http.Request, name, version string) {
	pkg, err := h.store.GetPackage(name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if pkg == nil {
		writeError(w, http.StatusNotFound, "package not found")
		return
	}

	ver, err := h.store.GetVersion(pkg.ID, version)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if ver == nil {
		writeError(w, http.StatusNotFound, "version not found")
		return
	}

	// Increment download counter (best-effort)
	_ = h.store.IncrementDownload(pkg.ID, ver.ID)

	// Redirect to tarball URL
	http.Redirect(w, r, ver.TarballURL, http.StatusFound)
}

func queryInt(r *http.Request, key string, def int) int {
	if v := r.URL.Query().Get(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			return n
		}
	}
	return def
}

// tokenOwner extracts the username from the Bearer token (simplified: store lookup optional here).
func tokenOwner(r *http.Request) string {
	parts := strings.Split(r.Header.Get("Authorization"), " ")
	if len(parts) == 2 {
		return "api-user"
	}
	return "unknown"
}
