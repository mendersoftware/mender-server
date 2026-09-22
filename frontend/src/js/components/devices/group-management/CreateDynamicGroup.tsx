// Copyright 2026 Northern.tech AS
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
import { useSelector } from 'react-redux';

import { Button, DialogActions, DialogContent, Typography } from '@mui/material';

import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';
import Form from '@northern.tech/common-ui/forms/Form';
import TextInput from '@northern.tech/common-ui/forms/TextInput';
import { getGroupNames } from '@northern.tech/store/selectors';

import { groupNameValidationRules } from './GroupDefinition';

export const CreateDynamicGroup = ({ onClose, onCreate }) => {
  // a dynamic group can't reuse the name of any existing group, static or dynamic
  const groups = useSelector(getGroupNames);

  const onSubmit = ({ groupName }) => onCreate(groupName);

  return (
    <BaseDialog open title="Create a dynamic group" fullWidth maxWidth="sm" onClose={onClose}>
      <Form onSubmit={onSubmit} defaultValues={{ groupName: '' }} validationMode="onSubmit">
        <DialogContent>
          <Typography variant="subtitle1">Name</Typography>
          <TextInput
            className="margin-top-x-small"
            helperText="Devices that match the filter will be automatically part of this group."
            id="groupName"
            InputLabelProps={{ shrink: true }}
            hint="Group name"
            rules={groupNameValidationRules({ existingGroups: groups, isDynamic: true })}
            width="100%"
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={onClose}>Cancel</Button>
          <Button variant="contained" type="submit">
            Create group
          </Button>
        </DialogActions>
      </Form>
    </BaseDialog>
  );
};

export default CreateDynamicGroup;
