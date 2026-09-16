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
import { useCallback } from 'react';
import { useFormState, useWatch } from 'react-hook-form';
import { useSelector } from 'react-redux';

// material ui
import { Alert, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import BaseDrawer from '@northern.tech/common-ui/BaseDrawer';
import { ContentSection } from '@northern.tech/common-ui/ContentSection';
import DocsLink from '@northern.tech/common-ui/DocsLink';
import EnterpriseNotification from '@northern.tech/common-ui/EnterpriseNotification';
import Form from '@northern.tech/common-ui/forms/Form';
import FormCheckbox from '@northern.tech/common-ui/forms/FormCheckbox';
import TextInput from '@northern.tech/common-ui/forms/TextInput';
import { BENEFITS, EXTERNAL_PROVIDER, canAccess } from '@northern.tech/store/constants';
import { getTenantCapabilities } from '@northern.tech/store/selectors';

import { SETTINGS_FORM_MAX_WIDTH } from '../constants';

const useStyles = makeStyles()(() => ({
  buttonWrapper: {
    '&.button-wrapper': { justifyContent: 'initial' }
  },
  formWrapper: { display: 'flex', flexDirection: 'column', maxWidth: SETTINGS_FORM_MAX_WIDTH }
}));

export const availableScopes = {
  deviceauth: { id: 'deviceauth', title: 'Device authentication', canAccess },
  inventory: { id: 'inventory', title: 'Device inventory', canAccess: ({ canSelectEvents }) => canSelectEvents }
};

const UrlInput = props => {
  const watchedUrl = useWatch({ name: props.id });
  const { errors } = useFormState();
  return (
    <>
      <TextInput {...props} />
      {!errors[props.id] && watchedUrl.startsWith('http://') && (
        <Alert severity="warning">The endpoint you provided is not protected by HTTPS; all the data will be transferred in plain text</Alert>
      )}
    </>
  );
};

const WebhookEventsSelector = ({ canSelectEvents }: { canSelectEvents: boolean }) => (
  <ContentSection title="Webhook Events" postTitle={<EnterpriseNotification id={BENEFITS.webhookEvents.id} />}>
    <Typography variant="body2">
      You can select which type(s) of events the webhook will receive. Device authentication includes when devices are provisioned, decommissioned, or
      authentication states changes.
    </Typography>
    <div className="flexbox column margin-left-small">
      {Object.values(availableScopes).map(({ canAccess, id, title }) => (
        <FormCheckbox className="margin-top-none" disabled={!canAccess({ canSelectEvents })} key={id} id={id} label={title} />
      ))}
    </div>
  </ContentSection>
);

const defaultValues = {
  description: '',
  url: '',
  secret: '',
  ...Object.keys(availableScopes).reduce((accu, scope) => ({ ...accu, [scope]: false }), {})
};

interface WebhookConfigurationObject {
  credentials: {
    type: string;
    webhook: {
      secret: string;
      url: string;
    };
  };
  description: string;
  id: string;
  provider: string;
  scopes: string[];
}

const WebhookConfiguration = ({ onCancel, onSubmit }: { onCancel: () => void; onSubmit: (arg: WebhookConfigurationObject) => void }) => {
  const { canDelta: canSelectEvents } = useSelector(getTenantCapabilities);
  const { classes } = useStyles();

  const onSubmitClick = useCallback(
    formState => {
      const webhookConfig = {
        id: 'new',
        provider: EXTERNAL_PROVIDER.webhook.provider,
        credentials: {
          type: EXTERNAL_PROVIDER.webhook.credentialsType,
          [EXTERNAL_PROVIDER.webhook.credentialsType]: { secret: formState.secret, url: formState.url }
        },
        description: formState.description,
        scopes: Object.keys(availableScopes).filter(scope => formState[scope])
      };
      onSubmit(webhookConfig);
    },
    [onSubmit]
  );

  return (
    <BaseDrawer open onClose={onCancel} size="md" slotProps={{ header: { title: 'Webhook details' } }}>
      <Form
        className={classes.formWrapper}
        classes={classes}
        defaultValues={defaultValues}
        handleCancel={onCancel}
        id="webhookConfig"
        onSubmit={onSubmitClick}
        showButtons
        submitLabel="Save"
      >
        <Typography variant="body2">
          Use webhooks to send data about device lifecycle events to third-party systems. You can have one integration set up at a time.
        </Typography>
        <ContentSection title="Target URL">
          <UrlInput
            hint="URL"
            InputLabelProps={{ shrink: true }}
            required
            requiredRendered={false}
            id="url"
            validations="isLength:1,isURL"
            helperText="This URL will receive the events."
          />
        </ContentSection>
        <ContentSection title="Description">
          <TextInput hint="Description (optional)" InputLabelProps={{ shrink: true }} id="description" InputProps={{ multiline: true }} />
        </ContentSection>
        <WebhookEventsSelector canSelectEvents={canSelectEvents} />
        <ContentSection title="Secret">
          <Typography variant="body2">
            The secret is used for signing the requests sent to your webhook, to verify their authenticity. It is highly recommended for security.{' '}
            <DocsLink path="server-integration/webhooks#signature-header" title="Learn more" /> about secret signature header.
          </Typography>
          <div className="flexbox">
            <TextInput
              hint="Secret (optional)"
              InputLabelProps={{ shrink: true }}
              id="secret"
              validations="isHexadecimal"
              helperText="The secret must be hexadecimal string (including only characters from A-F and 0-9)"
            />
          </div>
        </ContentSection>
      </Form>
    </BaseDrawer>
  );
};

export default WebhookConfiguration;
