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

import type { SelectOption, SelectProps } from './Select';
import { Select } from './Select';

interface ControlledSelectProps<T extends SelectOption = SelectOption> extends Omit<SelectProps<T>, 'onChange' | 'value'> {
  name: string;
  placeholder?: string;
}

export const ControlledSelect = <T extends SelectOption = SelectOption>({ name, placeholder = '', width = 240, ...remainder }: ControlledSelectProps<T>) => {
  const { control } = useFormContext();
  return (
    <Controller
      control={control}
      name={name}
      render={({ field: { value, onChange } }) => (
        <Select
          MenuItemProps={{ dense: false }}
          name={name}
          placeholder={placeholder}
          slotProps={{ input: { 'aria-label': placeholder } }}
          value={value}
          width={width}
          onChange={({ target: { value } }) => onChange(value)}
          {...remainder}
        />
      )}
    />
  );
};
