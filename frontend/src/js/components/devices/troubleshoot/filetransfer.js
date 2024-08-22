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
import React, { useCallback, useEffect, useState } from 'react';
import { useDispatch } from 'react-redux';

import { FileCopy as CopyPasteIcon } from '@mui/icons-material';
import { Button, Divider, IconButton, InputAdornment, Tab, Tabs, TextField, Tooltip } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { deviceFileUpload } from '../../../actions/deviceActions';
import { canAccess } from '../../../constants/appConstants';
import FileUpload from '../../common/forms/fileupload';

const tabs = [
  { key: 'upload', canAccess: ({ userCapabilities: { canTroubleshoot, canWriteDevices } }) => canTroubleshoot && canWriteDevices },
  { key: 'download', canAccess }
];

const useStyles = makeStyles()(theme => ({
  column: { maxWidth: theme.spacing(80) },
  inputWrapper: { alignItems: 'end', gap: theme.spacing(2), display: 'grid', gridTemplateColumns: `${theme.spacing(60)} min-content` }
}));

const CopyPasteButton = ({ onClick }) => (
  <InputAdornment position="end">
    <Tooltip title="Paste" placement="top">
      <IconButton onClick={onClick}>
        <CopyPasteIcon />
      </IconButton>
    </Tooltip>
  </InputAdornment>
);

export const FileTransfer = ({
  device: { id: deviceId },
  downloadPath,
  file,
  onDownload,
  setFile,
  setDownloadPath,
  setUploadPath,
  uploadPath,
  userCapabilities
}) => {
  const { classes } = useStyles();
  const [currentTab, setCurrentTab] = useState(tabs[0].key);
  const [isValidDestination, setIsValidDestination] = useState(true);
  const [availableTabs, setAvailableTabs] = useState(tabs);
  const dispatch = useDispatch();

  useEffect(() => {
    let destination = currentTab === 'download' ? downloadPath : uploadPath;
    const isValid = destination.length ? /^(?:\/|[a-z]+:\/\/)/.test(destination) : true;
    setIsValidDestination(isValid);
  }, [currentTab, downloadPath, uploadPath]);

  useEffect(() => {
    const availableTabs = tabs.reduce((accu, item) => {
      if (item.canAccess({ userCapabilities })) {
        accu.push(item);
      }
      return accu;
    }, []);
    setAvailableTabs(availableTabs);
    setCurrentTab(availableTabs[0].key);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(userCapabilities)]);

  const onPasteDownloadClick = async () => {
    const path = await navigator.clipboard.readText();
    setDownloadPath(path);
  };

  const onPasteUploadClick = async () => {
    const path = await navigator.clipboard.readText();
    setUploadPath(path);
  };

  const onFileSelect = selectedFile => {
    let path;
    if (selectedFile) {
      path = `${uploadPath}/${selectedFile.name}`;
    } else {
      path = file && uploadPath.includes(file.name) ? uploadPath.substring(0, uploadPath.lastIndexOf('/')) : uploadPath;
    }
    setUploadPath(path);
    setFile(selectedFile);
  };

  const onUploadClick = useCallback(() => dispatch(deviceFileUpload(deviceId, uploadPath, file)), [dispatch, deviceId, uploadPath, file]);

  const fileInputProps = {
    error: !isValidDestination,
    fullWidth: true,
    InputLabelProps: { shrink: true }
  };

  return (
    <div className={classes.column}>
      <Tabs onChange={(e, item) => setCurrentTab(item)} value={currentTab} visibleScrollbar>
        {availableTabs.map(({ key }) => (
          <Tab className="capitalized" key={key} label={key} value={key} />
        ))}
      </Tabs>
      <Divider />
      {currentTab === 'upload' ? (
        <div className={classes.column}>
          <div>
            <p className="margin-top">
              <b>Upload a file to the device</b>
            </p>
            <FileUpload
              enableContentReading={false}
              fileNameSelection={file?.name}
              onFileChange={() => undefined}
              onFileSelect={onFileSelect}
              placeholder={
                <>
                  Drag here or <a>browse</a> to upload a file
                </>
              }
            />
          </div>
          <p className="margin-top margin-bottom-none">
            <b>Destination directory on the device where the file will be transferred</b>
          </p>
          <div className={classes.inputWrapper}>
            <TextField
              {...fileInputProps}
              helperText={!isValidDestination && <div className="warning">Destination has to be an absolute path</div>}
              onChange={e => setUploadPath(e.target.value)}
              placeholder="Example: /opt/installed-by-single-file"
              value={uploadPath}
              InputProps={{ endAdornment: <CopyPasteButton onClick={onPasteUploadClick} /> }}
            />
            <Button variant="contained" color="primary" disabled={!(file && uploadPath && isValidDestination)} onClick={onUploadClick}>
              Upload
            </Button>
          </div>
        </div>
      ) : (
        <div>
          <p className="margin-top margin-bottom-none">
            <b>Path to the file on the device</b>
          </p>
          <div className={classes.inputWrapper}>
            <TextField
              {...fileInputProps}
              helperText={!isValidDestination && <div className="warning">Destination has to be an absolute path</div>}
              onChange={e => setDownloadPath(e.target.value)}
              placeholder="Example: /home/mender/"
              value={downloadPath}
              InputProps={{ endAdornment: <CopyPasteButton onClick={onPasteDownloadClick} /> }}
            />
            <Button
              variant="contained"
              color="primary"
              disabled={!(downloadPath && isValidDestination)}
              onClick={() => onDownload(downloadPath)}
              style={{ alignSelf: 'flex-end' }}
            >
              Download
            </Button>
          </div>
        </div>
      )}
    </div>
  );
};

export default FileTransfer;
