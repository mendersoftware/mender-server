// Copyright 2018 Northern.tech AS
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
import { ReactNode, useCallback, useEffect, useRef, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Link } from 'react-router-dom';

import { AutoAwesomeOutlined as AutoAwesomeIcon } from '@mui/icons-material';
import {
  Button,
  Checkbox,
  FormControl,
  FormControlLabel,
  FormHelperText,
  InputLabel,
  MenuItem,
  Select,
  Switch,
  TextField,
  Typography,
  textFieldClasses
} from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import DocsLink from '@northern.tech/common-ui/DocsLink';
import EnterpriseNotification from '@northern.tech/common-ui/EnterpriseNotification';
import { SupportLink } from '@northern.tech/common-ui/SupportLink';
import { BENEFITS, DEVICE_ONLINE_CUTOFF, TIMEOUTS, alertChannels, settingsKeys } from '@northern.tech/store/constants';
import {
  getDeviceIdentityAttributes,
  getGlobalSettings as getGlobalSettingsSelector,
  getIdAttribute,
  getOfflineThresholdSettings,
  getOrganization,
  getTenantCapabilities,
  getUserCapabilities,
  getUserRoles
} from '@northern.tech/store/selectors';
import { changeNotificationSetting, getDeviceAttributes, getGlobalSettings, saveGlobalSettings } from '@northern.tech/store/thunks';
import { useDebounce } from '@northern.tech/utils/debouncehook';
import { yes } from '@northern.tech/utils/helpers';

import ArtifactGenerationSettings from './ArtifactGeneration';

const maxWidth = 750;

const useStyles = makeStyles()(theme => ({
  formWrapper: { display: 'flex', flexDirection: 'column', gap: theme.spacing(4) },
  threshold: {
    columnGap: theme.spacing(2),
    display: 'grid',
    gridTemplateColumns: '100px 100px',
    marginLeft: 0,
    [`.${textFieldClasses.root}`]: { minWidth: 'auto' }
  }
}));

export const IdAttributeSelection = ({ attributes, dialog = false, onCloseClick, onSaveClick, selectedAttribute = '' }) => {
  const [attributeSelection, setAttributeSelection] = useState('name');

  useEffect(() => {
    setAttributeSelection(selectedAttribute);
  }, [selectedAttribute]);

  const changed = selectedAttribute !== attributeSelection;

  const onChangeIdAttribute = ({ target: { value: attributeSelection } }) => {
    setAttributeSelection(attributeSelection);
    if (dialog) {
      return;
    }
    onSaveClick(undefined, { attribute: attributeSelection, scope: attributes.find(({ value }) => value === attributeSelection).scope });
  };

  const undoChanges = e => {
    setAttributeSelection(selectedAttribute);
    if (dialog) {
      onCloseClick(e);
    }
  };

  const saveSettings = e => onSaveClick(e, { attribute: attributeSelection, scope: attributes.find(({ value }) => value === attributeSelection).scope });

  return (
    <div className="flexbox space-between" style={{ alignItems: 'start', maxWidth }}>
      <div className="flexbox column">
        <FormControl className="margin-top-none">
          <InputLabel id="device-id">Device identity attribute</InputLabel>
          <Select label="Device identity attribute" labelId="device-id" value={attributeSelection} onChange={onChangeIdAttribute}>
            {attributes.map(item => (
              <MenuItem key={item.value} value={item.value}>
                {item.label}
              </MenuItem>
            ))}
          </Select>
          <FormHelperText className="info margin-left-none" component="div">
            <div>Choose a device identity attribute to use to identify your devices throughout the UI.</div>
            <div className={`margin-top-x-small ${dialog ? 'margin-bottom-small' : ''}`}>
              <DocsLink path="client-installation/identity" title="Learn how to add custom identity attributes" /> to your devices.
            </div>
          </FormHelperText>
        </FormControl>
      </div>
      {dialog && (
        <div className="margin-left margin-top flexbox">
          <Button onClick={undoChanges} style={{ marginRight: 10 }}>
            Cancel
          </Button>
          <Button variant="contained" onClick={saveSettings} disabled={!changed} color="primary">
            Save
          </Button>
        </div>
      )}
    </div>
  );
};

const ToggleSetting = ({
  description,
  disabled = false,
  title,
  onClick,
  value
}: {
  description?: string;
  disabled?: boolean;
  onClick: () => void;
  title: string | ReactNode;
  value: boolean;
}) => (
  <div className="flexbox column">
    <FormControl variant="standard">
      <FormControlLabel
        disabled={disabled}
        classes={{ label: 'capitalized-start' }}
        className="align-self-start margin-left-none margin-top-none"
        control={<Switch className="margin-left-small" checked={value} onClick={onClick} />}
        label={title}
        labelPlacement="start"
      />
    </FormControl>
    {!!description && (
      <Typography className="margin-top-x-small" variant="body2">
        {description}
      </Typography>
    )}
  </div>
);

