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

import { act } from '@testing-library/react';

import { defaultState, undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import { EXTERNAL_PROVIDER } from '../../constants/deviceConstants';
import { IntegrationConfiguration, Integrations } from './integrations';

const integrations = [
  {
    id: 'iot-hub',
    provider: 'iot-hub',
    credentials: { type: EXTERNAL_PROVIDER['iot-hub'].credentialsType, connection_string: 'something' }
  },
  {
    id: 'iot-core',
    provider: 'iot-core',
    credentials: { type: EXTERNAL_PROVIDER['iot-core'].credentialsType, aws: 'something else' }
  }
];

const preloadedState = { ...defaultState, organization: { ...defaultState.organization, externalDeviceIntegrations: integrations } };

describe('IntegrationConfiguration Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(
      <IntegrationConfiguration integration={{ ...integrations[0], connection_string: '' }} onCancel={jest.fn} onDelete={jest.fn} onSave={jest.fn} />
    );
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
    await act(async () => {
      jest.runOnlyPendingTimers();
      jest.runAllTicks();
    });
  });
});

describe('Integrations Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<Integrations />, { preloadedState });
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
    await act(async () => {
      jest.runOnlyPendingTimers();
      jest.runAllTicks();
    });
  });
});
