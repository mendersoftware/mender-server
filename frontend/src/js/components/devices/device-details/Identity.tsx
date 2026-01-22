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
import DeviceNameInput from '@northern.tech/common-ui/DeviceNameInput';
import Time from '@northern.tech/common-ui/Time';
import { TwoColumnData } from '@northern.tech/common-ui/TwoColumnData';
import { DEVICE_STATES } from '@northern.tech/store/constants';

import AuthStatus from './AuthStatus';
import DeviceDataCollapse from './DeviceDataCollapse';
import DeviceTags from './DeviceTags';

export const DeviceIdentity = ({ device, setSnackbar }) => {
  const { created_ts, id, identity_data = {}, status = DEVICE_STATES.accepted } = device;

  const { mac, ...remainingIdentity } = identity_data;

  const content = {
    ID: id || '-',
    ...(mac ? { mac } : {}),
    ...remainingIdentity
  };

  if (created_ts) {
    const createdTime = <Time value={created_ts} />;
    content[status === DEVICE_STATES.preauth ? 'Date added' : 'First request'] = createdTime;
  }

  return (
    <DeviceDataCollapse header={null} title="Device identity">
      <TwoColumnData data={{ Name: <DeviceNameInput device={device} isHovered />, ...content }} setSnackbar={setSnackbar} />
    </DeviceDataCollapse>
  );
};

export default DeviceIdentity;

export const IdentityTab = ({ device, setSnackbar, userCapabilities, onDecommissionDevice }) => (
  <>
    <DeviceIdentity device={device} setSnackbar={setSnackbar} />
    <AuthStatus device={device} decommission={onDecommissionDevice} />
    <DeviceTags device={device} setSnackbar={setSnackbar} userCapabilities={userCapabilities} />
  </>
);
