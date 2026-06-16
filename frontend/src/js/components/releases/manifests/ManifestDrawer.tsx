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
import { useCallback, useEffect, useState } from 'react';
import Dropzone from 'react-dropzone';
import { FormProvider, useForm } from 'react-hook-form';
import { useSelector } from 'react-redux';

import { Close, UploadFile } from '@mui/icons-material';
import { ExpandLess as ExpandLessIcon, ExpandMore as ExpandMoreIcon } from '@mui/icons-material';
import { Alert, Button, FormControlLabel, IconButton, InputAdornment, Radio, RadioGroup, TextField, Typography, alpha } from '@mui/material';
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
import { checkReleasesExistence, generateManifest, getManifest, getSoftware, uploadManifest } from '@northern.tech/store/releasesSlice/thunks';
import { useAppDispatch } from '@northern.tech/store/store';
import { getExistingSoftwareTags } from '@northern.tech/store/thunks';
import type { ManifestContent, Software } from '@northern.tech/types/MenderTypes';
import { parseAllDocuments, stringify } from 'yaml';
import * as z from 'zod';

import { SoftwareArtifactFilter } from '../../deployments/deployment-wizard/ReleaseArtifactFilter';
import { isMenderArtifact } from '../dialogs/AddArtifact';
import { ComponentTypesTable } from './ComponentTypesTable';

const MenderManifestSizeLimit = 1024 ** 2;
const genericUploadErrorMessage = 'The Manifest could not be uploaded. Check that the .mender file contains a valid .yaml Manifest';

