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

import { Button, DialogActions, DialogContent, List, ListItem, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import DocsLink from '@northern.tech/common-ui/DocsLink';
import Loader from '@northern.tech/common-ui/Loader';
import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';
import { DEVICE_STATES, TIMEOUTS, onboardingSteps } from '@northern.tech/store/constants';
import { getDeviceCountsByStatus, getFeatures, getOnboardingState, getTenantCapabilities } from '@northern.tech/store/selectors';
import { advanceOnboarding, saveUserSettings, setDeviceListState } from '@northern.tech/store/thunks';

import docker from '../../../../assets/img/docker.png';
import raspberryPi4 from '../../../../assets/img/raspberrypi4.png';
import raspberryPi from '../../../../assets/img/raspberrypi.png';
import { HELPTOOLTIPS } from '../../helptips/HelpTooltips';
import { MenderHelpTooltip } from '../../helptips/MenderTooltip';
import PhysicalDeviceOnboarding from './PhysicalDeviceOnboarding';
import VirtualDeviceOnboarding from './VirtualDeviceOnboarding';

const useStyles = makeStyles()(theme => ({
  rpiQuickstart: {
    backgroundColor: theme.palette.background.lightgrey ? theme.palette.background.lightgrey : theme.palette.grey[100],
    '.os-list img': {
      height: 80,
      margin: theme.spacing(2)
    }
  },
  virtualLogo: { height: 40, marginLeft: theme.spacing(2) },
  deviceSection: {
    border: `1px solid ${theme.palette.divider}`,
    borderRadius: theme.spacing(0.5),
    gap: theme.spacing(1)
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

const DeviceConnectionExplainer = ({ setOnDevice, setVirtualDevice }) => {
  const { classes } = useStyles();
  return (
    <>
      <Typography variant="body1">
        You can connect almost any device and Linux OS with Mender, but to make things simple during evaluation we recommend you use a Raspberry Pi as a test
        device.
      </Typography>
      <div className={`margin-top-small padding-small rpi-quickstart ${classes.rpiQuickstart}`}>
        <Typography variant="subtitle1">Raspberry Pi quick start</Typography>
        <Typography variant="body1">We&apos;ll walk you through the steps to connect a Raspberry Pi and deploy your first update with Mender.</Typography>
        <div className="flexbox column centered">
          <div className="flexbox centered os-list">
            {[raspberryPi, raspberryPi4].map((tile, index) => (
              <img key={`tile-${index}`} src={tile} />
            ))}
          </div>
          <Button variant="contained" color="secondary" onClick={() => setOnDevice(true)}>
            Get started
          </Button>
        </div>
      </div>
      <div className="two-columns margin-top-small">
        <div className={`padding-small padding-bottom-none flexbox column ${classes.deviceSection}`}>
          <div className="flexbox center-aligned">
            <Typography variant="subtitle1" gutterBottom>
              Use a virtual device
            </Typography>
            <img src={docker} className={classes.virtualLogo} />
          </div>
          <Typography variant="body1">
            Don&apos;t have a Raspberry Pi?
            <br />
            You can use our Docker-run virtual device to go through the same tutorial.
          </Typography>
          <div>
            <Typography variant="body1" color="text.secondary">
              If you want to evaluate our commercial components such as mender-monitor, please use a physical device instead as the virtual client does not
              support these components at this time.
            </Typography>
            <Button variant="text" size="small" onClick={() => setVirtualDevice(true)}>
              Try a virtual device
            </Button>
          </div>
        </div>
        <div className={`padding-small ${classes.deviceSection}`}>
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
      </div>
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

  let content = <DeviceConnectionExplainer setOnDevice={setOnDevice} setVirtualDevice={setVirtualDevice} />;
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

  const isPhysicalAndNotFinal = progress < 2 && (!virtualDevice || progress < 1);
  return (
    <BaseDialog open title="Connecting a device" maxWidth="sm" onClose={onCancel}>
      <DialogContent>{content}</DialogContent>
      <DialogActions>
        {onDevice || virtualDevice ? (
          <>
            {isPhysicalAndNotFinal && <Button onClick={onCancel}>Cancel</Button>}
            <Button onClick={onBackClick} variant="outlined">
              Back
            </Button>
            {isPhysicalAndNotFinal ? (
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
