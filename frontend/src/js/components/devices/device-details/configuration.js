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

import { Block as BlockIcon, CheckCircle as CheckCircleIcon, Error as ErrorIcon, Refresh as RefreshIcon, SaveAlt as SaveAltIcon } from '@mui/icons-material';
import { Button, Checkbox, FormControlLabel, Typography } from '@mui/material';

import { setSnackbar } from '../../../actions/appActions';
import { abortDeployment, getDeviceLog, getSingleDeployment } from '../../../actions/deploymentActions';
import { applyDeviceConfig, setDeviceConfig } from '../../../actions/deviceActions';
import { saveGlobalSettings } from '../../../actions/userActions';
import { BENEFITS, TIMEOUTS } from '../../../constants/appConstants';
import { DEPLOYMENT_ROUTES, DEPLOYMENT_STATES } from '../../../constants/deploymentConstants';
import { DEVICE_STATES } from '../../../constants/deviceConstants';
import { deepCompare, groupDeploymentDevicesStats, groupDeploymentStats, isEmpty, toggle } from '../../../helpers';
import { getDeviceConfigDeployment, getTenantCapabilities } from '../../../selectors';
import Tracking from '../../../tracking';
import ConfigurationObject from '../../common/configurationobject';
import Confirm, { EditButton } from '../../common/confirm';
import LogDialog from '../../common/dialogs/log';
import { DOCSTIPS, DocsTooltip } from '../../common/docslink';
import EnterpriseNotification from '../../common/enterpriseNotification';
import KeyValueEditor from '../../common/forms/keyvalueeditor';
import { InfoHintContainer } from '../../common/info-hint';
import Loader from '../../common/loader';
import Time from '../../common/time';
import { HELPTOOLTIPS, MenderHelpTooltip } from '../../helptips/helptooltips';
import ConfigImportDialog from './configimportdialog';
import DeviceDataCollapse from './devicedatacollapse';

const buttonStyle = { marginLeft: 30 };
const iconStyle = { margin: 12 };
const textStyle = { textTransform: 'capitalize', textAlign: 'left' };

const defaultReportTimeStamp = '0001-01-01T00:00:00Z';

const configHelpTipsMap = {
  'mender-demo-raspberrypi-led': {
    position: 'right',
    component: ({ anchor, ...props }) => <MenderHelpTooltip style={anchor} id={HELPTOOLTIPS.configureRaspberryLedTip.id} contentProps={props} />
  },
  timezone: {
    position: 'right',
    component: ({ anchor, ...props }) => <MenderHelpTooltip style={anchor} id={HELPTOOLTIPS.configureTimezoneTip.id} contentProps={props} />
  }
};

export const ConfigUpToDateNote = ({ updated_ts = defaultReportTimeStamp }) => (
  <div className="flexbox margin-small">
    <CheckCircleIcon className="green" style={iconStyle} />
    <div>
      <Typography variant="subtitle2" style={textStyle}>
        Configuration up-to-date on the device
      </Typography>
      <Typography variant="caption" className="muted" style={textStyle}>
        Updated: {<Time value={updated_ts} />}
      </Typography>
    </div>
  </div>
);

export const ConfigEmptyNote = ({ updated_ts = '' }) => (
  <div className="flexbox column margin-small">
    <Typography variant="subtitle2">The device appears to either have an empty configuration or not to have reported a configuration yet.</Typography>
    <Typography variant="caption" className="muted" style={textStyle}>
      Updated: {<Time value={updated_ts} />}
    </Typography>
  </div>
);

export const ConfigEditingActions = ({ hasDeviceConfig, isSetAsDefault, onSetAsDefaultChange, onSubmit, onCancel }) => (
  <>
    <div style={{ maxWidth: 275 }}>
      <FormControlLabel
        control={<Checkbox color="primary" checked={isSetAsDefault} onChange={onSetAsDefaultChange} size="small" />}
        label="Save as default configuration"
        style={{ marginTop: 0 }}
      />
      <div className="muted">You can import these key value pairs when configuring other devices</div>
    </div>
    <Button variant="contained" color="primary" onClick={onSubmit} style={buttonStyle}>
      Save and apply to device
    </Button>
    {hasDeviceConfig && (
      <Button onClick={onCancel} style={buttonStyle}>
        Cancel changes
      </Button>
    )}
  </>
);

