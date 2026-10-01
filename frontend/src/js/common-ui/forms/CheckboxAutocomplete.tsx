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
import type { ReactNode, Ref } from 'react';
import type { FieldValues, Path } from 'react-hook-form';
import { Controller, useFormContext } from 'react-hook-form';

import type { AutocompleteProps } from '@mui/material';
import { Autocomplete, Checkbox, Chip, TextField } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { TruncatedTagList } from './helpers';

const listboxMaxHeight = 304;

const useStyles = makeStyles()(() => ({
  optionLabel: { minWidth: 0, wordBreak: 'break-word' },
  tagsSelect: { width: 270 }
}));

type CheckboxAutocompleteProps<T> = {
  chipDisplay?: boolean;
  error?: boolean;
  helperText?: ReactNode;
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
  className = '',
  error,
  helperText,
  inputRef,
  label = '',
  labelAttribute = 'title',
  onChange,
  options = [],
  placeholder = '',
  value = [],
  ...remainder
}: CheckboxAutocompleteProps<T>) => {
  const { classes } = useStyles();

  return (
    <Autocomplete
      autoSelect={false}
      className={`${chipDisplay ? classes.tagsSelect : ''} ${className}`}
      disableCloseOnSelect
      multiple
      value={value ?? []}
      onChange={(_e, data) => onChange(data)}
      options={options}
      getOptionLabel={option => (typeof option === 'string' ? option : option[labelAttribute])}
      isOptionEqualToValue={(option, val) => option === val || (option[labelAttribute] != null && option[labelAttribute] === val[labelAttribute])}
      renderOption={({ key, ...optionProps }, option, { selected }) => (
        <li key={key} {...optionProps}>
          <Checkbox className="padding-none margin-right-x-small" checked={selected} />
          <span className={classes.optionLabel}>{typeof option === 'string' ? option : option[labelAttribute]}</span>
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
      renderInput={params => (
        <TextField {...params} error={error} helperText={helperText} label={label} placeholder={value?.length ? '' : placeholder} inputRef={inputRef} />
      )}
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
