// Copyright 2025 Northern.tech AS
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
import { prettyDOM, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import { DeltaProgress } from './DeltaGeneration';

const preloadedState = {
  ...defaultState,
  app: {
    ...defaultState.app,
    features: {
      ...defaultState.app.features,
      isEnterprise: true
    }
  },
  releases: {
    ...defaultState.releases,
    deltaJobs: {},
    deltaJobsList: {
      jobIds: [],
      total: 0,
      sort: { key: 'started', direction: 'desc' },
      page: 1,
      perPage: 20
    }
  }
};

const preloadedStateWithJobs = {
  ...preloadedState,
  releases: {
    ...preloadedState.releases,
    deltaJobs: {
      'delta-job-1': {
        id: 'delta-job-1',
        to_release: 'mender-demo-artifact-3.3.1',
        from_release: 'mender-demo-artifact-3.2.1',
        device_types_compatible: ['qemux86-64'],
        started: '2022-07-11T20:49:00.000Z',
        status: 'success'
      },
      'delta-job-2': {
        id: 'delta-job-2',
        to_release: 'mender-demo-artifact-3.3.1',
        from_release: 'mender-demo-artifact-3.3.0',
        device_types_compatible: ['raspberrypi0w', 'raspberrypi0-wifi', 'raspberrypi3', 'raspberrypi4'],
        started: '2022-07-11T20:49:00.000Z',
        status: 'failed'
      }
    },
    deltaJobsList: {
      ...preloadedState.releases.deltaJobsList,
      jobIds: ['delta-job-1', 'delta-job-2'],
      total: 2
    }
  }
};

describe('DeltaProgress Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<DeltaProgress />, { preloadedState: preloadedStateWithJobs });
    const view = prettyDOM(baseElement.firstChild, 100000, { highlight: false })
      .replace(/(:?aria-labelledby|id)=":.*:"/g, '')
      .replace(/\\/g, '');
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('shows enterprise upgrade notification for non-enterprise users', async () => {
    render(<DeltaProgress />);
    await expect(screen.getByText(/This feature is not available on your plan/i)).toBeInTheDocument();
    await expect(screen.getByText(/Upgrade/i)).toBeInTheDocument();
  });

  it('shows empty state', async () => {
    render(<DeltaProgress />, { preloadedState });
    await waitFor(() => expect(screen.getByText(/No Delta Artifacts have been generated in the last 30 days/i)).toBeInTheDocument());
  });

  it('shows different status indicators correctly', async () => {
    const ui = <DeltaProgress />;
    render(ui, { preloadedState: preloadedStateWithJobs });
    await waitFor(() => expect(screen.queryByText('Delta Artifact information')).not.toBeInTheDocument());
    await expect(screen.queryByText(/Success/i)).toBeVisible();
    await expect(screen.getByText(/Failed/i)).toBeVisible();
    await expect(screen.getAllByRole('progressbar')).toHaveLength(2);
  });
});

describe('DeltaGenerationDetailsDrawer', () => {
  it('shows job details', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    render(<DeltaProgress />, { preloadedState: preloadedStateWithJobs });
    await user.click(screen.getByText(/success/i));
    await waitFor(() => expect(screen.queryByText('Delta Artifact information')).toBeInTheDocument());
  });
});
