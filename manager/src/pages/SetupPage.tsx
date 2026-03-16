import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Box,
  Card,
  CardContent,
  TextField,
  Button,
  Typography,
  Alert,
  CircularProgress,
  Stepper,
  Step,
  StepLabel,
  IconButton,
  InputAdornment,
} from '@mui/material'
import { Visibility, VisibilityOff, ContentCopy } from '@mui/icons-material'
import { setupApi } from '../api/setup'
import { getErrorMessage } from '../api/client'
import { useSetup } from '../contexts/SetupContext.hooks'

const steps = ['Organization', 'Administrator Account', 'Complete']

function generatePassword(length = 10): string {
  const upper = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'
  const lower = 'abcdefghijklmnopqrstuvwxyz'
  const numbers = '0123456789'
  const all = upper + lower + numbers

  // Ensure at least 1 of each type
  const password = [
    upper[Math.floor(Math.random() * upper.length)],
    lower[Math.floor(Math.random() * lower.length)],
    numbers[Math.floor(Math.random() * numbers.length)],
  ]

  // Fill rest randomly
  for (let i = 3; i < length; i++) {
    password.push(all[Math.floor(Math.random() * all.length)])
  }

  // Shuffle
  return password.sort(() => Math.random() - 0.5).join('')
}

export default function SetupPage() {
  const navigate = useNavigate()
  const { checkSetupStatus } = useSetup()
  const [activeStep, setActiveStep] = useState(0)
  const [error, setError] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [showPassword, setShowPassword] = useState(false)
  const [countdown, setCountdown] = useState(20)

  // Form state
  const [organizationName, setOrganizationName] = useState('')
  const [adminEmail, setAdminEmail] = useState('')
  const [adminName, setAdminName] = useState('')
  const [adminPassword, setAdminPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')

  // Auto-redirect countdown when setup is complete
  useEffect(() => {
    if (activeStep === 2) {
      if (countdown > 0) {
        const timer = setTimeout(() => setCountdown(countdown - 1), 1000)
        return () => clearTimeout(timer)
      } else {
        handleGoToLogin()
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [activeStep, countdown])

  const handleNext = () => {
    setError('')

    // Validate current step
    if (activeStep === 0) {
      if (!organizationName.trim()) {
        setError('Organization name is required')
        return
      }
    } else if (activeStep === 1) {
      if (!adminEmail.trim()) {
        setError('Admin email is required')
        return
      }
      if (!adminName.trim()) {
        setError('Admin name is required')
        return
      }
      if (!adminPassword) {
        setError('Password is required')
        return
      }
      if (adminPassword.length < 8) {
        setError('Password must be at least 8 characters')
        return
      }
      if (adminPassword !== confirmPassword) {
        setError('Passwords do not match')
        return
      }

      // Submit to server
      handleSubmit()
      return
    }

    setActiveStep((prev) => prev + 1)
  }

  const handleBack = () => {
    setError('')
    setActiveStep((prev) => prev - 1)
  }

  const handleSubmit = async () => {
    setError('')
    setIsLoading(true)

    try {
      await setupApi.createAdmin({
        organization_name: organizationName,
        admin_email: adminEmail,
        admin_name: adminName,
        admin_password: adminPassword,
      })
      setActiveStep(2) // Go to success step
    } catch (err) {
      setError(getErrorMessage(err))
    } finally {
      setIsLoading(false)
    }
  }

  const handleGoToLogin = async () => {
    // Refresh setup status before navigating so we don't redirect back to setup
    await checkSetupStatus()
    navigate('/login')
  }

  const handleGeneratePassword = () => {
    const newPassword = generatePassword(10)
    setAdminPassword(newPassword)
    setConfirmPassword(newPassword)
  }

  const handleCopyPassword = () => {
    navigator.clipboard.writeText(adminPassword)
  }

  return (
    <Box
      sx={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        bgcolor: 'background.default',
        p: 2,
      }}
    >
      <Card sx={{ maxWidth: 600, width: '100%' }}>
        <CardContent sx={{ p: 4 }}>
          <Typography variant="h4" component="h1" gutterBottom textAlign="center">
            Welcome to Code Together
          </Typography>
          <Typography
            variant="body2"
            color="text.secondary"
            textAlign="center"
            sx={{ mb: 4 }}
          >
            Let's set up your organization
          </Typography>

          <Stepper activeStep={activeStep} sx={{ mb: 4 }}>
            {steps.map((label) => (
              <Step key={label}>
                <StepLabel>{label}</StepLabel>
              </Step>
            ))}
          </Stepper>

          {error && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {error}
            </Alert>
          )}

          {activeStep === 0 && (
            <Box>
              <Typography variant="h6" gutterBottom>
                Organization Information
              </Typography>
              <TextField
                fullWidth
                label="Organization Name"
                value={organizationName}
                onChange={(e) => setOrganizationName(e.target.value)}
                margin="normal"
                required
                autoFocus
                helperText="Enter the name of your organization or team"
              />
            </Box>
          )}

          {activeStep === 1 && (
            <Box>
              <Typography variant="h6" gutterBottom>
                Administrator Account
              </Typography>
              <TextField
                fullWidth
                label="Admin Email"
                type="email"
                value={adminEmail}
                onChange={(e) => setAdminEmail(e.target.value)}
                margin="normal"
                required
                autoFocus
              />
              <TextField
                fullWidth
                label="Admin Name"
                value={adminName}
                onChange={(e) => setAdminName(e.target.value)}
                margin="normal"
                required
              />
              <Box sx={{ position: 'relative' }}>
                <TextField
                  fullWidth
                  label="Password"
                  type={showPassword ? 'text' : 'password'}
                  value={adminPassword}
                  onChange={(e) => setAdminPassword(e.target.value)}
                  margin="normal"
                  required
                  helperText="Minimum 8 characters"
                  InputProps={{
                    endAdornment: (
                      <InputAdornment position="end">
                        <IconButton
                          onClick={() => setShowPassword(!showPassword)}
                          edge="end"
                        >
                          {showPassword ? <VisibilityOff /> : <Visibility />}
                        </IconButton>
                        {adminPassword && (
                          <IconButton onClick={handleCopyPassword} edge="end">
                            <ContentCopy />
                          </IconButton>
                        )}
                      </InputAdornment>
                    ),
                  }}
                />
              </Box>
              <TextField
                fullWidth
                label="Confirm Password"
                type={showPassword ? 'text' : 'password'}
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                margin="normal"
                required
              />
              <Button
                variant="outlined"
                onClick={handleGeneratePassword}
                sx={{ mt: 2 }}
              >
                Generate Secure Password
              </Button>
            </Box>
          )}

          {activeStep === 2 && (
            <Box sx={{ textAlign: 'center' }}>
              <Typography variant="h5" gutterBottom color="success.main">
                ✓ Setup Complete!
              </Typography>
              <Typography variant="body1" color="text.secondary" sx={{ mb: 3 }}>
                Your organization has been set up successfully. You can now log in with
                your administrator account.
              </Typography>
              <Alert severity="success" sx={{ mb: 3, textAlign: 'left' }}>
                <Typography variant="body2" sx={{ mb: 1 }}>
                  <strong>Email:</strong> {adminEmail}
                </Typography>
                <Typography variant="body2">
                  <strong>Password:</strong> {adminPassword}
                </Typography>
              </Alert>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                Redirecting to login in {countdown} second{countdown !== 1 ? 's' : ''}...
              </Typography>
              <Button
                variant="contained"
                size="large"
                onClick={handleGoToLogin}
              >
                Go to Login Now
              </Button>
            </Box>
          )}

          {activeStep < 2 && (
            <Box sx={{ display: 'flex', justifyContent: 'space-between', mt: 3 }}>
              <Button onClick={handleBack} disabled={activeStep === 0 || isLoading}>
                Back
              </Button>
              <Button
                variant="contained"
                onClick={handleNext}
                disabled={isLoading}
                startIcon={isLoading ? <CircularProgress size={20} /> : null}
              >
                {activeStep === 1 ? 'Create Account' : 'Next'}
              </Button>
            </Box>
          )}
        </CardContent>
      </Card>
    </Box>
  )
}
