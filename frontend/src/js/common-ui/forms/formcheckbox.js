// Copyright 2017 Northern.tech AS
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
import { Controller } from 'react-hook-form';

import { Checkbox, FormControlLabel } from '@mui/material';

export const FormCheckbox = ({ className, control, disabled, id, handleClick, style, label, required }) => (
  <Controller
    name={id}
    rules={{ required }}
    control={control}
    render={({ field: { value = false, onChange } }) => {
      return (
        <FormControlLabel
          className={className}
          control={
            <Checkbox name={id} onClick={handleClick} disabled={disabled} checked={value} style={style} color="primary" onChange={() => onChange(!value)} />
          }
          label={label}
        />
      );
    }}
  />
);

export default FormCheckbox;
