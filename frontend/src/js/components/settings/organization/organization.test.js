// Copyright 2019 Northern.tech AS
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

import { drawerClasses } from '@mui/material';

import { act, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { defaultState, undefineds } from '../../../../../tests/mockData';
import { render } from '../../../../../tests/setupTests';
import { getSessionInfo } from '../../../auth';
import { CancelSubscriptionAlert, CancelSubscriptionButton, DeviceLimitExpansionNotification, TrialExpirationNote } from './billing';
import MyOrganization, { OrgHeader } from './organization';

describe('MyOrganization Component', () => {
  let preloadedState;
  beforeEach(() => {
    preloadedState = {
      ...defaultState,
      app: {
        ...defaultState.app,
        features: {
          ...defaultState.app.features,
          hasReporting: true,
          isEnterprise: true,
          isHosted: true
        },
        versionInformation: { Integration: '1.2.3' }
      },
      devices: {
        ...defaultState.devices,
        byStatus: {
          ...defaultState.devices.byStatus,
          accepted: {
            ...defaultState.devices.byStatus.accepted,
            total: 1
          }
        },
        limit: null,
        filteringAttributesConfig: {
          attributes: {
            identity: ['something1', 'something2'],
            inventory: ['else1'],
            system: ['entirely1']
          },
          count: 20,
          limit: 100
        }
      },
      organization: {
        ...defaultState.organization,
        card: {
          last4: '1234',
          expiration: { month: 8, year: 1230 },
          brand: 'Visa'
        },
        organization: {
          ...defaultState.organization.organization,
          addons: [
            { enabled: true, name: 'configure' },
            { enabled: true, name: 'monitor' }
          ],
          plan: 'enterprise',
          tenant_token: 'test',
          trial: true,
          trial_expiration: new Date('2021-01-01T00:00:00Z')
        }
      }
    };
    Date.now = jest.fn(() => new Date('2020-07-01T12:00:00.000Z'));
  });

  it('renders correctly', async () => {
    const { baseElement } = render(<MyOrganization />, {
      preloadedState: { ...preloadedState, users: { ...preloadedState.users, currentSession: getSessionInfo() } }
    });
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('supports modifying SSO settings', async () => {
    const config = '<div>not quite right</div>';
    const str = JSON.stringify(config);
    const blob = new Blob([str]);
    const file = new File([blob], 'values.xml', { type: 'application/xml' });
    File.prototype.text = jest.fn().mockResolvedValue(str);
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    const ui = <MyOrganization />;
    const { rerender } = render(ui, { preloadedState: { ...preloadedState, users: { ...preloadedState.users, currentSession: getSessionInfo() } } });
    await waitFor(() => expect(screen.getByText(/text editor/i)).toBeVisible());
    expect(screen.getByText(/text editor/i)).toBeVisible();
    await user.click(screen.getByText(/text editor/i));
    expect(screen.getByText(/import from a file/i)).toBeVisible();
    await act(async () => await user.upload(screen.getByText(/import from a file/i).previousSibling, file));
    await waitFor(() => expect(document.querySelector(`.${drawerClasses.root}`)).toBeVisible());
    await user.click(screen.getByTestId('CloseIcon'));
    await waitFor(() => rerender(ui));
    await waitFor(() => expect(document.querySelector(`.${drawerClasses.root}`)).not.toBeInTheDocument());
    await waitFor(() => expect(screen.getByRole('checkbox')).toBeChecked());
    while (screen.queryByText(/entity id/i)) {
      await user.click(screen.getByRole('checkbox'));
      await user.click(screen.getByRole('button', { name: /save/i }));
    }
  });
});

describe('smaller components', () => {
  [OrgHeader, TrialExpirationNote, DeviceLimitExpansionNotification, CancelSubscriptionAlert, CancelSubscriptionButton].forEach(Component => {
    it(`renders ${Component.displayName || Component.name} correctly`, () => {
      const { baseElement } = render(
        <Component
          trial_expiration="2019-10-05T13:00:00.000Z"
          isTrial={true}
          handleCancelSubscription={jest.fn}
          orgName="test"
          mailBodyTexts={{ billing: 'bill this', upgrade: 'upgrade here' }}
        />
      );
      const view = baseElement.lastChild?.firstChild;
      expect(view).toMatchSnapshot();
      expect(view).toEqual(expect.not.stringMatching(undefineds));
      expect(view).toBeTruthy();
    });
  });
});
