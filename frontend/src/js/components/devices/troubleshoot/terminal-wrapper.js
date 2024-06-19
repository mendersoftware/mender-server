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
import React, { useCallback, useEffect, useRef, useState } from 'react';
import Dropzone from 'react-dropzone';
import { useSelector } from 'react-redux';
import { Link } from 'react-router-dom';

import { Button } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import moment from 'moment';
import momentDurationFormatSetup from 'moment-duration-format';

import { BEGINNING_OF_TIME, TIMEOUTS } from '../../../constants/appConstants';
import { getCurrentSession, getFeatures, getIsPreview, getTenantCapabilities, getUserCapabilities } from '../../../selectors';
import Tracking from '../../../tracking';
import { useSession } from '../../../utils/sockethook';
import { MaybeTime } from '../../common/time';
import { getCode } from '../dialogs/make-gateway-dialog';
import Terminal from '../troubleshoot/terminal';
import ListOptions from '../widgets/listoptions';

momentDurationFormatSetup(moment);

const useStyles = makeStyles()(theme => ({
  connectionActions: { marginTop: theme.spacing() },
  connectionButton: { background: theme.palette.background.terminal, display: 'grid', placeContent: 'center' },
  sessionInfo: { gap: theme.spacing(3), marginBottom: theme.spacing(), '&>div': { gap: theme.spacing(2) } },
  terminalContent: {
    display: 'grid',
    gridTemplateRows: `max-content 0 minmax(${theme.spacing(60)}, 1fr) max-content`,
    flexGrow: 1,
    overflow: 'hidden',
    '&.device-connected': {
      gridTemplateRows: `max-content minmax(${theme.spacing(80)}, 1fr) max-content`
    }
  }
}));

const SessionInfo = ({ socketInitialized, startTime }) => {
  const [elapsed, setElapsed] = useState(moment());
  const timer = useRef();
  const { classes } = useStyles();

  useEffect(() => {
    clearInterval(timer.current);
    if (!socketInitialized) {
      return;
    }
    timer.current = setInterval(() => setElapsed(moment()), TIMEOUTS.halfASecond);
    return () => {
      clearInterval(timer.current);
    };
  }, [socketInitialized]);

  return (
    <div className={`flexbox ${classes.sessionInfo}`}>
      {[
        { key: 'status', title: 'Session status', content: socketInitialized ? 'connected' : 'disconnected' },
        { key: 'start', title: 'Connection start', content: <MaybeTime value={startTime} /> },
        {
          key: 'duration',
          title: 'Duration',
          content: startTime ? `${moment.duration(elapsed.diff(moment(startTime))).format('hh:mm:ss', { trim: false })}` : '-'
        }
      ].map(({ key, title, content }) => (
        <div key={key} className="flexbox">
          <div>{title}</div>
          <b>{content}</b>
        </div>
      ))}
    </div>
  );
};

