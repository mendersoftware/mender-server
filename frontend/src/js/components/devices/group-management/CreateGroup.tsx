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
import { useWatch } from 'react-hook-form';
import { useSelector } from 'react-redux';

import { Button, DialogActions, DialogContent } from '@mui/material';

import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';
import Form from '@northern.tech/common-ui/forms/Form';
import { getGroups } from '@northern.tech/store/selectors';

import GroupDefinition from './GroupDefinition';

const inputName = 'group';

const CreateGroupContent = ({ groups, onClose, selectedDevices }) => {
  const group = useWatch({ name: inputName });

  return (
    <>
      <DialogContent>
        <GroupDefinition groups={groups} name={inputName} selectedDevices={selectedDevices} />
      </DialogContent>
      <DialogActions>
        <Button className="margin-right-x-small" onClick={onClose}>
          Cancel
        </Button>
        <Button variant="contained" type="submit">
          {groups.includes(group) ? 'Add to group' : 'Create group'}
        </Button>
      </DialogActions>
    </>
  );
};

export const CreateGroup = ({ addListOfDevices, onClose, selectedDevices }) => {
  const title = `Add ${selectedDevices.length ? 'selected ' : ''}devices to group`;

  const { static: staticGroups } = useSelector(getGroups);
  const groups = staticGroups.map(g => g.groupId);

  const onSubmit = ({ group }) => addListOfDevices(selectedDevices, group);

  return (
    <BaseDialog open title={title} fullWidth maxWidth="sm" onClose={onClose}>
      <Form onSubmit={onSubmit} defaultValues={{ [inputName]: '' }} validationMode="onSubmit">
        <CreateGroupContent groups={groups} onClose={onClose} selectedDevices={selectedDevices} />
      </Form>
    </BaseDialog>
  );
};

export default CreateGroup;
