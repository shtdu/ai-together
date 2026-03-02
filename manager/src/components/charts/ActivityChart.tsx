import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts'
import { Box, ToggleButton, ToggleButtonGroup, Typography } from '@mui/material'
import { useState } from 'react'
import type { MetricDataPoint } from '../../types/api'

interface ActivityChartProps {
  data: MetricDataPoint[]
  isLoading?: boolean
}

type ChartView = 'active_time' | 'tokens'

export default function ActivityChart({ data, isLoading }: ActivityChartProps) {
  const [view, setView] = useState<ChartView>('active_time')

  const handleViewChange = (_: React.MouseEvent<HTMLElement>, newView: ChartView | null) => {
    if (newView) {
      setView(newView)
    }
  }

  // Format data for the chart
  const chartData = data.map((point) => ({
    timestamp: new Date(point.timestamp).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
    }),
    fullDate: new Date(point.timestamp).toLocaleString(),
    activeTimeHours: point.active_time_seconds / 3600,
    totalTokens: point.total_tokens,
    inputTokens: point.input_tokens,
    outputTokens: point.output_tokens,
    requestCount: point.request_count,
  }))

  if (isLoading) {
    return (
      <Box sx={{ height: 300, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Typography color="text.secondary">Loading...</Typography>
      </Box>
    )
  }

  if (data.length === 0) {
    return (
      <Box sx={{ height: 300, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Typography color="text.secondary">No data available</Typography>
      </Box>
    )
  }

  return (
    <Box>
      <Box sx={{ mb: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Typography variant="h6">
          {view === 'active_time' ? 'AI Active Time' : 'Token Volume'}
        </Typography>
        <ToggleButtonGroup
          value={view}
          exclusive
          onChange={handleViewChange}
          size="small"
        >
          <ToggleButton value="active_time">Active Time</ToggleButton>
          <ToggleButton value="tokens">Tokens</ToggleButton>
        </ToggleButtonGroup>
      </Box>

      <ResponsiveContainer width="100%" height={250}>
        {view === 'active_time' ? (
          <LineChart data={chartData}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="timestamp" fontSize={12} />
            <YAxis
              fontSize={12}
              tickFormatter={(value) => `${value.toFixed(1)}h`}
            />
            <Tooltip
              formatter={(value: number) => [`${value.toFixed(2)} hours`, 'Active Time']}
            />
            <Legend />
            <Line
              type="monotone"
              dataKey="activeTimeHours"
              name="AI Active Time"
              stroke="#1976d2"
              strokeWidth={2}
              dot={{ r: 3 }}
              activeDot={{ r: 5 }}
            />
          </LineChart>
        ) : (
          <LineChart data={chartData}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="timestamp" fontSize={12} />
            <YAxis
              fontSize={12}
              tickFormatter={(value) => {
                if (value >= 1000000) return `${(value / 1000000).toFixed(1)}M`
                if (value >= 1000) return `${(value / 1000).toFixed(0)}K`
                return value.toString()
              }}
            />
            <Tooltip
              formatter={(value: number) => [value.toLocaleString(), 'Tokens']}
            />
            <Legend />
            <Line
              type="monotone"
              dataKey="inputTokens"
              name="Input Tokens"
              stroke="#2196f3"
              strokeWidth={2}
              dot={{ r: 2 }}
            />
            <Line
              type="monotone"
              dataKey="outputTokens"
              name="Output Tokens"
              stroke="#4caf50"
              strokeWidth={2}
              dot={{ r: 2 }}
            />
          </LineChart>
        )}
      </ResponsiveContainer>
    </Box>
  )
}
