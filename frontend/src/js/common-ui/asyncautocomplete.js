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
import React, { useEffect, useState } from 'react';

import { Autocomplete, TextField } from '@mui/material';
import { useTheme } from '@mui/material/styles';

import { TIMEOUTS } from '@northern.tech/store/constants';
import { useDebounce } from '@northern.tech/utils/debouncehook';

import Loader from './loader';

export const AsyncAutocomplete = ({
  id,
  initialValue,
  isLoading,
  label,
  placeholder,
  styles,
  selectionAttribute,
  labelAttribute,
  onChange,
  onChangeSelection,
  options
}) => {
  const theme = useTheme();
  const [open, setOpen] = useState(false);
  const [inputValue, setInputValue] = useState(initialValue);
  const loading = open && isLoading;

  const debouncedValue = useDebounce(inputValue, TIMEOUTS.debounceShort);

  useEffect(() => {
    if (debouncedValue === undefined) {
      return;
    }
    const selection = options.find(option => option[selectionAttribute] === debouncedValue);
    if (selection) {
      onChangeSelection(selection);
    } else {
      onChange(debouncedValue);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [debouncedValue, JSON.stringify(options), onChange, onChangeSelection, selectionAttribute]);

  const onInputChange = (e, value, reason) => {
    if (reason === 'clear') {
      setInputValue('');
      return onChangeSelection();
    } else if ((reason === 'reset' && !e) || reason === 'blur') {
      return;
    }
    setInputValue(value);
  };

  return (
    <Autocomplete
      autoHighlight
      freeSolo
      getOptionLabel={option => option[labelAttribute]}
      isOptionEqualToValue={(option, value) => option[selectionAttribute] === value[selectionAttribute]}
      id={id}
      inputValue={inputValue || ''}
      loading={loading}
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
          style={styles.textField}
          InputProps={{
            ...params.InputProps,
            endAdornment: (
              <>
                {loading && <Loader show small table style={{ marginTop: theme.spacing(-4) }} />}
                {params.InputProps.endAdornment}
              </>
            )
          }}
        />
      )}
    />
  );
};

export default AsyncAutocomplete;
