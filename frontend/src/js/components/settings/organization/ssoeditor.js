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
import React, { useCallback, useEffect, useRef, useState } from 'react';
import Dropzone from 'react-dropzone';

// material ui
import { Close as CloseIcon, CloudUpload, FileCopyOutlined as CopyPasteIcon } from '@mui/icons-material';
import { Button, Divider, Drawer, IconButton } from '@mui/material';

import Editor, { loader } from '@monaco-editor/react';
import copy from 'copy-to-clipboard';

import { JSON_METADATA_FORMAT, XML_METADATA_FORMAT } from '../../../constants/organizationConstants.js';
import { createFileDownload } from '../../../helpers';
import Loader from '../../common/loader';

loader.config({ paths: { vs: '/ui/vs' } });

const editorProps = {
  height: 700,
  loading: <Loader show />,
  options: {
    autoClosingOvertype: 'auto',
    codeLens: false,
    contextmenu: false,
    enableSplitViewResizing: false,
    formatOnPaste: true,
    lightbulb: { enabled: false },
    lineNumbers: 'off',
    minimap: { enabled: false },
    quickSuggestions: false,
    readOnly: true,
    renderOverviewRuler: false,
    scrollBeyondLastLine: false,
    wordWrap: 'on'
  }
};

export const SSOEditor = ({ ssoItem, config, fileContent, hasSSOConfig, open, onCancel, onClose, onSave, setFileContent, token }) => {
  const [isEditing, setIsEditing] = useState(false);
  const [isMetadataValid, setIsMetadataValid] = useState(false);
  const editorRef = useRef();

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
        } catch (error) {
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

  const onSubmitClick = () => {
    onSave();
    setIsEditing(false);
  };

  const onCopyClick = () => copy(fileContent);

  const onDrop = acceptedFiles => {
    let reader = new FileReader();
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
    editorRef.current = { editor, monaco, modifiedEditor: editor };
  };

  return (
    <Drawer className={`${open ? 'fadeIn' : 'fadeOut'}`} anchor="right" open={open} onClose={onClose} PaperProps={{ style: { minWidth: '75vw' } }}>
      <div className="flexbox margin-bottom-small space-between">
        <h3>{ssoItem.title} metadata</h3>
        <div className="flexbox center-aligned">
          <Dropzone multiple={false} onDrop={onDrop}>
            {({ getRootProps, getInputProps }) => (
              <div {...getRootProps()}>
                <input {...getInputProps()} />
                <Button color="secondary" startIcon={<CloudUpload fontSize="small" />}>
                  Import from a file
                </Button>
              </div>
            )}
          </Dropzone>

          <IconButton onClick={onClose} size="large">
            <CloseIcon />
          </IconButton>
        </div>
      </div>
      <Divider light />
      <Editor
        {...editorProps}
        language={ssoItem.editorLanguage}
        defaultLanguage={ssoItem.editorLanguage}
        options={{
          ...editorProps.options,
          readOnly: hasSSOConfig && !isEditing
        }}
        className="editor modified"
        onChange={setFileContent}
        onMount={handleEditorDidMount}
        value={fileContent}
      />
      {!isMetadataValid && fileContent.length > 4 && <div className="error">There was an error parsing the metadata.</div>}
      <Divider className="margin-top-large margin-bottom" light />
      <div>
        {hasSSOConfig && !isEditing ? (
          <div className="flexbox center-aligned">
            <Button onClick={onEditClick}>Edit</Button>
            <Button onClick={onDownloadClick}>Download file</Button>
            <Button onClick={onCopyClick} startIcon={<CopyPasteIcon />}>
              Copy to clipboard
            </Button>
          </div>
        ) : (
          <>
            <Button onClick={onCancelClick}>Cancel</Button>
            <Button variant="contained" disabled={!isMetadataValid} onClick={onSubmitClick} color="secondary" style={{ marginLeft: 10 }}>
              Save
            </Button>
          </>
        )}
      </div>
    </Drawer>
  );
};

export default SSOEditor;
