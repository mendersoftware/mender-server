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
import { act, cleanup, prettyDOM, waitFor } from '@testing-library/react';
import { vi } from 'vitest';

import { defaultState } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import DeploymentReport from './Report';

describe('DeploymentReport Component', () => {
  afterEach(cleanup);

  it('renders correctly', async () => {
    const ui = <DeploymentReport type="finished" open={true} />;
    const { asFragment, rerender } = render(ui, {
      preloadedState: {
        ...defaultState,
        deployments: {
          ...defaultState.deployments,
          selectedDeviceIds: [defaultState.deployments.byId.d1.devices.a1.id],
          selectionState: {
            selectedId: defaultState.deployments.byId.d1.id
          }
        }
      }
    });
    act(() => vi.advanceTimersByTime(5000));
    await waitFor(() => rerender(ui));
    const view = prettyDOM(asFragment().childNodes[1], 100000, { highlight: false })
      .replace(/(:?aria-labelledby|id)=":.*:"/g, '')
      .replace(/\\/g, '');
    expect(view).toMatchSnapshot();
  });
  it('renders correctly for phased inprogress', async () => {
    const ui = <DeploymentReport type="inprogress" open={true} />;
    const { asFragment, rerender } = render(ui, {
      preloadedState: {
        ...defaultState,
        deployments: {
          ...defaultState.deployments,
          selectedDeviceIds: [defaultState.deployments.byId.d1.devices.a1.id],
          selectionState: {
            selectedId: defaultState.deployments.byId.d3.id
          }
        }
      }
    });
    act(() => vi.advanceTimersByTime(5000));
    await waitFor(() => rerender(ui));
    const view = prettyDOM(asFragment().childNodes[1], 100000, { highlight: false })
      .replace(/(:?aria-labelledby|id)=":.*:"/g, '')
      .replace(/\\/g, '');
    expect(view).toMatchSnapshot();
  });
});
