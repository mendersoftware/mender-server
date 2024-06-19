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
import Linkify from 'react-linkify';

import { Announcement as AnnounceIcon, Close as CloseIcon } from '@mui/icons-material';

const Announcement = ({ announcement = '', errorIconClassName, iconClassName, onHide, sectionClassName }) => (
  <div className={`flexbox centered fadeInSlow ${sectionClassName}`}>
    <AnnounceIcon className={errorIconClassName} fontSize="small" style={{ marginRight: 4, minWidth: 24 }} />
    <p style={{ margin: 4 }}>
      <Linkify.default properties={{ target: '_blank' }}>{announcement}</Linkify.default>
    </p>
    <CloseIcon className={iconClassName} style={{ marginLeft: 4, cursor: 'pointer' }} onClick={onHide} />
  </div>
);

export default Announcement;
