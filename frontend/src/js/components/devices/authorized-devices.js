// Copyright 2015 Northern.tech AS
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
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';

// material ui
import { Autorenew as AutorenewIcon, Delete as DeleteIcon, FilterList as FilterListIcon, LockOutlined } from '@mui/icons-material';
import { Button, MenuItem, Select } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { setSnackbar } from '../../actions/appActions';
import { deleteAuthset, setDeviceFilters, setDeviceListState, updateDevicesAuth } from '../../actions/deviceActions';
import { getIssueCountsByType } from '../../actions/monitorActions';
import { advanceOnboarding } from '../../actions/onboardingActions';
import { saveUserSettings, updateUserColumnSettings } from '../../actions/userActions';
import { SORTING_OPTIONS, TIMEOUTS } from '../../constants/appConstants';
import { ALL_DEVICES, DEVICE_ISSUE_OPTIONS, DEVICE_STATES, UNGROUPED_GROUP } from '../../constants/deviceConstants';
import { onboardingSteps } from '../../constants/onboardingConstants';
import { duplicateFilter, toggle } from '../../helpers';
import {
  getAvailableIssueOptionsByType,
  getDeviceCountsByStatus,
  getDeviceFilters,
  getFilterAttributes,
  getIdAttribute,
  getLimitMaxed,
  getMappedDevicesList,
  getOnboardingState,
  getSelectedGroupInfo,
  getTenantCapabilities,
  getUserCapabilities,
  getUserSettings
} from '../../selectors';
import { useDebounce } from '../../utils/debouncehook';
import { getOnboardingComponentFor } from '../../utils/onboardingmanager';
import useWindowSize from '../../utils/resizehook';
import { clearAllRetryTimers, setRetryTimer } from '../../utils/retrytimer';
import Loader from '../common/loader';
import { defaultHeaders, defaultTextRender, getDeviceIdentityText, routes as states } from './base-devices';
import DeviceList, { minCellWidth } from './devicelist';
import ColumnCustomizationDialog from './dialogs/custom-columns-dialog';
import ExpandedDevice from './expanded-device';
import DeviceQuickActions from './widgets/devicequickactions';
import Filters from './widgets/filters';
import DeviceIssuesSelection from './widgets/issueselection';
import ListOptions from './widgets/listoptions';

const deviceRefreshTimes = {
  [DEVICE_STATES.accepted]: TIMEOUTS.refreshLong,
  [DEVICE_STATES.pending]: TIMEOUTS.refreshDefault,
  [DEVICE_STATES.preauth]: TIMEOUTS.refreshLong,
  [DEVICE_STATES.rejected]: TIMEOUTS.refreshLong,
  default: TIMEOUTS.refreshDefault
};

const idAttributeTitleMap = {
  id: 'Device ID',
  name: 'Name'
};

const headersReducer = (accu, header) => {
  if (header.attribute.scope === accu.column.scope && (header.attribute.name === accu.column.name || header.attribute.alternative === accu.column.name)) {
    accu.header = { ...accu.header, ...header };
  }
  return accu;
};

const useStyles = makeStyles()(theme => ({
  filterCommon: {
    borderStyle: 'solid',
    borderWidth: 1,
    borderRadius: 5,
    borderColor: theme.palette.grey[100],
    background: theme.palette.background.default,
    [`.filter-list > .MuiChip-root`]: {
      marginBottom: theme.spacing()
    },
    [`.filter-list > .MuiChip-root > .MuiChip-label`]: {
      whiteSpace: 'normal'
    },
    ['&.filter-header']: {
      overflow: 'hidden',
      zIndex: 2
    },
    ['&.filter-toggle']: {
      background: 'transparent',
      borderBottomRightRadius: 0,
      borderBottomLeftRadius: 0,
      borderBottomColor: theme.palette.background.default,
      marginBottom: -1
    },
    ['&.filter-wrapper']: {
      padding: 20,
      borderTopLeftRadius: 0
    }
  },
  selection: {
    fontSize: 13,
    marginLeft: theme.spacing(0.5),
    marginTop: 2,
    '>div': {
      paddingLeft: theme.spacing(0.5)
    }
  }
}));

