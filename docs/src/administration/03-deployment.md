# Deployment

This page covers how to deploy Polaris for public access.

## Running as a System Service

The best way to run Polaris in production is as a systemd service so it starts automatically when your server boots.

Create a file at `/etc/systemd/system/polaris.service`:

```ini
[Unit]
Description=Polaris Social Network
After=network.target postgresql.service redis-server.service

[Service]
Type=simple
User=polaris
WorkingDirectory=/opt/polaris
ExecStart=/opt/polaris/polaris
Restart=always
RestartSec=5
EnvironmentFile=/opt/polaris/.env

[Install]
WantedBy=multi-user.target
```

Then enable and start the service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now polaris
```

## Making Polaris Accessible from the Internet

There are several ways to make your Polaris server accessible from the internet.

### Option 1: Direct Connection (with a Public IP)

If your server has a public IP address:

1. Set up a domain name pointing to that IP
2. Set up a reverse proxy (like Nginx or Caddy) to handle SSL certificates
3. Forward port 443 (HTTPS) from your router to your server

Example Nginx config:

```nginx
server {
    listen 443 ssl;
    server_name polaris.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/polaris.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/polaris.yourdomain.com/privkey.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Option 2: Cloudflare Tunnel (Recommended for Home Servers)

If your server is on a home network without a public IP (like an Orange Pi), Cloudflare Tunnel is the easiest option.

**Step 1: Install cloudflared**

```bash
# Download for ARM (Orange Pi)
wget https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-arm64
sudo mv cloudflared-linux-arm64 /usr/local/bin/cloudflared
sudo chmod +x /usr/local/bin/cloudflared
```

**Step 2: Authenticate**

```bash
cloudflared tunnel login
```

This opens a browser window where you log in to Cloudflare and select your domain.

**Step 3: Create a tunnel**

```bash
cloudflared tunnel create polaris
```

This creates a tunnel with a unique ID.

**Step 4: Configure the tunnel**

Create a config file at `~/.cloudflared/config.yml`:

```yaml
tunnel: your-tunnel-id
credentials-file: /home/youruser/.cloudflared/your-tunnel-id.json

ingress:
  - hostname: polaris.yourdomain.com
    service: http://localhost:8080
  # Documentation hosted as a subdomain
  - hostname: docs.yourdomain.com
    service: http://localhost:9000
  - service: http_status:404
```

**Step 5: Add DNS records**

```bash
cloudflared tunnel route dns polaris polaris.yourdomain.com
cloudflared tunnel route dns polaris docs.yourdomain.com
```

**Step 6: Run the tunnel as a service**

```bash
sudo cloudflared service install
sudo systemctl enable --now cloudflared
```

Now your Polaris server is accessible at `https://polaris.yourdomain.com`!

### Serving Documentation as a Subdomain

If you want to serve the documentation from a separate subdomain (like `docs.yourdomain.com`), you have two options:

#### Option A: Serve Docs from a Separate Web Server

1. Copy the `docs/` directory to your server
2. Install a simple static file server:

```bash
# Using Python (simple, no installation needed for Python 3)
cd /path/to/docs
python3 -m http.server 9000

# Or using nginx to serve the built mdbook output
```

3. Add a second ingress rule in your Cloudflare tunnel config (as shown above)

#### Option B: Build and Serve with mdbook

```bash
# Install mdbook
cargo install mdbook

# Build the documentation
cd /path/to/polaris/docs
mdbook build

# The output is in docs/book/
# Serve it with any static file server
cd book
python3 -m http.server 9000
```

Then in your Cloudflare tunnel config, point `docs.yourdomain.com` to `http://localhost:9000`.

## Docker Deployment

A Dockerfile is not yet provided, but you can create one easily:

```dockerfile
FROM golang:1.23 AS builder
WORKDIR /build
COPY . .
RUN go build -o polaris ./cmd/polaris

FROM node:20 AS frontend-builder
WORKDIR /build
COPY web/ ./web/
RUN cd web && pnpm install && pnpm build

FROM debian:bookworm-slim
RUN apt update && apt install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /build/polaris /usr/local/bin/
COPY --from=frontend-builder /build/web/dist/ /var/www/polaris/
COPY docs/ /opt/polaris/docs/
CMD ["polaris"]
```

## Security Checklist

Before opening your server to the public:

- [ ] Change the JWT secret to a strong random value
- [ ] Use a strong database password
- [ ] Enable HTTPS (through a reverse proxy or Cloudflare tunnel)
- [ ] Keep your server and dependencies updated
- [ ] Regularly back up your database
- [ ] Monitor your server's resource usage
