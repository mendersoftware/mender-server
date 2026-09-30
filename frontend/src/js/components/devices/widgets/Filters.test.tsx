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
import { screen } from '@testing-library/react';

import Filters from './Filters';

describe('Filters Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<Filters attributes={[{ key: 'testkey', value: 'testvalue' }]} filters={[]} onFilterChange={() => {}} open={true} />);
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
  it.each([
    { isEnterprise: false, isHosted: false, plan: 'os', showsNotification: true },
    { isEnterprise: false, isHosted: true, plan: 'os', showsNotification: true },
    { isEnterprise: false, isHosted: true, plan: 'professional', showsNotification: true },
    { isEnterprise: true, isHosted: true, plan: 'enterprise', showsNotification: false }
  ])('points to the right plan for further filtering capabilities - %o', async ({ isEnterprise, isHosted, plan, showsNotification }) => {
    const preloadedState = {
      ...defaultState,
      app: { ...defaultState.app, features: { ...defaultState.app.features, isEnterprise, isHosted } },
      organization: { ...defaultState.organization, organization: { ...defaultState.organization.organization, plan } }
    };
    render(<Filters open={true} />, { preloadedState });
    expect(screen.queryByText('Professional')).not.toBeInTheDocument();
    if (showsNotification) {
      expect(screen.getByText('Enterprise')).toBeVisible();
    } else {
      expect(screen.queryByText('Enterprise')).not.toBeInTheDocument();
    }
  });
});
