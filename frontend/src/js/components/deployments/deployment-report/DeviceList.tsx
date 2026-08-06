// Copyright 2017 Northern.tech AS
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
import { useEffect, useState } from 'react';

// material ui
import type { SvgIconComponent } from '@mui/icons-material';
import {
  AutoAwesomeOutlined as AutoAwesomeIcon,
  Cancel as CancelIcon,
  CheckCircle as CheckIcon,
  Error as ErrorIcon,
  PendingOutlined as PendingIcon
} from '@mui/icons-material';
import type { SvgIconOwnProps } from '@mui/material';
import { Button, LinearProgress, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import DetailsTable from '@northern.tech/common-ui/DetailsTable';
import DeviceIdentityDisplay from '@northern.tech/common-ui/DeviceIdentity';
import FileSize from '@northern.tech/common-ui/FileSize';
import { Link } from '@northern.tech/common-ui/Link';
import Loader from '@northern.tech/common-ui/Loader';
import Pagination from '@northern.tech/common-ui/Pagination';
import { MaybeTime } from '@northern.tech/common-ui/Time';
import { TwoColumnData } from '@northern.tech/common-ui/TwoColumnData';
import MenderTooltip from '@northern.tech/common-ui/helptips/MenderTooltip';
import {
  DEVICE_LIST_DEFAULTS,
  canAccess as canShow,
  deploymentStatesToSubstates,
  rootfsImageVersion as rootfsImageVersionAttribute
} from '@northern.tech/store/constants';
import { generateReleasesPath } from '@northern.tech/store/locationutils';
import { formatTime } from '@northern.tech/utils/helpers';

import DeltaIcon from '../../../../assets/img/deltaicon.svg';

const useStyles = makeStyles()(() => ({
  table: { minHeight: '10vh', maxHeight: '40vh', overflowX: 'auto' },
  totalSize: { ['& > div.two-columns.column-data']: { gridTemplateColumns: 'max-content max-content' } }
}));

const { page: defaultPage } = DEVICE_LIST_DEFAULTS;

const statusColorMap: Record<string, SvgIconOwnProps['color']> = {
  error: 'error',
  aborted: 'secondary',
  success: 'success',
  default: 'info'
};

type StateInfoEntry = {
  color?: SvgIconOwnProps['color'];
  icon?: SvgIconComponent;
  title: string;
};

const stateInfoMap: Record<string, StateInfoEntry> = {
  pending: { title: 'Pending', icon: PendingIcon },
  decommissioned: { title: 'Decommissioned', icon: PendingIcon },
  'already-installed': { title: 'Already installed', icon: PendingIcon },
  'pause_before_committing': { title: 'Paused before committing', icon: PendingIcon },
  'pause_before_installing': { title: 'Paused before installing', icon: PendingIcon },
  'pause_before_rebooting': { title: 'Paused before rebooting', icon: PendingIcon },
  aborted: { title: 'Paused before committing', color: statusColorMap.aborted, icon: CancelIcon },
  failure: { title: 'Fail', color: statusColorMap.error, icon: ErrorIcon },
  noartifact: { title: 'No compatible artifact found', icon: CancelIcon },
  artifact_too_big: { title: 'Skipped', icon: CancelIcon },
  incompatible_tier: { title: 'Incompatible tier', icon: CancelIcon },
  success: { title: 'Success', color: statusColorMap.success, icon: CheckIcon }
};

const deviceListColumns = [
  {
    key: 'idAttribute',
    title: 'id',
    renderTitle: ({ idAttribute }) => idAttribute.attribute,
    render: ({ device }) => (
      <Link to={`/devices?id=${device.id}`}>
        <DeviceIdentityDisplay device={device} isEditable={false} />
      </Link>
    ),
    canShow
  },
  {
    key: 'device-type',
    title: 'Device Type',
    render: ({ device }) => {
      const { attributes = {} } = device;
      const { device_type: deviceTypes = [] } = attributes;
      return deviceTypes.length ? deviceTypes.join(',') : '-';
    },
    canShow
  },
  {
    key: 'current-artifact',
    title: 'Current artifact',
    render: ({ device: { attributes = {} }, userCapabilities: { canReadReleases } }) => {
      const { artifact_name: softwareName } = attributes;
      return softwareName ? (
        canReadReleases ? (
          <Link to={generateReleasesPath({ pageState: { selectedRelease: softwareName } })}>{softwareName}</Link>
        ) : (
          softwareName
        )
      ) : (
        '-'
      );
    },
    canShow
  },
  {
    key: 'current-software',
    title: 'Current software',
    render: ({ device: { attributes = {} }, userCapabilities: { canReadReleases } }) => {
      const { [rootfsImageVersionAttribute]: rootfsImageVersion } = attributes;
      const softwareName = rootfsImageVersion;
      return softwareName ? (
        canReadReleases ? (
          <Link style={{ fontWeight: 'initial' }} href={generateReleasesPath({ pageState: { selectedRelease: softwareName } })}>
            {softwareName}
          </Link>
        ) : (
          softwareName
        )
      ) : (
        '-'
      );
    },
    canShow
  },
  { key: 'started', title: 'Started', render: ({ device: { started } }) => <MaybeTime value={formatTime(started)} />, sortable: false, canShow },
  { key: 'finished', title: 'Finished', render: ({ device: { finished } }) => <MaybeTime value={formatTime(finished)} />, sortable: false, canShow },
  {
    key: 'artifact_size',
    title: 'Artifact size',
    render: ({ device: { image = {} } }) => {
      const { size } = image;
      return <FileSize fileSize={size} />;
    },
    sortable: false,
    canShow
  },
  {
    key: 'delta',
    title: '',
    render: ({ device: { isDelta } }) =>
      isDelta ? (
        <MenderTooltip placement="bottom" title="Device is enabled for delta updates">
          <DeltaIcon />
        </MenderTooltip>
      ) : (
        ''
      ),
    canShow
  },
  {
    key: 'attempts',
    title: 'Attempts',
    render: ({ device: { attempts, retries } }) => `${attempts || 1} / ${retries + 1}`,
    canShow: ({ deployment: { retries } }) => !!retries
  },
  {
    key: 'status',
    title: 'Deployment status',
    render: ({ device: { substate = '', status = '' } }) => {
      const { color: progressColor = statusColorMap.default, icon: Icon, title = status } = stateInfoMap[status] ?? {};
      const statusTitle = (
        <Typography variant="body2" className="capitalized-start">
          {title}
        </Typography>
      );
      return (
        <>
          <div className="flexbox align-items-center">
            {Icon && <Icon className="margin-right-x-small" color={progressColor} />}
            {statusTitle}
          </div>
          {substate && <Typography variant="caption">{substate}</Typography>}
          {deploymentStatesToSubstates.inprogress.includes(status.toLowerCase()) && <LinearProgress variant="indeterminate" />}
        </>
      );
    },
    canShow
  },
  {
    key: 'log',
    title: '',
    render: ({ canAi, device: { id, log }, viewLog }) =>
      log ? (
        <Button className="nowrap" color="info" variant="outlined" endIcon={canAi ? <AutoAwesomeIcon /> : null} onClick={() => viewLog(id)} size="small">
          View log
        </Button>
      ) : null,
    canShow
  }
];

export const DeploymentDeviceList = ({ canAi, deployment, getDeploymentDevices, idAttribute, selectedDevices, userCapabilities, viewLog }) => {
  const [currentPage, setCurrentPage] = useState(defaultPage);
  const [isLoading, setIsLoading] = useState(false);
  const [perPage, setPerPage] = useState(10);
  const { device_count = 0, totalDeviceCount: totalDevices, statistics = {} } = deployment;
  const totalSize = statistics.total_size ?? 0;
  const totalDeviceCount = totalDevices ?? device_count;
  const { classes } = useStyles();

  useEffect(() => {
    setCurrentPage(defaultPage);
  }, [perPage]);

  useEffect(() => {
    if (!deployment.id) {
      return;
    }
    setIsLoading(true);
    getDeploymentDevices({ id: deployment.id, page: currentPage, perPage }).then(() => setIsLoading(false));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentPage, deployment.id, deployment.status, getDeploymentDevices, JSON.stringify(statistics.status), perPage]);

  const columns = deviceListColumns.filter(column => column.canShow({ deployment })).map(column => ({ ...column, extras: { idAttribute } }));
  const items = selectedDevices.map(device => ({ canAi, device, id: device.id, idAttribute, userCapabilities, viewLog }));
  return (
    <>
      <DetailsTable className={classes.table} columns={columns} items={items} />
      <div className="flexbox space-between align-items-center">
        <div className="flexbox">
          <Pagination
            className="margin-top-none"
            count={totalDeviceCount}
            rowsPerPage={perPage}
            onChangePage={setCurrentPage}
            onChangeRowsPerPage={setPerPage}
            page={currentPage}
          />
          <Loader show={isLoading} small />
        </div>
        <div className={classes.totalSize}>
          <TwoColumnData data={{ 'Total download size': <FileSize fileSize={totalSize} /> }} />
        </div>
      </div>
    </>
  );
};

export default DeploymentDeviceList;
