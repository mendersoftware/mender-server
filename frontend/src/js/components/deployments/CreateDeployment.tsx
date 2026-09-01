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
import { memo, useEffect, useRef, useState } from 'react';
import { FormProvider, useForm, useWatch } from 'react-hook-form';
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router';

import { ExpandMore } from '@mui/icons-material';
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Button,
  FormGroup,
  Typography,
  accordionClasses,
  accordionSummaryClasses,
  lighten
} from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import BaseDrawer from '@northern.tech/common-ui/BaseDrawer';
import Confirm from '@northern.tech/common-ui/Confirm';
import { DOCSTIPS, DocsTextLink } from '@northern.tech/common-ui/DocsLink';
import { InfoHintContainer } from '@northern.tech/common-ui/InfoHint';
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
  getOrganization,
  getReleaseListState,
  getReleasesById,
  getTenantCapabilities,
  getUserCapabilities
} from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { advanceOnboarding, createDeployment, getDeploymentsConfig, getRelease, getReleases } from '@northern.tech/store/thunks';
import { isEmpty, toggle } from '@northern.tech/utils/helpers';
import pluralize from 'pluralize';

import { getOnboardingComponentFor } from '../../utils/onboardingManager';
import DeviceLimit from './deployment-wizard/DeviceLimit';
import { RolloutPatternSelection } from './deployment-wizard/PhaseSettings';
import { ForceDeploy, Retries, RolloutOptions } from './deployment-wizard/RolloutOptions';
import { ScheduleRollout } from './deployment-wizard/ScheduleRollout';
import { Devices, ReleasesWarning, Software } from './deployment-wizard/SoftwareDevices';
import { rolloutModes, rolloutPatterns } from './deployment-wizard/phases/constants';
import type { DeploymentFormValues, DeploymentPrefill } from './deployment-wizard/types';
import { buildPhasePayload, deploymentFormSections, useDerivedData } from './deployment-wizard/utils';
import type { DeploymentResolverContext } from './deployment-wizard/validation';
import { deploymentResolver, getDeviceLimitDisabledReason, getPausesDisabledReason, getRolloutPatternDisabledReason } from './deployment-wizard/validation';

