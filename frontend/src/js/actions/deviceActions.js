/*eslint import/namespace: ['error', { allowComputed: true }]*/
import React from 'react';
import { Link } from 'react-router-dom';

import { isCancel } from 'axios';
import pluralize from 'pluralize';
import { v4 as uuid } from 'uuid';

import { commonErrorFallback, commonErrorHandler, setSnackbar } from '../actions/appActions';
import { getSingleDeployment } from '../actions/deploymentActions';
import { auditLogsApiUrl } from '../actions/organizationActions';
import { cleanUpUpload, progress } from '../actions/releaseActions';
import { saveGlobalSettings } from '../actions/userActions';
import GeneralApi, { MAX_PAGE_SIZE, apiUrl, headerNames } from '../api/general-api';
import { routes, sortingAlternatives } from '../components/devices/base-devices';
import { filtersFilter } from '../components/devices/widgets/filters';
import { SORTING_OPTIONS, TIMEOUTS, UPLOAD_PROGRESS, emptyChartSelection, yes } from '../constants/appConstants';
import * as DeviceConstants from '../constants/deviceConstants';
import { rootfsImageVersion } from '../constants/releaseConstants';
import { attributeDuplicateFilter, deepCompare, extractErrorMessage, getSnackbarMessage, mapDeviceAttributes } from '../helpers';
import {
  getDeviceById as getDeviceByIdSelector,
  getDeviceFilters,
  getDeviceTwinIntegrations,
  getGroups as getGroupsSelector,
  getIdAttribute,
  getTenantCapabilities,
  getUserCapabilities,
  getUserSettings
} from '../selectors';
import { chartColorPalette } from '../themes/Mender';
import { getDeviceMonitorConfig, getLatestDeviceAlerts } from './monitorActions';

const { DEVICE_FILTERING_OPTIONS, DEVICE_STATES, DEVICE_LIST_DEFAULTS, UNGROUPED_GROUP, emptyFilter } = DeviceConstants;
const { page: defaultPage, perPage: defaultPerPage } = DEVICE_LIST_DEFAULTS;

export const deviceAuthV2 = `${apiUrl.v2}/devauth`;
export const deviceConnect = `${apiUrl.v1}/deviceconnect`;
export const inventoryApiUrl = `${apiUrl.v1}/inventory`;
export const inventoryApiUrlV2 = `${apiUrl.v2}/inventory`;
export const deviceConfig = `${apiUrl.v1}/deviceconfig/configurations/device`;
export const reportingApiUrl = `${apiUrl.v1}/reporting`;
export const iotManagerBaseURL = `${apiUrl.v1}/iot-manager`;

const defaultAttributes = [
  { scope: 'identity', attribute: 'status' },
  { scope: 'inventory', attribute: 'artifact_name' },
  { scope: 'inventory', attribute: 'device_type' },
  { scope: 'inventory', attribute: 'mender_is_gateway' },
  { scope: 'inventory', attribute: 'mender_gateway_system_id' },
  { scope: 'inventory', attribute: rootfsImageVersion },
  { scope: 'monitor', attribute: 'alerts' },
  { scope: 'system', attribute: 'created_ts' },
  { scope: 'system', attribute: 'updated_ts' },
  { scope: 'system', attribute: 'check_in_time' },
  { scope: 'system', attribute: 'group' },
  { scope: 'tags', attribute: 'name' }
];

export const getSearchEndpoint = hasReporting => (hasReporting ? `${reportingApiUrl}/devices/search` : `${inventoryApiUrlV2}/filters/search`);

const getAttrsEndpoint = hasReporting => (hasReporting ? `${reportingApiUrl}/devices/search/attributes` : `${inventoryApiUrlV2}/filters/attributes`);

export const getGroups = () => (dispatch, getState) =>
  GeneralApi.get(`${inventoryApiUrl}/groups`).then(res => {
    const state = getState().devices.groups.byId;
    const dynamicGroups = Object.entries(state).reduce((accu, [id, group]) => {
      if (group.id || (group.filters?.length && id !== UNGROUPED_GROUP.id)) {
        accu[id] = group;
      }
      return accu;
    }, {});
    const groups = res.data.reduce((accu, group) => {
      accu[group] = { deviceIds: [], filters: [], total: 0, ...state[group] };
      return accu;
    }, dynamicGroups);
    const filters = [{ key: 'group', value: res.data, operator: DEVICE_FILTERING_OPTIONS.$nin.key, scope: 'system' }];
    return Promise.all([
      dispatch({ type: DeviceConstants.RECEIVE_GROUPS, groups }),
      dispatch(getDevicesByStatus(undefined, { filterSelection: filters, group: 0, page: 1, perPage: 1 }))
    ]).then(promises => {
      const ungroupedDevices = promises[promises.length - 1] || [];
      const result = ungroupedDevices[ungroupedDevices.length - 1] || {};
      if (!result.total) {
        return Promise.resolve();
      }
      return Promise.resolve(
        dispatch({
          type: DeviceConstants.ADD_DYNAMIC_GROUP,
          groupName: UNGROUPED_GROUP.id,
          group: {
            deviceIds: [],
            total: 0,
            ...getState().devices.groups.byId[UNGROUPED_GROUP.id],
            filters: [{ key: 'group', value: res.data, operator: DEVICE_FILTERING_OPTIONS.$nin.key, scope: 'system' }]
          }
        })
      );
    });
  });

export const addDevicesToGroup = (group, deviceIds, isCreation) => dispatch =>
  GeneralApi.patch(`${inventoryApiUrl}/groups/${group}/devices`, deviceIds)
    .then(() => dispatch({ type: DeviceConstants.ADD_TO_GROUP, group, deviceIds }))
    .finally(() => (isCreation ? Promise.resolve(dispatch(getGroups())) : {}));

export const removeDevicesFromGroup = (group, deviceIds) => dispatch =>
  GeneralApi.delete(`${inventoryApiUrl}/groups/${group}/devices`, deviceIds).then(() =>
    Promise.all([
      dispatch({
        type: DeviceConstants.REMOVE_FROM_GROUP,
        group,
        deviceIds
      }),
      dispatch(setSnackbar(`The ${pluralize('devices', deviceIds.length)} ${pluralize('were', deviceIds.length)} removed from the group`, TIMEOUTS.fiveSeconds))
    ])
  );

const getGroupNotification = (newGroup, selectedGroup) => {
  const successMessage = 'The group was updated successfully';
  if (newGroup === selectedGroup) {
    return [successMessage, TIMEOUTS.fiveSeconds];
  }
  return [
    <>
      {successMessage} - <Link to={`/devices?inventory=group:eq:${newGroup}`}>click here</Link> to see it.
    </>,
    5000,
    undefined,
    undefined,
    () => {}
  ];
};

export const addStaticGroup = (group, devices) => (dispatch, getState) =>
  Promise.resolve(
    dispatch(
      addDevicesToGroup(
        group,
        devices.map(({ id }) => id),
        true
      )
    )
  )
    .then(() =>
      Promise.resolve(
        dispatch({
          type: DeviceConstants.ADD_STATIC_GROUP,
          group: { deviceIds: [], total: 0, filters: [], ...getState().devices.groups.byId[group] },
          groupName: group
        })
      ).then(() =>
        Promise.all([
          dispatch(setDeviceListState({ setOnly: true })),
          dispatch(getGroups()),
          dispatch(setSnackbar(...getGroupNotification(group, getState().devices.groups.selectedGroup)))
        ])
      )
    )
    .catch(err => commonErrorHandler(err, `Group could not be updated:`, dispatch));

