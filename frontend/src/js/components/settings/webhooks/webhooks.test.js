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
import React from 'react';

import { defaultState, undefineds, webhookEvents } from '../../../../../tests/mockData';
import { render } from '../../../../../tests/setupTests';
import { EXTERNAL_PROVIDER } from '../../../constants/deviceConstants';
import Activity from './activity';
import { WebhookCreation } from './configuration';
import Management from './management';
import Webhooks from './webhooks';

describe('Webhooks Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<Webhooks />);
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('renders correctly with entries ', async () => {
    const preloadedState = {
      ...defaultState,
      organization: {
        ...defaultState.organization,
        externalDeviceIntegrations: [
          {
            id: '1',
            credentials: { [EXTERNAL_PROVIDER.webhook.credentialsType]: { url: 'https://example.com' } },
            provider: EXTERNAL_PROVIDER.webhook.provider
          }
        ],
        webhooks: {
          ...defaultState.organization.webhooks,
          events: webhookEvents
        }
      }
    };
    const { baseElement } = render(<Webhooks />, { preloadedState });
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  describe('static components', () => {
    const props = {
      adding: true,
      editing: true,
      events: webhookEvents,
      getWebhookEvents: jest.fn,
      onCancel: jest.fn,
      onSubmit: jest.fn,
      onRemove: jest.fn
    };
    [Activity, Management, WebhookCreation].forEach(Component => {
      it(`renders ${Component.displayName || Component.name} correctly`, () => {
        const { baseElement } = render(<Component {...props} />);
        const view = baseElement;
        expect(view).toMatchSnapshot();
        expect(view).toEqual(expect.not.stringMatching(undefineds));
      });
    });
  });
});
