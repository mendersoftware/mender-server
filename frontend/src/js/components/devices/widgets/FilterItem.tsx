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
import { Close as CloseIcon } from '@mui/icons-material';
import { FormHelperText, IconButton, MenuItem, Select, TextField } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { DEVICE_FILTERING_OPTIONS, TIMEOUTS, emptyFilter } from '@northern.tech/store/constants';

import { HELPTOOLTIPS } from '../../helptips/HelpTooltips';
import { MenderHelpTooltip } from '../../helptips/MenderTooltip';
import AttributeAutoComplete from './AttributeAutocomplete';

const filterOptionsByPlan = {
  os: { [DEVICE_FILTERING_OPTIONS.$eq.key]: DEVICE_FILTERING_OPTIONS.$eq },
  professional: DEVICE_FILTERING_OPTIONS,
  enterprise: DEVICE_FILTERING_OPTIONS
};

const filterNotifications = {
  name: <MenderHelpTooltip id={HELPTOOLTIPS.nameFilterTip.id} style={{ position: 'absolute', left: -50 }} />
};

const useStyles = makeStyles()(theme => ({
  filter: {
    flexWrap: 'wrap',
    gap: theme.spacing(1.5),
    [theme.breakpoints.down('xl')]: {
      width: '100%'
    }
  },
  filterItem: {
    minWidth: 240
  },
  valueFilter: {
    flexGrow: 1,
    [theme.breakpoints.up('xl')]: {
      flexGrow: 0
    }
  }
}));

export const FilterItem = ({ attributes, onChange, onSelect, plan, reset, onSave }) => {
  const [key, setKey] = useState(emptyFilter.key); // this refers to the selected filter with key as the id
  const [value, setValue] = useState(emptyFilter.value); // while this is the value that is applied with the filter
  const [operator, setOperator] = useState(emptyFilter.operator);
  const [scope, setScope] = useState(emptyFilter.scope);
  const timer = useRef();
  const { classes } = useStyles();

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
  const isFilterDefined = Object.values({ key, operator, scope, ...(showValue ? { value } : {}) }).every(thing => !!thing);

  return (
    <>
      <div className="flexbox center-aligned margin-top-small margin-bottom-small relative">
        {filterNotifications[key]}
        <div className={`flexbox margin-right-small ${classes.filter}`}>
          <AttributeAutoComplete
            attributes={attributes}
            className={classes.filterItem}
            filter={{ key, operator, scope, value }}
            label="Attribute"
            onKeyDown={onKeyDown}
            onRemove={removeFilter}
            onSelect={updateFilterKey}
          />
          <Select className={classes.filterItem} onChange={updateFilterOperator} value={operator}>
            {Object.values(filterOptions).map(option => (
              <MenuItem key={option.key} value={option.key}>
                {option.title}
              </MenuItem>
            ))}
          </Select>
          {showValue && <TextField className={classes.valueFilter} label="Value" value={value} onChange={updateFilterValue} onKeyDown={onKeyDown} />}
        </div>
        {isFilterDefined && (
          <IconButton onClick={removeFilter} size="small">
            <CloseIcon />
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