export const removeStaticGroup = groupName => (dispatch, getState) => {
  return GeneralApi.delete(`${inventoryApiUrl}/groups/${groupName}`).then(() => {
    const selectedGroup = getState().devices.groups.selectedGroup === groupName ? undefined : getState().devices.groups.selectedGroup;
    // eslint-disable-next-line no-unused-vars
    const { [groupName]: removal, ...groups } = getState().devices.groups.byId;
    return Promise.all([
      dispatch({
        type: DeviceConstants.REMOVE_STATIC_GROUP,
        groups
      }),
      dispatch(getGroups()),
      dispatch(selectGroup(selectedGroup)),
      dispatch(setSnackbar('Group was removed successfully', TIMEOUTS.fiveSeconds))
    ]);
  });
};

// for some reason these functions can not be stored in the deviceConstants...
const filterProcessors = {
  $gt: val => Number(val) || val,
  $gte: val => Number(val) || val,
  $lt: val => Number(val) || val,
  $lte: val => Number(val) || val,
  $in: val => ('' + val).split(',').map(i => i.trim()),
  $nin: val => ('' + val).split(',').map(i => i.trim()),
  $exists: yes,
  $nexists: () => false
};
const filterAliases = {
  $nexists: { alias: DEVICE_FILTERING_OPTIONS.$exists.key, value: false }
};
export const mapFiltersToTerms = (filters = []) =>
  filters.map(filter => ({
    scope: filter.scope,
    attribute: filter.key,
    type: filterAliases[filter.operator]?.alias || filter.operator,
    value: filterProcessors.hasOwnProperty(filter.operator) ? filterProcessors[filter.operator](filter.value) : filter.value
  }));
export const mapTermsToFilters = (terms = []) =>
  terms.map(term => {
    const aliasedFilter = Object.entries(filterAliases).find(
      aliasDefinition => aliasDefinition[1].alias === term.type && aliasDefinition[1].value === term.value
    );
    const operator = aliasedFilter ? aliasedFilter[0] : term.type;
    return { scope: term.scope, key: term.attribute, operator, value: term.value };
  });

export const getDynamicGroups = () => (dispatch, getState) =>
  GeneralApi.get(`${inventoryApiUrlV2}/filters?per_page=${MAX_PAGE_SIZE}`)
    .then(({ data: filters }) => {
      const state = getState().devices.groups.byId;
      const staticGroups = Object.entries(state).reduce((accu, [id, group]) => {
        if (!(group.id || group.filters?.length)) {
          accu[id] = group;
        }
        return accu;
      }, {});
      const groups = (filters || []).reduce((accu, filter) => {
        accu[filter.name] = {
          deviceIds: [],
          total: 0,
          ...state[filter.name],
          id: filter.id,
          filters: mapTermsToFilters(filter.terms)
        };
        return accu;
      }, staticGroups);
      return Promise.resolve(dispatch({ type: DeviceConstants.RECEIVE_DYNAMIC_GROUPS, groups }));
    })
    .catch(() => console.log('Dynamic group retrieval failed - likely accessing a non-enterprise backend'));

export const addDynamicGroup = (groupName, filterPredicates) => (dispatch, getState) =>
  GeneralApi.post(`${inventoryApiUrlV2}/filters`, { name: groupName, terms: mapFiltersToTerms(filterPredicates) })
    .then(res =>
      Promise.resolve(
        dispatch({
          type: DeviceConstants.ADD_DYNAMIC_GROUP,
          groupName,
          group: {
            deviceIds: [],
            total: 0,
            ...getState().devices.groups.byId[groupName],
            id: res.headers[headerNames.location].substring(res.headers[headerNames.location].lastIndexOf('/') + 1),
            filters: filterPredicates
          }
        })
      ).then(() => {
        const { cleanedFilters } = getGroupFilters(groupName, getState().devices.groups);
        return Promise.all([
          dispatch(setDeviceFilters(cleanedFilters)),
          dispatch(setSnackbar(...getGroupNotification(groupName, getState().devices.groups.selectedGroup))),
          dispatch(getDynamicGroups())
        ]);
      })
    )
    .catch(err => commonErrorHandler(err, `Group could not be updated:`, dispatch));

export const updateDynamicGroup = (groupName, filterPredicates) => (dispatch, getState) => {
  const filterId = getState().devices.groups.byId[groupName].id;
  return GeneralApi.delete(`${inventoryApiUrlV2}/filters/${filterId}`).then(() => Promise.resolve(dispatch(addDynamicGroup(groupName, filterPredicates))));
};

export const removeDynamicGroup = groupName => (dispatch, getState) => {
  let groups = { ...getState().devices.groups.byId };
  const filterId = groups[groupName].id;
  const selectedGroup = getState().devices.groups.selectedGroup === groupName ? undefined : getState().devices.groups.selectedGroup;
  return Promise.all([GeneralApi.delete(`${inventoryApiUrlV2}/filters/${filterId}`), dispatch(selectGroup(selectedGroup))]).then(() => {
    delete groups[groupName];
    return Promise.all([
      dispatch({
        type: DeviceConstants.REMOVE_DYNAMIC_GROUP,
        groups
      }),
      dispatch(setSnackbar('Group was removed successfully', TIMEOUTS.fiveSeconds))
    ]);
  });
};
/*
 * Device inventory functions
 */
const getGroupFilters = (group, groupsState, filters = []) => {
  const groupName = group === UNGROUPED_GROUP.id || group === UNGROUPED_GROUP.name ? UNGROUPED_GROUP.id : group;
  const selectedGroup = groupsState.byId[groupName];
  const groupFilterLength = selectedGroup?.filters?.length || 0;
  const cleanedFilters = groupFilterLength ? [...filters, ...selectedGroup.filters].filter(filtersFilter) : filters;
  return { cleanedFilters, groupName, selectedGroup, groupFilterLength };
};

export const selectGroup =
  (group, filters = []) =>
  (dispatch, getState) => {
    const { cleanedFilters, groupName, selectedGroup, groupFilterLength } = getGroupFilters(group, getState().devices.groups, filters);
    const state = getState();
    if (state.devices.groups.selectedGroup === groupName && ((filters.length === 0 && !groupFilterLength) || filters.length === cleanedFilters.length)) {
      return Promise.resolve();
    }
    let tasks = [];
    if (groupFilterLength) {
      tasks.push(dispatch(setDeviceFilters(cleanedFilters)));
    } else {
      tasks.push(dispatch(setDeviceFilters(filters)));
      tasks.push(dispatch(getGroupDevices(groupName, { perPage: 1, shouldIncludeAllStates: true })));
    }
    const selectedGroupName = selectedGroup || !Object.keys(state.devices.groups.byId).length ? groupName : undefined;
    tasks.push(dispatch({ type: DeviceConstants.SELECT_GROUP, group: selectedGroupName }));
    return Promise.all(tasks);
  };

const getEarliestTs = (dateA = '', dateB = '') => (!dateA || !dateB ? dateA || dateB : dateA < dateB ? dateA : dateB);

