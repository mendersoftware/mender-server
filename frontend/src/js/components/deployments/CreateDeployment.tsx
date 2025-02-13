// Copyright 2019 Northern.tech AS
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
import React, { useEffect, useRef, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';

import { ExpandMore } from '@mui/icons-material';
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Button,
  Checkbox,
  Divider,
  Drawer,
  FormControlLabel,
  FormGroup,
  Typography,
  accordionClasses
} from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import Confirm from '@northern.tech/common-ui/Confirm';
import { DrawerTitle } from '@northern.tech/common-ui/DrawerTitle';
import { ALL_DEVICES, onboardingSteps } from '@northern.tech/store/constants';
import {
  getAcceptedDevices,
  getDeviceCountsByStatus,
  getDevicesById,
  getFeatures,
  getGlobalSettings,
  getGroupNames,
  getGroupsByIdWithoutUngrouped,
  getIdAttribute,
  getIsEnterprise,
  getOnboardingState,
  getReleaseListState,
  getReleasesById,
  getTenantCapabilities
} from '@northern.tech/store/selectors';
import { advanceOnboarding, createDeployment, getDeploymentsConfig, getGroupDevices, getRelease, getReleases } from '@northern.tech/store/thunks';
import { toggle } from '@northern.tech/utils/helpers';
import pluralize from 'pluralize';

import DeltaIcon from '../../../assets/img/deltaicon.svg';
import { getOnboardingComponentFor } from '../../utils/onboardingManager';
import DeviceLimit from './deployment-wizard/DeviceLimit';
import { RolloutPatternSelection, getPhaseStartTime, validatePhases } from './deployment-wizard/PhaseSettings';
import { ForceDeploy, Retries, RolloutOptions } from './deployment-wizard/RolloutOptions';
import { ScheduleRollout } from './deployment-wizard/ScheduleRollout';
import { Devices, ReleasesWarning, Software } from './deployment-wizard/SoftwareDevices';

const useStyles = makeStyles()(theme => ({
  accordion: {
    backgroundColor: theme.palette.grey[400],
    marginTop: theme.spacing(4),
    '&:before': {
      display: 'none'
    },
    [`&.${accordionClasses.expanded}`]: {
      margin: 'unset',
      marginTop: theme.spacing(4)
    }
  },
  columns: {
    columnGap: 30,
    display: 'grid',
    gridTemplateColumns: 'max-content max-content',
    '&>p': {
      marginTop: theme.spacing(3)
    }
  },
  disabled: { color: theme.palette.text.disabled }
}));

const getAnchor = (element, heightAdjustment = 3) => ({
  top: element.offsetTop + element.offsetHeight / heightAdjustment,
  left: element.offsetLeft + element.offsetWidth
});

