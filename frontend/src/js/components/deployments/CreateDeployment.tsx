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
import { useEffect, useRef, useState } from 'react';
import { FormProvider, useForm } from 'react-hook-form';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';

import { ExpandMore } from '@mui/icons-material';
import { Accordion, AccordionDetails, AccordionSummary, Button, FormGroup, Typography, accordionClasses, lighten } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import BaseDrawer from '@northern.tech/common-ui/BaseDrawer';
import Confirm from '@northern.tech/common-ui/Confirm';
import { FormCheckbox } from '@northern.tech/common-ui/forms/FormCheckbox';
import { ALL_DEVICES, onboardingSteps } from '@northern.tech/store/constants';
import {
  getDeviceCountsByStatus,
  getDevicesById,
  getFeatures,
  getGlobalSettings,
  getGroupData,
  getGroupNames,
  getIdAttribute,
  getIsEnterprise,
  getOnboardingState,
  getReleaseListState,
  getReleasesById,
  getTenantCapabilities
} from '@northern.tech/store/selectors';
import { advanceOnboarding, createDeployment, getDeploymentsConfig, getGroupDevices, getRelease, getReleases } from '@northern.tech/store/thunks';
import { isEmpty, toggle } from '@northern.tech/utils/helpers';
import pluralize from 'pluralize';

import { getOnboardingComponentFor } from '../../utils/onboardingManager';
import DeviceLimit from './deployment-wizard/DeviceLimit';
import { RolloutPatternSelection, getPhaseStartTime, validatePhases } from './deployment-wizard/PhaseSettings';
import { ForceDeploy, Retries, RolloutOptions } from './deployment-wizard/RolloutOptions';
import { ScheduleRollout } from './deployment-wizard/ScheduleRollout';
import { Devices, ReleasesWarning, Software } from './deployment-wizard/SoftwareDevices';
import type { DeploymentFormValues } from './deployment-wizard/types';
import { deploymentFormSections } from './deployment-wizard/utils';

