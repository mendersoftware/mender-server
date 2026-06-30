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
import { useMemo } from 'react';
import { useController, useFormContext } from 'react-hook-form';

import { Autocomplete, Checkbox, Chip, TextField } from '@mui/material';

import { TruncatedTagList } from './helpers';

const listboxMaxHeight = 304;

export const CheckboxAutocomplete = ({
  chipDisplay = false,
  label = '',
  labelAttribute = 'title',
  name,
  options = [],
  placeholder = '',
  ...remainder
}) => {
  const { control } = useFormContext();
  const {
    field: { onChange: formOnChange, ref, value = [], ...field }
  } = useController({ control, name });

  const selectedValues = value ?? [];

  const sortedOptions = useMemo(() => {
    const selectedSet = new Set(selectedValues);
    return [...options].sort((a, b) => {
      const aSelected = selectedSet.has(a);
      const bSelected = selectedSet.has(b);
      if (aSelected && !bSelected) return -1;
      if (!aSelected && bSelected) return 1;
      return 0;
    });
  }, [selectedValues, options]);

  return (
    <Autocomplete
      autoSelect={false}
      disableCloseOnSelect
      multiple
      {...field}
      value={selectedValues}
      onChange={(_e, data) => formOnChange(data)}
      options={sortedOptions}
      getOptionLabel={option => (typeof option === 'string' ? option : option[labelAttribute])}
      isOptionEqualToValue={(option, val) => option === val || (option[labelAttribute] && option[labelAttribute] === val[labelAttribute])}
      renderOption={({ key, ...optionProps }, option, { selected }) => (
        <li key={key} {...optionProps}>
          <Checkbox checked={selected} sx={{ mr: 1 }} />
          {typeof option === 'string' ? option : option[labelAttribute]}
        </li>
      )}
      renderTags={
        chipDisplay
          ? (values, getTagProps) =>
              values.map((option, index) => {
                const { key, ...tagProps } = getTagProps({ index });
                return <Chip key={key} label={typeof option === 'string' ? option : option[labelAttribute]} size="small" {...tagProps} />;
              })
          : values => <TruncatedTagList labelAttribute={labelAttribute} values={values} />
      }
      renderInput={params => <TextField {...params} label={label} placeholder={selectedValues.length ? '' : placeholder} inputRef={ref} />}
      slotProps={{ listbox: { style: { maxHeight: listboxMaxHeight } } }}
      {...remainder}
    />
  );
};

export default CheckboxAutocomplete;
