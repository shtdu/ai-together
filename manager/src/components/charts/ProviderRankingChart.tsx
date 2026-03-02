import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip } from 'recharts'
import { Box, Typography, List, ListItem, ListItemText, Chip } from '@mui/material'
import type { ProviderRanking } from '../../types/api'

interface ProviderRankingChartProps {
  rankings: ProviderRanking[]
  isLoading?: boolean
}

const COLORS = ['#1976d2', '#2196f3', '#4caf50', '#ff9800', '#f44336', '#9c27b0', '#00bcd4', '#795548']

export default function ProviderRankingChart({ rankings, isLoading }: ProviderRankingChartProps) {
  if (isLoading) {
    return (
      <Box sx={{ height: 250, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Typography color="text.secondary">Loading...</Typography>
      </Box>
    )
  }

  if (rankings.length === 0) {
    return (
      <Box sx={{ height: 250, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Typography color="text.secondary">No provider data available</Typography>
      </Box>
    )
  }

  const chartData = rankings.map((r) => ({
    name: r.provider,
    value: r.total_tokens,
    percentage: r.percentage,
    requests: r.request_count,
  }))

  // Show pie chart if <= 5 providers, otherwise show list
  if (rankings.length <= 5) {
    return (
      <Box>
        <Typography variant="h6" gutterBottom>
          Provider Rankings (24h)
        </Typography>
        <ResponsiveContainer width="100%" height={200}>
          <PieChart>
            <Pie
              data={chartData}
              cx="50%"
              cy="50%"
              innerRadius={40}
              outerRadius={70}
              paddingAngle={2}
              dataKey="value"
              nameKey="name"
              label={({ name, percentage }) => `${name}: ${percentage}%`}
              labelLine={false}
            >
              {chartData.map((_, index) => (
                <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
              ))}
            </Pie>
            <Tooltip
              formatter={(value: number) => [value.toLocaleString(), 'Tokens']}
            />
          </PieChart>
        </ResponsiveContainer>
      </Box>
    )
  }

  // Show list for more providers
  return (
    <Box>
      <Typography variant="h6" gutterBottom>
        Provider Rankings (24h)
      </Typography>
      <List dense sx={{ maxHeight: 220, overflow: 'auto' }}>
        {rankings.map((ranking, index) => (
          <ListItem
            key={ranking.provider}
            secondaryAction={
              <Chip
                label={`${ranking.percentage}%`}
                size="small"
                sx={{ bgcolor: COLORS[index % COLORS.length], color: 'white' }}
              />
            }
          >
            <ListItemText
              primary={`${ranking.rank}. ${ranking.provider}`}
              secondary={`${ranking.total_tokens.toLocaleString()} tokens`}
            />
          </ListItem>
        ))}
      </List>
    </Box>
  )
}
