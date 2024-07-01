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
import React, { useMemo } from 'react';
import { useSelector } from 'react-redux';

import { makeStyles } from 'tss-react/mui';

import GatewayConnectionIcon from '../../../assets/img/gateway-connection.svg';
import GatewayIcon from '../../../assets/img/gateway.svg';
import { stringToBoolean } from '../../helpers';
import { getIdAttribute } from '../../selectors';
import { getDeviceIdentityText } from '../devices/base-devices';
import DeviceNameInput from './devicenameinput';

const useStyles = makeStyles()(theme => ({
  container: {
    gridTemplateColumns: '1fr max-content',
    columnGap: theme.spacing()
  },
  gatewayIcon: {
    color: theme.palette.grey[400],
    width: 'max-content',
    marginRight: theme.spacing()
  }
}));

const DeviceIdComponent = ({ style = {}, value }) => <div style={style}>{value}</div>;

const attributeComponentMap = {
  default: DeviceIdComponent,
  name: DeviceNameInput
};

const adornments = [
  {
    component: GatewayConnectionIcon,
    isApplicable: ({ attributes = {} }) => !stringToBoolean(attributes.mender_is_gateway) && !!attributes.mender_gateway_system_id
  },
  { component: GatewayIcon, isApplicable: ({ attributes = {} }) => stringToBoolean(attributes.mender_is_gateway) }
];

export const DeviceIdentityDisplay = props => {
  const { device = {}, isEditable = true, hasAdornment = true } = props;

  const { attribute: idAttribute } = useSelector(getIdAttribute);
  const stateDevice = useSelector(state => state.devices.byId[device?.id]) || {};
  const idValue = getDeviceIdentityText({ device: { ...device, ...stateDevice }, idAttribute });
  const { classes } = useStyles();

  const Component = !isEditable ? attributeComponentMap.default : attributeComponentMap[idAttribute] ?? attributeComponentMap.default;
  const { attributes = {} } = device;
  const EndAdornment = useMemo(
    () => adornments.find(item => item.isApplicable(device))?.component,
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [attributes.mender_is_gateway, attributes.mender_gateway_system_id]
  );
  return (
    // due to the specificity of the deviceListRow child class, applying the display styling through the container class doesn't work, thus the inline style in addition here
    <div className={classes.container} style={{ display: 'grid' }}>
      <Component {...props} value={idValue} />
      {hasAdornment && EndAdornment && <EndAdornment className={classes.gatewayIcon} />}
    </div>
  );
};

export default DeviceIdentityDisplay;