const reduceReceivedDevices = (devices, ids, state, status) =>
  devices.reduce(
    (accu, device) => {
      const stateDevice = getDeviceByIdSelector(state, device.id);
      const {
        attributes: storedAttributes = {},
        identity_data: storedIdentity = {},
        monitor: storedMonitor = {},
        tags: storedTags = {},
        group: storedGroup
      } = stateDevice;
      const { identity, inventory, monitor, system = {}, tags } = mapDeviceAttributes(device.attributes);
      device.tags = { ...storedTags, ...tags };
      device.group = system.group ?? storedGroup;
      device.monitor = { ...storedMonitor, ...monitor };
      device.identity_data = { ...storedIdentity, ...identity, ...(device.identity_data ? device.identity_data : {}) };
      device.status = status ? status : device.status || identity.status;
      device.check_in_time_rounded = system.check_in_time ?? stateDevice.check_in_time_rounded;
      device.check_in_time_exact = device.check_in_time ?? stateDevice.check_in_time_exact;
      device.created_ts = getEarliestTs(getEarliestTs(system.created_ts, device.created_ts), stateDevice.created_ts);
      device.updated_ts = device.attributes ? device.updated_ts : stateDevice.updated_ts;
      device.isNew = new Date(device.created_ts) > new Date(state.app.newThreshold);
      device.isOffline = new Date(device.check_in_time_rounded) < new Date(state.app.offlineThreshold) || device.check_in_time_rounded === undefined;
      // all the other mapped attributes return as empty objects if there are no attributes to map, but identity will be initialized with an empty state
      // for device_type and artifact_name, potentially overwriting existing info, so rely on stored information instead if there are no attributes
      device.attributes = device.attributes ? { ...storedAttributes, ...inventory } : storedAttributes;
      accu.devicesById[device.id] = { ...stateDevice, ...device };
      accu.ids.push(device.id);
      return accu;
    },
    { ids, devicesById: {} }
  );

export const getGroupDevices =
  (group, options = {}) =>
  (dispatch, getState) => {
    const { shouldIncludeAllStates, ...remainder } = options;
    const { cleanedFilters: filterSelection } = getGroupFilters(group, getState().devices.groups);
    return Promise.resolve(
      dispatch(getDevicesByStatus(shouldIncludeAllStates ? undefined : DEVICE_STATES.accepted, { ...remainder, filterSelection, group }))
    ).then(results => {
      if (!group) {
        return Promise.resolve();
      }
      const { deviceAccu, total } = results[results.length - 1];
      const stateGroup = getState().devices.groups.byId[group];
      if (!stateGroup && !total && !deviceAccu.ids.length) {
        return Promise.resolve();
      }
      return Promise.resolve(
        dispatch({
          type: DeviceConstants.RECEIVE_GROUP_DEVICES,
          group: {
            filters: [],
            ...stateGroup,
            deviceIds: deviceAccu.ids.length === total || deviceAccu.ids.length > stateGroup?.deviceIds ? deviceAccu.ids : stateGroup.deviceIds,
            total
          },
          groupName: group
        })
      );
    });
  };

export const getAllGroupDevices = (group, shouldIncludeAllStates) => (dispatch, getState) => {
  if (!group || (!!group && (!getState().devices.groups.byId[group] || getState().devices.groups.byId[group].filters.length))) {
    return Promise.resolve();
  }
  const { attributes, filterTerms } = prepareSearchArguments({
    filters: [],
    group,
    state: getState(),
    status: shouldIncludeAllStates ? undefined : DEVICE_STATES.accepted
  });
  const getAllDevices = (perPage = MAX_PAGE_SIZE, page = defaultPage, devices = []) =>
    GeneralApi.post(getSearchEndpoint(getState().app.features.hasReporting), {
      page,
      per_page: perPage,
      filters: filterTerms,
      attributes
    }).then(res => {
      const state = getState();
      const deviceAccu = reduceReceivedDevices(res.data, devices, state);
      dispatch({
        type: DeviceConstants.RECEIVE_DEVICES,
        devicesById: deviceAccu.devicesById
      });
      const total = Number(res.headers[headerNames.total]);
      if (total > perPage * page) {
        return getAllDevices(perPage, page + 1, deviceAccu.ids);
      }
      return Promise.resolve(
        dispatch({
          type: DeviceConstants.RECEIVE_GROUP_DEVICES,
          group: {
            filters: [],
            ...state.devices.groups.byId[group],
            deviceIds: deviceAccu.ids,
            total: deviceAccu.ids.length
          },
          groupName: group
        })
      );
    });
  return getAllDevices();
};

export const getAllDynamicGroupDevices = group => (dispatch, getState) => {
  if (!!group && (!getState().devices.groups.byId[group] || !getState().devices.groups.byId[group].filters.length)) {
    return Promise.resolve();
  }
  const { attributes, filterTerms: filters } = prepareSearchArguments({
    filters: getState().devices.groups.byId[group].filters,
    state: getState(),
    status: DEVICE_STATES.accepted
  });
  const getAllDevices = (perPage = MAX_PAGE_SIZE, page = defaultPage, devices = []) =>
    GeneralApi.post(getSearchEndpoint(getState().app.features.hasReporting), { page, per_page: perPage, filters, attributes }).then(res => {
      const state = getState();
      const deviceAccu = reduceReceivedDevices(res.data, devices, state);
      dispatch({
        type: DeviceConstants.RECEIVE_DEVICES,
        devicesById: deviceAccu.devicesById
      });
      const total = Number(res.headers[headerNames.total]);
      if (total > deviceAccu.ids.length) {
        return getAllDevices(perPage, page + 1, deviceAccu.ids);
      }
      return Promise.resolve(
        dispatch({
          type: DeviceConstants.RECEIVE_GROUP_DEVICES,
          group: {
            ...state.devices.groups.byId[group],
            deviceIds: deviceAccu.ids,
            total
          },
          groupName: group
        })
      );
    });
  return getAllDevices();
};

export const setDeviceFilters = filters => (dispatch, getState) => {
  if (deepCompare(filters, getDeviceFilters(getState()))) {
    return Promise.resolve();
  }
  return Promise.resolve(dispatch({ type: DeviceConstants.SET_DEVICE_FILTERS, filters }));
};

export const getDeviceById = id => (dispatch, getState) =>
  GeneralApi.get(`${inventoryApiUrl}/devices/${id}`)
    .then(res => {
      const device = reduceReceivedDevices([res.data], [], getState()).devicesById[id];
      device.etag = res.headers.etag;
      dispatch({ type: DeviceConstants.RECEIVE_DEVICE, device });
      return Promise.resolve(device);
    })
    .catch(err => {
      const errMsg = extractErrorMessage(err);
      if (errMsg.includes('Not Found')) {
        console.log(`${id} does not have any inventory information`);
        const device = reduceReceivedDevices(
          [
            {
              id,
              attributes: [
                { name: 'status', value: 'decomissioned', scope: 'identity' },
                { name: 'decomissioned', value: 'true', scope: 'inventory' }
              ]
            }
          ],
          [],
          getState()
        ).devicesById[id];
        dispatch({ type: DeviceConstants.RECEIVE_DEVICE, device });
      }
    });

