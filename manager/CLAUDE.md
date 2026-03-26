# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

The Manager client is the administrative web interface for the AI Together platform. It provides team-level dashboards, usage analytics, user management, and token usage tracking for AI coding tools (Claude Code, Codex, and OpenCode).

This is a **Vue 3 + TypeScript** single-page application (SPA) built with:
- **Vite** for build tooling and dev server
- **Tailwind CSS v4** for styling
- **Vue Router v5** for routing and navigation
- **Pinia v3** for state management
- **Axios** for HTTP client with interceptors
- **Chart.js + vue-chartjs** for data visualization

The application communicates with the backend server (`/server`) via REST API and includes authentication, role-based access control (RBAC), and real-time data refresh capabilities.

## Tech Stack Alignment with Member Module

The manager now uses the same web stack as the member client:
- Vue 3 with Composition API (`<script setup>`)
- Tailwind CSS v4 for styling
- Vue Router for navigation
- No formal component library (custom components with Tailwind)

## Pre-Push Validation

**Before pushing changes**, follow the validation checklist in **[../GITOPS.md](../GITOPS.md)**.

Manager-specific checks:
```bash
cd manager
pnpm lint                # Lint TypeScript/Vue code
pnpm build               # Verify production build
pnpm test                # Run tests (when configured)
```

## Development Workflow

### Prerequisites
- Node.js 18+ and pnpm
- Backend server running on port 9080

### Common Commands

```bash
# Install dependencies
pnpm install

# Development server (with hot reload)
pnpm dev          # Runs on http://localhost:3000
                  # Proxies /api and /auth to http://localhost:9080

# Build for production
pnpm build        # Outputs to dist/

# Preview production build
pnpm preview

# Lint code
pnpm lint
```

### Development Server Proxy

The Vite dev server proxies API requests to the backend:
- `/api/*` → `http://localhost:9080/api/*`
- `/auth/*` → `http://localhost:9080/auth/*`

No CORS configuration needed when using the proxy.

## Architecture

### Project Structure

```
src/
├── api/                 # API client modules
│   ├── client.ts        # Axios instance with interceptors
│   ├── auth.ts          # Authentication endpoints
│   ├── dashboard.ts     # Dashboard data endpoints
│   ├── analytics.ts     # Analytics endpoints
│   ├── setup.ts         # Setup endpoints
│   └── users.ts         # User management endpoints
├── components/
│   ├── charts/          # Chart.js visualization components
│   ├── common/          # Reusable UI components
│   └── layout/          # App shell (Header, Sidebar, AppLayout)
├── pages/               # Route components (Dashboard, Analytics, Users)
├── stores/              # Pinia stores for state management
│   ├── auth.ts          # Authentication store
│   └── setup.ts         # Setup state store
├── router/              # Vue Router configuration
│   └── index.ts
├── styles/              # Global styles
│   └── main.css
├── types/
│   ├── api.ts           # API request/response types
│   └── models.ts        # Domain model types
├── App.vue              # Root component with router guards
└── main.ts              # Application entry point
```

### Key Architectural Patterns

**1. API Layer (`src/api/`)**

Each API module exports an object with async methods that wrap Axios calls:
- Uses a shared `apiClient` from `client.ts` with automatic auth token injection
- Returns typed responses matching API schema
- Errors are handled by calling code

Example:
```typescript
export const dashboardApi = {
  getMetrics: async (range: '24h' | '7d' | '30d', interval: 'hour' | 'day') => {
    const response = await apiClient.get<DashboardMetricsResponse>(
      `/api/v1/dashboard/metrics`,
      { params: { range, interval } }
    )
    return response.data
  }
}
```

**2. Authentication Flow (`src/stores/auth.ts`)**

Authentication is managed through a Pinia store:
- Tokens stored in `localStorage` (`access_token`, `refresh_token`)
- `apiClient` interceptors automatically add `Authorization: Bearer <token>` to requests
- On 401 errors, the interceptor attempts token refresh via `/auth/refresh`
- Failed refresh redirects to `/login` and clears tokens

