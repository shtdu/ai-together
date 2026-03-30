import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw, RouteLocationNormalized } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useSetupStore } from '../stores/setup'

// Layouts
import AppLayout from '../components/layout/AppLayout.vue'

// Pages
import LoginPage from '../pages/LoginPage.vue'
import SetupPage from '../pages/SetupPage.vue'
import DashboardPage from '../pages/DashboardPage.vue'
import TokenProviderPage from '../pages/TokenProviderPage.vue'
import TokenUserPage from '../pages/TokenUserPage.vue'
import TokenHistoryPage from '../pages/TokenHistoryPage.vue'
import UserManagementPage from '../pages/UserManagementPage.vue'
import ProfilePage from '../pages/ProfilePage.vue'
import PersonalAnalyticsPage from '../pages/PersonalAnalyticsPage.vue'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: LoginPage,
    meta: { public: true },
  },
  {
    path: '/setup',
    name: 'setup',
    component: SetupPage,
    meta: { public: true },
  },
  {
    path: '/',
    component: AppLayout,
    children: [
      {
        path: '',
        name: 'dashboard',
        component: DashboardPage,
      },
      {
        path: 'profile',
        name: 'profile',
        component: ProfilePage,
      },
      {
        path: 'analytics/personal',
        name: 'analytics-personal',
        component: PersonalAnalyticsPage,
      },
      {
        path: 'analytics/providers',
        name: 'analytics-providers',
        component: TokenProviderPage,
        meta: { requiresAdmin: true },
      },
      {
        path: 'analytics/users',
        name: 'analytics-users',
        component: TokenUserPage,
        meta: { requiresAdmin: true },
      },
      {
        path: 'analytics/history',
        name: 'analytics-history',
        component: TokenHistoryPage,
        meta: { requiresAdmin: true },
      },
      {
        path: 'users',
        name: 'users',
        component: UserManagementPage,
        meta: { requiresAdmin: true },
      },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/',
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// Global navigation guard - registered at module load time
// Using return instead of next() for Vue Router v5
router.beforeEach(async (to: RouteLocationNormalized) => {
  const authStore = useAuthStore()
  const setupStore = useSetupStore()

  // Wait for stores to initialize if not already done
  if (!setupStore.isInitialized) {
    await setupStore.initialize()
  }
  if (!authStore.isInitialized) {
    await authStore.initialize()
  }

  // If setup is required, only allow setup route
  if (setupStore.setupRequired === true && to.name !== 'setup') {
    return { name: 'setup' }
  }

  // Public routes
  if (to.meta.public) {
    // If authenticated and trying to access login, redirect to dashboard
    if (to.name === 'login' && authStore.isAuthenticated) {
      return { name: 'dashboard' }
    }
    return
  }

  // Protected routes
  if (!authStore.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  // Admin-only routes
  if (to.meta.requiresAdmin && !authStore.isAdmin) {
    return { name: 'dashboard' }
  }

  return
})

export default router
