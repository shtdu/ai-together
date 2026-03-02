import { useState } from 'react';
import { Box, TextField, Button, Stack, Popover, List, ListItemButton, ListItemText } from '@mui/material';
import dayjs from 'dayjs';

interface DateRangePickerProps {
  startDate: string;
  endDate: string;
  onChange: (startDate: string, endDate: string) => void;
}

const presets = [
  { label: 'Today', days: 0 },
  { label: 'Last 7 days', days: 7 },
  { label: 'Last 14 days', days: 14 },
  { label: 'Last 30 days', days: 30 },
  { label: 'Last 90 days', days: 90 },
];

export default function DateRangePicker({ startDate, endDate, onChange }: DateRangePickerProps) {
  const [anchorEl, setAnchorEl] = useState<HTMLElement | null>(null);

  const handlePresetClick = (days: number) => {
    const end = dayjs().format('YYYY-MM-DD');
    const start = dayjs().subtract(days, 'day').format('YYYY-MM-DD');
    onChange(start, end);
    setAnchorEl(null);
  };

  return (
    <Stack direction="row" spacing={2} alignItems="center">
      <TextField
        label="Start Date"
        type="date"
        value={startDate}
        onChange={(e) => onChange(e.target.value, endDate)}
        size="small"
        InputLabelProps={{ shrink: true }}
      />
      <TextField
        label="End Date"
        type="date"
        value={endDate}
        onChange={(e) => onChange(startDate, e.target.value)}
        size="small"
        InputLabelProps={{ shrink: true }}
      />
      <Box>
        <Button
          variant="outlined"
          size="small"
          onClick={(e) => setAnchorEl(e.currentTarget)}
        >
          Presets
        </Button>
        <Popover
          open={Boolean(anchorEl)}
          anchorEl={anchorEl}
          onClose={() => setAnchorEl(null)}
          anchorOrigin={{ vertical: 'bottom', horizontal: 'left' }}
        >
          <List dense>
            {presets.map((preset) => (
              <ListItemButton
                key={preset.label}
                onClick={() => handlePresetClick(preset.days)}
              >
                <ListItemText primary={preset.label} />
              </ListItemButton>
            ))}
          </List>
        </Popover>
      </Box>
    </Stack>
  );
}
