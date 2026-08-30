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
import storeActions from '@northern.tech/store/actions';
import GeneralApi from '@northern.tech/store/api/general-api';
import { DEPLOYMENT_STATES } from '@northern.tech/store/constants';
import { undefineds } from '@northern.tech/testing/mockData';
import { act, waitFor } from '@testing-library/react';
import { vi } from 'vitest';

import Progress, { minimalRefreshDeploymentsLength } from './InProgressDeployments';

const { receivedDeploymentsForStatus } = storeActions;

// the spy outlives the individual test, so its calls have to be dropped to count the ones of the cycle under test
const spyOnRequests = () => {
  const get = vi.spyOn(GeneralApi, 'get');
  get.mockClear();
  return get;
};

const listRequestsFor = (get, status) =>
  get.mock.calls.filter(([url]) => typeof url === 'string' && url.includes('/deployments?status=') && url.includes(`status=${status}`));

const preloadedState = {
  ...defaultState,
  deployments: {
    ...defaultState.deployments,
    selectionState: {
      ...defaultState.deployments.selectionState,
      inprogress: { ...defaultState.deployments.selectionState.inprogress, selection: ['d1'] },
      pending: { ...defaultState.deployments.selectionState.pending, selection: ['d2'] }
    }
  }
};

describe('InProgressDeployments Component', () => {
  // a refresh cycle left running would keep fetching into the following test & be counted towards its requests
  afterEach(() => vi.clearAllTimers());

  it('renders correctly', async () => {
    const { baseElement } = render(<Progress />, { preloadedState });
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  describe('refresh cycle', () => {
    it('fetches each status once per cycle', async () => {
      const get = spyOnRequests();
      render(<Progress />, { preloadedState });
      await waitFor(() => expect(listRequestsFor(get, DEPLOYMENT_STATES.inprogress)).toHaveLength(1));
      expect(listRequestsFor(get, DEPLOYMENT_STATES.pending)).toHaveLength(1);
    });

    it('extends the interval while the deployments sit still', async () => {
      const get = spyOnRequests();
      render(<Progress />, { preloadedState });
      // holding on to the minimal interval would fetch about 15 times over this span, the growing one only a few
      await act(async () => vi.advanceTimersByTimeAsync(15 * minimalRefreshDeploymentsLength));
      const cycles = listRequestsFor(get, DEPLOYMENT_STATES.inprogress).length;
      expect(cycles).toBeGreaterThan(1);
      expect(cycles).toBeLessThan(8);
    });

    it('shortens the interval again once a deployment moves on', async () => {
      const get = spyOnRequests();
      const { store } = render(<Progress />, { preloadedState });
      // the interval has to have grown first, so that a refresh this soon can only come from the reset
      await act(async () => vi.advanceTimersByTimeAsync(15 * minimalRefreshDeploymentsLength));
      const settled = listRequestsFor(get, DEPLOYMENT_STATES.inprogress).length;
      await act(async () => {
        store.dispatch(receivedDeploymentsForStatus({ deploymentIds: ['d2'], status: DEPLOYMENT_STATES.pending, total: 5 }));
      });
      await act(async () => vi.advanceTimersByTimeAsync(minimalRefreshDeploymentsLength));
      expect(listRequestsFor(get, DEPLOYMENT_STATES.inprogress).length).toBeGreaterThan(settled);
    });
  });
});
