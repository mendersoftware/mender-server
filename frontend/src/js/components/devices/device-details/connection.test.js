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

import { defaultState, undefineds } from '../../../../../tests/mockData';
import { render } from '../../../../../tests/setupTests';
import { DEVICE_CONNECT_STATES } from '../../../constants/deviceConstants';
import DeviceConnection, { DeviceConnectionMissingNote, DeviceDisconnectedNote, PortForwardLink } from './connection';

describe('tiny DeviceConnection components', () => {
  [DeviceConnectionMissingNote, DeviceDisconnectedNote, PortForwardLink].forEach(async Component => {
    it(`renders ${Component.displayName || Component.name} correctly`, () => {
      const { baseElement } = render(<Component lastConnectionTs={defaultState.devices.byId.a1.updated_ts} />);
      const view = baseElement.firstChild;
      expect(view).toMatchSnapshot();
      expect(view).toEqual(expect.not.stringMatching(undefineds));
    });
  });
});

const preloadedState = {
  ...defaultState,
  app: {
    ...defaultState.app,
    features: {
      ...defaultState.app.features,
      hasAuditlogs: true,
      hasDeviceConnect: true,
      isEnterprise: true
    }
  },
  organization: {
    ...defaultState.organization,
    organization: {
      ...defaultState.organization.organization,
      addons: [{ enabled: true, name: 'troubleshoot' }]
    }
  }
};

describe('DeviceConnection Component', () => {
  let socketSpyFactory;
  const oldMatchMedia = window.matchMedia;

  beforeEach(() => {
    socketSpyFactory = jest.spyOn(window, 'WebSocket');
    socketSpyFactory.mockImplementation(() => ({
      addEventListener: jest.fn(),
      removeEventListener: jest.fn(),
      close: () => {},
      send: () => {}
    }));
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: jest.fn().mockImplementation(query => ({
        matches: false,
        media: query,
        onchange: null,
        addListener: jest.fn(), // Deprecated
        removeListener: jest.fn(), // Deprecated
        addEventListener: jest.fn(),
        removeEventListener: jest.fn(),
        dispatchEvent: jest.fn()
      }))
    });
  });

  afterEach(() => {
    socketSpyFactory.mockReset();
    window.matchMedia = oldMatchMedia;
  });

  it('renders correctly', async () => {
    const { baseElement } = render(<DeviceConnection device={defaultState.devices.byId.a1} />, { preloadedState });
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
  it('renders correctly when disconnected', async () => {
    const { baseElement } = render(<DeviceConnection device={{ ...defaultState.devices.byId.a1, connect_status: DEVICE_CONNECT_STATES.disconnected }} />, {
      preloadedState
    });
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
  it('renders correctly when connected', async () => {
    const { baseElement } = render(<DeviceConnection device={{ ...defaultState.devices.byId.a1, connect_status: DEVICE_CONNECT_STATES.connected }} />, {
      preloadedState
    });
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
});
