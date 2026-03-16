import { useState } from 'react'
import {
  Box,
  Typography,
  Paper,
  TextField,
  Button,
  Alert,
  Divider,
} from '@mui/material'
import Grid from '@mui/material/Grid2'
import { useAuth } from '../contexts/AuthContext.hooks'

export default function ProfilePage() {
  const { user } = useAuth()
  const [name, setName] = useState(user?.name || '')
  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  const handleUpdateProfile = (e: React.FormEvent) => {
    e.preventDefault()
    // TODO: Implement profile update
    setMessage({ type: 'success', text: 'Profile update will be implemented with backend API.' })
  }

  const handleUpdatePassword = (e: React.FormEvent) => {
    e.preventDefault()
    if (newPassword !== confirmPassword) {
      setMessage({ type: 'error', text: 'New passwords do not match.' })
      return
    }
    // TODO: Implement password update
    setMessage({ type: 'success', text: 'Password update will be implemented with backend API.' })
  }

  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        Profile
      </Typography>

      {message && (
        <Alert severity={message.type} sx={{ mb: 2 }} onClose={() => setMessage(null)}>
          {message.text}
        </Alert>
      )}

      <Grid container spacing={3}>
        <Grid size={{ xs: 12, md: 6 }}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              Account Information
            </Typography>
            <Box component="form" onSubmit={handleUpdateProfile}>
              <TextField
                fullWidth
                label="Email"
                value={user?.email || ''}
                margin="normal"
                disabled
                helperText="Email cannot be changed"
              />
              <TextField
                fullWidth
                label="Name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                margin="normal"
              />
              <TextField
                fullWidth
                label="Role"
                value={user?.role === 'manager' ? 'Admin' : 'Member'}
                margin="normal"
                disabled
              />
              <Button type="submit" variant="contained" sx={{ mt: 2 }}>
                Update Profile
              </Button>
            </Box>
          </Paper>
        </Grid>

        <Grid size={{ xs: 12, md: 6 }}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              Change Password
            </Typography>
            <Box component="form" onSubmit={handleUpdatePassword}>
              <TextField
                fullWidth
                label="Current Password"
                type="password"
                value={currentPassword}
                onChange={(e) => setCurrentPassword(e.target.value)}
                margin="normal"
              />
              <Divider sx={{ my: 2 }} />
              <TextField
                fullWidth
                label="New Password"
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                margin="normal"
              />
              <TextField
                fullWidth
                label="Confirm New Password"
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                margin="normal"
              />
              <Button type="submit" variant="contained" sx={{ mt: 2 }}>
                Change Password
              </Button>
            </Box>
          </Paper>
        </Grid>
      </Grid>
    </Box>
  )
}
