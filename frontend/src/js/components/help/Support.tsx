// Copyright 2018 Northern.tech AS
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

const Support = () => (
  <div>
    <h2>Contact support</h2>
    <p>Do you have a paid plan and need to contact support?</p>
    <p>
      Please reach out through our{' '}
      <a href="https://support.northern.tech" target="_blank" rel="noopener noreferrer">
        support portal
      </a>
      , including your organization name and how we can help you.
    </p>
    <h4>Reporting bugs</h4>
    <p>To help us best respond to bug reports, please include:</p>
    <ul>
      <li>Which software versions you are using</li>
      <li>If the issue can be reproduced, what are the steps to reproduce it?</li>
      <li>In case of a UI issue, any screenshots or screen recordings</li>
    </ul>
    <h3>Trial or open source users</h3>
    <div>
      <p>Find technical help at</p>
      {/* eslint-disable-next-line react/jsx-no-target-blank */}
      <a href="https://hub.mender.io" target="_blank" rel="noopener">
        <span>
          <img style={{ 'width': '50%' }} src={hubLogo} />
        </span>
      </a>
      <p>Mender Hub is a forum where you can find help, ask technical questions and start discussions with our active and knowledgeable community.</p>
      {/* eslint-disable-next-line react/jsx-no-target-blank */}
      <a href="https://hub.mender.io" target="_blank" rel="noopener">
        <span>Ask a question at Mender Hub</span> <LaunchIcon style={{ 'verticalAlign': 'text-bottom' }} fontSize="small" />
      </a>
    </div>
  </div>
);
export default Support;
