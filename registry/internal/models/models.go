package models

import "time"

// Package represents a Kylix package in the registry.
type Package struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Owner       string    `json:"owner"`
	Description string    `json:"description"`
	RepoURL     string    `json:"repo_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Version represents a specific version of a package.
type Version struct {
	ID          int64     `json:"id"`
	PackageID   int64     `json:"package_id"`
	Version     string    `json:"version"` // Semantic version (e.g., "1.2.3")
	TarballURL  string    `json:"tarball_url"`
	Dependencies string    `json:"dependencies"` // JSON array of {name, version}
	PublishedAt time.Time `json:"published_at"`
}

// User represents a registry user (for authentication).
type User struct {
	ID        int64     `json:"id"`
	GitHubID  string    `json:"github_id"`
	Username  string    `json:"username"`
	APIToken  string    `json:"api_token"`
	CreatedAt time.Time `json:"created_at"`
}

// Download represents a download event for statistics.
type Download struct {
	ID        int64     `json:"id"`
	PackageID int64     `json:"package_id"`
	VersionID int64     `json:"version_id"`
	Count     int64     `json:"count"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PackageWithVersions is a package with its versions list.
type PackageWithVersions struct {
	Package
	Versions []Version `json:"versions"`
}
