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

import { DEVICE_STATES } from '../../../constants/deviceConstants';
import { TwoColumnData } from '../../common/configurationobject';
import DeviceNameInput from '../../common/devicenameinput';
import Time from '../../common/time';
import AuthStatus from './authstatus';
import DeviceDataCollapse from './devicedatacollapse';
import DeviceTags from './devicetags';

const style = { maxWidth: '80%', gridTemplateColumns: 'minmax(max-content, 150px) auto' };
const previewStyle = { ...style, marginBottom: 5 };

export const DeviceIdentity = ({ device, setSnackbar }) => {
  const { created_ts, id, identity_data = {}, status = DEVICE_STATES.accepted } = device;

  const { mac, ...remainingIdentity } = identity_data;

  let content = {
    ID: id || '-',
    ...(mac ? { mac } : {}),
    ...remainingIdentity
  };

  if (created_ts) {
    const createdTime = <Time value={created_ts} />;
    content[status === DEVICE_STATES.preauth ? 'Date added' : 'First request'] = createdTime;
  }

  return (
    <DeviceDataCollapse
      header={
        <TwoColumnData
          chipLikeKey
          compact
          style={{ ...previewStyle, alignItems: 'center', gridTemplateColumns: 'minmax(max-content, 150px) max-content' }}
          config={{ Name: device }}
          ValueProps={{ device, isHovered: true }}
          ValueComponent={DeviceNameInput}
        />
      }
      title="Device identity"
    >
      <TwoColumnData config={content} compact setSnackbar={setSnackbar} style={style} />
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
