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
import React from 'react';
import { Link } from 'react-router-dom';

import pluralize from 'pluralize';

import preauthImage from '../../../assets/img/preauthorize.png';
import { DEVICE_STATES } from '../../constants/deviceConstants';
import { currentArtifact, rootfsImageVersion } from '../../constants/releaseConstants';
import Time, { ApproximateRelativeDate } from '../common/time';
import DeviceStatus from './device-status';

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
  const attributeValue = accessorTarget[column.attribute.name] || accessorTarget[column.attribute.alternative] || device[column.attribute.name];
  return typeof attributeValue === 'object' ? JSON.stringify(attributeValue) : attributeValue;
};

export const getDeviceIdentityText = ({ device = {}, idAttribute }) => {
  const { id = '', identity_data = {}, tags = {} } = device;
  // eslint-disable-next-line no-unused-vars
  const { status, ...remainingIds } = identity_data;
  const nonIdKey = Object.keys(remainingIds)[0];
  if (!idAttribute || idAttribute === 'id' || idAttribute === 'Device ID') {
    return id;
  } else if (idAttribute === 'name') {
    return tags[idAttribute] ?? `${id.substring(0, 6)}...`;
  }
  return identity_data[idAttribute] ?? identity_data[nonIdKey] ?? id;
};

const AttributeRenderer = ({ content, textContent }) => (
  <div title={textContent}>
    <div className="text-overflow">{content}</div>
  </div>
);

export const DefaultAttributeRenderer = ({ column, device, idAttribute }) => (
  <AttributeRenderer content={column.textRender({ device, column, idAttribute })} textContent={column.textRender({ device, column, idAttribute })} />
);

export const getDeviceSoftwareText = (attributes = {}) => attributes[rootfsImageVersion] || '-';
export const DeviceSoftware = ({ device }) => (
  <AttributeRenderer content={getDeviceSoftwareText(device.attributes)} textContent={getDeviceSoftwareText(device.attributes)} />
);

export const getDeviceArtifactText = (attributes = {}) => attributes.artifact_name || '-';
export const DeviceArtifact = ({ device }) => (
  <AttributeRenderer content={getDeviceArtifactText(device.attributes)} textContent={getDeviceArtifactText(device.attributes)} />
);

export const getDeviceTypeText = (attributes = {}) => (attributes.device_type?.length ? attributes.device_type.join(',') : '-');
export const DeviceTypes = ({ device }) => (
  <AttributeRenderer content={getDeviceTypeText(device.attributes)} textContent={getDeviceTypeText(device.attributes)} />
);

export const RelativeDeviceTime = ({ device }) => (
  <div>
    <ApproximateRelativeDate updateTime={device.check_in_time_rounded} />
  </div>
);

export const DeviceCreationTime = ({ device }) =>
  device.created_ts ? (
    <div>
      <Time value={device.created_ts} />
    </div>
  ) : (
    '-'
  );

export const DeviceStatusRenderer = ({ device }) => (
  <div>
    <DeviceStatus device={device} />
  </div>
);

export const AcceptedEmptyState = ({ allCount }) => (
  <div className="dashboard-placeholder">
    <p>No devices found</p>
    {!allCount && (
      <>
        <p>No devices have been authorized to connect to the Mender server yet.</p>
        <p>
          Visit the <Link to="/help/get-started">Help section</Link> to learn how to connect devices to the Mender server.
        </p>
      </>
    )}
  </div>
);

export const PreauthorizedEmptyState = ({ canManageDevices, limitMaxed, onClick }) => (
  <div className="dashboard-placeholder">
    <p>There are no preauthorized devices.</p>
    {canManageDevices && (
      <p>
        {limitMaxed ? 'Preauthorize devices' : <a onClick={onClick}>Preauthorize devices</a>} so that when they come online, they will connect to the server
        immediately
      </p>
    )}
    <img src={preauthImage} alt="preauthorize" />
  </div>
);

export const PendingEmptyState = ({ filters }) => (
  <div className="dashboard-placeholder">
    <p>
      {filters.length
        ? `There are no pending devices matching the selected ${pluralize('filters', filters.length)}`
        : 'There are no devices pending authorization'}
    </p>
    <p>
      Visit the <Link to="/help/get-started">Help section</Link> to learn how to connect devices to the Mender server.
    </p>
  </div>
);

