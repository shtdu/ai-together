import { Autocomplete, TextField, Chip } from '@mui/material';

interface UserOption {
  id: number;
  name: string;
  email: string;
}

interface UserSelectProps {
  label: string;
  options: UserOption[];
  value: number[];
  onChange: (value: number[]) => void;
  width?: number | string;
}

export default function UserSelect({ label, options, value, onChange, width = 250 }: UserSelectProps) {
  const selectedOptions = options.filter(opt => value.includes(opt.id));

  return (
    <Autocomplete
      multiple
      size="small"
      options={options}
      value={selectedOptions}
      onChange={(_, newValue) => onChange(newValue.map(v => v.id))}
      getOptionLabel={(option) => `${option.name} (${option.email})`}
      filterOptions={(options, { inputValue }) => {
        const lowerInput = inputValue.toLowerCase();
        return options.filter(
          opt =>
            opt.name.toLowerCase().includes(lowerInput) ||
            opt.email.toLowerCase().includes(lowerInput)
        );
      }}
      renderInput={(params) => (
        <TextField {...params} label={label} placeholder="Search users..." />
      )}
      renderTags={(tagValue, getTagProps) =>
        tagValue.map((option, index) => (
          <Chip
            {...getTagProps({ index })}
            key={option.id}
            label={option.name}
            size="small"
          />
        ))
      }
      sx={{ minWidth: width }}
      isOptionEqualToValue={(option, value) => option.id === value.id}
    />
  );
}
