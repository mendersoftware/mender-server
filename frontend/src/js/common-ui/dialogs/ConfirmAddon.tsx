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
import { Button, Dialog, DialogActions, DialogContent, DialogTitle } from '@mui/material';

interface ConfirmAddonProps {
  variant: 'remove' | 'add';
  name: string;
  onConfirm: () => void;
  onClose: () => void;
}
const title = {
  remove: 'Remove add-on?',
  add: 'Add new on?'
};

export const ConfirmAddon = (props: ConfirmAddonProps) => {
  const { variant, name, onConfirm, onClose } = props;
  return (
    <Dialog open={true}>
      <DialogTitle>{title[variant]}</DialogTitle>
      <DialogContent>
        You are requesting to {variant} the Mender <b>{name}</b> add-on from your plan.
      </DialogContent>
      <DialogContent>
        Once we receive the request, we will get in touch to confirm the change to your subscription and price, before{' '}
        {variant === 'remove' ? 'disabling' : 'enabling'} the add-on.
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button color="secondary" variant="contained" onClick={onConfirm}>
          Confirm
        </Button>
      </DialogActions>
    </Dialog>
  );
};