export const getDeviceInfo = deviceId => (dispatch, getState) => {
  const device = getState().devices.byId[deviceId] || {};
  const { hasDeviceConfig, hasDeviceConnect, hasMonitor } = getTenantCapabilities(getState());
  const { canConfigure } = getUserCapabilities(getState());
  const integrations = getDeviceTwinIntegrations(getState());
  let tasks = [dispatch(getDeviceAuth(deviceId)), ...integrations.map(integration => dispatch(getDeviceTwin(deviceId, integration)))];
  if (hasDeviceConfig && canConfigure && [DEVICE_STATES.accepted, DEVICE_STATES.preauth].includes(device.status)) {
    tasks.push(dispatch(getDeviceConfig(deviceId)));
  }
  if (device.status === DEVICE_STATES.accepted) {
    // Get full device identity details for single selected device
    tasks.push(dispatch(getDeviceById(deviceId)));
    if (hasDeviceConnect) {
      tasks.push(dispatch(getDeviceConnect(deviceId)));
    }
    if (hasMonitor) {
      tasks.push(dispatch(getLatestDeviceAlerts(deviceId)));
      tasks.push(dispatch(getDeviceMonitorConfig(deviceId)));
    }
  }
  return Promise.all(tasks);
};

const deriveInactiveDevices = deviceIds => (dispatch, getState) => {
  const yesterday = new Date();
  yesterday.setDate(yesterday.getDate() - 1);
  const yesterdaysIsoString = yesterday.toISOString();
  const state = getState().devices;
  // now boil the list down to the ones that were not updated since yesterday
  const devices = deviceIds.reduce(
    (accu, id) => {
      const device = state.byId[id];
      if (device && device.updated_ts > yesterdaysIsoString) {
        accu.active.push(id);
      } else {
        accu.inactive.push(id);
      }
      return accu;
    },
    { active: [], inactive: [] }
  );
  return dispatch({
    type: DeviceConstants.SET_INACTIVE_DEVICES,
    activeDeviceTotal: devices.active.length,
    inactiveDeviceTotal: devices.inactive.length
  });
};

/*
    Device Auth + admission
  */
export const getDeviceCount = status => (dispatch, getState) =>
  GeneralApi.post(getSearchEndpoint(getState().app.features.hasReporting), {
    page: 1,
    per_page: 1,
    filters: mapFiltersToTerms([{ key: 'status', value: status, operator: DEVICE_FILTERING_OPTIONS.$eq.key, scope: 'identity' }]),
    attributes: defaultAttributes
  }).then(response => {
    const count = Number(response.headers[headerNames.total]);
    switch (status) {
      case DEVICE_STATES.accepted:
      case DEVICE_STATES.pending:
      case DEVICE_STATES.preauth:
      case DEVICE_STATES.rejected:
        return dispatch({ type: DeviceConstants[`SET_${status.toUpperCase()}_DEVICES_COUNT`], count, status });
      default:
        return dispatch({ type: DeviceConstants.SET_TOTAL_DEVICES, count });
    }
  });

export const getAllDeviceCounts = () => dispatch =>
  Promise.all([DEVICE_STATES.accepted, DEVICE_STATES.pending].map(status => dispatch(getDeviceCount(status))));

export const getDeviceLimit = () => dispatch =>
  GeneralApi.get(`${deviceAuthV2}/limits/max_devices`).then(res =>
    dispatch({
      type: DeviceConstants.SET_DEVICE_LIMIT,
      limit: res.data.limit
    })
  );

export const setDeviceListState =
  (selectionState, shouldSelectDevices = true, forceRefresh, fetchAuth = true) =>
  (dispatch, getState) => {
    const currentState = getState().devices.deviceList;
    const refreshTrigger = forceRefresh ? !currentState.refreshTrigger : selectionState.refreshTrigger;
    let nextState = {
      ...currentState,
      setOnly: false,
      refreshTrigger,
      ...selectionState,
      sort: { ...currentState.sort, ...selectionState.sort }
    };
    let tasks = [];
    // eslint-disable-next-line no-unused-vars
    const { isLoading: currentLoading, deviceIds: currentDevices, selection: currentSelection, ...currentRequestState } = currentState;
    // eslint-disable-next-line no-unused-vars
    const { isLoading: nextLoading, deviceIds: nextDevices, selection: nextSelection, ...nextRequestState } = nextState;
    if (!nextState.setOnly && !deepCompare(currentRequestState, nextRequestState)) {
      const { direction: sortDown = SORTING_OPTIONS.desc, key: sortCol, scope: sortScope } = nextState.sort ?? {};
      const sortBy = sortCol ? [{ attribute: sortCol, order: sortDown, scope: sortScope }] : undefined;
      if (sortCol && sortingAlternatives[sortCol]) {
        sortBy.push({ ...sortBy[0], attribute: sortingAlternatives[sortCol] });
      }
      const applicableSelectedState = nextState.state === routes.allDevices.key ? undefined : nextState.state;
      nextState.isLoading = true;
      tasks.push(
        dispatch(getDevicesByStatus(applicableSelectedState, { ...nextState, sortOptions: sortBy }, fetchAuth))
          .then(results => {
            const { deviceAccu, total } = results[results.length - 1];
            const devicesState = shouldSelectDevices
              ? { ...getState().devices.deviceList, deviceIds: deviceAccu.ids, total, isLoading: false }
              : { ...getState().devices.deviceList, isLoading: false };
            return Promise.resolve(dispatch({ type: DeviceConstants.SET_DEVICE_LIST_STATE, state: devicesState }));
          })
          // whatever happens, change "loading" back to null
          .catch(() =>
            Promise.resolve(dispatch({ type: DeviceConstants.SET_DEVICE_LIST_STATE, state: { ...getState().devices.deviceList, isLoading: false } }))
          )
      );
    }
    tasks.push(dispatch({ type: DeviceConstants.SET_DEVICE_LIST_STATE, state: nextState }));
    return Promise.all(tasks);
  };

const convertIssueOptionsToFilters = (issuesSelection, filtersState = {}) =>
  issuesSelection.map(item => {
    if (typeof DeviceConstants.DEVICE_ISSUE_OPTIONS[item].filterRule.value === 'function') {
      return { ...DeviceConstants.DEVICE_ISSUE_OPTIONS[item].filterRule, value: DeviceConstants.DEVICE_ISSUE_OPTIONS[item].filterRule.value(filtersState) };
    }
    return DeviceConstants.DEVICE_ISSUE_OPTIONS[item].filterRule;
  });

export const convertDeviceListStateToFilters = ({ filters = [], group, groups = { byId: {} }, offlineThreshold, selectedIssues = [], status }) => {
  let applicableFilters = [...filters];
  if (typeof group === 'string' && !(groups.byId[group]?.filters || applicableFilters).length) {
    applicableFilters.push({ key: 'group', value: group, operator: DEVICE_FILTERING_OPTIONS.$eq.key, scope: 'system' });
  }
  const nonMonitorFilters = applicableFilters.filter(
    filter =>
      !Object.values(DeviceConstants.DEVICE_ISSUE_OPTIONS).some(
        ({ filterRule }) => filter.scope !== 'inventory' && filterRule.scope === filter.scope && filterRule.key === filter.key
      )
  );
  const deviceIssueFilters = convertIssueOptionsToFilters(selectedIssues, { offlineThreshold });
  applicableFilters = [...nonMonitorFilters, ...deviceIssueFilters];
  const effectiveFilters = status
    ? [...applicableFilters, { key: 'status', value: status, operator: DEVICE_FILTERING_OPTIONS.$eq.key, scope: 'identity' }]
    : applicableFilters;
  return { applicableFilters: nonMonitorFilters, filterTerms: mapFiltersToTerms(effectiveFilters) };
};

