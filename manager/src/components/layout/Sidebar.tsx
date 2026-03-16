import { useLocation, useNavigate } from 'react-router-dom'
import {
  List,
  ListItem,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Box,
  Toolbar,
} from '@mui/material'
import DashboardIcon from '@mui/icons-material/Dashboard'
import BarChartIcon from '@mui/icons-material/BarChart'
import HistoryIcon from '@mui/icons-material/History'
import GroupIcon from '@mui/icons-material/Group'
import StorageIcon from '@mui/icons-material/Storage'
import { useAuth } from '../../contexts/AuthContext.hooks'

interface NavItem {
  label: string
  path: string
  icon: React.ReactNode
  adminOnly?: boolean
}

const navItems: NavItem[] = [
  { label: 'Dashboard', path: '/', icon: <DashboardIcon /> },
  { label: 'Provider Stats', path: '/analytics/providers', icon: <StorageIcon />, adminOnly: true },
  { label: 'User Stats', path: '/analytics/users', icon: <BarChartIcon />, adminOnly: true },
  { label: 'History', path: '/analytics/history', icon: <HistoryIcon />, adminOnly: true },
  { label: 'Users', path: '/users', icon: <GroupIcon />, adminOnly: true },
]

export default function Sidebar() {
  const location = useLocation()
  const navigate = useNavigate()
  const { isAdmin } = useAuth()

  const filteredItems = navItems.filter((item) => !item.adminOnly || isAdmin)

  return (
    <Box sx={{ overflow: 'auto', height: '100%' }}>
      <Toolbar /> {/* Spacer for header */}
      <List sx={{ pt: 1 }}>
        {filteredItems.map((item) => (
          <ListItem key={item.path} disablePadding>
            <ListItemButton
              selected={location.pathname === item.path}
              onClick={() => navigate(item.path)}
              sx={{
                mx: 1,
                borderRadius: 1,
                '&.Mui-selected': {
                  bgcolor: 'primary.main',
                  color: 'primary.contrastText',
                  '&:hover': {
                    bgcolor: 'primary.dark',
                  },
                  '& .MuiListItemIcon-root': {
                    color: 'primary.contrastText',
                  },
                },
              }}
            >
              <ListItemIcon sx={{ minWidth: 40 }}>{item.icon}</ListItemIcon>
              <ListItemText primary={item.label} />
            </ListItemButton>
          </ListItem>
        ))}
      </List>
    </Box>
  )
}
