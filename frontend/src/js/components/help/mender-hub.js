// Copyright 2022 Northern.tech AS
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

import { Launch as LaunchIcon } from '@mui/icons-material';

import hubLogo from '../../../assets/img/mender-hub-logo.png';

const MenderHub = () => (
  <div>
    <h3>Find technical help at Mender Hub</h3>
    {/* eslint-disable-next-line react/jsx-no-target-blank */}
    <a href="https://hub.mender.io" target="_blank" rel="noopener">
      <span>
        <img style={{ 'width': '50%' }} src={hubLogo} />
      </span>
    </a>
    <p>
      {/* eslint-disable-next-line react/jsx-no-target-blank */}
      <a href="https://hub.mender.io" target="_blank" rel="noopener">
        Mender Hub
      </a>{' '}
      is a forum where you can ask technical questions and get help from our active and knowledgeable community.
    </p>
    <h4>Board integrations</h4>
    <p>Get help and instructions for integrating Mender with different boards and operating systems.</p>
    <h4>Community-supported resources</h4>
    <p>Find out-of-the-box and community-supported Update Modules and Configuration Scripts, and contribute your own.</p>
    <h4>Tutorials</h4>
    <p>Tutorials and how-to guides for various Mender use-cases including Yocto Project features and practices.</p>
    <h4>Feature requests</h4>
    <p>Submit your feature requests and feedback. Mender is an open source project and we encourage community contributions.</p>
    <p>
      {/* eslint-disable-next-line react/jsx-no-target-blank */}
      <a href="https://hub.mender.io" target="_blank" rel="noopener">
        Visit Mender Hub <LaunchIcon style={{ 'verticalAlign': 'text-bottom' }} fontSize="small" />
      </a>
    </p>
  </div>
);

export default MenderHub;
