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
import React from 'react';

import { List, ListItem, ListItemText, Table, TableBody, TableCell, TableHead, TableRow } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { FileSize, getFormattedSize } from '../../helpers';
import Time from '../common/time';

const METADATA_SPACING = 2;

const useStyles = makeStyles()(theme => ({
  metadataList: {
    display: 'flex',
    flexDirection: 'row'
  },
  table: {
    background: 'transparent'
  },
  metadataListItem: {
    paddingBottom: 11,
    borderBottom: `1px solid ${theme.palette.grey[600]}`,
    marginRight: '2vw'
  },
  payloadHeader: {
    background: theme.palette.grey[500],
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
  const metaData = [
    { title: 'Type', value: type_info.type },
    { title: 'Total Size (uncompressed)', value: getFormattedSize(summedSize) }
  ];
  return (
    <div className="file-details">
      <h4 className={classes.payloadHeader}>Payload {index}</h4>
      <List className={classes.metadataList}>
        {metaData.map((item, index) => (
          <ListItem disabled={true} className={classes.metadataListItem} classes={{ root: 'attributes', disabled: 'opaque' }} key={`metadata-item-${index}`}>
            <ListItemText primary={item.title} secondary={item.value} />
          </ListItem>
        ))}
      </List>
      <div className="file-meta">
        {Object.keys(metaDataObject).length ? (
          <div>
            <h4>Update Metadata</h4>
            <pre>
              <code>{JSON.stringify(metaDataObject, null, METADATA_SPACING)}</code>
            </pre>
          </div>
        ) : null}
        <h4>Files</h4>
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