export const RejectedEmptyState = ({ filters }) => (
  <div className="dashboard-placeholder">
    <p>{filters.length ? `There are no rejected devices matching the selected ${pluralize('filters', filters.length)}` : 'There are no rejected devices'}</p>
  </div>
);

export const defaultHeaders = {
  currentSoftware: {
    title: 'Current software',
    attribute: { name: rootfsImageVersion, scope: 'inventory' },
    component: DeviceSoftware,
    sortable: true,
    textRender: getDeviceSoftwareText
  },
  currentArtifact: {
    title: 'Current artifact',
    attribute: { name: currentArtifact, scope: 'inventory' },
    component: DeviceArtifact,
    sortable: true,
    textRender: getDeviceArtifactText
  },
  deviceCreationTime: {
    title: 'First request',
    attribute: { name: 'created_ts', scope: 'system' },
    component: DeviceCreationTime,
    sortable: true
  },
  deviceId: {
    title: 'Device ID',
    attribute: { name: 'id', scope: 'identity' },
    sortable: true,
    textRender: ({ device }) => device.id
  },
  deviceStatus: {
    title: 'Status',
    attribute: { name: 'status', scope: 'identity' },
    component: DeviceStatusRenderer,
    sortable: true,
    textRender: defaultTextRender
  },
  deviceType: {
    title: 'Device type',
    attribute: { name: 'device_type', scope: 'inventory' },
    component: DeviceTypes,
    sortable: true,
    textRender: getDeviceTypeText
  },
  lastCheckIn: {
    title: 'Latest activity',
    attribute: { name: 'check_in_time', scope: 'system' },
    component: RelativeDeviceTime,
    sortable: true
  }
};

const baseDevicesRoute = '/devices';

const acceptedDevicesRoute = {
  key: DEVICE_STATES.accepted,
  groupRestricted: false,
  route: `${baseDevicesRoute}/${DEVICE_STATES.accepted}`,
  title: () => DEVICE_STATES.accepted,
  emptyState: AcceptedEmptyState,
  defaultHeaders: [defaultHeaders.deviceType, defaultHeaders.currentSoftware, defaultHeaders.lastCheckIn]
};

export const routes = {
  allDevices: {
    ...acceptedDevicesRoute,
    route: baseDevicesRoute,
    key: 'any',
    title: () => 'any'
  },
  devices: acceptedDevicesRoute,
  [DEVICE_STATES.accepted]: acceptedDevicesRoute,
  [DEVICE_STATES.pending]: {
    key: DEVICE_STATES.pending,
    groupRestricted: true,
    route: `${baseDevicesRoute}/${DEVICE_STATES.pending}`,
    title: count => `${DEVICE_STATES.pending}${count ? ` (${count})` : ''}`,
    emptyState: PendingEmptyState,
    defaultHeaders: [defaultHeaders.deviceCreationTime, defaultHeaders.lastCheckIn]
  },
  [DEVICE_STATES.preauth]: {
    key: DEVICE_STATES.preauth,
    groupRestricted: true,
    route: `${baseDevicesRoute}/${DEVICE_STATES.preauth}`,
    title: () => DEVICE_STATES.preauth,
    emptyState: PreauthorizedEmptyState,
    defaultHeaders: [
      {
        ...defaultHeaders.deviceCreationTime,
        title: 'Date added'
      }
    ]
  },
  [DEVICE_STATES.rejected]: {
    key: DEVICE_STATES.rejected,
    groupRestricted: true,
    route: `${baseDevicesRoute}/${DEVICE_STATES.rejected}`,
    title: () => DEVICE_STATES.rejected,
    emptyState: RejectedEmptyState,
    defaultHeaders: [defaultHeaders.deviceCreationTime, defaultHeaders.lastCheckIn]
  }
};

export const sortingAlternatives = Object.values(routes)
  .reduce((accu, item) => [...accu, ...item.defaultHeaders], [])
  .reduce((accu, item) => {
    if (item.attribute.alternative) {
      accu[item.attribute.name] = item.attribute.alternative;
      accu[item.attribute.alternative] = item.attribute.name;
    }
    return accu;
  }, {});
