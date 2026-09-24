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
import { useId } from 'react';

import type { MenuItemProps as MuiMenuItemProps, MenuProps as MuiMenuProps, SelectProps as MuiSelectProps } from '@mui/material';
import { Checkbox, FormControl, FormHelperText, InputLabel, MenuItem, Select as MuiSelect } from '@mui/material';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type SelectOption = Record<string, any>;

export interface SelectProps<T extends SelectOption = SelectOption, Value = unknown> extends Omit<MuiSelectProps<Value>, 'placeholder'> {
  getOptionDisabled?: (option: T) => boolean;
  helperText?: ReactNode;
  hideEmptyOption?: boolean;
  labelAttribute?: string;
  MenuItemProps?: Partial<MuiMenuItemProps>;
  options?: T[];
  placeholder?: ReactNode;
  renderOption?: (option: T) => ReactNode;
  selectionAttribute?: string;
  width?: number | string;
}

// open the menu below the select, aligned with its left edge
const defaultMenuProps: Partial<MuiMenuProps> = {
  anchorOrigin: { vertical: 'bottom', horizontal: 'left' },
  transformOrigin: { vertical: 'top', horizontal: 'left' }
};

export const Select = <T extends SelectOption = SelectOption, Value = unknown>({
  autoWidth = false,
  children,
  disabled,
  displayEmpty,
  error,
  getOptionDisabled,
  helperText,
  hideEmptyOption = false,
  label,
  labelAttribute = 'title',
  labelId: labelIdProp,
  MenuItemProps,
  MenuProps,
  multiple = false,
  options = [],
  placeholder,
  renderOption,
  renderValue,
  selectionAttribute = 'id',
  value,
  width,
  ...remainder
}: SelectProps<T, Value>) => {
  const generatedId = useId();
  const labelId = labelIdProp ?? `${generatedId}-label`;
  const hasPlaceholder = !!placeholder;
  const selectValue = (value ?? (multiple ? [] : '')) as Value | '';
  const selectedValues: unknown[] = Array.isArray(selectValue) ? selectValue : [selectValue];
  const selectedOptions = options.filter(option => selectedValues.includes(option[selectionAttribute]));

  const renderSelection = () => {
    if (!selectedOptions.length) {
      return <span className="muted">{placeholder}</span>;
    }
    return multiple ? selectedOptions.map(option => option[labelAttribute]).join(', ') : selectedOptions[0][labelAttribute];
  };

  const selectionRenderer = renderValue ?? (!children && (hasPlaceholder || multiple) ? renderSelection : undefined);

  return (
    <FormControl disabled={disabled} error={error} style={{ width }}>
      {!!label && <InputLabel id={labelId}>{label}</InputLabel>}
      <MuiSelect
        autoWidth={autoWidth}
        displayEmpty={displayEmpty ?? hasPlaceholder}
        label={label}
        labelId={label ? labelId : labelIdProp}
        MenuProps={{ ...defaultMenuProps, ...MenuProps }}
        multiple={multiple}
        renderValue={selectionRenderer}
        value={selectValue}
        {...remainder}
      >
        {children}
        {!children && hasPlaceholder && !hideEmptyOption && !multiple && (
          <MenuItem {...MenuItemProps} value="">
            <span className="muted">{placeholder}</span>
          </MenuItem>
        )}
        {!children &&
          options.map(option => (
            <MenuItem {...MenuItemProps} disabled={getOptionDisabled?.(option)} key={option[selectionAttribute]} value={option[selectionAttribute]}>
              {multiple && <Checkbox checked={selectedValues.includes(option[selectionAttribute])} />}
              {renderOption ? renderOption(option) : option[labelAttribute]}
            </MenuItem>
          ))}
      </MuiSelect>
      {!!helperText && <FormHelperText>{helperText}</FormHelperText>}
    </FormControl>
  );
};

export default Select;
