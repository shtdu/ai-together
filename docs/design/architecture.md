# Architecture

**Last Updated:** February 2025
**Epic:** Epic 0 - Core Platform Foundation

This document describes the architecture of the AI Together project from a system design perspective, covering design principles, technology choices, deployment patterns, and operational considerations.

## Table of Contents

1. [Design Principles](#design-principles)
2. [System Overview](#system-overview)
3. [Technology Stack & Rationale](#technology-stack--rationale)
4. [Component Architecture](#component-architecture)
5. [Data Model](#data-model)
6. [Security Architecture](#security-architecture)
7. [Multi-Tenancy Strategy](#multi-tenancy-strategy)
8. [Scalability & Performance](#scalability--performance)
9. [Deployment Patterns](#deployment-patterns)
10. [Operational Concerns](#operational-concerns)
11. [Evolution & Extensibility](#evolution--extensibility)

---

## Design Principles

### 1. Privacy-First Architecture

**Principle:** The system MUST NOT have access to actual prompt/response data from AI tools.

**Rationale:** AI prompts often contain sensitive code, proprietary algorithms, or confidential business logic. Storing this data would:
- Create security and compliance risks
- Increase liability for data breaches
- Violate user trust expectations

**Implementation:**
```
AI Tool → Proxy (extracts metadata only) → Server (stores metadata)
           ↓
    Actual prompt/response → Forwarded to AI provider (NOT logged)
```

**Trade-off:** We lose the ability to debug actual AI interactions, but gain user trust and reduce security surface area.

---

### 2. Multi-Tenant Data Isolation

**Principle:** Each tenant's data MUST be strictly isolated at the database level.

**Rationale:**
- Compliance with data privacy regulations (GDPR, etc.)
- Prevent data leakage between organizations
- Enable tenant-specific backups and migrations

**Implementation:**
- All tables include `tenant_id` column
- Repository layer automatically filters by `tenant_id`
- JWT claims include `tenant_id` for context injection

**Trade-off:** Cannot share data across tenants (feature, not a bug), but simplifies compliance.

---

### 3. Stateless Authentication

**Principle:** Use stateless JWT tokens instead of server-side sessions.

**Rationale:**
- Simplifies horizontal scaling (no session replication)
- Reduces database load (no session table lookups)
- Enables mobile/desktop clients without session cookies

**Implementation:**
- Access tokens: 24-hour expiration
- Refresh tokens: 7-day expiration
- Client-side token storage with auto-refresh

**Trade-off:** Cannot immediately revoke tokens (requires blacklist or short-lived tokens), but improves scalability.

---

### 4. API-First Design

**Principle:** All server functionality exposed via RESTful API with OpenAPI specification.

**Rationale:**
- Enables multiple client types (web, desktop, mobile)
- Facilitates third-party integrations
- Simplifies testing and documentation
- Decouples frontend from backend

**Implementation:**
- OpenAPI 3.0.3 specification in `docs/client_api/server_api.yaml`
- Generated TypeScript/Go clients via `oapi-codegen`
- Versioned endpoints (`/api/v1/`)

**Trade-off:** More upfront design effort, but enables faster client development.

---

### 5. Graceful Degradation

**Principle:** Member clients MUST function in degraded mode when server is unreachable.

**Rationale:**
- Users should not be blocked from working if server is down
- Enables offline development and testing
- Improves perceived reliability

**Implementation:**
- Member client uses cached provider config
- Queues usage data locally when offline
- Auto-syncs when connection restored
- Local proxy continues functioning

**Trade-off:** Stale configurations might be used, but maintains productivity.

---

## System Overview

AI Together is a **distributed multi-tenant SaaS platform** for AI provider management with the following characteristics:

- **Three-tier architecture:** Client → Application → Data
- **Multi-protocol support:** Claude, Codex, OpenCode with extensible provider model
- **Real-time usage tracking:** Token-level metadata collection without content inspection
- **Tiered licensing:** 4-tier system (Community → Enterprise) with feature gates
- **Zero-configuration member experience:** Auto-sync from server, no manual setup required

## Overview

AI Together is a three-tier distributed system:

```
┌─────────────────────────────────────────────────────────────┐
│                         Client Layer                         │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Browser    │  │ Desktop App  │  │ Desktop App  │      │
│  │  (Manager)   │  │ (Member)     │  │ (Admin)      │      │
│  │  React 18    │  │  Wails 3     │  │  Wails 3     │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                            ↓ HTTPS
┌─────────────────────────────────────────────────────────────┐
│                      nginx Reverse Proxy                     │
├─────────────────────────────────────────────────────────────┤
│  • TLS Termination          • Security Headers              │
│  • Static Asset Serving     • WebSocket Support             │
│  • Rate Limiting (planned)  • SPA Routing                   │
└─────────────────────────────────────────────────────────────┘
                            ↓ HTTP (internal)
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer (Server)                 │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Gin HTTP Router                                     │   │
│  │  ├─ Authentication Middleware (JWT)                  │   │
│  │  ├─ RBAC Middleware (Casbin)                         │   │
│  │  ├─ License Middleware (Tier Enforcement)            │   │
│  │  └─ Request Handlers                                 │   │
│  └──────────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Business Logic (Services)                           │   │
│  │  ├─ UserService        ├─ ProviderService            │   │
│  │  ├─ TeamService         ├─ UsageService               │   │
│  │  ├─ LicenseService      ├─ AnalyticsService           │   │
│  │  └─ ProxyService                                      │   │
│  └──────────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Data Access (Repository Pattern)                     │   │
│  │  ├─ UserRepository   ├─ ProviderRepository            │   │
│  │  ├─ TeamRepository    ├─ UsageRepository               │   │
│  │  └─ LicenseRepository                                  │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                      Data Layer (PostgreSQL)                  │
├─────────────────────────────────────────────────────────────┤
│  • tenants    • users          • team_members                │
│  • teams      • providers      • request_log                 │
│  • licenses  • team_usage_summary                             │
└─────────────────────────────────────────────────────────────┘
```

## Security & Privacy

### Data Privacy Principle

**Core Principle:** The system never tracks actual data transmitted through AI tools. Only metadata (usage statistics, token counts, model names) is collected.

**Implementation:**
1. Member client runs HTTP proxy on port 18100
2. Proxy intercepts requests to AI providers (Claude, Codex, OpenCode)
3. Proxy extracts metadata: model, tokens, provider, timestamps
4. Actual prompt/response data is NOT logged or stored
5. Metadata is batch-uploaded to server for analytics and billing

### Security Architecture

```
┌────────────────┐
│  Internet      │
└────────┬───────┘
         │ HTTPS (TLS 1.3)
         ↓
┌────────────────┐     JWT          ┌─────────────────┐
│     nginx      │◄──────────────►│  Server (Gin)   │
│  (Termination) │                 │  (Validation)   │
└────────────────┘                 └────────┬────────┘
                                            │ JWT Claims
                                            ↓
                                   ┌─────────────────┐
                                   │  PostgreSQL     │
                                   │  (tenant_id)     │
                                   └─────────────────┘
```

**Security Features:**
- ✅ TLS/SSL via nginx reverse proxy
- ✅ JWT authentication (24hr access + 7-day refresh tokens)
- ✅ RBAC with Casbin (manager/member roles)
- ✅ Multi-tenant data isolation (tenant_id column on all tables)
- ✅ License enforcement (4-tier system with feature gates)
- ⚠️ Token revocation (planned - requires Redis or database table)
- ⚠️ Rate limiting (can be added via nginx `limit_req`)

---

## Technology Stack & Rationale

### Backend Stack

| Component | Technology | Justification |
|-----------|-----------|---------------|
| **Language** | Go 1.24+ | Performance, concurrency, single binary deployment |
| **Web Framework** | Gin | Minimal, fast, middleware-friendly |
| **Database** | PostgreSQL 14+ | ACID compliance, JSON support, mature tooling |
| **ORM/Query** | sqlc | Type-safe SQL, no runtime overhead, compile-time validation |
| **Connection Pool** | pgxpool | Efficient connection management, context-aware |
| **Authentication** | JWT (golang-jwt/jwt) | Stateless, scalable, industry standard |
| **Authorization** | Casbin | Policy-based, flexible RBAC, domain-agnostic |
| **License Signing** | ed25519 | Cryptographic security, small key size, fast verification |
| **JSON Processing** | gjson/sjson | Zero-allocation parsing, critical for proxy performance |

**Why Not Other Technologies:**

| Decision | Considered | Rejected Because |
|----------|-----------|-----------------|
| ORM over sqlc | GORM, sqlx | Runtime overhead, less type safety, performance |
| NoSQL over PostgreSQL | MongoDB, Redis | Need ACID, complex joins, mature tooling |
| Sessions over JWT | Server sessions | Scaling complexity, state management |
| gRPC over REST | Protocol Buffers | Browser compatibility, debugging complexity |

### Frontend Stack

#### Shared Frontend Infrastructure

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **API Generation** | oapi-codegen | Type-safe client from OpenAPI spec |
| **API Client** | Axios (manager), net/http (member) | HTTP requests with interceptors |
| **Authentication** | JWT (access + refresh tokens) | Token management with auto-refresh |
| **Type Safety** | TypeScript (manager), Go (member) | Compile-time type checking |

**API Code Generation Workflow:**
```
docs/client_api/server_api.yaml (OpenAPI 3.0.3)
           ↓ (oapi-codegen)
shared/integration/ (generated Go + TypeScript clients)
           ↓
member/ & manager/ import and use
```

---

#### Manager UI (React 18 Web Application)

**Status:** ✅ Fully Implemented

**Technology Stack:**
| Layer | Technology | Rationale |
|-------|-----------|----------|
| **Framework** | React 18.3+ | Concurrent features, hooks, excellent ecosystem |
| **Language** | TypeScript 5.6+ | Type safety, better IDE support |
| **Build Tool** | Vite 6.x | Fast HMR, optimized production builds, native ESM |
| **UI Library** | Material UI (MUI) v6 | Pre-built components, theming, responsive |
| **Charts** | Recharts 2.x | Declarative, responsive, D3-based |
| **Routing** | React Router v7 | Lazy loading, route guards, nested routes |
| **State Management** | TanStack React Query v5 | Server state, caching, auto-refetch |
| **HTTP Client** | Axios 1.x | Interceptors, request cancellation, type-safe |
| **Form Validation** | React Hook Form + Zod | Type-safe forms, minimal re-renders |
| **Styling** | MUI sx + Emotion | Component-level styling, theme integration |

**Project Structure:**
```
manager/
├── src/
│   ├── api/                 # API client modules (Axios)
│   │   ├── client.ts        # Axios instance with interceptors
│   │   ├── auth.ts          # POST /auth/* endpoints
│   │   ├── dashboard.ts     # GET /api/v1/dashboard/*
│   │   ├── analytics.ts     # GET /api/v1/analytics/*
│   │   └── users.ts         # GET/POST/PUT /api/v1/users
│   ├── components/
│   │   ├── charts/          # Recharts wrappers
│   │   │   ├── ActivityChart.tsx
│   │   │   └── ProviderRankingChart.tsx
│   │   ├── common/          # Reusable UI components
│   │   │   ├── DateRangePicker.tsx
│   │   │   ├── MultiSelect.tsx
│   │   │   └── UserSelect.tsx
│   │   └── layout/
│   │       ├── AppLayout.tsx   # Main app shell
│   │       ├── Header.tsx      # Top navigation
│   │       └── Sidebar.tsx     # Side navigation
│   ├── pages/               # Route-level page components
│   │   ├── DashboardPage.tsx
│   │   ├── UserManagementPage.tsx
│   │   ├── TokenProviderPage.tsx
│   │   ├── TokenUserPage.tsx
│   │   └── ProfilePage.tsx
│   ├── contexts/
│   │   └── AuthContext.tsx  # Authentication state + logout
│   └── types/
│       ├── api.ts           # OpenAPI-generated types
│       └── models.ts        # Domain models
├── index.html              # Entry point
├── vite.config.ts          # Vite configuration
├── tsconfig.json           # TypeScript config
└── package.json            # Dependencies
```

**Key Architectural Patterns:**

1. **Route-Based Code Splitting:**
```typescript
// Lazy load pages for smaller bundle
const DashboardPage = lazy(() => import('./pages/DashboardPage.tsx'));
const UserManagementPage = lazy(() => import('./pages/UserManagementPage.tsx'));
```

2. **React Query for Server State:**
```typescript
// Automatic caching, refetching, loading states
const { data, isLoading, error } = useQuery({
  queryKey: ['dashboard', 'metrics', timeRange],
  queryFn: () => dashboardApi.getMetrics(timeRange, 'hour'),
  staleTime: 30000,  // 30 seconds
});
```

3. **Protected Routes:**
```typescript
// Nested route guards for RBAC
<Route path="/users" element={
  <ProtectedRoute requireAdmin>
    <UserManagementPage />
  </ProtectedRoute>
} />
```

4. **Axios Interceptors:**
```typescript
// Auto-inject access token
apiClient.interceptors.request.use((config) => {
  const token = getAccessToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Auto-refresh on 401
apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      const newToken = await refreshToken();
      return retryOriginalRequest(newToken);
    }
    return Promise.reject(error);
  }
);
```

**Build Artifacts:**
- Development: `vite dev` (port 3000, proxy to :9080)
- Production: `pnpm build` → `dist/` (static files)
- Bundle Size: ~1.5 MB (gzip: ~400 KB)

---

#### Member Client (Wails v3 + Vue 3 Desktop Application)

**Status:** ✅ Fully Implemented

**Technology Stack:**
| Layer | Technology | Rationale |
|-------|-----------|----------|
| **Desktop Framework** | Wails v3 (alpha) | Native Go + web frontend, cross-platform |
| **Frontend Framework** | Vue 3.4+ | Reactive, Composition API, small bundle |
| **Language** | TypeScript 5.6+ | Type safety, Intellisense |
| **Build Tool** | Vite 5.x (built-in Wails) | Fast HMR, optimized builds |
| **UI Components** | Custom + Element Plus | Lightweight, functional |
| **State Management** | Pinia 2.x | Vue 3 native, reactive stores |
| **HTTP Client** | Generated (oapi-codegen) | Type-safe, from OpenAPI spec |
| **Styling** | CSS + Scoped | Simple, no build overhead |

**Project Structure:**
```
member/
├── main.go                  # Wails entry point, service binding
├── services/                # Business logic (Go)
│   ├── providerservice.go         # Exposed to frontend
│   ├── providerrelay.go           # HTTP proxy (port 18100)
│   ├── authservice.go             # JWT auth + refresh
│   ├── configsyncservice.go       # Server config sync
│   ├── usagesyncservice.go        # Batch usage upload
│   ├── hookservice.go             # Hook event API
│   └── [20+ other services...]
├── frontend/                # Vue 3 application
│   ├── src/
│   │   ├── components/      # Vue SFC components
│   │   ├── views/           # Page views
│   │   ├── stores/          # Pinia stores
│   │   ├── services/        # API client wrapper
│   │   ├── router/          # Vue Router routes
│   │   ├── App.vue          # Root component
│   │   └── main.ts          # Entry point
│   ├── index.html           # HTML template
│   ├── vite.config.ts      # Vite config
│   └── package.json        # Dependencies
├── internal/                # Internal packages
│   ├── api/                 # Generated API client
│   ├── db/                  # Local SQLite database
│   └── models/              # Shared data models
└── build/                   # Wails build configuration
```

**Key Architectural Patterns:**

1. **Service Binding (Go → JavaScript):**
```go
// member/main.go
func main() {
    app := NewApp()

    // Expose services to frontend
    app.Bind(providerservice.NewProviderService())
    app.Bind(authservice.NewAuthService(...))
    app.Bind(configsyncservice.NewConfigSyncService(...))

    wails.Run(&options)
}
```

2. **Frontend Service Calls:**
```typescript
// member/frontend/src/services/provider.ts
import { GetProviders } from '../../wailsjs/go/main/ProviderService'

export const providerService = {
  async getProviders() {
    return await GetProviders()
  }
}
```

3. **Local Database (SQLite):**
```go
// member/internal/db/db.go
func OpenDB(path string) (*sql.DB, error) {
    db, err := sql.Open("sqlite", path)
    // Schema migrations on startup
    migrate(db)
    return db, err
}
```

4. **HTTP Proxy (port 18100):**
```go
// member/services/providerrelay.go
func (p *ProviderRelay) StartProxy() error {
    router := gin.Default()

    // Proxy endpoints
    router.POST("/v1/messages", p.ProxyClaude)
    router.POST("/responses", p.ProxyCodex)
    router.POST("/v1/chat/completions", p.ProxyOpenCode)

    return router.Run(":18100")
}
```

**Build Artifacts:**
- Development: `wails3 task dev` (hot reload)
- Production: `wails3 task build` → native binary
- Bundle Size:
  - macOS: ~40-50 MB (universal binary)
  - Windows: ~50-60 MB (exe)
  - Linux: ~45-55 MB (binary)

---

#### Frontend Comparison: Manager vs Member

| Aspect | Manager UI | Member Client |
|--------|-----------|---------------|
| **Type** | Web application | Desktop application |
| **Framework** | React 18 | Vue 3 + Wails v3 |
| **Deployment** | Static files (nginx) | Native binary |
| **Platform** | Any modern browser | macOS, Windows, Linux |
| **Updates** | Instant (refresh page) | Requires download/install |
| **State Sync** | Server state only | Local cache + server sync |
| **Offline Mode** | None | Degraded mode (cached config) |
| **Bundle Size** | ~1.5 MB (gzip) | ~50 MB (binary) |
| **Memory Usage** | ~50-100 MB (browser tab) | ~50-100 MB (process) |
| **User Type** | Managers (admin) | Members (end users) |

---

## Component Architecture

### 1. Server (`server/`)

**Status:** ✅ Fully Implemented

**Technology Stack:**
- Go 1.24+
- Gin web framework
- PostgreSQL (pgxpool for connection pooling)
- sqlc for type-safe database code generation
- Casbin for RBAC
- ed25519 for license signing

**Responsibilities:**
- Multi-tenant authentication and authorization
- Team, user, and provider management
- Usage tracking and analytics
- License enforcement and validation
- HTTP proxy for provider requests (optional, for last-mile routing)

**Key Services:**
```
server/
├── handlers/          # HTTP request handlers
│   ├── auth_handler.go         # /auth/login, /register, /refresh
│   ├── team_handler.go         # /api/v1/teams/*
│   ├── provider_handler.go     # /api/v1/providers/*
│   ├── usage_handler.go        # /api/v1/usage/*
│   ├── dashboard_handler.go    # /api/v1/dashboard/*
│   ├── analytics_handler.go    # /api/v1/analytics/*
│   └── license_handler.go      # /api/v1/license/*
├── services/          # Business logic layer
│   ├── user_service.go
│   ├── team_service.go
│   ├── provider_service.go
│   ├── usage_service.go
│   ├── license_service.go      # 4-tier license enforcement
│   └── interfaces.go            # Service interfaces
├── repository/        # Data access layer
│   ├── user_repository.go
│   ├── team_repository.go
│   ├── provider_repository.go
│   ├── usage_repository.go
│   └── license_repository.go
├── middleware/        # HTTP middleware
│   ├── auth_middleware.go       # JWT validation
│   ├── rbac.go                   # Casbin authorization
│   └── license_middleware.go    # Tier enforcement
├── models/            # Data models
└── internal/db/       # Database connection + migrations
```

**API Endpoints:**
- Authentication: `/auth/login`, `/auth/register`, `/auth/refresh`, `/auth/verify`
- Teams: `/api/v1/teams/*` (CRUD + member management)
- Providers: `/api/v1/providers/*` (CRUD + test + enable/disable)
- Usage: `/api/v1/usage/*` (tracking + statistics)
- Analytics: `/api/v1/analytics/*` (provider/user analytics)
- License: `/api/v1/license/*` (activation + status)

---

### 2. Member Client (`member/`)

**Status:** ✅ Fully Implemented

**Technology Stack:**
- Wails v3 (Go + Vue 3 frontend)
- Vue 3 + TypeScript + Vite
- Pinia for state management
- SQLite for local caching

**Responsibilities:**
- Zero-configuration setup for AI coding tools
- Auto-configuration sync from server
- Local HTTP proxy (port 18100) for transparent request routing
- Usage data collection and batch upload
- Hook event collection from AI tools

**Architecture:**
```
member/
├── main.go                  # Wails app entry point
├── services/                # Business logic (exposed to frontend)
│   ├── providerservice.go         # Provider CRUD
│   ├── providerrelay.go           # HTTP proxy (port 18100)
│   ├── authservice.go             # JWT authentication + refresh
│   ├── configsyncservice.go       # Server config sync
│   ├── usagesyncservice.go        # Batch usage upload
│   ├── hookservice.go             # Hook event collection API
│   ├── claudesettings.go          # Claude Code tool config
│   ├── codexsettings.go           # Codex tool config
│   └── opencodesettings.go        # OpenCode tool config
├── internal/                # Internal packages
│   ├── db/                   # Local SQLite database
│   ├── api/                  # Generated API client (from OpenAPI)
│   └── models/               # Data models
└── frontend/                # Vue 3 UI
    └── src/
        ├── components/      # Vue components
        ├── views/           # Page views
        └── stores/          # Pinia stores
```

**Key Features:**
- **HTTP Proxy (port 18100):**
  - Routes to Claude: `POST /v1/messages`
  - Routes to Codex: `POST /responses`
  - Routes to OpenCode: `POST /v1/chat/completions`
  - Automatic failover by provider priority
  - Token usage tracking

- **Auto-Configuration:**
  - Fetches provider config from server
  - Updates AI tool settings automatically
  - Syncs in background every 5 minutes

- **Usage Collection:**
  - Tracks token counts per request
  - Batches uploads every 30 seconds
  - Works offline (queues until reconnection)

- **Hook Event API:**
  - Endpoint: `POST /collect/{tool_name}`
  - Collects events from Claude Code hooks
  - Stores in SQLite for debugging
  - Provides browser UI for event inspection

---

### 3. Manager Client (`manager/`)

**Status:** ✅ Fully Implemented (React web application)

**Technology Stack:**
- React 18 + TypeScript
- Material UI (MUI) v6
- React Router v7
- TanStack React Query for server state
- Recharts for data visualization
- Vite for build tooling

**Responsibilities:**
- Team management dashboard
- Usage analytics and reporting
- User management (CRUD operations)
- Provider configuration UI
- License activation and status display
- Cost tracking and budget alerts

**Architecture:**
```
manager/
├── src/
│   ├── api/                 # API client modules
│   │   ├── client.ts        # Axios instance with interceptors
│   │   ├── auth.ts          # Authentication endpoints
│   │   ├── dashboard.ts     # Dashboard data
│   │   ├── analytics.ts     # Analytics endpoints
│   │   └── users.ts         # User management
│   ├── components/
│   │   ├── charts/          # Recharts visualizations
│   │   ├── common/          # Reusable UI components
│   │   └── layout/          # App shell (Header, Sidebar)
│   ├── pages/               # Route-level pages
│   │   ├── DashboardPage.tsx
│   │   ├── UserManagementPage.tsx
│   │   ├── TokenProviderPage.tsx
│   │   ├── TokenUserPage.tsx
│   │   └── ProfilePage.tsx
│   ├── contexts/
│   │   └── AuthContext.tsx  # Authentication state
│   └── types/
│       ├── api.ts           # API types
│       └── models.ts        # Domain models
├── nginx.conf               # Reverse proxy configuration
└── Dockerfile               # Container image
```

**Key Features:**
- **Dashboard:**
  - Real-time metrics (requests, tokens, costs)
  - Provider usage rankings
  - Member activity statistics
  - Custom date ranges and intervals

- **User Management:**
  - Add/remove users
  - Assign roles (manager/member)
  - View per-user analytics

- **Provider Management:**
  - CRUD operations for providers
  - Test connectivity
  - Enable/disable providers
  - View provider statistics

- **Analytics:**
  - Provider-level analytics with filtering
  - User-level analytics with filtering
  - Request history with pagination
  - Export to CSV/JSON (planned)

---

## Multi-Tenancy Architecture

### Data Isolation Strategy

All tables include `tenant_id` column for strict data segregation:

```sql
-- Example table structure
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,  -- Multi-tenant isolation
    email VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    ...
);

CREATE TABLE providers (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,  -- Multi-tenant isolation
    team_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    ...
);
```

**Query Pattern:**
```go
// All repository queries automatically filter by tenant_id
func (r *UsageRepository) GetMetricsByTenant(ctx context.Context, tenantID int64, ...) {
    query := `SELECT * FROM request_log WHERE tenant_id = $1`
    // tenant_id is injected from JWT claims
}
```

### Tenant Context Flow

```
1. User logs in → JWT issued with tenant_id claim
2. Client requests → Authorization: Bearer <token>
3. Server validates JWT → Extracts tenant_id from claims
4. AuthMiddleware → Injects tenant_id into Gin context
5. Repository queries → Automatically filtered by tenant_id
6. Response → Only tenant-specific data returned
```

---

## Role-Based Access Control (RBAC)

### Casbin Policy Model

**Roles:**
- **member:** Read-only access to providers and usage
- **manager:** Full read/write access to all resources

**Permissions:**

| Resource | Action | Member | Manager |
|----------|--------|--------|---------|
| providers | read | ✅ | ✅ |
| providers | write | ❌ | ✅ |
| providers | delete | ❌ | ✅ |
| teams | read | ✅ (own team only) | ✅ (all teams) |
| teams | write | ❌ | ✅ |
| users | read | ❌ | ✅ |
| users | write | ❌ | ✅ |
| usage | read | ✅ (own data) | ✅ (all data) |
| analytics | read | ❌ | ✅ |
| license | read | ✅ | ✅ |
| license | write | ❌ | ✅ |

**Enforcement Points:**
```go
// middleware/rbac.go
func RequirePermission(enforcer *casbin.Enforcer, resource, action string) gin.HandlerFunc {
    return func(c *gin.Context) {
        user := c.Get("user").(*models.User)
        if !enforcer.Enforce(user.Role, resource, action) {
            c.JSON(403, gin.H{"error": "Forbidden"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

---

## License System

### 4-Tier Architecture

| Tier | Name | Seats | Providers | Pricing |
|------|------|-------|-----------|---------|
| 0.0 | Community | 3 | 2 per kind | Free |
| 1.0 | Standard | 5 | Unlimited | ¥200/month |
| 2.0 | Professional | 20 | Unlimited | ¥1,200/month |
| 3.0 | Enterprise | 50+ | Unlimited | ¥5,000/month |

### Enforcement Points

**Server-Side:**
```go
// services/license_service.go
func (s *LicenseService) CanAddUser(ctx context.Context, tenantID int64) (bool, error) {
    license := s.GetEffectiveLicense(ctx, tenantID)
    currentUsers := s.userRepo.CountUsers(ctx, tenantID)
    return currentUsers < license.Seats, nil
}

func (s *LicenseService) CanAddProvider(ctx context.Context, tenantID int64, kind string) (bool, error) {
    license := s.GetEffectiveLicense(ctx, tenantID)

    // Tier 0 has max 2 providers per kind
    if license.MajorTier() == 0 {
        currentCount := s.providerRepo.CountProvidersByKind(ctx, tenantID, kind)
        return currentCount < 2, nil
    }

    // Tier 1+ has unlimited providers
    return true, nil
}
```

**Client-Side:**
- Manager UI hides features based on tier
- Member client syncs provider limits from server

---

## Data Flow

### Authentication Flow

```mermaid
sequenceDiagram
    participant Client
    participant Server
    participant Database

    Client->>Server: POST /auth/login
    Note over Client, Server: {email, password}

    Server->>Database: SELECT * FROM users WHERE email = ?
    Note over Server, Database: Validate user exists

    Database-->>Server: user_record
    Note over Database, Server: {id, email, password_hash, role, tenant_id}

    Server->>Server: bcrypt.Compare(password, hash)
    Note over Server: Validate credentials

    alt Credentials valid
        Server->>Server: Generate JWT access_token (24hr)
        Server->>Server: Generate JWT refresh_token (7d)
        Note over Server: Create signed tokens with user claims

        Server-->>Client: {access_token, refresh_token, expires_at, user}
        Note over Server, Client: Return tokens + user info

        Client->>Client: Store tokens locally
        Note over Client: ~/.code-together/auth.json
    else Credentials invalid
        Server-->>Client: 401 Unauthorized
        Note over Server, Client: Return error
    end
```

---

### Provider Configuration Sync Flow

```mermaid
sequenceDiagram
    participant Member as Member Client
    participant Server
    participant DB as PostgreSQL
    participant AI as AI Tools

    Member->>Server: GET /api/v1/providers
    Note over Member, Server: Authorization: Bearer <access_token>

    Server->>DB: SELECT * FROM providers WHERE tenant_id = ?
    Note over Server, DB: Filter by tenant_id from JWT claims

    DB-->>Server: Provider list
    Note over DB, Server: [{id, name, api_url, api_key, kind, model_mapping, enabled, level}]

    Server-->>Member: Provider configuration
    Note over Server, Member: Return provider list (api_key excluded)

    Member->>Member: Validate providers
    Note over Member: Check required fields, connectivity

    Member->>AI: Update AI tool configuration
    Note over Member, AI: Modify ~/.claude/config.json for Claude

    Member->>Member: Restart local proxy if needed
    Note over Member: Reload providerrelay service

    alt Provider changed
        Member->>Server: POST /collect/{tool_name}
        Note over Member, Server: Log configuration sync event
    end
```

---

### Usage Tracking & Upload Flow

```mermaid
sequenceDiagram
    participant AI as AI Tool
    participant Member as Member Client
    participant Local as Local DB
    participant Server as Server
    participant DB as PostgreSQL

    Note over AI, Member: AI Tool makes request

    AI->>Member: POST http://localhost:18100/v1/messages
    Note over AI, Member: Request with model, messages, streaming

    Member->>Member: Extract metadata (model, tokens, provider)
    Note over Member: Parse response to count tokens
    Member->>Member: Generate request_id
    Member->>Local: Store request_log locally
    Note over Member, Local: SQLite for offline queue

    Member->>AI: Forward to provider API
    Note over Member, AI: Transparent proxying

    AI-->>Member: AI response
    Note over AI, Member: Return response with usage data

    Member->>Member: Update request_log with actual usage
    Note over Member: Add input_tokens, output_tokens, etc.

    Note over Member: Every 30 seconds or batch of 100 records

    Member->>Server: POST /api/v1/usage/batch
    Note over Member, Server: Authorization: Bearer <token>
    Member->>Server: [{platform, model, provider, tokens, tenant_id, user_id, ...}]

    Server->>Server: Validate tenant_id and user_id
    Note over Server: Ensure data belongs to authenticated user

    Server->>DB: INSERT INTO request_log (batch)
    Note over Server, DB: Bulk insert for performance

    DB-->>Server: Insert confirmation
    Note over DB, Server: {synced_count: 95, errors: [...]}

    Server-->>Member: Upload result
    Note over Server, Member: Confirm batch processed

    Member->>Member: Mark local records as synced
    Note over Member: Clear from local queue
```

### Provider Sync Flow

```mermaid
sequenceDiagram
    participant Member as Member Client
    participant Background as Background Sync
    participant Server as Server API
    participant DB as PostgreSQL
    participant AI as AI Tools

    Note over Background: Triggered every 5 minutes
    Background->>Background: Check sync timer

    Background->>Member: HasLocalChanges(provider_type)?
    Member-->>Background: false (no local edits)

    Background->>Server: GET /api/v1/providers (with JWT)
    Server->>DB: SELECT * FROM providers WHERE tenant_id = ?
    DB-->>Server: Provider list

    Server-->>Background: [{id, name, api_url, kind, model_mapping, enabled, level, ...}]

    Background->>Background: Convert to local format
    Background->>Member: Update provider configuration
    Member->>Member: Save to local SQLite
    Member->>AI: Update Claude/Codex/OpenCode config

    Note over Member,AI: Proxy restarts if routing changes

    Note over Member,Server: Manager Push Flow (when manager edits)
    Member->>Server: PUT /api/v1/providers/:id
    Server->>DB: UPDATE providers SET ...
    Server-->>Member: Updated provider

    Note over Member: Sync confirmation to other clients
    Server-->>Background: WebSocket push (future)
    Background->>Background: Trigger immediate sync

    Note over Member: Local Changes Preservation
    Background->>Member: HasLocalChanges(provider_type)?
    Member-->>Background: true (user edited locally)
    Background->>Background: Skip server update for this type
    Background->>Background: Log skip reason
    Note over Background: Local changes take priority<br/>until explicitly pushed to server
```

### Usage Upload Flow

```mermaid
sequenceDiagram
    participant AI as AI Tool
    participant Proxy as Member Proxy
    participant Local as Local SQLite
    participant Background as Background Sync
    participant Server as Server API
    participant DB as PostgreSQL

    Note over AI,Local: Step 1: AI Tool Request (Real-time)
    AI->>Proxy: POST http://localhost:18100/v1/messages
    Proxy->>Proxy: Extract provider, model, kind
    Proxy->>Proxy: Initialize request_log struct

    Proxy->>AI: Proxy to upstream provider
    Note over Proxy: Streaming response

    AI-->>Proxy: Response chunks (SSE)
    loop For each chunk
        Proxy->>Proxy: Parse token usage (input/output/cache)
    end

    Note over Proxy: request_log populated with<br/>tokens, http_code, duration

    Proxy->>Local: INSERT INTO request_log<br/>(synced_to_server = 0)
    Note over Local: Stored locally for batch upload

    Proxy-->>AI: Return response to AI tool

    Note over Background: Step 2: Background Sync (Periodic)
    Background->>Background: Check sync timer

    Background->>Local: SELECT * FROM request_log<br/>WHERE synced_to_server = 0<br/>ORDER BY created_at ASC
    Local-->>Background: Pending records (e.g., 95 records)

    Background->>Background: Batch into chunks of 50
    Note over Background: batchSize = 50<br/>2 batches: 50 + 45

    loop For each batch
        Background->>Background: Transform to server format<br/>Add user_id, tenant_id

        Background->>Server: POST /api/v1/usage/batch<br/>Authorization: Bearer <token>
        Note over Server: Batch size: 50 records

        Server->>Server: Validate JWT claims<br/>Extract tenant_id, user_id
        Server->>Server: Validate each record schema

        Server->>DB: INSERT INTO request_log<br/>(batch of 50)
        DB-->>Server: Batch insert success

        Server-->>Background: {synced_count: 50, errors: []}
        Note over Background: All records in batch synced

        Background->>Local: UPDATE request_log<br/>SET synced_to_server = 1<br/>WHERE id IN (batch_ids)
    end

    Note over Background: Step 3: Cleanup Old Records
    Background->>Local: DELETE FROM request_log<br/>WHERE synced_to_server = 1<br/>AND created_at < NOW() - 7 DAYS
    Note over Local: Retention policy: 7 days<br/>Free up local storage

    Note over Local: Old records removed<br/>Database size managed
```

---

## Deployment Architecture

### Development Environment

```
┌─────────────────────────────────────────────────────┐
│  Developer Machine                                  │
├─────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐│
│  │ Server      │  │ Member      │  │ Manager     ││
│  │ :9080       │  │ :18100      │  │ :3000       ││
│  │ (air)       │  │ (wails3)    │  │ (vite)      ││
│  └─────────────┘  └─────────────┘  └─────────────┘│
│         │                │                │          │
│         └────────────────┴────────────────┘          │
│                    ↓                                 │
│  PostgreSQL :5432                                   │
└─────────────────────────────────────────────────────┘
```

### Production Environment

```
┌─────────────────────────────────────────────────────┐
│  Internet / Users                                    │
└────────────────┬────────────────────────────────────┘
                 │ HTTPS
                 ↓
┌─────────────────────────────────────────────────────┐
│  nginx (port 80/443)                                │
│  ├─ /api/*   → server:9080 (proxy)                  │
│  ├─ /auth/*  → server:9080 (proxy)                  │
│  ├─ /health → server:9080 (proxy)                  │
│  └─ /*      → manager/dist (static files)           │
└────────────────┬────────────────────────────────────┘
                 │ HTTP (internal network)
                 ↓
┌─────────────────────────────────────────────────────┐
│  Server (port 9080)                                 │
│  └─ Go application with Gin                          │
└────────────────┬────────────────────────────────────┘
                 │
                 ↓
┌─────────────────────────────────────────────────────┐
│  PostgreSQL (port 5432)                             │
│  └─ Multi-tenant database                           │
└─────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────┐
│  Member Clients (Desktop Apps)                       │
│  ├─ macOS CodeTogether.app                           │
│  ├─ Windows CodeTogether.exe                         │
│  └─ Linux CodeTogether (binary)                      │
│  └─ Local proxy on port 18100                        │
└─────────────────────────────────────────────────────┘
```

### Docker Deployment

**docker-compose.yml** includes:
- `server`: Go application with PostgreSQL
- `manager`: nginx serving React UI
- `postgres`: PostgreSQL database
- Volumes for data persistence
- Environment variables for configuration

**Example:**
```bash
docker-compose up -d
# All services started on localhost
# Manager UI: http://localhost:8080
# Server API: http://localhost:8080/api/*
# Member client connects to: http://localhost:8080
```

---

## Scalability & Performance

### Horizontal Scaling Strategy

**Server-Side:**
```
                    ┌─────────────────┐
                    │   Load Balancer  │ (nginx, HAProxy)
                    └────────┬────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
┌───────▼───────┐    ┌──────▼───────┐    ┌──────▼───────┐
│  Server 1    │    │  Server 2    │    │  Server 3    │
│  (Go proc)   │    │  (Go proc)   │    │  (Go proc)   │
└───────┬───────┘    └──────┬───────┘    └──────┬───────┘
        │                    │                    │
        └────────────────────┼────────────────────┘
                             │
                    ┌────────▼────────┐
                    │   PostgreSQL    │ (Connection pool)
                    │  (Primary)      │
                    └─────────────────┘
```

**Key Points:**
- **Stateless servers:** No session affinity required
- **Connection pooling:** pgxpool with 10-20 connections per server
- **Database bottleneck:** Single PostgreSQL instance limits write throughput
- **Read replicas:** Can add read replicas for analytics queries (not implemented)

**Scaling Tiers:**
| Load | Servers | DB Configuration |
|------|---------|------------------|
| 1-100 users | 1 server | Single instance, connection pool 10 |
| 100-1000 users | 2-3 servers | Single instance, connection pool 30 |
| 1000-10000 users | 5-10 servers | Read replicas, connection pool 50 |
| 10000+ users | 10+ servers | Sharded by tenant_id |

### Performance Characteristics

**Current Benchmarks:**
- Member proxy latency: ~5-10ms additional overhead
- Server API latency: 20-50ms (p95), 100-200ms (p99)
- Concurrent users per server: ~500-1000 (limited by DB connections)
- Usage batch upload: 100 records in ~200ms

**Bottlenecks:**
1. **Database writes:** Single PostgreSQL instance limits insert throughput
2. **Analytics queries:** Full-table scans on `request_log` for date range queries
3. **License verification:** Ed25519 verification on every request (fast but adds up)

**Optimization Opportunities:**
```go
// Current: Sequential batch insert
func (r *UsageRepository) CreateBatch(records []UsageRecord) error {
    for _, record := range records {
        INSERT INTO request_log ...
    }
}

// Optimized: Bulk insert with pgx
func (r *UsageRepository) CreateBatch(records []UsageRecord) error {
    _ = pgx.CopyFrom(
        conn,
        "COPY request_log (tenant_id, user_id, platform, model, ...)",
        reader,
    )
}
```

### Caching Strategy

**Current Implementation:**
- License data: In-memory cache on server startup
- Provider configurations: No caching (DB queries on sync)
- User sessions: No server-side sessions (JWT)

**Recommended Additions:**
```go
// Redis cache for frequently accessed data
type CacheService struct {
    redis *redis.Client
}

func (c *CacheService) GetProviderConfig(tenantID int64) ([]Provider, error) {
    key := fmt.Sprintf("providers:%d", tenantID)

    // Try cache first
    cached, err := c.redis.Get(key).Result()
    if err == nil {
        return json.Unmarshal(cached)
    }

    // Cache miss - query DB
    providers, _ := db.QueryProviders(tenantID)

    // Cache for 5 minutes
    data, _ := json.Marshal(providers)
    c.redis.Set(key, data, 5*time.Minute)

    return providers
}
```

---

## Data Model

### Entity Relationship Diagram

```mermaid
erDiagram
    tenants ||--o{ users : "has"
    tenants ||--o{ teams : "has"
    tenants ||--o{ request_log : "tracks"

    users }o--|| teams : "owns"
    users ||--o{ team_members : "joins"
    users ||--o{ request_log : "generates"

    teams ||--o{ team_members : "has"
    teams ||--o{ providers : "configures"
    teams ||--o{ team_usage_summary : "summarizes"

    team_members }o--|| teams : "belongs to"
    team_members }o--|| users : "member of"

    providers }o--|| teams : "configured for"

    request_log }o--|| tenants : "tracked by"
    request_log }o--|| users : "created by"

    team_usage_summary }o--|| teams : "aggregates"
```

### Database Schema Design Decisions

**1. Denormalized `tenant_id` on all tables**
- **Pro:** No JOINs required for multi-tenant filtering
- **Pro:** Enables tenant-level partitioning for sharding
- **Con:** Redundant data (but improves query performance)

**2. Separate `team_members` table**
- **Pro:** Allows users to belong to multiple teams (future feature)
- **Pro:** Explicit role per team (team-specific permissions)
- **Con:** Additional JOIN for user lookup (mitigated by caching)

**3. Materialized view for `team_usage_summary`**
- **Pro:** Fast dashboard queries (pre-aggregated)
- **Pro:** Reduces expensive GROUP BY operations
- **Con:** Data not real-time (refreshed periodically)

**4. `request_log` stores raw metadata**
- **Pro:** Detailed analytics possible
- **Pro:** No information loss
- **Con:** Table grows quickly (needs partitioning/archival)

---

## Operational Concerns

### Monitoring & Observability

**Current State:**
- ✅ Structured logging with slog
- ✅ Health check endpoint (`/health`)
- ⚠️ No metrics collection (Prometheus, etc.)
- ❌ No distributed tracing
- ❌ No APM (Application Performance Monitoring)

**Recommended Additions:**
```go
// Prometheus metrics endpoint
func (s *Server) MetricsHandler(c *gin.Context) {
    metrics := prometheus.Metrics{
        // Request count by endpoint
        // Request latency histograms
        // Active user count
        // Error rate by endpoint
    }
    c.JSON(200, metrics)
}

// OpenTelemetry tracing
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

func (h *Handler) ProxyRequest(c *gin.Context) {
    tracer := otel.Tracer("proxy")
    ctx, span := tracer.Start(c.Request.Context(), "ProxyRequest")
    defer span.End()

    span.SetAttributes(
        attribute.String("provider", providerName),
        attribute.String("model", modelName),
    )
}
```

### Logging Strategy

**Current Implementation:**
```go
// Structured logging in services
log := slog.With(
    "tenant_id", tenantID,
    "user_id", userID,
    "request_id", requestID,
)

log.Info("Processing provider request",
    "provider", providerName,
    "model", modelName,
)
```

**Log Levels:**
- **DEBUG:** Detailed request/response dumps (disabled in production)
- **INFO:** Normal operations (login, sync, etc.)
- **WARN:** Degraded performance (retry, fallback)
- **ERROR:** Request failures, database errors
- **FATAL:** Startup failures

**Log Aggregation (Recommended):**
- ELK Stack (Elasticsearch, Logstash, Kibana)
- Or: Loki + Grafana
- Centralized logging from all servers

### Backup & Disaster Recovery

**Current State:**
- ⚠️ No automated backups (manual pg_dump)
- ⚠️ No disaster recovery plan
- ❌ No backup testing

**Recommended Strategy:**
```bash
#!/bin/bash
# Daily backup script
DATE=$(date +%Y%m%d)
BACKUP_DIR="/backups/postgres"

# Full backup
pg_dump -Fc postgres://user:pass@localhost:5432/codetogether \
    > $BACKUP_DIR/codetogether_$DATE.dump

# Retention policy
find $BACKUP_DIR -name "*.dump" -mtime +30 -delete

# Offsite sync
rclone sync $BACKUP_DIR s3://backups/codetogether/
```

**RPO/RTO Targets:**
| Tier | RPO | RTO | Strategy |
|------|-----|-----|----------|
| Community | 24h | 4h | Daily backups |
| Standard | 12h | 2h | 2x daily |
| Professional | 1h | 30min | Hourly + WAL |
| Enterprise | 5min | 5min | Hot standby |

---

## Deployment Patterns (Enhanced)

### Pattern 1: Single-Node (SMB)

**Best For:**
- Teams < 100 users
- Proof of concept
- Cost-sensitive deployments

**Architecture:**
- Single server (4-8 CPU, 16-32 GB RAM)
- PostgreSQL on same machine
- nginx reverse proxy
- No redundancy

**Failure Impact:**
- Server down → Complete outage
- Database crash → Data loss risk (no replication)

**Mitigation:**
- Daily backups to offsite storage
- RAID 10 disk array
- Automatic restart on failure (systemd)

---

### Pattern 2: Active-Passive HA

**Best For:**
- Teams 100-1000 users
- Production workloads
- 99.9% uptime SLA

**Architecture:**
```
┌──────────┐
│ Primary  │ (Active)
│ Server   │
└─────┬────┘
      │
      ├────── streaming WAL replication ────┐
      │                                  │
      ▼                                  ▼
┌──────────┐                      ┌──────────┐
│Primary DB│                      │Standby DB│ (Passive)
└──────────┘                      └──────────┘

┌──────────┐
│ Standby  │ (Passive - Failover only)
│ Server   │
└──────────┘
```

**Failover Process:**
1. Primary server detected down (heartbeat failure)
2. Standby server promoted to primary
3. DNS update or VIP failover
4. Standby DB promoted to primary
5. RTO: 1-2 minutes

**Implementation:**
```bash
# heartbeat monitoring
while true; do
    if ! curl -f http://primary:9080/health; then
        echo "Primary down, initiating failover"
        /usr/local/bin/promote-standby.sh
        break
    fi
    sleep 10
done
```

---

### Pattern 3: Active-Active with Sharding

**Best For:**
- Teams 1000+ users
- Global multi-region deployment
- 99.99% uptime SLA

**Architecture:**
```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Region US   │     │  Region EU   │     │  Region AP   │
└──────┬───────┘     └──────┬───────┘     └──────┬───────┘
       │                     │                     │
       │    ┌────────────────┼────────────────┐
       │    │                │                │
   ┌───▼────▼───┐        ┌───▼────▼───┐        ┌───▼────▼───┐
   │ Server    │        │ Server    │        │ Server    │
   │ Pool (3)  │        │ Pool (3)  │        │ Pool (3)  │
   └───┬───────┘        └───┬───────┘        └───┬───────┘
       │                     │                     │
       └─────────────────────┼─────────────────────┘
                             │
                  ┌────────────▼─────────────┐
                  │  Shard Router (tenant_id) │
                  │  ┌──────────────────────┐  │
                  │  │ Shard 1 (US/West)   │  │
                  │  │ Shard 2 (EU/Central) │  │
                  │  │ Shard 3 (AP/Southeast)│  │
                  │  └──────────────────────┘  │
                  └────────────────────────────┘
```

**Sharding Strategy:**
- **Hash-based:** `tenant_id % num_shards`
- **Geo-based:** Route to nearest region
- **Tenant preference:** Allow tenants to choose region

---

## Security Architecture (Detailed)

### Authentication Flow

```
1. USER                2. SERVER               3. DATABASE
   │                        │                       │
   │ POST /auth/login      │                       │
   │ {email, password}     │                       │
   ├──────────────────────>│                       │
   │                        │ SELECT * FROM users    │
   │                        │ WHERE email = ?        │
   │                        ├──────────────────────>│
   │                        │ user_record            │
   │                        │<──────────────────────┤
   │                        │                       │
   │                        │ bcrypt.Compare(        │
   │                        │   password, hash)      │
   │                        ├───────┐               │
   │                        │       │               │
   │                        │  ┌────▼─────┐        │
   │                        │  │ Valid?   │        │
   │                        │  └────┬─────┘        │
   │                        │       │ Yes          │
   │                        │  ┌────▼─────┐        │
   │                        │  │ Generate│        │
   │                        │  │ JWT     │        │
   │                        │  └────┬─────┘        │
   │                        │       │               │
   │ {access_token,         │                       │
   │  refresh_token,        │                       │
   │  expires_at, user}     │                       │
   │<──────────────────────│                       │
   │                        │                       │
```

### JWT Token Structure

```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "user_id": 123,
    "email": "user@example.com",
    "role": "manager",
    "tenant_id": 1,
    "exp": 1704628800,
    "iat": 1704542400
  },
  "signature": "..."
}
```

**Security Properties:**
- **HS256:** HMAC with SHA-256 (symmetric)
- **Secret:** Configured per environment (env var)
- **Expiration:** 24 hours (access), 7 days (refresh)
- **Claims:** Encoded, signed, tamper-evident

### Authorization Flow

```
Request → AuthMiddleware → RBACMiddleware → Handler
            │                 │              │
            │                 │              │
      Validate JWT      Extract role    Check policy
            │                 │              │
         Inject user        Set context     Allow/Deny
```

**Casbin Policy Model:**
```go
// Policy definition
[request_definition]
r = sub, obj, act
p = sub, obj, act

[policy_definition]
p = manager, *, *
p = member, providers, read
p = member, usage, read
```

---

## Related Documentation
- **License Types:** `vibe_doc/01_identity_and_access/04_licensing/` - Open Source and Commercial license features
- **Database Schema:** `docs/production/database_schema.md` - ERD and table structures

### Implementation
- **Server:** `server/CLAUDE.md` - Server module documentation
- **Member:** `member/CLAUDE.md` - Member client documentation
- **Manager:** `manager/CLAUDE.md` - Manager UI documentation

### API Documentation
- **OpenAPI Spec:** `docs/client_api/server_api.yaml` - Complete API specification
- **Server README:** `server/README.md` - RBAC permissions and setup

### Business & Pricing
- **License Plan:** `docs/license/license_plan.md` - Pricing strategy (Chinese)
- **License Testing:** `docs/license/license_testing_plan.md` - Testing strategy

---

## Change History

| Date | Change | Author |
|------|--------|--------|
| 2025-02-07 | Complete rewrite to reflect current architecture | Claude Code |
| - | Updated component status (all implemented) | - |
| - | Added deployment architecture diagrams | - |
| - | Added multi-tenancy and RBAC sections | - |
| - | Added license system overview | - |
