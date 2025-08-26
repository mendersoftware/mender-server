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
import { undefineds } from '@northern.tech/testing/mockData';
import { vi } from 'vitest';

import DeviceConfiguration from './DeviceConfiguration';

describe('DeviceConfiguration Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(
      <DeviceConfiguration
        item={{
          ...defaultState.organization.auditlog.events[2],
          object: { ...defaultState.organization.auditlog.events[2].object, id: defaultState.devices.byId.a1.id },
          change: '{"something":"here"}'
        }}
        onClose={vi.fn}
      />,
      {
        preloadedState: {
          ...defaultState,
          organization: {
            ...defaultState.organization,
            auditlog: {
              ...defaultState.organization.auditlog,
              selectionState: {
                ...defaultState.organization.auditlog.selectionState,
                selectedId: btoa(`${defaultState.organization.auditlog.events[2].action}|${defaultState.organization.auditlog.events[2].time}`)
              }
            }
          }
        }
      }
    );

    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
});
