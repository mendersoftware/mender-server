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
import type { ReactNode } from 'react';
import type { Control, RegisterOptions } from 'react-hook-form';
import { Controller, useFormContext } from 'react-hook-form';

import { NumberField } from './NumberField';

type NumberInputProps = {
  className?: string;
  control?: Control;
  defaultValue?: number | null;
  disabled?: boolean;
  helperText?: ReactNode;
  id: string;
  label?: ReactNode;
  max?: number;
  min?: number;
  onBlur?: (value: number | null) => void;
  required?: boolean;
  requiredRendered?: boolean;
  rules?: RegisterOptions;
  showSteps?: boolean;
  size?: 'small' | 'medium';
  step?: number;
  width?: number | string;
};

export const NumberInput = ({
  className = '',
  control,
  defaultValue,
  disabled,
  helperText,
  id,
  label,
  max,
  min,
  onBlur: onBlurExternal,
  required,
  requiredRendered = true,
  rules,
  showSteps,
  size,
  step,
  width = 400
}: NumberInputProps) => {
  const { getValues } = useFormContext();

  const mergedRules: RegisterOptions = {
    required: required ? `${typeof label === 'string' ? label : 'This field'} is required` : false,
    ...rules
  };

  const wrapperClassName = `${className} ${required && requiredRendered ? 'required' : ''}`.trim();

  return (
    <Controller
      name={id}
      control={control}
      defaultValue={defaultValue}
      rules={mergedRules}
      render={({ field: { value, onChange, onBlur }, fieldState: { error } }) => (
        <NumberField
          id={id}
          className={wrapperClassName}
          style={{ width }}
          label={label}
          value={value ?? null}
          onValueChange={onChange}
          onBlur={e => {
            onBlur(e);
            onBlurExternal?.((getValues(id) as number | null) ?? null);
          }}
          min={min}
          max={max}
          step={step}
          required={required}
          disabled={disabled}
          showSteps={showSteps}
          size={size}
          error={!!error?.message}
          helperText={(error?.message as string) || (helperText as string)}
        />
      )}
    />
  );
};

export default NumberInput;
