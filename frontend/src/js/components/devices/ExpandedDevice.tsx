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
import { useCallback, useEffect, useRef } from 'react';
import { useSelector } from 'react-redux';

import { Tab, Tabs, Typography } from '@mui/material';

import BaseDrawer from '@northern.tech/common-ui/BaseDrawer';
import DeviceIdentityDisplay from '@northern.tech/common-ui/DeviceIdentity';
import { RelativeTime } from '@northern.tech/common-ui/Time';
import storeActions from '@northern.tech/store/actions';
import { DEVICE_STATES, EXTERNAL_PROVIDER, TIMEOUTS, yes } from '@northern.tech/store/constants';
import {
  getDeviceById,
  getDeviceTwinIntegrations,
  getGlobalSettings,
  getSelectedGroupInfo,
  getTenantCapabilities,
  getUserCapabilities
} from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { decommissionDevice, getDeviceInfo } from '@northern.tech/store/thunks';
import copy from 'copy-to-clipboard';

import DeviceConfiguration from './device-details/Configuration';
import TroubleshootTab from './device-details/Connection';
import Deployments from './device-details/Deployments';
import DeviceInventory from './device-details/DeviceInventory';
import { DeviceSystem } from './device-details/DeviceSystem';
import { IntegrationTab } from './device-details/DeviceTwin';
import { IdentityTab } from './device-details/Identity';
import InstalledSoftware from './device-details/InstalledSoftware';
import MonitoringTab from './device-details/Monitoring';
import DeviceNotifications from './device-details/Notifications';
import DeviceQuickActions from './widgets/DeviceQuickActions';

const { setSnackbar } = storeActions;

const refreshDeviceLength = TIMEOUTS.refreshDefault;

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
    component: DeviceSystem,
    title: () => 'System',
    value: 'system',
    isApplicable: args => args.device.tier === 'system' && deviceStatusCheck(args)
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
  }
];

export const ExpandedDevice = ({ actionCallbacks, deviceId, onClose, setDetailsTab, tabSelection }) => {
  const timer = useRef();

  const { latest: latestAlerts = [] } = useSelector(state => state.monitor.alerts.byDeviceId[deviceId]) || {};
  const { selectedGroup, groupFilters = [] } = useSelector(getSelectedGroupInfo);
  const { defaultDeviceConfig: defaultConfig } = useSelector(getGlobalSettings);
  const device = useSelector(state => getDeviceById(state, deviceId));
  const integrations = useSelector(getDeviceTwinIntegrations);
  const tenantCapabilities = useSelector(getTenantCapabilities);
  const userCapabilities = useSelector(getUserCapabilities);
  const dispatch = useAppDispatch();

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

  // close expanded device
  const onDecommissionDevice = deviceId => dispatch(decommissionDevice({ deviceId })).finally(onClose);

  const copyLinkToClipboard = () => {
    const location = window.location.href.substring(0, window.location.href.indexOf('/devices') + '/devices'.length);
    copy(`${location}?id=${deviceId}`);
    setSnackbar('Link copied to clipboard');
  };

  const scrollToMonitor = () => setDetailsTab('monitor');

  const selectedStaticGroup = selectedGroup && !groupFilters.length ? selectedGroup : undefined;

  const onCloseClick = useCallback(() => {
    if (deviceId) {
      onClose();
    }
  }, [deviceId, onClose]);

  const availableTabs = tabs.filter(tab => tab.isApplicable({ device, integrations, tenantCapabilities, userCapabilities }));

  const { component: SelectedTab, value: selectedTab } = availableTabs.find(tab => tab.value === tabSelection) ?? tabs[0];

  const dispatchedSetSnackbar = useCallback((...args) => dispatch(setSnackbar(...args)), [dispatch]);

  const commonProps = {
    defaultConfig,
    device,
    integrations,
    onDecommissionDevice,
    setSnackbar: dispatchedSetSnackbar,
    userCapabilities
  };
  return (
    <BaseDrawer
      className="expandedDevice"
      open={!!deviceId}
      onClose={onCloseClick}
      size="lg"
      slotProps={{
        header: {
          title: <>Device information for {<DeviceIdentityDisplay device={device} isEditable={false} style={{ marginLeft: 4 }} />}</>,
          onLinkCopy: copyLinkToClipboard,
          preCloser: (
            <Typography variant="body2" className="flexbox align-items-center" color="textSecondary">
              Latest activity:
              <RelativeTime className="margin-left-small" updateTime={device.check_in_time} />
            </Typography>
          )
        }
      }}
      notification={<DeviceNotifications alerts={latestAlerts} device={device} onClick={scrollToMonitor} />}
    >
      <Tabs
        value={selectedTab}
        onChange={(e, tab) => setDetailsTab(tab)}
        textColor="primary"
        variant="scrollable"
        scrollButtons="auto"
        allowScrollButtonsMobile
      >
        {availableTabs.map(item => (
          <Tab key={item.value} label={item.title({ integrations })} value={item.value} />
        ))}
      </Tabs>
      <SelectedTab {...commonProps} />
      <DeviceQuickActions actionCallbacks={actionCallbacks} deviceId={device.id} selectedGroup={selectedStaticGroup} />
    </BaseDrawer>
  );
};

export default ExpandedDevice;