export const ConfigUpdateNote = ({ isUpdatingConfig, isAccepted }) => (
  <div>
    <Typography variant="subtitle2" style={textStyle}>
      {!isAccepted
        ? 'Configuration will be applied once the device is connected'
        : isUpdatingConfig
          ? 'Updating configuration on device...'
          : 'Configuration could not be updated on device'}
    </Typography>
    <Typography variant="caption" className="muted" style={textStyle}>
      Status: {isUpdatingConfig || !isAccepted ? 'pending' : 'failed'}
    </Typography>
  </div>
);

export const ConfigUpdateFailureActions = ({ hasLog, onSubmit, onCancel, setShowLog }) => (
  <>
    {hasLog && (
      <Button color="secondary" onClick={setShowLog} style={buttonStyle}>
        View log
      </Button>
    )}
    <Button color="secondary" onClick={onSubmit} startIcon={<RefreshIcon fontSize="small" />} style={buttonStyle}>
      Retry
    </Button>
    <a className="margin-left-large" onClick={onCancel}>
      cancel changes
    </a>
  </>
);

export const DeviceConfiguration = ({ defaultConfig = {}, device: { id: deviceId } }) => {
  const { device, deviceConfigDeployment: deployment } = useSelector(state => getDeviceConfigDeployment(state, deviceId));
  const { hasDeviceConfig } = useSelector(state => getTenantCapabilities(state));
  const { config = {}, status } = device;
  const { configured = {}, deployment_id, reported = {}, reported_ts, updated_ts } = config;
  const isRelevantDeployment = deployment.created > updated_ts && (!reported_ts || deployment.finished > reported_ts);
  const [changedConfig, setChangedConfig] = useState();
  const [editableConfig, setEditableConfig] = useState();
  const [isAborting, setIsAborting] = useState(false);
  const [isEditingConfig, setIsEditingConfig] = useState(false);
  const [isSetAsDefault, setIsSetAsDefault] = useState(false);
  const [isUpdatingConfig, setIsUpdatingConfig] = useState(false);
  const [showConfigImport, setShowConfigImport] = useState(false);
  const [showLog, setShowLog] = useState(false);
  const [updateFailed, setUpdateFailed] = useState();
  const [updateLog, setUpdateLog] = useState();
  const dispatch = useDispatch();
  const deploymentTimer = useRef();

  useEffect(() => {
    if (!isEmpty(config) && !isEmpty(changedConfig) && !isEditingConfig) {
      setIsEditingConfig(isUpdatingConfig || updateFailed);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(config), JSON.stringify(changedConfig), isEditingConfig, isUpdatingConfig, updateFailed]);

  useEffect(() => {
    if (deployment.devices && deployment.devices[device.id]?.log) {
      setUpdateLog(deployment.devices[device.id].log);
    }

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(deployment.devices), device.id]);

  useEffect(() => {
    clearInterval(deploymentTimer.current);
    if (isRelevantDeployment && deployment.status !== DEPLOYMENT_STATES.finished) {
      deploymentTimer.current = setInterval(() => dispatch(getSingleDeployment(deployment_id)), TIMEOUTS.refreshDefault);
    } else if (deployment_id && !isRelevantDeployment) {
      dispatch(getSingleDeployment(deployment_id));
    }
    return () => {
      clearInterval(deploymentTimer.current);
    };
  }, [deployment.status, deployment_id, dispatch, isRelevantDeployment]);

  useEffect(() => {
    if (!isRelevantDeployment) {
      return;
    }
    if (deployment.status === DEPLOYMENT_STATES.finished) {
      // we have to rely on the device stats here as the state change might not have propagated to the deployment status
      // leaving all stats at 0 and giving a false impression of deployment success
      const stats = groupDeploymentStats(deployment);
      const deviceStats = groupDeploymentDevicesStats(deployment);
      const updateFailed = !!((stats.failures || deviceStats.failures) && deployment.device_count);
      setUpdateFailed(updateFailed);
      setIsEditingConfig(updateFailed);
      setIsUpdatingConfig(false);
    } else if (deployment.status) {
      setChangedConfig(configured);
      setEditableConfig(configured);
      // we can't rely on the deployment.status to be !== 'finished' since `deployment` is initialized as an empty object
      // and thus the undefined status would also point to an ongoing update
      setIsUpdatingConfig(true);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(configured), JSON.stringify(deployment.stats), deployment.created, deployment.status, deployment.finished, isRelevantDeployment]);

  useEffect(() => {
    if (!isRelevantDeployment) {
      return;
    }
    if (!changedConfig && !isEmpty(config) && (!deployment_id || deployment.status)) {
      // let currentConfig = reported;
      const stats = groupDeploymentStats(deployment);
      if (deployment.status !== DEPLOYMENT_STATES.finished || stats.failures) {
        setEditableConfig(configured);
        setChangedConfig(configured);
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dispatch, JSON.stringify(config), deployment.status, !changedConfig, JSON.stringify(deployment), isRelevantDeployment]);

  const onConfigImport = ({ config, importType }) => {
    let updatedConfig = config;
    if (importType === 'default') {
      updatedConfig = defaultConfig.current;
    }
    setChangedConfig(updatedConfig);
    setEditableConfig(updatedConfig);
    setShowConfigImport(false);
  };

  const onSetSnackbar = useCallback((...args) => dispatch(setSnackbar(...args)), [dispatch]);

  const onSetAsDefaultChange = () => setIsSetAsDefault(toggle);

  const onShowLog = () =>
    dispatch(getDeviceLog(deployment_id, device.id)).then(result => {
      setShowLog(true);
      setUpdateLog(result[1]);
    });

  const onCancel = () => {
    if (!isEmpty(reported)) {
      setEditableConfig(reported);
      setChangedConfig(reported);
    }
    let requests = [];
    if (deployment_id && deployment.status !== DEPLOYMENT_STATES.finished) {
      requests.push(dispatch(abortDeployment(deployment_id)));
    }
    if (deepCompare(reported, changedConfig)) {
      requests.push(Promise.resolve());
    } else {
      requests.push(dispatch(setDeviceConfig(device.id, reported)));
      if (isSetAsDefault) {
        requests.push(dispatch(saveGlobalSettings({ defaultDeviceConfig: { current: defaultConfig.previous } })));
      }
    }
    return Promise.all(requests).then(() => {
      setIsUpdatingConfig(false);
      setUpdateFailed(false);
      setIsAborting(false);
    });
  };

  const onSubmit = () => {
    Tracking.event({ category: 'devices', action: 'apply_configuration' });
    setIsUpdatingConfig(true);
    setUpdateFailed(false);
    return dispatch(setDeviceConfig(device.id, changedConfig))
      .then(() => dispatch(applyDeviceConfig(device.id, { retries: 0 }, isSetAsDefault, changedConfig)))
      .catch(() => {
        setIsEditingConfig(true);
        setUpdateFailed(true);
        setIsUpdatingConfig(false);
      });
  };

  const onStartEdit = e => {
    e.stopPropagation();
    const nextEditableConfig = { ...configured, ...reported };
    setChangedConfig(nextEditableConfig);
    setEditableConfig(nextEditableConfig);
    setIsEditingConfig(true);
  };

  const onStartImportClick = e => {
    e.stopPropagation();
    setShowConfigImport(true);
  };

  const onAbortClick = () => setIsAborting(toggle);

  const hasDeviceConfiguration = !isEmpty(reported);
  let footer = hasDeviceConfiguration ? <ConfigUpToDateNote updated_ts={reported_ts} /> : <ConfigEmptyNote updated_ts={updated_ts} />;
  if (isEditingConfig) {
    footer = (
      <ConfigEditingActions
        hasDeviceConfig={hasDeviceConfiguration}
        isSetAsDefault={isSetAsDefault}
        onSetAsDefaultChange={onSetAsDefaultChange}
        onSubmit={onSubmit}
        onCancel={onCancel}
      />
    );
  }
  if (isUpdatingConfig || updateFailed) {
    const hasLog = deployment.devices && deployment.devices[device.id]?.log;
    footer = (
      <>
        <div className="flexbox">
          {isUpdatingConfig && <Loader show={true} style={{ marginRight: 15, marginTop: -15 }} />}
          {updateFailed && <ErrorIcon className="red" style={iconStyle} />}
          <ConfigUpdateNote isUpdatingConfig={isUpdatingConfig} isAccepted={status === DEVICE_STATES.accepted} />
        </div>
        {updateFailed ? (
          <ConfigUpdateFailureActions hasLog={hasLog} setShowLog={onShowLog} onSubmit={onSubmit} onCancel={onCancel} />
        ) : isAborting ? (
          <Confirm cancel={onAbortClick} action={onCancel} type="abort" classes="margin-left-large" />
        ) : (
          <>
            <Button color="secondary" onClick={onAbortClick} startIcon={<BlockIcon fontSize="small" />} style={buttonStyle}>
              Abort update
            </Button>
            <Button
              color="secondary"
              component={Link}
              to={`/deployments/${deployment.status || DEPLOYMENT_ROUTES.active.key}?open=true&id=${deployment_id}`}
              style={buttonStyle}
            >
              View deployment
            </Button>
          </>
        )}
      </>
    );
  }

  const helpTipsMap = Object.entries(configHelpTipsMap).reduce((accu, [key, value]) => {
    accu[key] = {
      ...value,
      props: { deviceId: device.id }
    };
    return accu;
  }, {});

  return (
    <DeviceDataCollapse
      isAddOn
      title={
        <div className="two-columns">
          <div className="flexbox center-aligned">
            <h4 className="margin-right">Device configuration</h4>
            {hasDeviceConfig && !(isEditingConfig || isUpdatingConfig) && <EditButton onClick={onStartEdit} />}
          </div>
          <div className="flexbox center-aligned">
            {isEditingConfig ? (
              <Button onClick={onStartImportClick} disabled={isUpdatingConfig} startIcon={<SaveAltIcon />} style={{ justifySelf: 'left' }}>
                Import configuration
              </Button>
            ) : null}
            <InfoHintContainer>
              <EnterpriseNotification id={BENEFITS.deviceConfiguration.id} />
              <MenderHelpTooltip id={HELPTOOLTIPS.configureAddOnTip.id} style={{ marginTop: 5 }} />
              <DocsTooltip id={DOCSTIPS.deviceConfig.id} />
            </InfoHintContainer>
          </div>
        </div>
      }
    >
      <div className="relative">
        {isEditingConfig ? (
          <KeyValueEditor
            disabled={isUpdatingConfig}
            errortext=""
            initialInput={editableConfig}
            inputHelpTipsMap={helpTipsMap}
            onInputChange={setChangedConfig}
          />
        ) : (
          hasDeviceConfig && <ConfigurationObject config={reported} setSnackbar={onSetSnackbar} />
        )}
        {hasDeviceConfig && <div className="flexbox center-aligned margin-bottom margin-top">{footer}</div>}
        {showLog && <LogDialog logData={updateLog} onClose={() => setShowLog(false)} type="configUpdateLog" />}
        {showConfigImport && <ConfigImportDialog onCancel={() => setShowConfigImport(false)} onSubmit={onConfigImport} />}
      </div>
    </DeviceDataCollapse>
  );
};

export default DeviceConfiguration;
