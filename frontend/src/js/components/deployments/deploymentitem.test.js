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

import { undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import DeploymentItem from './deploymentitem';
import { defaultHeaders as columnHeaders } from './deploymentslist';

describe('DeploymentItem Component', () => {
  it('renders correctly', async () => {
    const deployment = {
      id: 'd1',
      name: 'test deployment',
      artifact_name: 'test',
      created: '2019-01-01T13:30:00.000Z',
      artifacts: ['123'],
      device_count: 1,
      statistics: {
        status: {
          downloading: 0,
          decommissioned: 0,
          failure: 0,
          installing: 1,
          noartifact: 0,
          pending: 0,
          rebooting: 0,
          success: 0,
          'already-installed': 0
        }
      }
    };
    const { container } = render(<DeploymentItem columnHeaders={columnHeaders} deployment={deployment} type="progress" />);
    expect(container.firstChild.firstChild).toMatchSnapshot();
    expect(container).toEqual(expect.not.stringMatching(undefineds));
  });
});
