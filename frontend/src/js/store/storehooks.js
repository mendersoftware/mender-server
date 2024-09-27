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
import { useCallback, useEffect, useRef } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import dayjs from 'dayjs';
import durationDayJs from 'dayjs/plugin/duration';
import Cookies from 'universal-cookie';

import { getOnboardingComponentFor } from '../utils/onboardingmanager';
import storeActions from './actions';
import { getSessionInfo } from './auth';
import { DEPLOYMENT_STATES, DEVICE_STATES, TIMEOUTS, onboardingSteps, timeUnits } from './constants';
import {
  getDevicesByStatus as getDevicesByStatusSelector,
  getFeatures,
  getGlobalSettings as getGlobalSettingsSelector,
  getIsEnterprise,
  getOfflineThresholdSettings,
  getOnboardingState as getOnboardingStateSelector,
  getSortedFilteringAttributes,
  getUserCapabilities,
  getUserSettings as getUserSettingsSelector
} from './selectors';
import {
  getDeploymentsByStatus,
  getDeviceAttributes,
  getDeviceById,
  getDeviceLimit,
  getDevicesByStatus,
  getDynamicGroups,
  getGlobalSettings,
  getGroups,
  getIntegrations,
  getLatestReleaseInfo,
  getOnboardingState,
  getReleases,
  getRoles,
  getUserOrganization,
  getUserSettings,
  saveGlobalSettings,
  saveUserSettings
} from './thunks';
import { extractErrorMessage, getComparisonCompatibleVersion, stringToBoolean } from './utils';

const cookies = new Cookies();
dayjs.extend(durationDayJs);

const { setSnackbar, setDeviceListState, setFirstLoginAfterSignup, setTooltipsState, setShowStartupNotification } = storeActions;

const featureFlags = [
  'hasAuditlogs',
  'hasMultitenancy',
  'hasDeltaProgress',
  'hasDeviceConfig',
  'hasDeviceConnect',
  'hasReporting',
  'hasMonitor',
  'isEnterprise'
];

export const parseEnvironmentInfo = () => (dispatch, getState) => {
  const state = getState();
  let onboardingComplete = state.onboarding.complete || !!JSON.parse(window.localStorage.getItem('onboardingComplete') ?? 'false');
  let demoArtifactPort = 85;
  let environmentData = {};
  let environmentFeatures = {};
  let versionInfo = {};
  if (mender_environment) {
    const {
      features = {},
      demoArtifactPort: port,
      disableOnboarding,
      hostAddress,
      hostedAnnouncement,
      integrationVersion,
      isDemoMode,
      menderVersion,
      menderArtifactVersion,
      metaMenderVersion,
      recaptchaSiteKey,
      services = {},
      stripeAPIKey,
      trackerCode
    } = mender_environment;
    onboardingComplete = stringToBoolean(features.isEnterprise) || stringToBoolean(disableOnboarding) || onboardingComplete;
    demoArtifactPort = port || demoArtifactPort;
    environmentData = {
      hostedAnnouncement: hostedAnnouncement || state.app.hostedAnnouncement,
      hostAddress: hostAddress || state.app.hostAddress,
      recaptchaSiteKey: recaptchaSiteKey || state.app.recaptchaSiteKey,
      stripeAPIKey: stripeAPIKey || state.app.stripeAPIKey,
      trackerCode: trackerCode || state.app.trackerCode
    };
    environmentFeatures = {
      ...featureFlags.reduce((accu, flag) => ({ ...accu, [flag]: stringToBoolean(features[flag]) }), {}),
      isHosted: features.isHosted || window.location.hostname.includes('hosted.mender.io'),
      isDemoMode: stringToBoolean(isDemoMode || features.isDemoMode)
    };
    versionInfo = {
      docs: isNaN(integrationVersion.charAt(0)) ? '' : integrationVersion.split('.').slice(0, 2).join('.'),
      remainder: {
        Integration: getComparisonCompatibleVersion(integrationVersion),
        'Mender-Client': getComparisonCompatibleVersion(menderVersion),
        'Mender-Artifact': menderArtifactVersion,
        'Meta-Mender': metaMenderVersion,
        Deployments: services.deploymentsVersion,
        Deviceauth: services.deviceauthVersion,
        Inventory: services.inventoryVersion,
        GUI: services.guiVersion
      }
    };
  }
  return Promise.all([
    dispatch(storeActions.successfullyLoggedIn(getSessionInfo())),
    dispatch(storeActions.setOnboardingComplete(onboardingComplete)),
    dispatch(storeActions.setDemoArtifactPort(demoArtifactPort)),
    dispatch(storeActions.setFeatures(environmentFeatures)),
    dispatch(storeActions.setVersionInformation({ docsVersion: versionInfo.docs, value: versionInfo.remainder })),
    dispatch(storeActions.setEnvironmentData(environmentData)),
    dispatch(getLatestReleaseInfo())
  ]);
};

const maybeAddOnboardingTasks = ({ devicesByStatus, dispatch, onboardingState, tasks }) => {
  if (!onboardingState.showTips || onboardingState.complete) {
    return tasks;
  }
  const welcomeTip = getOnboardingComponentFor(onboardingSteps.ONBOARDING_START, {
    progress: onboardingState.progress,
    complete: onboardingState.complete,
    showTips: onboardingState.showTips
  });
  if (welcomeTip) {
    tasks.push(dispatch(setSnackbar({ message: 'open', autoHideDuration: TIMEOUTS.refreshDefault, children: welcomeTip, onClick: () => {}, onClose: true })));
  }
  // try to retrieve full device details for onboarding devices to ensure ips etc. are available
  // we only load the first few/ 20 devices, as it is possible the onboarding is left dangling
  // and a lot of devices are present and we don't want to flood the backend for this
  return devicesByStatus[DEVICE_STATES.accepted].deviceIds.reduce((accu, id) => {
    accu.push(dispatch(getDeviceById(id)));
    return accu;
  }, tasks);
};

