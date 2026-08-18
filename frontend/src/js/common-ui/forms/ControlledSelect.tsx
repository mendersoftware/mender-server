// Copyright 2026 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.
import type { ReactNode } from 'react';
import { Controller, useFormContext } from 'react-hook-form';

import type { SelectProps } from '@mui/material';
import { MenuItem, Select } from '@mui/material';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type SelectOption = Record<string, any>;

interface ControlledSelectProps<T extends SelectOption = SelectOption> extends Omit<SelectProps, 'onChange' | 'renderValue' | 'value'> {
  getOptionDisabled?: (option: T) => boolean;
  hideEmptyOption?: boolean;
  labelAttribute?: string;
  options?: T[];
  placeholder?: string;
  renderOption?: (option: T) => ReactNode;
  selectionAttribute?: string;
  width?: number | string;
}

export const ControlledSelect = <T extends SelectOption = SelectOption>({
  name,
  options = [],
  placeholder = '',
  hideEmptyOption = false,
  selectionAttribute = 'id',
  labelAttribute = 'title',
  getOptionDisabled,
  renderOption,
  width = 240,
  className,
  ...remainder
}: ControlledSelectProps<T>) => {
  const { control } = useFormContext();
  return (
    <Controller
      control={control}
      name={name}
      render={({ field: { value, onChange } }) => (
        <Select
          autoWidth={false}
          displayEmpty
          style={{ width }}
          value={value ?? ''}
          onChange={({ target: { value } }) => onChange(value)}
          renderValue={selected => {
            const selectedOption = options.find(option => option[selectionAttribute] === selected);
            return selectedOption ? selectedOption[labelAttribute] : <span className="muted">{placeholder}</span>;
          }}
          MenuProps={{
            anchorOrigin: { vertical: 'bottom', horizontal: 'left' },
            transformOrigin: { vertical: 'top', horizontal: 'left' },
            slotProps: { paper: { className } }
          }}
          slotProps={{ input: { 'aria-label': placeholder } }}
          {...remainder}
        >
          {!!placeholder && !hideEmptyOption && (
            <MenuItem dense={false} value="">
              <span className="muted">{placeholder}</span>
            </MenuItem>
          )}
          {options.map(option => (
            <MenuItem
              dense={false}
              key={option[selectionAttribute]}
              value={option[selectionAttribute]}
              disabled={getOptionDisabled ? getOptionDisabled(option) : undefined}
            >
              {renderOption ? renderOption(option) : option[labelAttribute]}
            </MenuItem>
          ))}
        </Select>
      )}
    />
  );
};
