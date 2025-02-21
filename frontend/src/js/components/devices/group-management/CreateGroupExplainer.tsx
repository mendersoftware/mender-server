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

import { Button, DialogActions, DialogContent } from '@mui/material';

import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';

import CreateGroupExplainerContent from './CreateGroupExplainerContent';

export const CreateGroupExplainer = ({ isEnterprise, onClose }) => (
  <BaseDialog title="Creating a group" open fullWidth maxWidth="md" onClose={onClose}>
    <DialogContent>
      <CreateGroupExplainerContent isEnterprise={isEnterprise} />
    </DialogContent>
    <DialogActions>
      <Button onClick={onClose}>Close</Button>
    </DialogActions>
  </BaseDialog>
);

export default CreateGroupExplainer;
