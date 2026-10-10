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
import { useCallback, useEffect, useRef, useState } from 'react';

import storeActions from '@northern.tech/store/actions';
import actions from '@northern.tech/store/actions';
import { getSessionInfo } from '@northern.tech/store/auth';
import { DEPLOYMENT_STATES, DEVICE_STATES, locations, timeUnits } from '@northern.tech/store/constants';
import type { DeviceSliceType } from '@northern.tech/store/devicesSlice';
import {
  getCurrentSession,
  getCurrentUser,
  getDevicesByStatus as getDevicesByStatusSelector,
  getFeatures,
  getGlobalSettings as getGlobalSettingsSelector,
  getIsServiceProvider,
  getOfflineThresholdSettings,
  getOnboardingState as getOnboardingStateSelector,
  getSortedFilteringAttributes,
  getUserCapabilities,
  getUserSettingsInitialized,
  getUserSettings as getUserSettingsSelector
} from '@northern.tech/store/selectors';
import { createAppAsyncThunk, useAppDispatch, useAppSelector } from '@northern.tech/store/store';
import type { AppDispatch } from '@northern.tech/store/store';
import {
  getAllDeviceCounts,
  getDeploymentsByStatus,
  getDeviceAttributes,
  getDeviceById,
  getDeviceLimits,
  getDevicesByStatus,
  getDynamicGroups,
  getGlobalSettings,
  getGroups,
  getIntegrations,
  getOnboardingState,
  getReleases,
  getRoles,
  getUserOrganization,
  getUserSettings,
  initializeSelf,
  saveGlobalSettings,
  saveUserSettings
} from '@northern.tech/store/thunks';
import type { UserSettings, UserSliceType } from '@northern.tech/store/usersSlice';
import { stringToBoolean } from '@northern.tech/store/utils';
import { TIMEOUTS } from '@northern.tech/utils/constants';
import { useDebounce } from '@northern.tech/utils/debouncehook';
import { extractErrorMessage } from '@northern.tech/utils/helpers';
import dayjs from 'dayjs';
import durationDayJs from 'dayjs/plugin/duration.js';
import Cookies from 'universal-cookie';

const cookies = new Cookies();
dayjs.extend(durationDayJs);

const { setDeviceListState, setFirstLoginAfterSignup, setTooltipsState, setShowStartupNotification } = storeActions;

// a host is hosted when it is one of the known hosted domains or a subdomain of one (e.g. staging.hosted.mender.io),
// except for per-PR preview deployments whose leading label looks like <component>-pr-<number> (e.g. os-pr-2012.staging.hosted.mender.io)
const hostedDomains = Object.values(locations).map(({ location }) => location);
const isPreviewDeployment = (hostname: string) => /^[^.]+-pr-\d+\./.test(hostname);
const isHostedHostname = (hostname: string) =>
  !isPreviewDeployment(hostname) && hostedDomains.some(domain => hostname === domain || hostname.endsWith(`.${domain}`));

const featureFlags = [
  'hasAiEnabled',
  'hasAuditlogs',
  'hasMultitenancy',
  'hasDeltaProgress',
  'hasDeviceConfig',
  'hasDeviceConnect',
  'hasDeviceFlags',
  'hasFeedbackEnabled',
  'hasManifestsEnabled',
  'hasMonitor',
  'hasMCUEnabled',
  'hasNewSearch',
  'isEnterprise'
] as const;

const environmentDatas = [
  'commit',
  'feedbackProbability',
  'hostAddress',
  'hostedAnnouncement',
  'recaptchaSiteKey',
  'sentry',
  'stripeAPIKey',
  'trackerCode'
] as const;

type VersionInfo = {
  docsVersion?: string;
  version?: string;
};

type FeatureFlagsState = Record<string, boolean>;

