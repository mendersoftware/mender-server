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
import type { Ref } from 'react';
import { useMemo } from 'react';
import type { FieldValues, Path } from 'react-hook-form';
import { Controller, useFormContext } from 'react-hook-form';

import type { AutocompleteProps } from '@mui/material';
import { Autocomplete, Checkbox, Chip, TextField } from '@mui/material';

import { TruncatedTagList } from './helpers';

const listboxMaxHeight = 304;

type CheckboxAutocompleteProps<T> = {
  chipDisplay?: boolean;
  inputRef?: Ref<HTMLInputElement>;
  label?: string;
  labelAttribute?: string;
  onChange: (value: T[]) => void;
  options?: T[];
  placeholder?: string;
  value?: T[];
} & Omit<AutocompleteProps<T, true, false, false>, 'multiple' | 'onChange' | 'options' | 'renderInput' | 'value'>;

export const CheckboxAutocomplete = <T,>({
  chipDisplay = false,
  inputRef,
  label = '',
  labelAttribute = 'title',
  onChange,
  options = [],
  placeholder = '',
  value = [],
  ...remainder
}: CheckboxAutocompleteProps<T>) => {
  const sortedOptions = useMemo(() => {
    const selectedSet = new Set(value);
    return [...options].sort((a, b) => {
      const aSelected = selectedSet.has(a);
      const bSelected = selectedSet.has(b);
      if (aSelected && !bSelected) return -1;
      if (!aSelected && bSelected) return 1;
      return 0;
    });
  }, [value, options]);

  return (
    <Autocomplete
      autoSelect={false}
      disableCloseOnSelect
      multiple
      value={value ?? []}
      onChange={(_e, data) => onChange(data)}
      options={sortedOptions}
      getOptionLabel={option => (typeof option === 'string' ? option : option[labelAttribute])}
      isOptionEqualToValue={(option, val) => option === val || (option[labelAttribute] != null && option[labelAttribute] === val[labelAttribute])}
      renderOption={({ key, ...optionProps }, option, { selected }) => (
        <li key={key} {...optionProps}>
          <Checkbox className="margin-right-x-small" checked={selected} />
          {typeof option === 'string' ? option : option[labelAttribute]}
        </li>
      )}
      renderValue={
        chipDisplay
          ? (values, getItemProps) =>
              values.map((option, index) => {
                const { key, ...tagProps } = getItemProps({ index });
                return <Chip key={key} label={typeof option === 'string' ? option : option[labelAttribute]} size="small" {...tagProps} />;
              })
          : values => <TruncatedTagList labelAttribute={labelAttribute} values={values} />
      }
      renderInput={params => <TextField {...params} label={label} placeholder={value?.length ? '' : placeholder} inputRef={inputRef} />}
      slotProps={{ listbox: { style: { maxHeight: listboxMaxHeight } } }}
      {...remainder}
    />
  );
};

type ControlledCheckboxAutocompleteProps<T, TFieldValues extends FieldValues = FieldValues> = {
  name: Path<TFieldValues>;
} & Omit<CheckboxAutocompleteProps<T>, 'onChange' | 'value'>;

export const ControlledCheckboxAutocomplete = <T, TFieldValues extends FieldValues = FieldValues>({
  name,
  ...rest
}: ControlledCheckboxAutocompleteProps<T, TFieldValues>) => {
  const { control } = useFormContext<TFieldValues>();

  return (
    <Controller
      control={control}
      name={name}
      render={({ field: { onChange, ref, value, ...field } }) => (
        <CheckboxAutocomplete {...field} {...rest} inputRef={ref} onChange={onChange} value={value ?? []} />
      )}
    />
  );
};
