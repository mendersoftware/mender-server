// Copyright 2021 Northern.tech AS
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
import React, { useEffect, useRef, useState } from 'react';
import { useDispatch } from 'react-redux';
import { Link } from 'react-router-dom';

import { CheckCircleOutlined, CloudUploadOutlined as CloudUpload, Refresh as RefreshIcon } from '@mui/icons-material';
import { Button } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import Editor, { DiffEditor, loader } from '@monaco-editor/react';
import pluralize from 'pluralize';

import { getDeviceTwin, setDeviceTwin } from '../../../actions/deviceActions';
import { TIMEOUTS } from '../../../constants/appConstants';
import { EXTERNAL_PROVIDER } from '../../../constants/deviceConstants';
import { deepCompare, isEmpty } from '../../../helpers';
import InfoHint from '../../common/info-hint';
import Loader from '../../common/loader';
import Time from '../../common/time';
import DeviceDataCollapse from './devicedatacollapse';

loader.config({ paths: { vs: '/ui/vs' } });

const useStyles = makeStyles()(theme => ({
  buttonSpacer: { marginLeft: theme.spacing(2) },
  title: { alignItems: 'baseline' },
  titleContainer: { width: '100%' },
  diffStatus: {
    minHeight: 75,
    display: 'grid',
    gridTemplateColumns: 'min-content 300px max-content',
    gridColumnGap: theme.spacing(2),
    alignItems: 'center',
    background: theme.palette.grey[100],
    width: 'min-content'
  }
}));

export const LastSyncNote = ({ updateTime }) => (
  <div className="muted slightly-smaller" style={{ marginTop: 2 }}>
    Last synced: <Time value={updateTime} />
  </div>
);

const NoDiffStatus = () => {
  const { classes } = useStyles();
  return (
    <div className={['padding', classes.diffStatus]}>
      <CheckCircleOutlined className="green" />
      <div>No difference between desired and reported configuration</div>
    </div>
  );
};

export const TwinError = ({ providerTitle, twinError }) => (
  <InfoHint
    content={
      <>
        {twinError}
        <br />
        Please check your connection string in the <Link to="/settings/integrations">Integration settings</Link>, and check that the device exists in your{' '}
        {providerTitle}
      </>
    }
  />
);

export const TwinSyncStatus = ({ diffCount, providerTitle, twinError, updateTime }) => {
  const classes = useStyles();
  if (twinError) {
    return <TwinError providerTitle={providerTitle} twinError={twinError} />;
  }
  return !diffCount ? (
    <NoDiffStatus />
  ) : (
    <div className={['padding', classes.diffStatus]}>
      <CloudUpload />
      <div>
        <b>
          Found {diffCount} {pluralize('difference', diffCount)}
        </b>{' '}
        between desired and reported configuration
      </div>
      <LastSyncNote updateTime={updateTime} />
    </div>
  );
};

export const Title = ({ providerTitle, twinTitle, updateTime }) => {
  const { classes } = useStyles();
  return (
    <div className={`flexbox center-aligned space-between ${classes.titleContainer}`}>
      <div className={`flexbox ${classes.title}`}>
        <h4 className="margin-right">
          {providerTitle} {twinTitle}
        </h4>
        <LastSyncNote updateTime={updateTime} />
      </div>
      <Link to="/settings/integrations">Integration settings</Link>
    </div>
  );
};

const editorProps = {
  height: 500,
  defaultLanguage: 'json',
  language: 'json',
  loading: <Loader show />,
  options: {
    autoClosingOvertype: 'auto',
    codeLens: false,
    contextmenu: false,
    enableSplitViewResizing: false,
    formatOnPaste: true,
    lightbulb: { enabled: false },
    minimap: { enabled: false },
    lineNumbersMinChars: 3,
    quickSuggestions: false,
    renderOverviewRuler: false,
    scrollBeyondLastLine: false,
    readOnly: true
  }
};
const maxWidth = 800;

const indentation = 4; // number of spaces, tab based indentation won't show in the editor, but be converted to 4 spaces

const stringifyTwin = twin => JSON.stringify(twin, undefined, indentation) ?? '';

