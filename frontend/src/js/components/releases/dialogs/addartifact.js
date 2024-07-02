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
import React, { useCallback, useEffect, useRef, useState } from 'react';
import Dropzone from 'react-dropzone';
import { useDispatch, useSelector } from 'react-redux';

import { CloudUpload, Delete as DeleteIcon, InsertDriveFile as InsertDriveFileIcon } from '@mui/icons-material';
import { Button, Dialog, DialogActions, DialogContent, DialogTitle, Divider, IconButton } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { setSnackbar } from '../../../actions/appActions';
import { createArtifact, uploadArtifact } from '../../../actions/releaseActions';
import { FileSize, unionizeStrings } from '../../../helpers';
import { getDeviceTypes } from '../../../selectors';
import Tracking from '../../../tracking';
import useWindowSize from '../../../utils/resizehook';
import { HELPTOOLTIPS, MenderHelpTooltip } from '../../helptips/helptooltips';
import ArtifactInformationForm from './artifactinformationform';
import ArtifactUploadConfirmation from './artifactupload';

const reFilename = new RegExp(/^[a-z0-9.,_-]+$/i);

const useStyles = makeStyles()(theme => ({
  dropzone: { ['&.dropzone']: { padding: theme.spacing(4) } },
  fileInfo: {
    alignItems: 'center',
    columnGap: theme.spacing(4),
    display: 'grid',
    gridTemplateColumns: 'max-content 1fr max-content max-content',
    marginBottom: theme.spacing(2),
    marginRight: theme.spacing(4)
  },
  fileSizeWrapper: { marginTop: 5 }
}));

const uploadTypes = {
  mender: {
    key: 'mender',
    component: ArtifactUploadConfirmation
  },
  singleFile: {
    key: 'singleFile',
    component: ArtifactInformationForm
  }
};

const fileInformationContent = {
  mender: {
    title: 'Mender Artifact',
    icon: InsertDriveFileIcon,
    infoId: 'menderArtifactUpload'
  },
  singleFile: {
    title: 'Single File',
    icon: InsertDriveFileIcon,
    infoId: 'singleFileUpload'
  }
};

export const FileInformation = ({ file, type, onRemove }) => {
  const { classes } = useStyles();
  if (!file) {
    return <div />;
  }
  const { icon: Icon, infoId, title } = fileInformationContent[type];
  return (
    <>
      <h4>Selected {title}</h4>
      <div className={classes.fileInfo}>
        <Icon size="large" />
        <div className="flexbox column">
          <div>{file.name}</div>
          <div className={`muted ${classes.fileSizeWrapper}`}>
            <FileSize fileSize={file.size} />
          </div>
        </div>
        <IconButton size="large" onClick={onRemove}>
          <DeleteIcon />
        </IconButton>
        <MenderHelpTooltip id={HELPTOOLTIPS[infoId].id} />
      </div>
      <Divider className="margin-right-large" />
    </>
  );
};

const commonExtensions = ['zip', 'txt', 'tar', 'html', 'tar.gzip', 'gzip'];
const shortenFileName = name => {
  const extension = commonExtensions.find(extension => name.endsWith(extension));
  if (extension) {
    const dotIndex = name.lastIndexOf(`.${extension}`);
    return name.substring(0, dotIndex);
  }
  return name;
};

export const ArtifactUpload = ({ setSnackbar, updateCreation }) => {
  const onboardingAnchor = useRef();
  const { classes } = useStyles();
  // eslint-disable-next-line no-unused-vars
  const size = useWindowSize();

  const onDrop = acceptedFiles => {
    const emptyFileInfo = { file: undefined, name: '', type: uploadTypes.mender.key };
    if (acceptedFiles.length === 1) {
      if (!reFilename.test(acceptedFiles[0].name)) {
        updateCreation(emptyFileInfo);
        setSnackbar('Only letters, digits and characters in the set ".,_-" are allowed in the filename.', null);
      } else {
        const { name } = acceptedFiles[0];
        updateCreation({
          file: acceptedFiles[0],
          name: shortenFileName(name),
          type: name.endsWith('.mender') ? uploadTypes.mender.key : uploadTypes.singleFile.key
        });
      }
    } else {
      updateCreation(emptyFileInfo);
      setSnackbar('The selected file is not supported.', null);
    }
  };

  return (
    <>
      <div className="flexbox column centered margin">
        Upload a premade Mender Artifact
        <p className="muted">OR</p>
        Upload a file to generate a single file application update Artifact
      </div>
      <Dropzone multiple={false} onDrop={onDrop}>
        {({ getRootProps, getInputProps }) => (
          <div {...getRootProps({ className: `fadeIn onboard dropzone ${classes.dropzone}` })} ref={onboardingAnchor}>
            <input {...getInputProps()} />
            <CloudUpload fontSize="large" className="muted" />
            <div>
              Drag and drop here or <b>browse</b> to upload
            </div>
          </div>
        )}
      </Dropzone>
    </>
  );
};

