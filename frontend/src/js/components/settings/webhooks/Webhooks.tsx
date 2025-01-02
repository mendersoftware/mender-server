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
import React, { useMemo, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import DetailsIndicator from '@northern.tech/common-ui/DetailsIndicator';
import DetailsTable from '@northern.tech/common-ui/DetailsTable';
import DocsLink from '@northern.tech/common-ui/DocsLink';
import { EXTERNAL_PROVIDER } from '@northern.tech/store/constants';
import { getWebhooks } from '@northern.tech/store/selectors';
import { deleteIntegration } from '@northern.tech/store/thunks';

import WebhookManagement from './Management';

const columns = [
  { key: 'url', title: 'URL', render: ({ url }) => url },
  { key: 'description', title: 'Description', render: ({ description }) => description },
  { key: 'manage', title: 'Manage', render: DetailsIndicator }
];

export const Webhooks = () => {
  const [selectedWebhook, setSelectedWebhook] = useState();
  const webhooks = useSelector(getWebhooks);
  const dispatch = useDispatch();

  const onCancel = () => setSelectedWebhook();

  const onRemoveClick = () => dispatch(deleteIntegration(selectedWebhook)).then(() => setSelectedWebhook());

  const mappedWebhooks = useMemo(
    () => webhooks.map(item => ({ ...item, url: item.credentials[EXTERNAL_PROVIDER.webhook.credentialsType].url, status: 'enabled' })),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [JSON.stringify(webhooks)]
  );

  if (!mappedWebhooks.length) {
    return null;
  }
  return (
    <div>
      <h2>Webhooks</h2>
      {webhooks.length ? (
        <DetailsTable columns={columns} items={mappedWebhooks} onItemClick={setSelectedWebhook} />
      ) : (
        <div className="flexbox centered">
          No webhooks are configured yet. Learn more about webhooks in our <DocsLink path="server-integration" title="documentation" />
        </div>
      )}
      <WebhookManagement onCancel={onCancel} onRemove={onRemoveClick} webhook={selectedWebhook} />
    </div>
  );
};

export default Webhooks;
