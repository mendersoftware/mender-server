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

import { Sort as SortIcon } from '@mui/icons-material';

import DetailsIndicator from '@northern.tech/common-ui/detailsindicator';
import DeviceIdentityDisplay from '@northern.tech/common-ui/deviceidentity';
import Loader from '@northern.tech/common-ui/loader';
import Pagination from '@northern.tech/common-ui/pagination';
import Time from '@northern.tech/common-ui/time';
import { DEPLOYMENT_ROUTES, SORTING_OPTIONS, canAccess } from '@northern.tech/store/constants';

import EventDetailsDrawer from './eventdetailsdrawer';

export const defaultRowsPerPage = 20;

const ArtifactLink = ({ item }) => <Link to={`/releases/${item.object.artifact.name}`}>View artifact</Link>;
const DeploymentLink = ({ item }) => <Link to={`${DEPLOYMENT_ROUTES.finished.route}?open=true&id=${item.object.id}`}>View deployment</Link>;
const DeviceLink = ({ item }) => <Link to={`/devices?id=${item.object.id}`}>View device</Link>;
const DeviceRejectedLink = ({ item }) => <Link to={`/devices/rejected?id=${item.object.id}`}>View device</Link>;
const TerminalSessionLink = () => <a>View session log</a>;
const UserChange = ({ item: { change = '-' } }) => {
  const formatChange = change => {
    const diff = change.split(/@@.*@@/);
    return diff.length > 1 ? diff[1].trim() : diff;
  };
  return (
    <div className="capitalized" style={{ alignItems: 'flex-start', whiteSpace: 'pre-line' }}>
      {formatChange(change)}
    </div>
  );
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
  user: { component: UserChange, actionFormatter: UserFormatter, accessCheck: defaultAccess }
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

export const AuditLogsList = ({
  eventItem,
  items,
  loading,
  onChangePage,
  onChangeRowsPerPage,
  onChangeSorting,
  selectionState,
  setAuditlogsState,
  userCapabilities
}) => {
  const { page, perPage, sort = {}, total: count } = selectionState;

  const onIssueSelection = selectedIssue =>
    setAuditlogsState({ selectedId: selectedIssue ? btoa(`${selectedIssue.action}|${selectedIssue.time}`) : undefined });

  return (
    !!items.length && (
      <div className="fadeIn deploy-table-contain auditlogs-list">
        <div className="auditlogs-list-item auditlogs-list-item-header muted">
          {auditLogColumns.map((column, index) => (
            <div
              className="columnHeader"
              key={`columnHeader-${index}`}
              onClick={() => (column.sortable ? onChangeSorting() : null)}
              style={column.sortable ? {} : { cursor: 'initial' }}
            >
              {column.title}
              {column.sortable ? <SortIcon className={`sortIcon selected ${(sort.direction === SORTING_OPTIONS.desc).toString()}`} /> : null}
            </div>
          ))}
          <div />
        </div>
        <div className="auditlogs-list">
          {items.map(item => {
            const allowsExpansion = !!item.change || item.action.includes('terminal') || item.action.includes('portforward');
            return (
              <div
                className={`auditlogs-list-item ${allowsExpansion ? 'clickable' : ''}`}
                key={`event-${item.time}`}
                onClick={() => onIssueSelection(allowsExpansion ? item : undefined)}
              >
                {auditLogColumns.map((column, index) => column.render(item, index, userCapabilities))}
                {allowsExpansion ? <DetailsIndicator /> : <div />}
              </div>
            );
          })}
        </div>
        <div className="flexbox margin-top">
          <Pagination
            className="margin-top-none"
            count={count}
            rowsPerPage={perPage}
            onChangeRowsPerPage={onChangeRowsPerPage}
            page={page}
            onChangePage={onChangePage}
          />
          <Loader show={loading} small />
        </div>
        <EventDetailsDrawer eventItem={eventItem} open={Boolean(eventItem)} onClose={() => onIssueSelection()} />
      </div>
    )
  );
};

export default AuditLogsList;
