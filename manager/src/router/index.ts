import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

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

export default createRouter({
  history: createWebHistory(),
  routes,
})
