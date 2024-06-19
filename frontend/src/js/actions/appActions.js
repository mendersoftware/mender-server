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
import moment from 'moment';
import Cookies from 'universal-cookie';

import GeneralApi from '../api/general-api';
import { getSessionInfo, getToken } from '../auth';
import {
  SET_ENVIRONMENT_DATA,
  SET_FEATURES,
  SET_FIRST_LOGIN_AFTER_SIGNUP,
  SET_OFFLINE_THRESHOLD,
  SET_SEARCH_STATE,
  SET_SNACKBAR,
  SET_VERSION_INFORMATION,
  TIMEOUTS
} from '../constants/appConstants';
import { DEPLOYMENT_STATES } from '../constants/deploymentConstants';
import { DEVICE_STATES, timeUnits } from '../constants/deviceConstants';
import { onboardingSteps } from '../constants/onboardingConstants';
import { SET_TOOLTIPS_STATE, SUCCESSFULLY_LOGGED_IN } from '../constants/userConstants';
import { deepCompare, extractErrorMessage, preformatWithRequestID, stringToBoolean } from '../helpers';
import { getFeatures, getIsEnterprise, getOfflineThresholdSettings, getUserCapabilities, getUserSettings as getUserSettingsSelector } from '../selectors';
import { getOnboardingComponentFor } from '../utils/onboardingmanager';
import { getDeploymentsByStatus } from './deploymentActions';
import {
  getDeviceAttributes,
  getDeviceById,
  getDeviceLimit,
  getDevicesByStatus,
  getDynamicGroups,
  getGroups,
  searchDevices,
  setDeviceListState
} from './deviceActions';
import { getOnboardingState, setDemoArtifactPort, setOnboardingComplete } from './onboardingActions';
import { getIntegrations, getUserOrganization } from './organizationActions';
import { getReleases } from './releaseActions';
import { getGlobalSettings, getRoles, getUserSettings, saveGlobalSettings, saveUserSettings, setShowStartupNotification } from './userActions';

const cookies = new Cookies();

export const commonErrorFallback = 'Please check your connection.';
export const commonErrorHandler = (err, errorContext, dispatch, fallback, mightBeAuthRelated = false) => {
  const errMsg = extractErrorMessage(err, fallback);
  if (mightBeAuthRelated || getToken()) {
    dispatch(setSnackbar(preformatWithRequestID(err.response, `${errorContext} ${errMsg}`), null, 'Copy to clipboard'));
  }
  return Promise.reject(err);
};

const getComparisonCompatibleVersion = version => (isNaN(version.charAt(0)) && version !== 'next' ? 'master' : version);

const featureFlags = [
  'hasAuditlogs',
  'hasMultitenancy',
  'hasDeltaProgress',
  'hasDeviceConfig',
  'hasDeviceConnect',
  'hasReporting',
  'hasMonitor',
  'hasMultiTenantAccess',
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
      // the check in features is purely kept as a local override, it shouldn't become relevant for production again
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
    dispatch({ type: SUCCESSFULLY_LOGGED_IN, value: getSessionInfo() }),
    dispatch(setOnboardingComplete(onboardingComplete)),
    dispatch(setDemoArtifactPort(demoArtifactPort)),
    dispatch({ type: SET_FEATURES, value: environmentFeatures }),
    dispatch({ type: SET_VERSION_INFORMATION, docsVersion: versionInfo.docs, value: versionInfo.remainder }),
    dispatch({ type: SET_ENVIRONMENT_DATA, value: environmentData }),
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
    tasks.push(dispatch(setSnackbar('open', TIMEOUTS.refreshDefault, '', welcomeTip, () => {}, true)));
  }
  // try to retrieve full device details for onboarding devices to ensure ips etc. are available
  // we only load the first few/ 20 devices, as it is possible the onboarding is left dangling
  // and a lot of devices are present and we don't want to flood the backend for this
  return devicesByStatus[DEVICE_STATES.accepted].deviceIds.reduce((accu, id) => {
    accu.push(dispatch(getDeviceById(id)));
    return accu;
  }, tasks);
};

