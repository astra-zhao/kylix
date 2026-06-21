# Kylix Package Registry

Official package registry server for Kylix programming language.

## Features

- **Package hosting**: Publish and download Kylix packages
- **Semantic versioning**: Version management with dependency resolution
- **API authentication**: API tokens for CLI + GitHub OAuth for web
- **Search & discovery**: Browse packages at `kylix.top/packages`
- **GitHub mirror**: Cache packages from GitHub repositories

## Architecture

- **Backend**: Go 1.21+ with `net/http` standard library
- **Database**: SQLite (dev/testing) + PostgreSQL (production)
- **Frontend**: htmx + Tailwind CSS (no Node.js build step)
- **CLI integration**: `kylix publish` / `kylix install`

## Directory Structure

```
registry/
├── cmd/registry/         # Server entry point
├── internal/
│   ├── api/             # HTTP handlers
│   ├── auth/            # Authentication (token + OAuth)
│   ├── db/              # Database abstraction
│   ├── models/          # Package metadata models
│   └── mirror/          # GitHub package mirroring
├── web/
│   ├── templates/       # HTML templates
│   └── static/          # CSS/JS assets
└── migrations/          # Database schema migrations
```

## Database Schema

### Core Tables

1. **packages** — Package metadata (name, owner, description, repo URL)
2. **versions** — Version releases (semver, tarball, dependencies, timestamp)
3. **users** — Authentication (GitHub ID, API tokens)
4. **downloads** — Download statistics (package + version)

## API Endpoints (v1)

```
GET  /api/v1/packages                   # Search packages
GET  /api/v1/packages/:name             # Package details
GET  /api/v1/packages/:name/versions    # List versions
POST /api/v1/packages                   # Publish (requires token)
GET  /api/v1/packages/:name/:ver/dl     # Download tarball
```

## Development

```bash
# Start registry server (SQLite, port 8080)
cd registry
go run ./cmd/registry

# Publish a package
kylix publish --registry=http://localhost:8080

# Search packages
curl http://localhost:8080/api/v1/packages?q=json

# Get package details
curl http://localhost:8080/api/v1/packages/jsonutil
```

## Configuration

Environment variables:

- `REGISTRY_DB_TYPE`: `sqlite` (default) or `postgres`
- `REGISTRY_DB_PATH`: SQLite path (default: `./registry.db`)
- `REGISTRY_POSTGRES_URL`: PostgreSQL connection string
- `REGISTRY_PORT`: HTTP port (default: `8080`)
- `REGISTRY_GITHUB_CLIENT_ID`: GitHub OAuth client ID
- `REGISTRY_GITHUB_CLIENT_SECRET`: GitHub OAuth secret

## Production Deployment

See [DEPLOY.md](DEPLOY.md) for production setup with PostgreSQL, HTTPS, and GitHub OAuth.

## License

MIT — same as Kylix compiler.
