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
import { Divider, Typography } from '@mui/material';

import { SynchronizedTwoColumnData } from '@northern.tech/common-ui/TwoColumnData';
import { isEmpty } from '@northern.tech/utils/helpers';

export const ArtifactMetadataList = ({ metaInfo = { title: '', content: {} } }) =>
  !isEmpty(metaInfo.content) && (
    <>
      <Typography variant="subtitle2">{metaInfo.title}</Typography>
      <Divider className="margin-top-small margin-bottom-small" />
      <SynchronizedTwoColumnData className="margin-bottom-small" data={metaInfo.content} />
    </>
  );

export default ArtifactMetadataList;
