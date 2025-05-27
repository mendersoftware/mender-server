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
import { Button, DialogActions, DialogContent } from '@mui/material';

import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';

const RemoveGroup = ({ onRemove, onClose }) => (
  <BaseDialog open title="Remove this group?" onClose={onClose}>
    <DialogContent>
      <p>This will remove the group from the list. Are you sure you want to continue?</p>
    </DialogContent>
    <DialogActions>
      <Button key="remove-action-button-1" onClick={onClose} style={{ marginRight: '10px' }}>
        Cancel
      </Button>
      <Button variant="contained" key="remove-action-button-2" color="primary" onClick={onRemove}>
        Remove group
      </Button>
    </DialogActions>
  </BaseDialog>
);

export default RemoveGroup;
