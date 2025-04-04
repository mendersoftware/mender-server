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

import { EXTERNAL_PROVIDER, TIMEOUTS } from '@northern.tech/store/constants';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import { defaultState, undefineds, webhookEvents } from '../../../../../tests/mockData';
import { render } from '../../../../../tests/setupTests';
import WebhookConfiguration from './Configuration';
import Webhooks from './Webhooks';

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

  it('works as expected', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
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
    const ui = <Webhooks />;
    const { rerender } = render(ui, { preloadedState });
    await user.click(screen.getByText('https://example.com'));
    await waitFor(() => rerender(ui));
    expect(screen.getByText(/webhook details/i)).toBeVisible();
    await user.click(screen.getAllByText(/device status updated/i)[0]);
    await waitFor(() => rerender(ui));
    expect(screen.getByText(/Payload/i)).toBeVisible();
    expect(screen.getByRole('button', { name: /delete webhook/i })).toBeDisabled();
    await user.click(screen.getByText(/back to webhook/i));
    await waitFor(() => rerender(ui));
    await user.click(screen.getByLabelText(/close/i));
    await waitFor(() => expect(screen.queryByText(/webhook details/i)).toBeNull());
  });

  it('can be configured', { timeout: TIMEOUTS.refreshDefault }, async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const onSubmit = vi.fn();
    render(<WebhookConfiguration onSubmit={onSubmit} />);
    expect(screen.getByText(/save/i)).not.toBeEnabled();
    await user.type(screen.getByLabelText(/url/i), 'http://foo.bar');
    await waitFor(() => expect(screen.queryByText(/not protected by HTTPS/i)).toBeInTheDocument());
    await user.clear(screen.getByLabelText(/url/i));
    await user.type(screen.getByLabelText(/url/i), 'https://foo.bar');
    expect(screen.queryByText(/not protected by HTTPS/i)).not.toBeInTheDocument();
    await user.type(screen.getByLabelText(/Description/i), 'https://foo.bar');
    await user.type(screen.getByLabelText(/secret/i), 'https://foo.bar');
    expect(screen.getByText(/has to be entered as a hexadecimal/i)).toBeVisible();
    await user.clear(screen.getByLabelText(/secret/i));
    await waitFor(() => expect(screen.getByRole('button', { name: /save/i })).toBeEnabled());
    await user.click(screen.getByRole('button', { name: /save/i }));
    expect(onSubmit).toHaveBeenCalledWith({
      credentials: { http: { secret: '', url: 'https://foo.bar' }, type: 'http' },
      description: 'https://foo.bar',
      id: 'new',
      provider: 'webhook',
      scopes: []
    });
  });
});
