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

import { Close as CloseIcon, Link as LinkIcon } from '@mui/icons-material';
import { Chip, Divider, Drawer, IconButton, Tab, Tabs, Tooltip, chipClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import copy from 'copy-to-clipboard';

import GatewayConnectionIcon from '../../../assets/img/gateway-connection.svg';
import GatewayIcon from '../../../assets/img/gateway.svg';
import { setSnackbar } from '../../actions/appActions';
import { decommissionDevice, getDeviceInfo, getGatewayDevices } from '../../actions/deviceActions';
import { saveGlobalSettings } from '../../actions/userActions';
import { TIMEOUTS, yes } from '../../constants/appConstants';
import { DEVICE_STATES, EXTERNAL_PROVIDER } from '../../constants/deviceConstants';
import { getDemoDeviceAddress, stringToBoolean } from '../../helpers';
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
} from '../../selectors';
import DeviceIdentityDisplay from '../common/deviceidentity';
import DocsLink from '../common/docslink';
import { MenderTooltipClickable } from '../common/mendertooltip';
import { RelativeTime } from '../common/time';
import DeviceConfiguration from './device-details/configuration';
import TroubleshootTab from './device-details/connection';
import Deployments from './device-details/deployments';
import DeviceInventory from './device-details/deviceinventory';
import DeviceSystem from './device-details/devicesystem';
import { IntegrationTab } from './device-details/devicetwin';
import { IdentityTab } from './device-details/identity';
import InstalledSoftware from './device-details/installedsoftware';
import MonitoringTab from './device-details/monitoring';
import DeviceNotifications from './device-details/notifications';
import DeviceQuickActions from './widgets/devicequickactions';

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
  const onDecommissionDevice = device_id => dispatch(decommissionDevice(device_id)).finally(onClose);

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
      <div className="flexbox center-aligned space-between">
        <div className="flexbox center-aligned">
          <h3 className="flexbox">
            Device information for {<DeviceIdentityDisplay device={device} isEditable={false} hasAdornment={false} style={{ marginLeft: 4 }} />}
          </h3>
          <IconButton onClick={copyLinkToClipboard} size="large">
            <LinkIcon />
          </IconButton>
        </div>
        <div className="flexbox center-aligned">
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
          <IconButton style={{ marginLeft: 'auto' }} onClick={onCloseClick} aria-label="close" size="large">
            <CloseIcon />
          </IconButton>
        </div>
      </div>
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