// get devices from inventory
export const getDevicesByStatus =
  (status, options = {}, fetchAuth = true) =>
  (dispatch, getState) => {
    const { filterSelection, group, selectedIssues = [], page = defaultPage, perPage = defaultPerPage, sortOptions = [], selectedAttributes = [] } = options;
    const state = getState();
    const { applicableFilters, filterTerms } = convertDeviceListStateToFilters({
      filters: filterSelection ?? getDeviceFilters(state),
      group: group ?? state.devices.groups.selectedGroup,
      groups: state.devices.groups,
      offlineThreshold: state.app.offlineThreshold,
      selectedIssues,
      status
    });
    const attributes = [...defaultAttributes, getIdAttribute(state), ...selectedAttributes];
    return GeneralApi.post(getSearchEndpoint(state.app.features.hasReporting), {
      page,
      per_page: perPage,
      filters: filterTerms,
      sort: sortOptions,
      attributes
    })
      .then(response => {
        const state = getState();
        const deviceAccu = reduceReceivedDevices(response.data, [], state, status);
        let total = !applicableFilters.length ? Number(response.headers[headerNames.total]) : null;
        if (status && state.devices.byStatus[status].total === deviceAccu.ids.length) {
          total = deviceAccu.ids.length;
        }
        let tasks = [
          dispatch({
            type: DeviceConstants.RECEIVE_DEVICES,
            devicesById: deviceAccu.devicesById
          })
        ];
        if (status) {
          tasks.push(
            dispatch({
              type: DeviceConstants[`SET_${status.toUpperCase()}_DEVICES`],
              deviceIds: deviceAccu.ids,
              status,
              total
            })
          );
        }
        // for each device, get device identity info
        const receivedDevices = Object.values(deviceAccu.devicesById);
        if (receivedDevices.length && fetchAuth) {
          tasks.push(dispatch(getDevicesWithAuth(receivedDevices)));
        }
        tasks.push(Promise.resolve({ deviceAccu, total: Number(response.headers[headerNames.total]) }));
        return Promise.all(tasks);
      })
      .catch(err => commonErrorHandler(err, `${status} devices couldn't be loaded.`, dispatch, commonErrorFallback));
  };

export const getAllDevicesByStatus = status => (dispatch, getState) => {
  const attributes = [...defaultAttributes, getIdAttribute(getState())];
  const getAllDevices = (perPage = MAX_PAGE_SIZE, page = 1, devices = []) =>
    GeneralApi.post(getSearchEndpoint(getState().app.features.hasReporting), {
      page,
      per_page: perPage,
      filters: mapFiltersToTerms([{ key: 'status', value: status, operator: DEVICE_FILTERING_OPTIONS.$eq.key, scope: 'identity' }]),
      attributes
    }).then(res => {
      const state = getState();
      const deviceAccu = reduceReceivedDevices(res.data, devices, state, status);
      dispatch({
        type: DeviceConstants.RECEIVE_DEVICES,
        devicesById: deviceAccu.devicesById
      });
      const total = Number(res.headers[headerNames.total]);
      if (total > state.deployments.deploymentDeviceLimit) {
        return Promise.resolve();
      }
      if (total > perPage * page) {
        return getAllDevices(perPage, page + 1, deviceAccu.ids);
      }
      let tasks = [
        dispatch({
          type: DeviceConstants[`SET_${status.toUpperCase()}_DEVICES`],
          deviceIds: deviceAccu.ids,
          forceUpdate: true,
          status,
          total: deviceAccu.ids.length
        })
      ];
      if (status === DEVICE_STATES.accepted && deviceAccu.ids.length === total) {
        tasks.push(dispatch(deriveInactiveDevices(deviceAccu.ids)));
        tasks.push(dispatch(deriveReportsData()));
      }
      return Promise.all(tasks);
    });
  return getAllDevices();
};

export const searchDevices =
  (passedOptions = {}) =>
  (dispatch, getState) => {
    const state = getState();
    let options = { ...state.app.searchState, ...passedOptions };
    const { page = defaultPage, searchTerm, sortOptions = [] } = options;
    const { columnSelection = [] } = getUserSettings(state);
    const selectedAttributes = columnSelection.map(column => ({ attribute: column.key, scope: column.scope }));
    const attributes = attributeDuplicateFilter([...defaultAttributes, getIdAttribute(state), ...selectedAttributes], 'attribute');
    return GeneralApi.post(getSearchEndpoint(state.app.features.hasReporting), {
      page,
      per_page: 10,
      filters: [],
      sort: sortOptions,
      text: searchTerm,
      attributes
    })
      .then(response => {
        const deviceAccu = reduceReceivedDevices(response.data, [], getState());
        return Promise.all([
          dispatch({ type: DeviceConstants.RECEIVE_DEVICES, devicesById: deviceAccu.devicesById }),
          Promise.resolve({ deviceIds: deviceAccu.ids, searchTotal: Number(response.headers[headerNames.total]) })
        ]);
      })
      .catch(err => commonErrorHandler(err, `devices couldn't be searched.`, dispatch, commonErrorFallback));
  };

const ATTRIBUTE_LIST_CUTOFF = 100;
const attributeReducer = (attributes = []) =>
  attributes.slice(0, ATTRIBUTE_LIST_CUTOFF).reduce(
    (accu, { name, scope }) => {
      if (!accu[scope]) {
        accu[scope] = [];
      }
      accu[scope].push(name);
      return accu;
    },
    { identity: [], inventory: [], system: [], tags: [] }
  );

export const getDeviceAttributes = () => (dispatch, getState) =>
  GeneralApi.get(getAttrsEndpoint(getState().app.features.hasReporting)).then(({ data }) => {
    // TODO: remove the array fallback once the inventory attributes endpoint is fixed
    const { identity: identityAttributes, inventory: inventoryAttributes, system: systemAttributes, tags: tagAttributes } = attributeReducer(data || []);
    return dispatch({
      type: DeviceConstants.SET_FILTER_ATTRIBUTES,
      attributes: { identityAttributes, inventoryAttributes, systemAttributes, tagAttributes }
    });
  });

export const getReportingLimits = () => dispatch =>
  GeneralApi.get(`${reportingApiUrl}/devices/attributes`)
    .catch(err => commonErrorHandler(err, `filterable attributes limit & usage could not be retrieved.`, dispatch, commonErrorFallback))
    .then(({ data }) => {
      const { attributes, count, limit } = data;
      const groupedAttributes = attributeReducer(attributes);
      return Promise.resolve(dispatch({ type: DeviceConstants.SET_FILTERABLES_CONFIG, count, limit, attributes: groupedAttributes }));
    });

export const ensureVersionString = (software, fallback) =>
  software.length && software !== 'artifact_name' ? (software.endsWith('.version') ? software : `${software}.version`) : fallback;

const getSingleReportData = (reportConfig, groups) => {
  const { attribute, group, software = '' } = reportConfig;
  const filters = [{ key: 'status', scope: 'identity', operator: DEVICE_FILTERING_OPTIONS.$eq.key, value: 'accepted' }];
  if (group) {
    const staticGroupFilter = { key: 'group', scope: 'system', operator: DEVICE_FILTERING_OPTIONS.$eq.key, value: group };
    const { cleanedFilters: groupFilters } = getGroupFilters(group, groups);
    filters.push(...(groupFilters.length ? groupFilters : [staticGroupFilter]));
  }
  const aggregationAttribute = ensureVersionString(software, attribute);
  return GeneralApi.post(`${reportingApiUrl}/devices/aggregate`, {
    aggregations: [{ attribute: aggregationAttribute, name: '*', scope: 'inventory', size: chartColorPalette.length }],
    filters: mapFiltersToTerms(filters)
  }).then(({ data }) => ({ data, reportConfig }));
};

