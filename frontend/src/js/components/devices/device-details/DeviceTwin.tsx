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
import { useEffect, useRef, useState } from 'react';
import { useDispatch } from 'react-redux';

import { CheckCircleOutlined, CloudUploadOutlined as CloudUpload, Refresh as RefreshIcon } from '@mui/icons-material';
import { Button, Typography } from '@mui/material';
import { alpha } from '@mui/material/styles';
import { makeStyles } from 'tss-react/mui';

import { DiffEditor } from '@monaco-editor/react';
import { CodeEditor, defaultEditorOptions, useEditorStyles, useEditorTheme } from '@northern.tech/common-ui/CodeEditor';
import { ContentSection } from '@northern.tech/common-ui/ContentSection';
import InfoHint from '@northern.tech/common-ui/InfoHint';
import { Link } from '@northern.tech/common-ui/Link';
import Loader from '@northern.tech/common-ui/Loader';
import Time from '@northern.tech/common-ui/Time';
import { EXTERNAL_PROVIDER, TIMEOUTS } from '@northern.tech/store/constants';
import { getDeviceTwin, setDeviceTwin } from '@northern.tech/store/thunks';
import { deepCompare, isEmpty } from '@northern.tech/utils/helpers';
import pluralize from 'pluralize';

const useStyles = makeStyles()(theme => ({
  buttonSpacer: { marginLeft: theme.spacing(2) },
  diffStatus: {
    backgroundColor:
      theme.palette.mode === 'dark'
        ? alpha(theme.palette.grey[300], theme.palette.action.selectedOpacity)
        : alpha(theme.palette.grey[400], theme.palette.action.hoverOpacity)
  }
}));

export const LastSyncNote = ({ updateTime }) => (
  <Typography variant="body2">
    Last synced: <Time value={updateTime} />
  </Typography>
);

const NoDiffStatus = () => {
  const { classes } = useStyles();
  return (
    <div className={`padding-medium flexbox ${classes.diffStatus}`}>
      <CheckCircleOutlined className="green margin-right-medium" />
      <Typography>No difference between desired and reported configuration</Typography>
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
  const { classes } = useStyles();
  if (twinError) {
    return <TwinError providerTitle={providerTitle} twinError={twinError} />;
  }
  return !diffCount ? (
    <NoDiffStatus />
  ) : (
    <div className={`padding-medium flexbox space-between align-items-center ${classes.diffStatus}`}>
      <div className="flexbox align-items-center">
        <CloudUpload className="margin-small" />
        <Typography className="margin-left-x-small">
          Found {diffCount} {pluralize('difference', diffCount)} between desired and reported configuration
        </Typography>
      </div>
      <LastSyncNote updateTime={updateTime} />
    </div>
  );
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
  const { classes: editorClasses } = useEditorStyles();
  const dispatch = useDispatch();
  const { editorThemeName, defineEditorTheme } = useEditorTheme(!isEditing);
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
    defineEditorTheme(monaco);
    const modifiedEditor = editor.getModifiedEditor();
    modifiedEditor.onDidChangeModelContent(() => setUpdated(modifiedEditor.getValue()));
    editor.onDidUpdateDiff(onDidUpdateDiff);
    editorRef.current = { editor, monaco, modifiedEditor };
  };

  const onDidUpdateDiff = () => {
    const changes = editorRef.current.editor.getLineChanges() ?? '';
    setDiffCount(changes.length);
    setInitialized(true);
  };

  const onApplyClick = () => {
    let update = {};
    try {
      update = JSON.parse(updated);
    } catch {
      setErrorMessage('There was an error parsing the device twin changes, please ensure that it is valid JSON.');
      return;
    }
    editorRef.current.modifiedEditor.getAction('editor.action.formatDocument').run();
    setUpdated(stringifyTwin(update));
    setErrorMessage('');
    dispatch(setDeviceTwin({ deviceId: device.id, integration, settings: update })).then(() => setIsEditing(false));
  };

  const onCancelClick = () => {
    const textContent = stringifyTwin(configuredTwin);
    setUpdated(textContent);
    editorRef.current.modifiedEditor.getModel().setValue(textContent);
    setIsEditing(false);
  };

  const onRefreshClick = () => {
    setIsRefreshing(true);
    dispatch(getDeviceTwin({ deviceId: device.id, integration })).finally(() => setTimeout(() => setIsRefreshing(false), TIMEOUTS.halfASecond));
  };

  const onEditClick = () => setIsEditing(true);

  const widthStyle = { maxWidth: isSync ? maxWidth : 'initial' };

  return (
    <ContentSection postTitle={<Link to="/settings/integrations">Integration settings</Link>} title={`${externalProvider.title} ${externalProvider.twinTitle}`}>
      <div className="flexbox column">
        {initialized ? (
          <TwinSyncStatus diffCount={diffCount} providerTitle={externalProvider.title} twinError={twinError} updateTime={updateTime} />
        ) : (
          <Loader show />
        )}
        <div className="margin-top-medium" style={widthStyle}>
          {!initialized || (!(isEmpty(reported) && isEmpty(configured)) && !isSync) ? (
            <>
              <div className="two-columns">
                <Typography variant="subtitle1">Desired configuration</Typography>
                <Typography variant="subtitle1">Reported configuration</Typography>
              </div>
              <div className={editorClasses.wrapper}>
                <DiffEditor
                  height={500}
                  language="json"
                  loading={<Loader show />}
                  original={reported}
                  modified={configured}
                  beforeMount={defineEditorTheme}
                  theme={editorThemeName}
                  onMount={handleDiffEditorDidMount}
                  options={{
                    ...defaultEditorOptions,
                    lineNumbersMinChars: 3,
                    readOnly: !isEditing
                  }}
                />
              </div>
            </>
          ) : (
            <>
              <Typography variant="subtitle1">{!deviceTwin.reported || isEditing ? 'Desired' : 'Reported'} configuration</Typography>
              <CodeEditor language="json" readOnly={!isEditing} onMount={handleEditorDidMount} value={reported || configured} onChange={setUpdated} />
            </>
          )}
          {!!errorMessage && <Typography className="warning">{errorMessage}</Typography>}
        </div>
        <div className="margin-top-small flexbox">
          <div className="flexbox">
            {isEditing ? (
              <>
                <Button onClick={onCancelClick}>Cancel</Button>
                <Button className={classes.buttonSpacer} onClick={onApplyClick} variant="contained">
                  Save
                </Button>
              </>
            ) : (
              <Button onClick={onEditClick} variant="contained">
                Edit desired configuration
              </Button>
            )}
          </div>
          <div className="flexbox margin-left-small">
            {!isEditing && (
              <Button className="margin-right-small" onClick={onRefreshClick} startIcon={<RefreshIcon />}>
                Refresh
              </Button>
            )}
            <Loader show={isRefreshing} small table />
          </div>
        </div>
      </div>
    </ContentSection>
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
