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
import { DEVICE_ISSUE_OPTIONS } from '@northern.tech/store/constants';
import { vi } from 'vitest';

import { undefineds } from '../../../../../tests/mockData';
import { render } from '../../../../../tests/setupTests';
import DeviceIssuesSelection from './IssueSelection';

describe('DeviceIssuesSelection Component', () => {
  it('renders correctly', async () => {
    const options = [
      { ...DEVICE_ISSUE_OPTIONS.authRequests, count: 2 },
      { ...DEVICE_ISSUE_OPTIONS.monitoring, count: 0 },
      { ...DEVICE_ISSUE_OPTIONS.offline, count: 8 }
    ];
    const { baseElement } = render(
      <DeviceIssuesSelection classes={{ selection: '' }} onChange={vi.fn} options={options} selection={[DEVICE_ISSUE_OPTIONS.offline.key]} />
    );
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
});
