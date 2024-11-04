// Copyright 2024 Northern.tech AS
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
import { FunctionComponent } from 'react';

import { Button, Dialog, DialogActions, DialogContent, DialogTitle } from '@mui/material';

interface Props {
  dismiss: () => void;
  submit: () => void;
  name: string;
  open: boolean;
}

export const DeleteRoleDialog: FunctionComponent<Props> = ({ dismiss, open, submit, name }) => (
  <Dialog open={open}>
    <DialogTitle>Delete role?</DialogTitle>
    <DialogContent style={{ overflow: 'hidden' }}>
      Are you sure you want to delete the role{' '}
      <b>
        <i>{name}</i>
      </b>
      ?
    </DialogContent>
    <DialogActions>
      <Button style={{ marginRight: 10 }} onClick={dismiss}>
        Cancel
      </Button>
      <Button variant="contained" color="primary" onClick={submit}>
        Delete role
      </Button>
    </DialogActions>
  </Dialog>
);
