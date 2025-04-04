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
import { ReactElement, useCallback, useMemo, useRef, useState } from 'react';
import { useSelector } from 'react-redux';

// material ui
import { Circle as CircleIcon } from '@mui/icons-material';
import { Button, Divider, Drawer, Slide } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { TwoColumnData } from '@northern.tech/common-ui/ConfigurationObject';
import DetailsIndicator from '@northern.tech/common-ui/DetailsIndicator';
import { DrawerTitle } from '@northern.tech/common-ui/DrawerTitle';
import { ClassesOverrides } from '@northern.tech/common-ui/List';
import Time from '@northern.tech/common-ui/Time';
import actions from '@northern.tech/store/actions';
import { Event } from '@northern.tech/store/api/types/MenderTypes';
import { EXTERNAL_PROVIDER, Webhook, emptyWebhook } from '@northern.tech/store/constants';
import { getTenantCapabilities, getWebhookEventInfo } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { getWebhookEvents } from '@northern.tech/store/thunks';

import WebhookActivity from './Activity';
import { availableScopes } from './Configuration';
import WebhookEventDetails from './EventDetails';

const { setSnackbar } = actions;

const useStyles = makeStyles()(theme => ({
  divider: { marginTop: theme.spacing(), marginBottom: theme.spacing() },
  statusIcon: { fontSize: 12, marginRight: theme.spacing() },
  twoColumnsMultiple: {
    gridTemplateColumns: 'max-content 1fr',
    marginBottom: theme.spacing(2),
    marginTop: theme.spacing(2),
    maxWidth: 'initial'
  },
  wrapper: { justifyContent: 'end' }
}));

const triggerMap = {
  'device-decommissioned': 'Device decommissioned',
  'device-provisioned': 'Device provisioned',
  'device-status-changed': 'Device status updated',
  'device-inventory-changed': 'Device inventory changed'
};

const DeliveryStatus = ({ entry, webhook = {}, classes }) => {
  const { delivery_statuses = [] } = entry;

  const status = useMemo(() => {
    const status = delivery_statuses.find(status => status.integration_id === webhook.id) ?? delivery_statuses[0];
    if (status) {
      return { code: status.status_code, signal: status.success ? 'green' : 'red' };
    }
    return { code: 418, signal: 'disabled' };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(delivery_statuses), webhook.id]);

  return (
    <div className="flexbox center-aligned">
      <CircleIcon className={`${status.signal} ${classes.statusIcon}`} />
      <div className={status.code >= 400 ? 'muted' : ''}>{status.code}</div>
    </div>
  );
};

interface WebhookColumnRenderer extends ClassesOverrides {
  webhook: Webhook;
}

export type WebhookColumns = {
  key: string;
  render: (entry: Event, { webhook, classes }: WebhookColumnRenderer) => ReactElement;
  title: string;
}[];

const columns: WebhookColumns = [
  { key: 'created_ts', title: 'Time', render: entry => <Time value={entry.time} /> },
  { key: 'trigger', title: 'Event trigger', render: entry => <div className="trigger-type">{triggerMap[entry.type] ?? entry.type}</div> },
  { key: 'status', title: 'Status', render: (entry, { webhook, classes }) => <DeliveryStatus classes={classes} entry={entry} webhook={webhook} /> },
  { key: 'details', title: '', render: (_, { classes }) => <DetailsIndicator classes={classes} /> }
];

export const WebhookManagement = ({ onCancel, onRemove, webhook }) => {
  const [selectedEvent, setSelectedEvent] = useState<Event>();
  const { events, eventTotal } = useSelector(getWebhookEventInfo);
  const { canDelta: canScopeWebhooks } = useSelector(getTenantCapabilities);
  const dispatch = useAppDispatch();
  const { classes } = useStyles();
  const containerRef = useRef();

  const dispatchedGetWebhookEvents = useCallback(options => dispatch(getWebhookEvents(options)).unwrap(), [dispatch]);
  const dispatchedSetSnackbar = useCallback(args => dispatch(setSnackbar(args)), [dispatch]);

  const { description, scopes = [], credentials = {} } = webhook ?? emptyWebhook;
  const {
    [EXTERNAL_PROVIDER.webhook.credentialsType]: { url = '', secret = '' }
  } = credentials;

  const webhookConfig = {
    'Destination URL': url,
    'Description': description,
    'Webhook events': scopes?.length
      ? scopes.map(scope => availableScopes[scope].title).join(', ')
      : canScopeWebhooks
        ? 'Backend information unclear'
        : availableScopes.deviceauth.title,
    'Secret': secret
  };

  const handleBack = () => setSelectedEvent();

  const onCancelClick = () => {
    setSelectedEvent();
    onCancel();
  };

  return (
    <Drawer anchor="right" open={!!webhook?.id} PaperProps={{ style: { minWidth: 750, width: '60vw' } }} onClose={onCancelClick}>
      <DrawerTitle
        title="Webhook details"
        preCloser={
          <Button className={selectedEvent ? 'muted' : ''} color="secondary" disabled={!!selectedEvent} onClick={() => onRemove(webhook)}>
            delete webhook
          </Button>
        }
        onClose={onCancelClick}
      />
      <Divider />
      <div className="relative" ref={containerRef}>
        <Slide in={!selectedEvent} container={containerRef.current} direction="right">
          <div className="absolute margin-top full-width" style={{ top: 0 }}>
            <h4>Settings</h4>
            <TwoColumnData className={classes.twoColumnsMultiple} config={webhookConfig} setSnackbar={dispatchedSetSnackbar} />
            <h4>Activity</h4>
            <WebhookActivity
              classes={classes}
              columns={columns}
              events={events}
              eventTotal={eventTotal}
              getWebhookEvents={dispatchedGetWebhookEvents}
              setSelectedEvent={setSelectedEvent}
              webhook={webhook}
            />
          </div>
        </Slide>
        <Slide in={!!selectedEvent} container={containerRef.current} direction="left">
          <div className="absolute margin-top full-width" style={{ top: 0 }}>
            <WebhookEventDetails
              classes={classes}
              columns={columns}
              entry={selectedEvent}
              onClickBack={handleBack}
              setSnackbar={setSnackbar}
              webhook={webhook}
            />
          </div>
        </Slide>
      </div>
    </Drawer>
  );
};

export default WebhookManagement;
