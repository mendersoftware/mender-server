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
import { useEffect, useState } from 'react';
import { Controller, useFormContext } from 'react-hook-form';

import { Autocomplete, TextField } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { countries } from '@northern.tech/store/constants';

interface CountrySelectProps {
  [other: string]: any;
  error: boolean;
  helperText?: string;
  id?: string;
  onChange: (...event: any[]) => void;
}
const useStyles = makeStyles()(() => ({
  autocomplete: { width: 500 }
}));

export const CountrySelect = (props: CountrySelectProps) => {
  const { id, onChange, defaultValue, error, helperText, ...restProps } = props;
  const { classes } = useStyles();
  return (
    <Autocomplete
      key={defaultValue}
      getOptionLabel={option => option.label}
      options={countries}
      className={classes.autocomplete}
      autoHighlight
      renderInput={params => <TextField {...params} error={error} helperText={helperText} label="Country or region" id={id || 'country'} />}
      onChange={(e, data) => onChange(data)}
      {...restProps}
      defaultValue={countries.find(country => country.code === defaultValue)}
    />
  );
};

export const ControlledCountrySelect = ({ control, id, required }) => {
  const [defaultCountry, setDefaultCountry] = useState('');
  const { getValues } = useFormContext();
  const values = getValues();
  useEffect(() => {
    setDefaultCountry(values.country);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);
  return (
    <Controller
      rules={{ required: required ? 'Country or region is required' : false }}
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      render={({ field: { onChange }, formState, fieldState: { error }, ...props }) => (
        <CountrySelect defaultValue={defaultCountry} onChange={onChange} id={id} error={!!error} helperText={error?.message} {...props} />
      )}
      name="country"
      control={control}
    />
  );
};
