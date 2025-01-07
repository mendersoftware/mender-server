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

export const EventDetailsDrawer = ({ eventItem = {}, onClose, open, mapChangeToContent, fallbackComponent }) => {
  const theme = useTheme();
  const { title, content: Component } = mapChangeToContent(eventItem, fallbackComponent);
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
