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

import { defaultState, undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import { EXTERNAL_PROVIDER } from '../../constants/deviceConstants';
import ExpandedDevice from './expanded-device';

const preloadedState = {
  ...defaultState,
  app: {
    ...defaultState.app,
    features: {
      ...defaultState.app.features,
      hasDeviceConfig: true,
      hasDeviceConnect: true,
      hasMonitor: true,
      hasMultitenancy: true
    }
  },
  devices: {
    ...defaultState.devices,
    byId: {
      ...defaultState.devices.byId,
      [defaultState.devices.byId.a1.id]: {
        ...defaultState.devices.byId.a1,
        check_in_time: '2019-01-01T09:25:00.000Z',
        attributes: {
          ...defaultState.devices.byId.a1.attributes,
          mender_is_gateway: true
        }
      }
    },
    groups: {
      ...defaultState.devices.groups,
      selectedGroup: 'testGroup'
    }
  },
  organization: {
    ...defaultState.organization,
    externalDeviceIntegrations: [{ ...EXTERNAL_PROVIDER['iot-hub'], id: 'test' }]
  },
  users: {
    ...defaultState.users,
    onboarding: {
      ...defaultState.onboarding,
      complete: false
    }
  }
};
describe('ExpandedDevice Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<ExpandedDevice deviceId={defaultState.devices.byId.a1.id} setDetailsTab={jest.fn} />, { preloadedState });
    const view = baseElement;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
});
