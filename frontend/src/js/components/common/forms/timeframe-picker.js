// Copyright 2020 Northern.tech AS
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
import { Controller, useFormContext } from 'react-hook-form';

import { DatePicker } from '@mui/x-date-pickers/DatePicker';

import dayjs from 'dayjs';

const ensureStartOfDay = date => {
  const momentDate = typeof date === 'string' ? dayjs(date.replace('Z', '')) : dayjs(date);
  return `${momentDate.format().split('T')[0]}T00:00:00.000`;
};

const ensureEndOfDay = date => {
  const momentDate = typeof date === 'string' ? dayjs(date.replace('Z', '')) : dayjs(date);
  return `${momentDate.format().split('T')[0]}T23:59:59.999`;
};

export const TimeframePicker = ({ tonight: propsTonight }) => {
  const [tonight] = useState(dayjs(propsTonight));
  const [maxStartDate, setMaxStartDate] = useState(tonight);
  const [minEndDate, setMinEndDate] = useState(tonight);

  const { control, setValue, watch, getValues } = useFormContext();

  const startDate = watch('startDate');
  const endDate = watch('endDate');

  useEffect(() => {
    let currentEndDate = getValues('endDate');
    const now = new Date().toISOString().replace('Z', '');
    if (startDate > currentEndDate) {
      setValue('endDate', ensureEndOfDay(startDate));
    } else if (currentEndDate > now) {
      setValue('endDate', now);
    }
    setMinEndDate(dayjs(startDate));
  }, [startDate, getValues, setValue]);

  useEffect(() => {
    let currentStartDate = getValues('startDate');
    if (endDate < currentStartDate) {
      setValue('startDate', ensureStartOfDay(endDate));
    }
    setMaxStartDate(dayjs(endDate));
  }, [endDate, getValues, setValue]);

  const handleChangeStartDate = date => ensureStartOfDay(date);

  const handleChangeEndDate = date => ensureEndOfDay(date);

  return (
    <div className="flexbox" style={{ flexWrap: 'wrap', gap: 15 }}>
      <Controller
        name="startDate"
        control={control}
        render={({ field: { onChange, value } }) => (
          <DatePicker
            disabled={!value}
            disableFuture
            format="MMMM Do"
            label="From"
            maxDate={maxStartDate}
            onChange={e => onChange(handleChangeStartDate(e))}
            value={value ? dayjs(value) : dayjs()}
          />
        )}
      />
      <Controller
        name="endDate"
        control={control}
        render={({ field: { onChange, value } }) => (
          <DatePicker
            disableFuture
            format="MMMM Do"
            label="To"
            minDate={minEndDate}
            onChange={e => onChange(handleChangeEndDate(e))}
            value={value ? dayjs(value) : dayjs()}
          />
        )}
      />
    </div>
  );
};

export default TimeframePicker;
