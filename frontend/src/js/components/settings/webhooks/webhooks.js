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
import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

// material ui
import { ArrowRightAlt as ArrowRightAltIcon } from '@mui/icons-material';

import { changeIntegration, createIntegration, deleteIntegration, getIntegrations, getWebhookEvents } from '../../../actions/organizationActions';
import { EXTERNAL_PROVIDER } from '../../../constants/deviceConstants';
import { emptyWebhook } from '../../../constants/organizationConstants';
import DetailsTable from '../../common/detailstable';
import DocsLink from '../../common/docslink';
import Time from '../../common/time';
import WebhookManagement from './management';

const columns = [
  { key: 'url', title: 'URL', render: ({ url }) => url },
  { key: 'status', title: 'Status', render: ({ status }) => status },
  {
    key: 'updated_ts',
    title: 'Last activity',
    render: ({ updated_ts }) => <Time value={updated_ts} />
  },
  {
    key: 'manage',
    title: 'Manage',
    render: () => (
      <div className="bold flexbox center-aligned link-color margin-right-small uppercased" style={{ whiteSpace: 'nowrap' }}>
        view details <ArrowRightAltIcon />
      </div>
    )
  }
];

export const Webhooks = ({ webhook = { ...emptyWebhook } }) => {
  const [adding, setAdding] = useState(false);
  const [editing, setEditing] = useState(false);
  const [selectedWebhook, setSelectedWebhook] = useState(webhook);
  const dispatch = useDispatch();
  const { events, webhooks } = useSelector(state => {
    const webhooks = state.organization.externalDeviceIntegrations.filter(
      integration => integration.id && integration.provider === EXTERNAL_PROVIDER.webhook.provider
    );
    const events = webhooks.length ? state.organization.webhooks.events : [];
    return { events, webhooks };
  });
  const eventTotal = useSelector(state => state.organization.webhooks.eventTotal);

  useEffect(() => {
    dispatch(getIntegrations());
  }, [dispatch]);

  useEffect(() => {
    setSelectedWebhook(webhook);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(webhook)]);

  const onEdit = item => {
    setAdding(false);
    setEditing(true);
    setSelectedWebhook(item);
  };

  const onCancel = () => {
    setAdding(false);
    setEditing(false);
  };

  const onSubmit = item => {
    if (adding) {
      dispatch(createIntegration(item));
    } else {
      dispatch(changeIntegration(item));
    }
    setAdding(false);
    setEditing(false);
  };

  const onRemoveClick = () => dispatch(deleteIntegration(selectedWebhook));

  const { mappedWebhooks, relevantColumns } = useMemo(() => {
    const mappedWebhooks = webhooks.map(item => ({ ...item, url: item.credentials[EXTERNAL_PROVIDER.webhook.credentialsType].url, status: 'enabled' }));
    const relevantColumns = columns.reduce((accu, item) => {
      if (mappedWebhooks.some(hook => hook[item.key]) || item === columns[columns.length - 1]) {
        accu.push(item);
      }
      return accu;
    }, []);
    return { mappedWebhooks, relevantColumns };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(webhooks)]);

  const dispatchedGetWebhookEvents = useCallback(options => dispatch(getWebhookEvents(options)), [dispatch]);

  return (
    <div>
      <h2>Webhooks</h2>
      {webhooks.length ? (
        <DetailsTable columns={relevantColumns} items={mappedWebhooks} onItemClick={onEdit} />
      ) : (
        <div className="flexbox centered">
          No webhooks are configured yet. Learn more about webhooks in our <DocsLink path="server-integration" title="documentation" />
        </div>
      )}
      <WebhookManagement
        adding={adding}
        editing={editing}
        events={events}
        eventTotal={eventTotal}
        getWebhookEvents={dispatchedGetWebhookEvents}
        onCancel={onCancel}
        onSubmit={onSubmit}
        onRemove={onRemoveClick}
        webhook={selectedWebhook}
      />
    </div>
  );
};

export default Webhooks;
