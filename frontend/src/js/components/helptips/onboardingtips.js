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
import React from 'react';
import { useDispatch } from 'react-redux';

import { Schedule as HelpIcon } from '@mui/icons-material';
import { Button } from '@mui/material';

import storeActions from '@northern.tech/store/actions';
import { ALL_DEVICES, onboardingSteps } from '@northern.tech/store/constants';
import { advanceOnboarding } from '@northern.tech/store/thunks';

import BaseOnboardingTip, { BaseOnboardingTooltip } from './baseonboardingtip';

const { setShowConnectingDialog, setShowDismissOnboardingTipsDialog } = storeActions;

export const DevicePendingTip = props => (
  <BaseOnboardingTip
    icon={<HelpIcon />}
    component={<div>If you followed the steps in &quot;Connecting a device&quot;, your device will show here shortly.</div>}
    {...props}
  />
);

export const GetStartedTip = props => {
  const dispatch = useDispatch();
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

export const DashboardOnboardingState = () => <div>Your device has requested to join the server. Click the row to expand the device details.</div>;

export const DevicesPendingAcceptingOnboarding = () => (
  <div>
    Your device has made a request to join the Mender server. You can inspect its identity details, such as mac address and public key, to verify it is
    definitely your device.
    <br />
    When you are ready, Accept it!
  </div>
);

export const DashboardOnboardingPendings = () => <div>Next accept your device</div>;

export const DevicesAcceptedOnboarding = props => {
  const dispatch = useDispatch();
  return (
    <BaseOnboardingTooltip {...props}>
      <div className="margin-top" style={{ marginBottom: -12 }}>
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
      </div>
    </BaseOnboardingTooltip>
  );
};

export const DevicesDeployReleaseOnboarding = () => (
  <div>
    From the Device actions, choose &apos;Create a deployment for this device&apos;. You&apos;ll deploy an update to the device, using some demo software we
    have provided.
  </div>
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

export const DeploymentUploadFinished = () => <div>Artifact upload is finished. Click to close the panel</div>;

export const DeploymentsPast = () => <div>Your deployment has finished, click here to view it</div>;

export const DeploymentsPastCompletedFailure = () => (
  <div>Your deployment has finished, but it looks like there was a problem. Click to view the deployment report, where you can see the error log.</div>
);
