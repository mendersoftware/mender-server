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

import { Info as InfoIcon } from '@mui/icons-material';
import { List } from '@mui/material';

import DocsLink from '../../common/docslink';
import Loader from '../../common/loader';
import { MenderTooltipClickable } from '../../common/mendertooltip';

export const DeviceInventoryLoader = () => (
  <List>
    <div className="waiting-inventory" key="waiting-inventory">
      <MenderTooltipClickable
        placement="left"
        disableFocusListener={false}
        title={
          <>
            <h3>Waiting for inventory data</h3>
            <p>Inventory data not yet received from the device - this can take up to 30 minutes with default installation.</p>
            <p>
              Also see the documentation for <DocsLink path="client-installation/configuration-file/polling-intervals" title="Polling intervals" />.
            </p>
          </>
        }
      >
        <InfoIcon />
      </MenderTooltipClickable>
      <p>Waiting for inventory data from the device</p>
      <Loader show={true} waiting={true} />
    </div>
  </List>
);

export default DeviceInventoryLoader;
