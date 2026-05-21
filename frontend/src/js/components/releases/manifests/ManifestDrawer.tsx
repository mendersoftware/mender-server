// Copyright 2026 Northern.tech AS
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
import { useState } from 'react';
import Dropzone from 'react-dropzone';
import { FormProvider, useForm } from 'react-hook-form';
import { useSelector } from 'react-redux';

import { Close, UploadFile } from '@mui/icons-material';
import { Alert, Button, IconButton, InputAdornment, TextField, Typography, alpha } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { mdiCloudUploadOutline } from '@mdi/js';
import BaseDrawer from '@northern.tech/common-ui/BaseDrawer';
import ChipSelect from '@northern.tech/common-ui/ChipSelect';
import { ContentSection } from '@northern.tech/common-ui/ContentSection';
import { DOCSTIPS, DocsTextLink } from '@northern.tech/common-ui/DocsLink';
import Link from '@northern.tech/common-ui/Link';
import MaterialDesignIcon from '@northern.tech/common-ui/MaterialDesignIcon';
import { TwoColumnData } from '@northern.tech/common-ui/TwoColumnData';
import TextInput from '@northern.tech/common-ui/forms/TextInput';
import { getManifestTags } from '@northern.tech/store/releasesSlice/selectors';
import { generateManifest, uploadManifest } from '@northern.tech/store/releasesSlice/thunks';
import { useAppDispatch } from '@northern.tech/store/store';
import type { ManifestContent } from '@northern.tech/types/MenderTypes';
import { parse } from 'yaml';
import * as z from 'zod';

import { isMenderArtifact } from '../dialogs/AddArtifact';
import { ComponentTypesTable } from './ManifestDetails';

const MenderManifestSizeLimit = 1024 ** 2;
const genericUploadErrorMessage = 'The Manifest could not be uploaded. Check that the .mender file contains a valid .yaml Manifest';

const ManifestContentSchema = z.object({
  api_version: z.string(),
  kind: z.literal('manifest'),
  name: z.string(),
  system_types_compatible: z.array(z.string()),
  component_types: z.record(
    z.string(),
    z.object({
      artifact_name: z.string().optional(),
      artifact_path: z.string().optional(),
      update_strategy: z.object({ order: z.number() })
    })
  )
}) satisfies z.ZodType<ManifestContent>;

const useStyles = makeStyles()(theme => ({
  dropzone: {
    ['&.dropzone']: {
      '&:hover, &:active, &.active': {
        backgroundColor: alpha(theme.palette.primary.light, 0.08),
        borderColor: theme.palette.primary.main
      },
      maxWidth: '850px',
      padding: theme.spacing(4),
      transition: theme.transitions.create(['background-color', 'border-color'])
    }
  },
  input: {
    maxWidth: '400px'
  }
}));
const isYaml = (filename: string) => filename.endsWith('.yaml') || filename.endsWith('.yml');
const validateFile = async (file: File) => {
  const { size, name } = file;
  if (isMenderArtifact(name) && size > MenderManifestSizeLimit) {
    return { ok: false, message: 'Manifest Artifacts must be smaller than 1MB' };
  } else if (isYaml(name)) {
    const text = await file.text();
    let parsed;
    try {
      parsed = parse(text);
    } catch (e) {
      return { ok: false, message: e instanceof Error ? e.message : String(e) };
    }
    const result = ManifestContentSchema.safeParse(parsed);
    if (!result.success) {
      return { ok: false, message: z.prettifyError(result.error) };
    }
    return { ok: true, data: result.data };
  }
  return { ok: true };
};
type AddManifestDrawerProps = {
  onClose: () => void;
  open: boolean;
};

type ManifestFormValues = {
  description: string;
  tags: string[];
};

