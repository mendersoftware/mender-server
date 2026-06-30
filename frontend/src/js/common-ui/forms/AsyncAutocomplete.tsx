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
import { useEffect, useState } from 'react';
import { useController, useFormContext } from 'react-hook-form';

import { Autocomplete, TextField, useTheme } from '@mui/material';

import { TIMEOUTS } from '@northern.tech/store/constants';
import { useDebounce } from '@northern.tech/utils/debouncehook';

import Loader from '../Loader';

export const AsyncAutocomplete = ({
  label = '',
  labelAttribute = 'title',
  name,
  isLoading = false,
  onSearch,
  options = [],
  placeholder = '',
  selectionAttribute = 'id',
  ...remainder
}) => {
  const theme = useTheme();
  const { control } = useFormContext();
  const {
    field: { onChange: formOnChange, ref, value, ...field }
  } = useController({ control, name });

  const [open, setOpen] = useState(false);
  const [inputValue, setInputValue] = useState(() => {
    if (!value) return '';
    return typeof value === 'string' ? value : value[labelAttribute] ?? '';
  });
  const loading = open && isLoading;

  const debouncedValue = useDebounce(inputValue, TIMEOUTS.debounceShort);

  useEffect(() => {
    if (!debouncedValue) {
      return;
    }
    onSearch(debouncedValue);
  }, [debouncedValue, onSearch]);

  const onInputChange = (_e, value, reason) => {
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
      {...field}
      inputValue={inputValue}
      value={value ?? null}
      loading={loading}
      onChange={(_e, data) => {
        formOnChange(data);
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
          inputRef={ref}
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

export default AsyncAutocomplete;
