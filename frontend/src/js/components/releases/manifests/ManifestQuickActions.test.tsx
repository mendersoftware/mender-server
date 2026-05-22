// Copyright 2026 Northern.tech AS
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
import * as StoreThunks from '@northern.tech/store/thunks';
import { undefineds } from '@northern.tech/testing/mockData';
import { act, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import ManifestQuickActions from './ManifestQuickActions';

describe('ManifestQuickActions Component', () => {
  it('renders correctly', async () => {
    const preloadedState = {
      ...defaultState,
      releases: {
        ...defaultState.releases,
        manifestsList: {
          ...defaultState.releases.manifestsList,
          manifestIds: [defaultState.releases.manifestsById.m1000.name],
          selection: [0]
        }
      }
    };
    const { baseElement } = render(<ManifestQuickActions />, { preloadedState });
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('renders the speed dial with actions when manifests are selected', async () => {
    const preloadedState = {
      ...defaultState,
      releases: {
        ...defaultState.releases,
        manifestsList: {
          ...defaultState.releases.manifestsList,
          manifestIds: [defaultState.releases.manifestsById.m1000.name],
          selection: [0]
        }
      }
    };
    render(<ManifestQuickActions />, { preloadedState });
    expect(screen.getByLabelText('manifest-actions')).toBeInTheDocument();
    expect(screen.getByText(/1 Manifest selected/i)).toBeInTheDocument();
  });

  it('triggers manifest removal after confirming the dialog', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const { removeManifests: removeManifestsSpy } = StoreThunks;
    const manifestName = defaultState.releases.manifestsById.m1000.name;
    const preloadedState = {
      ...defaultState,
      releases: {
        ...defaultState.releases,
        selectedManifest: manifestName
      }
    };
    const { container } = render(<ManifestQuickActions />, { preloadedState });
    await user.click(container.querySelector('.MuiSpeedDial-fab') as Element);
    await user.click(screen.getByLabelText('delete'));
    await waitFor(() => expect(screen.queryByRole('button', { name: /remove/i })).toBeInTheDocument());
    await user.click(screen.getByRole('button', { name: /remove/i }));
    await act(async () => {
      vi.runOnlyPendingTimers();
      vi.runAllTicks();
    });
    expect(removeManifestsSpy).toHaveBeenCalledWith([manifestName]);
  });
});
