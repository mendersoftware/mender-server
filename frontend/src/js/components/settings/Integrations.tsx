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
import { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { Button, Divider, FormControl, InputLabel, MenuItem, Select, TextField } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import Confirm from '@northern.tech/common-ui/Confirm';
import InfoHint from '@northern.tech/common-ui/InfoHint';
import { EXTERNAL_PROVIDER, TIMEOUTS } from '@northern.tech/store/constants';
import { getExternalIntegrations, getIsPreview } from '@northern.tech/store/selectors';
import { changeIntegration, createIntegration, deleteIntegration, getIntegrations } from '@northern.tech/store/thunks';
import { useDebounce } from '@northern.tech/utils/debouncehook';
import { customSort } from '@northern.tech/utils/helpers';

import WebhookConfiguration from './webhooks/Configuration';
import Webhooks from './webhooks/Webhooks';

const maxWidth = 750;

const useStyles = makeStyles()(theme => ({
  leftButton: { marginRight: theme.spacing() },
  inputWrapper: { alignItems: 'flex-end' },
  select: { minWidth: 300 },
  formWrapper: { display: 'flex', flexDirection: 'column', gap: theme.spacing(2) },
  textInput: { minWidth: 500, wordBreak: 'break-all' },
  widthLimit: { maxWidth }
}));

const ConnectionDetailsInput = ({ connectionConfig, isEditing, setConnectionConfig }) => {
  const { access_key_id = '', secret_access_key = '', region = '', device_policy_name = '' } = connectionConfig.aws || {};
  const [keyId, setKeyId] = useState(access_key_id);
  const [keySecret, setKeySecret] = useState(secret_access_key);
  const [awsRegion, setRegion] = useState(region);
  const [policy, setPolicy] = useState(device_policy_name);

  const debouncedId = useDebounce(keyId, TIMEOUTS.debounceDefault);
  const debouncedSecret = useDebounce(keySecret, TIMEOUTS.debounceDefault);
  const debouncedRegion = useDebounce(awsRegion, TIMEOUTS.debounceDefault);
  const debounced = useDebounce(policy, TIMEOUTS.debounceDefault);

  const { classes } = useStyles();

  useEffect(() => {
    setConnectionConfig({
      aws: {
        access_key_id: debouncedId,
        secret_access_key: debouncedSecret,
        region: debouncedRegion,
        device_policy_name: debounced
      }
    });
  }, [debounced, debouncedRegion, debouncedId, debouncedSecret, setConnectionConfig]);

  useEffect(() => {
    setKeyId(access_key_id);
    setKeySecret(secret_access_key);
    setRegion(region);
    setPolicy(device_policy_name);
  }, [access_key_id, secret_access_key, region, device_policy_name]);

  const onKeyChange = ({ target: { value = '' } }) => setKeyId(value);
  const onSecretChange = ({ target: { value = '' } }) => setKeySecret(value);
  const onRegionChange = ({ target: { value = '' } }) => setRegion(value);
  const onPolicyChange = ({ target: { value = '' } }) => setPolicy(value);

  const commonProps = { className: classes.textInput, disabled: !isEditing, multiline: true };
  return (
    <div className={classes.formWrapper}>
      <TextField {...commonProps} label="Key ID" onChange={onKeyChange} value={keyId} />
      <TextField {...commonProps} label="Key Secret" onChange={onSecretChange} value={keySecret} />
      <TextField {...commonProps} label="Region" onChange={onRegionChange} value={awsRegion} />
      <TextField {...commonProps} label="Device Policy Name" onChange={onPolicyChange} value={policy} />
    </div>
  );
};

const ConnectionStringInput = ({ connectionConfig, isEditing, setConnectionConfig, title }) => {
  const [value, setValue] = useState(connectionConfig.connection_string);
  const debouncedValue = useDebounce(value, TIMEOUTS.debounceDefault);

  const { classes } = useStyles();

  useEffect(() => {
    setConnectionConfig({ connection_string: debouncedValue });
  }, [debouncedValue, setConnectionConfig]);

  useEffect(() => {
    setValue(connectionConfig.connection_string);
  }, [connectionConfig.connection_string]);

  const updateConnectionConfig = ({ target: { value = '' } }) => setValue(value);

  return (
    <TextField
      className={classes.textInput}
      disabled={!isEditing}
      label={`${title} connection string`}
      multiline
      onChange={updateConnectionConfig}
      value={value}
    />
  );
};

const providerConfigMap = {
  'iot-core': ConnectionDetailsInput,
  'iot-hub': ConnectionStringInput
};

export const IntegrationConfiguration = ({ integration, isLast, onCancel, onDelete, onSave }) => {
  const { credentials = {}, provider } = integration;
  const [connectionConfig, setConnectionConfig] = useState(credentials);
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const { type, ...otherProps } = credentials;
  const [isEditing, setIsEditing] = useState(!Object.values(otherProps).some(i => i));
  const [isDeleting, setIsDeleting] = useState(false);

  const { classes } = useStyles();

  useEffect(() => {
    const { credentials = {} } = integration;
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const { type, ...otherProps } = credentials;
    setConnectionConfig(credentials);
    setIsEditing(!Object.values(otherProps).some(i => i));
  }, [integration]);

  const onCancelClick = () => {
    setIsEditing(false);
    setConnectionConfig(credentials);
    onCancel(integration);
  };
  const onDeleteClick = () => setIsDeleting(true);
  const onDeleteConfirm = () => onDelete(integration);
  const onEditClick = () => setIsEditing(true);
  const onSaveClick = () =>
    onSave({
      ...integration,
      credentials: {
        type: EXTERNAL_PROVIDER[provider].credentialsType,
        ...connectionConfig
      }
    });

  const ConfigInput = providerConfigMap[provider];
  const { configHint, title } = EXTERNAL_PROVIDER[provider];
  return (
    <>
      <h3 className="margin-bottom-none">{title}</h3>
      <div className={`flexbox space-between padding-top-small relative ${classes.widthLimit} ${classes.inputWrapper}`}>
        <ConfigInput connectionConfig={connectionConfig} isEditing={isEditing} setConnectionConfig={setConnectionConfig} title={title} />
        <div className="flexbox margin-bottom-x-small">
          {isEditing ? (
            <>
              <Button className={classes.leftButton} onClick={onCancelClick}>
                Cancel
              </Button>
              <Button variant="contained" onClick={onSaveClick} disabled={credentials === connectionConfig}>
                Save
              </Button>
            </>
          ) : (
            <>
              <Button className={classes.leftButton} onClick={onEditClick}>
                Edit
              </Button>
              <Button onClick={onDeleteClick}>Delete</Button>
            </>
          )}
        </div>
        {isDeleting && <Confirm type="integrationRemoval" classes="confirmation-overlay" action={onDeleteConfirm} cancel={() => setIsDeleting(false)} />}
      </div>
      <InfoHint className={`margin-bottom ${classes.widthLimit}`} content={configHint} />
      {!isLast && <Divider className={`margin-bottom ${classes.widthLimit}`} />}
    </>
  );
};

const determineAvailableIntegrations = (integrations, isPreRelease) =>
  Object.values(EXTERNAL_PROVIDER).reduce((accu, provider) => {
    const hasIntegrationConfigured = integrations.some(integration => integration.provider == provider.provider);
    if (provider.title && (provider.enabled || isPreRelease) && !hasIntegrationConfigured) {
      accu.push(provider);
    }
    return accu;
  }, []);

const IntegrationsContainer = ({ children }: { children: ReactNode }) => (
  <div>
    <h2 className="margin-top-small">Integrations</h2>
    {children}
  </div>
);

export const Integrations = () => {
  const [availableIntegrations, setAvailableIntegrations] = useState([]);
  const [configuredIntegrations, setConfiguredIntegrations] = useState([]);
  const [isConfiguringWebhook, setIsConfiguringWebhook] = useState(false);
  const integrations = useSelector(getExternalIntegrations);
  const isPreRelease = useSelector(getIsPreview);
  const dispatch = useDispatch();

  const { classes } = useStyles();

  useEffect(() => {
    const available = determineAvailableIntegrations(integrations, isPreRelease);
    setAvailableIntegrations(available);
    setConfiguredIntegrations(integrations.filter(integration => integration.provider !== EXTERNAL_PROVIDER.webhook.provider));
  }, [integrations, isPreRelease]);

  useEffect(() => {
    dispatch(getIntegrations());
  }, [dispatch]);

  const onConfigureIntegration = ({ target: { value: provider = '' } }) => {
    if (provider === EXTERNAL_PROVIDER.webhook.provider) {
      return setIsConfiguringWebhook(true);
    }
    setConfiguredIntegrations([...configuredIntegrations, { id: 'new', provider }]);
    setAvailableIntegrations(integrations => integrations.filter(integration => integration.provider !== provider));
  };

  const onCancelClick = ({ id, provider }) => {
    if (id === 'new') {
      setAvailableIntegrations(current => [...current, EXTERNAL_PROVIDER[provider]].sort(customSort(true, 'provider')));
      setConfiguredIntegrations(current =>
        current.filter(
          integration => !(integration.id === id && integration.provider === provider && integration.provider !== EXTERNAL_PROVIDER.webhook.provider)
        )
      );
    }
    setIsConfiguringWebhook(false);
  };

  const onSaveClick = integration => {
    if (integration.id === 'new') {
      setIsConfiguringWebhook(false);
      return dispatch(createIntegration(integration));
    }
    dispatch(changeIntegration(integration));
  };

  const isConfiguring = configuredIntegrations.some(({ id }) => id === 'new');
  if (!!availableIntegrations.length && !integrations.length && !isConfiguring) {
    return (
      <IntegrationsContainer>
        <FormControl>
          <InputLabel id="integration-select-label">Add an integration</InputLabel>
          <Select className={classes.select} label="Add an integration" labelId="integration-select-label" onChange={onConfigureIntegration} value="">
            {availableIntegrations.map(item => (
              <MenuItem key={item.provider} value={item.provider}>
                {item.title}
              </MenuItem>
            ))}
            <MenuItem value="webhook">Webhooks</MenuItem>
          </Select>
        </FormControl>
        {isConfiguringWebhook && <WebhookConfiguration onCancel={onCancelClick} onSubmit={onSaveClick} />}
      </IntegrationsContainer>
    );
  }
  return (
    <IntegrationsContainer>
      {configuredIntegrations.map((integration, index) => (
        <IntegrationConfiguration
          key={integration.provider}
          integration={integration}
          isLast={configuredIntegrations.length === index + 1}
          onCancel={onCancelClick}
          onDelete={integration => dispatch(deleteIntegration(integration))}
          onSave={onSaveClick}
        />
      ))}
      <Webhooks />
      {!isConfiguring && (
        <InfoHint content="You can only have one active integration at a time. To use a different integration, you'll need to delete the current one first." />
      )}
    </IntegrationsContainer>
  );
};

export default Integrations;
