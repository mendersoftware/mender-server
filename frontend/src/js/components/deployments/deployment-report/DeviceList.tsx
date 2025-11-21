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
import { Link } from 'react-router-dom';

// material ui
import {
  AutoAwesomeOutlined as AutoAwesomeIcon,
  Cancel as CancelIcon,
  CheckCircle as CheckIcon,
  Error as ErrorIcon,
  PendingOutlined as PendingIcon,
  SvgIconComponent
} from '@mui/icons-material';
import { Button, LinearProgress, LinearProgressProps, SvgIconOwnProps, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { TwoColumns } from '@northern.tech/common-ui/ConfigurationObject';
import DetailsTable from '@northern.tech/common-ui/DetailsTable';
import DeviceIdentityDisplay from '@northern.tech/common-ui/DeviceIdentity';
import FileSize from '@northern.tech/common-ui/FileSize';
import Loader from '@northern.tech/common-ui/Loader';
import Pagination from '@northern.tech/common-ui/Pagination';
import { MaybeTime } from '@northern.tech/common-ui/Time';
import MenderTooltip from '@northern.tech/common-ui/helptips/MenderTooltip';
import {
  DEVICE_LIST_DEFAULTS,
  canAccess as canShow,
  deploymentSubstates,
  rootfsImageVersion as rootfsImageVersionAttribute
} from '@northern.tech/store/constants';
import { formatTime } from '@northern.tech/utils/helpers';

import DeltaIcon from '../../../../assets/img/deltaicon.svg';

const useStyles = makeStyles()(() => ({
  table: { minHeight: '10vh', maxHeight: '40vh', overflowX: 'auto' }
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
  progress?: number;
  title: string;
};

const stateInfoMap: Record<string, StateInfoEntry> = {
  'already-installed': { title: 'Already installed', progress: 100 },
  'pause-before-committing': { title: 'Paused before committing' },
  'pause-before-installing': { title: 'Paused before installing' },
  'pause-before-rebooting': { title: 'Paused before rebooting' },
  aborted: { title: 'Paused before committing', progress: 100, color: statusColorMap.aborted, icon: CancelIcon },
  failure: { title: 'Fail', progress: 100, color: statusColorMap.error, icon: ErrorIcon },
  noartifact: { title: 'No compatible artifact found', progress: 0, icon: CancelIcon },
  success: { title: 'Success', progress: 100, color: statusColorMap.success, icon: CheckIcon }
};

const undefinedStates = [deploymentSubstates.pending, deploymentSubstates.decommissioned, deploymentSubstates.alreadyInstalled];

const deviceListColumns = [
  {
    key: 'idAttribute',
    title: 'id',
    renderTitle: ({ idAttribute }) => idAttribute.attribute,
    render: ({ device }) => (
      <Link style={{ fontWeight: 'initial' }} to={`/devices?id=${device.id}`}>
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
      const { artifact_name } = attributes;
      const softwareName = artifact_name;
      const encodedArtifactName = encodeURIComponent(softwareName);
      return softwareName ? (
        canReadReleases ? (
          <Link style={{ fontWeight: 'initial' }} to={`/releases/${encodedArtifactName}`}>
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
  {
    key: 'current-software',
    title: 'Current software',
    render: ({ device: { attributes = {} }, userCapabilities: { canReadReleases } }) => {
      const { [rootfsImageVersionAttribute]: rootfsImageVersion } = attributes;
      const softwareName = rootfsImageVersion;
      const encodedArtifactName = encodeURIComponent(softwareName);
      return softwareName ? (
        canReadReleases ? (
          <Link style={{ fontWeight: 'initial' }} to={`/releases/${encodedArtifactName}`}>
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
      const {
        color: progressColor = statusColorMap.default,
        icon: Icon = PendingIcon,
        progress: devicePercentage,
        title = status
      } = stateInfoMap[status] ?? {};
      const statusTitle = (
        <Typography variant="body2" className="capitalized-start">
          {title}
        </Typography>
      );
      return (
        <>
          <div className="flexbox center-aligned margin-bottom-x-small">
            <Icon className="margin-right-x-small" color={progressColor} />
            {statusTitle}
          </div>
          {substate && <Typography variant="caption">{substate}</Typography>}
          {!undefinedStates.includes(status.toLowerCase()) && (
            <div style={{ position: 'absolute', bottom: 0, width: '100%', paddingRight: 32 }}>
              <LinearProgress
                color={progressColor as LinearProgressProps['color']}
                value={devicePercentage}
                variant={devicePercentage !== undefined ? 'determinate' : 'indeterminate'}
              />
            </div>
          )}
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
        <Button className="nowrap" endIcon={canAi ? <AutoAwesomeIcon /> : null} onClick={() => viewLog(id)} size="small">
          View log
        </Button>
      ) : null,
    canShow
  }
];

const ValueFileSize = ({ value, ...props }) => <FileSize fileSize={value} {...props} />;

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

  const columns = deviceListColumns.reduce((accu, column) => (column.canShow({ deployment }) ? [...accu, { ...column, extras: { idAttribute } }] : accu), []);
  const items = selectedDevices.map(device => ({ canAi, device, id: device.id, idAttribute, userCapabilities, viewLog }));
  return (
    <>
      <DetailsTable className={classes.table} columns={columns} items={items} />
      <div className="flexbox space-between center-aligned margin-top">
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
        <TwoColumns compact items={{ 'Total download size': totalSize }} ValueComponent={ValueFileSize} />
      </div>
    </>
  );
};

export default DeploymentDeviceList;
