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
import type { CSSProperties, ReactNode } from 'react';
import { useState } from 'react';
import Dropzone from 'react-dropzone';

// material ui
import { CheckCircle as CheckCircleIcon, Delete as DeleteIcon, UploadFileOutlined as FileIcon } from '@mui/icons-material';
import { IconButton, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import storeActions from '@northern.tech/store/actions';
import { useAppDispatch } from '@northern.tech/store/store';

const { setSnackbar } = storeActions;

const useStyles = makeStyles()(theme => ({
  dropzone: { ['&.dropzone']: { padding: theme.spacing(3, 2) } },
  selection: {
    display: 'grid',
    gridTemplateColumns: 'max-content 1fr max-content max-content',
    alignItems: 'center',
    columnGap: theme.spacing(2),
    paddingLeft: theme.spacing(2),
    paddingRight: theme.spacing(2)
  }
}));

interface FileUploadProps {
  enableContentReading?: boolean;
  fileNameSelection?: string;
  /** indicates whether the selected file passed the consuming component's validation - shows a success indicator if it did */
  isValid?: boolean;
  onFileChange: (content?: string) => void;
  onFileSelect?: (file?: File) => void;
  placeholder: ReactNode;
  style?: CSSProperties;
}

export const FileUpload = ({
  enableContentReading = true,
  fileNameSelection,
  isValid,
  onFileChange,
  onFileSelect = () => undefined,
  placeholder,
  style = {}
}: FileUploadProps) => {
  const [filename, setFilename] = useState(fileNameSelection);
  const { classes } = useStyles();
  const dispatch = useAppDispatch();

  const onDrop = (acceptedFiles, rejectedFiles) => {
    if (acceptedFiles.length) {
      if (enableContentReading) {
        const reader = new FileReader();
        reader.readAsBinaryString(acceptedFiles[0]);
        reader.onload = () => {
          const str = (reader.result as string).replace(/\n|\r/g, '\n');
          onFileChange(str);
        };
        reader.onerror = error => {
          console.log('Error: ', error);
          setFilename(undefined);
        };
      }
      setFilename(acceptedFiles[0].name);
      onFileSelect(acceptedFiles[0]);
    }
    if (rejectedFiles.length) {
      dispatch(setSnackbar(`File '${rejectedFiles[0].name}' was rejected.`));
    }
  };

  const onClear = () => {
    onFileChange();
    onFileSelect();
    setFilename(undefined);
  };

  return filename ? (
    <div className={classes.selection} style={style}>
      <FileIcon color="primary" />
      <Typography component="div" variant="subtitle2">
        {filename}
      </Typography>
      <IconButton aria-label="remove the selected file" onClick={onClear} size="small">
        <DeleteIcon />
      </IconButton>
      {isValid ? <CheckCircleIcon color="success" titleAccess="the selected file was accepted" /> : <span />}
    </div>
  ) : (
    <div style={style}>
      <Dropzone activeClassName="active" rejectClassName="active" multiple={false} onDrop={onDrop}>
        {({ getRootProps, getInputProps }) => (
          <div {...getRootProps()} className={`dropzone onboard dashboard-placeholder flexbox centered ${classes.dropzone}`}>
            <input {...getInputProps()} />
            <FileIcon color="primary" />
            <Typography className="margin-left-small" component="div" variant="subtitle2">
              {placeholder}
            </Typography>
          </div>
        )}
      </Dropzone>
    </div>
  );
};

export default FileUpload;
