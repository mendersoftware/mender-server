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

import DeviceIdentityDisplay from '@northern.tech/common-ui/deviceidentity';
import Time from '@northern.tech/common-ui/time';
import { DEPLOYMENT_ROUTES, auditlogTypes, canAccess } from '@northern.tech/store/constants';

const ArtifactLink = ({ item }) => <Link to={`/releases/${item.object.artifact.name}`}>View artifact</Link>;
const DeploymentLink = ({ item }) => <Link to={`${DEPLOYMENT_ROUTES.finished.route}?open=true&id=${item.object.id}`}>View deployment</Link>;
const DeviceLink = ({ item }) => <Link to={`/devices?id=${item.object.id}`}>View device</Link>;
const DeviceRejectedLink = ({ item }) => <Link to={`/devices/rejected?id=${item.object.id}`}>View device</Link>;
const TerminalSessionLink = () => <a>View session log</a>;
const ChangeFallback = props => {
  const {
    item: { change = '-' }
  } = props;
  return <div>{change}</div>;
};

const FallbackFormatter = props => {
  let result = '';
  try {
    result = JSON.stringify(props);
  } catch (error) {
    console.log(error);
  }
  return <div>{result}</div>;
};

const ArtifactFormatter = ({ artifact }) => <div>{artifact.name}</div>;
const DeploymentFormatter = ({ deployment }) => <div>{deployment.name}</div>;
const DeviceFormatter = ({ id }) => <DeviceIdentityDisplay device={{ id }} />;
const UserFormatter = ({ user }) => <div>{user.email}</div>;
const TenantFormatter = ({ tenant }) => <div>{tenant.name}</div>;

const defaultAccess = canAccess;
const changeMap = {
  default: { component: 'div', actionFormatter: FallbackFormatter, title: 'defaultTitle', accessCheck: defaultAccess },
  artifact: { actionFormatter: ArtifactFormatter, component: ArtifactLink, accessCheck: ({ canReadReleases }) => canReadReleases },
  deployment: {
    actionFormatter: DeploymentFormatter,
    component: DeploymentLink,
    accessCheck: ({ canReadDeployments }) => canReadDeployments
  },
  deviceDecommissioned: { actionFormatter: DeviceFormatter, component: 'div', accessCheck: defaultAccess },
  deviceRejected: { actionFormatter: DeviceFormatter, component: DeviceRejectedLink, accessCheck: ({ canReadDevices }) => canReadDevices },
  deviceGeneral: { actionFormatter: DeviceFormatter, component: DeviceLink, accessCheck: ({ canReadDevices }) => canReadDevices },
  deviceTerminalSession: { actionFormatter: DeviceFormatter, component: TerminalSessionLink, accessCheck: defaultAccess },
  user: { actionFormatter: UserFormatter, component: ChangeFallback, accessCheck: defaultAccess },
  user_access_token: { actionFormatter: FallbackFormatter, component: ChangeFallback, accessCheck: defaultAccess },
  tenant: { actionFormatter: TenantFormatter, component: ChangeFallback, accessCheck: defaultAccess }
};

const mapChangeToContent = item => {
  let content = changeMap[item.object.type];
  if (content) {
    return content;
  } else if (item.object.type === 'device' && item.action.includes('terminal')) {
    content = changeMap.deviceTerminalSession;
  } else if (item.object.type === 'device' && item.action.includes('reject')) {
    content = changeMap.deviceRejected;
  } else if (item.object.type === 'device' && item.action.includes('decommission')) {
    content = changeMap.deviceDecommissioned;
  } else if (item.object.type === 'device') {
    content = changeMap.deviceGeneral;
  } else {
    content = changeMap.default;
  }
  return content;
};

const actorMap = {
  user: 'email',
  device: 'id'
};

export const UserDescriptor = (item, index) => <div key={`${item.time}-${index} `}>{item.actor[actorMap[item.actor.type]]}</div>;
export const ActionDescriptor = (item, index) => (
  <div className="uppercased" key={`${item.time}-${index}`}>
    {item.action}
  </div>
);
export const TypeDescriptor = (item, index) => (
  <div className="capitalized" key={`${item.time}-${index}`}>
    {auditlogTypes[item.object.type]?.title ?? item.object.type}
  </div>
);
export const ChangeDescriptor = (item, index) => {
  const FormatterComponent = mapChangeToContent(item).actionFormatter;
  return <FormatterComponent key={`${item.time}-${index}`} {...item.object} />;
};
export const ChangeDetailsDescriptor = (item, index, userCapabilities) => {
  const { component: Comp, accessCheck } = mapChangeToContent(item);
  const key = `${item.time}-${index}`;
  return accessCheck(userCapabilities) ? <Comp key={key} item={item} /> : <div key={key} />;
};
export const TimeWrapper = (item, index) => <Time key={`${item.time}-${index}`} value={item.time} />;
