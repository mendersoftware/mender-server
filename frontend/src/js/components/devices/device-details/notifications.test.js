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
import DeviceNotifications, { DeviceOfflineHeaderNotification, LastConnection, NoAlertsHeaderNotification, ServiceNotification } from './notifications';

describe('tiny components', () => {
  [LastConnection, ServiceNotification, NoAlertsHeaderNotification, DeviceOfflineHeaderNotification].forEach(async Component => {
    it(`renders ${Component.displayName || Component.name} correctly`, () => {
      const { baseElement } = render(
        <Component
          alerts={[1, 2]}
          onClick={jest.fn}
          offlineThresholdSettings={{ intervalUnit: 'hour', interval: 24 }}
          check_in_time={defaultState.devices.byId.a1.check_in_time_exact}
        />
      );
      const view = baseElement.firstChild;
      expect(view).toMatchSnapshot();
      expect(view).toEqual(expect.not.stringMatching(undefineds));
    });
  });
});

describe('DeviceNotifications Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(
      <DeviceNotifications
        device={{
          ...defaultState.devices.byId.a1
        }}
        alerts={defaultState.monitor.alerts.byDeviceId.a1.alerts}
        onClick={jest.fn}
      />
    );
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
});
