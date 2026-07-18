# Configuration

All Polaris configuration is done through environment variables. This page explains each one.

## Required Settings

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://polaris:polaris@localhost:5432/polaris?sslmode=disable` |
| `JWT_SECRET` | Secret key for signing login tokens | `dev-secret-change-in-production` (⚠️ change this!) |

## Optional Settings

| Variable | Description | Default |
|----------|-------------|---------|
| `LISTEN_ADDR` | Address and port to listen on | `:8080` |
| `REDIS_ADDR` | Redis server address | `localhost:6379` |
| `ADMIN_EMAIL` | Email for the auto-created admin account | &mdash; |
| `ADMIN_PASSWORD` | Password for the admin account | &mdash; |
| `INSTANCE_NAME` | Name shown on the About page | `Polaris` |
| `INSTANCE_DESCRIPTION` | Description shown on the About page | &mdash; |
| `DOCS_DIR` | Path to the documentation directory | `./docs` |

## Database Configuration

The `DATABASE_URL` follows this format:

```
postgres://username:password@hostname:port/database?options
```

Common options:

- `sslmode=disable` &mdash; no SSL (local development)
- `sslmode=require` &mdash; require SSL (remote connections)

## JWT Secret

The JWT secret is used to sign login tokens. If you change it, all existing sessions will be invalidated and users will need to log in again.

Generate a strong secret:

```bash
openssl rand -base64 32
```

## Admin Account

If `ADMIN_EMAIL` and `ADMIN_PASSWORD` are set, Polaris will create an admin account on first startup (if one doesn't already exist). The admin has:

- Unlimited invites (even if invites are normally required)
- A maximized trust score
- Access to administrative features

## Documentation Directory

Set `DOCS_DIR` to a directory of static HTML files to serve at `/docs/`. If the directory doesn't exist, the docs route is not served.

## Running with Environment Variables

```bash
export DATABASE_URL="postgres://polaris:your-password@localhost:5432/polaris?sslmode=disable"
export JWT_SECRET="your-random-secret"
export LISTEN_ADDR=":8080"
./polaris
```

Or use a `.env` file with a tool like `dotenv` or systemd's `EnvironmentFile`.
