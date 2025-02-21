// Copyright 2025 Northern.tech AS
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
import { ReactNode } from 'react';

import { Dialog, DialogProps, DialogTitle } from '@mui/material';
import { Breakpoint } from '@mui/system';

interface BaseDialogProps {
  children?: ReactNode;
  className?: string;
  disableEscapeKeyDown?: boolean;
  fullWidth?: boolean;
  maxWidth?: Breakpoint | false;
  onClose: (e?: any) => void;
  open: boolean;
  slotProps?: DialogProps['slotProps'];
  title: string | ReactNode;
}
export const BaseDialog = (props: BaseDialogProps) => {
  const { open, maxWidth, className = '', children, onClose, title, slotProps, disableEscapeKeyDown, fullWidth } = props;
  return (
    <Dialog
      className={className}
      disableEscapeKeyDown={disableEscapeKeyDown}
      open={open}
      maxWidth={maxWidth}
      fullWidth={fullWidth}
      onClose={onClose}
      slotProps={slotProps}
    >
      <DialogTitle>{title}</DialogTitle>
      {children}
    </Dialog>
  );
};
