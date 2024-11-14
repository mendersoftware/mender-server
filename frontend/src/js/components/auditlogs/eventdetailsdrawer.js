// Copyright 2021 Northern.tech AS
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

import { Divider, Drawer } from '@mui/material';
import { useTheme } from '@mui/material/styles';

import { DrawerTitle } from '@northern.tech/common-ui/DrawerTitle';
import { Code } from '@northern.tech/common-ui/copy-code';

import DeviceConfiguration from './eventdetails/deviceconfiguration';
import FileTransfer from './eventdetails/filetransfer';
import PortForward from './eventdetails/portforward';
import TerminalSession from './eventdetails/terminalsession';
import { UserChange } from './eventdetails/userchange';

const FallbackComponent = ({ item }) => {
  let content = '';
  try {
    content = JSON.stringify(item, null, 2);
  } catch (error) {
    content = `error parsing the logged event:\n${error}`;
  }
  return <Code style={{ whiteSpace: 'pre' }}>{content}</Code>;
};

const changeTypes = {
  user: 'user',
  device: 'device',
  tenant: 'tenant'
};

const configChangeDescriptor = {
  set_configuration: 'definition',
  deploy_configuration: 'deployment'
};

const mapChangeToContent = item => {
  const { type } = item.object || {};
  let content = { title: 'Entry details', content: FallbackComponent };
  if (type === changeTypes.user) {
    content = { title: `${item.action}d user`, content: UserChange };
  } else if (type === changeTypes.device && item.action.includes('terminal')) {
    content = { title: 'Remote session log', content: TerminalSession };
  } else if (type === changeTypes.device && item.action.includes('file')) {
    content = { title: 'File transfer', content: FileTransfer };
  } else if (type === changeTypes.device && item.action.includes('portforward')) {
    content = { title: 'Port forward', content: PortForward };
  } else if (type === changeTypes.device && item.action.includes('configuration')) {
    content = { title: `Device configuration ${configChangeDescriptor[item.action] || ''}`, content: DeviceConfiguration };
  } else if (type === changeTypes.device) {
    content = { title: 'Device change', content: FallbackComponent };
  } else if (type === changeTypes.tenant) {
    content = { title: `${item.action}d tenant`, content: UserChange };
  }
  return content;
};

export const EventDetailsDrawer = ({ eventItem = {}, onClose, open }) => {
  const theme = useTheme();
  const { title, content: Component } = mapChangeToContent(eventItem);
  return (
    <Drawer className={`${open ? 'fadeIn' : 'fadeOut'}`} anchor="right" open={open} onClose={onClose}>
      <DrawerTitle title={<div className="capitalized-start">{title}</div>} onClose={onClose} />
      <Divider />
      <Component item={eventItem} onClose={onClose} />
      <Divider light style={{ marginTop: theme.spacing(2) }} />
    </Drawer>
  );
};

export default EventDetailsDrawer;
