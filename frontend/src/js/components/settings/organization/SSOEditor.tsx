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
import Dropzone from 'react-dropzone';

// material ui
import { CloudUpload, FileCopyOutlined as CopyPasteIcon } from '@mui/icons-material';
import { Button, Divider, Drawer } from '@mui/material';

import { CodeEditor } from '@northern.tech/common-ui/CodeEditor';
import { DrawerTitle } from '@northern.tech/common-ui/DrawerTitle';
import { JSON_METADATA_FORMAT, XML_METADATA_FORMAT } from '@northern.tech/store/constants';
import { createFileDownload } from '@northern.tech/utils/helpers';
import copy from 'copy-to-clipboard';

export const SSOEditor = ({ ssoItem, config, fileContent, hasSSOConfig, open, onCancel, onClose, onSave, setFileContent, token }) => {
  const [isEditing, setIsEditing] = useState(false);
  const [isMetadataValid, setIsMetadataValid] = useState(false);
  const isReadOnly = hasSSOConfig && !isEditing;

  useEffect(() => {
    if (!fileContent) {
      return;
    }

    const parser = new DOMParser();
    let valid = false;
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

  const onEditClick = () => setIsEditing(true);

  const onDownloadClick = () => createFileDownload(fileContent, `metadata.${ssoItem.metadataFormat}`, token);

  const onCancelClick = useCallback(() => {
    if (isEditing) {
      setFileContent(config);
      if (!hasSSOConfig) {
        return onCancel();
      }
      return setIsEditing(false);
    }
    onClose();
  }, [config, hasSSOConfig, isEditing, onCancel, setFileContent, onClose]);

  const onSubmitClick = async () => {
    try {
      await onSave();
      setIsEditing(false);
    } catch {
      // error already handled in thunk - leave open
    }
  };

  const onCopyClick = () => copy(fileContent);

  const onDrop = acceptedFiles => {
    const reader = new FileReader();
    reader.fileName = acceptedFiles[0].name;
    reader.onerror = error => {
      console.log('Error: ', error);
      setIsEditing(false);
    };
    reader.onload = () => {
      setFileContent(reader.result);
      setIsEditing(true);
    };
    reader.readAsBinaryString(acceptedFiles[0]);
  };

  const handleEditorDidMount = (editor, monaco) => {
    monaco.languages.html.registerHTMLLanguageService(ssoItem.metadataFormat, {}, { documentFormattingEdits: true });
  };

  return (
    <Drawer
      className={`${open ? 'fadeIn' : 'fadeOut'}`}
      anchor="right"
      open={open}
      onClose={onClose}
      PaperProps={{ style: { minWidth: '75vw', display: 'flex', flexDirection: 'column' } }}
    >
      <DrawerTitle
        title={`${ssoItem.title} metadata`}
        preCloser={
          <Dropzone multiple={false} onDrop={onDrop}>
            {({ getRootProps, getInputProps }) => (
              <div {...getRootProps()}>
                <input {...getInputProps()} />
                <Button startIcon={<CloudUpload fontSize="small" />}>Import from a file</Button>
              </div>
            )}
          </Dropzone>
        }
        onClose={onClose}
      />
      <Divider light />
      <CodeEditor
        className="full-height"
        language={ssoItem.editorLanguage}
        readOnly={isReadOnly}
        onChange={setFileContent}
        onMount={handleEditorDidMount}
        value={fileContent}
      />
      {!isMetadataValid && fileContent.length > 4 && <div className="error">There was an error parsing the metadata.</div>}
      <Divider className="margin-top-large margin-bottom" light />
      <div>
        {hasSSOConfig && !isEditing ? (
          <div className="flexbox align-items-center">
            <Button onClick={onEditClick}>Edit</Button>
            <Button onClick={onDownloadClick}>Download file</Button>
            <Button onClick={onCopyClick} startIcon={<CopyPasteIcon />}>
              Copy to clipboard
            </Button>
          </div>
        ) : (
          <>
            <Button onClick={onCancelClick}>Cancel</Button>
            <Button variant="contained" disabled={!isMetadataValid} onClick={onSubmitClick} style={{ marginLeft: 10 }}>
              Save
            </Button>
          </>
        )}
      </div>
    </Drawer>
  );
};

export default SSOEditor;
