import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Typography,
  Box,
} from '@mui/material'
import type { MemberStats } from '../types/api'

interface MemberStatsTableProps {
  members: MemberStats[]
  isLoading?: boolean
}

export default function MemberStatsTable({ members, isLoading }: MemberStatsTableProps) {
  if (isLoading) {
    return (
      <Box sx={{ p: 3, textAlign: 'center' }}>
        <Typography color="text.secondary">Loading...</Typography>
      </Box>
    )
  }

  if (members.length === 0) {
    return (
      <Box sx={{ p: 3, textAlign: 'center' }}>
        <Typography color="text.secondary">No member data available</Typography>
      </Box>
    )
  }

  const formatTime = (hours: number): string => {
    if (hours < 1) {
      return `${Math.round(hours * 60)} min`
    }
    return `${hours.toFixed(1)} hrs`
  }

  const formatTokens = (tokens: number): string => {
    if (tokens >= 1000000) {
      return `${(tokens / 1000000).toFixed(1)}M`
    }
    if (tokens >= 1000) {
      return `${(tokens / 1000).toFixed(1)}K`
    }
    return tokens.toString()
  }

  const formatDate = (dateStr: string): string => {
    const date = new Date(dateStr)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffHours = diffMs / (1000 * 60 * 60)

    if (diffHours < 1) {
      return 'Just now'
    }
    if (diffHours < 24) {
      return `${Math.round(diffHours)}h ago`
    }
    const diffDays = Math.floor(diffHours / 24)
    if (diffDays < 7) {
      return `${diffDays}d ago`
    }
    return date.toLocaleDateString()
  }

  return (
    <Box>
      <Typography variant="h6" gutterBottom>
        Member Statistics
      </Typography>
      <TableContainer component={Paper} variant="outlined">
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Name</TableCell>
              <TableCell align="right">AI Active Time</TableCell>
              <TableCell align="right">Total Tokens</TableCell>
              <TableCell align="right">Avg Tokens/Day</TableCell>
              <TableCell align="right">Last Active</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {members.map((member) => (
              <TableRow key={member.user_id} hover>
                <TableCell>
                  <Typography variant="body2">{member.name}</Typography>
                  <Typography variant="caption" color="text.secondary">
                    {member.email}
                  </Typography>
                </TableCell>
                <TableCell align="right">{formatTime(member.active_time_hours)}</TableCell>
                <TableCell align="right">{formatTokens(member.total_tokens)}</TableCell>
                <TableCell align="right">{formatTokens(member.avg_tokens_per_day)}</TableCell>
                <TableCell align="right">{formatDate(member.last_active)}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  )
}
