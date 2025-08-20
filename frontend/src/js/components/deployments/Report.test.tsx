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
import { TIMEOUTS } from '@northern.tech/store/constants';
import { act, cleanup, prettyDOM, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import DeploymentReport from './Report';

const deploymentsState = {
  ...defaultState.deployments,
  selectedDeviceIds: [defaultState.deployments.byId.d1.devices.a1.id],
  selectionState: {
    selectedId: defaultState.deployments.byId.d1.id
  },
  byId: {
    ...defaultState.deployments.byId,
    d1: {
      ...defaultState.deployments.byId.d1,
      devices: {
        ...defaultState.deployments.byId.d1.devices,
        a1: {
          ...defaultState.deployments.byId.d1.devices.a1,
          log: 'Sample deployment log content'
        }
      }
    }
  }
};

const usersState = {
  ...defaultState.users,
  globalSettings: { ...defaultState.users.globalSettings, aiFeatures: { enabled: true } }
};

describe('DeploymentReport Component', () => {
  afterEach(cleanup);

  it('renders correctly', async () => {
    const ui = <DeploymentReport type="finished" open={true} />; // default render happens with empty version information, so it defaults to 'next' and shows the ai log button
    const { asFragment } = render(ui, { preloadedState: { ...defaultState, deployments: deploymentsState } });
    act(() => vi.advanceTimersByTime(TIMEOUTS.fiveSeconds));
    const view = prettyDOM(asFragment().childNodes[1], 100000, { highlight: false })
      .replace(/(:?aria-labelledby|id)=":.*:"/g, '')
      .replace(/\\/g, '');
    expect(view).toMatchSnapshot();
  });
  it('renders correctly for phased inprogress', async () => {
    const ui = <DeploymentReport type="inprogress" open={true} />;
    const { asFragment } = render(ui, {
      preloadedState: {
        ...defaultState,
        deployments: {
          ...deploymentsState,
          selectionState: {
            selectedId: defaultState.deployments.byId.d3.id
          }
        }
      }
    });
    const view = prettyDOM(asFragment().childNodes[1], 100000, { highlight: false })
      .replace(/(:?aria-labelledby|id)=":.*:"/g, '')
      .replace(/\\/g, '');
    expect(view).toMatchSnapshot();
  });

  describe('AI Log Analysis functionality', () => {
    let user;
    beforeEach(() => {
      user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    });

    it('renders AI log analysis when AI features are enabled', async () => {
      render(<DeploymentReport type="finished" open={true} />, {
        preloadedState: {
          ...defaultState,
          app: {
            ...defaultState.app,
            features: {
              ...defaultState.app.features,
              hasAiEnabled: true,
              isHosted: true
            }
          },
          deployments: deploymentsState
        }
      });
      await user.click(screen.getByRole('button', { name: /View log/i }));
      expect(await screen.findByText('AI summary (experimental)')).toBeInTheDocument();
      expect(await screen.getByRole('button', { name: /Generate summary/i })).toBeInTheDocument();
    });

    it('successfully generates and displays AI analysis', async () => {
      render(<DeploymentReport type="finished" open={true} />, {
        preloadedState: {
          ...defaultState,
          app: {
            ...defaultState.app,
            features: {
              ...defaultState.app.features,
              hasAiEnabled: true,
              isHosted: true
            }
          },
          deployments: deploymentsState,
          users: usersState
        }
      });
      await user.click(screen.getByRole('button', { name: /View log/i }));
      await waitFor(() => expect(screen.queryByText(/generate summary/i)).toBeInTheDocument(), { timeout: TIMEOUTS.threeSeconds });
      await user.click(screen.getByRole('button', { name: /Generate summary/i }));
      expect(screen.getByText('Generating summary...')).toBeInTheDocument();

      await waitFor(() => expect(screen.getByText('Summary of Deployment Failure:')).toBeInTheDocument(), { timeout: TIMEOUTS.fiveSeconds });
      expect(screen.getByText(/Network connectivity issues/)).toBeVisible();
    });
    it('handles rate limit error correctly', async () => {
      render(<DeploymentReport type="finished" open={true} />, {
        preloadedState: {
          ...defaultState,
          app: {
            ...defaultState.app,
            features: {
              ...defaultState.app.features,
              hasAiEnabled: true,
              isHosted: true
            }
          },
          deployments: {
            ...deploymentsState,
            selectionState: {
              selectedId: 'rate-limited-deployment'
            },
            byId: {
              ...defaultState.deployments.byId,
              'rate-limited-deployment': { ...deploymentsState.byId.d1, id: 'rate-limited-deployment' }
            }
          },
          users: usersState
        }
      });
      await user.click(screen.getByRole('button', { name: /View log/i }));
      await waitFor(() => expect(screen.queryByText(/generate summary/i)).toBeInTheDocument(), { timeout: TIMEOUTS.threeSeconds });
      await user.click(screen.getByRole('button', { name: /Generate summary/i }));

      await waitFor(() => expect(screen.getByText(/You have reached your limit/)).toBeInTheDocument(), {
        timeout: TIMEOUTS.fiveSeconds
      });
    });

    it('handles feedback submission correctly', async () => {
      render(<DeploymentReport type="finished" open={true} />, {
        preloadedState: {
          ...defaultState,
          app: {
            ...defaultState.app,
            features: {
              ...defaultState.app.features,
              hasAiEnabled: true,
              isHosted: true
            }
          },
          deployments: deploymentsState,
          users: usersState
        }
      });

      await user.click(await screen.findByText('View log'));
      await waitFor(() => expect(screen.queryByText(/generate summary/i)).toBeInTheDocument(), { timeout: TIMEOUTS.threeSeconds });
      await user.click(await screen.findByText('Generate summary'));
      await waitFor(() => expect(screen.getByText('Summary of Deployment Failure:')).toBeInTheDocument(), { timeout: TIMEOUTS.fiveSeconds });

      await user.click(screen.getByRole('button', { name: /thumbs-up/i }));

      expect(screen.getByText('Thank you for your feedback!')).toBeVisible();
      expect(screen.queryByText('Was this helpful?')).not.toBeInTheDocument();
    });
  });
});
