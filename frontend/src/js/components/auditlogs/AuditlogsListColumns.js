import React from 'react';
import { Link } from 'react-router-dom';

import DeviceIdentityDisplay from '@northern.tech/common-ui/deviceidentity';
import Time from '@northern.tech/common-ui/time';
import { DEPLOYMENT_ROUTES, canAccess } from '@northern.tech/store/constants';

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
  user: { component: ChangeFallback, actionFormatter: UserFormatter, accessCheck: defaultAccess },
  tenant: { actionFormatter: TenantFormatter, accessCheck: defaultAccess, component: ChangeFallback }
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

const UserDescriptor = (item, index) => <div key={`${item.time}-${index} `}>{item.actor[actorMap[item.actor.type]]}</div>;
const ActionDescriptor = (item, index) => (
  <div className="uppercased" key={`${item.time}-${index}`}>
    {item.action}
  </div>
);
const TypeDescriptor = (item, index) => (
  <div className="capitalized" key={`${item.time}-${index}`}>
    {item.object.type}
  </div>
);
const ChangeDescriptor = (item, index) => {
  const FormatterComponent = mapChangeToContent(item).actionFormatter;
  return <FormatterComponent key={`${item.time}-${index}`} {...item.object} />;
};
const ChangeDetailsDescriptor = (item, index, userCapabilities) => {
  const { component: Comp, accessCheck } = mapChangeToContent(item);
  const key = `${item.time}-${index}`;
  return accessCheck(userCapabilities) ? <Comp key={key} item={item} /> : <div key={key} />;
};
const TimeWrapper = (item, index) => <Time key={`${item.time}-${index}`} value={item.time} />;

const auditLogColumns = [
  { title: 'Performed by', sortable: false, render: UserDescriptor },
  { title: 'Action', sortable: false, render: ActionDescriptor },
  { title: 'Type', sortable: false, render: TypeDescriptor },
  { title: 'Changed', sortable: false, render: ChangeDescriptor },
  { title: 'More details', sortable: false, render: ChangeDetailsDescriptor },
  { title: 'Time', sortable: true, render: TimeWrapper }
];

export default auditLogColumns;
