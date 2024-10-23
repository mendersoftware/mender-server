// Copyright 2024 Northern.tech AS
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

import { Autocomplete } from '@mui/material';

// eslint-disable-next-line @typescript-eslint/no-unused-vars
export const ControlledAutoComplete = ({ freeSolo, name, onChange, onInputChange, ...remainder }) => {
  const { control } = useFormContext();

  return (
    <Controller
      control={control}
      name={name}
      render={({ field: { onChange: formOnChange, ...props } }) => {
        const onChangeHandler = (e, data) => formOnChange(data);
        return <Autocomplete {...(freeSolo ? { freeSolo, onInputChange: onChangeHandler } : { onChange: onChangeHandler })} {...props} {...remainder} />;
      }}
    />
  );
};
