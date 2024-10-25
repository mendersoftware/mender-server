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
import { Controller } from 'react-hook-form';

import { Autocomplete, TextField } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { countries } from '@northern.tech/store/constants';

interface CountrySelectProps {
  id?: string;
  onChange: (...event: any[]) => void;
  [other: string]: any;
}
const useStyles = makeStyles()(() => ({
  autocomplete: { width: 400 }
}));

export const CountrySelect = (props: CountrySelectProps) => {
  const { id, onChange, ...restProps } = props;
  const { classes } = useStyles();
  return (
    <Autocomplete
      getOptionLabel={option => option.label}
      options={countries}
      className={classes.autocomplete}
      autoHighlight
      renderInput={params => <TextField {...params} label="Country" id={id || 'country'} />}
      onChange={(e, data) => onChange(data)}
      {...restProps}
    />
  );
};

export const ControlledCountrySelect = ({ control, id, required }) => (
  <Controller
    rules={{ required }}
    render={({ field: { onChange }, ...props }) => <CountrySelect onChange={onChange} id={id} {...props} />}
    name="country"
    control={control}
  />
);
