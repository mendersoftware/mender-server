// Copyright 2018 Northern.tech AS
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
import { useDispatch, useSelector } from 'react-redux';
import { useLocation, useParams } from 'react-router-dom';

import { DialogContent } from '@mui/material';

import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';
import storeActions from '@northern.tech/store/actions';
import { DEVICE_FILTERING_OPTIONS, DEVICE_STATES, SORTING_OPTIONS, emptyFilter, onboardingSteps } from '@northern.tech/store/constants';
import { useLocationParams } from '@northern.tech/store/liststatehook';
import {
  getAcceptedDevices,
  getDeviceCountsByStatus,
  getDeviceFilters,
  getDeviceLimit,
  getFeatures,
  getGroups as getGroupsSelector,
  getIsEnterprise,
  getIsPreview,
  getLimitMaxed,
  getOnboardingState,
  getSelectedGroupInfo,
  getSortedFilteringAttributes,
  getTenantCapabilities,
  getUserCapabilities
} from '@northern.tech/store/selectors';
import {
  addDynamicGroup,
  addStaticGroup,
  removeDevicesFromGroup,
  removeDynamicGroup,
  removeStaticGroup,
  selectGroup,
  setDeviceListState,
  setOfflineThreshold,
  updateDynamicGroup
} from '@northern.tech/store/thunks';
import { toggle } from '@northern.tech/utils/helpers';

import { getOnboardingComponentFor } from '../../utils/onboardingManager';
import Global from '../settings/Global';
import AuthorizedDevices from './AuthorizedDevices';
import DeviceStatusNotification from './DeviceStatusNotification';
import Groups from './Groups';
import MakeGatewayDialog from './dialogs/MakeGatewayDialog';
import PreauthDialog, { DeviceLimitWarning } from './dialogs/PreauthDialog';
import CreateGroup from './group-management/CreateGroup';
import CreateGroupExplainer from './group-management/CreateGroupExplainer';
import RemoveGroup from './group-management/RemoveGroup';
import DeviceAdditionWidget from './widgets/DeviceAdditionWidget';

const { setDeviceFilters, setShowConnectingDialog } = storeActions;