export const defaultReportType = 'distribution';
export const defaultReports = [{ ...emptyChartSelection, group: null, attribute: 'artifact_name', type: defaultReportType }];

export const getReportsData = () => (dispatch, getState) => {
  const state = getState();
  const reports =
    getUserSettings(state).reports ||
    state.users.globalSettings[`${state.users.currentUser}-reports`] ||
    (Object.keys(state.devices.byId).length ? defaultReports : []);
  return Promise.all(reports.map(report => getSingleReportData(report, getState().devices.groups))).then(results => {
    const devicesState = getState().devices;
    const totalDeviceCount = devicesState.byStatus.accepted.total;
    const newReports = results.map(({ data, reportConfig }) => {
      let { items, other_count } = data[0];
      const { attribute, group, software = '' } = reportConfig;
      const dataCount = items.reduce((accu, item) => accu + item.count, 0);
      // the following is needed to show reports including both old (artifact_name) & current style (rootfs-image.version) device software
      const otherCount = !group && (software === rootfsImageVersion || attribute === 'artifact_name') ? totalDeviceCount - dataCount : other_count;
      return { items, otherCount, total: otherCount + dataCount };
    });
    return Promise.resolve(dispatch({ type: DeviceConstants.SET_DEVICE_REPORTS, reports: newReports }));
  });
};

const initializeDistributionData = (report, groups, devices, totalDeviceCount) => {
  const { attribute, group = '', software = '' } = report;
  const effectiveAttribute = software ? software : attribute;
  const { deviceIds, total = 0 } = groups[group] || {};
  const relevantDevices = groups[group] ? deviceIds.map(id => devices[id]) : Object.values(devices);
  const distributionByAttribute = relevantDevices.reduce((accu, item) => {
    if (!item.attributes || item.status !== DEVICE_STATES.accepted) return accu;
    if (!accu[item.attributes[effectiveAttribute]]) {
      accu[item.attributes[effectiveAttribute]] = 0;
    }
    accu[item.attributes[effectiveAttribute]] = accu[item.attributes[effectiveAttribute]] + 1;
    return accu;
  }, {});
  const distributionByAttributeSorted = Object.entries(distributionByAttribute).sort((pairA, pairB) => pairB[1] - pairA[1]);
  const items = distributionByAttributeSorted.map(([key, count]) => ({ key, count }));
  const dataCount = items.reduce((accu, item) => accu + item.count, 0);
  // the following is needed to show reports including both old (artifact_name) & current style (rootfs-image.version) device software
  const otherCount = (groups[group] ? total : totalDeviceCount) - dataCount;
  return { items, otherCount, total: otherCount + dataCount };
};

const deriveReportsData = () => (dispatch, getState) => {
  const state = getState();
  const {
    groups: { byId: groupsById },
    byId,
    byStatus: {
      accepted: { total }
    }
  } = state.devices;
  const reports =
    getUserSettings(state).reports || state.users.globalSettings[`${state.users.currentUser}-reports`] || (Object.keys(byId).length ? defaultReports : []);
  const newReports = reports.map(report => initializeDistributionData(report, groupsById, byId, total));
  return Promise.resolve(dispatch({ type: DeviceConstants.SET_DEVICE_REPORTS, reports: newReports }));
};

export const getReportsDataWithoutBackendSupport = () => (dispatch, getState) =>
  Promise.all([dispatch(getAllDevicesByStatus(DEVICE_STATES.accepted)), dispatch(getGroups()), dispatch(getDynamicGroups())]).then(() => {
    const { dynamic: dynamicGroups, static: staticGroups } = getGroupsSelector(getState());
    return Promise.all([
      ...staticGroups.map(group => dispatch(getAllGroupDevices(group))),
      ...dynamicGroups.map(group => dispatch(getAllDynamicGroupDevices(group)))
    ]).then(() => dispatch(deriveReportsData()));
  });

export const getDeviceConnect = id => dispatch =>
  GeneralApi.get(`${deviceConnect}/devices/${id}`).then(({ data }) => {
    let tasks = [
      dispatch({
        type: DeviceConstants.RECEIVE_DEVICE_CONNECT,
        device: { connect_status: data.status, connect_updated_ts: data.updated_ts, id }
      })
    ];
    tasks.push(Promise.resolve(data));
    return Promise.all(tasks);
  });

export const getSessionDetails = (sessionId, deviceId, userId, startDate, endDate) => () => {
  const createdAfter = startDate ? `&created_after=${Math.round(Date.parse(startDate) / 1000)}` : '';
  const createdBefore = endDate ? `&created_before=${Math.round(Date.parse(endDate) / 1000)}` : '';
  const objectSearch = `&object_id=${deviceId}`;
  return GeneralApi.get(`${auditLogsApiUrl}/logs?per_page=500${createdAfter}${createdBefore}&actor_id=${userId}${objectSearch}`).then(
    ({ data: auditLogEntries }) => {
      const { start, end } = auditLogEntries.reduce(
        (accu, item) => {
          if (item.meta?.session_id?.includes(sessionId)) {
            accu.start = new Date(item.action.startsWith('open') ? item.time : accu.start);
            accu.end = new Date(item.action.startsWith('close') ? item.time : accu.end);
          }
          return accu;
        },
        { start: startDate || endDate, end: endDate || startDate }
      );
      return Promise.resolve({ start, end });
    }
  );
};

export const getDeviceFileDownloadLink = (deviceId, path) => () =>
  Promise.resolve(`${deviceConnect}/devices/${deviceId}/download?path=${encodeURIComponent(path)}`);

export const deviceFileUpload = (deviceId, path, file) => (dispatch, getState) => {
  var formData = new FormData();
  formData.append('path', path);
  formData.append('file', file);
  const uploadId = uuid();
  const cancelSource = new AbortController();
  const uploads = { ...getState().app.uploads, [uploadId]: { inprogress: true, uploadProgress: 0, cancelSource } };
  return Promise.all([
    dispatch(setSnackbar('Uploading file')),
    dispatch({ type: UPLOAD_PROGRESS, uploads }),
    GeneralApi.uploadPut(`${deviceConnect}/devices/${deviceId}/upload`, formData, e => dispatch(progress(e, uploadId)), cancelSource.signal)
  ])
    .then(() => Promise.resolve(dispatch(setSnackbar('Upload successful', TIMEOUTS.fiveSeconds))))
    .catch(err => {
      if (isCancel(err)) {
        return dispatch(setSnackbar('The upload has been cancelled', TIMEOUTS.fiveSeconds));
      }
      return commonErrorHandler(err, `Error uploading file to device.`, dispatch);
    })
    .finally(() => dispatch(cleanUpUpload(uploadId)));
};

export const getDeviceAuth = id => dispatch =>
  Promise.resolve(dispatch(getDevicesWithAuth([{ id }]))).then(results => {
    if (results[results.length - 1]) {
      return Promise.resolve(results[results.length - 1][0]);
    }
    return Promise.resolve();
  });

