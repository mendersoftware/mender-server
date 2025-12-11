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
import { useSelector } from 'react-redux';

import { Schedule as HelpIcon } from '@mui/icons-material';
import { Button } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import storeActions from '@northern.tech/store/actions';
import { ALL_DEVICES, onboardingSteps } from '@northern.tech/store/constants';
import { getOnboardingState } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { advanceOnboarding, setOnboardingComplete } from '@northern.tech/store/thunks';
import DocsLink from '@northern.tech/common-ui/DocsLink';

import BaseOnboardingTip, { BaseOnboardingTooltip } from './BaseOnoardingTip';

const { setShowConnectingDialog, setShowDismissOnboardingTipsDialog } = storeActions;

export const DevicePendingTip = props => (
  <BaseOnboardingTip
    icon={<HelpIcon />}
    component={<div>If you followed the steps in &quot;Connecting a device&quot;, your device will show here shortly.</div>}
    {...props}
  />
);

export const GetStartedTip = props => {
  const dispatch = useAppDispatch();
  return (
    <BaseOnboardingTooltip {...props}>
      <div className="margin-top" style={{ marginBottom: -12 }}>
        <p>
          <b>Welcome to Mender!</b>
        </p>
        We can help you get started with connecting your first device and deploying an update to it.
        <div className="flexbox center-aligned margin-top-small space-between">
          <b className="clickable slightly-smaller" onClick={() => dispatch(setShowDismissOnboardingTipsDialog(true))}>
            No thanks, I don&apos;t need help
          </b>
          <Button onClick={() => dispatch(setShowConnectingDialog(true))}>Get started</Button>
        </div>
      </div>
    </BaseOnboardingTooltip>
  );
};

export const DevicesPendingDelayed = () => (
  <div>If your device still isn&apos;t showing, try following the connection steps again or see our documentation for more.</div>
);

export const DashboardOnboardingState = () => <div>Your device has requested to join the server. Click the row to open the device details.</div>;

export const DevicesPendingAcceptingOnboarding = () => (
  <div>
    Verify your device&#39;s details, like MAC address and public key, then click <b>Accept</b> to allow it to connect to the Mender Server.
  </div>
);

export const DashboardOnboardingPendings = () => <div>Next accept your device</div>;
const useStyles = makeStyles()(theme => ({
  link: {
    color: theme.palette.grey[100],
    '&:hover': {
      color: theme.palette.grey[100]
    }
  },
  buttonContainer: {
    justifyContent: 'flex-end'
  }
}));
export const DevicesAcceptedOnboarding = props => {
  const dispatch = useAppDispatch();
  const { approach } = useSelector(getOnboardingState);
  const isMcu = approach === 'mcu';
  const { classes } = useStyles();
  return (
    <BaseOnboardingTooltip {...props}>
      {isMcu ? (
        <div>
          <p>Your device is now authenticated and has connected to the server! It&apos;s ready to receive updates, report its data and more.</p>
          <p>
            If you would like to learn how to deploy your first update, follow the steps in the documentation and{' '}
            <DocsLink path="get-started/microcontroller-preview/deploy-a-firmware-update" className={`bold ${classes.link}`} >
              deploy a firmware update for Zephyr.
            </DocsLink>
          </p>
          <div className={`flexbox ${classes.buttonContainer}`}>
            <Button variant="contained" onClick={() => dispatch(setOnboardingComplete(true))}>
              End tour
            </Button>
          </div>
        </div>
      ) : (
        <>
          <div>
            <p>Your device is now authenticated and has connected to the server! It&apos;s ready to receive updates, report its data and more.</p>
            Would you like to learn how to deploy your first update?
          </div>
          <div className="flexbox center-aligned margin-top-small space-between">
            <b className="clickable slightly-smaller" onClick={() => dispatch(setShowDismissOnboardingTipsDialog(true))}>
              Dismiss the tutorial
            </b>
            <Button onClick={() => dispatch(advanceOnboarding(onboardingSteps.DEVICES_ACCEPTED_ONBOARDING))}>Yes, let&apos;s deploy!</Button>
          </div>
        </>
      )}
    </BaseOnboardingTooltip>
  );
};

export const DevicesDeployReleaseOnboarding = () => <div>Select &apos;Device actions&apos; to view the list of actions for your device.</div>;

export const DevicesDeployReleaseOnboardingStep2 = () => (
  <div>Choose &apos;Create deployment for this device&apos; to deploy an update using the demo software provided.</div>
);

export const SchedulingArtifactSelection = ({ selectedRelease }) => <div>{`Select the ${selectedRelease.name} release we included.`}</div>;

export const SchedulingAllDevicesSelection = () => (
  <div>
    Select &apos;All devices&apos; for now.<p>You can learn how to create device groups later.</p>
  </div>
);

export const SchedulingGroupSelection = ({ createdGroup }) => <div>{`Select the ${createdGroup} device group you just made.`}</div>;

export const SchedulingReleaseToDevices = ({ selectedDevice, selectedGroup, selectedRelease }) => (
  <div>{`Create the deployment! This will deploy the ${selectedRelease.name} Artifact to ${
    selectedDevice ? selectedDevice.id : selectedGroup || ALL_DEVICES
  }`}</div>
);

export const DeploymentsInprogress = () => <div>Your deployment is in progress. Click to view a report</div>;

export const DeploymentUploadFinished = () => <div>Your deployment has finished. Click to close the panel</div>;

export const DeploymentsPast = () => <div>Your deployment has finished, click here to view it</div>;

export const DeploymentsPastCompletedFailure = () => (
  <div>Your deployment has finished, but it looks like there was a problem. Click to view the deployment report, where you can see the error log.</div>
);
