// Copyright 2022 Northern.tech AS
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
import { routes } from '../components/devices/base-devices';
import { DEPLOYMENT_ROUTES, DEPLOYMENT_STATES, DEPLOYMENT_TYPES } from '../constants/deploymentConstants';
import { ALL_DEVICES, ATTRIBUTE_SCOPES, DEVICE_FILTERING_OPTIONS, DEVICE_LIST_DEFAULTS, UNGROUPED_GROUP, emptyFilter } from '../constants/deviceConstants';
import { AUDIT_LOGS_TYPES } from '../constants/organizationConstants';
import { deepCompare, getISOStringBoundaries } from '../helpers';

const SEPARATOR = ':';

const defaultSelector = result => result[0];

const commonFields = {
  ...Object.keys(DEVICE_LIST_DEFAULTS).reduce((accu, key) => ({ ...accu, [key]: { parse: Number, select: defaultSelector, target: key } }), {}),
  id: { parse: String, select: i => i, target: 'id' },
  issues: { parse: undefined, select: defaultSelector, target: 'selectedIssues' },
  open: { parse: Boolean, select: defaultSelector, target: 'open' }
};

const scopes = {
  identity: { delimiter: 'identity', filters: [] },
  inventory: { delimiter: 'inventory', filters: [] },
  monitor: { delimiter: 'monitor', filters: [] },
  system: { delimiter: 'system', filters: [] },
  tags: { delimiter: 'tags', filters: [] }
};

export const commonProcessor = searchParams => {
  let params = new URLSearchParams(searchParams);
  const pageState = Object.entries(commonFields).reduce((accu, [key, { parse, select, target }]) => {
    const values = params.getAll(key);
    if (!values.length) {
      return accu;
    }
    if (!parse) {
      accu[target] = values;
    } else {
      try {
        accu[target] = select(values.map(parse));
      } catch (error) {
        console.log('encountered faulty url param, continue...', error);
      }
    }
    return accu;
  }, {});
  Object.keys(commonFields).map(key => params.delete(key));
  const sort = params.has('sort')
    ? params.getAll('sort').reduce((sortAccu, scopedQuery) => {
        const items = scopedQuery.split(SEPARATOR).reverse();
        return ['direction', 'key', 'scope'].reduce((accu, key, index) => {
          if (items[index]) {
            accu[key] = items[index];
          }
          return accu;
        }, sortAccu);
      }, {})
    : undefined;
  params.delete('sort');
  return { pageState, params, sort };
};

const legacyDeviceQueryParse = (searchParams, filteringAttributes) => {
  let params = new URLSearchParams(searchParams);
  const result = Object.keys(scopes).reduce((accu, scope) => ({ ...accu, [scope]: [] }), {});
  if (params.get('group')) {
    result.inventory.push({ ...emptyFilter, key: 'group', scope: 'inventory', operator: DEVICE_FILTERING_OPTIONS.$eq.key, value: params.get('group') });
    params.delete('group');
  }
  const filters = [...params.keys()].reduce(
    (accu, key) =>
      params.getAll(key).reduce((innerAccu, value) => {
        const scope =
          Object.entries(filteringAttributes).reduce((foundScope, [currentScope, attributes]) => {
            if (foundScope) {
              return foundScope;
            }
            return attributes.includes(key) ? currentScope.substring(0, currentScope.indexOf('Attributes')) : foundScope;
          }, undefined) ?? ATTRIBUTE_SCOPES.inventory;
        innerAccu[scope].push({ ...emptyFilter, scope, key, operator: DEVICE_FILTERING_OPTIONS.$eq.key, value });
        return innerAccu;
      }, accu),
    result
  );
  [...params.keys()].map(key => params.delete(key));
  return { filters, params };
};

const scopedFilterParse = searchParams => {
  let params = new URLSearchParams(searchParams);
  const filters = Object.keys(scopes).reduce(
    (accu, scope) => {
      accu[scope] = [];
      if (!params.has(scope)) {
        return accu;
      }
      accu[scope] = params.getAll(scope).map(scopedQuery => {
        const items = scopedQuery.split(SEPARATOR);
        // URLSearchParams will automatically decode any URI encoding present in the query string, thus we have to also handle queries with a SEPARATOR separately
        return { ...emptyFilter, scope, key: items[0], operator: `$${items[1]}`, value: items.slice(2).join(SEPARATOR) };
      });
      return accu;
    },
    { ...scopes }
  );
  Object.keys(scopes).map(scope => params.delete(scope));
  return { filters, params };
};