export const getDevicesWithAuth = devices => (dispatch, getState) =>
  devices.length
    ? GeneralApi.get(`${deviceAuthV2}/devices?id=${devices.map(device => device.id).join('&id=')}`)
        .then(({ data: receivedDevices }) => {
          const { devicesById } = reduceReceivedDevices(receivedDevices, [], getState());
          return Promise.all([dispatch({ type: DeviceConstants.RECEIVE_DEVICES, devicesById }), Promise.resolve(receivedDevices)]);
        })
        .catch(err => commonErrorHandler(err, `Error: ${err}`, dispatch))
    : Promise.resolve([[], []]);

const maybeUpdateDevicesByStatus = (deviceId, authId) => (dispatch, getState) => {
  const devicesState = getState().devices;
  const device = devicesState.byId[deviceId];
  const hasMultipleAuthSets = authId ? device.auth_sets.filter(authset => authset.id !== authId).length > 0 : false;
  if (!hasMultipleAuthSets && Object.values(DEVICE_STATES).includes(device.status)) {
    const deviceIds = devicesState.byStatus[device.status].deviceIds.filter(id => id !== deviceId);
    return Promise.resolve(
      dispatch({
        type: DeviceConstants[`SET_${device.status.toUpperCase()}_DEVICES`],
        deviceIds,
        forceUpdate: true,
        status: device.status,
        total: Math.max(0, devicesState.byStatus[device.status].total - 1)
      })
    );
  }
  return Promise.resolve();
};

export const updateDeviceAuth = (deviceId, authId, status) => (dispatch, getState) =>
  GeneralApi.put(`${deviceAuthV2}/devices/${deviceId}/auth/${authId}/status`, { status })
    .then(() => Promise.all([dispatch(getDeviceAuth(deviceId)), dispatch(setSnackbar('Device authorization status was updated successfully'))]))
    .catch(err => commonErrorHandler(err, 'There was a problem updating the device authorization status:', dispatch))
    .then(() => Promise.resolve(dispatch(maybeUpdateDevicesByStatus(deviceId, authId))))
    .finally(() => dispatch(setDeviceListState({ refreshTrigger: !getState().devices.deviceList.refreshTrigger })));

export const updateDevicesAuth = (deviceIds, status) => (dispatch, getState) => {
  let devices = getState().devices.byId;
  const deviceIdsWithoutAuth = deviceIds.reduce((accu, id) => (devices[id].auth_sets ? accu : [...accu, { id }]), []);
  return dispatch(getDevicesWithAuth(deviceIdsWithoutAuth)).then(() => {
    devices = getState().devices.byId;
    // for each device, get id and id of authset & make api call to accept
    // if >1 authset, skip instead
    const deviceAuthUpdates = deviceIds.map(id => {
      const device = devices[id];
      if (device.auth_sets.length !== 1) {
        return Promise.reject();
      }
      // api call device.id and device.authsets[0].id
      return dispatch(updateDeviceAuth(device.id, device.auth_sets[0].id, status)).catch(err =>
        commonErrorHandler(err, 'The action was stopped as there was a problem updating a device authorization status: ', dispatch)
      );
    });
    return Promise.allSettled(deviceAuthUpdates).then(results => {
      const { skipped, count } = results.reduce(
        (accu, item) => {
          if (item.status === 'rejected') {
            accu.skipped = accu.skipped + 1;
          } else {
            accu.count = accu.count + 1;
          }
          return accu;
        },
        { skipped: 0, count: 0 }
      );
      const message = getSnackbarMessage(skipped, count);
      // break if an error occurs, display status up til this point before error message
      return dispatch(setSnackbar(message));
    });
  });
};

export const deleteAuthset = (deviceId, authId) => (dispatch, getState) =>
  GeneralApi.delete(`${deviceAuthV2}/devices/${deviceId}/auth/${authId}`)
    .then(() => Promise.all([dispatch(setSnackbar('Device authorization status was updated successfully'))]))
    .catch(err => commonErrorHandler(err, 'There was a problem updating the device authorization status:', dispatch))
    .then(() => Promise.resolve(dispatch(maybeUpdateDevicesByStatus(deviceId, authId))))
    .finally(() => dispatch(setDeviceListState({ refreshTrigger: !getState().devices.deviceList.refreshTrigger })));

export const preauthDevice = authset => dispatch =>
  GeneralApi.post(`${deviceAuthV2}/devices`, authset)
    .catch(err => {
      if (err.response.status === 409) {
        return Promise.reject('A device with a matching identity data set already exists');
      }
      commonErrorHandler(err, 'The device could not be added:', dispatch);
      return Promise.reject();
    })
    .then(() => Promise.resolve(dispatch(setSnackbar('Device was successfully added to the preauthorization list', TIMEOUTS.fiveSeconds))));

export const decommissionDevice = (deviceId, authId) => (dispatch, getState) =>
  GeneralApi.delete(`${deviceAuthV2}/devices/${deviceId}`)
    .then(() => Promise.resolve(dispatch(setSnackbar('Device was decommissioned successfully'))))
    .catch(err => commonErrorHandler(err, 'There was a problem decommissioning the device:', dispatch))
    .then(() => Promise.resolve(dispatch(maybeUpdateDevicesByStatus(deviceId, authId))))
    // trigger reset of device list list!
    .finally(() => dispatch(setDeviceListState({ refreshTrigger: !getState().devices.deviceList.refreshTrigger })));

export const getDeviceConfig = deviceId => dispatch =>
  GeneralApi.get(`${deviceConfig}/${deviceId}`)
    .then(({ data }) => {
      let tasks = [
        dispatch({
          type: DeviceConstants.RECEIVE_DEVICE_CONFIG,
          device: { id: deviceId, config: data }
        })
      ];
      tasks.push(Promise.resolve(data));
      return Promise.all(tasks);
    })
    .catch(err => {
      // if we get a proper error response we most likely queried a device without an existing config check-in and we can just ignore the call
      if (err.response?.data?.error.status_code !== 404) {
        return commonErrorHandler(err, `There was an error retrieving the configuration for device ${deviceId}.`, dispatch, commonErrorFallback);
      }
    });

export const setDeviceConfig = (deviceId, config) => dispatch =>
  GeneralApi.put(`${deviceConfig}/${deviceId}`, config)
    .catch(err => commonErrorHandler(err, `There was an error setting the configuration for device ${deviceId}.`, dispatch, commonErrorFallback))
    .then(() => Promise.resolve(dispatch(getDeviceConfig(deviceId))));

export const applyDeviceConfig = (deviceId, configDeploymentConfiguration, isDefault, config) => (dispatch, getState) =>
  GeneralApi.post(`${deviceConfig}/${deviceId}/deploy`, configDeploymentConfiguration)
    .catch(err => commonErrorHandler(err, `There was an error deploying the configuration to device ${deviceId}.`, dispatch, commonErrorFallback))
    .then(({ data }) => {
      const device = getDeviceByIdSelector(getState(), deviceId);
      const { canManageUsers } = getUserCapabilities(getState());
      let tasks = [
        dispatch({ type: DeviceConstants.RECEIVE_DEVICE, device: { ...device, config: { ...device.config, deployment_id: '' } } }),
        new Promise(resolve => setTimeout(() => resolve(dispatch(getSingleDeployment(data.deployment_id))), TIMEOUTS.oneSecond))
      ];
      if (isDefault && canManageUsers) {
        const { previous } = getState().users.globalSettings.defaultDeviceConfig ?? {};
        tasks.push(dispatch(saveGlobalSettings({ defaultDeviceConfig: { current: config, previous } })));
      }
      return Promise.all(tasks);
    });

