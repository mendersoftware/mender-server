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
import { Pause as PauseIcon, ArrowDropDownCircleOutlined as ScrollDownIcon } from '@mui/icons-material';
import { alpha, tableCellClasses, tableRowClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import DetailsTable from '@northern.tech/common-ui/DetailsTable';
import { Link } from '@northern.tech/common-ui/Link';
import { SynchronizedTwoColumnData } from '@northern.tech/common-ui/TwoColumnData';
import { deploymentDisplayStates, pauseMap } from '@northern.tech/store/constants';
import { groupDeploymentStats, isDarkMode } from '@northern.tech/store/utils';

const useStyles = makeStyles()(theme => ({
  progressStatus: {
    backgroundColor: isDarkMode(theme.palette.mode) ? alpha(theme.palette.grey[300], theme.palette.action.selectedOpacity) : theme.palette.grey[50],
    borderRadius: theme.spacing(0.5)
  },
  statusTable: {
    [`.${tableCellClasses.root}`]: { whiteSpace: 'nowrap' },
    [`.${tableCellClasses.body}`]: { borderBottom: 'none' },
    [`.${tableRowClasses.root}.${tableRowClasses.hover}:hover`]: { backgroundColor: 'transparent' }
  },
  scrollDown: { marginLeft: theme.spacing() }
}));

const nonPhaseStates = ['failure', 'finished', 'scheduled', 'success'];

const statusColumns = [
  { key: 'status', title: 'Status', cellProps: { style: { width: '100%' } }, render: ({ statusDescription }) => statusDescription },
  { key: 'deviceCount', title: '# devices', cellProps: { align: 'right' }, render: ({ deviceCount }) => deviceCount },
  ...Object.entries(deploymentDisplayStates)
    .filter(([key]) => !nonPhaseStates.includes(key))
    .map(([key, title]) => ({ key, title, cellProps: { align: 'right' }, render: ({ phaseStats }) => phaseStats[key].toLocaleString() }))
];

export const DeploymentPhaseNotification = ({ className = '', deployment = {}, onReviewClick }) => {
  const { classes } = useStyles();
  const { paused } = groupDeploymentStats(deployment);
  if (paused === 0) {
    return null;
  }
  return (
    <div
      className={`${classes.progressStatus} flexbox align-items-center margin-bottom clickable ${className}`}
      onClick={onReviewClick}
      style={{ padding: 15 }}
    >
      <PauseIcon />
      <div className="muted">
        Deployment is <span className="uppercased">paused</span>. <Link>Review its status</Link> to continue, retry or abort the deployment{' '}
      </div>
      <ScrollDownIcon fontSize="small" className={`link-color ${classes.scrollDown}`} />
    </div>
  );
};

export const DeploymentStatus = ({ className = '', deployment = {} }) => {
  const { classes } = useStyles();
  const { finished, max_devices, retries = 0, status = 'pending', statistics = {} } = deployment;
  const { status: stats = {} } = statistics;
  const phaseStats = groupDeploymentStats(deployment, true);

  let statusDescription = (
    <>
      {deploymentDisplayStates[status]}
      {status === 'pending' ? ' (awaiting devices)' : ''}
    </>
  );
  if (finished) {
    statusDescription = <div>Finished {!!phaseStats.failures && <span className="failures">with failures</span>}</div>;
  } else if (status === 'paused' && phaseStats.paused > 0) {
    // based on the order of the possible pause states we find the furthest possible and use that as the current pause state - if applicable
    const currentPauseState = Object.keys(pauseMap)
      .reverse()
      .find(key => stats[key] > 0);
    statusDescription = (
      <>
        {deploymentDisplayStates[status]} ({pauseMap[currentPauseState].title})
      </>
    );
  }

  const statsBasedDeviceCount = Object.values(phaseStats).reduce((sum, count) => sum + count, 0);

  return (
    <>
      <div className={`padding-medium margin-bottom ${classes.progressStatus} ${className}`}>
        <DetailsTable
          className={`margin-bottom-none ${classes.statusTable}`}
          columns={statusColumns}
          items={[{ deviceCount: statsBasedDeviceCount, phaseStats, statusDescription }]}
        />
      </div>
      <SynchronizedTwoColumnData
        className="margin-bottom"
        data={{ 'Update attempts per device': retries + 1, 'Maximum number of devices': max_devices || 'N/A' }}
        style={{ gridTemplateColumns: 'max-content 1fr' }}
      />
    </>
  );
};

export default DeploymentStatus;