// filters, selectedGroup
export const parseDeviceQuery = (searchParams, extraProps = {}) => {
  let queryParams = new URLSearchParams(searchParams);
  const { filteringAttributes = {}, pageState = {} } = extraProps;
  const pageStateExtension = pageState.id?.length === 1 ? { open: true } : {};

  let scopedFilters;
  const refersOldStyleAttributes = Object.values(filteringAttributes).some(scopeValues => scopeValues.some(scopedValue => queryParams.get(scopedValue)));
  if ((refersOldStyleAttributes && !Object.keys(scopes).some(scope => queryParams.get(scope))) || queryParams.get('group')) {
    const { filters, params } = legacyDeviceQueryParse(queryParams, filteringAttributes);
    scopedFilters = filters;
    queryParams = params;
  } else {
    const { filters, params } = scopedFilterParse(queryParams);
    scopedFilters = filters;
    queryParams = params;
  }

  let groupName = '';
  const groupFilterIndex = scopedFilters.inventory.findIndex(filter => filter.key === 'group' && filter.operator === DEVICE_FILTERING_OPTIONS.$eq.key);
  if (groupFilterIndex > -1) {
    groupName = scopedFilters.inventory[groupFilterIndex].value;
    scopedFilters.inventory.splice(groupFilterIndex, 1);
  }

  const detailsTab = queryParams.has('tab') ? queryParams.get('tab') : '';
  return { detailsTab, filters: Object.values(scopedFilters).flat(), groupName, ...pageStateExtension };
};

const formatSorting = (sort, { sort: sortDefault }) => {
  if (!sort || deepCompare(sort, sortDefault)) {
    return '';
  }
  const sortQuery = ['scope', 'key', 'direction']
    .reduce((accu, key) => {
      if (!sort[key]) {
        return accu;
      }
      accu.push(sort[key]);
      return accu;
    }, [])
    .join(SEPARATOR);
  return `sort=${sortQuery}`;
};

export const formatPageState = ({ selectedId, selectedIssues, page, perPage, sort }, { defaults }) =>
  Object.entries({ page, perPage, id: selectedId, issues: selectedIssues, open: selectedId ? true : undefined })
    .reduce(
      (accu, [key, value]) => {
        if (Array.isArray(value)) {
          accu.push(...value.map(item => `${key}=${encodeURIComponent(item)}`));
        } else if ((DEVICE_LIST_DEFAULTS[key] != value || !DEVICE_LIST_DEFAULTS.hasOwnProperty(key)) && value) {
          accu.push(`${key}=${encodeURIComponent(value)}`);
        }
        return accu;
      },
      [formatSorting(sort, defaults)]
    )
    .filter(i => i)
    .join('&');

const stripFilterOperator = operator => operator.replaceAll('$', '');

const formatFilters = filters => {
  const result = filters
    // group all filters by their scope to get a more organized result
    .reduce(
      (accu, filter) => {
        const { scope = ATTRIBUTE_SCOPES.inventory, operator = '$eq' } = filter;
        accu[scope].add(`${scopes[scope].delimiter}=${filter.key}${SEPARATOR}${stripFilterOperator(operator)}${SEPARATOR}${encodeURIComponent(filter.value)}`);
        return accu;
      },
      Object.keys(scopes).reduce((accu, item) => ({ ...accu, [item]: new Set() }), {})
    );
  // boil it all down to a single line containing all filters
  return Object.values(result)
    .map(filterSet => [...filterSet])
    .flat();
};

