package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"kylix/registry/internal/models"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func (s *SQLiteStore) Migrate() error {
	schema := `
CREATE TABLE IF NOT EXISTS packages (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT UNIQUE NOT NULL,
	owner TEXT NOT NULL,
	description TEXT,
	repo_url TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS versions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	package_id INTEGER NOT NULL,
	version TEXT NOT NULL,
	tarball_url TEXT NOT NULL,
	dependencies TEXT,
	published_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY(package_id) REFERENCES packages(id),
	UNIQUE(package_id, version)
);

CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	github_id TEXT UNIQUE,
	username TEXT NOT NULL,
	api_token TEXT UNIQUE NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS downloads (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	package_id INTEGER NOT NULL,
	version_id INTEGER NOT NULL,
	count INTEGER DEFAULT 1,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY(package_id) REFERENCES packages(id),
	FOREIGN KEY(version_id) REFERENCES versions(id),
	UNIQUE(package_id, version_id)
);

CREATE INDEX IF NOT EXISTS idx_packages_name ON packages(name);
CREATE INDEX IF NOT EXISTS idx_versions_package ON versions(package_id);
CREATE INDEX IF NOT EXISTS idx_users_token ON users(api_token);
	`
	_, err := s.db.Exec(schema)
	return err
}

// Packages

