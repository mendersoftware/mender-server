// Copyright 2022 Northern.tech AS
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
import React, { useCallback, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { Cancel as CancelIcon } from '@mui/icons-material';
import { Drawer, IconButton, LinearProgress, Tooltip, drawerClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import pluralize from 'pluralize';

import { cancelFileUpload } from '../actions/releaseActions';
import { FileSize } from '../helpers';

const useStyles = makeStyles()(theme => ({
  progress: {
    backgroundColor: theme.palette.grey[600],
    gridColumn: 1,
    margin: '15px 0'
  },
  drawer: {
    [`.${drawerClasses.paper}`]: {
      maxWidth: 'initial'
    }
  },
  progressBarContainer: {
    position: 'fixed',
    display: 'grid',
    bottom: 0,
    gridTemplateRows: 'min-content',
    height: 'max-content',
    padding: `0 ${theme.spacing(2)}px`,
    overflow: 'hidden',
    width: '100%',
    zIindex: 10000,
    transition: 'all 0.2s ease-in-out'
  },
  progressContainer: {
    backgroundColor: theme.palette.background.default,
    borderColor: theme.palette.grey[300],
    color: theme.palette.grey[600],
    display: 'grid',
    gridTemplateRows: '30px 30px',
    gridTemplateColumns: '1fr 60px',
    columnGap: theme.spacing(2),
    ['> .MuiIconButton-root']: {
      gridColumn: 2,
      gridRowStart: 1,
      gridRowEnd: 2,
      height: 60
    }
  }
}));

const UploadProgressBar = ({ classes, upload, uploadId }) => {
  const { name, size, uploadProgress } = upload;
  const dispatch = useDispatch();
  const onCancelClick = useCallback(() => dispatch(cancelFileUpload(uploadId)), [dispatch, uploadId]);
  return (
    <div className={classes.progressContainer}>
      <div className="info flexbox centered">
        {Math.round(uploadProgress)}% - {name} (<FileSize fileSize={size} />)
      </div>
      <LinearProgress className={classes.progress} variant="determinate" value={uploadProgress} />
      <Tooltip title="Abort" placement="top">
        <IconButton onClick={onCancelClick} size="large">
          <CancelIcon />
        </IconButton>
      </Tooltip>
    </div>
  );
};

const Uploads = () => {
  const [isHovering, setIsHovering] = useState(false);
  const { classes } = useStyles();

  const uploads = useSelector(state => state.app.uploadsById);

  const isUploading = !!Object.keys(uploads).length;
  const uploadProgress = Object.values(uploads).reduce((accu, item, currentIndex, items) => {
    accu += item.uploadProgress;
    if (currentIndex === items.length - 1) {
      return accu / (currentIndex + 1);
    }
    return accu;
  }, 0);

  const onClose = () => setIsHovering(false);
  const onMouseOver = () => setIsHovering(true);

  return (
    <div className={classes.progressBarContainer} onMouseEnter={onMouseOver}>
      {isUploading && !isHovering && <LinearProgress className={classes.progress} variant="determinate" value={uploadProgress} />}
      <Drawer anchor="bottom" className={classes.drawer} open={isUploading && isHovering} onClose={onClose}>
        <h3>{pluralize('Upload', Object.keys(uploads).length)} in progress</h3>
        {Object.entries(uploads).map(([uploadId, upload]) => (
          <UploadProgressBar classes={classes} key={uploadId} upload={upload} uploadId={uploadId} />
        ))}
      </Drawer>
    </div>
  );
};

export default Uploads;
