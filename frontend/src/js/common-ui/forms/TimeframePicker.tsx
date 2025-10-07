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
import { useEffect, useState } from 'react';
import { Controller, useFormContext } from 'react-hook-form';

import { Typography, useTheme } from '@mui/material';
import { DatePicker } from '@mui/x-date-pickers/DatePicker';
import { makeStyles } from 'tss-react/mui';

import dayjs from 'dayjs';

const useStyles = makeStyles()(theme => ({
  container: {
    alignItems: 'start',
    width: 'min-content'
  },
  inputs: {
    gap: theme.spacing(2),
    flexWrap: 'wrap',
    [theme.breakpoints.up('md')]: {
      flexWrap: 'nowrap'
    }
  }
}));

const ensureStartOfDay = date => {
  const momentDate = typeof date === 'string' ? dayjs(date.replace('Z', '')) : dayjs(date);
  return `${momentDate.format().split('T')[0]}T00:00:00.000`;
};

const ensureEndOfDay = date => {
  const momentDate = typeof date === 'string' ? dayjs(date.replace('Z', '')) : dayjs(date);
  return `${momentDate.format().split('T')[0]}T23:59:59.999`;
};

const no = () => false;

interface HasHelptextParams {
  endDate: string;
  startDate: string;
}

export const TimeframePicker = ({
  hasHelperText = no,
  helperText,
  tonight: propsTonight
}: {
  hasHelperText?: (params: HasHelptextParams) => boolean;
  helperText?: string;
  tonight: string;
}) => {
  const [tonight] = useState(dayjs(propsTonight));
  const [maxStartDate, setMaxStartDate] = useState(tonight);
  const [minEndDate, setMinEndDate] = useState(tonight);
  const [showsHelptext, setShowsHelptext] = useState(false);
  const { classes } = useStyles();

  const theme = useTheme();
  const isNextTheme = !theme.palette.background.lightgrey;

  const { control, setValue, watch, getValues } = useFormContext();

  const startDate = watch('startDate');
  const endDate = watch('endDate');

  useEffect(() => {
    const currentEndDate = getValues('endDate');
    const now = new Date().toISOString().replace('Z', '');
    if (startDate > currentEndDate) {
      setValue('endDate', ensureEndOfDay(startDate));
    } else if (currentEndDate > now) {
      setValue('endDate', now);
    }
    setMinEndDate(dayjs(startDate));
  }, [startDate, getValues, setValue]);

  useEffect(() => {
    const currentStartDate = getValues('startDate');
    if (endDate < currentStartDate) {
      setValue('startDate', ensureStartOfDay(endDate));
    }
    setMaxStartDate(dayjs(endDate));
  }, [endDate, getValues, setValue]);

  useEffect(() => {
    setShowsHelptext(hasHelperText({ startDate, endDate }));
  }, [endDate, hasHelperText, startDate]);

  const handleChangeStartDate = date => ensureStartOfDay(date);

  const handleChangeEndDate = date => ensureEndOfDay(date);

  return (
    <div className={`flexbox column ${classes.container}`}>
      <div className={`flexbox ${classes.inputs}`}>
        <Controller
          name="startDate"
          control={control}
          render={({ field: { onChange, value } }) => (
            <DatePicker
              disableFuture
              slotProps={{
                textField: props => ({
                  size: isNextTheme ? 'small' : 'medium',
                  inputProps: {
                    ...props.inputProps,
                    'aria-label': 'From'
                  }
                })
              }}
              format="YYYY-MM-DD"
              yearsOrder="desc"
              label="From"
              maxDate={maxStartDate}
              onChange={e => onChange(handleChangeStartDate(e))}
              value={value ? dayjs(value) : null}
            />
          )}
        />
        <Controller
          name="endDate"
          control={control}
          render={({ field: { onChange, value } }) => (
            <DatePicker
              disableFuture
              slotProps={{
                textField: props => ({
                  size: isNextTheme ? 'small' : 'medium',
                  inputProps: {
                    ...props.inputProps,
                    'aria-label': 'To'
                  }
                })
              }}
              format="YYYY-MM-DD"
              yearsOrder="desc"
              label="To"
              minDate={minEndDate}
              onChange={e => onChange(handleChangeEndDate(e))}
              value={value ? dayjs(value) : dayjs()}
            />
          )}
        />
      </div>
      {showsHelptext && (
        <Typography className="margin-left-small margin-top-x-small" color="info" variant="body2">
          {helperText}
        </Typography>
      )}
    </div>
  );
};

export default TimeframePicker;
