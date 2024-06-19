// Copyright 2015 Northern.tech AS
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
import React, { useCallback, useEffect, useRef, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

// material ui
import {
  Block as BlockIcon,
  CheckCircleOutline as CheckCircleOutlineIcon,
  Close as CloseIcon,
  Link as LinkIcon,
  Refresh as RefreshIcon
} from '@mui/icons-material';
import { Button, Divider, Drawer, IconButton, Tooltip } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import copy from 'copy-to-clipboard';
import moment from 'moment';
import momentDurationFormatSetup from 'moment-duration-format';

import { setSnackbar } from '../../actions/appActions';
import { getDeploymentDevices, getDeviceLog, getSingleDeployment, updateDeploymentControlMap } from '../../actions/deploymentActions';
import { getAuditLogs } from '../../actions/organizationActions';
import { getRelease } from '../../actions/releaseActions';
import { TIMEOUTS } from '../../constants/appConstants';
import { DEPLOYMENT_STATES, DEPLOYMENT_TYPES, deploymentStatesToSubstates } from '../../constants/deploymentConstants';
import { AUDIT_LOGS_TYPES } from '../../constants/organizationConstants';
import { statCollector, toggle } from '../../helpers';
import {
  getDeploymentRelease,
  getDevicesById,
  getIdAttribute,
  getOnboardingState,
  getSelectedDeploymentData,
  getTenantCapabilities,
  getUserCapabilities
} from '../../selectors';
import ConfigurationObject from '../common/configurationobject';
import Confirm from '../common/confirm';
import LogDialog from '../common/dialogs/log';
import LinedHeader from '../common/lined-header';
import BaseOnboardingTip from '../helptips/baseonboardingtip.js';
import { DeploymentUploadFinished } from '../helptips/onboardingtips.js';
import DeploymentStatus, { DeploymentPhaseNotification } from './deployment-report/deploymentstatus';
import DeviceList from './deployment-report/devicelist';
import DeploymentOverview from './deployment-report/overview';
import RolloutSchedule from './deployment-report/rolloutschedule';

momentDurationFormatSetup(moment);

const useStyles = makeStyles()(theme => ({
  divider: { marginTop: theme.spacing(2) },
  header: {
    ['&.dashboard-header span']: {
      backgroundColor: theme.palette.background.paper,
      backgroundImage: 'linear-gradient(rgba(255, 255, 255, 0.15), rgba(255, 255, 255, 0.15))'
    }
  }
}));

export const defaultColumnDataProps = {
  chipLikeKey: false,
  style: { alignItems: 'center', alignSelf: 'flex-start', gridTemplateColumns: 'minmax(140px, 1fr) minmax(220px, 1fr)', maxWidth: '25vw' }
};

export const DeploymentAbortButton = ({ abort, deployment }) => {
  const [aborting, setAborting] = useState(false);

  const toggleAborting = () => setAborting(toggle);

  return aborting ? (
    <Confirm cancel={toggleAborting} action={() => abort(deployment.id)} type="abort" />
  ) : (
    <Tooltip
      title="Devices that have not yet started the deployment will not start the deployment.&#10;Devices that have already completed the deployment are not affected by the abort.&#10;Devices that are in the middle of the deployment at the time of abort will finish deployment normally, but will perform a rollback."
      placement="bottom"
    >
      <Button color="secondary" startIcon={<BlockIcon fontSize="small" />} onClick={toggleAborting}>
        {deployment.filters?.length ? 'Stop' : 'Abort'} deployment
      </Button>
    </Tooltip>
  );
};

export const DeploymentReport = ({ abort, onClose, past, retry, type }) => {
  const [deviceId, setDeviceId] = useState('');
  const rolloutSchedule = useRef();
  const timer = useRef();
  const onboardingTooltipAnchor = useRef();
  const { classes } = useStyles();
  const dispatch = useDispatch();
  const { deployment, selectedDevices } = useSelector(getSelectedDeploymentData);
  const devicesById = useSelector(getDevicesById);
  const { attribute: idAttribute } = useSelector(getIdAttribute);
  const release = useSelector(getDeploymentRelease);
  const tenantCapabilities = useSelector(getTenantCapabilities);
  const userCapabilities = useSelector(getUserCapabilities);
  const onboardingState = useSelector(getOnboardingState);
  // we can't filter by auditlog action via the api, so
  // - fall back to the following filter
  // - hope the deployment creation event is retrieved with the call to auditlogs api on report open
  // - otherwise no creator will be shown
  const { actor = {} } =
    useSelector(state =>
      state.organization.auditlog.events.find(event => event.object.id === state.deployments.selectionState.selectedId && event.action === 'create')
    ) || {};
  const creator = actor.email;

  const { canAuditlog } = userCapabilities;
  const { hasAuditlogs } = tenantCapabilities;
  const { devices = {}, device_count = 0, totalDeviceCount: totalDevices, statistics = {}, type: deploymentType } = deployment;
  const { status: stats = {} } = statistics;
  const totalDeviceCount = totalDevices ?? device_count;

  const refreshDeployment = useCallback(() => {
    if (!deployment.id) {
      return;
    }
    return dispatch(getSingleDeployment(deployment.id));
  }, [deployment.id, dispatch]);

  useEffect(() => {
    if (!deployment.id) {
      return;
    }
    clearInterval(timer.current);
    const now = new Date();
    now.setSeconds(now.getSeconds() + TIMEOUTS.refreshDefault / TIMEOUTS.oneSecond);
    if (!deployment.finished || new Date(deployment.finished) > now) {
      timer.current = past ? null : setInterval(refreshDeployment, TIMEOUTS.fiveSeconds);
    }
    if ((deployment.type === DEPLOYMENT_TYPES.software || !release.device_types_compatible.length) && deployment.artifact_name) {
      dispatch(getRelease(deployment.artifact_name));
    }
    if (hasAuditlogs && canAuditlog) {
      dispatch(
        getAuditLogs({
          page: 1,
          perPage: 100,
          startDate: undefined,
          endDate: undefined,
          user: undefined,
          type: AUDIT_LOGS_TYPES.find(item => item.value === 'deployment'),
          detail: deployment.name
        })
      );
    }
    return () => {
      clearInterval(timer.current);
    };
  }, [
    canAuditlog,
    deployment.artifact_name,
    deployment.finished,
    deployment.id,
    deployment.name,
    deployment.status,
    deployment.type,
    dispatch,
    hasAuditlogs,
    past,
    refreshDeployment,
    release.device_types_compatible.length
  ]);

  useEffect(() => {
    const progressCount =
      statCollector(deploymentStatesToSubstates.paused, stats) +
      statCollector(deploymentStatesToSubstates.pending, stats) +
      statCollector(deploymentStatesToSubstates.inprogress, stats);

    if (!!device_count && progressCount <= 0 && timer.current) {
      // if no more devices in "progress" statuses, deployment has finished, stop counter
      clearInterval(timer.current);
      timer.current = setTimeout(refreshDeployment, TIMEOUTS.oneSecond);
      return () => {
        clearTimeout(timer.current);
      };
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [deployment.id, device_count, JSON.stringify(stats), refreshDeployment]);

  const scrollToBottom = () => rolloutSchedule.current?.scrollIntoView({ behavior: 'smooth' });

  const viewLog = useCallback(id => dispatch(getDeviceLog(deployment.id, id)).then(() => setDeviceId(id)), [deployment.id, dispatch]);

  const copyLinkToClipboard = () => {
    const location = window.location.href.substring(0, window.location.href.indexOf('/deployments') + '/deployments'.length);
    copy(`${location}?open=true&id=${deployment.id}`);
    dispatch(setSnackbar('Link copied to clipboard'));
  };

  const { log: logData } = devices[deviceId] || {};
  const finished = deployment.finished || deployment.status === DEPLOYMENT_STATES.finished;
  const isConfigurationDeployment = deploymentType === DEPLOYMENT_TYPES.configuration;
  let config = {};
  if (isConfigurationDeployment) {
    try {
      config = JSON.parse(atob(deployment.configuration));
    } catch (error) {
      config = {};
    }
  }

  const onUpdateControlChange = (updatedMap = {}) => {
    const { id, update_control_map = {} } = deployment;
    const { states } = update_control_map;
    const { states: updatedStates } = updatedMap;
    dispatch(updateDeploymentControlMap(id, { states: { ...states, ...updatedStates } }));
  };

  const props = {
    deployment,
    getDeploymentDevices: useCallback((id, options) => dispatch(getDeploymentDevices(id, options)), [dispatch]),
    idAttribute,
    selectedDevices,
    userCapabilities,
    totalDeviceCount,
    viewLog
  };
  let onboardingComponent = null;
  if (!onboardingState.complete && onboardingTooltipAnchor.current && finished) {
    const anchor = {
      left: onboardingTooltipAnchor.current.offsetLeft + (onboardingTooltipAnchor.current.offsetWidth / 100) * 50,
      top: onboardingTooltipAnchor.current.offsetTop + onboardingTooltipAnchor.current.offsetHeight
    };
    onboardingComponent = (
      <BaseOnboardingTip
        id={onboardingState.progress}
        progress={onboardingState.progress}
        component={<DeploymentUploadFinished></DeploymentUploadFinished>}
        anchor={anchor}
      />
    );
  }

  return (
    <Drawer anchor="right" open onClose={onClose} PaperProps={{ style: { minWidth: '75vw' } }}>
      {!!onboardingComponent && onboardingComponent}
      <div className="flexbox margin-bottom-small space-between">
        <div className="flexbox">
          <h3>{`Deployment ${type !== DEPLOYMENT_STATES.scheduled ? 'details' : 'report'}`}</h3>
          <h4 className="margin-left-small margin-right-small">ID: {deployment.id}</h4>
          <IconButton onClick={copyLinkToClipboard} style={{ alignSelf: 'center' }} size="large">
            <LinkIcon />
          </IconButton>
        </div>
        <div className="flexbox center-aligned">
          {!finished ? (
            <DeploymentAbortButton abort={abort} deployment={deployment} />
          ) : (stats.failure || stats.aborted) && !isConfigurationDeployment ? (
            <Tooltip
              title="This will create a new deployment with the same device group and Release.&#10;Devices with this Release already installed will be skipped, all others will be updated."
              placement="bottom"
            >
              <Button color="secondary" startIcon={<RefreshIcon fontSize="small" />} onClick={() => retry(deployment, Object.keys(devices))}>
                Recreate deployment?
              </Button>
            </Tooltip>
          ) : (
            <div className="flexbox centered margin-right">
              <CheckCircleOutlineIcon fontSize="small" className="green margin-right-small" />
              <h3>Finished</h3>
            </div>
          )}
          <IconButton ref={onboardingTooltipAnchor} onClick={onClose} aria-label="close" size="large">
            <CloseIcon />
          </IconButton>
        </div>
      </div>
      <Divider />
      <div>
        <DeploymentPhaseNotification deployment={deployment} onReviewClick={scrollToBottom} />
        <DeploymentOverview
          creator={creator}
          deployment={deployment}
          devicesById={devicesById}
          onScheduleClick={scrollToBottom}
          tenantCapabilities={tenantCapabilities}
        />
        {isConfigurationDeployment && (
          <>
            <LinedHeader className={classes.header} heading="Configuration" />
            <ConfigurationObject className="margin-top-small margin-bottom-large" config={config} />
          </>
        )}
        <LinedHeader className={classes.header} heading="Status" />
        <DeploymentStatus deployment={deployment} />
        {!!totalDeviceCount && (
          <>
            <LinedHeader className={classes.header} heading="Devices" />
            <DeviceList {...props} viewLog={viewLog} />
          </>
        )}
        <RolloutSchedule
          deployment={deployment}
          headerClass={classes.header}
          onUpdateControlChange={onUpdateControlChange}
          onAbort={abort}
          innerRef={rolloutSchedule}
        />
        {Boolean(deviceId.length) && (
          <LogDialog
            context={{ device: deviceId, releaseName: deployment.artifact_name, date: deployment.finished }}
            logData={logData}
            onClose={() => setDeviceId('')}
          />
        )}
      </div>
      <Divider className={classes.divider} light />
    </Drawer>
  );
};
export default DeploymentReport;
