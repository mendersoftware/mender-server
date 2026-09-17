// Copyright 2016 Northern.tech AS
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
import { useState } from 'react';
import { useSelector } from 'react-redux';

import { Button, DialogActions, DialogContent } from '@mui/material';

import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';
import { getGroups } from '@northern.tech/store/selectors';

import GroupDefinition from './GroupDefinition';

export const CreateGroup = ({ addListOfDevices, isCreation, onClose, selectedDevices }) => {
  const [invalid, setInvalid] = useState(true);
  const [isModification, setIsModification] = useState(!isCreation);
  const [newGroup, setNewGroup] = useState('');
  const title = `Add ${selectedDevices.length ? 'selected ' : ''}devices to group`;

  const { static: staticGroups } = useSelector(getGroups);
  const groups = staticGroups.map(g => g.groupId);

  const onNameChange = (isInvalid, newGroupName, isModification) => {
    setInvalid(isInvalid);
    setIsModification(isModification);
    setNewGroup(newGroupName);
  };

  return (
    <BaseDialog open title={title} fullWidth maxWidth="sm" onClose={onClose}>
      <DialogContent>
        <GroupDefinition groups={groups} newGroup={newGroup} onInputChange={onNameChange} selectedDevices={selectedDevices} />
      </DialogContent>
      <DialogActions>
        <Button style={{ marginRight: 10 }} onClick={onClose}>
          Cancel
        </Button>
        <Button variant="contained" onClick={() => addListOfDevices(selectedDevices, newGroup)} disabled={!newGroup.length || invalid}>
          {!isModification || groups.length === 0 ? 'Create group' : 'Add to group'}
        </Button>
      </DialogActions>
    </BaseDialog>
  );
};

export default CreateGroup;