const ManifestContentSchema = z.object({
  api_version: z.string(),
  kind: z.literal('manifest'),
  name: z.string(),
  system_types_compatible: z.array(z.string()),
  component_types: z.record(
    z.string(),
    z
      .object({
        artifact_name: z.string().optional(),
        artifact_path: z.string().optional(),
        update_strategy: z.object({ order: z.number() })
      })
      .refine(data => data.artifact_name || data.artifact_path, {
        message: 'At least one of artifact_name or artifact_path must be provided'
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
    const docs = parseAllDocuments(text);
    if (docs.length !== 1) {
      return { ok: false, message: 'Only single-document .yaml Manifests are supported' };
    }
    const doc = docs[0];
    if (doc.errors.length) {
      return { ok: false, message: doc.errors.map(e => e.message).join('\n') };
    }
    const parsed = doc.toJS();
    const result = ManifestContentSchema.safeParse(parsed);
    if (!result.success) {
      return { ok: false, message: z.prettifyError(result.error) };
    }
    return { ok: true, data: result.data };
  }
  return { ok: true };
};
type AddManifestDrawerProps = {
  copyFromManifest?: string;
  onClose: () => void;
  open: boolean;
};

type ManifestFormValues = {
  description: string;
  name: string;
  tags: string[];
};

type Mode = 'upload' | 'copy';
type ModeInfo = { key: Mode; title: string };
const modes: Record<Mode, ModeInfo> = {
  upload: { key: 'upload', title: 'Upload' },
  copy: { key: 'copy', title: 'Copy Existing' }
};

export const AddManifestDrawer = ({ copyFromManifest, onClose, open }: AddManifestDrawerProps) => {
  const [mode, setMode] = useState<Mode>(modes.upload.key);
  const [copyReleaseOpen, setCopyReleaseOpen] = useState<boolean>(false);
  const [errorMessage, setErrorMessage] = useState('');
  const [backendErrorMessage, setBackendErrorMessage] = useState('');
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [parsedManifest, setParsedManifest] = useState<ManifestContent | null>(null);
  const [existingReleases, setExistingReleases] = useState<Record<string, boolean>>({});
  const { classes } = useStyles();
  const dispatch = useAppDispatch();
  const existingTags = useSelector(getManifestTags);
  useEffect(() => {
    if (!parsedManifest?.component_types) {
      return;
    }
    const artifactNames = Object.values(parsedManifest.component_types)
      .filter(({ artifact_name }) => !!artifact_name)
      .map(({ artifact_name }) => artifact_name);
    const uniqueNames = [...new Set(artifactNames)];
    dispatch(checkReleasesExistence(uniqueNames)).unwrap().then(setExistingReleases);
  }, [dispatch, parsedManifest?.component_types]);

  const defaultValues: ManifestFormValues = { tags: [], description: '', name: '' };
  const methods = useForm<ManifestFormValues>({ mode: 'onChange', defaultValues });
  const { handleSubmit, reset, setValue } = methods;

  useEffect(() => {
    dispatch(getSoftware());
    dispatch(getExistingSoftwareTags());
  }, [dispatch]);

  const loadCopyManifest = useCallback(
    (name: string) =>
      dispatch(getManifest(name))
        .unwrap()
        .then(manifest => {
          setParsedManifest(manifest?.manifest || null);
          setValue('name', `${name}(copy)`);
        }),
    [dispatch, setValue]
  );
  const onSelectCopy = (item: Software) => loadCopyManifest(item.name);

  useEffect(() => {
    if (open && copyFromManifest) {
      setMode(modes.copy.key);
      loadCopyManifest(copyFromManifest);
    }
  }, [open, copyFromManifest, loadCopyManifest]);

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
    const file =
      mode === modes.copy.key && parsedManifest
        ? new File([stringify({ ...parsedManifest, name: formData.name })], `${formData.name}.yaml`, { type: 'application/yaml' })
        : selectedFile;
    if (!file) return;
    const action = isYaml(file.name) ? generateManifest({ file, meta: formData }) : uploadManifest({ file, meta: formData });
    dispatch(action)
      .unwrap()
      .then(() => onDrawerClose())
      .catch(e => setBackendErrorMessage(typeof e === 'string' ? e : (e?.message ?? genericUploadErrorMessage)));
  };

  const onComponentTypesChange = (component_types: ManifestContent['component_types']) =>
    setParsedManifest(current => (current ? { ...current, component_types } : current));

  const fileUploadInput = (
    <>
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
    </>
  );
  return (
    <BaseDrawer open={open} size="md" onClose={onDrawerClose} slotProps={{ header: { title: 'Upload a Manifest' } }}>
      <RadioGroup row className="margin-bottom-medium" value={mode} onChange={(_, newMode) => setMode(newMode as Mode)}>
        {Object.values(modes).map(({ key, title }) => (
          <FormControlLabel value={key} control={<Radio />} label={title} key={key} />
        ))}
      </RadioGroup>
      <FormProvider {...methods}>
        <form noValidate onSubmit={handleSubmit(onSubmit)}>
          {mode === modes.upload.key ? (
            fileUploadInput
          ) : (
            <>
              {copyReleaseOpen ? (
                <SoftwareArtifactFilter
                  selectedSoftware={parsedManifest?.name}
                  kind="manifest"
                  open={copyReleaseOpen}
                  onClose={() => setCopyReleaseOpen(false)}
                  onSelect={item => onSelectCopy(item)}
                />
              ) : (
                <>
                  <Button
                    size="large"
                    color="neutral"
                    variant="outlined"
                    endIcon={copyReleaseOpen ? <ExpandLessIcon /> : <ExpandMoreIcon />}
                    onClick={() => setCopyReleaseOpen(!copyReleaseOpen)}
                  >
                    <span className="text-overflow">{parsedManifest?.name ?? 'Select a Manifest'}</span>
                  </Button>
                  <Typography className="margin-top-x-small" color="textSecondary" variant="body2">
                    Use an existing Manifest as a template. You can edit the details before saving it as new.
                  </Typography>
                  <TwoColumnData
                    className="margin-top-medium margin-bottom-medium"
                    data={{
                      'Compatible types': parsedManifest?.system_types_compatible?.join(', ') || '-'
                    }}
                  />
                </>
              )}
              <ContentSection className="margin-bottom-medium" title="Name">
                <TextInput id="name" required validations="trim,isLength:1" />
              </ContentSection>
            </>
          )}

          {mode === modes.upload.key && parsedManifest?.name && (
            <TwoColumnData
              className="margin-top-medium margin-bottom-medium"
              data={{
                Name: parsedManifest.name,
                'Compatible types': parsedManifest.system_types_compatible?.join(', ')
              }}
            />
          )}

          <ContentSection className="margin-bottom-medium" title="Notes">
            <TextInput id="description" hint="Add notes here" InputLabelProps={{ shrink: true }} InputProps={{ multiline: true, maxRows: 4 }} />
          </ContentSection>
          <ContentSection className="margin-bottom-medium" title="Tags">
            <ChipSelect className={classes.input} options={existingTags} name="tags" placeholder="Add Tags" forcePopupIcon={existingTags.length !== 0} />
          </ContentSection>
          {parsedManifest?.component_types && (
            <ComponentTypesTable
              componentTypes={parsedManifest.component_types}
              isEditable={mode === modes.copy.key}
              onChange={onComponentTypesChange}
              existingReleases={existingReleases}
            />
          )}
          {backendErrorMessage && (
            <Alert slotProps={{ message: { className: 'capitalized-start' } }} className="margin-top-medium capitalized-start" severity="error">
              {backendErrorMessage}
            </Alert>
          )}
          <div className="margin-top">
            <Button onClick={onDrawerClose}>Cancel</Button>
            <Button
              type="submit"
              className="margin-left-small"
              variant="contained"
              disabled={!!errorMessage || (mode === modes.copy.key ? !parsedManifest : !selectedFile)}
            >
              Upload
            </Button>
          </div>
        </form>
      </FormProvider>
    </BaseDrawer>
  );
};