export const parseEnvironmentInfo = createAppAsyncThunk(`app/parseEnvironmentInfo`, (_, { dispatch, getState }) => {
  const state = getState();
  let onboardingComplete = getOnboardingStateSelector(state).complete || !!JSON.parse(window.localStorage.getItem('onboardingComplete') ?? 'false');
  let demoArtifactPort = 85;
  let environmentData: Record<string, unknown> = {};
  let environmentFeatures: FeatureFlagsState = {};
  let versionInfo: VersionInfo = {};
  const mender_environment = window.mender_environment;
  if (mender_environment) {
    const { features = {}, demoArtifactPort: port, disableOnboarding, version } = mender_environment;
    demoArtifactPort = Number(port) || demoArtifactPort;
    const appState = state.app;
    environmentData = environmentDatas.reduce((accu, flag) => ({ ...accu, [flag]: mender_environment[flag] || appState[flag as keyof typeof appState] }), {});
    environmentFeatures = {
      ...featureFlags.reduce<FeatureFlagsState>((accu, flag) => ({ ...accu, [flag]: stringToBoolean(features[flag]) }), {}),
      isHosted: stringToBoolean(features.isHosted) || isHostedHostname(window.location.hostname)
    };
    onboardingComplete = !environmentFeatures.isHosted || stringToBoolean(disableOnboarding) || onboardingComplete;
    versionInfo = {
      docsVersion: isNaN(parseInt(version.charAt(0))) ? '' : version.split('.').slice(0, 2).join('.'),
      version
    };
  }
  return Promise.all([
    dispatch(storeActions.successfullyLoggedIn(getSessionInfo())),
    dispatch(storeActions.setOnboardingComplete(onboardingComplete)),
    dispatch(storeActions.setDemoArtifactPort(demoArtifactPort)),
    dispatch(storeActions.setFeatures(environmentFeatures)),
    dispatch(storeActions.setVersionInformation(versionInfo)),
    dispatch(storeActions.setEnvironmentData(environmentData))
  ]);
});

// probes the organization endpoint to classify the deployment: reaching it means the deployment is multitenant,
// which - when not hosted - identifies an on-prem enterprise installation, while an unreachable endpoint identifies
// a single-tenant OS installation. hosted deployments keep their plan-derived enterprise status (getIsEnterprise).
// must run after parseEnvironmentInfo so the final isHosted value is available in the store.
export const probeOrganizationCapabilities = createAppAsyncThunk(`app/probeOrganizationCapabilities`, (_, { dispatch, getState }) =>
  dispatch(getUserOrganization())
    .unwrap()
    .then(() => {
      if (!getFeatures(getState()).isHosted) {
        dispatch(storeActions.setFeatures({ hasMultitenancy: true, isEnterprise: true }));
      }
    })
    .catch(() => {
      if (!getFeatures(getState()).isHosted) {
        dispatch(storeActions.setFeatures({ hasMultitenancy: false, isEnterprise: false }));
      }
    })
);

type OnboardingState = {
  complete?: boolean;
  showTips?: boolean | null;
};

type MaybeAddOnboardingTasksParams = {
  devicesByStatus: DeviceSliceType['byStatus'];
  dispatch: AppDispatch;
  onboardingState: OnboardingState;
  tasks: Promise<unknown>[];
};

const maybeAddOnboardingTasks = ({ devicesByStatus, dispatch, onboardingState, tasks }: MaybeAddOnboardingTasksParams): Promise<unknown>[] => {
  if (!onboardingState.showTips || onboardingState.complete) {
    return tasks;
  }
  // try to retrieve full device details for onboarding devices to ensure ips etc. are available
  // we only load the first few/ 20 devices, as it is possible the onboarding is left dangling
  // and a lot of devices are present and we don't want to flood the backend for this
  return devicesByStatus[DEVICE_STATES.accepted].deviceIds.reduce<Promise<unknown>[]>((accu, id) => {
    accu.push(dispatch(getDeviceById(id)));
    return accu;
  }, tasks);
};

