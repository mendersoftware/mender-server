// Copyright 2026 Northern.tech AS
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
import { undefineds } from '@northern.tech/testing/mockData';

import { DeviceSystem } from './DeviceSystem';

const components = {
  R123: [
    { name: 'artifact_name', value: 'rtos-image-v1', scope: 'inventory' },
    { name: 'component_type', value: 'rtos', scope: 'inventory' },
    { name: 'version', value: 'v1', scope: 'inventory' }
  ],
  S789: [
    { name: 'artifact_name', value: 'sensor-firmware-v3', scope: 'inventory' },
    { name: 'component_type', value: 'sensor', scope: 'inventory' },
    { name: 'version', value: 'v3.1.0', scope: 'inventory' }
  ]
};

describe('DeviceSystem Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<DeviceSystem device={{ ...defaultState.devices.byId.a1, components }} />);
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
});
