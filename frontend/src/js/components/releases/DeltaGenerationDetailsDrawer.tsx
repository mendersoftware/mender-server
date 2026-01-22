// Copyright 2024 Northern.tech AS
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
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useSelector } from 'react-redux';
import { Link } from 'react-router-dom';

import { Launch as LaunchIcon } from '@mui/icons-material';
import { Alert, Divider, Drawer, LinearProgress, tableCellClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { TwoColumnData } from '@northern.tech/common-ui/ConfigurationObject';
import { Code } from '@northern.tech/common-ui/CopyCode';
import DetailsTable from '@northern.tech/common-ui/DetailsTable';
import { DrawerTitle } from '@northern.tech/common-ui/DrawerTitle';
import FileSize from '@northern.tech/common-ui/FileSize';
import LinedHeader from '@northern.tech/common-ui/LinedHeader';
import Loader from '@northern.tech/common-ui/Loader';
import { MaybeTime } from '@northern.tech/common-ui/Time';
import storeActions from '@northern.tech/store/actions';
import { TIMEOUTS } from '@northern.tech/store/constants';
import { formatReleases, generateReleasesPath } from '@northern.tech/store/locationutils';
import { getDeltaJobById } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { getDeltaGenerationJobDetails, getDeltaGenerationJobs } from '@northern.tech/store/thunks';
import { DeltaJobDetailsItem, DeltaJobsListItem } from '@northern.tech/types/MenderTypes';
import { formatTime } from '@northern.tech/utils/helpers';
import copy from 'copy-to-clipboard';
import dayjs from 'dayjs';
import durationDayJs from 'dayjs/plugin/duration';

dayjs.extend(durationDayJs);

const { setSnackbar } = storeActions;

const useStyles = makeStyles()(theme => ({
  detailsContainer: {
    padding: theme.spacing(2),
    minWidth: '60vw'
  },
  table: {
    [`.${tableCellClasses.body}, .${tableCellClasses.head}`]: {
      paddingLeft: 0
    }
  }
}));

const deltaStateTitleMap = {
  artifact_uploaded: 'Success'
};

const deltaProgressMap = {
  success: 100,
  failed: 100,
  artifact_uploaded: 100
};

const deltaStatusColorMap = {
  artifact_uploaded: 'success',
  failed: 'secondary',
  pending: 'primary',
  success: 'success'
};

export const StatusIndicator = ({ status }) => {
  const statusKey = status?.toLowerCase();
  const statusTitle = deltaStateTitleMap[statusKey] ?? status;
  const progressColor = deltaStatusColorMap[statusKey] ?? deltaStatusColorMap.pending;
  const progressValue = deltaProgressMap[statusKey];
  return (
    <>
      <div className="capitalized-start">{statusTitle}</div>
      <LinearProgress
        className="absolute full-width"
        color={progressColor}
        style={{ bottom: 0 }}
        value={progressValue}
        variant={progressValue ? 'determinate' : 'indeterminate'}
      />
    </>
  );
};

const statusColumns = [
  {
    key: 'started',
    title: 'Started',
    cellProps: { style: { width: '15%' } },
    render: ({ started }) => <MaybeTime value={formatTime(started)} />
  },
  {
    key: 'finished',
    title: 'Finished',
    cellProps: { style: { width: '15%' } },
    render: ({ finished }) => <MaybeTime value={formatTime(finished)} />
  },
  {
    key: 'totalTime',
    title: 'Total time',
    cellProps: { style: { width: '10%' } },
    render: ({ total_time }) => {
      if (!total_time) {
        return '-';
      }
      const duration = dayjs.duration(`PT${total_time}`.toUpperCase());
      const format = duration.asSeconds() >= 60 ? 'HH:mm' : 'HH:mm:ss'; // allowing quick generations to also show something
      return duration.format(format);
    }
  },
  {
    key: 'toArtifactSize',
    title: 'Target Artifact size',
    cellProps: { style: { width: '12.5%' } },
    render: ({ target_size }) => (target_size ? <FileSize fileSize={target_size} /> : '-')
  },
  {
    key: 'deltaArtifactSize',
    title: 'Delta Artifact size',
    cellProps: { style: { width: '12.5%' } },
    render: ({ delta_artifact_size }) => (delta_artifact_size ? <FileSize fileSize={delta_artifact_size} /> : '-')
  },
  {
    key: 'dataSaved',
    title: 'Data saved',
    cellProps: { style: { width: '10%' } },
    render: ({ dataSaved }) => <FileSize fileSize={dataSaved} />
  },
  {
    key: 'status',
    title: 'Status',
    cellProps: { style: { width: '20%' } },
    render: StatusIndicator
  },
  {
    key: 'spacer',
    title: '',
    sortable: false,
    cellProps: { style: { width: '5%' } },
    render: () => ''
  }
];

type EnhancedJobDetailsItem = DeltaJobDetailsItem &
  DeltaJobsListItem & {
    finished?: string;
    fromRelease: string;
    toRelease: string;
  };

const PageLink = ({ area, target }) =>
  target ? (
    <Link className="flexbox center-aligned" to={`/${area}/${encodeURIComponent(target)}`} target="_blank">
      {target}
      <LaunchIcon className="margin-left-small link-color" fontSize="small" />
    </Link>
  ) : (
    '-'
  );

interface DeltaGenerationDetailsDrawerProps {
  jobId?: string;
  onClose: () => void;
  open: boolean;
}

export const DeltaGenerationDetailsDrawer = ({ jobId, onClose, open }: DeltaGenerationDetailsDrawerProps) => {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const dispatch = useAppDispatch();
  const deltaJob: EnhancedJobDetailsItem = useSelector(state => getDeltaJobById(state, jobId));
  const { classes } = useStyles();
  const timer = useRef<ReturnType<typeof setInterval> | undefined>();

  const refreshJobDetails = useCallback(() => {
    setIsLoading(true);
    setError(null);
    // We need to get the list too to infer the completion time
    Promise.all([dispatch(getDeltaGenerationJobs()).unwrap(), dispatch(getDeltaGenerationJobDetails(jobId)).unwrap()])
      .catch(err => setError(err.message || 'Failed to load delta generation details'))
      .finally(() => setIsLoading(false));
  }, [dispatch, jobId]);

  useEffect(() => {
    if (!jobId) {
      return;
    }
    clearInterval(timer.current);
    if ('failed' !== deltaJob?.status && 'success' !== deltaJob?.status) {
      timer.current = setInterval(refreshJobDetails, TIMEOUTS.refreshDefault);
    }
    refreshJobDetails();
  }, [deltaJob?.status, jobId, refreshJobDetails]);

  const copyLinkToClipboard = () => {
    const location = window.location.href.substring(0, window.location.href.indexOf('/releases'));
    copy(`${location}${generateReleasesPath({ pageState: { selectedRelease: '' } })}?${formatReleases({ pageState: { tab: 'delta', id: jobId } })}`);
    dispatch(setSnackbar('Link copied to clipboard'));
  };

  const combinedData: EnhancedJobDetailsItem | undefined = useMemo(() => {
    if (!deltaJob) {
      return;
    }
    const { target_size, delta_artifact_size, to_release, to_version, from_release, from_version } = deltaJob;
    const dataSaved = target_size && delta_artifact_size ? Math.max(0, target_size - delta_artifact_size) : 0;

    return {
      ...deltaJob,
      toRelease: to_release || to_version || '-',
      fromRelease: from_release || from_version || '-',
      dataSaved
    };
  }, [deltaJob]);

  if (!combinedData) {
    return null;
  }

  const staticDetailsLeft = {
    'To Release': <PageLink area="releases" target={combinedData.toRelease} />,
    'From Release': <PageLink area="releases" target={combinedData.fromRelease} />,
    'Device types compatible': combinedData.devices_types_compatible?.join(', ') || '-'
  };
  const staticDetailsRight = {
    'From deployment': <PageLink area="deployments" target={combinedData.deployment_id} />
  };

  return (
    <Drawer anchor="right" open={open} onClose={onClose}>
      <DrawerTitle
        title={
          <>
            Delta Artifact information
            <div className="margin-left-small margin-right-small">ID: {jobId}</div>
          </>
        }
        onClose={onClose}
        onLinkCopy={copyLinkToClipboard}
      />
      <Divider />
      <div className={classes.detailsContainer}>
        {error && (
          <Alert severity="error" className="margin-bottom">
            {error}
          </Alert>
        )}
        {isLoading ? (
          <Loader show={true} />
        ) : (
          <>
            <div className="two-columns">
              <TwoColumnData config={staticDetailsLeft} />
              <TwoColumnData config={staticDetailsRight} />
            </div>
            <LinedHeader className="margin-top-large" heading="Status" />
            <DetailsTable className={classes.table} columns={statusColumns} items={[combinedData]} />
            {combinedData.status === 'failed' && combinedData.details && <Code className="log">{combinedData.details}</Code>}
          </>
        )}
      </div>
    </Drawer>
  );
};

export default DeltaGenerationDetailsDrawer;
