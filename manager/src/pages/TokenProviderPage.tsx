import { useState, useMemo } from 'react';
import { Box, Typography, Paper, Card, CardContent, Stack, Button, Alert, CircularProgress } from '@mui/material';
import Grid from '@mui/material/Grid2';
import { DataGrid, GridColDef } from '@mui/x-data-grid';
import { useQuery } from '@tanstack/react-query';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell, Legend } from 'recharts';
import RefreshIcon from '@mui/icons-material/Refresh';
import DownloadIcon from '@mui/icons-material/Download';
import SearchIcon from '@mui/icons-material/Search';
import dayjs from 'dayjs';
import DateRangePicker from '../components/common/DateRangePicker';
import MultiSelect from '../components/common/MultiSelect';
import { analyticsApi } from '../api/analytics';

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884D8', '#82CA9D', '#FF6B6B', '#4ECDC4'];

function formatNumber(num: number): string {
  if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`;
  if (num >= 1000) return `${(num / 1000).toFixed(1)}K`;
  return num.toString();
}

function formatCost(cost: number): string {
  return `$${cost.toFixed(2)}`;
}

export default function TokenProviderPage() {
  const [startDate, setStartDate] = useState(dayjs().subtract(7, 'day').format('YYYY-MM-DD'));
  const [endDate, setEndDate] = useState(dayjs().format('YYYY-MM-DD'));
  const [selectedProviders, setSelectedProviders] = useState<string[]>([]);
  const [selectedModels, setSelectedModels] = useState<string[]>([]);

  // Search params (applied when Search is clicked)
  const [searchParams, setSearchParams] = useState({
    startDate: dayjs().subtract(7, 'day').format('YYYY-MM-DD'),
    endDate: dayjs().format('YYYY-MM-DD'),
    providers: [] as string[],
    models: [] as string[],
  });

  const { data: filterOptions, isLoading: filterLoading } = useQuery({
    queryKey: ['filterOptions'],
    queryFn: analyticsApi.getFilterOptions,
  });

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['providerAnalytics', searchParams],
    queryFn: () => analyticsApi.getProviderAnalytics({
      start_date: searchParams.startDate,
      end_date: searchParams.endDate,
      providers: searchParams.providers.length > 0 ? searchParams.providers : undefined,
      models: searchParams.models.length > 0 ? searchParams.models : undefined,
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
      providers: selectedProviders,
      models: selectedModels,
    });
  };

  const handleExportCSV = () => {
    if (!data || !data.distribution.by_provider) return;

    const headers = ['Provider', 'Requests', 'Total Tokens', 'Cost', 'Percentage'];
    const rows = data.distribution.by_provider.map((p: { name: string; requests: number; tokens: number; cost: number; percentage: number }) => [
      p.name,
      p.requests,
      p.tokens,
      p.cost.toFixed(4),
      p.percentage.toFixed(2) + '%',
    ]);

    const csv = [headers.join(','), ...rows.map((r: (string | number)[]) => r.join(','))].join('\n');
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `provider-analytics-${searchParams.startDate}-${searchParams.endDate}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  };

  // Transform trend_data for the chart
  const trendChartData = useMemo(() => {
    if (!data?.trend_data) return [];

    return data.trend_data
      .map((item: { date: string; by_provider: Record<string, number> }) => {
        const totalTokens = Object.values(item.by_provider).reduce((sum: number, val: number) => sum + val, 0);
        return {
          date: item.date,
          tokens: totalTokens,
          ...item.by_provider,
        };
      })
      .sort((a: { date: string }, b: { date: string }) => a.date.localeCompare(b.date));
  }, [data?.trend_data]);

  // Get unique providers from trend data for stacked bar chart
  const trendProviders = useMemo(() => {
    if (!data?.trend_data) return [];
    const providers = new Set<string>();
    data.trend_data.forEach((item: { by_provider: Record<string, number> }) => {
      Object.keys(item.by_provider).forEach((p: string) => providers.add(p));
    });
    return Array.from(providers);
  }, [data?.trend_data]);

  const columns: GridColDef[] = [
    { field: 'name', headerName: 'Provider', flex: 1 },
    { field: 'requests', headerName: 'Requests', type: 'number', width: 120 },
    { field: 'tokens', headerName: 'Total Tokens', type: 'number', width: 140, valueFormatter: (value: number) => formatNumber(value) },
    { field: 'percentage', headerName: 'Share', type: 'number', width: 100, valueFormatter: (value: number) => `${value.toFixed(1)}%` },
    { field: 'cost', headerName: 'Cost', type: 'number', width: 120, valueFormatter: (value: number) => formatCost(value) },
  ];

  if (error) {
    return (
      <Alert severity="error">
        Failed to load provider analytics. Please try again.
      </Alert>
    );
  }

  return (
    <Box>
      <Stack direction="row" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4">Provider Analytics</Typography>
        <Stack direction="row" spacing={1}>
          <Button startIcon={<RefreshIcon />} onClick={() => refetch()}>
            Refresh
          </Button>
          <Button startIcon={<DownloadIcon />} variant="outlined" onClick={handleExportCSV} disabled={!data || !data.distribution.by_provider}>
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
          <MultiSelect
            label="Providers"
            options={filterOptions?.providers || []}
            value={selectedProviders}
            onChange={setSelectedProviders}
          />
          <MultiSelect
            label="Models"
            options={filterOptions?.models || []}
            value={selectedModels}
            onChange={setSelectedModels}
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
        <>
          {/* Summary Cards */}
          {data && (
            <Grid container spacing={3} mb={3}>
              <Grid size={{ xs: 12, sm: 6, md: 3 }}>
                <Card>
                  <CardContent>
                    <Typography color="text.secondary" gutterBottom>Total Requests</Typography>
                    <Typography variant="h4">{formatNumber(data.summary.total_requests)}</Typography>
                  </CardContent>
                </Card>
              </Grid>
              <Grid size={{ xs: 12, sm: 6, md: 3 }}>
                <Card>
                  <CardContent>
                    <Typography color="text.secondary" gutterBottom>Total Tokens</Typography>
                    <Typography variant="h4">{formatNumber(data.summary.total_tokens)}</Typography>
                  </CardContent>
                </Card>
              </Grid>
              <Grid size={{ xs: 12, sm: 6, md: 3 }}>
                <Card>
                  <CardContent>
                    <Typography color="text.secondary" gutterBottom>Input / Output</Typography>
                    <Typography variant="h4">
                      {formatNumber(data.summary.total_input)} / {formatNumber(data.summary.total_output)}
                    </Typography>
                  </CardContent>
                </Card>
              </Grid>
              <Grid size={{ xs: 12, sm: 6, md: 3 }}>
                <Card>
                  <CardContent>
                    <Typography color="text.secondary" gutterBottom>Total Cost</Typography>
                    <Typography variant="h4">{formatCost(data.summary.total_cost)}</Typography>
                  </CardContent>
                </Card>
              </Grid>
            </Grid>
          )}

          {/* Charts */}
          {data && data.distribution.by_provider && data.distribution.by_provider.length > 0 ? (
            <Grid container spacing={3} mb={3}>
              {/* Daily Trend - Stacked Bar Chart by Provider */}
              <Grid size={{ xs: 12, md: 8 }}>
                <Paper sx={{ p: 2, height: 400 }}>
                  <Typography variant="h6" gutterBottom>Daily Token Usage by Provider</Typography>
                  <ResponsiveContainer width="100%" height={330}>
                    <BarChart data={trendChartData}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="date" tickFormatter={(v: string) => dayjs(v).format('MM/DD')} />
                      <YAxis tickFormatter={formatNumber} />
                      <Tooltip
                        formatter={(value: number) => formatNumber(value)}
                        labelFormatter={(label: string) => dayjs(label).format('YYYY-MM-DD')}
                      />
                      <Legend />
                      {trendProviders.map((provider: string, index: number) => (
                        <Bar
                          key={provider}
                          dataKey={provider}
                          stackId="a"
                          fill={COLORS[index % COLORS.length]}
                          name={provider}
                        />
                      ))}
                    </BarChart>
                  </ResponsiveContainer>
                </Paper>
              </Grid>

              {/* Provider Distribution */}
              <Grid size={{ xs: 12, md: 4 }}>
                <Paper sx={{ p: 2, height: 400 }}>
                  <Typography variant="h6" gutterBottom>Provider Distribution</Typography>
                  <ResponsiveContainer width="100%" height={330}>
                    <PieChart>
                      <Pie
                        data={data.distribution.by_provider}
                        dataKey="tokens"
                        nameKey="name"
                        cx="50%"
                        cy="50%"
                        outerRadius={100}
                        label={({ name, percentage }: { name: string; percentage: number }) => `${name} ${percentage.toFixed(0)}%`}
                        labelLine={false}
                      >
                        {data.distribution.by_provider.map((_: unknown, index: number) => (
                          <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                        ))}
                      </Pie>
                      <Tooltip formatter={(value: number) => formatNumber(value)} />
                    </PieChart>
                  </ResponsiveContainer>
                </Paper>
              </Grid>
            </Grid>
          ) : data ? (
            <Alert severity="info" sx={{ mb: 3 }}>
              No usage data available for the selected time period. Start using AI providers to see analytics here.
            </Alert>
          ) : null}

          {/* Data Table */}
          {data && data.distribution.by_provider && data.distribution.by_provider.length > 0 && (
            <Paper sx={{ p: 2 }}>
              <Typography variant="h6" gutterBottom>Provider Details</Typography>
              <DataGrid
                rows={data.distribution.by_provider.map((p: { name: string; requests: number; tokens: number; cost: number; percentage: number }, i: number) => ({ id: i, ...p }))}
                columns={columns}
                autoHeight
                pageSizeOptions={[5, 10, 25]}
                initialState={{
                  pagination: { paginationModel: { pageSize: 10 } },
                }}
                disableRowSelectionOnClick
              />
            </Paper>
          )}
        </>
      )}
    </Box>
  );
}
