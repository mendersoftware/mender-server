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
import React, { useEffect, useState } from 'react';
import Dropzone from 'react-dropzone';

// material ui
import { CloudUpload } from '@mui/icons-material';
import { Button } from '@mui/material';
import { listItemTextClasses } from '@mui/material/ListItemText';
import { makeStyles } from 'tss-react/mui';

import { SSO_TYPES, XML_METADATA_FORMAT } from '../../../constants/organizationConstants.js';
import { toggle } from '../../../helpers';
import ExpandableAttribute from '../../common/expandable-attribute';
import { HELPTOOLTIPS, MenderHelpTooltip } from '../../helptips/helptooltips';
import { maxWidth } from './organizationsettingsitem';
import SSOEditor from './ssoeditor';

const useStyles = makeStyles()(theme => ({
  configDetail: {
    maxWidth,
    [`.${listItemTextClasses.primary}`]: {
      color: theme.palette.text.disabled,
      fontSize: 'smaller'
    }
  },
  uploadIcon: {
    marginBottom: theme.spacing(-0.5),
    marginRight: theme.spacing()
  },
  tinyMargin: {
    marginLeft: theme.spacing(0.5),
    marginRight: theme.spacing(0.5)
  },
  wrapper: {
    alignItems: 'start',
    columnGap: theme.spacing(2),
    display: 'grid',
    gridTemplateColumns: `minmax(max-content, ${maxWidth}px) max-content max-content`,
    position: 'relative',
    ['&.has-sso']: {
      gridTemplateColumns: `${maxWidth - 45}px 1fr`
    }
  }
}));

export const SSOConfig = ({ ssoItem, config, onCancel, onSave, setSnackbar, token }) => {
  const [configDetails, setConfigDetails] = useState([]);
  const [fileContent, setFileContentState] = useState('');
  const [hasSSOConfig, setHasSSOConfig] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const { id, type, ...content } = config || {};
  const configContent = content.config || '';

  // file content should be text, otherwise editor will fail
  const setFileContent = content => setFileContentState(typeof content === 'object' ? JSON.stringify(content) : content);

  const { classes } = useStyles();

  useEffect(() => {
    setHasSSOConfig(!!config);
    setFileContent(configContent);
    if (config?.id) {
      setConfigDetails(SSO_TYPES[type].configDetails.map(item => ({ ...item, value: item.getValue(config.id) })));
    }
  }, [config, configContent, type]);

  const onCancelSSOSettings = () => {
    setHasSSOConfig(!!config);
    setFileContent(configContent);
    setIsEditing(false);
    onCancel();
  };

  const onSaveSSOSettings = () => {
    onSave(id, fileContent);
    setIsEditing(false);
  };

  const onDrop = acceptedFiles => {
    let reader = new FileReader();
    reader.fileName = acceptedFiles[0].name;
    reader.onerror = error => console.log('Error: ', error);
    reader.onload = () => {
      setIsEditing(true);
      setFileContent(reader.result);
    };
    reader.readAsBinaryString(acceptedFiles[0]);
  };

  const onOpenEditorClick = () => setIsEditing(toggle);

  // disable save button if there is no metadata content and metadata format is not xml that means SAML. As we allow to save empty SAML SSO config
  const saveButtonDisabled = !fileContent && ssoItem.metadataFormat !== XML_METADATA_FORMAT;

  return (
    <>
      <div className={`flexbox center-aligned ${classes.wrapper} ${hasSSOConfig ? 'has-sso' : ''}`}>
        {hasSSOConfig ? (
          <a onClick={onOpenEditorClick}>View metadata in the text editor</a>
        ) : (
          <>
            <MenderHelpTooltip id={HELPTOOLTIPS.ssoMetadata.id} style={{ position: 'absolute', left: -35 }} placement="left" />
            <Dropzone multiple={false} onDrop={onDrop}>
              {({ getRootProps, getInputProps }) => (
                <div {...getRootProps()} className="dropzone onboard dashboard-placeholder flexbox centered">
                  <input {...getInputProps()} />
                  <CloudUpload className={classes.uploadIcon} fontSize="small" /> Drag here or <a className={classes.tinyMargin}>browse</a> to upload a metadata
                  document
                </div>
              )}
            </Dropzone>
            <div>
              or <a onClick={onOpenEditorClick}>input with the text editor</a>
            </div>
          </>
        )}
        <div className="flexbox">
          {hasSSOConfig && !isEditing ? (
            <Button onClick={onOpenEditorClick}>Edit</Button>
          ) : (
            <>
              <Button onClick={onCancelSSOSettings}>Cancel</Button>
              <Button className={classes.tinyMargin} onClick={onSaveSSOSettings} disabled={saveButtonDisabled} variant="contained">
                Save
              </Button>
            </>
          )}
        </div>
      </div>
      {hasSSOConfig && (
        <div className="flexbox column margin-top">
          {configDetails.map(item => (
            <ExpandableAttribute
              className={classes.configDetail}
              copyToClipboard
              key={item.key}
              primary={item.label}
              secondary={item.value}
              setSnackbar={setSnackbar}
              disableGutters
              dividerDisabled
            />
          ))}
        </div>
      )}
      <SSOEditor
        open={isEditing}
        ssoItem={ssoItem}
        config={configContent}
        fileContent={fileContent}
        hasSSOConfig={hasSSOConfig}
        onCancel={onCancelSSOSettings}
        onClose={onOpenEditorClick}
        onSave={onSaveSSOSettings}
        setFileContent={setFileContent}
        token={token}
      />
    </>
  );
};

export default SSOConfig;
