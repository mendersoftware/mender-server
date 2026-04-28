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
import { useCallback, useEffect, useRef, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Link } from 'react-router-dom';

import { AutoAwesomeOutlined as AutoAwesomeIcon, Edit as EditIcon } from '@mui/icons-material';
import { Button, Checkbox, FormControlLabel, MenuItem, Select, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { DOCSTIPS, DocsTextLink } from '@northern.tech/common-ui/DocsLink';
import EnterpriseNotification from '@northern.tech/common-ui/EnterpriseNotification';
import { SettingsItem } from '@northern.tech/common-ui/SettingsItem';
import { SupportLink } from '@northern.tech/common-ui/SupportLink';
import { ToggleSetting } from '@northern.tech/common-ui/ToggleSetting';
import { NumberField } from '@northern.tech/common-ui/forms/NumberField';
import { BENEFITS, DEVICE_ONLINE_CUTOFF, TIMEOUTS, alertChannels, settingsKeys } from '@northern.tech/store/constants';
import {
  getDeviceIdentityAttributes,
  getFeatures,
  getGlobalSettings as getGlobalSettingsSelector,
  getIdAttribute,
  getIsPreview,
  getOfflineThresholdSettings,
  getOrganization,
  getTenantCapabilities,
  getUserCapabilities,
  getUserRoles
} from '@northern.tech/store/selectors';
import { changeNotificationSetting, getDeviceAttributes, getGlobalSettings, saveGlobalSettings } from '@northern.tech/store/thunks';
import type { Scope } from '@northern.tech/types/MenderTypes';
import { useDebounce } from '@northern.tech/utils/debouncehook';

import ArtifactGenerationSettings from './ArtifactGeneration';

const maxWidth = 750;
const maxOfflineIntervalDays = 1000;

const useStyles = makeStyles()(theme => ({
  formWrapper: { display: 'flex', flexDirection: 'column', gap: theme.spacing(4) }
}));

type DisplayableAttribute = {
  label: string;
  scope: Scope;
  value: string;
};

interface IdAttributeSelectionProps {
  attributes: DisplayableAttribute[];
  onSave: (attribute: { attribute: string; scope: Scope }) => void;
  selectedAttribute?: string;
}

export const IdAttributeSelection = ({ attributes, onSave, selectedAttribute = '' }: IdAttributeSelectionProps) => {
  const onChangeIdAttribute = ({ target: { value } }: { target: { value: string } }) => {
    const match = attributes.find(attr => attr.value === value);
    if (match) {
      onSave({ attribute: value, scope: match.scope });
    }
  };

  return (
    <div style={{ maxWidth }}>
      <SettingsItem
        title="Device identity attribute"
        description={
          <>
            Choose a device identity attribute to represent the devices throughout the UI.{' '}
            <DocsTextLink id={DOCSTIPS.deviceIdentity.id} typographyProps={{ variant: 'body2' }}>
              Learn how to add custom identity attributes.
            </DocsTextLink>
          </>
        }
      />
      <Select className="margin-top-x-small" value={selectedAttribute} onChange={onChangeIdAttribute}>
        {attributes.map(item => (
          <MenuItem key={item.value} value={item.value}>
            {item.label}
          </MenuItem>
        ))}
      </Select>
    </div>
  );
};

export const GlobalSettings = () => {
  const dispatch = useDispatch();
  const attributes = useSelector(getDeviceIdentityAttributes) as DisplayableAttribute[];
  const { isAdmin } = useSelector(getUserRoles);
  const notificationChannelSettings = useSelector(state => state.monitor.settings.global.channels);
  const offlineThresholdSettings = useSelector(getOfflineThresholdSettings);
  const { attribute: selectedAttribute } = useSelector(getIdAttribute);
  const settings = useSelector(getGlobalSettingsSelector);
  const tenantCapabilities = useSelector(getTenantCapabilities);
  const userCapabilities = useSelector(getUserCapabilities);
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
  const { hasAiEnabled } = useSelector(getFeatures);
  const isPreview = useSelector(getIsPreview);

  const dispatchedSaveGlobalSettings = useCallback((...args) => dispatch(saveGlobalSettings(...args)), [dispatch]);

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
    dispatchedSaveGlobalSettings({ offlineThreshold: { interval: debouncedOfflineThreshold, intervalUnit: DEVICE_ONLINE_CUTOFF.intervalName }, notify: true });
  }, [canManageUsers, debouncedOfflineThreshold, dispatchedSaveGlobalSettings]);

  useEffect(() => {
    dispatch(getGlobalSettings());
    dispatch(getDeviceAttributes());
  }, [dispatch]);

  useEffect(() => {
    const initTimer = setTimeout(() => (timer.current = true), TIMEOUTS.fiveSeconds);
    return () => {
      clearTimeout(initTimer);
    };
  }, []);

  const saveAttributeSetting = idAttribute => dispatchedSaveGlobalSettings({ id_attribute: idAttribute, notify: true });

  const onNotificationSettingsClick = useCallback(
    channel => {
      const checked = channelSettings[channel].enabled;
      setChannelSettings({ ...channelSettings, [channel]: { enabled: !checked } });
      dispatch(changeNotificationSetting({ enabled: !checked, channel }));
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [JSON.stringify(channelSettings), dispatch]
  );

  const onChangeOfflineInterval = (value: number | null) => {
    if (value != null && value > 0 && value <= maxOfflineIntervalDays) {
      setCurrentInterval(value);
      return setIntervalErrorText('');
    }
    setIntervalErrorText('Please enter a valid number between 1 and 1000.');
  };

  const toggleDeploymentConfirmation = () => saveGlobalSettings({ needsDeploymentConfirmation: !needsDeploymentConfirmation });

  const onEditDeltaClick = () => setShowDeltaConfig(true);

  const onToggleAiClick = current =>
    dispatchedSaveGlobalSettings({ aiFeatures: { ...aiFeatures, enabled: !current, trainingEnabled: aiFeatures.trainingEnabled || !current } });

  const onToggleAiTrainingClick = ({ target: { checked } }) => dispatchedSaveGlobalSettings({ aiFeatures: { ...aiFeatures, trainingEnabled: checked } });

  return (
    <div style={{ maxWidth }} className="margin-top-small">
      <Typography variant="h6">Global settings</Typography>
      <Typography className="margin-top-x-small margin-bottom-large" variant="body2">
        Global settings are applied organization-wide. Modifying these settings will affect all users.
      </Typography>
      <div className={classes.formWrapper}>
        <IdAttributeSelection attributes={attributes} onSave={saveAttributeSetting} selectedAttribute={selectedAttribute} />
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
            <div className="flexbox align-items-center">
              <Typography variant="subtitle1">Delta Artifacts generation</Typography>
              <EnterpriseNotification className="margin-left-small" id={BENEFITS.deltaGeneration.id} />
            </div>
            <Button
              className="margin-top-x-small"
              disabled={!(isEnterprise && hasDeltaArtifactGeneration)}
              onClick={onEditDeltaClick}
              variant="text"
              endIcon={<EditIcon />}
            >
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
          <Typography className="margin-bottom-small" variant="body2">
            Choose how long (days) a device can go without reporting to the server before it is considered &quot;offline&quot;
          </Typography>
          <NumberField
            min={1}
            max={maxOfflineIntervalDays}
            inputStyle={{ width: 120 }}
            onValueChange={onChangeOfflineInterval}
            error={!!intervalErrorText}
            value={currentInterval}
            helperText={intervalErrorText}
          />
        </div>
        {(isPreview || hasAiEnabled) && (
          <div>
            <ToggleSetting
              value={isAiEnabled}
              onClick={() => onToggleAiClick(isAiEnabled)}
              title={
                <div className="flexbox align-items-center">
                  <AutoAwesomeIcon className="margin-right-x-small" fontSize="small" color={isAiEnabled ? 'secondary' : 'inherit'} />
                  <Typography variant="subtitle1">AI features (experimental)</Typography>
                </div>
              }
              description="Enable AI features for all users. We'll try to remove any sensitive details, such as URLs and timestamps, before sending your data for AI analysis. AI features are rate limited to 50 requests per day. "
            />
            <FormControlLabel
              control={
                <Checkbox key={`aiEnabled-${isAiTrainingEnabled}`} disabled={!isAiEnabled} checked={isAiTrainingEnabled} onChange={onToggleAiTrainingClick} />
              }
              label="Allow us to use data for training"
            />
            <Typography variant="body2">This allows us to enhance the responses you get, collect your feedback, and refine the AI model.</Typography>
          </div>
        )}
      </div>
      <ArtifactGenerationSettings open={showDeltaConfig} onClose={() => setShowDeltaConfig(false)} />
    </div>
  );
};
