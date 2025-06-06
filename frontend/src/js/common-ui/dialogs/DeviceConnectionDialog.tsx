// Copyright 2019 Northern.tech AS
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
import { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';

import { Button, DialogActions, DialogContent } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { DEVICE_STATES, TIMEOUTS, onboardingSteps } from '@northern.tech/store/constants';
import { getDeviceCountsByStatus, getFeatures, getOnboardingState, getTenantCapabilities } from '@northern.tech/store/selectors';
import { advanceOnboarding, saveUserSettings, setDeviceListState } from '@northern.tech/store/thunks';

import docker from '../../../assets/img/docker.png';
import raspberryPi4 from '../../../assets/img/raspberrypi4.png';
import raspberryPi from '../../../assets/img/raspberrypi.png';
import DocsLink from '../DocsLink';
import InfoText from '../InfoText';
import Loader from '../Loader';
import { HELPTOOLTIPS } from '../helptips/HelpTooltips';
import { MenderHelpTooltip } from '../helptips/MenderTooltip';
import { BaseDialog } from './BaseDialog';
import PhysicalDeviceOnboarding from './PhysicalDeviceOnboarding';
import VirtualDeviceOnboarding from './VirtualDeviceOnboarding';

const useStyles = makeStyles()(theme => ({
  rpiQuickstart: {
    backgroundColor: theme.palette.background.lightgrey
  },
  virtualLogo: { height: 40, marginLeft: theme.spacing(2) }
}));

const docsLinks = [
  { key: 'debian', target: 'operating-system-updates-debian-family', title: 'Debian family' },
  { key: 'yocto', target: 'operating-system-updates-yocto-project', title: 'Yocto OSes' }
];

const MenderHubReference = () => (
  <p className="padding-bottom-none">
    Or visit {/* eslint-disable-next-line react/jsx-no-target-blank */}
    <a href="https://hub.mender.io/c/board-integrations" target="_blank" rel="noopener">
      Mender Hub
    </a>{' '}
    and search integrations for your device and OS.
  </p>
);

const OnPremDeviceConnectionExplainer = ({ isEnterprise }) => (
  <>
    <p>
      You can connect almost any device and Linux OS with Mender, but to make things simple during evaluation we recommend you to get started with a Debian
      based setup. This also works with a Raspberry Pi as a test device.
      <br />
      Follow the <DocsLink path="client-installation/install-with-debian-package" title="installation instructions" /> for Debian packages and select the{' '}
      {isEnterprise ? 'Enterprise' : 'Demo'} server tab to configure the client.
    </p>
    <div>For operating system updates, see the documentation to integrate the following with Mender:</div>
    <ul>
      {docsLinks.map(item => (
        <li key={item.key}>
          <DocsLink path={item.target} title={item.title} />
        </li>
      ))}
    </ul>
    <MenderHubReference />
  </>
);

const DeviceConnectionExplainer = ({ hasMonitor, setOnDevice, setVirtualDevice }) => {
  const { classes } = useStyles();
  return (
    <>
      <p>
        You can connect almost any device and Linux OS with Mender, but to make things simple during evaluation we recommend you use a Raspberry Pi as a test
        device.
      </p>
      <div className={`padding-small padding-top-none rpi-quickstart ${classes.rpiQuickstart}`}>
        <h3>Raspberry Pi quick start</h3>
        <p>We&apos;ll walk you through the steps to connect a Raspberry Pi and deploy your first update with Mender.</p>
        <div className="flexbox column centered">
          <div className="flexbox centered os-list">
            {[raspberryPi, raspberryPi4].map((tile, index) => (
              <img key={`tile-${index}`} src={tile} />
            ))}
          </div>
          <Button variant="contained" color="secondary" onClick={() => setOnDevice(true)}>
            Get Started
          </Button>
        </div>
      </div>
      <div className="two-columns margin-top">
        <div className="padding-small padding-top-none">
          <div className="flexbox center-aligned">
            <h3>Use a virtual device</h3>
            <img src={docker} className={classes.virtualLogo} />
          </div>
          <p className="margin-top-none">Don&apos;t have a Raspberry Pi?</p>
          <p>You can use our Docker-run virtual device to go through the same tutorial.</p>
          {hasMonitor && (
            <InfoText className="slightly-smaller">
              If you want to evaluate our commercial components such as mender-monitor, please use a physical device instead as the virtual client does not
              support these components at this time.
            </InfoText>
          )}
          <a onClick={() => setVirtualDevice(true)}>Try a virtual device</a>
        </div>
        <div className="padding-small padding-top-none">
          <h3>Other devices</h3>
          <div>See the documentation to integrate the following with Mender:</div>
          <ul>
            {docsLinks.map(item => (
              <li key={item.key}>
                <DocsLink path={item.target} title={item.title} />
              </li>
            ))}
          </ul>
          <MenderHubReference />
        </div>
      </div>
      <MenderHelpTooltip id={HELPTOOLTIPS.deviceSupportTip.id} style={{ position: 'absolute', bottom: '2.5%', left: '88%' }} />
    </>
  );
};

export const DeviceConnectionDialog = ({ onCancel }) => {
  const [onDevice, setOnDevice] = useState(false);
  const [progress, setProgress] = useState(1);
  const [virtualDevice, setVirtualDevice] = useState(false);
  const { pending: pendingCount } = useSelector(getDeviceCountsByStatus);
  const [pendingDevicesCount] = useState(pendingCount);
  const [hasMoreDevices, setHasMoreDevices] = useState(false);
  const { isEnterprise, hasMonitor } = useSelector(getTenantCapabilities);
  const { isHosted } = useSelector(getFeatures);
  const { complete: onboardingComplete, deviceType: onboardingDeviceType } = useSelector(getOnboardingState);
  const dispatch = useDispatch();
  const navigate = useNavigate();

  useEffect(() => {
    setHasMoreDevices(pendingCount > pendingDevicesCount);
  }, [pendingDevicesCount, pendingCount]);

  useEffect(() => {
    if ((virtualDevice || progress >= 2) && hasMoreDevices && !window.location.hash.includes('pending')) {
      dispatch(advanceOnboarding(onboardingSteps.DASHBOARD_ONBOARDING_START));
      dispatch(setDeviceListState({ state: DEVICE_STATES.pending }));
      navigate('/devices/pending');
    }
    if (virtualDevice || progress >= 2) {
      dispatch(saveUserSettings({ onboarding: { deviceConnection: new Date().toISOString() } }));
    }
  }, [dispatch, hasMoreDevices, navigate, progress, virtualDevice]);

  const onBackClick = () => {
    let updatedProgress = progress - 1;
    if (!updatedProgress) {
      updatedProgress = 1;
      setOnDevice(false);
      setVirtualDevice(false);
    }
    setProgress(updatedProgress);
  };

  const onAdvance = () => {
    dispatch(advanceOnboarding(onboardingSteps.DASHBOARD_ONBOARDING_START));
    setProgress(progress + 1);
  };

  let content = <DeviceConnectionExplainer hasMonitor={hasMonitor} setOnDevice={setOnDevice} setVirtualDevice={setVirtualDevice} />;
  if (onDevice) {
    content = <PhysicalDeviceOnboarding progress={progress} />;
  } else if (virtualDevice) {
    content = <VirtualDeviceOnboarding />;
  } else if (!isHosted) {
    content = <OnPremDeviceConnectionExplainer isEnterprise={isEnterprise} />;
  }

  if (hasMoreDevices && !onboardingComplete) {
    setTimeout(onCancel, TIMEOUTS.twoSeconds);
  }

  return (
    <BaseDialog open title="Connecting a device" maxWidth="sm" onClose={onCancel}>
      <DialogContent className="onboard-dialog padding-bottom-none margin-left margin-right">{content}</DialogContent>
      <DialogActions>
        <Button onClick={onCancel}>Cancel</Button>
        <div style={{ flexGrow: 1 }} />
        {(onDevice || virtualDevice) && (
          <div>
            <Button onClick={onBackClick}>Back</Button>
            {progress < 2 && (!virtualDevice || progress < 1) ? (
              <Button variant="contained" disabled={!(virtualDevice || (onDevice && onboardingDeviceType))} onClick={onAdvance}>
                Next
              </Button>
            ) : (
              <Button
                variant="contained"
                disabled={!onboardingComplete}
                onClick={onCancel}
                endIcon={!onboardingComplete && <Loader show small table style={{ top: -24 }} />}
              >
                {onboardingComplete ? 'Close' : 'Waiting for device'}
              </Button>
            )}
          </div>
        )}
      </DialogActions>
    </BaseDialog>
  );
};

export default DeviceConnectionDialog;
