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

import { Button } from '@mui/material';

import { InfoHintContainer } from '@northern.tech/common-ui/info-hint';
import Loader from '@northern.tech/common-ui/loader';

export const AuditlogsView = ({ total, csvLoading, createCsvDownload, infoHintComponent = null, auditLogsFilter, children }) => {
  return (
    <div className="fadeIn margin-left flexbox column" style={{ marginRight: '5%' }}>
      <div className="flexbox center-aligned">
        <h3 className="margin-right-small">Audit log</h3>
        <InfoHintContainer>{infoHintComponent}</InfoHintContainer>
      </div>
      {auditLogsFilter}
      <div className="flexbox center-aligned" style={{ justifyContent: 'flex-end' }}>
        <Loader show={csvLoading} />
        <Button variant="contained" color="secondary" disabled={csvLoading || !total} onClick={createCsvDownload} style={{ marginLeft: 15 }}>
          Download results as csv
        </Button>
      </div>
      {children}
    </div>
  );
};

export default AuditlogsView;