export const AddManifestDrawer = ({ onClose, open }: AddManifestDrawerProps) => {
  const [errorMessage, setErrorMessage] = useState('');
  const [backendErrorMessage, setBackendErrorMessage] = useState('');
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [parsedManifest, setParsedManifest] = useState<ManifestContent | null>(null);
  const { classes } = useStyles();
  const dispatch = useAppDispatch();
  const existingTags = useSelector(getManifestTags);

  const defaultValues: ManifestFormValues = { tags: [], description: '' };
  const methods = useForm<ManifestFormValues>({ mode: 'onChange', defaultValues });
  const { handleSubmit, reset } = methods;
  const onReset = () => {
    setSelectedFile(null);
    setErrorMessage('');
    setBackendErrorMessage('');
    setParsedManifest(null);
  };
  const onDrawerClose = () => {
    reset();
    onReset();
    onClose();
  };
  const onDrop = async (acceptedFiles: File[]) => {
    if (acceptedFiles.length === 1) {
      const { ok, message, data } = await validateFile(acceptedFiles[0]);
      setSelectedFile(acceptedFiles[0]);
      if (ok) {
        setParsedManifest(data ?? null);
        setErrorMessage('');
      } else {
        setParsedManifest(null);
        if (message) {
          setErrorMessage(message);
        }
      }
    } else {
      setParsedManifest(null);
      setErrorMessage('The selected file is not supported.');
    }
  };
  const onSubmit = (formData: ManifestFormValues) => {
    if (!selectedFile) return;
    const { name } = selectedFile;
    let action = uploadManifest({ file: selectedFile, meta: formData });
    if (isYaml(name)) {
      action = generateManifest({ file: selectedFile, meta: formData });
    }
    dispatch(action)
      .unwrap()
      .then(() => onDrawerClose())
      .catch(e => setBackendErrorMessage(typeof e === 'string' ? e : (e?.message ?? genericUploadErrorMessage)));
  };

  return (
    <BaseDrawer open={open} size="md" onClose={onDrawerClose} slotProps={{ header: { title: 'Upload a Manifest' } }}>
      <FormProvider {...methods}>
        <form noValidate onSubmit={handleSubmit(onSubmit)}>
          {selectedFile ? (
            <TextField
              fullWidth
              label="File"
              className={classes.input}
              value={selectedFile.name}
              error={!!errorMessage}
              helperText={
                errorMessage &&
                errorMessage.split('\n').map((line, index) => (
                  <span key={index} className="flexbox">
                    {line}
                  </span>
                ))
              }
              slotProps={{
                input: {
                  readOnly: true,
                  startAdornment: (
                    <InputAdornment position="start">
                      <UploadFile />
                    </InputAdornment>
                  ),
                  endAdornment: (
                    <InputAdornment position="end">
                      <IconButton onClick={onReset} size="small" aria-label="Remove file">
                        <Close fontSize="small" />
                      </IconButton>
                    </InputAdornment>
                  )
                }
              }}
            />
          ) : (
            <Dropzone
              multiple={false}
              onDrop={onDrop}
              accept={{
                'application/octet-stream': ['.mender'],
                'application/yaml': ['.yaml', '.yml']
              }}
            >
              {({ getRootProps, getInputProps, isDragActive }) => (
                <div {...getRootProps({ className: `dropzone onboard ${classes.dropzone}${isDragActive ? ' active' : ''}` })}>
                  <input {...getInputProps()} />
                  <div className="flexbox centered">
                    <MaterialDesignIcon fontSize="medium" path={mdiCloudUploadOutline} className="margin-right-x-small" />
                    <Typography>
                      Drag here or <Link>browse</Link> to upload a file
                    </Typography>
                  </div>
                </div>
              )}
            </Dropzone>
          )}
          <Typography className="margin-top-x-small" color="textSecondary" variant="body2">
            Upload a Manifest (.yaml), or an Artifact (.mender) containing a Manifest. <DocsTextLink id={DOCSTIPS.orchestratorManifest.id} />
          </Typography>
          {parsedManifest?.name && (
            <TwoColumnData
              className="margin-top-medium margin-bottom-medium"
              data={{
                Name: parsedManifest.name,
                'Compatible types': parsedManifest.system_types_compatible?.join(', ')
              }}
            />
          )}

          <ContentSection title="Notes">
            <TextInput id="description" hint="Add notes here" InputLabelProps={{ shrink: true }} InputProps={{ multiline: true, maxRows: 4 }} />
          </ContentSection>
          <ContentSection title="Tags">
            <ChipSelect className={classes.input} options={existingTags} name="tags" placeholder="Add Tags" forcePopupIcon={existingTags.length !== 0} />
          </ContentSection>
          {parsedManifest?.component_types && <ComponentTypesTable componentTypes={parsedManifest.component_types} />}
          {backendErrorMessage && (
            <Alert slotProps={{ message: { className: 'capitalized-start' } }} className="margin-top-medium capitalized-start" severity="error">
              {backendErrorMessage}
            </Alert>
          )}
          <div className="margin-top">
            <Button onClick={onClose}>Cancel</Button>
            <Button type="submit" className="margin-left-small" variant="contained" disabled={!!errorMessage || !selectedFile}>
              Upload
            </Button>
          </div>
        </form>
      </FormProvider>
    </BaseDrawer>
  );
};
