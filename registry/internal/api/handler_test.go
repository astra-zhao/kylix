package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"kylix/registry/internal/api"
	"kylix/registry/internal/auth"
	"kylix/registry/internal/db"
	"kylix/registry/internal/models"
)

// setupTestServer creates an in-memory SQLite server for testing.
func setupTestServer(t *testing.T) (*httptest.Server, db.Store) {
	t.Helper()
	store, err := db.NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("setup store: %v", err)
	}
	if err := store.Migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	authSvc := auth.NewService(store)
	handler := api.NewHandler(store, authSvc)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/packages", handler.HandlePackages)
	mux.HandleFunc("/api/v1/packages/", handler.HandlePackageDetail)

	return httptest.NewServer(mux), store
}

// createTestToken inserts a test user and returns their token.
func createTestToken(t *testing.T, store db.Store) string {
	t.Helper()
	token, err := auth.GenerateToken()
	if err != nil {
		t.Fatal(err)
	}
	user := &models.User{
		GitHubID: "test-123",
		Username: "testuser",
		APIToken: token,
	}
	if err := store.CreateUser(user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return token
}

func TestListPackages_Empty(t *testing.T) {
	srv, _ := setupTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/packages")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	pkgs := result["packages"].([]any)
	if len(pkgs) != 0 {
		t.Errorf("expected empty list, got %d packages", len(pkgs))
	}
}

func TestPublishPackage_RequiresAuth(t *testing.T) {
	srv, _ := setupTestServer(t)
	defer srv.Close()

	// POST without token should return 401
	body := `{"name":"testpkg","version":"1.0.0","tarball_url":"http://example.com/pkg.tar.gz"}`
	resp, err := http.Post(srv.URL+"/api/v1/packages", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestPublishAndRetrievePackage(t *testing.T) {
	srv, store := setupTestServer(t)
	defer srv.Close()

	token := createTestToken(t, store)

	// Publish a package
	body := `{"name":"kylix-http","version":"1.0.0","description":"HTTP client for Kylix","tarball_url":"http://example.com/kylix-http-1.0.0.tar.gz"}`
	req, _ := http.NewRequest("POST", srv.URL+"/api/v1/packages", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}

	// Retrieve the package
	resp2, err := http.Get(srv.URL + "/api/v1/packages/kylix-http")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp2.StatusCode)
	}

	var result map[string]any
	json.NewDecoder(resp2.Body).Decode(&result)
	pkg := result["package"].(map[string]any)
	if pkg["name"] != "kylix-http" {
		t.Errorf("expected name 'kylix-http', got %v", pkg["name"])
	}
}

func TestPublishDuplicateVersion(t *testing.T) {
	srv, store := setupTestServer(t)
	defer srv.Close()

	token := createTestToken(t, store)

	body := `{"name":"mypkg","version":"1.0.0","tarball_url":"http://example.com/mypkg.tar.gz"}`
	postPkg := func() *http.Response {
		req, _ := http.NewRequest("POST", srv.URL+"/api/v1/packages", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		resp, _ := http.DefaultClient.Do(req)
		return resp
	}

	r1 := postPkg()
	r1.Body.Close()
	if r1.StatusCode != http.StatusCreated {
		t.Fatalf("first publish: expected 201, got %d", r1.StatusCode)
	}

	r2 := postPkg()
	r2.Body.Close()
	if r2.StatusCode != http.StatusConflict {
		t.Errorf("duplicate publish: expected 409, got %d", r2.StatusCode)
	}
}

func TestListVersions(t *testing.T) {
	srv, store := setupTestServer(t)
	defer srv.Close()

	token := createTestToken(t, store)

	// Publish two versions
	for _, ver := range []string{"1.0.0", "1.1.0"} {
		body := `{"name":"semver-pkg","version":"` + ver + `","tarball_url":"http://example.com/pkg-` + ver + `.tar.gz"}`
		req, _ := http.NewRequest("POST", srv.URL+"/api/v1/packages", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		resp, _ := http.DefaultClient.Do(req)
		resp.Body.Close()
	}

	resp, err := http.Get(srv.URL + "/api/v1/packages/semver-pkg/versions")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	versions := result["versions"].([]any)
	if len(versions) != 2 {
		t.Errorf("expected 2 versions, got %d", len(versions))
	}
}

func TestSearchPackages(t *testing.T) {
	srv, store := setupTestServer(t)
	defer srv.Close()

	token := createTestToken(t, store)

	// Publish two packages
	for _, name := range []string{"kylix-json", "kylix-http"} {
		body := `{"name":"` + name + `","version":"1.0.0","description":"` + name + ` package","tarball_url":"http://example.com/` + name + `.tar.gz"}`
		req, _ := http.NewRequest("POST", srv.URL+"/api/v1/packages", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		resp, _ := http.DefaultClient.Do(req)
		resp.Body.Close()
	}

	// Search for "json"
	resp, err := http.Get(srv.URL + "/api/v1/packages?q=json")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	pkgs := result["packages"].([]any)
	if len(pkgs) != 1 {
		t.Errorf("expected 1 result for 'json', got %d", len(pkgs))
	}
}

func TestGetNonexistentPackage(t *testing.T) {
	srv, _ := setupTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/packages/doesnotexist")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}