export const useAppInit = userId => {
  const dispatch = useDispatch();
  const isEnterprise = useSelector(getIsEnterprise);
  const { hasMultitenancy, isHosted } = useSelector(getFeatures);
  // const user = useSelector(getCurrentUser);
  const devicesByStatus = useSelector(getDevicesByStatusSelector);
  const onboardingState = useSelector(getOnboardingStateSelector);
  let { columnSelection = [], trackingConsentGiven: hasTrackingEnabled, tooltips = {} } = useSelector(getUserSettingsSelector);
  const { canManageUsers } = useSelector(getUserCapabilities);
  const { interval, intervalUnit } = useSelector(getOfflineThresholdSettings);
  const { id_attribute } = useSelector(getGlobalSettingsSelector);
  const { identityAttributes } = useSelector(getSortedFilteringAttributes);
  const initRunning = useRef(false);

  const retrieveAppData = useCallback(() => {
    let tasks = [
      dispatch(parseEnvironmentInfo()),
      dispatch(getUserSettings()),
      dispatch(getGlobalSettings()),
      dispatch(getDeviceAttributes()),
      dispatch(getDeploymentsByStatus({ status: DEPLOYMENT_STATES.finished, shouldSelect: false })),
      dispatch(getDeploymentsByStatus({ status: DEPLOYMENT_STATES.inprogress })),
      dispatch(getDevicesByStatus({ status: DEVICE_STATES.accepted })),
      dispatch(getDevicesByStatus({ status: DEVICE_STATES.pending })),
      dispatch(getDevicesByStatus({ status: DEVICE_STATES.preauth })),
      dispatch(getDevicesByStatus({ status: DEVICE_STATES.rejected })),
      dispatch(getDynamicGroups()),
      dispatch(getGroups()),
      dispatch(getIntegrations()),
      dispatch(getReleases()),
      dispatch(getDeviceLimit()),
      dispatch(getRoles()),
      dispatch(setFirstLoginAfterSignup(stringToBoolean(cookies.get('firstLoginAfterSignup'))))
    ];
    const multitenancy = hasMultitenancy || isHosted || isEnterprise;
    if (multitenancy) {
      tasks.push(dispatch(getUserOrganization()));
    }
    return Promise.all(tasks);
  }, [dispatch, hasMultitenancy, isHosted, isEnterprise]);

  const interpretAppData = useCallback(() => {
    let settings = {};
    if (cookies.get('_ga') && typeof hasTrackingEnabled === 'undefined') {
      settings.trackingConsentGiven = true;
    }
    let tasks = [
      dispatch(setDeviceListState({ selectedAttributes: columnSelection.map(column => ({ attribute: column.key, scope: column.scope })) })),
      dispatch(setTooltipsState(tooltips)), // tooltips read state is primarily trusted from the redux store, except on app init - here user settings are the reference
      dispatch(saveUserSettings(settings))
    ];
    // checks if user id is set and if cookie for helptips exists for that user
    tasks = maybeAddOnboardingTasks({ devicesByStatus, dispatch, tasks, onboardingState });

    if (canManageUsers && intervalUnit && intervalUnit !== timeUnits.days) {
      const duration = dayjs.duration(interval, intervalUnit);
      const days = duration.asDays();
      if (days < 1) {
        tasks.push(Promise.resolve(setTimeout(() => dispatch(setShowStartupNotification(true)), TIMEOUTS.fiveSeconds)));
      } else {
        const roundedDays = Math.max(1, Math.round(days));
        tasks.push(dispatch(saveGlobalSettings({ offlineThreshold: { interval: roundedDays, intervalUnit: timeUnits.days } })));
      }
    }

    // the following is used as a migration and initialization of the stored identity attribute
    // changing the default device attribute to the first non-deviceId attribute, unless a stored
    // id attribute setting exists
    const identityOptions = identityAttributes.filter(attribute => !['id', 'Device ID', 'status'].includes(attribute));
    if (!id_attribute && identityOptions.length) {
      tasks.push(dispatch(saveGlobalSettings({ id_attribute: { attribute: identityOptions[0], scope: 'identity' } })));
    } else if (typeof id_attribute === 'string') {
      let attribute = id_attribute;
      if (attribute === 'Device ID') {
        attribute = 'id';
      }
      tasks.push(dispatch(saveGlobalSettings({ id_attribute: { attribute, scope: 'identity' } })));
    }
    return Promise.all(tasks);
  }, [
    columnSelection,
    dispatch,
    identityAttributes,
    hasTrackingEnabled,
    canManageUsers,
    devicesByStatus,
    id_attribute,
    interval,
    intervalUnit,
    onboardingState,
    tooltips
  ]);

  const initializeAppData = useCallback(
    () =>
      retrieveAppData()
        .then(interpretAppData)
        // this is allowed to fail if no user information are available
        .catch(err => console.log(extractErrorMessage(err)))
        .then(() => dispatch(getOnboardingState())),
    [dispatch, retrieveAppData, interpretAppData]
  );

  useEffect(() => {
    if (!userId || initRunning.current) {
      return;
    }
    initRunning.current = true;
    initializeAppData();
  }, [userId, initializeAppData]);
};
