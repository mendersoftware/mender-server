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
import { Divider, Table, TableBody, TableCell, TableHead, TableRow, Typography, lighten } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import FileSize from '@northern.tech/common-ui/FileSize';
import Time from '@northern.tech/common-ui/Time';
import { SynchronizedTwoColumnData } from '@northern.tech/common-ui/TwoColumnData';
import { getFormattedSize } from '@northern.tech/utils/helpers';

const METADATA_SPACING = 2;

const useStyles = makeStyles()(theme => ({
  table: {
    background: 'transparent'
  },
  payloadHeader: {
    background: lighten(theme.palette.background.paper, 0.25),
    margin: 0,
    padding: 10,
    position: 'absolute',
    top: -20
  }
}));

const attributes = ['Name', 'Checksum', 'Build date', 'Size (uncompressed)'];

export const ArtifactPayload = ({ index, payload: { files: payloadFiles, meta_data = {}, type_info } }) => {
  const { classes } = useStyles();
  const files = payloadFiles || [];
  const summedSize = files.reduce((accu, item) => accu + item.size, 0);
  const metaDataObject = meta_data;
  const metaData = {
    'Type': type_info.type,
    'Total Size (uncompressed)': getFormattedSize(summedSize)
  };
  return (
    <div className="file-details">
      <h4 className={classes.payloadHeader}>Payload {index}</h4>
      <SynchronizedTwoColumnData className="margin-top margin-bottom" data={metaData} />
      <div className="file-meta">
        {Object.keys(metaDataObject).length ? (
          <div>
            <h4>Update Metadata</h4>
            <pre>
              <code>{JSON.stringify(metaDataObject, null, METADATA_SPACING)}</code>
            </pre>
          </div>
        ) : null}
        <Typography variant="subtitle2">Files</Typography>
        <Divider className="margin-top-small margin-bottom-small" />
        {files.length ? (
          <Table className={classes.table}>
            <TableHead>
              <TableRow>
                {attributes.map((item, index) => (
                  <TableCell key={`file-header-${index}`} tooltip={item}>
                    {item}
                  </TableCell>
                ))}
              </TableRow>
            </TableHead>
            <TableBody className={classes.table}>
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
