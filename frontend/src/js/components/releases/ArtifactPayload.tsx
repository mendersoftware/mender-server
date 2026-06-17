// Copyright 2019 Northern.tech AS
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
import { Divider, Table, TableBody, TableCell, TableHead, TableRow, Typography } from '@mui/material';

import CopyCode from '@northern.tech/common-ui/CopyCode';
import FileSize from '@northern.tech/common-ui/FileSize';
import Time from '@northern.tech/common-ui/Time';
import { SynchronizedTwoColumnData } from '@northern.tech/common-ui/TwoColumnData';
import { getFormattedSize } from '@northern.tech/utils/helpers';

const METADATA_SPACING = 2;

const attributes = ['Name', 'Checksum', 'Build date', 'Size (uncompressed)'];

export const ArtifactPayload = ({ payload: { files: payloadFiles, meta_data = {}, type_info } }) => {
  const files = payloadFiles || [];
  const summedSize = files.reduce((accu, item) => accu + item.size, 0);
  const metaDataObject = meta_data;
  const metaData = {
    'Type': type_info.type,
    'Total Size (uncompressed)': getFormattedSize(summedSize)
  };
  return (
    <div>
      <SynchronizedTwoColumnData className="margin-bottom-small" data={metaData} />
      <div>
        {Object.keys(metaDataObject).length ? (
          <div>
            <Typography variant="subtitle2" className="margin-bottom-small">
              Update Metadata
            </Typography>
            <CopyCode code={JSON.stringify(metaDataObject, null, METADATA_SPACING)} size="medium" />
          </div>
        ) : null}
        <Typography variant="subtitle2">Files</Typography>
        <Divider className="margin-top-small margin-bottom-small" />
        {files.length ? (
          <Table>
            <TableHead>
              <TableRow>
                {attributes.map((item, index) => (
                  <TableCell key={`file-header-${index}`} tooltip={item}>
                    {item}
                  </TableCell>
                ))}
              </TableRow>
            </TableHead>
            <TableBody>
              {files.map((file, index) => (
                <TableRow key={index}>
                  <TableCell>{file.name}</TableCell>
                  <TableCell style={{ wordBreak: 'break-word' }}>{file.checksum}</TableCell>
                  <TableCell>
                    <Time value={file.date} />
                  </TableCell>
                  <TableCell>
                    <FileSize fileSize={file.size} />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : (
          <p>There are no files in this Artifact</p>
        )}
      </div>
    </div>
  );
};

export default ArtifactPayload;
