import { FormControl, InputLabel, Select, MenuItem, Checkbox, ListItemText, OutlinedInput, Chip, Box } from '@mui/material';

interface MultiSelectProps {
  label: string;
  options: string[];
  value: string[];
  onChange: (value: string[]) => void;
  width?: number | string;
}

export default function MultiSelect({ label, options, value, onChange, width = 200 }: MultiSelectProps) {
  return (
    <FormControl size="small" sx={{ minWidth: width }}>
      <InputLabel>{label}</InputLabel>
      <Select
        multiple
        value={value}
        onChange={(e) => onChange(e.target.value as string[])}
        input={<OutlinedInput label={label} />}
        renderValue={(selected) => (
          <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
            {selected.map((val) => (
              <Chip key={val} label={val} size="small" />
            ))}
          </Box>
        )}
        MenuProps={{
          PaperProps: {
            style: {
              maxHeight: 300,
            },
          },
        }}
      >
        {options.map((option) => (
          <MenuItem key={option} value={option}>
            <Checkbox checked={value.indexOf(option) > -1} size="small" />
            <ListItemText primary={option} />
          </MenuItem>
        ))}
      </Select>
    </FormControl>
  );
}
