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
import { Alert, Button } from '@mui/material';

import { DEVICE_STATES } from '@northern.tech/store/constants';
import pluralize from 'pluralize';

const stateActionMap = {
  [DEVICE_STATES.pending]: 'pending authorization',
  [DEVICE_STATES.rejected]: 'reject',
  [DEVICE_STATES.preauth]: 'preauthorized',
  [DEVICE_STATES.accepted]: 'accepted'
};

export const DeviceStatusNotification = ({ deviceCount, onClick, state }) => {
  const pluralized = pluralize('device', deviceCount);
  return (
    <Alert
      className="flexbox align-items-center margin-left-large margin-right"
      onClick={() => onClick(state)}
      severity="info"
      slotProps={{ message: { className: 'flexbox align-items-center space-between full-width' } }}
    >
      <div>
        {deviceCount} {pluralized} {pluralize('is', deviceCount)} {stateActionMap[state]}
      </div>
      <Button className="padding-bottom-none padding-top-none" size="small" variant="text">
        View details
      </Button>
    </Alert>
  );
};

export default DeviceStatusNotification;
