# Docker Deployment Guide

This guide explains how to deploy the complete AI Together platform (Server API + Manager UI) using Docker and Docker Compose.

# 🇨🇳 中国用户专用部署

如果您在中国大陆，使用专用配置文件可获得更快的部署速度：

```bash
# 下载中国用户专用配置（使用 docker.gh-proxy.com 镜像代理）
curl -O https://raw.githubusercontent.com/shtdu/code-together/main/docker-compose.cn.yml
curl -O https://raw.githubusercontent.com/shtdu/code-together/main/.env.example

# 配置并启动
cp .env.example .env
nano .env  # 设置 JWT_SECRET
docker-compose -f docker-compose.cn.yml up -d
```

**详细文档:** 请参考 [DOCKER_CN.md](./DOCKER_CN.md) 获取完整的中国用户部署指南

---

> **For users in China:** Use `docker-compose.cn.yml` with `docker.gh-proxy.com` mirror for faster deployment. See [DOCKER_CN.md](./DOCKER_CN.md) for details.

## Architecture

The deployment includes three services accessible through a **single HTTP port**:

```
                    ┌─────────────────────────────────┐
                    │   http://localhost:8080         │
                    │   (Single Entry Point)          │
                    └──────────────┬──────────────────┘
                                   │
                    ┌──────────────▼──────────────────┐
                    │  Manager (nginx)                │
                    │  - Serves UI at /               │
                    │  - Proxies /api/* → Server      │
                    │  - Proxies /auth/* → Server     │
                    └──────────────┬──────────────────┘
                                   │
                    ┌──────────────▼──────────────────┐
                    │  Server (Go API)                │
                    │  - Port 9080 (internal only)    │
                    │  - Business logic & database    │
                    └──────────────┬──────────────────┘
                                   │
                    ┌──────────────▼──────────────────┐
                    │  PostgreSQL Database            │
                    │  - Port 5432 (internal only)    │
                    └─────────────────────────────────┘
```

**Key Points:**
- Only **one port** (8080 by default) is exposed externally
- Manager UI serves the web interface and acts as a reverse proxy
- Server API is only accessible through the manager's proxy
- Database is completely internal

