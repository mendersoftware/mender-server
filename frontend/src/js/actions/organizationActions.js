// Copyright 2020 Northern.tech AS
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
import { jwtDecode } from 'jwt-decode';
import hashString from 'md5';
import Cookies from 'universal-cookie';

import Api, { apiUrl, headerNames } from '../api/general-api';
import { SET_ANNOUNCEMENT, SORTING_OPTIONS, TIMEOUTS, locations } from '../constants/appConstants';
import { DEVICE_LIST_DEFAULTS } from '../constants/deviceConstants';
import {
  RECEIVE_AUDIT_LOGS,
  RECEIVE_CURRENT_CARD,
  RECEIVE_EXTERNAL_DEVICE_INTEGRATIONS,
  RECEIVE_SETUP_INTENT,
  RECEIVE_SSO_CONFIGS,
  RECEIVE_WEBHOOK_EVENTS,
  SET_AUDITLOG_STATE,
  SET_ORGANIZATION,
  SSO_TYPES
} from '../constants/organizationConstants';
import { deepCompare } from '../helpers';
import { getCurrentSession, getTenantCapabilities } from '../selectors';
import { commonErrorFallback, commonErrorHandler, setFirstLoginAfterSignup, setSnackbar } from './appActions';
import { deviceAuthV2, iotManagerBaseURL } from './deviceActions';

const cookies = new Cookies();
export const auditLogsApiUrl = `${apiUrl.v1}/auditlogs`;
export const tenantadmApiUrlv1 = `${apiUrl.v1}/tenantadm`;
export const tenantadmApiUrlv2 = `${apiUrl.v2}/tenantadm`;
export const ssoIdpApiUrlv1 = `${apiUrl.v1}/useradm/sso/idp/metadata`;

const { page: defaultPage, perPage: defaultPerPage } = DEVICE_LIST_DEFAULTS;

export const cancelRequest = (tenantId, reason) => dispatch =>
  Api.post(`${tenantadmApiUrlv2}/tenants/${tenantId}/cancel`, { reason: reason }).then(() =>
    Promise.resolve(dispatch(setSnackbar('Deactivation request was sent successfully', TIMEOUTS.fiveSeconds, '')))
  );

export const getTargetLocation = key => {
  if (devLocations.includes(window.location.hostname)) {
    return '';
  }
  let subdomainSections = window.location.hostname.substring(0, window.location.hostname.indexOf(locations.us.location)).split('.');
  subdomainSections = subdomainSections.splice(0, subdomainSections.length - 1);
  if (!subdomainSections.find(section => section === key)) {
    subdomainSections = key === locations.us.key ? subdomainSections.filter(section => !locations[section]) : [...subdomainSections, key];
    return `https://${[...subdomainSections, ...locations.us.location.split('.')].join('.')}`;
  }
  return `https://${window.location.hostname}`;
};

const devLocations = ['localhost', 'docker.mender.io'];
export const createOrganizationTrial = data => dispatch => {
  const { key } = locations[data.location];
  const targetLocation = getTargetLocation(key);
  const target = `${targetLocation}${tenantadmApiUrlv2}/tenants/trial`;
  return Api.postUnauthorized(target, data)
    .catch(err => {
      if (err.response.status >= 400 && err.response.status < 500) {
        dispatch(setSnackbar(err.response.data.error, TIMEOUTS.fiveSeconds, ''));
        return Promise.reject(err);
      }
    })
    .then(({ headers }) => {
      cookies.remove('oauth');
      cookies.remove('externalID');
      cookies.remove('email');
      dispatch(setFirstLoginAfterSignup(true));
      return new Promise(resolve =>
        setTimeout(() => {
          window.location.assign(`${targetLocation}${headers.location || ''}`);
          return resolve();
        }, TIMEOUTS.fiveSeconds)
      );
    });
};

export const startCardUpdate = () => dispatch =>
  Api.post(`${tenantadmApiUrlv2}/billing/card`)
    .then(res => {
      dispatch({
        type: RECEIVE_SETUP_INTENT,
        intentId: res.data.intent_id
      });
      return Promise.resolve(res.data.secret);
    })
    .catch(err => commonErrorHandler(err, `Updating the card failed:`, dispatch));