const TroubleshootContent = ({ device, onDownload, setSocketClosed, setUploadPath, setFile, setSnackbar, setSocketInitialized, socketInitialized }) => {
  const [terminalInput, setTerminalInput] = useState('');
  const [startTime, setStartTime] = useState();
  const [snackbarAlreadySet, setSnackbarAlreadySet] = useState(false);
  const snackTimer = useRef();
  const { classes } = useStyles();
  const termRef = useRef({ terminal: React.createRef(), terminalRef: React.createRef() });

  const { isHosted } = useSelector(getFeatures);
  const { hasAuditlogs, isEnterprise } = useSelector(getTenantCapabilities);
  const { canAuditlog } = useSelector(getUserCapabilities);
  const canPreview = useSelector(getIsPreview);
  const { token } = useSelector(getCurrentSession);
  const onMessageReceived = useCallback(message => {
    if (!termRef.current.terminal.current) {
      return;
    }
    termRef.current.terminal.current.write(new Uint8Array(message));
  }, []);

  const onNotify = useCallback(
    content => {
      if (snackbarAlreadySet) {
        return;
      }
      setSnackbarAlreadySet(true);
      setSnackbar(content, TIMEOUTS.threeSeconds);
      snackTimer.current = setTimeout(() => setSnackbarAlreadySet(false), TIMEOUTS.threeSeconds + TIMEOUTS.debounceShort);
    },
    [setSnackbar, snackbarAlreadySet]
  );

  const onHealthCheckFailed = useCallback(() => {
    if (!socketInitialized) {
      return;
    }
    onNotify('Health check failed: connection with the device lost.');
  }, [onNotify, socketInitialized]);

  const onSocketClose = useCallback(
    event => {
      // abnormal socket close might happen without socket being initialized, in case of forbidden permissions
      // this should be checked before socketInitialized condition
      if (event.code === 1006) {
        // 1006: abnormal closure
        onNotify('Connection to the remote terminal is forbidden.');
      }

      if (!socketInitialized) {
        return;
      }
      if (event.wasClean) {
        onNotify(`Connection with the device closed.`);
      } else {
        onNotify('Connection with the device died.');
      }
      setSocketInitialized(false);
      setSocketClosed(true);
    },
    [onNotify, setSocketClosed, setSocketInitialized, socketInitialized]
  );

  const [connect, sendMessage, close, sessionState] = useSession({
    onClose: onSocketClose,
    onHealthCheckFailed,
    onMessageReceived,
    onNotify,
    onOpen: setSocketInitialized,
    token
  });

  useEffect(() => {
    if (socketInitialized === undefined) {
      return;
    }
    if (socketInitialized) {
      setStartTime(new Date());
      setSnackbar('Connection with the device established.', TIMEOUTS.fiveSeconds);
    } else {
      close();
    }
  }, [close, setSnackbar, socketInitialized]);

  useEffect(() => {
    return () => {
      clearTimeout(snackTimer.current);
      if (socketInitialized !== undefined) {
        close();
      }
    };
  }, [close, socketInitialized]);

  useEffect(() => {
    if (sessionState !== WebSocket.OPEN) {
      return;
    }
    return close;
  }, [close, sessionState]);

  const onConnectionToggle = () => {
    if (sessionState === WebSocket.CLOSED) {
      setStartTime();
      setSocketInitialized(undefined);
      setSocketClosed(false);
      connect(device.id);
      Tracking.event({ category: 'devices', action: 'open_terminal' });
    } else {
      setSocketInitialized(false);
      close();
    }
  };

  const onMakeGatewayClick = () => {
    const code = getCode(canPreview);
    setTerminalInput(code);
  };

  const onDrop = acceptedFiles => {
    if (acceptedFiles.length === 1) {
      setFile(acceptedFiles[0]);
      setUploadPath(`/tmp/${acceptedFiles[0].name}`);
    }
  };

  const commandHandlers = isHosted && isEnterprise ? [{ key: 'thing', onClick: onMakeGatewayClick, title: 'Promote to Mender gateway' }] : [];

  const visibilityToggle = !socketInitialized ? { maxHeight: 0, overflow: 'hidden' } : {};
  return (
    <div className={`${classes.terminalContent} ${socketInitialized ? 'device-connected' : ''}`}>
      <SessionInfo socketInitialized={socketInitialized} startTime={startTime} />
      <Dropzone activeClassName="active" rejectClassName="active" multiple={false} onDrop={onDrop} noClick>
        {({ getRootProps }) => (
          <div {...getRootProps()} style={{ position: 'relative', ...visibilityToggle }}>
            <Terminal
              onDownloadClick={onDownload}
              sendMessage={sendMessage}
              socketInitialized={socketInitialized}
              style={{ position: 'absolute', width: '100%', height: '100%', ...visibilityToggle }}
              textInput={terminalInput}
              xtermRef={termRef}
            />
          </div>
        )}
      </Dropzone>
      {!socketInitialized && (
        <div className={classes.connectionButton}>
          <Button variant="contained" color="secondary" onClick={onConnectionToggle}>
            Connect Terminal
          </Button>
        </div>
      )}
      <div className={`flexbox space-between ${classes.connectionActions}`}>
        <Button onClick={onConnectionToggle}>{socketInitialized ? 'Disconnect' : 'Connect'} Terminal</Button>
        {canAuditlog && hasAuditlogs && (
          <Button component={Link} to={`/auditlog?objectType=device&objectId=${device.id}&startDate=${BEGINNING_OF_TIME}`}>
            View Session Logs for this device
          </Button>
        )}
        {socketInitialized && !!commandHandlers.length && <ListOptions options={commandHandlers} title="Quick commands" />}
      </div>
    </div>
  );
};

export default TroubleshootContent;
