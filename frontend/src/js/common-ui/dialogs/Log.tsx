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
import type { ReactNode } from 'react';

import { Button, DialogActions, DialogContent, Divider, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';
import storeActions from '@northern.tech/store/actions';
import { useAppDispatch } from '@northern.tech/store/store';
import { createFileDownload } from '@northern.tech/utils/helpers';
import copy from 'copy-to-clipboard';

import { Code } from '../CopyCode';

const { setSnackbar } = storeActions;

const useStyles = makeStyles()(() => ({
  wrapper: {
    display: 'grid',
    '&.with-extras': {
      gridTemplateRows: 'minmax(70%, 1fr) min-content min-content',
      paddingBottom: 0
    }
  }
}));

interface LogContext {
  date?: string;
  device?: string;
  releaseName?: string;
}

const dialogTypes = {
  deviceLog: {
    title: 'Deployment log for device',
    filename: ({ device, releaseName, date }: LogContext) => `deployment-log-${device}-${releaseName}-${date}.log`
  },
  configUpdateLog: {
    title: 'Config update log for device',
    filename: () => 'configuration-update.log'
  }
};

interface LogDialogProps {
  children?: ReactNode;
  className?: string;
  context?: LogContext;
  logData: string;
  onClose: () => void;
  type?: keyof typeof dialogTypes;
}

export const LogDialog = ({ className = '', children, context = {}, logData = '', onClose, type = 'deviceLog' }: LogDialogProps) => {
  const { classes } = useStyles();
  const dispatch = useAppDispatch();
  const { filename, title } = dialogTypes[type];

  const exportLog = () => createFileDownload(logData, filename(context), '');

  const onCopyClick = () => {
    copy(logData);
    dispatch(setSnackbar('Copied to clipboard'));
  };

  return (
    <BaseDialog open title={title} maxWidth="lg" onClose={onClose}>
      <DialogContent className={`${classes.wrapper} ${children ? 'with-extras' : ''} ${className}`}>
        <Code className="full-height">
          <Typography component="code" variant="code1" className="copyable-content">
            {logData}
          </Typography>
        </Code>
        {!!children && (
          <>
            <Divider className="margin-top-small" />
            {children}
          </>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button variant="outlined" color="info" onClick={onCopyClick}>
          Copy to clipboard
        </Button>
        <Button variant="contained" onClick={exportLog}>
          Export log
        </Button>
      </DialogActions>
    </BaseDialog>
  );
};

export default LogDialog;
