# Server Docker Deployment

The Code Together server can be easily deployed using Docker.

## Quick Start

See the main [Docker Deployment Guide](../DOCKER.md) for complete instructions.

## Development

### Build Docker Image Locally

**Important:** The build context must be the repository root (not the server directory) due to the shared module dependency.

```bash
# From the repository root
docker build -f server/Dockerfile -t code-together-server .

# Or use docker-compose
docker-compose build server
```

### Run Locally

```bash
docker run -d \
  -p 8080:9080 \
  -e DATABASE_URL="postgres://user:pass@host:5432/db?sslmode=disable" \
  -e JWT_SECRET="your_secret" \
  code-together-server
```

### Using Docker Compose (Development)

```bash
# From the repository root
docker-compose up -d
```

This will:
- Build the server from source
- Start PostgreSQL database
- Run migrations automatically
- Create default admin on first run

## Production

For production deployment, use `docker-compose.prod.yml` which pulls pre-built images from GitHub Container Registry:

```bash
# From the repository root
docker-compose -f docker-compose.prod.yml up -d
```

See [DOCKER.md](../DOCKER.md) for detailed production deployment instructions.

## Default Admin Account

On first run, if the database is empty, a default admin account is automatically created:

- **Email:** `admin@codetogether.local`
- **Password:** Randomly generated (8 characters, shown in logs)

The credentials are:
1. Displayed in the server logs
2. Saved to `/app/admin_credentials.txt` inside the container

**⚠️ IMPORTANT:** Change the default password immediately after first login!

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | `9080` | Server port |
| `DEBUG` | No | `false` | Enable debug mode |
| `DATABASE_URL` | Yes | - | PostgreSQL connection string |
| `JWT_SECRET` | Yes | - | Secret for JWT token signing |
| `ADMIN_EMAIL` | No | `admin@codetogether.local` | Override default admin email |
| `ADMIN_NAME` | No | `Administrator` | Override default admin name |
| `ADMIN_PASSWORD` | No | Random | Override default admin password |

## Health Check

The server includes a health check endpoint:

```bash
curl http://localhost:9080/health
```

Returns:
```json
{
  "status": "ok",
  "database": "connected",
  "version": "1.0.0"
}
```

## Documentation

- [Main Docker Guide](../DOCKER.md) - Complete Docker deployment guide
- [CLAUDE.md](./CLAUDE.md) - Server development documentation
- [README.md](./README.md) - Server overview and RBAC permissions
