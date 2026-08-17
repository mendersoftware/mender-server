// Copyright 2019 Northern.tech AS
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
import { useState } from 'react';
import { useFormContext } from 'react-hook-form';

import { FormControl, FormHelperText, MenuItem, Select, Typography } from '@mui/material';
import { DateTimePicker } from '@mui/x-date-pickers';
import { makeStyles } from 'tss-react/mui';

import EnterpriseNotification from '@northern.tech/common-ui/EnterpriseNotification';
import { InfoHintContainer } from '@northern.tech/common-ui/InfoHint';
import { defaultTimeFormat } from '@northern.tech/common-ui/Time';
import { BENEFITS } from '@northern.tech/store/constants';
import dayjs from 'dayjs';

import { HELPTOOLTIPS } from '../../helptips/HelpTooltips';
import { MenderHelpTooltip } from '../../helptips/MenderTooltip';
import { deploymentFormSections, useValidatedSetValue } from './utils';

const useStyles = makeStyles()(() => ({
  textField: { minWidth: 400 },
  infoStyle: { minWidth: 400, borderBottom: 'none' },
  pickerStyle: { marginBottom: 15, width: 'min-content' }
}));

export const ScheduleRollout = ({ canSchedule, commonClasses, open = false }) => {
  const [isPickerOpen, setIsPickerOpen] = useState(open);
  const { classes } = useStyles();
  const {
    formState: { errors },
    watch
  } = useFormContext();
  const setValue = useValidatedSetValue();

  // the start time is tracked separately from the phases, as the individual phase start times are derived from it
  const start_time = watch(deploymentFormSections.startTime);

  const handleStartTimeChange = (value?: string) => setValue(deploymentFormSections.startTime, value);

  const handleStartChange = event => {
    // To be used with updated datetimepicker to open programmatically
    if (event.target.value) {
      setIsPickerOpen(true);
    } else {
      handleStartTimeChange();
    }
  };

  const startTime = start_time ? dayjs(start_time) : dayjs();
  return (
    <>
      <div className="flexbox margin-bottom-x-small">
        <Typography variant="subtitle1" className={canSchedule ? '' : commonClasses.disabled}>
          Select a start time
        </Typography>
        <MenderHelpTooltip className="margin-left-small" id={HELPTOOLTIPS.scheduleDeployment.id} small />
      </div>
      <div className={commonClasses.columns}>
        <FormControl className={classes.pickerStyle} disabled={!canSchedule} error={!!errors.startTime}>
          <Select className={classes.textField} onChange={handleStartChange} value={start_time ? 'custom' : 0}>
            <MenuItem value={0}>Start immediately</MenuItem>
            <MenuItem value="custom">Schedule the start date &amp; time</MenuItem>
          </Select>
          {!!errors.startTime && <FormHelperText>{errors.startTime.message}</FormHelperText>}
        </FormControl>
        <InfoHintContainer>
          <EnterpriseNotification id={BENEFITS.scheduledDeployments.id} />
        </InfoHintContainer>
      </div>
      {Boolean(isPickerOpen || start_time) && (
        <FormControl className={classes.pickerStyle} disabled={!canSchedule}>
          <DateTimePicker
            ampm={false}
            open={isPickerOpen}
            onOpen={() => setIsPickerOpen(true)}
            onClose={() => setIsPickerOpen(false)}
            label="Starting at"
            format={defaultTimeFormat}
            minDateTime={dayjs()}
            disabled={!canSchedule}
            onChange={date => handleStartTimeChange(date.toISOString())}
            slotProps={{ textField: { size: 'small', style: { minWidth: 400 } } }}
            value={startTime}
          />
        </FormControl>
      )}
    </>
  );
};