const useStyles = makeStyles()(theme => ({
  accordion: {
    backgroundColor: lighten(theme.palette.background.paper, 0.25),
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

export const defaultValues = {
  group: null,
  release: null,
  delta: false,
  forceDeploy: false,
  maxDevices: 0,
  retries: 1,
  phases: [],
  update_control_map: { states: {} }
};

export const CreateDeployment = props => {
  const { deploymentObject = {}, onDismiss, onScheduleSubmit, onValuesChange, open } = props;

  const { canRetry, canSchedule } = useSelector(getTenantCapabilities);
  const { isHosted } = useSelector(getFeatures);
  const { createdGroup, groups, hasDynamicGroups } = useSelector(getGroupData);
  const { hasDelta: hasDeltaEnabled } = useSelector(state => state.deployments.config) ?? {};
  const devicesById = useSelector(getDevicesById);
  const { accepted: acceptedDeviceCount, pending: hasPending } = useSelector(getDeviceCountsByStatus);
  const hasDevices = !!acceptedDeviceCount;
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
  const methods = useForm<DeploymentFormValues>({ mode: 'onChange', defaultValues });
  const { reset, watch } = methods;
  const formValues = watch();
  const [derivedState, setDerivedState] = useState({
    deploymentDeviceCount: deploymentObject.deploymentDeviceCount ?? 0,
    deploymentDeviceIds: deploymentObject.deploymentDeviceIds ?? [],
    devices: deploymentObject.devices ?? []
  });

  useEffect(() => {
    dispatch(getReleases({ page: 1, perPage: 100, searchOnly: true, searchTerm: '', selectedTags: [], type: '' }));
  }, [dispatch]);

  useEffect(() => {
    if (isHosted || isEnterprise) {
      dispatch(getDeploymentsConfig());
    }
  }, [dispatch, isEnterprise, isHosted]);

  const { group, phases, release } = formValues;
  useEffect(() => {
    if (open) {
      reset({
        group: deploymentObject.group ?? defaultValues.group,
        release: deploymentObject.release ?? defaultValues.release,
        delta: deploymentObject.delta ?? defaultValues.delta,
        forceDeploy: deploymentObject.forceDeploy ?? defaultValues.forceDeploy,
        maxDevices: deploymentObject.maxDevices ?? defaultValues.maxDevices,
        retries: (deploymentObject.retries ?? previousRetries ?? 0) + 1,
        phases: deploymentObject.phases ?? defaultValues.phases,
        update_control_map: deploymentObject.update_control_map ?? defaultValues.update_control_map
      });
      setDerivedState({
        deploymentDeviceCount: deploymentObject.deploymentDeviceCount ?? 0,
        deploymentDeviceIds: deploymentObject.deploymentDeviceIds ?? [],
        devices: deploymentObject.devices ?? []
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, reset]);

  // Notify parent of form value changes for URL param sync
  useEffect(() => {
    onValuesChange?.({
      group: formValues.group,
      release: formValues.release,
      delta: formValues.delta,
      forceDeploy: formValues.forceDeploy,
      maxDevices: formValues.maxDevices,
      retries: formValues.retries - 1,
      phases: formValues.phases,
      update_control_map: formValues.update_control_map
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(formValues), onValuesChange]);

  useEffect(() => {
    const { devices = [] } = deploymentObject;
    if (release) {
      dispatch(advanceOnboarding(onboardingSteps.SCHEDULING_ARTIFACT_SELECTION));
      dispatch(getRelease(release.name));
    }
    dispatch(advanceOnboarding(onboardingSteps.SCHEDULING_GROUP_SELECTION));
    let deploymentDeviceCount = devices.length ? devices.length : 0;
    if (group === ALL_DEVICES) {
      dispatch(advanceOnboarding(onboardingSteps.SCHEDULING_ALL_DEVICES_SELECTION));
      deploymentDeviceCount = acceptedDeviceCount;
    }
    if (groups[group]) {
      dispatch(getGroupDevices({ group, perPage: 1 }))
        .unwrap()
        .then(
          ({
            payload: {
              group: { total: deploymentDeviceCount }
            }
          }) => setDerivedState(prev => ({ ...prev, deploymentDeviceCount }))
        );
    }
    setDerivedState(prev => ({ ...prev, deploymentDeviceCount }));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [acceptedDeviceCount, group, release?.name, dispatch, JSON.stringify(groups)]);

  // Enrich devices from Redux store
  useEffect(() => {
    let { deploymentDeviceIds: deviceIds = [], devices = [] } = deploymentObject;
    let deviceCount = derivedState.deploymentDeviceCount;
    if (devices.length) {
      deviceIds = devices.map(({ id }) => id);
      deviceCount = deviceIds.length;
      devices = devices.map(({ id }) => ({ id, ...(devicesById[id] ?? {}) }));
    } else if (group === ALL_DEVICES) {
      deviceCount = acceptedDeviceCount;
    }
    setDerivedState(prev => ({ ...prev, deploymentDeviceIds: deviceIds, deploymentDeviceCount: deviceCount, devices }));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [acceptedDeviceCount, group, JSON.stringify(deploymentObject.devices), JSON.stringify(devicesById)]);

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

  const onScheduleSubmitClick = settings => {
    if (needsCheck && !isChecking) {
      return setIsChecking(true);
    }
    isCreating.current = true;
    const { delta, forceDeploy = false, maxDevices, phases, release, update_control_map } = settings;
    const retries = (settings.retries ?? 1) - 1;
    const { deploymentDeviceIds, devices, filter } = settings;
    const group = settings.group;
    const startTime = phases?.length ? phases[0].start_ts : undefined;
    const retrySetting = canRetry && retries ? { retries } : {};
    const newDeployment = {
      artifact_name: release.name,
      autogenerate_delta: delta ? delta : undefined,
      devices: (filter || group) && !devices.length ? undefined : deploymentDeviceIds,
      filter_id: filter?.id,
      all_devices: !filter && group === ALL_DEVICES,
      group: group === ALL_DEVICES || devices.length ? undefined : group,
      max_devices: maxDevices ? maxDevices : undefined,
      name: devices[0]?.id || (group ? decodeURIComponent(group) : ALL_DEVICES),
      phases: phases.length
        ? phases.map((phase, i, origPhases) => {
            phase.start_ts = getPhaseStartTime(origPhases, i, startTime);
            return phase;
          })
        : undefined,
      ...retrySetting,
      force_installation: forceDeploy,
      update_control_map: !isEmpty(update_control_map.states) ? update_control_map : undefined
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

  const { deploymentDeviceCount, deploymentDeviceIds, devices } = derivedState;

  const filter = groups[group]?.id ? groups[group] : undefined;
  const deploymentSettings = {
    ...formValues,
    ...derivedState,
    filter,
    releaseSelectionLocked: deploymentObject.releaseSelectionLocked
  };
  const disabled = isCreating.current || !(release && (deploymentDeviceCount || !!filter || group)) || !validatePhases(phases, deploymentDeviceCount);

  const hasReleases = !!Object.keys(releasesById).length;
  return (
    <BaseDrawer open={open} onClose={closeWizard} size="sm" slotProps={{ header: { title: 'Create a deployment' } }}>
      <FormProvider {...methods}>
        <FormGroup>
          {!hasReleases ? (
            <ReleasesWarning />
          ) : (
            <>
              <Devices
                deploymentDeviceCount={deploymentDeviceCount}
                devices={devices}
                devicesById={devicesById}
                filter={filter}
                groupRef={groupRef}
                groupNames={groupNames}
                hasDevices={hasDevices}
                hasDynamicGroups={hasDynamicGroups}
                hasPending={hasPending}
                idAttribute={idAttribute}
              />
              <Software
                commonClasses={classes}
                devices={devices}
                releaseRef={releaseRef}
                releaseSelectionLocked={deploymentObject.releaseSelectionLocked}
                releases={releases}
                releasesById={releasesById}
              />
            </>
          )}
          <ScheduleRollout canSchedule={canSchedule} commonClasses={classes} />
          <Accordion className={classes.accordion} square expanded={isExpanded} onChange={() => setIsExpanded(toggle)}>
            <AccordionSummary expandIcon={<ExpandMore />}>
              <Typography className={classes.disabled} variant="subtitle2">
                {isExpanded ? 'Hide' : 'Show'} advanced options
              </Typography>
            </AccordionSummary>
            <AccordionDetails>
              <DeviceLimit deploymentDeviceCount={deploymentDeviceCount} deploymentDeviceIds={deploymentDeviceIds} filter={filter} />
              <RolloutPatternSelection
                deploymentDeviceCount={deploymentDeviceCount}
                deploymentDeviceIds={deploymentDeviceIds}
                filter={filter}
                isEnterprise={isEnterprise}
                previousPhases={previousPhases}
              />
              <RolloutOptions isEnterprise={isEnterprise} />
              <Retries canRetry={canRetry} commonClasses={classes} hasNewRetryDefault={hasNewRetryDefault} onSaveRetriesSetting={onSaveRetriesSetting} />
              <ForceDeploy />
              {hasDeltaEnabled && (
                <FormCheckbox id={deploymentFormSections.delta} control={methods.control} label="Generate and deploy Delta Artifacts where available" />
              )}
            </AccordionDetails>
          </Accordion>
        </FormGroup>
        <div className="margin-top relative">
          {isChecking && (
            <Confirm
              classes="confirmation-overlay"
              cancel={() => setIsChecking(false)}
              action={() => onScheduleSubmitClick(deploymentSettings)}
              message={`This will deploy ${release?.name} to ${deploymentDeviceCount} ${pluralize('device', deploymentDeviceCount)}. Are you sure?`}
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
          deploymentObject={deploymentSettings}
          deploymentAnchor={deploymentAnchor}
          onboardingState={onboardingState}
          createdGroup={createdGroup}
          releasesById={releasesById}
          releases={releases}
          hasDevices={hasDevices}
        />
      </FormProvider>
    </BaseDrawer>
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
