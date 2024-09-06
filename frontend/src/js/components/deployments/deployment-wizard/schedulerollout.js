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
import React, { useState } from 'react';

import { FormControl, MenuItem, Select } from '@mui/material';
import { DateTimePicker } from '@mui/x-date-pickers';
import { makeStyles } from 'tss-react/mui';

import dayjs from 'dayjs';

import { BENEFITS } from '../../../constants/appConstants';
import EnterpriseNotification from '../../common/enterpriseNotification';
import { InfoHintContainer } from '../../common/info-hint';
import { HELPTOOLTIPS, MenderHelpTooltip } from '../../helptips/helptooltips';

const useStyles = makeStyles()(() => ({
  textField: { minWidth: 400 },
  infoStyle: { minWidth: 400, borderBottom: 'none' },
  pickerStyle: { marginBottom: 15, width: 'min-content' }
}));

export const ScheduleRollout = ({ canSchedule, commonClasses, setDeploymentSettings, deploymentObject, open = false }) => {
  const [isPickerOpen, setIsPickerOpen] = useState(open);
  const { classes } = useStyles();

  const { phases = [] } = deploymentObject;

  const handleStartTimeChange = value => {
    // if there is no existing phase, set phase and start time
    if (!phases.length) {
      setDeploymentSettings({ phases: [{ batch_size: 100, start_ts: value, delay: 0 }] });
    } else {
      //if there are existing phases, set the first phases to the new start time and adjust later phases in different function
      let newPhases = phases;
      newPhases[0].start_ts = value;
      setDeploymentSettings({ phases: newPhases });
    }
  };

  const handleStartChange = event => {
    // To be used with updated datetimepicker to open programmatically
    if (event.target.value) {
      setIsPickerOpen(true);
    } else {
      handleStartTimeChange();
    }
  };

  const start_time = phases.length ? phases[0].start_ts : undefined;

  const startTime = dayjs(start_time);
  return (
    <>
      <h4 className={`margin-bottom-none margin-top-none ${canSchedule ? '' : commonClasses.disabled}`}>Select a start time</h4>
      <div className={commonClasses.columns}>
        <FormControl className={classes.pickerStyle} disabled={!canSchedule}>
          <Select className={classes.textField} onChange={handleStartChange} value={start_time ? 'custom' : 0}>
            <MenuItem value={0}>Start immediately</MenuItem>
            <MenuItem value="custom">Schedule the start date &amp; time</MenuItem>
          </Select>
        </FormControl>
        <InfoHintContainer>
          <EnterpriseNotification id={BENEFITS.scheduledDeployments.id} />
          <MenderHelpTooltip id={HELPTOOLTIPS.scheduleDeployment.id} />
        </InfoHintContainer>
      </div>
      {Boolean(isPickerOpen || start_time) && (
        <FormControl className={`margin-top-none ${classes.pickerStyle}`} disabled={!canSchedule}>
          <DateTimePicker
            ampm={false}
            open={isPickerOpen}
            onOpen={() => setIsPickerOpen(true)}
            onClose={() => setIsPickerOpen(false)}
            label="Starting at"
            minDateTime={dayjs()}
            disabled={!canSchedule}
            onChange={date => handleStartTimeChange(date.toISOString())}
            slotProps={{ textField: { style: { minWidth: 400 } } }}
            value={startTime}
          />
        </FormControl>
      )}
    </>
  );
};
