// Copyright 2019 Northern.tech AS
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
import { HighlightOff as HighlightOffIcon } from '@mui/icons-material';
import { FormHelperText, IconButton, MenuItem, Select, TextField } from '@mui/material';

import { HELPTOOLTIPS } from '@northern.tech/common-ui/helptips/HelpTooltips';
import { MenderHelpTooltip } from '@northern.tech/common-ui/helptips/MenderTooltip';
import { DEVICE_FILTERING_OPTIONS, TIMEOUTS, emptyFilter } from '@northern.tech/store/constants';

import AttributeAutoComplete from './AttributeAutocomplete';

const filterOptionsByPlan = {
  os: { $eq: { title: 'equals' } },
  professional: DEVICE_FILTERING_OPTIONS,
  enterprise: DEVICE_FILTERING_OPTIONS
};

const filterNotifications = {
  name: <MenderHelpTooltip id={HELPTOOLTIPS.nameFilterTip.id} style={{ position: 'absolute', top: 20, left: -28 }} />
};

export const FilterItem = ({ attributes, onChange, onSelect, plan, reset, onSave }) => {
  const [key, setKey] = useState(emptyFilter.key); // this refers to the selected filter with key as the id
  const [value, setValue] = useState(emptyFilter.value); // while this is the value that is applied with the filter
  const [operator, setOperator] = useState(emptyFilter.operator);
  const [scope, setScope] = useState(emptyFilter.scope);
  const timer = useRef();

  useEffect(() => {
    clearTimeout(timer.current);
    setKey(emptyFilter.key);
    setValue(emptyFilter.value);
    setOperator(emptyFilter.operator);
    setScope(emptyFilter.scope);
  }, [attributes.length, reset]);

  useEffect(() => {
    clearTimeout(timer.current);
    onChange({ key, operator, scope, value });
    timer.current = setTimeout(
      () =>
        onSelect({
          key,
          operator,
          scope,
          value
        }),
      TIMEOUTS.debounceDefault
    );
    return () => {
      clearTimeout(timer.current);
    };
  }, [key, onChange, onSelect, operator, scope, value]);

  const updateFilterKey = ({ key, scope }) => {
    setKey(key);
    setScope(scope);
  };

  const updateFilterOperator = ({ target: { value: changedOperator } }) => {
    const newOperator = DEVICE_FILTERING_OPTIONS[changedOperator] || {};
    const opValue = newOperator.value ?? (operator.includes('exists') ? '' : value) ?? '';
    setOperator(changedOperator);
    setValue(opValue);
  };

  const updateFilterValue = ({ target: { value = '' } }) => setValue(value);

  const removeFilter = useCallback(() => {
    setKey(emptyFilter.key);
    setValue(emptyFilter.value);
    setOperator(emptyFilter.operator);
    setScope(emptyFilter.scope);
  }, []);

  const onKeyDown = e => {
    if (e.key !== 'Enter' || ![key, operator, scope, value].every(thing => !!thing)) {
      return;
    }
    e.preventDefault();
    onSave({ key, operator, scope, value });
  };

  const filterOptions = plan ? filterOptionsByPlan[plan] : DEVICE_FILTERING_OPTIONS;
  const operatorHelpMessage = (DEVICE_FILTERING_OPTIONS[operator] || {}).help || '';
  const showValue = typeof (filterOptions[operator] || {}).value === 'undefined';
  return (
    <>
      <div className="flexbox center-aligned margin-top-small margin-bottom-small relative">
        {filterNotifications[key]}
        <AttributeAutoComplete
          attributes={attributes}
          filter={{ key, operator, scope, value }}
          label="Attribute"
          onKeyDown={onKeyDown}
          onRemove={removeFilter}
          onSelect={updateFilterKey}
        />
        <Select className="margin-left-small margin-right-small" onChange={updateFilterOperator} value={operator}>
          {Object.entries(filterOptions).map(([optionKey, option]) => (
            <MenuItem key={optionKey} value={optionKey}>
              {option.title}
            </MenuItem>
          ))}
        </Select>
        {showValue && <TextField label="Value" value={value} onChange={updateFilterValue} onKeyDown={onKeyDown} />}
        {!!key && (
          <IconButton className="margin-left" onClick={removeFilter} size="small">
            <HighlightOffIcon />
          </IconButton>
        )}
      </div>
      {operatorHelpMessage && (
        <div className="margin-bottom-small">
          <FormHelperText>{operatorHelpMessage}</FormHelperText>
        </div>
      )}
    </>
  );
};

export default FilterItem;
