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
import React, { useCallback, useEffect, useRef } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';

import { Chip, Divider, Drawer, Tab, Tabs, Tooltip, chipClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import DeviceIdentityDisplay from '@northern.tech/common-ui/DeviceIdentity';
import DocsLink from '@northern.tech/common-ui/DocsLink';
import { DrawerTitle } from '@northern.tech/common-ui/DrawerTitle';
import { MenderTooltipClickable } from '@northern.tech/common-ui/MenderTooltip';
import { RelativeTime } from '@northern.tech/common-ui/Time';
import storeActions from '@northern.tech/store/actions';
import { DEVICE_STATES, EXTERNAL_PROVIDER, TIMEOUTS, yes } from '@northern.tech/store/constants';
import {
  getDeviceConfigDeployment,
  getDeviceTwinIntegrations,
  getDevicesById,
  getDocsVersion,
  getGlobalSettings,
  getSelectedGroupInfo,
  getTenantCapabilities,
  getUserCapabilities,
  getUserSettings
} from '@northern.tech/store/selectors';
import { decommissionDevice, getDeviceInfo, getGatewayDevices, saveGlobalSettings } from '@northern.tech/store/thunks';
import { getDemoDeviceAddress, stringToBoolean } from '@northern.tech/utils/helpers';
import copy from 'copy-to-clipboard';

import GatewayConnectionIcon from '../../../assets/img/gateway-connection.svg';
import GatewayIcon from '../../../assets/img/gateway.svg';
import DeviceConfiguration from './device-details/Configuration';
import TroubleshootTab from './device-details/Connection';
import Deployments from './device-details/Deployments';
import DeviceInventory from './device-details/DeviceInventory';
import DeviceSystem from './device-details/DeviceSystem';
import { IntegrationTab } from './device-details/DeviceTwin';
import { IdentityTab } from './device-details/Identity';
import InstalledSoftware from './device-details/InstalledSoftware';
import MonitoringTab from './device-details/Monitoring';
import DeviceNotifications from './device-details/Notifications';
import DeviceQuickActions from './widgets/DeviceQuickActions';

const { setSnackbar } = storeActions;

const useStyles = makeStyles()(theme => ({
  gatewayChip: {
    backgroundColor: theme.palette.grey[400],
    color: theme.palette.grey[900],
    path: {
      fill: theme.palette.grey[900]
    },
    [`.${chipClasses.icon}`]: {
      marginLeft: 10,
      width: 20
    },
    [`.${chipClasses.icon}.connected`]: {
      transform: 'scale(1.3)',
      width: 15
    }
  },
  deviceConnection: {
    marginRight: theme.spacing(2)
  },
  dividerTop: {
    marginBottom: theme.spacing(3),
    marginTop: theme.spacing(2)
  }
}));

const refreshDeviceLength = TIMEOUTS.refreshDefault;

const GatewayConnectionNotification = ({ gatewayDevices, onClick }) => {
  const { classes } = useStyles();

  const onGatewayClick = () => {
    const query =
      gatewayDevices.length > 1 ? gatewayDevices.map(device => `id=${device.id}`).join('&') : `id=${gatewayDevices[0].id}&open=true&tab=device-system`;
    onClick(query);
  };

  return (
    <MenderTooltipClickable
      placement="bottom"
      title={
        <div style={{ maxWidth: 350 }}>
          Connected to{' '}
          {gatewayDevices.length > 1 ? 'multiple devices' : <DeviceIdentityDisplay device={gatewayDevices[0]} isEditable={false} hasAdornment={false} />}
        </div>
      }
    >
      <Chip className={classes.gatewayChip} icon={<GatewayConnectionIcon className="connected" />} label="Connected to gateway" onClick={onGatewayClick} />
    </MenderTooltipClickable>
  );
};

const GatewayNotification = ({ device, onClick }) => {
  const ipAddress = getDemoDeviceAddress([device]);
  const { classes } = useStyles();
  return (
    <MenderTooltipClickable
      placement="bottom"
      title={
        <div style={{ maxWidth: 350 }}>
          For information about connecting other devices to this gateway, please refer to the{' '}
          <DocsLink path="get-started/mender-gateway" title="Mender Gateway documentation" />. This device is reachable via <i>{ipAddress}</i>.
        </div>
      }
    >
      <Chip className={classes.gatewayChip} icon={<GatewayIcon />} label="Gateway" onClick={onClick} />
    </MenderTooltipClickable>
  );
};

const deviceStatusCheck = ({ device: { status = DEVICE_STATES.accepted } }, states = [DEVICE_STATES.accepted]) => states.includes(status);

const tabs = [
  { component: IdentityTab, title: () => 'Identity', value: 'identity', isApplicable: yes },
  {
    component: DeviceInventory,
    title: () => 'Inventory',
    value: 'inventory',
    isApplicable: deviceStatusCheck
  },
  {
    component: InstalledSoftware,
    title: () => 'Software',
    value: 'software',
    isApplicable: deviceStatusCheck
  },
  {
    component: Deployments,
    title: () => 'Deployments',
    value: 'deployments',
    isApplicable: deviceStatusCheck
  },
  {
    component: DeviceConfiguration,
    title: () => 'Configuration',
    value: 'configuration',
    isApplicable: ({ userCapabilities: { canConfigure }, ...rest }) => canConfigure && deviceStatusCheck(rest, [DEVICE_STATES.accepted, DEVICE_STATES.preauth])
  },
  {
    component: MonitoringTab,
    title: () => 'Monitoring',
    value: 'monitor',
    isApplicable: deviceStatusCheck
  },
  {
    component: TroubleshootTab,
    title: () => 'Troubleshooting',
    value: 'troubleshoot',
    isApplicable: deviceStatusCheck
  },
  {
    component: IntegrationTab,
    title: ({ integrations }) => {
      if (integrations.length > 1) {
        return 'Device Twin';
      }
      const { title, twinTitle } = EXTERNAL_PROVIDER[integrations[0].provider];
      return `${title} ${twinTitle}`;
    },
    value: 'device-twin',
    isApplicable: ({ integrations, ...rest }) => !!integrations.length && deviceStatusCheck(rest, [DEVICE_STATES.accepted, DEVICE_STATES.preauth])
  },
  {
    component: DeviceSystem,
    title: () => 'System',
    value: 'system',
    isApplicable: ({ device: { attributes = {} } }) => stringToBoolean(attributes?.mender_is_gateway ?? '')
  }
];

export const ExpandedDevice = ({ actionCallbacks, deviceId, onClose, setDetailsTab, tabSelection }) => {
  const timer = useRef();
  const navigate = useNavigate();
  const { classes } = useStyles();

  const { latest: latestAlerts = [] } = useSelector(state => state.monitor.alerts.byDeviceId[deviceId]) || {};
  const { selectedGroup, groupFilters = [] } = useSelector(getSelectedGroupInfo);
  const { columnSelection = [] } = useSelector(getUserSettings);
  const { defaultDeviceConfig: defaultConfig } = useSelector(getGlobalSettings);
  const { device, deviceConfigDeployment } = useSelector(state => getDeviceConfigDeployment(state, deviceId));
  const devicesById = useSelector(getDevicesById);
  const docsVersion = useSelector(getDocsVersion);
  const integrations = useSelector(getDeviceTwinIntegrations);
  const tenantCapabilities = useSelector(getTenantCapabilities);
  const userCapabilities = useSelector(getUserCapabilities);
  const dispatch = useDispatch();

  const { attributes = {}, isOffline, gatewayIds = [] } = device;
  const { mender_is_gateway, mender_gateway_system_id } = attributes;
  const isGateway = stringToBoolean(mender_is_gateway);

  useEffect(() => {
    clearInterval(timer.current);
    if (!deviceId) {
      return;
    }
    timer.current = setInterval(() => dispatch(getDeviceInfo(deviceId)), refreshDeviceLength);
    dispatch(getDeviceInfo(deviceId));
    return () => {
      clearInterval(timer.current);
    };
  }, [deviceId, device.status, dispatch]);

  useEffect(() => {
    if (!(device.id && mender_gateway_system_id)) {
      return;
    }
    dispatch(getGatewayDevices(device.id));
  }, [device.id, dispatch, mender_gateway_system_id]);

  // close expanded device
  const onDecommissionDevice = deviceId => dispatch(decommissionDevice({ deviceId })).finally(onClose);

  const copyLinkToClipboard = () => {
    const location = window.location.href.substring(0, window.location.href.indexOf('/devices') + '/devices'.length);
    copy(`${location}?id=${deviceId}`);
    setSnackbar('Link copied to clipboard');
  };

  const scrollToMonitor = () => setDetailsTab('monitor');

  const selectedStaticGroup = selectedGroup && !groupFilters.length ? selectedGroup : undefined;

  const scrollToDeviceSystem = target => {
    if (target) {
      return navigate(`/devices?${target}`);
    }
    return setDetailsTab('device-system');
  };

  const onCloseClick = useCallback(() => {
    if (deviceId) {
      onClose();
    }
  }, [deviceId, onClose]);

  const availableTabs = tabs.reduce((accu, tab) => {
    if (tab.isApplicable({ device, integrations, tenantCapabilities, userCapabilities })) {
      accu.push(tab);
    }
    return accu;
  }, []);

  const { component: SelectedTab, value: selectedTab } = availableTabs.find(tab => tab.value === tabSelection) ?? tabs[0];

  const dispatchedSetSnackbar = useCallback((...args) => dispatch(setSnackbar(...args)), [dispatch]);
  const dispatchedSaveGlobalSettings = useCallback(settings => dispatch(saveGlobalSettings(settings)), [dispatch]);

  const commonProps = {
    classes,
    columnSelection,
    defaultConfig,
    device,
    deviceConfigDeployment,
    docsVersion,
    integrations,
    latestAlerts,
    onDecommissionDevice,
    saveGlobalSettings: dispatchedSaveGlobalSettings,
    setDetailsTab,
    setSnackbar: dispatchedSetSnackbar,
    tenantCapabilities,
    userCapabilities
  };
  return (
    <Drawer anchor="right" className="expandedDevice" open={!!deviceId} onClose={onCloseClick} PaperProps={{ style: { minWidth: '67vw' } }}>
      <DrawerTitle
        title={<>Device information for {<DeviceIdentityDisplay device={device} isEditable={false} hasAdornment={false} style={{ marginLeft: 4 }} />}</>}
        onLinkCopy={copyLinkToClipboard}
        preCloser={
          <>
            {isGateway && <GatewayNotification device={device} onClick={() => scrollToDeviceSystem()} />}
            {!!gatewayIds.length && (
              <GatewayConnectionNotification gatewayDevices={gatewayIds.map(gatewayId => devicesById[gatewayId])} onClick={scrollToDeviceSystem} />
            )}
            <div className={`${isOffline ? 'red' : 'muted'} margin-left margin-right flexbox`}>
              <Tooltip title="The last time the device communicated with the Mender server" placement="bottom">
                <div className="margin-right-small">Last check-in:</div>
              </Tooltip>
              <RelativeTime updateTime={device.check_in_time_exact ?? device.check_in_time} />
            </div>
          </>
        }
        onClose={onCloseClick}
      />
      <DeviceNotifications alerts={latestAlerts} device={device} onClick={scrollToMonitor} />
      <Divider className={classes.dividerTop} />
      <Tabs value={selectedTab} onChange={(e, tab) => setDetailsTab(tab)} textColor="primary">
        {availableTabs.map(item => (
          <Tab key={item.value} label={item.title({ integrations })} value={item.value} />
        ))}
      </Tabs>
      <SelectedTab {...commonProps} />
      <DeviceQuickActions actionCallbacks={actionCallbacks} deviceId={device.id} selectedGroup={selectedStaticGroup} />
    </Drawer>
  );
};

export default ExpandedDevice;
