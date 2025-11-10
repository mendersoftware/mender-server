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
import { Button, DialogActions, DialogContent, Divider } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import CopyCode from '@northern.tech/common-ui/CopyCode';
import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';
import { createFileDownload } from '@northern.tech/utils/helpers';

import { AiLogAnalysis } from './AiLogAnalysis';

const useStyles = makeStyles()(() => ({
  codeContainer: {
    overflowY: 'auto',
    '.code': {
      minHeight: '100%'
    }
  },
  wrapper: {
    display: 'grid',
    '&.ai-enabled': {
      gridTemplateRows: 'minmax(70%, 1fr) min-content min-content',
      paddingBottom: 0
    }
  }
}));

const getFilename = ({ device, releaseName, date }) => `deployment-log-${device}-${releaseName}-${date}.log`;

export const LogDialog = ({ canAi, deployment, deviceId, onClose }) => {
  const { classes } = useStyles();
  const { devices = {} } = deployment;
  const { log: logData } = devices[deviceId] || {};
  const context = { device: deviceId, releaseName: deployment.artifact_name, date: deployment.finished };

  const exportLog = () => createFileDownload(logData, getFilename(context), '');

  return (
    <BaseDialog open title="Deployment log for device" maxWidth="xl" onClose={onClose}>
      <DialogContent className={`${classes.wrapper} ${canAi ? 'ai-enabled' : ''}`}>
        <div className={classes.codeContainer}>
          <CopyCode code={logData} withDescription />
        </div>
        {canAi && (
          <>
            <Divider className="margin-top-small" />
            <AiLogAnalysis deployment={deployment} deviceId={deviceId} />
          </>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button variant="contained" onClick={exportLog}>
          Export log
        </Button>
      </DialogActions>
    </BaseDialog>
  );
};

export default LogDialog;
