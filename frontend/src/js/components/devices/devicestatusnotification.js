// Copyright 2021 Northern.tech AS
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

import { makeStyles } from 'tss-react/mui';

import pluralize from 'pluralize';

import { DEVICE_STATES } from '../../constants/deviceConstants';
import InfoText from '../common/infotext';

const useStyles = makeStyles()(theme => ({
  default: {
    flexGrow: 1,
    ['&.onboard']: {
      background: theme.palette.grey[400],
      padding: 10
    }
  }
}));

const stateActionMap = {
  [DEVICE_STATES.pending]: 'pending authorization',
  [DEVICE_STATES.rejected]: 'reject',
  [DEVICE_STATES.preauth]: 'preauthorized',
  [DEVICE_STATES.accepted]: 'accepted'
};

export const DeviceStatusNotification = ({ deviceCount, onClick, state }) => {
  const { classes } = useStyles();
  const pluralized = pluralize('device', deviceCount);
  return (
    <div className={`clickable margin-right onboard ${classes.default}`} onClick={() => onClick(state)}>
      <InfoText variant="dense">
        {deviceCount} {pluralized} {pluralize('is', deviceCount)} {stateActionMap[state]}
      </InfoText>
      <b className="margin-left-small link link-color">View {pluralized}</b>
    </div>
  );
};

export default DeviceStatusNotification;
