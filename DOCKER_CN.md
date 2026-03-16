# Docker 部署指南（中国用户）

本指南专为中国用户提供 AI Together 平台的 Docker 部署说明。

## ⚠️ 网络连接问题

Docker Hub 在中国大陆可能访问缓慢或无法访问。我们提供以下解决方案。

## 🚀 快速开始（推荐方案）

### 使用 docker.gh-proxy.com 镜像代理（最简单！）

**优势:**
- ✅ 无需配置 Docker daemon
- ✅ 稳定快速的镜像代理服务
- ✅ 只需修改 docker-compose 文件
- ✅ 支持自动更新

**步骤:**

```bash
# 1. 下载中国用户专用配置文件
curl -O https://raw.githubusercontent.com/shtdu/code-together/main/docker-compose.cn.yml
curl -O https://raw.githubusercontent.com/shtdu/code-together/main/.env.example

# 2. 配置环境变量
cp .env.example .env
nano .env  # 必须设置 JWT_SECRET

# 3. 启动服务（使用 docker.gh-proxy.com 代理）
docker-compose -f docker-compose.cn.yml up -d

# 4. 访问应用并设置管理员账号
open http://localhost:8080
# 首次运行时会自动跳转到设置向导，您可以在其中创建组织和管理员账号
```

**工作原理:**
- GitHub 镜像: `ghcr.io/shtdu/code-together/server:latest` → `docker.gh-proxy.com/ghcr.io/shtdu/code-together/server:latest`
- Docker Hub 镜像: `postgres:latest` → `docker.gh-proxy.com/library/postgres:latest`
- docker.gh-proxy.com 会自动从源镜像仓库拉取并缓存镜像

## 其他解决方案

### 方案 1: 使用 Docker Daemon 代理

配置 Docker 使用代理服务器来拉取镜像。

#### Linux (systemd)

```bash
# 1. 创建 Docker 服务的配置目录
sudo mkdir -p /etc/systemd/system/docker.service.d

# 2. 创建代理配置文件
sudo tee /etc/systemd/system/docker.service.d/http-proxy.conf > /dev/null <<EOF
[Service]
Environment="HTTP_PROXY=http://proxy.example.com:8080"
Environment="HTTPS_PROXY=http://proxy.example.com:8080"
Environment="NO_PROXY=localhost,127.0.0.1"
EOF

# 3. 重新加载配置并重启 Docker
sudo systemctl daemon-reload
sudo systemctl restart docker

# 4. 验证配置
sudo systemctl show --property=Environment docker
```

#### macOS (Docker Desktop)

```
1. 打开 Docker Desktop
2. 点击设置图标 (齿轮) → Resources → Proxies
3. 选择 "Manual proxy configuration"
4. 填写代理信息：
   - Web Server (HTTP): http://proxy.example.com:8080
   - Secure Web Server (HTTPS): http://proxy.example.com:8080
5. 点击 "Apply & Restart"
```

#### Windows (Docker Desktop)

```
1. 打开 Docker Desktop
2. 点击设置图标 → Resources → Proxies
3. 选择 "Manual proxy configuration"
4. 填写代理信息
5. 点击 "Apply & Restart"
```

### 方案 2: 使用 Docker 镜像加速器

配置 Docker 使用中国镜像加速服务。

#### 配置镜像加速器

编辑或创建 `/etc/docker/daemon.json`:

```json
{
  "registry-mirrors": [
    "https://docker.mirrors.ustc.edu.cn",
    "https://hub-mirror.c.163.com",
    "https://mirror.ccs.tencentyun.com"
  ]
}
```

然后重启 Docker:

```bash
sudo systemctl daemon-reload
sudo systemctl restart docker
```

**注意:** 镜像加速器主要用于 Docker Hub，对 ghcr.io 效果有限。

1. **手动拉取并导出镜像**:

   ```bash
   # 在可访问 Docker Hub 的服务器上
   docker pull genewoo/ai-together-server:latest
   docker pull genewoo/ai-together-manager:latest

   # 保存为文件
   docker save genewoo/ai-together-server:latest | gzip > server.tar.gz
   docker save genewoo/ai-together-manager:latest | gzip > manager.tar.gz
   ```

