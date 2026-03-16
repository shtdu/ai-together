# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

The Manager client is the administrative web interface for the AI Together platform. It provides team-level dashboards, usage analytics, user management, and token usage tracking for AI coding tools (Claude Code, Codex, and OpenCode).

This is a **React 18 + TypeScript** single-page application (SPA) built with:
- **Vite** for build tooling and dev server
- **Material UI (MUI) v6** for UI components
- **React Router v7** for routing and navigation
- **TanStack React Query** for server state management and caching
- **Axios** for HTTP client with interceptors
- **Recharts** for data visualization

The application communicates with the backend server (`/server`) via REST API and includes authentication, role-based access control (RBAC), and real-time data refresh capabilities.

## Pre-Push Validation

**Before pushing changes**, follow the validation checklist in **[../GITOPS.md](../GITOPS.md)**.

Manager-specific checks:
```bash
cd manager
pnpm lint                # Lint TypeScript code
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
│   └── users.ts         # User management endpoints
├── components/
│   ├── charts/          # Recharts visualization components
│   ├── common/          # Reusable UI components
│   ├── layout/          # App shell (Header, Sidebar, AppLayout)
│   └── *.tsx            # Feature components
├── contexts/
│   └── AuthContext.tsx  # Authentication state management
├── pages/               # Route components (Dashboard, Analytics, Users)
├── types/
│   ├── api.ts           # API request/response types
│   └── models.ts        # Domain model types
├── App.tsx              # Root routing configuration
├── main.tsx             # Application entry point
└── theme.ts             # MUI theme configuration
```

### Key Architectural Patterns

**1. API Layer (`src/api/`)**

Each API module exports an object with async methods that wrap Axios calls:
- Uses a shared `apiClient` from `client.ts` with automatic auth token injection
- Returns typed responses matching API schema
- Errors are handled by React Query error boundaries

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

**2. Authentication Flow (`src/contexts/AuthContext.tsx`)**

Authentication is managed through a React Context with the following flow:
- Tokens stored in `localStorage` (`access_token`, `refresh_token`)
- `apiClient` interceptors automatically add `Authorization: Bearer <token>` to requests
- On 401 errors, the interceptor attempts token refresh via `/auth/refresh`
- Failed refresh redirects to `/login` and clears tokens
- `AuthProvider` wraps the app and exposes `useAuth()` hook for auth state

**3. Data Fetching with React Query**

Pages use `useQuery` for data fetching with automatic caching and refetching:
```typescript
const { data, isLoading, error, refetch } = useQuery({
  queryKey: ['dashboard', 'metrics', timeRange],
  queryFn: () => dashboardApi.getMetrics(timeRange, 'hour'),
})
```

Query keys follow the pattern: `['resource', 'id?', 'params?']` for cache invalidation.

**4. Role-Based Access Control**

Routes are protected using nested `ProtectedRoute` components:
- Default: requires authentication (any role)
- `requireAdmin`: requires `user.role === 'manager'`
- Auth context provides `isAdmin` boolean for conditional rendering

**5. Responsive Layout**

The app shell (`AppLayout`) features:
- Fixed sidebar with resizable width (persists to `localStorage`)
- Mobile-responsive drawer (temporary on mobile, permanent on desktop)
- Fixed header with navigation toggle
- Main content area with `<Outlet />` for nested routes

### TypeScript Configuration

- Path alias: `@/*` maps to `src/*` (configured in `tsconfig.json`)
- Strict mode enabled with `noUnusedLocals` and `noUnusedParameters`
- JSX runtime: `react-jsx` (no need to import React)

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
- `GET /auth/profile` - Get current user profile

**Dashboard endpoints:**
- `GET /api/v1/dashboard/metrics?range={7d}&interval={hour}` - Activity metrics
- `GET /api/v1/dashboard/rankings` - Provider usage rankings
- `GET /api/v1/dashboard/members` - Member activity stats (admin only)

**Analytics endpoints:**
- `GET /api/v1/analytics/providers` - Provider-level analytics
- `GET /api/v1/analytics/users` - User-level analytics
- `GET /api/v1/analytics/history` - Request history with pagination

**User management endpoints:**
- `GET /api/v1/users` - List users
- `POST /api/v1/users` - Create user
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

All responses follow the schema in `src/types/api.ts`.

## Component Patterns

### Page Components

Pages are route-level components that:
- Use `useQuery` for data fetching
- Handle loading/error states
- Compose UI components and charts
- Avoid business logic (delegates to API layer)

### Chart Components

Charts use Recharts and accept:
- `data: T[]` - Array of data points
- `isLoading: boolean` - Show skeleton loading state
- Responsive containers (no fixed dimensions)

### Common Components

Reusable components in `src/components/common/`:
- `DateRangePicker` - Date range selection with MUI X DatePickers
- `MultiSelect` - Multi-select dropdown with chips
- `UserSelect` - User selection with search

## Styling

- **Material UI** uses the `theme.ts` configuration for colors, spacing, and typography
- **Tailwind CSS** is available for utility classes (configured but minimal usage)
- Component styling uses `sx` prop or `styled()` from MUI
- Responsive breakpoints: `xs`, `sm`, `md`, `lg`, `xl` (MUI breakpoint system)

## Error Handling

- API errors are caught by React Query and can be displayed with error boundaries
- Auth errors (401) trigger automatic logout via Axios interceptor
- Form validations use controlled components with error state
- Error messages extracted via `getErrorMessage()` utility from `api/client`

## Testing

No test framework is currently configured. When adding tests:
- Prefer Vitest for unit/integration tests (compatible with Vite)
- Use React Testing Library for component tests
- Mock API calls with MSW (Mock Service Worker)

## Build and Deployment

The production build outputs static assets to `dist/`:
- `npm run build` creates optimized bundle
- `dist/` can be served by any static web server or the backend server
- For production, the backend server should serve `dist/index.html` for SPA routing
