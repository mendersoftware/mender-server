// Copyright 2023 Northern.tech AS
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
import { useState } from 'react';
import { Controller, useFormContext } from 'react-hook-form';

import { Cancel as CancelIcon } from '@mui/icons-material';
import { Autocomplete, Chip, TextField } from '@mui/material';

import { duplicateFilter, unionizeStrings } from '@northern.tech/utils/helpers';

export const ChipSelect = ({ className = '', name, disabled = false, helperText, inputRef, label = '', options = [], placeholder = '' }) => {
  const [value, setValue] = useState('');

  const { control, getValues } = useFormContext();

  // to allow device types to automatically be selected on entered ',' we have to filter the input and transform any completed device types (followed by a ',')
  // while also checking for duplicates and allowing complete resets of the input
  const onTextInputChange = (inputValue, reason, setCurrentSelection) => {
    const value = inputValue || '';
    if (reason === 'clear') {
      setValue('');
      return setCurrentSelection([]);
    } else if (reason === 'reset') {
      return setValue('');
    }
    const lastIndex = value.lastIndexOf(',');
    const possibleSelection = value.substring(0, lastIndex).split(',').filter(duplicateFilter);
    const currentValue = value.substring(lastIndex + 1);
    const selection = getValues(name);
    const nextSelection = unionizeStrings(selection, possibleSelection);
    setValue(currentValue);
    setCurrentSelection(nextSelection);
  };

  const onTextInputLeave = (value, setCurrentSelection) => {
    const selection = getValues(name);
    const nextSelection = unionizeStrings(selection, [value]);
    setCurrentSelection(nextSelection);
    setValue('');
  };

  return (
    <Controller
      control={control}
      name={name}
      render={({ field: { onChange: formOnChange, value: currentSelection, ref, ...props } }) => (
        <Autocomplete
          id={`${name}-chip-select`}
          value={currentSelection}
          className={className}
          filterSelectedOptions
          freeSolo={true}
          includeInputInList={true}
          multiple
          // allow edits to the textinput without deleting existing device types by ignoring backspace
          onChange={(e, value) => (e.key !== 'Backspace' ? formOnChange(value) : null)}
          onInputChange={(e, v, reason) => onTextInputChange(null, reason, formOnChange)}
          options={options}
          readOnly={disabled}
          ref={ref}
          renderTags={(values, getTagProps) =>
            values.map((option, index) => {
              const { key, onDelete, ...tagProps } = getTagProps({ index });
              return (
                <Chip label={option} key={key} onDelete={onDelete} deleteIcon={<CancelIcon onClick={onDelete} aria-label={`${name}-delete`} />} {...tagProps} />
              );
            })
          }
          renderInput={params => (
            <TextField
              {...params}
              fullWidth
              slotProps={{
                htmlInput: { ...params.inputProps, value },
                input: params.InputProps
              }}
              key={`${name}-input`}
              label={label}
              variant={disabled ? 'standard' : 'outlined'}
              onBlur={e => onTextInputLeave(e.target.value, formOnChange)}
              onChange={e => onTextInputChange(e.target.value, 'input', formOnChange)}
              placeholder={currentSelection.length ? '' : placeholder}
              helperText={helperText}
              inputRef={inputRef}
            />
          )}
          {...props}
        />
      )}
    />
  );
};

export default ChipSelect;