export const confirmCardUpdate = () => (dispatch, getState) =>
  Api.post(`${tenantadmApiUrlv2}/billing/card/${getState().organization.intentId}/confirm`)
    .then(() =>
      Promise.all([
        dispatch(setSnackbar('Payment card was updated successfully')),
        dispatch({
          type: RECEIVE_SETUP_INTENT,
          intentId: null
        })
      ])
    )
    .catch(err => commonErrorHandler(err, `Updating the card failed:`, dispatch));

export const getCurrentCard = () => dispatch =>
  Api.get(`${tenantadmApiUrlv2}/billing`).then(res => {
    const { last4, exp_month, exp_year, brand } = res.data.card || {};
    return Promise.resolve(
      dispatch({
        type: RECEIVE_CURRENT_CARD,
        card: {
          brand,
          last4,
          expiration: { month: exp_month, year: exp_year }
        }
      })
    );
  });

export const startUpgrade = tenantId => dispatch =>
  Api.post(`${tenantadmApiUrlv2}/tenants/${tenantId}/upgrade/start`)
    .then(({ data }) => Promise.resolve(data.secret))
    .catch(err => commonErrorHandler(err, `There was an error upgrading your account:`, dispatch));

export const cancelUpgrade = tenantId => () => Api.post(`${tenantadmApiUrlv2}/tenants/${tenantId}/upgrade/cancel`);

export const completeUpgrade = (tenantId, plan) => dispatch =>
  Api.post(`${tenantadmApiUrlv2}/tenants/${tenantId}/upgrade/complete`, { plan })
    .catch(err => commonErrorHandler(err, `There was an error upgrading your account:`, dispatch))
    .then(() => Promise.resolve(dispatch(getUserOrganization())));

const prepareAuditlogQuery = ({ startDate, endDate, user: userFilter, type, detail: detailFilter, sort = {} }) => {
  const userId = userFilter?.id || userFilter;
  const detail = detailFilter?.id || detailFilter;
  const createdAfter = startDate ? `&created_after=${Math.round(Date.parse(startDate) / 1000)}` : '';
  const createdBefore = endDate ? `&created_before=${Math.round(Date.parse(endDate) / 1000)}` : '';
  const typeSearch = type ? `&object_type=${type.value}`.toLowerCase() : '';
  const userSearch = userId ? `&actor_id=${userId}` : '';
  const objectSearch = type && detail ? `&${type.queryParameter}=${encodeURIComponent(detail)}` : '';
  const { direction = SORTING_OPTIONS.desc } = sort;
  return `${createdAfter}${createdBefore}${userSearch}${typeSearch}${objectSearch}&sort=${direction}`;
};

export const getAuditLogs = selectionState => (dispatch, getState) => {
  const { page, perPage } = selectionState;
  const { hasAuditlogs } = getTenantCapabilities(getState());
  if (!hasAuditlogs) {
    return Promise.resolve();
  }
  return Api.get(`${auditLogsApiUrl}/logs?page=${page}&per_page=${perPage}${prepareAuditlogQuery(selectionState)}`)
    .then(res => {
      let total = res.headers[headerNames.total];
      total = Number(total || res.data.length);
      return Promise.resolve(dispatch({ type: RECEIVE_AUDIT_LOGS, events: res.data, total }));
    })
    .catch(err => commonErrorHandler(err, `There was an error retrieving audit logs:`, dispatch));
};

export const getAuditLogsCsvLink = () => (dispatch, getState) =>
  Promise.resolve(`${auditLogsApiUrl}/logs/export?limit=20000${prepareAuditlogQuery(getState().organization.auditlog.selectionState)}`);

export const setAuditlogsState = selectionState => (dispatch, getState) => {
  const currentState = getState().organization.auditlog.selectionState;
  let nextState = {
    ...currentState,
    ...selectionState,
    sort: { ...currentState.sort, ...selectionState.sort }
  };
  let tasks = [];
  // eslint-disable-next-line no-unused-vars
  const { isLoading: currentLoading, selectedIssue: currentIssue, ...currentRequestState } = currentState;
  // eslint-disable-next-line no-unused-vars
  const { isLoading: selectionLoading, selectedIssue: selectionIssue, ...selectionRequestState } = nextState;
  if (!deepCompare(currentRequestState, selectionRequestState)) {
    nextState.isLoading = true;
    tasks.push(dispatch(getAuditLogs(nextState)).finally(() => dispatch(setAuditlogsState({ isLoading: false }))));
  }
  tasks.push(dispatch({ type: SET_AUDITLOG_STATE, state: nextState }));
  return Promise.all(tasks);
};

