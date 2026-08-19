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
import type { BaseSyntheticEvent } from 'react';
import { useCallback, useEffect, useRef, useState } from 'react';
import Dropzone from 'react-dropzone';
import { FormProvider, useForm } from 'react-hook-form';
import { useSelector } from 'react-redux';

import { CloudUpload } from '@mui/icons-material';
import { Button, DialogActions, DialogContent } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { InputErrorNotification } from '@northern.tech/common-ui/InputErrorNotification';
import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';
import { getDeviceTypes } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { createArtifact, uploadArtifact } from '@northern.tech/store/thunks';

import Tracking from '../../../tracking';
import ArtifactInformationForm, { steps } from './ArtifactInformationForm';
import { FileInformation } from './FileInformation';

type Update = {
  destination: string;
  deviceTypes: string[];
  fileSystem: string;
  name: string;
  softwareName: string;
  softwareVersion: string;
};

const useStyles = makeStyles()(theme => ({
  dropzone: { ['&.dropzone']: { padding: theme.spacing(4) } }
}));

const commonExtensions = ['zip', 'txt', 'tar', 'html', 'tar.gzip', 'gzip'];
const shortenFileName = name => {
  const extension = commonExtensions.find(extension => name.endsWith(extension));
  if (extension) {
    const dotIndex = name.lastIndexOf(`.${extension}`);
    return name.substring(0, dotIndex);
  }
  return name;
};

const singleFileLimit = 256 * 1024 ** 2; //256MiB
const menderFileLimit = 10 * 1024 ** 3; //10GiB
const reFilename = new RegExp(/^[\w\-.,]+$/);

export const isMenderArtifact = (name: string): boolean => name.endsWith('.mender');

const validateFile = ({ name, size }: File): string => {
  if (!reFilename.test(name)) {
    return 'Only letters, digits and characters in the set ".,_-" are allowed in the filename.';
  } else if (isMenderArtifact(name) && size > menderFileLimit) {
    return 'Only artifacts smaller than 10GiB are supported.';
  } else if (!isMenderArtifact(name) && size > singleFileLimit) {
    return 'Artifact generation is only supported for files smaller than 256MiB.';
  }
  return '';
};

export const ArtifactUpload = ({ onFileSelect }: { onFileSelect: (file?: File) => void }) => {
  const { classes } = useStyles();
  const [errorMessage, setErrorMessage] = useState<string>('');

  const onDrop = acceptedFiles => {
    if (acceptedFiles.length !== 1) {
      onFileSelect(undefined);
      setErrorMessage('The selected file is not supported.');
      return;
    }
    const [file] = acceptedFiles;
    const validationError = validateFile(file);
    if (validationError) {
      onFileSelect(undefined);
      setErrorMessage(validationError);
      return;
    }
    setErrorMessage('');
    onFileSelect(file);
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
          <div {...getRootProps({ className: `fadeIn onboard dropzone ${classes.dropzone}` })}>
            <input {...getInputProps()} />
            <CloudUpload fontSize="large" className="muted" />
            <div>Drag and drop here or browse to upload</div>
          </div>
        )}
      </Dropzone>
      <InputErrorNotification className="flexbox centered" content={errorMessage} />
    </>
  );
};

const defaultFileSystem = 'rootfs-image';
const defaultVersion = '1.0.0';
const lastStep = steps.length - 1;

const getDefaultValues = (file?: File): Update => {
  const name = file ? shortenFileName(file.name) : '';
  return {
    destination: '',
    deviceTypes: [],
    fileSystem: defaultFileSystem,
    name,
    softwareName: name.replace('.', '-'),
    softwareVersion: defaultVersion
  };
};

export const AddArtifactDialog = ({ onCancel, onUploadStarted, selectedFile }) => {
  const [activeStep, setActiveStep] = useState<number>(0);
  const [file, setFile] = useState<File | undefined>(selectedFile);
  const shouldAdvance = useRef<boolean>(false);

  const deviceTypes = useSelector(getDeviceTypes);
  const dispatch = useAppDispatch();
  const methods = useForm<Update>({ defaultValues: getDefaultValues(selectedFile) });
  const { getValues, handleSubmit, reset } = methods;

  useEffect(() => {
    setFile(selectedFile);
  }, [selectedFile]);

  useEffect(() => {
    setActiveStep(0);
    reset(getDefaultValues(file));
  }, [file, reset]);

  const addArtifact = useCallback(
    (meta, file, type = 'upload') => {
      onUploadStarted();
      const upload = type === 'create' ? dispatch(createArtifact({ meta, file })) : dispatch(uploadArtifact({ meta, file }));
      // track in GA
      return upload.unwrap().then(() => Tracking.event({ category: 'artifacts', action: 'create' }));
    },
    [dispatch, onUploadStarted]
  );

  const onUploadClick = useCallback(() => addArtifact({ description: '' }, file, 'upload'), [addArtifact, file]);

  // the inputs of the current step are validated before this is called - which is why progressing is handled here only
  const onValid = useCallback(
    ({ destination, deviceTypes: selectedDeviceTypes, fileSystem, name, softwareName, softwareVersion }: Update) => {
      if (activeStep < lastStep) {
        shouldAdvance.current = true;
        return;
      }
      const meta = {
        description: '',
        device_types_compatible: selectedDeviceTypes,
        args: { dest_dir: destination, filename: file?.name, software_filesystem: fileSystem, software_name: softwareName, software_version: softwareVersion },
        name
      };
      addArtifact(meta, file, 'create');
    },
    [activeStep, addArtifact, file]
  );

  // clearing the submit state on step changes prevents the inputs of the following step from being validated on every keystroke
  const onStepChange = useCallback(
    (step: number) => {
      setActiveStep(step);
      reset(getValues(), { keepDefaultValues: true, keepTouched: true });
    },
    [getValues, reset]
  );

  // react-hook-form marks the form as submitted only once the submit handler returned, so the step can be advanced here only
  const onSubmitClick = useCallback(
    (event?: BaseSyntheticEvent) =>
      handleSubmit(onValid)(event).then(() => {
        if (!shouldAdvance.current) {
          return;
        }
        shouldAdvance.current = false;
        onStepChange(activeStep + 1);
      }),
    [activeStep, handleSubmit, onStepChange, onValid]
  );

  const onRemove = () => setFile(undefined);

  const isMender = !!file && isMenderArtifact(file.name);
  const type = isMender ? 'mender' : 'singleFile';
  const submitLabel = isMender ? 'Upload artifact' : activeStep < lastStep ? 'Next' : 'Upload';

  return (
    <BaseDialog open title="Upload an Artifact" fullWidth maxWidth="sm" onClose={onCancel}>
      <DialogContent>
        {!file ? (
          <ArtifactUpload onFileSelect={setFile} />
        ) : isMender ? (
          <FileInformation file={file} onRemove={onRemove} type={type} />
        ) : (
          <FormProvider {...methods}>
            <form noValidate onSubmit={onSubmitClick}>
              <ArtifactInformationForm activeStep={activeStep} deviceTypes={deviceTypes} file={file} onRemove={onRemove} type={type} />
            </form>
          </FormProvider>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={onCancel}>Cancel</Button>
        {!!activeStep && (
          <Button color="info" variant="outlined" onClick={() => onStepChange(activeStep - 1)}>
            Back
          </Button>
        )}
        {!!file && (
          <Button variant="contained" onClick={isMender ? onUploadClick : onSubmitClick}>
            {submitLabel}
          </Button>
        )}
      </DialogActions>
    </BaseDialog>
  );
};

export default AddArtifactDialog;
