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

import { Launch as LaunchIcon } from '@mui/icons-material';

import DocsLink from '../common/docslink';

const InlineLaunchIcon = () => <LaunchIcon style={{ 'verticalAlign': 'text-bottom' }} fontSize="small" />;

const GettingStarted = () => (
  <div>
    <h2>Getting started</h2>
    <h3>Get Started guide</h3>
    <p>
      To quickly get started with the common use cases of Mender, follow the{' '}
      <DocsLink path="get-started">
        Get Started chapter in our documentation <InlineLaunchIcon />
      </DocsLink>
      .
    </p>
    <h3>Installing Mender on your devices</h3>
    <p>
      We support several ways of integrating your devices with Mender. For a detailed introduction read the{' '}
      <DocsLink path="client-installation">
        Client Installation documentation <InlineLaunchIcon />
      </DocsLink>
      .
    </p>
    <p>
      Integrations for different boards and operating systems can be found on {/* eslint-disable-next-line react/jsx-no-target-blank */}
      <a href="https://hub.mender.io/c/board-integrations/6" target="_blank" rel="noopener">
        Mender Hub <InlineLaunchIcon />
      </a>
    </p>
    <h3>Creating Artifacts</h3>
    <p>
      Everything you need to package your software for deployment with Mender can be found here:{' '}
      <DocsLink path="artifact-creation">
        Artifact Creation documentation <InlineLaunchIcon />
      </DocsLink>
      .
    </p>
  </div>
);

export default GettingStarted;