func (s *SQLiteStore) CreatePackage(pkg *models.Package) error {
	now := time.Now()
	res, err := s.db.Exec(
		`INSERT INTO packages (name, owner, description, repo_url, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		pkg.Name, pkg.Owner, pkg.Description, pkg.RepoURL, now, now,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	pkg.ID = id
	pkg.CreatedAt = now
	pkg.UpdatedAt = now
	return nil
}

func (s *SQLiteStore) GetPackage(name string) (*models.Package, error) {
	pkg := &models.Package{}
	err := s.db.QueryRow(
		`SELECT id, name, owner, description, repo_url, created_at, updated_at
		 FROM packages WHERE name = ?`,
		name,
	).Scan(&pkg.ID, &pkg.Name, &pkg.Owner, &pkg.Description, &pkg.RepoURL, &pkg.CreatedAt, &pkg.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return pkg, nil
}

func (s *SQLiteStore) ListPackages(query string, limit, offset int) ([]*models.Package, error) {
	var rows *sql.Rows
	var err error

	if query == "" {
		rows, err = s.db.Query(
			`SELECT id, name, owner, description, repo_url, created_at, updated_at
			 FROM packages ORDER BY updated_at DESC LIMIT ? OFFSET ?`,
			limit, offset,
		)
	} else {
		searchQuery := "%" + query + "%"
		rows, err = s.db.Query(
			`SELECT id, name, owner, description, repo_url, created_at, updated_at
			 FROM packages WHERE name LIKE ? OR description LIKE ?
			 ORDER BY updated_at DESC LIMIT ? OFFSET ?`,
			searchQuery, searchQuery, limit, offset,
		)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var packages []*models.Package
	for rows.Next() {
		pkg := &models.Package{}
		if err := rows.Scan(&pkg.ID, &pkg.Name, &pkg.Owner, &pkg.Description, &pkg.RepoURL, &pkg.CreatedAt, &pkg.UpdatedAt); err != nil {
			return nil, err
		}
		packages = append(packages, pkg)
	}
	return packages, rows.Err()
}

func (s *SQLiteStore) UpdatePackage(pkg *models.Package) error {
	pkg.UpdatedAt = time.Now()
	_, err := s.db.Exec(
		`UPDATE packages SET owner = ?, description = ?, repo_url = ?, updated_at = ?
		 WHERE id = ?`,
		pkg.Owner, pkg.Description, pkg.RepoURL, pkg.UpdatedAt, pkg.ID,
	)
	return err
}

// Versions

func (s *SQLiteStore) CreateVersion(ver *models.Version) error {
	now := time.Now()
	res, err := s.db.Exec(
		`INSERT INTO versions (package_id, version, tarball_url, dependencies, published_at)
		 VALUES (?, ?, ?, ?, ?)`,
		ver.PackageID, ver.Version, ver.TarballURL, ver.Dependencies, now,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	ver.ID = id
	ver.PublishedAt = now
	return nil
}

func (s *SQLiteStore) GetVersion(packageID int64, version string) (*models.Version, error) {
	ver := &models.Version{}
	err := s.db.QueryRow(
		`SELECT id, package_id, version, tarball_url, dependencies, published_at
		 FROM versions WHERE package_id = ? AND version = ?`,
		packageID, version,
	).Scan(&ver.ID, &ver.PackageID, &ver.Version, &ver.TarballURL, &ver.Dependencies, &ver.PublishedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return ver, nil
}

func (s *SQLiteStore) ListVersions(packageID int64) ([]*models.Version, error) {
	rows, err := s.db.Query(
		`SELECT id, package_id, version, tarball_url, dependencies, published_at
		 FROM versions WHERE package_id = ? ORDER BY published_at DESC`,
		packageID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []*models.Version
	for rows.Next() {
		ver := &models.Version{}
		if err := rows.Scan(&ver.ID, &ver.PackageID, &ver.Version, &ver.TarballURL, &ver.Dependencies, &ver.PublishedAt); err != nil {
			return nil, err
		}
		versions = append(versions, ver)
	}
	return versions, rows.Err()
}

// Users

func (s *SQLiteStore) CreateUser(user *models.User) error {
	now := time.Now()
	res, err := s.db.Exec(
		`INSERT INTO users (github_id, username, api_token, created_at) VALUES (?, ?, ?, ?)`,
		user.GitHubID, user.Username, user.APIToken, now,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = id
	user.CreatedAt = now
	return nil
}

func (s *SQLiteStore) GetUserByToken(token string) (*models.User, error) {
	user := &models.User{}
	err := s.db.QueryRow(
		`SELECT id, github_id, username, api_token, created_at FROM users WHERE api_token = ?`,
		token,
	).Scan(&user.ID, &user.GitHubID, &user.Username, &user.APIToken, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *SQLiteStore) GetUserByGitHubID(githubID string) (*models.User, error) {
	user := &models.User{}
	err := s.db.QueryRow(
		`SELECT id, github_id, username, api_token, created_at FROM users WHERE github_id = ?`,
		githubID,
	).Scan(&user.ID, &user.GitHubID, &user.Username, &user.APIToken, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Downloads

func (s *SQLiteStore) IncrementDownload(packageID, versionID int64) error {
	_, err := s.db.Exec(
		`INSERT INTO downloads (package_id, version_id, count, updated_at)
		 VALUES (?, ?, 1, ?)
		 ON CONFLICT(package_id, version_id) DO UPDATE SET count = count + 1, updated_at = ?`,
		packageID, versionID, time.Now(), time.Now(),
	)
	return err
}

func (s *SQLiteStore) GetDownloadCount(packageID int64) (int64, error) {
	var total int64
	err := s.db.QueryRow(
		`SELECT COALESCE(SUM(count), 0) FROM downloads WHERE package_id = ?`,
		packageID,
	).Scan(&total)
	return total, err
}

// PostgreSQL stub (TODO: implement when needed)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(connStr string) (*PostgresStore, error) {
	return nil, fmt.Errorf("PostgreSQL support not yet implemented")
}

func (s *PostgresStore) Close() error                          { return nil }
func (s *PostgresStore) Migrate() error                        { return nil }
func (s *PostgresStore) CreatePackage(pkg *models.Package) error { return nil }
func (s *PostgresStore) GetPackage(name string) (*models.Package, error) { return nil, nil }
func (s *PostgresStore) ListPackages(query string, limit, offset int) ([]*models.Package, error) { return nil, nil }
func (s *PostgresStore) UpdatePackage(pkg *models.Package) error { return nil }
func (s *PostgresStore) CreateVersion(ver *models.Version) error { return nil }
func (s *PostgresStore) GetVersion(packageID int64, version string) (*models.Version, error) { return nil, nil }
func (s *PostgresStore) ListVersions(packageID int64) ([]*models.Version, error) { return nil, nil }
func (s *PostgresStore) CreateUser(user *models.User) error { return nil }
func (s *PostgresStore) GetUserByToken(token string) (*models.User, error) { return nil, nil }
func (s *PostgresStore) GetUserByGitHubID(githubID string) (*models.User, error) { return nil, nil }
func (s *PostgresStore) IncrementDownload(packageID, versionID int64) error { return nil }
func (s *PostgresStore) GetDownloadCount(packageID int64) (int64, error) { return 0, nil }
