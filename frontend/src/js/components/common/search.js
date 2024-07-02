// Copyright 2022 Northern.tech AS
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
import React, { useCallback, useEffect, useRef } from 'react';
import { Controller, FormProvider, useForm, useFormContext } from 'react-hook-form';

import { Search as SearchIcon } from '@mui/icons-material';
import { InputAdornment, TextField } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { TIMEOUTS } from '../../constants/appConstants';
import { useDebounce } from '../../utils/debouncehook';
import Loader from './loader';

const useStyles = makeStyles()(() => ({
  root: {
    input: {
      fontSize: '13px'
    }
  }
}));

const endAdornment = (
  <InputAdornment position="end">
    <Loader show small style={{ marginTop: -10 }} />
  </InputAdornment>
);

const startAdornment = (
  <InputAdornment position="start">
    <SearchIcon color="disabled" fontSize="small" />
  </InputAdornment>
);

// due to search not working reliably for single letter searches, only start at 2
const MINIMUM_SEARCH_LENGTH = 2;

export const ControlledSearch = ({ isSearching, name = 'search', onSearch, placeholder = 'Search devices', style = {} }) => {
  const { classes } = useStyles();
  const { control, watch } = useFormContext();
  const inputRef = useRef();

  const searchValue = watch('search', '');

  const debouncedSearchTerm = useDebounce(searchValue, TIMEOUTS.debounceDefault);

  useEffect(() => {
    if (debouncedSearchTerm.length < MINIMUM_SEARCH_LENGTH) {
      return;
    }
    onSearch(debouncedSearchTerm).then(() => inputRef.current.focus());
  }, [debouncedSearchTerm, onSearch]);

  const onTriggerSearch = useCallback(
    ({ key }) => {
      if (key === 'Enter' && (!debouncedSearchTerm || debouncedSearchTerm.length >= MINIMUM_SEARCH_LENGTH)) {
        onSearch(debouncedSearchTerm).then(() => inputRef.current.focus());
      }
    },
    [debouncedSearchTerm, onSearch]
  );

  const adornments = isSearching ? { startAdornment, endAdornment } : { startAdornment };
  return (
    <Controller
      name={name}
      control={control}
      render={({ field }) => (
        <TextField
          className={classes.root}
          InputProps={adornments}
          onKeyPress={onTriggerSearch}
          placeholder={placeholder}
          inputRef={inputRef}
          size="small"
          style={style}
          {...field}
        />
      )}
    />
  );
};

ControlledSearch.displayName = 'ConnectedSearch';

const Search = props => {
  const { searchTerm, onSearch, trigger } = props;
  const methods = useForm({ mode: 'onChange', defaultValues: { search: searchTerm ?? '' } });
  const { handleSubmit } = methods;
  return (
    <FormProvider {...methods}>
      <form noValidate onSubmit={handleSubmit(({ search }) => onSearch(search, !trigger))}>
        <ControlledSearch {...props} />
        <input className="hidden" type="submit" />
      </form>
    </FormProvider>
  );
};

export default Search;
