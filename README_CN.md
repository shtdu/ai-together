# AI Together

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/shtdu/ai-together)](https://goreportcard.com/report/github.com/shtdu/ai-together)

集中管理 AI 编码工具（Claude Code、Codex、OpenCode）的服务商代理平台，提供透明的请求路由、智能故障转移和团队协作功能。

## 主要特性

- 无需重启 AI 工具，平滑切换不同服务商
- 支持多服务商自动降级，保证使用体验
- 支持请求级别的用量统计，花费多少清晰可见
- 支持团队协作和集中配置管理
- 基于 [Wails 3](https://v3.wails.io) 构建的跨平台桌面应用

## 系统架构

AI Together 由四个主要组件组成：

- **Member Client** (`member/`) - 桌面客户端，运行本地 HTTP 代理服务
- **Server** (`server/`) - 后端 API 服务，提供集中管理和协作功能
- **Manager UI** (`manager/`) - 管理界面，用于团队分析和用户管理
- **Integration Tests** (`integration/`) - 独立的集成测试模块

### 工作原理

应用启动时在本地 18100 端口创建一个 HTTP 代理服务器，并自动更新 AI 工具配置指向该代理。代理内部只暴露兼容的关键端点：

- `/v1/messages` - 转发到配置的 Claude 服务商
- `/responses` - 转发到 Codex 服务商
- `/v1/chat/completions` - 转发到 OpenCode 服务商

请求由 proxyHandler 动态挑选符合当前优先级与启用状态的 provider，并在失败时自动回退。

## 下载与安装

[macOS](https://github.com/shtdu/ai-together/releases) | [Windows](https://github.com/shtdu/ai-together/releases)

## 使用手册

完整的使用手册请参考：[**成员客户端使用手册**](docs/manuals/member.md)

手册包含：
- 快速开始指南
- 核心功能说明
- 配置与设置
- 用户角色与权限
- 常见问题解答

## 开发环境

### 必需工具
- Go 1.24+
- Node.js 18+ 和 npm/pnpm
- Wails 3 CLI：`go install github.com/wailsapp/wails/v3/cmd/wails3@latest`

### 可选工具
- PostgreSQL 14+（仅 server 开发需要）
- Air（热重载）：`go install github.com/cosmtrek/air@latest`
- mingw-w64（Windows 交叉编译）：`brew install mingw-w64`

## 开发运行

### 快速开始

```bash
# 查看所有可用命令
make help

# 构建所有组件
make build

# 运行开发模式（server + member）
make dev

# 运行测试
make test

# 清理构建产物
make clean
```

### 模块开发

每个模块都有自己的 Makefile，提供更详细的命令：

```bash
# Server 模块
cd server && make help      # 查看所有 server 命令
make server-dev             # 从根目录运行 server 开发模式

# Member 客户端
cd member && make help      # 查看所有 member 命令
make member-dev             # 从根目录运行 member 开发模式

# Manager UI
cd manager && make help     # 查看所有 manager 命令
make manager-dev            # 从根目录运行 manager 开发模式

# Integration 测试
cd integration && make help # 查看所有 integration 命令
make integration-test       # 从根目录运行集成测试
```

## 构建流程

### 快速构建

```bash
# 构建所有组件（当前平台）
make build

# 构建发布版本（当前平台）
make release

# 构建所有平台的发布版本
make release-all
```

### 单独构建

```bash
# Server
make server-build
make server-linux        # 交叉编译 Linux

# Member 客户端
make member-build
make member-release      # 打包 .app / .exe

# Manager UI
make manager-build
```

### Windows 交叉编译（macOS）

```bash
# 安装交叉编译工具
brew install mingw-w64

# 构建 Windows 版本
make member-win
```

详细构建命令请参考各模块的 Makefile：
- `cd server && make help`
- `cd member && make help`
- `cd manager && make help`

## 常见问题

- **".app 无法打开"**: `cd member && make update-assets` 后再构建
- **macOS 交叉编译**: 需要终端拥有完全磁盘访问权限
- **Server 运行**: 需要配置 `server/.env` 文件中的 `DATABASE_URL`

更多问题请参考各模块的文档。

## 开发文档

- [CLAUDE.md](CLAUDE.md) - 项目概览和架构说明
- [member/CLAUDE.md](member/CLAUDE.md) - Member 客户端开发文档
- [server/CLAUDE.md](server/CLAUDE.md) - Server 开发文档
- [server/README.md](server/README.md) - Server RBAC 权限说明
- [manager/CLAUDE.md](manager/CLAUDE.md) - Manager UI 开发文档
- [integration/INTEGRATION_TEST_SETUP.md](integration/INTEGRATION_TEST_SETUP.md) - 集成测试设置指南
- [integration/integration.design.md](integration/integration.design.md) - 集成测试设计文档

## 贡献指南

我们欢迎贡献！请参阅 [CONTRIBUTING.md](CONTRIBUTING.md) 了解：
- 开发环境设置
- 代码风格指南
- Pull Request 流程
- 测试要求
- 提交信息规范

## 许可证

本项目采用 Apache License 2.0 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 行为准则

请阅读并遵守我们的[行为准则](CODE_OF_CONDUCT.md)。

## 安全性

关于安全策略和漏洞报告，请参阅 [SECURITY.md](SECURITY.md)。

## English Documentation

For English documentation, please see: [README.md](README.md)
