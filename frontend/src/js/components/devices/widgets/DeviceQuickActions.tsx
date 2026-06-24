// Copyright 2021 Northern.tech AS
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
import { useEffect, useRef, useState } from 'react';
import { useSelector } from 'react-redux';

import {
  AddCircle as AddCircleIcon,
  CheckCircle as CheckCircleIcon,
  HeightOutlined as HeightOutlinedIcon,
  HighlightOffOutlined as HighlightOffOutlinedIcon,
  RemoveCircleOutlined as RemoveCircleOutlineIcon,
  SyncOutlined as SyncOutlinedIcon
} from '@mui/icons-material';

import { mdiTrashCanOutline as TrashCan } from '@mdi/js';
import { ConfirmModal } from '@northern.tech/common-ui/ConfirmModal';
import MaterialDesignIcon from '@northern.tech/common-ui/MaterialDesignIcon';
import { BaseQuickActions, type QuickAction } from '@northern.tech/common-ui/QuickActions';
import { DEVICE_STATES, TIMEOUTS, UNGROUPED_GROUP, onboardingSteps } from '@northern.tech/store/constants';
import { advanceOnboarding } from '@northern.tech/store/onboardingSlice/thunks';
import {
  getDeviceById,
  getFeatures,
  getMappedDevicesList,
  getOnboardingState,
  getTenantCapabilities,
  getUserCapabilities
} from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { stringToBoolean, toggle } from '@northern.tech/utils/helpers';
import pluralize from 'pluralize';

import GatewayIcon from '../../../../assets/img/gateway.svg';
import { getOnboardingComponentFor } from '../../../utils/onboardingManager';

const defaultActions = {
  accept: {
    icon: <CheckCircleIcon className="green" />,
    key: 'accept',
    title: pluralized => `Accept ${pluralized}`,
    action: ({ onAuthorizationChange, selection }) => onAuthorizationChange(selection, DEVICE_STATES.accepted),
    checkRelevance: ({ device, userCapabilities: { canWriteDevices } }) =>
      canWriteDevices && [DEVICE_STATES.pending, DEVICE_STATES.rejected].includes(device.status)
  },
  dismiss: {
    icon: <RemoveCircleOutlineIcon />,
    key: 'dismiss',
    title: pluralized => `Dismiss ${pluralized}`,
    action: ({ onDeviceDismiss, selection }) => onDeviceDismiss(selection),
    checkRelevance: ({ device, userCapabilities: { canWriteDevices } }) =>
      canWriteDevices && [DEVICE_STATES.accepted, DEVICE_STATES.pending, DEVICE_STATES.preauth, DEVICE_STATES.rejected, 'noauth'].includes(device.status)
  },
  reject: {
    icon: <HighlightOffOutlinedIcon className="red" />,
    key: 'reject',
    title: pluralized => `Reject ${pluralized}`,
    action: ({ onAuthorizationChange, selection }) => onAuthorizationChange(selection, DEVICE_STATES.rejected),
    checkRelevance: ({ device, userCapabilities: { canWriteDevices } }) =>
      canWriteDevices && [DEVICE_STATES.accepted, DEVICE_STATES.pending].includes(device.status)
  },
  addToGroup: {
    icon: <AddCircleIcon className="green" />,
    key: 'group-add',
    title: pluralized => `Add selected ${pluralized} to a group`,
    action: ({ onAddDevicesToGroup, selection }) => onAddDevicesToGroup(selection),
    checkRelevance: ({ selectedGroup, userCapabilities: { canWriteDevices } }) => canWriteDevices && !selectedGroup
  },
  moveToGroup: {
    icon: <HeightOutlinedIcon className="rotated ninety" />,
    key: 'group-change',
    title: pluralized => `Move selected ${pluralized} to another group`,
    action: ({ onAddDevicesToGroup, selection }) => onAddDevicesToGroup(selection),
    checkRelevance: ({ selectedGroup, userCapabilities: { canWriteDevices } }) => canWriteDevices && !!selectedGroup
  },
  removeFromGroup: {
    icon: <MaterialDesignIcon path={TrashCan} />,
    key: 'group-remove',
    title: pluralized => `Remove selected ${pluralized} from this group`,
    action: ({ onRemoveDevicesFromGroup, selection }) => onRemoveDevicesFromGroup(selection),
    checkRelevance: ({ selectedGroup, userCapabilities: { canWriteDevices } }) => canWriteDevices && selectedGroup && selectedGroup !== UNGROUPED_GROUP.id
  },
  promoteToGateway: {
    icon: <GatewayIcon style={{ width: 20 }} />,
    key: 'promote-to-gateway',
    title: () => 'Promote to gateway',
    action: ({ onPromoteGateway, selection }) => onPromoteGateway(selection),
    checkRelevance: ({ device, features, tenantCapabilities: { isEnterprise } }) =>
      features.isHosted && isEnterprise && !stringToBoolean(device.attributes?.mender_is_gateway) && device.status === DEVICE_STATES.accepted
  },
  createDeployment: {
    icon: <SyncOutlinedIcon />,
    key: 'create-deployment',
    title: (pluralized, count) => `Create deployment for ${pluralize('this', count)} ${pluralized}`,
    action: ({ onCreateDeployment, selection }) => onCreateDeployment(selection),
    checkRelevance: ({ device, userCapabilities: { canDeploy, canReadReleases } }) =>
      canDeploy && canReadReleases && device && device.status === DEVICE_STATES.accepted && device.attributes?.device_type?.length
  }
};

