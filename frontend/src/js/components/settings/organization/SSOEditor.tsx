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
import { useCallback, useEffect, useState } from 'react';
import type { ChangeEvent } from 'react';

// material ui
import { UploadFile } from '@mui/icons-material';
import { Alert, Button, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import BaseDrawer from '@northern.tech/common-ui/BaseDrawer';
import { CodeEditor } from '@northern.tech/common-ui/CodeEditor';
import { JSON_METADATA_FORMAT, XML_METADATA_FORMAT } from '@northern.tech/store/constants';

const useStyles = makeStyles()(() => ({
  editor: {
    maxHeight: '500px'
  }
}));

const editorDescriptions = {
  [XML_METADATA_FORMAT]: 'Import your SAML metadata document as an .xml file, or paste and edit the XML configuration directly below.',
  [JSON_METADATA_FORMAT]: 'Import your OpenID Connect metadata document as a .json file, or paste and edit the JSON configuration directly below.'
};

export const SSOEditor = ({ ssoItem, config, fileContent, open, onClose, onSave, setFileContent }) => {
  const [isMetadataValid, setIsMetadataValid] = useState(false);
  const [saveError, setSaveError] = useState('');
  const { classes } = useStyles();
  useEffect(() => {
    setSaveError('');
    setIsMetadataValid(false);
    if (!fileContent) {
      return;
    }

    const parser = new DOMParser();
    let valid;
    switch (ssoItem.metadataFormat) {
      case JSON_METADATA_FORMAT:
        try {
          JSON.parse(fileContent);
          valid = true;
        } catch {
          valid = false;
        }
        break;
      case XML_METADATA_FORMAT:
      default:
        valid = !parser.parseFromString(fileContent, 'application/xml').getElementsByTagName('parsererror').length;
        break;
    }
    setIsMetadataValid(valid);
  }, [fileContent, ssoItem.metadataFormat]);

  const onCancelClick = useCallback(() => {
    setFileContent(config);
    setSaveError('');
    onClose();
  }, [config, setFileContent, onClose]);

  const onSubmitClick = async () => {
    setSaveError('');
    try {
      await onSave();
    } catch (error) {
      setSaveError(typeof error === 'string' ? error : 'There was an error saving the SSO configuration.');
    }
  };

  const onFileChange = async ({ target }: ChangeEvent<HTMLInputElement>) => {
    const [file] = target.files ?? [];
    target.value = ''; // reset so picking the same file again still fires onChange
    if (!file) {
      return;
    }
    setFileContent(await file.text());
  };

  const handleEditorDidMount = (_, monaco) => {
    monaco.languages.html.registerHTMLLanguageService(ssoItem.metadataFormat, {}, { documentFormattingEdits: true });
  };

  return (
    <BaseDrawer
      open={open}
      onClose={onClose}
      size="xl"
      slotProps={{
        paper: { style: { display: 'flex', flexDirection: 'column' } },
        header: {
          title: `${ssoItem.title} metadata`,
          preCloser: (
            <Button component="label" startIcon={<UploadFile fontSize="small" />}>
              Import from file
              <input hidden type="file" accept={`.${ssoItem.metadataFormat}`} onChange={onFileChange} />
            </Button>
          )
        }
      }}
    >
      <Typography>{editorDescriptions[ssoItem.metadataFormat]}</Typography>
      <CodeEditor className={classes.editor} language={ssoItem.editorLanguage} onChange={setFileContent} onMount={handleEditorDidMount} value={fileContent} />
      {!isMetadataValid && fileContent.length > 4 && (
        <Typography className="margin-bottom-small red" variant="body2">
          There was an error parsing the metadata.
        </Typography>
      )}
      {saveError && (
        <Alert className="margin-bottom-small" severity="error">
          {saveError}
        </Alert>
      )}
      <div className="flexbox align-items-center gap-x-small">
        <Button variant="outlined" color="info" onClick={onCancelClick}>
          Cancel
        </Button>
        <Button variant="contained" disabled={!isMetadataValid} onClick={onSubmitClick}>
          Save SSO configuration
        </Button>
      </div>
    </BaseDrawer>
  );
};

export default SSOEditor;