export const GlobalSettingsDialog = ({
  attributes,
  isAdmin,
  notificationChannelSettings,
  offlineThresholdSettings,
  onChangeNotificationSetting,
  onCloseClick,
  onSaveClick,
  saveGlobalSettings,
  selectedAttribute,
  settings,
  tenantCapabilities,
  userCapabilities
}) => {
  const [channelSettings, setChannelSettings] = useState(notificationChannelSettings);
  const [currentInterval, setCurrentInterval] = useState(offlineThresholdSettings.interval);
  const [intervalErrorText, setIntervalErrorText] = useState('');
  const [showDeltaConfig, setShowDeltaConfig] = useState(false);
  const debouncedOfflineThreshold = useDebounce(currentInterval, TIMEOUTS.threeSeconds);
  const timer = useRef(false);
  const { classes } = useStyles();
  const { aiFeatures = {}, needsDeploymentConfirmation = false } = settings;
  const { enabled: isAiEnabled, trainingEnabled: isAiTrainingEnabled } = aiFeatures;
  const { hasMonitor, isEnterprise } = tenantCapabilities;
  const { canManageReleases, canManageUsers } = userCapabilities;
  const { trial: isTrial = true } = useSelector(getOrganization);
  const { hasDelta: hasDeltaArtifactGeneration } = useSelector(state => state.deployments.config) ?? {};

  useEffect(() => {
    setChannelSettings(notificationChannelSettings);
  }, [notificationChannelSettings]);

  useEffect(() => {
    setCurrentInterval(offlineThresholdSettings.interval);
  }, [offlineThresholdSettings.interval]);

  useEffect(() => {
    if (!window.sessionStorage.getItem(settingsKeys.initialized) || !timer.current || !canManageUsers) {
      return;
    }
    saveGlobalSettings({ offlineThreshold: { interval: debouncedOfflineThreshold, intervalUnit: DEVICE_ONLINE_CUTOFF.intervalName }, notify: true });
  }, [canManageUsers, debouncedOfflineThreshold, saveGlobalSettings]);

  useEffect(() => {
    const initTimer = setTimeout(() => (timer.current = true), TIMEOUTS.fiveSeconds);
    return () => {
      clearTimeout(initTimer);
    };
  }, []);

  const onNotificationSettingsClick = useCallback(
    channel => {
      const checked = channelSettings[channel].enabled;
      setChannelSettings({ ...channelSettings, [channel]: { enabled: !checked } });
      onChangeNotificationSetting({ enabled: !checked, channel });
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [JSON.stringify(channelSettings)]
  );

  const onChangeOfflineInterval = ({ target: { validity, value } }) => {
    if (validity.valid) {
      setCurrentInterval(value || 1);
      return setIntervalErrorText('');
    }
    setIntervalErrorText('Please enter a valid number between 1 and 1000.');
  };

  const toggleDeploymentConfirmation = () => saveGlobalSettings({ needsDeploymentConfirmation: !needsDeploymentConfirmation });

  const onEditDeltaClick = () => setShowDeltaConfig(true);

  const onToggleAiClick = useCallback(current => saveGlobalSettings({ aiFeatures: { ...aiFeatures, enabled: !current } }), [aiFeatures, saveGlobalSettings]);

  const onToggleAiTrainingClick = useCallback(
    ({ target: { checked } }) => saveGlobalSettings({ aiFeatures: { ...aiFeatures, trainingEnabled: checked } }),
    [aiFeatures, saveGlobalSettings]
  );

  return (
    <div style={{ maxWidth }} className="margin-top-small">
      <Typography variant="h6">Global settings</Typography>
      <Typography className="margin-top-x-small margin-bottom-large" variant="body2">
        Global settings are applied organization-wide. Modifying these settings will affect all users.
      </Typography>
      <div className={classes.formWrapper}>
        <IdAttributeSelection attributes={attributes} onCloseClick={onCloseClick} onSaveClick={onSaveClick} selectedAttribute={selectedAttribute} />
        {canManageUsers && (
          <ToggleSetting
            title="Deployments confirmation"
            description="Always require confirmation on deployment creation"
            onClick={toggleDeploymentConfirmation}
            value={needsDeploymentConfirmation}
          />
        )}
        {canManageReleases && (
          <div>
            <div className="flexbox center-aligned">
              <ToggleSetting title="Delta Artifacts generation" disabled onClick={yes} value={hasDeltaArtifactGeneration} />
              <EnterpriseNotification className="margin-left-small" id={BENEFITS.deltaGeneration.id} />
            </div>
            <Button className="margin-top-small" disabled={!(isEnterprise && hasDeltaArtifactGeneration)} onClick={onEditDeltaClick} variant="outlined">
              Edit configuration
            </Button>
            {!isEnterprise && (
              <Typography className="margin-top-small" variant="body2">
                Automatic delta artifacts generation is not enabled in your account. If you want to start using this feature, <SupportLink variant="ourTeam" />{' '}
                or <Link to="/subscription">upgrade</Link>
                {isTrial ? '' : ' to Mender Enterprise'}.
              </Typography>
            )}
          </div>
        )}
        {isAdmin &&
          hasMonitor &&
          Object.keys(alertChannels).map(channel => (
            <ToggleSetting
              key={channel}
              value={channelSettings[channel].enabled}
              onClick={() => onNotificationSettingsClick(channel)}
              title={`${channel} notifications`}
              description={`${channel} notifications for deployment and monitoring issues for all users`}
            />
          ))}
        <div>
          <Typography className="margin-bottom-small" variant="subtitle1">
            Offline threshold
          </Typography>
          <FormControl variant="standard">
            <FormControlLabel
              className={classes.threshold}
              control={
                <TextField
                  type="number"
                  onChange={onChangeOfflineInterval}
                  slotProps={{ htmlInput: { min: '1', max: '1000' } }}
                  error={!!intervalErrorText}
                  value={currentInterval}
                  variant="outlined"
                />
              }
              label={<div className="capitalized-start">{DEVICE_ONLINE_CUTOFF.intervalName}</div>}
            />
            {!!intervalErrorText && <FormHelperText className="warning">{intervalErrorText}</FormHelperText>}
            <FormHelperText>Choose how long a device can go without reporting to the server before it is considered “offline”.</FormHelperText>
          </FormControl>
        </div>
        <div>
          <ToggleSetting
            value={isAiEnabled}
            onClick={() => onToggleAiClick(isAiEnabled)}
            title={
              <div className="flexbox center-aligned">
                <AutoAwesomeIcon className="margin-right-x-small" fontSize="small" color={isAiEnabled ? 'secondary' : 'inherit'} />
                <Typography variant="subtitle1">AI features (experimental)</Typography>
              </div>
            }
            description="Enable AI features for all users. We'll try to remove any sensitive details, such as URLs and timestamps, before sending your data for AI analysis. AI features are rate limited to 50 requests per day. "
          />
          <FormControlLabel
            control={<Checkbox disabled={!isAiEnabled} checked={isAiTrainingEnabled} onChange={onToggleAiTrainingClick} />}
            label="Allow us to use data for training"
          />
          <Typography variant="body2">This allows us to enhance the responses you get, collect your feedback, and refine the AI model.</Typography>
        </div>
      </div>
      <ArtifactGenerationSettings open={showDeltaConfig} onClose={() => setShowDeltaConfig(false)} />
    </div>
  );
};

export const GlobalSettingsContainer = ({ closeDialog, dialog }) => {
  const dispatch = useDispatch();
  const attributes = useSelector(getDeviceIdentityAttributes);
  const { isAdmin } = useSelector(getUserRoles);
  const notificationChannelSettings = useSelector(state => state.monitor.settings.global.channels);
  const offlineThresholdSettings = useSelector(getOfflineThresholdSettings);
  const { attribute: selectedAttribute } = useSelector(getIdAttribute);
  const settings = useSelector(getGlobalSettingsSelector);
  const tenantCapabilities = useSelector(getTenantCapabilities);
  const userCapabilities = useSelector(getUserCapabilities);

  const [updatedSettings, setUpdatedSettings] = useState({ ...settings });

  useEffect(() => {
    if (!settings) {
      dispatch(getGlobalSettings());
    }
    dispatch(getDeviceAttributes());
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dispatch, JSON.stringify(settings)]);

  useEffect(() => {
    setUpdatedSettings(current => ({ ...current, ...settings }));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(settings)]);

  const onCloseClick = e => {
    if (dialog) {
      return closeDialog(e);
    }
  };

  const onChangeNotificationSetting = useCallback((...args) => dispatch(changeNotificationSetting(...args)), [dispatch]);
  const onSaveGlobalSettings = useCallback((...args) => dispatch(saveGlobalSettings(...args)), [dispatch]);

  const saveAttributeSetting = (e, id_attribute) =>
    onSaveGlobalSettings({ ...updatedSettings, id_attribute, notify: true }).then(() => {
      if (dialog) {
        closeDialog(e);
      }
    });

  if (dialog) {
    return (
      <IdAttributeSelection
        attributes={attributes}
        dialog
        onCloseClick={onCloseClick}
        onSaveClick={saveAttributeSetting}
        selectedAttribute={selectedAttribute}
      />
    );
  }
  return (
    <GlobalSettingsDialog
      attributes={attributes}
      isAdmin={isAdmin}
      notificationChannelSettings={notificationChannelSettings}
      offlineThresholdSettings={offlineThresholdSettings}
      onChangeNotificationSetting={onChangeNotificationSetting}
      onCloseClick={onCloseClick}
      onSaveClick={saveAttributeSetting}
      saveGlobalSettings={onSaveGlobalSettings}
      settings={settings}
      selectedAttribute={selectedAttribute}
      tenantCapabilities={tenantCapabilities}
      userCapabilities={userCapabilities}
    />
  );
};
export default GlobalSettingsContainer;