export const getHeaders = (columnSelection = [], currentStateHeaders, idAttribute, openSettingsDialog) => {
  const headers = columnSelection.length
    ? columnSelection.map(column => {
        let header = { ...column, attribute: { ...column }, textRender: defaultTextRender, sortable: true };
        header = Object.values(defaultHeaders).reduce(headersReducer, { column, header }).header;
        header = currentStateHeaders.reduce(headersReducer, { column, header }).header;
        return header;
      })
    : currentStateHeaders;
  const { attribute, scope } = idAttribute;
  return [
    {
      title: idAttributeTitleMap[attribute] ?? attribute,
      customize: openSettingsDialog,
      attribute: { name: attribute, scope },
      sortable: true,
      textRender: getDeviceIdentityText
    },
    ...headers,
    defaultHeaders.deviceStatus
  ];
};

const calculateColumnSelectionSize = (changedColumns, customColumnSizes) =>
  changedColumns.reduce(
    (accu, column) => {
      const size = customColumnSizes.find(({ attribute }) => attribute.name === column.key && attribute.scope === column.scope)?.size || minCellWidth;
      accu.columnSizes.push({ attribute: { name: column.key, scope: column.scope }, size });
      accu.selectedAttributes.push({ attribute: column.key, scope: column.scope });
      return accu;
    },
    { columnSizes: [], selectedAttributes: [] }
  );

const OnboardingComponent = ({ deviceListRef, onboardingState }) => {
  let onboardingComponent = null;
  const element = deviceListRef.current?.querySelector('body .deviceListItem > div');
  if (element) {
    const anchor = { left: 200, top: element.offsetTop + element.offsetHeight };
    onboardingComponent = getOnboardingComponentFor(onboardingSteps.DEVICES_PENDING_ONBOARDING, onboardingState, { anchor }, onboardingComponent);
  } else if (deviceListRef.current) {
    const anchor = { top: deviceListRef.current.offsetTop + deviceListRef.current.offsetHeight / 3, left: deviceListRef.current.offsetWidth / 2 + 30 };
    onboardingComponent = getOnboardingComponentFor(
      onboardingSteps.DEVICES_PENDING_ONBOARDING_START,
      onboardingState,
      { anchor, place: 'top' },
      onboardingComponent
    );
  }
  return onboardingComponent;
};

