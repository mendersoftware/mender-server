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
import { FormEvent, useState } from 'react';

import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogProps, DialogTitle, TextField } from '@mui/material';

interface ConfirmModalProps {
  close: () => void;
  onConfirm: () => void;
  toType: string;
  header: string;
  description: string;
  open: boolean;
  className?: string;
  maxWidth?: DialogProps['maxWidth'];
}
export const ConfirmModal = (props: ConfirmModalProps) => {
  const { close, onConfirm, className, toType, header, description, open, maxWidth = 'xs' } = props;
  const [inputValue, setInputValue] = useState<string>('');
  return (
    <Dialog
      className={className || ''}
      open={open}
      onClose={close}
      maxWidth={maxWidth}
      PaperProps={{
        component: 'form',
        onSubmit: (event: FormEvent<HTMLFormElement>) => {
          event.preventDefault();
          onConfirm();
          close();
        }
      }}
    >
      <DialogTitle>{header}</DialogTitle>
      <DialogContent>
        <DialogContentText>{description}</DialogContentText>
        <DialogContentText>Type &#39;{toType}&#39; below to continue</DialogContentText>
        <TextField
          value={inputValue}
          onChange={e => setInputValue(e.target.value)}
          autoFocus
          required
          name="confirmation-text"
          id="confirmation-text"
          label={toType}
          type="text"
        />
      </DialogContent>
      <DialogActions>
        <Button onClick={close}>Cancel</Button>
        <Button color="secondary" type="submit" variant="contained" disabled={inputValue !== toType}>
          Confirm
        </Button>
      </DialogActions>
    </Dialog>
  );
};