/*
  Tenant management + Hosted Mender
*/
export const tenantDataDivergedMessage = 'The system detected there is a change in your plan or purchased add-ons. Please log out and log in again';
export const getUserOrganization = () => (dispatch, getState) =>
  Api.get(`${tenantadmApiUrlv1}/user/tenant`).then(res => {
    let tasks = [dispatch({ type: SET_ORGANIZATION, organization: res.data })];
    const { addons, plan, trial } = res.data;
    const { token } = getCurrentSession(getState());
    const jwt = jwtDecode(token);
    const jwtData = { addons: jwt['mender.addons'], plan: jwt['mender.plan'], trial: jwt['mender.trial'] };
    if (!deepCompare({ addons, plan, trial }, jwtData)) {
      const hash = hashString(tenantDataDivergedMessage);
      cookies.remove(`${jwt.sub}${hash}`);
      tasks.push(dispatch({ type: SET_ANNOUNCEMENT, announcement: tenantDataDivergedMessage }));
    }
    return Promise.all(tasks);
  });

export const sendSupportMessage = content => dispatch =>
  Api.post(`${tenantadmApiUrlv2}/contact/support`, content)
    .catch(err => commonErrorHandler(err, 'There was an error sending your request', dispatch, commonErrorFallback))
    .then(() => Promise.resolve(dispatch(setSnackbar('Your request was sent successfully', TIMEOUTS.fiveSeconds, ''))));

export const requestPlanChange = (tenantId, content) => dispatch =>
  Api.post(`${tenantadmApiUrlv2}/tenants/${tenantId}/plan`, content)
    .catch(err => commonErrorHandler(err, 'There was an error sending your request', dispatch, commonErrorFallback))
    .then(() => Promise.resolve(dispatch(setSnackbar('Your request was sent successfully', TIMEOUTS.fiveSeconds, ''))));

export const downloadLicenseReport = () => dispatch =>
  Api.get(`${deviceAuthV2}/reports/devices`)
    .catch(err => commonErrorHandler(err, 'There was an error downloading the report', dispatch, commonErrorFallback))
    .then(res => res.data);

export const createIntegration = integration => dispatch => {
  // eslint-disable-next-line no-unused-vars
  const { credentials, id, provider, ...remainder } = integration;
  return Api.post(`${iotManagerBaseURL}/integrations`, { provider, credentials, ...remainder })
    .catch(err => commonErrorHandler(err, 'There was an error creating the integration', dispatch, commonErrorFallback))
    .then(() => Promise.all([dispatch(setSnackbar('The integration was set up successfully')), dispatch(getIntegrations())]));
};

export const changeIntegration = integration => dispatch =>
  Api.put(`${iotManagerBaseURL}/integrations/${integration.id}/credentials`, integration.credentials)
    .catch(err => commonErrorHandler(err, 'There was an error updating the integration', dispatch, commonErrorFallback))
    .then(() => Promise.all([dispatch(setSnackbar('The integration was updated successfully')), dispatch(getIntegrations())]));

export const deleteIntegration = integration => (dispatch, getState) =>
  Api.delete(`${iotManagerBaseURL}/integrations/${integration.id}`, {})
    .catch(err => commonErrorHandler(err, 'There was an error removing the integration', dispatch, commonErrorFallback))
    .then(() => {
      const integrations = getState().organization.externalDeviceIntegrations.filter(item => integration.provider !== item.provider);
      return Promise.all([
        dispatch(setSnackbar('The integration was removed successfully')),
        dispatch({ type: RECEIVE_EXTERNAL_DEVICE_INTEGRATIONS, value: integrations })
      ]);
    });

export const getIntegrations = () => (dispatch, getState) =>
  Api.get(`${iotManagerBaseURL}/integrations`)
    .catch(err => commonErrorHandler(err, 'There was an error retrieving the integration', dispatch, commonErrorFallback))
    .then(({ data }) => {
      const existingIntegrations = getState().organization.externalDeviceIntegrations;
      const integrations = data.reduce((accu, item) => {
        const existingIntegration = existingIntegrations.find(integration => item.id === integration.id) ?? {};
        const integration = { ...existingIntegration, ...item };
        accu.push(integration);
        return accu;
      }, []);
      return Promise.resolve(dispatch({ type: RECEIVE_EXTERNAL_DEVICE_INTEGRATIONS, value: integrations }));
    });