export const formatDeviceSearch = ({ pageState, filters, selectedGroup }) => {
  let activeFilters = [...filters];
  if (selectedGroup && selectedGroup !== ALL_DEVICES) {
    const isUngroupedGroup = selectedGroup === UNGROUPED_GROUP.id;
    activeFilters = isUngroupedGroup
      ? activeFilters.filter(
          filter => !(filter.key === 'group' && filter.scope === ATTRIBUTE_SCOPES.system && filter.operator === DEVICE_FILTERING_OPTIONS.$nin.key)
        )
      : activeFilters;
    const groupName = isUngroupedGroup ? UNGROUPED_GROUP.name : selectedGroup;
    activeFilters.push({ scope: ATTRIBUTE_SCOPES.inventory, key: 'group', operator: DEVICE_FILTERING_OPTIONS.$eq.key, value: groupName });
  }
  const formattedFilters = formatFilters(activeFilters).filter(i => i);
  if (pageState.detailsTab && pageState.selectedId) {
    formattedFilters.push(`tab=${pageState.detailsTab}`);
  }
  return formattedFilters.join('&');
};

export const generateDevicePath = ({ pageState }) => {
  const { state: selectedState } = pageState;
  const path = ['/devices'];
  if (![routes.allDevices.key, ''].includes(selectedState)) {
    path.push(selectedState);
  }
  return path.join('/');
};

const formatDates = ({ endDate, params, startDate, today, tonight }) => {
  if (endDate && endDate !== tonight) {
    params.set('endDate', endDate.split('T')[0]);
  }
  if (startDate && startDate !== today) {
    params.set('startDate', startDate.split('T')[0]);
  }
  return params;
};

const paramReducer = (accu, [key, value]) => {
  if (value) {
    accu.set(key, value);
  }
  return accu;
};

export const formatAuditlogs = ({ pageState }, { today, tonight }) => {
  const { detail, endDate, startDate, type, user } = pageState;
  let params = new URLSearchParams();
  let userId = user;
  if (user?.id) {
    userId = user.id;
  }
  params = Object.entries({ objectId: detail, userId }).reduce(paramReducer, params);
  if (type) {
    params.set('objectType', type.value ?? type);
  }
  params = formatDates({ endDate, params, startDate, today, tonight });
  return params.toString();
};

const parseDateParams = (params, today, tonight) => {
  let endDate = tonight;
  if (params.get('endDate')) {
    endDate = getISOStringBoundaries(new Date(params.get('endDate'))).end;
  }
  let startDate = today;
  if (params.get('startDate')) {
    startDate = getISOStringBoundaries(new Date(params.get('startDate'))).start;
  }
  return { endDate, startDate };
};

export const parseAuditlogsQuery = (params, { today, tonight }) => {
  const type = AUDIT_LOGS_TYPES.find(typeObject => typeObject.value === params.get('objectType'));
  const { endDate, startDate } = parseDateParams(params, today, tonight);
  return {
    detail: params.get('objectId'),
    endDate,
    startDate,
    type,
    user: params.get('userId')
  };
};

const formatActiveDeployments = (pageState, { defaults }) =>
  [DEPLOYMENT_STATES.inprogress, DEPLOYMENT_STATES.pending]
    .reduce((accu, state) => {
      const { page, perPage } = pageState[state] ?? {};
      const stateDefaults = defaults[state] ?? {};
      const items = Object.entries({ page, perPage })
        .reverse()
        .reduce((keyAccu, [key, value]) => {
          if ((value && value !== stateDefaults[key]) || keyAccu.length) {
            keyAccu.unshift(value || stateDefaults[key]);
          }
          return keyAccu;
        }, []);
      if (items.length) {
        accu.push(`${state}=${items.join(SEPARATOR)}`);
      }
      return accu;
    }, [])
    .filter(i => i)
    .join('&');

export const formatDeployments = ({ deploymentObject, pageState }, { defaults, today, tonight }) => {
  const { state: selectedState, showCreationDialog } = pageState.general;
  let params = new URLSearchParams();
  if (showCreationDialog) {
    params.set('open', true);
    if (deploymentObject.release) {
      params.set('release', deploymentObject.release.name);
    }
    if (deploymentObject.devices?.length) {
      deploymentObject.devices.map(({ id }) => params.append('deviceId', id));
    }
  }
  let pageStateQuery;
  if (selectedState === DEPLOYMENT_ROUTES.finished.key) {
    const { endDate, search, startDate, type } = pageState[selectedState];
    params = formatDates({ endDate, params, startDate, today, tonight });
    params = Object.entries({ search, type }).reduce(paramReducer, params);
    pageStateQuery = formatPageState(pageState[selectedState], { defaults });
  } else if (selectedState === DEPLOYMENT_ROUTES.scheduled.key) {
    pageStateQuery = formatPageState(pageState[selectedState], { defaults });
  } else {
    pageStateQuery = formatActiveDeployments(pageState, { defaults });
  }
  return [pageStateQuery, params.toString()].filter(i => i).join('&');
};