export const useAppInit = (): { coreInitDone: boolean } => {
  const dispatch = useAppDispatch();
  const [coreInitDone, setCoreInitDone] = useState(false);
  const user = useAppSelector(getCurrentUser);
  const { token } = useAppSelector(getCurrentSession);
  const userSettingInitialized = useAppSelector(getUserSettingsInitialized);
  const userId = useDebounce(user.id, TIMEOUTS.debounceDefault);
  const { hasMultitenancy } = useAppSelector(getFeatures);
  const devicesByStatus = useAppSelector(getDevicesByStatusSelector);
  const onboardingState = useAppSelector(getOnboardingStateSelector);
  const { columnSelection = [], trackingConsentGiven: hasTrackingEnabled, tooltips = {} } = useAppSelector(getUserSettingsSelector);
  const { canManageUsers } = useAppSelector(getUserCapabilities);
  const { interval, intervalUnit } = useAppSelector(getOfflineThresholdSettings);
  const { id_attribute } = useAppSelector(getGlobalSettingsSelector);
  const { identityAttributes } = useAppSelector(getSortedFilteringAttributes);
  const isServiceProvider = useAppSelector(getIsServiceProvider);
  const selfInitRunning = useRef(false);
  const coreInitRunning = useRef(false);
  const fullInitRunning = useRef(false);

  const retrieveCoreData = useCallback(
    (): Promise<unknown[]> =>
      Promise.all([
        // parse the environment first so isHosted is set before the organization probe classifies the deployment type
        dispatch(parseEnvironmentInfo())
          .unwrap()
          .then(() => dispatch(probeOrganizationCapabilities()).unwrap()),
        dispatch(getUserSettings()).unwrap(),
        dispatch(getGlobalSettings()).unwrap(),
        Promise.resolve(dispatch(setFirstLoginAfterSignup(stringToBoolean(cookies.get('firstLoginAfterSignup')))))
      ]),
    [dispatch]
  );

  const retrieveAppData = useCallback((): Promise<unknown[] | unknown> => {
    if (isServiceProvider) {
      return Promise.resolve(dispatch(getRoles()));
    }
    const tasks: Promise<unknown>[] = [
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
      dispatch(getAllDeviceCounts()).unwrap()
    ];
    if (hasMultitenancy) {
      tasks.push(dispatch(getDeviceLimits()));
      tasks.push(dispatch(getRoles()));
    } else {
      // OS users have unlimited standard and test devices
      tasks.push(Promise.resolve(dispatch(actions.setDeviceLimits({ micro: 0, standard: -1, system: 0, test: -1 }))));
    }
    return Promise.all(tasks);
  }, [dispatch, hasMultitenancy, isServiceProvider]);

  const interpretAppData = useCallback((): Promise<unknown[]> => {
    const settings: Partial<UserSettings> = {};
    if (cookies.get('_ga') && typeof hasTrackingEnabled === 'undefined') {
      settings.trackingConsentGiven = true;
    }
    let tasks: Promise<unknown>[] = [
      Promise.resolve(dispatch(setDeviceListState({ selectedAttributes: columnSelection.map(column => ({ attribute: column.key, scope: column.scope })) }))),
      Promise.resolve(dispatch(setTooltipsState(tooltips as UserSliceType['tooltips']['byId']))), // tooltips read state is primarily trusted from the redux store, except on app init - here user settings are the reference
      dispatch(saveUserSettings(settings)).unwrap()
    ];
    // checks if user id is set and if cookie for helptips exists for that user
    tasks = maybeAddOnboardingTasks({ devicesByStatus, dispatch, tasks, onboardingState });

    if (canManageUsers && intervalUnit && intervalUnit !== timeUnits.days) {
      const duration = dayjs.duration(interval, intervalUnit as keyof typeof timeUnits);
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
      // Legacy migration path: id_attribute used to be a plain string
      let attribute: string = id_attribute;
      if (attribute === 'Device ID') {
        attribute = 'id';
      }
      tasks.push(dispatch(saveGlobalSettings({ id_attribute: { attribute, scope: 'identity' } })));
    }
    tasks.push(Promise.resolve(dispatch(storeActions.setAppInitDone())));
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
    (): Promise<unknown> =>
      retrieveAppData()
        .then(interpretAppData)
        // this is allowed to fail if no user information are available
        .catch(err => console.log(extractErrorMessage(err)))
        .then(() => dispatch(getOnboardingState())),
    [dispatch, retrieveAppData, interpretAppData]
  );

  useEffect(() => {
    if (selfInitRunning.current || !token || (userId && user.email?.length && userSettingInitialized)) {
      return;
    }
    selfInitRunning.current = true;
    dispatch(getUserOrganization());
    dispatch(initializeSelf());
  }, [dispatch, token, user.email, userId, userSettingInitialized]);

  useEffect(() => {
    if (!userId || coreInitRunning.current) {
      return;
    }
    coreInitRunning.current = true;
    retrieveCoreData().then(() => setCoreInitDone(true));
  }, [userId, retrieveCoreData]);

  useEffect(() => {
    if (fullInitRunning.current || !coreInitDone) {
      return;
    }
    fullInitRunning.current = true;
    initializeAppData();
  }, [initializeAppData, coreInitDone]);

  return { coreInitDone };
};