export const Authorized = ({
  addDevicesToGroup,
  onGroupClick,
  onGroupRemoval,
  onMakeGatewayClick,
  onPreauthClick,
  openSettingsDialog,
  removeDevicesFromGroup,
  showsDialog
}) => {
  const limitMaxed = useSelector(getLimitMaxed);
  const devices = useSelector(state => getMappedDevicesList(state, 'deviceList'));
  const { selectedGroup, groupFilters = [] } = useSelector(getSelectedGroupInfo);
  const { columnSelection = [] } = useSelector(getUserSettings);
  const attributes = useSelector(getFilterAttributes);
  const { accepted: acceptedCount, pending: pendingCount, rejected: rejectedCount } = useSelector(getDeviceCountsByStatus);
  const allCount = acceptedCount + rejectedCount;
  const availableIssueOptions = useSelector(getAvailableIssueOptionsByType);
  const customColumnSizes = useSelector(state => state.users.customColumns);
  const deviceListState = useSelector(state => state.devices.deviceList);
  const { total: deviceCount } = deviceListState;
  const filters = useSelector(getDeviceFilters);
  const idAttribute = useSelector(getIdAttribute);
  const onboardingState = useSelector(getOnboardingState);
  const settingsInitialized = useSelector(state => state.users.settingsInitialized);
  const tenantCapabilities = useSelector(getTenantCapabilities);
  const userCapabilities = useSelector(getUserCapabilities);
  const dispatch = useDispatch();
  const dispatchedSetSnackbar = useCallback((...args) => dispatch(setSnackbar(...args)), [dispatch]);

  const {
    refreshTrigger,
    selectedId,
    selectedIssues = [],
    isLoading: pageLoading,
    selection: selectedRows,
    sort = {},
    state: selectedState,
    detailsTab: tabSelection
  } = deviceListState;
  const { direction: sortDown = SORTING_OPTIONS.desc, key: sortCol } = sort;
  const { canManageDevices, canManageUsers } = userCapabilities;
  const { hasMonitor } = tenantCapabilities;
  const currentSelectedState = states[selectedState] ?? states.devices;
  const [columnHeaders, setColumnHeaders] = useState([]);
  const [isInitialized, setIsInitialized] = useState(false);
  const [devicesInitialized, setDevicesInitialized] = useState(!!devices.length);
  const [showFilters, setShowFilters] = useState(false);
  const [showCustomization, setShowCustomization] = useState(false);
  const deviceListRef = useRef();
  const timer = useRef();
  const navigate = useNavigate();

  // eslint-disable-next-line no-unused-vars
  const size = useWindowSize();

  const { classes } = useStyles();

  useEffect(() => {
    clearAllRetryTimers(dispatchedSetSnackbar);
    if (!filters.length && selectedGroup && groupFilters.length) {
      dispatch(setDeviceFilters(groupFilters));
    }
    return () => {
      clearInterval(timer.current);
      clearAllRetryTimers(dispatchedSetSnackbar);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dispatch, dispatchedSetSnackbar]);

  useEffect(() => {
    const columnHeaders = getHeaders(columnSelection, currentSelectedState.defaultHeaders, idAttribute, canManageUsers ? openSettingsDialog : undefined);
    setColumnHeaders(columnHeaders);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [canManageUsers, columnSelection, idAttribute.attribute, currentSelectedState.defaultHeaders, openSettingsDialog]);

  useEffect(() => {
    // only set state after all devices id data retrieved
    setIsInitialized(isInitialized => isInitialized || (settingsInitialized && devicesInitialized && pageLoading === false));
    setDevicesInitialized(devicesInitialized => devicesInitialized || pageLoading === false);
  }, [settingsInitialized, devicesInitialized, pageLoading]);

  const onDeviceStateSelectionChange = useCallback(
    newState => dispatch(setDeviceListState({ state: newState, page: 1, refreshTrigger: !refreshTrigger }, true, false, false)),
    [dispatch, refreshTrigger]
  );

  useEffect(() => {
    if (onboardingState.complete) {
      return;
    }
    if (pendingCount) {
      dispatch(advanceOnboarding(onboardingSteps.DEVICES_PENDING_ONBOARDING_START));
      return;
    }
    if (!acceptedCount) {
      return;
    }

    if (acceptedCount < 2 && !window.sessionStorage.getItem('pendings-redirect')) {
      window.sessionStorage.setItem('pendings-redirect', true);
      onDeviceStateSelectionChange(DEVICE_STATES.accepted);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [acceptedCount, allCount, pendingCount, onboardingState.complete, dispatch, onDeviceStateSelectionChange, dispatchedSetSnackbar]);

  useEffect(() => {
    setShowFilters(false);
  }, [selectedGroup]);
  const dispatchDeviceListState = useCallback(
    (options, shouldSelectDevices = true, forceRefresh = false, fetchAuth = false) => {
      return dispatch(setDeviceListState(options, shouldSelectDevices, forceRefresh, fetchAuth));
    },
    [dispatch]
  );

  const refreshDevices = useCallback(() => {
    const refreshLength = deviceRefreshTimes[selectedState] ?? deviceRefreshTimes.default;
    return dispatchDeviceListState({}, true, true).catch(err =>
      setRetryTimer(err, 'devices', `Devices couldn't be loaded.`, refreshLength, dispatchedSetSnackbar)
    );
  }, [dispatchedSetSnackbar, selectedState, dispatchDeviceListState]);

  useEffect(() => {
    if (!devicesInitialized) {
      return;
    }
    const refreshLength = deviceRefreshTimes[selectedState] ?? deviceRefreshTimes.default;
    clearInterval(timer.current);
    timer.current = setInterval(() => refreshDevices(), refreshLength);
  }, [devicesInitialized, refreshDevices, selectedState]);

  useEffect(() => {
    Object.keys(availableIssueOptions).map(key => dispatch(getIssueCountsByType(key, { filters, group: selectedGroup, state: selectedState })));
    availableIssueOptions[DEVICE_ISSUE_OPTIONS.authRequests.key]
      ? dispatch(getIssueCountsByType(DEVICE_ISSUE_OPTIONS.authRequests.key, { filters: [] }))
      : undefined;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedIssues.join(''), JSON.stringify(availableIssueOptions), selectedState, selectedGroup, dispatch, JSON.stringify(filters)]);

  /*
   * Devices
   */
  const devicesToIds = devices => devices.map(device => device.id);

  const onRemoveDevicesFromGroup = devices => {
    const deviceIds = devicesToIds(devices);
    removeDevicesFromGroup(deviceIds);
    // if devices.length = number on page but < deviceCount
    // move page back to pageNO 1
    if (devices.length === deviceIds.length) {
      handlePageChange(1);
    }
  };

  const onAuthorizationChange = (devices, changedState) => {
    const deviceIds = devicesToIds(devices);
    return dispatchDeviceListState({ isLoading: true })
      .then(() => dispatch(updateDevicesAuth(deviceIds, changedState)))
      .then(() => onSelectionChange([]));
  };

  const onDeviceDismiss = devices =>
    dispatchDeviceListState({ isLoading: true })
      .then(() => {
        const deleteRequests = devices.reduce((accu, device) => {
          if (device.auth_sets?.length) {
            accu.push(dispatch(deleteAuthset(device.id, device.auth_sets[0].id)));
          }
          return accu;
        }, []);
        return Promise.all(deleteRequests);
      })
      .then(() => onSelectionChange([]));

  const handlePageChange = useCallback(page => dispatchDeviceListState({ selectedId: undefined, page }), [dispatchDeviceListState]);

  const onPageLengthChange = perPage => dispatchDeviceListState({ perPage, page: 1, refreshTrigger: !refreshTrigger });

  const onSortChange = attribute => {
    let changedSortCol = attribute.name;
    let changedSortDown = sortDown === SORTING_OPTIONS.desc ? SORTING_OPTIONS.asc : SORTING_OPTIONS.desc;
    if (changedSortCol !== sortCol) {
      changedSortDown = SORTING_OPTIONS.desc;
    }
    dispatchDeviceListState({
      sort: { direction: changedSortDown, key: changedSortCol, scope: attribute.scope },
      refreshTrigger: !refreshTrigger
    });
  };

  const setDetailsTab = detailsTab => dispatchDeviceListState({ detailsTab, setOnly: true });

  const onDeviceIssuesSelectionChange = ({ target: { value: selectedIssues } }) =>
    dispatchDeviceListState({ selectedIssues, page: 1, refreshTrigger: !refreshTrigger });

  const onSelectionChange = (selection = []) => {
    if (!onboardingState.complete && selection.length) {
      dispatch(advanceOnboarding(onboardingSteps.DEVICES_PENDING_ACCEPTING_ONBOARDING));
    }
    dispatchDeviceListState({ selection, setOnly: true });
  };

  const onToggleCustomizationClick = () => setShowCustomization(toggle);

  const onChangeColumns = useCallback(
    (changedColumns, customColumnSizes) => {
      const { columnSizes, selectedAttributes } = calculateColumnSelectionSize(changedColumns, customColumnSizes);
      dispatch(updateUserColumnSettings(columnSizes));
      dispatch(saveUserSettings({ columnSelection: changedColumns }));
      // we don't need an explicit refresh trigger here, since the selectedAttributes will be different anyway & otherwise the shown list should still be valid
      dispatchDeviceListState({ selectedAttributes });
      setShowCustomization(false);
    },
    [dispatch, dispatchDeviceListState]
  );

  const onExpandClick = (device = {}) => {
    dispatchedSetSnackbar('');
    const { id } = device;
    dispatchDeviceListState({ selectedId: deviceListState.selectedId === id ? undefined : id, detailsTab: deviceListState.detailsTab || 'identity' });
    if (!onboardingState.complete) {
      dispatch(advanceOnboarding(onboardingSteps.DEVICES_PENDING_ONBOARDING));
    }
  };

  const onCreateDeploymentClick = devices => navigate(`/deployments?open=true&${devices.map(({ id }) => `deviceId=${id}`).join('&')}`);

  const onCloseExpandedDevice = useCallback(() => dispatchDeviceListState({ selectedId: undefined, detailsTab: '' }), [dispatchDeviceListState]);

  const onResizeColumns = useCallback(columns => dispatch(updateUserColumnSettings(columns)), [dispatch]);

  const actionCallbacks = {
    onAddDevicesToGroup: addDevicesToGroup,
    onAuthorizationChange,
    onCreateDeployment: onCreateDeploymentClick,
    onDeviceDismiss,
    onPromoteGateway: onMakeGatewayClick,
    onRemoveDevicesFromGroup
  };

  const listOptionHandlers = [{ key: 'customize', title: 'Customize', onClick: onToggleCustomizationClick }];
  const EmptyState = currentSelectedState.emptyState;

  const groupLabel = selectedGroup ? decodeURIComponent(selectedGroup) : ALL_DEVICES;
  const isUngroupedGroup = selectedGroup && selectedGroup === UNGROUPED_GROUP.id;
  const selectedStaticGroup = selectedGroup && !groupFilters.length ? selectedGroup : undefined;

  const openedDevice = useDebounce(selectedId, TIMEOUTS.debounceShort);
  return (
    <>
      <div className="margin-left-small">
        <div className="flexbox">
          <h3 className="margin-right">{isUngroupedGroup ? UNGROUPED_GROUP.name : groupLabel}</h3>
          <div className="flexbox space-between center-aligned" style={{ flexGrow: 1 }}>
            <div className="flexbox">
              <DeviceStateSelection onStateChange={onDeviceStateSelectionChange} selectedState={selectedState} states={states} />
              {hasMonitor && (
                <DeviceIssuesSelection
                  classes={classes}
                  onChange={onDeviceIssuesSelectionChange}
                  options={Object.values(availableIssueOptions)}
                  selection={selectedIssues}
                />
              )}
              {selectedGroup && !isUngroupedGroup && (
                <div className="margin-left muted flexbox centered">
                  {!groupFilters.length ? <LockOutlined fontSize="small" /> : <AutorenewIcon fontSize="small" />}
                  <span>{!groupFilters.length ? 'Static' : 'Dynamic'}</span>
                </div>
              )}
            </div>
            {canManageDevices && selectedGroup && !isUngroupedGroup && (
              <Button onClick={onGroupRemoval} startIcon={<DeleteIcon />}>
                Remove group
              </Button>
            )}
          </div>
        </div>
        <div className="flexbox space-between">
          {!isUngroupedGroup && (
            <div className={`flexbox centered filter-header ${showFilters ? `${classes.filterCommon} filter-toggle` : ''}`}>
              <Button
                color="secondary"
                disableRipple
                onClick={() => setShowFilters(toggle)}
                startIcon={<FilterListIcon />}
                style={{ backgroundColor: 'transparent' }}
              >
                {filters.length > 0 ? `Filters (${filters.length})` : 'Filters'}
              </Button>
            </div>
          )}
          <ListOptions options={listOptionHandlers} title="Table options" />
        </div>
        <Filters className={classes.filterCommon} onGroupClick={onGroupClick} open={showFilters} />
      </div>
      <Loader show={!isInitialized} />
      <div className="padding-bottom" ref={deviceListRef}>
        {devices.length > 0 ? (
          <DeviceList
            columnHeaders={columnHeaders}
            customColumnSizes={customColumnSizes}
            devices={devices}
            deviceListState={deviceListState}
            idAttribute={idAttribute}
            onChangeRowsPerPage={onPageLengthChange}
            onExpandClick={onExpandClick}
            onPageChange={handlePageChange}
            onResizeColumns={onResizeColumns}
            onSelect={onSelectionChange}
            onSort={onSortChange}
            pageLoading={pageLoading}
            pageTotal={deviceCount}
          />
        ) : (
          <EmptyState allCount={allCount} canManageDevices={canManageDevices} filters={filters} limitMaxed={limitMaxed} onClick={onPreauthClick} />
        )}
      </div>
      <ExpandedDevice
        actionCallbacks={actionCallbacks}
        deviceId={openedDevice}
        onClose={onCloseExpandedDevice}
        setDetailsTab={setDetailsTab}
        tabSelection={tabSelection}
      />
      {!selectedId && !showsDialog && <OnboardingComponent deviceListRef={deviceListRef} onboardingState={onboardingState} />}
      {canManageDevices && !!selectedRows.length && (
        <DeviceQuickActions actionCallbacks={actionCallbacks} deviceId={openedDevice} selectedGroup={selectedStaticGroup} />
      )}
      <ColumnCustomizationDialog
        attributes={attributes}
        columnHeaders={columnHeaders}
        customColumnSizes={customColumnSizes}
        idAttribute={idAttribute}
        open={showCustomization}
        onCancel={onToggleCustomizationClick}
        onSubmit={onChangeColumns}
      />
    </>
  );
};

export default Authorized;

export const DeviceStateSelection = ({ onStateChange, selectedState = '', states }) => {
  const { classes } = useStyles();
  const availableStates = useMemo(() => Object.values(states).filter(duplicateFilter), [states]);

  return (
    <div className="flexbox centered">
      Status:
      <Select className={classes.selection} disableUnderline onChange={e => onStateChange(e.target.value)} value={selectedState}>
        {availableStates.map(state => (
          <MenuItem key={state.key} value={state.key}>
            {state.title()}
          </MenuItem>
        ))}
      </Select>
    </div>
  );
};
