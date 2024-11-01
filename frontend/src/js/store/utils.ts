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
// @ts-nocheck
import { duplicateFilter, yes } from '@northern.tech/utils/helpers';

import { ATTRIBUTE_SCOPES, DEVICE_FILTERING_OPTIONS, DEVICE_ISSUE_OPTIONS, DEVICE_LIST_MAXIMUM_LENGTH, emptyUiPermissions } from './commonConstants';
import {
  DARK_MODE,
  DEPLOYMENT_STATES,
  defaultStats,
  deploymentDisplayStates,
  deploymentStatesToSubstates,
  deploymentStatesToSubstatesWithSkipped,
  emptyFilter
} from './constants';

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
export const mapFiltersToTerms = filters =>
  filters.map(filter => ({
    scope: filter.scope,
    attribute: filter.key,
    type: filterAliases[filter.operator]?.alias || filter.operator,
    value: filterProcessors.hasOwnProperty(filter.operator) ? filterProcessors[filter.operator](filter.value) : filter.value
  }));
export const mapTermsToFilters = terms =>
  terms.map(term => {
    const aliasedFilter = Object.entries(filterAliases).find(
      aliasDefinition => aliasDefinition[1].alias === term.type && aliasDefinition[1].value === term.value
    );
    const operator = aliasedFilter ? aliasedFilter[0] : term.type;
    return { scope: term.scope, key: term.attribute, operator, value: term.value };
  });

const convertIssueOptionsToFilters = (issuesSelection, filtersState = {}) =>
  issuesSelection.map(item => {
    if (typeof DEVICE_ISSUE_OPTIONS[item].filterRule.value === 'function') {
      return { ...DEVICE_ISSUE_OPTIONS[item].filterRule, value: DEVICE_ISSUE_OPTIONS[item].filterRule.value(filtersState) };
    }
    return DEVICE_ISSUE_OPTIONS[item].filterRule;
  });

