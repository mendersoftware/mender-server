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

import { ArrowForward as ArrowForwardIcon } from '@mui/icons-material';
import { Button, Chip, DialogActions, DialogContent, List, ListItem, Typography, lighten } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import DocsLink from '@northern.tech/common-ui/DocsLink';
import Loader from '@northern.tech/common-ui/Loader';
import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';
import { DEVICE_STATES, TIMEOUTS, onboardingSteps } from '@northern.tech/store/constants';
import { getDeviceCountsByStatus, getFeatures, getOnboardingState, getTenantCapabilities } from '@northern.tech/store/selectors';
import { advanceOnboarding, saveUserSettings, setDeviceListState } from '@northern.tech/store/thunks';

import raspberryPi from '../../../../assets/img/raspberrypi.png';
import zephyr from '../../../../assets/img/zephyr_logo.png';
import { HELPTOOLTIPS } from '../../helptips/HelpTooltips';
import { MenderHelpTooltip } from '../../helptips/MenderTooltip';
import { McuDeviceOnboarding } from './McuDeviceOnboarding';
import PhysicalDeviceOnboarding from './PhysicalDeviceOnboarding';
import VirtualDeviceOnboarding from './VirtualDeviceOnboarding';

const useStyles = makeStyles()(theme => ({
  rpiQuickstart: {
    backgroundColor: lighten(theme.palette.background.paper, 0.25),
    'img': { height: 30, marginRight: theme.spacing(1.5), marginLeft: theme.spacing(0.75) }
  },
  zephyrLogo: { height: '24px', marginRight: theme.spacing(2) },
  deviceSection: {
    border: `1px solid ${theme.palette.divider}`,
    borderRadius: theme.spacing(0.5),
    gap: theme.spacing(1)
  },
  bottomText: {
    marginTop: theme.spacing(3)
  }
}));

const docsLinks = [
  { key: 'debian', target: 'operating-system-updates-debian-family', title: 'Debian family' },
  { key: 'yocto', target: 'operating-system-updates-yocto-project', title: 'Yocto OSes' }
];

const MenderHubReference = () => (
  <Typography variant="body1">
    Or visit {/* eslint-disable-next-line react/jsx-no-target-blank */}
    <a href="https://hub.mender.io/c/board-integrations" target="_blank" rel="noopener">
      Mender Hub
    </a>{' '}
    and search integrations for your device and OS.
  </Typography>
);

const OnPremDeviceConnectionExplainer = ({ isEnterprise }) => (
  <>
    <Typography variant="body1">
      You can connect almost any device and Linux OS with Mender, but to make things simple during evaluation we recommend you to get started with a Debian
      based setup. This also works with a Raspberry Pi as a test device.
      <br />
      Follow the <DocsLink path="client-installation/install-with-debian-package" title="installation instructions" /> for Debian packages and select the{' '}
      {isEnterprise ? 'Enterprise' : 'Demo'} server tab to configure the client.
      <br />
      For operating system updates, see the documentation to integrate the following with Mender:
    </Typography>
    <List>
      {docsLinks.map(item => (
        <ListItem key={item.key} disablePadding className="padding-top-none padding-bottom-none">
          <DocsLink path={item.target} title={item.title} />
        </ListItem>
      ))}
    </List>
    <MenderHubReference />
  </>
);
const ZephyrMCUGuide = ({ setMcu }) => {
  const { classes } = useStyles();
  return (
    <div className={`padding-small padding-bottom-none flexbox column ${classes.deviceSection}`}>
      <div className="flexbox space-between">
        <div className="flexbox centered">
          <img src={zephyr} className={classes.zephyrLogo} />
          <Typography variant="subtitle1" gutterBottom>
            Zephyr MCU
          </Typography>
        </div>
        <Chip size="small" label="Micro" />
      </div>
      <Typography variant="body1">Connect an Espressif ESP32-S3 DevKitC, or any compatible microcontroller that supports MCUBoot in Zephyr.</Typography>
      <div>
        <Button variant="text" size="small" endIcon={<ArrowForwardIcon />} onClick={() => setMcu(true)}>
          Get started with Zephyr
        </Button>
      </div>
    </div>
  );
};

const OtherDevicesGuide = () => {
  const { classes } = useStyles();
  return (
    <div className={`padding-small flexbox column ${classes.deviceSection}`}>
      <Typography variant="subtitle1" gutterBottom>
        Other devices
      </Typography>
      <Typography variant="body1">See the documentation to integrate the following with Mender:</Typography>
      <List>
        {docsLinks.map(item => (
          <ListItem key={item.key} disablePadding className="padding-top-none padding-bottom-none">
            <DocsLink path={item.target} title={item.title} />
          </ListItem>
        ))}
      </List>
      <MenderHubReference />
    </div>
  );
};