### 方案 4: 使用阿里云容器镜像服务（推荐中国用户）

您可以将镜像同步到阿里云容器镜像服务：

#### 1. 在阿里云创建镜像仓库

```
1. 访问 https://cr.console.aliyun.com/
2. 创建命名空间: code-together
3. 创建镜像仓库: server 和 manager
4. 设置为公开仓库
```

#### 2. 同步镜像到阿里云

```bash
# 从 Docker Hub 拉取
docker pull genewoo/ai-together-server:latest
docker pull genewoo/ai-together-manager:latest

# 重新标记
docker tag genewoo/ai-together-server:latest \
  registry.cn-hangzhou.aliyuncs.com/你的命名空间/server:latest

docker tag genewoo/ai-together-manager:latest \
  registry.cn-hangzhou.aliyuncs.com/你的命名空间/manager:latest

# 推送到阿里云
docker login --username=你的阿里云账号 registry.cn-hangzhou.aliyuncs.com
docker push registry.cn-hangzhou.aliyuncs.com/你的命名空间/server:latest
docker push registry.cn-hangzhou.aliyuncs.com/你的命名空间/manager:latest
```

#### 3. 修改 docker-compose.prod.yml

创建 `docker-compose.aliyun.yml`:

```yaml
version: "3.8"

services:
  postgres:
    image: postgres:latest
    container_name: code-together-db
    restart: unless-stopped
    environment:
      POSTGRES_USER: ${DB_USER:-codetogether}
      POSTGRES_PASSWORD: ${DB_PASSWORD:-codetogether}
      POSTGRES_DB: ${DB_NAME:-code_together}
    volumes:
      - postgres_data:/var/lib/postgresql
    networks:
      - code-together-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER:-codetogether} -d ${DB_NAME:-code_together}"]
      interval: 5s
      timeout: 5s
      retries: 5

  server:
    # 使用阿里云镜像
    image: registry.cn-hangzhou.aliyuncs.com/你的命名空间/server:${IMAGE_TAG:-latest}
    container_name: code-together-server
    restart: unless-stopped
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      PORT: 9080
      DEBUG: ${DEBUG:-false}
      DATABASE_URL: postgres://${DB_USER:-codetogether}:${DB_PASSWORD:-codetogether}@postgres:5432/${DB_NAME:-code_together}?sslmode=disable
      JWT_SECRET: ${JWT_SECRET:?JWT_SECRET is required}
    networks:
      - code-together-network
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:9080/health"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s

  manager:
    # 使用阿里云镜像
    image: registry.cn-hangzhou.aliyuncs.com/你的命名空间/manager:${IMAGE_TAG:-latest}
    container_name: code-together-manager
    restart: unless-stopped
    depends_on:
      server:
        condition: service_healthy
    networks:
      - code-together-network
    ports:
      - "${APP_PORT:-8080}:80"
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost/"]
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 5s

volumes:
  postgres_data:
    driver: local

networks:
  code-together-network:
    driver: bridge
```

#### 4. 使用阿里云镜像部署

```bash
docker-compose -f docker-compose.aliyun.yml up -d
```

### 方案 5: 本地构建（最可靠）

直接从源代码构建，无需拉取镜像：

```bash
# 1. 克隆仓库
git clone https://github.com/shtdu/code-together.git
cd code-together

# 2. 配置环境变量
cp .env.example .env
nano .env  # 设置 JWT_SECRET

# 3. 使用开发版 docker-compose（本地构建）
docker-compose up -d
```

这种方式会在本地构建镜像，不需要从任何仓库拉取。

## 常用部署命令

### 使用 docker.gh-proxy.com（推荐）

```bash
# 启动
docker-compose -f docker-compose.cn.yml up -d

# 查看日志
docker-compose -f docker-compose.cn.yml logs -f

# 停止
docker-compose -f docker-compose.cn.yml stop

# 重启
docker-compose -f docker-compose.cn.yml restart

# 更新镜像
docker-compose -f docker-compose.cn.yml pull
docker-compose -f docker-compose.cn.yml up -d

# 完全清除（包括数据）
docker-compose -f docker-compose.cn.yml down -v
```

