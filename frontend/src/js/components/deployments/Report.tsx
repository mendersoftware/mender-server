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
import { useCallback, useEffect, useRef, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

// material ui
import { Block as BlockIcon, CheckCircleOutline as CheckCircleOutlineIcon, Refresh as RefreshIcon } from '@mui/icons-material';
import { Button, Divider, Drawer, Tooltip } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import ConfigurationObject from '@northern.tech/common-ui/ConfigurationObject';
import Confirm from '@northern.tech/common-ui/Confirm';
import { DrawerTitle } from '@northern.tech/common-ui/DrawerTitle';
import LinedHeader from '@northern.tech/common-ui/LinedHeader';
import LogDialog from '@northern.tech/common-ui/dialogs/Log';
import storeActions from '@northern.tech/store/actions';
import { AUDIT_LOGS_TYPES, DEPLOYMENT_STATES, DEPLOYMENT_TYPES, TIMEOUTS, deploymentStatesToSubstates, onboardingSteps } from '@northern.tech/store/constants';
import {
  getDeploymentRelease,
  getDevicesById,
  getIdAttribute,
  getOnboardingState,
  getSelectedDeploymentData,
  getTenantCapabilities,
  getUserCapabilities
} from '@northern.tech/store/selectors';
import { getAuditLogs, getDeploymentDevices, getDeviceLog, getRelease, getSingleDeployment, updateDeploymentControlMap } from '@northern.tech/store/thunks';
import { statCollector } from '@northern.tech/store/utils';
import { toggle } from '@northern.tech/utils/helpers';
import copy from 'copy-to-clipboard';

import { getOnboardingComponentFor } from '../../utils/onboardingManager';
import DeploymentStatus, { DeploymentPhaseNotification } from './deployment-report/DeploymentStatus';
import DeviceList from './deployment-report/DeviceList';
import DeploymentOverview from './deployment-report/Overview';
import RolloutSchedule from './deployment-report/RolloutSchedule';

const { setSnackbar } = storeActions;

const useStyles = makeStyles()(theme => ({
  divider: { marginTop: theme.spacing(2) },
  header: {
    ['&.dashboard-header span']: {
      backgroundColor: theme.palette.background.paper,
      backgroundImage: 'linear-gradient(rgba(255, 255, 255, 0.15), rgba(255, 255, 255, 0.15))'
    }
  }
}));

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

export const DeploymentReport = ({ abort, onClose, past, retry, type, open }) => {
  const [deviceId, setDeviceId] = useState('');
  const rolloutSchedule = useRef();
  const timer = useRef();
  const onboardingTooltipAnchor = useRef();
  const { classes } = useStyles();
  const dispatch = useDispatch();
  const { deployment, selectedDevices } = useSelector(getSelectedDeploymentData);
  const devicesById = useSelector(getDevicesById);
  const idAttribute = useSelector(getIdAttribute);
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

  const viewLog = useCallback(
    id => dispatch(getDeviceLog({ deploymentId: deployment.id, deviceId: id })).then(() => setDeviceId(id)),
    [deployment.id, dispatch]
  );

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
    } catch {
      config = {};
    }
  }

  const onUpdateControlChange = (updatedMap = {}) => {
    const { id, update_control_map = {} } = deployment;
    const { states } = update_control_map;
    const { states: updatedStates } = updatedMap;
    dispatch(updateDeploymentControlMap({ deploymentId: id, updateControlMap: { states: { ...states, ...updatedStates } } }));
  };

  const props = {
    deployment,
    getDeploymentDevices: useCallback((...args) => dispatch(getDeploymentDevices(...args)), [dispatch]),
    idAttribute,
    selectedDevices,
    userCapabilities,
    totalDeviceCount,
    viewLog
  };
  let onboardingComponent = null;
  if (!onboardingState.complete && onboardingTooltipAnchor.current && finished) {
    const anchor = {
      left: onboardingTooltipAnchor.current.offsetLeft + onboardingTooltipAnchor.current.offsetWidth + 55,
      top: onboardingTooltipAnchor.current.offsetTop + onboardingTooltipAnchor.current.offsetHeight / 2 + 15
    };
    onboardingComponent = getOnboardingComponentFor(onboardingSteps.DEPLOYMENTS_COMPLETED, onboardingState, { anchor });
  }

  return (
    <Drawer anchor="right" open={open} onClose={onClose} PaperProps={{ style: { minWidth: '75vw' } }}>
      {!!onboardingComponent && onboardingComponent}
      <DrawerTitle
        title={
          <>
            Deployment {type !== DEPLOYMENT_STATES.scheduled ? 'details' : 'report'}
            <i className="margin-left-small margin-right-small">ID: {deployment.id}</i>
          </>
        }
        onLinkCopy={copyLinkToClipboard}
        preCloser={
          !finished ? (
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
            <div className="flexbox centered margin-right" ref={onboardingTooltipAnchor}>
              <CheckCircleOutlineIcon fontSize="small" className="green margin-right-small" />
              <h3>Finished</h3>
            </div>
          )
        }
        onClose={onClose}
      />
      <Divider />
      <div>
        <DeploymentPhaseNotification deployment={deployment} onReviewClick={scrollToBottom} />
        <DeploymentOverview creator={creator} deployment={deployment} devicesById={devicesById} idAttribute={idAttribute} onScheduleClick={scrollToBottom} />
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
