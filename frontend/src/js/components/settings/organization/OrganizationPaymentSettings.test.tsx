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
import { act } from '@testing-library/react';
import { vi } from 'vitest';

import { defaultState, undefineds } from '../../../../../tests/mockData';
import { render } from '../../../../../tests/setupTests';
import OrganizationPaymentSettings from './OrganizationPaymentSettings';

const preloadedState = {
  organization: {
    ...defaultState.organization,
    card: {
      last4: '1234',
      expiration: { month: 8, year: 1230 },
      brand: 'Visa'
    }
  }
};
describe('OrganizationPaymentSettings Component', () => {
  beforeEach(() => {
    Date.now = vi.fn(() => new Date('2020-07-01T12:00:00.000Z'));
  });
  it('renders correctly', async () => {
    const { baseElement } = render(<OrganizationPaymentSettings />, { preloadedState });
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
    await act(async () => vi.runAllTicks());
  });
});
