import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuth } from './contexts/AuthContext'
import { useSetup } from './contexts/SetupContext'
import ProtectedRoute from './components/ProtectedRoute'
import AppLayout from './components/layout/AppLayout'
import LoginPage from './pages/LoginPage'
import SetupPage from './pages/SetupPage'
import DashboardPage from './pages/DashboardPage'
import TokenProviderPage from './pages/TokenProviderPage'
import TokenUserPage from './pages/TokenUserPage'
import TokenHistoryPage from './pages/TokenHistoryPage'
import UserManagementPage from './pages/UserManagementPage'
import ProfilePage from './pages/ProfilePage'

export default function App() {
  const { isAuthenticated, isLoading } = useAuth()
  const { setupRequired, isLoading: setupLoading } = useSetup()

  if (isLoading || setupLoading) {
    return null
  }

  // If setup is required, only show setup routes
  if (setupRequired) {
    return (
      <Routes>
        <Route path="/setup" element={<SetupPage />} />
        <Route path="*" element={<Navigate to="/setup" replace />} />
      </Routes>
    )
  }

  return (
    <Routes>
      {/* Public route */}
      <Route
        path="/login"
        element={
          isAuthenticated ? <Navigate to="/" replace /> : <LoginPage />
        }
      />

      {/* Protected routes */}
      <Route
        element={
          <ProtectedRoute>
            <AppLayout />
          </ProtectedRoute>
        }
      >
        <Route path="/" element={<DashboardPage />} />
        <Route path="/profile" element={<ProfilePage />} />

        {/* Admin-only routes */}
        <Route
          path="/analytics/providers"
          element={
            <ProtectedRoute requireAdmin>
              <TokenProviderPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/analytics/users"
          element={
            <ProtectedRoute requireAdmin>
              <TokenUserPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/analytics/history"
          element={
            <ProtectedRoute requireAdmin>
              <TokenHistoryPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/users"
          element={
            <ProtectedRoute requireAdmin>
              <UserManagementPage />
            </ProtectedRoute>
          }
        />
      </Route>

      {/* Catch-all redirect */}
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
