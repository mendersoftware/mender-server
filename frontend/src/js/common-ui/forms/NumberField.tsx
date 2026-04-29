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
import type { FocusEventHandler, ReactNode } from 'react';
import { useId } from 'react';

import { KeyboardArrowDown as KeyboardArrowDownIcon, KeyboardArrowUp as KeyboardArrowUpIcon } from '@mui/icons-material';
import { FormControl, FormHelperText, IconButton, InputAdornment, InputLabel, OutlinedInput } from '@mui/material';

import { NumberField as BaseNumberField } from '@base-ui/react/number-field';

type NumberFieldProps = BaseNumberField.Root.Props & {
  endAdornment?: ReactNode;
  error?: boolean;
  helperText?: ReactNode | string;
  label?: ReactNode;
  onBlur?: FocusEventHandler<HTMLInputElement>;
  showSteps?: boolean;
  size?: 'small' | 'medium';
};

export const NumberField = (props: NumberFieldProps) => {
  const { id: idProp, label, error, size = 'medium', showSteps = false, helperText, onBlur, endAdornment = null, ...other } = props;
  let id = useId();
  if (idProp) {
    id = idProp;
  }
  return (
    <BaseNumberField.Root
      allowOutOfRange
      format={{ useGrouping: false, ...other.format }}
      render={(props, state) => (
        <FormControl
          className={props.className}
          style={props.style}
          size={size}
          ref={props.ref}
          disabled={state.disabled}
          required={state.required}
          error={error}
          variant="outlined"
        >
          {props.children}
        </FormControl>
      )}
      {...other}
    >
      <InputLabel htmlFor={id}>{label}</InputLabel>
      <BaseNumberField.Input
        id={id}
        onBlur={onBlur}
        render={(props, state) => (
          <OutlinedInput
            label={label}
            inputRef={props.ref}
            value={state.inputValue}
            onBlur={props.onBlur}
            onChange={props.onChange}
            onKeyUp={props.onKeyUp}
            onKeyDown={props.onKeyDown}
            onFocus={props.onFocus}
            slotProps={{
              input: props
            }}
            endAdornment={
              showSteps ? (
                <InputAdornment
                  position="end"
                  sx={{
                    flexDirection: 'column',
                    maxHeight: 'unset',
                    alignSelf: 'stretch',
                    borderLeft: '1px solid',
                    borderColor: 'divider',
                    ml: 0,
                    '& button': {
                      py: 0,
                      flex: 1,
                      borderRadius: 0.5
                    }
                  }}
                >
                  <BaseNumberField.Increment render={<IconButton size={size} aria-label="Increase" />}>
                    <KeyboardArrowUpIcon fontSize={size} sx={{ transform: 'translateY(2px)' }} />
                  </BaseNumberField.Increment>

                  <BaseNumberField.Decrement render={<IconButton size={size} aria-label="Decrease" />}>
                    <KeyboardArrowDownIcon fontSize={size} sx={{ transform: 'translateY(-2px)' }} />
                  </BaseNumberField.Decrement>
                </InputAdornment>
              ) : (
                endAdornment
              )
            }
            sx={showSteps ? { pr: 0 } : undefined}
          />
        )}
      />
      <FormHelperText sx={{ ml: 0, '&:empty': { mt: 0 } }}>{helperText}</FormHelperText>
    </BaseNumberField.Root>
  );
};
