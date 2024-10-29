// Copyright 2017 Northern.tech AS
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
import React, { useState } from 'react';

import { Cancel as CancelIcon, CheckCircle as CheckCircleIcon, Check as CheckIcon, Close as CloseIcon, Edit as EditIcon } from '@mui/icons-material';
import { Button, IconButton } from '@mui/material';

const defaultRemoving = 'Removing...';

const confirmationType = {
  retry: {
    loading: 'Creating new deployment...',
    message: 'Confirm retry?'
  },
  abort: {
    loading: 'Aborting...',
    message: 'Confirm abort?'
  },
  chartRemoval: {
    loading: defaultRemoving,
    message: 'Remove this chart?'
  },
  decommissioning: {
    loading: 'Decommissioning...',
    message: 'Decommission this device and remove all of its data from the server. This cannot be undone. Are you sure?'
  },
  deploymentContinuation: {
    loading: 'Continuing...',
    message: 'All devices with no errors will continue to the next step of the updates. Confirm continue?'
  },
  deploymentAbort: {
    loading: 'Aborting...',
    message: 'This will abort the deployment and attempt to roll back all devices. Confirm abort?'
  },
  integrationRemoval: {
    loading: defaultRemoving,
    message: 'Remove the ingration. Are you sure?'
  },
  webhooksRemoval: {
    loading: defaultRemoving,
    message: 'Delete all webhooks?'
  }
};

export const Confirm = ({ action, cancel, classes = '', message = '', style = {}, type }) => {
  const [className, setClassName] = useState('fadeIn');
  const [loading, setLoading] = useState(false);

  const handleCancel = () => {
    setClassName('fadeOut');
    cancel();
  };
  const handleConfirm = () => {
    setLoading(true);
    action();
  };

  let notification = message;
  if (confirmationType[type]) {
    notification = loading ? confirmationType[type].loading : confirmationType[type].message;
  }
  return (
    <div className={`flexbox center-aligned ${className} ${classes}`} style={{ marginRight: '12px', justifyContent: 'flex-end', ...style }}>
      <span className="bold">{notification}</span>
      <IconButton id="confirmAbort" onClick={handleConfirm} size="large">
        <CheckCircleIcon className="green" />
      </IconButton>
      <IconButton onClick={handleCancel} size="large">
        <CancelIcon className="red" />
      </IconButton>
    </div>
  );
};

export const EditButton = ({ onClick, disabled = false }) => (
  <Button onClick={onClick} size="small" disabled={disabled} startIcon={<EditIcon />} style={{ padding: 5 }}>
    Edit
  </Button>
);

export const ConfirmationButtons = ({ onConfirm, onCancel }) => (
  <div className="flexbox">
    <IconButton onClick={onConfirm} size="small">
      <CheckIcon color="disabled" />
    </IconButton>
    <IconButton onClick={onCancel} size="small">
      <CloseIcon color="disabled" />
    </IconButton>
  </div>
);

export default Confirm;
