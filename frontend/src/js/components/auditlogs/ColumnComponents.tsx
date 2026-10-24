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
import { Typography } from '@mui/material';

import DeviceIdentityDisplay from '@northern.tech/common-ui/DeviceIdentity';
import { Link } from '@northern.tech/common-ui/Link';
import TextOverflowMultiline from '@northern.tech/common-ui/TextOverflowMultiline';
import Time from '@northern.tech/common-ui/Time';
import { DEPLOYMENT_ROUTES, auditlogTypes, canAccess } from '@northern.tech/store/constants';
import { generateReleasesPath } from '@northern.tech/store/locationutils';
import type { AuditLog } from '@northern.tech/types/MenderTypes';

import { stringifyEvent } from './eventdetails/utils';

// navigating links shouldn't also trigger the row click that opens the details drawer
const StoppedLink = ({ children, to }) => (
  <Link onClick={event => event.stopPropagation()} to={to}>
    {children}
  </Link>
);

const ArtifactLink = ({ item }) => (
  <StoppedLink to={generateReleasesPath({ pageState: { selectedRelease: item.object.artifact.name } })}>View software</StoppedLink>
);
const DeploymentLink = ({ item }) => <StoppedLink to={`${DEPLOYMENT_ROUTES.finished.route}?open=true&id=${item.object.id}`}>View deployment</StoppedLink>;
const DeviceLink = ({ item }) => <StoppedLink to={`/devices?id=${item.object.id}`}>View device</StoppedLink>;
const DeviceRejectedLink = ({ item }) => <StoppedLink to={`/devices/rejected?id=${item.object.id}`}>View device</StoppedLink>;
const TerminalSessionLink = () => <Link>View session log</Link>;
const ChangeFallback = props => {
  const {
    item: { change = '-' }
  } = props;
  return <TextOverflowMultiline variant="body2">{change}</TextOverflowMultiline>;
};

const FallbackFormatter = props => <Typography variant="body2">{stringifyEvent(props)}</Typography>;

const ArtifactFormatter = ({ artifact }) => <TextOverflowMultiline variant="body2">{artifact.name}</TextOverflowMultiline>;
const DeploymentFormatter = ({ deployment }) => <TextOverflowMultiline variant="body2">{deployment.name}</TextOverflowMultiline>;
const DeviceFormatter = ({ id }) => <DeviceIdentityDisplay device={{ id }} />;
const UserFormatter = ({ user }) => <TextOverflowMultiline variant="body2">{user.email}</TextOverflowMultiline>;
const TenantFormatter = ({ tenant }) => <TextOverflowMultiline variant="body2">{tenant.name}</TextOverflowMultiline>;

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

export const UserDescriptor = (item: AuditLog) => <TextOverflowMultiline variant="body2">{item.actor[actorMap[item.actor.type]]}</TextOverflowMultiline>;
export const ActionDescriptor = (item: AuditLog) => (
  <TextOverflowMultiline className="uppercased" variant="body2">
    {item.action}
  </TextOverflowMultiline>
);
export const TypeDescriptor = (item: AuditLog) => (
  <TextOverflowMultiline className="capitalized" variant="body2">
    {auditlogTypes[item.object.type]?.title ?? item.object.type}
  </TextOverflowMultiline>
);
export const ChangeDescriptor = (item: AuditLog) => {
  const FormatterComponent = mapChangeToContent(item).actionFormatter;
  return <FormatterComponent {...item.object} />;
};
export const ChangeDetailsDescriptor = (item: AuditLog, userCapabilities) => {
  const { component: Comp, accessCheck } = mapChangeToContent(item);
  return accessCheck(userCapabilities) ? <Comp item={item} /> : null;
};
export const TimeWrapper = (item: AuditLog) => <Time value={item.time} Component={Typography} component="time" variant="body2" />;
