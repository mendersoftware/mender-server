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

import { Announcement as AnnouncmentIcon, Close as CloseIcon } from '@mui/icons-material';

import DocsLink from '../common/docslink';

const OfferHeader = ({ onHide }) => (
  <div id="offerHeader" className="offerBox">
    <AnnouncmentIcon style={{ marginRight: '2px', height: '16px', verticalAlign: 'bottom' }} />
    <span>
      Learn more about Mender&apos;s new add-ons at&nbsp;
      {/* eslint-disable-next-line react/jsx-no-target-blank */}
      <a href="https://mender.io/plans/pricing" target="_blank">
        mender.io/plans/pricing
      </a>
      .&nbsp;
      <DocsLink path="add-ons" title="Visit our documentation" />
      &nbsp;to learn how to enable add-on features
    </span>
    <CloseIcon style={{ marginLeft: '4px', height: '16px', cursor: 'pointer' }} onClick={onHide} />
  </div>
);

export default OfferHeader;
