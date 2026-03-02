import { useState } from 'react';
import { Box, Typography, Paper, Stack, Button, Alert } from '@mui/material';
import { DataGrid, GridColDef, GridPaginationModel, GridSortModel } from '@mui/x-data-grid';
import { useQuery } from '@tanstack/react-query';
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

function formatDuration(ms: number): string {
  if (ms >= 1000) return `${(ms / 1000).toFixed(2)}s`;
  return `${ms}ms`;
}

export default function TokenHistoryPage() {
  const [startDate, setStartDate] = useState(dayjs().subtract(7, 'day').format('YYYY-MM-DD'));
  const [endDate, setEndDate] = useState(dayjs().format('YYYY-MM-DD'));
  const [selectedProviders, setSelectedProviders] = useState<string[]>([]);
  const [selectedModels, setSelectedModels] = useState<string[]>([]);
  const [selectedUsers, setSelectedUsers] = useState<number[]>([]);
  const [paginationModel, setPaginationModel] = useState<GridPaginationModel>({ page: 0, pageSize: 25 });
  const [sortModel, setSortModel] = useState<GridSortModel>([{ field: 'created_at', sort: 'desc' }]);

  // Search params (applied when Search is clicked)
  const [searchParams, setSearchParams] = useState({
    startDate: dayjs().subtract(7, 'day').format('YYYY-MM-DD'),
    endDate: dayjs().format('YYYY-MM-DD'),
    providers: [] as string[],
    models: [] as string[],
    userIds: [] as number[],
  });

  const { data: filterOptions, isLoading: filterLoading } = useQuery({
    queryKey: ['filterOptions'],
    queryFn: analyticsApi.getFilterOptions,
  });

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['history', searchParams, paginationModel, sortModel],
    queryFn: () => analyticsApi.getHistory({
      start_date: searchParams.startDate,
      end_date: searchParams.endDate,
      page: paginationModel.page + 1,
      limit: paginationModel.pageSize,
      providers: searchParams.providers.length > 0 ? searchParams.providers : undefined,
      models: searchParams.models.length > 0 ? searchParams.models : undefined,
      user_ids: searchParams.userIds.length > 0 ? searchParams.userIds : undefined,
      sort_by: sortModel[0]?.field || 'created_at',
      sort_order: sortModel[0]?.sort || 'desc',
    }),
  });

  const handleDateChange = (start: string, end: string) => {
    setStartDate(start);
    setEndDate(end);
  };

  const handleSearch = () => {
    setPaginationModel({ ...paginationModel, page: 0 });
    setSearchParams({
      startDate,
      endDate,
      providers: selectedProviders,
      models: selectedModels,
      userIds: selectedUsers,
    });
  };

  const handleExportCSV = () => {
    if (!data) return;

    const headers = ['ID', 'User', 'Provider', 'Model', 'Input Tokens', 'Output Tokens', 'Duration', 'Created At'];
    const rows = data.records.map(r => [
      r.id,
      r.user_name,
      r.provider,
      r.model,
      r.input_tokens,
      r.output_tokens,
      r.duration_ms,
      r.created_at,
    ]);

    const csv = [headers.join(','), ...rows.map(r => r.join(','))].join('\n');
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `token-history-${startDate}-${endDate}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const columns: GridColDef[] = [
    { field: 'id', headerName: 'ID', width: 80 },
    { field: 'user_name', headerName: 'User', flex: 1 },
    { field: 'provider', headerName: 'Provider', width: 120 },
    { field: 'model', headerName: 'Model', width: 200 },
    { field: 'input_tokens', headerName: 'Input', type: 'number', width: 100, valueFormatter: (value: number) => formatNumber(value) },
    { field: 'output_tokens', headerName: 'Output', type: 'number', width: 100, valueFormatter: (value: number) => formatNumber(value) },
    { field: 'duration_ms', headerName: 'Duration', type: 'number', width: 100, valueFormatter: (value: number) => formatDuration(value) },
    {
      field: 'created_at',
      headerName: 'Created At',
      width: 180,
      valueFormatter: (value: string) => dayjs(value).format('YYYY-MM-DD HH:mm:ss'),
    },
  ];

  if (error) {
    return (
      <Alert severity="error">
        Failed to load history. Please try again.
      </Alert>
    );
  }

  return (
    <Box>
      <Stack direction="row" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4">Token History</Typography>
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

      {/* Data Table */}
      <Paper sx={{ p: 2 }}>
        <DataGrid
          rows={data?.records || []}
          columns={columns}
          loading={isLoading}
          rowCount={data?.pagination?.total || 0}
          paginationMode="server"
          paginationModel={paginationModel}
          onPaginationModelChange={setPaginationModel}
          sortingMode="server"
          sortModel={sortModel}
          onSortModelChange={setSortModel}
          pageSizeOptions={[10, 25, 50, 100]}
          autoHeight
          disableRowSelectionOnClick
          sx={{ minHeight: 400 }}
        />
      </Paper>
    </Box>
  );
}
