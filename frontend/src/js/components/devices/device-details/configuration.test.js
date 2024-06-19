// Copyright 2021 Northern.tech AS
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

import { act, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { defaultState, undefineds } from '../../../../../tests/mockData';
import { render } from '../../../../../tests/setupTests';
import { TIMEOUTS } from '../../../constants/appConstants';
import Configuration, { ConfigEditingActions, ConfigEmptyNote, ConfigUpToDateNote, ConfigUpdateFailureActions, ConfigUpdateNote } from './configuration';

describe('tiny components', () => {
  [ConfigEditingActions, ConfigUpdateFailureActions, ConfigUpdateNote, ConfigEmptyNote, ConfigUpToDateNote].forEach(async Component => {
    it(`renders ${Component.displayName || Component.name} correctly`, () => {
      const { baseElement } = render(
        <Component
          isAccepted={true}
          isSetAsDefault={true}
          isUpdatingConfig={true}
          onCancel={jest.fn}
          onSetAsDefaultChange={jest.fn}
          onSubmit={jest.fn}
          setShowLog={jest.fn}
          updated_ts="testgroup"
        />
      );
      const view = baseElement.firstChild;
      expect(view).toMatchSnapshot();
      expect(view).toEqual(expect.not.stringMatching(undefineds));
    });
  });
});

describe('Configuration Component', () => {
  let preloadedState;
  const reportedTime = '2019-01-01T09:25:01.000Z';
  it('renders correctly', async () => {
    preloadedState = {
      ...defaultState,
      app: {
        ...defaultState.app,
        features: {
          ...defaultState.app.features,
          hasDeviceConfig: true
        }
      },
      devices: {
        ...defaultState.devices,
        byId: {
          ...defaultState.devices.byId,
          [defaultState.devices.byId.a1.id]: {
            ...defaultState.devices.byId.a1,
            config: {
              configured: { uiPasswordRequired: true, foo: 'bar', timezone: 'GMT+2' },
              reported: { uiPasswordRequired: true, foo: 'bar', timezone: 'GMT+2' },
              updated_ts: defaultState.devices.byId.a1.updated_ts,
              reported_ts: reportedTime
            }
          }
        }
      },
      organization: {
        ...defaultState.organization,
        organization: {
          ...defaultState.organization.organization,
          addons: [{ enabled: true, name: 'configure' }]
        }
      }
    };
    const { baseElement } = render(<Configuration device={preloadedState.devices.byId.a1} />, { preloadedState });
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('works as expected', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    let preloadedState = {
      ...defaultState,
      app: {
        ...defaultState.app,
        features: {
          ...defaultState.app.features,
          hasDeviceConfig: true
        }
      },
      devices: {
        ...defaultState.devices,
        byId: {
          ...defaultState.devices.byId,
          [defaultState.devices.byId.a1.id]: {
            ...defaultState.devices.byId.a1,
            config: {
              configured: {},
              reported: {},
              updated_ts: defaultState.devices.byId.a1.updated_ts,
              reported_ts: reportedTime
            }
          }
        }
      },
      organization: {
        ...defaultState.organization,
        organization: {
          ...defaultState.organization.organization,
          addons: [{ enabled: true, name: 'configure' }]
        }
      }
    };
    let ui = <Configuration device={preloadedState.devices.byId.a1} />;
    const { rerender } = render(ui, { preloadedState });
    expect(screen.queryByRole('button', { name: /import configuration/i })).not.toBeInTheDocument();
    while (screen.queryByRole('button', { name: /edit/i })) {
      await user.click(screen.getByRole('button', { name: /edit/i }));
      await waitFor(() => rerender(ui));
    }
    expect(screen.getByRole('button', { name: /import configuration/i })).toBeInTheDocument();
    const fabButton = document.querySelector('.MuiFab-root');
    expect(fabButton).toBeDisabled();
    await user.type(screen.getByPlaceholderText(/key/i), 'testKey');
    await user.type(screen.getByPlaceholderText(/value/i), 'evilValue');
    expect(fabButton).not.toBeDisabled();
    await user.click(screen.getByRole('checkbox', { name: /save/i }));
    await user.click(screen.getByRole('button', { name: /save/i }));
    await act(async () => jest.runOnlyPendingTimers());
    expect(screen.getByText(/Configuration could not be updated on device/i)).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: /Retry/i }));
    await waitFor(() => rerender(ui));
    ui = <Configuration device={preloadedState.devices.byId.a1} />;
    act(() => jest.advanceTimersByTime(TIMEOUTS.twoSeconds));
    await act(async () => {
      jest.runAllTimers();
      jest.runAllTicks();
    });
    await waitFor(() => rerender(ui));
    await waitFor(() => expect(document.querySelector('.loaderContainer')).not.toBeInTheDocument());
    const valueInput = screen.getByDisplayValue('evilValue');
    await user.clear(valueInput);
    await user.type(valueInput, 'testValue');
    await user.click(screen.getByRole('button', { name: /Retry/i }));
    await waitFor(() => expect(screen.queryByText(/Updating configuration/i)).toBeInTheDocument());
  });
});
