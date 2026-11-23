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
import { useCallback, useState } from 'react';
import { useSelector } from 'react-redux';

import { Cancel as CancelIcon } from '@mui/icons-material';
import { Drawer, IconButton, LinearProgress, Tooltip, Typography, drawerClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import FileSize from '@northern.tech/common-ui/FileSize';
import { getIsUploading, getUploads } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { cancelFileUpload } from '@northern.tech/store/thunks';
import pluralize from 'pluralize';

const useStyles = makeStyles()(() => ({
  drawer: {
    [`.${drawerClasses.paper}`]: {
      maxWidth: 'initial'
    }
  },
  progressBarContainer: {
    position: 'fixed',
    bottom: 0,
    width: '100%',
    zIndex: 10000
  }
}));

const UploadProgressBar = ({ upload, uploadId }) => {
  const { name, size, progress } = upload;
  const dispatch = useAppDispatch();
  const onCancelClick = useCallback(() => dispatch(cancelFileUpload(uploadId)), [dispatch, uploadId]);
  return (
    <div className="flexbox align-items-center gap-small">
      <div className="flexbox column full-width gap-x-small">
        <Typography variant="body2" color="text.secondary">
          {Math.round(progress)}% - {name} (<FileSize component="span" fileSize={size} />)
        </Typography>
        <LinearProgress variant="determinate" value={progress} />
      </div>
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

  const uploads = useSelector(getUploads);

  const isUploading = useSelector(getIsUploading);
  const uploadProgress = Object.values(uploads).reduce((accu, item, currentIndex, items) => {
    accu += item.progress;
    if (currentIndex === items.length - 1) {
      return accu / (currentIndex + 1);
    }
    return accu;
  }, 0);

  const onClose = () => setIsHovering(false);
  const onMouseOver = () => setIsHovering(true);

  return (
    <div className={classes.progressBarContainer} onMouseEnter={onMouseOver}>
      {isUploading && !isHovering && <LinearProgress className="margin-top-small margin-bottom-small" variant="determinate" value={uploadProgress} />}
      <Drawer anchor="bottom" className={classes.drawer} open={isUploading && isHovering} onClose={onClose}>
        <div className="flexbox column gap-small">
          <Typography variant="h5">{pluralize('Upload', Object.keys(uploads).length)} in progress</Typography>
          {Object.entries(uploads).map(([uploadId, upload]) => (
            <UploadProgressBar key={uploadId} upload={upload} uploadId={uploadId} />
          ))}
        </div>
      </Drawer>
    </div>
  );
};

export default Uploads;