export const getWebhookEvents =
  (config = {}) =>
  (dispatch, getState) => {
    const { isFollowUp, page = defaultPage, perPage = defaultPerPage } = config;
    return Api.get(`${iotManagerBaseURL}/events?page=${page}&per_page=${perPage}`)
      .catch(err => commonErrorHandler(err, 'There was an error retrieving activity for this integration', dispatch, commonErrorFallback))
      .then(({ data }) => {
        let tasks = [
          dispatch({
            type: RECEIVE_WEBHOOK_EVENTS,
            value: isFollowUp ? getState().organization.webhooks.events : data,
            total: (page - 1) * perPage + data.length
          })
        ];
        if (data.length >= perPage && !isFollowUp) {
          tasks.push(dispatch(getWebhookEvents({ isFollowUp: true, page: page + 1, perPage: 1 })));
        }
        return Promise.all(tasks);
      });
  };

const ssoConfigActions = {
  create: { success: 'stored', error: 'storing' },
  edit: { success: 'updated', error: 'updating' },
  read: { success: '', error: 'retrieving' },
  remove: { success: 'removed', error: 'removing' }
};

const ssoConfigActionErrorHandler = (err, type) => dispatch =>
  commonErrorHandler(err, `There was an error ${ssoConfigActions[type].error} the SSO configuration.`, dispatch, commonErrorFallback);

const ssoConfigActionSuccessHandler = type => dispatch => dispatch(setSnackbar(`The SSO configuration was ${ssoConfigActions[type].success} successfully`));

export const storeSsoConfig =
  ({ config, contentType }) =>
  dispatch =>
    Api.post(ssoIdpApiUrlv1, config, { headers: { 'Content-Type': contentType, Accept: 'application/json' } })
      .catch(err => dispatch(ssoConfigActionErrorHandler(err, 'create')))
      .then(() => Promise.all([dispatch(ssoConfigActionSuccessHandler('create')), dispatch(getSsoConfigs())]));

export const changeSsoConfig =
  ({ id, config, contentType }) =>
  dispatch =>
    Api.put(`${ssoIdpApiUrlv1}/${id}`, config, { headers: { 'Content-Type': contentType, Accept: 'application/json' } })
      .catch(err => dispatch(ssoConfigActionErrorHandler(err, 'edit')))
      .then(() => Promise.all([dispatch(ssoConfigActionSuccessHandler('edit')), dispatch(getSsoConfigs())]));

export const deleteSsoConfig =
  ({ id }) =>
  (dispatch, getState) =>
    Api.delete(`${ssoIdpApiUrlv1}/${id}`)
      .catch(err => dispatch(ssoConfigActionErrorHandler(err, 'remove')))
      .then(() => {
        const configs = getState().organization.ssoConfigs.filter(item => id !== item.id);
        return Promise.all([dispatch(ssoConfigActionSuccessHandler('remove')), dispatch({ type: RECEIVE_SSO_CONFIGS, value: configs })]);
      });

const getSsoConfigById = config => dispatch =>
  Api.get(`${ssoIdpApiUrlv1}/${config.id}`)
    .catch(err => dispatch(ssoConfigActionErrorHandler(err, 'read')))
    .then(({ data, headers }) => {
      const sso = Object.values(SSO_TYPES).find(({ contentType }) => contentType === headers['content-type']);
      return sso ? Promise.resolve({ ...config, config: data, type: sso.id }) : Promise.reject('Unsupported SSO config content type.');
    });

export const getSsoConfigs = () => dispatch =>
  Api.get(ssoIdpApiUrlv1)
    .catch(err => commonErrorHandler(err, 'There was an error retrieving SSO configurations', dispatch, commonErrorFallback))
    .then(({ data }) =>
      Promise.all(data.map(config => Promise.resolve(dispatch(getSsoConfigById(config)))))
        .then(configs => dispatch({ type: RECEIVE_SSO_CONFIGS, value: configs }))
        .catch(err => commonErrorHandler(err, err, dispatch, ''))
    );