const DeviceConnectionExplainer = ({ setOnDevice, setVirtualDevice, setMcu }) => {
  const { classes } = useStyles();
  const { hasMCUEnabled } = useSelector(getFeatures);
  return (
    <>
      <Typography variant="body1">
        You can connect almost any device from Linux to RTOSes. For simplicity, we recommend you use a Raspberry Pi as a test device.
      </Typography>
      <div className={`margin-top-small padding-small rpi-quickstart ${classes.rpiQuickstart} ${classes.deviceSection}`}>
        <div className="flexbox margin-bottom-small space-between">
          <div className="flexbox centered">
            <img src={raspberryPi} alt="rpi-logo" />
            <Typography variant="subtitle1">Raspberry Pi quick start</Typography>
          </div>
          <Chip size="small" label="Standard" />
        </div>
        <Typography variant="body1">
          A step-by-step guide for new users — connect your Raspberry Pi and get started with your first update using Mender.
        </Typography>
        <div className="flexbox margin-top-small">
          <Button variant="text" size="small" onClick={() => setOnDevice(true)} endIcon={<ArrowForwardIcon />}>
            Get started
          </Button>
        </div>
      </div>
      <div className="two-columns margin-top-small">
        {hasMCUEnabled ? <ZephyrMCUGuide setMcu={setMcu} /> : <OtherDevicesGuide />}
        <div className={`padding-small ${classes.deviceSection}`}>
          <Typography variant="subtitle1" gutterBottom>
            Don&#39;t have a device?
          </Typography>
          <Typography variant="body1">
            You can use our virtual device to explore the UI, deploy an update and get a quick feel for the features of Mender.
          </Typography>
          <div className="flexbox margin-top-small">
            <Button variant="text" size="small" onClick={() => setVirtualDevice(true)} endIcon={<ArrowForwardIcon />}>
              Try the virtual device
            </Button>
          </div>
        </div>
      </div>
      <Typography variant="body1" className={classes.bottomText}>
        <DocsLink path="overview/device-support" title="Visit our documentation" /> for full information about device support including Debian family and Yocto
        OSes.
      </Typography>
    </>
  );
};

export const DeviceConnectionDialog = ({ onCancel }) => {
  const [onDevice, setOnDevice] = useState(false);
  const [progress, setProgress] = useState(1);
  const [virtualDevice, setVirtualDevice] = useState(false);
  const [mcu, setMcu] = useState(false);
  const { pending: pendingCount } = useSelector(getDeviceCountsByStatus);
  const [pendingDevicesCount] = useState(pendingCount);
  const [hasMoreDevices, setHasMoreDevices] = useState(false);
  const { isEnterprise } = useSelector(getTenantCapabilities);
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
      setMcu(false);
    }
    setProgress(updatedProgress);
  };

  const onAdvance = () => {
    dispatch(advanceOnboarding(onboardingSteps.DASHBOARD_ONBOARDING_START));
    setProgress(progress + 1);
  };

  let content = <DeviceConnectionExplainer setOnDevice={setOnDevice} setVirtualDevice={setVirtualDevice} setMcu={setMcu} />;
  if (onDevice) {
    content = <PhysicalDeviceOnboarding progress={progress} />;
  } else if (virtualDevice) {
    content = <VirtualDeviceOnboarding />;
  } else if (mcu) {
    content = <McuDeviceOnboarding />;
  } else if (!isHosted) {
    content = <OnPremDeviceConnectionExplainer isEnterprise={isEnterprise} />;
  }

  if (hasMoreDevices && !onboardingComplete) {
    setTimeout(onCancel, TIMEOUTS.twoSeconds);
  }

  const isPhysicalAndNotFinal = progress < 2 && (!virtualDevice || progress < 1);
  return (
    <BaseDialog open title={mcu ? 'Connecting a Zephyr-based MCU' : 'Connecting a device'} maxWidth="sm" onClose={onCancel}>
      <DialogContent>{content}</DialogContent>
      <DialogActions>
        {onDevice || virtualDevice || mcu ? (
          <>
            {isPhysicalAndNotFinal && !mcu && <Button onClick={onCancel}>Cancel</Button>}
            <Button onClick={onBackClick} variant="outlined">
              Back
            </Button>
            {isPhysicalAndNotFinal && !mcu ? (
              <Button variant="contained" disabled={!(virtualDevice || (onDevice && onboardingDeviceType))} onClick={onAdvance}>
                Next
              </Button>
            ) : (
              <Button
                variant="contained"
                disabled={!onboardingComplete && !mcu}
                onClick={onCancel}
                endIcon={!onboardingComplete && !mcu && <Loader show small table style={{ top: -24 }} />}
              >
                {onboardingComplete || mcu ? 'Close' : 'Waiting for device'}
              </Button>
            )}
          </>
        ) : (
          <>
            <MenderHelpTooltip id={HELPTOOLTIPS.deviceSupportTip.id} style={{ marginLeft: 20 }} />
            <div style={{ flexGrow: 1 }} />
            <Button onClick={onCancel}>Cancel</Button>
          </>
        )}
      </DialogActions>
    </BaseDialog>
  );
};

export default DeviceConnectionDialog;