export const DeviceTwin = ({ device, integration }) => {
  const [configured, setConfigured] = useState('');
  const [diffCount, setDiffCount] = useState(0);
  const [isEditing, setIsEditing] = useState(false);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [errorMessage, setErrorMessage] = useState('');
  const [initialized, setInitialized] = useState(false);
  const [reported, setReported] = useState('');
  const [updated, setUpdated] = useState('');
  const [isSync, setIsSync] = useState(true);
  const editorRef = useRef(null);
  const { classes } = useStyles();
  const dispatch = useDispatch();

  const externalProvider = EXTERNAL_PROVIDER[integration.provider];
  const { [integration.id]: deviceTwin = {} } = device.twinsByIntegration ?? {};
  const { desired: configuredTwin = {}, reported: reportedTwin = {}, twinError, updated_ts: updateTime = device.created_ts } = deviceTwin;

  useEffect(() => {
    const textContent = stringifyTwin(configuredTwin);
    setConfigured(textContent);
    setUpdated(textContent);
    setReported(stringifyTwin(reportedTwin));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    setReported(stringifyTwin(reportedTwin));
    if (isEditing) {
      return;
    }
    const textContent = stringifyTwin(configuredTwin);
    setConfigured(textContent);
    setUpdated(textContent);
  }, [configuredTwin, reportedTwin, isEditing]);

  useEffect(() => {
    setIsSync(deepCompare(reported, configured));
  }, [configured, reported]);

  const handleEditorDidMount = (editor, monaco) => {
    editorRef.current = { editor, monaco, modifiedEditor: editor };
  };

  const handleDiffEditorDidMount = (editor, monaco) => {
    const modifiedEditor = editor.getModifiedEditor();
    modifiedEditor.onDidChangeModelContent(() => setUpdated(modifiedEditor.getValue()));
    editor.onDidUpdateDiff(onDidUpdateDiff);
    editorRef.current = { editor, monaco, modifiedEditor };
  };

  const onDidUpdateDiff = () => {
    const changes = editorRef.current.editor.getLineChanges();
    setDiffCount(changes.length);
    setInitialized(true);
  };

  const onApplyClick = () => {
    let update = {};
    try {
      update = JSON.parse(updated);
    } catch (error) {
      setErrorMessage('There was an error parsing the device twin changes, please ensure that it is valid JSON.');
      return;
    }
    editorRef.current.modifiedEditor.getAction('editor.action.formatDocument').run();
    setUpdated(stringifyTwin(update));
    setErrorMessage('');
    dispatch(setDeviceTwin(device.id, integration, update)).then(() => setIsEditing(false));
  };

  const onCancelClick = () => {
    const textContent = stringifyTwin(configuredTwin);
    setUpdated(textContent);
    editorRef.current.modifiedEditor.getModel().setValue(textContent);
    setIsEditing(false);
  };

  const onRefreshClick = () => {
    setIsRefreshing(true);
    dispatch(getDeviceTwin(device.id, integration)).finally(() => setTimeout(() => setIsRefreshing(false), TIMEOUTS.halfASecond));
  };

  const onEditClick = () => setIsEditing(true);

  const widthStyle = { maxWidth: isSync ? maxWidth : 'initial' };

  return (
    <DeviceDataCollapse
      header={
        <div className="flexbox column">
          {initialized ? (
            <TwinSyncStatus diffCount={diffCount} providerTitle={externalProvider.title} twinError={twinError} updateTime={updateTime} />
          ) : (
            <Loader show={!initialized} />
          )}
        </div>
      }
      title={<Title providerTitle={externalProvider.title} twinTitle={externalProvider.twinTitle} updateTime={updateTime} />}
    >
      <div className={`flexbox column ${isEditing ? 'twin-editing' : ''}`}>
        <div style={widthStyle}>
          {!initialized || (!(isEmpty(reported) && isEmpty(configured)) && !isSync) ? (
            <>
              <div className="two-columns">
                <h4>Desired configuration</h4>
                <h4>Reported configuration</h4>
              </div>
              <DiffEditor
                {...editorProps}
                original={reported}
                modified={configured}
                onMount={handleDiffEditorDidMount}
                options={{
                  ...editorProps.options,
                  readOnly: !isEditing
                }}
              />
            </>
          ) : (
            <>
              <h4>{!deviceTwin.reported || isEditing ? 'Desired' : 'Reported'} configuration</h4>
              <Editor
                {...editorProps}
                options={{
                  ...editorProps.options,
                  readOnly: !isEditing
                }}
                className="editor modified"
                onMount={handleEditorDidMount}
                value={reported || configured}
                onChange={setUpdated}
              />
            </>
          )}
          {!!errorMessage && <p className="warning">{errorMessage}</p>}
        </div>
        <div className="two-columns margin-top" style={isSync ? { gridTemplateColumns: `${maxWidth}px 1fr` } : widthStyle}>
          <div className="flexbox" style={{ alignItems: 'flex-start', justifyContent: 'flex-end' }}>
            {isEditing ? (
              <>
                <Button onClick={onCancelClick}>Cancel</Button>
                <Button className={classes.buttonSpacer} color="secondary" onClick={onApplyClick} variant="contained">
                  Save
                </Button>
              </>
            ) : (
              <Button color="secondary" onClick={onEditClick} variant="contained">
                Edit desired configuration
              </Button>
            )}
          </div>
          <div className="flexbox" style={{ justifyContent: 'flex-end' }}>
            <Loader show={isRefreshing} small table />
            {!isEditing && (
              <Button onClick={onRefreshClick} startIcon={<RefreshIcon />}>
                Refresh
              </Button>
            )}
          </div>
        </div>
      </div>
    </DeviceDataCollapse>
  );
};

export default DeviceTwin;

export const IntegrationTab = ({ device, integrations }) => (
  <div>
    {integrations.map(integration => (
      <DeviceTwin key={integration.id} device={device} integration={integration} />
    ))}
  </div>
);
