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
import type { ComponentType, KeyboardEvent } from 'react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

// material ui
import { CalendarMonth as CalendarIcon, Close as CloseIcon } from '@mui/icons-material';
import { FormHelperText, IconButton, InputAdornment, Popover, TextField } from '@mui/material';
import type { DateView } from '@mui/x-date-pickers';
import { DateCalendar } from '@mui/x-date-pickers';
import { makeStyles } from 'tss-react/mui';

import { Select } from '@northern.tech/common-ui/forms/Select';
import type { FilterOperator } from '@northern.tech/store/constants';
import { ATTRIBUTE_SCOPES, DEVICE_FILTERING_OPTIONS, TIMEOUTS, dateTimeAttributes, emptyFilter } from '@northern.tech/store/constants';
import { tryParseDate } from '@northern.tech/store/utils';
import type { Scope } from '@northern.tech/types/MenderTypes';
import type { Dayjs } from 'dayjs';

import { HELPTOOLTIPS } from '../../helptips/HelpTooltips';
import { MenderHelpTooltip } from '../../helptips/MenderTooltip';
import AttributeAutoComplete from './AttributeAutocomplete';

type FilterOptions = Partial<typeof DEVICE_FILTERING_OPTIONS>;
type ValueOption = { id: string; title: string };

type ValueInputProps = {
  classes: Record<'filterItem' | 'valueFilter', string>;
  onChange: (value: string) => void;
  onKeyDown: (event: KeyboardEvent) => void;
  options: ValueOption[];
  value: string;
};

type NonStandardAttributeFilter = {
  coerceValue: (value: string) => string;
  key: string;
  operators?: FilterOptions;
  options?: ValueOption[];
  scope: Scope;
  ValueInput: ComponentType<ValueInputProps> | null;
};

const TextValueInput = ({ classes, onChange, onKeyDown, value }: Omit<ValueInputProps, 'options'>) => (
  <TextField className={classes.valueFilter} label="Value" value={value} onChange={({ target: { value = '' } }) => onChange(value)} onKeyDown={onKeyDown} />
);

const OptionsValueInput = ({ classes, onChange, onKeyDown, options, value }: ValueInputProps) => (
  <Select
    className={`capitalized ${classes.filterItem}`}
    label="Value"
    MenuItemProps={{ className: 'capitalized' }}
    onChange={({ target: { value = '' } }) => onChange(value)}
    onKeyDown={onKeyDown}
    options={options}
    value={options.some(({ id }) => id === value) ? value : ''}
  />
);

const calendarViews: Record<DateView, string> = {
  year: 'YYYY',
  month: 'YYYY-MM',
  day: 'YYYY-MM-DD'
} as const;

const DateValueInput = ({ classes, onChange, onKeyDown, value }: Omit<ValueInputProps, 'options'>) => {
  const [text, setText] = useState(value);
  const [anchorEl, setAnchorEl] = useState<HTMLElement | null>(null);
  const [calendarView, setCalendarView] = useState<DateView>('day');
  const inputRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    setText(current => (!current || tryParseDate(current) ? value : current));
  }, [value]);

  const isInvalid = !!text && !tryParseDate(text);

  const onTextChange = ({ target: { value: changedText = '' } }) => {
    setText(changedText);
    if (!changedText || tryParseDate(changedText)) {
      onChange(changedText.trim());
    }
  };

  const onCalendarChange = (date: Dayjs | null) => {
    if (!date?.isValid() || calendarView === 'year') {
      return;
    }
    const changedText = date.format(calendarViews[calendarView]);
    setText(changedText);
    onChange(changedText);
    setAnchorEl(null);
  };

  const onOpenCalendar = () => {
    setCalendarView('day');
    setAnchorEl(inputRef.current);
  };

  return (
    <>
      <TextField
        className={classes.valueFilter}
        error={isInvalid}
        label="Value"
        onChange={onTextChange}
        onKeyDown={onKeyDown}
        placeholder={calendarViews.day}
        ref={inputRef}
        slotProps={{
          input: {
            endAdornment: (
              <InputAdornment position="end">
                <IconButton aria-label="Pick a date" edge="end" onClick={onOpenCalendar} size="small">
                  <CalendarIcon fontSize="small" />
                </IconButton>
              </InputAdornment>
            )
          }
        }}
        value={text}
      />
      <Popover anchorEl={anchorEl} anchorOrigin={{ vertical: 'bottom', horizontal: 'left' }} onClose={() => setAnchorEl(null)} open={!!anchorEl}>
        <DateCalendar
          onChange={onCalendarChange}
          onViewChange={setCalendarView}
          value={tryParseDate(text)}
          view={calendarView}
          views={Object.keys(calendarViews) as unknown as DateView[]}
        />
      </Popover>
    </>
  );
};

