import React from 'react';
import { Controller, useFormContext } from 'react-hook-form';

import { Autocomplete } from '@mui/material';

// eslint-disable-next-line no-unused-vars
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
