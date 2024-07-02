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
import React, { useState } from 'react';
import { useSelector } from 'react-redux';

import { Button, Dialog, DialogActions, DialogContent, DialogTitle } from '@mui/material';

import { getGroups, getSelectedGroupInfo } from '../../../selectors';
import GroupDefinition from './group-definition';

export const CreateGroup = ({ addListOfDevices, fromFilters, isCreation, onClose, selectedDevices }) => {
  const isCreationDynamic = isCreation && fromFilters;
  const [invalid, setInvalid] = useState(true);
  const [isModification, setIsModification] = useState(!isCreation);
  const [newGroup, setNewGroup] = useState('');
  const [title, setTitle] = useState(isCreationDynamic ? 'Create a new group' : `Add ${selectedDevices.length ? 'selected ' : ''}devices to group`);

  const { selectedGroup } = useSelector(getSelectedGroupInfo);
  // ensure that existing dynamic groups are only listed if a dynamic group should be created
  const { dynamic, static: staticGroups } = useSelector(getGroups);
  const groups = fromFilters ? [...staticGroups.map(g => g.groupId), ...dynamic.map(g => g.groupId)] : staticGroups.map(g => g.groupId);

  const onNameChange = (isInvalid, newGroupName, isModification) => {
    const title = !isCreationDynamic ? `Add ${selectedDevices.length ? 'selected ' : ''}devices to group` : 'Create a new group';
    setTitle(title);
    setInvalid(isInvalid);
    setIsModification(isModification);
    setNewGroup(newGroupName);
  };

  return (
    <Dialog disableEscapeKeyDown open={true} scroll="paper" fullWidth={true} maxWidth="sm">
      <DialogTitle style={{ paddingBottom: '15px', marginBottom: 0 }}>{title}</DialogTitle>
      <DialogContent className="dialog">
        <GroupDefinition
          groups={groups}
          isCreationDynamic={isCreationDynamic}
          newGroup={newGroup}
          onInputChange={(invalidName, name, isModification) => onNameChange(invalidName, name, isModification)}
          selectedDevices={selectedDevices}
          selectedGroup={selectedGroup}
        />
      </DialogContent>
      <DialogActions style={{ marginTop: 0 }}>
        <Button style={{ marginRight: 10 }} onClick={onClose}>
          Cancel
        </Button>
        <Button variant="contained" color="primary" onClick={() => addListOfDevices(selectedDevices, newGroup)} disabled={!newGroup.length || invalid}>
          {!isModification || isCreationDynamic || groups.length === 0 ? 'Create group' : 'Add to group'}
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export default CreateGroup;
