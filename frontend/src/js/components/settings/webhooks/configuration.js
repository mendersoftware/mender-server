// Copyright 2022 Northern.tech AS
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
import React, { useCallback, useEffect, useState } from 'react';

// material ui
import { Close as CloseIcon } from '@mui/icons-material';
import { Button, Divider, Drawer, IconButton, TextField } from '@mui/material';

import validator from 'validator';

import { EXTERNAL_PROVIDER } from '../../../constants/deviceConstants';
import { emptyWebhook } from '../../../constants/organizationConstants';
import MenderTooltip from '../../common/mendertooltip';
import { HELPTOOLTIPS, MenderHelpTooltip } from '../../helptips/helptooltips';

const WebhookConfiguration = ({ adding, editing, onCancel, onSubmit, webhook = { ...emptyWebhook } }) => {
  const [description, setDescription] = useState('');
  const [secret, setSecret] = useState('');
  const [secretFormatError, setSecretFormatError] = useState('');
  const [url, setUrl] = useState('');

  const { credentials = {}, description: hookDescription = '', id = 'new' } = webhook;
  const {
    [EXTERNAL_PROVIDER.webhook.credentialsType]: { secret: hookSecret = '', url: hookUrl = '' }
  } = credentials;
  useEffect(() => {
    setDescription(hookDescription);
    setUrl(hookUrl);
    setSecret(hookSecret);
  }, [adding, editing, hookDescription, hookSecret, hookUrl]);

  useEffect(() => {
    setSecretFormatError(!secret || validator.isHexadecimal(secret) ? '' : 'The secret has to be entered as a hexadecimal string');
  }, [secret]);

  const onSubmitClick = useCallback(() => {
    let webhookConfig = {
      id,
      provider: EXTERNAL_PROVIDER.webhook.provider,
      credentials: { type: EXTERNAL_PROVIDER.webhook.credentialsType, [EXTERNAL_PROVIDER.webhook.credentialsType]: { secret, url } },
      description
    };
    if (editing) {
      // eslint-disable-next-line no-unused-vars
      const { credentials, description, ...remainder } = webhookConfig;
      webhookConfig = { ...remainder, credentials: { ...credentials, [EXTERNAL_PROVIDER.webhook.credentialsType]: { url } } };
    }
    onSubmit(webhookConfig);
  }, [description, editing, id, onSubmit, secret, url]);

  const secretInputTip = editing ? 'Cannot edit webhook secrets after they have been saved' : 'The secret has to be entered as a hexadecimal string';
  const descriptionInput = (
    <TextField
      label="Description (optional)"
      id="webhook-description"
      disabled={editing}
      multiline
      onChange={e => setDescription(e.target.value)}
      value={description}
    />
  );
  const urlInput = <TextField label="Url" id="webhook-name" disabled={editing} value={url} onChange={e => setUrl(e.target.value)} />;
  const isSubmitDisabled = !url || (editing && url === hookUrl);
  return (
    <>
      <div className="flexbox column" style={{ width: 500 }}>
        {editing ? (
          <MenderTooltip arrow placement="bottom-start" title="Cannot edit webhook url after it has been saved">
            {urlInput}
          </MenderTooltip>
        ) : (
          urlInput
        )}
        {editing ? (
          <MenderTooltip arrow placement="bottom-start" title="Cannot edit webhook description after it has been saved">
            {descriptionInput}
          </MenderTooltip>
        ) : (
          descriptionInput
        )}
        <MenderTooltip arrow placement="bottom-start" title={secretInputTip}>
          <TextField
            label="Secret (optional)"
            id="webhook-secret"
            disabled={editing}
            error={!editing && !!secretFormatError}
            helperText={!editing && secretFormatError}
            onChange={e => setSecret(e.target.value)}
            type={editing ? 'password' : 'text'}
            value={secret}
          />
        </MenderTooltip>
      </div>
      <Divider className="margin-top-large" light />
      <div className="flexbox centered margin-top" style={{ justifyContent: 'flex-end' }}>
        <Button onClick={onCancel}>Cancel</Button>
        {!editing && (
          <Button color="secondary" variant="contained" disabled={isSubmitDisabled} onClick={onSubmitClick}>
            Submit
          </Button>
        )}
      </div>
    </>
  );
};

export const WebhookCreation = ({ adding, onCancel, ...props }) => (
  <Drawer anchor="right" open={adding} PaperProps={{ style: { minWidth: 600, width: '50vw' } }}>
    <div className="flexbox center-aligned margin-bottom-small space-between">
      <div className="flexbox center-aligned">
        <h3>Add a webhook</h3>
        <MenderHelpTooltip id={HELPTOOLTIPS.webhooks.id} />
      </div>
      <IconButton onClick={onCancel} aria-label="close">
        <CloseIcon />
      </IconButton>
    </div>
    <Divider />
    <WebhookConfiguration adding={adding} onCancel={onCancel} {...props} />
  </Drawer>
);

export default WebhookConfiguration;