export const AddArtifactDialog = ({ onCancel, onUploadStarted, releases, selectedFile }) => {
  const [activeStep, setActiveStep] = useState(0);
  const [creation, setCreation] = useState({
    customDeviceTypes: '',
    destination: '',
    file: undefined,
    fileSystem: 'rootfs-image',
    finalStep: false,
    isValid: false,
    isValidDestination: false,
    name: '',
    selectedDeviceTypes: [],
    softwareName: '',
    softwareVersion: '',
    type: uploadTypes.mender.key
  });

  const deviceTypes = useSelector(getDeviceTypes);
  const dispatch = useDispatch();

  const onCreateArtifact = useCallback((meta, file) => dispatch(createArtifact(meta, file)), [dispatch]);
  const onSetSnackbar = useCallback((...args) => dispatch(setSnackbar(...args)), [dispatch]);
  const onUploadArtifact = useCallback((meta, file) => dispatch(uploadArtifact(meta, file)), [dispatch]);

  useEffect(() => {
    setCreation(current => ({ ...current, file: selectedFile }));
  }, [selectedFile]);

  const addArtifact = useCallback(
    (meta, file, type = 'upload') => {
      const upload = type === 'create' ? onCreateArtifact(meta, file) : onUploadArtifact(meta, file);
      onUploadStarted();
      // track in GA
      return upload.then(() => Tracking.event({ category: 'artifacts', action: 'create' }));
    },
    [onCreateArtifact, onUploadStarted, onUploadArtifact]
  );

  const onUpload = useCallback(() => {
    const { customDeviceTypes, destination, file, fileSystem, name, selectedDeviceTypes, softwareName, softwareVersion } = creation;
    const { name: filename = '' } = file;
    let meta = { description: '' };
    if (filename.endsWith('.mender')) {
      return addArtifact(meta, file, 'upload');
    }
    const otherDeviceTypes = customDeviceTypes.split(',');
    const deviceTypes = unionizeStrings(selectedDeviceTypes, otherDeviceTypes);
    meta = {
      ...meta,
      device_types_compatible: deviceTypes,
      args: { dest_dir: destination, filename, software_filesystem: fileSystem, software_name: softwareName, software_version: softwareVersion },
      name
    };
    return addArtifact(meta, file, 'create');
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [addArtifact, JSON.stringify(creation)]);

  const onUpdateCreation = useCallback(update => setCreation(current => ({ ...current, ...update })), []);

  const onNextClick = useCallback(() => {
    onUpdateCreation({ isValid: false });
    setActiveStep(activeStep + 1);
  }, [activeStep, onUpdateCreation]);

  const onRemove = () => onUpdateCreation({ file: undefined, isValid: false });

  const { file, finalStep, isValid, type } = creation;
  const { component: ComponentToShow } = uploadTypes[type];
  const commonProps = { releases, setSnackbar: onSetSnackbar, updateCreation: onUpdateCreation };

  return (
    <Dialog open={true} fullWidth={true} maxWidth="sm">
      <DialogTitle>Upload an Artifact</DialogTitle>
      <DialogContent className="dialog-content margin-top margin-left margin-right margin-bottom">
        {!file ? (
          <ArtifactUpload {...commonProps} />
        ) : (
          <ComponentToShow {...commonProps} activeStep={activeStep} creation={creation} deviceTypes={deviceTypes} onRemove={onRemove} />
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={onCancel}>Cancel</Button>
        {!!activeStep && <Button onClick={() => setActiveStep(activeStep - 1)}>Back</Button>}
        <div style={{ flexGrow: 1 }} />
        {file && (
          <Button variant="contained" color="primary" disabled={!isValid} onClick={() => (finalStep ? onUpload() : onNextClick())}>
            {finalStep ? 'Upload artifact' : 'Next'}
          </Button>
        )}
      </DialogActions>
    </Dialog>
  );
};

export default AddArtifactDialog;
