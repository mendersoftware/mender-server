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
import React, { useState } from 'react';

// material ui
import { Close as CloseIcon } from '@mui/icons-material';
import { Button, Divider, Drawer, IconButton, Tab, Tabs } from '@mui/material';

import { emptyWebhook } from '../../../constants/organizationConstants';
import WebhookActivity from './activity';
import WebhookConfiguration from './configuration';

const tabs = ['settings', 'activity'];

export const WebhookManagement = ({ adding, editing, events, eventTotal, getWebhookEvents, onCancel, onSubmit, onRemove, webhook = { ...emptyWebhook } }) => {
  const [currentTab, setCurrentTab] = useState(tabs[0]);

  return (
    <Drawer anchor="right" open={adding || editing} PaperProps={{ style: { minWidth: 600, width: '50vw' } }}>
      <div className="flexbox margin-bottom-small space-between">
        <h3>Manage webhook</h3>
        <div className="flexbox center-aligned">
          <Button color="secondary" onClick={() => onRemove(webhook)}>
            delete webhook
          </Button>
          <IconButton onClick={onCancel} aria-label="close">
            <CloseIcon />
          </IconButton>
        </div>
      </div>
      <Divider />
      <Tabs className="margin-top margin-bottom" value={currentTab} onChange={(e, tab) => setCurrentTab(tab)} textColor="primary">
        {tabs.map(key => (
          <Tab key={key} label={<div className="capitalized-start">{key}</div>} value={key} />
        ))}
      </Tabs>
      {currentTab === tabs[0] ? (
        <WebhookConfiguration adding={adding} editing={editing} onCancel={onCancel} onSubmit={onSubmit} webhook={webhook} />
      ) : (
        <WebhookActivity events={events} eventTotal={eventTotal} getWebhookEvents={getWebhookEvents} webhook={webhook} />
      )}
    </Drawer>
  );
};

export default WebhookManagement;