const deploymentsPath = 'deployments/';
const parseDeploymentsPath = path => {
  const parts = path.split(deploymentsPath);
  if (parts.length > 1 && Object.keys(DEPLOYMENT_ROUTES).includes(parts[1])) {
    return parts[1];
  }
  return '';
};

const parseActiveDeployments = params =>
  [DEPLOYMENT_STATES.inprogress, DEPLOYMENT_STATES.pending].reduce((accu, state) => {
    if (!params.has(state)) {
      return accu;
    }
    const items = params.get(state).split(SEPARATOR);
    accu[state] = ['page', 'perPage'].reduce((stateAccu, key, index) => (items[index] ? { ...stateAccu, [key]: Number(items[index]) } : stateAccu), {});
    return accu;
  }, {});

const deploymentFields = {
  deviceId: { attribute: 'devices', parse: id => ({ id }), select: i => i },
  release: { attribute: 'release', parse: String, select: defaultSelector }
};

export const parseDeploymentsQuery = (params, { pageState, location, tonight }) => {
  // for deployments we get a startDate implicitly from a list of retrieved deployments if none is set, thus we're not defaulting to today
  const { endDate, startDate } = parseDateParams(params, '', tonight);
  const deploymentObject = Object.entries(deploymentFields).reduce(
    (accu, [key, { attribute, parse, select }]) => (params.has(key) ? { ...accu, [attribute]: select(params.getAll(key).map(parse)) } : accu),
    {}
  );
  const { state: selectedState, id, open, ...remainingPageState } = pageState;
  const tab = parseDeploymentsPath(location.pathname);
  const deploymentsTab = tab || selectedState || DEPLOYMENT_ROUTES.active.key;

  let state = {
    deploymentObject,
    general: {
      showCreationDialog: Boolean(open && !id),
      showReportDialog: Boolean(open && id),
      state: deploymentsTab
    }
  };
  if (deploymentsTab === DEPLOYMENT_ROUTES.finished.key) {
    const type = DEPLOYMENT_TYPES[params.get('type')] || '';
    const search = params.get('search') || '';
    state[deploymentsTab] = { ...remainingPageState, endDate, search, startDate, type };
  } else if (deploymentsTab === DEPLOYMENT_ROUTES.scheduled.key) {
    state[deploymentsTab] = { ...remainingPageState };
  } else {
    state = {
      ...state,
      ...parseActiveDeployments(params)
    };
  }
  return state;
};

export const generateDeploymentsPath = ({ pageState }) => {
  const { state: selectedState = DEPLOYMENT_ROUTES.active.key } = pageState.general;
  return `/deployments/${selectedState}`;
};

const releasesRoot = '/releases';
export const formatReleases = ({ pageState: { searchTerm, selectedTags = [], tab, type } }) =>
  Object.entries({ name: searchTerm, tab, type })
    .reduce(
      (accu, [key, value]) => (value ? [...accu, `${key}=${value}`] : accu),
      selectedTags.map(tag => `tag=${tag}`)
    )
    .join('&');

export const generateReleasesPath = ({ pageState: { selectedRelease } }) =>
  `${releasesRoot}${selectedRelease ? `/${encodeURIComponent(selectedRelease)}` : ''}`;

export const parseReleasesQuery = (queryParams, extraProps) => {
  const name = queryParams.has('name') ? queryParams.get('name') : '';
  const tab = queryParams.has('tab') ? queryParams.get('tab') : undefined;
  const tags = queryParams.has('tag') ? queryParams.getAll('tag') : [];
  const type = queryParams.has('type') ? queryParams.get('type') : '';
  let selectedRelease = decodeURIComponent(extraProps.location.pathname.substring(releasesRoot.length + 1));
  if (!selectedRelease && extraProps.pageState.id?.length) {
    selectedRelease = extraProps.pageState.id[0];
  }
  return { searchTerm: name, selectedRelease, tab, tags, type };
};