export const convertDeviceListStateToFilters = ({ filters = [], group, groups = { byId: {} }, offlineThreshold, selectedIssues = [], status }) => {
  let applicableFilters = [...filters];
  if (typeof group === 'string' && !(groups.byId[group]?.filters || applicableFilters).length) {
    applicableFilters.push({ key: 'group', value: group, operator: DEVICE_FILTERING_OPTIONS.$eq.key, scope: 'system' });
  }
  const nonMonitorFilters = applicableFilters.filter(
    filter =>
      !Object.values(DEVICE_ISSUE_OPTIONS).some(
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

const filterCompare = (filter, item) => Object.keys(emptyFilter).every(key => item[key].toString() === filter[key].toString());

export const filtersFilter = (item, index, array) => {
  const firstIndex = array.findIndex(filter => filterCompare(filter, item));
  return firstIndex === index;
};

export const listItemMapper = (byId, ids, { defaultObject = {}, cutOffSize = DEVICE_LIST_MAXIMUM_LENGTH }) => {
  return ids.slice(0, cutOffSize).reduce((accu, id) => {
    if (id && byId[id]) {
      accu.push({ ...defaultObject, ...byId[id] });
    }
    return accu;
  }, []);
};

export const mergePermissions = (existingPermissions = { ...emptyUiPermissions }, addedPermissions) =>
  Object.entries(existingPermissions).reduce(
    (accu, [key, value]) => {
      let values;
      if (!accu[key]) {
        accu[key] = value;
        return accu;
      }
      if (Array.isArray(value)) {
        values = [...value, ...accu[key]].filter(duplicateFilter);
      } else {
        values = mergePermissions(accu[key], { ...value });
      }
      accu[key] = values;
      return accu;
    },
    { ...addedPermissions }
  );

export const mapUserRolesToUiPermissions = (userRoles, roles) =>
  userRoles.reduce(
    (accu, roleId) => {
      if (!(roleId && roles[roleId])) {
        return accu;
      }
      return mergePermissions(accu, roles[roleId].uiPermissions);
    },
    { ...emptyUiPermissions }
  );

export const progress = ({ loaded, total }) => {
  let uploadProgress = (loaded / total) * 100;
  return (uploadProgress = uploadProgress < 50 ? Math.ceil(uploadProgress) : Math.round(uploadProgress));
};

export const extractErrorMessage = (err, fallback = '') =>
  err.response?.data?.error?.message || err.response?.data?.error || err.error || err.message || fallback;

export const preformatWithRequestID = (res, failMsg) => {
  // ellipsis line
  if (failMsg.length > 100) failMsg = `${failMsg.substring(0, 220)}...`;

  try {
    if (res?.data && Object.keys(res.data).includes('request_id')) {
      let shortRequestUUID = res.data['request_id'].substring(0, 8);
      return `${failMsg} [Request ID: ${shortRequestUUID}]`;
    }
  } catch (e) {
    console.log('failed to extract request id:', e);
  }
  return failMsg;
};

export const getComparisonCompatibleVersion = version => (isNaN(version.charAt(0)) && version !== 'next' ? 'master' : version);

export const stringToBoolean = content => {
  if (!content) {
    return false;
  }
  const string = content + '';
  switch (string.trim().toLowerCase()) {
    case 'true':
    case 'yes':
    case '1':
      return true;
    case 'false':
    case 'no':
    case '0':
    case null:
      return false;
    default:
      return Boolean(string);
  }
};

export const groupDeploymentDevicesStats = deployment => {
  const deviceStatCollector = (deploymentStates, devices) =>
    Object.values(devices).reduce((accu, device) => (deploymentStates.includes(device.status) ? accu + 1 : accu), 0);

  const inprogress = deviceStatCollector(deploymentStatesToSubstates.inprogress, deployment.devices);
  const pending = deviceStatCollector(deploymentStatesToSubstates.pending, deployment.devices);
  const successes = deviceStatCollector(deploymentStatesToSubstates.successes, deployment.devices);
  const failures = deviceStatCollector(deploymentStatesToSubstates.failures, deployment.devices);
  const paused = deviceStatCollector(deploymentStatesToSubstates.paused, deployment.devices);
  return { inprogress, paused, pending, successes, failures };
};

export const statCollector = (items, statistics) => items.reduce((accu, property) => accu + Number(statistics[property] || 0), 0);
export const groupDeploymentStats = (deployment, withSkipped) => {
  const { statistics = {} } = deployment;
  const { status = {} } = statistics;
  const stats = { ...defaultStats, ...status };
  let groupStates = deploymentStatesToSubstates;
  let result = {};
  if (withSkipped) {
    groupStates = deploymentStatesToSubstatesWithSkipped;
    result.skipped = statCollector(groupStates.skipped, stats);
  }
  result = {
    ...result,
    // don't include 'pending' as inprogress, as all remaining devices will be pending - we don't discriminate based on phase membership
    inprogress: statCollector(groupStates.inprogress, stats),
    pending: (deployment.max_devices ? deployment.max_devices - deployment.device_count : 0) + statCollector(groupStates.pending, stats),
    successes: statCollector(groupStates.successes, stats),
    failures: statCollector(groupStates.failures, stats),
    paused: statCollector(groupStates.paused, stats)
  };
  return result;
};

export const getDeploymentState = deployment => {
  const { status: deploymentStatus = DEPLOYMENT_STATES.pending } = deployment;
  const { inprogress: currentProgressCount, paused } = groupDeploymentStats(deployment);

  let status = deploymentDisplayStates[deploymentStatus];
  if (deploymentStatus === DEPLOYMENT_STATES.pending && currentProgressCount === 0) {
    status = 'queued';
  } else if (paused > 0) {
    status = deploymentDisplayStates.paused;
  }
  return status;
};

export const generateDeploymentGroupDetails = (filter, groupName) =>
  filter && filter.terms?.length
    ? `${groupName} (${filter.terms
        .map(filter => `${filter.attribute || filter.key} ${DEVICE_FILTERING_OPTIONS[filter.type || filter.operator].shortform} ${filter.value}`)
        .join(', ')})`
    : groupName;

export const mapDeviceAttributes = (attributes = []) =>
  attributes.reduce(
    (accu, attribute) => {
      if (!(attribute.value && attribute.name) && attribute.scope === ATTRIBUTE_SCOPES.inventory) {
        return accu;
      }
      accu[attribute.scope || ATTRIBUTE_SCOPES.inventory] = {
        ...accu[attribute.scope || ATTRIBUTE_SCOPES.inventory],
        [attribute.name]: attribute.value
      };
      if (attribute.name === 'device_type' && attribute.scope === ATTRIBUTE_SCOPES.inventory) {
        accu.inventory.device_type = [].concat(attribute.value);
      }
      return accu;
    },
    { inventory: { device_type: [], artifact_name: '' }, identity: {}, monitor: {}, system: {}, tags: {} }
  );

export const isDarkMode = mode => mode === DARK_MODE;
