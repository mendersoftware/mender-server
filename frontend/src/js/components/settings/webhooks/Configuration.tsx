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
import { Divider, Drawer, buttonClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { DOCSTIPS, DocsTooltip } from '@northern.tech/common-ui/DocsLink';
import { DrawerTitle } from '@northern.tech/common-ui/DrawerTitle';
import EnterpriseNotification from '@northern.tech/common-ui/EnterpriseNotification';
import InfoHint, { InfoHintContainer } from '@northern.tech/common-ui/InfoHint';
import Form from '@northern.tech/common-ui/forms/Form';
import FormCheckbox from '@northern.tech/common-ui/forms/FormCheckbox';
import TextInput from '@northern.tech/common-ui/forms/TextInput';
import { HELPTOOLTIPS, MenderHelpTooltip } from '@northern.tech/helptips/HelpTooltips';
import { BENEFITS, EXTERNAL_PROVIDER, canAccess } from '@northern.tech/store/constants';
import { getTenantCapabilities } from '@northern.tech/store/selectors';

const useStyles = makeStyles()(theme => ({
  buttonWrapper: {
    '&.button-wrapper': { justifyContent: 'initial' },
    [`.${buttonClasses.root}`]: { lineHeight: 'initial' }
  },
  formWrapper: { display: 'flex', flexDirection: 'column', gap: theme.spacing(2), paddingTop: theme.spacing(4) }
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
      {!errors[props.id] && watchedUrl.startsWith('http://') ? (
        <InfoHint content="The endpoint you provided is not protected by HTTPS; all the data will be transferred in plain text" />
      ) : (
        <div />
      )}
    </>
  );
};

const WebhookEventsSelector = ({ canSelectEvents }: { canSelectEvents: boolean }) => (
  <>
    <div className="flexbox center-aligned margin-top">
      <h4 className="margin-none margin-right-small">Webhook Events</h4>
      <InfoHintContainer>
        <EnterpriseNotification id={BENEFITS.webhookEvents.id} />
        <MenderHelpTooltip id={HELPTOOLTIPS.webhookEvents.id} />
      </InfoHintContainer>
    </div>
    <div className="flexbox column margin-left-small">
      {Object.values(availableScopes).map(({ canAccess, id, title }) => (
        <FormCheckbox className="margin-top-none" disabled={!canAccess({ canSelectEvents })} key={id} id={id} label={title} />
      ))}
    </div>
  </>
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
      let webhookConfig = {
        id: 'new',
        provider: EXTERNAL_PROVIDER.webhook.provider,
        credentials: {
          type: EXTERNAL_PROVIDER.webhook.credentialsType,
          [EXTERNAL_PROVIDER.webhook.credentialsType]: { secret: formState.secret, url: formState.url }
        },
        description: formState.description,
        scopes: Object.keys(availableScopes).reduce((accu, scope) => {
          if (formState[scope]) {
            accu.push(scope);
          }
          return accu;
        }, [])
      };
      onSubmit(webhookConfig);
    },
    [onSubmit]
  );

  return (
    <Drawer anchor="right" open PaperProps={{ style: { minWidth: 600, width: '50vw' } }}>
      <DrawerTitle title="Webhook details" postTitle={<MenderHelpTooltip className="margin-left-small" id={HELPTOOLTIPS.webhooks.id} />} onClose={onCancel} />
      <Divider />
      <Form
        className={classes.formWrapper}
        classes={classes}
        defaultValues={defaultValues}
        handleCancel={onCancel}
        id="webhookConfig"
        initialValues={defaultValues}
        onSubmit={onSubmitClick}
        showButtons
        submitLabel="Save"
      >
        <UrlInput label="Url" required id="url" validations="isLength:1,isURL" />
        <TextInput label="Description (optional)" id="description" InputProps={{ multiline: true }} />
        <WebhookEventsSelector canSelectEvents={canSelectEvents} />
        <div className="flexbox">
          <TextInput label="Secret (optional)" id="secret" validations="isHexadecimal" />
          <InfoHintContainer style={{ alignItems: 'center' }}>
            <MenderHelpTooltip id={HELPTOOLTIPS.webhookSecret.id} />
            <DocsTooltip id={DOCSTIPS.webhookSecret.id} />
          </InfoHintContainer>
        </div>
      </Form>
    </Drawer>
  );
};

export default WebhookConfiguration;
