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
import { Controller, useFormContext } from 'react-hook-form';

import { MenuItem, Select } from '@mui/material';

export const ControlledSelect = ({ name, options = [], placeholder = '', selectionAttribute = 'id', labelAttribute = 'title', width = 240, ...remainder }) => {
  const { control } = useFormContext();
  return (
    <Controller
      control={control}
      name={name}
      render={({ field: { value, onChange } }) => (
        <Select
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
            transformOrigin: { vertical: 'top', horizontal: 'left' }
          }}
          {...remainder}
        >
          {placeholder && (
            <MenuItem dense={false} value="">
              <span className="muted">{placeholder}</span>
            </MenuItem>
          )}
          {options.map(option => (
            <MenuItem dense={false} key={option[selectionAttribute]} value={option[selectionAttribute]}>
              {option[labelAttribute]}
            </MenuItem>
          ))}
        </Select>
      )}
    />
  );
};