## Quick Start (Production)

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) installed on your machine.
- [Docker Compose](https://docs.docker.com/compose/install/) installed.
- Basic familiarity with terminal/command line usage. for alternatives

### 1. Download Configuration Files

Download the following files from the repository:
- `docker-compose.prod.yml`
- `.env.example`

```bash
# Download files
curl -O https://raw.githubusercontent.com/shtdu/code-together/main/docker-compose.prod.yml
curl -O https://raw.githubusercontent.com/shtdu/code-together/main/.env.example
```

### 2. Configure Environment Variables

```bash
# Copy the example file
cp .env.example .env

# Edit .env with your preferred editor
nano .env
```

**IMPORTANT:** At minimum, you MUST change:
- `JWT_SECRET` - Set to a secure random string (use: `openssl rand -base64 32`)
- `GITHUB_REPOSITORY` - Your GitHub username/repository name

Example `.env`:
```env
JWT_SECRET=your_secure_random_jwt_secret_here
GITHUB_REPOSITORY=shtdu/code-together
IMAGE_TAG=latest
APP_PORT=8080
DB_PASSWORD=your_secure_database_password
```

### 3. Start the Services

```bash
docker-compose -f docker-compose.prod.yml up -d
```

### 4. Access the Application and Setup Admin Account

Open your browser to: **http://localhost:8080**

On first run, you will be automatically redirected to the **Setup Wizard** where you can create your organization and admin account.

**Setup Wizard Steps:**

1. **Organization Information**
   - Enter your organization or team name

2. **Administrator Account**
   - Enter admin email
   - Enter admin name
   - Create a password (minimum 8 characters)
   - **Tip:** Use the "Generate Secure Password" button for a strong random password

3. **Complete**
   - Review your credentials (email and password are displayed)
   - You will be automatically redirected to the login page in 20 seconds
   - Or click "Go to Login Now" to proceed immediately

**Important Notes:**
- Save your admin credentials securely
- The setup wizard only appears on first deployment (when database is empty)
- After setup, you can access the application at:
  - **Manager UI (Dashboard):** http://localhost:8080/
  - **Login Page:** http://localhost:8080/login
  - **API (proxied):** http://localhost:8080/api/*
  - **Health Check:** http://localhost:8080/health

All services are accessible through this single port!

## Configuration Options

### Environment Variables

All configuration is done through environment variables in the `.env` file:

#### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `JWT_SECRET` | Secret for JWT token signing | `openssl rand -base64 32` |
| `GITHUB_REPOSITORY` | Your GitHub repository | `shtdu/code-together` |

#### Optional Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `IMAGE_TAG` | `latest` | Docker image tag (use version tags for production) |
| `APP_PORT` | `8080` | External application port (UI + API) |
| `DEBUG` | `false` | Enable debug mode |
| `DB_USER` | `codetogether` | PostgreSQL username |
| `DB_PASSWORD` | `codetogether` | PostgreSQL password |
| `DB_NAME` | `code_together` | PostgreSQL database name |

## Common Operations

### View Logs

```bash
# All services
docker-compose -f docker-compose.prod.yml logs -f

# Server only
docker-compose -f docker-compose.prod.yml logs -f server

# Database only
docker-compose -f docker-compose.prod.yml logs -f postgres
```

### Stop Services

```bash
docker-compose -f docker-compose.prod.yml stop
```

### Restart Services

```bash
docker-compose -f docker-compose.prod.yml restart
```

### Upgrade to New Version

```bash
# 1. Pull latest image
docker-compose -f docker-compose.prod.yml pull

# 2. Restart services
docker-compose -f docker-compose.prod.yml up -d
```

### Backup Database

**Manual Backup:**

```bash
# Full database backup
docker exec code-together-db pg_dump -U codetogether code_together > backup.sql

# Compressed backup (recommended for large databases)
docker exec code-together-db pg_dump -U codetogether code_together | gzip > backup.sql.gz

# Backup with timestamp
docker exec code-together-db pg_dump -U codetogether code_together > backup_$(date +%Y%m%d_%H%M%S).sql
```

**Automated Backup Script:**

Create `backup.sh`:
```bash
#!/bin/bash
# Backup script for AI Together database

BACKUP_DIR="./backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/codetogether_$TIMESTAMP.sql.gz"

# Create backup directory if not exists
mkdir -p $BACKUP_DIR

# Perform backup
docker exec code-together-db pg_dump -U codetogether code_together | gzip > $BACKUP_FILE

# Keep only last 7 days of backups
find $BACKUP_DIR -name "codetogether_*.sql.gz" -mtime +7 -delete

echo "Backup completed: $BACKUP_FILE"
```

Make it executable and add to cron:
```bash
chmod +x backup.sh

# Add to crontab (daily at 2 AM)
crontab -e
# Add line: 0 2 * * * /path/to/backup.sh
```

### Restore Database

**From uncompressed backup:**
```bash
docker exec -i code-together-db psql -U codetogether code_together < backup.sql
```

**From compressed backup:**
```bash
gunzip < backup.sql.gz | docker exec -i code-together-db psql -U codetogether code_together
```

**Complete restore (drop and recreate):**
```bash
# Stop services first
docker-compose -f docker-compose.prod.yml stop server manager

# Drop and recreate database
docker exec -i code-together-db psql -U codetogether -c "DROP DATABASE IF EXISTS code_together;"
docker exec -i code-together-db psql -U codetogether -c "CREATE DATABASE code_together;"

# Restore from backup
docker exec -i code-together-db psql -U codetogether code_together < backup.sql

# Restart services
docker-compose -f docker-compose.prod.yml start server manager
```

### Export/Import Data Volume

**Backup entire database volume:**
```bash
# Create backup of PostgreSQL data directory
docker run --rm \
  -v code-together_postgres_data:/data \
  -v $(pwd):/backup \
  alpine tar czf /backup/postgres_data_backup.tar.gz -C /data .
```

**Restore database volume:**
```bash
# Stop database first
docker-compose -f docker-compose.prod.yml stop postgres

# Restore volume
docker run --rm \
  -v code-together_postgres_data:/data \
  -v $(pwd):/backup \
  alpine sh -c "cd /data && tar xzf /backup/postgres_data_backup.tar.gz"

# Start database
docker-compose -f docker-compose.prod.yml start postgres
```

### Migrate to Another Server

**On source server:**
```bash
# 1. Backup database
docker exec code-together-db pg_dump -U codetogether code_together | gzip > codetogether.sql.gz

# 2. Backup .env configuration
cp .env env_backup
```

**On target server:**
```bash
# 1. Setup new deployment
curl -O https://raw.githubusercontent.com/shtdu/code-together/main/docker-compose.prod.yml
curl -O https://raw.githubusercontent.com/shtdu/code-together/main/.env.example
cp .env.example .env

# 2. Copy .env from source (or configure manually)
# Make sure JWT_SECRET matches the source server!

# 3. Start services (this creates the database)
docker-compose -f docker-compose.prod.yml up -d

# 4. Wait for database to be ready
sleep 10

# 5. Restore database
gunzip < codetogether.sql.gz | docker exec -i code-together-db psql -U codetogether code_together

# 6. Restart services
docker-compose -f docker-compose.prod.yml restart
```

### Complete Cleanup

⚠️ **WARNING:** This will delete all data!

```bash
docker-compose -f docker-compose.prod.yml down -v
```

## Using Specific Versions

For production, it's recommended to use specific version tags instead of `latest`:

```env
# .env
IMAGE_TAG=v1.0.0
```

Available tags:
- `latest` - Latest build from main branch
- `v1.0.0`, `v1.1.0`, etc. - Specific releases
- `sha-abc123` - Specific commit builds

Check available tags at: https://github.com/shtdu/code-together/pkgs/container/code-together%2Fserver

## Development Mode

For local development with hot reload:

```bash
# Use the development compose file
docker-compose up -d

# This builds from source and enables debug mode
```

## Troubleshooting

### Cannot Pull Image

**Error:** `Error response from daemon: pull access denied`

**Solution:** The image is pulled from GitHub Container Registry. Make sure:
1. The `GITHUB_REPOSITORY` in `.env` matches your repository
2. The repository has GitHub Actions enabled and the image has been built
3. The package visibility is set to public (or you're authenticated)
4. If in China: You may need to use a proxy, mirror, or local build. See [DOCKER_CN.md](./DOCKER_CN.md)

### Setup Wizard Not Appearing

If the setup wizard doesn't appear and you're redirected to login page instead:

**This means the database already has users.** The setup wizard only appears on first deployment when the database is empty.

**Solution:**

1. **If you want to use existing data:** The admin account already exists. Check your records for the credentials or contact your system administrator.

2. **If you want to start fresh (WARNING: destroys all data):**
```bash
# Reset the database completely
docker-compose -f docker-compose.prod.yml down -v
docker-compose -f docker-compose.prod.yml up -d

# Access http://localhost:8080 - you should now see the setup wizard
```

### Database Connection Errors

**Solution:**
```bash
# Check database is healthy
docker-compose -f docker-compose.prod.yml ps

# View database logs
docker-compose -f docker-compose.prod.yml logs postgres

# Ensure DATABASE_URL matches DB credentials
grep DATABASE_URL .env
```

### Port Already in Use

**Error:** `Bind for 0.0.0.0:8080 failed: port is already allocated`

**Solution:**
```bash
# Change APP_PORT in .env
echo "APP_PORT=8090" >> .env

# Or stop the service using port 8080
lsof -ti:8080 | xargs kill -9
```

## Security Best Practices

1. ✅ **Change JWT_SECRET** - Never use the default value
2. ✅ **Change default admin password** - Immediately after first login
3. ✅ **Use strong database password** - Change from default `codetogether`
4. ✅ **Use version tags** - Don't use `latest` in production
5. ✅ **Enable HTTPS** - Use a reverse proxy (nginx, Caddy, Traefik)
6. ✅ **Regular backups** - Set up automated daily backups (see Backup section above)
7. ✅ **Update regularly** - Keep Docker images up to date
8. ✅ **Backup before updates** - Always backup before upgrading

## Service Details

### Images

- **Image:** `genewoo/ai-together-manager`
- **Tag:** `latest` (or specific version tags like `v1.0.0`)
- **Port:** `80` (Internal Nginx port)

### Server

- **Image:** `genewoo/ai-together-server`

### PostgreSQL
- **Image:** `postgres:latest`
- **Port:** 5432 (internal only, not exposed)
- **Function:** Data persistence

## Next Steps

After deployment:

1. **Change admin password** through the Manager UI
2. **Create teams** for your organization
3. **Add team members** and assign roles
4. **Configure providers** for Claude Code, Codex, and OpenCode
5. **Deploy member clients** to team members' machines

## Alternative Deployment Options

### For Users in China

**Quick Start:** Use `docker-compose.cn.yml` with docker.gh-proxy.com mirror:
```bash
docker-compose -f docker-compose.cn.yml up -d
```

See [DOCKER_CN.md](./DOCKER_CN.md) for:
- Using docker.gh-proxy.com mirror (recommended)
- Local build from source
- Docker daemon proxy configuration
- Aliyun Container Registry setup

### Local Build from Source

If you prefer to build locally instead of pulling images:

```bash
git clone https://github.com/shtdu/code-together.git
cd code-together
cp .env.example .env
# Edit .env (set JWT_SECRET)
docker-compose up -d  # Builds locally
```

**Note:** The server requires the `shared/` directory at the repository root due to Go module dependencies.

## Support

- **Issues:** https://github.com/shtdu/code-together/issues
- **Documentation:** See `README.md` and `CLAUDE.md`
- **Server Docs:** See `server/CLAUDE.md`
- **中国用户文档:** See `DOCKER_CN.md`