const interpretAppData = () => (dispatch, getState) => {
  const state = getState();
  let { columnSelection = [], trackingConsentGiven: hasTrackingEnabled, tooltips = {} } = getUserSettingsSelector(state);
  let settings = {};
  if (cookies.get('_ga') && typeof hasTrackingEnabled === 'undefined') {
    settings.trackingConsentGiven = true;
  }
  let tasks = [
    dispatch(setDeviceListState({ selectedAttributes: columnSelection.map(column => ({ attribute: column.key, scope: column.scope })) })),
    dispatch({ type: SET_TOOLTIPS_STATE, value: tooltips }), // tooltips read state is primarily trusted from the redux store, except on app init - here user settings are the reference
    dispatch(saveUserSettings(settings))
  ];
  tasks = maybeAddOnboardingTasks({ devicesByStatus: state.devices.byStatus, dispatch, tasks, onboardingState: state.onboarding });

  const { canManageUsers } = getUserCapabilities(getState());
  const { interval, intervalUnit } = getOfflineThresholdSettings(getState());
  if (canManageUsers && intervalUnit && intervalUnit !== timeUnits.days) {
    const duration = moment.duration(interval, intervalUnit);
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
  const identityOptions = state.devices.filteringAttributes.identityAttributes.filter(attribute => !['id', 'Device ID', 'status'].includes(attribute));
  const { id_attribute } = state.users.globalSettings;
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
};

const retrieveAppData = () => (dispatch, getState) => {
  let tasks = [
    dispatch(parseEnvironmentInfo()),
    dispatch(getUserSettings()),
    dispatch(getGlobalSettings()),
    dispatch(getDeviceAttributes()),
    dispatch(getDeploymentsByStatus(DEPLOYMENT_STATES.finished, undefined, undefined, undefined, undefined, undefined, undefined, false)),
    dispatch(getDeploymentsByStatus(DEPLOYMENT_STATES.inprogress)),
    dispatch(getDevicesByStatus(DEVICE_STATES.accepted)),
    dispatch(getDevicesByStatus(DEVICE_STATES.pending)),
    dispatch(getDevicesByStatus(DEVICE_STATES.preauth)),
    dispatch(getDevicesByStatus(DEVICE_STATES.rejected)),
    dispatch(getDynamicGroups()),
    dispatch(getGroups()),
    dispatch(getIntegrations()),
    dispatch(getReleases()),
    dispatch(getDeviceLimit()),
    dispatch(getRoles()),
    dispatch(setFirstLoginAfterSignup(stringToBoolean(cookies.get('firstLoginAfterSignup'))))
  ];
  const { hasMultitenancy, isHosted } = getFeatures(getState());
  const multitenancy = hasMultitenancy || isHosted || getIsEnterprise(getState());
  if (multitenancy) {
    tasks.push(dispatch(getUserOrganization()));
  }
  return Promise.all(tasks);
};

export const initializeAppData = () => dispatch =>
  dispatch(retrieveAppData())
    .then(() => dispatch(interpretAppData()))
    // this is allowed to fail if no user information are available
    .catch(err => console.log(extractErrorMessage(err)))
    .then(() => dispatch(getOnboardingState()));

/*
  General
*/
export const setSnackbar = (message, autoHideDuration, action, children, onClick, onClose) => dispatch =>
  dispatch({
    type: SET_SNACKBAR,
    snackbar: {
      open: message ? true : false,
      message,
      maxWidth: '900px',
      autoHideDuration,
      action,
      children,
      onClick,
      onClose
    }
  });

export const setFirstLoginAfterSignup = firstLoginAfterSignup => dispatch => {
  cookies.set('firstLoginAfterSignup', !!firstLoginAfterSignup, { maxAge: 60, path: '/', domain: '.mender.io', sameSite: false });
  dispatch({ type: SET_FIRST_LOGIN_AFTER_SIGNUP, firstLoginAfterSignup: !!firstLoginAfterSignup });
};

const dateFunctionMap = {
  getDays: 'getDate',
  setDays: 'setDate'
};
export const setOfflineThreshold = () => (dispatch, getState) => {
  const { interval, intervalUnit } = getOfflineThresholdSettings(getState());
  const today = new Date();
  const intervalName = `${intervalUnit.charAt(0).toUpperCase()}${intervalUnit.substring(1)}`;
  const setter = dateFunctionMap[`set${intervalName}`] ?? `set${intervalName}`;
  const getter = dateFunctionMap[`get${intervalName}`] ?? `get${intervalName}`;
  today[setter](today[getter]() - interval);
  let value;
  try {
    value = today.toISOString();
  } catch {
    return Promise.resolve(dispatch(setSnackbar('There was an error saving the offline threshold, please check your settings.')));
  }
  return Promise.resolve(dispatch({ type: SET_OFFLINE_THRESHOLD, value }));
};

export const setVersionInfo = info => (dispatch, getState) =>
  Promise.resolve(
    dispatch({
      type: SET_VERSION_INFORMATION,
      docsVersion: getState().app.docsVersion,
      value: {
        ...getState().app.versionInformation,
        ...info
      }
    })
  );

const versionRegex = new RegExp(/\d+\.\d+/);
const getLatestRelease = thing => {
  const latestKey = Object.keys(thing)
    .filter(key => versionRegex.test(key))
    .sort()
    .reverse()[0];
  return thing[latestKey];
};

const repoKeyMap = {
  integration: 'Integration',
  mender: 'Mender-Client',
  'mender-artifact': 'Mender-Artifact'
};

const deductSaasState = (latestRelease, guiTags, saasReleases) => {
  const latestGuiTag = guiTags.length ? guiTags[0].name : '';
  const latestSaasRelease = latestGuiTag.startsWith('saas-v') ? { date: latestGuiTag.split('-v')[1].replaceAll('.', '-'), tag: latestGuiTag } : saasReleases[0];
  return latestSaasRelease.date > latestRelease.release_date ? latestSaasRelease.tag : latestRelease.release;
};

export const getLatestReleaseInfo = () => (dispatch, getState) => {
  if (!getState().app.features.isHosted) {
    return Promise.resolve();
  }
  return Promise.all([GeneralApi.get('/versions.json'), GeneralApi.get('/tags.json')])
    .catch(err => {
      console.log('init error:', extractErrorMessage(err));
      return Promise.resolve([{ data: {} }, { data: [] }]);
    })
    .then(([{ data }, { data: guiTags }]) => {
      if (!guiTags.length) {
        return Promise.resolve();
      }
      const { releases, saas } = data;
      const latestRelease = getLatestRelease(getLatestRelease(releases));
      const { latestRepos, latestVersions } = latestRelease.repos.reduce(
        (accu, item) => {
          if (repoKeyMap[item.name]) {
            accu.latestVersions[repoKeyMap[item.name]] = getComparisonCompatibleVersion(item.version);
          }
          accu.latestRepos[item.name] = getComparisonCompatibleVersion(item.version);
          return accu;
        },
        { latestVersions: { ...getState().app.versionInformation }, latestRepos: {} }
      );
      const info = deductSaasState(latestRelease, guiTags, saas);
      return Promise.resolve(
        dispatch({
          type: SET_VERSION_INFORMATION,
          docsVersion: getState().app.docsVersion,
          value: {
            ...latestVersions,
            backend: info,
            GUI: info,
            latestRelease: {
              releaseDate: latestRelease.release_date,
              repos: latestRepos
            }
          }
        })
      );
    });
};

export const setSearchState = searchState => (dispatch, getState) => {
  const currentState = getState().app.searchState;
  let nextState = {
    ...currentState,
    ...searchState,
    sort: {
      ...currentState.sort,
      ...searchState.sort
    }
  };
  let tasks = [];
  // eslint-disable-next-line no-unused-vars
  const { isSearching: currentSearching, deviceIds: currentDevices, searchTotal: currentTotal, ...currentRequestState } = currentState;
  // eslint-disable-next-line no-unused-vars
  const { isSearching: nextSearching, deviceIds: nextDevices, searchTotal: nextTotal, ...nextRequestState } = nextState;
  if (nextRequestState.searchTerm && !deepCompare(currentRequestState, nextRequestState)) {
    nextState.isSearching = true;
    tasks.push(
      dispatch(searchDevices(nextState))
        .then(results => {
          const searchResult = results[results.length - 1];
          return dispatch(setSearchState({ ...searchResult, isSearching: false }));
        })
        .catch(() => dispatch(setSearchState({ isSearching: false, searchTotal: 0 })))
    );
  }
  tasks.push(dispatch({ type: SET_SEARCH_STATE, state: nextState }));
  return Promise.all(tasks);
};