export const CreateDeployment = props => {
  const { deploymentObject = {}, onDismiss, onScheduleSubmit, setDeploymentSettings } = props;

  const { canRetry, canSchedule, hasFullFiltering } = useSelector(getTenantCapabilities);
  const { isHosted } = useSelector(getFeatures);
  const { createdGroup, groups, hasDynamicGroups } = useSelector(state => {
    const groups = getGroupsByIdWithoutUngrouped(state);
    const createdGroup = Object.keys(groups).length ? Object.keys(groups)[0] : undefined;
    const hasDynamicGroups = Object.values(groups).some(group => !!group.id);
    return { createdGroup, hasDynamicGroups, groups };
  });
  const { hasDelta: hasDeltaEnabled } = useSelector(state => state.deployments.config) ?? {};
  const { total: acceptedDeviceCount } = useSelector(getAcceptedDevices);
  const hasDevices = !!acceptedDeviceCount;
  const devicesById = useSelector(getDevicesById);
  const { pending: hasPending } = useSelector(getDeviceCountsByStatus);
  const idAttribute = useSelector(getIdAttribute);
  const isEnterprise = useSelector(getIsEnterprise);
  const { needsDeploymentConfirmation: needsCheck, previousPhases = [], retries: previousRetries = 0 } = useSelector(getGlobalSettings);
  const onboardingState = useSelector(getOnboardingState) || {};
  const { complete: isOnboardingComplete } = onboardingState;
  const { searchedIds: releases } = useSelector(getReleaseListState);
  const releasesById = useSelector(getReleasesById);
  const groupNames = useSelector(getGroupNames);
  const dispatch = useDispatch();

  const isCreating = useRef(false);
  const [hasNewRetryDefault, setHasNewRetryDefault] = useState(false);
  const [isChecking, setIsChecking] = useState(false);
  const [isExpanded, setIsExpanded] = useState(false);
  const navigate = useNavigate();
  const releaseRef = useRef();
  const groupRef = useRef();
  const deploymentAnchor = useRef();
  const { classes } = useStyles();

  useEffect(() => {
    dispatch(getReleases({ page: 1, perPage: 100, searchOnly: true }));
  }, [dispatch]);

  useEffect(() => {
    if (isHosted || isEnterprise) {
      dispatch(getDeploymentsConfig());
    }
  }, [dispatch, isEnterprise, isHosted]);

  useEffect(() => {
    const { devices = [], group, release } = deploymentObject;
    if (release) {
      dispatch(advanceOnboarding(onboardingSteps.SCHEDULING_ARTIFACT_SELECTION));
      dispatch(getRelease(release.name));
    }
    dispatch(advanceOnboarding(onboardingSteps.SCHEDULING_GROUP_SELECTION));
    let nextDeploymentObject = { deploymentDeviceCount: devices.length ? devices.length : 0 };
    if (group === ALL_DEVICES) {
      dispatch(advanceOnboarding(onboardingSteps.SCHEDULING_ALL_DEVICES_SELECTION));
      nextDeploymentObject.deploymentDeviceCount = acceptedDeviceCount;
    }
    if (groups[group]) {
      dispatch(getGroupDevices({ group, perPage: 1 }))
        .unwrap()
        .then(
          ({
            payload: {
              group: { total: deploymentDeviceCount }
            }
          }) => setDeploymentSettings({ deploymentDeviceCount })
        );
    }
    setDeploymentSettings(nextDeploymentObject);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [acceptedDeviceCount, deploymentObject.group, deploymentObject.release?.name, dispatch, JSON.stringify(groups), setDeploymentSettings]);

  useEffect(() => {
    let { deploymentDeviceCount: deviceCount, deploymentDeviceIds: deviceIds = [], devices = [] } = deploymentObject;
    if (devices.length) {
      deviceIds = devices.map(({ id }) => id);
      deviceCount = deviceIds.length;
      devices = devices.map(({ id }) => ({ id, ...(devicesById[id] ?? {}) }));
    } else if (deploymentObject.group === ALL_DEVICES) {
      deviceCount = acceptedDeviceCount;
    }
    setDeploymentSettings({ deploymentDeviceIds: deviceIds, deploymentDeviceCount: deviceCount, devices });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [acceptedDeviceCount, JSON.stringify(deploymentObject), JSON.stringify(devicesById), setDeploymentSettings]);

  const cleanUpDeploymentsStatus = () => {
    if (!window.location.search) {
      return;
    }
    const location = window.location.pathname.slice('/ui'.length);
    navigate(location); // lgtm [js/client-side-unvalidated-url-redirection]
  };

  const onSaveRetriesSetting = hasNewRetryDefault => setHasNewRetryDefault(hasNewRetryDefault);

  const closeWizard = () => {
    cleanUpDeploymentsStatus();
    onDismiss();
  };

  const onDeltaToggle = ({ target: { checked } }) => setDeploymentSettings({ delta: checked });

  const onScheduleSubmitClick = settings => {
    if (needsCheck && !isChecking) {
      return setIsChecking(true);
    }
    isCreating.current = true;
    const { delta, deploymentDeviceIds, devices, filter, forceDeploy = false, group, maxDevices, phases, release, retries, update_control_map } = settings;
    const startTime = phases?.length ? phases[0].start_ts : undefined;
    const retrySetting = canRetry && retries ? { retries } : {};
    const newDeployment = {
      artifact_name: release.name,
      autogenerate_delta: delta,
      devices: (filter || group) && !devices.length ? undefined : deploymentDeviceIds,
      filter_id: filter?.id,
      all_devices: !filter && group === ALL_DEVICES,
      group: group === ALL_DEVICES || devices.length ? undefined : group,
      max_devices: maxDevices ? maxDevices : undefined,
      name: devices[0]?.id || (group ? decodeURIComponent(group) : ALL_DEVICES),
      phases: phases
        ? phases.map((phase, i, origPhases) => {
            phase.start_ts = getPhaseStartTime(origPhases, i, startTime);
            return phase;
          })
        : phases,
      ...retrySetting,
      force_installation: forceDeploy,
      update_control_map
    };
    if (!isOnboardingComplete) {
      dispatch(advanceOnboarding(onboardingSteps.SCHEDULING_RELEASE_TO_DEVICES));
    }
    return dispatch(createDeployment({ newDeployment, hasNewRetryDefault }))
      .then(() => {
        // successfully retrieved new deployment
        cleanUpDeploymentsStatus();
        onScheduleSubmit();
      })
      .finally(() => {
        isCreating.current = false;
        setIsChecking(false);
      });
  };

  const { delta, deploymentDeviceCount, group, phases } = deploymentObject;

  const deploymentSettings = {
    ...deploymentObject,
    filter: groups[group]?.id ? groups[group] : undefined
  };
  const disabled =
    isCreating.current ||
    !(deploymentSettings.release && (deploymentSettings.deploymentDeviceCount || !!deploymentSettings.filter || deploymentSettings.group)) ||
    !validatePhases(phases, deploymentSettings.deploymentDeviceCount);

  const sharedProps = {
    ...props,
    canRetry,
    canSchedule,
    groupNames,
    groupRef,
    groups,
    hasDevices,
    hasDynamicGroups,
    hasFullFiltering,
    hasPending,
    idAttribute,
    isEnterprise,
    previousPhases,
    previousRetries,
    releaseRef,
    releases,
    releasesById,
    commonClasses: classes,
    deploymentObject: deploymentSettings,
    hasNewRetryDefault,
    onSaveRetriesSetting,
    open: false,
    setDeploymentSettings
  };
  const hasReleases = !!Object.keys(releasesById).length;
  return (
    <Drawer anchor="right" open onClose={closeWizard} PaperProps={{ style: { minWidth: '50vw' } }}>
      <DrawerTitle title="Create a deployment" onClose={closeWizard} />
      <Divider className="margin-bottom" />
      <FormGroup>
        {!hasReleases ? (
          <ReleasesWarning />
        ) : (
          <>
            <Devices {...sharedProps} groupRef={groupRef} />
            <Software {...sharedProps} releaseRef={releaseRef} />
          </>
        )}
        <ScheduleRollout {...sharedProps} />
        <Accordion className={classes.accordion} square expanded={isExpanded} onChange={() => setIsExpanded(toggle)}>
          <AccordionSummary expandIcon={<ExpandMore />}>
            <Typography className={classes.disabled} variant="subtitle2">
              {isExpanded ? 'Hide' : 'Show'} advanced options
            </Typography>
          </AccordionSummary>
          <AccordionDetails>
            <DeviceLimit {...sharedProps} />
            <RolloutPatternSelection {...sharedProps} />
            <RolloutOptions {...sharedProps} />
            <Retries {...sharedProps} />
            <ForceDeploy {...sharedProps} />
            {hasDeltaEnabled && (
              <FormControlLabel
                control={<Checkbox color="primary" checked={delta} onChange={onDeltaToggle} size="small" />}
                label={
                  <>
                    Generate and deploy Delta Artifacts (where available) <DeltaIcon />
                  </>
                }
              />
            )}
          </AccordionDetails>
        </Accordion>
      </FormGroup>
      <div className="margin-top">
        {isChecking && (
          <Confirm
            classes="confirmation-overlay"
            cancel={() => setIsChecking(false)}
            action={() => onScheduleSubmitClick(deploymentSettings)}
            message={`This will deploy ${deploymentSettings.release?.name} to ${deploymentDeviceCount} ${pluralize(
              'device',
              deploymentDeviceCount
            )}. Are you sure?`}
            style={{ paddingLeft: 12, justifyContent: 'flex-start', maxHeight: 44 }}
          />
        )}
        <Button onClick={closeWizard} style={{ marginRight: 10 }}>
          Cancel
        </Button>
        <Button variant="contained" color="primary" ref={deploymentAnchor} disabled={disabled} onClick={() => onScheduleSubmitClick(deploymentSettings)}>
          Create deployment
        </Button>
      </div>
      <OnboardingComponent
        releaseRef={releaseRef}
        groupRef={groupRef}
        deploymentObject={deploymentObject}
        deploymentAnchor={deploymentAnchor}
        onboardingState={onboardingState}
        createdGroup={createdGroup}
        releasesById={releasesById}
        releases={releases}
        hasDevices={hasDevices}
      />
    </Drawer>
  );
};

export default CreateDeployment;

const OnboardingComponent = ({
  releaseRef,
  groupRef,
  deploymentAnchor,
  deploymentObject,
  onboardingState,
  createdGroup,
  releasesById,
  releases,
  hasDevices
}) => {
  const { deploymentDeviceCount, devices, group, release: deploymentRelease = null } = deploymentObject;

  let onboardingComponent = null;
  if (releaseRef.current && groupRef.current && deploymentAnchor.current) {
    const anchor = getAnchor(releaseRef.current);
    const groupAnchor = getAnchor(groupRef.current);
    onboardingComponent = getOnboardingComponentFor(onboardingSteps.SCHEDULING_ALL_DEVICES_SELECTION, onboardingState, { anchor: groupAnchor, place: 'right' });
    if (createdGroup) {
      onboardingComponent = getOnboardingComponentFor(
        onboardingSteps.SCHEDULING_GROUP_SELECTION,
        { ...onboardingState, createdGroup },
        { anchor: groupAnchor, place: 'right' },
        onboardingComponent
      );
    }
    if (deploymentDeviceCount && !deploymentRelease) {
      onboardingComponent = getOnboardingComponentFor(
        onboardingSteps.SCHEDULING_ARTIFACT_SELECTION,
        { ...onboardingState, selectedRelease: releasesById[releases[0]] || {} },
        { anchor, place: 'right' },
        onboardingComponent
      );
    }
    if (hasDevices && (deploymentDeviceCount || devices?.length) && deploymentRelease) {
      const buttonAnchor = getAnchor(deploymentAnchor.current, 2);
      onboardingComponent = getOnboardingComponentFor(
        onboardingSteps.SCHEDULING_RELEASE_TO_DEVICES,
        { ...onboardingState, selectedDevice: devices.length ? devices[0] : undefined, selectedGroup: group, selectedRelease: deploymentRelease },
        { anchor: buttonAnchor, place: 'right' },
        onboardingComponent
      );
    }
  }
  return onboardingComponent;
};
