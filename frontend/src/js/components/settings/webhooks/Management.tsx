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
import type { ReactElement } from 'react';
import { useCallback, useMemo, useRef, useState } from 'react';
import { useSelector } from 'react-redux';

// material ui
import { Circle as CircleIcon } from '@mui/icons-material';
import { Button, Slide } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import BaseDrawer from '@northern.tech/common-ui/BaseDrawer';
import DetailsIndicator from '@northern.tech/common-ui/DetailsIndicator';
import type { ClassesOverrides } from '@northern.tech/common-ui/List';
import Time from '@northern.tech/common-ui/Time';
import { TwoColumnData } from '@northern.tech/common-ui/TwoColumnData';
import actions from '@northern.tech/store/actions';
import type { Event } from '@northern.tech/store/api/types';
import type { Webhook } from '@northern.tech/store/constants';
import { EXTERNAL_PROVIDER, emptyWebhook } from '@northern.tech/store/constants';
import { getTenantCapabilities, getWebhookEventInfo } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { getWebhookEvents } from '@northern.tech/store/thunks';

import WebhookActivity from './Activity';
import { availableScopes } from './Configuration';
import WebhookEventDetails from './EventDetails';

const { setSnackbar } = actions;

const useStyles = makeStyles()(theme => ({
  divider: { marginTop: theme.spacing(), marginBottom: theme.spacing() },
  slide: { gridArea: '1 / 1', minWidth: 0 },
  slideContainer: { display: 'grid', overflowX: 'clip' },
  statusIcon: { fontSize: 12, marginRight: theme.spacing() },
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

  const delivery = delivery_statuses.find(status => status.integration_id === webhook.id);
  if (!delivery) {
    return '-';
  }

  return (
    <div className="flexbox align-items-center">
      <CircleIcon className={`${delivery.success ? 'green' : 'red'} ${classes.statusIcon}`} />
      <div>{delivery.status_code ?? '-'}</div>
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
  const { events, eventsTotal } = useSelector(getWebhookEventInfo);
  const { canDelta: canScopeWebhooks } = useSelector(getTenantCapabilities);
  const dispatch = useAppDispatch();
  const { classes } = useStyles();
  const containerRef = useRef<HTMLDivElement>(null);
  const lastSelectedEvent = useRef<Event>(undefined);

  const onEventSelect = useCallback((event?: Event) => {
    if (event) {
      lastSelectedEvent.current = event;
    }
    setSelectedEvent(event);
  }, []);

  const dispatchedGetWebhookEvents = useCallback(options => dispatch(getWebhookEvents(options)), [dispatch]);
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

  const handleBack = () => onEventSelect(undefined);

  const onCancelClick = () => {
    setSelectedEvent(undefined);
    onCancel();
  };

  return (
    <BaseDrawer
      open={!!webhook?.id}
      onClose={onCancelClick}
      size="lg"
      slotProps={{
        header: {
          title: 'Webhook details',
          preCloser: (
            <Button className={selectedEvent ? 'muted' : ''} color="error" disabled={!!selectedEvent} onClick={() => onRemove(webhook)} variant="outlined">
              Delete webhook
            </Button>
          )
        }
      }}
    >
      <div className={classes.slideContainer} ref={containerRef}>
        <Slide appear={false} in={!selectedEvent} container={() => containerRef.current} direction="right">
          <div className={`${classes.slide} margin-top`}>
            <h4>Settings</h4>
            <TwoColumnData className="margin-top margin-bottom" data={webhookConfig} setSnackbar={dispatchedSetSnackbar} />
            <h4>Activity</h4>
            <WebhookActivity
              classes={classes}
              columns={columns}
              events={events}
              eventsTotal={eventsTotal}
              getWebhookEvents={dispatchedGetWebhookEvents}
              setSelectedEvent={onEventSelect}
              webhook={webhook}
            />
          </div>
        </Slide>
        <Slide in={!!selectedEvent} container={() => containerRef.current} direction="left" mountOnEnter unmountOnExit>
          <div className={`${classes.slide} margin-top`}>
            <WebhookEventDetails
              classes={classes}
              columns={columns}
              entry={selectedEvent ?? lastSelectedEvent.current}
              onClickBack={handleBack}
              setSnackbar={dispatchedSetSnackbar}
              webhook={webhook}
            />
          </div>
        </Slide>
      </div>
    </BaseDrawer>
  );
};

export default WebhookManagement;
