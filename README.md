# Polaris

## Quick Start

```bash
# Start PostgreSQL + Redis
docker compose up -d

# Apply migrations (auto-runs on first start, or manually)
psql postgres://polaris:polaris@localhost:5432/polaris -f internal/db/migrations/001_initial.sql

# Start the server
make dev
```

Visit http://localhost:8080/health — expected response:

```json
{"postgres":true,"redis":true,"status":"ok"}
```

The API is JSON-only (no frontend HTML yet). Test auth with curl:

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","email":"alice@test.com","password":"password123"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@test.com","password":"password123"}'
```

Default port is `8080`. Set `LISTEN_ADDR=:9090` to change it.
