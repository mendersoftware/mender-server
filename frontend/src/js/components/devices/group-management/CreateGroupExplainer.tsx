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

import { Button, Dialog, DialogActions, DialogContent, DialogTitle } from '@mui/material';

import CreateGroupExplainerContent from './CreateGroupExplainerContent';

export const CreateGroupExplainer = ({ isEnterprise, onClose }) => (
  <Dialog className="dialog" disableEscapeKeyDown open={true} scroll="paper" fullWidth={true} maxWidth="md">
    <DialogTitle style={{ marginLeft: 15 }}>Creating a group</DialogTitle>
    <DialogContent>
      <CreateGroupExplainerContent isEnterprise={isEnterprise} />
    </DialogContent>
    <DialogActions>
      <Button onClick={onClose}>Close</Button>
    </DialogActions>
  </Dialog>
);

export default CreateGroupExplainer;
