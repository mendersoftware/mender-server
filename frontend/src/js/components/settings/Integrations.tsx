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
import type { ReactNode } from 'react';
import { useEffect, useState } from 'react';
import { FormProvider, useForm } from 'react-hook-form';
import { useSelector } from 'react-redux';

import { Alert, Button, Divider, FormControl, InputLabel, MenuItem, Select, FormHelperText, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { ConfirmModal } from '@northern.tech/common-ui/ConfirmModal';
import TextInput from '@northern.tech/common-ui/forms/TextInput';
import { EXTERNAL_PROVIDER } from '@northern.tech/store/constants';
import { getExternalIntegrations, getIsPreview } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { changeIntegration, createIntegration, deleteIntegration, getIntegrations } from '@northern.tech/store/thunks';
import type { Integration } from '@northern.tech/types/MenderTypes';
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
  confirmationWrapper: { height: 50 }, // roughly larger than even a large sized "delete" button
  widthLimit: { maxWidth }
}));

const ConnectionDetailsInput = ({ isEditing }) => {
  const { classes } = useStyles();

  const commonProps = { className: classes.textInput, disabled: !isEditing, width: SETTINGS_INPUT_WIDTH };
  return (
    <div className={classes.formWrapper}>
      <TextInput {...commonProps} label="Key ID" id="aws.access_key_id" />
      <TextInput {...commonProps} label="Key Secret" id="aws.secret_access_key" />
      <TextInput {...commonProps} label="Region" id="aws.region" />
      <TextInput {...commonProps} label="Device Policy Name" id="aws.device_policy_name" />
    </div>
  );
};

const ConnectionStringInput = ({ isEditing, title }) => {
  const { classes } = useStyles();
  return (
    <TextInput
      className={classes.textInput}
      disabled={!isEditing}
      id="connection_string"
      InputProps={{ multiline: true }}
      label={`${title} connection string`}
      width={SETTINGS_INPUT_WIDTH}
    />
  );
};

const providerConfigMap = {
  'iot-core': {
    Component: ConnectionDetailsInput,
    getFormValues: ({ aws }) => {
      const { access_key_id = '', secret_access_key = '', region = '', device_policy_name = '' } = aws || {};
      return { aws: { access_key_id, secret_access_key, region, device_policy_name } };
    }
  },
  'iot-hub': {
    Component: ConnectionStringInput,
    getFormValues: ({ connection_string = '' }) => ({ connection_string })
  }
};
const isUnconfigured = ({ type: _type, ...remainder }: Record<string, unknown> = {}) => !Object.values(remainder).some(i => i);

export const IntegrationConfiguration = ({ integration, isLast, onCancel, onDelete, onSave }) => {
  const { credentials = {}, provider } = integration;
  const [isEditing, setIsEditing] = useState(isUnconfigured(credentials));
  const [isDeleting, setIsDeleting] = useState(false);

  const { classes } = useStyles();

  const { Component: ConfigInput, getFormValues } = providerConfigMap[provider];
  const methods = useForm({ defaultValues: getFormValues(credentials) });
  const {
    formState: { isDirty },
    handleSubmit,
    reset
  } = methods;

  useEffect(() => {
    const { credentials = {} } = integration;
    reset(getFormValues(credentials));
    setIsEditing(isUnconfigured(credentials));
  }, [getFormValues, integration, reset]);

  const onCancelClick = () => {
    setIsEditing(false);
    reset();
    onCancel(integration);
  };
  const onDeleteClick = () => setIsDeleting(true);
  const onDeleteConfirm = () => onDelete(integration);
  const onEditClick = () => setIsEditing(true);
  const onSaveClick = handleSubmit(config =>
    onSave({
      ...integration,
      credentials: {
        type: EXTERNAL_PROVIDER[provider].credentialsType,
        ...config
      }
    })
  );

  const { configHint, title } = EXTERNAL_PROVIDER[provider];
  return (
    <FormProvider {...methods}>
      <Typography variant="subtitle1" className="margin-bottom-none">
        {title}
      </Typography>
      <div className={`flexbox column align-items-start padding-top-small ${classes.widthLimit}`}>
        <ConfigInput isEditing={isEditing} title={title} />
        <FormHelperText>{configHint}</FormHelperText>
        <div className="margin-bottom-x-small margin-top-small">
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
        {isDeleting && (
          <div className={`absolute full-width ${classes.confirmationWrapper}`}>
            <Confirm type="integrationRemoval" action={onDeleteConfirm} cancel={() => setIsDeleting(false)} />
          </div>
        )}
      </div>
      <InfoHint className={`margin-bottom ${classes.widthLimit}`} content={configHint} />
      {!isLast && <Divider className={`margin-bottom ${classes.widthLimit}`} />}
    </FormProvider>
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
  const dispatch = useAppDispatch();

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

  const onSaveClick = async (integration: Integration) => {
    try {
      if (integration.id === 'new') {
        await dispatch(createIntegration(integration)).unwrap();
        setIsConfiguringWebhook(false);
        return;
      }
      await dispatch(changeIntegration(integration)).unwrap();
    } catch {
      // error already handled in thunk - leave open
    }
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
