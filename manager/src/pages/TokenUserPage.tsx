import { useState } from 'react';
import { Box, Typography, Paper, Stack, Button, Alert, CircularProgress } from '@mui/material';
import Grid from '@mui/material/Grid2';
import { DataGrid, GridColDef } from '@mui/x-data-grid';
import { useQuery } from '@tanstack/react-query';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import RefreshIcon from '@mui/icons-material/Refresh';
import DownloadIcon from '@mui/icons-material/Download';
import SearchIcon from '@mui/icons-material/Search';
import dayjs from 'dayjs';
import DateRangePicker from '../components/common/DateRangePicker';
import MultiSelect from '../components/common/MultiSelect';
import UserSelect from '../components/common/UserSelect';
import { analyticsApi } from '../api/analytics';

function formatNumber(num: number): string {
  if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`;
  if (num >= 1000) return `${(num / 1000).toFixed(1)}K`;
  return num.toString();
}

function formatCost(cost: number): string {
  return `$${cost.toFixed(4)}`;
}

export default function TokenUserPage() {
  const [startDate, setStartDate] = useState(dayjs().subtract(7, 'day').format('YYYY-MM-DD'));
  const [endDate, setEndDate] = useState(dayjs().format('YYYY-MM-DD'));
  const [selectedUsers, setSelectedUsers] = useState<number[]>([]);
  const [selectedProviders, setSelectedProviders] = useState<string[]>([]);

  // Search params (applied when Search is clicked)
  const [searchParams, setSearchParams] = useState({
    startDate: dayjs().subtract(7, 'day').format('YYYY-MM-DD'),
    endDate: dayjs().format('YYYY-MM-DD'),
    userIds: [] as number[],
    providers: [] as string[],
  });

  const { data: filterOptions, isLoading: filterLoading } = useQuery({
    queryKey: ['filterOptions'],
    queryFn: analyticsApi.getFilterOptions,
  });

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['userAnalytics', searchParams],
    queryFn: () => analyticsApi.getUserAnalytics({
      start_date: searchParams.startDate,
      end_date: searchParams.endDate,
      user_ids: searchParams.userIds.length > 0 ? searchParams.userIds : undefined,
      providers: searchParams.providers.length > 0 ? searchParams.providers : undefined,
    }),
  });

  const handleDateChange = (start: string, end: string) => {
    setStartDate(start);
    setEndDate(end);
  };

  const handleSearch = () => {
    setSearchParams({
      startDate,
      endDate,
      userIds: selectedUsers,
      providers: selectedProviders,
    });
  };

  const handleExportCSV = () => {
    if (!data) return;

    const headers = ['Rank', 'User', 'Total Tokens', 'Total Requests', 'Percentage', 'Estimated Cost'];
    const rows = data.leaderboard.map(u => [
      u.rank,
      u.name,
      u.total_tokens,
      u.total_requests,
      u.percentage.toFixed(2) + '%',
      u.total_cost.toFixed(4),
    ]);

    const csv = [headers.join(','), ...rows.map(r => r.join(','))].join('\n');
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `user-analytics-${startDate}-${endDate}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  };

  // Columns for leaderboard table
  const leaderboardColumns: GridColDef[] = [
    { field: 'rank', headerName: 'Rank', width: 70 },
    { field: 'name', headerName: 'User', flex: 1 },
    { field: 'total_tokens', headerName: 'Total Tokens', type: 'number', width: 130, valueFormatter: (value: number) => formatNumber(value) },
    { field: 'total_requests', headerName: 'Requests', type: 'number', width: 100 },
    { field: 'percentage', headerName: 'Share', type: 'number', width: 100, valueFormatter: (value: number) => `${value.toFixed(1)}%` },
    { field: 'total_cost', headerName: 'Est. Cost', type: 'number', width: 100, valueFormatter: (value: number) => formatCost(value) },
  ];

  // Columns for details table
  const detailsColumns: GridColDef[] = [
    { field: 'user_name', headerName: 'User', flex: 1 },
    { field: 'date', headerName: 'Date', width: 110 },
    { field: 'provider', headerName: 'Provider', width: 120 },
    { field: 'model', headerName: 'Model', width: 180 },
    { field: 'total_tokens', headerName: 'Tokens', type: 'number', width: 100, valueFormatter: (value: number) => formatNumber(value) },
    { field: 'avg_latency_ms', headerName: 'Avg Latency', type: 'number', width: 110, valueFormatter: (value: number) => `${value.toFixed(0)}ms` },
    { field: 'total_cost', headerName: 'Cost', type: 'number', width: 90, valueFormatter: (value: number) => formatCost(value) },
  ];

  // Prepare ranking chart data (top 10)
  const rankingData = data?.leaderboard?.slice(0, 10).map(u => ({
    name: u.name,
    tokens: u.total_tokens,
  })) || [];

  if (error) {
    return (
      <Alert severity="error">
        Failed to load user analytics. Please try again.
      </Alert>
    );
  }

  return (
    <Box>
      <Stack direction="row" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4">User Analytics</Typography>
        <Stack direction="row" spacing={1}>
          <Button startIcon={<RefreshIcon />} onClick={() => refetch()}>
            Refresh
          </Button>
          <Button startIcon={<DownloadIcon />} variant="outlined" onClick={handleExportCSV} disabled={!data}>
            Export CSV
          </Button>
        </Stack>
      </Stack>

      {/* Filters */}
      <Paper sx={{ p: 2, mb: 3 }}>
        <Stack direction="row" spacing={2} alignItems="center" flexWrap="wrap" useFlexGap>
          <DateRangePicker
            startDate={startDate}
            endDate={endDate}
            onChange={handleDateChange}
          />
          <UserSelect
            label="Users"
            options={filterOptions?.users || []}
            value={selectedUsers}
            onChange={setSelectedUsers}
          />
          <MultiSelect
            label="Providers"
            options={filterOptions?.providers || []}
            value={selectedProviders}
            onChange={setSelectedProviders}
          />
          <Button
            variant="contained"
            startIcon={<SearchIcon />}
            onClick={handleSearch}
            disabled={filterLoading}
          >
            Search
          </Button>
        </Stack>
      </Paper>

      {isLoading ? (
        <Box display="flex" justifyContent="center" alignItems="center" minHeight={400}>
          <CircularProgress />
        </Box>
      ) : (
        <Grid container spacing={3}>
          {/* User Ranking Chart */}
          <Grid size={{ xs: 12, md: 5 }}>
            <Paper sx={{ p: 2, height: 450 }}>
              <Typography variant="h6" gutterBottom>Top 10 Users by Token Usage</Typography>
              <ResponsiveContainer width="100%" height={380}>
                <BarChart data={rankingData} layout="vertical">
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis type="number" tickFormatter={formatNumber} />
                  <YAxis dataKey="name" type="category" width={100} tick={{ fontSize: 12 }} />
                  <Tooltip formatter={(value: number) => formatNumber(value)} />
                  <Bar dataKey="tokens" fill="#8884d8" name="Tokens" />
                </BarChart>
              </ResponsiveContainer>
            </Paper>
          </Grid>

          {/* Leaderboard Table */}
          <Grid size={{ xs: 12, md: 7 }}>
            <Paper sx={{ p: 2, height: 450 }}>
              <Typography variant="h6" gutterBottom>User Leaderboard</Typography>
              <DataGrid
                rows={data?.leaderboard?.map((u) => ({ id: u.user_id, ...u })) || []}
                columns={leaderboardColumns}
                pageSizeOptions={[5, 10]}
                initialState={{
                  pagination: { paginationModel: { pageSize: 5 } },
                }}
                disableRowSelectionOnClick
                sx={{ height: 380 }}
              />
            </Paper>
          </Grid>

          {/* Usage Details Table */}
          <Grid size={12}>
            <Paper sx={{ p: 2 }}>
              <Typography variant="h6" gutterBottom>Usage Details</Typography>
              <DataGrid
                rows={data?.details?.map((d, i) => ({ id: i, ...d })) || []}
                columns={detailsColumns}
                autoHeight
                pageSizeOptions={[10, 25, 50]}
                initialState={{
                  pagination: { paginationModel: { pageSize: 10 } },
                }}
                disableRowSelectionOnClick
              />
            </Paper>
          </Grid>
        </Grid>
      )}
    </Box>
  );
}
