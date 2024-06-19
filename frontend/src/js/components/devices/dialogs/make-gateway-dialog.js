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

// material ui
import { Button, Dialog, DialogActions, DialogContent, DialogTitle } from '@mui/material';

import { getToken } from '../../../auth';
import CopyCode from '../../common/copy-code';
import DocsLink from '../../common/docslink';

export const getCode = isPreRelease => {
  const { target, flags } = isPreRelease
    ? { target: 'https://get.mender.io/staging', flags: ' -c --experimental' }
    : { target: 'https://get.mender.io', flags: '' };
  return `JWT_TOKEN='${getToken()}'

wget -O- ${target} | sudo bash -s -- --jwt-token $JWT_TOKEN mender-gateway --demo${flags}`;
};

export const MakeGatewayDialog = ({ isPreRelease, onCancel }) => (
  <Dialog open fullWidth maxWidth="md">
    <DialogTitle>Promoting a device to a gateway</DialogTitle>
    <DialogContent className="onboard-dialog dialog-content">
      You can test Mender Gateway by promoting a device to a gateway device, enabling other devices to securely contact the Mender Server through it.
      <p>
        On the device terminal, run the following command. You can use <DocsLink path="add-ons/remote-terminal" title="Remote Terminal" /> if mender-connect is
        enabled on the device.
      </p>
      <CopyCode code={getCode(isPreRelease)} withDescription />
      <p>
        Note: this is only intended for demo or testing purposes. For production installation please refer to the{' '}
        <DocsLink path="get-started/mender-gateway" title="full Mender Gateway documentation" />
      </p>
    </DialogContent>
    <DialogActions>
      <Button onClick={onCancel}>Cancel</Button>
      <div style={{ flexGrow: 1 }} />
      <Button variant="contained" onClick={onCancel} color="secondary">
        Close
      </Button>
    </DialogActions>
  </Dialog>
);

export default MakeGatewayDialog;
