// Copyright 2021 Northern.tech AS
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
import { useEffect, useState } from 'react';
import type { FieldValues, Path } from 'react-hook-form';
import { Controller, useFormContext } from 'react-hook-form';

import type { AutocompleteProps } from '@mui/material';
import { Autocomplete, TextField, useTheme } from '@mui/material';

import { TIMEOUTS } from '@northern.tech/store/constants';
import { useDebounce } from '@northern.tech/utils/debouncehook';

import Loader from '../Loader';

type AsyncAutocompleteProps<T> = {
  inputRef?: Ref<HTMLInputElement>;
  isLoading?: boolean;
  label?: string;
  labelAttribute?: string;
  onChange: (value: T | string | null) => void;
  onSearch: (value: string) => void;
  options?: T[];
  placeholder?: string;
  selectionAttribute?: string;
  value?: T | string | null;
} & Omit<AutocompleteProps<T, false, false, true>, 'freeSolo' | 'onChange' | 'options' | 'renderInput' | 'value'>;

export const AsyncAutocomplete = <T,>({
  inputRef,
  isLoading = false,
  label = '',
  labelAttribute = 'title',
  onChange,
  onSearch,
  options = [],
  placeholder = '',
  selectionAttribute = 'id',
  value,
  ...remainder
}: AsyncAutocompleteProps<T>) => {
  const theme = useTheme();
  const [open, setOpen] = useState(false);
  const [inputValue, setInputValue] = useState(() => {
    if (!value) return '';
    return typeof value === 'string' ? value : (value[labelAttribute] ?? '');
  });
  const loading = open && isLoading;

  const debouncedValue = useDebounce(inputValue, TIMEOUTS.debounceShort);

  useEffect(() => {
    if (!debouncedValue) {
      return;
    }
    onSearch(debouncedValue);
  }, [debouncedValue, onSearch]);

  const onInputChange = (_e: React.SyntheticEvent | null, value: string, reason: string) => {
    if (reason === 'clear') {
      setInputValue('');
    } else if ((reason === 'reset' && !_e) || reason === 'blur') {
      return;
    } else {
      setInputValue(value);
    }
  };

  return (
    <Autocomplete
      autoHighlight
      autoSelect={false}
      freeSolo
      getOptionLabel={option => (typeof option === 'string' ? option : option[labelAttribute])}
      isOptionEqualToValue={(option, val) => {
        if (typeof val === 'string') return typeof option === 'string' ? option === val : option[labelAttribute] === val;
        return option[selectionAttribute] === val[selectionAttribute];
      }}
      inputValue={inputValue}
      value={(value ?? null) as any}
      loading={loading}
      onChange={(_e, data) => {
        onChange(data as T | string | null);
        if (data) {
          setInputValue(typeof data === 'string' ? data : data[labelAttribute]);
        }
      }}
      onClose={() => setOpen(false)}
      onInputChange={onInputChange}
      onOpen={() => setOpen(true)}
      open={open}
      openOnFocus
      options={options}
      renderInput={params => (
        <TextField
          {...params}
          label={label}
          placeholder={placeholder}
          inputRef={inputRef}
          slotProps={{
            ...params.slotProps,
            input: {
              ...params.slotProps.input,
              endAdornment: (
                <>
                  {loading && <Loader show small table style={{ marginTop: theme.spacing(-4) }} />}
                  {params.slotProps.input.endAdornment}
                </>
              )
            }
          }}
        />
      )}
      {...remainder}
    />
  );
};

type ControlledAsyncAutocompleteProps<T, TFieldValues extends FieldValues = FieldValues> = {
  name: Path<TFieldValues>;
} & Omit<AsyncAutocompleteProps<T>, 'onChange' | 'value'>;

export const ControlledAsyncAutocomplete = <T, TFieldValues extends FieldValues = FieldValues>({
  name,
  ...rest
}: ControlledAsyncAutocompleteProps<T, TFieldValues>) => {
  const { control } = useFormContext<TFieldValues>();

  return (
    <Controller
      control={control}
      name={name}
      render={({ field: { onChange, ref, value, ...field } }) => <AsyncAutocomplete {...field} {...rest} inputRef={ref} onChange={onChange} value={value} />}
    />
  );
};
