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
import { Provider } from 'react-redux';
import { MemoryRouter } from 'react-router-dom';

import { ThemeProvider, createTheme } from '@mui/material/styles';

import { defaultState } from '@/testUtils';
import { DEVICE_FILTERING_OPTIONS } from '@northern.tech/store/constants';
import { getConfiguredStore } from '@northern.tech/store/store';
import { undefineds } from '@northern.tech/testing/mockData';
import { light } from '@northern.tech/testing/theme/light';
import { ATTRIBUTE_SCOPES } from '@northern.tech/utils/constants';
import { act, prettyDOM, render, waitFor } from '@testing-library/react';
import { vi } from 'vitest';

import DeviceGroups from './DeviceGroups';

const theme = createTheme(light);

const preloadedState = {
  ...defaultState,
  devices: {
    ...defaultState.devices,
    groups: {
      ...defaultState.devices.groups,
      selectedGroup: 'testGroup'
    },
    deviceList: {
      ...defaultState.devices.deviceList,
      deviceIds: defaultState.devices.byStatus.accepted.deviceIds
    }
  }
};

const renderWithRouter = (ui: React.ReactElement, { route = '/', preloadedState: state = preloadedState } = {}) => {
  const store = getConfiguredStore({ preloadedState: state });
  const Wrapper = ({ children }) => (
    <ThemeProvider theme={theme}>
      <MemoryRouter initialEntries={[route]}>
        <Provider store={store}>{children}</Provider>
      </MemoryRouter>
    </ThemeProvider>
  );
  return { store, ...render(ui, { wrapper: Wrapper }) };
};

describe('DeviceGroups Component', () => {
  it('renders correctly', async () => {
    const route = `/ui/devices/accepted?inventory=group:eq:${preloadedState.devices.groups.selectedGroup}`;
    const { baseElement } = renderWithRouter(<DeviceGroups />, { route, preloadedState });
    // special snapshot handling here to work around unstable ids in mui code...
    const view = prettyDOM(baseElement.firstChild, 100000, { highlight: false })
      .replace(/(:?aria-labelledby|id)=":.*:"/g, '')
      .replace(/\\/g, '');
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
    await act(async () => vi.runAllTicks());
  });

  it('applies id filter with correct scope when navigating to URL with multiple device ids', async () => {
    const deviceIds = [preloadedState.devices.byId.a1.id, preloadedState.devices.byId.b1.id];
    const route = `/ui/devices?id=${deviceIds[0]}&id=${deviceIds[1]}`;

    const enterpriseState = {
      ...defaultState,
      app: {
        ...defaultState.app,
        features: {
          ...defaultState.app.features,
          isEnterprise: true
        }
      },
      devices: {
        ...defaultState.devices,
        filters: [],
        deviceList: {
          ...defaultState.devices.deviceList,
          deviceIds: []
        }
      }
    };

    const { store } = renderWithRouter(<DeviceGroups />, { route, preloadedState: enterpriseState });

    await waitFor(() => {
      const state = store.getState();
      const idFilter = state.devices.filters.find(filter => filter.key === 'id' && filter.operator === DEVICE_FILTERING_OPTIONS.$in.key);
      expect(idFilter).toBeDefined();
      expect(idFilter!.scope).toBe(ATTRIBUTE_SCOPES.inventory);
      expect(idFilter!.value).toEqual(deviceIds);
    });
  });
});
