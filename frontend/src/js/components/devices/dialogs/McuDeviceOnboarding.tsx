// Copyright 2025 Northern.tech AS
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
import { useSelector } from 'react-redux';
import { Link } from 'react-router-dom';

import { Typography } from '@mui/material';

import CopyCode from '@northern.tech/common-ui/CopyCode';
import DocsLink from '@northern.tech/common-ui/DocsLink';
import { getOrganization } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { setOnboardingApproach } from '@northern.tech/store/thunks';

export const McuDeviceOnboarding = () => {
  const { tenant_token: tenantToken } = useSelector(getOrganization);
  const dispatch = useAppDispatch();
  useEffect(() => {
    dispatch(setOnboardingApproach('mcu'));
  }, [dispatch]);

  return (
    <div>
      <Typography variant="subtitle1" className="margin-bottom-small">
        Set up a Zephyr application with Mender
      </Typography>
      <Typography className="margin-bottom-small">
        To do this, you’ll need to visit our documentation and follow the guide there. Go to{' '}
        <DocsLink path="get-started/microcontroller-preview/prepare-an-esp32-s3-with-zephyr" title="Prepare an ESP32-S3 with Zephyr" />
      </Typography>
      <Typography className="margin-bottom-small" variant="body1">
        First, copy your <b>organization token</b> below as you will need it to complete the steps in the guide:
      </Typography>
      <CopyCode code={tenantToken} withDescription={true} />
      <Typography variant="body1">
        Once you have completed steps 1 and 2 in the documentation, your device should show in the Mender UI as <i>“Pending”</i>. Go to{' '}
        <Link to="/devices">Devices</Link> to see it.
      </Typography>
    </div>
  );
};
