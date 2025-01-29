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

import { getDeviceById as getDeviceByIdSelector, getIdAttribute } from '@northern.tech/store/selectors';
import { stringToBoolean } from '@northern.tech/utils/helpers';

import GatewayConnectionIcon from '../../assets/img/gateway-connection.svg';
import GatewayIcon from '../../assets/img/gateway.svg';
import DeviceNameInput from './DeviceNameInput';

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

const propertyNameMap = {
  inventory: 'attributes',
  identity: 'identity_data',
  system: 'system',
  monitor: 'monitor',
  tags: 'tags'
};

export const defaultTextRender = ({ column, device }) => {
  const propertyName = propertyNameMap[column.attribute.scope] ?? column.attribute.scope;
  const accessorTarget = device[propertyName] ?? device;
  const attributeValue = accessorTarget[column.attribute.name] || device[column.attribute.name];
  return (typeof attributeValue === 'object' ? JSON.stringify(attributeValue) : attributeValue) ?? '-';
};

export const getDeviceIdentityText = ({ device = {}, idAttribute }) => {
  const { id = '', identity_data = {}, tags = {} } = device;
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const { status, ...remainingIds } = identity_data;

  const nonIdKey = Object.keys(remainingIds)[0];
  if (!idAttribute || idAttribute === 'id' || idAttribute === 'Device ID') {
    return id;
  } else if (typeof idAttribute === 'string' || !Object.keys(idAttribute).length) {
    return identity_data[idAttribute] ?? identity_data[nonIdKey] ?? id;
  }
  const { attribute, scope } = idAttribute;
  // special handling for tags purely to handle the untagged devices case
  if (attribute === 'name' && scope === 'tags') {
    return tags[attribute] ?? `${id.substring(0, 6)}...`;
  }
  return defaultTextRender({ column: { attribute: { name: attribute, scope } }, device });
};

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

  const idAttribute = useSelector(getIdAttribute);
  const { attribute, scope } = idAttribute;
  const stateDevice = useSelector(state => getDeviceByIdSelector(state, device.id));
  const idValue = getDeviceIdentityText({ device: { ...device, ...stateDevice }, idAttribute });
  const { classes } = useStyles();

  let Component = attributeComponentMap.default;
  if (attribute === 'name' && scope === 'tags') {
    Component = isEditable ? attributeComponentMap.name : Component;
  }
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
