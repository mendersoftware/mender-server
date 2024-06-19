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
import React, { useEffect, useRef, useState } from 'react';
import { useSelector } from 'react-redux';

import {
  AddCircle as AddCircleIcon,
  CheckCircle as CheckCircleIcon,
  HeightOutlined as HeightOutlinedIcon,
  HighlightOffOutlined as HighlightOffOutlinedIcon,
  RemoveCircleOutline as RemoveCircleOutlineIcon,
  Replay as ReplayIcon
} from '@mui/icons-material';
import { ClickAwayListener, SpeedDial, SpeedDialAction, SpeedDialIcon } from '@mui/material';
import { speedDialActionClasses } from '@mui/material/SpeedDialAction';
import { makeStyles } from 'tss-react/mui';

import { mdiTrashCanOutline as TrashCan } from '@mdi/js';
import pluralize from 'pluralize';

import GatewayIcon from '../../../../assets/img/gateway.svg';
import { TIMEOUTS } from '../../../constants/appConstants';
import { DEVICE_STATES, UNGROUPED_GROUP } from '../../../constants/deviceConstants';
import { onboardingSteps } from '../../../constants/onboardingConstants';
import { stringToBoolean, toggle } from '../../../helpers';
import { getDeviceById, getFeatures, getMappedDevicesList, getOnboardingState, getTenantCapabilities, getUserCapabilities } from '../../../selectors';
import { getOnboardingComponentFor } from '../../../utils/onboardingmanager';
import MaterialDesignIcon from '../../common/materialdesignicon';

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
    icon: <RemoveCircleOutlineIcon className="red" />,
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
    icon: <ReplayIcon />,
    key: 'create-deployment',
    title: (pluralized, count) => `Create deployment for ${pluralize('this', count)} ${pluralized}`,
    action: ({ onCreateDeployment, selection }) => onCreateDeployment(selection),
    checkRelevance: ({ device, userCapabilities: { canDeploy, canReadReleases } }) =>
      canDeploy && canReadReleases && device && device.status === DEVICE_STATES.accepted
  }
};

const useStyles = makeStyles()(theme => ({
  container: {
    position: 'fixed',
    bottom: theme.spacing(6.5),
    right: theme.spacing(6.5),
    zIndex: 10,
    minWidth: 400,
    pointerEvents: 'none',
    [`.${speedDialActionClasses.staticTooltipLabel}`]: {
      minWidth: 'max-content'
    }
  },
  fab: { margin: theme.spacing(2) },
  innerContainer: {
    display: 'flex',
    alignItems: 'flex-end',
    justifyContent: 'flex-end'
  },
  label: {
    marginRight: theme.spacing(2),
    marginBottom: theme.spacing(4)
  }
}));

export const DeviceQuickActions = ({ actionCallbacks, deviceId, selectedGroup }) => {
  const [showActions, setShowActions] = useState(false);
  const features = useSelector(getFeatures);
  const tenantCapabilities = useSelector(getTenantCapabilities);
  const userCapabilities = useSelector(getUserCapabilities);
  const { selection: selectedRows } = useSelector(state => state.devices.deviceList);
  const singleDevice = useSelector(state => getDeviceById(state, deviceId));
  const devices = useSelector(state => getMappedDevicesList(state, 'deviceList'));
  const { classes } = useStyles();
  const deployActionRef = useRef();
  const onboardingState = useSelector(getOnboardingState);
  const [isInitialized, setIsInitialized] = useState(false);
  const timer = useRef();

  const handleShowActions = () => {
    setShowActions(!showActions);
  };

  const handleClickAway = () => {
    setShowActions(false);
  };

  useEffect(() => {
    clearTimeout(timer.current);
    timer.current = setTimeout(() => setIsInitialized(toggle), TIMEOUTS.debounceDefault);
    return () => {
      clearTimeout(timer.current);
    };
  }, []);

  const selectedDevices = deviceId ? [singleDevice] : selectedRows.map(row => devices[row]);
  const actions = Object.values(defaultActions).reduce((accu, action) => {
    if (selectedDevices.every(device => device && action.checkRelevance({ device, features, selectedGroup, tenantCapabilities, userCapabilities }))) {
      accu.push(action);
    }
    return accu;
  }, []);

  const pluralized = pluralize('devices', selectedDevices.length);

  let onboardingComponent;
  if (deployActionRef.current && isInitialized) {
    const anchor = {
      left: deployActionRef.current.firstElementChild.offsetLeft - 15,
      top: deployActionRef.current.offsetTop + deployActionRef.current.firstElementChild.offsetTop + deployActionRef.current.firstElementChild.offsetHeight / 2
    };
    onboardingComponent = getOnboardingComponentFor(
      onboardingSteps.DEVICES_DEPLOY_RELEASE_ONBOARDING,
      onboardingState,
      { anchor, place: 'left' },
      onboardingComponent
    );
  }
  return (
    <div className={classes.container}>
      <div className="relative">
        <div className={classes.innerContainer} ref={deployActionRef}>
          <div className={classes.label}>{deviceId ? 'Device actions' : `${selectedDevices.length} ${pluralized} selected`}</div>
          <ClickAwayListener onClickAway={handleClickAway}>
            <SpeedDial className={classes.fab} ariaLabel="device-actions" icon={<SpeedDialIcon />} onClick={handleShowActions} open={Boolean(showActions)}>
              {actions.map(action => (
                <SpeedDialAction
                  key={action.key}
                  aria-label={action.key}
                  icon={action.icon}
                  tooltipTitle={action.title(pluralized, selectedDevices.length)}
                  tooltipOpen
                  onClick={() => action.action({ ...actionCallbacks, selection: selectedDevices })}
                />
              ))}
            </SpeedDial>
          </ClickAwayListener>
        </div>
        {onboardingComponent}
      </div>
    </div>
  );
};

export default DeviceQuickActions;