const testDeviceFilterOptions = {
  [DEVICE_FILTERING_OPTIONS.$eq.key]: DEVICE_FILTERING_OPTIONS.$eq,
  [DEVICE_FILTERING_OPTIONS.$ne.key]: DEVICE_FILTERING_OPTIONS.$ne
};

const getNonStandardAttributeFilters = ({ enabledTiers }: { enabledTiers: string[] }): NonStandardAttributeFilter[] => {
  const filters: NonStandardAttributeFilter[] = [
    {
      key: 'test_device',
      scope: ATTRIBUTE_SCOPES.system,
      operators: testDeviceFilterOptions,
      coerceValue: () => 'true',
      ValueInput: null
    },
    ...dateTimeAttributes.map(({ attribute, scope }) => ({
      key: attribute,
      scope,
      coerceValue: (value: string) => (tryParseDate(value) ? value.trim() : emptyFilter.value),
      ValueInput: DateValueInput
    }))
  ];
  if (enabledTiers.length) {
    filters.push({
      key: 'tier',
      scope: ATTRIBUTE_SCOPES.system,
      operators: testDeviceFilterOptions,
      options: enabledTiers.map(tier => ({ id: tier, title: tier })),
      coerceValue: value => (enabledTiers.includes(value) ? value : emptyFilter.value),
      ValueInput: OptionsValueInput
    });
  }
  return filters;
};

const findNonStandardAttributeFilter = (filters: NonStandardAttributeFilter[], { key, scope }) =>
  filters.find(filter => filter.key === key && filter.scope === scope);

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

export const FilterItem = ({ attributes, enabledTiers, onChange, onSelect, reset, onSave }) => {
  const [key, setKey] = useState(emptyFilter.key); // this refers to the selected filter with key as the id
  const [value, setValue] = useState(emptyFilter.value); // while this is the value that is applied with the filter
  const [operator, setOperator] = useState(emptyFilter.operator);
  const [scope, setScope] = useState(emptyFilter.scope);
  const timer = useRef();
  const { classes } = useStyles();

  const nonStandardAttributeFilters = useMemo(() => getNonStandardAttributeFilters({ enabledTiers }), [enabledTiers]);
  const specialAttributeFilter = findNonStandardAttributeFilter(nonStandardAttributeFilters, { key, scope });
  const { operators: filterOptions = DEVICE_FILTERING_OPTIONS, options: valueOptions = [], ValueInput = TextValueInput } = specialAttributeFilter ?? {};

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

  const updateFilterKey = ({ key: nextKey, scope: nextScope }) => {
    setKey(nextKey);
    setScope(nextScope);
    if (specialAttributeFilter) {
      // non standard attributes might carry special values, so we need to clear them
      setValue(emptyFilter.value);
    }
    const nextSpecialFilter = findNonStandardAttributeFilter(nonStandardAttributeFilters, { key: nextKey, scope: nextScope });
    if (nextSpecialFilter) {
      const { coerceValue, operators = DEVICE_FILTERING_OPTIONS } = nextSpecialFilter;
      const nextOperator = Object.hasOwn(operators, operator) ? operator : (Object.keys(operators)[0] as FilterOperator);
      setOperator(nextOperator);
      setValue(coerceValue(value));
    }
  };

  const updateFilterOperator = ({ target: { value: changedOperator } }) => {
    const newOperator = DEVICE_FILTERING_OPTIONS[changedOperator] || {};
    const opValue = newOperator.value ?? (operator.includes('exists') ? '' : value) ?? '';
    setOperator(changedOperator);
    setValue(specialAttributeFilter ? specialAttributeFilter.coerceValue(opValue) : opValue);
  };

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

  const operatorHelpMessage = (DEVICE_FILTERING_OPTIONS[operator] || {}).help || '';
  const showValue = typeof (filterOptions[operator] || {}).value === 'undefined';
  const isFilterDefined = Object.values({ key, operator, scope, ...(showValue ? { value } : {}) }).every(thing => !!thing);

  return (
    <>
      <div className="flexbox align-items-center margin-top-small margin-bottom-small relative">
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
          <Select
            className={classes.filterItem}
            onChange={updateFilterOperator}
            options={Object.values(filterOptions)}
            selectionAttribute="key"
            value={operator}
          />
          {showValue && ValueInput && <ValueInput classes={classes} onChange={setValue} onKeyDown={onKeyDown} options={valueOptions} value={value} />}
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
