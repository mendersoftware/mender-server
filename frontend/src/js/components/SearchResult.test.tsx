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
import { defaultState, render } from '@/testUtils';
import { ALL_DEVICE_STATES } from '@northern.tech/store/constants';
import { undefineds } from '@northern.tech/testing/mockData';
import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import SearchResult from './SearchResult';

describe('SearchResult Component', () => {
  it('renders correctly', async () => {
    const state = {
      ...defaultState,
      app: {
        ...defaultState.app,
        features: {
          ...defaultState.app.features,
          hasDeviceConfig: true,
          hasDeviceConnect: true,
          hasMultitenancy: true,
          isHosted: true
        },
        searchState: {
          ...defaultState.app.searchState,
          isSearching: true,
          searchTerm: 'something',
          sort: {}
        }
      }
    };
    const { baseElement } = render(<SearchResult onToggleSearchResult={vi.fn} open setSearchState={vi.fn} setSnackbar={vi.fn} />, {
      preloadedState: state
    });
    const view = baseElement;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('navigates to the all devices state when a noauth device is clicked', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const noauthDeviceId = 'noauth-1';
    const state = {
      ...defaultState,
      app: {
        ...defaultState.app,
        searchState: {
          ...defaultState.app.searchState,
          isSearching: false,
          searchTerm: 'foo',
          searchTotal: 1,
          deviceIds: [noauthDeviceId]
        }
      },
      devices: {
        ...defaultState.devices,
        byId: {
          ...defaultState.devices.byId,
          [noauthDeviceId]: {
            ...defaultState.devices.byId.a1,
            id: noauthDeviceId,
            status: 'noauth'
          }
        }
      }
    };
    const onToggleSearchResult = vi.fn();
    const { store } = render(<SearchResult onToggleSearchResult={onToggleSearchResult} open />, {
      preloadedState: state
    });
    await user.click(screen.getByText(noauthDeviceId));
    expect(onToggleSearchResult).toHaveBeenCalled();
    const { state: devicesState, selectedId } = store.getState().devices.deviceList;
    expect(devicesState).toEqual(ALL_DEVICE_STATES);
    expect(selectedId).toEqual(noauthDeviceId);
  });
});
