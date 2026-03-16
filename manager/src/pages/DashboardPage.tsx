import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  Box,
  Typography,
  Paper,
  Card,
  CardContent,
  IconButton,
  ToggleButton,
  ToggleButtonGroup,
  Tooltip,
} from '@mui/material'
import Grid from '@mui/material/Grid2'
import RefreshIcon from '@mui/icons-material/Refresh'
import { dashboardApi } from '../api/dashboard'
import { useAuth } from '../contexts/AuthContext.hooks'
import ActivityChart from '../components/charts/ActivityChart'
import ProviderRankingChart from '../components/charts/ProviderRankingChart'
import MemberStatsTable from '../components/MemberStatsTable'

type TimeRange = '24h' | '7d' | '30d'

export default function DashboardPage() {
  const { isAdmin } = useAuth()
  const [timeRange, setTimeRange] = useState<TimeRange>('7d')

  const {
    data: metricsData,
    isLoading: metricsLoading,
    refetch: refetchMetrics,
  } = useQuery({
    queryKey: ['dashboard', 'metrics', timeRange],
    queryFn: () => dashboardApi.getMetrics(timeRange, 'hour'),
  })

  const {
    data: rankingsData,
    isLoading: rankingsLoading,
    refetch: refetchRankings,
  } = useQuery({
    queryKey: ['dashboard', 'rankings'],
    queryFn: () => dashboardApi.getRankings(),
  })

  const {
    data: membersData,
    isLoading: membersLoading,
    refetch: refetchMembers,
  } = useQuery({
    queryKey: ['dashboard', 'members'],
    queryFn: () => dashboardApi.getMembers(),
    enabled: isAdmin,
  })

  const handleRefresh = () => {
    refetchMetrics()
    refetchRankings()
    if (isAdmin) {
      refetchMembers()
    }
  }

  const handleTimeRangeChange = (_: React.MouseEvent<HTMLElement>, newRange: TimeRange | null) => {
    if (newRange) {
      setTimeRange(newRange)
    }
  }

  const summary = metricsData?.summary

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4">Dashboard</Typography>
        <Box sx={{ display: 'flex', gap: 2, alignItems: 'center' }}>
          <ToggleButtonGroup
            value={timeRange}
            exclusive
            onChange={handleTimeRangeChange}
            size="small"
          >
            <ToggleButton value="24h">24h</ToggleButton>
            <ToggleButton value="7d">7 Days</ToggleButton>
            <ToggleButton value="30d">30 Days</ToggleButton>
          </ToggleButtonGroup>
          <Tooltip title="Refresh data">
            <IconButton onClick={handleRefresh}>
              <RefreshIcon />
            </IconButton>
          </Tooltip>
        </Box>
      </Box>

      {/* Summary Cards */}
      <Grid container spacing={2} sx={{ mb: 3 }}>
        <Grid size={{ xs: 6, md: 3 }}>
          <Card>
            <CardContent>
              <Typography color="text.secondary" variant="body2">
                Total Requests
              </Typography>
              <Typography variant="h5">
                {summary?.total_requests?.toLocaleString() || '0'}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid size={{ xs: 6, md: 3 }}>
          <Card>
            <CardContent>
              <Typography color="text.secondary" variant="body2">
                Total Tokens
              </Typography>
              <Typography variant="h5">
                {summary?.total_tokens
                  ? summary.total_tokens >= 1000000
                    ? `${(summary.total_tokens / 1000000).toFixed(1)}M`
                    : summary.total_tokens >= 1000
                    ? `${(summary.total_tokens / 1000).toFixed(0)}K`
                    : summary.total_tokens
                  : '0'}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid size={{ xs: 6, md: 3 }}>
          <Card>
            <CardContent>
              <Typography color="text.secondary" variant="body2">
                AI Active Time
              </Typography>
              <Typography variant="h5">
                {summary?.total_active_time_hours
                  ? `${summary.total_active_time_hours.toFixed(1)}h`
                  : '0h'}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid size={{ xs: 6, md: 3 }}>
          <Card>
            <CardContent>
              <Typography color="text.secondary" variant="body2">
                Est. Cost
              </Typography>
              <Typography variant="h5">
                ${summary?.estimated_cost?.toFixed(2) || '0.00'}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Charts Row */}
      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid size={{ xs: 12, md: 8 }}>
          <Paper sx={{ p: 3 }}>
            <ActivityChart
              data={metricsData?.data_points || []}
              isLoading={metricsLoading}
            />
          </Paper>
        </Grid>
        <Grid size={{ xs: 12, md: 4 }}>
          <Paper sx={{ p: 3, height: '100%' }}>
            <ProviderRankingChart
              rankings={rankingsData?.rankings || []}
              isLoading={rankingsLoading}
            />
          </Paper>
        </Grid>
      </Grid>

      {/* Member Stats Table (Admin only) */}
      {isAdmin && (
        <Paper sx={{ p: 3 }}>
          <MemberStatsTable
            members={membersData?.members || []}
            isLoading={membersLoading}
          />
        </Paper>
      )}
    </Box>
  )
}
