# Manager Docker Deployment

The AI Together Manager UI is packaged as a Docker container with nginx serving the React SPA and acting as a reverse proxy to the backend server.

## Quick Start

See the main [Docker Deployment Guide](../DOCKER.md) for complete deployment instructions.

## Architecture

The manager container serves dual purposes:

1. **Static File Server** - Serves the React SPA (built with Vite)
2. **Reverse Proxy** - Proxies API requests to the backend server

### nginx Configuration

The nginx configuration (`nginx.conf`) includes:

- **`/`** - Serves React SPA with SPA fallback routing
- **`/api/*`** - Proxies to `http://server:9080/api/*`
- **`/auth/*`** - Proxies to `http://server:9080/auth/*`
- **`/health`** - Proxies to `http://server:9080/health`
- **Static asset caching** - 1 year cache for JS/CSS/images
- **Gzip compression** - Enabled for all text content
- **Security headers** - X-Frame-Options, X-Content-Type-Options, etc.

## Development

### Build Docker Image Locally

```bash
# From the manager directory
docker build -t code-together-manager .
```

### Run Locally

The manager requires the server to be running:

```bash
# Run with server using Docker Compose
cd ..
docker-compose up -d
```

The manager will be accessible at http://localhost:8080

### Local Development (without Docker)

For local development with hot reload:

```bash
# Install dependencies
pnpm install

# Run dev server (proxies to server on port 9080)
pnpm dev
```

This runs Vite dev server on port 3000 with HMR (Hot Module Replacement).

## Production Build

The Dockerfile uses a multi-stage build:

**Stage 1: Build**
- Uses `node:alpine` (latest stable Node.js)
- Installs pnpm and dependencies
- Builds the Vite app (`pnpm build`)
- Outputs to `/app/dist`

**Stage 2: Runtime**
- Uses `nginx:alpine`
- Copies built assets from stage 1
- Copies nginx configuration
- Final image size: ~25MB

## Environment Variables

The manager is a static SPA and doesn't use runtime environment variables. All configuration is baked into the build.

API endpoints are determined by nginx proxy configuration (not hardcoded in the app).

## nginx Proxy Headers

The nginx proxy forwards these headers to the backend:

- `Host` - Original request host
- `X-Real-IP` - Client IP address
- `X-Forwarded-For` - Proxy chain
- `X-Forwarded-Proto` - Original protocol (http/https)

## Troubleshooting

### Cannot access API endpoints

**Symptom:** UI loads but API calls fail with 502 or 504 errors

**Solution:**
```bash
# Check if server is running
docker-compose ps

# Check server logs
docker-compose logs server

# Verify server is healthy
docker exec code-together-server wget -O- http://localhost:9080/health
```

### SPA routing doesn't work (404 on refresh)

**Symptom:** Direct navigation to routes like `/dashboard` returns 404

**Solution:** This is handled by nginx configuration. Check that `nginx.conf` includes:
```nginx
location / {
    try_files $uri $uri/ /index.html;
}
```

### Static assets not loading

**Symptom:** Blank page, console errors for missing JS/CSS

**Solution:**
```bash
# Check if build succeeded
docker logs code-together-manager

# Verify files exist in container
docker exec code-together-manager ls -la /usr/share/nginx/html
```

## Documentation

- [Main Docker Guide](../DOCKER.md) - Complete Docker deployment guide
- [CLAUDE.md](./CLAUDE.md) - Manager development documentation
- [server/CLAUDE.md](../server/CLAUDE.md) - Server API documentation
