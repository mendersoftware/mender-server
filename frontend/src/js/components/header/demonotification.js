// Copyright 2020 Northern.tech AS
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

import { InfoOutlined as InfoIcon } from '@mui/icons-material';

import DocsLink from '../common/docslink';
import { MenderTooltipClickable } from '../common/mendertooltip';

const DemoNotification = ({ iconClassName, sectionClassName }) => (
  <MenderTooltipClickable
    className={`flexbox centered ${sectionClassName}`}
    title={
      <>
        <h3>Demo mode</h3>
        <p>
          Mender is currently running in <b>demo mode</b>.
        </p>
        <p>
          <DocsLink path="server-installation/production-installation-with-kubernetes" title="See the documentation for help switching to production mode" />.
        </p>
      </>
    }
  >
    <>
      <InfoIcon className={iconClassName} style={{ marginRight: 2, verticalAlign: 'bottom' }} />
      <a>Demo mode</a>
    </>
  </MenderTooltipClickable>
);

export default DemoNotification;
