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
import { useCallback, useEffect, useRef, useState } from 'react';

// material ui
import { Autocomplete, TextField, createFilterOptions } from '@mui/material';

import { TIMEOUTS, emptyFilter } from '@northern.tech/store/constants';

import { defaultHeaders } from '../BaseDevices';
import { getFilterLabelByKey } from './Filters';

export const getOptionLabel = option => {
  const header = Object.values(defaultHeaders).find(
    ({ attribute }) => attribute.scope === option.scope && (attribute.name === option.key || attribute.alternative === option.key)
  );
  return header?.title || option.title || option.value || option.key || option;
};

const FilterOption = ({ key, ...props }, option) => {
  let content = getOptionLabel(option);
  if (option.category === 'recently used') {
    content = (
      <div className="flexbox align-items-center space-between full-width">
        <div>{content}</div>
        <div className="muted slightly-smaller">({option.scope})</div>
      </div>
    );
  }
  return (
    <li key={key} {...props}>
      {content}
    </li>
  );
};

const optionsFilter = createFilterOptions();

const filterOptions = (options, params) => {
  const filtered = optionsFilter(options, params);
  if (filtered.length !== 1 && params.inputValue !== '') {
    filtered.push({
      inputValue: params.inputValue,
      key: 'custom',
      value: `Use "${params.inputValue}"`,
      category: 'custom',
      priority: 99
    });
  }
  return filtered;
};

export const AttributeAutoComplete = ({ attributes, disabled = false, filter = emptyFilter, label = 'Attribute', onRemove, onSelect, ...remainder }) => {
  const [key, setKey] = useState(filter.key); // this refers to the selected filter with key as the id
  const [options, setOptions] = useState([]);
  const [reset, setReset] = useState(true);
  const [scope, setScope] = useState(filter.scope);
  const timer = useRef();

  useEffect(
    () => () => {
      clearTimeout(timer.current);
    },
    []
  );

  useEffect(() => {
    setKey(emptyFilter.key);
    setScope(emptyFilter.scope);
    const attributesClean = attributes.map(attr => {
      if (!attr.category) {
        attr.category = attr.scope || '';
      }
      return attr;
    });
    const categoryPriorities = attributesClean.reduce((accu, { category, priority }) => {
      if (!accu[category]) {
        accu[category] = priority;
      }
      return accu;
    }, {});
    setOptions(
      attributesClean.sort((a, b) => {
        const aCategoryPriority = categoryPriorities[a.category] ?? Infinity;
        const bCategoryPriority = categoryPriorities[b.category] ?? Infinity;
        return aCategoryPriority !== bCategoryPriority
          ? aCategoryPriority - bCategoryPriority
          : a.category !== b.category
            ? a.category.localeCompare(b.category)
            : (a.key || '').localeCompare(b.key || '');
      })
    );
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [attributes.length, reset]);

  useEffect(() => {
    setKey(filter.key);
  }, [filter.key]);

  useEffect(() => {
    clearTimeout(timer.current);
    timer.current = setTimeout(() => onSelect({ key, scope }), TIMEOUTS.debounceDefault);
    return () => {
      clearTimeout(timer.current);
    };
  }, [key, onSelect, scope]);

  const updateFilterKey = (value, selectedScope) => {
    if (!value) {
      return removeFilter();
    }
    const { key = value, scope: fallbackScope } = attributes.find(filter => filter.key === value) ?? {};
    setKey(key);
    setScope(selectedScope || fallbackScope);
  };

  const removeFilter = useCallback(() => {
    if (key) {
      onRemove({ key, scope });
    }
    setReset(!reset);
  }, [key, onRemove, reset, setReset, scope]);

  return (
    <Autocomplete
      {...remainder}
      autoComplete
      autoHighlight
      autoSelect
      disabled={disabled}
      freeSolo
      filterSelectedOptions
      filterOptions={filterOptions}
      getOptionLabel={getOptionLabel}
      groupBy={option => option.category}
      renderOption={FilterOption}
      id="filter-selection"
      includeInputInList={true}
      onChange={(e, changedValue) => {
        const { inputValue, key = changedValue, scope } = changedValue || {};
        if (inputValue) {
          // only circumvent updateFilterKey if we deal with a custom attribute - those will be treated as inventory attributes
          setKey(inputValue);
          return setScope(emptyFilter.scope);
        }
        updateFilterKey(key, scope);
      }}
      options={options}
      renderInput={params => <TextField {...params} label={label} />}
      key={reset}
      value={getFilterLabelByKey(key, attributes)}
    />
  );
};

export default AttributeAutoComplete;
