package db

import (
	"kylix/registry/internal/models"
)

// Store is the database interface for the registry.
type Store interface {
	// Lifecycle
	Close() error
	Migrate() error

	// Packages
	CreatePackage(pkg *models.Package) error
	GetPackage(name string) (*models.Package, error)
	ListPackages(query string, limit, offset int) ([]*models.Package, error)
	UpdatePackage(pkg *models.Package) error

	// Versions
	CreateVersion(ver *models.Version) error
	GetVersion(packageID int64, version string) (*models.Version, error)
	ListVersions(packageID int64) ([]*models.Version, error)

	// Users
	CreateUser(user *models.User) error
	GetUserByToken(token string) (*models.User, error)
	GetUserByGitHubID(githubID string) (*models.User, error)

	// Downloads
	IncrementDownload(packageID, versionID int64) error
	GetDownloadCount(packageID int64) (int64, error)
}
