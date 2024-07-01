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

import { defaultState, undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import {
  AcceptedEmptyState,
  DefaultAttributeRenderer,
  DeviceCreationTime,
  DeviceSoftware,
  DeviceStatusRenderer,
  DeviceTypes,
  PendingEmptyState,
  PreauthorizedEmptyState,
  RejectedEmptyState,
  RelativeDeviceTime,
  defaultTextRender,
  getDeviceIdentityText
} from './base-devices';

describe('smaller components', () => {
  [
    AcceptedEmptyState,
    DefaultAttributeRenderer,
    DeviceCreationTime,
    DeviceSoftware,
    DeviceStatusRenderer,
    DeviceTypes,
    PendingEmptyState,
    PreauthorizedEmptyState,
    RejectedEmptyState,
    RelativeDeviceTime
  ].forEach(Component => {
    it(`renders ${Component.displayName || Component.name} correctly`, () => {
      const { baseElement } = render(
        <Component
          allCount={10}
          device={defaultState.devices.byId.a1}
          filters={[]}
          idAttribute={{ attribute: 'mac', scope: 'identity' }}
          column={{ title: 'mac', attribute: { name: 'mac', scope: 'identity' }, sortable: true, textRender: defaultTextRender }}
          limitMaxed={true}
          onClick={jest.fn}
        />
      );
      const view = baseElement.firstChild.firstChild;
      expect(view).toMatchSnapshot();
      expect(view).toEqual(expect.not.stringMatching(undefineds));
    });
  });

  it(`uses getDeviceIdentityText correctly`, () => {
    let result = getDeviceIdentityText({ device: defaultState.devices.byId.a1 });
    expect(result).toMatch(defaultState.devices.byId.a1.id);
    expect(result).toEqual(expect.not.stringMatching(undefineds));
    result = getDeviceIdentityText({ device: defaultState.devices.byId.a1, idAttribute: { attribute: 'mac', scope: 'identity' } });
    expect(result).toMatch(defaultState.devices.byId.a1.identity_data.mac);
    expect(result).toEqual(expect.not.stringMatching(undefineds));
  });
  it(`uses defaultTextRender correctly`, () => {
    let result = defaultTextRender({ column: { title: 'mac', attribute: { name: 'id', scope: 'identity' } }, device: defaultState.devices.byId.a1 });
    expect(result).toMatch(defaultState.devices.byId.a1.id);
    expect(result).toEqual(expect.not.stringMatching(undefineds));
    result = defaultTextRender({ column: { title: 'mac', attribute: { name: 'mac', scope: 'identity' } }, device: defaultState.devices.byId.a1 });
    expect(result).toMatch(defaultState.devices.byId.a1.identity_data.mac);
    expect(result).toEqual(expect.not.stringMatching(undefineds));
  });
});