export const DeviceGroups = () => {
  const [createGroupExplanation, setCreateGroupExplanation] = useState(false);
  const [fromFilters, setFromFilters] = useState(false);
  const [modifyGroupDialog, setModifyGroupDialog] = useState(false);
  const [openIdDialog, setOpenIdDialog] = useState(false);
  const [openPreauth, setOpenPreauth] = useState(false);
  const [showMakeGateway, setShowMakeGateway] = useState(false);
  const [removeGroup, setRemoveGroup] = useState(false);
  const [tmpDevices, setTmpDevices] = useState([]);
  const deviceConnectionRef = useRef();
  const { status: statusParam } = useParams();

  const { groupCount, selectedGroup, groupFilters = [] } = useSelector(getSelectedGroupInfo);
  const filteringAttributes = useSelector(getSortedFilteringAttributes);
  const { canManageDevices } = useSelector(getUserCapabilities);
  const tenantCapabilities = useSelector(getTenantCapabilities);
  const { groupNames, ...groupsByType } = useSelector(getGroupsSelector);
  const groups = groupNames;
  const { total: acceptedCount = 0 } = useSelector(getAcceptedDevices);
  const canPreview = useSelector(getIsPreview);
  const deviceLimit = useSelector(getDeviceLimit);
  const deviceListState = useSelector(state => state.devices.deviceList);
  const features = useSelector(getFeatures);
  const filters = useSelector(getDeviceFilters);
  const limitMaxed = useSelector(getLimitMaxed);
  const { pending: pendingCount } = useSelector(getDeviceCountsByStatus);
  const showDeviceConnectionDialog = useSelector(state => state.users.showConnectDeviceDialog);
  const onboardingState = useSelector(getOnboardingState);
  const isEnterprise = useSelector(getIsEnterprise);
  const dispatch = useDispatch();
  const isInitialized = useRef(false);
  const location = useLocation();

  const [locationParams, setLocationParams] = useLocationParams('devices', {
    filteringAttributes,
    filters,
    defaults: { sort: { direction: SORTING_OPTIONS.desc } }
  });

  const { refreshTrigger, selectedId, state: selectedState } = deviceListState;

  useEffect(() => {
    if (!isInitialized.current) {
      return;
    }
    setLocationParams({ pageState: deviceListState, filters, selectedGroup });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [
    deviceListState.detailsTab,
    deviceListState.page,
    deviceListState.perPage,
    deviceListState.selectedIssues,
    // eslint-disable-next-line react-hooks/exhaustive-deps
    JSON.stringify(deviceListState.sort),
    selectedId,
    filters,
    selectedGroup,
    selectedState,
    setLocationParams
  ]);

  useEffect(() => {
    // set isInitialized ref to false when location changes, otherwise when you go back setLocationParams will be set with a duplicate item
    isInitialized.current = false;
  }, [location]);

  useEffect(() => {
    const { groupName, filters = [], id = [], ...remainder } = locationParams;
    const { hasFullFiltering } = tenantCapabilities;
    if (groupName) {
      if (groupName != selectedGroup) {
        dispatch(selectGroup({ group: groupName, filters }));
      }
    } else if (filters.length) {
      // dispatch setDeviceFilters even when filters are empty, otherwise filter will not be reset
      dispatch(setDeviceFilters(filters));
    }
    // preset selectedIssues and selectedId with empty values, in case if remain properties are missing them
    const listState = { ...remainder };
    if (statusParam && (Object.values(DEVICE_STATES).some(state => state === statusParam) || statusParam === 'any')) {
      listState.state = statusParam;
    }

    if (id.length === 1 && Boolean(locationParams.open)) {
      listState.selectedId = id[0];
    } else if (id.length && hasFullFiltering) {
      dispatch(setDeviceFilters([...filters, { ...emptyFilter, key: 'id', operator: DEVICE_FILTERING_OPTIONS.$in.key, value: id }]));
    }
    dispatch(setDeviceListState(listState)).then(() => {
      if (isInitialized.current) {
        return;
      }
      isInitialized.current = true;
      dispatch(setDeviceListState({ shouldSelectDevices: true, forceRefresh: true }));
      dispatch(setOfflineThreshold());
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dispatch, JSON.stringify(tenantCapabilities), JSON.stringify(locationParams), statusParam]);

  /*
   * Groups
   */
  const removeCurrentGroup = () => {
    const request = groupFilters.length ? dispatch(removeDynamicGroup(selectedGroup)) : dispatch(removeStaticGroup(selectedGroup));
    return request.then(toggleGroupRemoval).catch(console.log);
  };

  // Edit groups from device selection
  const addDevicesToGroup = tmpDevices => {
    // (save selected devices in state, open dialog)
    setTmpDevices(tmpDevices);
    setModifyGroupDialog(toggle);
  };

  const createGroupFromDialog = (devices, group) => {
    const request = fromFilters ? dispatch(addDynamicGroup({ groupName: group, filterPredicates: filters })) : dispatch(addStaticGroup({ group, devices }));
    return request.then(() => {
      // reached end of list
      setCreateGroupExplanation(false);
      setModifyGroupDialog(false);
      setFromFilters(false);
    });
  };

  const onGroupClick = () => {
    if (selectedGroup && groupFilters.length) {
      return dispatch(updateDynamicGroup({ groupName: selectedGroup, filterPredicates: filters }));
    }
    setModifyGroupDialog(true);
    setFromFilters(true);
  };

  const onRemoveDevicesFromGroup = devices => {
    const isGroupRemoval = devices.length >= groupCount;
    let request;
    if (isGroupRemoval) {
      request = dispatch(removeStaticGroup(selectedGroup));
    } else {
      request = dispatch(removeDevicesFromGroup({ group: selectedGroup, deviceIds: devices }));
    }
    return request.catch(console.log);
  };

  const openSettingsDialog = e => {
    e.preventDefault();
    setOpenIdDialog(toggle);
  };

  const onCreateGroupClose = () => {
    setModifyGroupDialog(false);
    setFromFilters(false);
    setTmpDevices([]);
  };

  const onPreauthSaved = addMore => {
    setOpenPreauth(!addMore);
    dispatch(setDeviceListState({ page: 1, refreshTrigger: !refreshTrigger }));
  };

  const onShowDeviceStateClick = state => {
    dispatch(selectGroup());
    dispatch(setDeviceListState({ state }));
  };

  const onGroupSelect = groupName => {
    dispatch(selectGroup({ group: groupName }));
    dispatch(setDeviceListState({ page: 1, refreshTrigger: !refreshTrigger, selection: [] }));
  };

  const toggleGroupRemoval = () => setRemoveGroup(toggle);

  const toggleMakeGatewayClick = () => setShowMakeGateway(toggle);

  const changeLocation = useCallback(
    (newLocation: string) => {
      isInitialized.current = false;
      setLocationParams({ pageState: { ...deviceListState, state: newLocation }, filters, selectedGroup });
    },
    [setLocationParams, deviceListState, filters, selectedGroup]
  );

  let onboardingComponent;
  if (deviceConnectionRef.current && !(pendingCount || acceptedCount)) {
    const anchor = { top: deviceConnectionRef.current.offsetTop + deviceConnectionRef.current.offsetHeight / 2, left: deviceConnectionRef.current.offsetLeft };
    onboardingComponent = getOnboardingComponentFor(
      onboardingSteps.DEVICES_DELAYED_ONBOARDING,
      onboardingState,
      { anchor, place: 'left' },
      onboardingComponent
    );
  }
  return (
    <>
      <div className="tab-container with-sub-panels" style={{ paddingTop: 0, paddingBottom: 45, minHeight: 'max-content', alignContent: 'center' }}>
        <h3 className="flexbox center-aligned" style={{ marginBottom: 0, marginTop: 0, flexWrap: 'wrap' }}>
          Devices
        </h3>
        <span className="flexbox space-between margin-left-large margin-right center-aligned padding-top-small">
          {!!pendingCount && !selectedGroup && selectedState !== DEVICE_STATES.pending ? (
            <DeviceStatusNotification deviceCount={pendingCount} state={DEVICE_STATES.pending} onClick={onShowDeviceStateClick} />
          ) : (
            <div />
          )}
          {canManageDevices && (
            <DeviceAdditionWidget
              features={features}
              onConnectClick={() => dispatch(setShowConnectingDialog(true))}
              onMakeGatewayClick={toggleMakeGatewayClick}
              onPreauthClick={setOpenPreauth}
              tenantCapabilities={tenantCapabilities}
              innerRef={deviceConnectionRef}
            />
          )}
          {onboardingComponent}
        </span>
      </div>
      <div className="tab-container with-sub-panels" style={{ padding: 0, height: '100%' }}>
        <Groups
          className="leftFixed"
          acceptedCount={acceptedCount}
          changeGroup={onGroupSelect}
          groups={groupsByType}
          openGroupDialog={setCreateGroupExplanation}
          selectedGroup={selectedGroup}
        />
        <div className="rightFluid relative" style={{ paddingTop: 0 }}>
          {limitMaxed && <DeviceLimitWarning acceptedDevices={acceptedCount} deviceLimit={deviceLimit} />}
          <AuthorizedDevices
            changeLocation={changeLocation}
            addDevicesToGroup={addDevicesToGroup}
            onGroupClick={onGroupClick}
            onGroupRemoval={toggleGroupRemoval}
            onMakeGatewayClick={toggleMakeGatewayClick}
            onPreauthClick={setOpenPreauth}
            openSettingsDialog={openSettingsDialog}
            removeDevicesFromGroup={onRemoveDevicesFromGroup}
            showsDialog={showDeviceConnectionDialog || removeGroup || modifyGroupDialog || createGroupExplanation || openIdDialog || openPreauth}
          />
        </div>
        {removeGroup && <RemoveGroup onClose={toggleGroupRemoval} onRemove={removeCurrentGroup} />}
        {modifyGroupDialog && (
          <CreateGroup
            addListOfDevices={createGroupFromDialog}
            fromFilters={fromFilters}
            isCreation={fromFilters || !groups.length}
            selectedDevices={tmpDevices}
            onClose={onCreateGroupClose}
          />
        )}
        {createGroupExplanation && <CreateGroupExplainer isEnterprise={isEnterprise} onClose={() => setCreateGroupExplanation(false)} />}
        {openIdDialog && (
          <BaseDialog open title="Default device identity attribute" onClose={openSettingsDialog}>
            <DialogContent style={{ overflow: 'hidden' }}>
              <Global dialog closeDialog={openSettingsDialog} />
            </DialogContent>
          </BaseDialog>
        )}
        {openPreauth && (
          <PreauthDialog
            acceptedDevices={acceptedCount}
            deviceLimit={deviceLimit}
            limitMaxed={limitMaxed}
            onSubmit={onPreauthSaved}
            onCancel={() => setOpenPreauth(false)}
          />
        )}
        {showMakeGateway && <MakeGatewayDialog isPreRelease={canPreview} onCancel={toggleMakeGatewayClick} />}
      </div>
    </>
  );
};

export default DeviceGroups;
