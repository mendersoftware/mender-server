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
import { useDispatch, useSelector } from 'react-redux';
import { Link } from 'react-router-dom';

import { InfoOutlined as InfoIcon, Launch as LaunchIcon } from '@mui/icons-material';
import { Button, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { setSnackbar } from '../../../actions/appActions';
import { getDeviceFileDownloadLink } from '../../../actions/deviceActions';
import { BEGINNING_OF_TIME, BENEFITS, TIMEOUTS } from '../../../constants/appConstants';
import { ALL_DEVICES, DEVICE_CONNECT_STATES } from '../../../constants/deviceConstants';
import { AUDIT_LOGS_TYPES } from '../../../constants/organizationConstants';
import { checkPermissionsObject, uiPermissionsById } from '../../../constants/userConstants';
import { createDownload } from '../../../helpers';
import { getCurrentSession, getTenantCapabilities, getUserCapabilities } from '../../../selectors';
import { formatAuditlogs } from '../../../utils/locationutils';
import DocsLink from '../../common/docslink';
import EnterpriseNotification from '../../common/enterpriseNotification';
import Loader from '../../common/loader';
import MenderTooltip from '../../common/mendertooltip';
import Time from '../../common/time';
import FileTransfer from '../troubleshoot/filetransfer';
import TroubleshootContent from '../troubleshoot/terminal-wrapper';
import DeviceDataCollapse from './devicedatacollapse';

const useStyles = makeStyles()(theme => ({
  buttonStyle: { textTransform: 'none', textAlign: 'left' },
  connectionIcon: { marginRight: theme.spacing() },
  content: { maxWidth: 1280 },
  title: { marginRight: theme.spacing(0.5) },
  troubleshootButton: { marginRight: theme.spacing(2) }
}));

export const PortForwardLink = () => (
  <MenderTooltip
    arrow
    title={
      <div style={{ whiteSpace: 'normal' }}>
        <h3>Port forwarding</h3>
        <p>Port forwarding allows you to troubleshoot or use services on or via the device, without opening any ports on the device itself.</p>
        <p>
          To enable port forwarding you will need to install and configure the necessary software on the device and your workstation. Follow the link to learn
          more.
        </p>
      </div>
    }
  >
    <DocsLink className="flexbox centered margin-left" path="add-ons/port-forward">
      Enable port forwarding
      <LaunchIcon className="margin-left-small" fontSize="small" />
    </DocsLink>
  </MenderTooltip>
);

export const DeviceConnectionNote = ({ children }) => {
  const { classes } = useStyles();
  return (
    <div className="flexbox muted">
      <InfoIcon className={classes.connectionIcon} fontSize="small" />
      <Typography className={classes.buttonStyle} variant="body1">
        {children}
      </Typography>
    </div>
  );
};

export const DeviceConnectionMissingNote = () => (
  <DeviceConnectionNote>
    The troubleshoot add-on does not seem to be enabled on this device.
    <br />
    Please <DocsLink path="add-ons/remote-terminal" title="see the documentation" /> for a description on how it works and how to enable it.
  </DeviceConnectionNote>
);

export const DeviceDisconnectedNote = ({ lastConnectionTs }) => (
  <DeviceConnectionNote>
    The troubleshoot add-on is not currently connected on this device, it was last connected on <Time value={lastConnectionTs} />.
    <br />
    Please <DocsLink path="add-ons/remote-terminal" title="see the documentation" /> for more information.
  </DeviceConnectionNote>
);

export const TroubleshootButton = ({ disabled, item, onClick }) => {
  const { classes } = useStyles();
  return (
    <Button className={classes.troubleshootButton} onClick={() => onClick(item.key)} disabled={disabled} startIcon={item.icon}>
      <Typography className={classes.buttonStyle} variant="subtitle2">
        {item.title}
      </Typography>
    </Button>
  );
};

const deviceAuditlogType = AUDIT_LOGS_TYPES.find(type => type.value === 'device');

const tabs = {
  terminal: {
    title: 'Remote terminal',
    value: 'terminal',
    canShow: ({ canTroubleshoot, canWriteDevices, groupsPermissions }, { group }) =>
      (canTroubleshoot && canWriteDevices) || checkPermissionsObject(groupsPermissions, uiPermissionsById.connect.value, group, ALL_DEVICES),
    Component: TroubleshootContent
  },
  transfer: { title: 'File transfer', value: 'transfer', canShow: ({ canTroubleshoot }) => canTroubleshoot, Component: FileTransfer }
};

export const DeviceConnection = ({ className = '', device }) => {
  const [socketClosed, setSocketClosed] = useState();
  const [availableTabs, setAvailableTabs] = useState(Object.values(tabs));
  const [downloadPath, setDownloadPath] = useState('');
  const [file, setFile] = useState();
  const [socketInitialized, setSocketInitialized] = useState(undefined);
  const [uploadPath, setUploadPath] = useState('');
  const closeTimer = useRef();
  const initTimer = useRef();

  const userCapabilities = useSelector(getUserCapabilities);
  const { canAuditlog, canTroubleshoot } = userCapabilities;
  const { hasAuditlogs, hasDeviceConnect } = useSelector(getTenantCapabilities);
  const { token } = useSelector(getCurrentSession);
  const { classes } = useStyles();
  const { connect_status, connect_updated_ts, isOffline } = device;
  const [connectionStatus, setConnectionStatus] = useState(connect_status);

  const dispatch = useDispatch();
  const dispatchedSetSnackbar = useCallback((...args) => dispatch(setSnackbar(...args)), [dispatch]);

  useEffect(() => {
    if (!socketClosed) {
      return;
    }
    clearTimeout(closeTimer.current);
    closeTimer.current = setTimeout(() => setSocketClosed(false), TIMEOUTS.fiveSeconds);
  }, [socketClosed]);

  useEffect(() => {
    setConnectionStatus(connect_status);
    clearTimeout(initTimer.current);
    if (connectionStatus) {
      return;
    }
    initTimer.current = setTimeout(() => {
      setConnectionStatus(!connect_status || isOffline ? DEVICE_CONNECT_STATES.unknown : connect_status);
    }, TIMEOUTS.fiveSeconds);
    return () => clearTimeout(initTimer.current);
  }, [connect_status, connectionStatus, device.id, isOffline]);

  useEffect(() => {
    const allowedTabs = Object.values(tabs).reduce((accu, tab) => {
      if (tab.canShow(userCapabilities, device) && connectionStatus === DEVICE_CONNECT_STATES.connected) {
        accu.push(tab);
      }
      return accu;
    }, []);
    setAvailableTabs(allowedTabs);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [connectionStatus, JSON.stringify(device), JSON.stringify(userCapabilities)]);

  const onDownloadClick = useCallback(
    path => {
      setDownloadPath(path);
      dispatch(setSnackbar('Downloading file'));
      dispatch(getDeviceFileDownloadLink(device.id, path)).then(address => {
        const filename = path.substring(path.lastIndexOf('/') + 1) || 'file';
        createDownload(address, filename, token);
      });
    },
    [dispatch, device.id, token]
  );

  return (
    <DeviceDataCollapse
      isAddOn
      title={
        <div className="flexbox center-aligned">
          <h4>Troubleshooting</h4>
          <div className={`flexbox ${className}`}>
            {hasDeviceConnect && connectionStatus !== DEVICE_CONNECT_STATES.unknown && canTroubleshoot && <PortForwardLink />}
            {hasDeviceConnect && canAuditlog && hasAuditlogs && (
              <Link
                className="flexbox center-aligned margin-left"
                to={`/auditlog?${formatAuditlogs({ pageState: { type: deviceAuditlogType, detail: device.id, startDate: BEGINNING_OF_TIME } }, {})}`}
              >
                List all log entries for this device
              </Link>
            )}
          </div>
          <EnterpriseNotification className="margin-left-small" id={BENEFITS.deviceTroubleshoot.id} />
        </div>
      }
    >
      {hasDeviceConnect && (
        <div className={`flexbox column ${classes.content}`}>
          {!connectionStatus && (
            <div className="flexbox centered">
              <Loader show />
            </div>
          )}
          {connectionStatus === DEVICE_CONNECT_STATES.unknown && <DeviceConnectionMissingNote />}
          {connectionStatus === DEVICE_CONNECT_STATES.disconnected && <DeviceDisconnectedNote lastConnectionTs={connect_updated_ts} />}
          {availableTabs.map(({ Component, title, value }) => (
            <div key={value}>
              <h4 className="margin-top-large">{title}</h4>
              <Component
                device={device}
                downloadPath={downloadPath}
                file={file}
                onDownload={onDownloadClick}
                setDownloadPath={setDownloadPath}
                setFile={setFile}
                setSnackbar={dispatchedSetSnackbar}
                setSocketClosed={setSocketClosed}
                setSocketInitialized={setSocketInitialized}
                setUploadPath={setUploadPath}
                socketInitialized={socketInitialized}
                uploadPath={uploadPath}
                userCapabilities={userCapabilities}
              />
            </div>
          ))}
        </div>
      )}
    </DeviceDataCollapse>
  );
};

export default DeviceConnection;
