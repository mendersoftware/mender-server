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
import React from 'react';

import { Button, Dialog, DialogActions, DialogContent, DialogTitle } from '@mui/material';

const content = {
  artifact: ({ artifact }) => (
    <>
      Are you sure you want to remove <i>{artifact.name}</i>?
    </>
  ),
  release: ({ release }) => (
    <>
      All artifacts in the <i>{release.name}</i> release will be removed. Are you sure?
    </>
  )
};

const RemoveArtifactDialog = ({ artifact, onCancel, open, onRemove, release }) => {
  const type = artifact ? 'artifact' : 'release';
  const Content = content[type];
  return (
    <Dialog open={open}>
      <DialogTitle>Remove this {type}?</DialogTitle>
      <DialogContent>
        <Content artifact={artifact} release={release} />
      </DialogContent>
      <DialogActions>
        <Button onClick={onCancel}>Cancel</Button>
        <div style={{ flexGrow: 1 }} />
        <Button variant="contained" color="secondary" onClick={onRemove}>
          Remove {type}
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export default RemoveArtifactDialog;
