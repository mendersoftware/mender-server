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
import { ColumnWidthProvider } from '@northern.tech/common-ui/TwoColumnData';
import { adminUserCapabilities, undefineds } from '@northern.tech/testing/mockData';
import { act, cleanup, prettyDOM, waitFor } from '@testing-library/react';
import { vi } from 'vitest';

import ProgressDeviceList from './DeviceList';

describe('ProgressDeviceList Component', () => {
  afterEach(cleanup);

  it('renders correctly', async () => {
    const getDeploymentDevicesMock = vi.fn().mockResolvedValue(true);
    const ui = (
      <ColumnWidthProvider>
        <ProgressDeviceList
          selectedDevices={Object.values(defaultState.deployments.byId.d1.devices)}
          deployment={defaultState.deployments.byId.d1}
          getDeploymentDevices={getDeploymentDevicesMock}
          userCapabilities={adminUserCapabilities}
          idAttribute={{ attribute: 'foo', scope: 'bar' }}
        />
      </ColumnWidthProvider>
    );
    const { asFragment, rerender } = render(ui);
    await act(async () => vi.advanceTimersByTime(5000));
    await waitFor(() => rerender(ui));

    const view = prettyDOM(asFragment().childNodes[1], 100000, { highlight: false })
      .replace(/(:?aria-labelledby|id)=":.*:"/g, '')
      .replace(/\\/g, '');
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
});