**3. State Management with Pinia**

State is managed using Pinia stores:
```typescript
export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const isAuthenticated = computed(() => !!user.value)

  async function login(email: string, password: string) {
    // ...
  }

  return { user, isAuthenticated, login }
})
```

**4. Role-Based Access Control**

Routes are protected using Vue Router navigation guards:
- Meta field `public: true` marks public routes
- Meta field `requiresAdmin: true` for admin-only routes
- Auth store provides `isAdmin` boolean for conditional rendering

**5. Responsive Layout**

The app shell (`AppLayout`) features:
- Fixed sidebar with resizable width (persists to `localStorage`)
- Mobile-responsive drawer (hidden on mobile by default)
- Fixed header with navigation toggle
- Main content area with `<RouterView />` for nested routes

### TypeScript Configuration

- Path alias: `@/*` maps to `src/*` (configured in `tsconfig.json` and `vite.config.ts`)
- Strict mode enabled
- JSX runtime: `preserve` for Vue SFC support

### Package Manager

This project uses **pnpm** as the package manager (not npm or yarn).
- Lock file: `pnpm-lock.yaml`
- Use `pnpm install` instead of `npm install`
- All scripts are run with `pnpm` instead of `npm`

## API Integration

### Backend Server

The Manager client expects the backend server to provide:

**Authentication endpoints:**
- `POST /auth/login` - Email/password login
- `POST /auth/logout` - Invalidate refresh token
- `POST /auth/refresh` - Refresh access token
- `GET /api/v1/user/profile` - Get current user profile

**Dashboard endpoints:**
- `GET /api/v1/dashboard/metrics?range={7d}&interval={hour}` - Activity metrics
- `GET /api/v1/dashboard/rankings` - Provider usage rankings
- `GET /api/v1/dashboard/members` - Member activity stats (admin only)

**Analytics endpoints:**
- `GET /api/v1/analytics/providers` - Provider-level analytics
- `GET /api/v1/analytics/users` - User-level analytics
- `GET /api/v1/analytics/history` - Request history with pagination
- `GET /api/v1/analytics/filters` - Filter options for analytics

**User management endpoints:**
- `GET /api/v1/users` - List users
- `POST /api/v1/users` - Create user
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

**Setup endpoints:**
- `GET /api/v1/setup/status` - Check if setup is required
- `POST /api/v1/setup/admin` - Create initial admin user

All responses follow the schema in `src/types/api.ts`.

## Component Patterns

### Page Components

Pages are route-level components that:
- Use Pinia stores for state management
- Call API functions for data fetching
- Handle loading/error states
- Compose UI components and charts
- Avoid business logic (delegates to API layer)

### Chart Components

Charts use Chart.js with vue-chartjs and accept:
- `data: T[]` - Array of data points
- `isLoading: boolean` - Show loading state
- Responsive containers (no fixed dimensions)

### Common Components

Reusable components in `src/components/common/`:
- `DateRangePicker.vue` - Date range selection with presets
- `MultiSelect.vue` - Multi-select dropdown
- `UserSelect.vue` - User selection with search

## Styling

- **Tailwind CSS v4** is the primary styling solution
- Custom styles in `src/styles/main.css`
- Dark mode support using `dark:` prefix
- Responsive breakpoints: `sm`, `md`, `lg`, `xl`, `2xl` (Tailwind breakpoint system)

## Error Handling

- API errors are caught by calling code and displayed to users
- Auth errors (401) trigger automatic logout via Axios interceptor
- Form validations use reactive refs with error state
- Error messages extracted via `getErrorMessage()` utility from `api/client`

## Testing

No test framework is currently configured. When adding tests:
- Prefer Vitest for unit/integration tests (compatible with Vite)
- Use Vue Test Utils for component tests
- Mock API calls with MSW (Mock Service Worker)

## Build and Deployment

The production build outputs static assets to `dist/`:
- `pnpm build` creates optimized bundle
- `dist/` can be served by any static web server or the backend server
- For production, the backend server should serve `dist/index.html` for SPA routing
