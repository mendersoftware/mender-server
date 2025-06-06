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
import { List } from '@mui/material';

import ExpandableAttribute from '@northern.tech/common-ui/ExpandableAttribute';

export const ArtifactMetadataList = ({ metaInfo = { content: [] } }) =>
  !!metaInfo.content.length && (
    <>
      <p className="margin-bottom-none">{metaInfo.title}</p>
      <List className="list-horizontal-flex" style={{ paddingTop: 0 }}>
        {metaInfo.content.map(({ key, ...info }) => (
          <ExpandableAttribute key={key} {...info} />
        ))}
      </List>
    </>
  );

export default ArtifactMetadataList;
