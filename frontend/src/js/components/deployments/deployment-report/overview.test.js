// Copyright 2020 Northern.tech AS
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

import { defaultState } from '../../../../../tests/mockData';
import { render } from '../../../../../tests/setupTests';
import DeploymentOverview from './overview';

const creationDate = '2019-01-01';

describe('DeploymentOverview Component', () => {
  it('renders correctly', async () => {
    const deployment = {
      name: 'test deployment',
      artifact_name: 'test',
      created: creationDate,
      devices: { a: { id: '13' } },
      finished: creationDate,
      group: 'testGroup',
      statistics: { status: {} }
    };
    const { baseElement } = render(<DeploymentOverview devicesById={{}} deployment={deployment} tenantCapabilities={{ hasFullFiltering: true }} />);
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
  });
  it('renders correctly for dynamic groups', async () => {
    const { filters, id: name, ...remainder } = defaultState.devices.groups.byId.testGroupDynamic;
    const filter = { ...remainder, name, filters };
    const deployment = {
      name,
      artifact_name: 'test',
      created: creationDate,
      devices: { a: { id: '13' } },
      filter,
      finished: creationDate,
      statistics: { status: {} }
    };
    const { getByRole } = render(<DeploymentOverview devicesById={{}} deployment={deployment} tenantCapabilities={{ hasFullFiltering: true }} />);
    expect(getByRole('link', { name: RegExp(filter.name, 'i') })).toHaveAttribute('href', '/devices?inventory=group:eq:filter1&system=group:eq:things');
  });
});