const useStyles = makeStyles()(theme => ({
  accordion: {
    backgroundColor: lighten(theme.palette.background.paper, 0.25),
    marginTop: theme.spacing(4),
    '&:before': {
      display: 'none'
    },
    [`& .${accordionSummaryClasses.content}`]: {
      margin: theme.spacing(1, 0)
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
  }
}));

// these live in the collapsed advanced options, which have to be expanded before their errors can be seen
const advancedErrorFields = [deploymentFormSections.maxDevices, deploymentFormSections.phases, deploymentFormSections.retries];

const watchedFields = [
  deploymentFormSections.group,
  deploymentFormSections.release,
  deploymentFormSections.isPaused,
  deploymentFormSections.shouldLimit,
  deploymentFormSections.usesPattern
] as const;

const getAnchor = (element, heightAdjustment = 3) => ({
  top: element.offsetTop + element.offsetHeight / heightAdjustment,
  left: element.offsetLeft + element.offsetWidth
});

export const defaultValues: DeploymentFormValues = {
  group: null,
  release: null,
  delta: false,
  forceDeploy: false,
  isPaused: false,
  maxDevices: 0,
  retries: 1,
  phases: [],
  rolloutMode: rolloutModes.percentage.key,
  rolloutPattern: rolloutPatterns.custom.key,
  startTime: undefined,
  shouldLimit: false,
  update_control_map: { states: {} },
  usesPattern: false
};

export const CreateDeployment = ({
  deploymentObject = {},
  onDismiss,
  onScheduleSubmit,
  onValuesChange,
  open
}: {
  deploymentObject: DeploymentPrefill;
  onDismiss: () => void;
  onScheduleSubmit: () => void;
  onValuesChange: (change: Pick<DeploymentPrefill, 'release'>) => void;
  open: boolean;
}) => {
  const { canRetry, canSchedule } = useSelector(getTenantCapabilities);
  const { isHosted } = useSelector(getFeatures);
  const { createdGroup, hasDynamicGroups } = useSelector(getGroupData);
  const { hasDelta: hasDeltaEnabled } = useSelector(state => state.deployments.config) ?? {};
  const devicesById = useSelector(getDevicesById);
  const { accepted: acceptedDeviceCount, pending: hasPending } = useSelector(getDeviceCountsByStatus);
  const hasDevices = !!acceptedDeviceCount;
  const { canManageUsers } = useSelector(getUserCapabilities);
  const idAttribute = useSelector(getIdAttribute);
  const isEnterprise = useSelector(getIsEnterprise);
  const { trial: isTrial } = useSelector(getOrganization);
  const { needsDeploymentConfirmation: needsCheck, previousPhases = [], retries: previousRetries = 0 } = useSelector(getGlobalSettings);
  const onboardingState = useSelector(getOnboardingState) || {};
  const { complete: isOnboardingComplete } = onboardingState;
  const { searchedIds: releases } = useSelector(getReleaseListState);
  const releasesById = useSelector(getReleasesById);
  const groupNames = useSelector(getGroupNames);
  const dispatch = useAppDispatch();
  const [isChecking, setIsChecking] = useState(false);
  const [isExpanded, setIsExpanded] = useState(false);
  const navigate = useNavigate();
  const releaseRef = useRef();
  const groupRef = useRef();
  const deploymentAnchor = useRef();
  const formRef = useRef<HTMLDivElement>(null);
  const { classes } = useStyles();
  // the validation depends on the targeted devices, which are not part of the form - so the resolver gets them handed
  // in as context, kept in a ref to have it up to date whenever the resolver runs
  const validationContext = useRef<DeploymentResolverContext>({ deploymentDeviceCount: 0, devices: [], group: null, filter: undefined });
  const methods = useForm<DeploymentFormValues>({
    context: validationContext.current,
    defaultValues,
    reValidateMode: 'onChange',
    resolver: deploymentResolver,
    // the errors belong to controls the browser can't focus by itself, so we take care of that below
    shouldFocusError: false
  });
  const {
    control,
    formState: { dirtyFields, isSubmitted, isSubmitting },
    getValues,
    handleSubmit,
    reset,
    setValue,
    trigger,
    watch
  } = methods;
  const [group, release, isPaused, shouldLimit, usesPattern] = useWatch({ control, name: watchedFields }) as [
    DeploymentFormValues['group'],
    DeploymentFormValues['release'],
    boolean,
    boolean,
    boolean
  ];
  const { deploymentDeviceCount, deploymentDeviceIds, devices, filter, isDeviceCountResolved } = useDerivedData(watch, deploymentObject.devices);
  validationContext.current.deploymentDeviceCount = deploymentDeviceCount;
  validationContext.current.devices = devices;
  validationContext.current.filter = filter;
  validationContext.current.group = group;

  const target = { deploymentDeviceCount, devices, filter, group, isDeviceCountResolved };
  const deviceLimitDisabledReason = getDeviceLimitDisabledReason(target);
  const rolloutPatternDisabledReason = getRolloutPatternDisabledReason({ ...target, isPaused });
  const pausesDisabledReason = getPausesDisabledReason({ ...target, usesPattern });

  useEffect(() => {
    dispatch(getReleases({ page: 1, perPage: 100, searchOnly: true, searchTerm: '', selectedTags: [], type: '' }));
  }, [dispatch]);

  useEffect(() => {
    if (isHosted || isEnterprise) {
      dispatch(getDeploymentsConfig());
    }
  }, [dispatch, isEnterprise, isHosted]);

  useEffect(() => {
    if (open) {
      reset({
        ...defaultValues,
        group: deploymentObject.group ?? defaultValues.group,
        release: deploymentObject.release ?? defaultValues.release,
        isPaused: !isEmpty(deploymentObject.update_control_map?.states ?? {}),
        retries: previousRetries + 1,
        update_control_map: deploymentObject.update_control_map ?? defaultValues.update_control_map
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, reset]);

  // options the selected target has ruled out have to be dropped - they could otherwise neither apply nor be removed,
  // with their controls disabled while the validation keeps rejecting them
  useEffect(() => {
    if (deviceLimitDisabledReason && shouldLimit) {
      setValue(deploymentFormSections.shouldLimit, false, { shouldValidate: isSubmitted });
      setValue(deploymentFormSections.maxDevices, 0, { shouldValidate: isSubmitted });
    }
    if (rolloutPatternDisabledReason && usesPattern) {
      setValue(deploymentFormSections.usesPattern, false, { shouldValidate: isSubmitted });
      setValue(deploymentFormSections.phases, [], { shouldValidate: isSubmitted });
    }
    if (pausesDisabledReason && isPaused) {
      setValue(deploymentFormSections.isPaused, false, { shouldValidate: isSubmitted });
    }
  }, [deviceLimitDisabledReason, isPaused, isSubmitted, pausesDisabledReason, rolloutPatternDisabledReason, setValue, shouldLimit, usesPattern]);

  // the target device count is not part of the form, so a change there has to re-run the validation by hand
  useEffect(() => {
    if (isSubmitted) {
      trigger();
    }
  }, [deploymentDeviceCount, devices.length, isSubmitted, trigger]);

  // the global settings can arrive after the form was initialized, so keep the retries default in sync until the user changed the field
  useEffect(() => {
    if (!open || dirtyFields.retries) {
      return;
    }
    setValue('retries', previousRetries + 1);
  }, [dirtyFields.retries, open, previousRetries, setValue]);

  useEffect(() => {
    onValuesChange?.({ release });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [onValuesChange, release?.name]);

  useEffect(() => {
    if (release) {
      dispatch(advanceOnboarding(onboardingSteps.SCHEDULING_ARTIFACT_SELECTION));
      dispatch(getRelease(release.name));
    }
    dispatch(advanceOnboarding(onboardingSteps.SCHEDULING_GROUP_SELECTION));
    if (group === ALL_DEVICES) {
      dispatch(advanceOnboarding(onboardingSteps.SCHEDULING_ALL_DEVICES_SELECTION));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [group, release?.name, dispatch]);

  const cleanUpDeploymentsStatus = () => {
    if (!window.location.search) {
      return;
    }
    const location = window.location.pathname.slice('/ui'.length);
    navigate(location); // lgtm [js/client-side-unvalidated-url-redirection]
  };

  const closeWizard = () => {
    cleanUpDeploymentsStatus();
    onDismiss();
  };

  const onScheduleSubmitClick = () => {
    if (needsCheck && !isChecking) {
      return setIsChecking(true);
    }
    const currentValues = getValues();
    const { delta, forceDeploy = false, isPaused, maxDevices, phases, release, rolloutMode, rolloutPattern, startTime, update_control_map } = currentValues;
    const retries = (currentValues.retries ?? 1) - 1;
    const retrySetting = canRetry && retries ? { retries } : {};
    const phasePayload = buildPhasePayload({ phases, rolloutMode, rolloutPattern, startTime });
    const newDeployment = {
      artifact_name: release.name,
      autogenerate_delta: delta ? delta : undefined,
      devices: (filter || group) && !devices.length ? undefined : deploymentDeviceIds,
      filter_id: filter?.id,
      all_devices: !filter && group === ALL_DEVICES,
      group: group === ALL_DEVICES || devices.length ? undefined : group,
      max_devices: maxDevices ? maxDevices : undefined,
      name: devices[0]?.id || (group ? decodeURIComponent(group) : ALL_DEVICES),
      ...phasePayload,
      ...retrySetting,
      force_installation: forceDeploy,
      update_control_map: isPaused && !isEmpty(update_control_map.states) ? update_control_map : undefined
    };
    if (!isOnboardingComplete) {
      dispatch(advanceOnboarding(onboardingSteps.SCHEDULING_RELEASE_TO_DEVICES));
    }
    return dispatch(createDeployment({ newDeployment }))
      .unwrap()
      .then(() => {
        // successfully retrieved new deployment
        cleanUpDeploymentsStatus();
        onScheduleSubmit();
      })
      .catch(console.error)
      .finally(() => setIsChecking(false));
  };

  const scrollToError = () => formRef.current?.querySelector('.Mui-error')?.scrollIntoView?.({ behavior: 'smooth', block: 'center' });

  const onInvalidSubmit = submitErrors => {
    // an invalid deployment can't be worth confirming, so make way for the errors instead
    setIsChecking(false);
    if (!isExpanded && Object.keys(submitErrors).some(field => advancedErrorFields.includes(field))) {
      // scrolling has to wait until the accordion has finished expanding
      return setIsExpanded(true);
    }
    scrollToError();
  };

  const hasReleases = !!Object.keys(releasesById).length;
  return (
    <BaseDrawer open={open} onClose={closeWizard} size="md" slotProps={{ header: { title: 'Create a deployment' } }}>
      <FormProvider {...methods}>
        <FormGroup ref={formRef}>
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
                releaseRef={releaseRef}
                releaseSelectionLocked={deploymentObject.releaseSelectionLocked}
                releases={releases}
                releasesById={releasesById}
              />
            </>
          )}
          <ScheduleRollout canSchedule={canSchedule} commonClasses={classes} />
          <Accordion
            className={classes.accordion}
            square
            expanded={isExpanded}
            onChange={() => setIsExpanded(toggle)}
            slotProps={{ transition: { onEntered: scrollToError } }}
          >
            <AccordionSummary expandIcon={<ExpandMore />}>
              <Typography variant="subtitle2">{isExpanded ? 'Hide' : 'Show'} advanced options</Typography>
            </AccordionSummary>
            <AccordionDetails>
              <Retries canManageUsers={canManageUsers} canRetry={canRetry} commonClasses={classes} defaultRetries={previousRetries} />
              <DeviceLimit deploymentDeviceCount={deploymentDeviceCount} disabledReason={deviceLimitDisabledReason} />
              <RolloutPatternSelection
                deploymentDeviceCount={deploymentDeviceCount}
                disabledReason={rolloutPatternDisabledReason}
                filter={filter}
                isEnterprise={isEnterprise}
                previousPhases={previousPhases}
              />
              <RolloutOptions disabledReason={pausesDisabledReason} isEnterprise={isEnterprise} />
              <ForceDeploy />
              {!isTrial && hasDeltaEnabled && (
                <FormCheckbox
                  id={deploymentFormSections.delta}
                  control={control}
                  label={
                    <div className="flexbox align-items-center">
                      Generate and deploy Delta Artifacts where available
                      <InfoHintContainer>
                        <DocsTextLink id={DOCSTIPS.deltaArtifacts.id} />
                      </InfoHintContainer>
                    </div>
                  }
                  slotProps={{ checkbox: { className: 'margin-left-small', size: 'small' } }}
                />
              )}
            </AccordionDetails>
          </Accordion>
        </FormGroup>
        <div className="margin-top relative">
          {isChecking && (
            <Confirm
              classes="confirmation-overlay"
              cancel={() => setIsChecking(false)}
              action={handleSubmit(onScheduleSubmitClick, onInvalidSubmit)}
              message={`This will deploy ${release?.name} to ${deploymentDeviceCount} ${pluralize('device', deploymentDeviceCount)}. Are you sure?`}
              style={{ paddingLeft: 12, justifyContent: 'flex-start', maxHeight: 44 }}
            />
          )}
          <Button onClick={closeWizard} style={{ marginRight: 10 }}>
            Cancel
          </Button>
          <Button
            variant="contained"
            color="primary"
            ref={deploymentAnchor}
            disabled={isSubmitting}
            onClick={handleSubmit(onScheduleSubmitClick, onInvalidSubmit)}
          >
            Create deployment
          </Button>
        </div>
        <OnboardingComponent
          releaseRef={releaseRef}
          groupRef={groupRef}
          deploymentAnchor={deploymentAnchor}
          deploymentDeviceCount={deploymentDeviceCount}
          devices={devices}
          group={group}
          release={release}
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

export default memo(CreateDeployment);

const OnboardingComponent = ({
  releaseRef,
  groupRef,
  deploymentAnchor,
  deploymentDeviceCount,
  devices,
  group,
  release: deploymentRelease = null,
  onboardingState,
  createdGroup,
  releasesById,
  releases,
  hasDevices
}) => {
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
      const buttonAnchor = {
        top: deploymentAnchor.current.parentElement.offsetTop + deploymentAnchor.current.offsetHeight / 2,
        left: deploymentAnchor.current.parentElement.offsetLeft + deploymentAnchor.current.offsetLeft + deploymentAnchor.current.offsetWidth
      };
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