export const setDeviceTags = (deviceId, tags) => dispatch =>
  // to prevent tag set failures, retrieve the device & use the freshest etag we can get
  Promise.resolve(dispatch(getDeviceById(deviceId))).then(device => {
    const headers = device.etag ? { 'If-Match': device.etag } : {};
    return GeneralApi.put(
      `${inventoryApiUrl}/devices/${deviceId}/tags`,
      Object.entries(tags).map(([name, value]) => ({ name, value })),
      { headers }
    )
      .catch(err => commonErrorHandler(err, `There was an error setting tags for device ${deviceId}.`, dispatch, 'Please check your connection.'))
      .then(() => Promise.all([dispatch({ type: DeviceConstants.RECEIVE_DEVICE, device: { ...device, tags } }), dispatch(setSnackbar('Device name changed'))]));
  });

export const getDeviceTwin = (deviceId, integration) => (dispatch, getState) => {
  let providerResult = {};
  return GeneralApi.get(`${iotManagerBaseURL}/devices/${deviceId}/state`)
    .then(({ data }) => {
      providerResult = { ...data, twinError: '' };
    })
    .catch(err => {
      providerResult = {
        twinError: `There was an error getting the ${DeviceConstants.EXTERNAL_PROVIDER[
          integration.provider
        ].twinTitle.toLowerCase()} for device ${deviceId}. ${err}`
      };
    })
    .finally(() =>
      Promise.resolve(
        dispatch({
          type: DeviceConstants.RECEIVE_DEVICE,
          device: {
            ...getState().devices.byId[deviceId],
            twinsByIntegration: {
              ...getState().devices.byId[deviceId].twinsByIntegration,
              ...providerResult
            }
          }
        })
      )
    );
};

export const setDeviceTwin = (deviceId, integration, settings) => (dispatch, getState) =>
  GeneralApi.put(`${iotManagerBaseURL}/devices/${deviceId}/state/${integration.id}`, { desired: settings })
    .catch(err =>
      commonErrorHandler(
        err,
        `There was an error updating the ${DeviceConstants.EXTERNAL_PROVIDER[integration.provider].twinTitle.toLowerCase()} for device ${deviceId}.`,
        dispatch
      )
    )
    .then(() => {
      const { twinsByIntegration = {} } = getState().devices.byId[deviceId];
      const { [integration.id]: currentState = {} } = twinsByIntegration;
      return Promise.resolve(
        dispatch({
          type: DeviceConstants.RECEIVE_DEVICE,
          device: {
            ...getState().devices.byId[deviceId],
            twinsByIntegration: {
              ...twinsByIntegration,
              [integration.id]: {
                ...currentState,
                desired: settings
              }
            }
          }
        })
      );
    });

const prepareSearchArguments = ({ filters, group, state, status }) => {
  const { filterTerms } = convertDeviceListStateToFilters({ filters, group, offlineThreshold: state.app.offlineThreshold, selectedIssues: [], status });
  const { columnSelection = [] } = getUserSettings(state);
  const selectedAttributes = columnSelection.map(column => ({ attribute: column.key, scope: column.scope }));
  const attributes = [...defaultAttributes, getIdAttribute(state), ...selectedAttributes];
  return { attributes, filterTerms };
};

export const getSystemDevices =
  (id, options = {}) =>
  (dispatch, getState) => {
    const { page = defaultPage, perPage = defaultPerPage, sortOptions = [] } = options;
    const state = getState();
    let device = getDeviceByIdSelector(state, id);
    const { attributes: deviceAttributes = {} } = device;
    const { mender_gateway_system_id = '' } = deviceAttributes;
    const { hasFullFiltering } = getTenantCapabilities(state);
    if (!hasFullFiltering) {
      return Promise.resolve();
    }
    const filters = [
      { ...emptyFilter, key: 'mender_is_gateway', operator: DEVICE_FILTERING_OPTIONS.$ne.key, value: 'true', scope: 'inventory' },
      { ...emptyFilter, key: 'mender_gateway_system_id', value: mender_gateway_system_id, scope: 'inventory' }
    ];
    const { attributes, filterTerms } = prepareSearchArguments({ filters, state });

    return GeneralApi.post(getSearchEndpoint(state.app.features.hasReporting), {
      page,
      per_page: perPage,
      filters: filterTerms,
      sort: sortOptions,
      attributes
    })
      .catch(err => commonErrorHandler(err, `There was an error getting system devices device ${id}.`, dispatch, 'Please check your connection.'))
      .then(({ data, headers }) => {
        const state = getState();
        const { devicesById, ids } = reduceReceivedDevices(data, [], state);
        const device = {
          ...state.devices.byId[id],
          systemDeviceIds: ids,
          systemDeviceTotal: Number(headers[headerNames.total])
        };
        return Promise.resolve(
          dispatch({
            type: DeviceConstants.RECEIVE_DEVICES,
            devicesById: {
              ...devicesById,
              [id]: device
            }
          })
        );
      });
  };

export const getGatewayDevices = deviceId => (dispatch, getState) => {
  const state = getState();
  let device = getDeviceByIdSelector(state, deviceId);
  const { attributes = {} } = device;
  const { mender_gateway_system_id = '' } = attributes;
  const filters = [
    { ...emptyFilter, key: 'id', operator: DEVICE_FILTERING_OPTIONS.$ne.key, value: deviceId, scope: 'identity' },
    { ...emptyFilter, key: 'mender_is_gateway', value: 'true', scope: 'inventory' },
    { ...emptyFilter, key: 'mender_gateway_system_id', value: mender_gateway_system_id, scope: 'inventory' }
  ];
  const { attributes: attributeSelection, filterTerms } = prepareSearchArguments({ filters, state });
  return GeneralApi.post(getSearchEndpoint(state.app.features.hasReporting), {
    page: 1,
    per_page: MAX_PAGE_SIZE,
    filters: filterTerms,
    attributes: attributeSelection
  }).then(({ data }) => {
    const { ids } = reduceReceivedDevices(data, [], getState());
    let tasks = ids.map(deviceId => dispatch(getDeviceInfo(deviceId)));
    tasks.push(dispatch({ type: DeviceConstants.RECEIVE_DEVICE, device: { ...getState().devices.byId[deviceId], gatewayIds: ids } }));
    return Promise.all(tasks);
  });
};

export const geoAttributes = ['geo-lat', 'geo-lon'].map(attribute => ({ attribute, scope: 'inventory' }));
export const getDevicesInBounds = (bounds, group) => (dispatch, getState) => {
  const state = getState();
  const { filterTerms } = convertDeviceListStateToFilters({
    group: group === DeviceConstants.ALL_DEVICES ? undefined : group,
    groups: state.devices.groups,
    status: DEVICE_STATES.accepted
  });
  return GeneralApi.post(getSearchEndpoint(state.app.features.hasReporting), {
    page: 1,
    per_page: MAX_PAGE_SIZE,
    filters: filterTerms,
    attributes: geoAttributes,
    geo_bounding_box_filter: {
      geo_bounding_box: {
        location: {
          top_left: { lat: bounds._northEast.lat, lon: bounds._southWest.lng },
          bottom_right: { lat: bounds._southWest.lat, lon: bounds._northEast.lng }
        }
      }
    }
  }).then(({ data }) => {
    const { devicesById } = reduceReceivedDevices(data, [], getState());
    return Promise.resolve(dispatch({ type: DeviceConstants.RECEIVE_DEVICES, devicesById }));
  });
};