## 数据库备份与恢复

### 手动备份

```bash
# 完整数据库备份
docker exec code-together-db pg_dump -U codetogether code_together > backup.sql

# 压缩备份（推荐用于大型数据库）
docker exec code-together-db pg_dump -U codetogether code_together | gzip > backup.sql.gz

# 带时间戳的备份
docker exec code-together-db pg_dump -U codetogether code_together > backup_$(date +%Y%m%d_%H%M%S).sql
```

### 自动备份脚本

创建 `backup.sh`：
```bash
#!/bin/bash
# AI Together 数据库备份脚本

BACKUP_DIR="./backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/codetogether_$TIMESTAMP.sql.gz"

# 创建备份目录
mkdir -p $BACKUP_DIR

# 执行备份
docker exec code-together-db pg_dump -U codetogether code_together | gzip > $BACKUP_FILE

# 只保留最近 7 天的备份
find $BACKUP_DIR -name "codetogether_*.sql.gz" -mtime +7 -delete

echo "备份完成: $BACKUP_FILE"
```

设置定时任务：
```bash
chmod +x backup.sh

# 添加到 crontab（每天凌晨 2 点执行）
crontab -e
# 添加一行: 0 2 * * * /path/to/backup.sh
```

### 恢复数据库

**从未压缩的备份恢复:**
```bash
docker exec -i code-together-db psql -U codetogether code_together < backup.sql
```

**从压缩备份恢复:**
```bash
gunzip < backup.sql.gz | docker exec -i code-together-db psql -U codetogether code_together
```

**完全恢复（删除并重建数据库）:**
```bash
# 1. 停止服务
docker-compose -f docker-compose.cn.yml stop server manager

# 2. 删除并重建数据库
docker exec -i code-together-db psql -U codetogether -c "DROP DATABASE IF EXISTS code_together;"
docker exec -i code-together-db psql -U codetogether -c "CREATE DATABASE code_together;"

# 3. 从备份恢复
docker exec -i code-together-db psql -U codetogether code_together < backup.sql

# 4. 重启服务
docker-compose -f docker-compose.cn.yml start server manager
```

### 备份整个数据卷

**备份 PostgreSQL 数据目录:**
```bash
docker run --rm \
  -v code-together_postgres_data:/data \
  -v $(pwd):/backup \
  alpine tar czf /backup/postgres_data_backup.tar.gz -C /data .
```

**恢复数据卷:**
```bash
# 停止数据库
docker-compose -f docker-compose.cn.yml stop postgres

# 恢复数据卷
docker run --rm \
  -v code-together_postgres_data:/data \
  -v $(pwd):/backup \
  alpine sh -c "cd /data && tar xzf /backup/postgres_data_backup.tar.gz"

# 启动数据库
docker-compose -f docker-compose.cn.yml start postgres
```

### 迁移到另一台服务器

**在源服务器上:**
```bash
# 1. 备份数据库
docker exec code-together-db pg_dump -U codetogether code_together | gzip > codetogether.sql.gz

# 2. 备份配置文件
cp .env env_backup
```

**在目标服务器上:**
```bash
# 1. 下载配置文件
curl -O https://raw.githubusercontent.com/shtdu/code-together/main/docker-compose.cn.yml
curl -O https://raw.githubusercontent.com/shtdu/code-together/main/.env.example

# 2. 配置环境变量（从源服务器复制或手动配置）
cp .env.example .env
# 注意：JWT_SECRET 必须与源服务器一致！

# 3. 启动服务（创建数据库）
docker-compose -f docker-compose.cn.yml up -d

# 4. 等待数据库准备就绪
sleep 10

# 5. 恢复数据库
gunzip < codetogether.sql.gz | docker exec -i code-together-db psql -U codetogether code_together

# 6. 重启服务
docker-compose -f docker-compose.cn.yml restart
```

### 本地构建

