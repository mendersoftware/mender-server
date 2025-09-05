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
import { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { Typography } from '@mui/material';

import CopyCode from '@northern.tech/common-ui/CopyCode';
import DocsLink from '@northern.tech/common-ui/DocsLink';
import { getFeatures, getOrganization } from '@northern.tech/store/selectors';
import { setOnboardingApproach } from '@northern.tech/store/thunks';

export const getDemoDeviceCreationCommand = (tenantToken, demoArtifactPort) =>
  tenantToken
    ? `TENANT_TOKEN='${tenantToken}'\ndocker run -it -p ${demoArtifactPort}:${demoArtifactPort} -e SERVER_URL='https://${window.location.hostname}' \\\n-e TENANT_TOKEN=$TENANT_TOKEN --pull=always mendersoftware/mender-client-docker-addons`
    : './demo --client up';

export const VirtualDeviceOnboarding = () => {
  const dispatch = useDispatch();
  const { isHosted } = useSelector(getFeatures);
  const { tenant_token: tenantToken } = useSelector(getOrganization);
  const demoArtifactPort = useSelector(state => state.onboarding.demoArtifactPort);

  useEffect(() => {
    dispatch(setOnboardingApproach('virtual'));
  }, [dispatch]);

  const codeToCopy = getDemoDeviceCreationCommand(tenantToken, demoArtifactPort);

  return (
    <div>
      {isHosted ? (
        <>
          <Typography variant="subtitle1">1. Get Docker Engine</Typography>
          <Typography variant="body1">
            If you do not have it already, please install Docker on your local machine. For example if you are using Ubuntu follow this tutorial:{' '}
            <a href="https://docs.docker.com/engine/installation/linux/docker-ce/ubuntu/" target="_blank" rel="noopener noreferrer">
              https://docs.docker.com/engine/installation/linux/docker-ce/ubuntu/
            </a>
          </Typography>
        </>
      ) : (
        <>
          <Typography variant="subtitle1">1. Prerequisites</Typography>
          <Typography variant="body1">
            As you are running Mender on-premise, for these instructions we assume that you already have Docker installed and the Mender integration environment
            up and running on your machine.
            <br />
            To start a virtual device, change directory into the folder where you cloned Mender integration.
          </Typography>
        </>
      )}
      <Typography variant="subtitle1">2. Copy & paste and run the following command to start the virtual device:</Typography>
      <CopyCode code={codeToCopy} withDescription={true} />
      <Typography variant="body1">
        The device should appear in the Pending devices view in a couple of minutes. Visit{' '}
        <DocsLink path="get-started/preparation/prepare-a-virtual-device" title="our documentation" /> for more information on managing the virtual device.
      </Typography>
    </div>
  );
};

export default VirtualDeviceOnboarding;
