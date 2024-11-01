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
import React, { useEffect, useState } from 'react';
import { FormProvider, useForm } from 'react-hook-form';

import { makeStyles } from 'tss-react/mui';

import { TIMEOUTS } from '@northern.tech/store/constants';
import { useDebounce } from '@northern.tech/utils/debouncehook';

const useStyles = makeStyles()(theme => ({
  filters: {
    backgroundColor: theme.palette.background.lightgrey,
    columnGap: theme.spacing(2),
    display: 'flex',
    flexWrap: 'wrap',
    padding: `10px ${theme.spacing(3)} ${theme.spacing(3)}`,
    rowGap: theme.spacing(2),
    '.filter-item': {
      display: 'grid'
    },
    '.filter-item > div': {
      alignSelf: 'end'
    }
  },
  filterReset: { right: theme.spacing(3) }
}));

export const Filters = ({ className = '', defaultValues, filters = [], initialValues, onChange, fieldResetTrigger = '', dirtyField, clearDirty }) => {
  const { classes } = useStyles();
  const [values, setValues] = useState(initialValues);

  const methods = useForm({ mode: 'onChange', defaultValues });
  const { formState, reset, resetField, watch, setValue, getValues } = methods;
  const { isDirty } = formState;

  useEffect(() => {
    Object.entries(initialValues).map(([key, value]) => setValue(key, value));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(initialValues), setValue]);

  useEffect(() => {
    if (dirtyField && !formState.isDirty) {
      setValue(dirtyField, getValues(dirtyField), { shouldDirty: true });
      clearDirty('');
    }
  }, [clearDirty, dirtyField, formState, getValues, setValue]);

  useEffect(() => {
    if (!fieldResetTrigger) {
      return;
    }
    resetField(fieldResetTrigger);
  }, [fieldResetTrigger, resetField]);

  watch(setValues);
  const debouncedValues = useDebounce(values, TIMEOUTS.oneSecond);

  useEffect(() => {
    onChange(debouncedValues);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(debouncedValues), onChange]);

  return (
    <FormProvider {...methods}>
      <form className={`margin-bottom relative margin-top ${classes.filters} ${className}`} noValidate>
        {filters.map(({ key, title, Component, componentProps }) => (
          <div className="filter-item" key={key}>
            <h5 className="margin-top-small margin-bottom-none muted">{title}</h5>
            <Component name={key} {...componentProps} />
          </div>
        ))}
        {isDirty && (
          <span className={`link absolute ${classes.filterReset}`} onClick={() => reset()}>
            Clear filter
          </span>
        )}
      </form>
    </FormProvider>
  );
};

export default Filters;
