package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"kylix/registry/internal/api"
	"kylix/registry/internal/auth"
	"kylix/registry/internal/db"
)

var templates *template.Template

func main() {
	// Load HTML templates
	_, src, _, _ := runtime.Caller(0)
	webDir := filepath.Join(filepath.Dir(src), "..", "..", "web", "templates")
	var err error
	templates, err = template.ParseGlob(filepath.Join(webDir, "*.html"))
	if err != nil {
		log.Printf("Warning: could not load templates from %s: %v", webDir, err)
	}

	// Database setup
	dbType := getEnv("REGISTRY_DB_TYPE", "sqlite")
	var store db.Store

	switch dbType {
	case "sqlite":
		dbPath := getEnv("REGISTRY_DB_PATH", "./registry.db")
		store, err = db.NewSQLiteStore(dbPath)
		if err != nil {
			log.Fatalf("Failed to open SQLite: %v", err)
		}
		log.Printf("Using SQLite database: %s", dbPath)
	case "postgres":
		pgURL := getEnv("REGISTRY_POSTGRES_URL", "")
		if pgURL == "" {
			log.Fatal("REGISTRY_POSTGRES_URL required for postgres mode")
		}
		store, err = db.NewPostgresStore(pgURL)
		if err != nil {
			log.Fatalf("Failed to connect to PostgreSQL: %v", err)
		}
		log.Println("Using PostgreSQL database")
	default:
		log.Fatalf("Unknown REGISTRY_DB_TYPE: %s", dbType)
	}
	defer store.Close()

	if err := store.Migrate(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	authSvc := auth.NewService(store)
	apiHandler := api.NewHandler(store, authSvc)
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/v1/packages", apiHandler.HandlePackages)
	mux.HandleFunc("/api/v1/packages/", apiHandler.HandlePackageDetail)

	// Web routes
	mux.HandleFunc("/", handleRoot(store))
	mux.HandleFunc("/packages/", handlePackagePage(store))

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	port := getEnv("REGISTRY_PORT", "8080")
	log.Printf("Registry server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func handleRoot(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		if templates == nil {
			fmt.Fprintln(w, "Kylix Package Registry — API: /api/v1/packages")
			return
		}
		templates.ExecuteTemplate(w, "index.html", nil)
	}
}

func handlePackagePage(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/packages/")
		if name == "" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		pkg, err := store.GetPackage(name)
		if err != nil || pkg == nil {
			http.NotFound(w, r)
			return
		}
		versions, _ := store.ListVersions(pkg.ID)

		latestVersion := ""
		if len(versions) > 0 {
			latestVersion = versions[0].Version
		}

		data := struct {
			Name          string
			Owner         string
			Description   string
			RepoURL       string
			LatestVersion string
			Versions      interface{}
			UpdatedAt     time.Time
		}{
			Name:          pkg.Name,
			Owner:         pkg.Owner,
			Description:   pkg.Description,
			RepoURL:       pkg.RepoURL,
			LatestVersion: latestVersion,
			Versions:      versions,
			UpdatedAt:     pkg.UpdatedAt,
		}

		if templates == nil {
			fmt.Fprintf(w, "Package: %s@%s\n", pkg.Name, latestVersion)
			return
		}
		templates.ExecuteTemplate(w, "package.html", data)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