type ConfirmAction = {
  buttonText: string;
  description: (pluralized: string, selectedDevices: number[]) => string;
  header: (pluralized: string) => string;
};

const confirmActions: Record<'dismiss' | 'reject' | 'default', ConfirmAction> = {
  default: { description: () => '', header: () => '', buttonText: '' },
  dismiss: {
    description: (pluralized, selectedDevices) =>
      `Are you sure you want to dismiss ${selectedDevices.length} ${pluralized}? The ${pluralized} will be removed from the UI.`,
    header: pluralized => `Dismiss ${pluralized}?`,
    buttonText: 'Dismiss'
  },
  reject: {
    description: (pluralized, selectedDevices) =>
      `Are you sure you want to reject ${selectedDevices.length} ${pluralized}? The ${pluralized} will be blocked from communicating with the Mender server.`,
    header: pluralized => `Reject ${pluralized}?`,
    buttonText: 'Reject'
  }
};

export const DeviceQuickActions = ({ actionCallbacks, deviceId, selectedGroup }) => {
  const dispatch = useAppDispatch();
  const features = useSelector(getFeatures);
  const tenantCapabilities = useSelector(getTenantCapabilities);
  const userCapabilities = useSelector(getUserCapabilities);
  const { selection: selectedRows } = useSelector(state => state.devices.deviceList);
  const singleDevice = useSelector(state => getDeviceById(state, deviceId));
  const devices = useSelector(state => getMappedDevicesList(state, 'deviceList'));
  const deviceActionRef = useRef<HTMLDivElement>();
  const deviceActionLabelRef = useRef<HTMLDivElement>(null);
  const deploymentActionRef = useRef<HTMLDivElement>(null);
  const onboardingState = useSelector(getOnboardingState);
  const [isInitialized, setIsInitialized] = useState(false);
  const [confirmAction, setConfirmAction] = useState<{ key: string; selection: any[] } | null>(null);
  const timer = useRef();

  useEffect(() => {
    clearTimeout(timer.current);
    timer.current = setTimeout(() => setIsInitialized(toggle), TIMEOUTS.debounceDefault);
    return () => {
      clearTimeout(timer.current);
    };
  }, []);

  const selectedDevices = deviceId ? [singleDevice] : selectedRows.map(row => devices[row]);
  const pluralized = pluralize('devices', selectedDevices.length);

  const confirmableActions = new Set(['reject', 'dismiss']);

  const actions: QuickAction[] = Object.values(defaultActions)
    .filter(action =>
      selectedDevices.every(device => device && action.checkRelevance({ device, features, selectedGroup, tenantCapabilities, userCapabilities }))
    )
    .map(({ action, key, icon, title }) => ({
      key,
      icon,
      title: <div ref={key === 'create-deployment' ? deploymentActionRef : undefined}>{title(pluralized, selectedDevices.length)}</div>,
      onClick: () => {
        if (confirmableActions.has(key)) {
          setConfirmAction({ key, selection: selectedDevices });
        } else {
          action({ ...actionCallbacks, selection: selectedDevices });
        }
      }
    }));

  const handleToggle = () => dispatch(advanceOnboarding(onboardingSteps.DEVICES_DEPLOY_RELEASE_ONBOARDING));

  let onboardingComponent;
  let anchor;
  if (deviceActionLabelRef.current && deploymentActionRef.current && isInitialized) {
    anchor = {
      left: deploymentActionRef.current.parentElement.offsetLeft + deploymentActionRef.current.parentElement.parentElement.offsetWidth + 45,
      top: deploymentActionRef.current.parentElement.offsetTop + 15
    };
    onboardingComponent = getOnboardingComponentFor(onboardingSteps.DEVICES_DEPLOY_RELEASE_ONBOARDING_STEP_2, onboardingState, { anchor, place: 'left' }, null);
  } else if (deviceActionLabelRef.current && deviceActionRef.current && isInitialized) {
    anchor = {
      left: deviceActionLabelRef.current.offsetLeft - 55,
      top: deviceActionRef.current.offsetHeight - (deviceActionLabelRef.current.offsetHeight + 15) / 2
    };
    onboardingComponent = getOnboardingComponentFor(onboardingSteps.DEVICES_DEPLOY_RELEASE_ONBOARDING, onboardingState, { anchor, place: 'left' }, null);
  }

  const { description, header, buttonText } = confirmAction?.key ? confirmActions[confirmAction.key] : confirmActions.default;

  return (
    <>
      <BaseQuickActions
        actions={actions}
        ariaLabel="device-actions"
        label={deviceId ? 'Device actions' : `${selectedDevices.length} ${pluralized} selected`}
        onToggle={handleToggle}
        onboardingComponent={onboardingComponent}
        speedDialRef={deviceActionRef}
        titleRef={deviceActionLabelRef}
      />
      <ConfirmModal
        close={() => setConfirmAction(null)}
        confirmButtonText={buttonText}
        description={description(pluralized, selectedDevices)}
        header={header(pluralized, selectedDevices)}
        onConfirm={() => {
          if (confirmAction) {
            defaultActions[confirmAction.key].action({ ...actionCallbacks, selection: confirmAction.selection });
          }
          setConfirmAction(null);
        }}
        open={!!confirmAction}
      />
    </>
  );
};

export default DeviceQuickActions;
