# Registry Deployment Guide

## Prerequisites

- A VPS with Docker + Docker Compose installed (Ubuntu 22.04+ recommended).
- A domain (`packages.kylix.top`) with an A record pointing to your VPS IP.
- Let's Encrypt TLS certificates.

## Quick start

```bash
# 1. SSH into your VPS
ssh root@packages.kylix.top

# 2. Clone the repository
git clone https://github.com/astra-zhao/kylix.git /opt/kylix
cd /opt/kylix/registry/deploy

# 3. Configure secrets
cp .env.example .env
# Edit .env: set POSTGRES_PASSWORD and REGISTRY_AUTH_TOKEN_PEPPER to random strings.

# 4. Build and start
make build
make up

# 5. Run database migrations
make migrate

# 6. Test the API
curl https://packages.kylix.top/api/v1/packages
# Expected: {"ok":true,"packages":[]}

# 7. Publish a test package
kylix publish --registry=https://packages.kylix.top --token=YOUR_ADMIN_TOKEN
```

## TLS with Let's Encrypt (using Certbot)

```bash
apt install certbot python3-certbot-nginx
certbot --nginx -d packages.kylix.top
```

## Production checklist

- [ ] Secrets rotated (`.env`).
- [ ] Docker Compose exposed on a non-standard port behind nginx (already in nginx.conf).
- [ ] Prometheus / Grafana monitoring for `registry:8080/metrics` (TBD).
- [ ] Regular PostgreSQL backups (`pg_dump` cron job).
- [ ] Rate limiting in nginx config (uncomment `limit_req` zones).

## Troubleshooting

- `docker compose logs` — view registry + db logs.
- `curl -i http://localhost:8080/api/v1/packages` — direct health check.
- If `kylix publish` fails with 401, regenerate your token via the admin API.