```bash
# 克隆项目
git clone https://github.com/shtdu/code-together.git
cd code-together

# 配置
cp .env.example .env
nano .env  # 设置 JWT_SECRET

# 启动（本地构建）
docker-compose up -d

# 访问
open http://localhost:8080
```

**注意:** 服务器构建需要项目根目录下的 `shared/` 目录（Go 模块依赖）。

## 常见问题

### Q: 拉取镜像超时怎么办？

**A:** 按优先级尝试以下方案：
1. **使用 docker.gh-proxy.com（推荐）** - 使用 `docker-compose.cn.yml`
2. **使用本地构建** - 使用 `docker-compose.yml`
3. 配置 Docker daemon 代理
4. 使用阿里云镜像（企业用户）

### Q: GitHub 无法访问怎么办？

**A:** 如果无法访问 GitHub：
1. 使用代理或 VPN
2. 使用 GitHub 镜像站点（如 gitee.com）
3. 让能访问的朋友帮忙下载源码并传输

### Q: 已经部署了，如何更新镜像？

**A:**

使用 docker.gh-proxy.com（推荐）:
```bash
docker-compose -f docker-compose.cn.yml pull
docker-compose -f docker-compose.cn.yml up -d
```

本地构建:
```bash
git pull
docker-compose up -d --build
```

使用阿里云:
```bash
docker-compose -f docker-compose.aliyun.yml pull
docker-compose -f docker-compose.aliyun.yml up -d
```

## 性能优化建议

### 1. 使用 docker.gh-proxy.com（已包含在 docker-compose.cn.yml 中）

`docker-compose.cn.yml` 已经配置了所有镜像使用 docker.gh-proxy.com：
- PostgreSQL: `docker.gh-proxy.com/library/postgres:latest`
- Server: `docker.gh-proxy.com/ghcr.io/shtdu/code-together/server:latest`
- Manager: `docker.gh-proxy.com/ghcr.io/shtdu/code-together/manager:latest`

无需额外配置！

### 2. 设置 npm 国内镜像（如果本地构建）

在本地构建前设置 npm 镜像：

```bash
# 设置淘宝镜像
npm config set registry https://registry.npmmirror.com
```

### 3. 设置 Go 模块代理（如果本地构建）

在本地构建前设置 Go 代理：

```bash
# 设置 GOPROXY
export GOPROXY=https://goproxy.cn,direct
```

## 生产环境建议

对于中国大陆的生产环境部署，按优先级推荐：

### 1. 使用 docker.gh-proxy.com（最简单）
```bash
docker-compose -f docker-compose.cn.yml up -d
```
**优点:**
- ✅ 配置简单，一行命令
- ✅ 稳定的镜像代理服务
- ✅ 自动同步更新
- ✅ 无需额外账号或配置

### 2. 本地构建（最可靠）
```bash
git clone https://github.com/shtdu/code-together.git
docker-compose up -d
```
**优点:**
- ✅ 完全控制，不依赖外部服务
- ✅ 可以定制化修改
- ✅ 无网络限制

### 3. 阿里云镜像（企业用户）
**优点:**
- ✅ 最快的速度
- ✅ 最高的稳定性
- ✅ 可配置私有仓库
- ❌ 需要阿里云账号和配置

### 其他建议
- **配置 CDN 加速** - 如果需要公网访问
- **使用反向代理** - nginx/Caddy 配置 HTTPS
- **定期自动备份** - 使用 cron 每天自动备份数据库（见上方备份脚本）
- **异地备份** - 将备份文件同步到其他服务器或云存储
- **测试恢复** - 定期测试备份恢复流程确保可用

## 技术支持

- **项目地址:** https://github.com/shtdu/code-together
- **问题反馈:** https://github.com/shtdu/code-together/issues
- **主文档:** 见 `DOCKER.md`（英文）和 `README.md`

## 相关文档

- [DOCKER.md](./DOCKER.md) - 完整 Docker 部署指南（英文）
- [README.md](./README.md) - 项目概览
- [server/CLAUDE.md](./server/CLAUDE.md) - 服务器开发文档
- [manager/CLAUDE.md](./manager/CLAUDE.md) - 管理界面开发文档
