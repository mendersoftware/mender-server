// Copyright 2016 Northern.tech AS
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
import React from 'react';
import { Controller, useFormContext } from 'react-hook-form';

import { FormControl, FormHelperText, Input, InputLabel } from '@mui/material';

import { runValidations } from './form';

export const TextInput = ({
  autocomplete,
  className = '',
  control,
  controlRef,
  disabled,
  hint,
  id,
  InputLabelProps = {},
  InputProps = {},
  label,
  required,
  type,
  validations = '',
  numericValidations = {},
  value: passedValue = ''
}) => {
  const {
    clearErrors,
    formState: { errors },
    setError
  } = useFormContext();
  const errorKey = `${id}-error`;

  const validate = value => {
    const { isValid, errortext } = runValidations({ id, required, validations, value, wasMaybeTouched: !!errors[id] });
    if (isValid) {
      clearErrors(errorKey);
    } else {
      setError(errorKey, { type: 'validate', message: errortext });
    }
    return isValid;
  };
  return (
    <Controller
      name={id}
      control={control}
      rules={{ required, validate, ...numericValidations }}
      render={({ field: { value, onChange, onBlur, ref }, fieldState: { error } }) => (
        <FormControl className={`${className} ${required ? 'required' : ''}`} error={Boolean(error?.message || errors[errorKey])} style={{ width: 400 }}>
          <InputLabel htmlFor={id} {...InputLabelProps}>
            {label}
          </InputLabel>
          <Input
            autoComplete={autocomplete}
            id={id}
            name={id}
            disabled={disabled}
            inputRef={inputRef => {
              ref(inputRef);
              if (controlRef) {
                controlRef.current = inputRef;
              }
            }}
            value={value ?? passedValue}
            onChange={({ target: { value } }) => onChange(value)}
            onBlur={() => (validations.includes('trim') ? onChange((value ?? passedValue).trim()) : onBlur)}
            placeholder={hint}
            required={required}
            type={type}
            {...InputProps}
          />
          <FormHelperText>{(errors[errorKey] || error)?.message}</FormHelperText>
        </FormControl>
      )}
    />
  );
};

export default TextInput;
