# Installation

This guide walks you through installing a Threadlight server from scratch.

## Requirements

Before you start, you'll need:

- A computer running Linux (a cheap single-board computer like an Orange Pi or Raspberry Pi works fine)
- Go 1.23 or later (for building the server)
- PostgreSQL 16 or later (for the database)
- Redis 7 or later (for caching and session management)
- Node.js 20+ and pnpm (for building the web frontend)
- A domain name (optional, but recommended for public access)

## Step 1: Install Dependencies

### PostgreSQL

Install PostgreSQL using your system's package manager:

```bash
# On Debian/Ubuntu
sudo apt update
sudo apt install postgresql postgresql-client

# Start PostgreSQL
sudo systemctl enable --now postgresql
```

Create a database and user for Threadlight:

```bash
sudo -u postgres psql
```

Inside the PostgreSQL prompt:

```sql
CREATE USER polaris WITH PASSWORD 'your-strong-password';
CREATE DATABASE polaris OWNER polaris;
\q
```

### Redis

```bash
# On Debian/Ubuntu
sudo apt install redis-server
sudo systemctl enable --now redis-server
```

### Node.js and pnpm

```bash
# Install Node.js (using NodeSource or your package manager)
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install nodejs

# Install pnpm
sudo npm install -g pnpm
```

### Go

```bash
# Download and install Go
wget https://go.dev/dl/go1.23.0.linux-arm64.tar.gz
sudo tar -C /usr/local -xzf go1.23.0.linux-arm64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

> Note: On ARM machines (like Orange Pi), use the `arm64` version of Go. On regular PCs, use `amd64`.

## Step 2: Get the Code

```bash
git clone https://github.com/anomalyco/polaris.git
cd polaris
```

## Step 3: Build the Frontend

```bash
cd web
pnpm install
pnpm build
cd ..
```

This creates a `web/dist` folder with the compiled frontend files.

## Step 4: Build the Server

```bash
go build -o polaris ./cmd/polaris
```

This creates a binary called `polaris` in the current directory.

## Step 5: Set Up Environment Variables

Create a file called `.env` in the project directory:

```bash
# Database connection
DATABASE_URL=postgres://polaris:your-strong-password@localhost:5432/polaris?sslmode=disable

# Redis connection
REDIS_ADDR=localhost:6379

# Server address
LISTEN_ADDR=:8080

# JWT secret (change this to a random string!)
JWT_SECRET=replace-with-a-random-secret-string

# Admin account (created automatically on first start)
ADMIN_EMAIL=you@example.com
ADMIN_PASSWORD=your-admin-password

# Instance info (shown on the About page)
  INSTANCE_NAME=My Threadlight Server
INSTANCE_DESCRIPTION=A friendly community for thoughtful people.
