import { useState, useCallback, useEffect } from 'react'
import { Outlet } from 'react-router-dom'
import { Box, Drawer, Toolbar } from '@mui/material'
import Header from './Header'
import Sidebar from './Sidebar'

const MIN_DRAWER_WIDTH = 180
const MAX_DRAWER_WIDTH = 400
const DEFAULT_DRAWER_WIDTH = 240

export default function AppLayout() {
  const [mobileOpen, setMobileOpen] = useState(false)
  const [drawerWidth, setDrawerWidth] = useState(() => {
    const saved = localStorage.getItem('sidebarWidth')
    return saved ? parseInt(saved, 10) : DEFAULT_DRAWER_WIDTH
  })
  const [isResizing, setIsResizing] = useState(false)

  const handleDrawerToggle = () => {
    setMobileOpen(!mobileOpen)
  }

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    e.preventDefault()
    setIsResizing(true)
  }, [])

  const handleMouseMove = useCallback(
    (e: MouseEvent) => {
      if (!isResizing) return

      const newWidth = e.clientX
      if (newWidth >= MIN_DRAWER_WIDTH && newWidth <= MAX_DRAWER_WIDTH) {
        setDrawerWidth(newWidth)
      }
    },
    [isResizing]
  )

  const handleMouseUp = useCallback(() => {
    if (isResizing) {
      setIsResizing(false)
      localStorage.setItem('sidebarWidth', drawerWidth.toString())
    }
  }, [isResizing, drawerWidth])

  useEffect(() => {
    if (isResizing) {
      document.addEventListener('mousemove', handleMouseMove)
      document.addEventListener('mouseup', handleMouseUp)
      document.body.style.cursor = 'col-resize'
      document.body.style.userSelect = 'none'
    }

    return () => {
      document.removeEventListener('mousemove', handleMouseMove)
      document.removeEventListener('mouseup', handleMouseUp)
      document.body.style.cursor = ''
      document.body.style.userSelect = ''
    }
  }, [isResizing, handleMouseMove, handleMouseUp])

  return (
    <Box sx={{ display: 'flex', minHeight: '100vh' }}>
      <Header onMenuClick={handleDrawerToggle} />

      {/* Mobile drawer */}
      <Drawer
        variant="temporary"
        open={mobileOpen}
        onClose={handleDrawerToggle}
        ModalProps={{ keepMounted: true }}
        sx={{
          display: { xs: 'block', sm: 'none' },
          '& .MuiDrawer-paper': {
            boxSizing: 'border-box',
            width: drawerWidth,
          },
        }}
      >
        <Sidebar />
      </Drawer>

      {/* Desktop drawer */}
      <Box
        component="nav"
        sx={{
          width: { sm: drawerWidth },
          flexShrink: { sm: 0 },
          display: { xs: 'none', sm: 'block' },
        }}
      >
        <Box
          sx={{
            width: drawerWidth,
            height: '100vh',
            position: 'fixed',
            top: 0,
            left: 0,
            bgcolor: 'background.paper',
            borderRight: 1,
            borderColor: 'divider',
            display: 'flex',
          }}
        >
          <Box sx={{ flexGrow: 1, overflow: 'hidden' }}>
            <Sidebar />
          </Box>

          {/* Resize handle */}
          <Box
            onMouseDown={handleMouseDown}
            sx={{
              width: 4,
              cursor: 'col-resize',
              bgcolor: isResizing ? 'primary.main' : 'transparent',
              transition: 'background-color 0.2s',
              '&:hover': {
                bgcolor: 'primary.light',
              },
            }}
          />
        </Box>
      </Box>

      {/* Main content */}
      <Box
        component="main"
        sx={{
          flexGrow: 1,
          p: 3,
          width: { sm: `calc(100% - ${drawerWidth}px)` },
          bgcolor: 'background.default',
          minHeight: '100vh',
        }}
      >
        <Toolbar />
        <Outlet />
      </Box>
    </Box>
  )
}
