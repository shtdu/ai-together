# Code Together Design Constitution

**Version:** 1.0
**Last Updated:** February 2025

---

## Purpose

This document captures the fundamental design principles that govern the Code Together project. All technical decisions SHOULD align with these principles unless explicitly justified.

---

## Core Principles

### 1. Privacy-First Architecture ⭐⭐⭐

**Principle:** Never access or store prompt/response data from AI tools.

Track only metadata: model names, token counts, timestamps, provider names. The member proxy transparently forwards requests without inspecting bodies.

**Why:** AI prompts contain sensitive code and proprietary algorithms. Storing this data creates security risks and violates user trust.

**Trade-off:** We lose debugging visibility into AI interactions, but gain user trust and reduce security surface.

---

### 2. Multi-Tenant Data Isolation ⭐⭐⭐

**Principle:** Strict data segregation at the database level.

All tables include `tenant_id`. Repository layer automatically filters queries. No cross-tenant data sharing.

**Why:** Compliance with GDPR/privacy regulations. Prevents data leakage between organizations.

**Trade-off:** Cannot share data across tenants (feature, not a bug).

---

### 3. Stateless Authentication ⭐⭐

**Principle:** Use JWT tokens, not server-side sessions.

- Access tokens: 24-hour expiration
- Refresh tokens: 7-day expiration
- Client-side storage with auto-refresh

**Why:** Simplifies horizontal scaling. Reduces database load. Enables desktop/mobile clients.

**Trade-off:** Cannot immediately revoke tokens (requires blacklist).

---

### 4. API-First Design ⭐

**Principle:** All functionality exposed via RESTful API with OpenAPI spec.

OpenAPI 3.0.3 specification in `docs/client_api/server_api.yaml`. Generate clients via `oapi-codegen`. Versioned endpoints (`/api/v1/`).

**Why:** Enables multiple client types. Facilitates third-party integrations. Better documentation.

---

### 5. Graceful Degradation ⭐

**Principle:** Member clients function offline when server is unreachable.

Use cached provider config. Queue usage data locally. Auto-sync when connection restored.

**Why:** Users should not be blocked if server is down. Improves perceived reliability.

**Trade-off:** Stale configurations might be used.

---

## Technology Stack

### Backend (Fixed)
| Component | Technology | Why |
|-----------|-----------|-----|
| Language | Go 1.24+ | Performance, concurrency, single binary |
| Framework | Gin | Minimal, fast, middleware-friendly |
| Database | PostgreSQL 14+ | ACID, JSON support, mature tooling |
| Query | sqlc | Type-safe SQL, no runtime overhead |
| Auth | JWT | Stateless, scalable |
| Authorization | Casbin | Policy-based RBAC |
| License Signing | ed25519 | Fast, small signatures |

**Rule:** Backend stack changes require architecture team approval.

### Frontend
| Application | Framework | Key Libraries |
|------------|-----------|---------------|
| Manager UI | React 18 + TypeScript | MUI v6, React Query, Vite |
| Member Client | Wails v3 + Vue 3 | Pinia, TypeScript |

**Rule:** Library updates allowed (major versions need review).

---

## Architectural Patterns

### Three-Tier Architecture (Server)

```
Handlers → Services → Repositories
   ↓           ↓            ↓
Parse    Business     Database
Request    Logic      Operations
```

**Rules:**
- Handlers: Parse requests, call services, return responses (NO business logic)
- Services: Orchestrate repositories, validation, auth (NO HTTP access)
- Repositories: SQL queries, data mapping (NO business logic)

### Repository Pattern

Use interfaces for dependency injection. Enables testing and swappable implementations.

### Middleware Chain

Order matters: Logging → CORS → Recovery → Auth → RBAC → License → Handler

### Service Registration (Member Client)

Register services at startup. Constructor injection for dependencies. No circular dependencies.

---

## Anti-Patterns (Things to Avoid)

### ❌ 1. Business Logic in Handlers
Handlers should only parse requests and return responses. Delegate logic to services.

### ❌ 2. Direct Database Access from Services
Services use repositories. Never access database directly.

### ❌ 3. Ignoring Errors
Always handle errors explicitly. Use `fmt.Errorf` with wrapping for context.

### ❌ 4. Hardcoding Configuration
Use environment variables for all configuration.

### ❌ 5. Logging Sensitive Data
Never log passwords, API keys, tokens, or prompt/response content.

### ❌ 6. N+1 Queries
Use JOINs or batch queries instead of querying inside loops.

### ❌ 7. God Objects
Separate services by domain. Single service should not do everything.

---

## Decision-Making Framework

When making architectural decisions:

1. **Does it violate a Core Principle?** (⭐⭐⭐ principles require explicit approval)
2. **Does it align with our Technology Stack?**
3. **Does it follow established Patterns?**
4. **What is the maintenance burden?** (testing, documentation, longevity)
5. **What are the security implications?** (tenant isolation, data privacy, injection risks)
6. **Can we degrade gracefully?** (offline mode, caching)

---

## Quick Reference

### ✅ DO:
- Track metadata only (no prompts/responses)
- Always filter by `tenant_id`
- Use stateless JWT authentication
- Define APIs in OpenAPI spec first
- Handle offline gracefully
- Follow three-tier architecture
- Use repository pattern with interfaces
- Handle all errors explicitly
- Use structured logging (slog)
- Use environment variables for config

### ❌ DON'T:
- Store or log prompt/response content
- Query without `tenant_id` filter
- Use server-side sessions
- Put business logic in handlers
- Access database directly from services
- Ignore errors silently
- Hardcode configuration
- Log sensitive data (passwords, tokens)
- Create N+1 query problems
- Build god objects

---

## Related Documents

- **Architecture:** `docs/design/architecture.md` - Detailed system design
- **License Types:** `vibe_doc/01_identity_and_access/04_licensing/` - Open Source and Commercial license features
- **API Spec:** `docs/client_api/server_api.yaml` - OpenAPI specification
- **Module Docs:** `server/CLAUDE.md`, `member/CLAUDE.md`, `manager/CLAUDE.md`

---

**Remember:** This constitution serves the team, not the other way around. Question it when it doesn't make sense, but respect it when it does.
