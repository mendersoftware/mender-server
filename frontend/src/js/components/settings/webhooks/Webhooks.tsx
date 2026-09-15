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
import { useMemo, useState } from 'react';
import { useSelector } from 'react-redux';

import { Typography } from '@mui/material';

import DetailsIndicator from '@northern.tech/common-ui/DetailsIndicator';
import DetailsTable from '@northern.tech/common-ui/DetailsTable';
import { EXTERNAL_PROVIDER } from '@northern.tech/store/constants';
import { getWebhooks } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
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
  const dispatch = useAppDispatch();

  const onCancel = () => setSelectedWebhook();

  const onRemoveClick = () => dispatch(deleteIntegration(selectedWebhook)).then(() => setSelectedWebhook());

  const mappedWebhooks = useMemo(
    () => webhooks.map(item => ({ ...item, url: item.credentials?.[EXTERNAL_PROVIDER.webhook.credentialsType]?.url ?? '-' })),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [JSON.stringify(webhooks)]
  );

  if (!mappedWebhooks.length) {
    return null;
  }
  return (
    <div>
      <Typography variant="subtitle1">Webhooks</Typography>
      <DetailsTable columns={columns} items={mappedWebhooks} onItemClick={setSelectedWebhook} />
      <WebhookManagement onCancel={onCancel} onRemove={onRemoveClick} webhook={selectedWebhook} />
    </div>
  );
};

export default Webhooks;
